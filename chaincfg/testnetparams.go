// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"math/big"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/wire"
)

// TestNet3Params return the network parameters for the test currency network.
// This network is sometimes simply called "testnet".
// This is the third public iteration of testnet.
func TestNet3Params() *Params {
	// testNetPowLimit is the highest proof of work value a Decred block
	// can have for the test network.  It is the value 2^254 - 1.
	testNetPowLimit := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 254), bigOne)

	// genesisBlock defines the genesis block of the block chain which serves as
	// the public transaction ledger for the test network (version 3).
	genesisBlock := wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:   4,
			PrevBlock: chainhash.Hash{},
			// MerkleRoot: Calculated below.
			Timestamp:    time.Unix(1532420489, 0), // Tuesday, 24-Jul-18 08:21:29 UTC
			Bits:         bigToCompact(new(big.Int).Div(testNetPowLimit, big.NewInt(10))),
			SBits:        2 * 1e7, // 0.2 Coin
			Nonce:        0x18aea41a,
			StakeVersion: 0,
		},
		Transactions: []*wire.MsgTx{{
			SerType: wire.TxSerializeFull,
			Version: 1,
			TxIn: []*wire.TxIn{{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
				},
				SignatureScript: []byte{
					0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, 0x45, /* |.......E| */
					0x54, 0x68, 0x65, 0x20, 0x54, 0x69, 0x6d, 0x65, /* |The Time| */
					0x73, 0x20, 0x30, 0x33, 0x2f, 0x4a, 0x61, 0x6e, /* |s 03/Jan| */
					0x2f, 0x32, 0x30, 0x30, 0x39, 0x20, 0x43, 0x68, /* |/2009 Ch| */
					0x61, 0x6e, 0x63, 0x65, 0x6c, 0x6c, 0x6f, 0x72, /* |ancellor| */
					0x20, 0x6f, 0x6e, 0x20, 0x62, 0x72, 0x69, 0x6e, /* | on brin| */
					0x6b, 0x20, 0x6f, 0x66, 0x20, 0x73, 0x65, 0x63, /* |k of sec|*/
					0x6f, 0x6e, 0x64, 0x20, 0x62, 0x61, 0x69, 0x6c, /* |ond bail| */
					0x6f, 0x75, 0x74, 0x20, 0x66, 0x6f, 0x72, 0x20, /* |out for |*/
					0x62, 0x61, 0x6e, 0x6b, 0x73, /* |banks| */
				},
				Sequence: 0xffffffff,
				BlockHeight:     wire.NullBlockHeight,
				BlockIndex:      wire.NullBlockIndex,
				ValueIn:         wire.NullValueIn,
			}},
			TxOut: []*wire.TxOut{
				{
					Value: 0x00000000,
					PkScript: []byte{
						0x41, 0x04, 0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, /* |A.g....U| */
						0x48, 0x27, 0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, /* |H'.g..q0| */
						0xb7, 0x10, 0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, /* |..\..(.9| */
						0x09, 0xa6, 0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, /* |..yb...a| */
						0xde, 0xb6, 0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, /* |..I..?L.| */
						0x38, 0xc4, 0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, /* |8..U....| */
						0x12, 0xde, 0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, /* |..\8M...| */
						0x8d, 0x57, 0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, /* |.W.Lp+k.| */
						0x1d, 0x5f, 0xac, /* |._.| */
					},
				},
			},
			LockTime: 0,
			Expiry:   0,
		}},
	}
	// NOTE: This really should be TxHashFull, but it was defined incorrectly.
	//
	// Since the field is not used in any validation code, it does not have any
	// adverse effects, but correcting it would result in changing the block
	// hash which would invalidate the entire test network.  The next test
	// network should set the value properly.
	genesisBlock.Header.MerkleRoot = genesisBlock.Transactions[0].TxHash()

	return &Params{
		Name:        "testnet",
		Net:         wire.TestNet3,
		DefaultPort: "11999",
		DNSSeeds: []DNSSeed{
			{"testnet-seed.excc.co", true},
		},
		N: 144,
		K: 5,

		// Chain parameters.
		//
		// Note that the minimum difficulty reduction parameter only applies up
		// to and including block height 962927.
		GenesisBlock:             &genesisBlock,
		GenesisHash:              genesisBlock.BlockHash(),
		PowLimit:                 testNetPowLimit,
		PowLimitBits:             bigToCompact(testNetPowLimit),
		ReduceMinDifficulty:      false,
		MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
		GenerateSupported:        true,
		MaximumBlockSizes:        []int{393216},
		MaxTxSize:                393216,
		TargetTimePerBlock:       defaultTargetTimePerBlock,
		WorkDiffAlpha:            1,
		WorkDiffWindowSize:       144,
		WorkDiffWindows:          20,
		TargetTimespan:           defaultTargetTimePerBlock * 144, // TimePerBlock * WindowSize
		RetargetAdjustmentFactor: 4,

		// Subsidy parameters.
		BaseSubsidy:              38 * 1e8, // 38 Coin
		MulSubsidy:               100000,
		DivSubsidy:               103183,
		SubsidyReductionInterval: 16128, // 4 weeks
		WorkRewardProportion:     7,
		StakeRewardProportion:    3,

		// Checkpoints ordered from oldest to newest.  Note that only the latest
		// checkpoint is provided since with headers first syncing the most recent
		// checkpoint will be discovered before block syncing even starts.
		Checkpoints: []Checkpoint{
			{555, newHashFromStr("22d8ad8e7761a6e238388c5b40f85d273748d567a51a6bdd61f5b7ac03cee020")},
		},

		// AssumeValid is the hash of a block that has been externally verified
		// to be valid.  It allows several validation checks to be skipped for
		// blocks that are both an ancestor of the assumed valid block and an
		// ancestor of the best header.  This is intended to be updated
		// periodically with new releases.
		//
		// Block 22d8ad8e7761a6e238388c5b40f85d273748d567a51a6bdd61f5b7ac03cee020
		// Height: 555
		AssumeValid: *newHashFromStr("22d8ad8e7761a6e238388c5b40f85d273748d567a51a6bdd61f5b7ac03cee020"),

		// MinKnownChainWork is the minimum amount of known total work for the
		// chain at a given point in time.  This is intended to be updated
		// periodically with new releases.
		//
		// Block
		// Height:
		MinKnownChainWork: nil,

		// Consensus rule change deployments.
		//
		// The miner confirmation window is defined as:
		//   target proof of work timespan / target proof of work spacing
		RuleChangeActivationQuorum:     4032, // 10 % of RuleChangeActivationInterval * TicketsPerBlock
		RuleChangeActivationMultiplier: 3,    // 75%
		RuleChangeActivationDivisor:    4,
		RuleChangeActivationInterval:   2016 * 4, // 4 weeks
		Deployments:                    map[uint32][]ConsensusDeployment{},

		// Enforce current block version once majority of the network has
		// upgraded.
		// 75% (750 / 1000)
		// Reject previous block versions once a majority of the network has
		// upgraded.
		// 95% (950 / 1000)
		BlockEnforceNumRequired: 750,
		BlockRejectNumRequired:  950,
		BlockUpgradeNumToCheck:  1000,

		// AcceptNonStdTxs is a mempool param to either accept and relay non
		// standard txs to the network or reject them
		AcceptNonStdTxs: true,

		// Address encoding magics
		NetworkAddressPrefix: "T",
		PubKeyAddrID:         [2]byte{0x28, 0xf7}, // starts with Tk
		PubKeyHashAddrID:     [2]byte{0x0f, 0x21}, // starts with Ts
		PKHEdwardsAddrID:     [2]byte{0x0f, 0x01}, // starts with Te
		PKHSchnorrAddrID:     [2]byte{0x0e, 0xe3}, // starts with TS
		ScriptHashAddrID:     [2]byte{0x0e, 0xfc}, // starts with Tc
		PrivateKeyID:         0xef,                // starts with 9 (uncompressed) or c (compressed)

		// BIP32 hierarchical deterministic extended key magics
		HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x97}, // starts with tprv
		HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xd1}, // starts with tpub

		// BIP44 coin type used in the hierarchical deterministic path for
		// address generation.
		HDCoinType: 0,

		// Decred PoS parameters
		MinimumStakeDiff:        2 * 1e7, // 0.2 Coin
		TicketPoolSize:          1024,
		TicketsPerBlock:         5,
		TicketMaturity:          16,
		TicketExpiry:            6144, // 6*TicketPoolSize
		CoinbaseMaturity:        16,
		SStxChangeMaturity:      1,
		TicketPoolSizeWeight:    4,
		StakeDiffAlpha:          1,
		StakeDiffWindowSize:     144,
		StakeDiffWindows:        20,
		StakeVersionInterval:    144 * 2 * 7, // ~3.5 day
		MaxFreshStakePerBlock:   20,          // 4*TicketsPerBlock
		StakeEnabledHeight:      16 + 16,     // CoinbaseMaturity + TicketMaturity
		StakeValidationHeight:   192,         // ~8 hours
		StakeBaseSigScript:      []byte{0x00, 0x00},
		StakeMajorityMultiplier: 3,
		StakeMajorityDivisor:    4,

		// Decred organization related parameters.
		BlockOneLedger: tokenPayouts_TestNet3Params(),

		Algorithms: []wire.AlgorithmSpec{
			{Height: 0, HeaderSize: 108, Version: 0, Bits: bigToCompact(testNetPowLimit)},
			{
				Height:     28,
				HeaderSize: wire.MaxBlockHeaderPayload - wire.EquihashSolutionLen,
				Version:    1,
				Bits:       bigToCompact(new(big.Int).Sub(new(big.Int).Lsh(bigOne, 252), bigOne)),
			},
		},

		// Sanctioned Politeia keys.
		PiKeys: [][]byte{
			hexDecode("03beca9bbd227ca6bb5a58e03a36ba2b52fff09093bd7a50aee1193bccd257fb8a"),
			hexDecode("03e647c014f55265da506781f0b2d67674c35cb59b873d9926d483c4ced9a7bbd3"),
		},

		seeders: nil,
	}
}
