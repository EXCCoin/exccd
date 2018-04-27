package equihash

import (
	"encoding/hex"
	"errors"
	"strconv"
	"testing"
)

func decodeHex(s string) []byte {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return decoded
}

// countZeros tests

type testCountZerosParams struct {
	h        []byte
	expected int
}

func TestCountZeros(t *testing.T) {

	var paramsSet = []testCountZerosParams{
		{[]byte{1, 2}, 7},
		{[]byte{255, 255}, 0},
		{[]byte{126, 0, 2}, 1},
		{[]byte{54, 2}, 2},
		{[]byte{0, 0}, 16},
	}

	for _, params := range paramsSet {
		testCountZeros(t, params)
	}

}

func testCountZeros(t *testing.T, p testCountZerosParams) {
	r := countZeros(p.h)

	if r != p.expected {
		t.Error("Should be equal, actual:", r, "expected", p.expected)
	}
}

// expandArray and compressArray tests

type testExpandCompressParams struct {
	in      string
	out     string
	bitLen  int
	bytePad int
}

var expandCompressParamsSet = []testExpandCompressParams{
	{"ffffffffffffffffffffff", "07ff07ff07ff07ff07ff07ff07ff07ff", 11, 0},
	{"aaaaad55556aaaab55555aaaaad55556aaaab55555", "155555155555155555155555155555155555155555155555", 21, 0},
	{"000220000a7ffffe00123022b38226ac19bdf23456", "0000440000291fffff0001230045670089ab00cdef123456", 21, 0},
	{"cccf333cccf333cccf333cccf333cccf333cccf333cccf333cccf333", "3333333333333333333333333333333333333333333333333333333333333333", 14, 0},
	{"ffffffffffffffffffffff", "000007ff000007ff000007ff000007ff000007ff000007ff000007ff000007ff", 11, 2},
}

func byteSliceEq(a, b []byte) error {
	if len(a) != len(b) {
		return errors.New("a and b not same len")
	}
	for i, val := range a {
		if val != b[i] {
			av, bv := strconv.Itoa(int(val)), strconv.Itoa(int(b[i]))
			is := strconv.Itoa(i)
			msg := av + " != " + bv + " at " + is
			return errors.New(msg)
		}
	}
	return nil
}

func TestExpandCompressArray(t *testing.T) {
	for _, params := range expandCompressParamsSet {
		testExpandCompressArray(t, params)
	}
}

func testExpandCompressArray(t *testing.T, p testExpandCompressParams) {
	inHex, outHex := decodeHex(p.in), decodeHex(p.out)
	outLen := len(outHex)

	expanded, err := expandArray(inHex, outLen, p.bitLen, p.bytePad)
	if err != nil {
		t.Error(err)
	}

	err = byteSliceEq(outHex, expanded)
	if err != nil {
		t.Error(err)
	}

	compact, err := compressArray(expanded, len(inHex), p.bitLen, p.bytePad)
	if err != nil {
		t.Error(err)
	}

	err = byteSliceEq(inHex, compact)
	if err != nil {
		t.Error(err)
	}
}

// distinctIndices tests

func TestDistinctIndices(t *testing.T) {
	a := []byte{0, 1, 2, 3, 4, 5}
	b := []byte{0, 1, 2, 3, 4, 5}
	r := distinctIndices(a, b)
	if r {
		t.Error()
	}
	b = []byte{6, 7, 8, 9, 10}
	r = distinctIndices(a, b)
	if !r {
		t.Error()
	}
	a = []byte{7, 8, 9, 10, 11}
	r = distinctIndices(a, b)
	if r {
		t.Error()
	}
}

/*
func TestEquihashSolver(t *testing.T) {
	n := 96
	k := 5
	I := []byte("block header")
	nonce := 0
	solutions := [][]int{
		{976, 126621, 100174, 123328, 38477, 105390, 38834, 90500, 6411, 116489, 51107, 129167, 25557, 92292, 38525, 56514, 1110, 98024, 15426, 74455, 3185, 84007, 24328, 36473, 17427, 129451, 27556, 119967, 31704, 62448, 110460, 117894},
		{1008, 18280, 34711, 57439, 3903, 104059, 81195, 95931, 58336, 118687, 67931, 123026, 64235, 95595, 84355, 122946, 8131, 88988, 45130, 58986, 59899, 78278, 94769, 118158, 25569, 106598, 44224, 96285, 54009, 67246, 85039, 127667},
		{1278, 107636, 80519, 127719, 19716, 130440, 83752, 121810, 15337, 106305, 96940, 117036, 46903, 101115, 82294, 118709, 4915, 70826, 40826, 79883, 37902, 95324, 101092, 112254, 15536, 68760, 68493, 125640, 67620, 108562, 68035, 93430},
		{3976, 108868, 80426, 109742, 33354, 55962, 68338, 80112, 26648, 28006, 64679, 130709, 41182, 126811, 56563, 129040, 4013, 80357, 38063, 91241, 30768, 72264, 97338, 124455, 5607, 36901, 67672, 87377, 17841, 66985, 77087, 85291},
		{5970, 21862, 34861, 102517, 11849, 104563, 91620, 110653, 7619, 52100, 21162, 112513, 74964, 79553, 105558, 127256, 21905, 112672, 81803, 92086, 43695, 97911, 66587, 104119, 29017, 61613, 97690, 106345, 47428, 98460, 53655, 109002},
	}
}
*/
