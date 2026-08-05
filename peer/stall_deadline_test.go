package peer

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/wire"
)

func TestInventoryDeadlines(t *testing.T) {
	p := NewInboundPeer(&Config{})
	pendingResponses := make(map[string]time.Time)
	pendingData := make(map[wire.InvVect]time.Time)

	tx := wire.NewMsgTx()
	txHash := tx.TxHash()
	blockHash := chainhash.Hash{0xff}
	msg := wire.NewMsgGetData()
	if err := msg.AddInvVect(wire.NewInvVect(wire.InvTypeTx, &txHash)); err != nil {
		t.Fatal(err)
	}
	if err := msg.AddInvVect(wire.NewInvVect(wire.InvTypeBlock, &blockHash)); err != nil {
		t.Fatal(err)
	}
	before := time.Now()
	p.maybeAddDeadline(pendingResponses, pendingData, msg)
	after := time.Now()
	if len(pendingData) != 2 {
		t.Fatalf("got %d inventory deadlines, want 2", len(pendingData))
	}
	for _, deadline := range pendingData {
		if deadline.Before(before.Add(stallResponseTimeout*30)) ||
			deadline.After(after.Add(stallResponseTimeout*30)) {
			t.Fatalf("inventory deadline %v does not use the Equihash timeout", deadline)
		}
	}

	p.maybeRemoveDeadline(pendingResponses, pendingData, tx)
	if _, ok := pendingData[wire.InvVect{Type: wire.InvTypeTx, Hash: txHash}]; ok {
		t.Fatal("transaction response did not clear its deadline")
	}
	if len(pendingData) != 1 {
		t.Fatal("transaction response cleared another inventory deadline")
	}
}

// TestOptionalResponseDeadlines ensures requests which are allowed to produce
// no response cannot cause an honest peer to be disconnected as stalled.
func TestOptionalResponseDeadlines(t *testing.T) {
	p := NewInboundPeer(&Config{})
	optionalRequests := []wire.Message{
		wire.NewMsgMemPool(),
		wire.NewMsgGetBlocks(&chainhash.Hash{}),
		wire.NewMsgGetHeaders(),
		wire.NewMsgGetMiningState(),
	}
	for _, msg := range optionalRequests {
		pendingResponses := make(map[string]time.Time)
		pendingData := make(map[wire.InvVect]time.Time)
		p.maybeAddDeadline(pendingResponses, pendingData, msg)
		if len(pendingResponses) != 0 || len(pendingData) != 0 {
			t.Fatalf("optional %s request added a response deadline", msg.Command())
		}
	}

	// Init state has a guaranteed response and must retain its deadline.
	pendingResponses := make(map[string]time.Time)
	p.maybeAddDeadline(pendingResponses, make(map[wire.InvVect]time.Time),
		wire.NewMsgGetInitState())
	if _, ok := pendingResponses[wire.CmdInitState]; !ok {
		t.Fatal("getinitstate did not add its guaranteed response deadline")
	}
}

func TestInventoryDeadlineBurstLimit(t *testing.T) {
	const maxBurst = wire.MaxInvPerMsg + wire.MaxInvPerMsg/2
	p := NewInboundPeer(&Config{})
	pendingData := make(map[wire.InvVect]time.Time, maxBurst)
	for i := 0; i < maxBurst; i++ {
		var hash chainhash.Hash
		binary.LittleEndian.PutUint64(hash[:], uint64(i))
		pendingData[wire.InvVect{Type: wire.InvTypeBlock, Hash: hash}] = time.Now()
	}

	hash := chainhash.Hash{0xff}
	msg := wire.NewMsgGetData()
	if err := msg.AddInvVect(wire.NewInvVect(wire.InvTypeBlock, &hash)); err != nil {
		t.Fatal(err)
	}
	p.maybeAddDeadline(make(map[string]time.Time), pendingData, msg)
	p.WaitForDisconnect()
}
