package equihash

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func btoa(b byte) string {
	return strconv.Itoa(int(b))
}

func bytesEq(a, b []byte) error {
	if len(a) != len(b) {
		return errors.New("byte arrays not same len")
	}
	for i, val := range a {
		if val != b[i] {
			return errors.New(btoa(val) + " dont equal " + btoa(b[i]) + " at" + strconv.Itoa(i))
		}
	}
	return nil
}

func testExpandAndCompress(t *testing.T, scope, src, dst string, bitLen, bytePad int) {
	compact, err := ParseHex(src)
	if err != nil {
		t.Error(err)
	}
	expanded, err := ParseHex(dst)
	fmt.Println(expanded)
	if err != nil {
		t.Error(err)
	}
	out := make([]byte, len(expanded))
	err = ExpandArray(compact, out, uint32(bitLen), uint32(bytePad))
	if err != nil {
		t.Error(err)
	}
	err = bytesEq(expanded, out)
	if err != nil {
		t.Error(err)
	}
	//out = make([]byte, len(compact))
	//err = CompressArray(expanded, out, bitLen, bytePad)

}

func TestExpandAndCompressArrays(t *testing.T) {
	testExpandAndCompress(t,
		"8 11-bit chunks, all-ones",
		"ffffffffffffffffffffff",
		"07ff07ff07ff07ff07ff07ff07ff07ff",
		11,
		0)
}
