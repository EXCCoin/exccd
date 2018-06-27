// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"testing"

	"github.com/EXCCoin/exccd/chaincfg"
)

func TestBlockSubsidy(t *testing.T) {
	// TODO: Test is affected by block time and inflation. Fix it after applying inflation changes

	mainnet := &chaincfg.MainNetParams
	subsidyCache := NewSubsidyCache(0, mainnet)

	totalSubsidy := mainnet.BlockOneSubsidy()
	for i := int64(0); ; i++ {
		// Genesis block or first block.
		if i == 0 || i == 1 {
			continue
		}

		if i%mainnet.SubsidyReductionInterval == 0 {
			numBlocks := mainnet.SubsidyReductionInterval
			// First reduction internal, which is reduction interval - 2
			// to skip the genesis block and block one.
			if i == mainnet.SubsidyReductionInterval {
				numBlocks -= 2
			}
			height := i - numBlocks

			work := CalcBlockWorkSubsidy(subsidyCache, height,
				mainnet.TicketsPerBlock, mainnet)
			stake := CalcStakeVoteSubsidy(subsidyCache, height,
				mainnet) * int64(mainnet.TicketsPerBlock)
			if (work + stake) == 0 {
				break
			}
			totalSubsidy += (work + stake) * numBlocks
		}
	}

	expectedSubsidy := int64(2044657240590848) + mainnet.BlockOneSubsidy()
	if totalSubsidy != expectedSubsidy {
		t.Errorf("Bad total subsidy; want %v, got %v", expectedSubsidy, totalSubsidy)
	}
}
