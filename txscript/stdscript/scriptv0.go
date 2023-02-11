// Copyright (c) 2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package stdscript

import (
	"fmt"

	"github.com/EXCCoin/exccd/dcrec"
	"github.com/EXCCoin/exccd/txscript/v4"
)

const (
	// MaxDataCarrierSizeV0 is the maximum number of bytes allowed in pushed
	// data to be considered a standard version 0 provably pruneable nulldata
	// script.
	MaxDataCarrierSizeV0 = 256
)

// ExtractCompressedPubKeyV0 extracts a compressed public key from the passed
// script if it is a standard version 0 pay-to-compressed-secp256k1-pubkey
// script.  It will return nil otherwise.
func ExtractCompressedPubKeyV0(script []byte) []byte {
	// A pay-to-compressed-pubkey script is of the form:
	//  OP_DATA_33 <33-byte compressed pubkey> OP_CHECKSIG

	// All compressed secp256k1 public keys must start with 0x02 or 0x03.
	if len(script) == 35 &&
		script[34] == txscript.OP_CHECKSIG &&
		script[0] == txscript.OP_DATA_33 &&
		(script[1] == 0x02 || script[1] == 0x03) {

		return script[1:34]
	}
	return nil
}

// ExtractUncompressedPubKeyV0 extracts an uncompressed public key from the
// passed script if it is a standard version 0
// pay-to-uncompressed-secp256k1-pubkey script.  It will return nil otherwise.
func ExtractUncompressedPubKeyV0(script []byte) []byte {
	// A pay-to-uncompressed-pubkey script is of the form:
	//  OP_DATA_65 <65-byte uncompressed pubkey> OP_CHECKSIG

	// All non-hybrid uncompressed secp256k1 public keys must start with 0x04.
	if len(script) == 67 &&
		script[66] == txscript.OP_CHECKSIG &&
		script[0] == txscript.OP_DATA_65 &&
		script[1] == 0x04 {

		return script[1:66]
	}
	return nil
}

// ExtractPubKeyV0 extracts either a compressed or uncompressed public key from
// the passed script if it is either a standard version 0
// pay-to-compressed-secp256k1-pubkey or pay-to-uncompressed-secp256k1-pubkey
// script, respectively.  It will return nil otherwise.
func ExtractPubKeyV0(script []byte) []byte {
	if pubKey := ExtractCompressedPubKeyV0(script); pubKey != nil {
		return pubKey
	}
	return ExtractUncompressedPubKeyV0(script)
}

// IsPubKeyScriptV0 returns whether or not the passed script is either a
// standard version 0 pay-to-compressed-secp256k1-pubkey or
// pay-to-uncompressed-secp256k1-pubkey script.
func IsPubKeyScriptV0(script []byte) bool {
	return ExtractPubKeyV0(script) != nil
}

// ExtractPubKeyAltDetailsV0 extracts the public key and signature type from the
// passed script if it is a standard version 0 pay-to-alt-pubkey script.  It
// will return nil otherwise.
func ExtractPubKeyAltDetailsV0(script []byte) ([]byte, dcrec.SignatureType) {
	// A pay-to-alt-pubkey script is of the form:
	//  PUBKEY SIGTYPE OP_CHECKSIGALT
	//
	// The only two currently supported alternative signature types are ed25519
	// and schnorr + secp256k1 (with a compressed pubkey).
	//
	//  OP_DATA_32 <32-byte pubkey> <1-byte ed25519 sigtype> OP_CHECKSIGALT
	//  OP_DATA_33 <33-byte pubkey> <1-byte schnorr+secp sigtype> OP_CHECKSIGALT

	// The script can't possibly be a pay-to-alt-pubkey script if it doesn't
	// end with OP_CHECKSIGALT or have at least two small integer pushes
	// preceding it (although any reasonable pubkey will certainly be larger).
	// Fail fast to avoid more work below.
	if len(script) < 3 || script[len(script)-1] != txscript.OP_CHECKSIGALT {
		return nil, 0
	}

	if len(script) == 35 && script[0] == txscript.OP_DATA_32 &&
		txscript.IsSmallInt(script[33]) &&
		txscript.AsSmallInt(script[33]) == dcrec.STEd25519 {

		return script[1:33], dcrec.STEd25519
	}

	if len(script) == 36 && script[0] == txscript.OP_DATA_33 &&
		txscript.IsSmallInt(script[34]) &&
		txscript.AsSmallInt(script[34]) == dcrec.STSchnorrSecp256k1 &&
		txscript.IsStrictCompressedPubKeyEncoding(script[1:34]) {

		return script[1:34], dcrec.STSchnorrSecp256k1
	}

	return nil, 0
}

// ExtractPubKeyEd25519V0 extracts a public key from the passed script if it is
// a standard version 0 pay-to-ed25519-pubkey script.  It will return nil
// otherwise.
func ExtractPubKeyEd25519V0(script []byte) []byte {
	pk, sigType := ExtractPubKeyAltDetailsV0(script)
	if pk == nil || sigType != dcrec.STEd25519 {
		return nil
	}
	return pk
}

// IsPubKeyEd25519ScriptV0 returns whether or not the passed script is a
// standard version 0 pay-to-ed25519-pubkey script.
func IsPubKeyEd25519ScriptV0(script []byte) bool {
	return ExtractPubKeyEd25519V0(script) != nil
}

// ExtractPubKeySchnorrSecp256k1V0 extracts a public key from the passed script
// if it is a standard version 0 pay-to-schnorr-secp256k1-pubkey script.  It
// will return nil otherwise.
func ExtractPubKeySchnorrSecp256k1V0(script []byte) []byte {
	pk, sigType := ExtractPubKeyAltDetailsV0(script)
	if pk == nil || sigType != dcrec.STSchnorrSecp256k1 {
		return nil
	}
	return pk
}

// IsPubKeySchnorrSecp256k1ScriptV0 returns whether or not the passed script is
// a standard version 0 pay-to-schnorr-secp256k1-pubkey script.
func IsPubKeySchnorrSecp256k1ScriptV0(script []byte) bool {
	return ExtractPubKeySchnorrSecp256k1V0(script) != nil
}

// ExtractPubKeyHashV0 extracts the public key hash from the passed script if it
// is a standard version 0 pay-to-pubkey-hash-ecdsa-secp256k1 script.  It will
// return nil otherwise.
func ExtractPubKeyHashV0(script []byte) []byte {
	// A pay-to-pubkey-hash script is of the form:
	//  OP_DUP OP_HASH160 <20-byte hash> OP_EQUALVERIFY OP_CHECKSIG
	if len(script) == 25 &&
		script[0] == txscript.OP_DUP &&
		script[1] == txscript.OP_HASH160 &&
		script[2] == txscript.OP_DATA_20 &&
		script[23] == txscript.OP_EQUALVERIFY &&
		script[24] == txscript.OP_CHECKSIG {

		return script[3:23]
	}

	return nil
}

// IsPubKeyHashScriptV0 returns whether or not the passed script is a standard
// version 0 pay-to-pubkey-hash-ecdsa-secp256k1 script.
func IsPubKeyHashScriptV0(script []byte) bool {
	return ExtractPubKeyHashV0(script) != nil
}

// IsStandardAltSignatureTypeV0 returns whether or not the provided version 0
// script opcode represents a push of a standard alt signature type.
func IsStandardAltSignatureTypeV0(op byte) bool {
	if !txscript.IsSmallInt(op) {
		return false
	}

	sigType := txscript.AsSmallInt(op)
	return sigType == dcrec.STEd25519 || sigType == dcrec.STSchnorrSecp256k1
}

// ExtractPubKeyHashAltDetailsV0 extracts the public key hash and signature type
// from the passed script if it is a standard version 0 pay-to-alt-pubkey-hash
// script.  It will return nil otherwise.
func ExtractPubKeyHashAltDetailsV0(script []byte) ([]byte, dcrec.SignatureType) {
	// A pay-to-alt-pubkey-hash script is of the form:
	//  DUP HASH160 <20-byte hash> EQUALVERIFY SIGTYPE CHECKSIG
	//
	// The only two currently supported alternative signature types are ed25519
	// and schnorr + secp256k1 (with a compressed pubkey).
	//
	//  DUP HASH160 <20-byte hash> EQUALVERIFY <1-byte ed25519 sigtype> CHECKSIGALT
	//  DUP HASH160 <20-byte hash> EQUALVERIFY <1-byte schnorr+secp sigtype> CHECKSIGALT
	//
	//  Notice that OP_0 is not specified since signature type 0 disabled.
	if len(script) == 26 &&
		script[0] == txscript.OP_DUP &&
		script[1] == txscript.OP_HASH160 &&
		script[2] == txscript.OP_DATA_20 &&
		script[23] == txscript.OP_EQUALVERIFY &&
		IsStandardAltSignatureTypeV0(script[24]) &&
		script[25] == txscript.OP_CHECKSIGALT {

		return script[3:23], dcrec.SignatureType(txscript.AsSmallInt(script[24]))
	}

	return nil, 0
}

// ExtractPubKeyHashEd25519V0 extracts the public key hash from the passed
// script if it is a standard version 0 pay-to-pubkey-hash-ed25519 script.  It
// will return nil otherwise.
func ExtractPubKeyHashEd25519V0(script []byte) []byte {
	pkHash, sigType := ExtractPubKeyHashAltDetailsV0(script)
	if pkHash == nil || sigType != dcrec.STEd25519 {
		return nil
	}
	return pkHash
}

// IsPubKeyHashEd25519ScriptV0 returns whether or not the passed script is a
// standard version 0 pay-to-pubkey-hash-ed25519 script.
func IsPubKeyHashEd25519ScriptV0(script []byte) bool {
	return ExtractPubKeyHashEd25519V0(script) != nil
}

// ExtractPubKeyHashSchnorrSecp256k1V0 extracts the public key hash from the
// passed script if it is a standard version 0
// pay-to-pubkey-hash-schnorr-secp256k1 script.  It will return nil otherwise.
func ExtractPubKeyHashSchnorrSecp256k1V0(script []byte) []byte {
	pkHash, sigType := ExtractPubKeyHashAltDetailsV0(script)
	if pkHash == nil || sigType != dcrec.STSchnorrSecp256k1 {
		return nil
	}
	return pkHash
}

// IsPubKeyHashSchnorrSecp256k1ScriptV0 returns whether or not the passed script
// is a standard version 0 pay-to-pubkey-hash-schnorr-secp256k1 script.
func IsPubKeyHashSchnorrSecp256k1ScriptV0(script []byte) bool {
	return ExtractPubKeyHashSchnorrSecp256k1V0(script) != nil
}

// ExtractScriptHashV0 extracts the script hash from the passed script if it is
// a standard version 0 pay-to-script-hash script.  It will return nil
// otherwise.
func ExtractScriptHashV0(script []byte) []byte {
	// Defer to consensus code.
	return txscript.ExtractScriptHash(script)
}

// IsScriptHashScriptV0 returns whether or not the passed script is a standard
// version 0 pay-to-script-hash script.
func IsScriptHashScriptV0(script []byte) bool {
	return ExtractScriptHashV0(script) != nil
}

// MultiSigDetailsV0 houses details extracted from a standard version 0 ECDSA
// multisig script.
type MultiSigDetailsV0 struct {
	RequiredSigs uint16
	NumPubKeys   uint16
	PubKeys      [][]byte
	Valid        bool
}

// ExtractMultiSigScriptDetailsV0 attempts to extract details from the passed
// version 0 script if it is a standard ECDSA multisig script.  The returned
// details struct will have the valid flag set to false otherwise.
//
// The extract pubkeys flag indicates whether or not the pubkeys themselves
// should also be extracted and is provided because extracting them results in
// an allocation that the caller might wish to avoid.  The PubKeys member of the
// returned details struct will be nil when the flag is false.
func ExtractMultiSigScriptDetailsV0(script []byte, extractPubKeys bool) MultiSigDetailsV0 {
	// A multi-signature script is of the form:
	//  REQ_SIGS PUBKEY PUBKEY PUBKEY ... NUM_PUBKEYS OP_CHECKMULTISIG

	// The script can't possibly be a multisig script if it doesn't end with
	// OP_CHECKMULTISIG or have at least two small integer pushes preceding it.
	// Fail fast to avoid more work below.
	if len(script) < 3 || script[len(script)-1] != txscript.OP_CHECKMULTISIG {
		return MultiSigDetailsV0{}
	}

	// The first opcode must be a small integer specifying the number of
	// signatures required.
	const scriptVersion = 0
	tokenizer := txscript.MakeScriptTokenizer(scriptVersion, script)
	if !tokenizer.Next() || !txscript.IsSmallInt(tokenizer.Opcode()) {
		return MultiSigDetailsV0{}
	}
	requiredSigs := txscript.AsSmallInt(tokenizer.Opcode())

	// There must be at least one required signature.
	if requiredSigs == 0 {
		return MultiSigDetailsV0{}
	}

	// The next series of opcodes must either push public keys or be a small
	// integer specifying the number of public keys.  It should be noted that
	// although the consensus rules allow a higher maximum number of pubkeys,
	// this intentionally further restricts the maximum number to what can be
	// represented by a small integer push (up to a max of 16).
	var numPubKeys int
	var pubKeys [][]byte
	if extractPubKeys {
		pubKeys = make([][]byte, 0, txscript.MaxPubKeysPerMultiSig)
	}
	for tokenizer.Next() {
		data := tokenizer.Data()
		if !txscript.IsStrictCompressedPubKeyEncoding(data) {
			break
		}
		numPubKeys++
		if extractPubKeys {
			pubKeys = append(pubKeys, data)
		}
	}
	if tokenizer.Done() {
		return MultiSigDetailsV0{}
	}

	// The next opcode must be a small integer specifying the number of public
	// keys required.
	op := tokenizer.Opcode()
	if !txscript.IsSmallInt(op) || txscript.AsSmallInt(op) != numPubKeys {
		return MultiSigDetailsV0{}
	}

	// There must be at least as many pubkeys as required signatures.
	if numPubKeys < requiredSigs {
		return MultiSigDetailsV0{}
	}

	// There must only be a single opcode left unparsed which will be
	// OP_CHECKMULTISIG per the check above.
	if int32(len(tokenizer.Script()))-tokenizer.ByteIndex() != 1 {
		return MultiSigDetailsV0{}
	}

	return MultiSigDetailsV0{
		RequiredSigs: uint16(requiredSigs),
		NumPubKeys:   uint16(numPubKeys),
		PubKeys:      pubKeys,
		Valid:        true,
	}
}

// IsMultiSigScriptV0 returns whether or not the passed script is a standard
// version 0 ECDSA multisig script.
//
// NOTE: This function is only valid for version 0 scripts.  It will always
// return false for other script versions.
func IsMultiSigScriptV0(script []byte) bool {
	// Since this is only checking the form of the script, don't extract the
	// public keys to avoid the allocation.
	details := ExtractMultiSigScriptDetailsV0(script, false)
	return details.Valid
}

// finalOpcodeDataV0 returns the data associated with the final opcode in the
// passed version 0 script.  It will return nil if the script fails to parse.
func finalOpcodeDataV0(script []byte) []byte {
	// Avoid unnecessary work.
	if len(script) == 0 {
		return nil
	}

	var data []byte
	const scriptVersion = 0
	tokenizer := txscript.MakeScriptTokenizer(scriptVersion, script)
	for tokenizer.Next() {
		data = tokenizer.Data()
	}
	if tokenizer.Err() != nil {
		return nil
	}
	return data
}

// IsMultiSigSigScriptV0 returns whether or not the passed script appears to be
// a version 0 signature script which consists of a pay-to-script-hash
// multi-signature redeem script.  Determining if a signature script is actually
// a redemption of pay-to-script-hash requires the associated public key script
// which is often expensive to obtain.  Therefore, this makes a fast best effort
// guess that has a high probability of being correct by checking if the
// signature script ends with a data push and treating that data push as if it
// were a p2sh redeem script.
func IsMultiSigSigScriptV0(script []byte) bool {
	// The script can't possibly be a multisig signature script if it doesn't
	// end with OP_CHECKMULTISIG in the redeem script or have at least two small
	// integers preceding it, and the redeem script itself must be preceded by
	// at least a data push opcode.  Fail fast to avoid more work below.
	if len(script) < 4 || script[len(script)-1] != txscript.OP_CHECKMULTISIG {
		return false
	}

	// Parse through the script to find the last opcode and any data it might
	// push and treat it as a p2sh redeem script even though it might not
	// actually be one.
	possibleRedeemScript := finalOpcodeDataV0(script)
	if possibleRedeemScript == nil {
		return false
	}

	// Finally, return if that possible redeem script is a multisig script.
	return IsMultiSigScriptV0(possibleRedeemScript)
}

// MultiSigRedeemScriptFromScriptSigV0 attempts to extract a multi-signature
// redeem script from a version 0 P2SH-redeeming input.  The script is expected
// to already have been checked to be a version 0 multisignature script prior to
// calling this function.  The results are undefined for other script types.
func MultiSigRedeemScriptFromScriptSigV0(script []byte) []byte {
	// The redeemScript is always the last item on the stack of the script sig.
	return finalOpcodeDataV0(script)
}

// isCanonicalPushV0 returns whether or not the given version 0 opcode and
// associated data is a push instruction that uses the smallest instruction to
// do the job.
//
// For example, it is possible to push a value of 1 to the stack as "OP_1",
// "OP_DATA_1 0x01", "OP_PUSHDATA1 0x01 0x01", and others, however, the first
// only takes a single byte, while the rest take more.  Only the first is
// considered canonical.
func isCanonicalPushV0(opcode byte, data []byte) bool {
	dataLen := len(data)
	if opcode > txscript.OP_16 {
		return false
	}
	if opcode < txscript.OP_PUSHDATA1 && opcode > txscript.OP_0 &&
		(dataLen == 1 && data[0] <= 16) {
		return false
	}
	if opcode == txscript.OP_PUSHDATA1 && dataLen < txscript.OP_PUSHDATA1 {
		return false
	}
	if opcode == txscript.OP_PUSHDATA2 && dataLen <= 0xff {
		return false
	}
	if opcode == txscript.OP_PUSHDATA4 && dataLen <= 0xffff {
		return false
	}
	return true
}

// IsNullDataScriptV0 returns whether or not the passed script is a standard
// version 0 null data script.
func IsNullDataScriptV0(script []byte) bool {
	// A null script is of the form:
	//  OP_RETURN <optional data>
	//
	// Thus, it can either be a single OP_RETURN or an OP_RETURN followed by a
	// canonical data push up to MaxDataCarrierSizeV0 bytes.

	// The script can't possibly be a null data script if it doesn't start
	// with OP_RETURN.  Fail fast to avoid more work below.
	if len(script) < 1 || script[0] != txscript.OP_RETURN {
		return false
	}

	// Single OP_RETURN.
	if len(script) == 1 {
		return true
	}

	// OP_RETURN followed by a canonical data push up to MaxDataCarrierSizeV0
	// bytes in length.
	const scriptVersion = 0
	tokenizer := txscript.MakeScriptTokenizer(scriptVersion, script[1:])
	return tokenizer.Next() && tokenizer.Done() &&
		len(tokenizer.Data()) <= MaxDataCarrierSizeV0 &&
		isCanonicalPushV0(tokenizer.Opcode(), tokenizer.Data())
}

// extractStakePubKeyHashV0 extracts the public key hash from the passed script
// if it is a standard version 0 stake-tagged pay-to-pubkey-hash script with the
// provided stake opcode.  It will return nil otherwise.
func extractStakePubKeyHashV0(script []byte, stakeOpcode byte) []byte {
	// A stake-tagged pay-to-pubkey-hash is of the form:
	//   <stake opcode> <standard-pay-to-pubkey-hash script>

	// The script can't possibly be a stake-tagged pay-to-pubkey-hash if it
	// doesn't start with the given stake opcode.  Fail fast to avoid more work
	// below.
	if len(script) < 1 || script[0] != stakeOpcode {
		return nil
	}

	return ExtractPubKeyHashV0(script[1:])
}

// ExtractStakeSubmissionPubKeyHashV0 extracts the public key hash from the
// passed script if it is a standard version 0 stake submission
// pay-to-pubkey-hash script.  It will return nil otherwise.
func ExtractStakeSubmissionPubKeyHashV0(script []byte) []byte {
	// A stake submission pay-to-pubkey-hash script is of the form:
	//  OP_SSTX <standard-pay-to-pubkey-hash script>
	const stakeOpcode = txscript.OP_SSTX
	return extractStakePubKeyHashV0(script, stakeOpcode)
}

// IsStakeSubmissionPubKeyHashScriptV0 returns whether or not the passed script
// is a standard version 0 stake submission pay-to-pubkey-hash script.
func IsStakeSubmissionPubKeyHashScriptV0(script []byte) bool {
	return ExtractStakeSubmissionPubKeyHashV0(script) != nil
}

// ExtractStakePubKeyHashV0 extracts the public key hash from the passed script
// if it is any one of the supported standard version 0 stake-tagged
// pay-to-pubkey-hash scripts.  It will return nil otherwise.
func ExtractStakePubKeyHashV0(script []byte) []byte {
	if h := ExtractStakeSubmissionPubKeyHashV0(script); h != nil {
		return h
	}
	if h := ExtractStakeGenPubKeyHashV0(script); h != nil {
		return h
	}
	if h := ExtractStakeRevocationPubKeyHashV0(script); h != nil {
		return h
	}
	if h := ExtractStakeChangePubKeyHashV0(script); h != nil {
		return h
	}

	return nil
}

// extractStakeScriptHashV0 extracts the script hash from the passed script if
// it is a standard version 0 stake-tagged pay-to-script-hash script with the
// provided stake opcode.  It will return nil otherwise.
func extractStakeScriptHashV0(script []byte, stakeOpcode byte) []byte {
	// A stake-tagged pay-to-script-hash is of the form:
	//   <stake opcode> <standard-pay-to-script-hash script>

	// The script can't possibly be a stake-tagged pay-to-script-hash if it
	// doesn't start with the given stake opcode.  Fail fast to avoid more work
	// below.
	if len(script) < 1 || script[0] != stakeOpcode {
		return nil
	}

	return txscript.ExtractScriptHash(script[1:])
}

// ExtractStakeSubmissionScriptHashV0 extracts the script hash from the
// passed script if it is a standard version 0 stake submission
// pay-to-script-hash script.  It will return nil otherwise.
func ExtractStakeSubmissionScriptHashV0(script []byte) []byte {
	// A stake submission pay-to-script-hash script is of the form:
	//  OP_SSTX <standard-pay-to-script-hash script>
	const stakeOpcode = txscript.OP_SSTX
	return extractStakeScriptHashV0(script, stakeOpcode)
}

// IsStakeSubmissionScriptHashScriptV0 returns whether or not the passed script
// is a standard version 0 stake submission pay-to-script-hash script.
func IsStakeSubmissionScriptHashScriptV0(script []byte) bool {
	return ExtractStakeSubmissionScriptHashV0(script) != nil
}

// ExtractStakeGenPubKeyHashV0 extracts the public key hash from the
// passed script if it is a standard version 0 stake generation
// pay-to-pubkey-hash script.  It will return nil otherwise.
func ExtractStakeGenPubKeyHashV0(script []byte) []byte {
	// A stake generation pay-to-pubkey-hash script is of the form:
	//  OP_SSGEN <standard-pay-to-pubkey-hash script>
	const stakeOpcode = txscript.OP_SSGEN
	return extractStakePubKeyHashV0(script, stakeOpcode)
}

// IsStakeGenPubKeyHashScriptV0 returns whether or not the passed script is a
// standard version 0 stake generation pay-to-pubkey-hash script.
func IsStakeGenPubKeyHashScriptV0(script []byte) bool {
	return ExtractStakeGenPubKeyHashV0(script) != nil
}

// ExtractStakeGenScriptHashV0 extracts the script hash from the passed
// script if it is a standard version 0 stake generation pay-to-script-hash
// script.  It will return nil otherwise.
func ExtractStakeGenScriptHashV0(script []byte) []byte {
	// A stake generation pay-to-script-hash script is of the form:
	//  OP_SSGEN <standard-pay-to-script-hash script>
	const stakeOpcode = txscript.OP_SSGEN
	return extractStakeScriptHashV0(script, stakeOpcode)
}

// IsStakeGenScriptHashScriptV0 returns whether or not the passed script is a
// standard version 0 stake generation pay-to-script-hash script.
func IsStakeGenScriptHashScriptV0(script []byte) bool {
	return ExtractStakeGenScriptHashV0(script) != nil
}

// ExtractStakeRevocationPubKeyHashV0 extracts the public key hash from
// the passed script if it is a standard version 0 stake revocation
// pay-to-pubkey-hash script.  It will return nil otherwise.
func ExtractStakeRevocationPubKeyHashV0(script []byte) []byte {
	// A stake revocation pay-to-pubkey-hash script is of the form:
	//  OP_SSRTX <standard-pay-to-pubkey-hash script>
	const stakeOpcode = txscript.OP_SSRTX
	return extractStakePubKeyHashV0(script, stakeOpcode)
}

// IsStakeRevocationPubKeyHashScriptV0 returns whether or not the passed script
// is a standard version 0 stake revocation pay-to-pubkey-hash script.
func IsStakeRevocationPubKeyHashScriptV0(script []byte) bool {
	return ExtractStakeRevocationPubKeyHashV0(script) != nil
}

// ExtractStakeRevocationScriptHashV0 extracts the script hash from the
// passed script if it is a standard version 0 stake revocation
// pay-to-script-hash script.  It will return nil otherwise.
func ExtractStakeRevocationScriptHashV0(script []byte) []byte {
	// A stake revocation pay-to-script-hash script is of the form:
	//  OP_SSRTX <standard-pay-to-script-hash script>
	const stakeOpcode = txscript.OP_SSRTX
	return extractStakeScriptHashV0(script, stakeOpcode)
}

// IsStakeRevocationScriptHashScriptV0 returns whether or not the passed script
// is a standard version 0 stake revocation pay-to-script-hash script.
func IsStakeRevocationScriptHashScriptV0(script []byte) bool {
	return ExtractStakeRevocationScriptHashV0(script) != nil
}

// ExtractStakeChangePubKeyHashV0 extracts the public key hash from the
// passed script if it is a standard version 0 stake change pay-to-pubkey-hash
// script.  It will return nil otherwise.
func ExtractStakeChangePubKeyHashV0(script []byte) []byte {
	// A stake change pay-to-pubkey-hash script is of the form:
	//  OP_SSTXCHANGE <standard-pay-to-pubkey-hash script>
	const stakeOpcode = txscript.OP_SSTXCHANGE
	return extractStakePubKeyHashV0(script, stakeOpcode)
}

// IsStakeChangePubKeyHashScriptV0 returns whether or not the passed script is a
// standard version 0 stake change pay-to-pubkey-hash script.
func IsStakeChangePubKeyHashScriptV0(script []byte) bool {
	return ExtractStakeChangePubKeyHashV0(script) != nil
}

// ExtractStakeChangeScriptHashV0 extracts the script hash from the passed
// script if it is a standard version 0 stake change pay-to-script-hash
// script.  It will return nil otherwise.
func ExtractStakeChangeScriptHashV0(script []byte) []byte {
	// A stake change pay-to-script-hash script is of the form:
	//  OP_SSTXCHANGE <standard-pay-to-script-hash script>
	const stakeOpcode = txscript.OP_SSTXCHANGE
	return extractStakeScriptHashV0(script, stakeOpcode)
}

// IsStakeChangeScriptHashScriptV0 returns whether or not the passed script is a
// standard version 0 stake change pay-to-script-hash script.
func IsStakeChangeScriptHashScriptV0(script []byte) bool {
	return ExtractStakeChangeScriptHashV0(script) != nil
}

// ExtractStakeScriptHashV0 extracts the script hash from the passed script if
// it is any one of the supported standard version 0 stake-tagged
// pay-to-script-hash scripts.  It will return nil otherwise.
func ExtractStakeScriptHashV0(script []byte) []byte {
	if h := ExtractStakeSubmissionScriptHashV0(script); h != nil {
		return h
	}
	if h := ExtractStakeGenScriptHashV0(script); h != nil {
		return h
	}
	if h := ExtractStakeRevocationScriptHashV0(script); h != nil {
		return h
	}
	if h := ExtractStakeChangeScriptHashV0(script); h != nil {
		return h
	}

	return nil
}

// DetermineScriptTypeV0 returns the type of the passed version 0 script for
// the known standard types.  This includes both types that are required by
// consensus as well as those which are not.
//
// STNonStandard will be returned when the script does not parse.
func DetermineScriptTypeV0(script []byte) ScriptType {
	switch {
	case IsPubKeyScriptV0(script):
		return STPubKeyEcdsaSecp256k1
	case IsPubKeyEd25519ScriptV0(script):
		return STPubKeyEd25519
	case IsPubKeySchnorrSecp256k1ScriptV0(script):
		return STPubKeySchnorrSecp256k1
	case IsPubKeyHashScriptV0(script):
		return STPubKeyHashEcdsaSecp256k1
	case IsPubKeyHashEd25519ScriptV0(script):
		return STPubKeyHashEd25519
	case IsPubKeyHashSchnorrSecp256k1ScriptV0(script):
		return STPubKeyHashSchnorrSecp256k1
	case IsScriptHashScriptV0(script):
		return STScriptHash
	case IsMultiSigScriptV0(script):
		return STMultiSig
	case IsNullDataScriptV0(script):
		return STNullData
	case IsStakeSubmissionPubKeyHashScriptV0(script):
		return STStakeSubmissionPubKeyHash
	case IsStakeSubmissionScriptHashScriptV0(script):
		return STStakeSubmissionScriptHash
	case IsStakeGenPubKeyHashScriptV0(script):
		return STStakeGenPubKeyHash
	case IsStakeGenScriptHashScriptV0(script):
		return STStakeGenScriptHash
	case IsStakeRevocationPubKeyHashScriptV0(script):
		return STStakeRevocationPubKeyHash
	case IsStakeRevocationScriptHashScriptV0(script):
		return STStakeRevocationScriptHash
	case IsStakeChangePubKeyHashScriptV0(script):
		return STStakeChangePubKeyHash
	case IsStakeChangeScriptHashScriptV0(script):
		return STStakeChangeScriptHash
	}

	return STNonStandard
}

// DetermineRequiredSigsV0 attempts to identify the number of signatures
// required by the passed version 0 script for the known standard types.
//
// It will return 0 when the script does not parse or is not one of the known
// standard types.
func DetermineRequiredSigsV0(script []byte) uint16 {
	scriptType := DetermineScriptTypeV0(script)
	switch scriptType {
	case STPubKeyHashEcdsaSecp256k1, STScriptHash,
		STPubKeyHashEd25519, STPubKeyHashSchnorrSecp256k1,
		STPubKeyEcdsaSecp256k1, STPubKeyEd25519, STPubKeySchnorrSecp256k1,
		STStakeSubmissionPubKeyHash, STStakeSubmissionScriptHash,
		STStakeGenPubKeyHash, STStakeGenScriptHash,
		STStakeRevocationPubKeyHash, STStakeRevocationScriptHash,
		STStakeChangePubKeyHash, STStakeChangeScriptHash,
		STTreasuryGenPubKeyHash, STTreasuryGenScriptHash:

		return 1

	case STMultiSig:
		details := ExtractMultiSigScriptDetailsV0(script, false)
		if details.Valid {
			return details.RequiredSigs
		}
		return 0

	case STNullData, STTreasuryAdd:
		return 0
	}

	return 0
}

// MultiSigScriptV0 returns a valid version 0 script for a multisignature
// redemption where the specified threshold number of the keys in the given
// public keys are required to have signed the transaction for success.
//
// The provided public keys must be serialized in the compressed format or an
// error with kind ErrPubKeyType will be returned.
//
// An Error with kind ErrTooManyRequiredSigs will be returned if the threshold
// is larger than the number of keys provided.
func MultiSigScriptV0(threshold int, pubKeys ...[]byte) ([]byte, error) {
	if len(pubKeys) < threshold {
		str := fmt.Sprintf("unable to generate multisig script with %d "+
			"required signatures when there are only %d public keys available",
			threshold, len(pubKeys))
		return nil, makeError(ErrTooManyRequiredSigs, str)
	}

	builder := txscript.NewScriptBuilder().AddInt64(int64(threshold))
	for _, pubKey := range pubKeys {
		if !txscript.IsStrictCompressedPubKeyEncoding(pubKey) {
			str := fmt.Sprintf("unable to generate multisig script with "+
				"unsupported public key %x", pubKey)
			return nil, makeError(ErrPubKeyType, str)
		}

		builder.AddData(pubKey)
	}
	builder.AddInt64(int64(len(pubKeys)))
	builder.AddOp(txscript.OP_CHECKMULTISIG)

	return builder.Script()
}

// ProvablyPruneableScriptV0 returns a valid version 0 provably-pruneable script
// which consists of an OP_RETURN followed by the passed data.  An Error with
// kind ErrTooMuchNullData will be returned if the length of the passed data
// exceeds MaxDataCarrierSizeV0.
func ProvablyPruneableScriptV0(data []byte) ([]byte, error) {
	if len(data) > MaxDataCarrierSizeV0 {
		str := fmt.Sprintf("data size %d is larger than max allowed size %d",
			len(data), MaxDataCarrierSizeV0)
		return nil, makeError(ErrTooMuchNullData, str)
	}

	builder := txscript.NewScriptBuilder()
	return builder.AddOp(txscript.OP_RETURN).AddData(data).Script()
}

// AtomicSwapDataPushesV0 houses the data pushes found in hash-based atomic swap
// contracts using version 0 scripts.
type AtomicSwapDataPushesV0 struct {
	RecipientHash160 [20]byte
	RefundHash160    [20]byte
	SecretHash       [32]byte
	SecretSize       int64
	LockTime         int64
}

// ExtractAtomicSwapDataPushesV0 returns the data pushes from an atomic swap
// contract using version 0 scripts if it is one.  It will return nil otherwise.
//
// NOTE: Atomic swaps are not considered standard script types by the dcrd
// mempool policy and should be used with P2SH.  The atomic swap format is also
// expected to change to use a more secure hash function in the future.
func ExtractAtomicSwapDataPushesV0(redeemScript []byte) *AtomicSwapDataPushesV0 {
	// Local constants for convenience.
	const (
		maxMathOpCodeLen = txscript.MathOpCodeMaxScriptNumLen
		maxCltvLen       = txscript.CltvMaxScriptNumLen
	)

	// An atomic swap is of the form:
	//  IF
	//   SIZE <secret size> EQUALVERIFY
	//   SHA256 <32-byte secret hash> EQUALVERIFY DUP
	//   HASH160 <20-byte recipient hash>
	//  ELSE
	//   <locktime> CHECKLOCKTIMEVERIFY DROP DUP HASH160 <20-byte refund hash>
	//  ENDIF
	//  EQUALVERIFY CHECKSIG
	type templateMatch struct {
		expectCanonicalInt bool
		maxIntBytes        int
		opcode             byte
		extractedInt       int64
		extractedData      []byte
	}
	var template = [20]templateMatch{
		{opcode: txscript.OP_IF},
		{opcode: txscript.OP_SIZE},
		{expectCanonicalInt: true, maxIntBytes: maxMathOpCodeLen},
		{opcode: txscript.OP_EQUALVERIFY},
		{opcode: txscript.OP_SHA256},
		{opcode: txscript.OP_DATA_32},
		{opcode: txscript.OP_EQUALVERIFY},
		{opcode: txscript.OP_DUP},
		{opcode: txscript.OP_HASH160},
		{opcode: txscript.OP_DATA_20},
		{opcode: txscript.OP_ELSE},
		{expectCanonicalInt: true, maxIntBytes: maxCltvLen},
		{opcode: txscript.OP_CHECKLOCKTIMEVERIFY},
		{opcode: txscript.OP_DROP},
		{opcode: txscript.OP_DUP},
		{opcode: txscript.OP_HASH160},
		{opcode: txscript.OP_DATA_20},
		{opcode: txscript.OP_ENDIF},
		{opcode: txscript.OP_EQUALVERIFY},
		{opcode: txscript.OP_CHECKSIG},
	}

	const scriptVersion = 0
	var templateOffset int
	tokenizer := txscript.MakeScriptTokenizer(scriptVersion, redeemScript)
	for tokenizer.Next() {
		// Not an atomic swap script if it has more opcodes than expected in the
		// template.
		if templateOffset >= len(template) {
			return nil
		}

		op := tokenizer.Opcode()
		data := tokenizer.Data()
		tplEntry := &template[templateOffset]
		if tplEntry.expectCanonicalInt {
			switch {
			case data != nil:
				val, err := txscript.MakeScriptNum(data, tplEntry.maxIntBytes)
				if err != nil {
					return nil
				}
				tplEntry.extractedInt = int64(val)

			case txscript.IsSmallInt(op):
				tplEntry.extractedInt = int64(txscript.AsSmallInt(op))

			// Not an atomic swap script if the opcode does not push an int.
			default:
				return nil
			}
		} else {
			if op != tplEntry.opcode {
				return nil
			}

			tplEntry.extractedData = data
		}

		templateOffset++
	}
	if err := tokenizer.Err(); err != nil {
		return nil
	}
	if !tokenizer.Done() || templateOffset != len(template) {
		return nil
	}

	// At this point, the script appears to be an atomic swap, so populate and
	// return the extracted data.
	pushes := AtomicSwapDataPushesV0{
		SecretSize: template[2].extractedInt,
		LockTime:   template[11].extractedInt,
	}
	copy(pushes.SecretHash[:], template[5].extractedData)
	copy(pushes.RecipientHash160[:], template[9].extractedData)
	copy(pushes.RefundHash160[:], template[16].extractedData)
	return &pushes
}
