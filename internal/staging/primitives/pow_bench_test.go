// Copyright (c) 2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package primitives

import (
	"testing"

	"github.com/EXCCoin/exccd/chaincfg/chainhash"
)

// BenchmarkDiffBitsToUint256 benchmarks converting the compact representation
// used to encode difficulty targets to an unsigned 256-bit integer.
func BenchmarkDiffBitsToUint256(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		const input = 0x1b01330e
		DiffBitsToUint256(input)
	}
}

// BenchmarkUint256ToDiffBits benchmarks converting an unsigned 256-bit integer
// to the compact representation used to encode difficulty targets.
func BenchmarkUint256ToDiffBits(b *testing.B) {
	n := hexToUint256("1330e000000000000000000000000000000000000000000000000")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Uint256ToDiffBits(n)
	}
}

// BenchmarkCalcWork benchmarks calculating a work value from difficulty bits.
func BenchmarkCalcWork(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		const input = 0x1b01330e
		CalcWork(input)
	}
}

// BenchmarkHashToUint256 benchmarks converting a hash to an unsigned 256-bit
// integer that can be used to perform math comparisons.
func BenchmarkHashToUint256(b *testing.B) {
	h := "000000000000437482b6d47f82f374cde539440ddb108b0a76886f0d87d126b9"
	hash, err := chainhash.NewHashFromStr(h)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashToUint256(hash)
	}
}

// BenchmarkCheckProofOfWork benchmarks ensuring a given block hash satisfies
// the proof of work requirements for given difficulty bits.
func BenchmarkCheckProofOfWork(b *testing.B) {
	// Data from block 100k on the main network.
	h := "00000000000004289d9a7b0f7a332fb60a1c221faae89a107ce3ab93eead2f93"
	blockHash, err := chainhash.NewHashFromStr(h)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	const diffBits = 0x1a1194b4
	powLimit := hexToUint256(mockMainNetPowLimit())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CheckProofOfWork(blockHash, diffBits, powLimit)
	}
}
