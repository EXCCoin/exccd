// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type inboundTestClock struct {
	mu  sync.Mutex
	now time.Time
}

func newInboundTestClock() *inboundTestClock {
	return &inboundTestClock{now: time.Unix(2000000000, 0)}
}

func (c *inboundTestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *inboundTestClock) Advance(elapsed time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(elapsed)
	c.mu.Unlock()
}

func newTestInboundLimiter(clock *inboundTestClock) *inboundRateLimiter {
	var key [32]byte
	key[0] = 1
	limiter := newInboundRateLimiter(key)
	limiter.now = clock.Now
	limiter.groupLimiters = newInboundGroupCache(maxGroupLimiters,
		maxPerGroupTTL)
	limiter.logLimiter = newTokenBucket(dropLogRateLimit,
		dropLogBurstLimit, limiter.now)
	return limiter
}

func inboundTestAddr(ip string, port int) net.Addr {
	return &net.TCPAddr{IP: net.ParseIP(ip), Port: port}
}

// TestInboundTokenBucket ensures groups have independent burst capacity and
// replenish at the configured rate.  IPv6 addresses in the same /64 share a
// limiter under normal conditions.
func TestInboundTokenBucket(t *testing.T) {
	clock := newInboundTestClock()
	limiter := newTestInboundLimiter(clock)
	addr := inboundTestAddr("10.0.0.1", 1000)

	for i := 0; i < groupBurstLimit; i++ {
		if !limiter.allow(addr) {
			t.Fatalf("connection %d in initial burst was rate limited", i+1)
		}
	}
	if limiter.allow(addr) {
		t.Fatal("connection beyond burst limit was allowed")
	}

	clock.Advance(5 * time.Second)
	if !limiter.allow(addr) {
		t.Fatal("token did not replenish after five seconds")
	}

	// A different IPv4 address has an independent limiter.
	if !limiter.allow(inboundTestAddr("10.0.0.2", 1001)) {
		t.Fatal("different IPv4 group was unexpectedly rate limited")
	}

	firstIPv6 := inboundTestAddr("2001:db8:1:2::1", 1002)
	secondIPv6 := inboundTestAddr("2001:db8:1:2::2", 1003)
	for i := 0; i < groupBurstLimit; i++ {
		if !limiter.allow(firstIPv6) {
			t.Fatalf("IPv6 connection %d in initial burst was rate limited", i+1)
		}
	}
	if limiter.allow(secondIPv6) {
		t.Fatal("addresses in the same IPv6 /64 did not share a limiter")
	}
}

// TestInboundGroupCacheBoundedAndExpiring ensures prefix spraying cannot grow
// the limiter cache without bound and inactive entries expire.
func TestInboundGroupCacheBoundedAndExpiring(t *testing.T) {
	clock := newInboundTestClock()
	limiter := newTestInboundLimiter(clock)
	limiter.groupLimiters = newInboundGroupCache(2, time.Minute)

	addrA := inboundTestAddr("10.0.0.1", 1000)
	addrB := inboundTestAddr("10.0.0.2", 1001)
	addrC := inboundTestAddr("10.0.0.3", 1002)
	keyA, _ := limiter.groupKey(addrA)
	keyB, _ := limiter.groupKey(addrB)

	limiter.allow(addrA)
	limiter.allow(addrB)
	limiter.allow(addrA) // Make A most recently used.
	limiter.allow(addrC)

	limiter.groupMtx.Lock()
	_, hasA := limiter.groupLimiters.entries[keyA]
	_, hasB := limiter.groupLimiters.entries[keyB]
	cacheLen := len(limiter.groupLimiters.entries)
	limiter.groupMtx.Unlock()
	if cacheLen != 2 || !hasA || hasB {
		t.Fatalf("unexpected LRU state: len=%d hasA=%v hasB=%v",
			cacheLen, hasA, hasB)
	}

	clock.Advance(time.Minute + time.Second)
	limiter.allow(inboundTestAddr("10.0.0.4", 1003))
	limiter.groupMtx.Lock()
	cacheLen = len(limiter.groupLimiters.entries)
	limiter.groupMtx.Unlock()
	if cacheLen != 1 {
		t.Fatalf("expired cache contains %d entries, want 1", cacheLen)
	}
}

// TestInboundFloodPressure ensures flood mode coarsens groups, applies the
// configured probabilistic pressure, and recovers after the sliding window
// expires.
func TestInboundFloodPressure(t *testing.T) {
	clock := newInboundTestClock()
	limiter := newTestInboundLimiter(clock)
	first := inboundTestAddr("10.1.2.3", 1000)
	second := inboundTestAddr("10.1.2.4", 1001)

	firstKey, _ := limiter.groupKey(first)
	secondKey, _ := limiter.groupKey(second)
	if firstKey == secondKey {
		t.Fatal("normal IPv4 groups were unexpectedly coarsened")
	}

	for i := 0; i <= floodLow; i++ {
		limiter.recordAttempt(false)
	}
	if atomicFlood := atomicLoadFlood(limiter); !atomicFlood {
		t.Fatal("flood mode did not activate")
	}
	firstKey, _ = limiter.groupKey(first)
	secondKey, _ = limiter.groupKey(second)
	if firstKey != secondKey {
		t.Fatal("flood mode did not coarsen IPv4 groups to /24")
	}

	limiter.randomFloat = func() float64 { return 0 }
	if !limiter.shouldDropProbabilistic() {
		t.Fatal("connection below the minimum random threshold was not dropped")
	}
	limiter.randomFloat = func() float64 { return 0.999 }
	if limiter.shouldDropProbabilistic() {
		t.Fatal("connection above the maximum random threshold was dropped")
	}

	clock.Advance(61 * time.Second)
	limiter.recordAttempt(true)
	if atomicLoadFlood(limiter) {
		t.Fatal("flood mode did not clear after the window expired")
	}
	firstKey, _ = limiter.groupKey(first)
	secondKey, _ = limiter.groupKey(second)
	if firstKey == secondKey {
		t.Fatal("IPv4 groups remained coarsened after flood recovery")
	}
}

func atomicLoadFlood(limiter *inboundRateLimiter) bool {
	return atomic.LoadUint32(&limiter.flooding) != 0
}

// TestInboundDropLogSuppression ensures drop diagnostics are bounded and the
// scheduled summary resets suppression state.
func TestInboundDropLogSuppression(t *testing.T) {
	clock := newInboundTestClock()
	limiter := newTestInboundLimiter(clock)
	var callback func()
	var delay time.Duration
	var schedules int
	limiter.schedule = func(d time.Duration, f func()) {
		delay = d
		callback = f
		schedules++
	}
	addr := inboundTestAddr("10.0.0.1", 1000)

	for i := 0; i < dropLogBurstLimit; i++ {
		limiter.logDrop(addr, "test")
	}
	limiter.logDrop(addr, "test")
	limiter.logDrop(addr, "test")
	if schedules != 1 || callback == nil {
		t.Fatalf("got %d suppression schedules, want 1", schedules)
	}
	limiter.logMtx.Lock()
	dropped := limiter.droppedLogs
	limiter.logMtx.Unlock()
	if dropped != 2 {
		t.Fatalf("got %d suppressed drops, want 2", dropped)
	}

	clock.Advance(delay)
	callback()
	limiter.logMtx.Lock()
	dropped = limiter.droppedLogs
	limiter.logMtx.Unlock()
	if dropped != 0 {
		t.Fatalf("suppressed drop count was not reset: %d", dropped)
	}
	limiter.logDrop(addr, "test")
	if schedules != 1 {
		t.Fatal("log limiter did not replenish after suppression interval")
	}
}

// TestInboundRateLimitBeforeAccept ensures raw sockets are rejected by the
// limiter before the protocol callback receives them.
func TestInboundRateLimitBeforeAccept(t *testing.T) {
	listener := newMockListener("127.0.0.1:9108")
	accepted := make(chan net.Conn, groupBurstLimit+1)
	cm, err := New(&Config{
		Listeners:      []net.Listener{listener},
		MaxNormalConns: 20,
		OnAccept: func(conn net.Conn) {
			accepted <- conn
		},
		Dial: mockDialer,
	})
	if err != nil {
		t.Fatal(err)
	}
	clock := newInboundTestClock()
	cm.inboundLimiter = newTestInboundLimiter(clock)
	_, shutdown, wg := runConnMgrAsync(context.Background(), cm)
	defer func() {
		shutdown()
		wg.Wait()
	}()

	for i := 0; i < groupBurstLimit; i++ {
		listener.Connect("10.0.0.1", 1000+i)
		conn := receiveAcceptedConn(t, accepted)
		conn.Close()
	}
	listener.Connect("10.0.0.1", 2000)
	assertNoAcceptedConn(t, accepted)

	clock.Advance(5 * time.Second)
	listener.Connect("10.0.0.1", 2001)
	conn := receiveAcceptedConn(t, accepted)
	conn.Close()
}
