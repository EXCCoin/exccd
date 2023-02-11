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

// TestGenesisBlock tests the genesis block of the main network for validity by
// checking the encoded bytes and hashes.
func TestGenesisBlock(t *testing.T) {
	genesisBlockBytes, _ := hex.DecodeString("0100000000000000000000000000" +
		"00000000000000000000000000000000000000000000e0d744c8c5cc4ddb8db59" +
		"b140e6ff5e09aca011219c25ddb5bbaca5aaa432c8a0000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000" +
		"000000000d7a3002000c2eb0b000000000000000000000000905e4c5b00000000" +
		"00000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000000000000000000000000000000000000000000000000000000000" +
		"00000000000001010000000100000000000000000000000000000000000000000" +
		"00000000000000000000000ffffffff00ffffffff010000000000000000000020" +
		"801679e98561ada96caec2949a5d41c4cab3851eb740d951c10ecbcf265c1fd90" +
		"00000000000000001ffffffffffffffff00000000ffffffff02000000")

	// Encode the genesis block to raw bytes.
	params := MainNetParams()
	var buf bytes.Buffer
	err := params.GenesisBlock.Serialize(&buf)
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
	hash := params.GenesisBlock.BlockHash()
	if !params.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestGenesisBlock: Genesis block hash does not "+
			"appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(params.GenesisHash))
	}
}
