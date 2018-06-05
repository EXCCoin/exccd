// Copyright (c) 2018 The ExchangeCoin team
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cequihash

/*
#cgo CFLAGS: -O3 -I./implementation -I./implementation/blake2 -march=native -std=c99
#include <stdint.h>
#include "cequihash.h"
 */
import "C"
import (
	"unsafe"
	cptr "github.com/mattn/go-pointer"

)

type EquihashCallback interface {
	Validate(pointer unsafe.Pointer) int
}

//export equihash_proxy
func equihash_proxy(callback_data unsafe.Pointer, extra_data unsafe.Pointer) C.int {
	callback := cptr.Restore(callback_data).(*EquihashCallback)

	if callback == nil {
		return C.int(0)
	}

	return C.int((*callback).Validate(extra_data))
}

func ExtractSolution(n, k int, solptr unsafe.Pointer) []byte {
	size := EquihashSolutionSize(n, k)

	return C.GoBytes(solptr, C.int(size))
}

func SolveEquihash(n, k int, input []byte, nonce int64, callback EquihashCallback) error {
	callbackptr := cptr.Save(&callback)
	defer cptr.Unref(callbackptr)

	C.EquihashSolve(unsafe.Pointer(&input[0]), C.int(len(input)), C.int64_t(nonce), callbackptr, C.int(n), C.int(k))

	return nil
}

func ValidateEquihash(n, k int, input []byte, nonce int64, solution []byte) bool {
	equihashSolutionSize := EquihashSolutionSize(n, k)

	if len(solution) != equihashSolutionSize {
		return false
	}

	return C.EquihashValidate(C.int(n), C.int(k), unsafe.Pointer(&input[0]), C.int(len(input)), C.int64_t(nonce),
		unsafe.Pointer(&solution[0])) != 0
}

func EquihashSolutionSize(n, k int) int {
	return 1 << uint32(k) * (n/(k+1) + 1) / 8
}
