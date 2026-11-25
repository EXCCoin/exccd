// Copyright (c) 2026 The EXCC developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/EXCCoin/exccd/blockchain/stake/v4"
	"github.com/EXCCoin/exccd/blockchain/v4/chaingen"
	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/gcs/v3/blockcf2"
	"github.com/EXCCoin/exccd/txscript/v4"
	"github.com/EXCCoin/exccd/wire"
)

func TestMainNetRetirementTxs(t *testing.T) {
	if mainNetRetirementHeight != 1_728_000 {
		t.Fatalf("mainNetRetirementHeight = %d, want 1728000", mainNetRetirementHeight)
	}

	params := chaincfg.MainNetParams()
	txs := MainNetRetirementTxs()
	if len(txs) != 2 {
		t.Fatalf("got %d mainnet retirement txs, want 2", len(txs))
	}

	// Mainnet retirement must be inactive before mainNetRetirementHeight and
	// active on mainnet at and after mainNetRetirementHeight (never on other nets).
	for i, tx := range txs {
		if IsMainNetRetirementTx(params, mainNetRetirementHeight-1, tx) {
			t.Fatalf("tx %d unexpectedly active on mainnet before activation height", i)
		}
		if !IsMainNetRetirementTx(params, mainNetRetirementHeight, tx) {
			t.Fatalf("tx %d not active on mainnet at activation height %d", i, mainNetRetirementHeight)
		}
		if !IsMainNetRetirementTx(params, mainNetRetirementHeight+1, tx) {
			t.Fatalf("tx %d not active on mainnet after activation height", i)
		}
		if IsMainNetRetirementTx(chaincfg.TestNet3Params(), mainNetRetirementHeight, tx) {
			t.Fatalf("tx %d unexpectedly active on testnet3", i)
		}
		mutated := dcrutil.NewTx(tx.MsgTx().Copy())
		mutated.MsgTx().TxOut[0].Value--
		if IsMainNetRetirementTx(params, mainNetRetirementHeight, mutated) {
			t.Fatalf("mutated tx %d unexpectedly matched canonical retirement tx", i)
		}
	}

	recoveryOutpoints := MainNetRecoveryOutpoints()
	wantTotals := []int64{99_998_662_531_005, 2_099_999_989_990_000}
	wantInputs := []int{3, 4}
	wantTags := []string{"EXCC-RETIRE-20261124-F1", "EXCC-RETIRE-20261124-F2"}

	var grandTotal int64
	seenOutpoints := make(map[wire.OutPoint]struct{})
	for i, tx := range txs {
		msg := tx.MsgTx()
		if msg.Version != 1 || msg.SerType != wire.TxSerializeFull {
			t.Fatalf("tx %d version=%d serType=%d, want 1/Full", i, msg.Version, msg.SerType)
		}
		if len(msg.TxIn) != wantInputs[i] {
			t.Fatalf("tx %d inputs=%d, want %d", i, len(msg.TxIn), wantInputs[i])
		}
		if len(msg.TxOut) != 1 {
			t.Fatalf("tx %d outputs=%d, want 1", i, len(msg.TxOut))
		}
		var sumIn int64
		for j, in := range msg.TxIn {
			if in.PreviousOutPoint.Tree != wire.TxTreeRegular {
				t.Fatalf("tx %d in %d tree=%d, want regular", i, j, in.PreviousOutPoint.Tree)
			}
			if len(in.SignatureScript) != 0 {
				t.Fatalf("tx %d in %d non-empty sigScript", i, j)
			}
			if _, ok := recoveryOutpoints[in.PreviousOutPoint]; !ok {
				t.Fatalf("tx %d in %d (%v) missing from MainNetRecoveryOutpoints", i, j, in.PreviousOutPoint)
			}
			if _, dup := seenOutpoints[in.PreviousOutPoint]; dup {
				t.Fatalf("duplicate outpoint %v", in.PreviousOutPoint)
			}
			seenOutpoints[in.PreviousOutPoint] = struct{}{}
			sumIn += in.ValueIn
		}
		out := msg.TxOut[0]
		if sumIn != wantTotals[i] || out.Value != wantTotals[i] {
			t.Fatalf("tx %d sumIn=%d out=%d, want %d", i, sumIn, out.Value, wantTotals[i])
		}
		if out.Value > dcrutil.MaxAmount {
			t.Fatalf("tx %d output %d exceeds MaxAmount %d", i, out.Value, int64(dcrutil.MaxAmount))
		}
		if out.Version != 0 || !txscript.IsUnspendable(out.Value, out.PkScript) {
			t.Fatalf("tx %d output not version-0 provably unspendable", i)
		}
		wantScript, err := txscript.NewScriptBuilder().
			AddOp(txscript.OP_RETURN).
			AddData([]byte(wantTags[i])).
			Script()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(out.PkScript, wantScript) {
			t.Fatalf("tx %d pkScript=%x, want %x", i, out.PkScript, wantScript)
		}

		var buf bytes.Buffer
		if err := msg.Serialize(&buf); err != nil {
			t.Fatalf("tx %d serialize: %v", i, err)
		}
		t.Logf("MainNet Burn Tx %d (%s): TxHash=%s TxHashFull=%s Size=%d Value=%d Hex=%s",
			i+1, wantTags[i], msg.TxHash(), msg.TxHashFull(), buf.Len(), out.Value, hex.EncodeToString(buf.Bytes()))

		grandTotal += out.Value
	}
	if len(seenOutpoints) != 7 {
		t.Fatalf("got %d distinct inputs, want 7", len(seenOutpoints))
	}
	if grandTotal != 2_199_998_652_521_005 {
		t.Fatalf("grandTotal=%d, want 2199998652521005", grandTotal)
	}

	// Ensure MainNetRetirementTxs returns defensive deep copies.
	txs[0].MsgTx().TxOut[0].Value = 1
	if MainNetRetirementTxs()[0].MsgTx().TxOut[0].Value != wantTotals[0] {
		t.Fatal("mutating MainNetRetirementTxs copy corrupted canonical plan")
	}
}

func TestMainNetQuarantineBypassOnlyForCanonicalRetirement(t *testing.T) {
	params := chaincfg.MainNetParams()
	burnTxs := MainNetRetirementTxs()
	burn1 := burnTxs[0]

	view := NewUtxoViewpoint(nil)
	for _, in := range burn1.MsgTx().TxIn {
		view.entries[in.PreviousOutPoint] = &UtxoEntry{
			amount:      in.ValueIn,
			blockHeight: in.BlockHeight,
			blockIndex:  in.BlockIndex,
		}
	}

	const testHeight = mainNetInputRecoveryHeight + 100

	// Without allowQuarantined, even the canonical burn tx is rejected by quarantine.
	if _, err := checkTransactionInputs(nil, burn1, testHeight, view, true,
		params, nil, false, false, false, txInputRules{}); !errors.Is(err, ErrBadTxInput) {
		t.Fatalf("expected quarantine rejection without bypass, got %v", err)
	}

	// With allowQuarantined=true, canonical burn tx passes with zero fee.
	fee, err := checkTransactionInputs(nil, burn1, testHeight, view, true,
		params, nil, false, false, false, txInputRules{allowQuarantined: true})
	if err != nil {
		t.Fatalf("canonical retirement tx rejected with allowQuarantined: %v", err)
	}
	if fee != 0 {
		t.Fatalf("canonical retirement tx fee=%d, want 0", fee)
	}
}

type stakeNodeAdapter struct{ chain *BlockChain }

func (a stakeNodeAdapter) GetPrevStakeNode(h *wire.BlockHeader) (*stake.Node, error) {
	a.chain.chainLock.Lock()
	defer a.chain.chainLock.Unlock()
	prev := a.chain.index.LookupNode(&h.PrevBlock)
	return a.chain.fetchStakeNode(prev)
}

func TestRetirementConsensusLifecycle(t *testing.T) {
	// Use RegNet difficulty/maturity parameters with a private network magic so
	// header commitments stay inactive (matching MainNet and chaingen headers)
	// without triggering TestNet3-specific rules.
	params := cloneParams(chaincfg.RegNetParams())
	params.Net = wire.CurrencyNet(0x52455449)
	chain, teardownFunc, err := chainSetup("retirementlifecycletest", params)
	if err != nil {
		t.Fatalf("chainSetup: %v", err)
	}
	defer teardownFunc()
	gen, err := chaingen.MakeGenerator(params, stakeNodeAdapter{chain: chain})
	if err != nil {
		t.Fatalf("MakeGenerator: %v", err)
	}
	g := &chaingenHarness{
		Generator:          &gen,
		t:                  t,
		chain:              chain,
		deploymentVersions: make(map[string]uint32),
	}

	// Reach StakeValidationHeight so PoS voting and regular-tree disapproval
	// work identically to mainnet.
	g.AdvanceToStakeValidationHeight()

	// Construct an unspendable-without-key P2PKH script representing an
	// exchange/attacker address whose private key the protocol does not hold.
	lockedPubKeyHash := bytes.Repeat([]byte{0x42}, 20)
	lockedPkScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(lockedPubKeyHash).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	if err != nil {
		t.Fatal(err)
	}

	// Create a setup block that splits a mature coinbase output into:
	//   vout 0, 1: Family 1 target outputs (locked P2PKH, no key available)
	//   vout 2:    Family 2 target output  (locked P2PKH, no key available)
	//   vout 3:    Innocent output (OP_TRUE, spendable normally)
	outs := g.OldestCoinbaseOuts()
	seedOut := outs[0]
	const (
		amtA1 = int64(100_000_000)
		amtA2 = int64(200_000_000)
		amtB1 = int64(300_000_000)
		fee   = int64(10_000)
	)
	amtInnocent := int64(seedOut.Amount()) - amtA1 - amtA2 - amtB1 - fee
	if amtInnocent <= 0 {
		t.Fatalf("insufficient coinbase amount %d", seedOut.Amount())
	}

	setupTx := wire.NewMsgTx()
	setupTx.Version = 1
	seedPrevOut := seedOut.PrevOut()
	setupTx.AddTxIn(wire.NewTxIn(&seedPrevOut, int64(seedOut.Amount()), []byte{txscript.OP_DATA_1, txscript.OP_TRUE}))
	setupTx.AddTxOut(wire.NewTxOut(amtA1, lockedPkScript))
	setupTx.AddTxOut(wire.NewTxOut(amtA2, lockedPkScript))
	setupTx.AddTxOut(wire.NewTxOut(amtB1, lockedPkScript))
	setupTx.AddTxOut(wire.NewTxOut(amtInnocent, []byte{txscript.OP_TRUE}))

	setupBlock := g.NextBlock("bsetup", nil, outs[1:], func(b *wire.MsgBlock) {
		b.AddTransaction(setupTx)
		b.Transactions[0].TxOut[2].Value += fee
	})
	g.SaveTipCoinbaseOuts()
	g.AcceptTipBlock()

	setupTxHash := setupTx.TxHash()
	setupHeight := setupBlock.Header.Height
	const setupBlockIdx = uint32(1)

	opA1 := wire.OutPoint{Hash: setupTxHash, Index: 0, Tree: wire.TxTreeRegular}
	opA2 := wire.OutPoint{Hash: setupTxHash, Index: 1, Tree: wire.TxTreeRegular}
	opB1 := wire.OutPoint{Hash: setupTxHash, Index: 2, Tree: wire.TxTreeRegular}
	opInnocent := wire.OutPoint{Hash: setupTxHash, Index: 3, Tree: wire.TxTreeRegular}

	// Build two canonical zero-fee OP_RETURN burn transactions with empty sigScripts.
	makeTestBurn := func(tag string, inputs []struct {
		op  wire.OutPoint
		amt int64
	}) *dcrutil.Tx {
		msg := wire.NewMsgTx()
		msg.SerType = wire.TxSerializeFull
		msg.Version = 1
		var total int64
		for _, in := range inputs {
			txIn := wire.NewTxIn(&in.op, in.amt, nil)
			txIn.BlockHeight = setupHeight
			txIn.BlockIndex = setupBlockIdx
			msg.AddTxIn(txIn)
			total += in.amt
		}
		script, err := txscript.NewScriptBuilder().
			AddOp(txscript.OP_RETURN).
			AddData([]byte(tag)).
			Script()
		if err != nil {
			t.Fatal(err)
		}
		msg.AddTxOut(wire.NewTxOut(total, script))
		tx := dcrutil.NewTx(msg)
		tx.SetTree(wire.TxTreeRegular)
		return tx
	}

	burn1 := makeTestBurn("TEST-RETIRE-F1", []struct {
		op  wire.OutPoint
		amt int64
	}{
		{opA1, amtA1},
		{opA2, amtA2},
	})
	burn2 := makeTestBurn("TEST-RETIRE-F2", []struct {
		op  wire.OutPoint
		amt int64
	}{
		{opB1, amtB1},
	})

	// Schedule retirement 2 blocks after bsetup.
	retireHeight := int64(setupHeight) + 2
	plan, err := NewRetirementPlan(retireHeight, []*dcrutil.Tx{burn1, burn2})
	if err != nil {
		t.Fatalf("NewRetirementPlan: %v", err)
	}
	g.chain.SetRetirementPlan(plan)

	// -------------------------------------------------------------------------
	// 1. Pre-activation block (retireHeight - 1): retirement not required.
	// -------------------------------------------------------------------------
	reqPre, err := g.chain.RequiredRetirementTxs(retireHeight-1, true)
	if err != nil || len(reqPre) != 0 {
		t.Fatalf("pre-activation RequiredRetirementTxs = (%d, %v), want (0, nil)", len(reqPre), err)
	}
	outs = g.OldestCoinbaseOuts()
	g.NextBlock("bpre", &outs[0], outs[1:])
	g.SaveTipCoinbaseOuts()
	g.AcceptTipBlock()

	for _, op := range []wire.OutPoint{opA1, opA2, opB1, opInnocent} {
		entry, err := g.chain.FetchUtxoEntry(op)
		if err != nil || entry == nil || entry.IsSpent() {
			t.Fatalf("pre-activation outpoint %v unexpectedly missing/spent (entry=%v err=%v)", op, entry, err)
		}
	}

	// Verify RequiredRetirementTxs now demands both burn txs for retireHeight.
	reqAt, err := g.chain.RequiredRetirementTxs(retireHeight, true)
	if err != nil || len(reqAt) != 2 {
		t.Fatalf("at-activation RequiredRetirementTxs = (%d, %v), want (2, nil)", len(reqAt), err)
	}
	if reqAt[0].MsgTx().TxHashFull() != burn1.MsgTx().TxHashFull() ||
		reqAt[1].MsgTx().TxHashFull() != burn2.MsgTx().TxHashFull() {
		t.Fatal("RequiredRetirementTxs returned wrong canonical burn txs")
	}

	// Helper to prepend canonical txs right after coinbase.
	prependRegular := func(txs ...*wire.MsgTx) func(*wire.MsgBlock) {
		return func(b *wire.MsgBlock) {
			out := make([]*wire.MsgTx, 0, len(b.Transactions)+len(txs))
			out = append(out, b.Transactions[0])
			out = append(out, txs...)
			out = append(out, b.Transactions[1:]...)
			b.Transactions = out
		}
	}

	// -------------------------------------------------------------------------
	// 2. Negative consensus checks at retireHeight.
	// -------------------------------------------------------------------------
	preTipName := g.TipName()
	outs = g.OldestCoinbaseOuts()

	// 2a. Omitting burn transactions.
	g.NextBlock("bretire_missing", &outs[0], outs[1:])
	g.RejectTipBlock(ErrBadTxInput)

	// 2b. Including only burn1 (omitting burn2).
	g.SetTip(preTipName)
	g.NextBlock("bretire_partial", &outs[0], outs[1:], prependRegular(burn1.MsgTx().Copy()))
	g.RejectTipBlock(ErrBadTxInput)

	// 2c. Reversed order (burn2 before burn1).
	g.SetTip(preTipName)
	g.NextBlock("bretire_reversed", &outs[0], outs[1:], prependRegular(burn2.MsgTx().Copy(), burn1.MsgTx().Copy()))
	g.RejectTipBlock(ErrBadTxInput)

	// 2d. Miner skims 1 atom fee from burn1 OP_RETURN value.
	g.SetTip(preTipName)
	skimBurn1 := burn1.MsgTx().Copy()
	skimBurn1.TxOut[0].Value--
	g.NextBlock("bretire_skim", &outs[0], outs[1:], func(b *wire.MsgBlock) {
		prependRegular(skimBurn1, burn2.MsgTx().Copy())(b)
		b.Transactions[0].TxOut[2].Value++
	})
	g.RejectTipBlock(ErrBadTxInput)

	// 2e. Non-empty SignatureScript on burn1.
	g.SetTip(preTipName)
	sigBurn1 := burn1.MsgTx().Copy()
	sigBurn1.TxIn[0].SignatureScript = []byte{txscript.OP_TRUE}
	g.NextBlock("bretire_sig", &outs[0], outs[1:], prependRegular(sigBurn1, burn2.MsgTx().Copy()))
	g.RejectTipBlock(ErrBadTxInput)

	// 2f. Altered OP_RETURN tag on burn2.
	g.SetTip(preTipName)
	badTagBurn2 := burn2.MsgTx().Copy()
	badTagBurn2.TxOut[0].PkScript, _ = txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).AddData([]byte("TAMPERED")).Script()
	g.NextBlock("bretire_tag", &outs[0], outs[1:], prependRegular(burn1.MsgTx().Copy(), badTagBurn2))
	g.RejectTipBlock(ErrBadTxInput)

	// 2g. Burn txs placed after another regular tx instead of [1..2].
	g.SetTip(preTipName)
	g.NextBlock("bretire_wrongpos", &outs[0], outs[1:], func(b *wire.MsgBlock) {
		b.AddTransaction(burn1.MsgTx().Copy())
		b.AddTransaction(burn2.MsgTx().Copy())
	})
	g.RejectTipBlock(ErrBadTxInput)

	// 2h. Duplicate canonical burn1 appended after [burn1, burn2].
	g.SetTip(preTipName)
	g.NextBlock("bretire_dup", &outs[0], outs[1:], prependRegular(
		burn1.MsgTx().Copy(), burn2.MsgTx().Copy(), burn1.MsgTx().Copy()))
	g.RejectTipBlock(ErrDuplicateTx)

	// -------------------------------------------------------------------------
	// 3. Valid retirement block at retireHeight (with innocent co-located spend).
	// -------------------------------------------------------------------------
	innocentSpend := wire.NewMsgTx()
	innocentSpend.Version = 1
	innocentIn := wire.NewTxIn(&opInnocent, amtInnocent, nil)
	innocentIn.BlockHeight = setupHeight
	innocentIn.BlockIndex = setupBlockIdx
	innocentSpend.AddTxIn(innocentIn)
	innocentSpend.AddTxOut(wire.NewTxOut(amtInnocent-fee, []byte{txscript.OP_TRUE}))

	g.SetTip(preTipName)
	retireBlock := g.NextBlock("bretire", &outs[0], outs[1:], func(b *wire.MsgBlock) {
		prependRegular(burn1.MsgTx().Copy(), burn2.MsgTx().Copy())(b)
		b.AddTransaction(innocentSpend)
		b.Transactions[0].TxOut[2].Value += fee
	})
	g.SaveTipCoinbaseOuts()
	g.AcceptTipBlock()

	// All 3 retired outpoints must now be removed from the UTXO set.
	for _, op := range []wire.OutPoint{opA1, opA2, opB1} {
		entry, err := g.chain.FetchUtxoEntry(op)
		if err != nil {
			t.Fatalf("FetchUtxoEntry(%v): %v", op, err)
		}
		if entry != nil && !entry.IsSpent() {
			t.Fatalf("retired outpoint %v still unspent after retirement block", op)
		}
	}
	// Both OP_RETURN burn outputs must be provably pruned (absent from UTXO set).
	for i, btx := range []*dcrutil.Tx{burn1, burn2} {
		op := wire.OutPoint{Hash: *btx.Hash(), Index: 0, Tree: wire.TxTreeRegular}
		entry, err := g.chain.FetchUtxoEntry(op)
		if err != nil || entry != nil {
			t.Fatalf("burn %d OP_RETURN output present in UTXO set (entry=%v err=%v)", i+1, entry, err)
		}
	}
	// Innocent output spent cleanly and its new output exists in the UTXO set.
	innocentNewOp := wire.OutPoint{Hash: innocentSpend.TxHash(), Index: 0, Tree: wire.TxTreeRegular}
	entry, err := g.chain.FetchUtxoEntry(innocentNewOp)
	if err != nil || entry == nil || entry.IsSpent() || entry.Amount() != amtInnocent-fee {
		t.Fatalf("innocent spend output missing/wrong: entry=%v err=%v", entry, err)
	}

	// Verify the block's v2 GCS compact filter indexes the burned P2PKH script
	// so SPV wallets automatically match and download the burn block.
	retireBlockHash := retireBlock.BlockHash()
	filter, err := g.chain.FilterByBlockHash(&retireBlockHash)
	if err != nil {
		t.Fatalf("FilterByBlockHash: %v", err)
	}
	if !filter.Match(blockcf2.Key(&retireBlock.Header.MerkleRoot), lockedPkScript) {
		t.Fatal("GCS v2 filter for retirement block did not match retired P2PKH script")
	}

	// If next block approves bretire, 0 retirement txs are required.
	// If next block disapproves bretire, both burn txs are required again!
	reqAfterApprove, err := g.chain.RequiredRetirementTxs(retireHeight+1, true)
	if err != nil || len(reqAfterApprove) != 0 {
		t.Fatalf("RequiredRetirementTxs(approve) = (%d, %v), want (0, nil)", len(reqAfterApprove), err)
	}
	reqAfterDisapprove, err := g.chain.RequiredRetirementTxs(retireHeight+1, false)
	if err != nil || len(reqAfterDisapprove) != 2 {
		t.Fatalf("RequiredRetirementTxs(disapprove) = (%d, %v), want (2, nil)", len(reqAfterDisapprove), err)
	}

	// -------------------------------------------------------------------------
	// 4. Disapproval at retireHeight + 1 (VoteBits = 0).
	// -------------------------------------------------------------------------
	retireTipName := g.TipName()
	outs = g.OldestCoinbaseOuts()
	disapprovePrev := func(b *wire.MsgBlock) {
		g.ReplaceVoteBits(0)(b)
		b.Header.VoteBits = 0
	}

	// 4a. Disapproving bretire WITHOUT re-including burn txs MUST be rejected!
	g.NextBlock("bdisapprove_missing", &outs[0], outs[1:], disapprovePrev)
	g.AssertTipDisapprovesPrevious()
	g.RejectTipBlock(ErrBadTxInput)

	// 4b. Disapproving bretire WITH [burn1, burn2] MUST be accepted!
	g.SetTip(retireTipName)
	g.NextBlock("bdisapprove_ok", &outs[0], outs[1:], disapprovePrev,
		prependRegular(burn1.MsgTx().Copy(), burn2.MsgTx().Copy()))
	g.AssertTipDisapprovesPrevious()
	g.SaveTipCoinbaseOuts()
	g.AcceptTipBlock()

	for _, op := range []wire.OutPoint{opA1, opA2, opB1} {
		entry, err := g.chain.FetchUtxoEntry(op)
		if err != nil || (entry != nil && !entry.IsSpent()) {
			t.Fatalf("outpoint %v unspent after disapproving retirement block", op)
		}
	}

	// -------------------------------------------------------------------------
	// 5. Post-retirement block (retireHeight + 2): burn txs no longer allowed.
	// -------------------------------------------------------------------------
	disapproveTipName := g.TipName()
	outs = g.OldestCoinbaseOuts()

	// Including burn1 again after retirement completed MUST be rejected.
	g.NextBlock("bpost_extra_burn", &outs[0], outs[1:], prependRegular(burn1.MsgTx().Copy()))
	g.RejectTipBlock(ErrBadTxInput)

	// Normal approving block succeeds.
	g.SetTip(disapproveTipName)
	postBlock := g.NextBlock("bpost", &outs[0], outs[1:])
	g.SaveTipCoinbaseOuts()
	g.AcceptTipBlock()
	postBlockHash := postBlock.BlockHash()

	wantStats, err := g.chain.FetchUtxoStats()
	if err != nil {
		t.Fatalf("FetchUtxoStats: %v", err)
	}

	// -------------------------------------------------------------------------
	// 6. InvalidateBlock / ReconsiderBlock rollback & re-apply across H_retire.
	// -------------------------------------------------------------------------
	if err := g.chain.InvalidateBlock(&retireBlockHash); err != nil {
		t.Fatalf("InvalidateBlock(bretire): %v", err)
	}
	if got := g.chain.BestSnapshot().Height; got != retireHeight-1 {
		t.Fatalf("height after InvalidateBlock = %d, want %d", got, retireHeight-1)
	}
	// All 3 retired outpoints must be restored as unspent!
	for _, op := range []wire.OutPoint{opA1, opA2, opB1} {
		entry, err := g.chain.FetchUtxoEntry(op)
		if err != nil || entry == nil || entry.IsSpent() {
			t.Fatalf("outpoint %v not restored after InvalidateBlock", op)
		}
	}

	if err := g.chain.ReconsiderBlock(&retireBlockHash); err != nil {
		t.Fatalf("ReconsiderBlock(bretire): %v", err)
	}
	if got := g.chain.BestSnapshot().Hash; got != postBlockHash {
		t.Fatalf("tip after ReconsiderBlock = %s, want %s", got, postBlockHash)
	}
	afterReconsiderStats, err := g.chain.FetchUtxoStats()
	if err != nil {
		t.Fatalf("FetchUtxoStats after reconsider: %v", err)
	}
	if afterReconsiderStats.SerializedHash != wantStats.SerializedHash ||
		afterReconsiderStats.Total != wantStats.Total {
		t.Fatalf("UTXO stats mismatch after reconsider: got (%v, %d), want (%v, %d)",
			afterReconsiderStats.SerializedHash, afterReconsiderStats.Total,
			wantStats.SerializedHash, wantStats.Total)
	}

	// -------------------------------------------------------------------------
	// 7. Fresh-sync convergence: replay all mainchain blocks into a fresh DB.
	// -------------------------------------------------------------------------
	freshChain, teardownFresh, err := chainSetup("retirementfreshtest", params)
	if err != nil {
		t.Fatalf("chainSetup(fresh): %v", err)
	}
	defer teardownFresh()
	freshChain.SetRetirementPlan(plan)

	tipHeight := g.chain.BestSnapshot().Height
	for h := int64(1); h <= tipHeight; h++ {
		blk, err := g.chain.BlockByHeight(h)
		if err != nil {
			t.Fatalf("BlockByHeight(%d): %v", h, err)
		}
		if _, err := freshChain.ProcessBlock(blk); err != nil {
			t.Fatalf("fresh ProcessBlock(%d): %v", h, err)
		}
	}

	freshStats, err := freshChain.FetchUtxoStats()
	if err != nil {
		t.Fatalf("fresh FetchUtxoStats: %v", err)
	}
	if freshStats.SerializedHash != wantStats.SerializedHash ||
		freshStats.Total != wantStats.Total ||
		freshStats.Utxos != wantStats.Utxos {
		t.Fatalf("fresh sync UTXO stats mismatch: got (hash=%v val=%d utxos=%d), want (hash=%v val=%d utxos=%d)",
			freshStats.SerializedHash, freshStats.Total, freshStats.Utxos,
			wantStats.SerializedHash, wantStats.Total, wantStats.Utxos)
	}
}
