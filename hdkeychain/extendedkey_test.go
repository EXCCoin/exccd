// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package hdkeychain_test

// References:
//   [BIP32]: BIP0032 - Hierarchical Deterministic Wallets
//   https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"bytes"
	"github.com/EXCCoin/exccd/chaincfg"
	"github.com/EXCCoin/exccd/hdkeychain"
)

// TestBIP0032Vectors tests the vectors provided by [BIP32] to ensure the
// derivation works as intended.
func TestBIP0032Vectors(t *testing.T) {
	// The master seeds for each of the two test vectors in [BIP32].
	testVec1MasterHex := "000102030405060708090a0b0c0d0e0f"
	testVec2MasterHex := "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542"
	hkStart := uint32(0x80000000)

	tests := []struct {
		name     string
		master   string
		path     []uint32
		wantPub  string
		wantPriv string
		net      *chaincfg.Params
	}{
		// Test vector 1
		{
			name:     "test vector 1 chain m",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
			wantPriv: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "xpub68Gmy5EdvgibQVfPdqkBBCHxA5htiqg55crXYuXoQRKfDBFA1WEjWgP6LHhwBZeNK1VTsfTFUHCdrfp1bgwQ9xv5ski8PX9rL2dZXvgGDnw",
			wantPriv: "xprv9uHRZZhk6KAJC1avXpDAp4MDc3sQKNxDiPvvkX8Br5ngLNv1TxvUxt4cV1rGL5hj6KCesnDYUhd7oWgT11eZG7XnxHrnYeSvkzY7d2bhkJ7",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "xpub6ASuArnXKPbfEwhqN6e3mwBcDTgzisQN1wXN9BJcM47sSikHjJf3UFHKkNAWbWMiGj7Wf5uMash7SyYq527Hqck2AxYysAA7xmALppuCkwQ",
			wantPriv: "xprv9wTYmMFdV23N2TdNG573QoEsfRrWKQgWeibmLntzniatZvR9BmLnvSxqu53Kw1UmYPxLgboyZQaXwTCg8MSY3H2EU4pWcQDnRnrVA1xe8fs",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "xpub6D4BDPcP2GT577Vvch3R8wDkScZWzQzMMUm3PWbmWvVJrZwQY4VUNgqFJPMM3No2dFDFGTsxxpG5uJh7n7epu4trkrX7x7DogT5Uv6fcLW5",
			wantPriv: "xprv9z4pot5VBttmtdRTWfWQmoH1taj2axGVzFqSb8C9xaxKymcFzXBDptWmT7FwuEzG3ryjH4ktypQSAewRiNMjANTtpgP4mLTj34bhnZX7UiM",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "xpub6FHa3pjLCk84BayeJxFW2SP4XRrFd1JYnxeLeU8EqN3vDfZmbqBqaGJAyiLjTAwm6ZLRQUMv1ZACTj37sR62cfN7fe5JnJ7dh8zL4fiyLHV",
			wantPriv: "xprvA2JDeKCSNNZky6uBCviVfJSKyQ1mDYahRjijr5idH2WwLsEd4Hsb2Tyh8RfQMuPh7f7RtyzTtdrbdqqsunu5Mm3wDvUAKRHSC34sJ7in334",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "xpub6H1LXWLaKsWFhvm6RVpEL9P4KfRZSW7abD2ttkWP3SSQvnyA8FSVqNTEcYFgJS2UaFcxupHiYkro49S8yGasTvXEYBVPamhGW6cFJodrTHy",
			wantPriv: "xprvA41z7zogVVwxVSgdKUHDy1SKmdb533PjDz7J6N6mV6uS3ze1ai8FHa8kmHScGpWmj4WggLyQjgPie1rFSruoUihUZREPSL39UNdE3BBDu76",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
			wantPriv: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterHex,
			path:     []uint32{0},
			wantPub:  "xpub69H7F5d8KSRgmmdJg2KhpAK8SR3DjMwAdkxj3ZuxV27CprR9LgpeyGmXUbC6wb7ERfvrnKZjXoUmmDznezpbZb7ap6r1D3tgFxHmwMkQTPH",
			wantPriv: "xprv9vHkqa6EV4sPZHYqZznhT2NPtPCjKuDKGY38FBWLvgaDx45zo9WQRUT3dKYnjwih2yJD9mkrocEZXo1ex8G81dwSM1fwqWpWkeS3v86pgKt",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647},
			wantPub:  "xpub6ASAVgeehLbnwdqV6UKMHVzgqAG8Gr6riv3Fxxpj8ksbH9ebxaEyBLZ85ySDhKiLDBrQSARLq1uNRts8RuJiHjaDMBU4Zn9h8LZNnBC5y4a",
			wantPriv: "xprv9wSp6B7kry3Vj9m1zSnLvN3xH8RdsPP1Mh7fAaR7aRLcQMKTR2vidYEeEg2mUCTAwCd6vnxVrcjfy2kRgVsFawNzmjuHc2YmYRmagcEPdU9",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1},
			wantPub:  "xpub6DF8uhdarytz3FWdA8TvFSvvAh8dP3283MY7p2V4SeE2wyWmG5mg5EwVvmdMVCQcoNJxGoWaU9DCWh89LojfZ537wTfunKau47EL2dhHKon",
			wantPriv: "xprv9zFnWC6h2cLgpmSA46vutJzBcfJ8yaJGg8cX1e5StJh45BBciYTRXSd25UEPVuesF9yog62tGAQtHjXajPPdbRCHuWS6T8XA2ECKADdw4Ef",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646},
			wantPub:  "xpub6ERApfZwUNrhLCkDtcHTcxd75RbzS1ed54G1LkBUHQVHQKqhMkhgbmJbZRkrgZw4koxb5JaHWkY4ALHY2grBGRjaDMzQLcgJvLJuZZvRcEL",
			wantPriv: "xprvA1RpRA33e1JQ7ifknakTFpgNXPmW2YvmhqLQYMmrj4xJXXWYpDPS3xz7iAxn8L39njGVyuoseXzU6rcxFLJ8HFsTjSyQbLYnMpCqE2VbFWc",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H/2",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646, 2},
			wantPub:  "xpub6FnCn6nSzZAw5Tw7cgR9bi15UV96gLZhjDstkXXxvCLsUXBGXPdSnLFbdpq8p9HmGsApME5hQTZ3emM2rnY5agb9rXpVGyy3bdW6EEgAtqt",
			wantPriv: "xprvA2nrNbFZABcdryreWet9Ea4LvTJcGsqrMzxHx98MMrotbir7yrKCEXw7nadnHM8Dq38EGfSh6dqA9QWTyefMLEcBYJUuekgW4BYPJcr9E7j",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 1 - Testnet
		{
			name:     "test vector 1 chain m - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "tpubVhnMyQmZAhoosedBTX7oacwyCNc5qtdEMoNHudUCW1R6WZTvqCZQoNJHSn4H11puwdk4qyDv2ET637EDap4r8HH3odjBC5nEjmnPcvoqJyL",
			wantPriv: "tprvZUo1ZuEfLLFWfAYiMVaoDV1EeLmbSRuNzaSh7F4awft7dm8nHfFAFZyobWQyV8Qr26r8M2CmNw6nEb35HaECWFGy1vzx2ZGdyfBeacrhh3a",
			net:      &chaincfg.TestNetParams,
		},
		{
			name:     "test vector 1 chain m/0H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "tpubVk3my852Fxvo2FktfGzSTf45XCASDy7F3QoBsirKUCcNsqm7FS6VvWehE3cykKjSwr2diCmLksXB1WCMXqoECAjn1GXS7PgTFhuwvXudTMi",
			wantPriv: "tprvZX4RZcY8RbNVomgRZFTS6X7LyAKwpWPPgBsb5LShus5Q13RxhtnFNiLDNmmJtqnoj9jpiKXdmHwexM4nwAWPJKMV5og6GWyXgfpW1kmLVLk",
			net:      &chaincfg.TestNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "tpubVnDuAucueforrhoLPXtK4Pwjaa9YDzqXyjU2Tzd8QqQb7PGEyEWot5Yve85ZAGSnuZegVdDSsU1ebowB1Ay7spZiJUNHb2gitSSjDUVfujH",
			wantPriv: "tprvZZEYmQ61pJFZeDisHWMJhG112YK3pY7gcWYRfcDWrVscEaw6RhCZLHESnpxNVmZrBEVWX984qzu56Hb24WJN5UqvbadpLGkPMU8sYherBBg",
			net:      &chaincfg.TestNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "tpubVpqBDSSmMYfGisbRe8HgRPysoj24VYRXKGhhiKvHahn2XETMmzMEnX6rC9GPc8t7G5kR71C4FQad495TiGWewGiYtNLRfykQc8MsJiot81R",
			wantPriv: "tprvZbqpovusXB6yWPWxY6kg4G39FhBa65hfx3n6uwWg2NF3eS8DET2zEinNLsAzU15LghWu7c4zGQiyKVKmeXDZCaHaxCCNVCzKxjt6BEzWfSU",
			net:      &chaincfg.TestNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "tpubVs4a3sZiY2LFoM59LPVmJu9BtYJo88jikkazyHSku9LdtL5iqm3bz6ZmsUFn1w2qjPsbF1g1J9UjcZRToZwresBoo9tcWAeEcpGiTMQQxrX",
			wantPriv: "tprvZe5DeN2phemxarzgEMxkwmCTLWUJig1sPXfQAu39Loof1XkaJDjMSJFJ2BaSvfUmkVebjXJZBEB8ngEDqwkuPxsdMSHU3Hp37iMFghuJcM6",
			net:      &chaincfg.TestNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "tpubVtnLXZAxf9iTKgrbSw4Vcc9Bgmt6wdYkYzyZDZpu7Dj8bTV7NBJGFCiqWJAisC7ZD6A8kMboqMBLCypUuRShW8LvfhJhJeDsRmtdhV5V5MZ",
			wantPriv: "tprvZfnz83e4pnAA7Cn8LuXVFUCT8k3cYApuBn3xRBRHYtC9if9xpdz1hQQMf3MeqabrMu3rWtHW2GiFnrEbP1mdWvXAgw3hACZkQ3ucRu4c6MQ",
			net:      &chaincfg.TestNetParams,
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

		extKey, err := hdkeychain.NewMaster(masterSeed, test.net)
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

		privStr, _ := extKey.String()
		if privStr != test.wantPriv {
			t.Errorf("Serialize #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, privStr, test.wantPriv)
			continue
		}

		pubKey, err := extKey.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v ", i,
				test.name, err)
			continue
		}

		// Neutering a second time should have no effect.
		pubKey, err = pubKey.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v", i,
				test.name, err)
			return
		}

		pubStr, _ := pubKey.String()
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
	testVec1MasterPrivKey := "dprv3hCznBesA6jBucms1ZhyGeFfvJfBSwfs7ZFrxS8tdYzbjDZe2UwSaL7EbYo1qa88DmtyyG5cL9tdGxHkD89JmeZTbz5sVYU4DgtijmzG6Bu"
	testVec2MasterPrivKey := "dprv3hCznBesA6jBtPKJbQTxRZAKG2gyj8tZKEPaCsV4e9YYFBAgRP2eTSPAeu4r8dTMt9q51j2Vdt5zNqj7jbtovvocrP1qLj6WUTLF9yGk28v"

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
			wantPriv: "dprv3hCznBesA6jBucms1ZhyGeFfvJfBSwfs7ZFrxS8tdYzbjDZe2UwSaL7EbYo1qa88DmtyyG5cL9tdGxHkD89JmeZTbz5sVYU4DgtijmzG6Bu",
		},
		{
			name:     "test vector 1 chain m/0",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "dprv3jP4XB95sQTZTXiF7ovKEWSbrQmhV5E8QrVR7RLqRWZZnzbWZD8YUXLMk9kEc9AWtuRgtR7AexwGjWQUPNLuTgFRTkCNL8x5mZzB7TPd7gQ",
		},
		{
			name:     "test vector 1 chain m/0/1",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1},
			wantPriv: "dprv3mpHByDDob83Pic6ffzzQt3i2Cptfgr6w1KwVzTFtCP6YboerCARAwGT8u4yrJJHMSkxTHJbc72Nu3o4JX49sW2sF3j9UCYVy1TZ4NYYC1x",
		},
		{
			name:     "test vector 1 chain m/0/1/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2},
			wantPriv: "dprv3oqHzsyfBPkwtDJ2C6WqiMKYS2T65y27cwRdQWGxECt1iRm7HKf34CaMWQRsdEx2wTaWD4Cd3wniVFp9wTugpch8xTTVHjiaYMTAvPp1RBz",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2},
			wantPriv: "dprv3qR6vptAiiG9s7xiBAGvG2Mh8x7V7NTH9r6DhRtvhhMapEhxKC35BAdwGZzYqNjWt25ufcxNTNRsu1UHhcEJ1MEEqmXvdScjpEn96B8ug4q",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2/1000000000",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2, 1000000000},
			wantPriv: "dprv3rzXzRMQznMkdiwnQFPd3XexQsUP2cGJTcCy7WKaquAfsw7b5N2xRPTfnGemMMuRT5JwgRmAi55nbNniF6EYBKEWcmNsqp8CZePBRDr6VQ6",
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterPrivKey,
			path:     []uint32{},
			wantPriv: "dprv3hCznBesA6jBtPKJbQTxRZAKG2gyj8tZKEPaCsV4e9YYFBAgRP2eTSPAeu4r8dTMt9q51j2Vdt5zNqj7jbtovvocrP1qLj6WUTLF9yGk28v",
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "dprv3kUk3uLpe7ZGQeubjBqqVMZqczSES7NqsQ8wNgWWiZQVw9cezeNicGxREQLemhJCHp4tujSE53W7jMMwHtPrs9y1rZRtjDGFbsEoXCocXou",
		},
		{
			name:     "test vector 2 chain m/0/2147483647",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647},
			wantPriv: "dprv3mdoJWNCgMCQQoAym9oPiP4QmH8x5gqTCfPtgD6iAowyzh4S2eHFaukJxuS86MtY5d8WzugybyUfozciYeze7xYmc7CS6zMRD6NSCsynng9",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1},
			wantPriv: "dprv3p4RHw6gdeC5GZfNqsyxWhZRS6kBKKYWRQFVzzTXLejhsfCnvVqrNskLfXvu3GG1rW5dBBd2kh73F9NS2bQxTRGPnAMwYhYZV2aVg7EhWbA",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646},
			wantPriv: "dprv3rXerqmXaFeZv6CUqoGVporTxf1zaTvKVUErvGUhymRRquy3k1SSYLZKke5JNKLBK6tkEcMDVJB3Wbbyh6a3inumTfszcL52E85kwPgwf3o",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPriv: "dprv3tK9Z8NmNnJJ5pStaGV1SKSV1ikA5kiQtV5HmqRRyG7otBixdmcFgv8tQjGNUzZxNB2jbhmbwYDamNBVUT2i5mbE8MvvpK9Vit9jHAcKcgQ",
		},

		// Custom tests to trigger specific conditions.
		{
			// Seed 000000000000000000000000000000da.
			name:     "Derived privkey with zero high byte m/0",
			master:   "dprv3jFfEhxvVxy6NJWopujhfg7syQL71xCRgNoGUpQTtjTpCwzigwtCwssQGbRQsby7PBs1Yp8Wu7isu396qeNof13EZuxbCTJVF1xkoFLFgMx",
			path:     []uint32{0},
			wantPriv: "dprv3mpHByDDob83MR4HJhATU4i9Y5MzRXSCjS1mTzUrSPj6q94hg2NXNse6uEg6Y9RAmsFXHbzLx7Kyr8J7qnV5rZ9mdXiLiLSQtVnMFZ8WTW5",
		},
	}

tests:
	for i, test := range tests {
		extKey, err := hdkeychain.NewKeyFromString(test.master)
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

		privStr, _ := extKey.String()
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
	testVec1MasterPubKey := "dpubZF8BRmciAzYoTjXZ3bbRWLVCwUKtTquact3Tr6ye77Rgmw76VyqMb9TB9KpfrvUYEM5d1Au4fQzE2BbtxRjwzGsqnWHmtQP9UV1kxYHcpqN"
	testVec2MasterPubKey := "dpubZF4LSCdF9YKZfNzTVYhz4RBxsjYXqms8AQnMBHXZ8GUKoRSigG7kQnKiJt5pzk93Q8FxcdVBEkQZruSXduGtWnkwXzGnjbSovQ97d9inEWC"

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
			wantPub: "dpubZF8BRmciAzYoTjXZ3bbRWLVCwUKtTquact3Tr6ye77Rgmw76VyqMb9TB9KpfrvUYEM5d1Au4fQzE2BbtxRjwzGsqnWHmtQP9UV1kxYHcpqN",
		},
		{
			name:    "test vector 1 chain m/0",
			master:  testVec1MasterPubKey,
			path:    []uint32{0},
			wantPub: "dpubZJHomN5nEoDeYbnJpyNT6tT1hJGJSPzZ72nFxbu1etrCY9wxd4e5QUiHyt489eWfjeVkFQe4q5hGayqDx5tDiFjx2iBX5yu3QBJg7G9o16d",
		},
		{
			name:    "test vector 1 chain m/0/1",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1},
			wantPub: "dpubZKMq3QyaJtddjXkQUFXK8RSVkKoSrPdiRYpE3zvm6Bm72YGpQjrv4hfGWg9sYQnPdpj5d5pM9ctgkawVPgzwFAc22eeb5wBEuvgwuXBxZ6A",
		},
		{
			name:    "test vector 1 chain m/0/1/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2},
			wantPub: "dpubZMHV2NU1finY6g1Bz6eecue4TCvnWAGqusP5Yj8rcsBXrEvBczF3ZdvGPDfPr4vguXbyMRpCcZrzzrNBRrkMKvvL1nLyvaBWJknzhRFT5i4",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2},
			wantPub: "dpubZPW6TLNqBDgmYHBUhWRdhCaYqso5fckd8pexQfUEVNozw7sWFiJjFwWG2eBMkAn6s6jYG1QSkNU5PvHgwVorw96YxgQjQ3JZ8GmdM6Rye3y",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2/1000000000",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2, 1000000000},
			wantPub: "dpubZR9pbnPM8zxPEXZFww9MazBmL6sqBj77XEGi58fThKUUeak8GFvGxCKBv56JdcjXhRUBBhLRfsfGRh61F7NaTNihvBHMJe6L9Bh89tUR7Du",
		},

		// Test vector 2
		{
			name:    "test vector 2 chain m",
			master:  testVec2MasterPubKey,
			path:    []uint32{},
			wantPub: "dpubZF4LSCdF9YKZfNzTVYhz4RBxsjYXqms8AQnMBHXZ8GUKoRSigG7kQnKiJt5pzk93Q8FxcdVBEkQZruSXduGtWnkwXzGnjbSovQ97d9inEWC",
		},
		{
			name:    "test vector 2 chain m/0",
			master:  testVec2MasterPubKey,
			path:    []uint32{0},
			wantPub: "dpubZHwDcH4ryrSJeZvfx2X7PLDkH4TJ4pCvNUkegwqGnBArJXaqYNA41ZmnZnBnrz6qCn5FJzrmEjGRGepG7ETdaz3JwcxxFKKXNgfHPBLASuS",
		},
		{
			name:    "test vector 2 chain m/0/2147483647",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647},
			wantPub: "dpubZKhiqFdqE3XC58qQ4THN8f5GKHeJYTAerkPRbn2LrmTdXdxo687HPw9mNhjitepMZj2BMZ9xBqK4rEraZgN44FvT6PkrxdozHgTnHKEnZbL",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1},
			wantPub: "dpubZMpqVYWGJ8tgC8c6UMCZRi5RgY3m9VxfoxMJcMaCLh8PnBqnwviwhpKAB7aU1nKPWuaw72aU1tWyudnNGazpT7SvcKZMmzSnoXi8Sz3DNgb",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646},
			wantPub: "dpubZPT82GSuJdPbu9hj9uqqCUdviti3oX5eASnwJx55H7kAS1U35KGTA3DxmhhXu7cmAXn7xrM6p1ZHdLkA39oXRVxXvWgPuWCeS3kyWT8ioWK",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPub: "dpubZQkz4MWdTXsExN6y8inzx7QHGhjENcwmRCydE5AuvRkJxuXTvxnG2mwspex6YpqFvpNoD62rtnziPxRCJ9d9DTrAXqPmFCyjvBXvFcC4Lty",
		},
	}

tests:
	for i, test := range tests {
		extKey, err := hdkeychain.NewKeyFromString(test.master)
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

		pubStr, _ := extKey.String()
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
	wantErr := errors.New("seed length must be between 128 and 512 bits")

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
		{name: "15 bytes", length: 15, err: wantErr},
		{name: "65 bytes", length: 65, err: wantErr},
	}

	for i, test := range tests {
		seed, err := hdkeychain.GenerateSeed(test.length)
		if !reflect.DeepEqual(err, test.err) {
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
		address    string
	}{
		{
			name:      "test vector 1 master node private",
			extKey:    "xprv9s21ZrQH143K3GqwVJofp3rvzLsJ8J6RHt1L85LoDyZifXqcUeSPnMThY1QHsSorPoYcbyCs18A2kpyfw94Wpu3LHUV2ZAiHJ452vaRBJef",
			isPrivate: true,
			parentFP:  0,
			privKey:   "33a63922ea4e6686c9fc31daf136888297537f66c1aabe3363df06af0b8274c7",
			pubKey:    "039f2e1d7b50b8451911c64cf745f9ba16193b319212a64096e5679555449d8f37",
			address:   "22tk78UXaPYZY6Z5ZS2hbqYKYvW188VKVBJW",
		},
		{
			name:       "test vector 2 chain m/0/2147483647/1/2147483646/2",
			extKey:     "dpubZRuRErXqhdJaZWD1AzXB6d5w2zw7UZ7ALxiS1gHbnQbVEohBzQzsVwGRzq97pmuE7ToA6DGn2QTH4DexxzdnMvkiYUpk8Nh2KEuYUPMnucA",
			isPrivate:  false,
			parentFP:   4220580796,
			privKeyErr: hdkeychain.ErrNotPrivExtKey,
			pubKey:     "03dceb0b07698ec3d6ac08ae7297e7f5e63d7fda99d3fce1ded31d36badcdd4d36",
			address:    "22u8UvFKiamQ4tKVGjsZt69KhCxfpe5XMoTN",
		},
	}

	for i, test := range tests {
		key, err := hdkeychain.NewKeyFromString(test.extKey)
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

		serializedKey, _ := key.String()
		if serializedKey != test.extKey {
			t.Errorf("String #%d (%s): mismatched serialized key "+
				"-- want %s, got %s", i, test.name, test.extKey,
				serializedKey)
			continue
		}

		privKey, err := key.ECPrivKey()
		if !reflect.DeepEqual(err, test.privKeyErr) {
			t.Errorf("ECPrivKey #%d (%s): mismatched error: want "+
				"%v, got %v", i, test.name, test.privKeyErr, err)
			continue
		}
		if test.privKeyErr == nil {
			privKeyStr := hex.EncodeToString(privKey.Serialize())
			if privKeyStr != test.privKey {
				t.Errorf("ECPrivKey #%d (%s): mismatched "+
					"private key -- want %s, got %s", i,
					test.name, test.privKey, privKeyStr)
				continue
			}
		}

		pubKey, err := key.ECPubKey()
		if err != nil {
			t.Errorf("ECPubKey #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}
		pubKeyStr := hex.EncodeToString(pubKey.SerializeCompressed())
		if pubKeyStr != test.pubKey {
			t.Errorf("ECPubKey #%d (%s): mismatched public key -- "+
				"want %s, got %s", i, test.name, test.pubKey,
				pubKeyStr)
			continue
		}

		addr, err := key.Address(&chaincfg.MainNetParams)
		if err != nil {
			t.Errorf("Address #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}
		if addr.EncodeAddress() != test.address {
			t.Errorf("Address #%d (%s): mismatched address -- want "+
				"%s, got %s", i, test.name, test.address,
				addr.EncodeAddress())
			continue
		}
	}
}

// TestNet ensures the network related APIs work as intended.
func TestNet(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		origNet   *chaincfg.Params
		newNet    *chaincfg.Params
		newPriv   string
		newPub    string
		isPrivate bool
	}{
		// Private extended keys.
		{
			name:      "mainnet -> simnet",
			key:       "xprv9s21ZrQH143K3gjdi76nKycGU7vJokj9LL5ZbyYjk4mKBLY93FNmzzJhexcj9R9vH8JDkYtBVkoyuTWmUA5EXppdDgzbNgo81gr9nJZExtg",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.SimNetParams,
			newPriv:   "sprvZ9xkGEZkBei2p9e1uBZRQMGtfGEQNGApP1W19PyNRqg9nuEs2X4ynkvAXWaBiGb5WKiaqcbiKgmyB1HYgcX3mnxiUs7UWeWEfe4tnSyYwB3",
			newPub:    "spubVNx6fk6e22GL2diV1D6RmVDdDJ4tmitfkERbwnNyzBD8fha1a4PELZEeNoUfNofdyJS2Y19tFgHZQ62tzKwELiBA3xVeZowLr4DJQ8rCgjL",
			isPrivate: true,
		},
		{
			name:      "simnet -> mainnet",
			key:       "sprvZ9xkGEZkBei2p9e1uBZRQMGtfGEQNGApP1W19PyNRqg9nuEs2X4ynkvAXWaBiGb5WKiaqcbiKgmyB1HYgcX3mnxiUs7UWeWEfe4tnSyYwB3",
			origNet:   &chaincfg.SimNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPriv:   "xprv9s21ZrQH143K3gjdi76nKycGU7vJokj9LL5ZbyYjk4mKBLY93FNmzzJhexcj9R9vH8JDkYtBVkoyuTWmUA5EXppdDgzbNgo81gr9nJZExtg",
			newPub:    "xpub661MyMwAqRbcGAp6p8dnh7Z129koDDSzhZ1AQMxMJQJJ48sHanh2YndBWFXCoxEUk71fSwSMRkKa8YG7msVR6k34nnNmRrEEC6zZQ3wh1Ra",
			isPrivate: true,
		},

		// Public extended keys.
		{
			name:      "mainnet -> simnet",
			key:       "xpub661MyMwAqRbcGAp6p8dnh7Z129koDDSzhZ1AQMxMJQJJ48sHanh2YndBWFXCoxEUk71fSwSMRkKa8YG7msVR6k34nnNmRrEEC6zZQ3wh1Ra",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.SimNetParams,
			newPub:    "spubVNx6fk6e22GL2diV1D6RmVDdDJ4tmitfkERbwnNyzBD8fha1a4PELZEeNoUfNofdyJS2Y19tFgHZQ62tzKwELiBA3xVeZowLr4DJQ8rCgjL",
			isPrivate: false,
		},
		{
			name:      "simnet -> mainnet",
			key:       "spubVNx6fk6e22GL2diV1D6RmVDdDJ4tmitfkERbwnNyzBD8fha1a4PELZEeNoUfNofdyJS2Y19tFgHZQ62tzKwELiBA3xVeZowLr4DJQ8rCgjL",
			origNet:   &chaincfg.SimNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPub:    "xpub661MyMwAqRbcGAp6p8dnh7Z129koDDSzhZ1AQMxMJQJJ48sHanh2YndBWFXCoxEUk71fSwSMRkKa8YG7msVR6k34nnNmRrEEC6zZQ3wh1Ra",
			isPrivate: false,
		},
	}

	for i, test := range tests {
		extKey, err := hdkeychain.NewKeyFromString(test.key)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected error "+
				"creating extended key: %v", i, test.name,
				err)
			continue
		}

		if !extKey.IsForNet(test.origNet) {
			t.Errorf("IsForNet #%d (%s): key is not for expected "+
				"network %v", i, test.name, test.origNet.Name)
			continue
		}

		extKey.SetNet(test.newNet)
		if !extKey.IsForNet(test.newNet) {
			t.Errorf("SetNet/IsForNet #%d (%s): key is not for "+
				"expected network %v", i, test.name,
				test.newNet.Name)
			continue
		}

		if test.isPrivate {
			privStr, _ := extKey.String()
			if privStr != test.newPriv {
				t.Errorf("Serialize #%d (%s): mismatched serialized "+
					"private extended key -- got: %s, want: %s", i,
					test.name, privStr, test.newPriv)
				continue
			}

			extKey, err = extKey.Neuter()
			if err != nil {
				t.Errorf("Neuter #%d (%s): unexpected error: %v ", i,
					test.name, err)
				continue
			}
		}

		pubStr, _ := extKey.String()
		if pubStr != test.newPub {
			t.Errorf("Neuter #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.newPub)
			continue
		}
	}
}

// TestErrors performs some negative tests for various invalid cases to ensure
// the errors are handled properly.
func TestErrors(t *testing.T) {
	// Should get an error when seed has too few bytes.
	net := &chaincfg.MainNetParams
	_, err := hdkeychain.NewMaster(bytes.Repeat([]byte{0x00}, 15), net)
	if err != hdkeychain.ErrInvalidSeedLen {
		t.Errorf("NewMaster: mismatched error -- got: %v, want: %v",
			err, hdkeychain.ErrInvalidSeedLen)
	}

	// Should get an error when seed has too many bytes.
	_, err = hdkeychain.NewMaster(bytes.Repeat([]byte{0x00}, 65), net)
	if err != hdkeychain.ErrInvalidSeedLen {
		t.Errorf("NewMaster: mismatched error -- got: %v, want: %v",
			err, hdkeychain.ErrInvalidSeedLen)
	}

	// Generate a new key and neuter it to a public extended key.
	seed, err := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	if err != nil {
		t.Errorf("GenerateSeed: unexpected error: %v", err)
		return
	}
	extKey, err := hdkeychain.NewMaster(seed, net)
	if err != nil {
		t.Errorf("NewMaster: unexpected error: %v", err)
		return
	}
	pubKey, err := extKey.Neuter()
	if err != nil {
		t.Errorf("Neuter: unexpected error: %v", err)
		return
	}

	// Deriving a hardened child extended key should fail from a public key.
	_, err = pubKey.Child(hdkeychain.HardenedKeyStart)
	if err != hdkeychain.ErrDeriveHardFromPublic {
		t.Errorf("Child: mismatched error -- got: %v, want: %v",
			err, hdkeychain.ErrDeriveHardFromPublic)
	}

	// NewKeyFromString failure tests.
	tests := []struct {
		name      string
		key       string
		err       error
		neuter    bool
		neuterErr error
	}{
		{
			name: "invalid key length",
			key:  "dpub1234",
			err:  hdkeychain.ErrInvalidKeyLen,
		},
		{
			name: "bad checksum",
			key:  "dpubZF6AWaFizAuUcbkZSs8cP8Gxzr6Sg5tLYYM7gEjZMC5GDaSHB4rW4F51zkWyo9U19BnXhc99kkEiPg248bYin8m9b8mGss9nxV6N2QpU8vj",
			err:  hdkeychain.ErrBadChecksum,
		},
		{
			name: "pubkey not on curve",
			key:  "dpubZ9169KDAEUnyoTzA7pDGtXbxpji5LuUk8johUPVGY2CDsz6S7hahGNL6QkeYrUeAPnaJD1MBmrsUnErXScGZdjL6b2gjCRX1Z1GNhRPHBJ8",
			err:  errors.New("pubkey [0,50963827496501355358210603252497135226159332537351223778668747140855667399507] isn't on secp256k1 curve"),
		},
		{
			name:      "unsupported version",
			key:       "4s9bfpYH9CkJboPNLFC4BhTENPrjfmKwUxesnqxHBjv585bCLzVdQKuKQ5TouA57FkdDskrR695Z5U2wWwDUUVWXPg7V57sLpc9dMgx74LpJmxQ5",
			err:       nil,
			neuter:    true,
			neuterErr: chaincfg.ErrUnknownHDKeyID,
		},
	}

	for i, test := range tests {
		extKey, err := hdkeychain.NewKeyFromString(test.key)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("NewKeyFromString #%d (%s): mismatched error "+
				"-- got: %v, want: %v", i, test.name, err,
				test.err)
			continue
		}

		if test.neuter {
			_, err := extKey.Neuter()
			if !reflect.DeepEqual(err, test.neuterErr) {
				t.Errorf("Neuter #%d (%s): mismatched error "+
					"-- got: %v, want: %v", i, test.name,
					err, test.neuterErr)
				continue
			}
		}
	}
}

// TestZero ensures that zeroing an extended key works as intended.
func TestZero(t *testing.T) {
	tests := []struct {
		name   string
		master string
		extKey string
		net    *chaincfg.Params
	}{
		// Test vector 1
		{
			name:   "test vector 1 chain m",
			master: "000102030405060708090a0b0c0d0e0f",
			extKey: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
			net:    &chaincfg.MainNetParams,
		},

		// Test vector 2
		{
			name:   "test vector 2 chain m",
			master: "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			extKey: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
			net:    &chaincfg.MainNetParams,
		},
	}

	// Use a closure to test that a key is zeroed since the tests create
	// keys in different ways and need to test the same things multiple
	// times.
	testZeroed := func(i int, testName string, key *hdkeychain.ExtendedKey) bool {
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
		_, errZeroed := key.String()
		if errZeroed.Error() != wantKey {
			t.Errorf("String #%d (%s): mismatched serialized key "+
				"-- want %s, got %s", i, testName, wantKey,
				errZeroed)
			return false
		}

		wantErr := hdkeychain.ErrNotPrivExtKey
		_, err := key.ECPrivKey()
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("ECPrivKey #%d (%s): mismatched error: want "+
				"%v, got %v", i, testName, wantErr, err)
			return false
		}

		wantErr = errors.New("pubkey string is empty")
		_, err = key.ECPubKey()
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("ECPubKey #%d (%s): mismatched error: want "+
				"%v, got %v", i, testName, wantErr, err)
			return false
		}

		wantAddr := "22u2HQGuGaB5AdL5ir6UUQkbPUsBjZKYhrKZ"
		addr, err := key.Address(&chaincfg.MainNetParams)
		if err != nil {
			t.Errorf("Address #%d (%s): unexpected error: %v", i,
				testName, err)
			return false
		}
		if addr.EncodeAddress() != wantAddr {
			t.Errorf("Address #%d (%s): mismatched address -- want "+
				"%s, got %s", i, testName, wantAddr,
				addr.EncodeAddress())
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
		key, err := hdkeychain.NewMaster(masterSeed, test.net)
		if err != nil {
			t.Errorf("NewMaster #%d (%s): unexpected error when "+
				"creating new master key: %v", i, test.name,
				err)
			continue
		}
		neuteredKey, err := key.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}

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
		key, err = hdkeychain.NewKeyFromString(test.extKey)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected "+
				"error: %v", i, test.name, err)
			continue
		}
		neuteredKey, err = key.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}

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
