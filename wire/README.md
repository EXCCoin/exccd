wire
====

[![Build Status](http://img.shields.io/travis/EXCCoin/exccd.svg)](https://travis-ci.org/EXCCoin/exccd)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/EXCCoin/exccd/wire)

Package wire implements the ExchangeCoin wire protocol.  A comprehensive suite of
tests with 100% test coverage is provided to ensure proper functionality.

This package has intentionally been designed so it can be used as a standalone
package for any projects needing to interface with ExchangeCoin peers at the wire
protocol level.

## Installation and Updating

```bash
$ go get -u github.com/EXCCoin/exccd/wire
```

## ExchangeCoin Message Overview

The ExchangeCoin protocol consists of exchanging messages between peers. Each message
is preceded by a header which identifies information about it such as which
ExchangeCoin network it is a part of, its type, how big it is, and a checksum to
verify validity. All encoding and decoding of message headers is handled by this
package.

To accomplish this, there is a generic interface for ExchangeCoin messages named
`Message` which allows messages of any type to be read, written, or passed
around through channels, functions, etc. In addition, concrete implementations
of most of the currently supported ExchangeCoin messages are provided. For these
supported messages, all of the details of marshalling and unmarshalling to and
from the wire using ExchangeCoin encoding are handled so the caller doesn't have to
concern themselves with the specifics.

## Reading Messages Example

In order to unmarshal ExchangeCoin messages from the wire, use the `ReadMessage`
function. It accepts any `io.Reader`, but typically this will be a `net.Conn`
to a remote node running a ExchangeCoin peer.  Example syntax is:

```Go
	// Use the most recent protocol version supported by the package and the
	// main ExchangeCoin network.
	pver := wire.ProtocolVersion
	exccnet := wire.MainNet

	// Reads and validates the next ExchangeCoin message from conn using the
	// protocol version pver and the ExchangeCoin network exccnet.  The returns
	// are a wire.Message, a []byte which contains the unmarshalled
	// raw payload, and a possible error.
	msg, rawPayload, err := wire.ReadMessage(conn, pver, exccnet)
	if err != nil {
		// Log and handle the error
	}
```

See the package documentation for details on determining the message type.

## Writing Messages Example

In order to marshal ExchangeCoin messages to the wire, use the `WriteMessage`
function. It accepts any `io.Writer`, but typically this will be a `net.Conn`
to a remote node running a ExchangeCoin peer. Example syntax to request addresses
from a remote peer is:

```Go
	// Use the most recent protocol version supported by the package and the
	// main ExchangeCoin network.
	pver := wire.ProtocolVersion
	exccnet := wire.MainNet

	// Create a new getaddr ExchangeCoin message.
	msg := wire.NewMsgGetAddr()

	// Writes a ExchangeCoin message msg to conn using the protocol version
	// pver, and the ExchangeCoin network exccnet.  The return is a possible
	// error.
	err := wire.WriteMessage(conn, msg, pver, exccnet)
	if err != nil {
		// Log and handle the error
	}
```

## License

Package wire is licensed under the [copyfree](http://copyfree.org) ISC
License.
