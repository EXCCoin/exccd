package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/minio/blake2b-simd"
	"hash"
	"math/big"
	"reflect"
	"sort"
)

const (
	// size of word
	wordSize = 32
	wordMask = (1 << wordSize) - 1
	byteMask = 0xFF
	// N is the number of hash digests used to find a mining solution
	N = 200
	// K is the exponent for xor'ing 2^k hash digests for solution
	K             = 9
	defaultPrefix = "ZcashPoW"

	CollisionBitLength = N / (K + 1)
	SolutionWidth      = (1 << K) * (CollisionBitLength + 1) / 8
)

var (
	errBadArg           = errors.New("invalid argument")
	errWriteLen         = errors.New("didn't write full len")
	errKLarge           = errors.New("k should be less than n")
	errCollisionLen     = errors.New("collision length too big")
	errSmallBitLen      = errors.New("bitLen < 8")
	errSmallWordSize    = errors.New("wordSize < 7+bitLen")
	errBadOutLen        = errors.New("outLen != 8*outWidth*len(in)/bitLen")
	errDuplicateIndices = errors.New("duplicate indices")
	errPairWiseOrdering = errors.New("bad pair-wise ordering")
	errBadWord          = errors.New("bad word")
	errNullHash         = errors.New("empty hash")
	errEmptyKeys        = errors.New("empty keys")
	errEmptyIndices     = errors.New("empty indices")
	bigZero             = big.NewInt(0)
)

// the generic person prefix encoder; which encodes the prefix and gbp parameters
func person(n, k int) []byte {
	nb, kb := writeU32(uint32(n)), writeU32(uint32(k))
	return append([]byte(defaultPrefix), append(nb, kb...)...)
}

// hashKeys represents a slice of hashKeys; used for creating a type for sorting hash keys
type hashKeys []hashKey

// returns the length of hash keys for sorting
func (k hashKeys) Len() int {
	return len(k)
}

// returns true if the ith hash is less than the jth hash
func (k hashKeys) Less(i, j int) bool {
	return bytes.Compare(k[i].hash, k[j].hash) < 0
}

// swaps the hash keys at the ith and jth position in the slice
func (k hashKeys) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

// encodes the solution position to the hash
func hashXi(h hash.Hash, xi int) error {
	if h == nil {
		return errNullHash
	}
	return writeU32ToHash(h, uint32(xi))
}

// creates and returns the hash digest
func hashDigest(h hash.Hash) []byte {
	return h.Sum(nil)
}

// newHash creates a blake2b hash using the equihash params (n, k) and the person prefix
func newHash(n, k int) (hash.Hash, error) {
	h, err := blake2b.New(&blake2b.Config{
		Person: person(n, k),
		Size:   uint8((512 / n) * n / 8),
	})
	return h, err
}

// expands the hash array based on its parameters
// TODO(jaupe) provide better description
func expandArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errSmallBitLen
	}
	if wordSize < 7+bitLen {
		return nil, errSmallWordSize
	}
	outWidth := (bitLen+7)/8 + bytePad
	if outLen != 8*outWidth*len(in)/bitLen {
		return nil, errBadOutLen
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

// a better descriptive type to represent a equihash solution; that is a list of indices that xor to 0
type equihashSolution []int

// hashKey contains the xor'd hash and the indices that we're used to xor
type hashKey struct {
	hash    []byte
	indices []int
}

// minInt returns the smallest number between the two arguments
func minInt(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

// xor  performs xor against two byte slices component-wise and
// returns a new slice with the xor result
func xor(a, b []byte) []byte {
	n := minInt(len(a), len(b))
	x := make([]byte, n)
	for i := 0; i < n; i++ {
		x[i] = a[i] ^ b[i]
	}
	return x
}

// hasCollision returns true if there's a hash collision between
// both x and y slices and the given bit position (i)
func hasCollision(x, y []byte, i, length int) bool {
	start, end := (i-1)*length/8, i*length/8
	for j := start; j < end; j++ {
		if x[j] == y[j] {
			return true
		}
	}
	return false
}

// collisionOffset returns the offset between the last hash (n-1)
func collisionOffset(keys []hashKey, i, collisionLen int) int {
	n := len(keys)
	if n == 0 {
		return -1
	}
	last := keys[n-1]
	ha := last.hash
	for j := 1; j < n; j++ {
		hb := keys[n-1-j].hash
		if !hasCollision(ha, hb, i, collisionLen) {
			return j
		}
	}
	return n
}

// hasDistinctIndices returns true if the indices are unique between
// two lists of indices; returns false if not unique
func hasDistinctIndices(a, b []int) bool {
	for _, av := range a {
		for _, bv := range b {
			if av == bv {
				return false
			}
		}
	}
	return true
}

// concatenates the solution indices of two disjoint indices list
func concatIndices(x, y []int) []int {
	if len(x) == 0 {
		return y
	}
	if len(y) == 0 {
		return x
	}
	if x[0] < y[0] {
		return append(x, y...)
	}
	return append(y, x...)
}

func indicesPerHashOutput(n int) int {
	return 512 / n
}

func hashLength(n, k int) int {
	return (k + 1) * ((collisionLength(n, k) + 7) / 8)
}

// Generate hash keys based on equihash params and pre-populated hash digest
func generateHashKeys(n, k int, digest hash.Hash) ([]hashKey, error) {
	err := validateEquihashParams(n, k)
	if err != nil {
		return nil, err
	}
	if digest == nil {
		return nil, errNullHash
	}
	var keys []hashKey
	var tmpHash []byte
	collisionLen, indicesPerHash := collisionLength(n, k), indicesPerHashOutput(n)
	hashLen := hashLength(n, k)
	for i := 0; i < powOf2(collisionLen+1); i++ {
		r := i % indicesPerHash
		if r == 0 {
			currDigest := copyHash(digest)
			err := hashXi(currDigest, i/indicesPerHash)
			if err != nil {
				return nil, err
			}
			tmpHash = hashDigest(currDigest)
		}
		d := tmpHash[r*n/8 : (r+1)*n/8]
		expanded, err := expandArray(d, hashLen, collisionLen, 0)
		if err != nil {
			return nil, errors.New("expandArray err: " + err.Error() + "\n")
		}
		keys = append(keys, hashKey{expanded, []int{i}})
	}
	return keys, nil
}

// reduces the hash keys based on the parameters (n and k)
func reduceHashKeys(n, k int, keys []hashKey) ([]hashKey, error) {
	err := validateEquihashParams(n, k)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errEmptyKeys
	}
	collisionLen := collisionLength(n, k)
	for i := 1; i < k; i++ {
		// sort tuples by hash
		sort.Sort(hashKeys(keys))

		var xc []hashKey
		for len(keys) > 0 {
			// 2b) Find next set of unordered pairs with collisions on first n/(k+1) bits
			xSize := len(keys)
			j := collisionOffset(keys, i, collisionLen)

			// 2c) Store tuples (X_i ^ X_j, (i, j)) on the table
			for l := 0; l < j-1; l++ {
				for m := l + 1; m < l; m++ {
					x1l, x1m := keys[xSize-1-l], keys[xSize-1-m]
					if hasDistinctIndices(x1l.indices, x1m.indices) {
						concat := concatIndices(x1l.indices, x1m.indices)
						a, b := x1l.hash, x1m.hash
						xc = append(xc, hashKey{xor(a, b), concat})
					}
				}
			}
			// 2d) drop this set
			keys = keys[:len(keys)-j]
		}
		// 2e) replace previous list with new list
		keys = xc
	}
	return keys, nil
}

// find solutions based on the reduced hash keys
func findSolutions(n, k int, keys []hashKey) ([]equihashSolution, error) {
	err := validateEquihashParams(n, k)
	if err != nil {
		return nil, err
	}
	// ensure keys are sorted after reduction
	sort.Sort(hashKeys(keys))
	// find solutions
	var solutions []equihashSolution
	hashLen := hashLength(n, k)
	for len(keys) > 0 {
		xn := len(keys)
		j := solutionOffset(keys, n, k)
		for l := 0; l < j-1; l++ {
			for m := l + 1; m < j; m++ {
				a, b := keys[xn-1-l], keys[xn-1-m]
				res := xor(a.hash, b.hash)
				f1 := countZeros(res) == 8*hashLen
				f2 := hasDistinctIndices(a.indices, b.indices)
				if f1 && f2 {
					indices := concatIndices(a.indices, b.indices)
					solutions = append(solutions, indices)
				}
			}
		}
		keys = keys[:len(keys)-j]
	}
	return solutions, nil
}

// equihash is the general birthday problem - which is the cryptopuzzle used for mining
// digest is the hash to copy that is already pre-populated
// n is the number of hashes to used to solve the problem; the more hashes, the more time it takes to solve
// k is the number used to select 2^k hashes at a time to see if - when xor'd - equals 0;
// the higher the number; the probability to find a solution increases.
// it returns the indices of the N hashes that solve the equihash puzzle
func equihash(digest hash.Hash, n, k int) ([]equihashSolution, error) {
	// validateEquihashArgs
	if digest == nil {
		return nil, errNullHash
	}
	err := validateEquihashParams(n, k)
	if err != nil {
		return nil, err
	}
	keys, err := generateHashKeys(n, k, digest)
	if err != nil {
		return nil, err
	}

	keys, err = reduceHashKeys(n, k, keys)
	if err != nil {
		return nil, err
	}

	return findSolutions(n, k, keys)
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

// returns the first index of offset index that doesn't collide with the first hash
func solutionOffset(keys []hashKey, n, k int) int {
	keysLen, collLen := len(keys), collisionLength(n, k)
	for j := 1; j < keysLen; j++ {
		lc := hasCollision(keys[keysLen-1].hash, keys[keysLen-1-j].hash, k, collLen)
		rc := hasCollision(keys[keysLen-1].hash, keys[keysLen-1-j].hash, k+1, collLen)
		if !(lc && rc) {
			return j
		}
	}
	return keysLen
}

// pow returns pow of base 2 for only positive k
// TODO(jaupe) look into handling overflow
func powOf2(k int) int {
	if k < 1 {
		return 1
	}
	return 1 << uint(k)
}

// hasDuplicateIndices checks for duplicate indices within the same slice
// returns true if there are duplicate numbers (indices)
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

// writes a bytes slice to a hash from start to end of the slice (full slice)
// TODO(jaupe) rewrite when slice is partially written to hash,
// by re-writing what was not written
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

// write a 32-bit unsigned int using little endian to the hash
func writeU32ToHash(h hash.Hash, v uint32) error {
	return writeBytesToHash(h, writeU32(v))
}

// encodes a 32-bit unsigned int to an allocated byte slice
func writeU32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

// copyHash does a deep copy of a hash and returns the deep copy
func copyHash(src hash.Hash) hash.Hash {
	if src == nil {
		return nil
	}
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

// generateWord generates a word to create a slice of words for validating a solution
func generateWord(n int, h hash.Hash, idx int) (*big.Int, error) {
	if h == nil {
		return nil, errNullHash
	}

	bytesPerWord := n / 8
	wordsPerHash := indicesPerHashOutput(n)

	hidx := idx / wordsPerHash
	hrem := idx % wordsPerHash

	idxdata := writeU32(uint32(hidx))
	ctx1 := copyHash(h)
	err := writeBytesToHash(ctx1, idxdata)
	if err != nil {
		return nil, err
	}
	digest := hashDigest(ctx1)

	// fold word
	word := big.NewInt(0)
	for i := hrem * bytesPerWord; i < hrem*bytesPerWord+bytesPerWord; i++ {
		word = word.Lsh(word, 8)
		word = word.Or(word, big.NewInt(int64(digest[i])&0xFF))
	}
	return word, nil
}

func solutionLength(k int) int {
	return powOf2(k)
}

// generates a slice of words used for validating a solution
func generateWords(n, k int, indices []int, h hash.Hash) ([]*big.Int, error) {
	if h == nil {
		return nil, errNullHash
	}
	if len(indices) == 0 {
		return nil, errEmptyIndices
	}
	solutionLen := solutionLength(k)
	var words []*big.Int
	for i := 0; i < solutionLen; i++ {
		word, err := generateWord(n, h, indices[i])
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

func validateNonEmptySolutionParams(header []byte, solutionIndices []int) error {
	if len(header) == 0 {
		return errors.New("empty header")
	}
	if len(solutionIndices) == 0 {
		return errors.New("empty solution indices")
	}
	return nil
}

func validateSolutionIndices(k int, indices []int) error {
	solutionLen := powOf2(k)
	if len(indices) != solutionLen {
		return errBadArg
	}
	if hasDuplicateIndices(indices) {
		return errDuplicateIndices
	}
	return nil
}

func validateSolutionParams(n, k int, header []byte, indices []int) error {
	err := validateEquihashParams(n, k)
	if err != nil {
		return err
	}

	err = validateNonEmptySolutionParams(header, indices)
	if err != nil {
		return err
	}

	return validateSolutionIndices(k, indices)
}

func newValidateHash(n, k int, header []byte) (hash.Hash, error) {
	h, err := newHash(n, k)

	if err != nil {
		return nil, err
	}
	err = writeBytesToHash(h, header)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func validateSolutionOrdering(k int, indices []int) error {
	solutionLen := powOf2(k)
	for s := 0; s < k; s++ {
		d := 1 << uint(s)
		for i := 0; i < solutionLen; i += 2 * d {
			if indices[i] >= indices[i+d] {
				return errPairWiseOrdering
			}
		}
	}
	return nil
}

func validateWords(n, k int, words []*big.Int) (bool, error) {
	solutionLen := powOf2(k)
	bitsPerStage := n / (k + 1)
	for s := 0; s < k; s++ {
		d := 1 << uint(s)
		for i := 0; i < solutionLen; i += 2 * d {
			w := words[i].Xor(words[i], words[i+d])
			if !isBigIntZero(w.Rsh(w, uint(n-(s+1)*bitsPerStage))) {
				return false, errBadWord
			}
			words[i] = w
		}
	}
	return isBigIntZero(words[0]), nil
}

func validateIndices(n, k int, indices []int, digest hash.Hash) (bool, error) {
	// check pair-wise ordering of solution indices
	err := validateSolutionOrdering(k, indices)
	if err != nil {
		return false, err
	}

	words, err := generateWords(n, k, indices, digest)
	if err != nil {
		return false, err
	}

	return validateWords(n, k, words)
}

// ValidateSolution validates that a mining solution is correct
func ValidateSolution(n, k int, header []byte, solutionIndices []int) (bool, error) {
	err := validateSolutionParams(n, k, header, solutionIndices)
	if err != nil {
		return false, err
	}

	// create hash digest and words
	digest, err := newValidateHash(n, k, header)
	if err != nil {
		return false, err
	}

	return validateIndices(n, k, solutionIndices, digest)
}

// isBigIntZero returns true if the big int (arbitrary sized int) equals zero.
// returns false if not equal to zero.
func isBigIntZero(w *big.Int) bool {
	return w.Cmp(bigZero) == 0
}

// validateEquihashParams validates the two main parameters of equihash
func validateEquihashParams(n, k int) error {
	if n < 2 {
		return errors.New("n < 2")
	}
	if k < 3 {
		return errors.New("k < 3")
	}
	if (n % 8) != 0 {
		return errors.New("n%8 != 0")
	}
	if (n % (k + 1)) != 0 {
		return errors.New("n%(k+1) != 0")
	}
	if k >= n {
		return errKLarge
	}
	if collisionLength(n, k)+1 >= 32 {
		return errCollisionLen
	}
	return nil
}

// collisionLength returns the number of bits used for detecting collision length
func collisionLength(n, k int) int {
	return n / (k + 1)
}
