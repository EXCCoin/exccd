package cequihash

/*
#cgo CFLAGS: -O3 -I./implementation -march=native -std=c99

#include <stdlib.h>
void* ExpandArray(int, int, void*);
 */
import "C"
import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

func expandArray(n, k int, solution []byte) []int {
	ptr := C.ExpandArray(C.int(n), C.int(k), C.CBytes(solution))
	defer C.free(ptr)

	buf := bytes.NewBuffer(solution)
	var tmp [4]byte
	var result []int

	num_read, _ := buf.Read(tmp[:])

	for num_read > 0 {
		index := binary.LittleEndian.Uint32(tmp[:])
		result = append(result, int(index))
		num_read, _ = buf.Read(tmp[:])
	}

	return result
}

type SolutionAppenderData struct {
	EquihashCallback
	n int
	k int
	fullSolution [][]int
}

func (data SolutionAppenderData)solutionAppender(solution unsafe.Pointer) int {
	equihashSolutionSize := (1 << uint(data.k)) * (data.n / (data.k + 1) + 1) / 8;
	solutionBytes := C.GoBytes(solution, C.int(equihashSolutionSize))
	data.fullSolution = append(data.fullSolution, expandArray(data.n, data.k, solutionBytes))
	return 0
}
