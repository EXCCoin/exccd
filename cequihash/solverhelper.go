// Copyright (c) 2018 The ExchangeCoin team
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package cequihash

/*
#cgo CFLAGS: -O3 -march=native -std=c99

#include "cequihash.h"
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

func expandArray(n, k int, solution unsafe.Pointer) []uint32 {
	ptr := C.GetIndices(C.int(n), C.int(k), solution)
	defer C.free(ptr)

	indexCount := 1 << uint32(k)

	buf := bytes.NewBuffer(C.GoBytes(ptr, C.int(indexCount*4)))
	var tmp [4]byte
	var result []uint32

	num_read, _ := buf.Read(tmp[:])

	for num_read > 0 {
		index := binary.BigEndian.Uint32(tmp[:])
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
	ptr := C.PutIndices(C.int(n), C.int(k), unsafe.Pointer(&input[0]), C.int(len(input)), C.uint32_t(nonce),
		unsafe.Pointer(&solutionIndices[0]), C.int(len(solutionIndices)))
	defer C.free(ptr)

	cBitLen := n / (k + 1)
	lenIndices := len(solutionIndices) * int(unsafe.Sizeof(solutionIndices[0]))
	minLen := (cBitLen + 1) * lenIndices / int(8*unsafe.Sizeof(solutionIndices[0]))

	result := C.GoBytes(ptr, C.int(minLen))

	return result
}
