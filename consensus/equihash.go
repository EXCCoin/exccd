package blockchain

import (
	"unsafe"
	"errors"

	"github.com/EXCCoin/exccd/chaincfg"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/wire"
)

func checkEquihashSolution(header *wire.BlockHeader, params *chaincfg.Params) (bool, error) {
	if header == nil {
		return false, errors.New("empty header")
	}
	if params == nil {
		return false, errors.New("empty chain params")
	}
	return false, nil
}

func checkProofOfWork(hash chainhash.Hash, nBits int, params *Params) bool {
	return false
}

type Equihash struct {
	N                    int
	K                    int
	IndicesPerHashOutput uint64
	HashOutput           uint64
	CollisionBitLength   uint64
	CollisionByteLength  uint64
	HashLength           uint64
	FullWidth            uint64
	FinalFullWidth       uint64
	TruncatedWidth       uint64
	FinalTruncatedWidth  uint64
	SolutionWidth        uint64
}

func (eh *Equihash) BasicSolve() (bool, error) {
	GenerateHashList()

	
	/**
	if n == 96 && k == 3 {
		return false, nil
	}
	if n == 200 && k == 9 {
		return false, EquihashBasicSolve_200_9()
	}
	if n == 96 && k == 5 {
		return false, nil
	}
	if n == 48 && k == 5 {
		return false, nil
	}
	*/
	initSize := 1 << (eh.CollisionBitLength + 1)
	lenIndices = unsafe.Sizeof(initSize)
	input := make([]FullStepRow, 0, initSize)
	tmpHash := make([]byte, eh.HashOutput)
	for (g := 0; len(input) < initSize; g++) {
		generateHash(state, g, tmpHash, eh.HashOutput)
		for (i := 0; i < eh.IndicesPerHashOutput && len(input) < initSize) {
			fullstepRow := NewFullStepRow()
			input = append(input, fullstepRow)
			if cancelled(ListGeneration) {
				return false, errors.New("solver cancelled")
			}
		}

	}
	return false, errors.New("unrecognised k and n params")
}

func EquihashSolutionSize(n, k int) int {
	return (1 << k) * (n/(k+1) + 1) / 8
}

type StepRow struct {
	hash []byte
	hashLen int
}

func GenerateHash(state *HashState, g int, hash *chainhash.Hash) {
	
}

func ExpandArray(in, out []byte, bitLen, bytePad int) error {
	if bitLen < 8 {
		return errors.New("bitLen must be >= 8")
	}
	if 8 * 32 < 7 + bitLen {
		return errors.New("")
	}
	outWidth := (bitLen+7)/8 + bytePad
	if len(out) != 8*outWidth*len(in)/bitLen {
		return errors.New("")
	}
	bitLenMask := (1 << bitLen) - 1

	accBits, accVal := 0, 0

	j := 0
	for (i := 0; i < len(in); i++) {
		accVal = (accVal << 8) | in[i]
		accBits += 8


	}
	return nil
}

func NewStepRow()
