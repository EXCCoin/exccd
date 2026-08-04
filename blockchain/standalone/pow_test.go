// Copyright (c) 2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package standalone

import (
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/wire"
)

// TestHashToBig ensures HashToBig properly converts a hash treated as a little
// endian unsigned 256-bit value to a big integer encoded with big endian.
func TestHashToBig(t *testing.T) {
	tests := []struct {
		name string // test description
		hash string // hash to convert
		want string // expected bit integer bytes in hex
	}{{
		name: "mainnet block 1 hash",
		hash: "000000000000437482b6d47f82f374cde539440ddb108b0a76886f0d87d126b9",
		want: "000000000000437482b6d47f82f374cde539440ddb108b0a76886f0d87d126b9",
	}, {
		name: "mainnet block 2 hash",
		hash: "000000000000c41019872ff7db8fd2e9bfa05f42d3f8fee8e895e8c1e5b8dcba",
		want: "000000000000c41019872ff7db8fd2e9bfa05f42d3f8fee8e895e8c1e5b8dcba",
	}}

	for _, test := range tests {
		hash, err := chainhash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("%q: unexpected err parsing test hash: %v", test.name, err)
			continue
		}

		want, success := new(big.Int).SetString(test.want, 16)
		if !success {
			t.Errorf("%q: unexpected err parsing test result", test.name)
			continue
		}

		result := HashToBig(hash)
		if result.Cmp(want) != 0 {
			t.Errorf("%s: unexpected result -- got %x, want %x", test.name,
				result, want)
			continue
		}
	}
}

// TestBigToCompact ensures converting from big integers to the compact
// representation used for target difficulties produces the correct results.
func TestBigToCompact(t *testing.T) {
	tests := []struct {
		name  string // test description
		input string // big integer to test
		want  uint32 // expected compact value
	}{{
		name:  "mainnet block 1",
		input: "0x000000000001ffff000000000000000000000000000000000000000000000000",
		want:  0x1b01ffff,
	}, {
		name:  "mainnet block 288",
		input: "0x000000000001330e000000000000000000000000000000000000000000000000",
		want:  0x1b01330e,
	}, {
		name:  "higher diff (exponent 24, sign bit 0, mantissa 0x5fb28a)",
		input: "0x00000000000000005fb28a000000000000000000000000000000000000000000",
		want:  0x185fb28a,
	}, {
		name:  "zero",
		input: "0",
		want:  0,
	}, {
		name:  "-1 (exponent 1, sign bit 1, mantissa 0x10000)",
		input: "-1",
		want:  0x1810000,
	}, {
		name:  "-128 (exponent 2, sign bit 1, mantissa 0x08000)",
		input: "-128",
		want:  0x2808000,
	}, {
		name:  "-32768 (exponent 3, sign bit 1, mantissa 0x08000)",
		input: "-32768",
		want:  0x3808000,
	}, {
		name:  "-8388608 (exponent 4, sign bit 1, mantissa 0x08000)",
		input: "-8388608",
		want:  0x4808000,
	}}

	for _, test := range tests {
		input, success := new(big.Int).SetString(test.input, 0)
		if !success {
			t.Errorf("%q: unexpected err parsing test input", test.name)
			continue
		}

		result := BigToCompact(input)
		if result != test.want {
			t.Errorf("%q: mismatched result -- got %x, want %x", test.name,
				result, test.want)
			continue
		}
	}
}

// TestCompactToBig ensures converting from the compact representation used for
// target difficulties to big integers produces the correct results.
func TestCompactToBig(t *testing.T) {
	tests := []struct {
		name  string // test description
		input uint32 // compact target difficulty bits to test
		want  string // expected big int
	}{{
		name:  "mainnet block 1",
		input: 0x1b01ffff,
		want:  "0x000000000001ffff000000000000000000000000000000000000000000000000",
	}, {
		name:  "mainnet block 288",
		input: 0x1b01330e,
		want:  "0x000000000001330e000000000000000000000000000000000000000000000000",
	}, {
		name:  "higher diff (exponent 24, sign bit 0, mantissa 0x5fb28a)",
		input: 0x185fb28a,
		want:  "0x00000000000000005fb28a000000000000000000000000000000000000000000",
	}, {
		name:  "zero",
		input: 0,
		want:  "0",
	}, {
		name:  "-1 (exponent 1, sign bit 1, mantissa 0x10000)",
		input: 0x1810000,
		want:  "-1",
	}, {
		name:  "-128 (exponent 2, sign bit 1, mantissa 0x08000)",
		input: 0x2808000,
		want:  "-128",
	}, {
		name:  "-32768 (exponent 3, sign bit 1, mantissa 0x08000)",
		input: 0x3808000,
		want:  "-32768",
	}, {
		name:  "-8388608 (exponent 4, sign bit 1, mantissa 0x08000)",
		input: 0x4808000,
		want:  "-8388608",
	}}

	for _, test := range tests {
		want, success := new(big.Int).SetString(test.want, 0)
		if !success {
			t.Errorf("%q: unexpected err parsing expected value", test.name)
			continue
		}

		result := CompactToBig(test.input)
		if result.Cmp(want) != 0 {
			t.Errorf("%q: mismatched result -- got %x, want %x", test.name,
				result, want)
			continue
		}
	}
}

// TestCalcWork ensures calculating a work value from a compact target
// difficulty produces the correct results.
func TestCalcWork(t *testing.T) {
	tests := []struct {
		name  string // test description
		input uint32 // compact target difficulty bits to test
		want  string // expected big int
	}{{
		name:  "mainnet block 1",
		input: 0x1b01ffff,
		want:  "0x0000000000000000000000000000000000000000000000000000800040002000",
	}, {
		name:  "mainnet block 288",
		input: 0x1b01330e,
		want:  "0x0000000000000000000000000000000000000000000000000000d56f2dcbe105",
	}, {
		name:  "higher diff (exponent 24)",
		input: 0x185fb28a,
		want:  "0x000000000000000000000000000000000000000000000002acd33ddd458512da",
	}, {
		name:  "zero",
		input: 0,
		want:  "0",
	}, {
		name:  "negative target difficulty",
		input: 0x1810000,
		want:  "0",
	}}

	for _, test := range tests {
		want, success := new(big.Int).SetString(test.want, 0)
		if !success {
			t.Errorf("%q: unexpected err parsing expected value", test.name)
			continue
		}

		result := CalcWork(test.input)
		if result.Cmp(want) != 0 {
			t.Errorf("%q: mismatched result -- got %x, want %x", test.name,
				result, want)
			continue
		}
	}
}

// mockMainNetPowLimit returns the pow limit for the main network as of the
// time this comment was written.  It is used to ensure the tests are stable
// independent of any potential changes to chain parameters.
func mockMainNetPowLimit() string {
	return "00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
}

// TestValidateEquihashSolution checks known mainnet headers on both sides of
// the algorithm transition and the first historical exception.
func TestValidateEquihashSolution(t *testing.T) {
	params := chaincfg.MainNetParams()
	tests := []struct {
		name   string
		header string
	}{{
		name:   "last algorithm v0 block",
		header: "05000000a06af4a3c604b72842ac9e05edc264d9f3ed399fb65973ec38755b6be8010000885740afd78a2dac6cf9510342870956ef041bd0e3d526c488d9bad1436f7c24c8649f590cce2216f34f3a29708b3d8d44b5bc051594f021437d1deff913184001007a765f3b2c1c0500000131ac00004d74021e9b91f39826000000fd550100ec0c00008aa9175cd70cd8006e2c7bb537aa50e291432728475ec70fe28a3e4cfe4258544c661f8042c4603705000000000df91619997bd7247f3f85821656dba5da45b846db853c36353b3d3282fb94056af6b24f2401422a2da9293b6e4fade452045c8cb4787b2712499a1614990de5b4e343ae704df19169a605d63abb1db8e931163ab122fc0ed3f62e790f1c02e9cee765",
	}, {
		name:   "first algorithm v1 block",
		header: "050000001efdc5d33aa6ee6bb7146507b0750180e44773479f0b35f47d70ea7f31000000c3b7853f8d4b1870b3a597450c24befb8b301e1e61741a3929cb9bb096d3461dc7f17e6f257b3a0df6fb609817cfe340dd4980cb14c5a4e61648017d25277dce0100f274708796e1050001002cac0000ffff011f9b91f39826000000fe550100160b00005bab175c14bf00000000000000000000bf79b7e400000000000000000000000000000000000000000500000006185840935dd40f87f281e90404c872674a45fd38afc2d857082ac6c52ba69d3144feb7aa2356753eafa14e95244be250312325f323926e5ed4c43ab921a30cd514ade979a11fd760dce25a7907e9de871716ceaba885c6753484c0308f8cba13e6e473",
	}, {
		name:   "block before invalid solution",
		header: "090000002c7fd369b5b9b8635cab6d82845b34f10aa501b3b649ad6b2b011b34c26510031fb3a996194aff7b5da29f275ac8a9f23950927722ee55e2a3a5f3a7b5ad1f97f31745f7b817c5e55be9bb39354cac4099c78ec69e591f4c7fe62b80199a6255010028adb8ec546c0500000072c10000fa1b0c20c11afbfd00000000eb820f0063080000534993640000000003968f15912fb144000000000000000000000000000000000000000000000000090000001e14effcd02cc8a02614ab99f3ce580e76c8f215318f727bc64748d0329e2b7dab7f9ef02887e8fd1de9c9a3db22b9fa403d26707adef8130d4481cd4252f302278b36740e45e35d7374e946821eee3a57a03baf953f5d45b3f11ae7f7d57a23331825ff",
	}, {
		name:   "first historical exception",
		header: "09000000aabad8ffdae7c37f33616e4dae133354718f73f7c55315fbc3f3d3a9533a0306fc3ea8340d7b6cb357585e83ef7367000eddc87aa6cb54fd470b04c083f37b2ac789abc97707099649fba6aa862c25d09c6109f60894b1282af8f4cdbb859a520100d76534d306810400000077c10000fa1b0c20c11afbfd00000000ec820f008c7900001f4a9364000000007f13d9a3ffb139cd00000000000000000000000000000000000000000000000009000000070165bccad80639c78af161f0e608738125c5f7b0295a53bf1d9436a68a1b5fa8997ece10c2c0f802a4a5a1450355cc6fe2195d3f279f4ae23e7dbd98616484db9accbd9e6ba75f7a8c6b7114fee3f6df69224a957252eb4c8dd5d7cf031169ddb96ed3",
	}}

	for _, test := range tests {
		headerBytes, err := hex.DecodeString(test.header)
		if err != nil {
			t.Fatalf("%s: decode header: %v", test.name, err)
		}
		var header wire.BlockHeader
		if err := header.FromBytes(headerBytes); err != nil {
			t.Fatalf("%s: deserialize header: %v", test.name, err)
		}

		if err := CheckProofOfWorkHash(&header, header.Bits, params); err != nil {
			t.Errorf("%s: hash proof of work: %v", test.name, err)
			continue
		}

		if err := CheckProofOfWork(&header, header.Bits, params); err != nil {
			t.Errorf("%s: proof of work: %v", test.name, err)
			continue
		}
	}
}

// TestHistoricalEquihashExceptionHashes locks the complete, mainnet-only set
// of historical exceptions.
func TestHistoricalEquihashExceptionHashes(t *testing.T) {
	hashes := []string{
		"080dca3e1603a4ce3fdb6eb7e86fb460d95ec46df7132b80f8a5547ad877e732",
		"0b62ebec9ac835c8c31ddc4cc2c6eff47444c9de47fbffa43b1acc25b2f45a13",
		"028653b73e92f6d4984ae0af271efa3dfd3e7aa02a80a506995438e4a8eb4afb",
		"000009016dc87ba964b819dcf6612cba4a2439bc62c560aede577336a7804897",
		"00000fe85253156a3ec146cdde074504a57ddb065588845cfcccba851137b69e",
		"0000143be1d1bfd29a3e7de6c49dc0754385896ec14cf52d5320c45bc7c728f2",
		"0000135ecf086f7420f5fb8789146601adef22cb7dfb76bc9c3d5764c800a7c7",
		"00000304f1b93f032e1c4dd57c5ea53dd5c8b6ef5d5122b1f4a1b693c17b3094",
		"000010140fe77f88c96178e58fff159a0cdc13abee894cdee6369dff37887963",
		"000005bbd1590593d2953b3dd2bf38f64b324b12fcc3701844efd006d636ce3b",
		"00000a7ce7d66e87ce2706644e77acf7ebe16d2e98b680b598dbacdfc456cc3e",
		"000012a78f34078a0c6619d66763684d3d90a291e804443b7267aedf8552b7bb",
		"00000bce885cf7e61d95fc74d80178d09e4e51105121994130fc6609427f56cd",
	}
	headers := []string{
		"09000000aabad8ffdae7c37f33616e4dae133354718f73f7c55315fbc3f3d3a9533a0306fc3ea8340d7b6cb357585e83ef7367000eddc87aa6cb54fd470b04c083f37b2ac789abc97707099649fba6aa862c25d09c6109f60894b1282af8f4cdbb859a520100d76534d306810400000077c10000fa1b0c20c11afbfd00000000ec820f008c7900001f4a9364000000007f13d9a3ffb139cd00000000000000000000000000000000000000000000000009000000070165bccad80639c78af161f0e608738125c5f7b0295a53bf1d9436a68a1b5fa8997ece10c2c0f802a4a5a1450355cc6fe2195d3f279f4ae23e7dbd98616484db9accbd9e6ba75f7a8c6b7114fee3f6df69224a957252eb4c8dd5d7cf031169ddb96ed3",
		"0900000032e777d87a54a5f8802b13f76dc45ed960b46fe8b76edb3fcea403163eca0d08dcb3da62749298de04e4b87121afa8e0abc7b5a0cd93fda611e388f8c8c6cb8fd460dc3d5fd8ea8312820ff791148de756dbd23eaf054f5a8e62e24e49d68ccf0100a00ea9dfafae0400140072c10000fa1b0c20c11afbfd00000000ed820f00a4210000ea4a9364010000003658e8ac0e9a804f000000000000000000000000000000000000000000000000090000000e9faa4e7a8b83c3e4e567b2a23ea3879c79a1a485d6d422361926c816e434d03f53fa7f4609bc39279589778acd19d87d7e3d92a4aed679a4b4257b143a3485744bbf0011faa281aae7134ceef9abdb9fd5991437e1b384dd886c77f196613cfdc6081a",
		"09000000135af4b225cc1a3ba4fffb47dec94474f4efc6c24cdc1dc3c835c89aeceb620ba185ef58dd163f71e1d345d9f6edd1bef48a1da85f2c472cd4cae161be94a58937b414c87b5db0b09ce32f63a90dbe8ed757a664e6f805736fcfa4f782bc40920100d12b4832793704000f0077c10000fa1b0c20c11afbfd00000000ee820f00792c00002a4b9364000000008b16ec5e68f6a02e00000000000000000000000000000000000000000000000009000000022c38a669d724b52f7a45de40c0dc8358cb4559e9cd13fd861470815b18c14a136498e59e1235d35c53776520f797def9960bf662ab24c2ce73d89fa48704ff4a7bd1e97d611b1110054e0f7d2f321c16d9d1a43feefb71eced4d68c111f132ffa6219c",
		"090000005319d8fd543ba866fa425806c08677a124a9ebd75a3969871561596aa0150000ed4ad42462301f9f14c98c88f95e0c765a6e2520c93fee8fba85eb7d9f74435bd2ac09b20f1c2af4bce755287858d16a74f31683fd1aa2f68bda45f0ffcf80400100c0d11cb0ce04050001001b9d0000db27181e2672460c00000000476110008e150000f8ca1465b78b9a22c75e02000000000000000000000000000000000000000000000000000000000009000000101d9538de270e50fd341298f67b5063f9296e01107524094633a5ed2ff2a41cb99cd58e6424bf420314a7fd616a94ecbf0f146ceabcc5ac45b0b988998e2380c52a00709a801809c5399b6cbcfe3b6b2129e6d45a5ff7286007be0f113adbc391f8ec4e",
		"0900000038581d7e430be9ebdf9fd52751143a0e29d88a578e58d69ecf615b5f3e020000e77c37d81aeb46a80564a2a80f4bb13d7b0bd1ac5ecec41ac4746ffe0257d28904835a79d03b1e14ea905b0537bf54a1395813af4f6f24ff9a990613fc995ec9010009a90a47c6b50500070089aa00003b0b171e587aae0f0000000010631000a3170000f5c91565f8bec42c92960500000000000000000000000000000000000000000000000000000000000900000007d23bd9d8af5ba9a2fdad5f40f10fa696cb30cbdecd0033ed0b7296895703169a0691110645e5f8edb637e1bd91cd3372181220ba62a2672c8bc9390567f6b541e65c422ddb0b50f59b6617e139930225d6e1e16dae5102d01b367122359a24e792e9d1",
		"090000007f2c866d4bec1f551cb39ff467bd4dc280dfdde2110e14d6c0f5fd6f1e160000dc31bf7db8ad11b999504576168e52a78a69ca0d3f3583706ca11f2473c05d398a3f7daba4e5d838a9719130320b7f9b1167030a8fb8b95aaabb7fcc4ed89c9e010072e0580d5789050002007bab0000835b181e2157191400000000e7641000e311000085fe1665c1bc0276b85e090000000000000000000000000000000000000000000000000000000000090000000bc2380defb7cd8ad91faf78b78372364e9bbb212d199e069d434568e08d32f287595ea7f4a6d463a38b7995d090fb08c61f1a17e772fd0ccb96e76aeb3b840ee90c0bc14a023b9336b4c8650065cc9e12a92abd9cc622b7d94ebefba15642b937c71bdc",
		"090000002db25b4f4d2fa796379d64592a4a4874ad0c3f792d26922251ad5f87f80f0000174037500044ce853399becc3a296f4167838202f65acf8d3c060f66f5b9f452182d717723296dc8bbd358cdef0fe6509c730e3c9a47ce474ee6d0d79b00fa5c010004560a1a338b05000200d7a30000e600151e0f48de1600000000166810004e0d0000e8b618652795250d30cc0e00000000000000000000000000000000000000000000000000000000000900000006a0caecdb14ee25113d7f4cd174d47d2503d59cef86ee75ed0b68894bed27c78ae36d5a15d12cb77b3200ceb39cf5d9f4cd07cec4b9333423a25ad5a03c80e8d478f85c64a2510509c12808075a1f34ffd1be444ca5e3a50c02c4231d1db88ae1a656e4",
		"090000004a382470452f1932e34be1bb5051806de3d5ec951fbd82061e21e31a4f1200005b0c25cf9b853972c747ea4d5952a4fbeb70b5fdfe91f80a6a7a97b13cabe5e6ff556d43b3fe2950577a2929e0e644d86163478e2420d59d083ad6cc4664e091010099ff95f989c7050003001ca00000a287131ed9ba6d1600000000ae6a1000dc0b0000cd461a65b3ed4d4db4bd1300000000000000000000000000000000000000000000000000000000000900000005113302bcda4a06335e596176eae3fe53a9ba197d5512b3bc6c726458e5eee5b6691a3d6f6a74ef761848cf2de795c8e8150704de7e4a26d2ad9b159620c204b3228f87e1a2185ba98f28328cd9c625804cd73c97e492d56948a69b07f21c7c79d4fb49",
		"09000000346a18a3157f0e8fb60c3b4d0e21691d6391a3fecb0753bde382f0af770a0000a26ec09bed1f1eb8bf27766bd1d39bb1304e75c0f4fa3bf84677564bc5e342f70c55c6ef6fc54268c9c0f4ee671bf191c302b6312db8f0c774ed99623928f9040100dfe1403a0c2704000300b39d00004ac7131ef653121500000000486c10001d0e0000d72e1b65ac66ab8909a316000000000000000000000000000000000000000000000000000000000009000000052c76151b830327ae490efce4b1fa3bbce07dece3bbdf7af7164ee0e4c3a0c6a84303942495e1259d09b9961d6c6bf682ee144ed4518b34d44b071074c52350597f930de9cfd6335312e6334233a26c060dc679ce47d0489202ec96391a4a1f75294822",
		"09000000dae7fcacab07ff6d1e0579b53354c95238f560ea244b816dee391e0df90a00005faaa3e62f20538f24e7515d0904d595be0336592bb52b8e23af022331d15b584cae7db7a49765b0f7cbe275ec6f1c0eda30f015d5ec32147b7aa2630e8df9460100afb14271c60305000600c6a40000d5a0151efb417b1600000000ec6f10009313000016531d6548daf5e310280000000000000000000000000000000000000000000000000000000000000900000002aa8d1087f115b79ad16b1f231464bc88442d6e10c97f3fdd051c181e2ffd0a6068a5b8d048b7e9e58dbc26da4743e5ec4e0aa7884145c28d1842fa0bc1428632c6e7635ccc3020bda5f00e4d3142c9371a393d51096ed7a0d8c5a1488a1e4067e558b5",
		"09000000e363d501827529ec5372797caa0cf6411c9f3dfc65e59d6fde32756b870800005040630fb0f9bb0e497d814fb8554b80e84595d87e1ede9afb47cd2d04e604775855f111ee0ac3ee870efeddeaa7f736a07769e78d77c5a86967dcef613aaff9010022b2aedd0e600500040046a50000d5a0151efb417b160000000000701000070d0000cb5c1d657d94b427f149000000000000000000000000000000000000000000000000000000000000090000001d74ac4d010f15ef450c16c633813e0d49de19ea1deb40c0f21e5dfbf113c894a9775bb1ba426908eabd32cca244d94fd3b72051ef70f161589f1370a758e46f06cc0f1ada5521516622f133fd11ea081ce22e11b99bf26391b3129a3beece695dc7a49d",
		"090000005647a1aae85e7c76a208c3723cd47e0e04a5192e63896cca7a64af7da10e0000fc6b7009311d9bcce66123398e48994f99cf2f2debd7382bf0fc4a3828953d944e095e29e6fc71dac299cfd69bc218f56752fc8c96533f42ea584620a682804101001dd428ca508905000000cda800006a53131ea1d7601800000000ba701000ab250000b6cc1d65d77a8e07b3c80100000000000000000000000000000000000000000000000000000000000900000009d56255f8a978bb0abf456128463ec50245a3dd42bffab53c0da824dde406deb1ddd26ab38320ca35e6892242da75d4b0171163f88dc0e74fec0f4c55e3b239ad0f66d9fca86b66d569b82116df57ef71cc4a3fbf08758233df6ffdb59e52c507e0cc3b",
		"09000000208b7fad659ec4182d6ab3dedd37621d883b87b5e0485c11decfc93de9070000d7b94d27d6fb88f9e0f288f1ab2dd7c33841e4fdafcd8cb6650969c2205845fdb9619f018d4de3e5f977eaea2b0b10203f698222d94e4afb9611c733db1e0f08010035c8ab1be49305000000d3a6000022d6131e61cdc51900000000d6711000420d0000417a1e65f6f00c8d0b0b040000000000000000000000000000000000000000000000000000000000090000000b469d21ae4b066b9a8b2a5d9269c7733201fb0aa3d999b4b227d6792bfce251b30adae0d8e531c6f749d9495c44cbe6145316067321a523490278c6d8377469714c6b26ca73afdd3f78f14817e6e0471b1219212ac2ca26f127c6c8aaf3653c0dc3fd3b",
	}

	if len(headers) != len(hashes) {
		t.Fatalf("got %d headers for %d hashes", len(headers), len(hashes))
	}
	params := chaincfg.MainNetParams()
	for i, hashString := range hashes {
		hash, err := chainhash.NewHashFromStr(hashString)
		if err != nil {
			t.Fatalf("parse %s: %v", hashString, err)
		}
		headerBytes, err := hex.DecodeString(headers[i])
		if err != nil {
			t.Fatalf("decode header %s: %v", hash, err)
		}
		var header wire.BlockHeader
		if err := header.FromBytes(headerBytes); err != nil {
			t.Fatalf("deserialize header %s: %v", hash, err)
		}
		if got := header.BlockHash(); got != *hash {
			t.Fatalf("header hash = %s, want %s", got, hash)
		}
		if err := CheckProofOfWork(&header, header.Bits, params); err != nil {
			t.Errorf("historical block %s: %v", hash, err)
		}
		if !isHistoricalEquihashException(*hash, wire.MainNet) {
			t.Errorf("mainnet exception missing: %s", hash)
		}
		if isHistoricalEquihashException(*hash, wire.TestNet3) {
			t.Errorf("testnet unexpectedly accepts exception: %s", hash)
		}
		if err := ValidateEquihashSolution(&header, chaincfg.TestNet3Params()); !errors.Is(err, ErrInvalidEquihashSolution) {
			t.Errorf("testnet header %s: error = %v, want %v", hash, err,
				ErrInvalidEquihashSolution)
		}

		mutatedSolution := header
		mutatedSolution.EquihashSolution[0] ^= 1
		if err := ValidateEquihashSolution(&mutatedSolution, params); !errors.Is(err, ErrInvalidEquihashSolution) {
			t.Errorf("solution mutation %s: error = %v, want %v", hash, err,
				ErrInvalidEquihashSolution)
		}
		mutatedHeader := header
		mutatedHeader.ExtraData[0] ^= 1
		if err := ValidateEquihashSolution(&mutatedHeader, params); !errors.Is(err, ErrInvalidEquihashSolution) {
			t.Errorf("header mutation %s: error = %v, want %v", hash, err,
				ErrInvalidEquihashSolution)
		}
	}

	if isHistoricalEquihashException(chainhash.Hash{}, wire.MainNet) {
		t.Fatal("unknown hash accepted as historical exception")
	}
}

// TestCheckProofOfWorkRange ensures target difficulties that are outside of
// the acceptable ranges are detected as an error and those inside are not.
func TestCheckProofOfWorkRange(t *testing.T) {
	tests := []struct {
		name     string // test description
		bits     uint32 // compact target difficulty bits to test
		powLimit string // proof of work limit
		err      error  // expected error
	}{{
		name:     "mainnet block 1",
		bits:     0x1b01ffff,
		powLimit: mockMainNetPowLimit(),
		err:      nil,
	}, {
		name:     "mainnet block 288",
		bits:     0x1b01330e,
		powLimit: mockMainNetPowLimit(),
		err:      nil,
	}, {
		name:     "smallest allowed",
		bits:     0x1010000,
		powLimit: mockMainNetPowLimit(),
		err:      nil,
	}, {
		name:     "max allowed (exactly the pow limit)",
		bits:     0x1d00ffff,
		powLimit: mockMainNetPowLimit(),
		err:      nil,
	}, {
		name:     "zero",
		bits:     0,
		powLimit: mockMainNetPowLimit(),
		err:      ErrUnexpectedDifficulty,
	}, {
		name:     "negative",
		bits:     0x1810000,
		powLimit: mockMainNetPowLimit(),
		err:      ErrUnexpectedDifficulty,
	}, {
		name:     "pow limit + 1",
		bits:     0x1d010000,
		powLimit: mockMainNetPowLimit(),
		err:      ErrUnexpectedDifficulty,
	}}

	for _, test := range tests {
		powLimit, success := new(big.Int).SetString(test.powLimit, 16)
		if !success {
			t.Errorf("%q: unexpected err parsing test pow limit", test.name)
			continue
		}

		err := CheckProofOfWorkRange(test.bits, powLimit)
		if !errors.Is(err, test.err) {
			t.Errorf("%q: unexpected err -- got %v, want %v", test.name, err,
				test.err)
			continue
		}
	}
}
