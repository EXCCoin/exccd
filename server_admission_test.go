// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"net"
	"strconv"
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

// TestManualConnectReportsAdmissionErrors ensures the RPC path does not report
// success when connmgr rejects a request before dialing it.
func TestManualConnectReportsAdmissionErrors(t *testing.T) {
	oldCfg := cfg
	cfg = &config{
		MaxPeers: 1,
		lookup: func(host string) ([]net.IP, error) {
			return []net.IP{net.ParseIP(host)}, nil
		},
	}
	defer func() { cfg = oldCfg }()

	dialStarted := make(chan struct{})
	releaseDial := make(chan struct{})
	cm, err := connmgr.New(&connmgr.Config{
		MaxNormalConns: 1,
		Dial: func(context.Context, string, string) (net.Conn, error) {
			close(dialStarted)
			<-releaseDial
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
		close(releaseDial)
		cancel()
		<-done
	}()

	first := &connmgr.ConnReq{
		Addr: &net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 9108},
	}
	if err := cm.ConnectAsync(ctx, first); err != nil {
		t.Fatal(err)
	}
	select {
	case <-dialStarted:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for first dial")
	}

	s := &server{connManager: cm}
	reply := make(chan error, 1)
	s.handleQuery(emptyPeerState(), connectNodeMsg{
		addr:  "10.0.0.2:9108",
		reply: reply,
	})
	if err := <-reply; !errors.Is(err, connmgr.ErrMaxNormalConns) {
		t.Fatalf("manual connect error is %v, want %v", err,
			connmgr.ErrMaxNormalConns)
	}

	count := 0
	if err := cm.ForEachConnReq(func(*connmgr.ConnReq) error {
		count++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d registered requests, want 1", count)
	}
}

// TestPersistentPeerAddrs ensures startup configuration applies the same
// persistent total and per-host bounds as runtime RPC requests.
func TestPersistentPeerAddrs(t *testing.T) {
	oldCfg := cfg
	cfg = &config{
		lookup: func(host string) ([]net.IP, error) {
			return []net.IP{net.ParseIP(host)}, nil
		},
	}
	defer func() { cfg = oldCfg }()

	tooMany := make([]string, connmgr.MaxPersistent+1)
	for i := range tooMany {
		tooMany[i] = net.JoinHostPort("10.0.0."+strconv.Itoa(i+1), "9108")
	}
	_, err := persistentPeerAddrs(tooMany, 2)
	if !errors.Is(err, connmgr.ErrMaxPersistent) {
		t.Fatalf("total persistent error is %v, want %v", err,
			connmgr.ErrMaxPersistent)
	}

	_, err = persistentPeerAddrs([]string{
		"10.0.0.1:9108",
		"10.0.0.1:9109",
	}, 1)
	if !errors.Is(err, connmgr.ErrMaxPersistentPerHost) {
		t.Fatalf("per-host persistent error is %v, want %v", err,
			connmgr.ErrMaxPersistentPerHost)
	}

	addrs, err := persistentPeerAddrs([]string{
		"127.0.0.1:9108",
		"127.0.0.1:9109",
	}, 1)
	if err != nil {
		t.Fatalf("loopback peers should be exempt: %v", err)
	}
	if len(addrs) != 2 {
		t.Fatalf("got %d addresses, want 2", len(addrs))
	}
}
