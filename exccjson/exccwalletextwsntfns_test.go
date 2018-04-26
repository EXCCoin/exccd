// Copyright (c) 2018 The ExchangeCoin team
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

// TestChainSvrWsNtfns tests all of the chain server websocket-specific
// notifications marshal and unmarshal into valid results include handling of
// optional fields being omitted in the marshalled command, while optional
// fields with defaults have the default assigned on unmarshalled commands.
func TestDcrwalletChainSvrWsNtfns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		newNtfn      func() (interface{}, error)
		staticNtfn   func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "ticketpurchase",
			newNtfn: func() (interface{}, error) {
				return exccjson.NewCmd("ticketpurchased", "123", 5)
			},
			staticNtfn: func() interface{} {
				return exccjson.NewTicketPurchasedNtfn("123", 5)
			},
			marshalled: `{"jsonrpc":"1.0","method":"ticketpurchased","params":["123",5],"id":null}`,
			unmarshalled: &exccjson.TicketPurchasedNtfn{
				TxHash: "123",
				Amount: 5,
			},
		},
		{
			name: "votecreated",
			newNtfn: func() (interface{}, error) {
				return exccjson.NewCmd("votecreated", "123", "1234", 100, "12345", 1)
			},
			staticNtfn: func() interface{} {
				return exccjson.NewVoteCreatedNtfn("123", "1234", 100, "12345", 1)
			},
			marshalled: `{"jsonrpc":"1.0","method":"votecreated","params":["123","1234",100,"12345",1],"id":null}`,
			unmarshalled: &exccjson.VoteCreatedNtfn{
				TxHash:    "123",
				BlockHash: "1234",
				Height:    100,
				SStxIn:    "12345",
				VoteBits:  1,
			},
		},
		{
			name: "revocationcreated",
			newNtfn: func() (interface{}, error) {
				return exccjson.NewCmd("revocationcreated", "123", "1234")
			},
			staticNtfn: func() interface{} {
				return exccjson.NewRevocationCreatedNtfn("123", "1234")
			},
			marshalled: `{"jsonrpc":"1.0","method":"revocationcreated","params":["123","1234"],"id":null}`,
			unmarshalled: &exccjson.RevocationCreatedNtfn{
				TxHash: "123",
				SStxIn: "1234",
			},
		},
		{
			name: "winningtickets",
			newNtfn: func() (interface{}, error) {
				return exccjson.NewCmd("winningtickets", "123", 100, map[string]string{"a": "b"})
			},
			staticNtfn: func() interface{} {
				return exccjson.NewWinningTicketsNtfn("123", 100, map[string]string{"a": "b"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"winningtickets","params":["123",100,{"a":"b"}],"id":null}`,
			unmarshalled: &exccjson.WinningTicketsNtfn{
				BlockHash:   "123",
				BlockHeight: 100,
				Tickets:     map[string]string{"a": "b"},
			},
		},
		{
			name: "spentandmissedtickets",
			newNtfn: func() (interface{}, error) {
				return exccjson.NewCmd("spentandmissedtickets", "123", 100, 3, map[string]string{"a": "b"})
			},
			staticNtfn: func() interface{} {
				return exccjson.NewSpentAndMissedTicketsNtfn("123", 100, 3, map[string]string{"a": "b"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"spentandmissedtickets","params":["123",100,3,{"a":"b"}],"id":null}`,
			unmarshalled: &exccjson.SpentAndMissedTicketsNtfn{
				Hash:      "123",
				Height:    100,
				StakeDiff: 3,
				Tickets:   map[string]string{"a": "b"},
			},
		},
		{
			name: "newtickets",
			newNtfn: func() (interface{}, error) {
				return exccjson.NewCmd("newtickets", "123", 100, 3, []string{"a", "b"})
			},
			staticNtfn: func() interface{} {
				return exccjson.NewNewTicketsNtfn("123", 100, 3, []string{"a", "b"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"newtickets","params":["123",100,3,["a","b"]],"id":null}`,
			unmarshalled: &exccjson.NewTicketsNtfn{
				Hash:      "123",
				Height:    100,
				StakeDiff: 3,
				Tickets:   []string{"a", "b"},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the notification as created by the new static
		// creation function.  The ID is nil for notifications.
		marshalled, err := exccjson.MarshalCmd("1.0", nil, test.staticNtfn())
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

		// Ensure the notification is created without error via the
		// generic new notification creation function.
		cmd, err := test.newNtfn()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCmd error: %v ",
				i, test.name, err)
		}

		// Marshal the notification as created by the generic new
		// notification creation function.    The ID is nil for
		// notifications.
		marshalled, err = exccjson.MarshalCmd("1.0", nil, cmd)
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
