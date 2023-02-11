module github.com/decred/dcrd/chaincfg/v3

go 1.11

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/chaincfg/chainhash v1.0.3
	github.com/decred/dcrd/wire v1.5.0
)

replace (
	github.com/decred/dcrd/chaincfg/chainhash => ../chaincfg/chainhash
	github.com/decred/dcrd/dcrec/edwards/v2 => ../dcrec/edwards
	github.com/decred/dcrd/dcrec/secp256k1/v2 => ../dcrec/secp256k1
	github.com/decred/dcrd/wire => ../wire
)
