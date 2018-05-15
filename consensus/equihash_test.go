package consensus

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strconv"
	"testing"

	blake2b "github.com/minio/blake2b-simd"
)

const (
	difficulty = 1
)

var (
	i                   = []byte("block header")
	expandCompressTests = []expandCompressTest{
		{decodeHex("ffffffffffffffffffffff"), decodeHex("07ff07ff07ff07ff07ff07ff07ff07ff"), 11, 0},
		{decodeHex("aaaaad55556aaaab55555aaaaad55556aaaab55555"), decodeHex("155555155555155555155555155555155555155555155555"), 21, 0},
		{decodeHex("000220000a7ffffe00123022b38226ac19bdf23456"), decodeHex("0000440000291fffff0001230045670089ab00cdef123456"), 21, 0},
		{decodeHex("cccf333cccf333cccf333cccf333cccf333cccf333cccf333cccf333"), decodeHex("3333333333333333333333333333333333333333333333333333333333333333"), 14, 0},
		{decodeHex("ffffffffffffffffffffff"), decodeHex("000007ff000007ff000007ff000007ff000007ff000007ff000007ff000007ff"), 11, 2},
	}
	miningTests        = createMiningTests()
	n                  = N
	k                  = K
	miningResultHeader = []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem.")
	miningResultTests  = createMiningResultTests()
)

func createMiningTests() []miningTest {
	return []miningTest{
		{96, 5, i, 0,
			[][]int{
				[]int{976, 126621, 100174, 123328, 38477, 105390, 38834, 90500, 6411, 116489, 51107, 129167, 25557, 92292, 38525, 56514, 1110, 98024, 15426, 74455, 3185, 84007, 24328, 36473, 17427, 129451, 27556, 119967, 31704, 62448, 110460, 117894},
				[]int{1008, 18280, 34711, 57439, 3903, 104059, 81195, 95931, 58336, 118687, 67931, 123026, 64235, 95595, 84355, 122946, 8131, 88988, 45130, 58986, 59899, 78278, 94769, 118158, 25569, 106598, 44224, 96285, 54009, 67246, 85039, 127667},
				[]int{1278, 107636, 80519, 127719, 19716, 130440, 83752, 121810, 15337, 106305, 96940, 117036, 46903, 101115, 82294, 118709, 4915, 70826, 40826, 79883, 37902, 95324, 101092, 112254, 15536, 68760, 68493, 125640, 67620, 108562, 68035, 93430},
				[]int{3976, 108868, 80426, 109742, 33354, 55962, 68338, 80112, 26648, 28006, 64679, 130709, 41182, 126811, 56563, 129040, 4013, 80357, 38063, 91241, 30768, 72264, 97338, 124455, 5607, 36901, 67672, 87377, 17841, 66985, 77087, 85291},
				[]int{5970, 21862, 34861, 102517, 11849, 104563, 91620, 110653, 7619, 52100, 21162, 112513, 74964, 79553, 105558, 127256, 21905, 112672, 81803, 92086, 43695, 97911, 66587, 104119, 29017, 61613, 97690, 106345, 47428, 98460, 53655, 109002},
			},
		},
	}
}

func createMiningResultTests() []miningResult {
	return []miningResult{
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568},
			true},
		// Change one index
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{2262, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568},
			false},
		// Swap two arbitrary indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{45858, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 2261, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568},
			false},
		// Reverse the first pair of indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{15185, 2261, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568},
			false},
		// Swap the first and second pairs of indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{36112, 104243, 2261, 15185, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568},
			false},
		// Swap the second-to-last and last pairs of indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 104973, 122568, 69567, 114474},
			false},
		// Swap the first half and second half
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568, 2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026},
			false},
		// Sort the indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{2261, 15185, 15972, 23460, 23779, 32642, 36112, 45858, 49807, 52426, 68190, 69567, 69878, 76925, 80080, 80391, 81830, 85191, 90330, 91132, 92842, 104243, 104973, 111026, 114474, 115059, 116805, 118332, 118390, 122568, 122819, 130041},
			false},
		// Duplicate indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{2261, 2261, 15185, 15185, 36112, 36112, 104243, 104243, 23779, 23779, 118390, 118390, 118332, 118332, 130041, 130041, 32642, 32642, 69878, 69878, 76925, 76925, 80080, 80080, 45858, 45858, 116805, 116805, 92842, 92842, 111026, 111026},
			false},
		// Duplicate first half
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1,
			[]int{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026},
			false},
		{96, 5, []byte("block header"), 1,
			[]int{1911, 96020, 94086, 96830, 7895, 51522, 56142, 62444, 15441, 100732, 48983, 64776, 27781, 85932, 101138, 114362, 4497, 14199, 36249, 41817, 23995, 93888, 35798, 96337, 5530, 82377, 66438, 85247, 39332, 78978, 83015, 123505}, true},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 2,
			[]int{6005, 59843, 55560, 70361, 39140, 77856, 44238, 57702, 32125, 121969, 108032, 116542, 37925, 75404, 48671, 111682, 6937, 93582, 53272, 77545, 13715, 40867, 73187, 77853, 7348, 70313, 24935, 24978, 25967, 41062, 58694, 110036}, true},
		{200, 9, []byte("block header"), 0,
			[]int{0, 4313, 223176, 448870, 1692641, 214911, 551567, 1696002, 1768726, 500589, 938660, 724628, 1319625, 632093, 1474613, 665376, 1222606, 244013, 528281, 1741992, 1779660, 313314, 996273, 435612, 1270863, 337273, 1385279, 1031587, 1147423, 349396, 734528, 902268, 1678799, 10902, 1231236, 1454381, 1873452, 120530, 2034017, 948243, 1160178, 198008, 1704079, 1087419, 1734550, 457535, 698704, 649903, 1029510, 75564, 1860165, 1057819, 1609847, 449808, 527480, 1106201, 1252890, 207200, 390061, 1557573, 1711408, 396772, 1026145, 652307, 1712346, 10680, 1027631, 232412, 974380, 457702, 1827006, 1316524, 1400456, 91745, 2032682, 192412, 710106, 556298, 1963798, 1329079, 1504143, 102455, 974420, 639216, 1647860, 223846, 529637, 425255, 680712, 154734, 541808, 443572, 798134, 322981, 1728849, 1306504, 1696726, 57884, 913814, 607595, 1882692, 236616, 1439683, 420968, 943170, 1014827, 1446980, 1468636, 1559477, 1203395, 1760681, 1439278, 1628494, 195166, 198686, 349906, 1208465, 917335, 1361918, 937682, 1885495, 494922, 1745948, 1320024, 1826734, 847745, 894084, 1484918, 1523367, 7981, 1450024, 861459, 1250305, 226676, 329669, 339783, 1935047, 369590, 1564617, 939034, 1908111, 1147449, 1315880, 1276715, 1428599, 168956, 1442649, 766023, 1171907, 273361, 1902110, 1169410, 1786006, 413021, 1465354, 707998, 1134076, 977854, 1604295, 1369720, 1486036, 330340, 1587177, 502224, 1313997, 400402, 1667228, 889478, 946451, 470672, 2019542, 1023489, 2067426, 658974, 876859, 794443, 1667524, 440815, 1099076, 897391, 1214133, 953386, 1932936, 1100512, 1362504, 874364, 975669, 1277680, 1412800, 1227580, 1857265, 1312477, 1514298, 12478, 219890, 534265, 1351062, 65060, 651682, 627900, 1331192, 123915, 865936, 1218072, 1732445, 429968, 1097946, 947293, 1323447, 157573, 1212459, 923792, 1943189, 488881, 1697044, 915443, 2095861, 333566, 732311, 336101, 1600549, 575434, 1978648, 1071114, 1473446, 50017, 54713, 367891, 2055483, 561571, 1714951, 715652, 1347279, 584549, 1642138, 1002587, 1125289, 1364767, 1382627, 1387373, 2054399, 97237, 1677265, 707752, 1265819, 121088, 1810711, 1755448, 1858538, 444653, 1130822, 514258, 1669752, 578843, 729315, 1164894, 1691366, 15609, 1917824, 173620, 587765, 122779, 2024998, 804857, 1619761, 110829, 1514369, 410197, 493788, 637666, 1765683, 782619, 1186388, 494761, 1536166, 1582152, 1868968, 825150, 1709404, 1273757, 1657222, 817285, 1955796, 1014018, 1961262, 873632, 1689675, 985486, 1008905, 130394, 897076, 419669, 535509, 980696, 1557389, 1244581, 1738170, 197814, 1879515, 297204, 1165124, 883018, 1677146, 1545438, 2017790, 345577, 1821269, 761785, 1014134, 746829, 751041, 930466, 1627114, 507500, 588000, 1216514, 1501422, 991142, 1378804, 1797181, 1976685, 60742, 780804, 383613, 645316, 770302, 952908, 1105447, 1878268, 504292, 1961414, 693833, 1198221, 906863, 1733938, 1315563, 2049718, 230826, 2064804, 1224594, 1434135, 897097, 1961763, 993758, 1733428, 306643, 1402222, 532661, 627295, 453009, 973231, 1746809, 1857154, 263652, 1683026, 1082106, 1840879, 768542, 1056514, 888164, 1529401, 327387, 1708909, 961310, 1453127, 375204, 878797, 1311831, 1969930, 451358, 1229838, 583937, 1537472, 467427, 1305086, 812115, 1065593, 532687, 1656280, 954202, 1318066, 1164182, 1963300, 1232462, 1722064, 17572, 923473, 1715089, 2079204, 761569, 1557392, 1133336, 1183431, 175157, 1560762, 418801, 927810, 734183, 825783, 1844176, 1951050, 317246, 336419, 711727, 1630506, 634967, 1595955, 683333, 1461390, 458765, 1834140, 1114189, 1761250, 459168, 1897513, 1403594, 1478683, 29456, 1420249, 877950, 1371156, 767300, 1848863, 1607180, 1819984, 96859, 1601334, 171532, 2068307, 980009, 2083421, 1329455, 2030243, 69434, 1965626, 804515, 1339113, 396271, 1252075, 619032, 2080090, 84140, 658024, 507836, 772757, 154310, 1580686, 706815, 1024831, 66704, 614858, 256342, 957013, 1488503, 1615769, 1515550, 1888497, 245610, 1333432, 302279, 776959, 263110, 1523487, 623933, 2013452, 68977, 122033, 680726, 1849411, 426308, 1292824, 460128, 1613657, 234271, 971899, 1320730, 1559313, 1312540, 1837403, 1690310, 2040071, 149918, 380012, 785058, 1675320, 267071, 1095925, 1149690, 1318422, 361557, 1376579, 1587551, 1715060, 1224593, 1581980, 1354420, 1850496, 151947, 748306, 1987121, 2070676, 273794, 981619, 683206, 1485056, 766481, 2047708, 930443, 2040726, 1136227, 1945705, 1722044, 1971986}, true},
	}

}

type expandCompressTest struct {
	compact  []byte
	expanded []byte
	bitLen   int
	bytePad  int
}

type miningTest struct {
	n         int
	k         int
	I         []byte
	nonce     int
	solutions [][]int
}

type miningResult struct {
	n        int
	k        int
	I        []byte
	nonce    int
	solution []int
	valid    bool
}

func createDigest(n, k, nonce int, I []byte) (hash.Hash, error) {
	//bytesPerWord := n / 8
	//wordsPerWord := 512 / n
	return nil, nil
}

func (mt *miningTest) createDigest() (hash.Hash, error) {
	h, err := newHash(mt.n, mt.k)
	if err != nil {
		return nil, err
	}
	err = writeHashU32(h, uint32(mt.nonce))
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (mt *miningTest) header() []byte {
	nonce := writeU32(uint32(mt.nonce))
	tail := make([]byte, 28)
	return append(mt.I, append(nonce, tail...)...)
}

func intSlicesCmp(a, b []int) bool {

	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func newMiningParams(n, k int, I []byte, nonce int, solns [][]int) miningTest {
	return miningTest{n, k, I, nonce, solns}
}

func hashStr(hash []byte) string {
	return hex.EncodeToString(hash)
}

func sliceMemoryEq(a, b []byte) bool {
	return &a[cap(a)-1] == &b[cap(b)-1]
}

func valErr(x, y string, i int) error {
	txt := x + " != " + y + " at " + strconv.Itoa(i)
	return errors.New(txt)
}

func intSliceEq(x, y []int) error {
	for i, v := range x {
		if v != y[i] {
			return valErr(strconv.Itoa(v), strconv.Itoa(y[i]), i)
		}
	}
	return nil
}

func byteSliceEq(a, b []byte) error {
	if len(a) != len(b) {
		return errors.New("a and b not same len")
	}
	for i, val := range a {
		if val != b[i] {
			av, bv := strconv.Itoa(int(val)), strconv.Itoa(int(b[i]))
			return valErr(av, bv, i)
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

func testExpandCompressArray(t *testing.T, p expandCompressTest) {
	outLen := len(p.expanded)
	expanded, err := expandArray(p.compact, outLen, p.bitLen, p.bytePad)
	if err != nil {
		t.Error(err)
	}
	err = byteSliceEq(expanded, p.expanded)
	if err != nil {
		t.Error(err)
	}
	compact, err := compressArray(expanded, len(p.compact), p.bitLen, p.bytePad)
	if err != nil {
		t.Error(err)
	}
	err = byteSliceEq(p.compact, compact)
	if err != nil {
		t.Error(err)
	}
}

func TestExpandCompressArrays(t *testing.T) {
	for _, params := range expandCompressTests {
		testExpandCompressArray(t, params)
	}
}

func TestDistinctIndices(t *testing.T) {
	a := []int{0, 1, 2, 3, 4, 5}
	b := []int{0, 1, 2, 3, 4, 5}
	r := distinctIndices(a, b)
	if r {
		t.Error()
	}
	b = []int{6, 7, 8, 9, 10}
	r = distinctIndices(a, b)
	if !r {
		t.Error()
	}
	a = []int{7, 8, 9, 10, 11}
	r = distinctIndices(a, b)
	if r {
		t.Error()
	}
}

func TestHasCollision(t *testing.T) {
	h := make([]byte, 32)
	r := hasCollision(h, h, 1, len(h))
	if r == false {
		t.Fail()
	}
}

func TestHasCollision_AStartPos(t *testing.T) {
	ha, hb := []byte{}, []byte{1, 2, 3, 4, 5}
	r := hasCollision(ha, hb, 1, 0)
	if r {
		t.Errorf("r = %v\n", r)
	}
}

func TestHasCollision_BStartPos(t *testing.T) {
	hb, ha := []byte{}, []byte{1, 2, 3, 4, 5}
	r := hasCollision(ha, hb, 1, 0)
	if r {
		t.Errorf("r = %v\n", r)
	}
}

func TestHasCollision_HashLen(t *testing.T) {
	hb, ha := []byte{}, []byte{1, 2, 3, 4, 5}
	r := hasCollision(ha, hb, 1, 0)
	if r {
		t.Fail()
	}
}

func TestPow(t *testing.T) {
	exp := 1
	for i := 0; i < 64; i++ {
		val := pow(i)
		if exp != val {
			t.Errorf("binPowInt(%v) == %v and not %v\n", i, val, exp)
		}
		exp *= 2
	}
}

func TestBinPowInt_NegIndices(t *testing.T) {
	for i := 0; i < 64; i++ {
		k := -1 - i
		if pow(k) != 1 {
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
	a := make([]int, 1, 5)
	b := make([]int, 1, 10)
	small, large := minSlices(a, b)
	err := intSliceEq(a, small)
	if err != nil {
		t.Error(err)
	}
	err = intSliceEq(b, large)
	if err != nil {
		t.Error(err)
	}
}

func TestMinSlices_B(t *testing.T) {
	a := make([]int, 1, 10)
	b := make([]int, 1, 5)
	small, large := minSlices(a, b)
	err := intSliceEq(b, small)
	if err != nil {
		t.Error(err)
	}
	err = intSliceEq(a, large)
	if err != nil {
		t.Error(err)
	}
}

func TestMinSlices_Eq(t *testing.T) {
	a := make([]int, 1, 5)
	b := make([]int, 1, 5)
	small, large := minSlices(a, b)
	err := intSliceEq(a, small)
	if err != nil {
		t.Error(err)
	}
	err = intSliceEq(b, large)
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

func TestHashSize(t *testing.T) {
	if 64 != hashSize {
		t.Errorf("hashSize should equal 64 and not %v\n", hashSize)
	}
}

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

func TestJoinBytes(t *testing.T) {
	a, b := []byte{1, 2, 3, 4}, []byte{5, 6, 7, 8}
	act, exp := joinBytes(a, b), []byte{1, 2, 3, 4, 5, 6, 7, 8}
	err := byteSliceEq(act, exp)
	if err != nil {
		t.Error(err)
	}
}

func hashKeyEq(x, y hashKey) bool {
	if bytes.Compare(x.hash, y.hash) != 0 {
		return false
	}
	return true
}

func solutionsEq(x, y [][]int) error {
	if len(x) != len(y) {
		return errors.New("incorrect solutions lengths")
	}
	for i, xs := range x {
		err := solutionEq(xs, y[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func solutionEq(x, y []int) error {
	return intSliceEq(x, y)
}

func TestExccPerson_2(t *testing.T) {
	p := exccPerson(N, K)
	n := len(personPrefix)
	err := byteSliceEq(p[:n], personPrefix)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = byteSliceEq(p[n:n+4], writeU32(uint32(N)))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = byteSliceEq(p[n+4:n+8], writeU32(uint32(K)))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestWriteU32(t *testing.T) {
	exp := []byte{1, 0, 0, 0}
	err := byteSliceEq(writeU32(1), exp)
	if err != nil {
		t.Error(err)
	}
}

func TestPutU32(t *testing.T) {
	exp, act := []byte{1, 0, 0, 0}, make([]byte, 4)
	putU32(act, 1)
	err := byteSliceEq(act, exp)
	if err != nil {
		t.Error(err)
	}
}

func testHashKeys(n int) []hashKey {
	keys := make([]hashKey, 0, n)
	hashLen := 4
	for i := 0; i < n; i++ {
		digest := make([]byte, hashLen)
		for j := 0; j < hashLen; j++ {
			v := rand.Intn(math.MaxInt8)
			digest = append(digest, byte(v))
		}
		keys = append(keys, hashKey{digest, []int{i}})
	}
	return keys
}

func TestSortHashKeys(t *testing.T) {
	n := 8
	keys := testHashKeys(n)
	sortHashKeys(keys)
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			x, y := keys[i].hash, keys[j].hash
			cmp := bytes.Compare(x, y)
			if cmp != -1 {
				t.Errorf("%v >= %v\n", x, y)
			}
		}
	}
}

func randDigest(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Int())
	}
	return b
}

func TestCopyHash(t *testing.T) {
	h, err := newHash(n, k)
	if err != nil {
		t.Error(err)
	}
	b := []byte{1, 2, 3, 4}
	err = writeHashBytes(h, b)
	if err != nil {
		t.Error(err)
	}
	exp := hashDigest(h)
	h = copyHash(h)
	act := hashDigest(h)
	if bytes.Compare(exp, act) != 0 {
		t.Error("digests are not equal")
	}
}

func testCompressArray(t *testing.T, p expandCompressTest) error {
	bitLen, bytePad := p.bitLen, p.bytePad
	act, err := compressArray(p.expanded, len(p.compact), bitLen, bytePad)
	if err != nil {
		t.Error(err)
	}
	return byteSliceEq(p.compact, act)
}

func TestCompressArray(t *testing.T) {
	for _, p := range expandCompressTests {
		err := testCompressArray(t, p)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

func testExpandArray(t *testing.T, p expandCompressTest) error {
	outLen, bitLen, bytePad := len(p.expanded), p.bitLen, p.bytePad
	act, err := expandArray(p.compact, outLen, bitLen, bytePad)
	if err != nil {
		t.Error(err)
	}
	return byteSliceEq(p.expanded, act)
}

func TestExpandArray(t *testing.T) {
	for _, p := range expandCompressTests {
		err := testExpandArray(t, p)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

func testValidateCalculationSolution(t *testing.T, m miningTest) error {
	n, k := m.n, m.n
	header := m.header()
	person := exccPerson(n, k)
	for _, solution := range m.solutions {
		r, err := ValidateSolution(n, k, person, header, solution)
		if err != nil {
			return err
		}
		if !r {
			t.Errorf("solution %v is not valid\n", solution)
		}
	}
	return nil
}

func TestValidateCalculationSolution(t *testing.T) {
	for _, test := range miningTests {
		err := testValidateCalculationSolution(t, test)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

	}
}

func testSolveGBP(t *testing.T, test miningTest) error {
	digest, err := test.createDigest()
	if err != nil {
		return err
	}
	n, k := test.n, test.k
	solutions, err := gbp(digest, n, k)
	if err != nil {
		return err
	}
	for i, solution := range solutions {
		err := solutionEq(solution, test.solutions[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func TestSolveGBP(t *testing.T) {
	for _, test := range miningTests {
		err := testSolveGBP(t, test)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

func TestExccPerson(t *testing.T) {
	person := exccPerson(N, K)
	exp := []byte{90, 99, 97, 115, 104, 80, 111, 87, 96, 0, 0, 0, 5, 0, 0, 0}
	if bytes.Compare(person, exp) != 0 {
		t.Errorf("%v != %v\n", person, exp)
	}
}

func TestBlake2bPerson(t *testing.T) {
	size := N / 8
	c := &blake2b.Config{
		Key:    nil,
		Person: exccPerson(N, K),
		Salt:   nil,
		Size:   uint8(size),
	}
	h, err := blake2b.New(c)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	hash := hashDigest(h)
	exp := []byte{20, 36, 1, 103, 212, 8, 139, 129, 145, 123, 113, 170}
	if bytes.Compare(hash, exp) != 0 {
		fmt.Printf("%v != %v\n", hash, exp)
	}
}
