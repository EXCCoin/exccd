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

// TestTestNetGenesisBlock tests the genesis block of the test network (version
// 3) for validity by checking the encoded bytes and hashes.
func TestTestNetGenesisBlock(t *testing.T) {
	testNetGenesisBlockBytes, _ := hex.DecodeString(
		"04000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000532904483b63ea815ebdf93e" +
		"b2fa68d294137c626b81504f3fc6d325" +
		"378136e7000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"0000000066660620002d310100000000" +
		"000000000000000089e1565b1aa4ae18" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000000000000000000" +
		"00000000000000000101000000010000" +
		"00000000000000000000000000000000" +
		"0000000000000000000000000000ffff" +
		"ffff00ffffffff010000000000000000" +
		"0000434104678afdb0fe5548271967f1" +
		"a67130b7105cd6a828e03909a67962e0" +
		"ea1f61deb649f6bc3f4cef38c4f35504" +
		"e51ec112de5c384df7ba0b8d578a4c70" +
		"2b6bf11d5fac000000000000000001ff" +
		"ffffffffffffff00000000ffffffff4d" +
		"04ffff001d0104455468652054696d65" +
		"732030332f4a616e2f32303039204368" +
		"616e63656c6c6f72206f6e206272696e" +
		"6b206f66207365636f6e64206261696c" +
		"6f757420666f722062616e6b7300")

	// Encode the genesis block to raw bytes.
	params := TestNet3Params()
	var buf bytes.Buffer
	err := params.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestTestNetGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), testNetGenesisBlockBytes) {
		t.Fatalf("TestTestNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(testNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := params.GenesisBlock.BlockHash()
	if !params.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestTestNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(params.GenesisHash))
	}
}
