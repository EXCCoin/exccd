txscript
========

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/txscript/v4)

Package txscript implements the Exchangecoin transaction script language.  There is
a comprehensive test suite.

This package has intentionally been designed so it can be used as a standalone
package for any projects needing to use or validate Exchangecoin transaction scripts.

## Exchangecoin Scripts

Exchangecoin provides a stack-based, FORTH-like language for the scripts in the Exchangecoin
transactions.  This language is not Turing complete although it is still fairly
powerful.

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd/txscript/v3` module.  Use
the standard go tooling for working with modules to incorporate it.

## Examples

* [Counting Opcodes in Scripts](https://pkg.go.dev/github.com/EXCCoin/exccd/txscript/v4#example-ScriptTokenizer)
  Demonstrates creating a script tokenizer instance and using it to count the
  number of opcodes a script contains.

## License

Package txscript is licensed under the [copyfree](http://copyfree.org) ISC
License.
