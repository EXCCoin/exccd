package cequihash

/*
#cgo CFLAGS: -O3 -I./implementation -I./implementation/blake2 -march=native -std=c99
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

func SolveEquihash(n, k int, input []byte, nonce int, callback EquihashCallback) error {
	callbackptr := cptr.Save(&callback)
	defer cptr.Unref(callbackptr)

	C.EquihashSolve(unsafe.Pointer(C.CBytes(input)), C.int(len(input)), C.int(nonce), callbackptr, C.int(n), C.int(k))

	return nil
}

func ValidateEquihash(n, k int, input []byte, solution []byte) bool {
	pow2K := 1 << uint32(k)
	equihashSolutionSize := pow2K * (n / (k + 1) + 1) / 8

	if len(solution) != equihashSolutionSize {
		return false
	}

	return C.EquihashValidate(C.int(n), C.int(k), unsafe.Pointer(C.CBytes(input)), C.int(len(input)),
		unsafe.Pointer(C.CBytes(solution))) != 0
}
