// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/EXCCoin/exccd/connmgr/v3"
	"github.com/EXCCoin/exccd/peer/v3"
)

func emptyPeerState() *peerState {
	return &peerState{
		inboundPeers:    make(map[int32]*serverPeer),
		outboundPeers:   make(map[int32]*serverPeer),
		persistentPeers: make(map[int32]*serverPeer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
		subCache: &naSubmissionCache{
			cache: make(map[string]*naSubmission),
			limit: 1,
		},
	}
}

// TestServerDoesNotReapplyConnectionLimits ensures the protocol layer does
// not reject connections already admitted by connmgr.  In particular,
// persistent connections are exempt from the normal total limit and outbound
// whitelisted connections are exempt from the per-host limit.
func TestServerDoesNotReapplyConnectionLimits(t *testing.T) {
	oldCfg := cfg
	cfg = &config{
		MaxPeers:      1,
		MaxSameIP:     1,
		DisableListen: true,
	}
	defer func() { cfg = oldCfg }()

	existingPeer, err := peer.NewOutboundPeer(&peer.Config{}, "10.0.0.1:9108")
	if err != nil {
		t.Fatal(err)
	}
	state := emptyPeerState()
	state.outboundPeers[1] = &serverPeer{Peer: existingPeer}

	admittedPeer, err := peer.NewOutboundPeer(&peer.Config{}, "10.0.0.1:9109")
	if err != nil {
		t.Fatal(err)
	}
	sp := &serverPeer{
		Peer:          admittedPeer,
		server:        &server{},
		persistent:    true,
		isWhitelisted: true,
	}
	if !sp.server.handleAddPeerMsg(state, sp) {
		t.Fatal("server rejected a connection already admitted by connmgr")
	}
	if state.persistentPeers[sp.ID()] != sp {
		t.Fatal("admitted persistent peer was not registered")
	}
}

// TestManualConnectDefersLimitsToConnManager ensures the RPC path does not use
// the late peer-state count, which excludes pending attempts and cannot apply
// connmgr whitelist semantics.
func TestManualConnectDefersLimitsToConnManager(t *testing.T) {
	oldCfg := cfg
	cfg = &config{
		MaxPeers: 1,
		lookup: func(host string) ([]net.IP, error) {
			return []net.IP{net.ParseIP(host)}, nil
		},
	}
	defer func() { cfg = oldCfg }()

	cm, err := connmgr.New(&connmgr.Config{
		Dial: func(context.Context, string, string) (net.Conn, error) {
			return nil, errors.New("test dial failure")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		cm.Run(ctx)
		close(done)
	}()
	defer func() {
		cancel()
		<-done
	}()

	state := emptyPeerState()
	state.inboundPeers[1] = &serverPeer{}
	s := &server{connManager: cm}
	reply := make(chan error, 1)
	s.handleQuery(state, connectNodeMsg{
		addr:  "10.0.0.2:9108",
		reply: reply,
	})
	if err := <-reply; err != nil {
		t.Fatalf("manual connect was rejected by late server state: %v", err)
	}
}
