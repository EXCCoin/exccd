sampleconfig
============

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/sampleconfig)

Package sampleconfig provides a single function that returns the contents of
the sample configuration file for exccd.  This is provided for tools that perform
automatic configuration and would like to ensure the generated configuration
file not only includes the specifically configured values, but also provides
samples of other configuration options.

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd` module.  Use the standard
go tooling for working with modules to incorporate it.

## License

Package sampleconfig is licensed under the [copyfree](http://copyfree.org) ISC
License.
