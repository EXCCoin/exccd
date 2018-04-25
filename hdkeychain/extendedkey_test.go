// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package hdkeychain_test

// References:
//   [BIP32]: BIP0032 - Hierarchical Deterministic Wallets
//   https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

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
			wantPub:  "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EHxoeTG",
			wantPriv: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGByB3yWc",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "xpub69GmQNHCKJbkH4WHGtBXeMZjvvhWKxTs3SSsgRqx1eMJLFjiXKcTfk1veJvjAuKXxBN5j7pgeL9Umsjoz68TMCqiwrR9cxfP4xcKBtxkb9d",
			wantPriv: "xprv9vHQzrkJUw3T4aRpAreXHDd1Nts1vVk1gDXGt3SLTJpKTTQZynJD7whSo354KRNtjV5GjEayekZxiicFPQqcTMTS2PZon5xTVvWsH8qAior",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "xpub6BEFPjMnr89vRYCeMWuSzF7iCtpFA2S1jtvnpy4FL5MCZcmpVV3W12fir9YpfEFGDjBCSbKsRhXdHneZUWudeDGwiLwfmVZD7hkEUDrELiV",
			wantPriv: "xprv9xEtzDpu1kbdD48BFVNSd7AyerykkZiANg1C2aedmjpDgpSfwwjFTEMEzrRdzjNKVQ22U7EVQER3nGJQXrEsqsZA1TDCWjcsajSNoPsNrd2",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "xpub6DMFwjbJmTo7DakqdNELmY2gqksoXjYiv2Me624iEsmrSdWw8oYQV2Lqek1MQchf7352P38c1jzZd3tknUQy6UJXcJpiZTUSZa2JawRdUpx",
			wantPriv: "xprv9zMuYE4Qw6Ep16gNXLhLQQ5xHj3K8GpsYoS3Hdf6gYEsZqBnbGE9wE2MoTuxGUttXeqWPe1Y2k8utQ94ij7sMmsZg8gfNgiMvBYXTVu9ha6",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "xpub6EvN2iY38NF8k7MVt7S99MSPFS4HW4aJvTbDu9zH6dzxftydQesJB7MeacwcdyXrHBRskjE2Fn8uRxAc61RtorFMKcAKQB525ZrFuySrjqh",
			wantPriv: "xprvA1w1dD19HzgqXdH2n5u8nDVehQDo6brTZEfd6mafYJTyo6eUs7Z3dK3AjLGHYhynJHCtFEra8rqJc4yN8PEwYwwAstZAwJEpaTvo9N9XNkk",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "xpub6H6tydVkGNmeTMdW3MQo244wzmyfG9sCrEbdeHCj1GWUgt9uKjzjmS8QQTY6Wpi7Ryp64eU5qfa6gKUi3CJReuedH7FMKpojCWqwLCC4cL7",
			wantPriv: "xprvA47Ya7xrS1DMEsZ2wKsnev8DSk9Arh9MV1g2qto7SvyVp5pknCgVDdovZCj2VDCQanhoqB9n2b72GBtpWndMfhpsJLzMBP9cAnrv4XJFtGG",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WCzUQn6",
			wantPriv: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsu94ep",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterHex,
			path:     []uint32{0},
			wantPub:  "xpub68ALFFUDaY6t35Rn6DcEnFUaDGyV8KViC2AU6WUYuUk2vQcN57pQbqXSnavAHZ7vcgaVRvyYQEd7qBAspsWbuFJsb89Lm4Q9JaMgoZG3wAa",
			wantPriv: "xprv9uAyqjwKkAYapbMJzC5ER7XqfF8zirmrpoEsJ84wM9D43cHDXaWA43CxwKGr5ujPDywqoPAfg3NubkBk7zx8MJ8j82yHPXKyoGVxnL9V1XT",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647},
			wantPub:  "xpub6AUeabxcoN1BhRp9xeaNxzYbcjVWTGNbDeZcvoNjRDK41BGPaTabdwTXg6Qw2pawbHbXdAGsVGgomrzbsqJtM3hRW6C3zXQWiKWRHKCVCVN",
			wantPriv: "xprv9wVJB6RixzStUwjgrd3Nbrbs4hf23oejrRe28Qy7rsn58NwF2vGM6993po1UohKnKJNE7np2WsX7Jzsu8RsReFWCvedH2mob8QidBdqPvsH",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1},
			wantPub:  "xpub6C4D8jqTGdUu8otMFpTk2jbkLKTAAYDwBuENusnAUjnVAd9bHXuA96kueTbEz8CUaq722e934aG5GCCj5ZNLidPkewz4KRhcEKJ3CYW2A1v",
			wantPriv: "xprv9y4rjEJZSFvbvKot9nvjfbf1nHcfm5W5pgJn7VNYvQFWHppSjzaubJSRoACGzqSj2cmsRvfLrbTm3EcAU92JkyYvczkEzEdsCSG2L4b3xcD",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646},
			wantPub:  "xpub6E3QCHQ1oSon5yPoph46FeeFxQAfPhr9HcDECnkWvuPtcRGtwsaC4SNKWvYyhrjT8ZabCQ5nuzTrjLZc74CinziTnketHNGYJBQ21pct3pJ",
			wantPriv: "xprvA143nms7y5FUsVKLifX5tWhXQNLAzF8HvPHdQQLuNZrujcwkQLFwWe3qffku9cqYAUtW71KP3mvGfru2KhefoprMJqdtY691jfHwgHRtbb8",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H/2",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646, 2},
			wantPub:  "xpub6FoT5r8Je2oYgtWpN1ihgTjD1YWFwzu7UobiGkT3PDXfHoZMMMmn2m67JDiBC5yP9kFTkDPg8WD9R2MB29XFQ2LDbZP1SVHL7uUYct483Bc",
			wantPriv: "xprvA2p6gLbQofFFUQSMFzBhKKnUTWfmYYBG7ag7UN3RpszgR1ECopTXUxmdSyWpfHoqhvCsfekfpgVFufWc91eX9aMFHL3RpFznaTWqhB9amxp",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 1 - Testnet
		{
			name:     "test vector 1 chain m - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "tpubVhnMyQmZAhoosedBTX7oacwyCNc5qtdEMoNHudUCW1R6WZTvqCZQoNJHSn4H11puwdk4qyDv2ET637EDap4r8HH3odjBC5nEjmnPcsDfLwm",
			wantPriv: "tprvZUo1ZuEfLLFWfAYiMVaoDV1EeLmbSRuNzaSh7F4awft7dm8nHfFAFZyobWQyV8Qr26r8M2CmNw6nEb35HaECWFGy1vzx2ZGdyfBeaaHudoi",
			net:      &chaincfg.TestNet2Params,
		},
		{
			name:     "test vector 1 chain m/0H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "tpubVm3mQR7aeaowtpbnJKRnvpKsJ3A3q5u31EPY1FAU5Re1zvFfmFUE5aHXY4qmjfQcb1uFZf8mvvU1vi89vEzHPQfR5NETLqByzdthaYfQGja",
			wantPriv: "tprvZY4QzuagpDFegLXKCHtnZgP8k1KZRdBBe1TwCrkrX67387vXDi9yXmy3gnz6tBTyNKcSZmu4wLtVsYzbKZhSVZH89uP7VxV4RboFfozTBMQ",
			net:      &chaincfg.TestNet2Params,
		},
		{
			name:     "test vector 1 chain m/0H/1 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "tpubVo1FPnCBBQN83JJ9Nx9iGhsqa1Gnf9sBhgsT9nNmPrdvEHHmjQuGQrwKjuTsDzLLrZiNH8dxiHrASd2uQfmTgR6dqrkyVN5p3P2crvfgpEQ",
			wantPriv: "tprvZa1tzGfHM2opppDgGvchuZw71ySJFh9LLTwrMPy9qX6wMUxdBsb1s4cqtcLgZVTQ8EZCJeYagpjaw6gkU16ht5Nr8y2WEc9UWQimC4Y6MFs",
			net:      &chaincfg.TestNet2Params,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "tpubVq8FwnRh6k1JqLrLeoUc3znpCsLM2rytspJJQqPEJf4a7J2tNjQAtrcSYVvPyNnjjscCDaShJLK6mtH6idGo8g8Djpe2HL13VFJgygvRmc9",
			wantPriv: "tprvZc8uYGtoGNT1crmsYmwbgrr5eqVrdQG3WbNhcSyckKXbEVhjqC5vM4HxhDpzqEyyAVNgEBKdKLTT3EXQesyhPyhFoeVy6ZExqrpurCCRvrF",
			net:      &chaincfg.TestNet2Params,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "tpubVrhN2mNRTeTLMsSzuYgQRpCWcYWq1C1UtFXtDyJoARHgLZVaeaj4awdFUNrfCjcvv1y3bGY7YNTSanYx2AHir453T7yd83bd1F8eJgtdq3F",
			wantPriv: "tprvZdi1dFqXdGu39PNXoX9Q4gFn4WgLbjHdX2cHRauBc5khTmAS73Qp39Jmd6BL7U4rw7k45nAfRT9qkuMi4Y6mb9ks1QNUfAmRW9DBY5g5ELT",
			net:      &chaincfg.TestNet2Params,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "tpubVtstygL8beyr57j14nf4JWq5MtSCmHJNp2YHy6XF53oCMYfrZfrWBGQ1JDT95aoC4pMFuBnB8Ftdq9s3yMAFh7UKQd4f3hLL8C8KipwSek6",
			wantPriv: "tprvZftYaAoEmHRYrdeXxm83wNtLorbiMpaXSochAi7dWiGDUkLi28YFdU5XSxe53yHVDdEyfiTsKBRZR2HASwVBhueZRroeuFgD6U9JTC1mUyU",
			net:      &chaincfg.TestNet2Params,
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
	testVec1MasterPrivKey := "dprv3hCznBesA6jBucms1ZhyGeFfvJfBSwfs7ZFrxS8tdYzbjDZe2UwSaL7EbYo1qa88DmtyyG5cL9tdGxHkD89JmeZTbz5sVYU4Dgtijiio4Sc"
	testVec2MasterPrivKey := "dprv3hCznBesA6jBtPKJbQTxRZAKG2gyj8tZKEPaCsV4e9YYFBAgRP2eTSPAeu4r8dTMt9q51j2Vdt5zNqj7jbtovvocrP1qLj6WUTLF9xYQt4y"

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
			wantPriv: "dprv3hCznBesA6jBucms1ZhyGeFfvJfBSwfs7ZFrxS8tdYzbjDZe2UwSaL7EbYo1qa88DmtyyG5cL9tdGxHkD89JmeZTbz5sVYU4Dgtijiio4Sc",
		},
		{
			name:     "test vector 1 chain m/0",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "dprv3jFfEhxvVxy6NJWopujhfg7syQL71xCRgNoGUpQTtjTpCwzigwtCwssQGbRQsby7PBs1Yp8Wu7isu396qeNof13EZuxbCTJVF1xkoFAQHWj",
		},
		{
			name:     "test vector 1 chain m/0/1",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1},
			wantPriv: "dprv3mWLns1v1fdLhxStaJHh3BqxmTi14RHHeWdNU6oU8sSkTDmAr54yK6La2APy3rAZr9ZJAdm5asTJaqBZ3vBYVSPHqyL8kbcCp5jgqfxBs4x",
		},
		{
			name:     "test vector 1 chain m/0/1/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2},
			wantPriv: "dprv3oDxSziXR1rQVWwWWBRKgCQU3vN6dnR3ekzHzvZRdgfVvrYSE35saJh8UdfSxCgtMn7pnbeMXWbyBbwxoncC9LMrnuH1AoJSB259c4XgmnN",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2},
			wantPriv: "dprv3rYHNih25i8MqeRjhFq8mLnK4a3J63a9zkYTCFJd8kaJuNc5aDAAuG1XopkU7h93HvfbNvQaWdQLtwFmUEbDN3GCZ2Mxw6tq5ZSh8d1Chyw",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2/1000000000",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2, 1000000000},
			wantPriv: "dprv3tKkzgLFKaX2VcQTir9JeNHWxikiKokSJtdyj7sYoiDkU3np2rc3DGYPRVmDhb2FFaAk98fnqRotQYTVCRaoyAZiHaoyNoPCFeYA9pEshBT",
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterPrivKey,
			path:     []uint32{},
			wantPriv: "dprv3hCznBesA6jBtPKJbQTxRZAKG2gyj8tZKEPaCsV4e9YYFBAgRP2eTSPAeu4r8dTMt9q51j2Vdt5zNqj7jbtovvocrP1qLj6WUTLF9xYQt4y",
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "dprv3jMy45BuuDETfxi59P8NTSjHPrNVq4wPRfLgRd57923L2hosj5NUEqiLYQ4i7fJtUpiXZLr2wUeToJY2Tm5sCpAJdajEHDmieVJiPQNXwu9",
		},
		{
			name:     "test vector 2 chain m/0/2147483647",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647},
			wantPriv: "dprv3mgHPRgAnNboAb9edL4RPscKYrNLG77BhPvFe3eiTGPSiigDeXct3WeiZ2QqRrm9TiseBuYWGEG79xkBzazpBGfym1vRXjcEo5KUi4rbhZ1",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1},
			wantPriv: "dprv3onqUUAjN1xQUAW1BzwiBa5wisDCE8hX8YcAEgVe1DPvbgHVLauDt2NsZPQX6tJs6ozcQSU9GdsffhueTbxTxFMPEMPxM2iHiooMz2oQHWS",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646},
			wantPriv: "dprv3r8NTJgGzAjY5cU7sLo5rhT8o2wdco2iqjks6nJoiDACTucrPMcsciccv9skwGMX69uRa8EaZofskV7YyzBDbVi6v4RXbJ4DyeZ6JpUgdUi",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPriv: "dprv3sp4xvFP9mL9UUEddSZUxrtNnhe5UcHs5wrpxdZVEFCXoT4EpYeHZJjCDhvVEQFK2KfSHXFmew6MeBuvtrJfQv1BnkiSV7xxUji66uvWasp",
		},

		// Custom tests to trigger specific conditions.
		{
			// Seed 000000000000000000000000000000da.
			name:     "Derived privkey with zero high byte m/0",
			master:   "dprv3jFfEhxvVxy6NJWopujhfg7syQL71xCRgNoGUpQTtjTpCwzigwtCwssQGbRQsby7PBs1Yp8Wu7isu396qeNof13EZuxbCTJVF1xkoFAQHWj",
			path:     []uint32{0},
			wantPriv: "dprv3mWLns1v1fdLfeu5DKTA6NWQHLF6pFsPSwKCS6q4h4nkjm2DfuH5X2iDnW15jhHTGa3rzxSpvskuXugcbBcUUVWCETKKzjW7ja4V2jL4aw4",
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
	testVec1MasterPubKey := "dpubZF8BRmciAzYoTjXZ3bbRWLVCwUKtTquact3Tr6ye77Rgmw76VyqMb9TB9KpfrvUYEM5d1Au4fQzE2BbtxRjwzGsqnWHmtQP9UV1kxZaqvb6"
	testVec2MasterPubKey := "dpubZF4LSCdF9YKZfNzTVYhz4RBxsjYXqms8AQnMBHXZ8GUKoRSigG7kQnKiJt5pzk93Q8FxcdVBEkQZruSXduGtWnkwXzGnjbSovQ97dCxqaXc"

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
			wantPub: "dpubZF8BRmciAzYoTjXZ3bbRWLVCwUKtTquact3Tr6ye77Rgmw76VyqMb9TB9KpfrvUYEM5d1Au4fQzE2BbtxRjwzGsqnWHmtQP9UV1kxZaqvb6",
		},
		{
			name:    "test vector 1 chain m/0",
			master:  testVec1MasterPubKey,
			path:    []uint32{0},
			wantPub: "dpubZHm6cmVU9pvfDCe3BY7iESzsEnV6xfi4DfoYvycnWLM9cryzKA84DqJ2CphYq6cfiEXgo9C3YLJA4ou81mavw9NDtNc3bLCWVqJz8Fx8qxB",
		},
		{
			name:    "test vector 1 chain m/0/1",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1},
			wantPub: "dpubZKtA6UTDuxeXV2PcYqoe68u7cgDhbTNbA4dUJoaAvfWzuCcRQCyG5S6dbpDZb2p3B5Y2XxLtD94Nemc8QRV4RspmvGwHvE2FZsfE5Pqpeor",
		},
		{
			name:    "test vector 1 chain m/0/1/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2},
			wantPub: "dpubZMwLXm5dRVEJRvJHU8gNV7RwHeXMRRUnYFD4f6C8uNFfqksD1FCDARTwNPsQB3Pg4LuoKXkZbPnE6woUyedwNYVPvZToT5x4Kt6rs4GKa9c",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2},
			wantPub: "dpubZPfASfojwk6MhtAtkM6wPdQBr1ycVjoyqs3N51zR1keK6FcBhjBTtdW3Wn3kDLBZqgLnGozu8Gh3FV8GrFGpu3knmGVoF1Z6yGdqLU1Rz1S",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2/1000000000",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2, 1000000000},
			wantPub: "dpubZR5Pf8cbUGikESevygwydenBaTsgcvoYnRSi7tygu23PxmVEG4GeMQj54oHFoPyRdt7Pg4sMad56yprQszbNyZVewaNEhDkn112C3mqB1fd",
		},

		// Test vector 2
		{
			name:    "test vector 2 chain m",
			master:  testVec2MasterPubKey,
			path:    []uint32{},
			wantPub: "dpubZF4LSCdF9YKZfNzTVYhz4RBxsjYXqms8AQnMBHXZ8GUKoRSigG7kQnKiJt5pzk93Q8FxcdVBEkQZruSXduGtWnkwXzGnjbSovQ97dCxqaXc",
		},
		{
			name:    "test vector 2 chain m/0",
			master:  testVec2MasterPubKey,
			path:    []uint32{0},
			wantPub: "dpubZHJs2Z3PtHbbpaXQCi5wBKPhU8tC5ztBKUYBCYNGKk8eZ1EmBs3MhnLJbxHFMAahGnDnZT7qZxC7AXKP8PB6BDNUZgkG77moNMRmXyQ6s6s",
		},
		{
			name:    "test vector 2 chain m/0/2147483647",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647},
			wantPub: "dpubZJgFEUcAZawGaLZdFEX6FfQBQVgU4bUC5qvDERUTD5dfcB2AQPnJ1dKp1R2DrAzC36BznZG43317s2oBJv3PuaZmA6HqmwMu6vNna4Gfumf",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1},
			wantPub: "dpubZLbgtFNyjt3k2cJtg4a3dD2iXPKFTLgNKP8rLC1p5UE3AyfRHLTcYrZ6brg8eUmGvKRrXZ7A3XyVfwGxvYtjfz8514dUoJPkmSnBmC6qQK6",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646},
			wantPub: "dpubZNyWTupEG35S6d4uN93vWXpGxQuxtW9zuThQbnWpWTHwRCzxREqSSc9eDYivRGiZnEkEhPece5ciSoHtW6Khc729f6eAxjPnBgU38U9hgYw",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPub: "dpubZRuRErXqhdJaZWD1AzXB6d5w2zw7UZ7ALxiS1gHbnQbVEohBzQzsVwGRzq97pmuE7ToA6DGn2QTH4DexxzdnMvkiYUpk8Nh2KEuYUM2RCeU",
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
			extKey:    "xprv9s21ZrQH143K3GqwVJofp3rvzLsJ8J6RHt1L85LoDyZifXqcUeSPnMThY1QHsSorPoYcbyCs18A2kpyfw94Wpu3LHUV2ZAiHJ452vYkjjAt",
			isPrivate: true,
			parentFP:  0,
			privKey:   "33a63922ea4e6686c9fc31daf136888297537f66c1aabe3363df06af0b8274c7",
			pubKey:    "039f2e1d7b50b8451911c64cf745f9ba16193b319212a64096e5679555449d8f37",
			address:   "2saAJkfGvLsBmkXkZaCAo84XCvv8bnx6nGkq",
		},
		{
			name:       "test vector 2 chain m/0/2147483647/1/2147483646/2",
			extKey:     "dpubZRuRErXqhdJaZWD1AzXB6d5w2zw7UZ7ALxiS1gHbnQbVEohBzQzsVwGRzq97pmuE7ToA6DGn2QTH4DexxzdnMvkiYUpk8Nh2KEuYUM2RCeU",
			isPrivate:  false,
			parentFP:   4220580796,
			privKeyErr: hdkeychain.ErrNotPrivExtKey,
			pubKey:     "03dceb0b07698ec3d6ac08ae7297e7f5e63d7fda99d3fce1ded31d36badcdd4d36",
			address:    "2sZyo3fV2RKL8HwfChMam279rzJqFURtpCT4",
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
			key:       "xprv9s21ZrQH143K3gjdi76nKycGU7vJokj9LL5ZbyYjk4mKBLY93FNmzzJhexcj9R9vH8JDkYtBVkoyuTWmUA5EXppdDgzbNgo81gr9nNNw4dw",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.SimNetParams,
			newPriv:   "sprvZ9xkGEZkBei2p9e1uBZRQMGtfGEQNGApP1W19PyNRqg9nuEs2X4ynkvAXWaBiGb5WKiaqcbiKgmyB1HYgcX3mnxiUs7UWeWEfe4tnSpbXLv",
			newPub:    "spubVNx6fk6e22GL2diV1D6RmVDdDJ4tmitfkERbwnNyzBD8fha1a4PELZEeNoUfNofdyJS2Y19tFgHZQ62tzKwELiBA3xVeZowLr4DJQ7xGuao",
			isPrivate: true,
		},
		{
			name:      "simnet -> mainnet",
			key:       "sprvZ9xkGEZkBei2p9e1uBZRQMGtfGEQNGApP1W19PyNRqg9nuEs2X4ynkvAXWaBiGb5WKiaqcbiKgmyB1HYgcX3mnxiUs7UWeWEfe4tnSpbXLv",
			origNet:   &chaincfg.SimNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPriv:   "xprv9s21ZrQH143K3gjdi76nKycGU7vJokj9LL5ZbyYjk4mKBLY93FNmzzJhexcj9R9vH8JDkYtBVkoyuTWmUA5EXppdDgzbNgo81gr9nNNw4dw",
			newPub:    "xpub661MyMwAqRbcGAp6p8dnh7Z129koDDSzhZ1AQMxMJQJJ48sHanh2YndBWFXCoxEUk71fSwSMRkKa8YG7msVR6k34nnNmRrEEC6zZQ4x1ZYo",
			isPrivate: true,
		},

		// Public extended keys.
		{
			name:      "mainnet -> simnet",
			key:       "xpub661MyMwAqRbcGAp6p8dnh7Z129koDDSzhZ1AQMxMJQJJ48sHanh2YndBWFXCoxEUk71fSwSMRkKa8YG7msVR6k34nnNmRrEEC6zZQ4x1ZYo",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.SimNetParams,
			newPub:    "spubVNx6fk6e22GL2diV1D6RmVDdDJ4tmitfkERbwnNyzBD8fha1a4PELZEeNoUfNofdyJS2Y19tFgHZQ62tzKwELiBA3xVeZowLr4DJQ7xGuao",
			isPrivate: false,
		},
		{
			name:      "simnet -> mainnet",
			key:       "spubVNx6fk6e22GL2diV1D6RmVDdDJ4tmitfkERbwnNyzBD8fha1a4PELZEeNoUfNofdyJS2Y19tFgHZQ62tzKwELiBA3xVeZowLr4DJQ7xGuao",
			origNet:   &chaincfg.SimNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPub:    "xpub661MyMwAqRbcGAp6p8dnh7Z129koDDSzhZ1AQMxMJQJJ48sHanh2YndBWFXCoxEUk71fSwSMRkKa8YG7msVR6k34nnNmRrEEC6zZQ4x1ZYo",
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
			key:  "dpubZ9169KDAEUnyoTzA7pDGtXbxpji5LuUk8johUPVGY2CDsz6S7hahGNL6QkeYrUeAPnaJD1MBmrsUnErXScGZdjL6b2gjCRX1Z1GNhLdVCjv",
			err:  errors.New("pubkey [0,50963827496501355358210603252497135226159332537351223778668747140855667399507] isn't on secp256k1 curve"),
		},
		{
			name:      "unsupported version",
			key:       "4s9bfpYH9CkJboPNLFC4BhTENPrjfmKwUxesnqxHBjv585bCLzVdQKuKQ5TouA57FkdDskrR695Z5U2wWwDUUVWXPg7V57sLpc9dMgx74LsVZGEB",
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
			extKey: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGByB3yWc",
			net:    &chaincfg.MainNetParams,
		},

		// Test vector 2
		{
			name:   "test vector 2 chain m",
			master: "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			extKey: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsu94ep",
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

		wantAddr := "2sZw5xfBdXXCzFWeaDSuiDx4YiXy3t8xSu2m"
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
