// Copyright (c) 2018 The ExchangeCoin team
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cequihash

/*
#cgo amd64 CXXFLAGS: -march=x86-64 -mtune=generic
#cgo ppc64le CXXFLAGS: -mtune=powerpc64le -DNO_WARN_X86_INTRINSICS
#cgo CXXFLAGS: -O3 -std=c++17 -Wall -Wno-strict-aliasing -Wno-shift-count-overflow -Werror
#include "cequihash.h"
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

func expandArray(n, k int, solution unsafe.Pointer) []uint32 {
	ptr := C.IndicesFromSolution(C.int(n), C.int(k), solution)
	defer C.free(ptr)

	indexCount := 1 << uint32(k)

	buf := bytes.NewBuffer(C.GoBytes(ptr, C.int(indexCount*4)))
	var tmp [4]byte
	var result []uint32

	num_read, _ := buf.Read(tmp[:])

	for num_read > 0 {
		index := binary.LittleEndian.Uint32(tmp[:])
		result = append(result, index)
		num_read, _ = buf.Read(tmp[:])
	}

	return result
}

type SolutionHolder struct {
	fullSolution [][]uint32
}

type SolutionAppenderData struct {
	n        int
	k        int
	solution *SolutionHolder
}

func (data SolutionAppenderData) Validate(solution unsafe.Pointer) int {
	if uintptr(solution) == 0 {
		return 0
	}

	solutionIndexes := expandArray(data.n, data.k, solution)
	data.solution.fullSolution = append(data.solution.fullSolution, solutionIndexes)
	return 0
}

func compressIndices(n, k int, nonce uint32, input []byte, solutionIndices []uint32) []byte {
	ptr := C.SolutionFromIndices(C.int(n), C.int(k), unsafe.Pointer(&solutionIndices[0]), C.uint32_t(len(solutionIndices)))
	defer C.free(ptr)

	cBitLen := n / (k + 1)
	lenIndices := len(solutionIndices) * int(unsafe.Sizeof(solutionIndices[0]))
	minLen := (cBitLen + 1) * lenIndices / int(8*unsafe.Sizeof(solutionIndices[0]))

	result := C.GoBytes(ptr, C.int(minLen))

	return result
}
