rpctest
=======

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/rpctest)

Package rpctest provides a exccd-specific RPC testing harness crafting and
executing integration tests by driving a `exccd` instance via the `RPC`
interface. Each instance of an active harness comes equipped with a simple
in-memory HD wallet capable of properly syncing to the generated chain,
creating new addresses, and crafting fully signed transactions paying to an
arbitrary set of outputs. 

This package was designed specifically to act as an RPC testing harness for
`exccd`. However, the constructs presented are general enough to be adapted to
any project wishing to programmatically drive a `exccd` instance of its
systems/integration tests. 

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd` module.  Use the standard
go tooling for working with modules to incorporate it.

## License


Package rpctest is licensed under the [copyfree](http://copyfree.org) ISC
License.

