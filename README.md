exccd
====

[![Build Status](https://travis-ci.org/EXCCoin/exccd.png?branch=master)](https://travis-ci.org/EXCCoin/exccd)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/EXCCoin/exccd)

exccd is a EXCC full node implementation written in Go (golang).

This acts as a chain daemon for the [EXCC](https://cryptoxchanger.io/) cryptocurrency.
The exccd maintains the entire past transactional ledger of EXCC and allows
relaying of transactions to other EXCC nodes across the world. To read more
about EXCC please see the
[project documentation](https://cryptoxchanger.io/faq).

Note: To send or receive funds and join Proof-of-Stake mining, you will also need
[exccwallet](https://github.com/EXCCoin/exccwallet).

This project is currently under active development and is in a Beta state.

It is forked from [dcrd](https://github.com/decred/dcrd) which is a Decred
full node implementation written in Go.  dcrd is a ongoing project under active
development.  Because exccd is constantly synced with dcrd codebase, it will
get the benefit of dcrd's ongoing upgrades to peer and connection handling,
database optimization and other blockchain related technology improvements.

## Requirements

[Go](http://golang.org) 1.9 or newer.

## Getting Started

- exccd (and utilities) will now be installed in either ```$GOROOT/bin``` or
  ```$GOPATH/bin``` depending on your configuration.  If you did not already
  add the bin directory to your system path during Go installation, we
  recommend you do so now.

## Updating

#### Windows

Install a newer MSI

#### Linux/BSD/MacOSX/POSIX - Build from Source

- **Dep**

  Dep is used to manage project dependencies and provide reproducible builds.
  To install:

  `go get -u github.com/golang/dep/cmd/dep`

Unfortunately, the use of `dep` prevents a handy tool such as `go get` from
automatically downloading, building, and installing the source in a single
command.  Instead, the latest project and dependency sources must be first
obtained manually with `git` and `dep`, and then `go` is used to build and
install the project.

**Getting the source**:

For a first time installation, the project and dependency sources can be
obtained manually with `git` and `dep` (create directories as needed):

```
git clone https://github.com/EXCCoin/exccd $GOPATH/src/github.com/EXCCoin/exccd
cd $GOPATH/src/github.com/EXCCoin/exccd
dep ensure
go install . ./cmd/...
```

To update an existing source tree, pull the latest changes and install the
matching dependencies:

```
cd $GOPATH/src/github.com/EXCCoin/exccd
git pull
dep ensure
go install . ./cmd/...
```

## Docker

All tests and linters may be run in a docker container using the script
`run_tests.sh`.  This script defaults to using the current supported version of
go.  You can run it with the major version of Go you would like to use as the
only arguement to test a previous on a previous version of Go (generally Exchangecoin
supports the current version of Go and the previous one).

```
./run_tests.sh 1.9
```

To run the tests locally without docker:

```
go test ./...
```

## Issue Tracker

The [integrated github issue tracker](https://github.com/EXCCoin/exccd/issues)
is used for this project.

## Documentation

The documentation is a work-in-progress.  It is located in the
[docs](https://github.com/EXCCoin/exccd/tree/master/docs) folder.

## License

exccd is licensed under the [copyfree](http://copyfree.org) ISC License.
