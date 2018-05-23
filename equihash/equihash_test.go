package consensus

import (
	"bytes"
	"encoding/hex"
	"errors"
	"hash"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"testing"

	blake2b "github.com/minio/blake2b-simd"
)

const (
	prefix = "ZcashPoW"
)

var (
	expandCompressTests = createExpandCompressTests()
	miningTests         = createMiningTests()
	validationTests     = createValidationTests()
)

//compressArray compresses (shrinks) an array
// it is the reverse function of expandArray
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
				accVal = accVal | b
			}
			j += inWidth
			accBits += bitLen
		}
		accBits -= 8
		out[i] = byte((accVal >> uint(accBits)) & 0xFF)
	}

	return out, nil
}

func createExpandCompressTests() []expandCompressTest {
	return []expandCompressTest{
		{11, 0, decodeHex("ffffffffffffffffffffff"), decodeHex("07ff07ff07ff07ff07ff07ff07ff07ff")},
		{21, 0, decodeHex("aaaaad55556aaaab55555aaaaad55556aaaab55555"), decodeHex("155555155555155555155555155555155555155555155555")},
		{21, 0, decodeHex("000220000a7ffffe00123022b38226ac19bdf23456"), decodeHex("0000440000291fffff0001230045670089ab00cdef123456")},
		{14, 0, decodeHex("cccf333cccf333cccf333cccf333cccf333cccf333cccf333cccf333"), decodeHex("3333333333333333333333333333333333333333333333333333333333333333")},
		{11, 2, decodeHex("ffffffffffffffffffffff"), decodeHex("000007ff000007ff000007ff000007ff000007ff000007ff000007ff000007ff")},
	}
}

func createValidationTests() []validationTest {
	return []validationTest{
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568}, true},
		// Change one index
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{2262, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568}, false},
		// Swap two arbitrary indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{45858, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 2261, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568}, false},
		// Reverse the first pair of indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{15185, 2261, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568}, false},
		// Swap the first and second pairs of indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{36112, 104243, 2261, 15185, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568}, false},
		// Swap the second-to-last and last pairs of indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 104973, 122568, 69567, 114474}, false},
		// Swap the first half and second half
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568, 2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026}, false},
		// Sort the indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{2261, 15185, 15972, 23460, 23779, 32642, 36112, 45858, 49807, 52426, 68190, 69567, 69878, 76925, 80080, 80391, 81830, 85191, 90330, 91132, 92842, 104243, 104973, 111026, 114474, 115059, 116805, 118332, 118390, 122568, 122819, 130041}, false},
		// Duplicate indices
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{2261, 2261, 15185, 15185, 36112, 36112, 104243, 104243, 23779, 23779, 118390, 118390, 118332, 118332, 130041, 130041, 32642, 32642, 69878, 69878, 76925, 76925, 80080, 80080, 45858, 45858, 116805, 116805, 92842, 92842, 111026, 111026}, false},
		// Duplicate first half
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, []int{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026}, false},
		{96, 5, []byte("block header"), 1, []int{1911, 96020, 94086, 96830, 7895, 51522, 56142, 62444, 15441, 100732, 48983, 64776, 27781, 85932, 101138, 114362, 4497, 14199, 36249, 41817, 23995, 93888, 35798, 96337, 5530, 82377, 66438, 85247, 39332, 78978, 83015, 123505}, true},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 2, []int{6005, 59843, 55560, 70361, 39140, 77856, 44238, 57702, 32125, 121969, 108032, 116542, 37925, 75404, 48671, 111682, 6937, 93582, 53272, 77545, 13715, 40867, 73187, 77853, 7348, 70313, 24935, 24978, 25967, 41062, 58694, 110036}, true},
		{200, 9, []byte("block header"), 0, []int{4313, 223176, 448870, 1692641, 214911, 551567, 1696002, 1768726, 500589, 938660, 724628, 1319625, 632093, 1474613, 665376, 1222606, 244013, 528281, 1741992, 1779660, 313314, 996273, 435612, 1270863, 337273, 1385279, 1031587, 1147423, 349396, 734528, 902268, 1678799, 10902, 1231236, 1454381, 1873452, 120530, 2034017, 948243, 1160178, 198008, 1704079, 1087419, 1734550, 457535, 698704, 649903, 1029510, 75564, 1860165, 1057819, 1609847, 449808, 527480, 1106201, 1252890, 207200, 390061, 1557573, 1711408, 396772, 1026145, 652307, 1712346, 10680, 1027631, 232412, 974380, 457702, 1827006, 1316524, 1400456, 91745, 2032682, 192412, 710106, 556298, 1963798, 1329079, 1504143, 102455, 974420, 639216, 1647860, 223846, 529637, 425255, 680712, 154734, 541808, 443572, 798134, 322981, 1728849, 1306504, 1696726, 57884, 913814, 607595, 1882692, 236616, 1439683, 420968, 943170, 1014827, 1446980, 1468636, 1559477, 1203395, 1760681, 1439278, 1628494, 195166, 198686, 349906, 1208465, 917335, 1361918, 937682, 1885495, 494922, 1745948, 1320024, 1826734, 847745, 894084, 1484918, 1523367, 7981, 1450024, 861459, 1250305, 226676, 329669, 339783, 1935047, 369590, 1564617, 939034, 1908111, 1147449, 1315880, 1276715, 1428599, 168956, 1442649, 766023, 1171907, 273361, 1902110, 1169410, 1786006, 413021, 1465354, 707998, 1134076, 977854, 1604295, 1369720, 1486036, 330340, 1587177, 502224, 1313997, 400402, 1667228, 889478, 946451, 470672, 2019542, 1023489, 2067426, 658974, 876859, 794443, 1667524, 440815, 1099076, 897391, 1214133, 953386, 1932936, 1100512, 1362504, 874364, 975669, 1277680, 1412800, 1227580, 1857265, 1312477, 1514298, 12478, 219890, 534265, 1351062, 65060, 651682, 627900, 1331192, 123915, 865936, 1218072, 1732445, 429968, 1097946, 947293, 1323447, 157573, 1212459, 923792, 1943189, 488881, 1697044, 915443, 2095861, 333566, 732311, 336101, 1600549, 575434, 1978648, 1071114, 1473446, 50017, 54713, 367891, 2055483, 561571, 1714951, 715652, 1347279, 584549, 1642138, 1002587, 1125289, 1364767, 1382627, 1387373, 2054399, 97237, 1677265, 707752, 1265819, 121088, 1810711, 1755448, 1858538, 444653, 1130822, 514258, 1669752, 578843, 729315, 1164894, 1691366, 15609, 1917824, 173620, 587765, 122779, 2024998, 804857, 1619761, 110829, 1514369, 410197, 493788, 637666, 1765683, 782619, 1186388, 494761, 1536166, 1582152, 1868968, 825150, 1709404, 1273757, 1657222, 817285, 1955796, 1014018, 1961262, 873632, 1689675, 985486, 1008905, 130394, 897076, 419669, 535509, 980696, 1557389, 1244581, 1738170, 197814, 1879515, 297204, 1165124, 883018, 1677146, 1545438, 2017790, 345577, 1821269, 761785, 1014134, 746829, 751041, 930466, 1627114, 507500, 588000, 1216514, 1501422, 991142, 1378804, 1797181, 1976685, 60742, 780804, 383613, 645316, 770302, 952908, 1105447, 1878268, 504292, 1961414, 693833, 1198221, 906863, 1733938, 1315563, 2049718, 230826, 2064804, 1224594, 1434135, 897097, 1961763, 993758, 1733428, 306643, 1402222, 532661, 627295, 453009, 973231, 1746809, 1857154, 263652, 1683026, 1082106, 1840879, 768542, 1056514, 888164, 1529401, 327387, 1708909, 961310, 1453127, 375204, 878797, 1311831, 1969930, 451358, 1229838, 583937, 1537472, 467427, 1305086, 812115, 1065593, 532687, 1656280, 954202, 1318066, 1164182, 1963300, 1232462, 1722064, 17572, 923473, 1715089, 2079204, 761569, 1557392, 1133336, 1183431, 175157, 1560762, 418801, 927810, 734183, 825783, 1844176, 1951050, 317246, 336419, 711727, 1630506, 634967, 1595955, 683333, 1461390, 458765, 1834140, 1114189, 1761250, 459168, 1897513, 1403594, 1478683, 29456, 1420249, 877950, 1371156, 767300, 1848863, 1607180, 1819984, 96859, 1601334, 171532, 2068307, 980009, 2083421, 1329455, 2030243, 69434, 1965626, 804515, 1339113, 396271, 1252075, 619032, 2080090, 84140, 658024, 507836, 772757, 154310, 1580686, 706815, 1024831, 66704, 614858, 256342, 957013, 1488503, 1615769, 1515550, 1888497, 245610, 1333432, 302279, 776959, 263110, 1523487, 623933, 2013452, 68977, 122033, 680726, 1849411, 426308, 1292824, 460128, 1613657, 234271, 971899, 1320730, 1559313, 1312540, 1837403, 1690310, 2040071, 149918, 380012, 785058, 1675320, 267071, 1095925, 1149690, 1318422, 361557, 1376579, 1587551, 1715060, 1224593, 1581980, 1354420, 1850496, 151947, 748306, 1987121, 2070676, 273794, 981619, 683206, 1485056, 766481, 2047708, 930443, 2040726, 1136227, 1945705, 1722044, 1971986}, true},
	}
}

func createMiningTests() []miningTest {
	return []miningTest{
		{96, 5, []byte("block header"), 0, [][]int{
			{976, 126621, 100174, 123328, 38477, 105390, 38834, 90500, 6411, 116489, 51107, 129167, 25557, 92292, 38525, 56514, 1110, 98024, 15426, 74455, 3185, 84007, 24328, 36473, 17427, 129451, 27556, 119967, 31704, 62448, 110460, 117894},
			{1008, 18280, 34711, 57439, 3903, 104059, 81195, 95931, 58336, 118687, 67931, 123026, 64235, 95595, 84355, 122946, 8131, 88988, 45130, 58986, 59899, 78278, 94769, 118158, 25569, 106598, 44224, 96285, 54009, 67246, 85039, 127667},
			{1278, 107636, 80519, 127719, 19716, 130440, 83752, 121810, 15337, 106305, 96940, 117036, 46903, 101115, 82294, 118709, 4915, 70826, 40826, 79883, 37902, 95324, 101092, 112254, 15536, 68760, 68493, 125640, 67620, 108562, 68035, 93430},
			{3976, 108868, 80426, 109742, 33354, 55962, 68338, 80112, 26648, 28006, 64679, 130709, 41182, 126811, 56563, 129040, 4013, 80357, 38063, 91241, 30768, 72264, 97338, 124455, 5607, 36901, 67672, 87377, 17841, 66985, 77087, 85291},
			{5970, 21862, 34861, 102517, 11849, 104563, 91620, 110653, 7619, 52100, 21162, 112513, 74964, 79553, 105558, 127256, 21905, 112672, 81803, 92086, 43695, 97911, 66587, 104119, 29017, 61613, 97690, 106345, 47428, 98460, 53655, 109002},
		}},
		{96, 5, []byte("block header"), 1, [][]int{
			{1911, 96020, 94086, 96830, 7895, 51522, 56142, 62444, 15441, 100732, 48983, 64776, 27781, 85932, 101138, 114362, 4497, 14199, 36249, 41817, 23995, 93888, 35798, 96337, 5530, 82377, 66438, 85247, 39332, 78978, 83015, 123505},
		}},
		{96, 5, []byte("block header"), 2, [][]int{
			{165, 27290, 87424, 123403, 5344, 35125, 49154, 108221, 8882, 90328, 77359, 92348, 54692, 81690, 115200, 121929, 18968, 122421, 32882, 128517, 56629, 88083, 88022, 102461, 35665, 62833, 95988, 114502, 39965, 119818, 45010, 94889},
		}},
		{96, 5, []byte("block header"), 10, [][]int{
			{1855, 37525, 81472, 112062, 11831, 38873, 45382, 82417, 11571, 47965, 71385, 119369, 13049, 64810, 26995, 34659, 6423, 67533, 88972, 105540, 30672, 80244, 39493, 94598, 17858, 78496, 35376, 118645, 50186, 51838, 70421, 103703},
			{3671, 125813, 31502, 78587, 25500, 83138, 74685, 98796, 8873, 119842, 21142, 55332, 25571, 122204, 31433, 80719, 3955, 49477, 4225, 129562, 11837, 21530, 75841, 120644, 4653, 101217, 19230, 113175, 16322, 24384, 21271, 96965},
		}},
		{96, 5, []byte("block header"), 11, [][]int{
			{2570, 20946, 61727, 130667, 16426, 62291, 107177, 112384, 18464, 125099, 120313, 127545, 35035, 73082, 118591, 120800, 13800, 32837, 23607, 86516, 17339, 114578, 22053, 85510, 14913, 42826, 25168, 121262, 33673, 114773, 77592, 83471},
		}},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 0, [][]int{
			{3130, 83179, 30454, 107686, 71240, 88412, 109700, 114639, 10024, 32706, 38019, 113013, 18399, 92942, 21094, 112263, 4146, 30807, 10631, 73192, 22216, 90216, 45581, 125042, 11256, 119455, 93603, 110112, 59851, 91545, 97403, 111102},
			{3822, 35317, 47508, 119823, 37652, 117039, 69087, 72058, 13147, 111794, 65435, 124256, 22247, 66272, 30298, 108956, 13157, 109175, 37574, 50978, 31258, 91519, 52568, 107874, 14999, 103687, 27027, 109468, 36918, 109660, 42196, 100424},
		}},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 1, [][]int{
			{2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568},
			{16700, 46276, 21232, 43153, 22398, 58511, 47922, 71816, 23370, 26222, 39248, 40137, 65375, 85794, 69749, 73259, 23599, 72821, 42250, 52383, 35267, 75893, 52152, 57181, 27137, 101117, 45804, 92838, 29548, 29574, 37737, 113624},
		}},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 2, [][]int{
			{6005, 59843, 55560, 70361, 39140, 77856, 44238, 57702, 32125, 121969, 108032, 116542, 37925, 75404, 48671, 111682, 6937, 93582, 53272, 77545, 13715, 40867, 73187, 77853, 7348, 70313, 24935, 24978, 25967, 41062, 58694, 110036},
		}},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 10, [][]int{
			{968, 90691, 70664, 112581, 17233, 79239, 66772, 92199, 27801, 44198, 58712, 122292, 28227, 126747, 70925, 118108, 2876, 76082, 39335, 113764, 26643, 60579, 50853, 70300, 19640, 31848, 28672, 87870, 33574, 50308, 40291, 61593},
			{1181, 61261, 75793, 96302, 36209, 113590, 79236, 108781, 8275, 106510, 11877, 74550, 45593, 80595, 71247, 95783, 2991, 99117, 56413, 71287, 10235, 68286, 22016, 104685, 51588, 53344, 56822, 63386, 63527, 75772, 93100, 108542},
			{2229, 30387, 14573, 115700, 20018, 124283, 84929, 91944, 26341, 64220, 69433, 82466, 29778, 101161, 59334, 79798, 2533, 104985, 50731, 111094, 10619, 80909, 15555, 119911, 29028, 42966, 51958, 86784, 34561, 97709, 77126, 127250},
			{15465, 59017, 93851, 112478, 24940, 128791, 26154, 107289, 24050, 78626, 51948, 111573, 35117, 113754, 36317, 67606, 21508, 91486, 28293, 126983, 23989, 39722, 60567, 97243, 26720, 56243, 60444, 107530, 40329, 56467, 91943, 93737},
		}},
		{96, 5, []byte("Equihash is an asymmetric PoW based on the Generalised Birthday problem."), 11, [][]int{
			{1120, 77433, 58243, 76860, 11411, 96068, 13150, 35878, 15049, 88928, 20101, 104706, 29215, 73328, 39498, 83529, 9233, 124174, 66731, 97423, 10823, 92444, 25647, 127742, 12207, 46292, 22018, 120758, 14411, 46485, 21828, 57591},
		}},
		{96, 5, []byte("Test case with 3+-way collision in the final round."), 0x00000000000000000000000000000000000000000000000000000000000007f0, [][]int{
			{1162, 129543, 57488, 82745, 18311, 115612, 20603, 112899, 5635, 103373, 101651, 125986, 52160, 70847, 65152, 101720, 5810, 43165, 64589, 105333, 11347, 63836, 55495, 96392, 40767, 81019, 53976, 94184, 41650, 114374, 45109, 57038},
			{2321, 121781, 36792, 51959, 21685, 67596, 27992, 59307, 13462, 118550, 37537, 55849, 48994, 58515, 78703, 100100, 11189, 98120, 45242, 116128, 33260, 47351, 61550, 116649, 11927, 20590, 35907, 107966, 28779, 57407, 54793, 104108},
			{2321, 121781, 36792, 51959, 21685, 67596, 27992, 59307, 13462, 118550, 37537, 55849, 48994, 78703, 58515, 100100, 11189, 98120, 45242, 116128, 33260, 47351, 61550, 116649, 11927, 20590, 35907, 107966, 28779, 57407, 54793, 104108},
			{2321, 121781, 36792, 51959, 21685, 67596, 27992, 59307, 13462, 118550, 37537, 55849, 48994, 100100, 58515, 78703, 11189, 98120, 45242, 116128, 33260, 47351, 61550, 116649, 11927, 20590, 35907, 107966, 28779, 57407, 54793, 104108},
			{4488, 83544, 24912, 62564, 43206, 62790, 68462, 125162, 6805, 8886, 46937, 54588, 15509, 126232, 19426, 27845, 5959, 56839, 38806, 102580, 11255, 63258, 23442, 39750, 13022, 22271, 24110, 52077, 17422, 124996, 35725, 101509},
			{8144, 33053, 33933, 77498, 21356, 110495, 42805, 116575, 27360, 48574, 100682, 102629, 50754, 64608, 96899, 120978, 11924, 74422, 49240, 106822, 12787, 68290, 44314, 50005, 38056, 49716, 83299, 95307, 41798, 82309, 94504, 96161},
		}},
	}
}

type expandCompressTest struct {
	bitLen   int
	bytePad  int
	compact  []byte
	expanded []byte
}

type miningTest struct {
	n         int
	k         int
	I         []byte
	nonce     int
	solutions [][]int
}

type validationTest struct {
	n        int
	k        int
	I        []byte
	nonce    int
	solution []int
	valid    bool
}

func createDigest(n, k, nonce int, I []byte) (hash.Hash, error) {
	bytesPerWord := n / 8
	wordsPerHash := 512 / n
	size := bytesPerWord * wordsPerHash
	digest, err := blake2b.New(&blake2b.Config{
		Size:   uint8(size),
		Person: testPerson(n, k),
	})
	if err != nil {
		return nil, err
	}
	err = writeHashBytes(digest, I)
	if err != nil {
		return nil, err
	}
	err = hashNonce(digest, nonce)
	if err != nil {
		return nil, err
	}
	return digest, nil
}

func testPerson(n, k int) []byte {
	return person(prefix, n, k)
}

func (mt *miningTest) createDigest() (hash.Hash, error) {
	n, k := mt.n, mt.k
	return createDigest(mt.n, mt.k, mt.nonce, testPerson(n, k))
}

func testHeader(I []byte, nonce int) []byte {
	nb := writeU32(uint32(nonce))
	tail := make([]byte, 28)
	return append(I, append(nb, tail...)...)
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
	r := hasDistinctIndices(a, b)
	if r {
		t.Error()
	}
	b = []int{6, 7, 8, 9, 10}
	r = hasDistinctIndices(a, b)
	if !r {
		t.Error()
	}
	a = []int{7, 8, 9, 10, 11}
	r = hasDistinctIndices(a, b)
	if r {
		t.Error()
	}
}

func TestHasCollision(t *testing.T) {
	h := make([]byte, 32)
	r := hasCollision(h, h, 1, len(h))
	if !r {
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
		val := powOf2(i)
		if exp != val {
			t.Errorf("binPowInt(%v) == %v and not %v\n", i, val, exp)
		}
		exp *= 2
	}
}

func TestBinPowInt_NegIndices(t *testing.T) {
	for i := 0; i < 64; i++ {
		k := -1 - i
		if powOf2(k) != 1 {
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

func solutionEq(x, y []int) error {
	return intSliceEq(x, y)
}

func TestExccPerson_2(t *testing.T) {
	p := exccPerson(N, K)
	n := len(exccPrefix)
	err := byteSliceEq(p[:n], []byte(exccPrefix))
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

func generateHashKeys(n int) []hashKey {
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
	keys := generateHashKeys(n)
	sort.Sort(hashKeys(keys))
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			x, y := keys[i].hash, keys[j].hash
			if bytes.Compare(x, y) != -1 {
				t.Errorf("%v >= %v\n", x, y)
			}
		}
	}
}

func TestCopyHash(t *testing.T) {
	h, err := newHash(N, K, prefix)
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
	if !bytes.Equal(exp, act) {
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

func TestHasDuplicateIndices_EmptySlice(t *testing.T) {
	if hasDuplicateIndices([]int{}) {
		t.FailNow()
	}
}

func TestHasDuplicateIndices_Pass(t *testing.T) {
	in := []int{}
	for i := 0; i <= 10; i++ {
		in = append(in, i)
		if hasDuplicateIndices(in) {
			t.FailNow()
		}
	}
}

func TestHasDuplicateIndices_Fail(t *testing.T) {
	in := []int{}
	for i := 0; i <= 10; i++ {
		for j := 0; j < 2; j++ {
			in = append(in, i)
		}
		if !hasDuplicateIndices(in) {
			t.FailNow()
		}
	}
}

func TestIsWordZero(t *testing.T) {
	word := big.NewInt(0)
	if !isBigIntZero(word) {
		t.FailNow()
	}
	word = big.NewInt(1)
	if isBigIntZero(word) {
		t.FailNow()
	}
}

func testValidatePreparedSolution(t *testing.T, v validationTest) error {
	n, k := v.n, v.k
	header := testHeader(v.I, v.nonce)
	person := testPerson(n, k)
	solution := v.solution
	r, err := ValidateSolution(n, k, person, header, solution, prefix)
	if err != nil {
		return err
	}
	if r != v.valid {
		return errors.New("expected solution")
	}
	return nil
}

func TestValidatePreparedSolutions(t *testing.T) {
	for _, test := range validationTests[:1] {
		err := testValidatePreparedSolution(t, test)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

func testValidateCalculationSolution(t *testing.T, m miningTest) {
	n, k, I, nonce, solns := m.n, m.k, m.I, m.nonce, m.solutions
	p, header := testPerson(n, k), testHeader(I, nonce)
	for _, soln := range solns {
		r, err := ValidateSolution(n, k, p, header, soln, prefix)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if !r {
			t.FailNow()
		}
	}
}

func TestValidateCalculationSolutions(t *testing.T) {
	for _, test := range miningTests {
		testValidateCalculationSolution(t, test)
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
	for _, test := range miningTests[:1] {
		err := testSolveGBP(t, test)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

func TestPersonal(t *testing.T) {
	p := testPerson(N, K)
	exp := []byte{90, 99, 97, 115, 104, 80, 111, 87, 96, 0, 0, 0, 5, 0, 0, 0}
	err := byteSliceEq(p, exp)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestExccPerson(t *testing.T) {
	p := exccPerson(N, K)
	exp := append([]byte(exccPrefix), []byte{96, 0, 0, 0, 5, 0, 0, 0}...)
	if !bytes.Equal(p, exp) {
		t.FailNow()
	}
}

func TestBlake2bPerson(t *testing.T) {
	size := N / 8
	c := &blake2b.Config{
		Key:    nil,
		Person: testPerson(N, K),
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
	err = byteSliceEq(hash, exp)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestValidateSolution_SmallN(t *testing.T) {
	k, person, header, solns := 5, []byte{}, []byte{}, []int{}
	for n := 0; n < 2; n++ {
		_, err := ValidateSolution(n, k, person, header, solns, prefix)
		if err == nil {
			t.FailNow()
		}
	}
}

func TestValidateSolution_SmallK(t *testing.T) {
	n, person, header, solns := 96, []byte{}, []byte{}, []int{}
	for k := 0; n < 3; n++ {
		_, err := ValidateSolution(n, k, person, header, solns, prefix)
		if err == nil {
			t.FailNow()
		}
	}
}

func TestValidateSolution_NMod8(t *testing.T) {
	n, person, header, solns := 96, []byte{}, []byte{}, []int{}
	for _, k := range []int{4, 8, 16, 32, 64} {
		_, err := ValidateSolution(n, k, person, header, solns, prefix)
		if err == nil {
			t.FailNow()
		}
	}
}

func TestValidateSolution_NModK(t *testing.T) {
	n, person, header, solns := 96, []byte{}, []byte{}, []int{}
	for _, k := range []int{3, 7, 15, 31, 63} {
		_, err := ValidateSolution(n, k, person, header, solns, prefix)
		if err == nil {
			t.FailNow()
		}
	}
}

func TestValidateSolution_EmptySolutionSize(t *testing.T) {
	I := []byte("block header")
	n, k, person, header, solns := N, K, testPerson(N, K), testHeader(I, 1), []int{}
	_, err := ValidateSolution(n, k, person, header, solns, prefix)
	if err == nil {
		t.FailNow()
	}
}

/*
func TestMine(t *testing.T) {
	n, k, d, p := N, K, 1, prefix
	res, err := Mine(n, k, d, p)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if res.nonce < 0 {
		t.FailNow()
	}
	if len(res.currHash) == 0 {
		t.FailNow()
	}
	if len(res.previousHash) == 0 {
		t.FailNow()
	}
}
*/
