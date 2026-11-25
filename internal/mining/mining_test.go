// Copyright (c) 2020-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	"github.com/EXCCoin/exccd/blockchain/v4"
	"github.com/EXCCoin/exccd/blockchain/v4/chaingen"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/txscript/v4"
	"github.com/EXCCoin/exccd/txscript/v4/stdaddr"
	"github.com/EXCCoin/exccd/wire"
)

func solveTemplate(t *testing.T, block *wire.MsgBlock, params *chaincfg.Params) {
	t.Helper()
	if !chaingen.SolveBlockWithEquihash(&block.Header, params) {
		t.Fatal("unable to solve block template")
	}
}

// TestNewBlockTemplateBasicErrorScenarios tests various basic error scenarios
// that can occur during new block template generation.
func TestNewBlockTemplateBasicErrorScenarios(t *testing.T) {
	t.Parallel()

	// Create a new mining harness instance.
	harness, _, err := newMiningHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("error creating mining harness: %v", err)
	}

	// Create a test address for use in template generation.
	address, err := stdaddr.DecodeAddress("TsebVL4jcgJ2uVqZS2HutWTb4KnJLc9FdLR",
		harness.chainParams)
	if err != nil {
		t.Fatalf("error decoding address: %v", err)
	}

	// Test error retrieving standard verify flags.
	standardVerifyFlags := harness.policy.StandardVerifyFlags
	var errFlags = errors.New("error retrieving standard verify flags")
	harness.policy.StandardVerifyFlags = func() (txscript.ScriptFlags, error) {
		return 0, errFlags
	}
	_, err = harness.generator.NewBlockTemplate(address)
	if !errors.Is(err, errFlags) {
		t.Fatalf("unexpected error retrieving standard verify flags -- got %v, "+
			"want %v", err, errFlags)
	}
	harness.policy.StandardVerifyFlags = standardVerifyFlags

	// Test error retrieving treasury agenda.
	var errTreasuryAgenda = errors.New("error retrieving treasury agenda")
	harness.chain.isTreasuryAgendaActiveErr = errTreasuryAgenda
	_, err = harness.generator.NewBlockTemplate(address)
	if !errors.Is(err, errTreasuryAgenda) {
		t.Fatalf("unexpected error retrieving treasury agenda -- got %v, want %v",
			err, errTreasuryAgenda)
	}
	harness.chain.isTreasuryAgendaActiveErr = nil
}

// TestNewBlockTemplate tests the generation of a new block template containing
// regular and vote transactions.
func TestNewBlockTemplate(t *testing.T) {
	t.Parallel()

	// Create a new mining harness instance.
	harness, spendableOuts, err := newMiningHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("error creating mining harness: %v", err)
	}

	// Create a test address for use in template generation.
	address, err := stdaddr.DecodeAddress("TsebVL4jcgJ2uVqZS2HutWTb4KnJLc9FdLR",
		harness.chainParams)
	if err != nil {
		t.Fatalf("error decoding address: %v", err)
	}

	// Define a munger to apply transaction fees.
	applyTxFee := func(fee int64) func(*wire.MsgTx) {
		return func(tx *wire.MsgTx) {
			tx.TxOut[0].Value -= fee
		}
	}

	// Create additional transactions from the first spendable output provided by
	// the harness.
	const numTxs = 6
	txs := make([]*dcrutil.Tx, numTxs)
	baseTx, err := harness.CreateSignedTx(spendableOuts, uint32(numTxs))
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	harness.AddFakeUTXO(baseTx, harness.chain.bestState.Height, 1,
		harness.chain.isTreasuryAgendaActive)
	for i := 0; i < numTxs; i++ {
		tx, err := harness.CreateSignedTx([]spendableOutput{
			txOutToSpendableOut(baseTx, uint32(i), wire.TxTreeRegular)}, 1,
			applyTxFee(5000))
		if err != nil {
			t.Fatalf("unable to create transaction: %v", err)
		}
		txs[i] = tx
	}

	// Create ticket purchase transactions spending the outputs of the prior
	// regular transactions.
	const numVotes = 5
	tickets := make([]*dcrutil.Tx, numVotes)
	ticketHashes := make([]chainhash.Hash, numVotes)
	for i := 0; i < numVotes; i++ {
		ticket, err := harness.CreateTicketPurchase(txs[i], 40000)
		if err != nil {
			t.Fatalf("unable to create ticket purchase transaction: %v", err)
		}
		tickets[i] = ticket
		ticketHashes[i] = ticket.MsgTx().TxHash()
	}

	// Add the ticket outputs as utxos to fake their existence.  Use one after
	// the stake enabled height for the height of the fake utxos to ensure they
	// are mature for the votes cast at stake validation height below.
	harness.chain.bestState.Height = harness.chainParams.StakeEnabledHeight + 1
	for i, ticket := range tickets {
		harness.AddFakeUTXO(ticket, harness.chain.bestState.Height, uint32(i+1),
			harness.chain.isTreasuryAgendaActive)
	}

	// Create votes on a block at stake validation height using the previously
	// created tickets.
	harness.chain.bestState.Height = harness.chainParams.StakeValidationHeight
	harness.chain.bestState.NextWinningTickets = ticketHashes
	votes := make([]*dcrutil.Tx, numVotes)
	for i, ticket := range tickets {
		vote, err := harness.CreateVote(ticket)
		if err != nil {
			t.Fatalf("unable to create vote: %v", err)
		}
		votes[i] = vote
	}

	// Add vote transactions to the tx source.
	for _, vote := range votes {
		_, err = harness.AddTransactionToTxSource(vote)
		if err != nil {
			t.Fatalf("unable to add transaction to the tx source: %v", err)
		}
	}

	// Add remaining regular transactions to the tx source.
	for i := numVotes; i < numTxs; i++ {
		_, err = harness.AddTransactionToTxSource(txs[i])
		if err != nil {
			t.Fatalf("unable to add transaction to the tx source: %v", err)
		}
	}

	// Generate a new block template.
	blockTemplate, err := harness.generator.NewBlockTemplate(address)
	if err != nil {
		t.Fatalf("unexpected err generating block template: %v", err)
	}
	if got := blockTemplate.Block.Header.Version; got != generatedBlockVersion {
		t.Fatalf("block version = %d, want %d", got, generatedBlockVersion)
	}

	// Validate the number of transactions in the generated block template.
	gotTx := len(blockTemplate.Block.Transactions)
	wantTx := numTxs - numVotes + 1 // + 1 for coinbase.
	if gotTx != wantTx {
		t.Fatalf("unexpected number of transactions in template --  got %v, want %v",
			gotTx, wantTx)
	}

	// Validate the number of stake transactions in the generated block template.
	gotStx := len(blockTemplate.Block.STransactions)
	wantStx := numVotes
	if gotStx != wantStx {
		t.Fatalf("unexpected number of stake transactions in template --  got %v, "+
			"want %v", gotStx, wantStx)
	}

	// Validate that the block is sane.  These checks are context free.
	solveTemplate(t, blockTemplate.Block, harness.chainParams)
	block := dcrutil.NewBlock(blockTemplate.Block)
	err = blockchain.CheckBlockSanity(block, harness.generator.cfg.TimeSource,
		harness.chainParams)
	if err != nil {
		t.Fatalf("unexpected error when checking block sanity: %v", err)
	}
}

// TestNewBlockTemplateAutoRevocations tests the generation of a new block with
// automatic ticket revocations enabled.
func TestNewBlockTemplateAutoRevocations(t *testing.T) {
	t.Parallel()

	// Create a new mining harness instance.
	harness, spendableOuts, err := newMiningHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("error creating mining harness: %v", err)
	}

	// Activate the automatic ticket revocations agenda.
	harness.chain.isAutoRevocationsAgendaActive = true

	// Create a test address for use in template generation.
	address, err := stdaddr.DecodeAddress("TsebVL4jcgJ2uVqZS2HutWTb4KnJLc9FdLR",
		harness.chainParams)
	if err != nil {
		t.Fatalf("error decoding address: %v", err)
	}

	// Define a munger to apply transaction fees.
	applyTxFee := func(fee int64) func(*wire.MsgTx) {
		return func(tx *wire.MsgTx) {
			tx.TxOut[0].Value -= fee
		}
	}

	// Create additional transactions from the first spendable output provided by
	// the harness.
	const numTxs = 11
	txs := make([]*dcrutil.Tx, numTxs)
	baseTx, err := harness.CreateSignedTx(spendableOuts, uint32(numTxs))
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	harness.AddFakeUTXO(baseTx, harness.chain.bestState.Height, 1,
		harness.chain.isTreasuryAgendaActive)
	for i := 0; i < numTxs; i++ {
		tx, err := harness.CreateSignedTx([]spendableOutput{
			txOutToSpendableOut(baseTx, uint32(i), wire.TxTreeRegular)}, 1,
			applyTxFee(5000))
		if err != nil {
			t.Fatalf("unable to create transaction: %v", err)
		}
		txs[i] = tx
	}

	// Create ticket purchase transactions spending the outputs of the prior
	// regular transactions.
	const numTickets = 10
	tickets := make([]*dcrutil.Tx, numTickets)
	ticketHashes := make([]chainhash.Hash, numTickets)
	for i := 0; i < numTickets; i++ {
		ticket, err := harness.CreateTicketPurchase(txs[i], 40000)
		if err != nil {
			t.Fatalf("unable to create ticket purchase transaction: %v", err)
		}
		tickets[i] = ticket
		ticketHashes[i] = ticket.MsgTx().TxHash()
	}

	// Add the ticket outputs as utxos to fake their existence.  Use one after
	// the stake enabled height for the height of the fake utxos to ensure they
	// are mature for the votes cast at stake validation height below.
	harness.chain.bestState.Height = harness.chainParams.StakeEnabledHeight + 1
	for i, ticket := range tickets {
		harness.AddFakeUTXO(ticket, harness.chain.bestState.Height, uint32(i+1),
			harness.chain.isTreasuryAgendaActive)
	}

	// Create votes on a block at stake validation height using the previously
	// created tickets.
	//
	// Create votes for 3 out of 5 winning tickets.  The 2 remaining winning
	// tickets will be missed and will need to be revoked.
	winningTickets := ticketHashes[:5]
	voteTickets := winningTickets[:3]
	missedThisBlock := winningTickets[3:]
	harness.chain.bestState.Height = harness.chainParams.StakeValidationHeight
	harness.chain.bestState.NextWinningTickets = winningTickets
	numVotes := len(voteTickets)
	votes := make([]*dcrutil.Tx, numVotes)
	for i, ticket := range tickets[:numVotes] {
		vote, err := harness.CreateVote(ticket)
		if err != nil {
			t.Fatalf("unable to create vote: %v", err)
		}
		votes[i] = vote
	}

	// Add vote transactions to the tx source.
	for _, vote := range votes {
		_, err = harness.AddTransactionToTxSource(vote)
		if err != nil {
			t.Fatalf("unable to add transaction to the tx source: %v", err)
		}
	}

	// Add 2 of the remaining non-winning tickets to the list of tickets expiring
	// next block to simulate expiring tickets that will need to be revoked.
	expiredThisBlock := ticketHashes[5:7]
	harness.chain.bestState.NextExpiringTickets = expiredThisBlock

	// Add 2 of the remaining non-winning tickets to the list of tickets that were
	// already missed.
	missedPreviously := ticketHashes[7:9]
	harness.chain.bestState.MissedTickets = missedPreviously

	// Use 1 ticket as a ticket that is ineligible to vote or be revoked.
	ineligibleTicket := ticketHashes[9]

	// Add an additional vote transaction with a bad input.  This ensures that a
	// revocation is still created for a missed ticket due to a vote transaction
	// failing validation.
	vote, err := harness.CreateVote(tickets[numVotes])
	if err != nil {
		t.Fatalf("unable to create vote: %v", err)
	}
	_, err = harness.AddTransactionToTxSource(vote)
	if err != nil {
		t.Fatalf("unable to add transaction to the tx source: %v", err)
	}
	vote.MsgTx().TxIn[0].ValueIn = 0

	// Get the best header bytes.
	bestHeader, err := harness.chain.HeaderByHash(&harness.chain.bestState.Hash)
	if err != nil {
		t.Fatalf("unable to get tip block header %v: %v",
			harness.chain.bestState.Hash, err)
	}
	bestHeaderBytes, err := bestHeader.Bytes()
	if err != nil {
		t.Fatalf("failed to serialize header for block %v: %v",
			bestHeader.BlockHash(), err)
	}

	// Add revocations for tickets that were previously missed.
	for _, ticketHash := range harness.chain.bestState.MissedTickets {
		txDesc, err := harness.generator.createRevocationFromTicket(&ticketHash,
			harness.chain.utxos, bestHeaderBytes,
			harness.chain.isTreasuryAgendaActive)
		if err != nil {
			t.Fatalf("failed to create revocation for ticket %v: %v", ticketHash, err)
		}
		_, err = harness.AddTransactionToTxSource(txDesc.Tx)
		if err != nil {
			t.Fatalf("unable to add transaction to the tx source: %v", err)
		}
	}

	// Add a revocation that tries to spend an ineligible ticket.
	txDesc, err := harness.generator.createRevocationFromTicket(&ineligibleTicket,
		harness.chain.utxos, bestHeaderBytes, harness.chain.isTreasuryAgendaActive)
	if err != nil {
		t.Fatalf("failed to create revocation for ticket %v: %v", ineligibleTicket,
			err)
	}
	_, err = harness.AddTransactionToTxSource(txDesc.Tx)
	if err != nil {
		t.Fatalf("unable to add transaction to the tx source: %v", err)
	}

	// Add remaining regular transactions to the tx source.
	for i := numVotes; i < numTxs; i++ {
		_, err = harness.AddTransactionToTxSource(txs[i])
		if err != nil {
			t.Fatalf("unable to add transaction to the tx source: %v", err)
		}
	}

	// Generate a new block template.
	blockTemplate, err := harness.generator.NewBlockTemplate(address)
	if err != nil {
		t.Fatalf("unexpected err generating block template: %v", err)
	}

	// Validate the number of transactions in the generated block template.
	gotTx := len(blockTemplate.Block.Transactions)
	wantTx := numTxs - numVotes + 1 // + 1 for coinbase.
	if gotTx != wantTx {
		t.Fatalf("unexpected number of transactions in template --  got %v, want %v",
			gotTx, wantTx)
	}

	// Determine the expected number of revocation transactions.
	numAutoRevocations := len(missedThisBlock) + len(expiredThisBlock)
	numTxSourceRevocations := len(missedPreviously)
	numRevocations := numAutoRevocations + numTxSourceRevocations

	// Validate the number of stake transactions in the generated block template.
	gotStx := len(blockTemplate.Block.STransactions)
	wantStx := numVotes + numRevocations
	if gotStx != wantStx {
		t.Fatalf("unexpected number of stake transactions in template --  got %v, "+
			"want %v", gotStx, wantStx)
	}

	// Validate that the block is sane.  These checks are context free.
	solveTemplate(t, blockTemplate.Block, harness.chainParams)
	block := dcrutil.NewBlock(blockTemplate.Block)
	err = blockchain.CheckBlockSanity(block, harness.generator.cfg.TimeSource,
		harness.chainParams)
	if err != nil {
		t.Fatalf("unexpected error when checking block sanity: %v", err)
	}
}

// TestNewBlockTemplateAutoRevocationsVotesOnly tests that auto revocations are
// properly added to blocks that only contain vote transactions.
func TestNewBlockTemplateAutoRevocationsVotesOnly(t *testing.T) {
	t.Parallel()

	// Create a new mining harness instance.
	harness, spendableOuts, err := newMiningHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("error creating mining harness: %v", err)
	}

	// Activate the automatic ticket revocations agenda.
	harness.chain.isAutoRevocationsAgendaActive = true

	// Create a test address for use in template generation.
	address, err := stdaddr.DecodeAddress("TsebVL4jcgJ2uVqZS2HutWTb4KnJLc9FdLR",
		harness.chainParams)
	if err != nil {
		t.Fatalf("error decoding address: %v", err)
	}

	// Define a munger to apply transaction fees.
	applyTxFee := func(fee int64) func(*wire.MsgTx) {
		return func(tx *wire.MsgTx) {
			tx.TxOut[0].Value -= fee
		}
	}

	// Create additional transactions from the first spendable output provided by
	// the harness.
	const numTxs = 6
	txs := make([]*dcrutil.Tx, numTxs)
	baseTx, err := harness.CreateSignedTx(spendableOuts, uint32(numTxs))
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	harness.AddFakeUTXO(baseTx, harness.chain.bestState.Height, 1,
		harness.chain.isTreasuryAgendaActive)
	for i := 0; i < numTxs; i++ {
		tx, err := harness.CreateSignedTx([]spendableOutput{
			txOutToSpendableOut(baseTx, uint32(i), wire.TxTreeRegular)}, 1,
			applyTxFee(5000))
		if err != nil {
			t.Fatalf("unable to create transaction: %v", err)
		}
		txs[i] = tx
	}

	// Create ticket purchase transactions spending the outputs of the prior
	// regular transactions.
	const numTickets = 5
	tickets := make([]*dcrutil.Tx, numTickets)
	ticketHashes := make([]chainhash.Hash, numTickets)
	for i := 0; i < numTickets; i++ {
		ticket, err := harness.CreateTicketPurchase(txs[i], 40000)
		if err != nil {
			t.Fatalf("unable to create ticket purchase transaction: %v", err)
		}
		tickets[i] = ticket
		ticketHashes[i] = ticket.MsgTx().TxHash()
	}

	// Add the ticket outputs as utxos to fake their existence.  Use one after
	// the stake enabled height for the height of the fake utxos to ensure they
	// are mature for the votes cast at stake validation height below.
	harness.chain.bestState.Height = harness.chainParams.StakeEnabledHeight + 1
	for i, ticket := range tickets {
		harness.AddFakeUTXO(ticket, harness.chain.bestState.Height, uint32(i+1),
			harness.chain.isTreasuryAgendaActive)
	}

	// Create votes on a block at stake validation height using the previously
	// created tickets.
	//
	// Create votes for 3 out of 5 winning tickets.  The 2 remaining winning
	// tickets will be missed and will need to be revoked.
	winningTickets := ticketHashes[:5]
	voteTickets := winningTickets[:3]
	missedThisBlock := winningTickets[3:]
	harness.chain.bestState.Height = harness.chainParams.StakeValidationHeight
	harness.chain.bestState.NextWinningTickets = winningTickets
	numVotes := len(voteTickets)
	votes := make([]*dcrutil.Tx, numVotes)
	for i, ticket := range tickets[:numVotes] {
		vote, err := harness.CreateVote(ticket)
		if err != nil {
			t.Fatalf("unable to create vote: %v", err)
		}
		votes[i] = vote
	}

	// Add vote transactions to the tx source.
	for _, vote := range votes {
		_, err = harness.AddTransactionToTxSource(vote)
		if err != nil {
			t.Fatalf("unable to add transaction to the tx source: %v", err)
		}
	}

	// Generate a new block template.
	blockTemplate, err := harness.generator.NewBlockTemplate(address)
	if err != nil {
		t.Fatalf("unexpected err generating block template: %v", err)
	}

	// Determine the expected number of revocation transactions.
	numAutoRevocations := len(missedThisBlock)

	// Validate the number of stake transactions in the generated block template.
	gotStx := len(blockTemplate.Block.STransactions)
	wantStx := numVotes + numAutoRevocations
	if gotStx != wantStx {
		t.Fatalf("unexpected number of stake transactions in template --  got %v, "+
			"want %v", gotStx, wantStx)
	}

	// Validate that the block is sane.  These checks are context free.
	solveTemplate(t, blockTemplate.Block, harness.chainParams)
	block := dcrutil.NewBlock(blockTemplate.Block)
	err = blockchain.CheckBlockSanity(block, harness.generator.cfg.TimeSource,
		harness.chainParams)
	if err != nil {
		t.Fatalf("unexpected error when checking block sanity: %v", err)
	}
}

// TestSortParentsByVotes ensures the function that sorts parent blocks based on
// the number of votes available for them and the current tip block works as
// intended, including reorg prevention for an equal number of votes.
func TestSortParentsByVotes(t *testing.T) {
	params := chaincfg.TestNet3Params()

	tests := []struct {
		name      string  // test description
		numBlocks uint8   // number of sibling blocks to generate
		votes     []uint8 // mock votes to generate per block
		curTip    uint8   // index of the current tip block
		want      []uint8 // indexes of expected sorted blocks
	}{{
		name:      "0 blocks, 0 votes (sanity check for impossible condition)",
		numBlocks: 0,
		votes:     nil,
		curTip:    0,
		want:      nil,
	}, {
		name:      "1 block, 5 votes",
		numBlocks: 1,
		votes:     []uint8{5},
		curTip:    0,
		want:      []uint8{0},
	}, {
		name:      "1 block, 4 votes",
		numBlocks: 1,
		votes:     []uint8{4},
		curTip:    0,
		want:      []uint8{0},
	}, {
		name:      "1 block, 3 votes",
		numBlocks: 1,
		votes:     []uint8{3},
		curTip:    0,
		want:      []uint8{0},
	}, {
		name:      "1 block, 2 votes (no useful blocks)",
		numBlocks: 1,
		votes:     []uint8{2},
		curTip:    0,
		want:      nil,
	}, {
		name:      "2 blocks, votes{0:5, 1:5}, tip block 0 (no reorg)",
		numBlocks: 2,
		votes:     []uint8{5, 5},
		curTip:    0,
		want:      []uint8{0, 1},
	}, {
		name:      "2 blocks, votes{0:5, 1:5}, tip block 1 (no reorg)",
		numBlocks: 2,
		votes:     []uint8{5, 5},
		curTip:    1,
		want:      []uint8{1, 0},
	}, {
		name:      "2 blocks, votes{0:4, 1:5}, tip block 0",
		numBlocks: 2,
		votes:     []uint8{4, 5},
		curTip:    0,
		want:      []uint8{1, 0},
	}, {
		name:      "2 blocks, votes{0:4, 1:5}, tip block 1",
		numBlocks: 2,
		votes:     []uint8{4, 5},
		curTip:    1,
		want:      []uint8{1, 0},
	}, {
		name:      "3 blocks, votes{0:4, 1:5, 2:2}, tip block 0 (2 not useful)",
		numBlocks: 3,
		votes:     []uint8{4, 5, 2},
		curTip:    0,
		want:      []uint8{1, 0},
	}, {
		name:      "3 blocks, votes{0:3, 1:5, 2:4}, tip block 0",
		numBlocks: 3,
		votes:     []uint8{3, 5, 4},
		curTip:    0,
		want:      []uint8{1, 2, 0},
	}, {
		name:      "3 blocks, votes{0:3, 1:5, 2:4}, tip block 2",
		numBlocks: 3,
		votes:     []uint8{3, 5, 4},
		curTip:    2,
		want:      []uint8{1, 2, 0},
	}, {
		name:      "3 blocks, votes{0:3, 1:3, 2:4}, tip block 1",
		numBlocks: 3,
		votes:     []uint8{3, 3, 4},
		curTip:    1,
		want:      []uint8{2, 0, 1},
	}}

	// nextHash is a convenience function that generates a unique hash each time
	// it is called.
	var genHashIndex uint32
	nextHash := func() chainhash.Hash {
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], genHashIndex)
		genHashIndex++

		return chainhash.HashH(buf[:])
	}

	for _, test := range tests {
		// Generate mock blocks and votes based on test data.
		blockHashes := make([]chainhash.Hash, test.numBlocks)
		for i := uint8(0); i < test.numBlocks; i++ {
			blockHashes[i] = nextHash()
		}
		votes := make(map[chainhash.Hash][]VoteDesc, len(test.votes))
		for i, numVotes := range test.votes {
			blockVotedOn := blockHashes[i]
			voteDescs := make([]VoteDesc, 0, numVotes)
			for i := uint8(0); i < numVotes; i++ {
				voteDescs = append(voteDescs, VoteDesc{
					VoteHash:       nextHash(),
					TicketHash:     nextHash(),
					ApprovesParent: true,
				})
			}
			votes[blockVotedOn] = voteDescs
		}
		var curTip chainhash.Hash
		if len(blockHashes) == 0 {
			curTip = nextHash()
		} else {
			curTip = blockHashes[test.curTip]
		}
		var want []chainhash.Hash
		if len(test.want) > 0 {
			want = make([]chainhash.Hash, 0, len(test.want))
			for _, blockIndex := range test.want {
				want = append(want, blockHashes[blockIndex])
			}
		}

		// Create a fake transaction source instance populated with the votes.
		txSource := &fakeTxSource{
			votes: votes,
		}

		// Ensure the result matches the expected values.
		got := SortParentsByVotes(txSource, curTip, blockHashes, params)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%q: unexpected result -- got %v, want %v", test.name, got,
				want)
			continue
		}
	}
}

func TestRejectedSpendTemplates(t *testing.T) {
	op := wire.OutPoint{Index: 1, Tree: wire.TxTreeRegular}
	stakeOp := op
	stakeOp.Tree = wire.TxTreeStake
	tx := wire.NewMsgTx()
	tx.AddTxIn(wire.NewTxIn(&stakeOp, 1, nil))
	tx.AddTxIn(wire.NewTxIn(&op, 1, nil))
	called := false
	g := &BlkTmplGenerator{cfg: &Config{
		Policy:                    &Policy{RejectedSpends: RejectedSpends{op: {}}},
		CheckConnectBlockTemplate: func(*dcrutil.Block) error { called = true; return nil },
	}}
	if g.maybeInsertStakeTx(dcrutil.NewTx(tx), true, false) {
		t.Fatal("stake transaction with excluded input was selected")
	}
	for _, b := range []*wire.MsgBlock{{Transactions: []*wire.MsgTx{tx}}, {STransactions: []*wire.MsgTx{tx}}} {
		if err := g.checkConnectBlockTemplate(dcrutil.NewBlock(b)); err == nil || called {
			t.Fatal("template with blocked input reached consensus validation")
		}
	}
	tx.TxIn = tx.TxIn[:1]
	if err := g.checkConnectBlockTemplate(dcrutil.NewBlock(&wire.MsgBlock{STransactions: []*wire.MsgTx{tx}})); err != nil || !called {
		t.Fatalf("unrelated outpoint should pass policy: %v", err)
	}
}

func TestNewBlockTemplateWithRetirementTxs(t *testing.T) {
	t.Parallel()

	harness, spendableOuts, err := newMiningHarness(chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("error creating mining harness: %v", err)
	}
	// Enable header commitments so calcBlockCommitmentRootV1 exercises blockcf2
	// with the injected retirement transaction inputs in blockUtxos.
	harness.chain.isHeaderCommitmentsAgendaActive = true

	address, err := stdaddr.DecodeAddress("TsebVL4jcgJ2uVqZS2HutWTb4KnJLc9FdLR",
		harness.chainParams)
	if err != nil {
		t.Fatalf("error decoding address: %v", err)
	}

	// Create 2 spendable outputs: vout 0 will be retired, vout 1 is a normal mempool spend.
	baseTx, err := harness.CreateSignedTx(spendableOuts, 2)
	if err != nil {
		t.Fatalf("CreateSignedTx: %v", err)
	}
	baseHeight := harness.chain.bestState.Height
	harness.AddFakeUTXO(baseTx, baseHeight, 1, harness.chain.isTreasuryAgendaActive)

	retireOp := wire.OutPoint{Hash: *baseTx.Hash(), Index: 0, Tree: wire.TxTreeRegular}
	retireAmt := baseTx.MsgTx().TxOut[0].Value

	// Build a canonical zero-fee OP_RETURN burn tx spending retireOp with empty sigScript.
	burnMsg := wire.NewMsgTx()
	burnMsg.SerType = wire.TxSerializeFull
	burnMsg.Version = 1
	burnIn := wire.NewTxIn(&retireOp, retireAmt, nil)
	burnIn.BlockHeight = uint32(baseHeight)
	burnIn.BlockIndex = 1
	burnMsg.AddTxIn(burnIn)
	opRetScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).
		AddData([]byte("TEST-MINING-RETIRE")).
		Script()
	if err != nil {
		t.Fatal(err)
	}
	burnMsg.AddTxOut(wire.NewTxOut(retireAmt, opRetScript))
	burnTx := dcrutil.NewTx(burnMsg)
	burnTx.SetTree(wire.TxTreeRegular)

	// Put a normal mempool tx spending vout 1 AND a conflicting mempool tx spending retireOp.
	applyTxFee := func(fee int64) func(*wire.MsgTx) {
		return func(tx *wire.MsgTx) {
			tx.TxOut[0].Value -= fee
		}
	}
	normalTx, err := harness.CreateSignedTx([]spendableOutput{
		txOutToSpendableOut(baseTx, 1, wire.TxTreeRegular),
	}, 1, applyTxFee(5000))
	if err != nil {
		t.Fatalf("CreateSignedTx(normal): %v", err)
	}
	conflictTx, err := harness.CreateSignedTx([]spendableOutput{
		txOutToSpendableOut(baseTx, 0, wire.TxTreeRegular),
	}, 1, applyTxFee(5000))
	if err != nil {
		t.Fatalf("CreateSignedTx(conflict): %v", err)
	}
	if _, err := harness.AddTransactionToTxSource(normalTx); err != nil {
		t.Fatalf("AddTransactionToTxSource(normal): %v", err)
	}
	if _, err := harness.AddTransactionToTxSource(conflictTx); err != nil {
		t.Fatalf("AddTransactionToTxSource(conflict): %v", err)
	}

	nextHeight := harness.chain.bestState.Height + 1
	// Block retireOp via local RejectedSpends policy (matching mainnet quarantine policy).
	harness.policy.RejectedSpends = RejectedSpends{retireOp: {}}
	harness.generator.cfg.RequiredRetirementTxs = func(h int64, includeRegular bool) ([]*dcrutil.Tx, error) {
		if h == nextHeight {
			cp := dcrutil.NewTx(burnTx.MsgTx().Copy())
			cp.SetTree(wire.TxTreeRegular)
			return []*dcrutil.Tx{cp}, nil
		}
		return nil, nil
	}
	harness.generator.cfg.IsCanonicalRetirementTx = func(h int64, tx *dcrutil.Tx) bool {
		return h == nextHeight && tx != nil && tx.MsgTx().TxHashFull() == burnTx.MsgTx().TxHashFull()
	}

	tmpl, err := harness.generator.NewBlockTemplate(address)
	if err != nil {
		t.Fatalf("NewBlockTemplate failed with retirement tx + RejectedSpends: %v", err)
	}

	// Expected regular tree: [coinbase, burnTx, normalTx]. conflictTx must be excluded.
	if got := len(tmpl.Block.Transactions); got != 3 {
		t.Fatalf("got %d regular txs, want 3", got)
	}
	if got := tmpl.Block.Transactions[1].TxHashFull(); got != burnTx.MsgTx().TxHashFull() {
		t.Fatalf("Transactions[1] = %s, want canonical burn tx %s", got, burnTx.MsgTx().TxHashFull())
	}
	if got := tmpl.Block.Transactions[2].TxHash(); got != *normalTx.Hash() {
		t.Fatalf("Transactions[2] = %s, want normal tx %s", got, normalTx.Hash())
	}

	solveTemplate(t, tmpl.Block, harness.chainParams)
	if err := blockchain.CheckBlockSanity(dcrutil.NewBlock(tmpl.Block),
		harness.generator.cfg.TimeSource, harness.chainParams); err != nil {
		t.Fatalf("CheckBlockSanity failed: %v", err)
	}

	// Ensure a canonical burn tx placed in the stake tree is STILL blocked by RejectedSpends.
	stakeBad := &wire.MsgBlock{
		Header:        wire.BlockHeader{Height: uint32(nextHeight)},
		STransactions: []*wire.MsgTx{burnTx.MsgTx().Copy()},
	}
	if err := harness.generator.checkConnectBlockTemplate(dcrutil.NewBlock(stakeBad)); err == nil {
		t.Fatal("canonical retirement tx in stake tree unexpectedly bypassed RejectedSpends")
	}

	// Ensure a non-canonical regular tx spending retireOp is STILL blocked by RejectedSpends.
	regBad := &wire.MsgBlock{
		Header:       wire.BlockHeader{Height: uint32(nextHeight)},
		Transactions: []*wire.MsgTx{conflictTx.MsgTx().Copy()},
	}
	if err := harness.generator.checkConnectBlockTemplate(dcrutil.NewBlock(regBad)); err == nil {
		t.Fatal("non-canonical regular tx spending rejected outpoint unexpectedly passed")
	}
}
