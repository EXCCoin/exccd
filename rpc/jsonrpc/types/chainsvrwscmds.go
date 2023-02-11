// Copyright (c) 2014-2015 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// NOTE: This file is intended to house the RPC commands that are supported by
// a chain server, but are only available via websockets.

package types

import "github.com/EXCCoin/exccd/dcrjson/v4"

// AuthenticateCmd defines the authenticate JSON-RPC command.
type AuthenticateCmd struct {
	Username   string
	Passphrase string
}

// NewAuthenticateCmd returns a new instance which can be used to issue an
// authenticate JSON-RPC command.
func NewAuthenticateCmd(username, passphrase string) *AuthenticateCmd {
	return &AuthenticateCmd{
		Username:   username,
		Passphrase: passphrase,
	}
}

// OutPoint describes a transaction outpoint that will be marshalled to and
// from JSON.  Contains Decred addition.
type OutPoint struct {
	Hash  string `json:"hash"`
	Tree  int8   `json:"tree"`
	Index uint32 `json:"index"`
}

// LoadTxFilterCmd defines the loadtxfilter request parameters to load or
// reload a transaction filter.
type LoadTxFilterCmd struct {
	Reload    bool
	Addresses []string
	OutPoints []OutPoint
}

// NewLoadTxFilterCmd returns a new instance which can be used to issue a
// loadtxfilter JSON-RPC command.
func NewLoadTxFilterCmd(reload bool, addresses []string, outPoints []OutPoint) *LoadTxFilterCmd {
	return &LoadTxFilterCmd{
		Reload:    reload,
		Addresses: addresses,
		OutPoints: outPoints,
	}
}

// NotifyBlocksCmd defines the notifyblocks JSON-RPC command.
type NotifyBlocksCmd struct{}

// NewNotifyBlocksCmd returns a new instance which can be used to issue a
// notifyblocks JSON-RPC command.
func NewNotifyBlocksCmd() *NotifyBlocksCmd {
	return &NotifyBlocksCmd{}
}

// NotifyWorkCmd defines the notifywork JSON-RPC command.
type NotifyWorkCmd struct{}

// NewNotifyWorkCmd returns a new instance which can be used to issue a
// notifywork JSON-RPC command.
func NewNotifyWorkCmd() *NotifyWorkCmd {
	return &NotifyWorkCmd{}
}

// NotifyTSpendCmd defines the notifytspend JSON-RPC command.
type NotifyTSpendCmd struct{}

// NewNotifyTSpendCmd returns a new instance which can be used to issue a
// notifytspend JSON-RPC command.
func NewNotifyTSpendCmd() *NotifyTSpendCmd {
	return &NotifyTSpendCmd{}
}

// NotifyWinningTicketsCmd is a type handling custom marshaling and
// unmarshaling of notifywinningtickets JSON websocket extension
// commands.
type NotifyWinningTicketsCmd struct {
}

// NewNotifyWinningTicketsCmd creates a new NotifyWinningTicketsCmd.
func NewNotifyWinningTicketsCmd() *NotifyWinningTicketsCmd {
	return &NotifyWinningTicketsCmd{}
}

// NotifySpentAndMissedTicketsCmd is a type handling custom marshaling and
// unmarshaling of notifyspentandmissedtickets JSON websocket extension
// commands.
type NotifySpentAndMissedTicketsCmd struct {
}

// NewNotifySpentAndMissedTicketsCmd creates a new NotifySpentAndMissedTicketsCmd.
func NewNotifySpentAndMissedTicketsCmd() *NotifySpentAndMissedTicketsCmd {
	return &NotifySpentAndMissedTicketsCmd{}
}

// NotifyNewTicketsCmd is a type handling custom marshaling and
// unmarshaling of notifynewtickets JSON websocket extension
// commands.
type NotifyNewTicketsCmd struct {
}

// NewNotifyNewTicketsCmd creates a new NotifyNewTicketsCmd.
func NewNotifyNewTicketsCmd() *NotifyNewTicketsCmd {
	return &NotifyNewTicketsCmd{}
}

// RebroadcastMissedCmd is a type handling custom marshaling and
// unmarshaling of rebroadcastmissed JSON RPC commands.
type RebroadcastMissedCmd struct{}

// NewRebroadcastMissedCmd returns a new instance which can be used to
// issue a JSON-RPC rebroadcastmissed command.
func NewRebroadcastMissedCmd() *RebroadcastMissedCmd {
	return &RebroadcastMissedCmd{}
}

// RebroadcastWinnersCmd is a type handling custom marshaling and
// unmarshaling of rebroadcastwinners JSON RPC commands.
type RebroadcastWinnersCmd struct{}

// NewRebroadcastWinnersCmd returns a new instance which can be used to
// issue a JSON-RPC rebroadcastwinners command.
func NewRebroadcastWinnersCmd() *RebroadcastWinnersCmd {
	return &RebroadcastWinnersCmd{}
}

// StopNotifyBlocksCmd defines the stopnotifyblocks JSON-RPC command.
type StopNotifyBlocksCmd struct{}

// NewStopNotifyBlocksCmd returns a new instance which can be used to issue a
// stopnotifyblocks JSON-RPC command.
func NewStopNotifyBlocksCmd() *StopNotifyBlocksCmd {
	return &StopNotifyBlocksCmd{}
}

// StopNotifyWorkCmd defines the stopnotifywork JSON-RPC command.
type StopNotifyWorkCmd struct{}

// NewStopNotifyWorkCmd returns a new instance which can be used to issue a
// stopnotifywork JSON-RPC command.
func NewStopNotifyWorkCmd() *StopNotifyWorkCmd {
	return &StopNotifyWorkCmd{}
}

// StopNotifyTSpendCmd defines the stopnotifytspend JSON-RPC command.
type StopNotifyTSpendCmd struct{}

// NewStopNotifyTSpendCmd returns a new instance which can be used to issue a
// stopnotifytspend JSON-RPC command.
func NewStopNotifyTSpendCmd() *StopNotifyTSpendCmd {
	return &StopNotifyTSpendCmd{}
}

// NotifyNewTransactionsCmd defines the notifynewtransactions JSON-RPC command.
type NotifyNewTransactionsCmd struct {
	Verbose *bool `jsonrpcdefault:"false"`
}

// NewNotifyNewTransactionsCmd returns a new instance which can be used to issue
// a notifynewtransactions JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewNotifyNewTransactionsCmd(verbose *bool) *NotifyNewTransactionsCmd {
	return &NotifyNewTransactionsCmd{
		Verbose: verbose,
	}
}

// SessionCmd defines the session JSON-RPC command.
type SessionCmd struct{}

// NewSessionCmd returns a new instance which can be used to issue a session
// JSON-RPC command.
func NewSessionCmd() *SessionCmd {
	return &SessionCmd{}
}

// StopNotifyNewTransactionsCmd defines the stopnotifynewtransactions JSON-RPC command.
type StopNotifyNewTransactionsCmd struct{}

// NewStopNotifyNewTransactionsCmd returns a new instance which can be used to issue
// a stopnotifynewtransactions JSON-RPC command.
//
// The parameters which are pointers indicate they are optional.  Passing nil
// for optional parameters will use the default value.
func NewStopNotifyNewTransactionsCmd() *StopNotifyNewTransactionsCmd {
	return &StopNotifyNewTransactionsCmd{}
}

// RescanCmd defines the rescan JSON-RPC command.
type RescanCmd struct {
	BlockHashes []string
}

// NewRescanCmd returns a new instance which can be used to issue a rescan
// JSON-RPC command.
func NewRescanCmd(blockHashes []string) *RescanCmd {
	return &RescanCmd{BlockHashes: blockHashes}
}

func init() {
	// The commands in this file are only usable by websockets.
	flags := dcrjson.UFWebsocketOnly

	dcrjson.MustRegister(Method("authenticate"), (*AuthenticateCmd)(nil), flags)
	dcrjson.MustRegister(Method("loadtxfilter"), (*LoadTxFilterCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifyblocks"), (*NotifyBlocksCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifywork"), (*NotifyWorkCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifytspend"), (*NotifyTSpendCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifynewtransactions"), (*NotifyNewTransactionsCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifynewtickets"), (*NotifyNewTicketsCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifyspentandmissedtickets"), (*NotifySpentAndMissedTicketsCmd)(nil), flags)
	dcrjson.MustRegister(Method("notifywinningtickets"), (*NotifyWinningTicketsCmd)(nil), flags)
	dcrjson.MustRegister(Method("rebroadcastmissed"), (*RebroadcastMissedCmd)(nil), flags)
	dcrjson.MustRegister(Method("rebroadcastwinners"), (*RebroadcastWinnersCmd)(nil), flags)
	dcrjson.MustRegister(Method("session"), (*SessionCmd)(nil), flags)
	dcrjson.MustRegister(Method("stopnotifyblocks"), (*StopNotifyBlocksCmd)(nil), flags)
	dcrjson.MustRegister(Method("stopnotifywork"), (*StopNotifyWorkCmd)(nil), flags)
	dcrjson.MustRegister(Method("stopnotifytspend"), (*StopNotifyTSpendCmd)(nil), flags)
	dcrjson.MustRegister(Method("stopnotifynewtransactions"), (*StopNotifyNewTransactionsCmd)(nil), flags)
	dcrjson.MustRegister(Method("rescan"), (*RescanCmd)(nil), flags)
}
