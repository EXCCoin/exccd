package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"reflect"
	"sort"

	blake2b "github.com/minio/blake2b-simd"
)

const (
	wordSize = 32
	wordMask = (1 << wordSize) - 1
	byteMask = 0xFF
)

var (
	errBadArg   = errors.New("invalid argument")
	errWriteLen = errors.New("didn't write full len")
)

func bytesCmp(x, y []byte) bool {
	return bytes.Compare(x, y) < 0
}

type hashKeys []hashKey

func (k hashKeys) Len() int {
	return len(k)
}

func (k hashKeys) Less(i, j int) bool {
	return bytesCmp(k[i].hash, k[j].hash)
}

func (k hashKeys) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func hashXi(h hash.Hash, xi int) error {
	return writeUint32ToHash(h, uint32(xi))
}

func hashDigest(h hash.Hash) []byte {
	return h.Sum(nil)
}

func expandArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 && wordSize < 7+bitLen {
		return nil, errBadArg
	}
	outWidth := (bitLen+7)/8 + bytePad
	if outLen != 8*outWidth*len(in)/bitLen {
		return nil, errBadArg
	}

	out, bitLenMask := make([]byte, outLen), (1<<uint(bitLen))-1
	accBits, accValue, j := 0, 0, 0
	for _, val := range in {
		accValue = (accValue<<8)&wordMask | int(val&0xFF)
		accBits += 8

		if accBits >= bitLen {
			accBits -= bitLen
			for x := bytePad; x < outWidth; x++ {
				a := accValue >> uint(accBits+8*(outWidth-x-1))
				b := (bitLenMask >> uint(8*(outWidth-x-1))) & byteMask
				out[j+x] = byte(a & b)
			}
			j += outWidth
		}
	}

	return out, nil
}

type equihashSolution []int

type hashKey struct {
	hash    []byte
	indices []int
}

func minLen(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func xor(a, b []byte) []byte {
	n := minLen(len(a), len(b))
	x := make([]byte, n)
	for i := 0; i < n; i++ {
		x[i] = a[i] ^ b[i]
	}
	return x
}

func hasCollision(x, y []byte, i, length int) bool {
	start, end := (i-1)*length/8, i*length/8
	for j := start; j < end; j++ {
		if x[j] == y[j] {
			return true
		}
	}
	return false
}

func collisionOffset(tuples []hashKey, i, collisionLen int) int {
	n := len(tuples)
	last := tuples[n-1]
	ha := last.hash
	for j := 1; j < n; j++ {
		hb := tuples[n-1-j].hash
		if !hasCollision(ha, hb, i, collisionLen) {
			return j
		}
	}
	return n
}

func distinctIndices(a, b []int) bool {
	for _, av := range a {
		for _, bv := range b {
			if av == bv {
				return false
			}
		}
	}
	return true
}

func concatIndices(x, y []int) []int {
	if x[0] < y[0] {
		return append(x, y...)
	}
	return append(y, x...)
}

func gbp(digest hash.Hash, n, k int) ([]equihashSolution, error) {
	collisionLength := n / (k + 1)
	hashLength := (k + 1) * (collisionLength + 7) / 8
	indicesPeHashOutput := 512 / n

	//  1) Generate list (X)
	X := []hashKey{}
	var tmpHash []byte
	for i := 0; i < pow(collisionLength+1); i++ {
		r := i % indicesPeHashOutput
		if r == 0 {
			currDigest := copyHash(digest)
			err := hashXi(currDigest, i/indicesPeHashOutput)
			if err != nil {
				return nil, err
			}
			tmpHash = currDigest.Sum(nil)
		}
		d := tmpHash[r*n/8 : (r+1)*n/8]
		expanded, err := expandArray(d, hashLength, collisionLength, 0)
		if err != nil {
			return nil, err
		}
		X = append(X, hashKey{expanded, []int{i}})
	}

	// 3) Repeat step 2 until 2n/(k+1) bits remain
	for i := 1; i < k; i++ {
		// sort tuples by hash
		sort.Sort(hashKeys(X))

		xc := []hashKey{}
		for len(X) > 0 {
			// 2b) Find next set of unordered pairs with collisions on first n/(k+1) bits
			xSize := len(X)
			j := collisionOffset(X, i, collisionLength)

			//2c) Store tuples (X_i ^ X_j, (i, j)) on the table
			for l := 0; l < j-1; l++ {
				for m := l + 1; m < l; m++ {
					x1l, x1m := X[xSize-1-l], X[xSize-1-m]
					if distinctIndices(x1l.indices, x1m.indices) {
						concat := concatIndices(x1l.indices, x1m.indices)
						a, b := x1l.hash, x1m.hash
						xc = append(xc, hashKey{xor(a, b), concat})
					}
				}
			}
			// 2d) drop this set
			X = X[:len(X)-j]
		}
		// 2e) replace previous list with new list
		X = xc
	}

	sort.Sort(hashKeys(X))
	//find solutions
	solns := []equihashSolution{}
	for len(X) > 0 {
		xn := len(X)
		j := solutionOffset(X, k, collisionLength)
		for l := 0; l < j-1; l++ {
			for m := l + 1; m < j; m++ {
				a, b := X[xn-1-l], X[xn-1-m]
				res := xor(a.hash, b.hash)
				f1 := countZeros(res) == 8*hashLength
				f2 := distinctIndices(a.indices, b.indices)
				if f1 && f2 {
					indices := concatIndices(a.indices, b.indices)
					solns = append(solns, indices)
				}
			}
		}
		X = X[:len(X)-j]
	}

	return solns, nil
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

func solutionOffset(x []hashKey, k, collisionLen int) int {
	xSize := len(x)
	for j := 1; j < xSize; j++ {
		lc := hasCollision(x[xSize-1].hash, x[xSize-1-j].hash, k, collisionLen)
		rc := hasCollision(x[xSize-1].hash, x[xSize-1-j].hash, k+1, collisionLen)
		if !(lc && rc) {
			return j
		}
	}
	return xSize
}

// pow returns pow of base 2 for only positive k
func pow(k int) int {
	if k < 1 {
		return 1
	}
	return 1 << uint(k)
}

func hasDuplicateIndices(indices []int) bool {
	if len(indices) <= 1 {
		return false
	}
	set := make(map[int]bool)
	for _, index := range indices {
		if set[index] {
			return true
		}
		set[index] = true
	}
	return false
}

func writeBytesToHash(h hash.Hash, b []byte) error {
	n, err := h.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errWriteLen
	}
	return nil
}

func writeUint32ToHash(h hash.Hash, v uint32) error {
	return writeBytesToHash(h, writeUint32(v))
}

func writeUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func copyHash(src hash.Hash) hash.Hash {
	typ := reflect.TypeOf(src)
	val := reflect.ValueOf(src)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	elem := reflect.New(typ).Elem()
	elem.Set(val)
	return elem.Addr().Interface().(hash.Hash)
}

func generateWord(n int, h hash.Hash, idx int) (int, error) {
	bytesPerWord, wordsPerHash := n/8, 512/n
	hidx, hrem := idx/wordsPerHash, idx%wordsPerHash
	indexBytes := writeUint32(uint32(hidx))
	ctx := copyHash(h)
	err := writeBytesToHash(ctx, indexBytes)
	if err != nil {
		return 0, err
	}
	digest := ctx.Sum(nil)
	// fold word
	word := 0
	for i := hrem * bytesPerWord; i < hrem*bytesPerWord+bytesPerWord; i++ {
		word = word<<8 | (int(digest[i]) & 0xFF)
	}
	return word, nil
}

func generateWords(n, solutionLen int, indices []int, h hash.Hash) ([]int, error) {
	words := []int{}
	for i := 0; i < solutionLen; i++ {
		word, err := generateWord(n, h, indices[i])
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

func ValidateSolution(n, k int, person, header []byte, solutionIndices []int) (bool, error) {
	if n < 2 {
		return false, errBadArg
	}
	if !(k >= 3) {
		return false, errBadArg
	}
	if n%8 != 0 {
		return false, errBadArg
	}
	if n%(k+1) != 0 {
		return false, errBadArg
	}

	solutionLen := pow(k)
	if len(solutionIndices) != solutionLen {
		return false, errBadArg
	}

	if hasDuplicateIndices(solutionIndices) {
		return false, nil
	}

	bytesPerWord := n / 8
	wordsPerHash := 512 / n
	outLen := wordsPerHash * bytesPerWord

	// create hash digest and words
	c := blake2b.Config()
	digest, err := blake2b.New(c)
	if err != nil {
		return false, err
	}
	err = writeBytesToHash(digest, header)
	if err != nil {
		return false, err
	}

	// check pair-wise ordering of solution indices
	for s := 0; s < k; s++ {
		d := 1 << uint(s)
		for i := 0; i < solutionLen; i += 2 * d {
			if solutionIndices[i] >= solutionIndices[i+d] {
				return false, nil
			}
		}
	}

	words, err := generateWords(n, solutionLen, solutionIndices, digest)
	if err != nil {
		return false, err
	}

	// check XOR conditions
	bitsPerStage := n / (k + 1)
	for s := 0; s < k; s++ {
		d := 1 << uint(s)
		for i := 0; i < solutionLen; i += 2 * d {
			w := words[i] ^ words[i+d]
			if (w >> uint(n-(s+1)*bitsPerStage)) != 0 {
				return false, nil
			}
			words[i] = w
		}
	}
	return words[0] == 0, nil
}
