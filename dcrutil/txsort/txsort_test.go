// Copyright (c) 2015-2016 The btcsuite developers
// Copyright (c) 2017-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txsort

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/EXCCoin/exccd/wire"
)

// TestSort ensures the transaction sorting works as expected.
func TestSort(t *testing.T) {
	tests := []struct {
		name         string
		hexFile      string
		isSorted     bool
		unsortedHash string
		sortedHash   string
	}{
		{
			name:         "block 100004 tx[4] - already sorted",
			hexFile:      "tx100004-4.hex",
			isSorted:     true,
			unsortedHash: "ff77ef63c87653b896044dd847a926c4a9d25593a3cb05ba94be3b69bfc817b3",
			sortedHash:   "ff77ef63c87653b896044dd847a926c4a9d25593a3cb05ba94be3b69bfc817b3",
		},
		{
			name:         "block 101790 tx[3] - sorts inputs only, based on tree",
			hexFile:      "tx101790-3.hex",
			isSorted:     false,
			unsortedHash: "0db834ff9bcf2ec45cf745f9c9491ab0f405ab7237e3ee801af8a31b6eb7cc9f",
			sortedHash:   "97fbdff94b5da583ba4c3fe85dff618f6006261b679aab956d5383f33955ba60",
		},
		{
			name:         "block 150007 tx[23] - sorts inputs only, based on hash",
			hexFile:      "tx150007-23.hex",
			isSorted:     false,
			unsortedHash: "23f9c58732117638ca2d60b1ec05868edfc3a99c3059a4c0a118f28d4c8c61f4",
			sortedHash:   "154036c36bc87e84396d70dd482686c8b880393fbb1d2f69279d67fcefa3b7c2",
		},
		{
			name:         "block 108930 tx[1] - sorts inputs only, based on index",
			hexFile:      "tx108930-1.hex",
			isSorted:     false,
			unsortedHash: "69aa335f8a1fed1f397d995c70b51efe9c9c2af24c39535ea567dd4a21d777ba",
			sortedHash:   "128a1d6f061b1bcae63deb20e2935b72f902db7dd30e76ea4295789fff2a6c28",
		},
		{
			name:         "block 100082 tx[5] - sorts outputs only, based on amount",
			hexFile:      "tx100082-5.hex",
			isSorted:     false,
			unsortedHash: "4707764d63ad99dcfaebee91c1761a63ce17060f3e1b9b5826b6d0f5e80fffc0",
			sortedHash:   "421f1d0ba77c7408d4b63978b45fe9a92cfc0845b49f3d2b5a77cee41b7e8b0c",
		},
		{
			// Tx manually modified to make the first output (output 0)
			// have script version 1.
			name:         "modified block 150043 tx[14] - sorts outputs only, based on script version",
			hexFile:      "tx150043-14m.hex",
			isSorted:     false,
			unsortedHash: "6a4dab4afc3c8bb059b6f1414f058e8393f4a2ee4252c58baea530488436afb7",
			sortedHash:   "e55527a77e29ad40c12c605205e9f62945037fa74079b9f602f1213a5e726753",
		},
		{
			name:         "block 150043 tx[14] - sorts outputs only, based on output script",
			hexFile:      "tx150043-14.hex",
			isSorted:     false,
			unsortedHash: "73eaa4e89f0e2fc89bc6937e13660fc0d88badf9112c782133f2adc2d2be9e13",
			sortedHash:   "ae94bfc983440ccf34c45ac9785c149afbcfc089ac9b5db2d348c13a5d001002",
		},
		{
			name:         "block 150626 tx[24] - sorts outputs only, based on amount and output script",
			hexFile:      "tx150626-24.hex",
			isSorted:     false,
			unsortedHash: "79824fbfee8d7e1a5582c601695a089ecb11199a5b006ae98a0c33350d02de02",
			sortedHash:   "72e9a426f6c0ac801e84961d47dc407c06e24243afe72a1ca9f019c10c628f80",
		},
		{
			name:         "block 150002 tx[7] - sorts both inputs and outputs",
			hexFile:      "tx150002-7.hex",
			isSorted:     false,
			unsortedHash: "b0e16c9e714c2e6013d48d7c49ab64c8fc3f8be6db80da3097a4ccfd13f98290",
			sortedHash:   "c22918de8faa4aa28f26ef7848f1d9d581d90d876be1e8dab84bc4fbcda029c1",
		},
	}

	for _, test := range tests {
		// Load and deserialize the test transaction.
		filePath := filepath.Join("testdata", test.hexFile)
		txHexBytes, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("ReadFile (%s): failed to read test file: %v",
				test.name, err)
			continue
		}
		txBytes, err := hex.DecodeString(string(txHexBytes))
		if err != nil {
			t.Errorf("DecodeString (%s): failed to decode tx: %v",
				test.name, err)
			continue
		}
		var tx wire.MsgTx
		err = tx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			t.Errorf("Deserialize (%s): unexpected error %v",
				test.name, err)
			continue
		}

		// Ensure the sort order of the original transaction matches the
		// expected value.
		if got := IsSorted(&tx); got != test.isSorted {
			t.Errorf("IsSorted (%s): sort does not match "+
				"expected - got %v, want %v", test.name, got,
				test.isSorted)
			continue
		}

		// Sort the transaction and ensure the resulting hash is the
		// expected value.
		sortedTx := Sort(&tx)
		if got := sortedTx.TxHash().String(); got != test.sortedHash {
			t.Errorf("Sort (%s): sorted hash does not match "+
				"expected - got %v, want %v", test.name, got,
				test.sortedHash)
			continue
		}

		// Ensure the original transaction is not modified.
		if got := tx.TxHash().String(); got != test.unsortedHash {
			t.Errorf("Sort (%s): unsorted hash does not match "+
				"expected - got %v, want %v", test.name, got,
				test.unsortedHash)
			continue
		}

		// Now sort the transaction using the mutable version and ensure
		// the resulting hash is the expected value.
		InPlaceSort(&tx)
		if got := tx.TxHash().String(); got != test.sortedHash {
			t.Errorf("SortMutate (%s): sorted hash does not match "+
				"expected - got %v, want %v", test.name, got,
				test.sortedHash)
			continue
		}
	}
}
