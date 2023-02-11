// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
)

// TestGetSigOpCount tests that the GetSigOpCount function behaves as expected.
func TestGetSigOpCount(t *testing.T) {
	// This should correspond to MaxPubKeysPerMultiSig. It's intentionally
	// not referred to here so that any changes to MaxPubKeysPerMultisig
	// are flagged during tests.
	maxMultiSigOps := 20

	// Build out a script that tests every opcode not on the "special list"
	// below. These will be tested separately.
	specialOpCodes := map[byte]struct{}{
		OP_CHECKSIG:            {},
		OP_CHECKSIGVERIFY:      {},
		OP_CHECKSIGALT:         {},
		OP_CHECKSIGALTVERIFY:   {},
		OP_CHECKMULTISIG:       {},
		OP_CHECKMULTISIGVERIFY: {},
	}
	otherOpCodesScript := ""
	for i := 0; i <= 255; i++ {
		if _, ok := specialOpCodes[byte(i)]; ok {
			continue
		}
		otherOpCodesScript = fmt.Sprintf("%s 0x%.2x", otherOpCodesScript, i)
	}

	testCases := []struct {
		name              string
		script            string
		wantCount         int
		wantTreasuryCount int
	}{{
		name:              "all opcodes that dont count for a sigop",
		script:            otherOpCodesScript,
		wantCount:         0,
		wantTreasuryCount: 0,
	}, {
		name:              "all opcodes that count for a sigop",
		script:            "CHECKSIG CHECKSIGVERIFY CHECKSIGALT CHECKSIGALTVERIFY",
		wantCount:         4,
		wantTreasuryCount: 4,
	}, {
		name:              "multisig with < MaxPubKeysPerMultiSig",
		script:            "2 DATA_33 0x00{33} 0x00{33} 0x00{33} 3 OP_CHECKMULTISIG",
		wantCount:         maxMultiSigOps,
		wantTreasuryCount: maxMultiSigOps,
	}, {
		name:              "multisigverify with less than MaxPubKeysPerMultiSig",
		script:            "2 DATA_33 0x00{33} 0x00{33} 0x00{33} 3 OP_CHECKMULTISIGVERIFY",
		wantCount:         maxMultiSigOps,
		wantTreasuryCount: maxMultiSigOps,
	}, {
		name:              "valid p2pkh output",
		script:            "DUP HASH160 DATA_20 0x00{20} EQUALVERIFY CHECKSIG",
		wantCount:         1,
		wantTreasuryCount: 1,
	}, {
		name:              "valid p2sh output",
		script:            "HASH160 DATA_20 0x00{20} EQUAL",
		wantCount:         0,
		wantTreasuryCount: 0,
	}, {
		name:              "valid stake change output",
		script:            "SSTXCHANGE DUP HASH160 DATA_20 0x00{20} EQUALVERIFY CHECKSIG",
		wantCount:         1,
		wantTreasuryCount: 1,
	}}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			script := mustParseShortFormV0(tc.script)
			gotCount := GetSigOpCount(script, false)
			if gotCount != tc.wantCount {
				t.Fatalf("unexpected sigOpCount with treasury=false. "+
					"want=%d got=%d", tc.wantCount, gotCount)
			}

			gotTreasuryCount := GetSigOpCount(script, true)
			if gotTreasuryCount != tc.wantTreasuryCount {
				t.Fatalf("unexpected sigOpCount with treasury=true. "+
					"want=%d got=%d", tc.wantTreasuryCount,
					gotTreasuryCount)
			}
		})
	}
}

// TestGetPreciseSigOps ensures the more precise signature operation counting
// mechanism which includes signatures in P2SH scripts works as expected.
func TestGetPreciseSigOps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scriptSig []byte
		nSigOps   int
	}{{
		name:      "scriptSig doesn't parse",
		scriptSig: mustParseShortFormV0("PUSHDATA1 0x02"),
	}, {
		name:      "scriptSig isn't push only",
		scriptSig: mustParseShortFormV0("1 DUP"),
		nSigOps:   0,
	}, {
		name:      "scriptSig length 0",
		scriptSig: nil,
		nSigOps:   0,
	}, {
		// No script at end but still push only.
		name:      "No script at the end",
		scriptSig: mustParseShortFormV0("1 1"),
		nSigOps:   0,
	}, {
		name:      "pushed script doesn't parse",
		scriptSig: mustParseShortFormV0("DATA_2 PUSHDATA1 0x02"),
	}}

	// The signature in the p2sh script is nonsensical for the tests since
	// this script will never be executed.  What matters is that it matches
	// the right pattern. Without treasury enabled.
	pkScript := mustParseShortFormV0("HASH160 DATA_20 0x433ec2ac1ffa1b7b7d0" +
		"27f564529c57197f9ae88 EQUAL")
	for _, test := range tests {
		count := GetPreciseSigOpCount(test.scriptSig, pkScript,
			noTreasury)
		if count != test.nSigOps {
			t.Errorf("%s: expected count of %d, got %d", test.name,
				test.nSigOps, count)
		}
	}

	// The signature in the p2sh script is nonsensical for the tests since
	// this script will never be executed.  What matters is that it matches
	// the right pattern. With treasury enabled.
	pkScript = mustParseShortFormV0("HASH160 DATA_20 0x433ec2ac1ffa1b7b7d0" +
		"27f564529c57197f9ae88 EQUAL")
	for _, test := range tests {
		count := GetPreciseSigOpCount(test.scriptSig, pkScript,
			withTreasury)
		if count != test.nSigOps {
			t.Errorf("%s: expected count of %d, got %d", test.name,
				test.nSigOps, count)
		}
	}
}

// TestRemoveOpcodeByData ensures that removing data carrying opcodes based on
// the data they contain works as expected.
func TestRemoveOpcodeByData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		before []byte
		remove []byte
		err    error
		after  []byte
	}{{
		name:   "nothing to do",
		before: mustParseShortFormV0("NOP"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("NOP"),
	}, {
		name:   "simple case",
		before: mustParseShortFormV0("DATA_4 0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  nil,
	}, {
		name:   "simple case (miss)",
		before: mustParseShortFormV0("DATA_4 0x01020304"),
		remove: []byte{1, 2, 3, 5},
		after:  mustParseShortFormV0("DATA_4 0x01020304"),
	}, {
		name: "stakesubmission simple case p2pkh",
		before: mustParseShortFormV0("SSTX DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("SSTX DUP HASH160 EQUALVERIFY CHECKSIG"),
	}, {
		name: "stakesubmission simple case p2pkh (miss)",
		before: mustParseShortFormV0("SSTX DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
		remove: []byte{1, 2, 3, 4, 5},
		after: mustParseShortFormV0("SSTX DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
	}, {
		name: "stakesubmission simple case p2sh",
		before: mustParseShortFormV0("SSTX HASH160 DATA_20 0x00{16} 0x01020304 " +
			"EQUAL"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("SSTX HASH160 EQUAL"),
	}, {
		name: "stakesubmission simple case p2sh (miss)",
		before: mustParseShortFormV0("SSTX HASH160 DATA_20 0x00{16} 0x01020304 " +
			"EQUAL"),
		remove: []byte{1, 2, 3, 4, 5},
		after: mustParseShortFormV0("SSTX HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
	}, {
		name: "stakegen simple case p2pkh",
		before: mustParseShortFormV0("SSGEN DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("SSGEN DUP HASH160 EQUALVERIFY CHECKSIG"),
	}, {
		name: "stakegen simple case p2pkh (miss)",
		before: mustParseShortFormV0("SSGEN DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
		remove: []byte{1, 2, 3, 4, 5},
		after: mustParseShortFormV0("SSGEN DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
	}, {
		name: "stakegen simple case p2sh",
		before: mustParseShortFormV0("SSGEN HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("SSGEN HASH160 EQUAL"),
	}, {
		name: "stakegen simple case p2sh (miss)",
		before: mustParseShortFormV0("SSGEN HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
		remove: []byte{1, 2, 3, 4, 5},
		after: mustParseShortFormV0("SSGEN HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
	}, {
		name: "stakerevoke simple case p2pkh",
		before: mustParseShortFormV0("SSRTX DUP HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUALVERIFY CHECKSIG"),
		remove: []byte{1, 2, 3, 4},
		after:  []byte{OP_SSRTX, OP_DUP, OP_HASH160, OP_EQUALVERIFY, OP_CHECKSIG},
	}, {
		name: "stakerevoke simple case p2pkh (miss)",
		before: mustParseShortFormV0("SSRTX DUP HASH160 DATA_20 0x00{20} " +
			"EQUALVERIFY CHECKSIG"),
		remove: bytes.Repeat([]byte{0}, 21),
		after: mustParseShortFormV0("SSRTX DUP HASH160 DATA_20 0x00{20} " +
			"EQUALVERIFY CHECKSIG"),
	}, {
		name: "stakerevoke simple case p2sh",
		before: mustParseShortFormV0("SSRTX HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("SSRTX HASH160 EQUAL"),
	}, {
		name: "stakerevoke simple case p2sh (miss)",
		before: mustParseShortFormV0("SSRTX HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
		remove: []byte{1, 2, 3, 4, 5},
		after: mustParseShortFormV0("SSRTX HASH160 DATA_20 0x00{16} " +
			"0x01020304 EQUAL"),
	}, {
		// padded to keep it canonical.
		name:   "simple case (pushdata1)",
		before: mustParseShortFormV0("PUSHDATA1 0x4c 0x00{72} 0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  nil,
	}, {
		name:   "simple case (pushdata1 miss)",
		before: mustParseShortFormV0("PUSHDATA1 0x4c 0x00{72} 0x01020304"),
		remove: []byte{1, 2, 3, 5},
		after:  mustParseShortFormV0("PUSHDATA1 0x4c 0x00{72} 0x01020304"),
	}, {
		name:   "simple case (pushdata1 miss noncanonical)",
		before: mustParseShortFormV0("PUSHDATA1 0x04 0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("PUSHDATA1 0x04 0x01020304"),
	}, {
		name:   "simple case (pushdata2)",
		before: mustParseShortFormV0("PUSHDATA2 0x0001 0x00{252} 0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  nil,
	}, {
		name:   "simple case (pushdata2 miss)",
		before: mustParseShortFormV0("PUSHDATA2 0x0001 0x00{252} 0x01020304"),
		remove: []byte{1, 2, 3, 4, 5},
		after:  mustParseShortFormV0("PUSHDATA2 0x0001 0x00{252} 0x01020304"),
	}, {
		name:   "simple case (pushdata2 miss noncanonical)",
		before: mustParseShortFormV0("PUSHDATA2 0x0400 0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("PUSHDATA2 0x0400 0x01020304"),
	}, {
		// This is padded to make the push canonical.
		name: "simple case (pushdata4)",
		before: mustParseShortFormV0("PUSHDATA4 0x00000100 0x00{65532} " +
			"0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  nil,
	}, {
		name:   "simple case (pushdata4 miss noncanonical)",
		before: mustParseShortFormV0("PUSHDATA4 0x04000000 0x01020304"),
		remove: []byte{1, 2, 3, 4},
		after:  mustParseShortFormV0("PUSHDATA4 0x04000000 0x01020304"),
	}, {
		// This is padded to make the push canonical.
		name: "simple case (pushdata4 miss)",
		before: mustParseShortFormV0("PUSHDATA4 0x00000100 0x00{65532} " +
			"0x01020304"),
		remove: []byte{1, 2, 3, 4, 5},
		after: mustParseShortFormV0("PUSHDATA4 0x00000100 0x00{65532} " +
			"0x01020304"),
	}, {
		name:   "invalid opcode",
		before: []byte{OP_UNKNOWN240},
		remove: []byte{1, 2, 3, 4},
		after:  []byte{OP_UNKNOWN240},
	}, {
		name:   "invalid length (instruction)",
		before: []byte{OP_PUSHDATA1},
		remove: []byte{1, 2, 3, 4},
		err:    ErrMalformedPush,
	}, {
		name:   "invalid length (data)",
		before: []byte{OP_PUSHDATA1, 255, 254},
		remove: []byte{1, 2, 3, 4},
		err:    ErrMalformedPush,
	}}

	// tstRemoveOpcodeByData is a convenience function to ensure the provided
	// script parses before attempting to remove the passed data.
	const scriptVersion = 0
	tstRemoveOpcodeByData := func(script []byte, data []byte) ([]byte, error) {
		if err := checkScriptParses(scriptVersion, script); err != nil {
			return nil, err
		}

		return removeOpcodeByData(script, data), nil
	}

	for _, test := range tests {
		result, err := tstRemoveOpcodeByData(test.before, test.remove)
		if !errors.Is(err, test.err) {
			t.Errorf("%s: unexpected error -- got %v, want %v", test.name, err,
				test.err)
			continue
		}

		if !bytes.Equal(result, test.after) {
			t.Errorf("%s: value does not equal expected -- got: %x, want %x",
				test.name, result, test.after)
		}
	}
}

// TestIsPayToScriptHash ensures the IsPayToScriptHash function returns the
// expected results.
func TestIsPayToScriptHash(t *testing.T) {
	t.Parallel()

	// Convience function that combines fmt.Sprintf with mustParseShortFormV0
	// to create more compact tests.
	p := func(format string, a ...interface{}) []byte {
		return mustParseShortFormV0(fmt.Sprintf(format, a...))
	}

	// Script hash for a 2-of-3 multisig composed of the following public keys:
	// pk1: 0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798
	// pk2: 02f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9
	// pk3: 03fff97bd5755eeea420453a14355235d382f6472f8568a18b2f057a1460297556
	p2sh := "f86b5a7c6d32566aa4dccc04d1533530b4d64cf3"

	tests := []struct {
		name   string // test description
		script []byte // script to examine
		want   bool   // expected p2sh?
	}{{
		name:   "almost v0 p2sh -- wrong hash length",
		script: p("HASH160 DATA_21 0x00%s EQUAL", p2sh),
		want:   false,
	}, {
		name:   "almost v0 p2sh -- trailing opcode",
		script: p("HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		want:   false,
	}, {
		name:   "v0 p2sh",
		script: p("HASH160 DATA_20 0x%s EQUAL", p2sh),
		want:   true,
	}}

	for _, test := range tests {
		got := IsPayToScriptHash(test.script)
		if got != test.want {
			t.Errorf("%q: unexpected result -- got %v, want %v", test.name,
				got, test.want)
		}
	}
}

// TestIsAnyKindOfScriptHash ensures the isAnyKindOfScriptHash function returns
// the expected results.
func TestIsAnyKindOfScriptHash(t *testing.T) {
	t.SkipNow()

	t.Parallel()

	// Convience function that combines fmt.Sprintf with mustParseShortFormV0
	// to create more compact tests.
	p := func(format string, a ...interface{}) []byte {
		return mustParseShortFormV0(fmt.Sprintf(format, a...))
	}

	// Script hash for a 2-of-3 multisig composed of the following public keys:
	// pk1: 0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798
	// pk2: 02f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9
	// pk3: 03fff97bd5755eeea420453a14355235d382f6472f8568a18b2f057a1460297556
	p2sh := "f86b5a7c6d32566aa4dccc04d1533530b4d64cf3"

	tests := []struct {
		name   string // test description
		script []byte // script to examine
		isTrsy bool   // whether script involves a treasury opcode
		want   bool   // expected result
	}{{
		name:   "almost v0 p2sh -- wrong hash length",
		script: p("HASH160 DATA_21 0x00%s EQUAL", p2sh),
		want:   false,
	}, {
		name:   "almost v0 p2sh -- trailing opcode",
		script: p("HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		want:   false,
	}, {
		name:   "v0 p2sh",
		script: p("HASH160 DATA_20 0x%s EQUAL", p2sh),
		want:   true,
	}, {
		name:   "almost v0 stake submission  p2sh -- wrong hash length",
		script: p("SSTX HASH160 DATA_21 0x00%s EQUAL", p2sh),
		want:   false,
	}, {
		name:   "almost v0 stake submission  p2sh -- trailing opcode",
		script: p("SSTX HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		want:   false,
	}, {
		name:   "v0 stake submission p2sh",
		script: p("SSTX HASH160 DATA_20 0x%s EQUAL", p2sh),
		want:   true,
	}, {
		name:   "almost v0 stake gen p2sh -- wrong hash length",
		script: p("SSGEN HASH160 DATA_21 0x00%s EQUAL", p2sh),
		want:   false,
	}, {
		name:   "almost v0 stake gen p2sh -- trailing opcode",
		script: p("SSGEN HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		want:   false,
	}, {
		name:   "v0 stake gen p2sh",
		script: p("SSGEN HASH160 DATA_20 0x%s EQUAL", p2sh),
		want:   true,
	}, {
		name:   "almost v0 stake revocation p2sh -- wrong hash length",
		script: p("SSRTX HASH160 DATA_21 0x00%s EQUAL", p2sh),
		want:   false,
	}, {
		name:   "almost v0 stake revocation p2sh -- trailing opcode",
		script: p("SSRTX HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		want:   false,
	}, {
		name:   "v0 stake revocation p2sh",
		script: p("SSRTX HASH160 DATA_20 0x%s EQUAL", p2sh),
		want:   true,
	}, {
		name:   "almost v0 stake change p2sh -- wrong hash length",
		script: p("SSTXCHANGE HASH160 DATA_21 0x00%s EQUAL", p2sh),
		want:   false,
	}, {
		name:   "almost v0 stake change p2sh -- trailing opcode",
		script: p("SSTXCHANGE HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		want:   false,
	}, {
		name:   "v0 stake change p2sh",
		script: p("SSTXCHANGE HASH160 DATA_20 0x%s EQUAL", p2sh),
		want:   true,
	}, {
		name:   "almost v0 treasury gen p2sh -- wrong hash length",
		script: p("TGEN HASH160 DATA_21 0x00%s EQUAL", p2sh),
		isTrsy: true,
		want:   false,
	}, {
		name:   "almost v0 treasury gen p2sh -- trailing opcode",
		script: p("TGEN HASH160 DATA_20 0x%s EQUAL TRUE", p2sh),
		isTrsy: true,
		want:   false,
	}, {
		name:   "v0 treasury gen p2sh",
		script: p("TGEN HASH160 DATA_20 0x%s EQUAL", p2sh),
		isTrsy: true,
		want:   true,
	}}

	for _, test := range tests {
		// Run the tests with and without treasury enabled.
		for _, flags := range []ScriptFlags{0, ScriptVerifyTreasury} {
			vm := Engine{flags: flags}
			got := vm.isAnyKindOfScriptHash(test.script)
			want := test.want && (!test.isTrsy || flags != 0)
			if got != want {
				t.Errorf("%q: unexpected result -- got %v, want %v", test.name,
					got, want)
			}
		}
	}
}

// TestHasCanonicalPushes ensures the isCanonicalPush function properly
// determines what is considered a canonical push for the purposes of
// removeOpcodeByData and script null data checks.
func TestHasCanonicalPushes(t *testing.T) {
	t.Parallel()

	const scriptVersion = 0
	type canonicalPushTest struct {
		name     string // test description
		script   string // short form script to test
		expected bool   // expected result
	}
	tests := []canonicalPushTest{{
		name: "does not parse",
		script: "0x046708afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e" +
			"0ea1f61d",
		expected: false,
	}, {
		name:     "non-canonical push",
		script:   "PUSHDATA1 0x04 0x01020304",
		expected: false,
	}}
	for i := 0; i < 65535; i++ {
		tests = append(tests, canonicalPushTest{
			name:     fmt.Sprintf("canonical push of integer %d", i),
			script:   fmt.Sprintf("%d", i),
			expected: true,
		})
	}
	for i := 0; i <= MaxScriptElementSize; i++ {
		tests = append(tests, canonicalPushTest{
			name:     fmt.Sprintf("canonical push of %d bytes of data", i),
			script:   fmt.Sprintf("'a'{%d}", i),
			expected: true,
		})
	}

	for _, test := range tests {
		script := mustParseShortForm(scriptVersion, test.script)
		if err := checkScriptParses(scriptVersion, script); err != nil {
			if test.expected {
				t.Errorf("%s: script parse failed: %v", test.name, err)
			}
			continue
		}
		tokenizer := MakeScriptTokenizer(scriptVersion, script)
		for tokenizer.Next() {
			result := isCanonicalPush(tokenizer.Opcode(), tokenizer.Data())
			if result != test.expected {
				t.Errorf("%s: wrong result -- got %v, want: %v", test.name,
					result, test.expected)
				break
			}
		}
	}
}

// TestIsPushOnlyScript ensures the IsPushOnlyScript function returns the
// expected results.
func TestIsPushOnlyScript(t *testing.T) {
	t.Parallel()

	type pushOnlyTest struct {
		name     string // test description
		script   string // short form script to test
		expected bool   // expected result
	}
	tests := []pushOnlyTest{{
		name: "does not parse",
		script: "0x046708afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0" +
			"ea1f61d",
		expected: false,
	}}
	for i := 0; i < 65535; i++ {
		tests = append(tests, pushOnlyTest{
			name:     fmt.Sprintf("canonical push of integer %d", i),
			script:   fmt.Sprintf("%d", i),
			expected: true,
		})
	}
	for i := 0; i <= MaxScriptElementSize; i++ {
		tests = append(tests, pushOnlyTest{
			name:     fmt.Sprintf("canonical push of %d bytes of data", i),
			script:   fmt.Sprintf("'a'{%d}", i),
			expected: true,
		})
	}

	for _, test := range tests {
		script := mustParseShortFormV0(test.script)
		result := IsPushOnlyScript(script)
		if result != test.expected {
			t.Errorf("%s: wrong result -- got: %v, want: %v", test.name, result,
				test.expected)
		}
	}
}

// TestIsUnspendable ensures the IsUnspendable function returns the expected
// results.
func TestIsUnspendable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		amount   int64
		pkScript string
		expected bool
	}{{
		name:     "unspendable due to being provably pruneable",
		amount:   100,
		pkScript: "RETURN DATA_4 0x74657374",
		expected: true,
	}, {
		name:   "unspendable due to zero amount",
		amount: 0,
		pkScript: "DUP HASH160 DATA_20 0x2995a0fe6843fa9b954597f0dca7a44df6fa" +
			"0b5c EQUALVERIFY CHECKSIG",
		expected: true,
	}, {
		name:   "spendable",
		amount: 100,
		pkScript: "DUP HASH160 DATA_20 0x2995a0fe6843fa9b954597f0dca7a44df6fa" +
			"0b5c EQUALVERIFY CHECKSIG",
		expected: false,
	}}

	for _, test := range tests {
		pkScript := mustParseShortFormV0(test.pkScript)
		result := IsUnspendable(test.amount, pkScript)
		if result != test.expected {
			t.Errorf("%s: unexpected result -- got %v, want %v", test.name,
				result, test.expected)
			continue
		}
	}
}

// TestGenerateSSGenBlockRef ensures the block reference script for use in stake
// vote transactions is generated correctly for various block hashes and
// heights.
func TestGenerateSSGenBlockRef(t *testing.T) {
	var tests = []struct {
		blockHash string
		height    uint32
		expected  []byte
	}{{
		"0000000000004740ad140c86753f9295e09f9cc81b1bb75d7f5552aeeedb7012",
		1000,
		hexToBytes("6a241270dbeeae52557f5db71b1bc89c9fe095923f75860c14ad40470" +
			"00000000000e8030000"),
	}, {
		"000000000000000033eafc268a67c8d1f02343d7a96cf3fe2a4915ef779b52f9",
		290000,
		hexToBytes("6a24f9529b77ef15492afef36ca9d74323f0d1c8678a26fcea3300000" +
			"00000000000d06c0400"),
	}}

	for _, test := range tests {
		h, err := chainhash.NewHashFromStr(test.blockHash)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
			continue
		}
		s, err := GenerateSSGenBlockRef(*h, test.height)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
			continue
		}
		if !bytes.Equal(s, test.expected) {
			t.Errorf("unexpected script -- got %x, want %x", s, test.expected)
			continue
		}
	}
}

// TestGenerateSSGenVotes ensures the expected vote script for use in stake
// vote transactions is generated correctly for various vote bits.
func TestGenerateSSGenVotes(t *testing.T) {
	var tests = []struct {
		votebits uint16
		expected []byte
	}{
		{65535, hexToBytes("6a02ffff")},
		{256, hexToBytes("6a020001")},
		{127, hexToBytes("6a027f00")},
		{0, hexToBytes("6a020000")},
	}
	for _, test := range tests {
		s, err := GenerateSSGenVotes(test.votebits)
		if err != nil {
			t.Errorf("unexpected err: %v", err)
			continue
		}
		if !bytes.Equal(s, test.expected) {
			t.Errorf("unexpected script -- got %x, want %x", s, test.expected)
			continue
		}
	}
}
