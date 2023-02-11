// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcrutil

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/EXCCoin/base58"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/chaincfg/v3"
	"github.com/EXCCoin/exccd/dcrec"
	"github.com/EXCCoin/exccd/dcrec/edwards/v2"
	"github.com/EXCCoin/exccd/dcrec/secp256k1/v4"
)

var (
	// ErrMalformedPrivateKey describes an error where a WIF-encoded private key
	// cannot be decoded due to being improperly formatted.  This may occur if
	// the byte length is incorrect or an unexpected magic number was
	// encountered.
	ErrMalformedPrivateKey = errors.New("malformed private key")

	// ErrChecksumMismatch describes an error where decoding failed due to a bad
	// checksum.
	ErrChecksumMismatch = errors.New("checksum mismatch")
)

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

// ErrWrongWIFNetwork describes an error in which the provided WIF is not for
// the expected network.
type ErrWrongWIFNetwork byte

// Error implements the error interface.
func (e ErrWrongWIFNetwork) Error() string {
	return fmt.Sprintf("WIF is not for the network identified by %#02x",
		byte(e))
}

// WIF contains the individual components described by the Wallet Import Format
// (WIF).  A WIF string is typically used to represent a private key and its
// associated address in a way that may be easily copied and imported into or
// exported from wallet software.  WIF strings may be decoded into this
// structure by calling DecodeWIF or created with a user-provided private key
// by calling NewWIF.
type WIF struct {
	// scheme is the type of signature scheme used.
	scheme dcrec.SignatureType

	// privKey is the private key being imported or exported.
	privKey []byte

	// pubKey is the public key of the privKey
	pubKey []byte

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
func NewUncompressedWIF(privKey []byte, net *chaincfg.Params) (*WIF, error) {
	if net == nil {
		return nil, errors.New("no network")
	}
	return &WIF{dcrec.STEcdsaSecp256k1, privKey, []byte{}, false, net.PrivateKeyID}, nil
}

// NewWIF creates a new WIF structure to export an address and its private key
// as a string encoded in the Wallet Import Format. The net parameter specifies
// the magic bytes of the network for which the WIF string is intended.
// The address intended to be imported or exported was created
// by serializing the public key COMPRESSED.
func NewWIF(privKey []byte, net byte, scheme dcrec.SignatureType) (*WIF, error) {
	var pubBytes []byte
	switch scheme {
	case dcrec.STEcdsaSecp256k1, dcrec.STSchnorrSecp256k1:
		priv := secp256k1.PrivKeyFromBytes(privKey)
		pubBytes = priv.PubKey().SerializeCompressed()
	case dcrec.STEd25519:
		_, pub, err := edwards.PrivKeyFromScalar(privKey)
		if err != nil {
			return nil, err
		}
		pubBytes = pub.SerializeCompressed()
	default:
		return nil, fmt.Errorf("unsupported signature type '%v'", scheme)
	}
	return &WIF{scheme, privKey, pubBytes, true, net}, nil
}

// IsForNet returns whether or not the decoded WIF structure is associated
// with the passed network.
func (w *WIF) IsForNet(net *chaincfg.Params) bool {
	return w.netID == net.PrivateKeyID
}

// DecodeWIF creates a new WIF structure by decoding the string encoding of
// the import format which is required to be for the provided network.
//
// The WIF string must be a base58-encoded string of the following byte
// sequence:
//
//   - 1 byte to identify the network, must be 0x80 for mainnet or 0xef for
//     either testnet or the simnet test network
//   - 32 bytes of a binary-encoded, big-endian, zero-padded private key
//   - Optional 1 byte (greater or equal to 0x01) if the address being imported or exported
//     was created by taking the RIPEMD160 after SHA256 hash of a serialized
//     compressed (33-byte) public key. The byte also indicates EC type
//     0x1 for Secp256k1, 0x2 for Edwards, 0x3 for SecSchnorr
//   - 4 bytes of checksum, must equal the first four bytes of the double SHA256
//     of every byte before the checksum in this sequence
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
	var ecType dcrec.SignatureType

	// Length of base58 decoded WIF must be 32 bytes + an optional 1 byte
	// (0x01) if compressed, plus 1 byte for netID + 4 bytes of checksum.
	switch decodedLen {
	case 1 + privKeyBytesLen + 1 + cksumBytesLen:
		compress = true
		ecType = dcrec.SignatureType(int(decoded[33]) - ecTypeOffset)
	case 1 + privKeyBytesLen + cksumBytesLen:
		compress = false
		ecType = dcrec.STEcdsaSecp256k1
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

	var privKeyBytes, pubKeyBytes []byte
	var scheme dcrec.SignatureType

	switch ecType {
	case dcrec.STEcdsaSecp256k1:
		privKeyBytes = decoded[1 : 1+secp256k1.PrivKeyBytesLen]
		privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
		pubKeyBytes = privKey.PubKey().SerializeCompressed()
		scheme = dcrec.STEcdsaSecp256k1
	case dcrec.STEd25519:
		privKeyBytes = decoded[1 : 1+edwards.PrivScalarSize]
		_, pubKey, err := edwards.PrivKeyFromScalar(privKeyBytes)
		if err != nil {
			return nil, err
		}
		pubKeyBytes = pubKey.SerializeCompressed()
		scheme = dcrec.STEd25519
	case dcrec.STSchnorrSecp256k1:
		privKeyBytes = decoded[1 : 1+secp256k1.PrivKeyBytesLen]
		privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
		pubKeyBytes = privKey.PubKey().SerializeCompressed()
		scheme = dcrec.STSchnorrSecp256k1
	default:
		return nil, ErrMalformedPrivateKey
	}

	netID := decoded[0]
	return &WIF{scheme, privKeyBytes, pubKeyBytes, compress, netID}, nil
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
	a = append(a, w.privKey...)
	if w.CompressPubKey {
		a = append(a, byte(int(w.scheme)+ecTypeOffset))
	}

	cksum := chainhash.DoubleHashB(a)
	a = append(a, cksum[:cksumBytesLen]...)
	return base58.Encode(a)
}

// PrivKey returns the serialized private key described by the WIF.  The bytes
// must not be modified.
func (w *WIF) PrivKey() []byte {
	return w.privKey
}

// PubKey returns the compressed serialization of the associated public key for
// the WIF's private key.
func (w *WIF) PubKey() []byte {
	return w.pubKey
}

// DSA describes the signature scheme of the key.
func (w *WIF) DSA() dcrec.SignatureType {
	return w.scheme
}
