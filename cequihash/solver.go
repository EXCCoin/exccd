package cequihash

/*
#cgo CFLAGS: -O3 -I./implementation -I./implementation/blake2 -march=native -std=c99
#include "cequihash.h"
 */
import "C"
import (
	"errors"
	"unsafe"
	cptr "github.com/mattn/go-pointer"

)

type EquihashCallback interface {
	Validate(pointer unsafe.Pointer) int
}

//export equihash_proxy
func equihash_proxy(callback_data unsafe.Pointer, extra_data unsafe.Pointer) C.int {
	callback := cptr.Restore(callback_data).(EquihashCallback)

	return C.int(callback.Validate(extra_data))
}

func SolveEquihash(n, k int, input []byte, callback EquihashCallback) error {
	if len(input) < 140 {
		return errors.New("Too short input for the equihash")
	}

	callbackptr := cptr.Save(callback)

	C.EquihashSolve(unsafe.Pointer(C.CBytes(input)), callbackptr, C.int(n), C.int(k))

	return nil
}

func ValidateEquihash(n, k int, input []byte, solution []byte) bool {
	return C.EquihashValidate(C.int(n), C.int(k), unsafe.Pointer(C.CBytes(input)), unsafe.Pointer(C.CBytes(solution))) != 0
}