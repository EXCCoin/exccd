// Copyright (c) 2015-2016 The btcsuite developers
// Copyright (c) 2016-2020 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package txscript

import (
	"crypto/rand"
	"encoding/binary"
	"sync"

	"github.com/dchest/siphash"
	"github.com/EXCCoin/exccd/chaincfg/chainhash"
	"github.com/EXCCoin/exccd/dcrec/secp256k1/v4"
	"github.com/EXCCoin/exccd/dcrec/secp256k1/v4/ecdsa"
	"github.com/EXCCoin/exccd/wire"
)

// ProactiveEvictionDepth is the depth of the block at which the signatures for
// the transactions within the block are nearly guaranteed to no longer be
// useful.
const ProactiveEvictionDepth = 2

// shortTxHashKeySize is the size of the byte array required for key material
// for the SipHash keyed shortTxHash function.
const shortTxHashKeySize = 16

// sigCacheEntry represents an entry in the SigCache. Entries within the
// SigCache are keyed according to the sigHash of the signature. In the
// scenario of a cache-hit (according to the sigHash), an additional comparison
// of the signature and public key will be executed in order to ensure a complete
// match. In the occasion that two sigHashes collide, the newer sigHash will
// simply overwrite the existing entry.
type sigCacheEntry struct {
	sig         *ecdsa.Signature
	pubKey      *secp256k1.PublicKey
	shortTxHash uint64
}

// SigCache implements an ECDSA signature verification cache with a randomized
// entry eviction policy. Only valid signatures will be added to the cache. The
// benefits of SigCache are two fold. Firstly, usage of SigCache mitigates a DoS
// attack wherein an attack causes a victim's client to hang due to worst-case
// behavior triggered while processing attacker crafted invalid transactions. A
// detailed description of the mitigated DoS attack can be found here:
// https://bitslog.wordpress.com/2013/01/23/fixed-bitcoin-vulnerability-explanation-why-the-signature-cache-is-a-dos-protection/.
// Secondly, usage of the SigCache introduces a signature verification
// optimization which speeds up the validation of transactions within a block,
// if they've already been seen and verified within the mempool.
type SigCache struct {
	sync.RWMutex
	validSigs      map[chainhash.Hash]sigCacheEntry
	maxEntries     uint
	shortTxHashKey [shortTxHashKeySize]byte
}

// NewSigCache creates and initializes a new instance of SigCache. Its sole
// parameter 'maxEntries' represents the maximum number of entries allowed to
// exist in the SigCache at any particular moment. Random entries are evicted
// to make room for new entries that would cause the number of entries in the
// cache to exceed the max.
func NewSigCache(maxEntries uint) (*SigCache, error) {
	// Create a cryptographically secure random key for generating short tx hashes.
	shortTxHashKey, err := createShortTxHashKey()
	if err != nil {
		return nil, err
	}

	return &SigCache{
		validSigs:      make(map[chainhash.Hash]sigCacheEntry, maxEntries),
		maxEntries:     maxEntries,
		shortTxHashKey: shortTxHashKey,
	}, nil
}

// Exists returns true if an existing entry of 'sig' over 'sigHash' for public
// key 'pubKey' is found within the SigCache. Otherwise, false is returned.
//
// NOTE: This function is safe for concurrent access. Readers won't be blocked
// unless there exists a writer, adding an entry to the SigCache.
func (s *SigCache) Exists(sigHash chainhash.Hash, sig *ecdsa.Signature, pubKey *secp256k1.PublicKey) bool {
	s.RLock()
	entry, ok := s.validSigs[sigHash]
	s.RUnlock()

	return ok && entry.pubKey.IsEqual(pubKey) && entry.sig.IsEqual(sig)
}

// Add adds an entry for a signature over 'sigHash' under public key 'pubKey'
// to the signature cache. In the event that the SigCache is 'full', an
// existing entry is randomly chosen to be evicted in order to make space for
// the new entry.
//
// NOTE: This function is safe for concurrent access. Writers will block
// simultaneous readers until function execution has concluded.
func (s *SigCache) Add(sigHash chainhash.Hash, sig *ecdsa.Signature, pubKey *secp256k1.PublicKey, tx *wire.MsgTx) {
	s.Lock()
	defer s.Unlock()

	if s.maxEntries == 0 {
		return
	}

	// If adding this new entry will put us over the max number of allowed
	// entries, then evict an entry.
	if uint(len(s.validSigs)+1) > s.maxEntries {
		// Remove a random entry from the map. Relying on the random
		// starting point of Go's map iteration. It's worth noting that
		// the random iteration starting point is not 100% guaranteed
		// by the spec, however most Go compilers support it.
		// Ultimately, the iteration order isn't important here because
		// in order to manipulate which items are evicted, an adversary
		// would need to be able to execute preimage attacks on the
		// hashing function in order to start eviction at a specific
		// entry.
		for sigEntry := range s.validSigs {
			delete(s.validSigs, sigEntry)
			break
		}
	}
	s.validSigs[sigHash] = sigCacheEntry{sig, pubKey, shortTxHash(tx, s.shortTxHashKey)}
}

// createShortTxHashKey returns a cryptographically secure random key of size
// shortTxHashKeySize that can be used for generating short transaction hashes
// with the shortTxHash function.
func createShortTxHashKey() ([shortTxHashKeySize]byte, error) {
	var key [shortTxHashKeySize]byte
	_, err := rand.Read(key[:])
	if err != nil {
		return key, err
	}

	return key, nil
}

// shortTxHash generates a short hash from the standard transaction hash. The
// hash function used is SipHash-2-4, a keyed function, and it produces a 64-bit
// hash.  The key that is used must be a cryptographically secure random key.
func shortTxHash(msg *wire.MsgTx, key [shortTxHashKeySize]byte) uint64 {
	k0 := binary.LittleEndian.Uint64(key[0:8])
	k1 := binary.LittleEndian.Uint64(key[8:16])
	txHash := msg.TxHash()
	return siphash.Hash(k0, k1, txHash[:])
}

// EvictEntries removes all entries from the SigCache that correspond to the
// transactions in the given block.  The block that is passed should be
// ProactiveEvictionDepth blocks deep, which is the depth at which the
// signatures for the transactions within the block are nearly guaranteed to no
// longer be useful.
//
// EvictEntries wraps the unexported evictEntries method, which is run from a
// goroutine.  evictEntries is only invoked if validSigs is not empty.  This
// avoids starting a new goroutine when there is nothing to evict, such as when
// syncing is ongoing.
func (s *SigCache) EvictEntries(block *wire.MsgBlock) {
	s.RLock()
	if len(s.validSigs) == 0 {
		s.RUnlock()
		return
	}
	s.RUnlock()

	go s.evictEntries(block)
}

// evictEntries removes all entries from the SigCache that correspond to the
// transactions in the given block.  The block that is passed should be
// ProactiveEvictionDepth blocks deep, which is the depth at which the
// signatures for the transactions within the block are nearly guaranteed to no
// longer be useful.
//
// Proactively evicting entries reduces the likelihood of the SigCache reaching
// maximum capacity quickly and then relying on random eviction, which may
// randomly evict entries that are still useful.
//
// This method must be run from a goroutine and should not be run during block
// validation.
func (s *SigCache) evictEntries(block *wire.MsgBlock) {
	// Create a set consisting of the short tx hashes that are in the block.
	numTxns := len(block.Transactions) + len(block.STransactions)
	shortTxHashSet := make(map[uint64]struct{}, numTxns)
	for _, tx := range block.Transactions {
		shortTxHashSet[shortTxHash(tx, s.shortTxHashKey)] = struct{}{}
	}
	for _, stx := range block.STransactions {
		shortTxHashSet[shortTxHash(stx, s.shortTxHashKey)] = struct{}{}
	}

	// Iterate through the entries in validSigs and remove any that are associated
	// with a transaction in the block.  This is done by iterating through every
	// entry in validSigs, since the alternative of also keying the map by the
	// shortTxHash would take extra space.
	s.Lock()
	for sigHash, sigEntry := range s.validSigs {
		if _, ok := shortTxHashSet[sigEntry.shortTxHash]; ok {
			delete(s.validSigs, sigHash)
		}
	}
	s.Unlock()
}
