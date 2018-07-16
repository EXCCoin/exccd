// Equihash solver
// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2016 John Tromp

// Equihash presents the following problem
//
// Fix N, K, such that N is a multiple of K+1
// Let integer n = N/(K+1), and view N-bit words
// as having K+1 "digits" of n bits each
// Fix M = 2^{n+1} N-bit hashes H_0, ... , H_{M-1}
// as outputs of a hash function applied to an (n+1)-bit index
//
// Problem: find a binary tree on 2^K distinct indices,
// for which the exclusive-or of leaf hashes is all 0s
// Additionally, it should satisfy the Wagner conditions:
// 1) for each height i subtree, the exclusive-or
// of its 2^i leaf hashes starts with i*n 0 bits,
// 2) the leftmost leaf of any left subtree is less
// than the leftmost leaf of the corresponding right subtree
//
// The algorithm below solves this by storing trees
// as a directed acyclic graph of K layers
// The n digit bits are split into
// BUCKBITS=n-RESTBITS bucket bits and RESTBITS leftover bits
// Each layer i, consisting of height i subtrees
// whose xor starts with i 0-digits, is partitioned into
// 2^BUCKBITS buckets according to the next BUCKBITS in the xor
// Within each bucket, trees whose xor match in the
// remaining RESTBITS bits of the digit are combined
// to produce trees in the next layer
// To eliminate trees with duplicated indices,
// we simply test if the last 32 bits of the xor are 0,
// and if so, assume that this is due to index duplication
// In practice this works very well to avoid bucket overflow
// and produces negligible false positives

#ifndef __EQUI_MINER_H
#define __EQUI_MINER_H

#include <stdio.h>
#include <assert.h>
#include <stdint.h> // for types uint32_t,uint64_t
#include <string.h> // for functions memset
#include "blake2.h"
#include <type_traits>
#include <algorithm>
#include <vector>
#include <unistd.h>
#include <ctype.h>
#include "portable_endian.h"

#ifndef HEADERNONCELEN
#define HEADERNONCELEN ((uint32_t)180)
#endif

#if __cplusplus >= 201703L
#define CONSTEXPR   constexpr
#else
#define CONSTEXPR
#endif

enum verify_code {
    POW_OK = 0,
    POW_INVALID_HEADER_LENGTH = 1,
    POW_DUPLICATE = 2,
    POW_OUT_OF_ORDER = 3,
    POW_NONZERO_XOR = 4,
    POW_SOL_SIZE_MISMATCH = 5,
    POW_UNKNOWN_PARAMS = 6,
};

extern "C" int equihashProxy(const void* blockData, void* solution);

// The algorithm proceeds in K+1 rounds, one for each digit
// All data is stored in two heaps,
// heap0 of type digit0, and heap1 of type digit1
// The following table shows the layout of these heaps
// in each round, which is an optimized version
// of xenoncat's fixed memory layout, avoiding any waste
// Each line shows only a single slot, which is actually
// replicated NSLOTS * NBUCKETS times
//
//             heap0         heap1
// round  hashes   tree   hashes tree
// 0      A A A A A A 0   . . . . . .
// 1      A A A A A A 0   B B B B B 1
// 2      C C C C C 2 0   B B B B B 1
// 3      C C C C C 2 0   D D D D 3 1
// 4      E E E E 4 2 0   D D D D 3 1
// 5      E E E E 4 2 0   F F F 5 3 1
// 6      G G 6 . 4 2 0   F F F 5 3 1
// 7      G G 6 . 4 2 0   H H 7 5 3 1
// 8      I 8 6 . 4 2 0   H H 7 5 3 1
//
// Round 0 generates hashes and stores them in the buckets
// of heap0 according to the initial n-RESTBITS bits
// These hashes are denoted A above and followed by the
// tree tag denoted 0
// In round 1 we combine each pair of slots in the same bucket
// with matching RESTBITS of digit 0 and store the resulting
// 1-tree in heap1 with its xor hash denoted B
// Upon finishing round 1, the A space is no longer needed,
// and is re-used in round 2 to store both the shorter C hashes,
// and their tree tags denoted 2
// Continuing in this manner, each round reads buckets from one
// heap, and writes buckets in the other heap.
// In the final round K, all pairs leading to 0 xors are identified
// and their leafs recovered through the DAG of tree nodes

// main solver object
template <uint32_t WN, uint32_t WK>
struct TrompEquihash {
    //----------------------------------------------
    // Constants and typedefs
    //----------------------------------------------
    static constexpr bool CANTOR = (WN == 200 && WK == 9) ? true:false;
    static constexpr uint32_t RESTBITS = (WN == 200 && WK == 9) ? 10 : 4;

    static constexpr uint32_t NDIGITS         = (WK+1);
    static constexpr uint32_t DIGITBITS       = (WN/(NDIGITS));
    static constexpr uint32_t PROOFSIZE       = (uint32_t)(1 << WK);
    static constexpr uint32_t BASE            = (uint32_t)(1 << DIGITBITS);
    static constexpr uint32_t NHASHES         = (uint32_t)(2 * BASE);
    static constexpr uint32_t HASHESPERBLAKE  = (uint32_t)(512 / WN);
    static constexpr uint32_t HASHOUT         = (uint32_t)(HASHESPERBLAKE * WN / 8);
    // 2_log of number of buckets
    static constexpr uint32_t BUCKBITS = (DIGITBITS-RESTBITS);
    // 2_log of number of slots per bucket
    static constexpr uint32_t SLOTBITS = (RESTBITS+1+1);

    static constexpr double SAVEMEM = (RESTBITS < 8) ? 1.0f : 9.0f/14.0f;
    static constexpr uint32_t NBUCKETS  = (1 << BUCKBITS);
    static constexpr uint32_t BUCKMASK  = (NBUCKETS - 1);
    static constexpr uint32_t SLOTRANGE = (1 << SLOTBITS);
    static constexpr uint32_t SLOTMASK  = (SLOTRANGE - 1);
    static constexpr uint32_t NSLOTS    = (SLOTRANGE * SAVEMEM);
    static constexpr uint32_t NRESTS    = (1 << RESTBITS);
    static constexpr uint32_t MAXSOLS   = (8);

    // tree node identifying its children as two different slots in
    // a bucket on previous layer with matching rest bits (x-tra hash)
    static constexpr uint32_t CANTORBITS = (CANTOR) ? (2*SLOTBITS-2):0;
    static constexpr uint32_t CANTORMASK = (CANTOR) ? ((1<<CANTORBITS) - 1):0;
    static constexpr uint32_t CANTORMAXSQRT = (CANTOR) ? (2 * NSLOTS):0;
    static constexpr uint32_t NSLOTPAIRS = (CANTOR) ? ((NSLOTS-1) * (NSLOTS+2) / 2):0;
    static constexpr uint32_t TREEMINBITS = (CANTOR) ? (BUCKBITS + CANTORBITS):(BUCKBITS + 2 * SLOTBITS );
    static_assert(!CANTOR || (NSLOTPAIRS <= 1 << CANTORBITS), "cantor throws a fit");

    static constexpr uint32_t NBLAKES = 1;

    static_assert(TREEMINBITS <= 32, "tree doesnt fit in 32 bits");

    using tree_t = typename std::conditional<TREEMINBITS <= 16, uint16_t, uint32_t>::type;

    static constexpr uint32_t TREEBYTES = sizeof(tree_t);
    static constexpr uint32_t TREEBITS = TREEBYTES*8;
    static constexpr uint32_t COMPRESSED_SOL_SIZE = (PROOFSIZE * (DIGITBITS + 1) / 8);

    static constexpr uint32_t HASHWORDS0 = ((WN - DIGITBITS + RESTBITS) + TREEBITS - 1) / TREEBITS;
    static constexpr uint32_t HASHWORDS1 = ((WN - 2*DIGITBITS + RESTBITS) + TREEBITS - 1) / TREEBITS;

    //----------------------------------------------
    // Helper structs
    //----------------------------------------------
    struct tree {
        tree_t bid_s0_s1;

        // constructor for height 0 trees stores index instead
        tree(const uint32_t idx) {
            bid_s0_s1 = idx;
        }

        static uint32_t cantor(uint32_t s0, uint32_t s1) {
            return s1 * (s1 + 1) / 2 + s0;
        }

        tree(const uint32_t bid, const uint32_t s0, const uint32_t s1) {
            // CANTOR saves 2 bits by Cantor pairing
            if CONSTEXPR (CANTOR) {
                bid_s0_s1 = (bid << CANTORBITS) | cantor(s0, s1);
            } else {
                bid_s0_s1 = (((bid << SLOTBITS) | s0) << SLOTBITS) | s1;
            }
        }

        // retrieve hash index from tree(const uint32_t idx) constructor
        uint32_t getindex() const {
            return bid_s0_s1;
        }

        // retrieve bucket index
        uint32_t bucketid() const {
            if CONSTEXPR (CANTOR) {
                return bid_s0_s1 >> (2 * SLOTBITS - 2);
            } else {
                return bid_s0_s1 >> (2 * SLOTBITS);
            }
        }
        // retrieve first slot index
        uint32_t slotid0(uint32_t s1) const {
            if CONSTEXPR (CANTOR) {
                return (bid_s0_s1 & CANTORMASK) - cantor(0, s1);
            } else {
                return (bid_s0_s1 >> SLOTBITS) & SLOTMASK;
            }
        }

        // retrieve second slot index
        uint32_t slotid1() const {
            if CONSTEXPR (CANTOR) {
                uint32_t k, q, sqr = 8 * (bid_s0_s1 & CANTORMASK) + 1;;
                // this k=sqrt(sqr) computing loop averages 3.4 iterations out of maximum 9
                for (k = CANTORMAXSQRT; (q = sqr / k) < k; k = (k + q) / 2);
                return (k - 1) / 2;
            } else {
                return bid_s0_s1 & SLOTMASK;
            }
        }

        // returns false for trees sharing a child subtree
        bool prob_disjoint(const tree other) const {
            if CONSTEXPR (CANTOR) {
                if (bucketid() != other.bucketid())
                    return true;
                uint32_t s1 = slotid1(), s0 = slotid0(s1);
                uint32_t os1 = other.slotid1(), os0 = other.slotid0(os1);
                return s1 != os1 && s0 != os0;
            } else {
                tree xort(bid_s0_s1 ^ other.bid_s0_s1);
                return xort.bucketid() || (xort.slotid0(0) && xort.slotid1());
                // next two tests catch much fewer cases and are therefore skipped
                // && slotid0() != other.slotid1() && slotid1() != other.slotid0()
            }
        }
    };

    // each bucket slot occupies a variable number of hash/tree units,
    // all but the last of which hold the xor over all leaf hashes,
    // or what's left of it after stripping the initial i*n 0s
    // the last unit holds the tree node itself
    // the hash is sometimes accessed 32 bits at a time (word)
    // and sometimes 8 bits at a time (bytes)
    union htunit {
        tree tag;
        tree_t word;
        uint8_t bytes[sizeof(tree_t)];
    };

    // A slot is up to HASHWORDS0 hash units followed by a tag
    typedef htunit slot0[HASHWORDS0 + 1];
    typedef htunit slot1[HASHWORDS1 + 1];
    // a bucket is NSLOTS treenodes
    typedef slot0 bucket0[NSLOTS];
    typedef slot1 bucket1[NSLOTS];
    // the N-bit hash consists of K+1 n-bit "digits"
    // each of which corresponds to a layer of NBUCKETS buckets
    typedef bucket0 digit0[NBUCKETS];
    typedef bucket1 digit1[NBUCKETS];
    typedef uint32_t bsizes[NBUCKETS];

    // manages hash and tree data
    struct htalloc {
        bucket0 *heap0;
        bucket1 *heap1;
        uint32_t alloced;

        htalloc() {
            alloced = 0;
        }

        void alloctrees() {
            static_assert(2 * DIGITBITS >= TREEBITS, "needed to ensure hashes shorten by 1 unit every 2 digits");
            heap0 = (bucket0 *) alloc(NBUCKETS, sizeof(bucket0));
            heap1 = (bucket1 *) alloc(NBUCKETS, sizeof(bucket1));
        }

        void dealloctrees() {
            free(heap0);
            free(heap1);
        }

        void *alloc(const uint32_t n, const uint32_t sz) {
            void *mem = calloc(n, sz);
            assert(mem);
            alloced += n * sz;
            return mem;
        }
    };

    //----------------------------------------------
    // Equihash solver data
    //----------------------------------------------
    typedef uint32_t proof[PROOFSIZE];
    blake2b_state blake_ctx; // holds blake2b midstate after call to setheadernounce
    htalloc hta;             // holds allocated heaps
    bsizes *nslots;          // counts number of slots used in buckets
    proof *sols;             // store found solutions here (only first MAXSOLS)
    uint32_t nsols;              // number of solutions found
    const void* user_data;

    TrompEquihash(const void* userData):user_data(userData) {
        static_assert(sizeof(htunit) == sizeof(tree_t), "");
        static_assert(WK & 1, "K assumed odd in candidate() calling indices1()");
        hta.alloctrees();
        nslots = (bsizes *) hta.alloc(2 * NBUCKETS, sizeof(uint32_t));
        sols = (proof *) hta.alloc(MAXSOLS, sizeof(proof));
    }

    ~TrompEquihash() {
        hta.dealloctrees();
        free(nslots);
        free(sols);
    }

    // size (in bytes) of hash in round 0 <= r < WK
    static uint32_t hashsize(const uint32_t r) {
        const uint32_t hashbits = WN - (r + 1) * DIGITBITS + RESTBITS;
        return (hashbits + 7) / 8;
    }

    // convert bytes into words,rounding up
    static uint32_t hashwords(uint32_t bytes) {
        return (bytes + TREEBYTES - 1) / TREEBYTES;
    }

    // prepare blake2b midstate for new run and initialize counters
    void setheadernonce(const unsigned char *input, const uint32_t len, int64_t nonce) {
        setheader(&blake_ctx, input, len, nonce);
        nsols = 0;
    }

    // get heap0 bucket size in threadsafe manner
    uint32_t getslot0(const uint32_t bucketi) {
        return nslots[0][bucketi]++;
    }

    // get heap1 bucket size in threadsafe manner
    uint32_t getslot1(const uint32_t bucketi) {
        return nslots[1][bucketi]++;
    }

    // get old heap0 bucket size and clear it for next round
    uint32_t getnslots0(const uint32_t bid) {
        uint32_t &nslot = nslots[0][bid];
        const uint32_t n = std::min(nslot, NSLOTS);
        nslot = 0;
        return n;
    }

    // get old heap1 bucket size and clear it for next round
    uint32_t getnslots1(const uint32_t bid) {
        uint32_t &nslot = nslots[1][bid];
        const uint32_t n = std::min(nslot, NSLOTS);
        nslot = 0;
        return n;
    }

    // recognize most (but not all) remaining dupes while Wagner-ordering the indices
    bool orderindices(uint32_t *indices, uint32_t size) {
        if (indices[0] > indices[size]) {
            for (uint32_t i = 0; i < size; i++) {
                const uint32_t tmp = indices[i];
                indices[i] = indices[size + i];
                indices[size + i] = tmp;
            }
        }
        return false;
    }

    // listindices combines index tree reconstruction with probably dupe test
    bool listindices0(uint32_t r, const tree t, uint32_t *indices) {
        if (r == 0) {
            *indices = t.getindex();
            return false;
        }
        const slot1 *buck = hta.heap1[t.bucketid()];
        const uint32_t size = 1 << --r;
        uint32_t tagi = hashwords(hashsize(r));
        uint32_t s1 = t.slotid1(), s0 = t.slotid0(CANTOR ? s1 : 0);
        tree t0 = buck[s0][tagi].tag, t1 = buck[s1][tagi].tag;
        return !t0.prob_disjoint(t1)
               || listindices1(r, t0, indices) || listindices1(r, t1, indices + size)
               || orderindices(indices, size) || indices[0] == indices[size];
    }

    // need separate instance for accessing (differently typed) heap1
    bool listindices1(uint32_t r, const tree t, uint32_t *indices) {
        const slot0 *buck = hta.heap0[t.bucketid()];
        const uint32_t size = 1 << --r;
        uint32_t tagi = hashwords(hashsize(r));
        uint32_t s1 = t.slotid1(), s0 = t.slotid0(CANTOR ? s1 : 0);
        tree t0 = buck[s0][tagi].tag, t1 = buck[s1][tagi].tag;
        return listindices0(r, t0, indices) || listindices0(r, t1, indices + size)
               || orderindices(indices, size) || indices[0] == indices[size];
    }

    // check a candidate that resulted in 0 xor
    // add as solution, with proper subtree ordering, if it has unique indices
    void candidate(const tree t) {
        proof prf;
        // listindices combines index tree reconstruction with probably dupe test
        if (listindices1(WK, t, prf) || duped(prf))
            return; // assume WK odd
        // and now we have ourselves a genuine solution
        uint32_t soli = nsols++;
        // copy solution into final place
        if (soli < MAXSOLS)
            memcpy(sols[soli], prf, sizeof(proof));
    }

    // thread-local object that precomputes various slot metrics for each round
    // facilitating access to various bits in the variable size slots
    struct htlayout {
        htalloc hta;
        uint32_t prevhtunits;
        uint32_t nexthtunits;
        uint32_t dunits;
        uint32_t prevbo;

        htlayout(TrompEquihash *eq, uint32_t r) : hta(eq->hta), prevhtunits(0), dunits(0) {
            uint32_t nexthashbytes = hashsize(r);        // number of bytes occupied by round r hash
            nexthtunits = hashwords(nexthashbytes); // number of TREEBITS words taken up by those bytes
            prevbo = 0;                  // byte offset for accessing hash form previous round
            if (r) {     // similar measure for previous round
                uint32_t prevhashbytes = hashsize(r - 1);
                prevhtunits = hashwords(prevhashbytes);
                prevbo = prevhtunits * sizeof(htunit) - prevhashbytes; // 0-1 or 0-3
                dunits = prevhtunits - nexthtunits; // number of words by which hash shrinks
            }
        }

        // extract remaining bits in digit slots in same bucket still need to collide on
        uint32_t getxhash0(const htunit *slot) const {
            if CONSTEXPR (DIGITBITS % 8 == 4 && RESTBITS == 4) {
                return slot->bytes[prevbo] >> 4;
            } else if CONSTEXPR (DIGITBITS % 8 == 4 && RESTBITS == 8) {
                return (slot->bytes[prevbo] & 0xf) << 4 | slot->bytes[prevbo + 1] >> 4;
            } else if CONSTEXPR (DIGITBITS % 8 == 4 && RESTBITS == 10) {
                return (slot->bytes[prevbo] & 0x3f) << 4 | slot->bytes[prevbo + 1] >> 4;
            } else if CONSTEXPR (DIGITBITS % 8 == 0 && RESTBITS == 4) {
                return slot->bytes[prevbo] & 0xf;
            } else if CONSTEXPR (RESTBITS == 0) {
                return 0;
            }
        }

        // similar but accounting for possible change in hashsize modulo 4 bits
        uint32_t getxhash1(const htunit *slot) const {
            if CONSTEXPR (DIGITBITS % 4 == 0 && RESTBITS == 4) {
                return slot->bytes[prevbo] & 0xf;
            } else if CONSTEXPR (DIGITBITS % 4 == 0 && RESTBITS == 8) {
                return slot->bytes[prevbo];
            } else if CONSTEXPR (DIGITBITS % 4 == 0 && RESTBITS == 10) {
                return (slot->bytes[prevbo] & 0x3) << 8 | slot->bytes[prevbo + 1];
            } else if CONSTEXPR (RESTBITS == 0) {
                return 0;
            }
        }

        // test whether two hashes match in last TREEBITS bits
        bool equal(const htunit *hash0, const htunit *hash1) const {
            return hash0[prevhtunits - 1].word == hash1[prevhtunits - 1].word;
        }
    };

    // this thread-local object performs in-bucket collisions
    // by linking together slots that have identical rest bits
    // (which is in essense a 2nd stage bucket sort)
    struct collisiondata {
        // the bitmap is an early experiment in a bitmap encoding
        // that works only for at most 64 slots
        // it might as well be obsoleted as it performs worse even in that case
#ifdef XBITMAP
        static_assert(NSLOTS <= 64, "cant use XBITMAP with more than 64 slots")

        u64 xhashmap[NRESTS];
        u64 xmap;
#else
        // This maintains NRESTS = 2^RESTBITS lists whose starting slot
        // are in xhashslots[] and where subsequent (next-lower-numbered)
        // slots in each list are found through nextxhashslot[]
        // since 0 is already a valid slot number, use ~0 as nil value

        using xslot = typename std::conditional<RESTBITS <= 6, uint8_t, uint16_t>::type;
        static const xslot xnil = ~0;
        xslot xhashslots[NRESTS];
        xslot nextxhashslot[NSLOTS];
        xslot nextslot;
#endif
        uint32_t s0;

        void clear() {
#ifdef XBITMAP
            memset(xhashmap, 0, NRESTS * sizeof(u64));
#else
            memset(xhashslots, xnil, NRESTS * sizeof(xslot));
            memset(nextxhashslot, xnil, NSLOTS * sizeof(xslot));
#endif
        }

        void addslot(uint32_t s1, uint32_t xh) {
#ifdef XBITMAP
            xmap = xhashmap[xh];
            xhashmap[xh] |= (u64)1 << s1;
            s0 = -1;
#else
            nextslot = xhashslots[xh];
            nextxhashslot[s1] = nextslot;
            xhashslots[xh] = s1;
#endif
        }

        bool nextcollision() const {
#ifdef XBITMAP
            return xmap != 0;
#else
            return nextslot != xnil;
#endif
        }

        uint32_t slot() {
#ifdef XBITMAP
            const uint32_t ffs = __builtin_ffsll(xmap);
            s0 += ffs; xmap >>= ffs;
#else
            nextslot = nextxhashslot[s0 = nextslot];
#endif
            return s0;
        }
    };

    // number of hashes extracted from NBLAKES blake2b outputs
    static const uint32_t HASHESPERBLOCK = NBLAKES * HASHESPERBLAKE;
    // number of blocks of parallel blake2b calls
    static const uint32_t NBLOCKS = (NHASHES + HASHESPERBLOCK - 1) / HASHESPERBLOCK;

    void digitZero() {
        htlayout htl(this, 0);
        const uint32_t hashbytes = hashsize(0);
        uint8_t hashes[NBLAKES * 64];
        blake2b_state state0 = blake_ctx;  // local copy on stack can be copied faster
        for (uint32_t block = 0; block < NBLOCKS; block++) {
            static_assert(NBLAKES == 1, "Support for other NBLAKES values is not implemented");

            blake2b_state state = state0;  // make another copy since blake2b_final modifies it
            uint32_t leb = htole32(block);
            blake2b_update(&state, (uint8_t *) &leb, sizeof(uint32_t));
            blake2b_final(&state, hashes, HASHOUT);

            for (uint32_t i = 0; i < NBLAKES; i++) {
                for (uint32_t j = 0; j < HASHESPERBLAKE; j++) {
                    const uint8_t *ph = hashes + i * 64 + j * WN / 8;
                    // figure out bucket for this hash by extracting leading BUCKBITS bits

                    uint32_t bucketid;
                    if CONSTEXPR (BUCKBITS <= 8) {
                        bucketid = (uint32_t) (ph[0] >> (8 - BUCKBITS));
                    } else if CONSTEXPR (BUCKBITS > 8 && BUCKBITS <= 16) {
                        bucketid = ((uint32_t) ph[0] << (BUCKBITS - 8)) | ph[1] >> (16 - BUCKBITS);
                    } else if CONSTEXPR (BUCKBITS > 16) {
                        bucketid = ((((uint32_t) ph[0] << 8) | ph[1]) << (BUCKBITS - 16)) | ph[2] >> (24 - BUCKBITS);
                    }
                    // grab next available slot in that bucket
                    const uint32_t slot = getslot0(bucketid);
                    if (slot >= NSLOTS) {
                        continue;
                    }
                    // location for slot's tag
                    htunit *s = hta.heap0[bucketid][slot] + htl.nexthtunits;
                    // hash should end right before tag
                    memcpy(s->bytes - hashbytes, ph + WN / 8 - hashbytes, hashbytes);
                    // round 0 tags store hash-generating index
                    s->tag = tree((block * NBLAKES + i) * HASHESPERBLAKE + j);
                }
            }
        }
    }

    void digitodd(const uint32_t r) {
        htlayout htl(this, r);
        collisiondata cd;
        // threads process buckets in round-robin fashion
        for (uint32_t bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear(); // could have made this the constructor, and declare here
            slot0 *buck = htl.hta.heap0[bucketid]; // point to first slot of this bucket
            uint32_t bsize = getnslots0(bucketid);    // grab and reset bucket size
            for (uint32_t s1 = 0; s1 < bsize; s1++) {   // loop over slots
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htl.getxhash0(slot1));// identify list of previous colliding slots
                for (; cd.nextcollision();) {
                    const uint32_t s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (htl.equal(slot0, slot1)) {     // expect difference in last 32 bits unless duped
                        continue;
                    }
                    uint32_t xorbucketid;                   // determine bucket for s0 xor s1
                    const uint8_t *bytes0 = slot0->bytes, *bytes1 = slot1->bytes;

                    if CONSTEXPR (WN == 200 && BUCKBITS == 12 && RESTBITS == 8) {
                        xorbucketid = (((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) & 0xf) << 8)
                                      | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]);
                    }
                    if CONSTEXPR (WN == 200 && BUCKBITS == 10 && RESTBITS == 10) {
                        xorbucketid = (((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) & 0xf) << 6)
                                      | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) >> 2;
                    }
                    if CONSTEXPR (WN % 24 == 0 && BUCKBITS == 20 && RESTBITS == 4) {
                        xorbucketid = ((((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) << 8)
                                        | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2])) << 4)
                                      | (bytes0[htl.prevbo + 3] ^ bytes1[htl.prevbo + 3]) >> 4;
                    }
                    if CONSTEXPR (WN == 96 && BUCKBITS == 12 && RESTBITS == 4) {
                        xorbucketid = ((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) << 4)
                                      | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) >> 4;
                    }
                    if CONSTEXPR (WN == 48 && BUCKBITS == 4 && RESTBITS == 4) {
                        xorbucketid = (uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) >> 4;
                    }

                    // grab next available slot in that bucket
                    const uint32_t xorslot = getslot1(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    // start of slot for s0 ^ s1
                    htunit *xs = htl.hta.heap1[xorbucketid][xorslot];
                    // store xor of hashes possibly minus initial 0 word due to collision
                    for (uint32_t i = htl.dunits; i < htl.prevhtunits; i++)
                        xs++->word = slot0[i].word ^ slot1[i].word;
                    // store tree node right after hash
                    xs->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digiteven(const uint32_t r) {
        htlayout htl(this, r);
        collisiondata cd;
        for (uint32_t bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot1 *buck = htl.hta.heap1[bucketid];
            uint32_t bsize = getnslots1(bucketid);
            for (uint32_t s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htl.getxhash1(slot1));
                for (; cd.nextcollision();) {
                    const uint32_t s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (htl.equal(slot0, slot1)) {
                        continue;
                    }
                    uint32_t xorbucketid;
                    const uint8_t *bytes0 = slot0->bytes, *bytes1 = slot1->bytes;
                    if CONSTEXPR (WN == 200 && BUCKBITS == 12 && RESTBITS == 8) {
                        xorbucketid = ((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) << 4)
                                      | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) >> 4;
                    }
                    if CONSTEXPR (WN == 200 && BUCKBITS == 10 && RESTBITS == 10) {
                        xorbucketid = ((uint32_t) (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) << 2)
                                      | (bytes0[htl.prevbo + 3] ^ bytes1[htl.prevbo + 3]) >> 6;
                    }
                    if CONSTEXPR (WN % 24 == 0 && BUCKBITS == 20 && RESTBITS == 4) {
                        xorbucketid = ((((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) << 8)
                                        | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2])) << 4)
                                      | (bytes0[htl.prevbo + 3] ^ bytes1[htl.prevbo + 3]) >> 4;
                    }
                    if CONSTEXPR (WN == 96 && BUCKBITS == 12 && RESTBITS == 4) {
                        xorbucketid = ((uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) << 4)
                                      | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) >> 4;
                    }
                    if CONSTEXPR (WN == 48 && BUCKBITS == 4 && RESTBITS == 4) {
                        xorbucketid = (uint32_t) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) >> 4;
                    }
                    const uint32_t xorslot = getslot0(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = htl.hta.heap0[xorbucketid][xorslot];
                    for (uint32_t i = htl.dunits; i < htl.prevhtunits; i++) {
                        xs++->word = slot0[i].word ^ slot1[i].word;
                    }
                    xs->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    // final round looks simpler
    void digitK() {
        collisiondata cd;
        htlayout htl(this, WK);
        uint32_t nc = 0;
        for (uint32_t bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot0 *buck = htl.hta.heap0[bucketid];   // assume WK odd
            uint32_t bsize = getnslots0(bucketid);      // assume WK odd
            for (uint32_t s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htl.getxhash0(slot1));  // assume WK odd
                for (; cd.nextcollision();) {
                    const uint32_t s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    // there is only 1 word of hash left
                    if (htl.equal(slot0, slot1) && slot0[1].tag.prob_disjoint(slot1[1].tag)) {
                        candidate(tree(bucketid, s0, s1)); // so a match gives a solution candidate
                        nc++;
                    }
                }
            }
        }
    }

    int worker() {
        digitZero();
        if (equihashProxy(user_data, 0))
            return 0;

        for (uint32_t r = 1; r < WK; r++) {
            r&1 ? digitodd(r) : digiteven(r);
            if (equihashProxy(user_data, 0))
                return 0;
        }

        digitK();
        if (equihashProxy(user_data, 0))
            return 0;
        return 1;
    }

    static bool duped(uint32_t* prf) {
        uint32_t sortprf[PROOFSIZE];
        memcpy(sortprf, prf, sizeof(sortprf));
        std::sort(sortprf, sortprf + PROOFSIZE);

        constexpr size_t cBitLen { WN/(WK+1) };
        constexpr size_t maxValue { (1 << (cBitLen + 1)) - 1};

        if (sortprf[0] > maxValue) {
            return true;
        }

        for (uint32_t i = 1; i < PROOFSIZE; i++) {
            if (sortprf[i] <= sortprf[i - 1]) {
                return true;
            }

            if (sortprf[i] > maxValue) {
                return true;
            }
        }
        return false;
    }

    static void hashNonce(blake2b_state *S, uint32_t nonce) {
        uint32_t expandedNonce[8] = {0};
        expandedNonce[0] = htole32(nonce);

        blake2b_update(S, (uint8_t *) &expandedNonce, sizeof(expandedNonce));
    }

    static void setheader(blake2b_state *ctx, const uint8_t *input, uint32_t input_len, int64_t nonce) {
        blake2b_param P;
        memset(&P, 0, sizeof(blake2b_param));

        P.fanout = 1;
        P.depth = 1;
        P.digest_length = (512 / WN) * WN / 8;
        memcpy(P.personal, "ZcashPoW", 8);
        *(uint32_t *) (P.personal + 8) = htole32(WN);
        *(uint32_t *) (P.personal + 12) = htole32(WK);

        blake2b_init_param(ctx, &P);
        blake2b_update(ctx, input, input_len);

        if (nonce >= 0) {
            hashNonce(ctx, (uint32_t)nonce);
        }
    }

    static void genhash(const blake2b_state *ctx, uint32_t idx, uint8_t *hash) {
        blake2b_state state = *ctx;
        uint32_t leb = htole32(idx / HASHESPERBLAKE);
        blake2b_update(&state, (uint8_t *) &leb, sizeof(uint32_t));
        uint8_t blakehash[HASHOUT];
        blake2b_final(&state, blakehash, HASHOUT);
        memcpy(hash, blakehash + (idx % HASHESPERBLAKE) * WN / 8, WN / 8);
    }

    static void CompressArray(const unsigned char* in, size_t in_len,
                              unsigned char* out, size_t out_len,
                              size_t bit_len, size_t byte_pad)
    {
        assert(bit_len >= 8);
        assert(8*sizeof(uint32_t) >= 7+bit_len);

        size_t in_width { (bit_len+7)/8 + byte_pad };
        assert(out_len == bit_len*in_len/(8*in_width));

        uint32_t bit_len_mask { ((uint32_t)1 << bit_len) - 1 };

        // The acc_bits least-significant bits of acc_value represent a bit sequence
        // in big-endian order.
        size_t acc_bits = 0;
        uint32_t acc_value = 0;

        size_t j = 0;
        for (size_t i = 0; i < out_len; i++) {
            // When we have fewer than 8 bits left in the accumulator, read the next
            // input element.
            if (acc_bits < 8) {
                acc_value = acc_value << bit_len;
                for (size_t x = byte_pad; x < in_width; x++) {
                    acc_value = acc_value | ((in[j+x] & ((bit_len_mask >> (8*(in_width-x-1))) & 0xFF)) << (8*(in_width-x-1)));
                }
                j += in_width;
                acc_bits += bit_len;
            }

            acc_bits -= 8;
            out[i] = (acc_value >> acc_bits) & 0xFF;
        }
    }

    static void EhIndexToArray(const uint32_t i, unsigned char* array)
    {
        uint32_t bei = htobe32(i);
        memcpy(array, &bei, sizeof(uint32_t));
    }

    static std::vector<unsigned char> GetMinimalFromIndices(const uint32_t* indices, uint32_t indices_len)
    {
        size_t cBitLen { WN/(WK+1) };
        assert(((cBitLen+1)+7)/8 <= sizeof(uint32_t));
        size_t lenIndices { indices_len*sizeof(uint32_t) };
        size_t minLen { (cBitLen+1)*lenIndices/(8*sizeof(uint32_t)) };
        size_t bytePad { sizeof(uint32_t) - ((cBitLen+1)+7)/8 };
        std::vector<unsigned char> array(lenIndices);

        for (uint32_t i = 0; i < indices_len; i++) {
            EhIndexToArray(indices[i], array.data()+(i*sizeof(uint32_t)));
        }

        std::vector<unsigned char> ret(minLen);
        CompressArray(array.data(), lenIndices,
                      ret.data(), minLen, cBitLen+1, bytePad);
        return ret;
    }

    static int verifyrec(const blake2b_state *ctx, uint32_t *indices, uint8_t *hash, uint32_t r) {
        if (r == 0) {
            TrompEquihash::genhash(ctx, *indices, hash);
            return POW_OK;
        }
        uint32_t *indices1 = indices + (1 << (r - 1));
        if (*indices >= *indices1)
            return POW_OUT_OF_ORDER;

        uint8_t hash0[WN / 8], hash1[WN / 8];
        int vrf0 = verifyrec(ctx, indices, hash0, r - 1);
        if (vrf0 != POW_OK)
            return vrf0;

        int vrf1 = verifyrec(ctx, indices1, hash1, r - 1);
        if (vrf1 != POW_OK)
            return vrf1;

        for (uint32_t i = 0; i < WN / 8; i++)
            hash[i] = hash0[i] ^ hash1[i];

        uint32_t i, b = r < WK ? r * DIGITBITS : WN;

        for (i = 0; i < b / 8; i++)
            if (hash[i])
                return POW_NONZERO_XOR;

        if ((b % 8) && hash[i] >> (8 - (b % 8)))
            return POW_NONZERO_XOR;
        return POW_OK;
    }
};

template<uint32_t WN, uint32_t WK>
void compress_solution(const uint32_t* sol, uint8_t *csol) {
    auto compressed = TrompEquihash<WN, WK>::GetMinimalFromIndices(sol, TrompEquihash<WN, WK>::PROOFSIZE);
    memcpy(csol, compressed.data(), TrompEquihash<WN, WK>::COMPRESSED_SOL_SIZE);
}

template<uint32_t WN, uint32_t WK>
int verify(uint32_t* indices, uint32_t proofsize, const unsigned char *input, const uint32_t input_len, int64_t nonce) {
    if (input_len > HEADERNONCELEN)
        return POW_INVALID_HEADER_LENGTH;
    if (proofsize != TrompEquihash<WN, WK>::PROOFSIZE)
        return POW_SOL_SIZE_MISMATCH;
    if (TrompEquihash<WN, WK>::duped(indices))
        return POW_DUPLICATE;
    blake2b_state ctx;
    TrompEquihash<WN, WK>::setheader(&ctx, input, input_len, nonce);
    uint8_t hash[WN / 8];
    return TrompEquihash<WN, WK>::verifyrec(&ctx, indices, hash, WK);
}

template<uint32_t WN, uint32_t WK>
int solve(const unsigned char* input, uint32_t input_len, int64_t nonce, const void* userData) {
    TrompEquihash<WN, WK> eq(userData);
    eq.setheadernonce(input, input_len, nonce);
    eq.worker();

    uint32_t maxsols = std::min(TrompEquihash<WN, WK>::MAXSOLS, eq.nsols);
    uint8_t csol[TrompEquihash<WN, WK>::COMPRESSED_SOL_SIZE];

    for (uint32_t nsols = 0; nsols < maxsols; nsols++) {
        compress_solution<WN, WK>(eq.sols[nsols], csol);

        int rc = equihashProxy(eq.user_data, csol);

        if (rc == 1) {
            return 1;
        } else if (rc == 2) {
            return 0;
        }
    }

    return eq.nsols;
}

#undef CONSTEXPR

#endif // __EQUI_MINER_H
