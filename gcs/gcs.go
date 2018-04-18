// Copyright (c) 2016-2017 The btcsuite developers
// Copyright (c) 2016-2017 The Lightning Network Developers
// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gcs

import (
	"encoding/binary"
	"errors"
	"math"
	"sort"

	"github.com/aead/siphash"
	"github.com/dchest/blake256"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
)

// Inspired by https://github.com/rasky/gcs

var (
	// ErrNTooBig signifies that the filter can't handle N items.
	ErrNTooBig = errors.New("N does not fit in uint32")

	// ErrPTooBig signifies that the filter can't handle `1/2**P`
	// collision probability.
	ErrPTooBig = errors.New("P is too large")

	// ErrNoData signifies that an empty slice was passed.
	ErrNoData = errors.New("no data provided")

	// ErrMisserialized signifies a filter was misserialized and is missing the
	// N and/or P parameters of a serialized filter.
	ErrMisserialized = errors.New("misserialized filter")
)

// KeySize is the size of the byte array required for key material for the
// SipHash keyed hash function.
const KeySize = siphash.KeySize

// Filter describes an immutable filter that can be built from a set of data
// elements, serialized, deserialized, and queried in a thread-safe manner. The
// serialized form is compressed as a Golomb Coded Set (GCS), but does not
// include N or P to allow the user to encode the metadata separately if
// necessary. The hash function used is SipHash, a keyed function; the key used
// in building the filter is required in order to match filter values and is
// not included in the serialized form.
type Filter struct {
	n           uint32
	p           uint8
	modulusNP   uint64
	filterNData []byte // 4 bytes n big endian, remainder is filter data
}

// NewFilter builds a new GCS filter with the collision probability of
// `1/(2**P)`, key `key`, and including every `[]byte` in `data` as a member of
// the set.
func NewFilter(P uint8, key [KeySize]byte, data [][]byte) (*Filter, error) {
	// Some initial parameter checks: make sure we have data from which to
	// build the filter, and make sure our parameters will fit the hash
	// function we're using.
	if len(data) == 0 {
		return nil, ErrNoData
	}
	if len(data) > math.MaxInt32 {
		return nil, ErrNTooBig
	}
	if P > 32 {
		return nil, ErrPTooBig
	}

	// Create the filter object and insert metadata.
	modP := uint64(1 << P)
	modPMask := modP - 1
	f := Filter{
		n:         uint32(len(data)),
		p:         P,
		modulusNP: uint64(len(data)) * modP,
	}

	// Allocate filter data.
	values := make(uint64Slice, 0, len(data))

	// Insert the hash (modulo N*P) of each data element into a slice and
	// sort the slice.
	for _, d := range data {
		v := siphash.Sum64(d, &key) % f.modulusNP
		values = append(values, v)
	}
	sort.Sort(values)

	var b bitWriter

	// Write the sorted list of values into the filter bitstream,
	// compressing it using Golomb coding.
	var value, lastValue, remainder uint64
	for _, v := range values {
		// Calculate the difference between this value and the last,
		// modulo P.
		remainder = (v - lastValue) & modPMask

		// Calculate the difference between this value and the last,
		// divided by P.
		value = (v - lastValue - remainder) >> f.p
		lastValue = v

		// Write the P multiple into the bitstream in unary; the
		// average should be around 1 (2 bits - 0b10).
		for value > 0 {
			b.writeOne()
			value--
		}
		b.writeZero()

		// Write the remainder as a big-endian integer with enough bits
		// to represent the appropriate collision probability.
		b.writeNBits(remainder, uint(f.p))
	}

	// Save the filter data internally as n + filter bytes
	ndata := make([]byte, 4+len(b.bytes))
	binary.BigEndian.PutUint32(ndata, f.n)
	copy(ndata[4:], b.bytes)
	f.filterNData = ndata

	return &f, nil
}

// FromBytes deserializes a GCS filter from a known N, P, and serialized filter
// as returned by Bytes().
func FromBytes(N uint32, P uint8, d []byte) (*Filter, error) {
	// Basic sanity check.
	if P > 32 {
		return nil, ErrPTooBig
	}

	// Save the filter data internally as n + filter bytes
	ndata := make([]byte, 4+len(d))
	binary.BigEndian.PutUint32(ndata, N)
	copy(ndata[4:], d)

	f := &Filter{
		n:           N,
		p:           P,
		modulusNP:   uint64(N) * uint64(1<<P),
		filterNData: ndata,
	}
	return f, nil
}

// FromNBytes deserializes a GCS filter from a known P, and serialized N and
// filter as returned by NBytes().
func FromNBytes(P uint8, d []byte) (*Filter, error) {
	if len(d) < 4 {
		return nil, ErrMisserialized
	}

	n := binary.BigEndian.Uint32(d[:4])
	f := &Filter{
		n:           n,
		p:           P,
		modulusNP:   uint64(n) * uint64(1<<P),
		filterNData: d,
	}
	return f, nil
}

// FromPBytes deserializes a GCS filter from a known N, and serialized P and
// filter as returned by NBytes().
func FromPBytes(N uint32, d []byte) (*Filter, error) {
	if len(d) < 1 {
		return nil, ErrMisserialized
	}
	return FromBytes(N, d[0], d[1:])
}

// FromNPBytes deserializes a GCS filter from a serialized N, P, and filter as
// returned by NPBytes().
func FromNPBytes(d []byte) (*Filter, error) {
	if len(d) < 5 {
		return nil, ErrMisserialized
	}
	return FromBytes(binary.BigEndian.Uint32(d[:4]), d[4], d[5:])
}

// Bytes returns the serialized format of the GCS filter, which does not
// include N or P (returned by separate methods) or the key used by SipHash.
func (f *Filter) Bytes() []byte {
	return f.filterNData[4:]
}

// NBytes returns the serialized format of the GCS filter with N, which does
// not include P (returned by a separate method) or the key used by SipHash.
func (f *Filter) NBytes() []byte {
	return f.filterNData
}

// PBytes returns the serialized format of the GCS filter with P, which does
// not include N (returned by a separate method) or the key used by SipHash.
func (f *Filter) PBytes() []byte {
	filterData := make([]byte, len(f.filterNData)-3)
	filterData[0] = f.p
	copy(filterData[1:], f.filterNData[4:])
	return filterData
}

// NPBytes returns the serialized format of the GCS filter with N and P, which
// does not include the key used by SipHash.
func (f *Filter) NPBytes() []byte {
	filterData := make([]byte, len(f.filterNData)+1)
	copy(filterData[:4], f.filterNData)
	filterData[4] = f.p
	copy(filterData[5:], f.filterNData[4:])
	return filterData
}

// P returns the filter's collision probability as a negative power of 2 (that
// is, a collision probability of `1/2**20` is represented as 20).
func (f *Filter) P() uint8 {
	return f.p
}

// N returns the size of the data set used to build the filter.
func (f *Filter) N() uint32 {
	return f.n
}

// Match checks whether a []byte value is likely (within collision probability)
// to be a member of the set represented by the filter.
func (f *Filter) Match(key [KeySize]byte, data []byte) bool {
	// Create a filter bitstream.
	b := newBitReader(f.filterNData[4:])

	// Hash our search term with the same parameters as the filter.
	term := siphash.Sum64(data, &key) % f.modulusNP

	// Go through the search filter and look for the desired value.
	var lastValue uint64
	for lastValue < term {
		// Read the difference between previous and new value from
		// bitstream.
		value, err := f.readFullUint64(&b)
		if err != nil {
			return false
		}

		// Add the previous value to it.
		value += lastValue
		if value == term {
			return true
		}

		lastValue = value
	}

	return false
}

// MatchAny checks whether any []byte value is likely (within collision
// probability) to be a member of the set represented by the filter faster than
// calling Match() for each value individually.
func (f *Filter) MatchAny(key [KeySize]byte, data [][]byte) bool {
	if len(data) == 0 {
		return false
	}

	// Create a filter bitstream.
	b := newBitReader(f.filterNData[4:])

	// Create an uncompressed filter of the search values.
	values := make(uint64Slice, 0, len(data))
	for _, d := range data {
		v := siphash.Sum64(d, &key) % f.modulusNP
		values = append(values, v)
	}
	sort.Sort(values)

	// Zip down the filters, comparing values until we either run out of
	// values to compare in one of the filters or we reach a matching
	// value.
	var lastValue1, lastValue2 uint64
	lastValue2 = values[0]
	i := 1
	for lastValue1 != lastValue2 {
		// Check which filter to advance to make sure we're comparing
		// the right values.
		switch {
		case lastValue1 > lastValue2:
			// Advance filter created from search terms or return
			// false if we're at the end because nothing matched.
			if i < len(values) {
				lastValue2 = values[i]
				i++
			} else {
				return false
			}
		case lastValue2 > lastValue1:
			// Advance filter we're searching or return false if
			// we're at the end because nothing matched.
			value, err := f.readFullUint64(&b)
			if err != nil {
				return false
			}
			lastValue1 += value
		}
	}

	// If we've made it this far, an element matched between filters so we
	// return true.
	return true
}

// readFullUint64 reads a value represented by the sum of a unary multiple of
// the filter's P modulus (`2**P`) and a big-endian P-bit remainder.
func (f *Filter) readFullUint64(b *bitReader) (uint64, error) {
	v, err := b.readUnary()
	if err != nil {
		return 0, err
	}

	rem, err := b.readNBits(uint(f.p))
	if err != nil {
		return 0, err
	}

	// Add the multiple and the remainder.
	return v<<f.p + rem, nil
}

// Hash returns the BLAKE256 hash of the filter.
func (f *Filter) Hash() chainhash.Hash {
	h := blake256.New()
	h.Write(f.filterNData)

	var hash chainhash.Hash
	copy(hash[:], h.Sum(nil))
	return hash
}

// MakeHeaderForFilter makes a filter chain header for a filter, given the
// filter and the previous filter chain header.
func MakeHeaderForFilter(filter *Filter, prevHeader *chainhash.Hash) chainhash.Hash {
	filterTip := make([]byte, 2*chainhash.HashSize)
	filterHash := filter.Hash()

	// In the buffer we created above we'll compute hash || prevHash as an
	// intermediate value.
	copy(filterTip, filterHash[:])
	copy(filterTip[chainhash.HashSize:], prevHeader[:])

	// The final filter hash is the blake256 of the hash computed above.
	h := blake256.New()
	h.Write(filterTip)
	var hash chainhash.Hash
	copy(hash[:], h.Sum(nil))
	return hash
}
