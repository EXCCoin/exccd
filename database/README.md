database
========

[![Build Status](https://github.com/EXCCoin/exccd/workflows/Build%20and%20Test/badge.svg)](https://github.com/EXCCoin/exccd/actions)
[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![Doc](https://img.shields.io/badge/doc-reference-blue.svg)](https://pkg.go.dev/github.com/EXCCoin/exccd/database/v2)

Package database provides a block and metadata storage database.

Please note that this package is intended to enable exccd to support different
database backends and is not something that a client can directly access as only
one entity can have the database open at a time (for most database backends),
and that entity will be exccd.

When a client wants programmatic access to the data provided by exccd, they'll
likely want to use the [rpcclient](https://github.com/EXCCoin/exccd/tree/master/rpcclient)
package which makes use of the
[JSON-RPC API](https://github.com/EXCCoin/exccd/tree/master/docs/json_rpc_api.mediawiki).

However, this package could be extremely useful for any applications requiring
Exchangecoin block storage capabilities.

The default backend, ffldb, has a strong focus on speed, efficiency, and
robustness.  It makes use of leveldb for the metadata, flat files for block
storage, and strict checksums in key areas to ensure data integrity.

## Feature Overview

- Key/value metadata store
- Exchangecoin block storage
- Efficient retrieval of block headers and regions (transactions, scripts, etc)
- Read-only and read-write transactions with both manual and managed modes
- Nested buckets
- Iteration support including cursors with seek capability
- Supports registration of backend databases
- Comprehensive test coverage

## Installation

```bash
$ go get github.com/EXCCoin/exccd/database2
```

## Examples

* [Basic Usage Example](https://pkg.go.dev/github.com/EXCCoin/exccd/database/v2#example-package-BasicUsage)
  Demonstrates creating a new database and using a managed read-write
  transaction to store and retrieve metadata.

* [Block Storage and Retrieval Example](https://pkg.go.dev/github.com/EXCCoin/exccd/database/v2#example-package-BlockStorageAndRetrieval)
  Demonstrates creating a new database, using a managed read-write transaction
  to store a block, and then using a managed read-only transaction to fetch the
  block.

## License

Package database is licensed under the [copyfree](http://copyfree.org) ISC
License.
