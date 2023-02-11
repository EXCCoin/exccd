// Copyright (c) 2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package stdaddr_test

import (
	"fmt"

	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/txscript/v4/stdaddr"
)

// This example demonstrates decoding addresses, generating their payment
// scripts and associated script versions, determining supported capabilities by
// checking if interfaces are implemented, obtaining the associated underlying
// hash160 for addresses that support it, converting public key addresses to
// their public key hash variant, and generating stake-related scripts for
// addresses that can be used in the staking system.
func ExampleDecodeAddress() {
	// Ordinarily addresses would be read from the user or the result of a
	// derivation, but they are hard coded here for the purposes of this
	// example.
	netParams := chaincfg.TestNet3Params()
	addrsToDecode := []string{
		"TsbbBthkKvpMEd55BU3ndzGH9w8rX1HmQ3V",
	}
	for idx, encodedAddr := range addrsToDecode {
		addr, err := stdaddr.DecodeAddress(encodedAddr, netParams)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Obtain the payment script and associated script version that would
		// ordinarily by used in a transaction output to send funds to the
		// address.
		scriptVer, script := addr.PaymentScript()
		fmt.Printf("addr%d: %s\n", idx, addr)
		fmt.Printf("  payment script version: %d\n", scriptVer)
		fmt.Printf("  payment script: %x\n", script)

		// Access the RIPEMD-160 hash from addresses that involve it.
		if h160er, ok := addr.(stdaddr.Hash160er); ok {
			fmt.Printf("  hash160: %x\n", *h160er.Hash160())
		}

		// Demonstrate converting public key addresses to the public key hash
		// variant when supported.  This is primarily provided for convenience
		// when the caller already happens to have the public key address handy
		// such as in cases where public keys are shared through some protocol.
		if pkHasher, ok := addr.(stdaddr.AddressPubKeyHasher); ok {
			fmt.Printf("  p2pkh addr: %s\n", pkHasher.AddressPubKeyHash())
		}

		// Obtain stake-related scripts and associated script versions that
		// would ordinarily be used in stake transactions such as ticket
		// purchases and votes for supported addresses.
		//
		// Note that only very specific addresses can be used as destinations in
		// the staking system and this approach provides a capabilities based
		// mechanism to determine support.
		if stakeAddr, ok := addr.(stdaddr.StakeAddress); ok {
			// Obtain the voting rights script and associated script version
			// that would ordinarily by used in a ticket purchase transaction to
			// give voting rights to the address.
			voteScriptVer, voteScript := stakeAddr.VotingRightsScript()
			fmt.Printf("  voting rights script version: %d\n", voteScriptVer)
			fmt.Printf("  voting rights script: %x\n", voteScript)

			// Obtain the rewards commitment script and associated script
			// version that would ordinarily by used in a ticket purchase
			// transaction to commit the original funds locked plus the reward
			// to the address.
			//
			// Ordinarily the reward amount and fee limits would need to be
			// calculated correctly, but they are hard coded here for the
			// purposes of this example.
			const rewardAmount = 1e8
			const voteFeeLimit = 0
			const revokeFeeLimit = 16777216
			rewardScriptVer, rewardScript := stakeAddr.RewardCommitmentScript(
				rewardAmount, voteFeeLimit, revokeFeeLimit)
			fmt.Printf("  reward script version: %d\n", rewardScriptVer)
			fmt.Printf("  reward script: %x\n", rewardScript)
		}
	}

	// Output:
	// addr0: TsbbBthkKvpMEd55BU3ndzGH9w8rX1HmQ3V
	//   payment script version: 0
	//   payment script: 76a91473f1d619dc632d15db63adfaf726b4633c5a057088ac
	//   hash160: 73f1d619dc632d15db63adfaf726b4633c5a0570
	//   voting rights script version: 0
	//   voting rights script: ba76a91473f1d619dc632d15db63adfaf726b4633c5a057088ac
	//   reward script version: 0
	//   reward script: 6a1e73f1d619dc632d15db63adfaf726b4633c5a057000e1f505000000000058
}
