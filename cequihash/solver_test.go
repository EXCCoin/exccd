// Copyright (c) 2026 The ExchangeCoin developers
// Use of this source code is governed by an ISC license.

package cequihash

import (
	"encoding/binary"
	"slices"
	"testing"
	"unsafe"
)

func TestSupportedProofParameters(t *testing.T) {
	for _, tc := range []struct{ n, k, size int }{{48, 5, 36}, {96, 5, 68}, {144, 5, 100}, {200, 9, 1344}} {
		if size := EquihashSolutionSize(tc.n, tc.k); size != tc.size {
			t.Fatalf("(%d,%d) size %d, want %d", tc.n, tc.k, size, tc.size)
		}
		indices := make([]uint32, 1<<uint(tc.k))
		for i := range indices {
			indices[i] = uint32(i)
		}
		packed := compressIndices(tc.n, tc.k, 0, nil, indices)
		if len(packed) != tc.size || !slices.Equal(expandArray(tc.n, tc.k, unsafe.Pointer(&packed[0])), indices) {
			t.Fatalf("(%d,%d) proof round trip", tc.n, tc.k)
		}
		if ValidateEquihash(tc.n, tc.k, nil, packed) {
			t.Fatal("accepted empty input")
		}
	}
	for _, pair := range [][2]int{{0, 0}, {48, 9}, {200, 5}, {48, -1}} {
		if EquihashSolutionSize(pair[0], pair[1]) != 0 {
			t.Fatalf("supported unexpected parameters %v", pair)
		}
	}
}

type foundSolution []byte

func (s *foundSolution) Validate(ptr unsafe.Pointer) int {
	if ptr == nil {
		return 0
	}
	*s = ExtractSolution(48, 5, ptr)
	return 1
}

func TestSolveHeaderNonce(t *testing.T) {
	for _, size := range []int{144, 180} {
		header := make([]byte, size)
		var proof foundSolution
		for nonce := uint32(1); nonce <= 32; nonce++ {
			SolveEquihash(48, 5, header, nonce, 1, &proof)
			if len(proof) > 0 {
				binary.LittleEndian.PutUint32(header[140:], nonce)
				if !ValidateEquihash(48, 5, header, proof) {
					t.Fatalf("invalid proof for %d-byte header", size)
				}
				break
			}
		}
		if len(proof) == 0 {
			t.Fatalf("no proof for %d-byte header", size)
		}
	}
}
