package equihash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
)

var (
	errLen      = errors.New("slices not same len")
	errBitLen   = errors.New("bit len < 8")
	errOutWidth = errors.New("incorrect outwidth size")
)

const (
	wordSize = 32
	wordMask = (1 << wordSize) - 1
)

func xor(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, errLen
	}
	out := make([]byte, len(a))
	for i, val := range a {
		out[i] = val ^ b[i]
	}
	return out, nil
}

func compressArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errBitLen
	}
	if wordSize < 7+bitLen {
		return nil, errBitLen
	}
	inWidth := (bitLen+7)/8 + bytePad
	if outLen != bitLen*len(in)/(8*inWidth) {
		return nil, errLen
	}
	out := make([]byte, outLen)
	bitLenMask := (1 << uint(bitLen)) - 1
	accBits, accVal, j := 0, 0, 0

	for i := 0; i < outLen; i++ {
		if accBits < 8 {
			accVal = ((accVal << uint(bitLen)) & wordMask) | int(in[j])
			for x := bytePad; x < inWidth; x++ {
				b := int(in[j+x]) & ((bitLenMask >> uint(8*(inWidth-x-1))) & 0xFF)
				accVal = accVal | b
			}
			j += inWidth
			accBits += bitLen
		}
		accBits -= 8
		out[i] = byte((accVal >> uint(accBits)) & 0xFF)
	}

	return out, nil
}

func expandArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errBitLen
	}
	if wordSize < 7+bitLen {
		return nil, errBitLen
	}
	outWidth := (bitLen+7)/8 + bytePad
	if outLen != 8*outWidth*len(in)/bitLen {
		return nil, errOutWidth
	}
	bitLenMask := (1 << uint(bitLen)) - 1
	accBits, accVal, j := 0, uint(0), 0
	out := make([]byte, outLen)
	for i := 0; i < len(in); i++ {
		accVal = ((accVal << 8) & wordMask) | uint(in[i])
		accBits += 8
		fmt.Printf("i = %v\nin[%v] = %v\naccVal = %v\n accBits = %v\n", i, i, in[i], accVal, accBits)

		if accBits >= bitLen {
			accBits -= bitLen
			for x := bytePad; x < outWidth; x++ {
				a := accVal >> (uint(accBits + (8 * (outWidth - x - 1))))
				b := (bitLenMask >> uint((8 * (outWidth - x - 1)))) & 0xFF
				v := byte(a) & byte(b)
				fmt.Printf("a = %v\nb = %v\nv = %v\n", a, b, v)
				out[j+x] = v
			}
			j += outWidth
		}
	}

	return out, nil
}

func binPowInt(k int) int {
	if k < 1 {
		return 1
	}
	val := 2
	for i := 0; i < k; i++ {
		val *= 2
	}
	return val
}

func hashXi(digest hash.Hash, xi int) (hash.Hash, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, xi)
	n, err := digest.Write(b)
	if err != nil {
		return nil, err
	}
	return digest, nil
}

func distinctIndices(a, b []byte) bool {
	for _, l := range a {
		for _, r := range b {
			if l == r {
				return false
			}
		}
	}
	return true
}

func newBlakeHash() (hash.Hash, error) {
	return nil, errors.New("nyi")
}

func EquihashSolver(digest hash.Hash, n, k int) ([][]byte, error) {
	collisionLen := n / (k + 1)
	hashLen := (k + 1) * int((collisionLen+7)/8)
	indicesPerHashOutput := 512 / n
	x := [][]byte{}

	spaceLen := binPowInt(collisionLen + 1)
	// 1. generate first list
	for i := 0; i < spaceLen; i++ {
		r := i % indicesPerHashOutput
		if r == 0 {

			tmpHash := digest.Copy()
			hashXi(currDigest, i/indicesPerHashOutput)
			tmpHash := digest.digest()
			slice := []byte(tmpHash[r*n/8 : (r+1)*n/8])
			x = append(x, expandArray(slice, hashLen, collisionLen))
		}
	}

	for i := 1; i < k; i++ {
		sortX(x)
		// finding collisions
		xc = []byte{}
		for len(X) > 0 {
			j := 1
			for j < len(X) {
				if !hasCollision(X[len(X)-1][0], X[len(X)-1-j][0], i, collisionLen) {
					break
				}
				j++
			}

			for l := 0; l < j-1; l++ {
				for m := l + 1; m < j; m++ {
					if distinctindices(X[len(X)-1-l][1], X[len(X)-1-m][1]) {

					}
				}
			}
		}
	}
	return nil, errors.New("nyi")
}

func compressArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errBitLen
	}
	if wordSize < 7+bitLen {
		return nil, errBitLen
	}

	inWidth := (bitLen+7)/8 + bytePad
	if outLen != 8*inWidth*len(in)/bitLen {
		return nil, errOutWidth
	}
	bitLenMask := (1 << uint(bitLen)) - 1
	accBits, accVal, j := 0, 0, 0
	out := make([]byte, outLen)
	for i := 0; i < outLen; i++ {
		accBits -= bitLen
		if accBits < 8 {
			accVal = accVal << uint(bitLen)
			for x := bytePad; x < inWidth; x++ {
				a := in[j+x]
				b := (bitLenMask >> uint(8*(inWidth-x-1))) & 0xFF
				v := a & byte(b)
				out[j+x] = v
			}
			j += inWidth
		}

		accBits -= 8
		out[i] = byte((accVal >> uint(accBits)) & 0xFF)
	}
	return out, nil
}
