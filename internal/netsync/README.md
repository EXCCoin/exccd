netsync
=======

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/internal/netsync)

Package netsync implements a concurrency safe block syncing protocol.

## Overview

The provided implementation of SyncManager communicates with connected peers to
perform an initial block download, keep the chain in sync, and announce new
blocks connected to the chain. Currently the sync manager selects a single sync
peer that it downloads all blocks from until it is up to date with the longest
chain the sync peer is aware of.

## License

Package netsync is licensed under the [copyfree](http://copyfree.org) ISC
License.
