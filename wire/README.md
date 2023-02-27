wire
====

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/wire)

Package wire implements the Exchangecoin wire protocol.  A comprehensive suite of
tests with 100% test coverage is provided to ensure proper functionality.

This package has intentionally been designed so it can be used as a standalone
package for any projects needing to interface with Exchangecoin peers at the wire
protocol level.

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd/wire` module.  Use the
standard go tooling for working with modules to incorporate it.

## Exchangecoin Message Overview

The Exchangecoin protocol consists of exchanging messages between peers. Each message
is preceded by a header which identifies information about it such as which
Exchangecoin network it is a part of, its type, how big it is, and a checksum to
verify validity. All encoding and decoding of message headers is handled by this
package.

To accomplish this, there is a generic interface for Exchangecoin messages named
`Message` which allows messages of any type to be read, written, or passed
around through channels, functions, etc. In addition, concrete implementations
of most of the currently supported Exchangecoin messages are provided. For these
supported messages, all of the details of marshalling and unmarshalling to and
from the wire using Exchangecoin encoding are handled so the caller doesn't have to
concern themselves with the specifics.

## Reading Messages Example

In order to unmarshal Exchangecoin messages from the wire, use the `ReadMessage`
function. It accepts any `io.Reader`, but typically this will be a `net.Conn`
to a remote node running a Exchangecoin peer.  Example syntax is:

```Go
	// Use the most recent protocol version supported by the package and the
	// main Exchangecoin network.
	pver := wire.ProtocolVersion
	dcrnet := wire.MainNet

	// Reads and validates the next Exchangecoin message from conn using the
	// protocol version pver and the Exchangecoin network dcrnet.  The returns
	// are a wire.Message, a []byte which contains the unmarshalled
	// raw payload, and a possible error.
	msg, rawPayload, err := wire.ReadMessage(conn, pver, dcrnet)
	if err != nil {
		// Log and handle the error
	}
```

See the package documentation for details on determining the message type.

## Writing Messages Example

In order to marshal Exchangecoin messages to the wire, use the `WriteMessage`
function. It accepts any `io.Writer`, but typically this will be a `net.Conn`
to a remote node running a Exchangecoin peer. Example syntax to request addresses
from a remote peer is:

```Go
	// Use the most recent protocol version supported by the package and the
	// main Exchangecoin network.
	pver := wire.ProtocolVersion
	dcrnet := wire.MainNet

	// Create a new getaddr Exchangecoin message.
	msg := wire.NewMsgGetAddr()

	// Writes a Exchangecoin message msg to conn using the protocol version
	// pver, and the Exchangecoin network dcrnet.  The return is a possible
	// error.
	err := wire.WriteMessage(conn, msg, pver, dcrnet)
	if err != nil {
		// Log and handle the error
	}
```

## License

Package wire is licensed under the [copyfree](http://copyfree.org) ISC
License.
