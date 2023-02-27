connmgr
=======

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/connmgr/v3)

Package connmgr implements a generic Exchangecoin network connection manager.

## Overview

This package handles all the general connection concerns such as maintaining a
set number of outbound connections, sourcing peers, banning, limiting max
connections, tor lookup, etc.

The package provides a generic connection manager which is able to accept
connection requests from a source or a set of given addresses, dial them and
notify the caller on connections.  The main intended use is to initialize a pool
of active connections and maintain them to remain connected to the P2P network.

In addition the connection manager provides the following utilities:

- Notifications on connections or disconnections
- Handle failures and retry new addresses from the source
- Connect only to specified addresses
- Permanent connections with increasing backoff retry timers
- Disconnect or Remove an established connection

## Installation and Updating

This package is part of the `github.com/EXCCoin/exccd/connmgr/v3` module.  Use the
standard go tooling for working with modules to incorporate it.

## License

Package connmgr is licensed under the [copyfree](http://copyfree.org) ISC License.
