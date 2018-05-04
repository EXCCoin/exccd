package equihash

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"log"
	"sort"

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
	person          = []byte("exccPoW")
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

func hasCollision(ha, hb []byte, i, l int) (bool, error) {
	if len(ha) != len(hb) {
		return false, errHashLen
	}
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

func distinctIndices(a, b []int) bool {
	for _, l := range a {
		for _, r := range b {
			if l == r {
				return false
			}
		}
	}
	return true
}

//TODO(jaupe) optimize function by not making a deep copy
func writeHashU32(h hash.Hash, v uint32) error {
	return writeHashBytes(h, writeU32(v))
}

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

func newDigest() (hash.Hash, error) {
	return blake2b.New(hashSize, nil)
}

// Equihash computes the hash digest
func equihash(b []byte) ([]byte, error) {
	digest, err := newDigest()
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

func hashLen(k, collLen int) int {
	return (k + 1) * int((collLen+7)/8)
}

type hashBuilder struct {
	digest string
}

func (hb *hashBuilder) copy() (hash.Hash, error) {
	h, err := blake2b.New(64, nil)
	if err != nil {
		return nil, err
	}
	err = writeHashStr(h, hb.digest)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (hb *hashBuilder) append(s string) {
	hb.digest += s
}

func (hb *hashBuilder) build() ([]byte, error) {
	return equihash([]byte(hb.digest))
}

func inSlice(b []byte, r, n int) []byte {
	i, j := r*n/8, (r+1)*n/8
	return b[i:j]
}

func hashDigest(h hash.Hash) []byte {
	return h.Sum(nil)
}

type solution struct {
	digest  []byte
	indices []int
}

func negIndex(s []solution, i int) int {
	return len(s) - i
}

func gbp(hb hashBuilder, n, k int) ([][]int, error) {
	collLen := collisionLen(n, k)
	hLen := hashLen(k, collLen)
	indicesPerHash := 512 / n

	x := []solution{}

	for i := 0; i < pow(collLen+1); i++ {
		r := i % indicesPerHash
		if r == 0 {
			hash, err := hb.copy()
			if err != nil {
				return nil, err
			}
			err = hashXi(hash, i/indicesPerHash)
			if err != nil {
				return nil, err
			}
			digest, err := expandArray(inSlice(hashDigest(hash), r, n), hLen, collLen, 0)
			sol := solution{
				digest:  digest,
				indices: []int{i},
			}
			x = append(x, sol)
		}
	}

	for i := 1; i < k; i++ {
		sortSolutions(x)
		xc := []solution{}
		for len(x) > 0 {
			j := 1

			for j < len(x) {
				ha := x[len(x)-1].digest
				hb := x[len(x)-1-j].digest
				coll, err := hasCollision(ha, hb, i, collLen)
				if err != nil {
					return nil, err
				}
				if !coll {
					break
				}
				j++
			}

			for l := 0; l < j-1; l++ {
				for m := l + 1; m < j; m++ {
					a, b := x[len(x)-l], x[len(x)-m]
					ai, bi := a.indices, b.indices
					if distinctIndices(ai, bi) {
						i, j := 0, 0
						if a.digest[0] < b.digest[0] {
							i, j = ai[0], bi[0]
						} else {
							i, j = bi[0], ai[0]
						}
						digest := xor(a.digest, b.digest)
						indices := []int{i, j}
						xc = append(xc, solution{digest, indices})
					}
				}
			}

			for j > 0 {
				x = x[:len(x)-1]
				j--
			}
		}
		x = xc
	}

	sortSolutions(x)
	solns := [][]int{}
	for len(x) > 0 {
		j := 1
		for j < len(x) {
			i := len(x) - 1
			ha, hb := x[i].digest, x[i-j].digest
			c1, err := hasCollision(ha, hb, k, collLen)
			if err != nil {
				return nil, err
			}
			c2, err := hasCollision(ha, hb, k+1, collLen)
			if err != nil {
				return nil, err
			}
			if !(c1 && c2) {
				break
			}
			j++
		}

		for l := 0; l < j-1; l++ {
			for m := l + 1; m < j; j++ {
				i, j := len(x)-l, len(x)-m
				res := xor(x[i].digest, x[j].digest)
				ii, ji := x[i].indices, x[j].indices
				if countZeros(res) == 8*hLen && distinctIndices(ii, ji) {
					if ii[0] < ji[0] {
						solns = append(solns, append(ii, ji...))
					} else {
						solns = append(solns, append(ji, ii...))
					}
				}
			}
		}

		for j > 0 {
			x = x[:len(x)-1] //pop last
			j--
		}
	}
	return solns, nil
}

type solutions []solution

func (s solutions) Len() int {
	return len(s)
}

func (s solutions) Less(i, j int) bool {
	x, y := s[i].digest, s[j].digest
	switch bytes.Compare(x, y) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (s solutions) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

type sortByteArrays [][]byte

func (b sortByteArrays) Len() int {
	return len(b)
}

func (b sortByteArrays) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortByteArrays) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

func sortSolutions(s []solution) []solution {
	sort.Sort(solutions(s))
	return s
}

func difficultyFilter(prevHash []byte, nonce int, soln []int, diff int) bool {
	h, err := blockHash(prevHash, nonce, soln)
	if err != nil {
		return false
	}
	count := countZeros(h)
	return count >= diff
}

func blockHash(prevHash []byte, nonce int, soln []int) ([]byte, error) {
	digest, err := newDigest()
	if err != nil {
		return nil, err
	}
	err = writeHashBytes(digest, prevHash)
	if err != nil {
		return nil, err
	}
	hashNonce(digest, nonce)
	return hashDigest(digest), nil
}

func hashNonce(h hash.Hash, nonce int) error {
	for i := 0; i < 8; i++ {
		err := writeHashU32(h, uint32(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func writeU32(v uint32) []byte {
	b := make([]byte, 4)
	putU32(b, v)
	return b
}

func putU32(b []byte, v uint32) {
	binary.LittleEndian.PutUint32(b, v)
}

func joinBytes(x, y []byte) []byte {
	b := make([]byte, len(x)+len(y))
	for i, v := range x {
		b[i] = v
	}
	for i, v := range y {
		b[len(x)+i] = v
	}
	return b
}

func exccPerson(n, k int) []byte {
	return joinBytes(person, writeParams(n, k))
}

func writeParams(n, k int) []byte {
	b := make([]byte, 8)
	putU32(b[:4], uint32(n))
	putU32(b[4:], uint32(k))
	return b
}

func hashXi(h hash.Hash, xi int) error {
	return writeHashU32(h, uint32(xi))
}
