// Copyright (c) 2018-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"math/big"
	"time"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/wire"
)

// RegNetParams returns the network parameters for the regression test network.
// This should not be confused with the public test network or the simulation
// test network.  The purpose of this network is primarily for unit tests and
// RPC server tests.  On the other hand, the simulation test network is intended
// for full integration tests between different applications such as wallets,
// voting service providers, mining pools, block explorers, and other services
// that build on Decred.
//
// Since this network is only intended for unit testing, its values are subject
// to change even if it would cause a hard fork.
func RegNetParams() *Params {
	// regNetPowLimit is the highest proof of work value a Decred block
	// can have for the regression test network.  It is the value 2^255 - 1.
	regNetPowLimit := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// genesisBlock defines the genesis block of the block chain which serves as
	// the public transaction ledger for the regression test network.
	genesisBlock := wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:   1,
			PrevBlock: chainhash.Hash{}, // All zero.
			// MerkleRoot: Calculated below.
			StakeRoot:    chainhash.Hash{}, // All zero.
			VoteBits:     0,
			FinalState:   [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Voters:       0,
			FreshStake:   0,
			Revocations:  0,
			Timestamp:    time.Unix(1538524800, 0), // 2018-10-03 00:00:00 +0000 UTC
			PoolSize:     0,
			Bits:         0x207fffff, // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
			SBits:        0,
			Nonce:        0,
			StakeVersion: 0,
			Height:       0,
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
		Name:        "regnet",
		Net:         wire.RegNet,
		DefaultPort: "11997",
		DNSSeeds:    nil, // NOTE: There must NOT be any seeds.

		// Chain parameters
		GenesisBlock:             &genesisBlock,
		GenesisHash:              genesisBlock.BlockHash(),
		PowLimit:                 regNetPowLimit,
		PowLimitBits:             bigToCompact(regNetPowLimit),
		ReduceMinDifficulty:      false,
		MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
		GenerateSupported:        true,
		MaximumBlockSizes:        []int{1000000, 1310720},
		MaxTxSize:                1000000,
		TargetTimePerBlock:       time.Second,
		WorkDiffAlpha:            1,
		WorkDiffWindowSize:       8,
		WorkDiffWindows:          4,
		TargetTimespan:           time.Second * 8, // TimePerBlock * WindowSize
		RetargetAdjustmentFactor: 4,

		// Subsidy parameters.
		BaseSubsidy:              50000000000,
		MulSubsidy:               100,
		DivSubsidy:               101,
		SubsidyReductionInterval: 128,
		WorkRewardProportion:     6,
		StakeRewardProportion:    3,

		// Checkpoints ordered from oldest to newest.
		Checkpoints: nil,

		// AssumeValid is the hash of a block that has been externally verified
		// to be valid.
		//
		// Not set for regression test network since its chain is dynamic.
		AssumeValid: chainhash.Hash{},

		// MinKnownChainWork is the minimum amount of known total work for the
		// chain at a given point in time.
		//
		// Not set for regression test network since its chain is dynamic.
		MinKnownChainWork: nil,

		// Consensus rule change deployments.
		//
		// The miner confirmation window is defined as:
		//   target proof of work timespan / target proof of work spacing
		RuleChangeActivationQuorum:     160, // 10 % of RuleChangeActivationInterval * TicketsPerBlock
		RuleChangeActivationMultiplier: 3,   // 75%
		RuleChangeActivationDivisor:    4,
		RuleChangeActivationInterval:   320, // Full ticket pool -- 320 seconds
		Deployments: map[uint32][]ConsensusDeployment{},

		// Enforce current block version once majority of the network has
		// upgraded.
		// 51% (51 / 100)
		// Reject previous block versions once a majority of the network has
		// upgraded.
		// 75% (75 / 100)
		BlockEnforceNumRequired: 51,
		BlockRejectNumRequired:  75,
		BlockUpgradeNumToCheck:  100,

		// AcceptNonStdTxs is a mempool param to either accept and relay non
		// standard txs to the network or reject them
		AcceptNonStdTxs: true,

		// Address encoding magics
		NetworkAddressPrefix: "R",
		PubKeyAddrID:         [2]byte{0x25, 0xe5}, // starts with Rk
		PubKeyHashAddrID:     [2]byte{0x0e, 0x00}, // starts with Rs
		PKHEdwardsAddrID:     [2]byte{0x0d, 0xe0}, // starts with Re
		PKHSchnorrAddrID:     [2]byte{0x0d, 0xc2}, // starts with RS
		ScriptHashAddrID:     [2]byte{0x0d, 0xdb}, // starts with Rc
		PrivateKeyID:         0xef,                // starts with 9 (uncompressed) or c (compressed)

		// BIP32 hierarchical deterministic extended key magics
		HDPrivateKeyID: [4]byte{0xea, 0xb4, 0x04, 0x48}, // starts with rprv
		HDPublicKeyID:  [4]byte{0xea, 0xb4, 0xf9, 0x87}, // starts with rpub

		// BIP44 coin type used in the hierarchical deterministic path for
		// address generation.
		HDCoinType: 1,

		// Decred PoS parameters
		MinimumStakeDiff:        20000,
		TicketPoolSize:          64,
		TicketsPerBlock:         5,
		TicketMaturity:          16,
		TicketExpiry:            384, // 6*TicketPoolSize
		CoinbaseMaturity:        16,
		SStxChangeMaturity:      1,
		TicketPoolSizeWeight:    4,
		StakeDiffAlpha:          1,
		StakeDiffWindowSize:     8,
		StakeDiffWindows:        8,
		StakeVersionInterval:    8 * 2 * 7,
		MaxFreshStakePerBlock:   20,            // 4*TicketsPerBlock
		StakeEnabledHeight:      16 + 16,       // CoinbaseMaturity + TicketMaturity
		StakeValidationHeight:   16 + (64 * 2), // CoinbaseMaturity + TicketPoolSize*2
		StakeBaseSigScript:      []byte{0x73, 0x57},
		StakeMajorityMultiplier: 3,
		StakeMajorityDivisor:    4,

		BlockOneLedger:          tokenPayouts_RegNetParams(),

		PiKeys: [][]byte{},

		Algorithms: []wire.AlgorithmSpec{
			{Height: 0, HeaderSize: 108, Version: 0, Bits: bigToCompact(regNetPowLimit)},
			{Height: 4, HeaderSize: wire.MaxBlockHeaderPayload - wire.EquihashSolutionLen, Version: 1, Bits: bigToCompact(regNetPowLimit)},
		},

		seeders: nil, // NOTE: There must NOT be any seeds.
	}
}
