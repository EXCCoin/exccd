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

func ipv4Slash16Group(addr net.Addr) string {
	tcpAddr := addr.(*net.TCPAddr)
	return tcpAddr.IP.To4().Mask(net.CIDRMask(16, 32)).String()
}

// TestPendingOutboundGroupReservation ensures concurrent automatic address
// selection accounts for pending attempts, not only established peers.
func TestPendingOutboundGroupReservation(t *testing.T) {
	addrs := []*net.TCPAddr{
		{IP: net.ParseIP("10.1.0.1"), Port: 1001},
		{IP: net.ParseIP("10.1.0.2"), Port: 1002},
		{IP: net.ParseIP("10.2.0.1"), Port: 1003},
	}
	var addrIndex uint32
	getAddr := func() (net.Addr, error) {
		i := atomic.AddUint32(&addrIndex, 1) - 1
		if int(i) >= len(addrs) {
			return addrs[len(addrs)-1], nil
		}
		return addrs[i], nil
	}

	firstGroupStarted := make(chan struct{})
	releaseFirstGroup := make(chan struct{})
	var startedOnce sync.Once
	dialer := func(ctx context.Context, addr net.Addr) (net.Conn, error) {
		if ipv4Slash16Group(addr) == "10.1.0.0" {
			startedOnce.Do(func() { close(firstGroupStarted) })
			<-releaseFirstGroup
		}
		return mockDialerAddr(ctx, addr)
	}
	connected := make(chan *ConnReq, 2)
	cm, err := New(&Config{
		TargetOutbound: 2,
		DialAddr:       dialer,
		GetNewAddress:  getAddr,
		OutboundGroup:  ipv4Slash16Group,
		OnConnection: func(req *ConnReq, conn net.Conn) {
			connected <- req
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		close(releaseFirstGroup)
		shutdown()
		wg.Wait()
	}()

	select {
	case <-firstGroupStarted:
	case <-time.After(admissionTestTimeout):
		t.Fatal("first outbound group did not start dialing")
	}
	select {
	case req := <-connected:
		if got := ipv4Slash16Group(req.Addr); got != "10.2.0.0" {
			t.Fatalf("connected pending duplicate group %q", got)
		}
	case <-time.After(admissionTestTimeout):
		t.Fatal("second outbound group did not connect")
	}
	if got := atomic.LoadUint32(&addrIndex); got < 3 {
		t.Fatalf("address selector made %d calls, want at least 3", got)
	}
}

// TestOutboundGroupReleasedAfterFailure ensures a failed pending attempt does
// not permanently block its network group.
func TestOutboundGroupReleasedAfterFailure(t *testing.T) {
	var dials uint32
	dialer := func(ctx context.Context, addr net.Addr) (net.Conn, error) {
		if atomic.AddUint32(&dials, 1) == 1 {
			return nil, errors.New("first dial failed")
		}
		return mockDialerAddr(ctx, addr)
	}
	var ports uint32
	connected := make(chan *ConnReq, 1)
	cm, err := New(&Config{
		TargetOutbound: 1,
		RetryDuration:  time.Millisecond,
		DialAddr:       dialer,
		GetNewAddress: func() (net.Addr, error) {
			port := atomic.AddUint32(&ports, 1)
			return &net.TCPAddr{IP: net.ParseIP("10.1.0.1"), Port: int(port)}, nil
		},
		OutboundGroup: ipv4Slash16Group,
		OnConnection: func(req *ConnReq, conn net.Conn) {
			connected <- req
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		shutdown()
		wg.Wait()
	}()

	select {
	case <-connected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("failed outbound group reservation was not released")
	}
	if got := atomic.LoadUint32(&dials); got != 2 {
		t.Fatalf("got %d dial attempts, want 2", got)
	}
}

// TestOutboundGroupReleasedAfterCancellation ensures canceling a blocked
// pending attempt immediately makes its group available to another attempt.
func TestOutboundGroupReleasedAfterCancellation(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var dials uint32
	dialer := func(ctx context.Context, addr net.Addr) (net.Conn, error) {
		if atomic.AddUint32(&dials, 1) == 1 {
			close(firstStarted)
			<-releaseFirst
			return nil, errors.New("canceled")
		}
		return mockDialerAddr(ctx, addr)
	}
	var ports uint32
	connected := make(chan *ConnReq, 1)
	cm, err := New(&Config{
		TargetOutbound: 1,
		DialAddr:       dialer,
		GetNewAddress: func() (net.Addr, error) {
			port := atomic.AddUint32(&ports, 1)
			return &net.TCPAddr{IP: net.ParseIP("10.1.0.1"), Port: int(port)}, nil
		},
		OutboundGroup: ipv4Slash16Group,
		OnConnection: func(req *ConnReq, conn net.Conn) {
			connected <- req
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		close(releaseFirst)
		shutdown()
		wg.Wait()
	}()

	select {
	case <-firstStarted:
	case <-time.After(admissionTestTimeout):
		t.Fatal("first outbound attempt did not start")
	}
	var pendingID uint64
	err = cm.ForEachConnReq(func(req *ConnReq) error {
		if req.State() == ConnPending {
			pendingID = req.ID()
		}
		return nil
	})
	if err != nil || pendingID == 0 {
		t.Fatalf("unable to find pending request: %v", err)
	}
	cm.Remove(pendingID)
	go cm.newConnReq(ctx)

	select {
	case <-connected:
	case <-time.After(admissionTestTimeout):
		t.Fatal("canceled outbound group reservation was not released")
	}
}
