jsonrpc/types
=============

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/rpc/jsonrpc/types/v3)

Package types implements concrete types for marshalling to and from the exccd
JSON-RPC commands, return values, and notifications.  A comprehensive suite of
tests is provided to ensure proper functionality.

The provided types are automatically registered with
[dcrjson](https://github.com/EXCCoin/exccd/tree/master/dcrjson) when the package
is imported.  Although this package was primarily written for exccd, it has
intentionally been designed so it can be used as a standalone package for any
projects needing to marshal to and from exccd JSON-RPC requests and responses.

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd/rpc/jsonrpc/types/v2`
module.  Use the standard go tooling for working with modules to incorporate it.

## License

Package types is licensed under the [copyfree](http://copyfree.org) ISC License.
