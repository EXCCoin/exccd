package equihash

import (
	"encoding/hex"
	"testing"
)

func byteSliceEq(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, val := range a {
		if val != b[i] {
			return false
		}
	}
	return true
}

func decodeHex(s string) []byte {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return decoded
}

func TestExpandArray(t *testing.T) {
	in := decodeHex("ffffffffffffffffffffff")
	exp := decodeHex("07ff07ff07ff07ff07ff07ff07ff07ff")
	bitLen, outLen, bytePad := 11, len(exp), 0
	out, err := expandArray(in, outLen, bitLen, bytePad)
	if err != nil {
		t.Error(err)
	}
	if !byteSliceEq(out, exp) {
		t.FailNow()
	}
}
