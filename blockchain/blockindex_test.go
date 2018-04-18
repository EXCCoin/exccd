// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"reflect"
	"testing"
	"time"

	"github.com/EXCCoin/exccd/chaincfg"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/wire"
)

// mustParseHash converts the passed big-endian hex string into a
// chainhash.Hash and will panic if there is an error.  It only differs from the
// one available in chainhash in that it will panic so errors in the source code
// be detected.  It will only (and must only) be called with hard-coded, and
// therefore known good, hashes.
func mustParseHash(s string) *chainhash.Hash {
	hash, err := chainhash.NewHashFromStr(s)
	if err != nil {
		panic("invalid hash in source file: " + s)
	}
	return hash
}

// TestBlockNodeHeader ensures that block nodes reconstruct the correct header
// and fetching the header from the chain reconstructs it from memory.
func TestBlockNodeHeader(t *testing.T) {
	// Create a fake chain and block header with all fields set to nondefault
	// values.
	params := &chaincfg.SimNetParams
	bc := newFakeChain(params)
	testHeader := wire.BlockHeader{
		Version:      1,
		PrevBlock:    bc.bestNode.hash,
		MerkleRoot:   *mustParseHash("09876543210987654321"),
		StakeRoot:    *mustParseHash("43210987654321098765"),
		VoteBits:     0x03,
		FinalState:   [6]byte{0xaa},
		Voters:       4,
		FreshStake:   5,
		Revocations:  6,
		PoolSize:     20000,
		Bits:         0x1234,
		SBits:        123456789,
		Height:       1,
		Size:         393216,
		Timestamp:    time.Unix(1454954400, 0),
		Nonce:        7,
		ExtraData:    [32]byte{0xbb},
		StakeVersion: 5,
	}
	node := newBlockNode(&testHeader, bc.bestNode)
	bc.index.AddNode(node)

	// Ensure reconstructing the header for the node produces the same header
	// used to create the node.
	gotHeader := node.Header()
	if !reflect.DeepEqual(gotHeader, testHeader) {
		t.Fatalf("node.Header: mismatched headers: got %+v, want %+v",
			gotHeader, testHeader)
	}

	// Ensure fetching the header from the chain produces the same header used
	// to create the node.
	testHeaderHash := testHeader.BlockHash()
	gotHeader, err := bc.FetchHeader(&testHeaderHash)
	if err != nil {
		t.Fatalf("FetchHeader: unexpected error: %v", err)
	}
	if !reflect.DeepEqual(gotHeader, testHeader) {
		t.Fatalf("FetchHeader: mismatched headers: got %+v, want %+v",
			gotHeader, testHeader)
	}
}

// TestCalcPastMedianTime ensures the CalcPastMedianTie function works as
// intended including when there are less than the typical number of blocks
// which happens near the beginning of the chain.
func TestCalcPastMedianTime(t *testing.T) {
	tests := []struct {
		name       string
		timestamps []int64
		expected   int64
	}{
		{
			name:       "one block",
			timestamps: []int64{1517188771},
			expected:   1517188771,
		},
		{
			name:       "two blocks, in order",
			timestamps: []int64{1517188771, 1517188831},
			expected:   1517188771,
		},
		{
			name:       "three blocks, in order",
			timestamps: []int64{1517188771, 1517188831, 1517188891},
			expected:   1517188831,
		},
		{
			name:       "three blocks, out of order",
			timestamps: []int64{1517188771, 1517188891, 1517188831},
			expected:   1517188831,
		},
		{
			name:       "four blocks, in order",
			timestamps: []int64{1517188771, 1517188831, 1517188891, 1517188951},
			expected:   1517188831,
		},
		{
			name:       "four blocks, out of order",
			timestamps: []int64{1517188831, 1517188771, 1517188951, 1517188891},
			expected:   1517188831,
		},
		{
			name: "eleven blocks, in order",
			timestamps: []int64{1517188771, 1517188831, 1517188891, 1517188951,
				1517189011, 1517189071, 1517189131, 1517189191, 1517189251,
				1517189311, 1517189371},
			expected: 1517189071,
		},
		{
			name: "eleven blocks, out of order",
			timestamps: []int64{1517188831, 1517188771, 1517188891, 1517189011,
				1517188951, 1517189071, 1517189131, 1517189191, 1517189251,
				1517189371, 1517189311},
			expected: 1517189071,
		},
		{
			name: "fifteen blocks, in order",
			timestamps: []int64{1517188771, 1517188831, 1517188891, 1517188951,
				1517189011, 1517189071, 1517189131, 1517189191, 1517189251,
				1517189311, 1517189371, 1517189431, 1517189491, 1517189551,
				1517189611},
			expected: 1517189311,
		},
		{
			name: "fifteen blocks, out of order",
			timestamps: []int64{1517188771, 1517188891, 1517188831, 1517189011,
				1517188951, 1517189131, 1517189071, 1517189251, 1517189191,
				1517189371, 1517189311, 1517189491, 1517189431, 1517189611,
				1517189551},
			expected: 1517189311,
		},
	}

	params := &chaincfg.SimNetParams
	for _, test := range tests {
		// Create a synthetic chain with the correct number of nodes and the
		// timestamps as specified by the test.
		bc := newFakeChain(params)
		node := bc.bestNode
		for _, timestamp := range test.timestamps {
			node = newFakeNode(node, 0, 0, 0, time.Unix(timestamp, 0))
			bc.index.AddNode(node)
			bc.bestNode = node
		}

		// Ensure the median time is the expected value.
		gotTime, err := bc.index.CalcPastMedianTime(node)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
			continue
		}
		wantTime := time.Unix(test.expected, 0)
		if !gotTime.Equal(wantTime) {
			t.Errorf("%s: mismatched timestamps -- got: %v, want: %v",
				test.name, gotTime, wantTime)
			continue
		}
	}
}

// TestChainTips ensures the chain tip tracking in the block index works
// as expected.
func TestChainTips(t *testing.T) {
	params := &chaincfg.SimNetParams
	bc := newFakeChain(params)
	genesis := bc.bestNode

	// Construct a synthetic simnet chain consisting of the following structure.
	// 0 -> 1 -> 2  -> 3  -> 4
	//  |    \-> 2a -> 3a -> 4a -> 5a -> 6a -> 7a -> ... -> 26a
	//  |    |     \-> 3b -> 4b -> 5b
	//  |    \-> 2c -> 3c -> 4c -> 5c -> 6c -> 7c -> ... -> 26c
	//  \-> 1d
	//  \-> 1e
	branches := make([][]*blockNode, 6)
	branches[0] = chainedFakeNodes(genesis, 4)
	branches[1] = chainedFakeNodes(branches[0][0], 25)
	branches[2] = chainedFakeNodes(branches[1][0], 3)
	branches[3] = chainedFakeNodes(branches[0][0], 25)
	branches[4] = chainedFakeNodes(genesis, 1)
	branches[5] = chainedFakeNodes(genesis, 1)

	// Add all of the nodes to the index.
	for _, branch := range branches {
		for _, node := range branch {
			bc.index.AddNode(node)
		}
	}

	// Create a map of all of the chain tips the block index believes exist.
	chainTips := make(map[*blockNode]struct{})
	bc.index.RLock()
	for _, nodes := range bc.index.chainTips {
		for _, node := range nodes {
			chainTips[node] = struct{}{}
		}
	}
	bc.index.RUnlock()

	// The expected chain tips are the tips of all of the branches.
	tip := func(nodes []*blockNode) *blockNode {
		return nodes[len(nodes)-1]
	}
	expectedTips := make(map[*blockNode]struct{})
	for _, branch := range branches {
		expectedTips[tip(branch)] = struct{}{}
	}

	// Ensure the chain tips are the expected values.
	if len(chainTips) != len(expectedTips) {
		t.Fatalf("block index reports %d chain tips, but %d were expected",
			len(chainTips), len(expectedTips))
	}
	for node := range expectedTips {
		if _, ok := chainTips[node]; !ok {
			t.Fatalf("block index does not contain expected tip %s (height %d)",
				node.hash, node.height)
		}
	}
}
