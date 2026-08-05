// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"net"
	"sync"
)

const (
	defaultMaxPerOutboundGroup = uint32(1)
	maxOutboundAddrTries       = 100
)

var errNoSuitableOutboundAddr = errors.New("no suitable outbound address")

type outboundGroupState struct {
	key [32]byte

	mu     sync.Mutex
	counts map[[32]byte]uint32
}

func newOutboundGroupState() (*outboundGroupState, error) {
	g := &outboundGroupState{counts: make(map[[32]byte]uint32)}
	if _, err := io.ReadFull(rand.Reader, g.key[:]); err != nil {
		return nil, err
	}
	return g, nil
}

func (cm *ConnManager) outboundGroupKey(addr net.Addr) [32]byte {
	group := cm.cfg.OutboundGroup(addr)
	mac := hmac.New(sha256.New, cm.outboundGroups.key[:])
	_, _ = mac.Write([]byte(group))
	var key [32]byte
	copy(key[:], mac.Sum(nil))
	return key
}

func (cm *ConnManager) releaseOutboundGroup(key [32]byte) {
	g := cm.outboundGroups
	g.mu.Lock()
	count := g.counts[key]
	if count <= 1 {
		delete(g.counts, key)
	} else {
		g.counts[key] = count - 1
	}
	g.mu.Unlock()
}

// reserveOutboundGroup tracks a manual or persistent outbound attempt.  These
// attempts are not themselves rejected by group policy, but automatic address
// selection avoids their groups.
func (cm *ConnManager) reserveOutboundGroup(addr net.Addr) func() {
	if cm.cfg.OutboundGroup == nil {
		return nil
	}
	key := cm.outboundGroupKey(addr)
	g := cm.outboundGroups
	g.mu.Lock()
	g.counts[key]++
	g.mu.Unlock()
	return func() { cm.releaseOutboundGroup(key) }
}

// pickOutboundAddr selects and atomically reserves an automatic outbound group.
func (cm *ConnManager) pickOutboundAddr() (net.Addr, func(), error) {
	if cm.cfg.OutboundGroup == nil {
		addr, err := cm.cfg.GetNewAddress()
		return addr, nil, err
	}

	g := cm.outboundGroups
	g.mu.Lock()
	defer g.mu.Unlock()

	for i := 0; i < maxOutboundAddrTries; i++ {
		addr, err := cm.cfg.GetNewAddress()
		if err != nil {
			return nil, nil, err
		}
		key := cm.outboundGroupKey(addr)
		if g.counts[key] >= defaultMaxPerOutboundGroup {
			continue
		}
		g.counts[key]++
		return addr, func() { cm.releaseOutboundGroup(key) }, nil
	}

	return nil, nil, errNoSuitableOutboundAddr
}
