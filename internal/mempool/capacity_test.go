// Copyright (c) 2026 The EXCC developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"errors"
	"maps"
	"testing"

	"github.com/EXCCoin/exccd/blockchain/stake/v4"
	"github.com/EXCCoin/exccd/blockchain/v4"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/internal/mining"
	"github.com/EXCCoin/exccd/wire"
)

func assertCapacity(t *testing.T, p *TxPool) {
	t.Helper()
	c := &p.capacity
	if len(c.entries) != len(p.pool)+len(p.staged) {
		t.Fatalf("tracked count %d, main %d, staged %d", len(c.entries), len(p.pool), len(p.staged))
	}
	var size int64
	refs := make(map[parentKey]int)
	for _, pool := range []map[chainhash.Hash]*TxDesc{p.pool, p.staged} {
		for hash, desc := range pool {
			if e := c.entries[hash]; e == nil || e.desc != desc {
				t.Fatalf("missing capacity entry %v", hash)
			}
			size += int64(desc.Tx.MsgTx().SerializeSize())
			for _, in := range desc.Tx.MsgTx().TxIn {
				op := &in.PreviousOutPoint
				refs[parentKey{op.Hash, op.Tree}]++
			}
		}
	}
	if size != c.bytes || size > c.maxBytes || len(c.entries) > c.maxTxs || !maps.Equal(refs, c.references) {
		t.Fatalf("inconsistent capacity accounting: expected %d bytes, tracked %d", size, c.bytes)
	}
	for hash, e := range c.entries {
		isLeaf := refs[parentKey{hash, e.desc.Tx.Tree()}] == 0
		if isLeaf != (e.index >= 0) || isLeaf && c.leaves[e.index] != e {
			t.Fatalf("incorrect leaf index for %v", hash)
		}
	}
	for i, entry := range c.leaves {
		if entry.index != i || i > 0 && c.leaves.Less(i, (i-1)/2) {
			t.Fatalf("broken eviction heap at %d", i)
		}
	}
}

func TestFeeAndCapacityAdmission(t *testing.T) {
	t.Parallel()
	h, inputs, err := newPoolHarness(chaincfg.MainNetParams())
	if err != nil {
		t.Fatal(err)
	}
	p := h.txPool
	funding, err := h.CreateSignedTx(inputs, 6)
	if err != nil {
		t.Fatal(err)
	}
	h.AddFakeUTXO(funding, h.chain.BestHeight()-1, 0)
	spend := func(parent *dcrutil.Tx, index uint32, outs uint32, fee int64) *dcrutil.Tx {
		t.Helper()
		input := txOutToSpendableOut(parent, index, wire.TxTreeRegular)
		tx, err := h.CreateSignedTx([]spendableOutput{input}, outs, func(tx *wire.MsgTx) {
			desiredSum := int64(input.amount) - fee
			for _, output := range tx.TxOut[1:] {
				desiredSum -= output.Value
			}
			tx.TxOut[0].Value = desiredSum
		})
		if err != nil {
			t.Fatal(err)
		}
		return tx
	}
	process := func(tx *dcrutil.Tx, want error) {
		t.Helper()
		_, err := p.ProcessTransaction(tx, false, false, true, 0)
		if !errors.Is(err, want) {
			t.Fatalf("process %s: got %v, want %v", tx.Hash(), err, want)
		}
		assertCapacity(t, p)
	}
	for _, fee := range []int64{0, 100} {
		tx := spend(funding, 0, 1, fee)
		for _, oldRateLimit := range []bool{true, false} {
			_, err := p.ProcessTransaction(tx, false, oldRateLimit, true, 0)
			if !errors.Is(err, ErrInsufficientFee) {
				t.Fatalf("peer/RPC minimum fee: %v", err)
			}
			_, err = p.MaybeAcceptTransaction(tx, false, oldRateLimit)
			if !errors.Is(err, ErrInsufficientFee) {
				t.Fatalf("disconnected block minimum fee: %v", err)
			}
		}
	}

	parent := spend(funding, 0, 1, 1000)
	child := spend(parent, 0, 1, 3000)
	middle := spend(funding, 1, 1, 2000)
	high := spend(funding, 2, 1, 4000)
	p.capacity.maxTxs = 2
	process(parent, nil)
	process(child, nil)
	process(middle, ErrMempoolFull)
	process(spend(child, 0, 1, 10000), ErrMempoolFull) // Both ancestors pinned.
	process(high, nil)
	if p.HaveTransaction(child.Hash()) || !p.HaveTransaction(parent.Hash()) || !p.HaveTransaction(high.Hash()) {
		t.Fatal("lowest eligible leaf was not displaced")
	}

	p.RemoveTransaction(parent, true, false, false)
	p.RemoveTransaction(high, true, false, false)
	assertCapacity(t, p)
	p.capacity.maxTxs = 10
	p.capacity.maxBytes = int64(parent.MsgTx().SerializeSize() + child.MsgTx().SerializeSize())
	process(parent, nil)
	process(child, nil)
	// A larger candidate cannot evict the high-fee child. Reverse the
	// dependency by mining the parent, then replace it with an independent
	// low-fee transaction to exercise an unsuccessful multi-victim plan.
	p.RemoveTransaction(parent, false, false, false)
	assertCapacity(t, p)
	low := spend(funding, 3, 1, 900)
	p.capacity.maxBytes = int64(low.MsgTx().SerializeSize() + child.MsgTx().SerializeSize())
	process(low, nil)
	partial := spend(funding, 4, 3, 2000)
	process(partial, ErrMempoolFull)
	if !p.HaveTransaction(low.Hash()) || !p.HaveTransaction(child.Hash()) {
		t.Fatal("partial eviction applied to a rejected transaction")
	}
	large := spend(funding, 4, 3, 10000)
	process(large, nil)
	if p.Count() != 1 || !p.HaveTransaction(large.Hash()) {
		t.Fatal("multi-victim byte admission did not replace both leaves")
	}
	process(spend(funding, 5, 12, 20000), ErrMempoolFull)

	p.RemoveTransaction(large, true, false, false)
	p.capacity.maxBytes = DefaultMaxMempoolBytes
	p.capacity.maxTxs = 2
	process(parent, nil)
	ticket, err := h.CreateTicketPurchaseFromTx(parent, 40000)
	if err != nil {
		t.Fatal(err)
	}
	process(ticket, nil)
	if !p.isTransactionStaged(ticket.Hash()) {
		t.Fatal("ticket should be staged")
	}
	process(high, nil)
	if p.HaveTransaction(ticket.Hash()) || !p.HaveTransaction(parent.Hash()) {
		t.Fatal("staged ticket eviction did not preserve parent")
	}

	// A higher-fee orphan on the second output can displace its sibling
	// accepted from the first. Only announce candidates still in the pool.
	p.RemoveTransaction(parent, true, false, false)
	p.RemoveTransaction(high, true, false, false)
	root := spend(funding, 5, 2, 2000)
	sibling := spend(root, 0, 1, 1000)
	kept := spend(root, 1, 1, 3000)
	rejected := spend(root, 1, 1, 0)
	for _, tx := range []*dcrutil.Tx{sibling, rejected, kept} {
		if accepted, err := p.ProcessTransaction(tx, true, true, true, 0); err != nil || len(accepted) != 0 {
			t.Fatalf("store orphan: %v", err)
		}
	}
	accepted, err := p.ProcessTransaction(root, false, true, true, 0)
	if err != nil || len(accepted) != 2 || *accepted[0].Hash() != *root.Hash() || *accepted[1].Hash() != *kept.Hash() {
		t.Fatalf("incorrect announced transactions: %v, %v", accepted, err)
	}
	assertCapacity(t, p)
}

func TestRetentionOrdering(t *testing.T) {
	entry := func(kind stake.TxType, fee, size int64) *capacityEntry {
		return &capacityEntry{desc: &TxDesc{TxDesc: mining.TxDesc{
			Type: kind, Fee: fee, TxSize: size}}}
	}
	// Cross products here exceed the width of uint64.
	base := entry(stake.TxTypeRegular, 1000000000000001, 1000000)
	equal := entry(stake.TxTypeSStx, 2000000000000002, 2000000)
	higher := entry(stake.TxTypeRegular, 2000000000000003, 2000000)
	revocation := entry(stake.TxTypeSSRtx, 0, 100)
	oldVote := entry(stake.TxTypeSSGen, 0, 200)
	if compareRetention(base, equal) != 0 || compareRetention(higher, equal) <= 0 ||
		compareRetention(revocation, higher) <= 0 || compareRetention(oldVote, revocation) <= 0 {
		t.Fatal("incorrect retention ordering")
	}
}

// A vote is protected only after the current tip and lottery ticket agree.
func TestCapacityWinningVotes(t *testing.T) {
	h, outputs, err := newPoolHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := h.CreateTicketPurchase(outputs[0], 40000)
	if err != nil {
		t.Fatal(err)
	}
	h.chain.SetHeight(h.chainParams.StakeValidationHeight)
	vote, err := h.CreateVote(ticket)
	if err != nil {
		t.Fatal(err)
	}
	desc := &TxDesc{TxDesc: mining.TxDesc{Tx: vote, Type: stake.TxTypeSSGen, TxSize: int64(vote.MsgTx().SerializeSize())}}
	best := &blockchain.BestState{Hash: *h.chain.BestHash(), Height: h.chain.BestHeight(), NextWinningTickets: []chainhash.Hash{*ticket.Hash()}}
	if !isWinningVote(desc, best) {
		t.Fatal("locally selected vote did not get its reserve")
	}
	p := newPoolCapacity(Policy{MaxMempoolTxs: 1})
	p.insert(&capacityEntry{desc: desc, index: -1})
	other, err := h.CreateTx(outputs[0])
	if err != nil {
		t.Fatal(err)
	}
	ordinary := &TxDesc{TxDesc: mining.TxDesc{Tx: other, Type: stake.TxTypeRegular, Fee: 10000, TxSize: int64(other.MsgTx().SerializeSize())}}
	snapshot := func() *blockchain.BestState { return best }
	if _, err := p.makeRoom(ordinary, snapshot); !errors.Is(err, ErrMempoolFull) || len(p.leaves) != 1 {
		t.Fatalf("lost reserved vote: %v", err)
	}
	p.remove(vote.Hash())
	p.insert(&capacityEntry{desc: ordinary, index: -1})
	if victims, err := p.makeRoom(desc, snapshot); err != nil || len(victims) != 1 || victims[0] != ordinary {
		t.Fatalf("winning vote could not displace fee traffic: %v", err)
	}
	best.NextWinningTickets = nil
	if isWinningVote(desc, best) {
		t.Fatal("unselected ticket received reserve")
	}
	best.NextWinningTickets = []chainhash.Hash{*ticket.Hash()}
	best.Hash[0] ^= 1
	if isWinningVote(desc, best) {
		t.Fatal("wrong tip received reserve")
	}
}

func TestCapacityVoteReacceptance(t *testing.T) {
	h, outputs, err := newPoolHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := h.CreateTicketPurchase(outputs[0], 40000)
	if err != nil {
		t.Fatal(err)
	}
	h.chain.utxos.AddTxOuts(ticket, h.chainParams.StakeEnabledHeight+1, 0, noTreasury, noAutoRevocations)
	h.chain.SetHeight(h.chainParams.StakeValidationHeight)
	p := h.txPool
	p.capacity.maxTxs = 1
	p.cfg.BestSnapshot = func() *blockchain.BestState {
		return &blockchain.BestState{Hash: *h.chain.BestHash(), Height: h.chain.BestHeight(),
			NextWinningTickets: []chainhash.Hash{*ticket.Hash()}}
	}
	setTip := func(tag byte) chainhash.Hash {
		hash := chainhash.Hash{tag}
		h.chain.blocks[hash] = dcrutil.NewBlock(&wire.MsgBlock{})
		h.chain.SetBestHash(&hash)
		return hash
	}
	submit := func() *dcrutil.Tx {
		t.Helper()
		vote, err := h.CreateVote(ticket)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ProcessTransaction(vote, false, true, true, 0); err != nil {
			t.Fatal(err)
		}
		assertCapacity(t, p)
		return vote
	}
	first := setTip(1)
	firstVote := submit()
	oldView := p.VotesForBlocks([]chainhash.Hash{first})[0]
	second := setTip(2)
	secondVote := submit()
	if p.HaveTransaction(firstVote.Hash()) || len(p.VoteHashesForBlock(&first)) != 0 ||
		len(oldView) != 1 || oldView[0].VoteHash != *firstVote.Hash() {
		t.Fatal("evicted side vote remained active or its reader snapshot changed")
	}
	setTip(1)
	submit()
	active := p.VoteHashesForBlock(&first)
	if len(active) != 1 || active[0] != *firstVote.Hash() || len(p.VoteHashesForBlock(&second)) != 0 || p.HaveTransaction(secondVote.Hash()) {
		t.Fatal("reaccepted vote not reflected in the mining view")
	}
	p.RemoveTransaction(firstVote, false, false, false)
	if len(p.VoteHashesForBlock(&first)) != 1 {
		t.Fatal("normal removal lost historical vote metadata")
	}

	// Keep all legitimate side-tip candidates indexed when the last
	// submitted vote for a ticket leaves, then remove every conflict on
	// confirmation. Stakebases never compete for a previous output.
	p.capacity.maxTxs = 3
	submit()
	setTip(2)
	submit()
	setTip(3)
	thirdVote := submit()
	op := thirdVote.MsgTx().TxIn[1].PreviousOutPoint
	if len(p.alternateSpenders[op]) != 2 || p.outpoints[thirdVote.MsgTx().TxIn[0].PreviousOutPoint] != nil {
		t.Fatal("incorrect vote spender indexes")
	}
	p.RemoveTransaction(thirdVote, false, false, false)
	if len(p.alternateSpenders[op]) != 1 || p.outpoints[op] == nil {
		t.Fatal("removing vote lost the other spenders")
	}
	p.RemoveDoubleSpends(thirdVote, false, false)
	if p.Count() != 0 || len(p.alternateSpenders) != 0 || len(p.outpoints) != 0 {
		t.Fatal("confirmed ticket spend left conflicting votes")
	}
	assertCapacity(t, p)
}

func TestReinsertMissingAncestor(t *testing.T) {
	h, outputs, err := newPoolHarness(chaincfg.MainNetParams())
	if err != nil {
		t.Fatal(err)
	}
	// Mine a parent and child in the synthetic UTXO view, then receive
	// the grandchild into the pool.
	parent, err := h.CreateSignedTx(outputs, 1, func(tx *wire.MsgTx) {
		tx.TxOut[0].Value = int64(outputs[0].amount)
	})
	if err != nil {
		t.Fatal(err)
	}
	family, err := h.CreateTxChain(txOutToSpendableOut(parent, 0, wire.TxTreeRegular), 2)
	if err != nil {
		t.Fatal(err)
	}
	child, grandchild := family[0], family[1]
	h.AddFakeUTXO(child, h.chain.BestHeight()-1, 0)
	p := h.txPool
	if _, err := p.ProcessTransaction(grandchild, false, true, true, 0); err != nil {
		t.Fatal(err)
	}
	h.chain.utxos.RemoveEntry(grandchild.MsgTx().TxIn[0].PreviousOutPoint)
	p.ReinsertTransaction(parent, 0) // Below the local relay fee.
	p.ReinsertTransaction(child, 0)  // Its parent is unavailable.
	if p.HaveTransaction(grandchild.Hash()) {
		t.Fatal("retained grandchild with an unavailable input")
	}
	assertCapacity(t, p)
}
