package consensus

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"math/big"
	"reflect"
	"sort"

	blake2b "github.com/minio/blake2b-simd"
)

const (
	wordSize = 32
	wordMask = (1 << wordSize) - 1
	byteMask = 0xFF
	hashSize = 64
	// N is the number of hash digests used to find a mining solution
	N = 96
	// K is the exponent for xor'ing 2^k hash digests for solution
	K = 5
)

var (
	errBadArg           = errors.New("invalid argument")
	errWriteLen         = errors.New("didn't write full len")
	errLen              = errors.New("slices not same len")
	errBitLen           = errors.New("bit len < 8")
	errOutWidth         = errors.New("incorrect outwidth size")
	errKLarge           = errors.New("k should be less than n")
	errCollisionLen     = errors.New("collision length too big")
	errHashLen          = errors.New("hash len is too small")
	errHashStartPos     = errors.New("hash len < start pos")
	errHashEndPos       = errors.New("hash len < end pos")
	errNonce            = errors.New("no valid nonce")
	errStartEndEq       = errors.New("start and end positions equal")
	errLenZero          = errors.New("len is 0")
	errNil              = errors.New("unexpected nil pointer")
	errEmptySlice       = errors.New("empty slice")
	errSmallBitLen      = errors.New("bitLen < 8")
	errSmallWordSize    = errors.New("wordSize < 7+bitLen")
	errBadOutLen        = errors.New("outLen != 8*outWidth*len(in)/bitLen")
	errDuplicateIndices = errors.New("duplicate indices")
	errPairWiseOrdering = errors.New("bad pair-wise ordering")
	errBadWord          = errors.New("bad word")
	exccPrefix          = "excc"
	bigZero             = big.NewInt(0)
)

func person(prefix string, n, k int) []byte {
	nb, kb := writeUint32(uint32(n)), writeUint32(uint32(k))
	return append([]byte(prefix), append(nb, kb...)...)
}

func exccPerson(n, k int) []byte {
	return person(exccPrefix, n, k)
}

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

func sortHashKeys(k []hashKey) {
	sort.Sort(hashKeys(k))
}

func gbp(digest hash.Hash, n, k int) ([]equihashSolution, error) {
	collisionLength := n / (k + 1)
	hashLength := (k + 1) * ((collisionLength + 7) / 8)
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
			tmpHash = hashDigest(currDigest)
		}
		d := tmpHash[r*n/8 : (r+1)*n/8]
		expanded, err := expandArray(d, hashLength, collisionLength, 0)
		if err != nil {
			return nil, errors.New("expandArray err: " + err.Error() + "\n")
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

func generateWord(n int, digestWithoutIdx hash.Hash, idx int) (*big.Int, error) {
	bytesPerWord := n / 8
	wordsPerHash := 512 / n

	hidx := idx / wordsPerHash
	hrem := idx % wordsPerHash

	idxdata := writeUint32(uint32(hidx))
	ctx1 := copyHash(digestWithoutIdx)
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

func compressArray(in []byte, outLen, bitLen, bytePad int) ([]byte, error) {
	if bitLen < 8 {
		return nil, errors.New("bitLen < 8")
	}
	if wordSize < 7+bitLen {
		return nil, errors.New("wordSize < 7+bitLen")
	}
	inWidth := (bitLen+7)/8 + bytePad
	if outLen != bitLen*len(in)/(8*inWidth) {
		return nil, errors.New("bitLen*len(in)/(8*inWidth)")
	}
	out := make([]byte, outLen)
	bitLenMask := (1 << uint(bitLen)) - 1
	accBits, accVal, j := 0, 0, 0

	for i := 0; i < outLen; i++ {
		if accBits < 8 {
			accVal = ((accVal << uint(bitLen)) & wordMask) | int(in[j])
			for x := bytePad; x < inWidth; x++ {
				v := int(in[j+x])
				a1 := bitLenMask >> (uint(8 * (inWidth - x - 1)))
				b := ((v & a1) & 0xFF) << uint(8*(inWidth-x-1))
				accVal = accVal | int(b)
			}
			j += inWidth
			accBits += bitLen
		}
		accBits -= 8
		out[i] = byte((accVal >> uint(accBits)) & 0xFF)
	}

	return out, nil
}

func generateWords(n, solutionLen int, indices []int, h hash.Hash) ([]*big.Int, error) {
	words := []*big.Int{}
	for i := 0; i < solutionLen; i++ {
		word, err := generateWord(n, h, indices[i])
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	return words, nil
}

func minSlices(a, b []int) ([]int, []int) {
	if len(a) <= len(b) {
		return a, b
	}
	return b, a
}

func joinBytes(a, b []byte) []byte {
	return append(a, b...)
}

// ValidateSolution validates that a mining solution is correct
func ValidateSolution(n, k int, person, header []byte, solutionIndices []int, prefix string) (bool, error) {
	if n < 2 {
		return false, errors.New("n < 2")
	}
	if k < 3 {
		return false, errors.New("k < 3")
	}
	if (n % 8) != 0 {
		return false, errors.New("n%8 != 0")
	}
	if (n % (k + 1)) != 0 {
		return false, errors.New("n%(k+1) != 0")
	}
	if len(person) == 0 {
		return false, errors.New("empty person")
	}
	if len(header) == 0 {
		return false, errors.New("empty header")
	}
	if len(solutionIndices) == 0 {
		return false, errors.New("empty solution indices")
	}
	solutionLen := pow(k)
	if len(solutionIndices) != solutionLen {
		return false, errBadArg
	}
	if hasDuplicateIndices(solutionIndices) {
		return false, errDuplicateIndices
	}

	bytesPerWord := n / 8
	wordsPerHash := 512 / n
	outLen := wordsPerHash * bytesPerWord

	// create hash digest and words
	digest, err := blake2b.New(&blake2b.Config{
		Person: person,
		Size:   uint8(outLen),
	})
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
				return false, errPairWiseOrdering
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
			w := words[i].Xor(words[i], words[i+d])
			if !isBigIntZero(w.Rsh(w, uint(n-(s+1)*bitsPerStage))) {
				return false, errBadWord
			}
			words[i] = w
		}
	}
	return isBigIntZero(words[0]), nil
}

func isBigIntZero(w *big.Int) bool {
	return w.Cmp(bigZero) == 0
}

func newBlake2bHash(n, k int, prefix string, prevHash []byte) (hash.Hash, error) {
	return newHash(n, k, prefix)
}

// MiningResult provides the details of the mining result
type MiningResult struct {
	previousHash []byte
	currHash     []byte
	nonce        int
}

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

func hashNonce(h hash.Hash, nonce int) error {
	for i := 0; i < 8; i++ {
		err := writeHashU32(h, uint32(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func putU32(b []byte, v uint32) {
	binary.LittleEndian.PutUint32(b, v)
}

func writeU32(v uint32) []byte {
	b := make([]byte, 4)
	putU32(b, v)
	return b
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

func newHash(n, k int, prefix string) (hash.Hash, error) {
	h, err := blake2b.New(&blake2b.Config{
		Key:    nil,
		Person: person(prefix, n, k),
		Size:   uint8(hashDigestSize(n)),
	})
	return h, err
}

func blockHash(n, k int, prefix string, prevHash []byte, nonce int, soln []int) ([]byte, error) {
	h, err := newHash(n, k, prefix)
	if err != nil {
		return nil, err
	}
	err = writeHashBytes(h, prevHash)
	if err != nil {
		return nil, err
	}
	hashNonce(h, nonce)
	for _, xi := range soln {
		err := hashXi(h, xi)
		if err != nil {
			return nil, err
		}
	}
	hb := hashDigest(h)
	// double hash
	h, err = newHash(n, k, prefix)
	if err != nil {
		return nil, err
	}
	err = writeHashBytes(h, hb)
	if err != nil {
		return nil, err
	}
	return hashDigest(h), nil
}

func difficultyFilter(n, k int, prefix string, prevHash []byte, nonce int, soln []int, d int) bool {
	h, err := blockHash(n, k, prefix, prevHash, nonce, soln)
	if err != nil {
		return false
	}
	count := countZeros(h)
	return d < count
}

func hashDigestSize(n int) int {
	return (512 / n) * n / 8
}

// Mine mines for equihash solution based on N number of hashes digests.
// It finds 2^k indices of hash digest that equal 0 when xor'd
func Mine(n, k, d int, prefix string) (*MiningResult, error) {
	err := validateParams(n, k)
	if err != nil {
		return nil, err
	}

	digest := sha256.New()
	prevHash := hashDigest(digest)
	var x []int
	nonce := 0
	for x == nil {
		digest, err = newBlake2bHash(n, k, prefix, prevHash)
		if err != nil {
			return nil, err
		}
		nonce = 0

		for {
			currDigest := copyHash(digest)
			err = hashNonce(currDigest, nonce)
			if err != nil {
				return nil, err
			}
			solns, err := gbp(currDigest, n, k)
			if err != nil {
				return nil, err
			}
			for _, soln := range solns {
				if difficultyFilter(n, k, prefix, prevHash, nonce, soln, d) {
					x = soln
					break
				}
			}
			if x != nil {
				break
			}
			nonce++
		}
	}
	currHash, err := blockHash(n, k, prefix, prevHash, nonce, x)
	if err != nil {
		return nil, err
	}
	return &MiningResult{prevHash, currHash, nonce}, nil
}
