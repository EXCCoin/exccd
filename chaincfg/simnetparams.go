// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"math/big"
	"time"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/wire"
)

// SimNetParams returns the network parameters for the simulation test network.
// This network is similar to the normal test network except it is intended for
// private use within a group of individuals doing simulation testing and full
// integration tests between different applications such as wallets, voting
// service providers, mining pools, block explorers, and other services that
// build on Decred.
//
// The functionality is intended to differ in that the only nodes which are
// specifically specified are used to create the network rather than following
// normal discovery rules.  This is important as otherwise it would just turn
// into another public testnet.
func SimNetParams() *Params {
	// simNetPowLimit is the highest proof of work value a Decred block can have
	// for the simulation test network.  It is the value 2^255 - 1.
	simNetPowLimit := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 255), bigOne)

	// Original settings
	simTicketPoolSize       := uint16(64)
	simCoinbaseMaturity     := uint16(16)
	simTicketMaturity       := uint16(16)
	simStakeVersionInterval := int64(8 * 2 * 7)

	// genesisBlock defines the genesis block of the block chain which serves
	// as the public transaction ledger for the simulation test network.
	genesisBlock := wire.MsgBlock{
		Header: wire.BlockHeader{
			Version:   1,
			PrevBlock: chainhash.Hash{}, // All zero.
			// MerkleRoot: Calculated below.
			StakeRoot:    chainhash.Hash{}, // All zero.
			VoteBits:     0,
			FinalState:   [6]byte{}, // All zero.
			Voters:       0,
			FreshStake:   0,
			Revocations:  0,
			Timestamp:    time.Unix(1401292357, 0), // 2009-01-08 20:54:25 -0600 CST
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
			TxIn: []*wire.TxIn{
				{
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
				},
			},
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
	genesisBlock.Header.MerkleRoot = genesisBlock.Transactions[0].TxHashFull()

	return &Params{
		Name:        "simnet",
		Net:         wire.SimNet,
		DefaultPort: "11998",
		DNSSeeds:    nil, // NOTE: There must NOT be any seeds.
		N:           48,
		K:           5,

		// Chain parameters
		GenesisBlock:             &genesisBlock,
		GenesisHash:              genesisBlock.BlockHash(),
		PowLimit:                 simNetPowLimit,
		PowLimitBits:             bigToCompact(simNetPowLimit),
		ReduceMinDifficulty:      false,
		MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
		GenerateSupported:        true,
		MaximumBlockSizes:        []int{1310720},
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
		// Not set for simnet test network since its chain is dynamic.
		AssumeValid: chainhash.Hash{},

		// MinKnownChainWork is the minimum amount of known total work for the
		// chain at a given point in time.
		//
		// Not set for simnet test network since its chain is dynamic.
		MinKnownChainWork: nil,

		// Consensus rule change deployments.
		//
		// The miner confirmation window is defined as:
		//   target proof of work timespan / target proof of work spacing
		RuleChangeActivationQuorum:     160, // 10 % of RuleChangeActivationInterval * TicketsPerBlock
		RuleChangeActivationMultiplier: 3,   // 75%
		RuleChangeActivationDivisor:    4,
		RuleChangeActivationInterval:   320, // 320 seconds
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
		NetworkAddressPrefix: "S",
		PubKeyAddrID:         [2]byte{0x27, 0x6f}, // starts with Sk
		PubKeyHashAddrID:     [2]byte{0x0e, 0x91}, // starts with Ss
		PKHEdwardsAddrID:     [2]byte{0x0e, 0x71}, // starts with Se
		PKHSchnorrAddrID:     [2]byte{0x0e, 0x53}, // starts with SS
		ScriptHashAddrID:     [2]byte{0x0e, 0x6c}, // starts with Sc
		PrivateKeyID:         0xef,                // starts with 9 (uncompressed) or c (compressed)

		// BIP32 hierarchical deterministic extended key magics
		HDPrivateKeyID: [4]byte{0x04, 0x20, 0xb9, 0x03}, // starts with sprv
		HDPublicKeyID:  [4]byte{0x04, 0x20, 0xbd, 0x3d}, // starts with spub

		// BIP44 coin type used in the hierarchical deterministic path for
		// address generation.
		HDCoinType: 2,

		// Decred PoS parameters
		MinimumStakeDiff:        20000,
		TicketPoolSize:          simTicketPoolSize,
		TicketsPerBlock:         5,
		TicketMaturity:          simTicketMaturity,
		TicketExpiry:            uint32(6 * simTicketPoolSize), // 6*TicketPoolSize
		CoinbaseMaturity:        simCoinbaseMaturity,
		SStxChangeMaturity:      1,
		TicketPoolSizeWeight:    4,
		StakeDiffAlpha:          1,
		StakeDiffWindowSize:     8,
		StakeDiffWindows:        8,
		StakeVersionInterval:    simStakeVersionInterval,
		MaxFreshStakePerBlock:   20,            // 4*TicketsPerBlock
		StakeEnabledHeight:      int64(simCoinbaseMaturity + simTicketMaturity),   // CoinbaseMaturity + TicketMaturity
		StakeValidationHeight:   int64(simCoinbaseMaturity + simTicketPoolSize*2), // CoinbaseMaturity + TicketPoolSize*2
		StakeBaseSigScript:      []byte{0xDE, 0xAD, 0xBE, 0xEF},
		StakeMajorityMultiplier: 3,
		StakeMajorityDivisor:    4,

		// Decred organization related parameters
		//
		// Treasury address is a 3-of-3 P2SH going to a wallet with seed:
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// aardvark adroitness aardvark adroitness
		// briefcase
		// (seed 0x0000000000000000000000000000000000000000000000000000000000000000)
		//
		// This same wallet owns the three ledger outputs for simnet.
		//
		// P2SH details for simnet treasury:
		//
		// redeemScript: 532103e8c60c7336744c8dcc7b85c27789950fc52aa4e48f895ebbfb
		// ac383ab893fc4c2103ff9afc246e0921e37d12e17d8296ca06a8f92a07fbe7857ed1d4
		// f0f5d94e988f21033ed09c7fa8b83ed53e6f2c57c5fa99ed2230c0d38edf53c0340d0f
		// c2e79c725a53ae
		//   (3-of-3 multisig)
		// Pubkeys used:
		//   SkQmxbeuEFDByPoTj41TtXat8tWySVuYUQpd4fuNNyUx51tF1csSs
		//   SkQn8ervNvAUEX5Ua3Lwjc6BAuTXRznDoDzsyxgjYqX58znY7w9e4
		//   SkQkfkHZeBbMW8129tZ3KspEh1XBFC1btbkgzs6cjSyPbrgxzsKqk
		//
		// Organization address is ScuQxvveKGfpG1ypt6u27F99Anf7EW3cqhq
		BlockOneLedger:              tokenPayouts_SimNetParams(),

		// Commands to generate simnet Pi keys:
		// $ treasurykey.go -simnet
		// Private key: 62deae1ab2b1ebd96a28c80e870aee325bed359e83d8db2464ef999e616a9eef
		// Public  key: 02a36b785d584555696b69d1b2bbeff4010332b301e3edd316d79438554cacb3e7
		// WIF        : PsUUktzTqNKDRudiz3F4Chh5CKqqmp5W3ckRDhwECbwrSuWZ9m5fk
		//
		// $ treasurykey.go -simnet
		// Private key: cc0d8258d68acf047732088e9b70e2c97c53f711518042d267fc6975f39b791b
		// Public  key: 02b2c110e7b560aa9e1545dd18dd9f7e74a3ba036297a696050c0256f1f69479d7
		// WIF        : PsUVZDkMHvsH8RmYtCxCWs78xsLU9qAyZyLvV9SJWAdoiJxSFhvFx
		PiKeys: [][]byte{
			hexDecode("02a36b785d584555696b69d1b2bbeff4010332b301e3edd316d79438554cacb3e7"),
			hexDecode("02b2c110e7b560aa9e1545dd18dd9f7e74a3ba036297a696050c0256f1f69479d7"),
		},

		Algorithms: []wire.AlgorithmSpec{
			{Height: 0, HeaderSize: 108, Version: 0, Bits: bigToCompact(simNetPowLimit)},
			{Height: 4, HeaderSize: wire.MaxBlockHeaderPayload - wire.EquihashSolutionLen, Version: 1, Bits: bigToCompact(simNetPowLimit)},
		},

		seeders: nil, // NOTE: There must NOT be any seeds.
	}
}
