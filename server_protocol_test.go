package main

import (
	"testing"
	"time"

	"github.com/EXCCoin/exccd/peer/v3"
	"github.com/EXCCoin/exccd/wire"
)

func TestInitialStateMessageSequencing(t *testing.T) {
	oldCfg := cfg
	cfg = &config{}
	defer func() { cfg = oldCfg }()

	tests := []struct {
		name    string
		pver    uint32
		prepare func(*serverPeer)
		invoke  func(*serverPeer)
	}{
		{
			name: "new protocol getminingstate",
			pver: wire.InitStateVersion,
			invoke: func(sp *serverPeer) {
				sp.OnGetMiningState(sp.Peer, wire.NewMsgGetMiningState())
			},
		},
		{
			name: "new protocol miningstate",
			pver: wire.InitStateVersion,
			invoke: func(sp *serverPeer) {
				sp.OnMiningState(sp.Peer, wire.NewMsgMiningState())
			},
		},
		{
			name: "repeated getminingstate",
			pver: wire.InitStateVersion - 1,
			prepare: func(sp *serverPeer) {
				sp.getMiningStateSent = true
			},
			invoke: func(sp *serverPeer) {
				sp.OnGetMiningState(sp.Peer, wire.NewMsgGetMiningState())
			},
		},
		{
			name: "repeated miningstate",
			pver: wire.InitStateVersion - 1,
			prepare: func(sp *serverPeer) {
				sp.initStateReceived = true
			},
			invoke: func(sp *serverPeer) {
				sp.OnMiningState(sp.Peer, wire.NewMsgMiningState())
			},
		},
		{
			name: "repeated getinitstate",
			pver: wire.InitStateVersion,
			prepare: func(sp *serverPeer) {
				sp.initStateSent = true
			},
			invoke: func(sp *serverPeer) {
				sp.OnGetInitState(sp.Peer, wire.NewMsgGetInitState())
			},
		},
		{
			name: "repeated initstate",
			pver: wire.InitStateVersion,
			prepare: func(sp *serverPeer) {
				sp.initStateReceived = true
			},
			invoke: func(sp *serverPeer) {
				sp.OnInitState(sp.Peer, wire.NewMsgInitState())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := peer.NewInboundPeer(&peer.Config{ProtocolVersion: test.pver})
			s := &server{banPeers: make(chan *serverPeer, 1)}
			sp := &serverPeer{Peer: p, server: s}
			if test.prepare != nil {
				test.prepare(sp)
			}

			test.invoke(sp)

			select {
			case banned := <-s.banPeers:
				if banned != sp {
					t.Fatal("wrong peer banned")
				}
			default:
				t.Fatal("peer was not banned")
			}

			disconnected := make(chan struct{})
			go func() {
				sp.WaitForDisconnect()
				close(disconnected)
			}()
			select {
			case <-disconnected:
			case <-time.After(time.Second):
				t.Fatal("peer was not disconnected")
			}
		})
	}
}
