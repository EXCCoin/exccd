// Copyright (c) 2026 The EXCC developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/wire"
)

func TestRebroadcastMembership(t *testing.T) {
	const limit = 2
	pending := make(pendingRebroadcast)
	live := make(map[chainhash.Hash]*dcrutil.Tx)
	fetch := func(hash *chainhash.Hash) (*dcrutil.Tx, error) {
		if tx := live[*hash]; tx != nil {
			return tx, nil
		}
		return nil, errors.New("left the pool")
	}
	snapshot := func() map[chainhash.Hash]struct{} {
		active := make(map[chainhash.Hash]struct{}, len(live))
		for hash := range live {
			active[hash] = struct{}{}
		}
		return active
	}
	for i := byte(0); i < 8; i++ {
		tx := wire.NewMsgTx()
		tx.LockTime = uint32(i)
		dtx := dcrutil.NewTx(tx)
		if i > 0 {
			clear(live)
		}
		live[*dtx.Hash()] = dtx
		pending.add(*wire.NewInvVect(wire.InvTypeTx, dtx.Hash()), limit, snapshot)
		if len(pending) > 2*limit {
			t.Fatalf("stale inventory grew to %d", len(pending))
		}
	}
	var visited int
	pending.visit(fetch, func(iv wire.InvVect, tx *dcrutil.Tx) {
		visited++
		if live[iv.Hash] != tx {
			t.Fatal("relayed a stale transaction")
		}
	})
	if visited != 1 || len(pending) != 1 {
		t.Fatalf("visited %d, retained %d", visited, len(pending))
	}
}
