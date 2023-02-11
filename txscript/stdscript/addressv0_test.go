// Copyright (c) 2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package stdscript

import "fmt"

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
		pubKeyID:    [2]byte{0x02, 0xdc}, // starts with 2s    -- no such addresses should exist in RL
		pkhEcdsaID:  [2]byte{0x21, 0xB9}, // starts with 22
		pkhEd25519ID:[2]byte{0x35, 0xcf}, // starts with 2e
		pkhSchnorrID:[2]byte{0x2f, 0x0d}, // starts with 2S
		scriptHashID:[2]byte{0x34, 0xAF}, // starts with 2c
		privKeyID:   0x80,                // starts with 5 (uncompressed) or K (compressed)
	}
}

// mockTestNetParams returns mock testnet address parameters to use throughout
// the tests.  They match the Decred testnet params as of the time this comment
// was written.
func mockTestNetParams() *mockAddrParams {
	return &mockAddrParams{
		pubKeyID:    [2]byte{0x28, 0xf7}, // starts with Tk
		pkhEcdsaID:  [2]byte{0x0f, 0x21}, // starts with Ts
		pkhEd25519ID:[2]byte{0x0f, 0x01}, // starts with Te
		pkhSchnorrID:[2]byte{0x0e, 0xe3}, // starts with TS
		scriptHashID:[2]byte{0x0e, 0xfc}, // starts with Tc
		privKeyID:   0xef,                // starts with 9 (uncompressed) or c (compressed)
	}
}

// addressV0Tests houses several version 0 test scripts used to ensure various
// script types and address extraction is working as expected.  It's defined as
// a test global versus inside a specific test function scope so it can remain
// separate from tests for other future script versions.
var addressV0Tests = func() []addressTest {
	mainNetParams := mockMainNetParams()
	testNetParams := mockTestNetParams()

	// Convenience function that combines fmt.Sprintf with mustParseShortForm
	// to create more compact tests.
	p := func(format string, a ...interface{}) []byte {
		const scriptVersion = 0
		return mustParseShortForm(scriptVersion, fmt.Sprintf(format, a...))
	}

	// ---------------------------------------------------------------------
	// Define some data shared in the tests for convenience.
	// ---------------------------------------------------------------------

	// Uncompressed and compressed/hybrid even/odd secp256k1 public keys along
	// with hash160s of the compressed even ones.
	pkUE := "0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f817" +
		"98483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
	pkUO := "04fff97bd5755eeea420453a14355235d382f6472f8568a18b2f057a14602975" +
		"56ae12777aacfbb620f3be96017f45c560de80f0f6518fe4a03c870c36b075f297"
	pkCE := "02" + pkUE[2:66]
	h160CE := "e280cb6e66b96679aec288b1fbdbd4db08077a1b"
	pkCE2 := "02f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9"
	h160CE2 := "01557763e0252dc0ff9e0996ad1d04b167bb993c"
	pkCO := "03" + pkUO[2:66]
	pkHE := "05" + pkUE[2:]
	pkHO := "06" + pkUO[2:]

	// Ed25519 public key and hash.
	pkEd := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
	h160Ed := "456d8ee57a4b9121987b4ecab8c3bcb5797e8a53"

	// Script hash for a 2-of-3 multisig composed of pkCE, pkCE2, and pkCO.
	p2sh := "f86b5a7c6d32566aa4dccc04d1533530b4d64cf3"

	return []addressTest{{
		// ---------------------------------------------------------------------
		// Misc negative tests.
		// ---------------------------------------------------------------------

		name:     "malformed v0 script that does not parse",
		script:   p("DATA_5 0x01020304"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "empty v0 script",
		script:   nil,
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name:     "v0 p2pk-ecdsa-secp256k1 hybrid odd",
		script:   p("DATA_33 0x%s CHECKSIG", pkHO),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "v0 p2pk-ecdsa-secp256k1 hybrid even",
		script:   p("DATA_33 0x%s CHECKSIG", pkHE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-ecdsa-secp256k1 -- trailing opcode",
		script:   p("DATA_33 0x%s CHECKSIG TRUE", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-ecdsa-secp256k1 -- pubkey not pushed",
		script:   p("0x%s CHECKSIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-ecdsa-secp256k1 -- malformed pubkey prefix",
		script:   p("DATA_33 0x08%s CHECKSIG", pkCE[2:]),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PK ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 p2pk-ecdsa-secp256k1 uncompressed",
		script:    p("DATA_65 0x%s CHECKSIG", pkUE),
		params:    mainNetParams,
		wantType:  STPubKeyEcdsaSecp256k1,
		wantAddrs: []string{"2sHTV8v3sg9pLGJ6RvQDyAidVn8pNKz9HR8JvYLeJ2Yb39Ecn9We2"},
	}, {
		name:      "mainnet v0 p2pk-ecdsa-secp256k1 compressed even",
		script:    p("DATA_33 0x%s CHECKSIG", pkCE),
		params:    mainNetParams,
		wantType:  STPubKeyEcdsaSecp256k1,
		wantAddrs: []string{"2sHTV8v3sg9pLGJ6RvQDyAidVn8pNKz9HR8JvYLeJ2Yb39Ecn9We2"},
	}, {
		name:      "mainnet v0 p2pk-ecdsa-secp256k1 compressed odd",
		script:    p("DATA_33 0x%s CHECKSIG", pkCO),
		params:    mainNetParams,
		wantType:  STPubKeyEcdsaSecp256k1,
		wantAddrs: []string{"2sMmKYEy8kvhy7znWAZQYHj8maMzd9YaSWfpcEMtHbQHqgjKxPiX9"},
	}, {
		name:      "testnet v0 p2pk-ecdsa-secp256k1 uncompressed",
		script:    p("DATA_65 0x%s CHECKSIG", pkUE),
		params:    testNetParams,
		wantType:  STPubKeyEcdsaSecp256k1,
		wantAddrs: []string{"TkKmUYE9BRkzYvDHnjYbYAXVfWfcnp9FdRe4N1YMppRTArJ6nzapC"},
	}, {
		name:      "testnet v0 p2pk-ecdsa-secp256k1 compressed even",
		script:    p("DATA_33 0x%s CHECKSIG", pkCE),
		params:    testNetParams,
		wantType:  STPubKeyEcdsaSecp256k1,
		wantAddrs: []string{"TkKmUYE9BRkzYvDHnjYbYAXVfWfcnp9FdRe4N1YMppRTArJ6nzapC"},
	}, {
		name:      "testnet v0 p2pk-ecdsa-secp256k1 compressed odd",
		script:    p("DATA_33 0x%s CHECKSIG", pkCO),
		params:    testNetParams,
		wantType:  STPubKeyEcdsaSecp256k1,
		wantAddrs: []string{"TkQ5JwZ4SWXtBmuyryhn7HXzwJto3dhgnXBa3hZbpPH9yPnrbvJTg"},
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Alt tests.
		// ---------------------------------------------------------------------

		name:     "v0 p2pk-alt unsupported signature type 0",
		script:   p("DATA_33 0x%s 0 CHECKSIGALT", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "v0 p2pk-alt unsupported signature type 3",
		script:   p("DATA_33 0x%s 3 CHECKSIGALT", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-alt -- signature type not small int",
		script:   p("DATA_33 0x%s DATA_1 2 CHECKSIGALT", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-alt -- NOP for signature type",
		script:   p("DATA_33 0x%s NOP CHECKSIGALT", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Ed25519 tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 p2pk-ed25519 -- trailing opcode",
		script:   p("DATA_32 0x%s 1 CHECKSIGALT TRUE", pkEd),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-ed25519 -- pubkey not pushed",
		script:   p("0x%s 1 CHECKSIGALT", pkEd),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-ed25519 -- wrong signature type",
		script:   p("DATA_32 0x%s 2 CHECKSIGALT", pkEd),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PK Ed25519 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 p2pk-ed25519",
		script:    p("DATA_32 0x%s 1 CHECKSIGALT", pkEd),
		params:    mainNetParams,
		wantType:  STPubKeyEd25519,
		wantAddrs: []string{"2sHW5LfhGgZokqA4SzE9UUNQpuzuY4g1J8aoTGPhuHeNSsnesn9Dw"},
	}, {
		name:      "testnet v0 p2pk-ed25519",
		script:    p("DATA_32 0x%s 1 CHECKSIGALT", pkEd),
		params:    testNetParams,
		wantType:  STPubKeyEd25519,
		wantAddrs: []string{"TkKp4jynaSAyyV5FooNX3UBGzeXhxYq7e96YtjbRS5XEaar95isMx"},
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name:     "v0 p2pk-schnorr-secp256k1 uncompressed",
		script:   p("DATA_65 0x%s 2 CHECKSIGALT", pkUE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "v0 p2pk-schnorr-secp256k1 hybrid odd",
		script:   p("DATA_65 0x%s 2 CHECKSIGALT", pkHO),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "v0 p2pk-schnorr-secp256k1 hybrid even",
		script:   p("DATA_65 0x%s 2 CHECKSIGALT", pkHE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-schnorr-secp256k1 -- trailing opcode",
		script:   p("DATA_33 0x%s 2 CHECKSIGALT TRUE", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-schnorr-secp256k1 -- pubkey not pushed",
		script:   p("0x%s 2 CHECKSIGALT", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2pk-schnorr-secp256k1 -- malformed pubkey prefix",
		script:   p("DATA_33 0x08%s 2 CHECKSIGALT", pkCE[2:]),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PK Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 p2pk-schnorr-secp256k1 compressed even",
		script:    p("DATA_33 0x%s 2 CHECKSIGALT", pkCE),
		params:    mainNetParams,
		wantType:  STPubKeySchnorrSecp256k1,
		wantAddrs: []string{"2sHXNdGWmjrfgiZjbLDPDAdGhKy3bmHYyAhb4kpwfpXu3vDsXfuKW"},
	}, {
		name:      "mainnet v0 p2pk-schnorr-secp256k1 compressed odd",
		script:    p("DATA_33 0x%s 2 CHECKSIGALT", pkCO),
		params:    mainNetParams,
		wantType:  STPubKeySchnorrSecp256k1,
		wantAddrs: []string{"2sMqD2bS2pdZKaGRfaNZnHdmy8CDraqz8GF6kSrBfPPbrTiarq3Nu"},
	}, {
		name:      "testnet v0 p2pk-schnorr-secp256k1 compressed even",
		script:    p("DATA_33 0x%s 2 CHECKSIGALT", pkCE),
		params:    testNetParams,
		wantType:  STPubKeySchnorrSecp256k1,
		wantAddrs: []string{"TkKqN2ac5VTquNUvx9MknAS8s4Vr2FSfKBDLWE2fCcQmBdHKwyMBG"},
	}, {
		name:      "testnet v0 p2pk-schnorr-secp256k1 compressed odd",
		script:    p("DATA_33 0x%s 2 CHECKSIGALT", pkCO),
		params:    testNetParams,
		wantType:  STPubKeySchnorrSecp256k1,
		wantAddrs: []string{"TkQ9CRuXLaEjYEBd2PWwMHSe8rj2H516UGkrBv3uCBGTzAn6gnRmD"},
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 p2pkh-ecdsa-secp256k1 -- wrong hash length",
		script:   p("DUP HASH160 DATA_21 0x00%s EQUALVERIFY CHECKSIG", h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PKH ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 p2pkh-ecdsa-secp256k1",
		script:    p("DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG", h160CE),
		params:    mainNetParams,
		wantType:  STPubKeyHashEcdsaSecp256k1,
		wantAddrs: []string{"22u6UvKbhmoYPMC85LKcjfFmdrWsaGrE2qnB"},
	}, {
		name:      "testnet v0 p2pkh-ecdsa-secp256k1",
		script:    p("DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG", h160CE),
		params:    testNetParams,
		wantType:  STPubKeyHashEcdsaSecp256k1,
		wantAddrs: []string{"TsmfmUitQApgnNxQypdGd2x36djCCo1VQAn"},
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH Alt tests.
		// ---------------------------------------------------------------------

		name: "v0 p2pkh-alt unsupported signature type 0",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 0 CHECKSIGALT",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name: "v0 p2pkh-alt unsupported signature type 3",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 3 CHECKSIGALT",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name: "almost v0 p2pkh-alt -- signature type not a small int",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY DATA_1 2 CHECKSIGALT",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name: "almost v0 p2pkh-alt -- NOP for signature type",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY NOP CHECKSIGALT",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH Ed25519 tests.
		// ---------------------------------------------------------------------

		name: "almost v0 p2pkh-ed25519 -- wrong hash length",
		script: p("DUP HASH160 DATA_21 0x00%s EQUALVERIFY 1 CHECKSIGALT",
			h160Ed),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PKH Ed25519 tests.
		// ---------------------------------------------------------------------

		name: "mainnet v0 p2pkh-ed25519",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 1 CHECKSIGALT",
			h160Ed),
		params:    mainNetParams,
		wantType:  STPubKeyHashEd25519,
		wantAddrs: []string{"2eZt5xAGLspEZeTLF29nVJsYW4nbR6jP9SCJ"},
	}, {
		name: "testnet v0 p2pkh-ed25519",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 1 CHECKSIGALT",
			h160Ed),
		params:    testNetParams,
		wantType:  STPubKeyHashEd25519,
		wantAddrs: []string{"TeeXvqZJrc7KnFZCT27fHfzcrTTzSHCgpHu"},
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "almost v0 p2pkh-schnorr-secp256k1 -- wrong hash length",
		script: p("DUP HASH160 DATA_21 0x00%s EQUALVERIFY 2 CHECKSIGALT",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PKH Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "mainnet v0 p2pkh-schnorr-secp256k1",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 2 CHECKSIGALT",
			h160CE),
		params:    mainNetParams,
		wantType:  STPubKeyHashSchnorrSecp256k1,
		wantAddrs: []string{"2Sa1gyt6inM5Sqb1nuq5D5oaL7z2enVbMub8"},
	}, {
		name: "mainnetv0 p2pkh-schnorr-secp256k1 2",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 2 CHECKSIGALT",
			h160CE2),
		params:    mainNetParams,
		wantType:  STPubKeyHashSchnorrSecp256k1,
		wantAddrs: []string{"2SZfAPwSDQH92ffNxUKzzDpmjV9auMR2zURn"},
	}, {
		name: "testnet v0 p2pkh-schnorr-secp256k1",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 2 CHECKSIGALT",
			h160CE),
		params:    testNetParams,
		wantType:  STPubKeyHashSchnorrSecp256k1,
		wantAddrs: []string{"TSpiNRzQY9dxTg5mcJccryBAKcTDifW1vHw"},
	}, {
		name: "testnetv0 p2pkh-schnorr-secp256k1 2",
		script: p("DUP HASH160 DATA_20 0x%s EQUALVERIFY 2 CHECKSIGALT",
			h160CE2),
		params:    testNetParams,
		wantType:  STPubKeyHashSchnorrSecp256k1,
		wantAddrs: []string{"TSUBnVKuA5hYHkSwAoYPzzNZgn1UHaohyC6"},
	}, {
		// ---------------------------------------------------------------------
		// Negative P2SH tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 p2sh -- wrong hash length",
		script:   p("HASH160 DATA_21 0x00%s EQUAL", p2sh),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 p2sh -- trailing opcode",
		script:   p("HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2SH tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 p2sh",
		script:    p("HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    mainNetParams,
		wantType:  STScriptHash,
		wantAddrs: []string{"2caH4eu9E49X8J4xPzrYWrWiVGoN5opNizFk"},
	}, {
		name:      "testnet v0 p2sh",
		script:    p("HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    testNetParams,
		wantType:  STScriptHash,
		wantAddrs: []string{"TcvALEAYZsT2PJqowebx2Yedhaza6c8vzfg"},
	}, {
		// ---------------------------------------------------------------------
		// Negative ECDSA multisig secp256k1 tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 multisig 1-of-2 -- mixed (un)compressed pubkeys",
		script:   p("1 DATA_65 0x%s DATA_33 0x%s 2 CHECKMULTISIG", pkUE, pkCO),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- no req sigs",
		script:   p("0 0 CHECKMULTISIG"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- invalid pubkey",
		script:   p("1 DATA_32 0x%s 1 CHECKMULTISIG", pkCE[2:]),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- hybrid pubkey",
		script:   p("1 DATA_65 0x%s 1 CHECKMULTISIG", pkHO),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- invalid number of signatures",
		script:   p("DUP DATA_33 0x%s 1 CHECKMULTISIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- ends with CHECKSIG instead",
		script:   p("1 DATA_33 0x%s 1 CHECKSIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- num required sigs not small int",
		script:   p("DATA_1 1 DATA_33 0x%s 1 CHECKMULTISIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- num public keys not small int",
		script:   p("1 DATA_33 0x%s DATA_1 1 CHECKMULTISIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- missing num public keys",
		script:   p("1 DATA_33 0x%s CHECKMULTISIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- num pubkeys does not match given keys",
		script:   p("2 DATA_33 0x%s DATA_33 0x%s 3 CHECKMULTISIG", pkCE, pkCO),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- fewer pubkeys than num required sigs",
		script:   p("1 0 CHECKMULTISIG"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- CHECKMULTISIGVERIFY",
		script:   p("1 DATA_33 0x%s 1 CHECKMULTISIGVERIFY", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- extra NOP prior to final opcode",
		script:   p("1 DATA_33 0x%s 1 NOP CHECKMULTISIG", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- trailing opcode",
		script:   p("1 DATA_33 0x%s 1 CHECKMULTISIG TRUE", pkCE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 multisig -- no pubkeys specified",
		script:   p("1 CHECKMULTISIG"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive ECDSA multisig secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 multisig 1-of-1 compressed pubkey",
		script:    p("1 DATA_33 0x%s 1 CHECKMULTISIG", pkCE),
		params:    mainNetParams,
		wantType:  STMultiSig,
		wantAddrs: []string{"2sHTV8v3sg9pLGJ6RvQDyAidVn8pNKz9HR8JvYLeJ2Yb39Ecn9We2"},
	}, {
		name:     "mainnet v0 multisig 1-of-2 compressed pubkeys",
		script:   p("1 DATA_33 0x%s DATA_33 0x%s 2 CHECKMULTISIG", pkCE, pkCE2),
		params:   mainNetParams,
		wantType: STMultiSig,
		wantAddrs: []string{
			"2sHTV8v3sg9pLGJ6RvQDyAidVn8pNKz9HR8JvYLeJ2Yb39Ecn9We2",
			"2sHUTGMBVTX15bK1csJn12YgPRBYSiNtvrApwnnCH41Apm2XLdsNm",
		},
	}, {
		name: "mainnet v0 multisig 2-of-3 compressed pubkeys",
		script: p("2 DATA_33 0x%s DATA_33 0x%s DATA_33 0x%s 3 CHECKMULTISIG",
			pkCE, pkCE2, pkCO),
		params:   mainNetParams,
		wantType: STMultiSig,
		wantAddrs: []string{
			"2sHTV8v3sg9pLGJ6RvQDyAidVn8pNKz9HR8JvYLeJ2Yb39Ecn9We2",
			"2sHUTGMBVTX15bK1csJn12YgPRBYSiNtvrApwnnCH41Apm2XLdsNm",
			"2sMmKYEy8kvhy7znWAZQYHj8maMzd9YaSWfpcEMtHbQHqgjKxPiX9",
		},
	}, {
		name:      "testnet v0 multisig 1-of-1 compressed pubkey",
		script:    p("1 DATA_33 0x%s 1 CHECKMULTISIG", pkCE),
		params:    testNetParams,
		wantType:  STMultiSig,
		wantAddrs: []string{"TkKmUYE9BRkzYvDHnjYbYAXVfWfcnp9FdRe4N1YMppRTArJ6nzapC"},
	}, {
		name:     "testnet v0 multisig 1-of-2 compressed pubkeys",
		script:   p("1 DATA_33 0x%s DATA_33 0x%s 2 CHECKMULTISIG", pkCE, pkCE2),
		params:   testNetParams,
		wantType: STMultiSig,
		wantAddrs: []string{
			"TkKmUYE9BRkzYvDHnjYbYAXVfWfcnp9FdRe4N1YMppRTArJ6nzapC",
			"TkKnSffGoD8BJFECygT9a2MYZ9iLsCY1GrgaPFyuoqt2xU5veVwLq",
		},
	}, {
		name: "testnet v0 multisig 2-of-3 compressed pubkeys",
		script: p("2 DATA_33 0x%s DATA_33 0x%s DATA_33 0x%s 3 CHECKMULTISIG",
			pkCE, pkCE2, pkCO),
		params:   testNetParams,
		wantType: STMultiSig,
		wantAddrs: []string{
			"TkKmUYE9BRkzYvDHnjYbYAXVfWfcnp9FdRe4N1YMppRTArJ6nzapC",
			"TkKnSffGoD8BJFECygT9a2MYZ9iLsCY1GrgaPFyuoqt2xU5veVwLq",
			"TkQ5JwZ4SWXtBmuyryhn7HXzwJto3dhgnXBa3hZbpPH9yPnrbvJTg",
		},
	}, {
		// ---------------------------------------------------------------------
		// Negative nulldata tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 nulldata -- NOP instead of data push",
		script:   p("RETURN NOP"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 nulldata -- non-canonical small int push (DATA_1 vs 12)",
		script:   p("RETURN DATA_1 0x0c"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 nulldata -- non-canonical small int push (PUSHDATA1 vs 12)",
		script:   p("RETURN PUSHDATA1 0x01 0x0c"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name: "almost v0 nulldata -- non-canonical 60-byte push (PUSHDATA1 vs DATA_60)",
		script: p("RETURN PUSHDATA1 0x3c 0x046708afdb0fe5548271967f1a67130b7105" +
			"cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef3046708afdb0fe5548271" +
			"967f1a67130b7105cd6a"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 nulldata -- non-canonical 12-byte push (PUSHDATA2)",
		script:   p("RETURN PUSHDATA2 0x0c00 0x046708afdb0fe5548271967f"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 nulldata -- non-canonical 12-byte push (PUSHDATA4)",
		script:   p("RETURN PUSHDATA4 0x0c000000 0x046708afdb0fe5548271967f"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 nulldata -- exceeds max standard push",
		script:   p("RETURN PUSHDATA2 0x0101 0x01{257}"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		name:     "almost v0 nulldata -- trailing opcode",
		script:   p("RETURN 4 TRUE"),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive nulldata tests.
		// ---------------------------------------------------------------------

		name:     "mainnet v0 nulldata no data push",
		script:   p("RETURN"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		name:     "mainnet v0 nulldata single zero push",
		script:   p("RETURN 0"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		name:     "mainnet v0 nulldata small int push",
		script:   p("RETURN 1"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		name:     "mainnet v0 nulldata max small int push",
		script:   p("RETURN 16"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		name:     "mainnet v0 nulldata small data push",
		script:   p("RETURN DATA_8 0x046708afdb0fe554"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		name: "mainnet v0 nulldata 60-byte push",
		script: p("RETURN 0x3c 0x046708afdb0fe5548271967f1a67130b7105cd6a828e03" +
			"909a67962e0ea1f61deb649f6bc3f4cef3046708afdb0fe5548271967f1a6713" +
			"0b7105cd6a"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		name:     "mainnet v0 nulldata max standard push",
		script:   p("RETURN PUSHDATA2 0x0001 0x01{256}"),
		params:   mainNetParams,
		wantType: STNullData,
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission P2PKH tests.
		// ---------------------------------------------------------------------

		name: "almost v0 stake sub p2pkh-ecdsa-secp256k1 -- wrong hash length",
		script: p("SSTX DUP HASH160 DATA_21 0x00%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission P2PKH tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 stake submission p2pkh-ecdsa-secp256k1",
		script:    p("SSTX DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG", h160CE),
		params:    mainNetParams,
		wantType:  STStakeSubmissionPubKeyHash,
		wantAddrs: []string{"22u6UvKbhmoYPMC85LKcjfFmdrWsaGrE2qnB"},
	}, {
		name:      "testnet v0 stake submission p2pkh-ecdsa-secp256k1",
		script:    p("SSTX DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG", h160CE),
		params:    testNetParams,
		wantType:  STStakeSubmissionPubKeyHash,
		wantAddrs: []string{"TsmfmUitQApgnNxQypdGd2x36djCCo1VQAn"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission P2SH tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 stake submission p2sh -- wrong hash length",
		script:   p("SSTX HASH160 DATA_21 0x00%s EQUAL", p2sh),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission P2SH tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 stake submission p2sh",
		script:    p("SSTX HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    mainNetParams,
		wantType:  STStakeSubmissionScriptHash,
		wantAddrs: []string{"2caH4eu9E49X8J4xPzrYWrWiVGoN5opNizFk"},
	}, {
		name:      "testnet v0 stake submission p2sh",
		script:    p("SSTX HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    testNetParams,
		wantType:  STStakeSubmissionScriptHash,
		wantAddrs: []string{"TcvALEAYZsT2PJqowebx2Yedhaza6c8vzfg"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission generation P2PKH tests.
		// ---------------------------------------------------------------------

		name: "almost v0 stake gen p2pkh-ecdsa-secp256k1 -- wrong hash length",
		script: p("SSGEN DUP HASH160 DATA_21 0x00%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission generation P2PKH tests.
		// ---------------------------------------------------------------------

		name: "mainnet v0 stake gen p2pkh-ecdsa-secp256k1",
		script: p("SSGEN DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:    mainNetParams,
		wantType:  STStakeGenPubKeyHash,
		wantAddrs: []string{"22u6UvKbhmoYPMC85LKcjfFmdrWsaGrE2qnB"},
	}, {
		name: "testnet v0 stake gen p2pkh-ecdsa-secp256k1",
		script: p("SSGEN DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:    testNetParams,
		wantType:  STStakeGenPubKeyHash,
		wantAddrs: []string{"TsmfmUitQApgnNxQypdGd2x36djCCo1VQAn"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission generation P2SH tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 stake gen p2sh -- wrong hash length",
		script:   p("SSGEN HASH160 DATA_21 0x00%s EQUAL", p2sh),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission generation P2SH tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 stake gen p2sh",
		script:    p("SSGEN HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    mainNetParams,
		wantType:  STStakeGenScriptHash,
		wantAddrs: []string{"2caH4eu9E49X8J4xPzrYWrWiVGoN5opNizFk"},
	}, {
		name:      "testnet v0 stake gen p2sh",
		script:    p("SSGEN HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    testNetParams,
		wantType:  STStakeGenScriptHash,
		wantAddrs: []string{"TcvALEAYZsT2PJqowebx2Yedhaza6c8vzfg"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission revocation P2PKH tests.
		// ---------------------------------------------------------------------

		name: "almost v0 stake revoke p2pkh-ecdsa-secp256k1 -- wrong hash length",
		script: p("SSRTX DUP HASH160 DATA_21 0x00%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission revocation P2PKH tests.
		// ---------------------------------------------------------------------

		name: "mainnet v0 stake revoke p2pkh-ecdsa-secp256k1",
		script: p("SSRTX DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:    mainNetParams,
		wantType:  STStakeRevocationPubKeyHash,
		wantAddrs: []string{"22u6UvKbhmoYPMC85LKcjfFmdrWsaGrE2qnB"},
	}, {
		name: "testnet v0 stake revoke p2pkh-ecdsa-secp256k1",
		script: p("SSRTX DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:    testNetParams,
		wantType:  STStakeRevocationPubKeyHash,
		wantAddrs: []string{"TsmfmUitQApgnNxQypdGd2x36djCCo1VQAn"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission revocation P2SH tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 stake revoke p2sh -- wrong hash length",
		script:   p("SSRTX HASH160 DATA_21 0x00%s EQUAL", p2sh),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission revocation P2SH tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 stake revoke p2sh",
		script:    p("SSRTX HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    mainNetParams,
		wantType:  STStakeRevocationScriptHash,
		wantAddrs: []string{"2caH4eu9E49X8J4xPzrYWrWiVGoN5opNizFk"},
	}, {
		name:      "testnet v0 stake revoke p2sh",
		script:    p("SSRTX HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    testNetParams,
		wantType:  STStakeRevocationScriptHash,
		wantAddrs: []string{"TcvALEAYZsT2PJqowebx2Yedhaza6c8vzfg"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission change P2PKH tests.
		// ---------------------------------------------------------------------

		name: "almost v0 stake change p2pkh-ecdsa-secp256k1 -- wrong hash length",
		script: p("SSTXCHANGE DUP HASH160 DATA_21 0x00%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission change P2PKH tests.
		// ---------------------------------------------------------------------

		name: "mainnet v0 stake change p2pkh-ecdsa-secp256k1",
		script: p("SSTXCHANGE DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:    mainNetParams,
		wantType:  STStakeChangePubKeyHash,
		wantAddrs: []string{"22u6UvKbhmoYPMC85LKcjfFmdrWsaGrE2qnB"},
	}, {
		name: "testnet v0 stake change p2pkh-ecdsa-secp256k1",
		script: p("SSTXCHANGE DUP HASH160 DATA_20 0x%s EQUALVERIFY CHECKSIG",
			h160CE),
		params:    testNetParams,
		wantType:  STStakeChangePubKeyHash,
		wantAddrs: []string{"TsmfmUitQApgnNxQypdGd2x36djCCo1VQAn"},
	}, {
		// ---------------------------------------------------------------------
		// Negative stake submission change P2SH tests.
		// ---------------------------------------------------------------------

		name:     "almost v0 stake change p2sh -- wrong hash length",
		script:   p("SSTXCHANGE HASH160 DATA_21 0x00%s EQUAL", p2sh),
		params:   mainNetParams,
		wantType: STNonStandard,
	}, {
		// ---------------------------------------------------------------------
		// Positive stake submission change P2SH tests.
		// ---------------------------------------------------------------------

		name:      "mainnet v0 stake change p2sh",
		script:    p("SSTXCHANGE HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    mainNetParams,
		wantType:  STStakeChangeScriptHash,
		wantAddrs: []string{"2caH4eu9E49X8J4xPzrYWrWiVGoN5opNizFk"},
	}, {
		name:      "testnet v0 stake change p2sh",
		script:    p("SSTXCHANGE HASH160 DATA_20 0x%s EQUAL", p2sh),
		params:    testNetParams,
		wantType:  STStakeChangeScriptHash,
		wantAddrs: []string{"TcvALEAYZsT2PJqowebx2Yedhaza6c8vzfg"},
	}}
}()
