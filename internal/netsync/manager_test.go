package netsync

import (
	"encoding/binary"
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	peerpkg "github.com/EXCCoin/exccd/peer/v3"
	"github.com/EXCCoin/exccd/wire"
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

func TestDonePeerRerequestsKnownInventory(t *testing.T) {
	donePeer := peerpkg.NewInboundPeer(&peerpkg.Config{})
	backupPeer := peerpkg.NewInboundPeer(&peerpkg.Config{})
	txHash := chainhash.Hash{0x01}
	blockHash := chainhash.Hash{0x02}
	unknownTxHash := chainhash.Hash{0x03}
	unknownBlockHash := chainhash.Hash{0x04}
	backupPeer.AddKnownInventory(&wire.InvVect{
		Type: wire.InvTypeTx,
		Hash: txHash,
	})
	backupPeer.AddKnownInventory(&wire.InvVect{
		Type: wire.InvTypeBlock,
		Hash: blockHash,
	})

	doneState := &syncMgrPeer{
		Peer: donePeer,
		requestedTxns: map[chainhash.Hash]struct{}{
			txHash: {}, unknownTxHash: {},
		},
		requestedBlocks: map[chainhash.Hash]struct{}{
			blockHash: {}, unknownBlockHash: {},
		},
	}
	backupState := &syncMgrPeer{
		Peer:            backupPeer,
		requestedTxns:   make(map[chainhash.Hash]struct{}),
		requestedBlocks: make(map[chainhash.Hash]struct{}),
	}
	m := SyncManager{
		peers: map[*peerpkg.Peer]*syncMgrPeer{
			donePeer:   doneState,
			backupPeer: backupState,
		},
		requestedTxns: map[chainhash.Hash]struct{}{
			txHash: {}, unknownTxHash: {},
		},
		requestedBlocks: map[chainhash.Hash]struct{}{
			blockHash: {}, unknownBlockHash: {},
		},
	}

	m.handleDonePeerMsg(donePeer)

	if _, ok := m.peers[donePeer]; ok {
		t.Fatal("disconnected peer was retained")
	}
	if _, ok := backupState.requestedTxns[txHash]; !ok {
		t.Fatal("transaction was not assigned to backup peer")
	}
	if _, ok := backupState.requestedBlocks[blockHash]; !ok {
		t.Fatal("block was not assigned to backup peer")
	}
	if _, ok := m.requestedTxns[txHash]; !ok {
		t.Fatal("global transaction request was removed")
	}
	if _, ok := m.requestedBlocks[blockHash]; !ok {
		t.Fatal("global block request was removed")
	}
	if _, ok := m.requestedTxns[unknownTxHash]; ok {
		t.Fatal("unannounced transaction request was retained")
	}
	if _, ok := m.requestedBlocks[unknownBlockHash]; ok {
		t.Fatal("unannounced block request was retained")
	}
}
