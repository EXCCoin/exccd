module github.com/decred/dcrd/txscript/v4

go 1.16

require (
	github.com/dchest/siphash v1.2.2
	github.com/decred/base58 v1.0.3
	github.com/decred/dcrd/chaincfg/chainhash v1.0.3
	github.com/decred/dcrd/chaincfg/v3 v3.1.0
	github.com/decred/dcrd/crypto/blake256 v1.0.0
	github.com/decred/dcrd/crypto/ripemd160 v1.0.1
	github.com/decred/dcrd/dcrec v1.0.0
	github.com/decred/dcrd/dcrec/edwards/v2 v2.0.2
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1
	github.com/decred/dcrd/wire v1.5.0
	github.com/decred/slog v1.2.0
)

replace (
	github.com/decred/dcrd/chaincfg/chainhash => ../chaincfg/chainhash
	github.com/decred/dcrd/chaincfg/v2 => ../chaincfg
	github.com/decred/dcrd/crypto/ripemd160 => ../crypto/ripemd160
	github.com/decred/dcrd/dcrec => ../dcrec
	github.com/decred/dcrd/dcrec/edwards/v2 => ../dcrec/edwards
	github.com/decred/dcrd/dcrec/secp256k1/v2 => ../dcrec/secp256k1
	github.com/decred/dcrd/dcrutil/v2 => ../dcrutil
	github.com/decred/dcrd/wire => ../wire
)
