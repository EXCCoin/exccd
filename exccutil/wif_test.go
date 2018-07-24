// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2013, 2014 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package exccutil_test

import (
	"testing"

	"github.com/EXCCoin/exccd/chaincfg"
	"github.com/EXCCoin/exccd/chaincfg/chainec"
	"github.com/EXCCoin/exccd/exccec/secp256k1"
	"github.com/EXCCoin/exccd/exccutil"
)

func TestEncodeDecodeUncompressedWIF(t *testing.T) {
	priv1, _ := secp256k1.PrivKeyFromBytes([]byte{
		0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
		0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
		0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
		0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

	priv2, _ := secp256k1.PrivKeyFromBytes([]byte{
		0xdd, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
		0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
		0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
		0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})

	wif1, err := exccutil.NewUncompressedWIF(priv1, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	wif2, err := exccutil.NewUncompressedWIF(priv2, &chaincfg.TestNetParams)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		wif     *exccutil.WIF
		encoded string
	}{
		{
			wif1,
			"5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ",
		},
		{
			wif2,
			"93GXcP5BqkAxXrV7N2EjmjAeGwfvrX7ALLXTnfZExxtn4QXMjjs",
		},
	}

	for _, test := range tests {
		// Test that encoding the WIF structure matches the expected string.
		s := test.wif.String()
		if s != test.encoded {
			t.Errorf("TestEncodeDecodePrivateKey failed: want '%s', got '%s'",
				test.encoded, s)
			continue
		}

		// Test that decoding the expected string results in the original WIF
		// structure.
		w, err := exccutil.DecodeWIF(test.encoded)
		if err != nil {
			t.Error(err)
			continue
		}
		if got := w.String(); got != test.encoded {
			t.Errorf("NewWIF failed: want '%v', got '%v'", test.wif, got)
		}
	}
}

func TestEncodeDecodeWIF(t *testing.T) {
	suites := []int{
		chainec.ECTypeSecp256k1,
		chainec.ECTypeEdwards,
		chainec.ECTypeSecSchnorr,
	}
	for _, ecType := range suites {
		var priv1, priv2 chainec.PrivateKey
		switch ecType {
		case chainec.ECTypeSecp256k1:
			priv1, _ = chainec.Secp256k1.PrivKeyFromBytes([]byte{
				0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
				0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
				0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
				0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

			priv2, _ = chainec.Secp256k1.PrivKeyFromBytes([]byte{
				0xdd, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
				0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
				0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
				0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})
		case chainec.ECTypeEdwards:
			priv1, _ = chainec.Edwards.PrivKeyFromScalar([]byte{
				0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
				0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
				0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
				0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

			priv2, _ = chainec.Edwards.PrivKeyFromScalar([]byte{
				0x0c, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
				0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
				0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
				0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})
		case chainec.ECTypeSecSchnorr:
			priv1, _ = chainec.SecSchnorr.PrivKeyFromBytes([]byte{
				0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
				0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
				0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
				0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

			priv2, _ = chainec.SecSchnorr.PrivKeyFromBytes([]byte{
				0xdd, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
				0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
				0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
				0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})
		}

		wif1, err := exccutil.NewWIF(priv1, &chaincfg.MainNetParams, ecType)
		if err != nil {
			t.Fatal(err)
		}
		wif2, err := exccutil.NewWIF(priv2, &chaincfg.TestNetParams, ecType)
		if err != nil {
			t.Fatal(err)
		}

		var tests []struct {
			wif     *exccutil.WIF
			encoded string
		}

		switch ecType {
		case chainec.ECTypeSecp256k1:
			tests = []struct {
				wif     *exccutil.WIF
				encoded string
			}{
				{
					wif1,
					"KwdMAjGmerYanjeui5SHS7JkmpZvVipYvB2LJGU1ZxJwYvP98617",
				},
				{
					wif2,
					"cV1Y7ARUr9Yx7BR55nTdnR7ZXNJphZtCCMBTEZBJe1hXt2kB684q",
				},
			}
		case chainec.ECTypeEdwards:
			tests = []struct {
				wif     *exccutil.WIF
				encoded string
			}{
				{
					wif1,
					"KwdMAjGmerYanjeui5SHS7JkmpZvVipYvB2LJGU1ZxJwYvWxyf5d",
				},
				{
					wif2,
					"cN1GXHxgB3dbmzxcgKJux56FfyU7chi5BtbejZsYHFk8ptuDeyhf",
				},
			}
		case chainec.ECTypeSecSchnorr:
			tests = []struct {
				wif     *exccutil.WIF
				encoded string
			}{
				{
					wif1,
					"KwdMAjGmerYanjeui5SHS7JkmpZvVipYvB2LJGU1ZxJwYvbMrQyG",
				},
				{
					wif2,
					"cV1Y7ARUr9Yx7BR55nTdnR7ZXNJphZtCCMBTEZBJe1hXt32Z7LP9",
				},
			}
		}

		for _, test := range tests {
			// Test that encoding the WIF structure matches the expected string.
			s := test.wif.String()
			if s != test.encoded {
				t.Errorf("TestEncodeDecodePrivateKey failed: want '%s', got '%s'",
					test.encoded, s)
				continue
			}

			// Test that decoding the expected string results in the original WIF
			// structure.
			w, err := exccutil.DecodeWIF(test.encoded)
			if err != nil {
				t.Error(err)
				continue
			}
			if got := w.String(); got != test.encoded {
				t.Errorf("NewWIF failed: want '%v', got '%v'", test.wif, got)
			}

			w.SerializePubKey()
		}
	}
}
