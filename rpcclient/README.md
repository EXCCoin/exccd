rpcclient
=========

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/rpcclient/v7)

rpcclient implements a Websocket-enabled Exchangecoin JSON-RPC client package written
in [Go](https://golang.org/).  It provides a robust and easy to use client for
interfacing with a Exchangecoin RPC server that uses a exccd compatible Exchangecoin
JSON-RPC API.

## Status

This package is currently under active development.  It is already stable and
the infrastructure is complete.  However, there are still several RPCs left to
implement and the API is not stable yet.

## Documentation

* [API Reference](https://pkg.go.dev/github.com/EXCCoin/exccd/rpcclient/v7)
* [exccd Websockets Example](https://github.com/EXCCoin/exccd/tree/master/rpcclient/examples/exccdwebsockets)
  Connects to a exccd RPC server using TLS-secured websockets, registers for
  block connected and block disconnected notifications, and gets the current
  block count

## Major Features

* Supports Websockets (exccd/exccwallet) and HTTP POST mode (bitcoin core-like)
* Provides callback and registration functions for exccd notifications
* Translates to and from higher-level and easier to use Go types
* Offers a synchronous (blocking) and asynchronous API
* When running in Websockets mode (the default):
  * Automatic reconnect handling (can be disabled)
  * Outstanding commands are automatically reissued
  * Registered notifications are automatically reregistered
  * Back-off support on reconnect attempts

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd/rpcclient/v7` module.  Use
the standard go tooling for working with modules to incorporate it.

## License

Package rpcclient is licensed under the [copyfree](http://copyfree.org) ISC
License.
