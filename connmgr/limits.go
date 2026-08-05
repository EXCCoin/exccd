// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
)

// semaphore is a context-aware counting semaphore.
type semaphore chan struct{}

func makeSemaphore(capacity uint32) semaphore {
	return make(chan struct{}, capacity)
}

func (s semaphore) acquire(ctx context.Context, quit <-chan struct{}) bool {
	select {
	case s <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	case <-quit:
		return false
	}
}

func (s semaphore) tryAcquire() bool {
	select {
	case s <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s semaphore) release() {
	<-s
}

// connReservation owns release callbacks for a single connection attempt.  It
// permits resources to be added while an attempt is being prepared and still
// guarantees they are each released exactly once when preparation races with
// cancellation or shutdown.
type connReservation struct {
	mu       sync.Mutex
	released bool
	releases []func()
}

func (r *connReservation) add(release func()) {
	if release == nil {
		return
	}

	r.mu.Lock()
	if r.released {
		r.mu.Unlock()
		release()
		return
	}
	r.releases = append(r.releases, release)
	r.mu.Unlock()
}

func (r *connReservation) release() {
	r.mu.Lock()
	if r.released {
		r.mu.Unlock()
		return
	}
	r.released = true
	releases := r.releases
	r.releases = nil
	r.mu.Unlock()

	for i := len(releases) - 1; i >= 0; i-- {
		releases[i]()
	}
}

// limitedConn releases its admission reservation when the connection closes.
type limitedConn struct {
	net.Conn
	reservation *connReservation
}

func (c *limitedConn) Close() error {
	err := c.Conn.Close()
	c.reservation.release()
	return err
}

// addrHostKey returns a stable host key and parsed IP for an address.
func addrHostKey(addr net.Addr) (string, net.IP, error) {
	if addr == nil {
		return "", nil, fmt.Errorf("nil network address")
	}
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		ip := tcpAddr.IP
		if ip4 := ip.To4(); ip4 != nil {
			ip = ip4
		}
		return ip.String(), ip, nil
	}

	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", nil, err
	}
	if zone := strings.LastIndexByte(host, '%'); zone >= 0 {
		host = host[:zone]
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			ip = ip4
		}
		return ip.String(), ip, nil
	}
	return strings.ToLower(host), nil, nil
}

func (cm *ConnManager) isWhitelisted(addr net.Addr) bool {
	return cm.cfg.IsWhitelisted != nil && cm.cfg.IsWhitelisted(addr)
}

// reserveHost reserves a per-host connection slot unless the address is
// exempt.  The returned callback releases the slot.
func (cm *ConnManager) reserveHost(addr net.Addr) (func(), error) {
	if cm.cfg.MaxConnsPerHost == 0 || cm.isWhitelisted(addr) {
		return nil, nil
	}

	host, ip, err := addrHostKey(addr)
	if err != nil {
		return nil, err
	}
	if ip != nil && ip.IsLoopback() {
		return nil, nil
	}

	cm.hostMtx.Lock()
	count := cm.hostCounts[host]
	if count >= cm.cfg.MaxConnsPerHost {
		cm.hostMtx.Unlock()
		return nil, MakeError(ErrMaxConnsPerHost, fmt.Sprintf(
			"a maximum of %d connections per host is allowed",
			cm.cfg.MaxConnsPerHost))
	}
	cm.hostCounts[host] = count + 1
	cm.hostMtx.Unlock()

	return func() {
		cm.hostMtx.Lock()
		count := cm.hostCounts[host]
		if count <= 1 {
			delete(cm.hostCounts, host)
		} else {
			cm.hostCounts[host] = count - 1
		}
		cm.hostMtx.Unlock()
	}, nil
}

// reserveOutbound reserves all admission resources for an outbound attempt.
// Automatic connections wait for an overall slot while manual connections
// fail quickly when the limit is reached.
func (cm *ConnManager) reserveOutbound(ctx context.Context, c *ConnReq,
	waitNormal bool) (*connReservation, error) {

	r := new(connReservation)
	if !c.Permanent {
		var acquired bool
		if waitNormal {
			acquired = cm.normalSem.acquire(ctx, cm.quit)
		} else {
			acquired = cm.normalSem.tryAcquire()
		}
		if !acquired {
			return nil, MakeError(ErrMaxNormalConns, fmt.Sprintf(
				"a maximum of %d normal connections is allowed",
				cm.cfg.MaxNormalConns))
		}
		r.add(cm.normalSem.release)
	}

	if c.Addr != nil {
		r.add(cm.reserveOutboundGroup(c.Addr))
		releaseHost, err := cm.reserveHost(c.Addr)
		if err != nil {
			r.release()
			return nil, err
		}
		r.add(releaseHost)
	}
	return r, nil
}

// reserveInbound reserves admission resources before an accepted connection is
// handed to the protocol layer.
func (cm *ConnManager) reserveInbound(conn net.Conn) (*connReservation, error) {
	r := new(connReservation)
	remoteAddr := conn.RemoteAddr()

	releaseHost, err := cm.reserveHost(remoteAddr)
	if err != nil {
		return nil, err
	}
	r.add(releaseHost)

	if !cm.isWhitelisted(remoteAddr) {
		if !cm.normalSem.tryAcquire() {
			r.release()
			return nil, MakeError(ErrMaxNormalConns, fmt.Sprintf(
				"a maximum of %d normal connections is allowed",
				cm.cfg.MaxNormalConns))
		}
		r.add(cm.normalSem.release)
	}
	return r, nil
}
