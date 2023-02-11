module github.com/decred/dcrd/blockchain/stake/v4

go 1.16

require (
	github.com/decred/dcrd/chaincfg/chainhash v1.0.3
	github.com/decred/dcrd/chaincfg/v3 v3.1.0
	github.com/decred/dcrd/database/v3 v3.0.0
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1
	github.com/decred/dcrd/dcrutil/v4 v4.0.0
	github.com/decred/dcrd/txscript/v4 v4.0.0
	github.com/decred/dcrd/wire v1.5.0
	github.com/decred/slog v1.2.0
)

replace (
	github.com/decred/dcrd/chaincfg/chainhash => ../../chaincfg/chainhash
	github.com/decred/dcrd/chaincfg/v2 => ../../chaincfg
	github.com/decred/dcrd/database/v2 => ../../database
	github.com/decred/dcrd/dcrec => ../../dcrec
	github.com/decred/dcrd/dcrutil/v2 => ../../dcrutil
	github.com/decred/dcrd/txscript/v2 => ../../txscript
	github.com/decred/dcrd/wire => ../../wire
)