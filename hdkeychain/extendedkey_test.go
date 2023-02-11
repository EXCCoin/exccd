// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package hdkeychain

// References:
//   [BIP32]: BIP0032 - Hierarchical Deterministic Wallets
//   https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
)

// mockNetParams implements the NetworkParams interface and is used throughout
// the tests to mock multiple networks.
type mockNetParams struct {
	privKeyID [4]byte
	pubKeyID  [4]byte
}

// HDPrivKeyVersion returns the extended private key version bytes associated
// with the mock params.
//
// This is part of the NetworkParams interface.
func (p *mockNetParams) HDPrivKeyVersion() [4]byte {
	return p.privKeyID
}

// HDPubKeyVersion returns the extended public key version bytes associated with
// the mock params.
//
// This is part of the NetworkParams interface.
func (p *mockNetParams) HDPubKeyVersion() [4]byte {
	return p.pubKeyID
}

// mockMainNetParams returns mock mainnet parameters to use throughout the
// tests.  They match the Decred mainnet params as of the time this comment was
// written.
func mockMainNetParams() *mockNetParams {
	return &mockNetParams{
		privKeyID: [4]byte{0x04, 0x88, 0xAD, 0xE4}, // starts with xprv
		pubKeyID:  [4]byte{0x04, 0x88, 0xB2, 0x1E}, // starts with xpub
	}
}

// mockTestNetParams returns mock testnet parameters to use throughout the
// tests.  They match the Decred testnet version 3 params as of the time this
// comment was written.
func mockTestNetParams() *mockNetParams {
	return &mockNetParams{
		privKeyID: [4]byte{0x04, 0x35, 0x83, 0x97}, // starts with tprv
		pubKeyID:  [4]byte{0x04, 0x35, 0x87, 0xd1}, // starts with tpub
	}
}

// TestBIP0032Vectors tests the vectors provided by [BIP32] to ensure the
// derivation works as intended.
func TestBIP0032Vectors(t *testing.T) {
	// The master seeds for each of the two test vectors in [BIP32].
	//
	// Note that the 3rd seed has been changed to ensure the condition it is
	// intended to test applies with the modified hash function used in Decred.
	//
	// In particular, it results in hardened child 0 of the master key having
	// a child key with leading zeroes.
	testVec1MasterHex := "000102030405060708090a0b0c0d0e0f"
	testVec2MasterHex := "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542"
	testVec3MasterHex := "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c5100005930"
	hkStart := uint32(0x80000000)

	mainNetParams := mockMainNetParams()
	testNetParams := mockTestNetParams()
	tests := []struct {
		name     string
		master   string
		path     []uint32
		wantPub  string
		wantPriv string
		net      NetworkParams
	}{
		// Test vector 1
		{
			name:     "test vector 1 chain m",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
			wantPriv: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
			net:      mainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "xpub68Gmy5EdvgibQVfPdqkBBCHxA5htiqg55crXYuXoQRKfDBFA1WEjWgP6LHhwBZeNK1VTsfTFUHCdrfp1bgwQ9xv5ski8PX9rL2dZXvgGDnw",
			wantPriv: "xprv9uHRZZhk6KAJC1avXpDAp4MDc3sQKNxDiPvvkX8Br5ngLNv1TxvUxt4cV1rGL5hj6KCesnDYUhd7oWgT11eZG7XnxHrnYeSvkzY7d2bhkJ7",
			net:      mainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "xpub6ASuArnXKPbfEwhqN6e3mwBcDTgzisQN1wXN9BJcM47sSikHjJf3UFHKkNAWbWMiGj7Wf5uMash7SyYq527Hqck2AxYysAA7xmALppuCkwQ",
			wantPriv: "xprv9wTYmMFdV23N2TdNG573QoEsfRrWKQgWeibmLntzniatZvR9BmLnvSxqu53Kw1UmYPxLgboyZQaXwTCg8MSY3H2EU4pWcQDnRnrVA1xe8fs",
			net:      mainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "xpub6D4BDPcP2GT577Vvch3R8wDkScZWzQzMMUm3PWbmWvVJrZwQY4VUNgqFJPMM3No2dFDFGTsxxpG5uJh7n7epu4trkrX7x7DogT5Uv6fcLW5",
			wantPriv: "xprv9z4pot5VBttmtdRTWfWQmoH1taj2axGVzFqSb8C9xaxKymcFzXBDptWmT7FwuEzG3ryjH4ktypQSAewRiNMjANTtpgP4mLTj34bhnZX7UiM",
			net:      mainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "xpub6FHa3pjLCk84BayeJxFW2SP4XRrFd1JYnxeLeU8EqN3vDfZmbqBqaGJAyiLjTAwm6ZLRQUMv1ZACTj37sR62cfN7fe5JnJ7dh8zL4fiyLHV",
			wantPriv: "xprvA2JDeKCSNNZky6uBCviVfJSKyQ1mDYahRjijr5idH2WwLsEd4Hsb2Tyh8RfQMuPh7f7RtyzTtdrbdqqsunu5Mm3wDvUAKRHSC34sJ7in334",
			net:      mainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "xpub6H1LXWLaKsWFhvm6RVpEL9P4KfRZSW7abD2ttkWP3SSQvnyA8FSVqNTEcYFgJS2UaFcxupHiYkro49S8yGasTvXEYBVPamhGW6cFJodrTHy",
			wantPriv: "xprvA41z7zogVVwxVSgdKUHDy1SKmdb533PjDz7J6N6mV6uS3ze1ai8FHa8kmHScGpWmj4WggLyQjgPie1rFSruoUihUZREPSL39UNdE3BBDu76",
			net:      mainNetParams,
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
			wantPriv: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
			net:      mainNetParams,
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterHex,
			path:     []uint32{0},
			wantPub:  "xpub69H7F5d8KSRgmmdJg2KhpAK8SR3DjMwAdkxj3ZuxV27CprR9LgpeyGmXUbC6wb7ERfvrnKZjXoUmmDznezpbZb7ap6r1D3tgFxHmwMkQTPH",
			wantPriv: "xprv9vHkqa6EV4sPZHYqZznhT2NPtPCjKuDKGY38FBWLvgaDx45zo9WQRUT3dKYnjwih2yJD9mkrocEZXo1ex8G81dwSM1fwqWpWkeS3v86pgKt",
			net:      mainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647},
			wantPub:  "xpub6ASAVgeehLbnwdqV6UKMHVzgqAG8Gr6riv3Fxxpj8ksbH9ebxaEyBLZ85ySDhKiLDBrQSARLq1uNRts8RuJiHjaDMBU4Zn9h8LZNnBC5y4a",
			wantPriv: "xprv9wSp6B7kry3Vj9m1zSnLvN3xH8RdsPP1Mh7fAaR7aRLcQMKTR2vidYEeEg2mUCTAwCd6vnxVrcjfy2kRgVsFawNzmjuHc2YmYRmagcEPdU9",
			net:      mainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1},
			wantPub:  "xpub6DF8uhdarytz3FWdA8TvFSvvAh8dP3283MY7p2V4SeE2wyWmG5mg5EwVvmdMVCQcoNJxGoWaU9DCWh89LojfZ537wTfunKau47EL2dhHKon",
			wantPriv: "xprv9zFnWC6h2cLgpmSA46vutJzBcfJ8yaJGg8cX1e5StJh45BBciYTRXSd25UEPVuesF9yog62tGAQtHjXajPPdbRCHuWS6T8XA2ECKADdw4Ef",
			net:      mainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646},
			wantPub:  "xpub6ERApfZwUNrhLCkDtcHTcxd75RbzS1ed54G1LkBUHQVHQKqhMkhgbmJbZRkrgZw4koxb5JaHWkY4ALHY2grBGRjaDMzQLcgJvLJuZZvRcEL",
			wantPriv: "xprvA1RpRA33e1JQ7ifknakTFpgNXPmW2YvmhqLQYMmrj4xJXXWYpDPS3xz7iAxn8L39njGVyuoseXzU6rcxFLJ8HFsTjSyQbLYnMpCqE2VbFWc",
			net:      mainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H/2",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646, 2},
			wantPub:  "xpub6FnCn6nSzZAw5Tw7cgR9bi15UV96gLZhjDstkXXxvCLsUXBGXPdSnLFbdpq8p9HmGsApME5hQTZ3emM2rnY5agb9rXpVGyy3bdW6EEgAtqt",
			wantPriv: "xprvA2nrNbFZABcdryreWet9Ea4LvTJcGsqrMzxHx98MMrotbir7yrKCEXw7nadnHM8Dq38EGfSh6dqA9QWTyefMLEcBYJUuekgW4BYPJcr9E7j",
			net:      mainNetParams,
		},

		// Test vector 3
		{
			name:     "test vector 3 chain m",
			master:   testVec3MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFXZg2bWSKmcomc2Uwf9JrhB1zY4xJGQmJNk6xWVUC2cQmqzzLhdNcVMqRu2iKEC2qnLfyRqNfjQqYLkmu6R6pehwQzQiE8k",
			wantPriv: "xprv9s21ZrQH143K33VCvZyRxdg5DaBzYCRTVUFRC9fLjvsnRaQxQyBDeEHvvbQhtvjudakdsjFPiSmXBoMhFA6F7d43x4G1qtuXmu7YUwF4EZL",
			net:      mainNetParams,
		},
		{
			name:     "test vector 3 chain m/0H",
			master:   testVec3MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "xpub67wAGevRzekvsLr8oCt9LeHioAoTYMqtAUUssMZsPkMQCFhgE2oAzvVmsGnd1Q8FfeztrbzejXdnSfErNHQoaFCiUqbMLMG5nM1BaxAw152",
			wantPriv: "xprv9twos9PYAHCdermfhBM8yWLzF8xy8u82oFZH4yAFqQpRKTNXgVUvT8BJ1zFvG4rdZkMQU6HcPkqwM3xLWmMGJ3rS65eLQSPQ1TsYLFvEAqk",
			net:      mainNetParams,
		},
		{
			name:     "test vector 3 chain m/0H/0H",
			master:   testVec3MasterHex,
			path:     []uint32{hkStart, hkStart},
			wantPub:  "xpub69vz3nrvUh6ij2ETFpN6Lz6LMcWeiUmPqKJhbdW2fw3rd1ga62d4ZcJ9zfFDhpSfxbp5c66o33zeJcTQLmqesGrWicQtHgbDZQq8wMKsJuk",
			wantPriv: "xprv9vwdeHL2eKYRWY9z9nq5yr9boagAK23YU6P6oF6R7bWskDMRYVJp1oyg9Pctiaa7YbMmvGbNDwsVn4nwCc5hpnyfD9oDurL7LS9Jh2Yd144",
			net:      mainNetParams,
		},

		// Test vector 1 - Testnet
		{
			name:     "test vector 1 chain m - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "tpubVhnMyQmZAhoosedBTX7oacwyCNc5qtdEMoNHudUCW1R6WZTvqCZQoNJHSn4H11puwdk4qyDv2ET637EDap4r8HH3odjBC5nEjmnPcvoqJyL",
			wantPriv: "tprvZUo1ZuEfLLFWfAYiMVaoDV1EeLmbSRuNzaSh7F4awft7dm8nHfFAFZyobWQyV8Qr26r8M2CmNw6nEb35HaECWFGy1vzx2ZGdyfBeacrhh3a",
			net:      testNetParams,
		},
		{
			name:     "test vector 1 chain m/0H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "tpubVk3my852Fxvo2FktfGzSTf45XCASDy7F3QoBsirKUCcNsqm7FS6VvWehE3cykKjSwr2diCmLksXB1WCMXqoECAjn1GXS7PgTFhuwvXudTMi",
			wantPriv: "tprvZX4RZcY8RbNVomgRZFTS6X7LyAKwpWPPgBsb5LShus5Q13RxhtnFNiLDNmmJtqnoj9jpiKXdmHwexM4nwAWPJKMV5og6GWyXgfpW1kmLVLk",
			net:      testNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "tpubVnDuAucueforrhoLPXtK4Pwjaa9YDzqXyjU2Tzd8QqQb7PGEyEWot5Yve85ZAGSnuZegVdDSsU1ebowB1Ay7spZiJUNHb2gitSSjDUVfujH",
			wantPriv: "tprvZZEYmQ61pJFZeDisHWMJhG112YK3pY7gcWYRfcDWrVscEaw6RhCZLHESnpxNVmZrBEVWX984qzu56Hb24WJN5UqvbadpLGkPMU8sYherBBg",
			net:      testNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "tpubVpqBDSSmMYfGisbRe8HgRPysoj24VYRXKGhhiKvHahn2XETMmzMEnX6rC9GPc8t7G5kR71C4FQad495TiGWewGiYtNLRfykQc8MsJiot81R",
			wantPriv: "tprvZbqpovusXB6yWPWxY6kg4G39FhBa65hfx3n6uwWg2NF3eS8DET2zEinNLsAzU15LghWu7c4zGQiyKVKmeXDZCaHaxCCNVCzKxjt6BEzWfSU",
			net:      testNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "tpubVs4a3sZiY2LFoM59LPVmJu9BtYJo88jikkazyHSku9LdtL5iqm3bz6ZmsUFn1w2qjPsbF1g1J9UjcZRToZwresBoo9tcWAeEcpGiTMQQxrX",
			wantPriv: "tprvZe5DeN2phemxarzgEMxkwmCTLWUJig1sPXfQAu39Loof1XkaJDjMSJFJ2BaSvfUmkVebjXJZBEB8ngEDqwkuPxsdMSHU3Hp37iMFghuJcM6",
			net:      testNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "tpubVtnLXZAxf9iTKgrbSw4Vcc9Bgmt6wdYkYzyZDZpu7Dj8bTV7NBJGFCiqWJAisC7ZD6A8kMboqMBLCypUuRShW8LvfhJhJeDsRmtdhV5V5MZ",
			wantPriv: "tprvZfnz83e4pnAA7Cn8LuXVFUCT8k3cYApuBn3xRBRHYtC9if9xpdz1hQQMf3MeqabrMu3rWtHW2GiFnrEbP1mdWvXAgw3hACZkQ3ucRu4c6MQ",
			net:      testNetParams,
		},
	}

tests:
	for i, test := range tests {
		masterSeed, err := hex.DecodeString(test.master)
		if err != nil {
			t.Errorf("DecodeString #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}

		extKey, err := NewMaster(masterSeed, test.net)
		if err != nil {
			t.Errorf("NewMaster #%d (%s): unexpected error when "+
				"creating new master key: %v", i, test.name,
				err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.Child(childNum)
			if err != nil {
				t.Errorf("err: %v", err)
				continue tests
			}
		}

		privStr := extKey.String()
		if privStr != test.wantPriv {
			t.Errorf("Serialize #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, privStr, test.wantPriv)
			continue
		}

		pubKey := extKey.Neuter()

		// Neutering a second time should have no effect.
		// Test for referential equality.
		if pubKey != pubKey.Neuter() {
			t.Errorf("Neuter of extended public key returned " +
				"different object address")
			continue
		}

		pubStr := pubKey.String()
		if pubStr != test.wantPub {
			t.Errorf("Neuter #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.wantPub)
			continue
		}
	}
}

// TestPrivateDerivation tests several vectors which derive private keys from
// other private keys works as intended.
func TestPrivateDerivation(t *testing.T) {
	// The private extended keys for test vectors in [BIP32].
	testVec1MasterPrivKey := "xprv9xrCN3r9gRJF4pX6no8T6oLEEigiEM1qqzioEp9eN8L9HdXKNLjtsgsPAhfVcjZJuAKWNJ2vE8tAZPt2MMnr32B4iht4pnzssqjQbVzsxLx"
	testVec2MasterPrivKey := "xprv9zS5ZbA5KkunCtNiHtxQX558M6A8JDQZtvXFhaB7GYTatZqm3xQ5psaFRWetovE1PFPDPE8zqacUa84pirRi1DgbVet6sodRFkpshuLsik6"

	tests := []struct {
		name     string
		master   string
		path     []uint32
		wantPriv string
	}{
		// Test vector 1
		{
			name:     "test vector 1 chain m",
			master:   testVec1MasterPrivKey,
			path:     []uint32{},
			wantPriv: "xprv9xrCN3r9gRJF4pX6no8T6oLEEigiEM1qqzioEp9eN8L9HdXKNLjtsgsPAhfVcjZJuAKWNJ2vE8tAZPt2MMnr32B4iht4pnzssqjQbVzsxLx",
		},
		{
			name:     "test vector 1 chain m/0",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "xprvA1FzK1zN36dUcR2vD4fwSkF1jyZ4TtMEfbp1qvWBzHwukGBsSNW5KGHUJ19ufbGTvHhQSaLtCTzymZ43F2Y59xxkehWY42SzspciRb3eUMg",
		},
		{
			name:     "test vector 1 chain m/0/1",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1},
			wantPriv: "xprvA2g3sCD2RGUKSWc1tdVFcWeub7MiLEMbesvXTLegaYRLKMXFdPRMysnNpLcpF21umJCZTTQdfzdF27e15xJa3E2gpMXkNVKGybYAT3hadFJ",
		},
		{
			name:     "test vector 1 chain m/0/1/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2},
			wantPriv: "xprvA67sEY33eysWrhUKaB5TkGgLm2kS1R1P8e4ngb868S4vkj1CCsuVuX2ECpDmnSi7tyRD8y5uwyxgtc5aiJQqjUrLJXAm9Wdyj1Nun2GrEKV",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2},
			wantPriv: "xprvA7KnhHaSUFmZRyvbYwW7WfFmfWas1ViXAUrxQu2oDdHmLwBy2qgSwREJT6Pgf83daHYh3KUXFmjZ5RaypvgSdevStoWeXUpCqMrZUvVSaZ9",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2/1000000000",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2, 1000000000},
			wantPriv: "xprvA9Wwku2rF9WzcwMJcWTbQ75zQcoGWJCRKK9RcA1e3JNQkmZj4939ycqXQpJL9a7mq4C4xPeMSLfvCzTXEBSSjv4ht2gbZ3LYNWTqbU6mmrG",
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterPrivKey,
			path:     []uint32{},
			wantPriv: "xprv9zS5ZbA5KkunCtNiHtxQX558M6A8JDQZtvXFhaB7GYTatZqm3xQ5psaFRWetovE1PFPDPE8zqacUa84pirRi1DgbVet6sodRFkpshuLsik6",
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "xprvA2BusGTuJJjNQ3fcXGGVFZEb2ZPrvBDwJdPKFvpUWqvyNsyVGETrQi5c16LtH9LyNg2sG4JDhLBgSWZqZrxSf3okvBc6T5zN8TvmZRJFttf",
		},
		{
			name:     "test vector 2 chain m/0/2147483647",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647},
			wantPriv: "xprvA2XaZMHrpLxbJSy5UBXKK3dm8ij3pX5pAee41rXMMAB6eignVxexuCt9Dkgm6Cc3QSjbp3bhiF3vTdiRgfYwdkKcfCFp2RxdiRd1XZhPYjj",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1},
			wantPriv: "xprvA4MPQoPsc2Vd8MrfE1NrydxQwQe3PkAeB8ECKa5fghhHw8E1xvJbQ1wueTg2E27YvZvMRtuDeq8GmwTeyrhpRgviyknkqknyyC6UeoCU8WD",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646},
			wantPriv: "xprvA6vWhxjbVXB1FUWNWmK6i1rYkTBnkJoQ9oNATUcPA7nywwQYoJxbrDajarujhY6ykBZjp6tg2AyLhizPfDEF6w2es7swBnxnhTsScAoZ17p",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPriv: "xprvA8bBBZTnC1JjEax8N5XUWYeCHLtivL1dRorkpH2h1rYLaqoPScD8RmCzXmWsLBxspdghmshvxFctC4x12jkEd6rFiD1ddgURysC67DbrrEi",
		},

		// Custom tests to trigger specific conditions.
		{
			// Seed 000000000000000000000000000000da.
			name:     "Derived privkey with zero high byte m/0",
			master:   "xprv9xsYCYH2rfQ5x6g2mCuztfoZPRoU6enUJn1WStwSA7oiuoWpbRWYKmFW4x6mohLwYJzQzhhHbAfYtV3XUXqwEMmkWCdHxpwh69pXSKTfLif",
			path:     []uint32{0},
			wantPriv: "xprv9znyfxhTjAZajLYSnJELB41hEKmCxxr4ysKADcXveNW5RFTUCbe27Df2fc8Mxebyuay9msJ3MtPJcwPTTkoNKFMHgUapqwRnzL526racfpv",
		},
	}

	mainNetParms := mockMainNetParams()
tests:
	for i, test := range tests {
		extKey, err := NewKeyFromString(test.master, mainNetParms)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected error "+
				"creating extended key: %v", i, test.name,
				err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.Child(childNum)
			if err != nil {
				t.Errorf("err: %v", err)
				continue tests
			}
		}

		privStr := extKey.String()
		if privStr != test.wantPriv {
			t.Errorf("Child #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, privStr, test.wantPriv)
			continue
		}
	}
}

// TestPublicDerivation tests several vectors which derive public keys from
// other public keys works as intended.
func TestPublicDerivation(t *testing.T) {
	// The public extended keys for test vectors in [BIP32].
	testVec1MasterPubKey := "xpub6BqYmZP3WnrYHJbZtpfTTwGxnkXCdojhDDeQ3CZFvTs8ARrTut49RVBs1wMz6SE9D12o6Vh57BqyJmyRPfy7WaD8XQgUnGieUPR4N6yeAts"
	testVec2MasterPubKey := "xpub6DRRy6gyA8U5RNTBPvVQtD1ru7zchg8RG9SrVxaipszZmNAubViLNftjGo1mJYj1Bq5vZNMc8ieUu289rqfVdbC6H57imH8GXpgSji86mc6"

	tests := []struct {
		name    string
		master  string
		path    []uint32
		wantPub string
	}{
		// Test vector 1
		{
			name:    "test vector 1 chain m",
			master:  testVec1MasterPubKey,
			path:    []uint32{},
			wantPub: "xpub6BqYmZP3WnrYHJbZtpfTTwGxnkXCdojhDDeQ3CZFvTs8ARrTut49RVBs1wMz6SE9D12o6Vh57BqyJmyRPfy7WaD8XQgUnGieUPR4N6yeAts",
		},
		{
			name:    "test vector 1 chain m/0",
			master:  testVec1MasterPubKey,
			path:    []uint32{0},
			wantPub: "xpub6EFLiXXFsUBmpu7PK6CwotBkJ1PYsM562pjceJuoYdUtd4X1yupKs4bx9GiCcTKRs1uakptcCNcLC8dAsrmmaVSpNMjun8qnub178Pv41qi",
		},
		{
			name:    "test vector 1 chain m/0/1",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1},
			wantPub: "xpub6FfQGhjvFe2cezgUzf2Fyebe99CCjh5T26r8Fj4J8sxKC9rQAvjcXg6rfbWgQNyEJvaw1JsCti7eJwP6urSAU9jU8prNSK9K1xwq5dLi2T7",
		},
		{
			name:    "test vector 1 chain m/0/1/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2},
			wantPub: "xpub6K7De3ZwVMRp5BYngCcU7Qd5K4avQsjEVrzPUyXhgmbudXLLkRDkTKLi44pj2khGaTVoyyK7J2C4jG3nx6SdrpQegT1MWmQJ1BxWwZGKCaj",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2},
			wantPub: "xpub6LK96o7LJdKreU14ey37soCWDYRMQxSNXhnZDHSQmxpkDjX7aNzhVDYnJP4kEed6aL2CCGcDBWeHUvgyKroABzqVozDm3HjxbwoNR7rsgGZ",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2/1000000000",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2, 1000000000},
			wantPub: "xpub6NWJAQZk5X5HqRRmiXzbmF2ixedkukvGgY52QYRFbduPdZtsbgMQXRA1G4yWuXoWJa66rwAr8Kh3jmNjjM5b6Hic5pCcvAUFyg7babx8qre",
		},

		// Test vector 2
		{
			name:    "test vector 2 chain m",
			master:  testVec2MasterPubKey,
			path:    []uint32{},
			wantPub: "xpub6DRRy6gyA8U5RNTBPvVQtD1ru7zchg8RG9SrVxaipszZmNAubViLNftjGo1mJYj1Bq5vZNMc8ieUu289rqfVdbC6H57imH8GXpgSji86mc6",
		},
		{
			name:    "test vector 2 chain m/0",
			master:  testVec2MasterPubKey,
			path:    []uint32{0},
			wantPub: "xpub6FBGGmzo8gHfcXk5dHoVchBKabEMKdwnfrJv4KE65BTxFgJdomn6xWQ5rLyr62yM6zEabE17UwFcRz5hcvJ9Mv2wvjSyCV8MU64cneYM67h",
		},
		{
			name:    "test vector 2 chain m/0/2147483647",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647},
			wantPub: "xpub6FWvxrpkeiWtWw3YaD4KgBaVgkZYDyofXsZepEvxuVi5XX1w3VyDT1Cd52SB1zBEsMPRVBedMLygnKgzQp7hetCQF3rKjg3Fib8X2hayCgQ",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1},
			wantPub: "xpub6HLjpJvmSQ3vLqw8L2usLmu9VSUXoCtVYM9o7xVHF3EGovZAWTcqwpGPVkZcoz6QGMZJyhWawQEisoMCnrhUxT7CYhMLRw6XPydtJLEzZQW",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646},
			wantPub: "xpub6Kus7UGVKtjJTxaqcnr759oHJV2H9mXFX2HmFs1ziTKxpjjhLrGrQ1uDS81Mmu7ygChTEmV4SEsMPB2FkoV1S2xXBFrf7niZVaheFstMS3s",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPub: "xpub6MaXb4zg2Ns2T52bU74UsgavqNjDKnjUo2nMcfSJaC5KTe8Xz9XNyZXUP5VKu1BEhP3qKh8mbnnrL55VN9gh37RG9BSpr4vZXo9zSGuSh93",
		},
	}

	mainNetParams := mockMainNetParams()
tests:
	for i, test := range tests {
		extKey, err := NewKeyFromString(test.master, mainNetParams)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected error "+
				"creating extended key: %v", i, test.name,
				err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.Child(childNum)
			if err != nil {
				t.Errorf("err: %v", err)
				continue tests
			}
		}

		pubStr := extKey.String()
		if pubStr != test.wantPub {
			t.Errorf("Child #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.wantPub)
			continue
		}
	}
}

// TestGenenerateSeed ensures the GenerateSeed function works as intended.
func TestGenenerateSeed(t *testing.T) {
	tests := []struct {
		name   string
		length uint8
		err    error
	}{
		// Test various valid lengths.
		{name: "16 bytes", length: 16},
		{name: "17 bytes", length: 17},
		{name: "20 bytes", length: 20},
		{name: "32 bytes", length: 32},
		{name: "64 bytes", length: 64},

		// Test invalid lengths.
		{name: "15 bytes", length: 15, err: ErrInvalidSeedLen},
		{name: "65 bytes", length: 65, err: ErrInvalidSeedLen},
	}

	for i, test := range tests {
		seed, err := GenerateSeed(test.length)
		if !errors.Is(err, test.err) {
			t.Errorf("GenerateSeed #%d (%s): unexpected error -- "+
				"want %v, got %v", i, test.name, test.err, err)
			continue
		}

		if test.err == nil && len(seed) != int(test.length) {
			t.Errorf("GenerateSeed #%d (%s): length mismatch -- "+
				"got %d, want %d", i, test.name, len(seed),
				test.length)
			continue
		}
	}
}

// TestExtendedKeyAPI ensures the API on the ExtendedKey type works as intended.
func TestExtendedKeyAPI(t *testing.T) {
	tests := []struct {
		name       string
		extKey     string
		isPrivate  bool
		parentFP   uint32
		privKey    string
		privKeyErr error
		pubKey     string
	}{
		{
			name:      "test vector 1 master node private",
			extKey:    "xprv9s21ZrQH143K3GqwVJofp3rvzLsJ8J6RHt1L85LoDyZifXqcUeSPnMThY1QHsSorPoYcbyCs18A2kpyfw94Wpu3LHUV2ZAiHJ452vaRBJef",
			isPrivate: true,
			parentFP:  0,
			privKey:   "33a63922ea4e6686c9fc31daf136888297537f66c1aabe3363df06af0b8274c7",
			pubKey:    "039f2e1d7b50b8451911c64cf745f9ba16193b319212a64096e5679555449d8f37",
		},
		{
			name:       "test vector 2 chain m/0/2147483647/1/2147483646/2",
			extKey:     "xpub6Cz6m4eorCzY5FpYPWjQ4GFiFZ89eT7tJ5td6rGxaLn4LMywgznSsyAz3iAr32w8412J11DUg5CbEKh8t4AX5fiMeYxYtuQztZWJw5j3uxt",
			isPrivate:  false,
			parentFP:   3043329185,
			privKeyErr: ErrNotPrivExtKey,
			pubKey:     "029144a9ca3d1be341e6d5b94d172c396a0bdce74a4fe49106231147a5a6fbfa3b",
		},
	}

	mainNetParams := mockMainNetParams()
	for i, test := range tests {
		key, err := NewKeyFromString(test.extKey, mainNetParams)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected "+
				"error: %v", i, test.name, err)
			continue
		}

		if key.IsPrivate() != test.isPrivate {
			t.Errorf("IsPrivate #%d (%s): mismatched key type -- "+
				"want private %v, got private %v", i, test.name,
				test.isPrivate, key.IsPrivate())
			continue
		}

		parentFP := key.ParentFingerprint()
		if parentFP != test.parentFP {
			t.Errorf("ParentFingerprint #%d (%s): mismatched "+
				"parent fingerprint -- want %d, got %d", i,
				test.name, test.parentFP, parentFP)
			continue
		}

		serializedKey := key.String()
		if serializedKey != test.extKey {
			t.Errorf("String #%d (%s): mismatched serialized key "+
				"-- want %s, got %s", i, test.name, test.extKey,
				serializedKey)
			continue
		}

		privKey, err := key.SerializedPrivKey()
		if !errors.Is(err, test.privKeyErr) {
			t.Errorf("SerializedPrivKey #%d (%s): mismatched "+
				"error: want %v, got %v", i, test.name,
				test.privKeyErr, err)
			continue
		}
		if test.privKeyErr == nil {
			privKeyStr := hex.EncodeToString(privKey)
			if privKeyStr != test.privKey {
				t.Errorf("SerializedPrivKey #%d (%s): mismatched "+
					"private key -- want %s, got %s", i,
					test.name, test.privKey, privKeyStr)
				continue
			}
		}

		pubKeyStr := hex.EncodeToString(key.SerializedPubKey())
		if pubKeyStr != test.pubKey {
			t.Errorf("SerializedPubKey #%d (%s): mismatched public "+
				"key -- want %s, got %s", i, test.name, test.pubKey,
				pubKeyStr)
			continue
		}
	}
}

// TestErrors performs some negative tests for various invalid cases to ensure
// the errors are handled properly.
func TestErrors(t *testing.T) {
	// Should get an error when seed has too few bytes.
	net := mockMainNetParams()
	_, err := NewMaster(bytes.Repeat([]byte{0x00}, 15), net)
	if !errors.Is(err, ErrInvalidSeedLen) {
		t.Errorf("NewMaster: mismatched error -- got: %v, want: %v",
			err, ErrInvalidSeedLen)
	}

	// Should get an error when seed has too many bytes.
	_, err = NewMaster(bytes.Repeat([]byte{0x00}, 65), net)
	if !errors.Is(err, ErrInvalidSeedLen) {
		t.Errorf("NewMaster: mismatched error -- got: %v, want: %v",
			err, ErrInvalidSeedLen)
	}

	// Generate a new key and neuter it to a public extended key.
	seed, err := GenerateSeed(RecommendedSeedLen)
	if err != nil {
		t.Errorf("GenerateSeed: unexpected error: %v", err)
		return
	}
	extKey, err := NewMaster(seed, net)
	if err != nil {
		t.Errorf("NewMaster: unexpected error: %v", err)
		return
	}
	pubKey := extKey.Neuter()

	// Deriving a hardened child extended key should fail from a public key.
	_, err = pubKey.Child(HardenedKeyStart)
	if !errors.Is(err, ErrDeriveHardFromPublic) {
		t.Errorf("Child: mismatched error -- got: %v, want: %v",
			err, ErrDeriveHardFromPublic)
	}

	// NewKeyFromString failure tests.
	tests := []struct {
		name string
		key  string
		err  error
	}{
		{
			name: "invalid key length",
			key:  "xpub1234",
			err:  ErrInvalidKeyLen,
		},
		{
			name: "bad checksum",
			key:  "xpubZF6AWaFizAuUcbkZSs8cP8Gxzr6Sg5tLYYM7gEjZMC5GDaSHB4rW4F51zkWyo9U19BnXhc99kkEiPg248bYin8m9b8mGss9nxV6N2QpU8vj",
			err:  ErrBadChecksum,
		},
		{
			name: "unsupported version",
			key:  "tpubVp6YBY259b19iRKsbdpXMBMRD7rVPzhYyZ9XP2hmn3T9mPjb44LuAbc618MYRQpnnm6NGkTFKMWQoqSHzvpDAbn8UikMzejg9783pqgtgY1",
			err:  ErrWrongNetwork,
		},
	}

	mainNetParams := mockMainNetParams()
	for i, test := range tests {
		_, err := NewKeyFromString(test.key, mainNetParams)
		if !errors.Is(err, test.err) {
			t.Errorf("NewKeyFromString #%d (%s): mismatched error "+
				"-- got: %v, want: %v", i, test.name, err,
				test.err)
			continue
		}
	}
}

// TestZero ensures that zeroing an extended key works as intended.
func TestZero(t *testing.T) {
	mainNetParams := mockMainNetParams()
	tests := []struct {
		name   string
		master string
		extKey string
		net    NetworkParams
	}{
		// Test vector 1
		{
			name:   "test vector 1 chain m",
			master: "000102030405060708090a0b0c0d0e0f",
			extKey: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
			net:    mainNetParams,
		},

		// Test vector 2
		{
			name:   "test vector 2 chain m",
			master: "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			extKey: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
			net:    mainNetParams,
		},
	}

	// Use a closure to test that a key is zeroed since the tests create
	// keys in different ways and need to test the same things multiple
	// times.
	testZeroed := func(i int, testName string, key *ExtendedKey) bool {
		// Zeroing a key should result in it no longer being private
		if key.IsPrivate() {
			t.Errorf("IsPrivate #%d (%s): mismatched key type -- "+
				"want private %v, got private %v", i, testName,
				false, key.IsPrivate())
			return false
		}

		parentFP := key.ParentFingerprint()
		if parentFP != 0 {
			t.Errorf("ParentFingerprint #%d (%s): mismatched "+
				"parent fingerprint -- want %d, got %d", i,
				testName, 0, parentFP)
			return false
		}

		wantKey := "zeroed extended key"
		serializedKey := key.String()
		if serializedKey != wantKey {
			t.Errorf("String #%d (%s): mismatched serialized key "+
				"-- want %s, got %s", i, testName, wantKey,
				serializedKey)
			return false
		}

		wantErr := ErrNotPrivExtKey
		_, err := key.SerializedPrivKey()
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("SerializedPrivKey #%d (%s): mismatched "+
				"error: want %v, got %v", i, testName, wantErr,
				err)
			return false
		}

		serializedPubKey := key.SerializedPubKey()
		if len(serializedPubKey) != 0 {
			t.Errorf("ECPubKey #%d (%s): mismatched serialized "+
				"pubkey: want nil, got %x", i, testName,
				serializedPubKey)
			return false
		}

		return true
	}

	for i, test := range tests {
		// Create new key from seed and get the neutered version.
		masterSeed, err := hex.DecodeString(test.master)
		if err != nil {
			t.Errorf("DecodeString #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}
		key, err := NewMaster(masterSeed, test.net)
		if err != nil {
			t.Errorf("NewMaster #%d (%s): unexpected error when "+
				"creating new master key: %v", i, test.name,
				err)
			continue
		}
		neuteredKey := key.Neuter()

		// Ensure both non-neutered and neutered keys are zeroed
		// properly.
		key.Zero()
		if !testZeroed(i, test.name+" from seed not neutered", key) {
			continue
		}
		neuteredKey.Zero()
		if !testZeroed(i, test.name+" from seed neutered", key) {
			continue
		}

		// Deserialize key and get the neutered version.
		key, err = NewKeyFromString(test.extKey, mainNetParams)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected "+
				"error: %v", i, test.name, err)
			continue
		}
		neuteredKey = key.Neuter()

		// Ensure both non-neutered and neutered keys are zeroed
		// properly.
		key.Zero()
		if !testZeroed(i, test.name+" deserialized not neutered", key) {
			continue
		}
		neuteredKey.Zero()
		if !testZeroed(i, test.name+" deserialized neutered", key) {
			continue
		}
	}
}

// TestBIP0032Vector4 tests the BIP32 test vector 4 against keys derived from
// the ChildBIP32Std method, which derives keys that retain the leading zeros.
func TestBIP0032Vector4(t *testing.T) {
	testVec4MasterHex := "3ddd5602285899a946114506157c7997e5444528f3003f6134712147db19b678"
	hkStart := uint32(0x80000000)

	mainNetParams := mockMainNetParams()
	tests := []struct {
		name         string
		master       string
		path         []uint32
		wantPub      string
		wantPriv     string
		wantPrivSer  string
		leadingZeros int
		net          NetworkParams
	}{
		{
			name:        "test vector 4 chain m",
			master:      testVec4MasterHex,
			path:        []uint32{},
			wantPub:     "xpub661MyMwAqRbcGczjuMoRm6dXaLDEhW1u34gKenbeYqAix21mdUKJyuyu5F1rzYGVxyL6tmgBUAEPrEz92mBXjByMRiJdba9wpnN37RLLAXa",
			wantPriv:    "xprv9s21ZrQH143K48vGoLGRPxgo2JNkJ3J3fqkirQC2zVdk5Dgd5w14S7fRDyHH4dWNHUgkvsvNDCkvAwcSHNAQwhwgNMgZhLtQC63zxwhQmRv",
			wantPrivSer: "12c0d59c7aa3a10973dbd3f478b65f2516627e3fe61e00c345be9a477ad2e215",
			net:         mainNetParams,
		},
		{ // This test fails when using Child instead of ChildBIP32Std.
			name:         "test vector 4 chain m/0H -- leading zeros in private key serialization",
			master:       testVec4MasterHex,
			path:         []uint32{hkStart},
			wantPub:      "xpub69AUMk3qDBi3uW1sXgjCmVjJ2G6WQoYSnNHyzkmdCHEhSZ4tBok37xfFEqHd2AddP56Tqp4o56AePAgCjYdvpW2PU2jbUPFKsav5ut6Ch1m",
			wantPriv:     "xprv9vB7xEWwNp9kh1wQRfCCQMnZUEG21LpbR9NPCNN1dwhiZkjjeGRnaALmPXCX7SgjFTiCTT6bXes17boXtjq3xLpcDjzEuGLQBM5ohqkao9G",
			wantPrivSer:  "00d948e9261e41362a688b916f297121ba6bfb2274a3575ac0e456551dfd7f7e",
			leadingZeros: 1, // 1 zero byte
			net:          mainNetParams,
		},
		{
			name:        "test vector 4 chain m/0H/1H -- completely different key",
			master:      testVec4MasterHex,
			path:        []uint32{hkStart, hkStart + 1},
			wantPub:     "xpub6BJA1jSqiukeaesWfxe6sNK9CCGaujFFSJLomWHprUL9DePQ4JDkM5d88n49sMGJxrhpjazuXYWdMf17C9T5XnxkopaeS7jGk1GyyVziaMt",
			wantPriv:    "xprv9xJocDuwtYCMNAo3Zw76WENQeAS6WGXQ55RCy7tDJ8oALr4FWkuVoHJeHVAcAqiZLE7Je3vZJHxspZdFHfnBEjHqU5hG1Jaj32dVoS6XLT1",
			wantPrivSer: "3a2086edd7d9df86c3487a5905a1712a9aa664bce8cc268141e07549eaa8661d",
			net:         mainNetParams,
		},
	}

tests:
	for i, test := range tests {
		masterSeed, err := hex.DecodeString(test.master)
		if err != nil {
			t.Errorf("DecodeString #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}

		extKey, err := NewMaster(masterSeed, test.net)
		if err != nil {
			t.Errorf("NewMaster #%d (%s): unexpected error when "+
				"creating new master key: %v", i, test.name, err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.ChildBIP32Std(childNum)
			if err != nil {
				t.Errorf("ChildBIP32Std #%d (%s): %v", i, test.name, err)
				continue tests
			}
		}

		priv, err := extKey.SerializedPrivKey()
		if err != nil {
			t.Errorf("SerializedPrivKey #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}
		if len(priv) != 32 {
			t.Errorf("SerializedPrivKey #%d (%s): serialized private key "+
				"length %d, want 32", i, test.name, len(priv))
			continue
		}

		privStr := hex.EncodeToString(priv)
		if privStr != test.wantPrivSer {
			t.Errorf("Serialize #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, privStr, test.wantPrivSer)
			continue
		}

		extKeyStr := extKey.String()
		if extKeyStr != test.wantPriv {
			t.Errorf("Serialize #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, extKeyStr, test.wantPriv)
			continue
		}

		pubKey := extKey.Neuter()

		// Neutering a second time should have no effect.
		// Test for referential equality.
		if pubKey != pubKey.Neuter() {
			t.Errorf("Neuter #%d (%s): of extended public key returned "+
				"different object address", i, test.name)
			continue
		}

		pubStr := pubKey.String()
		if pubStr != test.wantPub {
			t.Errorf("Neuter #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.wantPub)
			continue
		}

		// If this path generates a private key with leading zeros, ensure they
		// are stripped with the legacy Decred child derivation.
		if test.leadingZeros > 0 {
			wantLen := 32 - test.leadingZeros
			extKey, err := NewMaster(masterSeed, test.net)
			if err != nil {
				t.Errorf("NewMaster #%d (%s): unexpected error when "+
					"creating new master key: %v", i, test.name, err)
				continue
			}

			for _, childNum := range test.path {
				var err error
				extKey, err = extKey.Child(childNum) // modified/legacy BIP32
				if err != nil {
					t.Errorf("Child #%d (%s): %v", i, test.name, err)
					continue tests
				}
			}

			priv, err := extKey.SerializedPrivKey()
			if err != nil {
				t.Errorf("SerializedPrivKey #%d (%s): unexpected error: %v",
					i, test.name, err)
				continue
			}
			if len(priv) != wantLen {
				t.Errorf("SerializedPrivKey #%d (%s): serialized private key "+
					"length %d, want %d", i, test.name, len(priv), wantLen)
			}
		}
	}
}
