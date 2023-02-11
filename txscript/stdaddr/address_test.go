// Copyright (c) 2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package stdaddr

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/EXCCoin/base58"
	"github.com/EXCCoin/exccd/crypto/ripemd160"
	"github.com/EXCCoin/exccd/dcrec/edwards/v2"
	"github.com/EXCCoin/exccd/dcrec/secp256k1/v4"
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
// the tests.  They match the Decred testnet params as of the time this comment
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

// mockRegNetParams returns mock regression test address parameters to use
// throughout the tests.  They match the Decred regnet params as of the time
// this comment was written.
func mockRegNetParams() *mockAddrParams {
	return &mockAddrParams{
		pubKeyID:     [2]byte{0x25, 0xe5}, // starts with Rk
		pkhEcdsaID:   [2]byte{0x0e, 0x00}, // starts with Rs
		pkhEd25519ID: [2]byte{0x0d, 0xe0}, // starts with Re
		pkhSchnorrID: [2]byte{0x0d, 0xc2}, // starts with RS
		scriptHashID: [2]byte{0x0d, 0xdb}, // starts with Rc
		privKeyID:    0xef,                // starts with 9 (uncompressed) or c (compressed)
	}
}

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// TestAddresses ensures that address-related APIs work as intended including
// that they are properly encoded and decoded, that they produce the expected
// payment-related scripts, and that error paths fail as expected.  For
// addresses that implement the stake address interface, the stake-related
// scripts are also tested.
func TestAddresses(t *testing.T) {
	mainNetParams := mockMainNetParams()
	testNetParams := mockTestNetParams()
	regNetParams := mockRegNetParams()

	type newAddrFn func() (Address, error)
	tests := []struct {
		name         string        // test description
		makeAddr     newAddrFn     // function to construct new address via API
		makeErr      error         // expected error from new address function
		addr         string        // expected address and address to decode
		net          AddressParams // params for network
		decodeErr    error         // expected error from decode
		version      uint16        // expected scripts version
		payScript    string        // hex-encoded expected payment script
		voteScript   string        // hex-encoded expected voting rights script
		rewardAmount int64         // reward commitment amount
		voteFeeLimit int64         // reward commitment vote fee limit
		revFeeLimit  int64         // reward commitment revoke fee limit
		rewardScript string        // hex-encoded expected reward commitment script
		changeScript string        // hex-encoded expected stake change script
		commitScript string        // hex-encoded expected vote commitment script
		revokeScript string        // hex-encoded expected revoke commitment script
		trsyScript   string        // hex-encoded expected pay from treasury script
		pubKey       string        // hex-encoded expected public key
	}{{
		// ---------------------------------------------------------------------
		// Misc decoding error tests.
		// ---------------------------------------------------------------------

		name:      "bad checksum",
		addr:      "TsmWaPM77WSyA3aiQ2Q1KnwGDVWvEkhip23",
		net:       testNetParams,
		decodeErr: ErrBadAddressChecksum,
	}, {
		name:      "parse valid mainnet address with testnet rejected",
		addr:      "2ca8LhcQmHsRMNnnJfXK6yR4gKNysttCyc5b",
		net:       testNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		name:      "mainnet p2pk with no data for pubkey",
		addr:      "Aiz5jz1s",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		name:      "invalid base58 (l not in base58 alphabet)",
		addr:      "2cl8LhcQmHsRMNnnJfXK6yR4gKNysttCyc5b",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pk-ecdsa-secp256k1 uncompressed (0x04) rejected via constructor",
		makeAddr: func() (Address, error) {
			pkHex := "0464c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b163" +
				"74dace6778f0b87ca4279b565d2130ce59f75bfbb2b88da794143d7cfd3" +
				"e80808a1fa3203904"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKeyFormat,
	}, {
		name:      "mainnet p2pk-ecdsa-secp256k1 uncompressed (0x04) rejected via decode",
		addr:      "HiQeNVx8PNYP8ysyunUoicyNdfRUrEu1kzPE6v5gECBHBYgDzXCg8BsDGjmaHCpV97ytaQGHz5XDMJgJVHjv9YeSXWkHfwmBJj",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		name: "mainnet p2pk-ecdsa-secp256k1 hybrid (0x06) rejected via constructor",
		makeAddr: func() (Address, error) {
			pkHex := "0664c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b163" +
				"74dace6778f0b87ca4279b565d2130ce59f75bfbb2b88da794143d7cfd3" +
				"e80808a1fa3203904"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKeyFormat,
	}, {
		name: "mainnet p2pk-ecdsa-secp256k1 hybrid (0x07) rejected via constructor",
		makeAddr: func() (Address, error) {
			pkHex := "07348d8aeb4253ca52456fe5da94ab1263bfee16bb8192497f6663" +
				"89ca964f84798375129d7958843b14258b905dc94faed324dd8a9d67ffa" +
				"c8cc0a85be84bac5d"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKeyFormat,
	}, {
		name: "p2pk-ecdsa-secp256k1 unsupported script version",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(9999, pk, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2pk-ecdsa-secp256k1 unsupported script version via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEcdsaSecp256k1(9999, pk, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2pk-ecdsa-secp256k1 malformed pubkey",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKey,
	}, {
		name:      "p2pk-ecdsa-secp256k1 malformed pubkey via decode",
		addr:      "3tWTcxjUnAKTzHh8pHPYpSsUKVbTvziNGHtbBFQkY12khQWuW83p",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PK ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pk-ecdsa-secp256k1 compressed (0x02)",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHTeeDBQu8wKuzvnWyaShfJkRLGTExvwZDnjCDbVC5K5LRgBDfvH",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2edac",
		pubKey:    "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed",
	}, {
		name: "mainnet p2pk-ecdsa-secp256k1 compressed (0x03)",
		makeAddr: func() (Address, error) {
			pkHex := "03e925aafc1edd44e7c7f1ea4fb7d265dc672f204c3d0c81930389c10b81fb75de"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sMm9V9S5uoBtwXfeNzoy2sQiUMxkCeeP8H5uGBgFkn3PR7EfDHKS",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "2103e925aafc1edd44e7c7f1ea4fb7d265dc672f204c3d0c81930389c10b81fb75deac",
		pubKey:    "03e925aafc1edd44e7c7f1ea4fb7d265dc672f204c3d0c81930389c10b81fb75de",
	}, {
		name: "mainnet p2pk-ecdsa-secp256k1 compressed via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEcdsaSecp256k1(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHTeeDBQu8wKuzvnWyaShfJkRLGTExvwZDnjCDbVC5K5LRgBDfvH",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2edac",
		pubKey:    "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed",
	}, {
		name: "mainnet p2pk-ecdsa-secp256k1 compressed from uncompressed via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "0464c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b163" +
				"74dace6778f0b87ca4279b565d2130ce59f75bfbb2b88da794143d7cfd3" +
				"e80808a1fa3203904"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEcdsaSecp256k1(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHTKu5sVGULAXvFymMGbsdpAaevydey5ReEhkahW5n4Cav6jxEAT",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "210264c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b16374dace6778f0ac",
		pubKey:    "0264c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b16374dace6778f0",
	}, {
		name: "testnet p2pk-ecdsa-secp256k1 compressed (0x02)",
		makeAddr: func() (Address, error) {
			pkHex := "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, testNetParams)
		},
		makeErr:   nil,
		addr:      "TkKmMiY5iDh4U3KkSopYgkU1AzhAcQZiSoVhYhFymZHGMi9JQkuDP",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06eac",
		pubKey:    "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e",
	}, {
		name: "testnet p2pk-ecdsa-secp256k1 compressed (0x03)",
		makeAddr: func() (Address, error) {
			pkHex := "030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af5"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEcdsaSecp256k1Raw(0, pk, testNetParams)
		},
		makeErr:   nil,
		addr:      "TkQ3RrFierkUUbgipYwgeVfV8ch3fktfrDamGyDYESPBXMaUxJmcc",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af5ac",
		pubKey:    "030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af5",
	}, {
		name: "testnet p2pk-ecdsa-secp256k1 compressed via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEcdsaSecp256k1(0, pk, testNetParams)
		},
		makeErr:   nil,
		addr:      "TkKmMiY5iDh4U3KkSopYgkU1AzhAcQZiSoVhYhFymZHGMi9JQkuDP",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06eac",
		pubKey:    "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e",
	}, {
		name: "testnet p2pk-ecdsa-secp256k1 compressed from uncompressed via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "046a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e12" +
				"20eacf4be06ed548c8c16fb5eb9007cb94220b3bb89491d5a1fd2d77867" +
				"fca64217acecf2244"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEcdsaSecp256k1(0, pk, testNetParams)
		},
		makeErr:   nil,
		addr:      "TkKmMiY5iDh4U3KkSopYgkU1AzhAcQZiSoVhYhFymZHGMi9JQkuDP",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06eac",
		pubKey:    "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e",
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Ed25519 tests.
		// ---------------------------------------------------------------------

		name: "p2pk-ed25519 unsupported script version",
		makeAddr: func() (Address, error) {
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEd25519Raw(9999, pk, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2pk-ed25519 unsupported script version via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk, err := edwards.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEd25519(9999, pk, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2pk-ed25519 malformed pubkey",
		makeAddr: func() (Address, error) {
			return NewAddressPubKeyEd25519Raw(0, nil, mainNetParams)
		},
		makeErr: ErrInvalidPubKey,
	}, {
		name:      "p2pk-ed25519 malformed pubkey (only 31 bytes) via decode",
		addr:      "3tWUQtEa3P4SDQwjER81wkTxe4kiYLgNAso3pt2X5k3NFHRVQeNv",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PK Ed25519 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pk-ed25519",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEd25519Raw(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHW5LfhGgZokqA4SzE9UUNQpuzuY4g1J8aoTGPhuHeNSsnesn9Dw",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "20cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc51be",
		pubKey:    "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
	}, {
		name: "mainnet p2pk-ed25519 via concrete constructor",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk, err := edwards.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEd25519(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHW5LfhGgZokqA4SzE9UUNQpuzuY4g1J8aoTGPhuHeNSsnesn9Dw",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "20cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc51be",
		pubKey:    "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
	}, {
		name: "testnet p2pk-ed25519",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEd25519Raw(0, pk, testNetParams)
		},
		makeErr:   nil,
		addr:      "TkKp4jynaSAyyV5FooNX3UBGzeXhxYq7e96YtjbRS5XEaar95isMx",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "20cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc51be",
		pubKey:    "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
	}, {
		name: "testnet p2pk-ed25519 via concrete constructor",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk, err := edwards.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeyEd25519(0, pk, testNetParams)
		},
		makeErr:   nil,
		addr:      "TkKp4jynaSAyyV5FooNX3UBGzeXhxYq7e96YtjbRS5XEaar95isMx",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "20cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc51be",
		pubKey:    "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
	}, {
		name: "regnet p2pk-ed25519",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			pkHex := "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeyEd25519Raw(0, pk, regNetParams)
		},
		makeErr:   nil,
		addr:      "Rk44TM8ZWqLsuaLr37pH7jNvB31oEPuzGBrvSvZ729Qs9Gfp3Stvx",
		net:       regNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "20cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc51be",
		pubKey:    "cecc1507dc1ddd7295951c290888f095adb9044d1b73d696e6df065d683bd4fc",
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pk-schnorr-secp256k1 uncompressed (0x04) rejected via constructor",
		makeAddr: func() (Address, error) {
			pkHex := "0464c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b163" +
				"74dace6778f0b87ca4279b565d2130ce59f75bfbb2b88da794143d7cfd3" +
				"e80808a1fa3203904"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeySchnorrSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKeyFormat,
	}, {
		name:      "mainnet p2pk-schnorr-secp256k1 uncompressed (0x04) rejected via decode",
		addr:      "HiQjU9uCJtiQD7osQuYHWJRFiBCTuqtaTw8QFMtMgAW2ny4nUENeXDiV5VxfVZrK6PZynKPDpL7bwc6XLFNpV8k7ePDJmkkVCh",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		name: "mainnet p2pk-schnorr-secp256k1 hybrid (0x06) rejected via constructor",
		makeAddr: func() (Address, error) {
			pkHex := "0664c44653d6567eff5753c5d24a682ddc2b2cadfe1b0c6433b163" +
				"74dace6778f0b87ca4279b565d2130ce59f75bfbb2b88da794143d7cfd3" +
				"e80808a1fa3203904"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeySchnorrSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKeyFormat,
	}, {
		name: "mainnet p2pk-schnorr-secp256k1 hybrid (0x07) rejected via constructor",
		makeAddr: func() (Address, error) {
			pkHex := "07348d8aeb4253ca52456fe5da94ab1263bfee16bb8192497f6663" +
				"89ca964f84798375129d7958843b14258b905dc94faed324dd8a9d67ffa" +
				"c8cc0a85be84bac5d"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeySchnorrSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKeyFormat,
	}, {
		name: "p2pk-schnorr-secp256k1 unsupported script version",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeySchnorrSecp256k1Raw(9999, pk, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2pk-schnorr-secp256k1 unsupported script version via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeySchnorrSecp256k1(9999, pk, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2pk-schnorr-secp256k1 malformed pubkey",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2"
			pk := hexToBytes(pkHex)
			return NewAddressPubKeySchnorrSecp256k1Raw(0, pk, mainNetParams)
		},
		makeErr: ErrInvalidPubKey,
	}, {
		name:      "p2pk-schnorr-secp256k1 malformed pubkey via decode",
		addr:      "3tWUW3oD87XtmFVGnLX4Z3Hdesm2qRvvN8H5kq3yXxJumCxbvCpo",
		net:       mainNetParams,
		decodeErr: ErrUnsupportedAddress,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PK Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet p2pk-schnorr-secp256k1 compressed (0x02)",
		addr:      "22tvQL1muRkfgEcg1tx8Lq9KRLNiGiJaRCNx",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a91473f13c131565ffc78d9acd9f8060b866a36941fc88ac",
		pubKey:    "0244bc9272edeb558510b7c3dcbb3f710dd08dab985bb0b7945696ced779f8ed90",
	}, {
		name:      "mainnet p2pk-schnorr-secp256k1 compressed (0x03)",
		addr:      "DkRQx3y6YoJPnMKom23nuDFdfhmEnu8oDLTp4YVyWC6RjND19UxHk",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "2103e925aafc1edd44e7c7f1ea4fb7d265dc672f204c3d0c81930389c10b81fb75de52be",
		pubKey:    "03e925aafc1edd44e7c7f1ea4fb7d265dc672f204c3d0c81930389c10b81fb75de",
	}, {
		name: "mainnet p2pk-schnorr-secp256k1 compressed via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeySchnorrSecp256k1(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHXY8ZeJxqngNGZwvnjghZwwyAVggGLdJo4sQhtrz4d67QqyDiR6",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed52be",
		pubKey:    "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed",
	}, {
		name: "mainnet p2pk-schnorr-secp256k1 compressed from uncompressed via concrete constructor",
		makeAddr: func() (Address, error) {
			pkHex := "048f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da1" +
				"43a02b0fe2ed91badd9f6403cc485dc3d5ca83ee8917dca57414866b083" +
				"087c1c83b7a8e3304"
			pk, err := secp256k1.ParsePubKey(hexToBytes(pkHex))
			if err != nil {
				return nil, err
			}
			return NewAddressPubKeySchnorrSecp256k1(0, pk, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2sHXY8ZeJxqngNGZwvnjghZwwyAVggGLdJo4sQhtrz4d67QqyDiR6",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed52be",
		pubKey:    "028f53838b7639563f27c94845549a41e5146bcd52e7fef0ea6da143a02b0fe2ed",
	}, {
		name:      "testnet p2pk-schnorr-secp256k1 compressed (0x02)",
		addr:      "TkKqFCtYcHPupVbPcDdhvkNeNYXPqqs88Z4ygukH9MGaNV8e1WWhX",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e52be",
		pubKey:    "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e",
	}, {
		name:      "testnet p2pk-schnorr-secp256k1 compressed (0x03)",
		addr:      "TkQ7KLcBYvTKq3xMyxkqtVa8LAXGuCC5XyA3RBhqcENVY8ZiDjuAB",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af552be",
		pubKey:    "030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af5",
	}, {
		name:      "regnet p2pk-schnorr-secp256k1 compressed (0x02)",
		addr:      "Rk45dp3KYgZokaryqY5U11aHYw1V7gwzkbqMF6hxjRACwAxDivM6N",
		net:       regNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e52be",
		pubKey:    "026a40c403e74670c4de7656a09caa2353d4b383a9ce66eef51e1220eacf4be06e",
	}, {
		name:      "regnet p2pk-schnorr-secp256k1 compressed (0x03)",
		addr:      "Rk8MhwkxVKdDm9DxDHCbxkmmWZ1NB3GxA1vQyNfXCJG86pPPnnn9M",
		net:       regNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "21030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af552be",
		pubKey:    "030844ee70d8384d5250e9bb3a6a73d4b5bec770e8b31d6a0ae9fb739009d91af5",
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "p2pkh-ecdsa-secp256k1 wrong hash length",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("000ef030107fd26e0b6bf40512bca2ceb1dd80adaa")
			return NewAddressPubKeyHashEcdsaSecp256k1(0, hash, mainNetParams)
		},
		makeErr: ErrInvalidHashLen,
	}, {
		name: "p2pkh-ecdsa-secp256k1 unsupported script version",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("2789d58cfa0957d206f025c2af056fc8a77cebb0")
			return NewAddressPubKeyHashEcdsaSecp256k1(9999, hash, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PKH ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pkh-ecdsa-secp256k1",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("2789d58cfa0957d206f025c2af056fc8a77cebb0")
			return NewAddressPubKeyHashEcdsaSecp256k1(0, hash, mainNetParams)
		},
		makeErr:      nil,
		addr:         "22toSLnoc4sfmoP3h7ZCV5jJQwtp58GAnq4s",
		net:          mainNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "76a9142789d58cfa0957d206f025c2af056fc8a77cebb088ac",
		voteScript:   "ba76a9142789d58cfa0957d206f025c2af056fc8a77cebb088ac",
		rewardAmount: 1e8,
		revFeeLimit:  16777216,
		rewardScript: "6a1e2789d58cfa0957d206f025c2af056fc8a77cebb000e1f505000000000058",
		changeScript: "bd76a9142789d58cfa0957d206f025c2af056fc8a77cebb088ac",
		commitScript: "bb76a9142789d58cfa0957d206f025c2af056fc8a77cebb088ac",
		revokeScript: "bc76a9142789d58cfa0957d206f025c2af056fc8a77cebb088ac",
		trsyScript:   "bc76a9142789d58cfa0957d206f025c2af056fc8a77cebb088ac",
	}, {
		name: "mainnet p2pkh-ecdsa-secp256k1 2",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("229ebac30efd6a69eec9c1a48e048b7c975c25f2")
			return NewAddressPubKeyHashEcdsaSecp256k1(0, hash, mainNetParams)
		},
		makeErr:      nil,
		addr:         "22tnzLSgPpDjNzWBdLN3xADdW5FDSBoZvZJP",
		net:          mainNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "76a914229ebac30efd6a69eec9c1a48e048b7c975c25f288ac",
		voteScript:   "ba76a914229ebac30efd6a69eec9c1a48e048b7c975c25f288ac",
		rewardAmount: 9556193632,
		revFeeLimit:  33554432,
		rewardScript: "6a1e229ebac30efd6a69eec9c1a48e048b7c975c25f260f19739020000000059",
		changeScript: "bd76a914229ebac30efd6a69eec9c1a48e048b7c975c25f288ac",
		commitScript: "bb76a914229ebac30efd6a69eec9c1a48e048b7c975c25f288ac",
		revokeScript: "bc76a914229ebac30efd6a69eec9c1a48e048b7c975c25f288ac",
		trsyScript:   "bc76a914229ebac30efd6a69eec9c1a48e048b7c975c25f288ac",
	}, {
		name: "testnet p2pkh-ecdsa-secp256k1",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("f15da1cb8d1bcb162c6ab446c95757a6e791c916")
			return NewAddressPubKeyHashEcdsaSecp256k1(0, hash, testNetParams)
		},
		makeErr:      nil,
		addr:         "Tso2MVTUeVrjHTBFedFhiyM7yVTbifRSmQJ",
		net:          testNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		voteScript:   "ba76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		rewardAmount: 2428220961,
		revFeeLimit:  16777216,
		rewardScript: "6a1ef15da1cb8d1bcb162c6ab446c95757a6e791c91621b6bb90000000000058",
		changeScript: "bd76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		commitScript: "bb76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		revokeScript: "bc76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		trsyScript:   "bc76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
	}, {
		name: "regnet p2pkh-sep256k1-ecdsa",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("f15da1cb8d1bcb162c6ab446c95757a6e791c916")
			return NewAddressPubKeyHashEcdsaSecp256k1(0, hash, regNetParams)
		},
		makeErr:      nil,
		addr:         "RsWM2w5LPJip56uxcZ1Scq7Tcbg97FQcqVC",
		net:          regNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		voteScript:   "ba76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		rewardAmount: 2428220961,
		revFeeLimit:  16777216,
		rewardScript: "6a1ef15da1cb8d1bcb162c6ab446c95757a6e791c91621b6bb90000000000058",
		changeScript: "bd76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		commitScript: "bb76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		revokeScript: "bc76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
		trsyScript:   "bc76a914f15da1cb8d1bcb162c6ab446c95757a6e791c91688ac",
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH ed25519 tests.
		// ---------------------------------------------------------------------

		name: "p2pkh-ed25519 wrong hash length",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("000ef030107fd26e0b6bf40512bca2ceb1dd80adaa")
			return NewAddressPubKeyHashEd25519(0, hash, mainNetParams)
		},
		makeErr: ErrInvalidHashLen,
	}, {
		name: "p2pkh-ed25519 unsupported script version",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("0ef030107fd26e0b6bf40512bca2ceb1dd80adaa")
			return NewAddressPubKeyHashEd25519(9999, hash, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PKH Ed25519 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pkh-ed25519",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			hash := hexToBytes("456d8ee57a4b9121987b4ecab8c3bcb5797e8a53")
			return NewAddressPubKeyHashEd25519(0, hash, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2eZt5xAGLspEZeTLF29nVJsYW4nbR6jP9SCJ",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a914456d8ee57a4b9121987b4ecab8c3bcb5797e8a538851be",
	}, {
		name: "mainnet p2pkh-ed25519 2",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...02.
			hash := hexToBytes("09788a8dcb216efa354e487d57d76255b1af4320")
			return NewAddressPubKeyHashEd25519(0, hash, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2eZncvmsRz8mFie7HVfi6E37ZrFsbG66xzjm",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a91409788a8dcb216efa354e487d57d76255b1af43208851be",
	}, {
		name: "testnet p2pkh-ed25519",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			hash := hexToBytes("456d8ee57a4b9121987b4ecab8c3bcb5797e8a53")
			return NewAddressPubKeyHashEd25519(0, hash, testNetParams)
		},
		makeErr:   nil,
		addr:      "TeeXvqZJrc7KnFZCT27fHfzcrTTzSHCgpHu",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a914456d8ee57a4b9121987b4ecab8c3bcb5797e8a538851be",
	}, {
		name: "regnet p2pkh-ed25519",
		makeAddr: func() (Address, error) {
			// From pubkey for privkey 0x00...01.
			hash := hexToBytes("456d8ee57a4b9121987b4ecab8c3bcb5797e8a53")
			return NewAddressPubKeyHashEd25519(0, hash, regNetParams)
		},
		makeErr:   nil,
		addr:      "ReMrcHBAbQyQZuHuQwsQBXkxVZgXps3tLEw",
		net:       regNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a914456d8ee57a4b9121987b4ecab8c3bcb5797e8a538851be",
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PKH Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "p2pkh-schnorr-secp256k1 wrong hash length",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("000ef030107fd26e0b6bf40512bca2ceb1dd80adaa")
			return NewAddressPubKeyHashSchnorrSecp256k1(0, hash, mainNetParams)
		},
		makeErr: ErrInvalidHashLen,
	}, {
		name: "p2pkh-schnorr-secp256k1 unsupported script version",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("0ef030107fd26e0b6bf40512bca2ceb1dd80adaa")
			return NewAddressPubKeyHashSchnorrSecp256k1(9999, hash, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2PKH Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2pkh-schnorr-secp256k1",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("2789d58cfa0957d206f025c2af056fc8a77cebb0")
			return NewAddressPubKeyHashSchnorrSecp256k1(0, hash, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2SZieQMJd5RCqHmwQh4exWH77DMy9dyfKv1y",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a9142789d58cfa0957d206f025c2af056fc8a77cebb08852be",
	}, {
		name: "mainnet p2pkh-schnorr-secp256k1 2",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("229ebac30efd6a69eec9c1a48e048b7c975c25f2")
			return NewAddressPubKeyHashSchnorrSecp256k1(0, hash, mainNetParams)
		},
		makeErr:   nil,
		addr:      "2SZiCQ1BQpmGSUu5LusWRamSCLiNWhZ3RaVv",
		net:       mainNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a914229ebac30efd6a69eec9c1a48e048b7c975c25f28852be",
	}, {
		name: "testnet p2pkh-schnorr-secp256k1",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("f15da1cb8d1bcb162c6ab446c95757a6e791c916")
			return NewAddressPubKeyHashSchnorrSecp256k1(0, hash, testNetParams)
		},
		makeErr:   nil,
		addr:      "TSr4xSiznUfzxkJcH7F3xuaFCUBdEegjq5H",
		net:       testNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a914f15da1cb8d1bcb162c6ab446c95757a6e791c9168852be",
	}, {
		name: "regnet p2pkh-sep256k1-schnorr",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("f15da1cb8d1bcb162c6ab446c95757a6e791c916")
			return NewAddressPubKeyHashSchnorrSecp256k1(0, hash, regNetParams)
		},
		makeErr:   nil,
		addr:      "RSZPdtLrXHY5kQ3KF2znrmLaqaQAd7XamRg",
		net:       regNetParams,
		decodeErr: nil,
		version:   0,
		payScript: "76a914f15da1cb8d1bcb162c6ab446c95757a6e791c9168852be",
	}, {
		// ---------------------------------------------------------------------
		// Negative P2SH tests.
		// ---------------------------------------------------------------------

		name: "p2sh wrong hash length",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("00f815b036d9bbbce5e9f2a00abd1bf3dc91e95510")
			return NewAddressScriptHashFromHash(0, hash, mainNetParams)
		},
		makeErr: ErrInvalidHashLen,
	}, {
		name: "p2sh from script hash unsupported script version",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("f815b036d9bbbce5e9f2a00abd1bf3dc91e95510")
			return NewAddressScriptHashFromHash(9999, hash, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		name: "p2sh from redeem script unsupported script version",
		makeAddr: func() (Address, error) {
			script := hexToBytes("a9148d5923973651abb5709a15203f7f603d337d016287")
			return NewAddressScriptHash(9999, script, mainNetParams)
		},
		makeErr: ErrUnsupportedScriptVersion,
	}, {
		// ---------------------------------------------------------------------
		// Positive P2SH tests.
		// ---------------------------------------------------------------------

		name: "mainnet p2sh",
		makeAddr: func() (Address, error) {
			script := hexToBytes("512103aa43f0a6c15730d886cc1f0342046d2017548" +
				"3d90d7ccb657f90c489111d794c51ae")
			return NewAddressScriptHash(0, script, mainNetParams)
		},
		makeErr:      nil,
		addr:         "2ca7JWn4BRnF5PnGB9rRABP4T3JufyrjFybG",
		net:          mainNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "a9148d5923973651abb5709a15203f7f603d337d016287",
		rewardAmount: 1e8,
		revFeeLimit:  16777216,
		rewardScript: "6a1e8d5923973651abb5709a15203f7f603d337d016200e1f505000000800058",
		voteScript:   "baa9148d5923973651abb5709a15203f7f603d337d016287",
		changeScript: "bda9148d5923973651abb5709a15203f7f603d337d016287",
		commitScript: "bba9148d5923973651abb5709a15203f7f603d337d016287",
		revokeScript: "bca9148d5923973651abb5709a15203f7f603d337d016287",
		trsyScript:   "bca9148d5923973651abb5709a15203f7f603d337d016287",
	}, {
		name: "mainnet p2sh 2",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("c7da5095683436f4435fc4e7163dcafda1a2d007")
			return NewAddressScriptHashFromHash(0, hash, mainNetParams)
		},
		makeErr:      nil,
		addr:         "2caCdriUF5h2pv9Emw5mtcSA4SNfbiCciheY",
		net:          mainNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "a914c7da5095683436f4435fc4e7163dcafda1a2d00787",
		voteScript:   "baa914c7da5095683436f4435fc4e7163dcafda1a2d00787",
		rewardAmount: 9556193632,
		revFeeLimit:  33554432,
		rewardScript: "6a1ec7da5095683436f4435fc4e7163dcafda1a2d00760f19739020000800059",
		changeScript: "bda914c7da5095683436f4435fc4e7163dcafda1a2d00787",
		commitScript: "bba914c7da5095683436f4435fc4e7163dcafda1a2d00787",
		revokeScript: "bca914c7da5095683436f4435fc4e7163dcafda1a2d00787",
		trsyScript:   "bca914c7da5095683436f4435fc4e7163dcafda1a2d00787",
	}, {
		name: "testnet p2sh",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("36c1ca10a8a6a4b5d4204ac970853979903aa284")
			return NewAddressScriptHashFromHash(0, hash, testNetParams)
		},
		makeErr:      nil,
		addr:         "TccWLgcquqvwrfBocq5mcK5kBiyw8S5DsVb",
		net:          testNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "a91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		voteScript:   "baa91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		rewardAmount: 2428220961,
		revFeeLimit:  16777216,
		rewardScript: "6a1e36c1ca10a8a6a4b5d4204ac970853979903aa28421b6bb90000000800058",
		changeScript: "bda91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		commitScript: "bba91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		revokeScript: "bca91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		trsyScript:   "bca91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
	}, {
		name: "regnet p2sh",
		makeAddr: func() (Address, error) {
			hash := hexToBytes("36c1ca10a8a6a4b5d4204ac970853979903aa284")
			return NewAddressScriptHashFromHash(0, hash, regNetParams)
		},
		makeErr:      nil,
		addr:         "RcKq28Eheeo2eJvWakqWWAr5pqCUWwm9hvG",
		net:          regNetParams,
		decodeErr:    nil,
		version:      0,
		payScript:    "a91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		voteScript:   "baa91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		rewardAmount: 2428220961,
		revFeeLimit:  16777216,
		rewardScript: "6a1e36c1ca10a8a6a4b5d4204ac970853979903aa28421b6bb90000000800058",
		changeScript: "bda91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		commitScript: "bba91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		revokeScript: "bca91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
		trsyScript:   "bca91436c1ca10a8a6a4b5d4204ac970853979903aa28487",
	}}

	for _, test := range tests {
		// Create address from test constructor and ensure it produces the
		// expected encoded address when the constructor is specified.
		if test.makeAddr != nil {
			addr, err := test.makeAddr()
			if !errors.Is(err, test.makeErr) {
				t.Errorf("%s: mismatched err -- got %v, want %v", test.name, err,
					test.makeErr)
				continue
			}
			if err != nil {
				continue
			}

			// Ensure encoding the address is the same as the original.
			encoded := addr.String()
			if encoded != test.addr {
				t.Errorf("%s: unexpected address -- got %v, want %v", test.name,
					encoded, test.addr)
				continue
			}
		}

		// Decode address and ensure the expected error is received.
		decodedAddr, err := DecodeAddress(test.addr, test.net)
		if !errors.Is(err, test.decodeErr) {
			t.Errorf("%s: mismatched err -- got %v, want %v", test.name, err,
				test.decodeErr)
			continue
		}
		if err != nil {
			continue
		}

		// Ensure the payment script version and contents are the expected
		// values.
		wantPayScript, err := hex.DecodeString(test.payScript)
		if err != nil {
			t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
			continue
		}
		gotPayScriptVersion, gotPayScript := decodedAddr.PaymentScript()
		if gotPayScriptVersion != test.version {
			t.Errorf("%s: mismatched payment script version -- got %d, want %d",
				test.name, gotPayScriptVersion, test.version)
			continue
		}
		if !bytes.Equal(gotPayScript, wantPayScript) {
			t.Errorf("%s: mismatched payment script -- got %x, want %x",
				test.name, gotPayScript, wantPayScript)
			continue
		}

		// Ensure stake-specific interface results produce the expected values.
		if stakeAddr, ok := decodedAddr.(StakeAddress); ok {
			// Ensure the voting rights script version and contents are the
			// expected values.
			wantScript, err := hex.DecodeString(test.voteScript)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			gotScriptVer, gotScript := stakeAddr.VotingRightsScript()
			if gotScriptVer != test.version {
				t.Errorf("%s: mismatched voting rights script version -- got "+
					"%d, want %d", test.name, gotScriptVer,
					test.version)
				continue
			}
			if !bytes.Equal(gotScript, wantScript) {
				t.Errorf("%s: mismatched voting rights script -- got %x, want %x",
					test.name, gotScript, wantScript)
				continue
			}

			// Ensure the reward commitment script version and contents are the
			// expected values.
			wantScript, err = hex.DecodeString(test.rewardScript)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			gotScriptVer, gotScript = stakeAddr.RewardCommitmentScript(
				test.rewardAmount, test.voteFeeLimit, test.revFeeLimit)
			if gotScriptVer != test.version {
				t.Errorf("%s: mismatched reward cmt script version -- got %d, "+
					"want %d", test.name, gotScriptVer, test.version)
				continue
			}
			if !bytes.Equal(gotScript, wantScript) {
				t.Errorf("%s: mismatched reward cmt script -- got %x, want %x",
					test.name, gotScript, wantScript)
				continue
			}

			// Ensure the stake change script version and contents are the
			// expected values.
			wantScript, err = hex.DecodeString(test.changeScript)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			gotScriptVer, gotScript = stakeAddr.StakeChangeScript()
			if gotScriptVer != test.version {
				t.Errorf("%s: mismatched change script version -- got %d, "+
					"want %d", test.name, gotScriptVer, test.version)
				continue
			}
			if !bytes.Equal(gotScript, wantScript) {
				t.Errorf("%s: mismatched change script -- got %x, want %x",
					test.name, gotScript, wantScript)
				continue
			}

			// Ensure the vote commitment script version and contents are the
			// expected values.
			wantScript, err = hex.DecodeString(test.commitScript)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			gotScriptVer, gotScript = stakeAddr.PayVoteCommitmentScript()
			if gotScriptVer != test.version {
				t.Errorf("%s: mismatched vote commit script version -- got %d, "+
					"want %d", test.name, gotScriptVer, test.version)
				continue
			}
			if !bytes.Equal(gotScript, wantScript) {
				t.Errorf("%s: mismatched vote commit script -- got %x, want %x",
					test.name, gotScript, wantScript)
				continue
			}

			// Ensure the revoke commitment script version and contents are the
			// expected values.
			wantScript, err = hex.DecodeString(test.revokeScript)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			gotScriptVer, gotScript = stakeAddr.PayRevokeCommitmentScript()
			if gotScriptVer != test.version {
				t.Errorf("%s: mismatched revoke cmt script version -- got %d, "+
					"want %d", test.name, gotScriptVer, test.version)
				continue
			}
			if !bytes.Equal(gotScript, wantScript) {
				t.Errorf("%s: mismatched revoke cmt script -- got %x, want %x",
					test.name, gotScript, wantScript)
				continue
			}

			// Ensure the pay from treasury script version and contents are the
			// expected values.
			wantScript, err = hex.DecodeString(test.trsyScript)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			if !bytes.Equal(gotScript, wantScript) {
				t.Errorf("%s: mismatched treasury change script -- got %x, "+
					"want %x", test.name, gotScript, wantScript)
				continue
			}
		}

		// Ensure encoding the address is the same as the original.
		encoded := decodedAddr.String()
		if encoded != test.addr {
			t.Errorf("%s: decoding and encoding produced different addresses "+
				"-- got %v, want %v", test.name, encoded, test.addr)
			continue
		}

		// Ensure the method to get the serialized public key for the addresses
		// that support it returns the expected value.
		if pubKeyer, ok := decodedAddr.(SerializedPubKeyer); ok {
			wantPubKey, err := hex.DecodeString(test.pubKey)
			if err != nil {
				t.Errorf("%s: unexpected hex decode err: %v", test.name, err)
				continue
			}
			gotPubKey := pubKeyer.SerializedPubKey()
			if !bytes.Equal(gotPubKey, wantPubKey) {
				t.Errorf("%s: mismatched public key -- got %x, want %x",
					test.name, gotPubKey, wantPubKey)
				continue
			}
		}

		// Ensure the AddressPubKeyHash method for the address types that
		// support it returns the expected address.
		var pkhAddr Address
		var wantPkhAddr string
		switch a := decodedAddr.(type) {
		case *AddressPubKeyEcdsaSecp256k1V0:
			id := test.net.AddrIDPubKeyHashECDSAV0()
			wantPkhAddr = base58.CheckEncode(Hash160(a.serializedPubKey), id)
			pkhAddr = a.AddressPubKeyHash()

		case *AddressPubKeyEd25519V0:
			id := test.net.AddrIDPubKeyHashEd25519V0()
			wantPkhAddr = base58.CheckEncode(Hash160(a.serializedPubKey), id)
			pkhAddr = a.AddressPubKeyHash()

		case *AddressPubKeySchnorrSecp256k1V0:
			id := test.net.AddrIDPubKeyHashSchnorrV0()
			wantPkhAddr = base58.CheckEncode(Hash160(a.serializedPubKey), id)
			pkhAddr = a.AddressPubKeyHash()
		}
		if pkhAddr != nil {
			gotAddr := pkhAddr.String()
			if gotAddr != wantPkhAddr {
				t.Errorf("%s: mismatched pkh address -- got %s, want %s",
					test.name, gotAddr, wantPkhAddr)
			}
		}

		// Ensure the Hash160 method for the addresses that support it returns
		// the expected value.
		if h160er, ok := decodedAddr.(Hash160er); ok {
			decodedBytes := base58.Decode(test.addr)
			wantH160 := decodedBytes[2 : 2+ripemd160.Size]
			if gotH160 := h160er.Hash160()[:]; !bytes.Equal(gotH160, wantH160) {
				t.Errorf("%s: mismatched hash160 -- got %x, want %x", test.name,
					gotH160, wantH160)
				return
			}
		}
	}
}

// TestDecodeAddressV0Corners ensures that some additional errors that are
// specific to decoding version 0 addresses directly, as opposed to via the
// generic API, work as intended.  This is necessary because the generic address
// decoding function contains additional logic to avoid even attempting to
// decode addresses which can't possibly be one of the supported version 0
// address types, while the version 0 decoding logic specifically attempts to
// decode the address in order to provide more detailed errors.
func TestDecodeAddressV0Corners(t *testing.T) {
	mainNetParams := mockMainNetParams()

	tests := []struct {
		name      string        // test description
		addr      string        // expected address and address to decode
		net       AddressParams // params for network
		decodeErr error         // expected error from decode
	}{{
		// ---------------------------------------------------------------------
		// Misc decoding error tests.
		// ---------------------------------------------------------------------

		name:      "mainnet p2pk with no data for pubkey",
		addr:      "Aiz5jz1s",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddressData,
	}, {
		name:      "invalid base58 (l not in base58 alphabet)",
		addr:      "DsUZxxoHlSty8DCfwfartwTYbuhmVct7tJu",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddress,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK ECDSA secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet p2pk-ecdsa-secp256k1 uncompressed (0x04) rejected via decode",
		addr:      "HiQeNVx8PNYP8ysyunUoicyNdfRUrEu1kzPE6v5gECBHBYgDzXCg8BsDGjmaHCpV97ytaQGHz5XDMJgJVHjv9YeSXWkHfwmBJj",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddressData,
	}, {
		name:      "p2pk-ecdsa-secp256k1 malformed pubkey via decode",
		addr:      "3tWTcxjUnAKTzHh8pHPYpSsUKVbTvziNGHtbBFQkY12khQWuW83p",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddressData,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Ed25519 tests.
		// ---------------------------------------------------------------------

		name:      "p2pk-ed25519 malformed pubkey (only 31 bytes) via decode",
		addr:      "3tWUQtEa3P4SDQwjER81wkTxe4kiYLgNAso3pt2X5k3NFHRVQeNv",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddressData,
	}, {
		// ---------------------------------------------------------------------
		// Negative P2PK Schnorr secp256k1 tests.
		// ---------------------------------------------------------------------

		name:      "mainnet p2pk-schnorr-secp256k1 uncompressed (0x04) rejected via decode",
		addr:      "HiQjU9uCJtiQD7osQuYHWJRFiBCTuqtaTw8QFMtMgAW2ny4nUENeXDiV5VxfVZrK6PZynKPDpL7bwc6XLFNpV8k7ePDJmkkVCh",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddressData,
	}, {
		name:      "p2pk-schnorr-secp256k1 malformed pubkey via decode",
		addr:      "3tWUW3oD87XtmFVGnLX4Z3Hdesm2qRvvN8H5kq3yXxJumCxbvCpo",
		net:       mainNetParams,
		decodeErr: ErrMalformedAddressData,
	}}

	for _, test := range tests {
		_, err := DecodeAddressV0(test.addr, test.net)
		if !errors.Is(err, test.decodeErr) {
			t.Errorf("%s: mismatched err -- got %v, want %v", test.name, err,
				test.decodeErr)
			continue
		}
	}
}

// TestProbablyV0Base58Addr ensures the function that determines if an address
// is probably a base58 address works as intended by checking off by ones and
// ensuring all allowed characters in the modified base58 alphabet are accepted.
func TestProbablyV0Base58Addr(t *testing.T) {
	tests := []struct {
		name string // test description
		str  string // string to test
		want bool   // expected result
	}{{
		name: "all allowed base58 chars part 1",
		str:  "123456789ABCDEFGHJKLMNPQRSTUVWXYZab",
		want: true,
	}, {
		name: "all allowed base58 chars part 2",
		str:  "QRSTUVWXYZabcdefghijkmnopqrstuvwxyz",
		want: true,
	}, {
		name: "invalid base58 (0 not in base58 alphabet, one less than '1')",
		str:  "DsUZxxoH0Sty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 ({ not in base58 alphabet, one more than 'z')",
		str:  "DsUZxxoH{Sty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 (I not in base58 alphabet)",
		str:  "DsUZxxoHISty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 (O not in base58 alphabet)",
		str:  "DsUZxxoHOSty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 (l not in base58 alphabet)",
		str:  "DsUZxxoHlSty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 (: not in base58 alphabet, one more than '9')",
		str:  "DsUZxxoH:Sty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 (@ not in base58 alphabet, one less than 'A')",
		str:  "DsUZxxoH@Sty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 ([ not in base58 alphabet, one more than 'Z')",
		str:  "DsUZxxoH[Sty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}, {
		name: "invalid base58 (` not in base58 alphabet, one less than 'a')",
		str:  "DsUZxxoH`Sty8DCfwfartwTYbuhmVct7tJu",
		want: false,
	}}

	for _, test := range tests {
		got := probablyV0Base58Addr(test.str)
		if got != test.want {
			t.Errorf("%q: unexpected result -- got %v, want %v", test.name, got,
				test.want)
			continue
		}
	}
}
