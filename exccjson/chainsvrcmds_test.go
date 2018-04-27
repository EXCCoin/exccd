// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package exccjson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/EXCCoin/exccd/exccjson"
)

// TestChainSvrCmds tests all of the chain server commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestChainSvrCmds(t *testing.T) {
	t.Parallel()

	testID := int(1)
	tests := []struct {
		name         string
		newCmd       func() (interface{}, error)
		staticCmd    func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "addnode",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("addnode", "127.0.0.1", exccjson.ANRemove)
			},
			staticCmd: func() interface{} {
				return exccjson.NewAddNodeCmd("127.0.0.1", exccjson.ANRemove)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"addnode","params":["127.0.0.1","remove"],"id":1}`,
			unmarshalled: &exccjson.AddNodeCmd{Addr: "127.0.0.1", SubCmd: exccjson.ANRemove},
		},
		{
			name: "createrawtransaction",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("createrawtransaction", `[{"txid":"123","vout":1}]`,
					`{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				txInputs := []exccjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return exccjson.NewCreateRawTransactionCmd(txInputs, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createrawtransaction","params":[[{"txid":"123","vout":1,"tree":0}],{"456":0.0123}],"id":1}`,
			unmarshalled: &exccjson.CreateRawTransactionCmd{
				Inputs:  []exccjson.TransactionInput{{Txid: "123", Vout: 1}},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("createrawtransaction", `[{"txid":"123","vout":1,"tree":0}]`,
					`{"456":0.0123}`, int64(12312333333))
			},
			staticCmd: func() interface{} {
				txInputs := []exccjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return exccjson.NewCreateRawTransactionCmd(txInputs, amounts, exccjson.Int64(12312333333))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createrawtransaction","params":[[{"txid":"123","vout":1,"tree":0}],{"456":0.0123},12312333333],"id":1}`,
			unmarshalled: &exccjson.CreateRawTransactionCmd{
				Inputs:   []exccjson.TransactionInput{{Txid: "123", Vout: 1}},
				Amounts:  map[string]float64{"456": .0123},
				LockTime: exccjson.Int64(12312333333),
			},
		},
		{
			name: "decoderawtransaction",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("decoderawtransaction", "123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewDecodeRawTransactionCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decoderawtransaction","params":["123"],"id":1}`,
			unmarshalled: &exccjson.DecodeRawTransactionCmd{HexTx: "123"},
		},
		{
			name: "decodescript",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("decodescript", "00")
			},
			staticCmd: func() interface{} {
				return exccjson.NewDecodeScriptCmd("00")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodescript","params":["00"],"id":1}`,
			unmarshalled: &exccjson.DecodeScriptCmd{HexScript: "00"},
		},
		{
			name: "getaddednodeinfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getaddednodeinfo", true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetAddedNodeInfoCmd(true, nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getaddednodeinfo","params":[true],"id":1}`,
			unmarshalled: &exccjson.GetAddedNodeInfoCmd{DNS: true, Node: nil},
		},
		{
			name: "getaddednodeinfo optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getaddednodeinfo", true, "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetAddedNodeInfoCmd(true, exccjson.String("127.0.0.1"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaddednodeinfo","params":[true,"127.0.0.1"],"id":1}`,
			unmarshalled: &exccjson.GetAddedNodeInfoCmd{
				DNS:  true,
				Node: exccjson.String("127.0.0.1"),
			},
		},
		{
			name: "getbestblockhash",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getbestblockhash")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBestBlockHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getbestblockhash","params":[],"id":1}`,
			unmarshalled: &exccjson.GetBestBlockHashCmd{},
		},
		{
			name: "getblock",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblock", "123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockCmd("123", nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123"],"id":1}`,
			unmarshalled: &exccjson.GetBlockCmd{
				Hash:      "123",
				Verbose:   exccjson.Bool(true),
				VerboseTx: exccjson.Bool(false),
			},
		},
		{
			name: "getblock required optional1",
			newCmd: func() (interface{}, error) {
				// Intentionally use a source param that is
				// more pointers than the destination to
				// exercise that path.
				verbosePtr := exccjson.Bool(true)
				return exccjson.NewCmd("getblock", "123", &verbosePtr)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockCmd("123", exccjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123",true],"id":1}`,
			unmarshalled: &exccjson.GetBlockCmd{
				Hash:      "123",
				Verbose:   exccjson.Bool(true),
				VerboseTx: exccjson.Bool(false),
			},
		},
		{
			name: "getblock required optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblock", "123", true, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockCmd("123", exccjson.Bool(true), exccjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123",true,true],"id":1}`,
			unmarshalled: &exccjson.GetBlockCmd{
				Hash:      "123",
				Verbose:   exccjson.Bool(true),
				VerboseTx: exccjson.Bool(true),
			},
		},
		{
			name: "getblockchaininfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblockchaininfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockChainInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockchaininfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetBlockChainInfoCmd{},
		},
		{
			name: "getblockcount",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblockcount")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockcount","params":[],"id":1}`,
			unmarshalled: &exccjson.GetBlockCountCmd{},
		},
		{
			name: "getblockhash",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblockhash", 123)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockHashCmd(123)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockhash","params":[123],"id":1}`,
			unmarshalled: &exccjson.GetBlockHashCmd{Index: 123},
		},
		{
			name: "getblockheader",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblockheader", "123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblockheader","params":["123"],"id":1}`,
			unmarshalled: &exccjson.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: exccjson.Bool(true),
			},
		},
		{
			name: "getblocksubsidy",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblocksubsidy", 123, 256)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockSubsidyCmd(123, 256)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocksubsidy","params":[123,256],"id":1}`,
			unmarshalled: &exccjson.GetBlockSubsidyCmd{
				Height: 123,
				Voters: 256,
			},
		},
		{
			name: "getblocktemplate",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblocktemplate")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblocktemplate","params":[],"id":1}`,
			unmarshalled: &exccjson.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getblocktemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblocktemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"]}`)
			},
			staticCmd: func() interface{} {
				template := exccjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				}
				return exccjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocktemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"]}],"id":1}`,
			unmarshalled: &exccjson.GetBlockTemplateCmd{
				Request: &exccjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				},
			},
		},
		{
			name: "getblocktemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblocktemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":500,"sizelimit":100000000,"maxversion":2}`)
			},
			staticCmd: func() interface{} {
				template := exccjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   500,
					SizeLimit:    100000000,
					MaxVersion:   2,
				}
				return exccjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocktemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":500,"sizelimit":100000000,"maxversion":2}],"id":1}`,
			unmarshalled: &exccjson.GetBlockTemplateCmd{
				Request: &exccjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   int64(500),
					SizeLimit:    int64(100000000),
					MaxVersion:   2,
				},
			},
		},
		{
			name: "getblocktemplate optional - template request with tweaks 2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getblocktemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":true,"sizelimit":100000000,"maxversion":2}`)
			},
			staticCmd: func() interface{} {
				template := exccjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   true,
					SizeLimit:    100000000,
					MaxVersion:   2,
				}
				return exccjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocktemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":true,"sizelimit":100000000,"maxversion":2}],"id":1}`,
			unmarshalled: &exccjson.GetBlockTemplateCmd{
				Request: &exccjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   true,
					SizeLimit:    int64(100000000),
					MaxVersion:   2,
				},
			},
		},
		{
			name: "getcfilter",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getcfilter", "123", "extended")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetCFilterCmd("123", "extended")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getcfilter","params":["123","extended"],"id":1}`,
			unmarshalled: &exccjson.GetCFilterCmd{
				Hash:       "123",
				FilterType: "extended",
			},
		},
		{
			name: "getcfilterheader",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getcfilterheader", "123", "extended")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetCFilterHeaderCmd("123", "extended")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getcfilterheader","params":["123","extended"],"id":1}`,
			unmarshalled: &exccjson.GetCFilterHeaderCmd{
				Hash:       "123",
				FilterType: "extended",
			},
		},
		{
			name: "getchaintips",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getchaintips")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetChainTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getchaintips","params":[],"id":1}`,
			unmarshalled: &exccjson.GetChainTipsCmd{},
		},
		{
			name: "getconnectioncount",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getconnectioncount")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getconnectioncount","params":[],"id":1}`,
			unmarshalled: &exccjson.GetConnectionCountCmd{},
		},
		{
			name: "getdifficulty",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getdifficulty")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getdifficulty","params":[],"id":1}`,
			unmarshalled: &exccjson.GetDifficultyCmd{},
		},
		{
			name: "getgenerate",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getgenerate")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetGenerateCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getgenerate","params":[],"id":1}`,
			unmarshalled: &exccjson.GetGenerateCmd{},
		},
		{
			name: "gethashespersec",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("gethashespersec")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetHashesPerSecCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"gethashespersec","params":[],"id":1}`,
			unmarshalled: &exccjson.GetHashesPerSecCmd{},
		},
		{
			name: "getinfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getinfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetInfoCmd{},
		},
		{
			name: "getmempoolinfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getmempoolinfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getmempoolinfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetMempoolInfoCmd{},
		},
		{
			name: "getmininginfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getmininginfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetMiningInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getmininginfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetMiningInfoCmd{},
		},
		{
			name: "getnetworkinfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnetworkinfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getnetworkinfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetNetworkInfoCmd{},
		},
		{
			name: "getnettotals",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnettotals")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getnettotals","params":[],"id":1}`,
			unmarshalled: &exccjson.GetNetTotalsCmd{},
		},
		{
			name: "getnetworkhashps",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnetworkhashps")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNetworkHashPSCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnetworkhashps","params":[],"id":1}`,
			unmarshalled: &exccjson.GetNetworkHashPSCmd{
				Blocks: exccjson.Int(120),
				Height: exccjson.Int(-1),
			},
		},
		{
			name: "getnetworkhashps optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnetworkhashps", 200)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNetworkHashPSCmd(exccjson.Int(200), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnetworkhashps","params":[200],"id":1}`,
			unmarshalled: &exccjson.GetNetworkHashPSCmd{
				Blocks: exccjson.Int(200),
				Height: exccjson.Int(-1),
			},
		},
		{
			name: "getnetworkhashps optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnetworkhashps", 200, 123)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNetworkHashPSCmd(exccjson.Int(200), exccjson.Int(123))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnetworkhashps","params":[200,123],"id":1}`,
			unmarshalled: &exccjson.GetNetworkHashPSCmd{
				Blocks: exccjson.Int(200),
				Height: exccjson.Int(123),
			},
		},
		{
			name: "getpeerinfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getpeerinfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getpeerinfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetPeerInfoCmd{},
		},
		{
			name: "getrawmempool",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawmempool")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawMempoolCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawmempool","params":[],"id":1}`,
			unmarshalled: &exccjson.GetRawMempoolCmd{
				Verbose: exccjson.Bool(false),
			},
		},
		{
			name: "getrawmempool optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawmempool", false)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawMempoolCmd(exccjson.Bool(false), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawmempool","params":[false],"id":1}`,
			unmarshalled: &exccjson.GetRawMempoolCmd{
				Verbose: exccjson.Bool(false),
			},
		},
		{
			name: "getrawmempool optional 2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawmempool", false, "all")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawMempoolCmd(exccjson.Bool(false), exccjson.String("all"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawmempool","params":[false,"all"],"id":1}`,
			unmarshalled: &exccjson.GetRawMempoolCmd{
				Verbose: exccjson.Bool(false),
				TxType:  exccjson.String("all"),
			},
		},
		{
			name: "getrawtransaction",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawtransaction", "123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawtransaction","params":["123"],"id":1}`,
			unmarshalled: &exccjson.GetRawTransactionCmd{
				Txid:    "123",
				Verbose: exccjson.Int(0),
			},
		},
		{
			name: "getrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawtransaction", "123", 1)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawTransactionCmd("123", exccjson.Int(1))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawtransaction","params":["123",1],"id":1}`,
			unmarshalled: &exccjson.GetRawTransactionCmd{
				Txid:    "123",
				Verbose: exccjson.Int(1),
			},
		},
		{
			name: "gettxout",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("gettxout", "123", 1)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxout","params":["123",1],"id":1}`,
			unmarshalled: &exccjson.GetTxOutCmd{
				Txid:           "123",
				Vout:           1,
				IncludeMempool: exccjson.Bool(true),
			},
		},
		{
			name: "gettxout optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("gettxout", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetTxOutCmd("123", 1, exccjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxout","params":["123",1,true],"id":1}`,
			unmarshalled: &exccjson.GetTxOutCmd{
				Txid:           "123",
				Vout:           1,
				IncludeMempool: exccjson.Bool(true),
			},
		},
		{
			name: "gettxoutsetinfo",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("gettxoutsetinfo")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"gettxoutsetinfo","params":[],"id":1}`,
			unmarshalled: &exccjson.GetTxOutSetInfoCmd{},
		},
		{
			name: "getwork",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getwork")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetWorkCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getwork","params":[],"id":1}`,
			unmarshalled: &exccjson.GetWorkCmd{
				Data: nil,
			},
		},
		{
			name: "getwork optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getwork", "00112233")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetWorkCmd(exccjson.String("00112233"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getwork","params":["00112233"],"id":1}`,
			unmarshalled: &exccjson.GetWorkCmd{
				Data: exccjson.String("00112233"),
			},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("help")
			},
			staticCmd: func() interface{} {
				return exccjson.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &exccjson.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("help", "getblock")
			},
			staticCmd: func() interface{} {
				return exccjson.NewHelpCmd(exccjson.String("getblock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getblock"],"id":1}`,
			unmarshalled: &exccjson.HelpCmd{
				Command: exccjson.String("getblock"),
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("ping")
			},
			staticCmd: func() interface{} {
				return exccjson.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &exccjson.PingCmd{},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address", nil, nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address"],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(1),
				Skip:        exccjson.Int(0),
				Count:       exccjson.Int(100),
				VinExtra:    exccjson.Int(0),
				Reverse:     exccjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address", 0)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address",
					exccjson.Int(0), nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(0),
				Skip:        exccjson.Int(0),
				Count:       exccjson.Int(100),
				VinExtra:    exccjson.Int(0),
				Reverse:     exccjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address", 0, 5)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address",
					exccjson.Int(0), exccjson.Int(5), nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(0),
				Skip:        exccjson.Int(5),
				Count:       exccjson.Int(100),
				VinExtra:    exccjson.Int(0),
				Reverse:     exccjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address",
					exccjson.Int(0), exccjson.Int(5), exccjson.Int(10), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(0),
				Skip:        exccjson.Int(5),
				Count:       exccjson.Int(10),
				VinExtra:    exccjson.Int(0),
				Reverse:     exccjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, 1)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address",
					exccjson.Int(0), exccjson.Int(5), exccjson.Int(10), exccjson.Int(1), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,1],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(0),
				Skip:        exccjson.Int(5),
				Count:       exccjson.Int(10),
				VinExtra:    exccjson.Int(1),
				Reverse:     exccjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, 1, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address",
					exccjson.Int(0), exccjson.Int(5), exccjson.Int(10),
					exccjson.Int(1), exccjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,1,true],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(0),
				Skip:        exccjson.Int(5),
				Count:       exccjson.Int(10),
				VinExtra:    exccjson.Int(1),
				Reverse:     exccjson.Bool(true),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, 1, true, []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return exccjson.NewSearchRawTransactionsCmd("1Address",
					exccjson.Int(0), exccjson.Int(5), exccjson.Int(10),
					exccjson.Int(1), exccjson.Bool(true), &[]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,1,true,["1Address"]],"id":1}`,
			unmarshalled: &exccjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     exccjson.Int(0),
				Skip:        exccjson.Int(5),
				Count:       exccjson.Int(10),
				VinExtra:    exccjson.Int(1),
				Reverse:     exccjson.Bool(true),
				FilterAddrs: &[]string{"1Address"},
			},
		},
		{
			name: "sendrawtransaction",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendrawtransaction", "1122")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendrawtransaction","params":["1122"],"id":1}`,
			unmarshalled: &exccjson.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: exccjson.Bool(false),
			},
		},
		{
			name: "sendrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendrawtransaction", "1122", false)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendRawTransactionCmd("1122", exccjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendrawtransaction","params":["1122",false],"id":1}`,
			unmarshalled: &exccjson.SendRawTransactionCmd{
				HexTx:         "1122",
				AllowHighFees: exccjson.Bool(false),
			},
		},
		{
			name: "setgenerate",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("setgenerate", true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSetGenerateCmd(true, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"setgenerate","params":[true],"id":1}`,
			unmarshalled: &exccjson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: exccjson.Int(-1),
			},
		},
		{
			name: "setgenerate optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("setgenerate", true, 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSetGenerateCmd(true, exccjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"setgenerate","params":[true,6],"id":1}`,
			unmarshalled: &exccjson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: exccjson.Int(6),
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("stop")
			},
			staticCmd: func() interface{} {
				return exccjson.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &exccjson.StopCmd{},
		},
		{
			name: "submitblock",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("submitblock", "112233")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitblock","params":["112233"],"id":1}`,
			unmarshalled: &exccjson.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitblock optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("submitblock", "112233", `{"workid":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := exccjson.SubmitBlockOptions{
					WorkID: "12345",
				}
				return exccjson.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitblock","params":["112233",{"workid":"12345"}],"id":1}`,
			unmarshalled: &exccjson.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &exccjson.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "validateaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("validateaddress", "1Address")
			},
			staticCmd: func() interface{} {
				return exccjson.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateaddress","params":["1Address"],"id":1}`,
			unmarshalled: &exccjson.ValidateAddressCmd{
				Address: "1Address",
			},
		},
		{
			name: "verifychain",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("verifychain")
			},
			staticCmd: func() interface{} {
				return exccjson.NewVerifyChainCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifychain","params":[],"id":1}`,
			unmarshalled: &exccjson.VerifyChainCmd{
				CheckLevel: exccjson.Int64(3),
				CheckDepth: exccjson.Int64(288),
			},
		},
		{
			name: "verifychain optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("verifychain", 2)
			},
			staticCmd: func() interface{} {
				return exccjson.NewVerifyChainCmd(exccjson.Int64(2), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifychain","params":[2],"id":1}`,
			unmarshalled: &exccjson.VerifyChainCmd{
				CheckLevel: exccjson.Int64(2),
				CheckDepth: exccjson.Int64(288),
			},
		},
		{
			name: "verifychain optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("verifychain", 2, 500)
			},
			staticCmd: func() interface{} {
				return exccjson.NewVerifyChainCmd(exccjson.Int64(2), exccjson.Int64(500))
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifychain","params":[2,500],"id":1}`,
			unmarshalled: &exccjson.VerifyChainCmd{
				CheckLevel: exccjson.Int64(2),
				CheckDepth: exccjson.Int64(500),
			},
		},
		{
			name: "verifymessage",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("verifymessage", "1Address", "301234", "test")
			},
			staticCmd: func() interface{} {
				return exccjson.NewVerifyMessageCmd("1Address", "301234", "test")
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifymessage","params":["1Address","301234","test"],"id":1}`,
			unmarshalled: &exccjson.VerifyMessageCmd{
				Address:   "1Address",
				Signature: "301234",
				Message:   "test",
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := exccjson.MarshalCmd("1.0", testID, test.staticCmd())
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			t.Errorf("\n%s\n%s", marshalled, test.marshalled)
			continue
		}

		// Ensure the command is created without error via the generic
		// new command creation function.
		cmd, err := test.newCmd()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCmd error: %v ",
				i, test.name, err)
		}

		// Marshal the command as created by the generic new command
		// creation function.
		marshalled, err = exccjson.MarshalCmd("1.0", testID, cmd)
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request exccjson.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = exccjson.UnmarshalCmd(&request)
		if err != nil {
			t.Errorf("UnmarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !reflect.DeepEqual(cmd, test.unmarshalled) {
			t.Errorf("Test #%d (%s) unexpected unmarshalled command "+
				"- got %s, want %s", i, test.name,
				fmt.Sprintf("(%T) %+[1]v", cmd),
				fmt.Sprintf("(%T) %+[1]v\n", test.unmarshalled))
			continue
		}
	}
}

// TestChainSvrCmdErrors ensures any errors that occur in the command during
// custom mashal and unmarshal are as expected.
func TestChainSvrCmdErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     interface{}
		marshalled string
		err        error
	}{
		{
			name:       "template request with invalid type",
			result:     &exccjson.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &exccjson.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        exccjson.Error{Code: exccjson.ErrInvalidType},
		},
		{
			name:       "invalid template request sizelimit field",
			result:     &exccjson.TemplateRequest{},
			marshalled: `{"sizelimit":"invalid"}`,
			err:        exccjson.Error{Code: exccjson.ErrInvalidType},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		err := json.Unmarshal([]byte(test.marshalled), &test.result)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Test #%d (%s) wrong error type - got `%T` (%v), got `%T`",
				i, test.name, err, err, test.err)
			continue
		}

		if terr, ok := test.err.(exccjson.Error); ok {
			gotErrorCode := err.(exccjson.Error).Code
			if gotErrorCode != terr.Code {
				t.Errorf("Test #%d (%s) mismatched error code "+
					"- got %v (%v), want %v", i, test.name,
					gotErrorCode, terr, terr.Code)
				continue
			}
		}
	}
}
