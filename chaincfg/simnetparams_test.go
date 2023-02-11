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
	simNetGenesisBlockBytes, _ := hex.DecodeString("01000000000000000000000000" +
		"00000000000000000000000000000000000000000000006cb3df3e601d42b43182" +
		"6e87ab35a5c2a6a25081479060c5fe6d90d24a9b97510000000000000000000000" +
		"000000000000000000000000000000000000000000000000000000000000000000" +
		"00000000ffff7f2000000000000000000000000000000000450686530000000000" +
		"000000000000000000000000000000000000000000000000000000000000000000" +
		"000000000000000000000000000000000000000000000000000000000000000000" +
		"000000000000000000000000000000000000000000000000000000000000000000" +
		"000000000000000000000000000000000000000000000000000000000000000000" +
		"000000010100000001000000000000000000000000000000000000000000000000" +
		"0000000000000000ffffffff00ffffffff0100000000000000000000434104678a" +
		"fdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc" +
		"3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac000000" +
		"000000000001000000000000000000000000000000004d04ffff001d0104455468" +
		"652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e" +
		"206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b" +
		"7300")

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
