// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package exccutil

import (
	"bytes"
	"errors"

	"github.com/EXCCoin/base58"
	"github.com/EXCCoin/exccd/chaincfg"
	"github.com/EXCCoin/exccd/chaincfg/chainec"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
)

// ErrMalformedPrivateKey describes an error where a WIF-encoded private
// key cannot be decoded due to being improperly formatted. This may occur
// if the byte length is incorrect or an unexpected EC Type was
// encountered.
var ErrMalformedPrivateKey = errors.New("malformed private key")

const (
	// ecTypeOffset is the offset of EC Type used to maintain compatibility with
	// Bitcoin magic number used to identify a WIF encoding for
	// an address created from a compressed serialized public key.
	ecTypeOffset int = 1

	// privKeyBytes is size of private key in bytes
	privKeyBytesLen = 32

	// cksumBytesLen is size of checksum in bytes
	cksumBytesLen = 4
)

// WIF contains the individual components described by the Wallet Import Format
// (WIF).  A WIF string is typically used to represent a private key and its
// associated address in a way that  may be easily copied and imported into or
// exported from wallet software.  WIF strings may be decoded into this
// structure by calling DecodeWIF or created with a user-provided private key
// by calling NewWIF.
type WIF struct {
	// ecType is the type of ECDSA used.
	ecType int

	// PrivKey is the private key being imported or exported.
	PrivKey chainec.PrivateKey

	// CompressPubKey specifies whether the address controlled by the
	// imported or exported private key was created by hashing a
	// compressed (33-byte) serialized public key, rather than an
	// uncompressed (65-byte) one.
	CompressPubKey bool

	// netID is the network identifier byte used when
	// WIF encoding the private key.
	netID byte
}

// NewUncompressedWIF creates a new WIF structure to export an address and its private key
// as a string encoded in the Wallet Import Format.
// The address intended to be imported or exported was created
// by serializing the Secp256k1 public key UNCOMPRESSED (legacy compatibility).
func NewUncompressedWIF(privKey chainec.PrivateKey, net *chaincfg.Params) (*WIF, error) {
	if net == nil {
		return nil, errors.New("no network")
	}
	return &WIF{chainec.ECTypeSecp256k1, privKey, false, net.PrivateKeyID}, nil
}

// NewWIF creates a new WIF structure to export an address and its private key
// as a string encoded in the Wallet Import Format.
// The address intended to be imported or exported was created
// by serializing the public key COMPRESSED.
func NewWIF(privKey chainec.PrivateKey, net *chaincfg.Params, ecType int) (*WIF, error) {
	if net == nil {
		return nil, errors.New("no network")
	}
	return &WIF{ecType, privKey, true, net.PrivateKeyID}, nil
}

// IsForNet returns whether or not the decoded WIF structure is associated
// with the passed network.
func (w *WIF) IsForNet(net *chaincfg.Params) bool {
	return w.netID == net.PrivateKeyID
}

// DecodeWIF creates a new WIF structure by decoding the string encoding of
// the import format.
//
// The WIF string must be a base58-encoded string of the following byte
// sequence:
//
//  * 1 byte to identify the network, must be 0x80 for mainnet or 0xef for
//    either testnet or the simnet test network
//  * 32 bytes of a binary-encoded, big-endian, zero-padded private key
//  * Optional 1 byte (greater or equal to 0x01) if the address being imported or exported
//    was created by taking the RIPEMD160 after SHA256 hash of a serialized
//    compressed (33-byte) public key. The byte also indicates EC type
//    0x1 for Secp256k1, 0x2 for Edwards, 0x3 for SecSchnorr
//  * 4 bytes of checksum, must equal the first four bytes of the double SHA256
//    of every byte before the checksum in this sequence
//
// If the base58-decoded byte sequence does not match this, DecodeWIF will
// return a non-nil error.  ErrMalformedPrivateKey is returned when the WIF
// is of an impossible length or the expected compressed pubkey EC Type is invalid.
// ErrChecksumMismatch is returned if the expected WIF checksum does not match
// the calculated checksum.
func DecodeWIF(wif string) (*WIF, error) {
	decoded := base58.Decode(wif)
	decodedLen := len(decoded)

	var compress bool
	var ecType int

	// Length of base58 decoded WIF must be 32 bytes + an optional 1 byte
	// (0x01) if compressed, plus 1 byte for netID + 4 bytes of checksum.
	switch decodedLen {
	case 1 + privKeyBytesLen + 1 + cksumBytesLen:
		compress = true
		ecType = int(decoded[33]) - ecTypeOffset
	case 1 + privKeyBytesLen + cksumBytesLen:
		compress = false
		ecType = chainec.ECTypeSecp256k1
	default:
		return nil, ErrMalformedPrivateKey
	}

	// Checksum is first four bytes of double SHA256 of the identifier byte
	// and privKey.  Verify this matches the final 4 bytes of the decoded
	// private key.
	var tosum []byte
	if compress {
		tosum = decoded[:1+privKeyBytesLen+1]
	} else {
		tosum = decoded[:1+privKeyBytesLen]
	}
	cksum := chainhash.DoubleHashB(tosum)[:cksumBytesLen]
	if !bytes.Equal(cksum, decoded[decodedLen-cksumBytesLen:]) {
		return nil, ErrChecksumMismatch
	}

	var privKey chainec.PrivateKey

	switch ecType {
	case chainec.ECTypeSecp256k1:
		privKeyBytes := decoded[1 : 1+chainec.Secp256k1.PrivKeyBytesLen()]
		privKey, _ = chainec.Secp256k1.PrivKeyFromScalar(privKeyBytes)
	case chainec.ECTypeEdwards:
		privKeyBytes := decoded[1 : 1+32]
		privKey, _ = chainec.Edwards.PrivKeyFromScalar(privKeyBytes)
	case chainec.ECTypeSecSchnorr:
		privKeyBytes := decoded[1 : 1+chainec.SecSchnorr.PrivKeyBytesLen()]
		privKey, _ = chainec.SecSchnorr.PrivKeyFromScalar(privKeyBytes)
	default:
		return nil, ErrMalformedPrivateKey
	}

	netID := decoded[0]
	return &WIF{ecType, privKey, compress, netID}, nil
}

// String creates the Wallet Import Format string encoding of a WIF structure.
// See DecodeWIF for a detailed breakdown of the format and requirements of
// a valid WIF string.
func (w *WIF) String() string {
	// Precalculate size.  Maximum number of bytes before base58 encoding
	// is one byte for the network, 32 bytes of private key, possibly one
	// extra byte if the pubkey is to be compressed or uses non-secp256k1 EC,
	// and finally four bytes of checksum.
	encodeLen := 1 + privKeyBytesLen + cksumBytesLen
	if w.CompressPubKey {
		encodeLen++
	}

	a := make([]byte, 0, encodeLen)
	a = append(a, w.netID)
	a = append(a, w.PrivKey.Serialize()...)
	if w.CompressPubKey {
		a = append(a, byte(w.ecType+ecTypeOffset))
	}

	cksum := chainhash.DoubleHashB(a)
	a = append(a, cksum[:cksumBytesLen]...)
	return base58.Encode(a)
}

// SerializePubKey serializes the associated public key of the imported or
// exported private key in either a compressed or uncompressed format.  The
// serialization format chosen depends on the value of w.ecType and w.CompressPubKey.
func (w *WIF) SerializePubKey() []byte {
	pkx, pky := w.PrivKey.Public()
	var pk chainec.PublicKey

	switch w.ecType {
	case chainec.ECTypeSecp256k1:
		pk = chainec.Secp256k1.NewPublicKey(pkx, pky)
	case chainec.ECTypeEdwards:
		pk = chainec.Edwards.NewPublicKey(pkx, pky)
	case chainec.ECTypeSecSchnorr:
		pk = chainec.SecSchnorr.NewPublicKey(pkx, pky)
	}

	if w.CompressPubKey {
		return pk.SerializeCompressed()
	}
	return pk.SerializeUncompressed()
}

// DSA returns the ECDSA type for the private key.
func (w *WIF) DSA() int {
	return w.ecType
}
