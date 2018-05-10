package consensus

import (
	"encoding/binary"
	"errors"
	"hash"
	"reflect"

	"golang.org/x/crypto/blake2b"
)

var (
	errBadArg   = errors.New("invalid argument")
	errWriteLen = errors.New("didn't write full len")
)

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

func validateSolution(n, k int, person, header []byte, solutionIndices []int) (bool, error) {
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
	digest, err := blake2b.New(outLen, person)
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

	return words[0] == 0, nil
}
