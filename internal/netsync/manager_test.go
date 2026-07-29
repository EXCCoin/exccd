package netsync

import (
	"encoding/binary"
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	peerpkg "github.com/EXCCoin/exccd/peer/v3"
)

func fullRequestMap(size int) map[chainhash.Hash]struct{} {
	m := make(map[chainhash.Hash]struct{}, size)
	for i := 0; i < size; i++ {
		var hash chainhash.Hash
		binary.LittleEndian.PutUint64(hash[:], uint64(i))
		m[hash] = struct{}{}
	}
	return m
}

func TestRequestFromPeerHonorsRequestLimits(t *testing.T) {
	p := peerpkg.NewInboundPeer(&peerpkg.Config{})
	sp := &syncMgrPeer{
		Peer:            p,
		requestedBlocks: make(map[chainhash.Hash]struct{}),
		requestedTxns:   make(map[chainhash.Hash]struct{}),
	}
	m := SyncManager{
		requestedBlocks: fullRequestMap(maxRequestedBlocks),
		requestedTxns:   fullRequestMap(maxRequestedTxns),
		peers:           map[*peerpkg.Peer]*syncMgrPeer{p: sp},
	}

	requests := []chainhash.Hash{{0xff}}
	if err := m.requestFromPeer(p, requests, requests, requests); err != nil {
		t.Fatal(err)
	}
	if got := len(m.requestedBlocks); got != maxRequestedBlocks {
		t.Fatalf("requested blocks grew to %d", got)
	}
	if got := len(m.requestedTxns); got != maxRequestedTxns {
		t.Fatalf("requested transactions grew to %d", got)
	}
}
