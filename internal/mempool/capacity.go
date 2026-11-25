// Copyright (c) 2026 The EXCC developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"bytes"
	"cmp"
	"container/heap"
	"math/bits"
	"slices"

	"github.com/EXCCoin/exccd/blockchain/stake/v4"
	"github.com/EXCCoin/exccd/blockchain/v4"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/dcrutil/v4"
)

const (
	// DefaultMaxMempoolBytes bounds combined serialized transaction bytes in
	// the main and staged pools. In-memory index overhead is additional.
	DefaultMaxMempoolBytes = 64 << 20

	// DefaultMaxMempoolTxs bounds the combined main and staged tx count.
	DefaultMaxMempoolTxs = 50000
)

type capacityEntry struct {
	desc  *TxDesc
	index int // position in leaves, or -1 when it has dependents
}

// compareRetention ranks votes above revocations, then fee-paying traffic.
// Within each class compare exact fees per serialized byte.
func compareRetention(a, b *capacityEntry) int {
	priority := func(d *TxDesc) int {
		switch d.Type {
		case stake.TxTypeSSGen:
			return 2
		case stake.TxTypeSSRtx:
			return 1
		}
		return 0
	}
	if c := cmp.Compare(priority(a.desc), priority(b.desc)); c != 0 {
		return c
	}
	ahi, alo := bits.Mul64(uint64(a.desc.Fee), uint64(b.desc.TxSize))
	bhi, blo := bits.Mul64(uint64(b.desc.Fee), uint64(a.desc.TxSize))
	if c := cmp.Compare(ahi, bhi); c != 0 {
		return c
	}
	return cmp.Compare(alo, blo)
}

type evictionHeap []*capacityEntry

func (h evictionHeap) Len() int { return len(h) }
func (h evictionHeap) Less(i, j int) bool {
	if c := compareRetention(h[i], h[j]); c != 0 {
		return c < 0
	}
	return bytes.Compare(h[i].desc.Tx.Hash()[:], h[j].desc.Tx.Hash()[:]) < 0
}
func (h evictionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index, h[j].index = i, j
}
func (h *evictionHeap) Push(item interface{}) {
	entry := item.(*capacityEntry)
	entry.index = len(*h)
	*h = append(*h, entry)
}
func (h *evictionHeap) Pop() interface{} {
	last := len(*h) - 1
	entry := (*h)[last]
	entry.index = -1
	(*h)[last] = nil
	*h = (*h)[:last]
	return entry
}

type parentKey struct {
	hash chainhash.Hash
	tree int8
}

// poolCapacity tracks main and staged transactions under the pool write lock.
// Only transactions without dependents appear in the eviction heap.
type poolCapacity struct {
	// Count input references even when their parent is mined. A reorg can
	// reintroduce a parent after several existing main or staged dependents.
	references map[parentKey]int
	entries    map[chainhash.Hash]*capacityEntry
	leaves     evictionHeap
	bytes      int64
	maxBytes   int64
	maxTxs     int
}

func newPoolCapacity(p Policy) poolCapacity {
	if p.MaxMempoolTxs <= 0 {
		p.MaxMempoolTxs = DefaultMaxMempoolTxs
	}
	if p.MaxMempoolBytes <= 0 {
		p.MaxMempoolBytes = DefaultMaxMempoolBytes
	}
	return poolCapacity{
		references: make(map[parentKey]int),
		entries:    make(map[chainhash.Hash]*capacityEntry),
		maxBytes:   p.MaxMempoolBytes,
		maxTxs:     p.MaxMempoolTxs,
	}
}

func (c *poolCapacity) updateParents(tx *dcrutil.Tx, delta int) {
	for _, in := range tx.MsgTx().TxIn {
		op := &in.PreviousOutPoint
		key := parentKey{op.Hash, op.Tree}
		before := c.references[key]
		after := before + delta
		if after == 0 {
			delete(c.references, key)
		} else {
			c.references[key] = after
		}
		parent := c.entries[op.Hash]
		if parent == nil || parent.desc.Tx.Tree() != op.Tree {
			continue
		}
		if before == 0 {
			heap.Remove(&c.leaves, parent.index)
		} else if after == 0 {
			heap.Push(&c.leaves, parent)
		}
	}
}

// isWinningVote uses the local tip and its lottery results. A claimed height
// or a vote for another block does not receive the current winners' reserve.
func isWinningVote(desc *TxDesc, best *blockchain.BestState) bool {
	if desc.Type != stake.TxTypeSSGen || best == nil {
		return false
	}
	vote := desc.Tx.MsgTx()
	hash, height := stake.SSGenBlockVotedOn(vote)
	return hash == best.Hash && int64(height) == best.Height &&
		slices.Contains(best.NextWinningTickets, vote.TxIn[1].PreviousOutPoint.Hash)
}

func (c *poolCapacity) insert(entry *capacityEntry) {
	tx := entry.desc.Tx
	c.updateParents(tx, 1)
	c.entries[*tx.Hash()] = entry
	c.bytes += entry.desc.TxSize
	if c.references[parentKey{*tx.Hash(), tx.Tree()}] == 0 {
		heap.Push(&c.leaves, entry)
	}
}

func (c *poolCapacity) remove(hash *chainhash.Hash) {
	entry := c.entries[*hash]
	if entry.index >= 0 {
		heap.Remove(&c.leaves, entry.index)
	}
	delete(c.entries, *hash)
	c.bytes -= entry.desc.TxSize
	c.updateParents(entry.desc.Tx, -1)
}

// makeRoom plans all removals before changing the live pool, its mining view,
// fee estimates or any other externally observable state. Pin the candidate's
// parents and consider newly released ancestors as their children are removed.
func (c *poolCapacity) makeRoom(desc *TxDesc, snapshot func() *blockchain.BestState) ([]*TxDesc, error) {
	fits := func() bool {
		return len(c.entries) < c.maxTxs && desc.TxSize <= c.maxBytes-c.bytes
	}
	if fits() {
		return nil, nil
	}
	if desc.TxSize > c.maxBytes {
		return nil, txRuleError(ErrMempoolFull, "transaction exceeds mempool byte limit")
	}
	var best *blockchain.BestState
	if snapshot != nil {
		best = snapshot()
	}
	incoming := &capacityEntry{desc: desc, index: -1}
	winning := isWinningVote(desc, best)
	var protected []*capacityEntry
	c.updateParents(desc.Tx, 1)
	var evicted []*capacityEntry
	defer func() {
		for _, e := range protected {
			heap.Push(&c.leaves, e)
		}
		for i := len(evicted) - 1; i >= 0; i-- {
			c.insert(evicted[i])
		}
		c.updateParents(desc.Tx, -1)
	}()
	for !fits() {
		if len(c.leaves) == 0 {
			return nil, txRuleError(ErrMempoolFull, "pool capacity is reserved for dependencies or winning votes")
		}
		victim := c.leaves[0]
		if isWinningVote(victim.desc, best) {
			protected = append(protected, heap.Pop(&c.leaves).(*capacityEntry))
			continue
		}
		if !winning && compareRetention(incoming, victim) <= 0 {
			return nil, txRuleError(ErrMempoolFull, "insufficient room in mempool at this fee rate and stake priority")
		}
		evicted = append(evicted, victim)
		c.remove(victim.desc.Tx.Hash())
	}
	victims := make([]*TxDesc, len(evicted))
	for i, e := range evicted {
		victims[i] = e.desc
	}
	return victims, nil
}
