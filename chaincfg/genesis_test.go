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

	genesisBlockBytes, _ := hex.DecodeString("01000000000000000000000000000000" +
											"00000000000000000000000000000000" +
											"00000000e0d744c8c5cc4ddb8db59b14" +
											"0e6ff5e09aca011219c25ddb5bbaca5a" +
											"aa432c8a000000000000000000000000" +
											"00000000000000000000000000000000" +
											"00000000000000000000000000000000" +
											"00000000ffff011b00c2eb0b00000000" +
											"0000000000000000a0d7b85600000000" +
											"00000000000000000000000000000000" +
											"00000000000000000000000000000000" +
											"00000000010100000001000000000000" +
											"00000000000000000000000000000000" +
											"00000000000000000000ffffffff00ff" +
											"ffffff01000000000000000000002080" +
											"1679e98561ada96caec2949a5d41c4ca" +
											"b3851eb740d951c10ecbcf265c1fd900" +
											"0000000000000001ffffffffffffffff" +
											"00000000ffffffff02000000")

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

	testNetGenesisBlockBytes, _ := hex.DecodeString("04000000000000000000000000000000" +
													   "00000000000000000000000000000000" +
													   "00000000532904483b63ea815ebdf93e" +
													   "b2fa68d294137c626b81504f3fc6d325" +
													   "378136e7000000000000000000000000" +
													   "00000000000000000000000000000000" +
													   "00000000000000000000000000000000" +
													   "00000000ffff001e002d310100000000" +
													   "000000000000000040bcc8581aa4ae18" +
													   "00000000000000000000000000000000" +
													   "00000000000000000000000000000000" +
													   "00000000010100000001000000000000" +
													   "00000000000000000000000000000000" +
													   "00000000000000000000ffffffff00ff" +
													   "ffffff01000000000000000000004341" +
													   "04678afdb0fe5548271967f1a67130b7" +
													   "105cd6a828e03909a67962e0ea1f61de" +
													   "b649f6bc3f4cef38c4f35504e51ec112" +
													   "de5c384df7ba0b8d578a4c702b6bf11d" +
													   "5fac0000000000000000010000000000" +
													   "00000000000000000000004d04ffff00" +
													   "1d0104455468652054696d6573203033" +
													   "2f4a616e2f32303039204368616e6365" +
													   "6c6c6f72206f6e206272696e6b206f66" +
													   "207365636f6e64206261696c6f757420" +
													   "666f722062616e6b7300")

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

	simNetGenesisBlockBytes, _ := hex.DecodeString("01000000000000000000000000000000" +
													"00000000000000000000000000000000" +
													"00000000e0d744c8c5cc4ddb8db59b14" +
													"0e6ff5e09aca011219c25ddb5bbaca5a" +
													"aa432c8a000000000000000000000000" +
													"00000000000000000000000000000000" +
													"00000000000000000000000000000000" +
													"00000000ffff7f200000000000000000" +
													"00000000000000004506865300000000" +
													"00000000000000000000000000000000" +
													"00000000000000000000000000000000" +
													"00000000010100000001000000000000" +
													"00000000000000000000000000000000" +
													"00000000000000000000ffffffff00ff" +
													"ffffff01000000000000000000004341" +
													"04678afdb0fe5548271967f1a67130b7" +
													"105cd6a828e03909a67962e0ea1f61de" +
													"b649f6bc3f4cef38c4f35504e51ec112" +
													"de5c384df7ba0b8d578a4c702b6bf11d" +
													"5fac0000000000000000010000000000" +
													"00000000000000000000004d04ffff00" +
													"1d0104455468652054696d6573203033" +
													"2f4a616e2f32303039204368616e6365" +
													"6c6c6f72206f6e206272696e6b206f66" +
													"207365636f6e64206261696c6f757420" +
													"666f722062616e6b7300")

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
