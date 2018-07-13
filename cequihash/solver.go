// Copyright (c) 2018 The ExchangeCoin team
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cequihash

/*
#cgo CXXFLAGS: -O3 -march=native -std=c++17 -Wall -Wno-deprecated-declarations -D_POSIX_C_SOURCE=200112L -Werror
#include "cequihash.h"
*/
import "C"
import (
	cptr "github.com/mattn/go-pointer"
	"unsafe"
)

// This code is a wrapper for C implementation of equihash.
// The C part relies on the callback which validates found solution for some additional criteria.
// The callback (from the point of view of C code) receives two parameters:
// -- pointer to some (unknown to C code) structure. No attempts to access this structure or check its content
//    are performed by C code. The pointer is just passed back and forth between C and Go.
// -- pointer to the solution. Note that since we're dealing with Equihash algorithm there is no need to pass
//    any information about solution size - it can be calculated on the fly from Equihash N and K parameters.
// The C code uses equihashProxy() Go function as an entry point for callback. This function converts pointer to
// data structure into Go EquihashCallback interface and calls Validate method with solution as a parameter.
//
// All of the above describes "business" call path, when actual solution is passed. There is additional call path
// which is necessary to prematurely finish C Equihash solver. This is done as follows:
// From time to time C code invokes callback with NULL solution. The purpose of this call is to check additional
// exit conditions. If callback returns non-zero value for such an invocation, then C Equihash solver exits ASAP.
//
// The call path at the Go side is identical for both of the above cases, so Validate method must be prepared to
// receive 0 as a solution. Corresponding check may look like this:
//
// ...Validate(solution unsafe.Pointer) int {
//    if uintptr(solution) == 0 {
//        if someExternalCondition {
//            return 1 // stop the solver
//        }
//    }
// ...
// }

type EquihashCallback interface {
	Validate(pointer unsafe.Pointer) int
}

//export equihashProxy
func equihashProxy(callback_data unsafe.Pointer, extra_data unsafe.Pointer) C.int {
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

func SolveEquihash(n, k int, input []byte, nonce int64, callback EquihashCallback) {
	callbackptr := cptr.Save(&callback)
	defer cptr.Unref(callbackptr)

	C.EquihashSolve(C.int(n), C.int(k), unsafe.Pointer(&input[0]), C.int(len(input)), C.int64_t(nonce), callbackptr)
}

func ValidateEquihash(n, k int, input []byte, nonce int64, solution []byte) bool {
	equihashSolutionSize := EquihashSolutionSize(n, k)

	if len(solution) < equihashSolutionSize {
		return false
	}

	return C.EquihashValidate(C.int(n), C.int(k), unsafe.Pointer(&input[0]), C.int(len(input)), C.int64_t(nonce),
		unsafe.Pointer(&solution[0])) == 0
}

func EquihashSolutionSize(n, k int) int {
	return 1 << uint32(k) * (n/(k+1) + 1) / 8
}
