package equihash

import (
	"errors"
	"hash"
	"strconv"

	"golang.org/x/crypto/blake2b"
)

var (
	errLen          = errors.New("slices not same len")
	errBitLen       = errors.New("bit len < 8")
	errOutWidth     = errors.New("incorrect outwidth size")
	errKLarge       = errors.New("k should be less than n")
	errCollisionLen = errors.New("collision length too big")
	errHashLen      = errors.New("hash len is too small")
	errHashStartPos = errors.New("hash len < start pos")
	errHashEndPos   = errors.New("hash len < end pos")
	errWriteLen     = errors.New("didn't write full len")
	emptySlice      = []byte{}
)

const (
	wordSize = 32
	wordMask = (1 << wordSize) - 1
	hashSize = 64
)

func validateParams(n, k int) error {
	if k >= n {
		return errKLarge
	}
	if collisionLen(n, k)+1 >= 32 {
		return errCollisionLen
	}
	return nil
}

func collisionLen(n, k int) int {
	return n / (k + 1)
}

// countZeros counts leading zero bits in byte array
func countZeros(h []byte) int {
	for i, val := range h {
		for j := 0; j < 8; j++ {
			mask := 1 << uint(7-j)
			if (int(val) & mask) > 0 {
				return (i * 8) + j
			}
		}
	}
	return len(h) * 8
}

// minSlices returns the slices sorted by their length
// the first returned has the smallest length and the
// and the second is the highest of the two
func minSlices(a, b []byte) ([]byte, []byte) {
	if len(a) <= len(b) {
		return a, b
	}
	return b, a
}

// xor runs xor piece-wise against 2 slices
// returns empty slice if slices are not same len
func xor(a, b []byte) []byte {
	if len(a) == 0 && len(b) == 0 {
		return emptySlice
	}
	if len(a) == 0 {
		return emptySlice
	}
	if len(b) == 0 {
		return emptySlice
	}
	out := make([]byte, len(a))
	for i, val := range a {
		out[i] = val ^ b[i]
	}
	return out
}

func hasCollision(ha, hb []byte, i int) (bool, error) {
	if len(ha) != len(hb) {
		return false, errHashLen
	}
	l := len(ha)
	start, end := (i-1)*l/8, i*l/8
	if len(ha) < start || len(hb) < start {
		return false, errHashStartPos
	}
	if len(hb) < end || len(hb) < end {
		return false, errHashEndPos
	}
	gate := true
	for j := start; j < end; j++ {
		gate = gate && (ha[j] == hb[j])
	}
	return gate, nil
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
				g := int(in[j+x]) & ((bitLenMask >> uint(8*(inWidth-x-1))) & 0xFF)
				g = g << uint(8*(inWidth-x-1))
				accVal = accVal | g
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

		if accBits >= bitLen {
			accBits -= bitLen
			for x := bytePad; x < outWidth; x++ {
				a := accVal >> (uint(accBits + (8 * (outWidth - x - 1))))
				b := (bitLenMask >> uint((8 * (outWidth - x - 1)))) & 0xFF
				v := byte(a) & byte(b)
				out[j+x] = v
			}
			j += outWidth
		}
	}

	return out, nil
}

// binPowInt returns pow of base 2 for only positive k
func pow(k int) int {
	if k < 1 {
		return 1
	}
	return 1 << uint(k)
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

//TODO(jaupe) optimize function by not making a byte array copy
func writeHashStr(h hash.Hash, s string) error {
	return writeHashBytes(h, []byte(s))
}

func writeHashBytes(h hash.Hash, b []byte) error {
	n, err := h.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errWriteLen
	}
	return nil
}

// Equihash computes the hash digest
func equihash(b []byte) ([]byte, error) {
	h, err := blake2b.New(64, nil)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func hashNonce(h hash.Hash, nonce int) error {
	for i := 0; i < 8; i++ {
		k := nonce >> uint(32*i)
		s := "<I" + strconv.Itoa(k)
		err := writeHashStr(h, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func hashXi(h hash.Hash, xi string) error {
	err := writeHashStr(h, "<I"+xi)
	if err != nil {
		return err
	}
	return nil
}

func hashLen(k, collLen int) int {
	return (k + 1) * int((collLen+7)/8)
}

type hashBuilder struct {
	digest string
}

func (hb *hashBuilder) copy() hashBuilder {
	return hashBuilder{string(hb.digest)}
}

func (hb *hashBuilder) append(s string) {
	hb.digest += s
}

func (hb *hashBuilder) build() ([]byte, error) {
	return equihash([]byte(hb.digest))
}

/*
func gbp(h hash.Hash, n, k int) error {
	collLen := collisionLen(n, k)
	hLen := hashLen(k, collLen)
	indicesPerHash := 512 / n

	input := [][]string{}
	tmpHash := ""

	for i := 0; i < pow(collLen+1); i++ {
		r := i % indicesPerHash
		if r == 0 {

		}
	}

	return nil
}
*/

/*
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
*/
