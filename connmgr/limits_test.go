// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const admissionTestTimeout = 250 * time.Millisecond

func receiveAcceptedConn(t *testing.T, accepted <-chan net.Conn) net.Conn {
	t.Helper()
	select {
	case conn := <-accepted:
		return conn
	case <-time.After(admissionTestTimeout):
		t.Fatal("timeout waiting for accepted connection")
		return nil
	}
}

func assertNoAcceptedConn(t *testing.T, accepted <-chan net.Conn) {
	t.Helper()
	select {
	case conn := <-accepted:
		conn.Close()
		t.Fatal("unexpected accepted connection")
	case <-time.After(20 * time.Millisecond):
	}
}

func waitConnReqState(t *testing.T, req *ConnReq, want ConnState) {
	t.Helper()
	deadline := time.Now().Add(admissionTestTimeout)
	for time.Now().Before(deadline) {
		if req.State() == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("connection state is %v, want %v", req.State(), want)
}

// TestInboundAdmissionLimits ensures raw accepted sockets count against the
// overall limit before the protocol layer performs a version handshake.  It
// also ensures closing a connection releases exactly one slot.
func TestInboundAdmissionLimits(t *testing.T) {
	listener := newMockListener("127.0.0.1:9108")
	accepted := make(chan net.Conn, 4)
	cm, err := New(&Config{
		Listeners:      []net.Listener{listener},
		MaxNormalConns: 2,
		OnAccept: func(conn net.Conn) {
			accepted <- conn
		},
		Dial: mockDialer,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		shutdown()
		wg.Wait()
	}()

	listener.Connect("10.0.0.1", 1001)
	first := receiveAcceptedConn(t, accepted)
	listener.Connect("10.0.0.2", 1002)
	second := receiveAcceptedConn(t, accepted)

	listener.Connect("10.0.0.3", 1003)
	assertNoAcceptedConn(t, accepted)

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	listener.Connect("10.0.0.3", 1004)
	third := receiveAcceptedConn(t, accepted)

	// A duplicate close must not release another slot.
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	listener.Connect("10.0.0.4", 1005)
	assertNoAcceptedConn(t, accepted)

	second.Close()
	third.Close()
}

// TestInboundPerHostLimit ensures pending protocol handshakes are limited per
// host before they reach the accept callback.
func TestInboundPerHostLimit(t *testing.T) {
	listener := newMockListener("127.0.0.1:9108")
	accepted := make(chan net.Conn, 4)
	cm, err := New(&Config{
		Listeners:       []net.Listener{listener},
		MaxNormalConns:  10,
		MaxConnsPerHost: 2,
		OnAccept: func(conn net.Conn) {
			accepted <- conn
		},
		Dial: mockDialer,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Keep this test focused on per-host lifecycle behavior.
	cm.inboundLimiter.burstLimit = 100
	_, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		shutdown()
		wg.Wait()
	}()

	listener.Connect("10.0.0.1", 1001)
	first := receiveAcceptedConn(t, accepted)
	listener.Connect("10.0.0.1", 1002)
	second := receiveAcceptedConn(t, accepted)
	listener.Connect("10.0.0.1", 1003)
	assertNoAcceptedConn(t, accepted)

	listener.Connect("10.0.0.2", 1004)
	otherHost := receiveAcceptedConn(t, accepted)
	first.Close()
	listener.Connect("10.0.0.1", 1005)
	replacement := receiveAcceptedConn(t, accepted)

	second.Close()
	otherHost.Close()
	replacement.Close()
}

// TestOutboundReservationLifecycle ensures pending attempts consume the
// overall limit and cancellation, dial failure, active disconnect, and
// duplicate disconnect all release their reservation exactly once.
func TestOutboundReservationLifecycle(t *testing.T) {
	blockedDial := make(chan struct{})
	blockedStarted := make(chan struct{})
	var blockedOnce sync.Once
	var dials uint32
	dialer := func(ctx context.Context, addr net.Addr) (net.Conn, error) {
		atomic.AddUint32(&dials, 1)
		tcpAddr := addr.(*net.TCPAddr)
		switch tcpAddr.Port {
		case 1001:
			blockedOnce.Do(func() { close(blockedStarted) })
			<-blockedDial
			return nil, errors.New("canceled dial")
		case 1003:
			return nil, errors.New("dial failed")
		default:
			return mockDialerAddr(ctx, addr)
		}
	}
	connected := make(chan *ConnReq, 4)
	disconnected := make(chan *ConnReq, 4)
	cm, err := New(&Config{
		MaxNormalConns: 1,
		DialAddr:       dialer,
		OnConnection: func(req *ConnReq, conn net.Conn) {
			connected <- req
		},
		OnDisconnection: func(req *ConnReq) {
			disconnected <- req
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		close(blockedDial)
		shutdown()
		wg.Wait()
	}()

	newReq := func(port int) *ConnReq {
		return &ConnReq{Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: port}}
	}

	pending := newReq(1001)
	go cm.Connect(ctx, pending)
	select {
	case <-blockedStarted:
	case <-time.After(admissionTestTimeout):
		t.Fatal("pending dial did not start")
	}

	rejected := newReq(1002)
	go cm.Connect(ctx, rejected)
	waitConnReqState(t, rejected, ConnFailed)
	if got := atomic.LoadUint32(&dials); got != 1 {
		t.Fatalf("got %d dials while pending slot was occupied, want 1", got)
	}

	cm.Remove(pending.ID())
	waitConnReqState(t, pending, ConnCanceled)
	active := newReq(1004)
	go cm.Connect(ctx, active)
	select {
	case <-connected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("connection did not start after pending cancellation")
	}

	cm.Remove(active.ID())
	select {
	case <-disconnected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("active connection was not removed")
	}
	cm.Remove(active.ID()) // Must not release another slot.

	failed := newReq(1003)
	go cm.Connect(ctx, failed)
	waitConnReqState(t, failed, ConnFailed)

	last := newReq(1005)
	go cm.Connect(ctx, last)
	select {
	case <-connected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("dial failure did not release reservation")
	}

	overLimit := newReq(1006)
	go cm.Connect(ctx, overLimit)
	waitConnReqState(t, overLimit, ConnFailed)
	cm.Remove(last.ID())
}

// TestPermanentRetryReleasesHostPermit ensures failed and disconnected
// persistent attempts release their per-host permit before retrying.
func TestPermanentRetryReleasesHostPermit(t *testing.T) {
	var attempts uint32
	dialer := func(ctx context.Context, addr net.Addr) (net.Conn, error) {
		if atomic.AddUint32(&attempts, 1) == 1 {
			return nil, errors.New("first attempt failed")
		}
		return mockDialerAddr(ctx, addr)
	}
	connected := make(chan *ConnReq, 2)
	disconnected := make(chan *ConnReq, 1)
	cm, err := New(&Config{
		MaxConnsPerHost: 1,
		RetryDuration:   time.Millisecond,
		DialAddr:        dialer,
		OnConnection: func(req *ConnReq, conn net.Conn) {
			connected <- req
		},
		OnDisconnection: func(req *ConnReq) {
			disconnected <- req
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		shutdown()
		wg.Wait()
	}()

	req := &ConnReq{
		Addr:      &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1001},
		Permanent: true,
	}
	go cm.Connect(ctx, req)
	select {
	case <-connected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("persistent connection did not retry after failure")
	}

	cm.Disconnect(req.ID())
	select {
	case <-disconnected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("persistent connection did not disconnect")
	}
	select {
	case <-connected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("persistent connection did not reconnect")
	}
	cm.Remove(req.ID())
}
