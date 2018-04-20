bloom
=====

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/EXCCoin/exccd/bloom)

Package bloom provides an API for dealing with excc-specific bloom filters.

A comprehensive suite of tests is provided to ensure proper functionality.

## Installation and Updating

```bash
$ go get -u github.com/EXCCoin/exccd/bloom
```

## Examples

* [NewFilter Example](https://godoc.org/github.com/EXCCoin/exccd/bloom#example-package--NewFilter)
  Demonstrates how to create a new bloom filter, add a transaction hash to it,
  and check if the filter matches the transaction.

## License

Package bloom is licensed under the [copyfree](http://copyfree.org) ISC
License.
