package equihash

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"log"
	"sort"

	"github.com/golang/glog"

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
	errNonce        = errors.New("no valid nonce")
	errStartEndEq   = errors.New("start and end positions equal")
	errLenZero      = errors.New("len is 0")
	errNil          = errors.New("unexpected nil pointer")
	errBadArg       = errors.New("bad arg")
	errEmptySlice   = errors.New("empty slice")
	personPrefix    = []byte("excc")
)

const (
	N        = 96
	K        = 5
	wordSize = 32
	wordMask = (1 << wordSize) - 1
	hashSize = 64
)

type solution struct {
	digest  []byte
	indices []int
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
		return nil
	}
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return nil
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
	if start == end {
		return false, errStartEndEq
	}
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

func hasCollision2(a, b []byte, l int) bool {
	for j := 0; j < l; j++ {
		if a[j] == b[j] {
			return false
		}
	}
	return true
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

	bitLenMask := (1 << uint(bitLen)) - 1

	// The accBits least-significant bits of accVal represent a bit sequence
	// in big-endian order.
	accBits, accVal := 0, 0

	out, j := make([]byte, outLen), 0
	for i := 0; i < outLen; i++ {
		// When we have fewer than 8 bits left in the accumulator, read the next
		// input element.
		if accBits < 8 {
			accVal = accVal << uint(bitLen)
			for x := bytePad; x < inWidth; x++ {
				// Apply bit_len_mask across byte boundaries
				g := int(in[j+x]) & ((bitLenMask >> uint(8*(inWidth-x-1))) & 0xFF)
				// Big-endian
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

	// The accBits least-significant bits of accVal represent a bit sequence
	// in big-endian order.
	accBits, accVal := 0, uint(0)

	out, j := make([]byte, outLen), 0
	for i := 0; i < len(in); i++ {
		accVal = (accVal << 8) | uint(in[i])
		accBits += 8

		// When we have bitLen or more bits in the accumulator, write the next
		// output element.
		if accBits >= bitLen {
			accBits -= bitLen
			for x := bytePad; x < outWidth; x++ {
				// Big-endian
				a := accVal >> (uint(accBits + (8 * (outWidth - x - 1))))
				// Apply bit_len_mask across byte boundaries
				b := (bitLenMask >> uint((8 * (outWidth - x - 1)))) & 0xFF
				out[j+x] = byte(a) & byte(b)
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

func newHash() (hash.Hash, error) {
	return blake2b.New(hashSize, nil)
}

func newKeyedHash(n, k int) (hash.Hash, error) {
	return blake2b.New((512/n)*n/8, exccPerson(n, k))
}

type hashBuilder struct {
	n      int
	k      int
	prefix []byte
}

func newHashBuilder(n, k int, prefix []byte) hashBuilder {
	return hashBuilder{n, k, prefix}
}

func copyByteSlice(in []byte) []byte {
	out := make([]byte, len(in))
	for i, val := range in {
		out[i] = val
	}
	return out
}

func (hb *hashBuilder) copy() hashBuilder {
	return hashBuilder{hb.n, hb.k, copyByteSlice(hb.prefix)}
}

func (hb *hashBuilder) append(b []byte) {
	hb.prefix = joinBytes(hb.prefix, b)
}

func (hb *hashBuilder) writeHashXi(xi int) {
	hb.writeUint32(uint32(xi))
}

func (hb *hashBuilder) writeNonce(nonce int) {
	hb.writeUint32(uint32(nonce))
}

func (hb *hashBuilder) writeUint32(x uint32) {
	b := writeU32(x)
	hb.append(b)
}

func (hb *hashBuilder) digest() ([]byte, error) {
	h, err := newKeyedHash(hb.n, hb.k)
	if err != nil {
		return nil, err
	}
	err = writeHashBytes(h, hb.prefix)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func inSlice(b []byte, r, n int) []byte {
	i, j := r*n/8, (r+1)*n/8
	return b[i:j]
}

func hashDigest(h hash.Hash) []byte {
	return h.Sum(nil)
}

func negIndex(s []solution, i int) int {
	return len(s) - i
}

/*
func equihash(hb hashBuilder, n, k int) ([][]int, error) {
	collLen := collisionLen(n, k)
	hLen := hashLen(k, collLen)
	indicesPerHash := 512 / n

	x := []solution{}

	for i := 0; i < pow(collLen+1); i++ {
		r := i % indicesPerHash
		if r == 0 {
			copyHB := hb.copy()
			copyHB.writeHashXi(i / indicesPerHash)
			h, err := copyHB.digest()
			if err != nil {
				return nil, err
			}
			digest, err := expandArray(inSlice(h, r, n), hLen, collLen, 0)
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
*/

type solutions []solution

func (s solutions) Len() int {
	return len(s)
}

func (s solutions) Less(i, j int) bool {
	x, y := s[i].digest, s[j].digest
	return bytesCompare(x, y)
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

func difficultyFilter(prevHash []byte, nonce int, soln []int, d int) bool {
	h, err := blockHash(prevHash, nonce, soln)
	if err != nil {
		return false
	}
	count := countZeros(h)
	return d < count
}

func blockHash(prevHash []byte, nonce int, soln []int) ([]byte, error) {
	h, err := newHash()
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
	h, err = newHash()
	if err != nil {
		return nil, err
	}
	err = writeHashBytes(h, hb)
	if err != nil {
		return nil, err
	}
	return hashDigest(h), nil
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

//TODO(woj)
func exccPerson(n, k int) []byte {
	return joinBytes(personPrefix, writeParams(n, k))
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

func genesisHash() []byte {
	h := sha256.New()
	return hashDigest(h)
}

func u32b(x int) []byte {
	return writeU32(uint32(x))
}

/*
func mine(n, k, d int) (*miningResult, error) {
	err := validateParams(n, k)
	if err != nil {
		return nil, err
	}
	prevHash := genesisHash()
	var solution []int
	for {
		hb := newHashBuilder(n, k, prevHash)
		nonce := 0
		for nonce != math.MaxInt32 {
			copyHB := hb.copy()
			if err != nil {
				return nil, err
			}
			copyHB.writeNonce(nonce)
			solns, err := equihash(copyHB, n, k)
			if err != nil {
				return nil, err
			}
			for _, s := range solns {
				if difficultyFilter(prevHash, nonce, s, d) {
					solution = s
					break
				}
			}
			if solution != nil {
				break
			}
			nonce++
		}

		currHash, err := blockHash(prevHash, nonce, solution)
		if err != nil {
			return nil, err
		}
		prevHash = currHash
	}
}
*/

type miningResult struct {
	prevHash    []byte
	currentHash []byte
	nonce       int
}

type hashKey struct {
	digest     []byte
	lenIndices int
}

func collisionBitLen(n, k int) int {
	return n / (k + 1)
}

func collisionByteLen(n, k int) int {
	return (collisionBitLen(n, k) + 7) / 8
}

func hashLen(n, k int) int {
	return (k + 1) * collisionByteLen(n, k)
}

func indicesPerHashOutput(n int) int {
	return 512 / n
}

func hashOutput(n int) int {
	return indicesPerHashOutput(n) * n / 8
}

func stepRow(in []byte, n, k, hashLen int) (hashKey, error) {
	bitLen, bytePad := collisionBitLen(n, k), 0
	digest, err := expandArray(in, hashLen, bitLen, bytePad)
	if err != nil {
		return hashKey{}, err
	}
	return hashKey{digest, 0}, nil
}

func generateHashDigest(n, g int) ([]byte, error) {
	hashLen := hashOutput(n)
	h, err := blake2b.New(hashLen, nil)
	if err != nil {
		return nil, err
	}
	err = writeHashU32(h, uint32(g))
	if err != nil {
		return nil, err
	}
	return hashDigest(h), nil
}

func hashSlice(buf []byte, i, n int) []byte {
	if len(buf) == 0 {
		return nil
	}
	start := i * n / 8
	end := start + n/8
	return buf[start:end]
}

func hasSpace(keys []hashKey, n int) bool {
	return len(keys) < n
}

func generateHashKeys(n, k, g int, keys []hashKey) ([]hashKey, error) {
	hash, err := generateHashDigest(n, g)
	if err != nil {
		return nil, err
	}
	outLen, bitLen, bytePad := hashLen(n, k), 8, 0
	initSize, indicesLen := initHashLen(n, k), indicesPerHashOutput(n)
	for i := 0; i < indicesLen && hasSpace(keys, initSize); i++ {
		s := hashSlice(hash, i, n)
		digest, err := expandArray(s, outLen, bitLen, bytePad)
		if err != nil {
			return nil, err
		}
		keys = append(keys, hashKey{digest, i})
	}
	return keys, nil
}

func hashBuffer(n, k int) []byte {
	return make([]byte, hashLen(n, k))
}

func initHashLen(n, k int) int {
	return 1 << uint(collisionBitLen(n, k)+1)
}

func generateHashes(n, k int) ([]hashKey, error) {
	initLen := initHashLen(n, k)
	hashKeys := make([]hashKey, 0, initLen)

	for g := 0; len(hashKeys) < initLen; g++ {
		keys, err := generateHashKeys(n, k, g, hashKeys)
		if err != nil {
			return nil, err
		}
		hashKeys = append(hashKeys, keys...)
	}

	return hashKeys, nil
}

func hashCollisionPair() {

}

func bytesCompare(x, y []byte) bool {
	return bytes.Compare(x, y) < 0
}

type hashKeys []hashKey

func (k hashKeys) Len() int {
	return len(k)
}

func (k hashKeys) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	return bytesCompare(k[i].digest, k[j].digest)
}

func (k hashKeys) Swap(i, j int) {
	k[j], k[i] = k[i], k[j]
}

func finalFullWidth(n, k int) int {
	return 2*collisionByteLen(n, k) + 32*(1<<uint(k-1))
}

func indicesBefore(a, b []byte, size, lenIndices int) bool {
	i, j := size, size+lenIndices
	x, y := a[i:j], b[i:j]
	return bytesCompare(x, y)
}

func newXorHashKey(x, y []byte, n, k int) (hashKey, error) {
	/*
		if key == nil {
			return errNil
		}
		w := finalFullWidth(n, k)
		if size+lenIndices <= w {
			return errBadArg
		}
		if size-trim+(2*lenIndices) <= hashLen(n, k) {
			return errBadArg
		}
		for i := trim; i < size; i++ {
			out[i-trim] = x[i] ^ y[i]
		}
	*/
	return hashKey{}, errnyi()
}

func trim(g, i, n int) int {
	return g*indicesPerHashOutput(n) + i
}

func popBack(k []hashKey) ([]hashKey, hashKey) {
	i := len(k) - 1
	return k[:i], k[i]
}

func processHashes(keys []hashKey, n, k int) ([]hashKey, error) {
	if len(keys) == 0 {
		return nil, errLen
	}
	// loop until 2n/(k+1) bits remain
	hashLen, lenIndices := hashOutput(n), 32
	//collByteLen, tmpHash := collisionByteLen(n, k), make([]byte, hashLen)
	for r := 1; r < k && len(keys) > 0; r++ {
		// 2a) sort the list
		sort.Sort(hashKeys(keys))
		i, l, posFree := 0, 0, 0
		//_ := []hashKey{}
		xc := []hashKey{}
		for i < len(keys)-1 {
			// 2b) find next set of unordered pairs with collisions on the next n/(k+1) bits
			j, a := 1, keys[i].digest
			b := keys[i+j].digest
			for i+j < len(keys) && hasCollision2(a, b, l) {
				j++
			}

			// 2c) Calculate tuples (X_i ^ X_j, (i, j))
			for l := 0; l < j-1; l++ {
				for m := l + 1; m < j; m++ {
					a, b := keys[i+l], keys[i+m]
					if distinctIndices(a.digest, b.digest) {
						key, err := newXorHashKey(a.digest, b.digest, n, k)
						if err != nil {
							return nil, err
						}
						xc = append(xc, key)
					}
				}
			}

			// 2d) Store tuples on the table in place if possible
			for posFree < i+j && len(xc) > 0 {
				head, tail := popBack(xc)
				keys[posFree] = tail
				posFree++
				xc = head
			}

			i += j
		}

		// 2e) handle edge case where final table entry has no collision
		for posFree < len(keys) && len(xc) > 0 {
			head, tail := popBack(xc)
			keys[posFree] = tail
			posFree++
			xc = head
		}

		if len(xc) > 0 {
			keys = append(keys, xc...)
		} else if posFree < len(keys) {
			// remove empty space at back
			keys = keys[:posFree]
		}

		hashLen = hashLen - collisionByteLen(n, k)
		lenIndices *= 2
	}
	return nil, nil
}

func findSolution(keys []hashKey, d uint) ([]byte, error) {
	if len(keys) == 0 {
		return nil, errEmptySlice
	}
	if len(keys) == 1 {
		return nil, errBadArg
	}
	return nil, errnyi()
}

func sortHashKeys(keys []hashKey) {
	sort.Sort(hashKeys(keys))
}

func solve(n, k, d int) ([]byte, error) {
	keys, err := generateHashes(n, k)
	if err != nil {
		return nil, err
	}
	glog.Info(keys)
	keys, err = processHashes(keys, n, k)
	if err != nil {
		return nil, err
	}
	return findSolution(keys, uint(d))
}

func errnyi() error {
	return errors.New("nyi")
}
