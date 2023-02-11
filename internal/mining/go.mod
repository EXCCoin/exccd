module github.com/decred/dcrd/mining/v2

go 1.11

replace (
	github.com/decred/dcrd/blockchain/stake/v3 => ./../../blockchain/stake
	github.com/decred/dcrd/blockchain/standalone/v2 => ./../../blockchain/standalone
	github.com/decred/dcrd/blockchain/v3 => ./../../blockchain
	github.com/decred/dcrd/cequihash => ./../../cequihash
	github.com/decred/dcrd/chaincfg/chainhash => ./../../chaincfg/chainhash
	github.com/decred/dcrd/chaincfg/v3 => ../../chaincfg
	github.com/decred/dcrd/dcrec/secp256k1 => ./../../dcrec/secp256k1
	github.com/decred/dcrd/dcrutil/v3 => ./../../dcrutil
	github.com/decred/dcrd/internal/mining => ./../../internal/mining
	github.com/decred/dcrd/wire => ./../../wire
)

require (
	github.com/decred/dcrd/blockchain/stake/v4 v4.0.0
	github.com/decred/dcrd/blockchain/standalone/v2 v2.1.0
	github.com/decred/dcrd/blockchain/v4 v4.1.0
	github.com/decred/dcrd/cequihash v0.0.0-00010101000000-000000000000 // indirect
	github.com/decred/dcrd/chaincfg/chainhash v1.0.3
	github.com/decred/dcrd/chaincfg/v3 v3.1.1
	github.com/decred/dcrd/dcrec v1.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0
	github.com/decred/dcrd/dcrutil/v4 v4.0.0
	github.com/decred/dcrd/gcs/v3 v3.0.0
	github.com/decred/dcrd/lru v1.1.1
	github.com/decred/dcrd/txscript/v4 v4.0.0
	github.com/decred/dcrd/wire v1.5.0
	github.com/decred/slog v1.2.0
	github.com/mattn/go-pointer v0.0.1 // indirect
)
