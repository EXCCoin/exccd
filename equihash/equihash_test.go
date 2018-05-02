package equihash

import (
	"encoding/hex"
	"errors"
	"strconv"
	"testing"
)

func hashStr(hash []byte) string {
	return hex.EncodeToString(hash)
}

func sliceMemoryEq(a, b []byte) bool {
	return &a[cap(a)-1] == &b[cap(b)-1]
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

func decodeHex(s string) []byte {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return decoded
}

func testExpandCompressArray(t *testing.T, p expandCompressParams) {
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

type expandCompressParams struct {
	in      string
	out     string
	bitLen  int
	bytePad int
}

var expandCompressTests = []expandCompressParams{
	{"ffffffffffffffffffffff", "07ff07ff07ff07ff07ff07ff07ff07ff", 11, 0},
	{"aaaaad55556aaaab55555aaaaad55556aaaab55555", "155555155555155555155555155555155555155555155555", 21, 0},
	{"000220000a7ffffe00123022b38226ac19bdf23456", "0000440000291fffff0001230045670089ab00cdef123456", 21, 0},
	{"cccf333cccf333cccf333cccf333cccf333cccf333cccf333cccf333", "3333333333333333333333333333333333333333333333333333333333333333", 14, 0},
	{"ffffffffffffffffffffff", "000007ff000007ff000007ff000007ff000007ff000007ff000007ff000007ff", 11, 2},
}

func TestExpandCompressArrays(t *testing.T) {
	for _, params := range expandCompressTests {
		testExpandCompressArray(t, params)
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
func TestCountZeros(t *testing.T) {
	in := []byte{1, 1, 1, 1, 1}
	count := countZeros(in)
	if count != 0 {
		t.Fail()
	}
	for i := 0; i < len(in); i++ {
		in[i] = 0
		count = countZeros(in)
		if (i + 1) != count {
			t.Fail()
		}
	}
}
func TestHasCollision(t *testing.T) {
	h := make([]byte, 1, 32)
	r, err := hasCollision(h, h, 1)
	if err != nil {
		t.Error(err)
	}
	if r == false {
		t.Fail()
	}
}
func TestHasCollision_AStartPos(t *testing.T) {
	ha, hb := []byte{}, []byte{1, 2, 3, 4, 5}
	r, err := hasCollision(ha, hb, 1)
	if r {
		t.Errorf("r = %v\n", r)
	}
	if err == nil {
		t.Error(err)
	}
}
func TestHasCollision_BStartPos(t *testing.T) {
	hb, ha := []byte{}, []byte{1, 2, 3, 4, 5}
	r, err := hasCollision(ha, hb, 1)
	if r {
		t.Errorf("r = %v\n", r)
	}
	if err == nil {
		t.Error(err)
	}
}
func TestHasCollision_HashLen(t *testing.T) {
	hb, ha := []byte{}, []byte{1, 2, 3, 4, 5}
	r, err := hasCollision(ha, hb, 1)
	if r {
		t.Fail()
	}
	if err == nil {
		t.Fail()
	}
}
func TestBinPowInt(t *testing.T) {
	pow := 1
	for i := 0; i < 64; i++ {
		val := binPowInt(i)
		if pow != val {
			t.Errorf("binPowInt(%v) == %v and not %v\n", i, val, pow)
		}
		pow *= 2
	}
}
func TestBinPowInt_NegIndices(t *testing.T) {
	for i := 0; i < 64; i++ {
		k := -1 - i
		if binPowInt(k) != 1 {
			t.Errorf("binPowInt(%v) != 1\n", k)
		}
	}
}
func TestCollisionLen(t *testing.T) {
	n, k := 90, 2
	r := collisionLen(n, k)
	if r != 30 {
		t.FailNow()
	}
	n, k = 200, 90
	r = collisionLen(n, k)
	if r != 2 {
		t.Fail()
	}
}
func TestMinSlices_A(t *testing.T) {
	a := make([]byte, 1, 5)
	b := make([]byte, 1, 10)
	small, large := minSlices(a, b)
	err := byteSliceEq(a, small)
	if err != nil {
		t.Error(err)
	}
	err = byteSliceEq(b, large)
	if err != nil {
		t.Error(err)
	}
}
func TestMinSlices_B(t *testing.T) {
	a := make([]byte, 1, 10)
	b := make([]byte, 1, 5)
	small, large := minSlices(a, b)
	err := byteSliceEq(b, small)
	if err != nil {
		t.Error(err)
	}
	err = byteSliceEq(a, large)
	if err != nil {
		t.Error(err)
	}
}
func TestMinSlices_Eq(t *testing.T) {
	a := make([]byte, 1, 5)
	b := make([]byte, 1, 5)
	small, large := minSlices(a, b)
	err := byteSliceEq(a, small)
	if err != nil {
		t.Error(err)
	}
	err = byteSliceEq(b, large)
	if err != nil {
		t.Error(err)
	}
}
func TestValidateParams_ErrKTooLarge(t *testing.T) {
	n, k := 200, 200
	err := validateParams(n, k)
	if err == nil {
		t.Fail()
	}
	n, k = 200, 201
	err = validateParams(n, k)
	if err == nil {
		t.Fail()
	}
}
func TestValidateParams_ErrCollision(t *testing.T) {
	n, k := 200, 200
	err := validateParams(n, k)
	if err == nil {
		t.Fail()
	}
}
func TestValidateParams(t *testing.T) {
	n, k := 200, 90
	err := validateParams(n, k)
	if err != nil {
		t.Error(err)
	}
}

func testXorEmptySlice(t *testing.T, a, b []byte) {
	r := xor(a, b)
	if len(r) != 0 {
		t.Errorf("r should be empty")
	}
}
func TestXor_EmptySlices(t *testing.T) {
	a, b := []byte{1, 2, 3, 4, 5}, []byte{}
	testXorEmptySlice(t, a, b)
	b, a = a, b
	testXorEmptySlice(t, a, b)
	a, b = []byte{}, []byte{}
	testXorEmptySlice(t, a, b)
}
func TestXor_NilSlices(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	var b []byte
	testXorEmptySlice(t, a, b)
	b, a = a, b
	testXorEmptySlice(t, a, b)
	var x []byte
	var y []byte
	testXorEmptySlice(t, x, y)
}

func testXor(t *testing.T, a, b, exp []byte) {
	act := xor(a, b)
	err := byteSliceEq(act, exp)
	if err != nil {
		t.Error(err)
	}
}
func TestXor_Pass(t *testing.T) {
	a, b := []byte{0, 1, 0, 1, 0, 1}, []byte{1, 0, 1, 0, 1, 0}
	exp := []byte{1, 1, 1, 1, 1, 1}
	testXor(t, a, b, exp)
	a, b = []byte{1, 0, 1, 0, 1, 0}, []byte{0, 1, 0, 1, 0, 1}
	testXor(t, a, b, exp)
	a, b, exp = []byte{0, 0, 1, 1}, []byte{1, 1, 0, 0}, []byte{1, 1, 1, 1}
	testXor(t, a, b, exp)
	a, b = []byte{1, 1, 1, 1}, []byte{0, 0, 0, 0}
	exp = []byte{1, 1, 1, 1}
	testXor(t, a, b, exp)
	a, b = b, a
	testXor(t, a, b, exp)
}
func TestXor_Fail(t *testing.T) {
	a, b := []byte{0, 1, 0, 1}, []byte{0, 1, 0, 1}
	exp := []byte{0, 0, 0, 0}
	testXor(t, a, b, exp)
	a, b = []byte{0, 0, 0, 0}, []byte{0, 0, 0, 0}
	testXor(t, a, b, exp)
	a, b = []byte{1, 1, 1, 1}, []byte{1, 1, 1, 1}
	testXor(t, a, b, exp)
}

func loweralpha() string {
	p := make([]byte, 26)
	for i := range p {
		p[i] = 'a' + byte(i)
	}
	return string(p)
}
func TestEquihash_LowCollisions(t *testing.T) {
	alpha, s, set := loweralpha(), "", make(map[string]bool)
	for i := 0; i < 526; i++ {
		for _, c := range alpha {
			s += string(c)
			h, err := Equihash([]byte(s))
			if err != nil {
				t.Error(err)
			}
			hs := hashStr(h)
			if set[hs] {
				t.Errorf("error collision: %v with %v\n", s, hs)
			}
			set[hs] = true
		}
	}
}
func TestHashSize(t *testing.T) {
	if 64 != hashSize {
		t.Errorf("hashSize should equal 64 and not %v\n", hashSize)
	}
}
