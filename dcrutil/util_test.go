// Copyright (c) 2020-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcrutil

import (
	"testing"
)

// mockAddrParams implements the AddressParams interface and is used throughout
// the tests to mock multiple networks.
type mockAddrParams struct {
	pubKeyID     [2]byte
	pkhEcdsaID   [2]byte
	pkhEd25519ID [2]byte
	pkhSchnorrID [2]byte
	scriptHashID [2]byte
	privKeyID    byte
}

// AddrIDPubKeyV0 returns the magic prefix bytes associated with the mock params
// for version 0 pay-to-pubkey addresses.
//
// This is part of the AddressParams interface.
func (p *mockAddrParams) AddrIDPubKeyV0() [2]byte {
	return p.pubKeyID
}

// AddrIDPubKeyHashECDSAV0 returns the magic prefix bytes associated with the
// mock params for version 0 pay-to-pubkey-hash addresses where the underlying
// pubkey is secp256k1 and the signature algorithm is ECDSA.
//
// This is part of the AddressParams interface.
func (p *mockAddrParams) AddrIDPubKeyHashECDSAV0() [2]byte {
	return p.pkhEcdsaID
}

// AddrIDPubKeyHashEd25519V0 returns the magic prefix bytes associated with the
// mock params for version 0 pay-to-pubkey-hash addresses where the underlying
// pubkey and signature algorithm are Ed25519.
//
// This is part of the AddressParams interface.
func (p *mockAddrParams) AddrIDPubKeyHashEd25519V0() [2]byte {
	return p.pkhEd25519ID
}

// AddrIDPubKeyHashSchnorrV0 returns the magic prefix bytes associated with the
// mock params for version 0 pay-to-pubkey-hash addresses where the underlying
// pubkey is secp256k1 and the signature algorithm is Schnorr.
//
// This is part of the AddressParams interface.
func (p *mockAddrParams) AddrIDPubKeyHashSchnorrV0() [2]byte {
	return p.pkhSchnorrID
}

// AddrIDScriptHashV0 returns the magic prefix bytes associated with the mock
// params for version 0 pay-to-script-hash addresses.
//
// This is part of the AddressParams interface.
func (p *mockAddrParams) AddrIDScriptHashV0() [2]byte {
	return p.scriptHashID
}

// mockMainNetParams returns mock mainnet address parameters to use throughout
// the tests.  They match the Decred mainnet params as of the time this comment
// was written.
func mockMainNetParams() *mockAddrParams {
	return &mockAddrParams{
		pubKeyID:     [2]byte{0x02, 0xdc}, // starts with 2s    -- no such addresses should exist in RL
		pkhEcdsaID:   [2]byte{0x21, 0xB9}, // starts with 22
		pkhEd25519ID: [2]byte{0x35, 0xcf}, // starts with 2e
		pkhSchnorrID: [2]byte{0x2f, 0x0d}, // starts with 2S
		scriptHashID: [2]byte{0x34, 0xAF}, // starts with 2c
		privKeyID:    0x80,                // starts with 5 (uncompressed) or K (compressed)
	}
}

// mockTestNetParams returns mock testnet address parameters to use throughout
// the tests.  They match the Decred mainnet params as of the time this comment
// was written.
func mockTestNetParams() *mockAddrParams {
	return &mockAddrParams{
		pubKeyID:     [2]byte{0x28, 0xf7}, // starts with Tk
		pkhEcdsaID:   [2]byte{0x0f, 0x21}, // starts with Ts
		pkhEd25519ID: [2]byte{0x0f, 0x01}, // starts with Te
		pkhSchnorrID: [2]byte{0x0e, 0xe3}, // starts with TS
		scriptHashID: [2]byte{0x0e, 0xfc}, // starts with Tc
		privKeyID:    0xef,                // starts with 9 (uncompressed) or c (compressed)
	}
}

// TestVerifyMessage ensures the verifying a message works as intended.
func TestVerifyMessage(t *testing.T) {
	mainNetParams := mockMainNetParams()
	testNetParams := mockTestNetParams()

	msg := "verifymessage test"

	var tests = []struct {
		name    string
		addr    string
		sig     string
		params  AddressParams
		isValid bool
	}{{
		name:    "valid",
		addr:    "TsULH3kCRvDjwTh9CoMnhiNMUkk2JvaecRB",
		sig:     "HxvjYfirrn3ccDhg26XSi/6y5+tIbuuOFLgF70pTeDSvWYE9/II6pZFaIvp95pfyaBhFhht7Dt5TvRe7g5BTwfQ=",
		params:  testNetParams,
		isValid: true,
	}, {
		name:    "wrong address",
		addr:    "TsakkhCjU7t4D46AkeNoCshtunX5rSfghJX",
		sig:     "IHJ2BFnXaXgvKsBvdNTTlHYmbM8Dy/xihFNLq/F+9LbcPYicaSy7+wwNU7LliD4PJyy9rKSqeOQDl0xAV416CKI=",
		params:  testNetParams,
		isValid: false,
	}, {
		name:    "wrong signature",
		addr:    "TsVGSmQMRDX8wQu82SuWQedmXCWW9hEPApz",
		sig:     "IAmUid56BTjvPYSuwgv1n8tDl8f7dqNZOhpvie4H3OaYVjdhRWFeRSjsUN3D7mfRhiV/9OIsWKepJ9ghrvHUbh8=",
		params:  testNetParams,
		isValid: false,
	}, {
		name:    "wrong params",
		addr:    "TsVRVzw5Nbjp4eVVhq2y9ankcNp78imTGKd",
		sig:     "IAmUid56BTjvPYSuwgv1n8tDl8f7dqNZOhpvie4H3OaYVjdhRWFeRSjsUN3D7mfRhiV/9OIsWKepJ9ghrvHUbh8=",
		params:  mainNetParams,
		isValid: false,
	}}

	for _, test := range tests {
		err := VerifyMessage(test.addr, test.sig, msg, test.params)
		if (test.isValid && err != nil) || (!test.isValid && err == nil) {
			t.Fatalf("%s: failed: %v", test.name, err)
		}
	}
}
