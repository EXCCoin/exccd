// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2020 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package standalone_test

import (
	"fmt"
	"math/big"

	"github.com/EXCCoin/exccd/blockchain/standalone/v2"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
)

// This example demonstrates how to convert the compact "bits" in a block header
// which represent the target difficulty to a big integer and display it using
// the typical hex notation.
func ExampleCompactToBig() {
	// Convert the bits from block 1 in the main chain.
	bits := uint32(453115903)
	targetDifficulty := standalone.CompactToBig(bits)

	// Display it in hex.
	fmt.Printf("%064x\n", targetDifficulty.Bytes())

	// Output:
	// 000000000001ffff000000000000000000000000000000000000000000000000
}

// This example demonstrates how to convert a target difficulty into the compact
// "bits" in a block header which represent that target difficulty.
func ExampleBigToCompact() {
	// Convert the target difficulty from block 1 in the main chain to compact
	// form.
	t := "000000000001ffff000000000000000000000000000000000000000000000000"
	targetDifficulty, success := new(big.Int).SetString(t, 16)
	if !success {
		fmt.Println("invalid target difficulty")
		return
	}
	bits := standalone.BigToCompact(targetDifficulty)

	fmt.Println(bits)

	// Output:
	// 453115903
}

// This example demonstrates calculating a merkle root from a slice of leaf
// hashes.
func ExampleCalcMerkleRoot() {
	// Create a slice of the leaf hashes.
	leaves := make([]chainhash.Hash, 3)
	for i := range leaves {
		// The hash would ordinarily be calculated from the TxHashFull function
		// on a transaction, however, it's left as a zero hash for the purposes
		// of this example.
		leaves[i] = chainhash.Hash{}
	}

	merkleRoot := standalone.CalcMerkleRoot(leaves)
	fmt.Printf("Result: %s", merkleRoot)

	// Output:
	// Result: 715d8b746a60de6ca9d0ebb1ecaa8992a8c95af32b895cf8c1d4fd004e1156db
}
