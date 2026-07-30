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
// the algorithm transition and the first historical invalid solution.
func TestValidateEquihashSolution(t *testing.T) {
	params := chaincfg.MainNetParams()
	tests := []struct {
		name    string
		header  string
		invalid bool
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
		name:    "historical invalid solution",
		header:  "09000000aabad8ffdae7c37f33616e4dae133354718f73f7c55315fbc3f3d3a9533a0306fc3ea8340d7b6cb357585e83ef7367000eddc87aa6cb54fd470b04c083f37b2ac789abc97707099649fba6aa862c25d09c6109f60894b1282af8f4cdbb859a520100d76534d306810400000077c10000fa1b0c20c11afbfd00000000ec820f008c7900001f4a9364000000007f13d9a3ffb139cd00000000000000000000000000000000000000000000000009000000070165bccad80639c78af161f0e608738125c5f7b0295a53bf1d9436a68a1b5fa8997ece10c2c0f802a4a5a1450355cc6fe2195d3f279f4ae23e7dbd98616484db9accbd9e6ba75f7a8c6b7114fee3f6df69224a957252eb4c8dd5d7cf031169ddb96ed3",
		invalid: true,
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

		err = CheckProofOfWork(&header, header.Bits, params)
		if (err != nil) != test.invalid {
			t.Errorf("%s: error = %v, want invalid %v", test.name, err,
				test.invalid)
		}
		if test.invalid && !errors.Is(err, ErrInvalidEquihashSolution) {
			t.Errorf("%s: error = %v, want %v", test.name, err,
				ErrInvalidEquihashSolution)
		}
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
