module chainec

go 1.19

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/dcrec/edwards/v2 v2.0.2
	github.com/decred/dcrd/dcrec/secp256k1/v2 v2.0.0
)

require (
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/decred/dcrd/chaincfg/chainhash v1.0.2 // indirect
)

replace (
	github.com/decred/dcrd/dcrec/edwards/v2 => ../../dcrec/edwards
	github.com/decred/dcrd/dcrec/secp256k1/v2 => ../../dcrec/secp256k1
	github.com/decred/dcrd/chaincfg/chainhash => ../chainhash
)