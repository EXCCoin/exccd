// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
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

// TestChainSvrWsCmds tests all of the chain server websocket-specific commands
// marshal and unmarshal into valid results include handling of optional fields
// being omitted in the marshalled command, while optional fields with defaults
// have the default assigned on unmarshalled commands.
func TestChainSvrWsCmds(t *testing.T) {
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
			name: "authenticate",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("authenticate", "user", "pass")
			},
			staticCmd: func() interface{} {
				return exccjson.NewAuthenticateCmd("user", "pass")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"authenticate","params":["user","pass"],"id":1}`,
			unmarshalled: &exccjson.AuthenticateCmd{Username: "user", Passphrase: "pass"},
		},
		{
			name: "notifywinningtickets",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifywinningtickets")
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifyWinningTicketsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifywinningtickets","params":[],"id":1}`,
			unmarshalled: &exccjson.NotifyWinningTicketsCmd{},
		},
		{
			name: "notifyspentandmissedtickets",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifyspentandmissedtickets")
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifySpentAndMissedTicketsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyspentandmissedtickets","params":[],"id":1}`,
			unmarshalled: &exccjson.NotifySpentAndMissedTicketsCmd{},
		},
		{
			name: "notifynewtickets",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifynewtickets")
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifyNewTicketsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifynewtickets","params":[],"id":1}`,
			unmarshalled: &exccjson.NotifyNewTicketsCmd{},
		},
		{
			name: "notifystakedifficulty",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifystakedifficulty")
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifyStakeDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifystakedifficulty","params":[],"id":1}`,
			unmarshalled: &exccjson.NotifyStakeDifficultyCmd{},
		},
		{
			name: "notifyblocks",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifyblocks")
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"notifyblocks","params":[],"id":1}`,
			unmarshalled: &exccjson.NotifyBlocksCmd{},
		},
		{
			name: "stopnotifyblocks",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("stopnotifyblocks")
			},
			staticCmd: func() interface{} {
				return exccjson.NewStopNotifyBlocksCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopnotifyblocks","params":[],"id":1}`,
			unmarshalled: &exccjson.StopNotifyBlocksCmd{},
		},
		{
			name: "notifynewtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifynewtransactions")
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifyNewTransactionsCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifynewtransactions","params":[],"id":1}`,
			unmarshalled: &exccjson.NotifyNewTransactionsCmd{
				Verbose: exccjson.Bool(false),
			},
		},
		{
			name: "notifynewtransactions optional",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("notifynewtransactions", true)
			},
			staticCmd: func() interface{} {
				return exccjson.NewNotifyNewTransactionsCmd(exccjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"notifynewtransactions","params":[true],"id":1}`,
			unmarshalled: &exccjson.NotifyNewTransactionsCmd{
				Verbose: exccjson.Bool(true),
			},
		},
		{
			name: "stopnotifynewtransactions",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("stopnotifynewtransactions")
			},
			staticCmd: func() interface{} {
				return exccjson.NewStopNotifyNewTransactionsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stopnotifynewtransactions","params":[],"id":1}`,
			unmarshalled: &exccjson.StopNotifyNewTransactionsCmd{},
		},
		{
			name: "rescan",
			newCmd: func() (interface{}, error) {
				return exccjson.NewCmd("rescan", "0000000000000000000000000000000000000000000000000000000000000123")
			},
			staticCmd: func() interface{} {
				return exccjson.NewRescanCmd("0000000000000000000000000000000000000000000000000000000000000123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"rescan","params":["0000000000000000000000000000000000000000000000000000000000000123"],"id":1}`,
			unmarshalled: &exccjson.RescanCmd{
				BlockHashes: "0000000000000000000000000000000000000000000000000000000000000123",
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
