// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"github.com/EXCCoin/exccd/blockchain/chaingen"
	"github.com/EXCCoin/exccd/cequihash"
	"github.com/EXCCoin/exccd/chaincfg"
	"github.com/EXCCoin/exccd/exccutil"
	"github.com/EXCCoin/exccd/wire"
	"unsafe"
)

// cloneParams returns a deep copy of the provided parameters so the caller is
// free to modify them without worrying about interfering with other tests.
func cloneParams(params *chaincfg.Params) *chaincfg.Params {
	// Encode via gob.
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	enc.Encode(params)

	// Decode via gob to make a deep copy.
	var paramsCopy chaincfg.Params
	dec := gob.NewDecoder(buf)
	dec.Decode(&paramsCopy)
	return &paramsCopy
}

//Temporary function for block data conversion

type JSONBlock struct {
	MsgBlock wire.MsgBlock
}

type solutionValidatorData struct {
	params *chaincfg.Params
	solved *bool
	header *wire.BlockHeader
}

func (data solutionValidatorData) Validate(solution unsafe.Pointer) int {
	solutionBytes := cequihash.ExtractSolution(data.params.N, data.params.K, solution)

	copy(data.header.EquihashSolution[:], solutionBytes)

	hash := data.header.BlockHash()

	if HashToBig(&hash).Cmp(CompactToBig(data.header.Bits)) <= 0 {
		*data.solved = true
		return 1
	} else {
		return 0
	}
}

func appendExtraNonce(headerData []byte, header *wire.BlockHeader) []byte {
	result := make([]byte, len(headerData)+32)
	copy(result, headerData)
	result = append(result, header.ExtraData[:]...)

	return result
}

const (
	// maxNonce is the maximum value a nonce can be in a block header.
	maxNonce = ^uint32(0) // 2^32 - 1

	// maxExtraNonce is the maximum value an extra nonce used in a coinbase
	// transaction can be.
	maxExtraNonce = ^uint64(0) // 2^64 - 1
)

func solve(t *testing.T, header *wire.BlockHeader, chainParams *chaincfg.Params) error {
	enOffset, err := wire.RandomUint64()
	if err != nil {
		t.Logf("Error generating extended nonce offset. Using default (0)")
		enOffset = 0
	}

	if err != nil {
		return err
	}

	solved := false
	validator := solutionValidatorData{chainParams, &solved, header}

	for extraNonce := uint64(0); extraNonce < maxExtraNonce && !solved; extraNonce++ {
		// Update the extra nonce in the block template header with the
		// new value.
		binary.LittleEndian.PutUint64(header.ExtraData[:], extraNonce+enOffset)

		// Update equihash solver input bytes
		headerBytes, _ := header.SerializeAllHeaderBytes()

		// Search through the entire nonce range for a solution while
		// periodically checking for early quit and stale block
		// conditions along with updates to the speed monitor.
		for i := uint32(0); i <= maxNonce && !solved; i++ {
			t.Logf("Trying nonce %d", i)

			header.Nonce = i

			cequihash.SolveEquihash(chainParams.N, chainParams.K, headerBytes, int64(i), validator)
		}
	}

	return nil
}

// Note: this test usually should be skipped. It is used only to convert and re-generate
// valid test data in cases of block header format or other relevant changes (for example, hash function)
func TestConvertToNewFormat(t *testing.T) {

	t.SkipNow()

	// Load up the rest of the blocks up to HEAD~1.
	filename := filepath.Join("testdata/", "blocks0to168.json.gz")
	fi, err := os.Open(filename)
	if err != nil {
		t.Errorf("Unable to open %s: %v", filename, err)
	}

	ofilename := filepath.Join("testdata/", "blocks0to168.exccd.json.gz")
	fo, err := os.Create(ofilename)
	if err != nil {
		t.Errorf("Unable to open %s: %v", filename, err)
	}

	jsonStream, err := gzip.NewReader(fi)

	if err != nil {
		t.Fatalf("Unable to open input file %s (%v)", ofilename, err)
	}

	jsonOutStream := gzip.NewWriter(fo)

	defer jsonStream.Close()
	defer fi.Close()
	defer jsonOutStream.Close()

	decoder := json.NewDecoder(jsonStream)
	encoder := json.NewEncoder(jsonOutStream)

	counter := 0

	params := cloneParams(&chaincfg.SimNetParams)
	params.GenesisBlock.Header.MerkleRoot = *mustParseHash("a216ea043f0d481a072424af646787794c32bcefd3ed181a090319bbf8a37105")
	genesisHash := params.GenesisBlock.BlockHash()
	params.GenesisHash = &genesisHash
	hash := &genesisHash

	for decoder.More() {
		var bl JSONBlock

		counter++

		err := decoder.Decode(&bl)

		if err != nil {
			t.Fatalf("Unable to decode block (%d) %v", counter, err)
		}

		// Update block header with missing or incorrect values
		bl.MsgBlock.Header.Size = uint32(bl.MsgBlock.SerializeSize())
		bl.MsgBlock.Header.PrevBlock.SetBytes(hash.CloneBytes())
		merkles := BuildMerkleTreeStore(exccutil.NewBlock(&bl.MsgBlock).Transactions())
		bl.MsgBlock.Header.MerkleRoot = *merkles[len(merkles)-1]

		t.Logf("Solving block %d...", counter)

		// Solve block and store valid equihash solution
		err = solve(t, &bl.MsgBlock.Header, params)

		if err != nil {
			t.Fatalf("Unable to find solution for block (%d) %v", counter, err)
		}

		err = validateEquihashSolution(&bl.MsgBlock.Header, params)

		if err != nil {
			t.Logf("...not solved\n")

		} else {
			t.Logf("...solved\n")
		}

		err = encoder.Encode(bl)

		if err != nil {
			t.Fatalf("Unable to encode block %v", err)
		}

		prevHash := bl.MsgBlock.BlockHash()
		hash = &prevHash
	}
	t.Logf("Total number of records: %d\n", counter)
}

// TestBlockchainFunction tests the various blockchain API to ensure proper
// functionality.
// TODO: once upon a time enable test and make it pass
func TestBlockchainFunctions(t *testing.T) {
	// TODO: At present test does not pass because of some discrepancies in TxOut of block #2. Need to fix this.
	t.SkipNow()

	// Update simnet parameters to reflect what is expected by the legacy
	// data.
	params := cloneParams(&chaincfg.SimNetParams)
	params.GenesisBlock.Header.MerkleRoot = *mustParseHash("a216ea043f0d481a072424af646787794c32bcefd3ed181a090319bbf8a37105")
	genesisHash := params.GenesisBlock.BlockHash()
	params.GenesisHash = &genesisHash

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("validateunittests", params)
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	defer teardownFunc()

	// Load up the rest of the blocks up to HEAD~1.
	filename := filepath.Join("testdata/", "blocks0to168.exccd.json.gz")
	fi, err := os.Open(filename)
	if err != nil {
		t.Errorf("Unable to open %s: %v", filename, err)
	}
	bcStream, err := gzip.NewReader(fi)
	if err != nil {
		t.Errorf("Unable to read archive %s: %v", filename, err)
	}

	defer bcStream.Close()
	defer fi.Close()

	// Create decoder from the buffer and a map to store the data.
	decoder := json.NewDecoder(bcStream)

	// Insert blocks 1 to 168 and perform various tests.
	for i := 1; i <= 168; i++ {
		var jsbl JSONBlock
		err := decoder.Decode(&jsbl)

		if err != nil {
			t.Fatalf("Unable to decode block (%d) %v", i, err)
		}

		bl := exccutil.NewBlock(&jsbl.MsgBlock)
		_, _, err = chain.ProcessBlock(bl, BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock error at height %v: %v", i, err.Error())
		}
	}

	val, err := chain.TicketPoolValue()
	if err != nil {
		t.Errorf("Failed to get ticket pool value: %v", err)
	}
	expectedVal := exccutil.Amount(3495091704)
	if val != expectedVal {
		t.Errorf("Failed to get correct result for ticket pool value; "+
			"want %v, got %v", expectedVal, val)
	}

	a, _ := exccutil.DecodeAddress("SsbKpMkPnadDcZFFZqRPY8nvdFagrktKuzB")
	hs, err := chain.TicketsWithAddress(a)
	if err != nil {
		t.Errorf("Failed to do TicketsWithAddress: %v", err)
	}
	expectedLen := 223
	if len(hs) != expectedLen {
		t.Errorf("Failed to get correct number of tickets for "+
			"TicketsWithAddress; want %v, got %v", expectedLen, len(hs))
	}

	totalSubsidy := chain.TotalSubsidy()
	expectedSubsidy := int64(35783267326630)
	if expectedSubsidy != totalSubsidy {
		t.Errorf("Failed to get correct total subsidy for "+
			"TotalSubsidy; want %v, got %v", expectedSubsidy,
			totalSubsidy)
	}
}

// TestForceHeadReorg ensures forcing header reorganization works as expected.
// TODO: once upon a time enable test and make it pass
func TestForceHeadReorg(t *testing.T) {
	// Test data contain invalid equihash solution
	t.SkipNow()

	// Create a test generator instance initialized with the genesis block
	// as the tip as well as some cached payment scripts to be used
	// throughout the tests.
	params := &chaincfg.SimNetParams
	g, err := chaingen.MakeGenerator(params)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("forceheadreorgtest", params)
	if err != nil {
		t.Fatalf("Failed to setup chain instance: %v", err)
	}
	defer teardownFunc()

	// Define some convenience helper functions to process the current tip
	// block associated with the generator.
	//
	// accepted expects the block to be accepted to the main chain.
	//
	// expectTip expects the provided block to be the current tip of the
	// main chain.
	//
	// acceptedToSideChainWithExpectedTip expects the block to be accepted
	// to a side chain, but the current best chain tip to be the provided
	// value.
	//
	// forceTipReorg forces the chain instance to reorganize the current tip
	// of the main chain from the given block to the given block.  An error
	// will result if the provided from block is not actually the current
	// tip.
	accepted := func() {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := exccutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)",
			g.TipName(), block.Hash(), blockHeight)

		isMainChain, isOrphan, err := chain.ProcessBlock(block, BFNone)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) should "+
				"have been accepted: %v", g.TipName(),
				block.Hash(), blockHeight, err)
		}

		// Ensure the main chain and orphan flags match the values
		// specified in the test.
		if !isMainChain {
			t.Fatalf("block %q (hash %s, height %d) unexpected main "+
				"chain flag -- got %v, want true", g.TipName(),
				block.Hash(), blockHeight, isMainChain)
		}
		if isOrphan {
			t.Fatalf("block %q (hash %s, height %d) unexpected "+
				"orphan flag -- got %v, want false", g.TipName(),
				block.Hash(), blockHeight, isOrphan)
		}
	}
	expectTip := func(tipName string) {
		// Ensure hash and height match.
		wantTip := g.BlockByName(tipName)
		best := chain.BestSnapshot()
		if best.Hash != wantTip.BlockHash() ||
			best.Height != int64(wantTip.Header.Height) {
			t.Fatalf("block %q (hash %s, height %d) should be "+
				"the current tip -- got (hash %s, height %d)",
				tipName, wantTip.BlockHash(),
				wantTip.Header.Height, best.Hash, best.Height)
		}
	}
	acceptedToSideChainWithExpectedTip := func(tipName string) {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := exccutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)",
			g.TipName(), block.Hash(), blockHeight)

		isMainChain, isOrphan, err := chain.ProcessBlock(block, BFNone)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) should "+
				"have been accepted: %v", g.TipName(),
				block.Hash(), blockHeight, err)
		}

		// Ensure the main chain and orphan flags match the values
		// specified in the test.
		if isMainChain {
			t.Fatalf("block %q (hash %s, height %d) unexpected main "+
				"chain flag -- got %v, want false", g.TipName(),
				block.Hash(), blockHeight, isMainChain)
		}
		if isOrphan {
			t.Fatalf("block %q (hash %s, height %d) unexpected "+
				"orphan flag -- got %v, want false", g.TipName(),
				block.Hash(), blockHeight, isOrphan)
		}

		expectTip(tipName)
	}
	forceTipReorg := func(fromTipName, toTipName string) {
		from := g.BlockByName(fromTipName)
		to := g.BlockByName(toTipName)
		err = chain.ForceHeadReorganization(from.BlockHash(), to.BlockHash())
		if err != nil {
			t.Fatalf("failed to force header reorg from block %q "+
				"(hash %s, height %d) to block %q (hash %s, "+
				"height %d): %v", fromTipName, from.BlockHash(),
				from.Header.Height, toTipName, to.BlockHash(),
				to.Header.Height, err)
		}
	}

	// Shorter versions of useful params for convenience.
	ticketsPerBlock := params.TicketsPerBlock
	coinbaseMaturity := params.CoinbaseMaturity
	stakeEnabledHeight := params.StakeEnabledHeight
	stakeValidationHeight := params.StakeValidationHeight

	// ---------------------------------------------------------------------
	// Premine.
	// ---------------------------------------------------------------------

	// Add the required premine block.
	//
	//   genesis -> bp
	g.CreatePremineBlock("bp", 0)
	g.AssertTipHeight(1)
	accepted()

	// ---------------------------------------------------------------------
	// Generate enough blocks to have mature coinbase outputs to work with.
	//
	//   genesis -> bp -> bm0 -> bm1 -> ... -> bm#
	// ---------------------------------------------------------------------

	for i := uint16(0); i < coinbaseMaturity; i++ {
		blockName := fmt.Sprintf("bm%d", i)
		g.NextBlock(blockName, nil, nil)
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(coinbaseMaturity) + 1)

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the stake enabled height while
	// creating ticket purchases that spend from the coinbases matured
	// above.  This will also populate the pool of immature tickets.
	//
	//   ... -> bm# ... -> bse0 -> bse1 -> ... -> bse#
	// ---------------------------------------------------------------------

	var ticketsPurchased int
	for i := int64(0); int64(g.Tip().Header.Height) < stakeEnabledHeight; i++ {
		outs := g.OldestCoinbaseOuts()
		ticketOuts := outs[1:]
		ticketsPurchased += len(ticketOuts)
		blockName := fmt.Sprintf("bse%d", i)
		g.NextBlock(blockName, nil, ticketOuts)
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeEnabledHeight))

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the stake validation height while
	// continuing to purchase tickets using the coinbases matured above and
	// allowing the immature tickets to mature and thus become live.
	// ---------------------------------------------------------------------

	targetPoolSize := g.Params().TicketPoolSize * ticketsPerBlock
	for i := int64(0); int64(g.Tip().Header.Height) < stakeValidationHeight; i++ {
		// Only purchase tickets until the target ticket pool size is
		// reached.
		outs := g.OldestCoinbaseOuts()
		ticketOuts := outs[1:]
		if ticketsPurchased+len(ticketOuts) > int(targetPoolSize) {
			ticketsNeeded := int(targetPoolSize) - ticketsPurchased
			if ticketsNeeded > 0 {
				ticketOuts = ticketOuts[1 : ticketsNeeded+1]
			} else {
				ticketOuts = nil
			}
		}
		ticketsPurchased += len(ticketOuts)

		blockName := fmt.Sprintf("bsv%d", i)
		g.NextBlock(blockName, nil, ticketOuts)
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight))

	// ---------------------------------------------------------------------
	// Generate enough blocks to have a known distance to the first mature
	// coinbase outputs for all tests that follow.  These blocks continue
	// to purchase tickets to avoid running out of votes.
	//
	//   ... -> bsv# -> bbm0 -> bbm1 -> ... -> bbm#
	// ---------------------------------------------------------------------

	for i := uint16(0); i < coinbaseMaturity; i++ {
		outs := g.OldestCoinbaseOuts()
		blockName := fmt.Sprintf("bbm%d", i)
		g.NextBlock(blockName, nil, outs[1:])
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight) + uint32(coinbaseMaturity))

	// Collect spendable outputs into two different slices.  The outs slice
	// is intended to be used for regular transactions that spend from the
	// output, while the ticketOuts slice is intended to be used for stake
	// ticket purchases.
	var outs []*chaingen.SpendableOut
	var ticketOuts [][]chaingen.SpendableOut
	for i := uint16(0); i < coinbaseMaturity; i++ {
		coinbaseOuts := g.OldestCoinbaseOuts()
		outs = append(outs, &coinbaseOuts[0])
		ticketOuts = append(ticketOuts, coinbaseOuts[1:])
	}

	// ---------------------------------------------------------------------
	// Forced header reorganization test.
	// ---------------------------------------------------------------------

	// Start by building a couple of blocks at current tip (value in parens
	// is which output is spent):
	//
	//   ... -> b1(0) -> b2(1)
	g.NextBlock("b1", outs[0], ticketOuts[0])
	accepted()

	g.NextBlock("b2", outs[1], ticketOuts[1])
	accepted()

	// Create some forks from b1.  There should not be a reorg since b2 was
	// seen first.
	//
	//   ... -> b1(0) -> b2(1)
	//               \-> b3(1)
	//               \-> b4(1)
	//               \-> b5(1)
	g.SetTip("b1")
	g.NextBlock("b3", outs[1], ticketOuts[1])
	acceptedToSideChainWithExpectedTip("b2")

	g.SetTip("b1")
	g.NextBlock("b4", outs[1], ticketOuts[1])
	acceptedToSideChainWithExpectedTip("b2")

	g.SetTip("b1")
	g.NextBlock("b5", outs[1], ticketOuts[1])
	acceptedToSideChainWithExpectedTip("b2")

	// Force tip reorganization to b3.
	//
	//   ... -> b1(0) -> b3(1)
	//               \-> b2(1)
	//               \-> b4(1)
	//               \-> b5(1)
	forceTipReorg("b2", "b3")
	expectTip("b3")

	// Force tip reorganization to b4.
	//
	//   ... -> b1(0) -> b4(1)
	//               \-> b2(1)
	//               \-> b3(1)
	//               \-> b5(1)
	forceTipReorg("b3", "b4")
	expectTip("b4")

	// Force tip reorganization to b5.
	//
	//   ... -> b1(0) -> b5(1)
	//               \-> b2(1)
	//               \-> b3(1)
	//               \-> b4(1)
	forceTipReorg("b4", "b5")
	expectTip("b5")
}
