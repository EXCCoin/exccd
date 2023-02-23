peer
====

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/peer/v2)

Package peer provides a common base for creating and managing Exchangecoin network
peers.

This package has intentionally been designed so it can be used as a standalone
package for any projects needing a full featured Exchangecoin peer base to build on.

## Overview

This package builds upon the wire package, which provides the fundamental
primitives necessary to speak the Exchangecoin wire protocol, in order to simplify
the process of creating fully functional peers.  In essence, it provides a
common base for creating concurrent safe fully validating nodes, Simplified
Payment Verification (SPV) nodes, proxies, etc.

A quick overview of the major features peer provides are as follows:

 - Provides a basic concurrent safe Exchangecoin peer for handling Exchangecoin
   communications via the peer-to-peer protocol
 - Full duplex reading and writing of Exchangecoin protocol messages
 - Automatic handling of the initial handshake process including protocol
   version negotiation
 - Asynchronous message queueing of outbound messages with optional channel for
   notification when the message is actually sent
 - Flexible peer configuration
   - Caller is responsible for creating outgoing connections and listening for
     incoming connections so they have flexibility to establish connections as
     they see fit (proxies, etc)
   - User agent name and version
   - Exchangecoin network
   - Service support signalling (full nodes, etc)
   - Maximum supported protocol version
   - Ability to register callbacks for handling Exchangecoin protocol messages
 - Inventory message batching and send trickling with known inventory detection
   and avoidance
 - Automatic periodic keep-alive pinging and pong responses
 - Random nonce generation and self connection detection
 - Snapshottable peer statistics such as the total number of bytes read and
   written, the remote address, user agent, and negotiated protocol version
 - Helper functions pushing addresses, getblocks, getheaders, and reject
   messages
   - These could all be sent manually via the standard message output function,
     but the helpers provide additional nice functionality such as duplicate
     filtering and address randomization
 - Ability to wait for shutdown/disconnect
 - Comprehensive test coverage

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd/peer/v2` module.  Use the
standard go tooling for working with modules to incorporate it.

## Examples

* [New Outbound Peer Example](https://pkg.go.dev/github.com/EXCCoin/exccd/peer/v2#example-package-NewOutboundPeer)
  Demonstrates the basic process for initializing and creating an outbound peer.
  Peers negotiate by exchanging version and verack messages.  For demonstration,
  a simple handler for the version message is attached to the peer.

## License

Package peer is licensed under the [copyfree](http://copyfree.org) ISC License.
