package blockchain

import (
	"testing"

	"github.com/EXCCoin/exccd/wire"
)

func testBlockHeader() *wire.BlockHeader {
	return nil
}

func TestCheckEquihashSolution_EmptyHeader(t *testing.T) {
	good, err := checkEquihashSolution(nil, nil)
	if good {
		t.Error("empty header shouldn't be good solution")
	}
	if err == nil {
		t.Error("expected error for empty header")
	}
}

func TestCheckEquihashSolution_EmptyChainParams(t *testing.T) {
	good, err := checkEquihashSolution(nil, nil)
	if good {
		t.Error("empty header shouldn't be good solution")
	}
	if err == nil {
		t.Error("expected error for empty header")
	}
}

func TestEquihashBasicSolve_0_0(t *testing.T) {
	solved, err := EquihashBasicSolve(0, 0)
	if err == nil {
		t.Error("expector error")
	}
	if solved {
		t.Error("should not be solved")
	}
}

func TestEquihashBasicSolve_0_0(t *testing.T) {
	solved, err := EquihashBasicSolve(0, 0)
	if err == nil {
		t.Error("expector error")
	}
	if solved {
		t.Error("should not be solved")
	}
}

func TestEquihashBasicSolve_96_