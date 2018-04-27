// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
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

// TestWalletSvrCmds tests all of the wallet server commands marshal and
// unmarshal into valid results include handling of optional fields being
// omitted in the marshalled command, while optional fields with defaults have
// the default assigned on unmarshalled commands.
func TestWalletSvrCmds(t *testing.T) {
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
			name: "addmultisigaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("addmultisigaddress", 2, []string{"031234", "035678"})
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return exccjson.NewAddMultisigAddressCmd(2, keys, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"addmultisigaddress","params":[2,["031234","035678"]],"id":1}`,
			unmarshalled: &exccjson.AddMultisigAddressCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
				Account:   nil,
			},
		},
		{
			name: "addmultisigaddress optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("addmultisigaddress", 2, []string{"031234", "035678"}, "test")
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return exccjson.NewAddMultisigAddressCmd(2, keys, exccjson.String("test"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"addmultisigaddress","params":[2,["031234","035678"],"test"],"id":1}`,
			unmarshalled: &exccjson.AddMultisigAddressCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
				Account:   exccjson.String("test"),
			},
		},
		{
			name: "createmultisig",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("createmultisig", 2, []string{"031234", "035678"})
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return exccjson.NewCreateMultisigCmd(2, keys)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createmultisig","params":[2,["031234","035678"]],"id":1}`,
			unmarshalled: &exccjson.CreateMultisigCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
			},
		},
		{
			name: "dumpprivkey",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("dumpprivkey", "1Address")
			},
			staticCmd: func() interface{} {
				return exccjson.NewDumpPrivKeyCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"dumpprivkey","params":["1Address"],"id":1}`,
			unmarshalled: &exccjson.DumpPrivKeyCmd{
				Address: "1Address",
			},
		},
		{
			name: "estimatefee",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("estimatefee", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewEstimateFeeCmd(6)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatefee","params":[6],"id":1}`,
			unmarshalled: &exccjson.EstimateFeeCmd{
				NumBlocks: 6,
			},
		},
		{
			name: "estimatepriority",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("estimatepriority", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewEstimatePriorityCmd(6)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatepriority","params":[6],"id":1}`,
			unmarshalled: &exccjson.EstimatePriorityCmd{
				NumBlocks: 6,
			},
		},
		{
			name: "getaccount",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getaccount", "1Address")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetAccountCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaccount","params":["1Address"],"id":1}`,
			unmarshalled: &exccjson.GetAccountCmd{
				Address: "1Address",
			},
		},
		{
			name: "getaccountaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getaccountaddress", "acct")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetAccountAddressCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaccountaddress","params":["acct"],"id":1}`,
			unmarshalled: &exccjson.GetAccountAddressCmd{
				Account: "acct",
			},
		},
		{
			name: "getaddressesbyaccount",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getaddressesbyaccount", "acct")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetAddressesByAccountCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaddressesbyaccount","params":["acct"],"id":1}`,
			unmarshalled: &exccjson.GetAddressesByAccountCmd{
				Account: "acct",
			},
		},
		{
			name: "getbalance",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getbalance")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBalanceCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getbalance","params":[],"id":1}`,
			unmarshalled: &exccjson.GetBalanceCmd{
				Account: nil,
				MinConf: exccjson.Int(1),
			},
		},
		{
			name: "getbalance optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getbalance", "acct")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBalanceCmd(exccjson.String("acct"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getbalance","params":["acct"],"id":1}`,
			unmarshalled: &exccjson.GetBalanceCmd{
				Account: exccjson.String("acct"),
				MinConf: exccjson.Int(1),
			},
		},
		{
			name: "getbalance optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getbalance", "acct", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetBalanceCmd(exccjson.String("acct"), exccjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getbalance","params":["acct",6],"id":1}`,
			unmarshalled: &exccjson.GetBalanceCmd{
				Account: exccjson.String("acct"),
				MinConf: exccjson.Int(6),
			},
		},
		{
			name: "getnewaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnewaddress")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNewAddressCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnewaddress","params":[],"id":1}`,
			unmarshalled: &exccjson.GetNewAddressCmd{
				Account:   nil,
				GapPolicy: nil,
			},
		},
		{
			name: "getnewaddress optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getnewaddress", "acct", "ignore")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetNewAddressCmd(exccjson.String("acct"), exccjson.String("ignore"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnewaddress","params":["acct","ignore"],"id":1}`,
			unmarshalled: &exccjson.GetNewAddressCmd{
				Account:   exccjson.String("acct"),
				GapPolicy: exccjson.String("ignore"),
			},
		},
		{
			name: "getrawchangeaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawchangeaddress")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawChangeAddressCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawchangeaddress","params":[],"id":1}`,
			unmarshalled: &exccjson.GetRawChangeAddressCmd{
				Account: nil,
			},
		},
		{
			name: "getrawchangeaddress optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getrawchangeaddress", "acct")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetRawChangeAddressCmd(exccjson.String("acct"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawchangeaddress","params":["acct"],"id":1}`,
			unmarshalled: &exccjson.GetRawChangeAddressCmd{
				Account: exccjson.String("acct"),
			},
		},
		{
			name: "getreceivedbyaccount",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getreceivedbyaccount", "acct")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetReceivedByAccountCmd("acct", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaccount","params":["acct"],"id":1}`,
			unmarshalled: &exccjson.GetReceivedByAccountCmd{
				Account: "acct",
				MinConf: exccjson.Int(1),
			},
		},
		{
			name: "getreceivedbyaccount optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getreceivedbyaccount", "acct", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetReceivedByAccountCmd("acct", exccjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaccount","params":["acct",6],"id":1}`,
			unmarshalled: &exccjson.GetReceivedByAccountCmd{
				Account: "acct",
				MinConf: exccjson.Int(6),
			},
		},
		{
			name: "getreceivedbyaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getreceivedbyaddress", "1Address")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetReceivedByAddressCmd("1Address", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaddress","params":["1Address"],"id":1}`,
			unmarshalled: &exccjson.GetReceivedByAddressCmd{
				Address: "1Address",
				MinConf: exccjson.Int(1),
			},
		},
		{
			name: "getreceivedbyaddress optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("getreceivedbyaddress", "1Address", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetReceivedByAddressCmd("1Address", exccjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaddress","params":["1Address",6],"id":1}`,
			unmarshalled: &exccjson.GetReceivedByAddressCmd{
				Address: "1Address",
				MinConf: exccjson.Int(6),
			},
		},
		{
			name: "gettransaction",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("gettransaction", "123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettransaction","params":["123"],"id":1}`,
			unmarshalled: &exccjson.GetTransactionCmd{
				Txid:             "123",
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "gettransaction optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("gettransaction", "123", true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewGetTransactionCmd("123", exccjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettransaction","params":["123",true],"id":1}`,
			unmarshalled: &exccjson.GetTransactionCmd{
				Txid:             "123",
				IncludeWatchOnly: exccjson.Bool(true),
			},
		},
		{
			name: "importprivkey",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("importprivkey", "abc")
			},
			staticCmd: func() interface{} {
				return exccjson.NewImportPrivKeyCmd("abc", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc"],"id":1}`,
			unmarshalled: &exccjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   nil,
				Rescan:  exccjson.Bool(true),
			},
		},
		{
			name: "importprivkey optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("importprivkey", "abc", "label")
			},
			staticCmd: func() interface{} {
				return exccjson.NewImportPrivKeyCmd("abc", exccjson.String("label"), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc","label"],"id":1}`,
			unmarshalled: &exccjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   exccjson.String("label"),
				Rescan:  exccjson.Bool(true),
			},
		},
		{
			name: "importprivkey optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("importprivkey", "abc", "label", false)
			},
			staticCmd: func() interface{} {
				return exccjson.NewImportPrivKeyCmd("abc", exccjson.String("label"), exccjson.Bool(false), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc","label",false],"id":1}`,
			unmarshalled: &exccjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   exccjson.String("label"),
				Rescan:  exccjson.Bool(false),
			},
		},
		{
			name: "importprivkey optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("importprivkey", "abc", "label", false, 12345)
			},
			staticCmd: func() interface{} {
				return exccjson.NewImportPrivKeyCmd("abc", exccjson.String("label"), exccjson.Bool(false), exccjson.Int(12345))
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc","label",false,12345],"id":1}`,
			unmarshalled: &exccjson.ImportPrivKeyCmd{
				PrivKey:  "abc",
				Label:    exccjson.String("label"),
				Rescan:   exccjson.Bool(false),
				ScanFrom: exccjson.Int(12345),
			},
		},
		{
			name: "keypoolrefill",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("keypoolrefill")
			},
			staticCmd: func() interface{} {
				return exccjson.NewKeyPoolRefillCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"keypoolrefill","params":[],"id":1}`,
			unmarshalled: &exccjson.KeyPoolRefillCmd{
				NewSize: exccjson.Uint(100),
			},
		},
		{
			name: "keypoolrefill optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("keypoolrefill", 200)
			},
			staticCmd: func() interface{} {
				return exccjson.NewKeyPoolRefillCmd(exccjson.Uint(200))
			},
			marshalled: `{"jsonrpc":"1.0","method":"keypoolrefill","params":[200],"id":1}`,
			unmarshalled: &exccjson.KeyPoolRefillCmd{
				NewSize: exccjson.Uint(200),
			},
		},
		{
			name: "listaccounts",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listaccounts")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListAccountsCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listaccounts","params":[],"id":1}`,
			unmarshalled: &exccjson.ListAccountsCmd{
				MinConf: exccjson.Int(1),
			},
		},
		{
			name: "listaccounts optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listaccounts", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListAccountsCmd(exccjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listaccounts","params":[6],"id":1}`,
			unmarshalled: &exccjson.ListAccountsCmd{
				MinConf: exccjson.Int(6),
			},
		},
		{
			name: "listlockunspent",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listlockunspent")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListLockUnspentCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"listlockunspent","params":[],"id":1}`,
			unmarshalled: &exccjson.ListLockUnspentCmd{},
		},
		{
			name: "listreceivedbyaccount",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaccount")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAccountCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAccountCmd{
				MinConf:          exccjson.Int(1),
				IncludeEmpty:     exccjson.Bool(false),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaccount optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaccount", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAccountCmd(exccjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[6],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAccountCmd{
				MinConf:          exccjson.Int(6),
				IncludeEmpty:     exccjson.Bool(false),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaccount optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaccount", 6, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAccountCmd(exccjson.Int(6), exccjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[6,true],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAccountCmd{
				MinConf:          exccjson.Int(6),
				IncludeEmpty:     exccjson.Bool(true),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaccount optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaccount", 6, true, false)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAccountCmd(exccjson.Int(6), exccjson.Bool(true), exccjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[6,true,false],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAccountCmd{
				MinConf:          exccjson.Int(6),
				IncludeEmpty:     exccjson.Bool(true),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaddress")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAddressCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAddressCmd{
				MinConf:          exccjson.Int(1),
				IncludeEmpty:     exccjson.Bool(false),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaddress", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAddressCmd(exccjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[6],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAddressCmd{
				MinConf:          exccjson.Int(6),
				IncludeEmpty:     exccjson.Bool(false),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaddress", 6, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAddressCmd(exccjson.Int(6), exccjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[6,true],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAddressCmd{
				MinConf:          exccjson.Int(6),
				IncludeEmpty:     exccjson.Bool(true),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listreceivedbyaddress", 6, true, false)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListReceivedByAddressCmd(exccjson.Int(6), exccjson.Bool(true), exccjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[6,true,false],"id":1}`,
			unmarshalled: &exccjson.ListReceivedByAddressCmd{
				MinConf:          exccjson.Int(6),
				IncludeEmpty:     exccjson.Bool(true),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listsinceblock",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listsinceblock")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListSinceBlockCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":[],"id":1}`,
			unmarshalled: &exccjson.ListSinceBlockCmd{
				BlockHash:           nil,
				TargetConfirmations: exccjson.Int(1),
				IncludeWatchOnly:    exccjson.Bool(false),
			},
		},
		{
			name: "listsinceblock optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listsinceblock", "123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListSinceBlockCmd(exccjson.String("123"), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":["123"],"id":1}`,
			unmarshalled: &exccjson.ListSinceBlockCmd{
				BlockHash:           exccjson.String("123"),
				TargetConfirmations: exccjson.Int(1),
				IncludeWatchOnly:    exccjson.Bool(false),
			},
		},
		{
			name: "listsinceblock optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listsinceblock", "123", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListSinceBlockCmd(exccjson.String("123"), exccjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":["123",6],"id":1}`,
			unmarshalled: &exccjson.ListSinceBlockCmd{
				BlockHash:           exccjson.String("123"),
				TargetConfirmations: exccjson.Int(6),
				IncludeWatchOnly:    exccjson.Bool(false),
			},
		},
		{
			name: "listsinceblock optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listsinceblock", "123", 6, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListSinceBlockCmd(exccjson.String("123"), exccjson.Int(6), exccjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":["123",6,true],"id":1}`,
			unmarshalled: &exccjson.ListSinceBlockCmd{
				BlockHash:           exccjson.String("123"),
				TargetConfirmations: exccjson.Int(6),
				IncludeWatchOnly:    exccjson.Bool(true),
			},
		},
		{
			name: "listtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listtransactions")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListTransactionsCmd(nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":[],"id":1}`,
			unmarshalled: &exccjson.ListTransactionsCmd{
				Account:          nil,
				Count:            exccjson.Int(10),
				From:             exccjson.Int(0),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listtransactions", "acct")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListTransactionsCmd(exccjson.String("acct"), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct"],"id":1}`,
			unmarshalled: &exccjson.ListTransactionsCmd{
				Account:          exccjson.String("acct"),
				Count:            exccjson.Int(10),
				From:             exccjson.Int(0),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listtransactions", "acct", 20)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListTransactionsCmd(exccjson.String("acct"), exccjson.Int(20), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct",20],"id":1}`,
			unmarshalled: &exccjson.ListTransactionsCmd{
				Account:          exccjson.String("acct"),
				Count:            exccjson.Int(20),
				From:             exccjson.Int(0),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listtransactions", "acct", 20, 1)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListTransactionsCmd(exccjson.String("acct"), exccjson.Int(20),
					exccjson.Int(1), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct",20,1],"id":1}`,
			unmarshalled: &exccjson.ListTransactionsCmd{
				Account:          exccjson.String("acct"),
				Count:            exccjson.Int(20),
				From:             exccjson.Int(1),
				IncludeWatchOnly: exccjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional4",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listtransactions", "acct", 20, 1, true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListTransactionsCmd(exccjson.String("acct"), exccjson.Int(20),
					exccjson.Int(1), exccjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct",20,1,true],"id":1}`,
			unmarshalled: &exccjson.ListTransactionsCmd{
				Account:          exccjson.String("acct"),
				Count:            exccjson.Int(20),
				From:             exccjson.Int(1),
				IncludeWatchOnly: exccjson.Bool(true),
			},
		},
		{
			name: "listunspent",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listunspent")
			},
			staticCmd: func() interface{} {
				return exccjson.NewListUnspentCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[],"id":1}`,
			unmarshalled: &exccjson.ListUnspentCmd{
				MinConf:   exccjson.Int(1),
				MaxConf:   exccjson.Int(9999999),
				Addresses: nil,
			},
		},
		{
			name: "listunspent optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listunspent", 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListUnspentCmd(exccjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[6],"id":1}`,
			unmarshalled: &exccjson.ListUnspentCmd{
				MinConf:   exccjson.Int(6),
				MaxConf:   exccjson.Int(9999999),
				Addresses: nil,
			},
		},
		{
			name: "listunspent optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listunspent", 6, 100)
			},
			staticCmd: func() interface{} {
				return exccjson.NewListUnspentCmd(exccjson.Int(6), exccjson.Int(100), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[6,100],"id":1}`,
			unmarshalled: &exccjson.ListUnspentCmd{
				MinConf:   exccjson.Int(6),
				MaxConf:   exccjson.Int(100),
				Addresses: nil,
			},
		},
		{
			name: "listunspent optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("listunspent", 6, 100, []string{"1Address", "1Address2"})
			},
			staticCmd: func() interface{} {
				return exccjson.NewListUnspentCmd(exccjson.Int(6), exccjson.Int(100),
					&[]string{"1Address", "1Address2"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[6,100,["1Address","1Address2"]],"id":1}`,
			unmarshalled: &exccjson.ListUnspentCmd{
				MinConf:   exccjson.Int(6),
				MaxConf:   exccjson.Int(100),
				Addresses: &[]string{"1Address", "1Address2"},
			},
		},
		{
			name: "lockunspent",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("lockunspent", true, `[{"txid":"123","vout":1}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []exccjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				return exccjson.NewLockUnspentCmd(true, txInputs)
			},
			marshalled: `{"jsonrpc":"1.0","method":"lockunspent","params":[true,[{"txid":"123","vout":1,"tree":0}]],"id":1}`,
			unmarshalled: &exccjson.LockUnspentCmd{
				Unlock: true,
				Transactions: []exccjson.TransactionInput{
					{Txid: "123", Vout: 1},
				},
			},
		},
		{
			name: "sendfrom",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendfrom", "from", "1Address", 0.5)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendFromCmd("from", "1Address", 0.5, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5],"id":1}`,
			unmarshalled: &exccjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     exccjson.Int(1),
				Comment:     nil,
				CommentTo:   nil,
			},
		},
		{
			name: "sendfrom optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendfrom", "from", "1Address", 0.5, 6)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendFromCmd("from", "1Address", 0.5, exccjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5,6],"id":1}`,
			unmarshalled: &exccjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     exccjson.Int(6),
				Comment:     nil,
				CommentTo:   nil,
			},
		},
		{
			name: "sendfrom optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendfrom", "from", "1Address", 0.5, 6, "comment")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendFromCmd("from", "1Address", 0.5, exccjson.Int(6),
					exccjson.String("comment"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5,6,"comment"],"id":1}`,
			unmarshalled: &exccjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     exccjson.Int(6),
				Comment:     exccjson.String("comment"),
				CommentTo:   nil,
			},
		},
		{
			name: "sendfrom optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendfrom", "from", "1Address", 0.5, 6, "comment", "commentto")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendFromCmd("from", "1Address", 0.5, exccjson.Int(6),
					exccjson.String("comment"), exccjson.String("commentto"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5,6,"comment","commentto"],"id":1}`,
			unmarshalled: &exccjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     exccjson.Int(6),
				Comment:     exccjson.String("comment"),
				CommentTo:   exccjson.String("commentto"),
			},
		},
		{
			name: "sendmany",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendmany", "from", `{"1Address":0.5}`)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return exccjson.NewSendManyCmd("from", amounts, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendmany","params":["from",{"1Address":0.5}],"id":1}`,
			unmarshalled: &exccjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     exccjson.Int(1),
				Comment:     nil,
			},
		},
		{
			name: "sendmany optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendmany", "from", `{"1Address":0.5}`, 6)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return exccjson.NewSendManyCmd("from", amounts, exccjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendmany","params":["from",{"1Address":0.5},6],"id":1}`,
			unmarshalled: &exccjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     exccjson.Int(6),
				Comment:     nil,
			},
		},
		{
			name: "sendmany optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendmany", "from", `{"1Address":0.5}`, 6, "comment")
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return exccjson.NewSendManyCmd("from", amounts, exccjson.Int(6), exccjson.String("comment"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendmany","params":["from",{"1Address":0.5},6,"comment"],"id":1}`,
			unmarshalled: &exccjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     exccjson.Int(6),
				Comment:     exccjson.String("comment"),
			},
		},
		{
			name: "sendtoaddress",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendtoaddress", "1Address", 0.5)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendToAddressCmd("1Address", 0.5, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendtoaddress","params":["1Address",0.5],"id":1}`,
			unmarshalled: &exccjson.SendToAddressCmd{
				Address:   "1Address",
				Amount:    0.5,
				Comment:   nil,
				CommentTo: nil,
			},
		},
		{
			name: "sendtoaddress optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("sendtoaddress", "1Address", 0.5, "comment", "commentto")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSendToAddressCmd("1Address", 0.5, exccjson.String("comment"),
					exccjson.String("commentto"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendtoaddress","params":["1Address",0.5,"comment","commentto"],"id":1}`,
			unmarshalled: &exccjson.SendToAddressCmd{
				Address:   "1Address",
				Amount:    0.5,
				Comment:   exccjson.String("comment"),
				CommentTo: exccjson.String("commentto"),
			},
		},
		{
			name: "settxfee",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("settxfee", 0.0001)
			},
			staticCmd: func() interface{} {
				return exccjson.NewSetTxFeeCmd(0.0001)
			},
			marshalled: `{"jsonrpc":"1.0","method":"settxfee","params":[0.0001],"id":1}`,
			unmarshalled: &exccjson.SetTxFeeCmd{
				Amount: 0.0001,
			},
		},
		{
			name: "signmessage",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("signmessage", "1Address", "message")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSignMessageCmd("1Address", "message")
			},
			marshalled: `{"jsonrpc":"1.0","method":"signmessage","params":["1Address","message"],"id":1}`,
			unmarshalled: &exccjson.SignMessageCmd{
				Address: "1Address",
				Message: "message",
			},
		},
		{
			name: "signrawtransaction",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("signrawtransaction", "001122")
			},
			staticCmd: func() interface{} {
				return exccjson.NewSignRawTransactionCmd("001122", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122"],"id":1}`,
			unmarshalled: &exccjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   nil,
				PrivKeys: nil,
				Flags:    exccjson.String("ALL"),
			},
		},
		{
			name: "signrawtransaction optional1",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("signrawtransaction", "001122", `[{"txid":"123","vout":1,"tree":0,"scriptPubKey":"00","redeemScript":"01"}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []exccjson.RawTxInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: "01",
					},
				}

				return exccjson.NewSignRawTransactionCmd("001122", &txInputs, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122",[{"txid":"123","vout":1,"tree":0,"scriptPubKey":"00","redeemScript":"01"}]],"id":1}`,
			unmarshalled: &exccjson.SignRawTransactionCmd{
				RawTx: "001122",
				Inputs: &[]exccjson.RawTxInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: "01",
					},
				},
				PrivKeys: nil,
				Flags:    exccjson.String("ALL"),
			},
		},
		{
			name: "signrawtransaction optional2",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("signrawtransaction", "001122", `[]`, `["abc"]`)
			},
			staticCmd: func() interface{} {
				txInputs := []exccjson.RawTxInput{}
				privKeys := []string{"abc"}
				return exccjson.NewSignRawTransactionCmd("001122", &txInputs, &privKeys, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122",[],["abc"]],"id":1}`,
			unmarshalled: &exccjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   &[]exccjson.RawTxInput{},
				PrivKeys: &[]string{"abc"},
				Flags:    exccjson.String("ALL"),
			},
		},
		{
			name: "signrawtransaction optional3",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("signrawtransaction", "001122", `[]`, `[]`, "ALL")
			},
			staticCmd: func() interface{} {
				txInputs := []exccjson.RawTxInput{}
				privKeys := []string{}
				return exccjson.NewSignRawTransactionCmd("001122", &txInputs, &privKeys,
					exccjson.String("ALL"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122",[],[],"ALL"],"id":1}`,
			unmarshalled: &exccjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   &[]exccjson.RawTxInput{},
				PrivKeys: &[]string{},
				Flags:    exccjson.String("ALL"),
			},
		},
		{
			name: "verifyseed",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("verifyseed", "abc")
			},
			staticCmd: func() interface{} {
				return exccjson.NewVerifySeedCmd("abc", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifyseed","params":["abc"],"id":1}`,
			unmarshalled: &exccjson.VerifySeedCmd{
				Seed:    "abc",
				Account: nil,
			},
		},
		{
			name: "verifyseed optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("verifyseed", "abc", 5)
			},
			staticCmd: func() interface{} {
				account := exccjson.Uint32(5)
				return exccjson.NewVerifySeedCmd("abc", account)
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifyseed","params":["abc",5],"id":1}`,
			unmarshalled: &exccjson.VerifySeedCmd{
				Seed:    "abc",
				Account: exccjson.Uint32(5),
			},
		},
		{
			name: "walletlock",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("walletlock")
			},
			staticCmd: func() interface{} {
				return exccjson.NewWalletLockCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"walletlock","params":[],"id":1}`,
			unmarshalled: &exccjson.WalletLockCmd{},
		},
		{
			name: "walletpassphrase",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("walletpassphrase", "pass", 60)
			},
			staticCmd: func() interface{} {
				return exccjson.NewWalletPassphraseCmd("pass", 60)
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletpassphrase","params":["pass",60],"id":1}`,
			unmarshalled: &exccjson.WalletPassphraseCmd{
				Passphrase: "pass",
				Timeout:    60,
			},
		},
		{
			name: "walletpassphrasechange",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("walletpassphrasechange", "old", "new")
			},
			staticCmd: func() interface{} {
				return exccjson.NewWalletPassphraseChangeCmd("old", "new")
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletpassphrasechange","params":["old","new"],"id":1}`,
			unmarshalled: &exccjson.WalletPassphraseChangeCmd{
				OldPassphrase: "old",
				NewPassphrase: "new",
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
