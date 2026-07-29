// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// TestSimNetGenesisBlock tests the genesis block of the simulation test network
// for validity by checking the encoded bytes and hashes.
func TestSimNetGenesisBlock(t *testing.T) {
	simNetGenesisBlockBytes, _ := hex.DecodeString("04000000000000000000000000000000" +
		"000000000000000000000000000000000000000019474bfe726a2fddc7ddf31d" +
		"8c639d340c802a8d89b2ca55340e5fc7b820cf75000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"00000000cccc0c20002d310100000000000000000000000089e1565b1aa4ae18" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000101000000010000" +
		"000000000000000000000000000000000000000000000000000000000000ffff" +
		"ffff00ffffffff0100000000000000000000434104678afdb0fe5548271967f1" +
		"a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504" +
		"e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac000000000000000001ff" +
		"ffffffffffffff00000000ffffffff4d04ffff001d0104455468652054696d65" +
		"732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e" +
		"6b206f66207365636f6e64206261696c6f757420666f722062616e6b7300")

	// Encode the genesis block to raw bytes.
	params := SimNetParams()
	var buf bytes.Buffer
	err := params.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestSimNetGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), simNetGenesisBlockBytes) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(simNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := params.GenesisBlock.BlockHash()
	if !params.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(params.GenesisHash))
	}
}
