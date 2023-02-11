// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"math/big"
	"time"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/wire"
)

// MainNetParams returns the network parameters for the main Decred network.
func MainNetParams() *Params {
	// mainPowLimit is the highest proof of work value a Decred block can
	// have for the main network.  It is the value 2^225 - 1.
	mainPowLimit := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 254), bigOne)

	// genesis block difficulty ratio
	initialDifficulty := big.NewInt(100)

	// genesisBlock defines the genesis block of the block chain which serves as
	// the public transaction ledger for the main network.
	//
	// The genesis block for Decred mainnet, testnet, and simnet are not
	// evaluated for proof of work. The only values that are ever used elsewhere
	// in the blockchain from it are:
	// (1) The genesis block hash is used as the PrevBlock.
	// (2) The difficulty starts off at the value given by Bits.
	// (3) The stake difficulty starts off at the value given by SBits.
	// (4) The timestamp, which guides when blocks can be built on top of it
	//      and what the initial difficulty calculations come out to be.
	//
	// The genesis block is valid by definition and none of the fields within it
	// are validated for correctness.
	genesisBlock := wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:   1,
			PrevBlock: chainhash.Hash{}, // All zero.
			// MerkleRoot: Calculated below.
			StakeRoot:    chainhash.Hash{},
			Timestamp:    time.Unix(1531731600, 0), // Monday, 16-Jul-18 09:00:00 UTC
			Bits:         bigToCompact(new(big.Int).Div(mainPowLimit, initialDifficulty)),
			SBits:        2 * 1e8,                  // 2 Coin
			Nonce:        0x00000000,
			StakeVersion: 0,
		},
		Transactions: []*wire.MsgTx{{
			SerType: wire.TxSerializeFull,
			Version: 1,
			TxIn: []*wire.TxIn{{
				// Fully null.
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
					Tree:  0,
				},
				SignatureScript: hexDecode("0000"),
				Sequence:        0xffffffff,
				BlockHeight:     wire.NullBlockHeight,
				BlockIndex:      wire.NullBlockIndex,
				ValueIn:         wire.NullValueIn,
			}},
			TxOut: []*wire.TxOut{{
				Version: 0x0000,
				Value:   0x00000000,
				PkScript: hexDecode("801679e98561ada96caec2949a5d41c4cab3851e" +
					"b740d951c10ecbcf265c1fd9"),
			}},
			LockTime: 0,
			Expiry:   0,
		}},
	}
	genesisBlock.Header.MerkleRoot = genesisBlock.Transactions[0].TxHashFull()

	return &Params{
		Name:        "mainnet",
		Net:         wire.MainNet,
		DefaultPort: "9666",
		DNSSeeds: []DNSSeed{
			{"seed.excc.co", true},
			{"seed.xchange.me", true},
			{"excc-seed.pragmaticcoders.com", true},
		},
		N: wire.MainEquihashN,
		K: wire.MainEquihashK,

		// Chain parameters
		GenesisBlock:             &genesisBlock,
		GenesisHash:              genesisBlock.BlockHash(),
		PowLimit:                 mainPowLimit,
		PowLimitBits:             bigToCompact(mainPowLimit),
		ReduceMinDifficulty:      false,
		MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
		GenerateSupported:        false,
		MaximumBlockSizes:        []int{393216},
		MaxTxSize:                393216,
		TargetTimePerBlock:       defaultTargetTimePerBlock,
		WorkDiffAlpha:            1,
		WorkDiffWindowSize:       144,
		WorkDiffWindows:          20,
		TargetTimespan:           defaultTargetTimePerBlock * 144, // TimePerBlock * WindowSize
		RetargetAdjustmentFactor: 4,

		// Subsidy parameters.
		BaseSubsidy:              38 * 1e8,
		MulSubsidy:               100000,
		DivSubsidy:               103183,
		SubsidyReductionInterval: 16128, // 4 weeks
		WorkRewardProportion:     7,
		StakeRewardProportion:    3,

		// Checkpoints ordered from oldest to newest.  Note that only the latest
		// checkpoint is provided since with headers first syncing the most recent
		// checkpoint will be discovered before block syncing even starts.
		Checkpoints: []Checkpoint{
			{120, newHashFromStr("0076826c6b55e5c2d27afb67f6d087b782d7c86bea3a5f2a3acbded8f5cdab55")},
			{87550, newHashFromStr("000187f6de1d0a99db6eb832f09d803cd95723b2e95d2a7d051e7a6f3809ba63")},
		},

		// AssumeValid is the hash of a block that has been externally verified
		// to be valid.  It allows several validation checks to be skipped for
		// blocks that are both an ancestor of the assumed valid block and an
		// ancestor of the best header.  This is intended to be updated
		// periodically with new releases.
		//
		// Block 00000000000000001251efb83ad1a5c71351b50fe9195f009cf0bf5a7cd99d52
		// Height: 606000
		AssumeValid: *newHashFromStr("00000000000000001251efb83ad1a5c71351b50fe9195f009cf0bf5a7cd99d52"),

		// MinKnownChainWork is the minimum amount of known total work for the
		// chain at a given point in time.  This is intended to be updated
		// periodically with new releases.
		//
		// Block 0000076fff5da5604856e135e81cfc4459ef3fdbb04b720710d30950091d8b9e
		// Height: 899030
		MinKnownChainWork: hexToBigInt("000000000000000000000000000000000000000000000000000000b60cab914c"),

		// The miner confirmation window is defined as:
		//   target proof of work timespan / target proof of work spacing
		RuleChangeActivationQuorum:     4032, // 10 % of RuleChangeActivationInterval * TicketsPerBlock
		RuleChangeActivationMultiplier: 3,    // 75%
		RuleChangeActivationDivisor:    4,
		RuleChangeActivationInterval:   2016 * 4, // 4 weeks
		Deployments: map[uint32][]ConsensusDeployment{},

		// Enforce current block version once majority of the network has
		// upgraded.
		// 75% (750 / 1000)
		//
		// Reject previous block versions once a majority of the network has
		// upgraded.
		// 95% (950 / 1000)
		BlockEnforceNumRequired: 750,
		BlockRejectNumRequired:  950,
		BlockUpgradeNumToCheck:  1000,

		// AcceptNonStdTxs is a mempool param to either accept and relay non
		// standard txs to the network or reject them
		AcceptNonStdTxs: false,

		// Address encoding magics
		NetworkAddressPrefix: "2",
		PubKeyAddrID:     [2]byte{0x02, 0xdc}, // starts with 2s	-- no such addresses should exist in RL
		PubKeyHashAddrID: [2]byte{0x21, 0xB9}, // starts with 22
		PKHEdwardsAddrID: [2]byte{0x35, 0xcf}, // starts with 2e
		PKHSchnorrAddrID: [2]byte{0x2f, 0x0d}, // starts with 2S
		ScriptHashAddrID: [2]byte{0x34, 0xAF}, // starts with 2c
		PrivateKeyID:     0x80,                // starts with 5 (uncompressed) or K (compressed)

		// BIP32 hierarchical deterministic extended key magics
		HDPrivateKeyID: [4]byte{0x04, 0x88, 0xAD, 0xE4}, // starts with xprv
		HDPublicKeyID:  [4]byte{0x04, 0x88, 0xB2, 0x1E}, // starts with xpub

		// BIP44 coin type used in the hierarchical deterministic path for
		// address generation.
		HDCoinType: 0,

		// Decred PoS parameters
		MinimumStakeDiff:        2 * 1e8, // 2 Coin
		TicketPoolSize:          8192,
		TicketsPerBlock:         5,
		TicketMaturity:          256,
		TicketExpiry:            40960, // 5*TicketPoolSize
		CoinbaseMaturity:        256,
		SStxChangeMaturity:      1,
		TicketPoolSizeWeight:    4,
		StakeDiffAlpha:          1, // Minimal
		StakeDiffWindowSize:     144,
		StakeDiffWindows:        20,
		StakeVersionInterval:    144 * 2 * 7, // ~1 week
		MaxFreshStakePerBlock:   20,          // 4*TicketsPerBlock
		StakeEnabledHeight:      256 + 256,   // CoinbaseMaturity + TicketMaturity
		StakeValidationHeight:   768,         // ~32 hours
		StakeBaseSigScript:      []byte{0x00, 0x00},
		StakeMajorityMultiplier: 3,
		StakeMajorityDivisor:    4,

		// Decred organization related parameters
		BlockOneLedger:              tokenPayouts_MainNetParams(),

		// Sanctioned Politeia keys.
		PiKeys: [][]byte{
			hexDecode("03f6e7041f1cf51ee10e0a01cd2b0385ce3cd9debaabb2296f7e9dee9329da946c"),
			hexDecode("0319a37405cb4d1691971847d7719cfce70857c0f6e97d7c9174a3998cf0ab86dd"),
		},

		seeders: []string{
			"seed.excc.co",
			"seed.xchange.me",
			"excc-seed.pragmaticcoders.com",
		},

		Algorithms: []wire.AlgorithmSpec{
			{Height: 0, HeaderSize: 108, Version: 0, Bits: bigToCompact(mainPowLimit)},
			{
				Height:     87550,
				HeaderSize: wire.MaxBlockHeaderPayload - wire.EquihashSolutionLen,
				Version:    1,
				Bits:       bigToCompact(new(big.Int).Sub(new(big.Int).Lsh(bigOne, 241), bigOne)),
			},
		},
	}
}
