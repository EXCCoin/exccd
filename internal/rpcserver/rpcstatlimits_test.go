// Copyright (c) 2026 The ExchangeCoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcserver

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/dcrjson/v4"
	"github.com/EXCCoin/exccd/dcrutil/v4"
	"github.com/EXCCoin/exccd/rpc/jsonrpc/types/v3"
)

func TestStatsRequestLimits(t *testing.T) {
	cfg := defaultMockConfig(chaincfg.MainNetParams())
	// Invalid parameters must be rejected before looking up any fees.
	cfg.TxMempooler = nil
	s := &Server{cfg: *cfg}
	blocks, windows := uint32(maxFeeInfoBlocks+1), uint32(maxTicketFeeWindows+1)
	start := uint32(1)
	feeEnd, vwapEnd := uint32(maxFeeRangeBlocks+1), uint32(maxVWAPHeaders+1)
	count := maxSearchTxCount + 1
	for _, test := range []struct {
		name    string
		handler commandHandler
		cmd     interface{}
	}{
		{"tx blocks", handleTxFeeInfo, &types.TxFeeInfoCmd{Blocks: &blocks}},
		{"ticket blocks", handleTicketFeeInfo, &types.TicketFeeInfoCmd{Blocks: &blocks}},
		{"ticket windows", handleTicketFeeInfo, &types.TicketFeeInfoCmd{Windows: &windows}},
		{"tx range", handleTxFeeInfo, &types.TxFeeInfoCmd{RangeStart: &start, RangeEnd: &feeEnd}},
		{"ticket VWAP range", handleTicketVWAP, &types.TicketVWAPCmd{Start: &start, End: &vwapEnd}},
		{"indexed address page", handleSearchRawTransactions, &types.SearchRawTransactionsCmd{Count: &count}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.handler(context.Background(), s, test.cmd)
			var rpcErr *dcrjson.RPCError
			if !errors.As(err, &rpcErr) || rpcErr.Code != dcrjson.ErrRPCInvalidParameter {
				t.Fatalf("got %v, want invalid parameter", err)
			}
		})
	}
}

func TestStatsRangeBoundaries(t *testing.T) {
	const tip = math.MaxUint32
	for _, limit := range []int64{maxVWAPHeaders, maxFeeRangeBlocks} {
		first := uint32(tip - limit + 1)
		start, end, err := statsHeightRange(&first, nil, tip, 144, limit)
		if err != nil || end-start+1 != limit || end+1 != int64(tip)+1 {
			t.Fatalf("inclusive boundary %d: [%d,%d] %v", limit, start, end, err)
		}
		first--
		if _, _, err := statsHeightRange(&first, nil, tip, 144, limit); err == nil {
			t.Fatalf("accepted more than %d heights", limit)
		}
	}
	for _, params := range []*chaincfg.Params{chaincfg.MainNetParams(), chaincfg.TestNet3Params(), chaincfg.SimNetParams(), chaincfg.RegNetParams()} {
		for _, r := range []struct{ lookback, limit int64 }{{params.WorkDiffWindowSize, maxFeeRangeBlocks}, {params.WorkDiffWindows * params.WorkDiffWindowSize, maxVWAPHeaders}} {
			start, end, err := statsHeightRange(nil, nil, 10000, r.lookback, r.limit)
			if err != nil || end-start != r.lookback {
				t.Fatalf("%s default range [%d,%d], %v", params.Name, start, end, err)
			}
		}
	}
	blocks := uint32(maxFeeInfoBlocks)
	if got, err := feeBlockCount(&blocks, maxFeeInfoBlocks-1); err != nil || got != blocks {
		t.Fatalf("last available block: %d %v", got, err)
	}
	if _, err := feeBlockCount(&blocks, maxFeeInfoBlocks-2); err == nil {
		t.Fatal("accepted a request below genesis")
	}
}

func TestVWAPWideAccumulator(t *testing.T) {
	cfg := defaultMockConfig(chaincfg.MainNetParams())
	chain := defaultMockRPCChain()
	cfg.Chain = chain
	chain.headerByHeight.FreshStake = 20
	chain.headerByHeight.SBits = dcrutil.MaxAmount
	start, end := uint32(1), uint32(512)
	got, err := handleTicketVWAP(context.Background(), &Server{cfg: *cfg}, &types.TicketVWAPCmd{Start: &start, End: &end})
	if err != nil || got != dcrutil.Amount(dcrutil.MaxAmount).ToCoin() {
		t.Fatalf("ticket VWAP %v, %v", got, err)
	}
}

func TestLimitedMiningAndHistoryMethods(t *testing.T) {
	for method, want := range map[string]bool{"regentemplate": false, "submitblock": false, "searchrawtransactions": true} {
		_, got := rpcLimited[method]
		if got != want || rpcHandlers[types.Method(method)] == nil {
			t.Errorf("%s limited=%t, want %t and admin handler", method, got, want)
		}
	}
}
