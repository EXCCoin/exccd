// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/txscript/v4/stdaddr"
	"github.com/EXCCoin/exccd/wire"
)

// lotteryDataForNode is a helper function that returns winning tickets
// along with the ticket pool size and PRNG checksum for a given node.
//
// This function is NOT safe for concurrent access and MUST be called
// with the chainLock held for writes.
func (b *BlockChain) lotteryDataForNode(node *blockNode) ([]chainhash.Hash, int, [6]byte, error) {
	if node.height < b.chainParams.StakeEnabledHeight {
		return nil, 0, [6]byte{}, nil
	}
	stakeNode, err := b.fetchStakeNode(node)
	if err != nil {
		return nil, 0, [6]byte{}, err
	}

	return stakeNode.Winners(), stakeNode.PoolSize(), stakeNode.FinalState(), nil
}

// lotteryDataForBlock takes a node block hash and returns the next tickets
// eligible for voting, the number of tickets in the ticket pool, and the
// final state of the PRNG.
//
// This function is NOT safe for concurrent access and must have the chainLock
// held for write access.
func (b *BlockChain) lotteryDataForBlock(hash *chainhash.Hash) ([]chainhash.Hash, int, [6]byte, error) {
	node := b.index.LookupNode(hash)
	if node == nil {
		return nil, 0, [6]byte{}, unknownBlockError(hash)
	}

	winningTickets, poolSize, finalState, err := b.lotteryDataForNode(node)
	if err != nil {
		return nil, 0, [6]byte{}, err
	}

	return winningTickets, poolSize, finalState, nil
}

// LotteryDataForBlock returns lottery data for a given block in the block
// chain, including side chain blocks.
//
// It is safe for concurrent access.
func (b *BlockChain) LotteryDataForBlock(hash *chainhash.Hash) ([]chainhash.Hash, int, [6]byte, error) {
	b.chainLock.Lock()
	winningTickets, poolSize, finalState, err := b.lotteryDataForBlock(hash)
	b.chainLock.Unlock()
	return winningTickets, poolSize, finalState, err
}

// LiveTickets returns all currently live tickets from the stake database.
//
// This function is safe for concurrent access.
func (b *BlockChain) LiveTickets() ([]chainhash.Hash, error) {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	return sn.LiveTickets(), nil
}

// MissedTickets returns all currently missed tickets from the stake database.
//
// This function is safe for concurrent access.
func (b *BlockChain) MissedTickets() ([]chainhash.Hash, error) {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	return sn.MissedTickets(), nil
}

// TicketsWithAddress returns a slice of ticket hashes that are currently live
// corresponding to the given address.
//
// This function is safe for concurrent access.
func (b *BlockChain) TicketsWithAddress(address stdaddr.StakeAddress) ([]chainhash.Hash, error) {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	var ticketsWithAddr []chainhash.Hash
	votingRightsScriptVer, votingRightsScript := address.VotingRightsScript()
	tickets := sn.LiveTickets()
	for _, hash := range tickets {
		outpoint := wire.OutPoint{Hash: hash, Index: 0, Tree: wire.TxTreeStake}
		utxo, err := b.utxoCache.FetchEntry(outpoint)
		if err != nil {
			return nil, err
		}

		if utxo.ScriptVersion() == votingRightsScriptVer &&
			bytes.Equal(utxo.PkScript(), votingRightsScript) {

			ticketsWithAddr = append(ticketsWithAddr, hash)
		}
	}

	return ticketsWithAddr, nil
}

// CheckLiveTicket returns whether or not a ticket exists in the live ticket
// treap of the best node.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckLiveTicket(hash chainhash.Hash) bool {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	return sn.ExistsLiveTicket(hash)
}

// CheckLiveTickets returns a slice of bools representing whether each ticket
// exists in the live ticket treap of the best node.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckLiveTickets(hashes []chainhash.Hash) []bool {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	existsSlice := make([]bool, len(hashes))
	for i := range hashes {
		existsSlice[i] = sn.ExistsLiveTicket(hashes[i])
	}

	return existsSlice
}

// CheckMissedTickets returns a slice of bools representing whether each ticket
// hash has been missed in the live ticket treap of the best node.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckMissedTickets(hashes []chainhash.Hash) []bool {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	existsSlice := make([]bool, len(hashes))
	for i := range hashes {
		existsSlice[i] = sn.ExistsMissedTicket(hashes[i])
	}

	return existsSlice
}

// CheckExpiredTicket returns whether or not a ticket was ever expired.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckExpiredTicket(hash chainhash.Hash) bool {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	return sn.ExistsExpiredTicket(hash)
}

// CheckExpiredTickets returns whether or not a ticket in a slice of
// tickets was ever expired.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckExpiredTickets(hashes []chainhash.Hash) []bool {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	existsSlice := make([]bool, len(hashes))
	for i := range hashes {
		existsSlice[i] = sn.ExistsExpiredTicket(hashes[i])
	}

	return existsSlice
}

// TicketPoolValue returns the current value of all the locked funds in the
// ticket pool.
//
// This function is safe for concurrent access.  All live tickets are at least
// 256 blocks deep on mainnet, so the UTXO set should generally always have
// the asked for transactions.
func (b *BlockChain) TicketPoolValue() (dcrutil.Amount, error) {
	b.chainLock.RLock()
	sn := b.bestChain.Tip().stakeNode
	b.chainLock.RUnlock()

	var amt int64
	for _, hash := range sn.LiveTickets() {
		outpoint := wire.OutPoint{Hash: hash, Index: 0, Tree: wire.TxTreeStake}
		utxo, err := b.utxoCache.FetchEntry(outpoint)
		if err != nil {
			return 0, err
		}

		amt += utxo.Amount()
	}
	return dcrutil.Amount(amt), nil
}
