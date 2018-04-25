package consensus

import "github.com/EXCCoin/exccd/chaincfg/chainhash"

type Params struct {
	GenesisHash      chainhash.Hash
	PowAverageWindow int64
	PowMaxUpAdjust   int64
	PowMaxDownAdjust int64
}
