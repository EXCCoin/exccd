// Copyright (c) 2026 The EXCC developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/txscript/v4"
	"github.com/EXCCoin/exccd/wire"
)

// mainNetRetirementHeight is the mainnet height at which the two canonical
// Option B OP_RETURN burn transactions become mandatory.
const mainNetRetirementHeight int64 = 1_728_000

// RetirementPlan defines a deterministic set of canonical regular-tree burn
// transactions that retire specified unspent outputs starting at
// ActivationHeight.
type RetirementPlan struct {
	ActivationHeight int64
	Txs              []*dcrutil.Tx
	txFullHashes     map[chainhash.Hash]int
	outpoints        []wire.OutPoint
}

type retirementInputSpec struct {
	txid        string
	vout        uint32
	atoms       int64
	blockHeight uint32
	blockIndex  uint32
}

type retirementTxSpec struct {
	tag    string
	inputs []retirementInputSpec
}

// Verified mainnet unspent outputs at canonical anchor 1,726,905
// (0000193e7b4c05dd28f56bc7c9180364dab3357b5a4a23f2a776fc44d7b61901).
// Split by incident family so each burn transaction stays under MaxAmount
// (21,000,000 EXCC) with zero fee and zero spendable outputs.
var mainNetRetirementSpecs = []retirementTxSpec{
	{
		// Family 1 (origin height 1,725,679): 999,986.62531005 EXCC.
		tag: "EXCC-RETIRE-20261124-F1",
		inputs: []retirementInputSpec{
			{
				txid:        "4d3b79348f9b795928c8506b3cb9a8b082c054feb800d65fdf22c8303d17c236",
				vout:        0,
				atoms:       88_715_528_331_005,
				blockHeight: 1_726_805,
				blockIndex:  1,
			},
			{
				txid:        "4d3b79348f9b795928c8506b3cb9a8b082c054feb800d65fdf22c8303d17c236",
				vout:        1,
				atoms:       6_283_134_200_000,
				blockHeight: 1_726_805,
				blockIndex:  1,
			},
			{
				txid:        "e0f645c9e04b8c79de41aae4d9b7cb3a6c634bd11d9f6085f8e6dcb58d23c4f1",
				vout:        0,
				atoms:       5_000_000_000_000,
				blockHeight: 1_725_948,
				blockIndex:  1,
			},
		},
	},
	{
		// Family 2 (origin height 1,726,718): 20,999,999.89990000 EXCC.
		tag: "EXCC-RETIRE-20261124-F2",
		inputs: []retirementInputSpec{
			{
				txid:        "d55f267cceea4eeb97c26cc6235cfcbf7b57d576266dbf1c06de08696e476733",
				vout:        0,
				atoms:       100_000_000_000,
				blockHeight: 1_726_804,
				blockIndex:  1,
			},
			{
				txid:        "d8e6904a1e6d5c443c29c8edab386ba51a23e3b20b28ef1c581fd224e9998fda",
				vout:        0,
				atoms:       50_000_000_000_000,
				blockHeight: 1_726_810,
				blockIndex:  1,
			},
			{
				txid:        "d8e6904a1e6d5c443c29c8edab386ba51a23e3b20b28ef1c581fd224e9998fda",
				vout:        1,
				atoms:       30_000_000_000_000,
				blockHeight: 1_726_810,
				blockIndex:  1,
			},
			{
				txid:        "d8e6904a1e6d5c443c29c8edab386ba51a23e3b20b28ef1c581fd224e9998fda",
				vout:        2,
				atoms:       2_019_899_989_990_000,
				blockHeight: 1_726_810,
				blockIndex:  1,
			},
		},
	},
}

// buildBurnTx constructs a deterministic zero-fee OP_RETURN burn transaction.
func buildBurnTx(spec retirementTxSpec) (*dcrutil.Tx, error) {
	msgTx := wire.NewMsgTx()
	msgTx.SerType = wire.TxSerializeFull
	msgTx.Version = 1

	var totalAtoms int64
	for _, in := range spec.inputs {
		hash, err := chainhash.NewHashFromStr(in.txid)
		if err != nil {
			return nil, fmt.Errorf("invalid retirement input hash %q: %w", in.txid, err)
		}
		op := wire.NewOutPoint(hash, in.vout, wire.TxTreeRegular)
		txIn := wire.NewTxIn(op, in.atoms, nil)
		txIn.BlockHeight = in.blockHeight
		txIn.BlockIndex = in.blockIndex
		msgTx.AddTxIn(txIn)

		var ok bool
		totalAtoms, ok = addAmount(totalAtoms, in.atoms)
		if !ok || totalAtoms <= 0 || totalAtoms > dcrutil.MaxAmount {
			return nil, fmt.Errorf("retirement tx %q total %d out of range", spec.tag, totalAtoms)
		}
	}

	pkScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).
		AddData([]byte(spec.tag)).
		Script()
	if err != nil {
		return nil, fmt.Errorf("building retirement OP_RETURN script: %w", err)
	}
	if !txscript.IsUnspendable(totalAtoms, pkScript) {
		return nil, fmt.Errorf("retirement script for %q is not provably unspendable", spec.tag)
	}
	msgTx.AddTxOut(wire.NewTxOut(totalAtoms, pkScript))

	tx := dcrutil.NewTx(msgTx)
	tx.SetTree(wire.TxTreeRegular)
	return tx, nil
}

// NewRetirementPlan compiles a RetirementPlan from canonical burn transactions.
func NewRetirementPlan(activationHeight int64, txs []*dcrutil.Tx) (*RetirementPlan, error) {
	plan := &RetirementPlan{
		ActivationHeight: activationHeight,
		Txs:              make([]*dcrutil.Tx, 0, len(txs)),
		txFullHashes:     make(map[chainhash.Hash]int, len(txs)),
	}
	seenOutpoints := make(map[wire.OutPoint]struct{})
	for i, src := range txs {
		msgCopy := src.MsgTx().Copy()
		tx := dcrutil.NewTx(msgCopy)
		tx.SetTree(wire.TxTreeRegular)
		if len(msgCopy.TxIn) == 0 || len(msgCopy.TxOut) != 1 {
			return nil, fmt.Errorf("retirement tx %d must have >=1 input and 1 OP_RETURN output", i)
		}
		out := msgCopy.TxOut[0]
		if out.Version != 0 || !txscript.IsUnspendable(out.Value, out.PkScript) {
			return nil, fmt.Errorf("retirement tx %d output must be version-0 provably unspendable", i)
		}
		var sumIn int64
		for _, txIn := range msgCopy.TxIn {
			if len(txIn.SignatureScript) != 0 {
				return nil, fmt.Errorf("retirement tx %d must have empty signature scripts", i)
			}
			if _, dup := seenOutpoints[txIn.PreviousOutPoint]; dup {
				return nil, fmt.Errorf("duplicate retirement input %v", txIn.PreviousOutPoint)
			}
			seenOutpoints[txIn.PreviousOutPoint] = struct{}{}
			plan.outpoints = append(plan.outpoints, txIn.PreviousOutPoint)
			var ok bool
			sumIn, ok = addAmount(sumIn, txIn.ValueIn)
			if !ok || sumIn > dcrutil.MaxAmount {
				return nil, fmt.Errorf("retirement tx %d input sum overflows MaxAmount", i)
			}
		}
		if sumIn != out.Value {
			return nil, fmt.Errorf("retirement tx %d must have zero fee (in=%d out=%d)", i, sumIn, out.Value)
		}
		fullHash := msgCopy.TxHashFull()
		if _, dup := plan.txFullHashes[fullHash]; dup {
			return nil, fmt.Errorf("duplicate retirement tx %v", fullHash)
		}
		plan.txFullHashes[fullHash] = i
		plan.Txs = append(plan.Txs, tx)
	}
	return plan, nil
}

var defaultMainNetRetirementPlan = mustBuildMainNetRetirementPlan()

func mustBuildMainNetRetirementPlan() *RetirementPlan {
	txs := make([]*dcrutil.Tx, 0, len(mainNetRetirementSpecs))
	for _, spec := range mainNetRetirementSpecs {
		tx, err := buildBurnTx(spec)
		if err != nil {
			panic(err)
		}
		txs = append(txs, tx)
	}
	plan, err := NewRetirementPlan(mainNetRetirementHeight, txs)
	if err != nil {
		panic(err)
	}
	return plan
}

// MainNetRetirementTxs returns deep copies of the two canonical Option B
// mainnet OP_RETURN burn transactions.
func MainNetRetirementTxs() []*dcrutil.Tx {
	out := make([]*dcrutil.Tx, len(defaultMainNetRetirementPlan.Txs))
	for i, tx := range defaultMainNetRetirementPlan.Txs {
		cp := dcrutil.NewTx(tx.MsgTx().Copy())
		cp.SetTree(wire.TxTreeRegular)
		out[i] = cp
	}
	return out
}

func (p *RetirementPlan) isActive(height int64) bool {
	return p != nil && p.ActivationHeight > 0 && height >= p.ActivationHeight
}

func (p *RetirementPlan) matchesTx(height int64, tx *dcrutil.Tx) bool {
	if !p.isActive(height) || tx == nil {
		return false
	}
	_, ok := p.txFullHashes[tx.MsgTx().TxHashFull()]
	return ok
}

func txHasUnspentInput(tx *dcrutil.Tx, view *UtxoViewpoint) bool {
	for _, txIn := range tx.MsgTx().TxIn {
		entry := view.LookupEntry(txIn.PreviousOutPoint)
		if entry != nil && !entry.IsSpent() {
			return true
		}
	}
	return false
}

// requiredTxs returns the ordered canonical retirement transactions whose
// inputs remain unspent in view.
func (p *RetirementPlan) requiredTxs(view *UtxoViewpoint) []*dcrutil.Tx {
	var required []*dcrutil.Tx
	for _, tx := range p.Txs {
		if txHasUnspentInput(tx, view) {
			required = append(required, tx)
		}
	}
	return required
}

// SetRetirementPlan overrides the chain's active retirement plan (for tests
// and private rehearsal networks). Mainnet uses its built-in plan unless
// explicitly overridden in isolated tests.
func (b *BlockChain) SetRetirementPlan(plan *RetirementPlan) {
	b.chainLock.Lock()
	b.retirementPlan = plan
	b.chainLock.Unlock()
}

func (b *BlockChain) activeRetirementPlan(height int64) *RetirementPlan {
	plan := b.retirementPlan
	if plan == nil && b.chainParams.Net == wire.MainNet {
		plan = defaultMainNetRetirementPlan
	}
	if !plan.isActive(height) {
		return nil
	}
	return plan
}

func (b *BlockChain) isCanonicalRetirementTx(height int64, tx *dcrutil.Tx) bool {
	return b.activeRetirementPlan(height).matchesTx(height, tx)
}

// IsCanonicalRetirementTx reports whether tx is an exact canonical retirement
// transaction active for the given network and height.
func (b *BlockChain) IsCanonicalRetirementTx(height int64, tx *dcrutil.Tx) bool {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return b.isCanonicalRetirementTx(height, tx)
}

// IsMainNetRetirementTx reports whether tx matches a canonical mainnet
// retirement transaction when mainnet retirement is active at height.
func IsMainNetRetirementTx(params *chaincfg.Params, height int64, tx *dcrutil.Tx) bool {
	if params == nil || params.Net != wire.MainNet {
		return false
	}
	return defaultMainNetRetirementPlan.matchesTx(height, tx)
}

// checkBlockRetirementTxs verifies that any active retirement transactions
// whose inputs remain unspent in view (after parent disapproval) appear
// verbatim at the start of the regular transaction tree immediately after the
// coinbase transaction.
func (b *BlockChain) checkBlockRetirementTxs(height int64, block *dcrutil.Block, view *UtxoViewpoint) error {
	plan := b.activeRetirementPlan(height)
	if plan == nil {
		return nil
	}

	// Fetch any retirement outpoints not already loaded into view (preserving
	// any entries restored by parent disapproval).
	filteredSet := make(ViewFilteredSet, len(plan.outpoints))
	for i := range plan.outpoints {
		filteredSet.add(view, &plan.outpoints[i])
	}
	if err := view.fetchUtxosMain(filteredSet); err != nil {
		return err
	}

	required := plan.requiredTxs(view)
	regTxns := block.Transactions()
	if len(regTxns) < 1+len(required) {
		str := fmt.Sprintf("block at height %d missing %d required retirement transaction(s)",
			height, len(required))
		return ruleError(ErrBadTxInput, str)
	}
	for i, want := range required {
		got := regTxns[1+i]
		if got.MsgTx().TxHashFull() != want.MsgTx().TxHashFull() {
			str := fmt.Sprintf("block at height %d regular tx %d is %v, want canonical retirement tx %v",
				height, 1+i, got.MsgTx().TxHashFull(), want.MsgTx().TxHashFull())
			return ruleError(ErrBadTxInput, str)
		}
	}

	// Reject any canonical retirement transaction appearing outside its
	// required prefix positions.
	for i := 1 + len(required); i < len(regTxns); i++ {
		if plan.matchesTx(height, regTxns[i]) {
			str := fmt.Sprintf("block at height %d contains unexpected retirement tx %v at index %d",
				height, regTxns[i].Hash(), i)
			return ruleError(ErrBadTxInput, str)
		}
	}
	return nil
}

// RequiredRetirementTxs returns deep copies of the canonical retirement
// transactions that must be included immediately after the coinbase in a block
// at nextHeight. When includeRegularTxns is false, the current tip's regular
// transaction tree is treated as disapproved before checking UTXO state.
func (b *BlockChain) RequiredRetirementTxs(nextHeight int64, includeRegularTxns bool) ([]*dcrutil.Tx, error) {
	b.chainLock.RLock()
	plan := b.activeRetirementPlan(nextHeight)
	b.chainLock.RUnlock()
	if plan == nil {
		return nil, nil
	}

	var required []*dcrutil.Tx
	for _, tx := range plan.Txs {
		view, err := b.FetchUtxoView(tx, includeRegularTxns)
		if err != nil {
			return nil, err
		}
		if txHasUnspentInput(tx, view) {
			cp := dcrutil.NewTx(tx.MsgTx().Copy())
			cp.SetTree(wire.TxTreeRegular)
			required = append(required, cp)
		}
	}
	return required, nil
}
