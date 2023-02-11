module github.com/decred/dcrd/blockchain/v4

go 1.16

require (
	github.com/decred/dcrd/blockchain/stake/v4 v4.0.0
	github.com/decred/dcrd/blockchain/standalone/v2 v2.1.0
	github.com/decred/dcrd/chaincfg/chainhash v1.0.3
	github.com/decred/dcrd/chaincfg/v3 v3.1.1
	github.com/decred/dcrd/database/v3 v3.0.0
	github.com/decred/dcrd/dcrec v1.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1
	github.com/decred/dcrd/dcrutil/v4 v4.0.0
	github.com/decred/dcrd/gcs/v3 v3.0.0
	github.com/decred/dcrd/lru v1.1.1
	github.com/decred/dcrd/txscript/v4 v4.0.0
	github.com/decred/dcrd/wire v1.5.0
	github.com/decred/slog v1.2.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)

replace (
	github.com/decred/dcrd/blockchain/stake/v2 => ./stake
	github.com/decred/dcrd/blockchain/standalone/v2 => ./standalone
	github.com/decred/dcrd/chaincfg/chainhash => ../chaincfg/chainhash
	github.com/decred/dcrd/chaincfg/v3 => ../chaincfg
	github.com/decred/dcrd/database/v2 => ../database
	github.com/decred/dcrd/dcrec => ../dcrec
	github.com/decred/dcrd/dcrec/secp256k1/v2 => ../dcrec/secp256k1
	github.com/decred/dcrd/dcrutil/v2 => ../dcrutil
	github.com/decred/dcrd/gcs/v2 => ../gcs
	github.com/decred/dcrd/txscript/v2 => ../txscript
	github.com/decred/dcrd/wire => ../wire
)