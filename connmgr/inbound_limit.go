// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"container/list"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// Keep the cache bounded under prefix-spraying attacks.  The TTL lets old
	// groups age out during normal operation.
	maxGroupLimiters = 10000
	maxPerGroupTTL   = 20 * time.Minute

	// Permit one connection per five seconds on average with bursts of three
	// from each inbound group.
	groupRateLimit  = 0.2
	groupBurstLimit = 3

	// Permit one drop log per minute on average with bursts of four.
	dropLogRateLimit  = float64(1) / 60
	dropLogBurstLimit = 4

	// Enter flood mode above an average of five admitted attempts per second
	// over the previous minute.  Reach maximum pressure at four times that
	// threshold.
	floodLow        = 5 * 60
	floodHighFactor = 3

	floodMinDropProb = 0.2
	floodMaxDropProb = 0.85
	floodRamp        = 0.1
)

// tokenBucket is a small concurrency-safe token bucket.
type tokenBucket struct {
	mu sync.Mutex

	rate   float64
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

func newTokenBucket(rate float64, burst uint32, now func() time.Time) *tokenBucket {
	return &tokenBucket{
		rate:   rate,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   now(),
		now:    now,
	}
}

// refill replenishes tokens according to elapsed time.  The caller must hold
// the bucket mutex.
func (b *tokenBucket) refill(now time.Time) {
	if now.After(b.last) {
		b.tokens += now.Sub(b.last).Seconds() * b.rate
		if b.tokens > b.burst {
			b.tokens = b.burst
		}
		b.last = now
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill(b.now())
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (b *tokenBucket) untilNextAllowed() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill(b.now())
	if b.tokens >= 1 || b.rate <= 0 {
		return 0
	}
	return time.Duration((1 - b.tokens) / b.rate * float64(time.Second))
}

type inboundGroupKey [32]byte

type inboundGroupEntry struct {
	key     inboundGroupKey
	limiter *tokenBucket
	expires time.Time
}

// inboundGroupCache is a bounded TTL LRU.  It is externally synchronized by
// inboundRateLimiter.groupMtx.
type inboundGroupCache struct {
	maxEntries int
	ttl        time.Duration
	entries    map[inboundGroupKey]*list.Element
	recent     list.List
}

func newInboundGroupCache(maxEntries int, ttl time.Duration) *inboundGroupCache {
	return &inboundGroupCache{
		maxEntries: maxEntries,
		ttl:        ttl,
		entries:    make(map[inboundGroupKey]*list.Element),
	}
}

func (c *inboundGroupCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*inboundGroupEntry)
	delete(c.entries, entry.key)
	c.recent.Remove(elem)
}

func (c *inboundGroupCache) get(key inboundGroupKey, now time.Time) *tokenBucket {
	elem := c.entries[key]
	if elem == nil {
		return nil
	}
	entry := elem.Value.(*inboundGroupEntry)
	if !now.Before(entry.expires) {
		c.removeElement(elem)
		return nil
	}
	entry.expires = now.Add(c.ttl)
	c.recent.MoveToFront(elem)
	return entry.limiter
}

func (c *inboundGroupCache) put(key inboundGroupKey, limiter *tokenBucket,
	now time.Time) {

	if elem := c.entries[key]; elem != nil {
		entry := elem.Value.(*inboundGroupEntry)
		entry.limiter = limiter
		entry.expires = now.Add(c.ttl)
		c.recent.MoveToFront(elem)
		return
	}

	// Entries are ordered by most recent access, so expired entries are
	// contiguous at the back.
	for elem := c.recent.Back(); elem != nil; elem = c.recent.Back() {
		entry := elem.Value.(*inboundGroupEntry)
		if now.Before(entry.expires) {
			break
		}
		c.removeElement(elem)
	}
	if c.maxEntries <= 0 {
		return
	}
	for len(c.entries) >= c.maxEntries {
		c.removeElement(c.recent.Back())
	}

	entry := &inboundGroupEntry{
		key:     key,
		limiter: limiter,
		expires: now.Add(c.ttl),
	}
	c.entries[key] = c.recent.PushFront(entry)
}

// inboundRateLimiter tracks bounded per-group rate limits and recent admitted
// attempts used to activate additional flood pressure.
type inboundRateLimiter struct {
	key        [32]byte
	now        func() time.Time
	burstLimit uint32

	randomCounter uint64
	randomFloat   func() float64

	groupMtx      sync.Mutex
	groupLimiters *inboundGroupCache

	floodMtx      sync.RWMutex
	attempts      [60]uint32
	attemptsStart int64
	totalAttempts uint64
	flooding      uint32

	logMtx      sync.Mutex
	logLimiter  *tokenBucket
	droppedLogs uint64
	schedule    func(time.Duration, func())
}

func newInboundRateLimiter(key [32]byte) *inboundRateLimiter {
	l := &inboundRateLimiter{
		key:           key,
		now:           time.Now,
		burstLimit:    groupBurstLimit,
		groupLimiters: newInboundGroupCache(maxGroupLimiters, maxPerGroupTTL),
		schedule: func(delay time.Duration, callback func()) {
			time.AfterFunc(delay, callback)
		},
	}
	l.randomFloat = l.secureFloat64
	l.logLimiter = newTokenBucket(dropLogRateLimit, dropLogBurstLimit, l.now)
	return l
}

// secureFloat64 returns a pseudorandom float in [0,1).  The key is generated
// from crypto/rand when the connection manager is created.  Domain-separated
// HMAC output avoids another fallible random read on the accept hot path.
func (l *inboundRateLimiter) secureFloat64() float64 {
	counter := atomic.AddUint64(&l.randomCounter, 1)
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)
	mac := hmac.New(sha256.New, l.key[:])
	_, _ = mac.Write([]byte("inbound-flood\x00"))
	_, _ = mac.Write(counterBytes[:])
	sum := mac.Sum(nil)
	random53 := binary.BigEndian.Uint64(sum[:8]) >> 11
	return float64(random53) / float64(uint64(1)<<53)
}

// groupKey returns a keyed hash of the inbound rate-limit group.  IPv4 uses
// /32 and IPv6 uses /64 normally.  Flood mode coarsens them to /24 and /56.
func (l *inboundRateLimiter) groupKey(addr net.Addr) (inboundGroupKey, error) {
	_, ip, err := addrHostKey(addr)
	if err != nil {
		return inboundGroupKey{}, err
	}

	flooding := atomic.LoadUint32(&l.flooding) != 0
	var preimage []byte
	if ip4 := ip.To4(); ip4 != nil {
		if flooding {
			ip4 = ip4.Mask(net.CIDRMask(24, 32))
		}
		preimage = ip4
	} else if ip16 := ip.To16(); ip16 != nil {
		bits := 64
		if flooding {
			bits = 56
		}
		preimage = ip16.Mask(net.CIDRMask(bits, 128))
	} else {
		preimage = []byte("unknown")
	}

	mac := hmac.New(sha256.New, l.key[:])
	_, _ = mac.Write([]byte("inbound-group\x00"))
	_, _ = mac.Write(preimage)
	var key inboundGroupKey
	copy(key[:], mac.Sum(nil))
	return key, nil
}

// recordAttempt advances the one-minute sliding window and records attempts
// that passed the per-group limiter.
func (l *inboundRateLimiter) recordAttempt(rateLimited bool) {
	const buckets = int64(len(l.attempts))
	nowUnix := l.now().Unix()
	idx := nowUnix % buckets
	if idx < 0 {
		idx += buckets
	}

	l.floodMtx.Lock()
	if l.attemptsStart != nowUnix {
		numExpired := nowUnix - l.attemptsStart
		if numExpired >= buckets || l.attemptsStart == 0 {
			for i := range l.attempts {
				l.attempts[i] = 0
			}
			l.totalAttempts = 0
		} else if numExpired > 0 {
			tail := l.attemptsStart + 1
			for i := int64(0); i < numExpired; i++ {
				oldIdx := (tail + i) % buckets
				l.totalAttempts -= uint64(l.attempts[oldIdx])
				l.attempts[oldIdx] = 0
			}
		}
		l.attemptsStart = nowUnix
	}
	if !rateLimited {
		l.attempts[idx]++
		l.totalAttempts++
	}
	if l.totalAttempts > floodLow {
		atomic.StoreUint32(&l.flooding, 1)
	} else {
		atomic.StoreUint32(&l.flooding, 0)
	}
	l.floodMtx.Unlock()
}

// allow applies the direct per-group rate limit and updates flood detection.
func (l *inboundRateLimiter) allow(addr net.Addr) bool {
	key, err := l.groupKey(addr)
	if err != nil {
		return false
	}

	now := l.now()
	l.groupMtx.Lock()
	limiter := l.groupLimiters.get(key, now)
	if limiter == nil {
		limiter = newTokenBucket(groupRateLimit, l.burstLimit, l.now)
		l.groupLimiters.put(key, limiter, now)
	}
	l.groupMtx.Unlock()

	allowed := limiter.allow()
	l.recordAttempt(!allowed)
	return allowed
}

// shouldDropProbabilistic applies a smooth quadratic S-curve from 20% to 85%
// as admitted connection pressure rises during a flood.
func (l *inboundRateLimiter) shouldDropProbabilistic() bool {
	if atomic.LoadUint32(&l.flooding) == 0 {
		return false
	}

	l.floodMtx.RLock()
	totalAttempts := l.totalAttempts
	l.floodMtx.RUnlock()
	if totalAttempts < floodLow {
		totalAttempts = floodLow
	}

	const factor = (1 + floodRamp) * (floodMaxDropProb - floodMinDropProb)
	const rampWidth = floodLow * floodHighFactor
	norm := float64(totalAttempts-floodLow) / float64(rampWidth)
	if norm > 1 {
		norm = 1
	}
	nSquared := norm * norm
	probability := floodMinDropProb + factor*nSquared/(floodRamp+nSquared)
	return l.randomFloat() < probability
}

func (l *inboundRateLimiter) flushSuppressedDropLogs() {
	l.logMtx.Lock()
	dropped := l.droppedLogs
	l.droppedLogs = 0
	l.logMtx.Unlock()

	// The first suppressed drop was already included in the message that
	// announced suppression.
	if dropped > 1 {
		log.Debugf("Dropped %d connections while suppressed", dropped-1)
	}
}

// logDrop logs dropped connections without allowing an attacker to flood the
// logs.  The first suppressed message announces the suppression and a later
// callback summarizes additional drops.
func (l *inboundRateLimiter) logDrop(addr net.Addr, reason string) {
	l.logMtx.Lock()
	if l.logLimiter.allow() {
		l.logMtx.Unlock()
		log.Debugf("Dropped connection from %v: %v", addr, reason)
		return
	}

	l.droppedLogs++
	firstSuppressed := l.droppedLogs == 1
	wait := l.logLimiter.untilNextAllowed()
	l.logMtx.Unlock()
	if !firstSuppressed {
		return
	}

	log.Debugf("Dropped connection from %v: %v -- suppressing drop logs for %v",
		addr, reason, wait.Round(time.Second))
	l.schedule(wait, l.flushSuppressedDropLogs)
}
