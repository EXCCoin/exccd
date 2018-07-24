// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// TestGenesisBlock tests the genesis block of the main network for validity by
// checking the encoded bytes and hashes.
func TestGenesisBlock(t *testing.T) {
	genesisBlockBytes, _ := hex.DecodeString(
		"01000000000000000000000000000000" +
			"00000000000000000000000000000000" +
			"00000000e0d744c8c5cc4ddb8db59b14" +
			"0e6ff5e09aca011219c25ddb5bbaca5a" +
			"aa432c8a000000000000000000000000" +
			"00000000000000000000000000000000" +
			"00000000000000000000000000000000" +
			"00000000d7a3002000c2eb0b00000000" +
			"0000000000000000905e4c5b00000000" +
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
			"000020801679e98561ada96caec2949a" +
			"5d41c4cab3851eb740d951c10ecbcf26" +
			"5c1fd9000000000000000001ffffffff" +
			"ffffffff00000000ffffffff02000000")

	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := MainNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), genesisBlockBytes) {
		t.Fatalf("TestGenesisBlock: Genesis block does not appear valid - "+
			"got %v, want %v", spew.Sdump(buf.Bytes()),
			spew.Sdump(genesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := MainNetParams.GenesisBlock.BlockHash()
	if !MainNetParams.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestGenesisBlock: Genesis block hash does not "+
			"appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(MainNetParams.GenesisHash))
	}
}

// TestTestNetGenesisBlock tests the genesis block of the test network (version
// 9) for validity by checking the encoded bytes and hashes.
func TestTestNetGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := TestNet2Params.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestTestNetGenesisBlock: %v", err)
	}

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
			"2b6bf11d5fac00000000000000000100" +
			"0000000000000000000000000000004d" +
			"04ffff001d0104455468652054696d65" +
			"732030332f4a616e2f32303039204368" +
			"616e63656c6c6f72206f6e206272696e" +
			"6b206f66207365636f6e64206261696c" +
			"6f757420666f722062616e6b7300")

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), testNetGenesisBlockBytes) {
		t.Fatalf("TestTestNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(testNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := TestNet2Params.GenesisBlock.BlockHash()
	if !TestNet2Params.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestTestNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(TestNet2Params.GenesisHash))
	}
}

// TestSimNetGenesisBlock tests the genesis block of the simulation test network
// for validity by checking the encoded bytes and hashes.
func TestSimNetGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := SimNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestSimNetGenesisBlock: %v", err)
	}

	simNetGenesisBlockBytes, _ := hex.DecodeString(
		"01000000000000000000000000000000" +
			"00000000000000000000000000000000" +
			"00000000e0d744c8c5cc4ddb8db59b14" +
			"0e6ff5e09aca011219c25ddb5bbaca5a" +
			"aa432c8a000000000000000000000000" +
			"00000000000000000000000000000000" +
			"00000000000000000000000000000000" +
			"00000000ffff00210000000000000000" +
			"00000000000000004506865300000000" +
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
			"2b6bf11d5fac00000000000000000100" +
			"0000000000000000000000000000004d" +
			"04ffff001d0104455468652054696d65" +
			"732030332f4a616e2f32303039204368" +
			"616e63656c6c6f72206f6e206272696e" +
			"6b206f66207365636f6e64206261696c" +
			"6f757420666f722062616e6b7300")

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), simNetGenesisBlockBytes) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(simNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := SimNetParams.GenesisBlock.BlockHash()
	if !SimNetParams.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(SimNetParams.GenesisHash))
	}
}
