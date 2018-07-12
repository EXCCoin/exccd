// Wagner's algorithm for Generalized Birthday Paradox, a memory-hard proof-of-work
// Copyright (c) 2016 John Tromp

#include "miner.h"
#include <unistd.h>
#include <ctype.h>
#include <algorithm>

#ifdef __APPLE__
    #include <machine/endian.h>
    #include <libkern/OSByteOrder.h>
    #define htole32(x) OSSwapHostToLittleInt32(x)
#else
#include <endian.h>
#endif

#if defined __builtin_bswap32 && defined __LITTLE_ENDIAN
#undef htobe32
    #define htobe32(x) __builtin_bswap32(x)
#elif defined __APPLE__
#undef htobe32
    #define htobe32(x) OSSwapHostToBigInt32(x)
#endif

// Select algorithm parameters according to specified configuration
#ifdef CONF_48_5
#define WN          48
#define WK          5
#define RESTBITS    4
#elif defined(CONF_96_5)
#define WN          96
#define WK          5
#define RESTBITS    4
#elif defined(CONF_144_5)
#define WN          144
#define WK          5
#define RESTBITS    4
#elif defined(CONF_200_9)
#define WN          200
#define WK          9
#endif

#define SUFFIX  JOIN(WN,WK)

//Make unique names for multiple instantiations
#define verify              JOIN(verify_,SUFFIX)
#define solve               JOIN(solve_,SUFFIX)
#define compress_solution   JOIN(compress_solution_,SUFFIX)
#define setheader           JOIN(setheader_,SUFFIX)
#define duped               JOIN(duped_,SUFFIX)
#define genhash             JOIN(genhash_,SUFFIX)
#define equi                JOIN(equi_,SUFFIX)
#define htunit              JOIN(htunit_,SUFFIX)
#define tree                JOIN(tree_,SUFFIX)
#define htalloc             JOIN(htalloc_,SUFFIX)

#define NDIGITS        (WK+1)
#define DIGITBITS    (WN/(NDIGITS))
#ifndef RESTBITS
#define CANTOR
#define RESTBITS    10
#endif

#define PROOFSIZE       (u32)(1 << WK)
#define BASE            (u32)(1 << DIGITBITS)
#define NHASHES         (u32)(2 * BASE)
#define HASHESPERBLAKE  (u32)(512 / WN)
#define HASHOUT         (u32)(HASHESPERBLAKE * WN / 8)

// 2_log of number of buckets
#define BUCKBITS (DIGITBITS-RESTBITS)

// 2_log of number of slots per bucket
#define SLOTBITS (RESTBITS+1+1)

// by default buckets have a capacity of twice their expected size
// but this factor reduced it accordingly
#ifndef SAVEMEM

#if RESTBITS < 8
// can't save much memory in such small buckets
#define SAVEMEM 1
#else
// an expected size of at least 512 has such relatively small
// standard deviation that we can reduce capacity with negligible discarding
// this value reduces (200,9) memory to under 144MB
// must be under sqrt(2)/2 with -DCANTOR
#define SAVEMEM 9/14
#endif // RESTBITS == 4

#endif // ifndef SAVEMEM

#define NBUCKETS   (u32)(1 << BUCKBITS)
#define BUCKMASK   (u32)(NBUCKETS - 1)
#define SLOTRANGE  (u32)(1 << SLOTBITS)
#define SLOTMASK   (u32)(SLOTRANGE - 1)
#define NSLOTS     (u32)(SLOTRANGE * SAVEMEM)
#define NRESTS     (u32)(1 << RESTBITS)
#define MAXSOLS    (u32)(8)

// tree node identifying its children as two different slots in
// a bucket on previous layer with matching rest bits (x-tra hash)
#ifdef CANTOR
#define CANTORBITS (2*SLOTBITS-2)
#define CANTORMASK ((1<<CANTORBITS) - 1)
#define CANTORMAXSQRT (2 * NSLOTS)
#define NSLOTPAIRS ((NSLOTS-1) * (NSLOTS+2) / 2)
static_assert(NSLOTPAIRS <= 1 << CANTORBITS, "cantor throws a fit");
#define TREEMINBITS (BUCKBITS + CANTORBITS)
#else
#define TREEMINBITS (BUCKBITS + 2 * SLOTBITS )
#endif

#if TREEMINBITS <= 16
#define tree_t u16
#elif TREEMINBITS <= 32
#define tree_t u32
#else
#error tree doesnt fit in 32 bits
#endif

#define TREEBYTES sizeof(tree_t)
#define TREEBITS (8*TREEBYTES)

extern "C" int equihashProxy(const void* blockData, void* solution);

static int compu32(const void *pa, const void *pb) {
    u32 a = *(u32 *) pa, b = *(u32 *) pb;
    return a < b ? -1 : a == b ? 0 : +1;
}

static bool duped(u32* prf) {
    u32 sortprf[PROOFSIZE];
    memcpy(sortprf, prf, sizeof(sortprf));
    //TODO: (siy) try to replace it with std::sort
    qsort(sortprf, PROOFSIZE, sizeof(u32), &compu32);
    for (u32 i = 1; i < PROOFSIZE; i++)
        if (sortprf[i] <= sortprf[i - 1])
            return true;
    return false;
}

static void hashNonce(blake2b_state *S, uint32_t nonce) {
    uint32_t expandedNonce[8] = {0};
    expandedNonce[0] = htole32(nonce);

    blake2b_update(S, (uint8_t *) &expandedNonce, sizeof(expandedNonce));
}

static void setheader(blake2b_state *ctx, const u8 *input, u32 input_len, int64_t nonce) {
    uint32_t le_N = htole32(WN);
    uint32_t le_K = htole32(WK);
    uchar personal[] = "ZcashPoW01230123";
    memcpy(personal + 8, &le_N, 4);
    memcpy(personal + 12, &le_K, 4);
    blake2b_param P[1];
    P->digest_length = HASHOUT;
    P->key_length = 0;
    P->fanout = 1;
    P->depth = 1;
    P->leaf_length = 0;
    P->node_offset = 0;
    P->node_depth = 0;
    P->inner_length = 0;
    memset(P->reserved, 0, sizeof(P->reserved));
    memset(P->salt, 0, sizeof(P->salt));
    memcpy(P->personal, (const uint8_t *) personal, 16);
    blake2b_init_param(ctx, P);
    blake2b_update(ctx, input, input_len);

    if (nonce >= 0) {
        hashNonce(ctx, (uint32_t)nonce);
    }
}

static void genhash(const blake2b_state *ctx, u32 idx, uchar *hash) {
    blake2b_state state = *ctx;
    u32 leb = htole32(idx / HASHESPERBLAKE);
    blake2b_update(&state, (uchar *) &leb, sizeof(u32));
    uchar blakehash[HASHOUT];
    blake2b_final(&state, blakehash, HASHOUT);
    memcpy(hash, blakehash + (idx % HASHESPERBLAKE) * WN / 8, WN / 8);
}

struct tree {
    tree_t bid_s0_s1;

    // constructor for height 0 trees stores index instead
    tree(const u32 idx) {
        bid_s0_s1 = idx;
    }

    static u32 cantor(u32 s0, u32 s1) {
        return s1 * (s1 + 1) / 2 + s0;
    }

    tree(const u32 bid, const u32 s0, const u32 s1) {
// CANTOR saves 2 bits by Cantor pairing
#ifdef CANTOR
        bid_s0_s1 = (bid << CANTORBITS) | cantor(s0, s1);
#else
        bid_s0_s1 = (((bid << SLOTBITS) | s0) << SLOTBITS) | s1;
#endif
    }

    // retrieve hash index from tree(const u32 idx) constructor
    u32 getindex() const {
        return bid_s0_s1;
    }

    // retrieve bucket index
    u32 bucketid() const {
#ifdef CANTOR
        return bid_s0_s1 >> (2 * SLOTBITS - 2);
#else
        return bid_s0_s1 >> (2*SLOTBITS);
#endif
    }
    // retrieve first slot index
#ifdef CANTOR

    u32 slotid0(u32 s1) const {
        return (bid_s0_s1 & CANTORMASK) - cantor(0, s1);
    }

#else
    u32 slotid0() const {
      return (bid_s0_s1 >> SLOTBITS) & SLOTMASK;
    }
#endif

    // retrieve second slot index
    u32 slotid1() const {
#ifdef CANTOR
        u32 k, q, sqr = 8 * (bid_s0_s1 & CANTORMASK) + 1;;
        // this k=sqrt(sqr) computing loop averages 3.4 iterations out of maximum 9
        for (k = CANTORMAXSQRT; (q = sqr / k) < k; k = (k + q) / 2);
        return (k - 1) / 2;
#else
        return bid_s0_s1 & SLOTMASK;
#endif
    }

    // returns false for trees sharing a child subtree
    bool prob_disjoint(const tree other) const {
#ifdef CANTOR
        if (bucketid() != other.bucketid())
            return true;
        u32 s1 = slotid1(), s0 = slotid0(s1);
        u32 os1 = other.slotid1(), os0 = other.slotid0(os1);
        return s1 != os1 && s0 != os0;
#else
        tree xort(bid_s0_s1 ^ other.bid_s0_s1);
        return xort.bucketid() || (xort.slotid0() && xort.slotid1());
        // next two tests catch much fewer cases and are therefore skipped
        // && slotid0() != other.slotid1() && slotid1() != other.slotid0()
#endif
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
    uchar bytes[sizeof(tree_t)];
};

#define WORDS(bits)    ((bits + TREEBITS-1) / TREEBITS)
#define HASHWORDS0 WORDS(WN - DIGITBITS + RESTBITS)
#define HASHWORDS1 WORDS(WN - 2*DIGITBITS + RESTBITS)

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
typedef u32 bsizes[NBUCKETS];

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

// convenience function
static u32 min(const u32 a, const u32 b) {
    return a < b ? a : b;
}

// size (in bytes) of hash in round 0 <= r < WK
static u32 hashsize(const u32 r) {
    const u32 hashbits = WN - (r + 1) * DIGITBITS + RESTBITS;
    return (hashbits + 7) / 8;
}

// convert bytes into words,rounding up
static u32 hashwords(u32 bytes) {
    return (bytes + TREEBYTES - 1) / TREEBYTES;
}

// manages hash and tree data
struct htalloc {
    bucket0 *heap0;
    bucket1 *heap1;
    u32 alloced;

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

    void *alloc(const u32 n, const u32 sz) {
        void *mem = calloc(n, sz);
        assert(mem);
        alloced += n * sz;
        return mem;
    }
};

// main solver object, shared between all threads
struct equi {
    typedef u32 proof[PROOFSIZE];
    blake_state blake_ctx; // holds blake2b midstate after call to setheadernounce
    htalloc hta;             // holds allocated heaps
    bsizes *nslots;          // counts number of slots used in buckets
    proof *sols;             // store found solutions here (only first MAXSOLS)
    u32 nsols;              // number of solutions found
    const void* user_data;

    equi(const void* userData):user_data(userData) {
        static_assert(sizeof(htunit) == sizeof(tree_t), "");
        static_assert(WK & 1, "K assumed odd in candidate() calling indices1()");
        hta.alloctrees();
        nslots = (bsizes *) hta.alloc(2 * NBUCKETS, sizeof(u32));
        sols = (proof *) hta.alloc(MAXSOLS, sizeof(proof));
    }

    ~equi() {
        hta.dealloctrees();
        free(nslots);
        free(sols);
    }

    // prepare blake2b midstate for new run and initialize counters
    void setheadernonce(const unsigned char *input, const u32 len, int64_t nonce) {
        setheader(&blake_ctx, input, len, nonce);
        nsols = 0;
    }

    // get heap0 bucket size in threadsafe manner
    u32 getslot0(const u32 bucketi) {
        return nslots[0][bucketi]++;
    }

    // get heap1 bucket size in threadsafe manner
    u32 getslot1(const u32 bucketi) {
        return nslots[1][bucketi]++;
    }

    // get old heap0 bucket size and clear it for next round
    u32 getnslots0(const u32 bid) {
        u32 &nslot = nslots[0][bid];
        const u32 n = min(nslot, NSLOTS);
        nslot = 0;
        return n;
    }

    // get old heap1 bucket size and clear it for next round
    u32 getnslots1(const u32 bid) {
        u32 &nslot = nslots[1][bid];
        const u32 n = min(nslot, NSLOTS);
        nslot = 0;
        return n;
    }

    // recognize most (but not all) remaining dupes while Wagner-ordering the indices
    bool orderindices(u32 *indices, u32 size) {
        if (indices[0] > indices[size]) {
            for (u32 i = 0; i < size; i++) {
                const u32 tmp = indices[i];
                indices[i] = indices[size + i];
                indices[size + i] = tmp;
            }
        }
        return false;
    }

    // listindices combines index tree reconstruction with probably dupe test
    bool listindices0(u32 r, const tree t, u32 *indices) {
        if (r == 0) {
            *indices = t.getindex();
            return false;
        }
        const slot1 *buck = hta.heap1[t.bucketid()];
        const u32 size = 1 << --r;
        u32 tagi = hashwords(hashsize(r));
#ifdef CANTOR
        u32 s1 = t.slotid1(), s0 = t.slotid0(s1);
#else
        u32 s1 = t.slotid1(), s0 = t.slotid0();
#endif
        tree t0 = buck[s0][tagi].tag, t1 = buck[s1][tagi].tag;
        return !t0.prob_disjoint(t1)
               || listindices1(r, t0, indices) || listindices1(r, t1, indices + size)
               || orderindices(indices, size) || indices[0] == indices[size];
    }

    // need separate instance for accessing (differently typed) heap1
    bool listindices1(u32 r, const tree t, u32 *indices) {
        const slot0 *buck = hta.heap0[t.bucketid()];
        const u32 size = 1 << --r;
        u32 tagi = hashwords(hashsize(r));
#ifdef CANTOR
        u32 s1 = t.slotid1(), s0 = t.slotid0(s1);
#else
        u32 s1 = t.slotid1(), s0 = t.slotid0();
#endif
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
        u32 soli = nsols++;
        // copy solution into final place
        if (soli < MAXSOLS)
            memcpy(sols[soli], prf, sizeof(proof));
    }

    // thread-local object that precomputes various slot metrics for each round
    // facilitating access to various bits in the variable size slots
    struct htlayout {
        htalloc hta;
        u32 prevhtunits;
        u32 nexthtunits;
        u32 dunits;
        u32 prevbo;

        htlayout(equi *eq, u32 r) : hta(eq->hta), prevhtunits(0), dunits(0) {
            u32 nexthashbytes = hashsize(r);        // number of bytes occupied by round r hash
            nexthtunits = hashwords(nexthashbytes); // number of TREEBITS words taken up by those bytes
            prevbo = 0;                  // byte offset for accessing hash form previous round
            if (r) {     // similar measure for previous round
                u32 prevhashbytes = hashsize(r - 1);
                prevhtunits = hashwords(prevhashbytes);
                prevbo = prevhtunits * sizeof(htunit) - prevhashbytes; // 0-1 or 0-3
                dunits = prevhtunits - nexthtunits; // number of words by which hash shrinks
            }
        }

        // extract remaining bits in digit slots in same bucket still need to collide on
        u32 getxhash0(const htunit *slot) const {
#if DIGITBITS % 8 == 4 && RESTBITS == 4
            return slot->bytes[prevbo] >> 4;
#elif DIGITBITS % 8 == 4 && RESTBITS == 8
            return (slot->bytes[prevbo] & 0xf) << 4 | slot->bytes[prevbo+1] >> 4;
#elif DIGITBITS % 8 == 4 && RESTBITS == 10
            return (slot->bytes[prevbo] & 0x3f) << 4 | slot->bytes[prevbo + 1] >> 4;
#elif DIGITBITS % 8 == 0 && RESTBITS == 4
            return slot->bytes[prevbo] & 0xf;
#elif RESTBITS == 0
            return 0;
#else
#error not implemented
#endif
        }

        // similar but accounting for possible change in hashsize modulo 4 bits
        u32 getxhash1(const htunit *slot) const {
#if DIGITBITS % 4 == 0 && RESTBITS == 4
            return slot->bytes[prevbo] & 0xf;
#elif DIGITBITS % 4 == 0 && RESTBITS == 8
            return slot->bytes[prevbo];
#elif DIGITBITS % 4 == 0 && RESTBITS == 10
            return (slot->bytes[prevbo] & 0x3) << 8 | slot->bytes[prevbo + 1];
#elif RESTBITS == 0
            return 0;
#else
#error not implemented
#endif
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
        #if NSLOTS > 64
#error cant use XBITMAP with more than 64 slots
#endif
        u64 xhashmap[NRESTS];
        u64 xmap;
#else
        // This maintains NRESTS = 2^RESTBITS lists whose starting slot
        // are in xhashslots[] and where subsequent (next-lower-numbered)
        // slots in each list are found through nextxhashslot[]
        // since 0 is already a valid slot number, use ~0 as nil value
#if RESTBITS <= 6
        typedef uchar xslot;
#else
        typedef u16 xslot;
#endif
        static const xslot xnil = ~0;
        xslot xhashslots[NRESTS];
        xslot nextxhashslot[NSLOTS];
        xslot nextslot;
#endif
        u32 s0;

        void clear() {
#ifdef XBITMAP
            memset(xhashmap, 0, NRESTS * sizeof(u64));
#else
            memset(xhashslots, xnil, NRESTS * sizeof(xslot));
            memset(nextxhashslot, xnil, NSLOTS * sizeof(xslot));
#endif
        }

        void addslot(u32 s1, u32 xh) {
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

        u32 slot() {
#ifdef XBITMAP
            const u32 ffs = __builtin_ffsll(xmap);
            s0 += ffs; xmap >>= ffs;
#else
            nextslot = nextxhashslot[s0 = nextslot];
#endif
            return s0;
        }
    };

#ifndef NBLAKES
#define NBLAKES 1
#endif

// number of hashes extracted from NBLAKES blake2b outputs
    static const u32 HASHESPERBLOCK = NBLAKES * HASHESPERBLAKE;
// number of blocks of parallel blake2b calls
    static const u32 NBLOCKS = (NHASHES + HASHESPERBLOCK - 1) / HASHESPERBLOCK;

    void digit0() {
        htlayout htl(this, 0);
        const u32 hashbytes = hashsize(0);
        uchar hashes[NBLAKES * 64];
        blake_state state0 = blake_ctx;  // local copy on stack can be copied faster
        for (u32 block = 0; block < NBLOCKS; block++) {
#if NBLAKES == 1
            blake_state state = state0;  // make another copy since blake2b_final modifies it
            u32 leb = htole32(block);
            blake2b_update(&state, (uchar *) &leb, sizeof(u32));
            blake2b_final(&state, hashes, HASHOUT);
#else
#error not implemented
#endif
            for (u32 i = 0; i < NBLAKES; i++) {
                for (u32 j = 0; j < HASHESPERBLAKE; j++) {
                    const uchar *ph = hashes + i * 64 + j * WN / 8;
                    // figure out bucket for this hash by extracting leading BUCKBITS bits
#if BUCKBITS <= 8
                    const u32 bucketid = (u32)(ph[0] >> (8-BUCKBITS));
#elif BUCKBITS > 8 && BUCKBITS <= 16
                    const u32 bucketid = ((u32) ph[0] << (BUCKBITS - 8)) | ph[1] >> (16 - BUCKBITS);
#elif BUCKBITS > 16
                    const u32 bucketid = ((((u32)ph[0] << 8) | ph[1]) << (BUCKBITS-16)) | ph[2] >> (24-BUCKBITS);
#else
#error not implemented
#endif
                    // grab next available slot in that bucket
                    const u32 slot = getslot0(bucketid);
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

    void digitodd(const u32 r) {
        htlayout htl(this, r);
        collisiondata cd;
        // threads process buckets in round-robin fashion
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear(); // could have made this the constructor, and declare here
            slot0 *buck = htl.hta.heap0[bucketid]; // point to first slot of this bucket
            u32 bsize = getnslots0(bucketid);    // grab and reset bucket size
            for (u32 s1 = 0; s1 < bsize; s1++) {   // loop over slots
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htl.getxhash0(slot1));// identify list of previous colliding slots
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (htl.equal(slot0, slot1)) {     // expect difference in last 32 bits unless duped
                        continue;
                    }
                    u32 xorbucketid;                   // determine bucket for s0 xor s1
                    const uchar *bytes0 = slot0->bytes, *bytes1 = slot1->bytes;
#if WN == 200 && BUCKBITS == 12 && RESTBITS == 8
                    xorbucketid = (((u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) & 0xf) << 8)
                                       | (bytes0[htl.prevbo+2] ^ bytes1[htl.prevbo+2]);
#elif WN == 200 && BUCKBITS == 10 && RESTBITS == 10
                    xorbucketid = (((u32) (bytes0[htl.prevbo + 1] ^ bytes1[htl.prevbo + 1]) & 0xf) << 6)
                                  | (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) >> 2;
#elif WN % 24 == 0 && BUCKBITS == 20 && RESTBITS == 4
                    xorbucketid = ((((u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) << 8)
                                        | (bytes0[htl.prevbo+2] ^ bytes1[htl.prevbo+2])) << 4)
                                        | (bytes0[htl.prevbo+3] ^ bytes1[htl.prevbo+3]) >> 4;
#elif WN == 96 && BUCKBITS == 12 && RESTBITS == 4
                    xorbucketid = ((u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) << 4)
                                      | (bytes0[htl.prevbo+2] ^ bytes1[htl.prevbo+2]) >> 4;
#elif WN == 48 && BUCKBITS == 4 && RESTBITS == 4
                    xorbucketid = (u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) >> 4;
#else
#error not implemented
#endif
                    // grab next available slot in that bucket
                    const u32 xorslot = getslot1(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    // start of slot for s0 ^ s1
                    htunit *xs = htl.hta.heap1[xorbucketid][xorslot];
                    // store xor of hashes possibly minus initial 0 word due to collision
                    for (u32 i = htl.dunits; i < htl.prevhtunits; i++)
                        xs++->word = slot0[i].word ^ slot1[i].word;
                    // store tree node right after hash
                    xs->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digiteven(const u32 r) {
        htlayout htl(this, r);
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot1 *buck = htl.hta.heap1[bucketid];
            u32 bsize = getnslots1(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htl.getxhash1(slot1));
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (htl.equal(slot0, slot1)) {
                        continue;
                    }
                    u32 xorbucketid;
                    const uchar *bytes0 = slot0->bytes, *bytes1 = slot1->bytes;
#if WN == 200 && BUCKBITS == 12 && RESTBITS == 8
                    xorbucketid = ((u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) << 4)
                                      | (bytes0[htl.prevbo+2] ^ bytes1[htl.prevbo+2]) >> 4;
#elif WN == 200 && BUCKBITS == 10 && RESTBITS == 10
                    xorbucketid = ((u32) (bytes0[htl.prevbo + 2] ^ bytes1[htl.prevbo + 2]) << 2)
                                  | (bytes0[htl.prevbo + 3] ^ bytes1[htl.prevbo + 3]) >> 6;
#elif WN % 24 == 0 && BUCKBITS == 20 && RESTBITS == 4
                    xorbucketid = ((((u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) << 8)
                                        | (bytes0[htl.prevbo+2] ^ bytes1[htl.prevbo+2])) << 4)
                                        | (bytes0[htl.prevbo+3] ^ bytes1[htl.prevbo+3]) >> 4;
#elif WN == 96 && BUCKBITS == 12 && RESTBITS == 4
                    xorbucketid = ((u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) << 4)
                                      | (bytes0[htl.prevbo+2] ^ bytes1[htl.prevbo+2]) >> 4;
#elif WN == 48 && BUCKBITS == 4 && RESTBITS == 4
                    xorbucketid = (u32)(bytes0[htl.prevbo+1] ^ bytes1[htl.prevbo+1]) >> 4;
#else
#error not implemented
#endif
                    const u32 xorslot = getslot0(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = htl.hta.heap0[xorbucketid][xorslot];
                    for (u32 i = htl.dunits; i < htl.prevhtunits; i++)
                        xs++->word = slot0[i].word ^ slot1[i].word;
                    xs->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

#if WN == 200 && WK == 9

    // functions digit1 through digit9 are unrolled versions specific to the
    // (N=200,K=9) parameters with 10 RESTBITS
    void digit1() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot0 *buck = heaps.heap0[bucketid];
            u32 bsize = getnslots0(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htobe32(slot1->word) >> 20 & 0x3ff);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (slot0[5].word == slot1[5].word) {
                        continue;
                    }
                    u32 xorbucketid = htobe32(slot0->word ^ slot1->word) >> 10 & BUCKMASK;
                    const u32 xorslot = getslot1(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    u64 *x = (u64 *) heaps.heap1[xorbucketid][xorslot];
                    u64 *x0 = (u64 *) slot0, *x1 = (u64 *) slot1;
                    *x++ = x0[0] ^ x1[0];
                    *x++ = x0[1] ^ x1[1];
                    *x++ = x0[2] ^ x1[2];
                    ((htunit *) x)->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit2() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot1 *buck = heaps.heap1[bucketid];
            u32 bsize = getnslots1(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htobe32(slot1->word) & 0x3ff);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (slot0[5].word == slot1[5].word) {
                        continue;
                    }
                    u32 xor1 = slot0[1].word ^slot1[1].word;
                    u32 xorbucketid = htobe32(xor1) >> 22;
                    const u32 xorslot = getslot0(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = heaps.heap0[xorbucketid][xorslot];
                    xs++->word = xor1;
                    u64 *x = (u64 *) xs, *x0 = (u64 *) slot0, *x1 = (u64 *) slot1;
                    *x++ = x0[1] ^ x1[1];
                    *x++ = x0[2] ^ x1[2];
                    ((htunit *) x)->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit3() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot0 *buck = heaps.heap0[bucketid];
            u32 bsize = getnslots0(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htobe32(slot1->word) >> 12 & 0x3ff);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (slot0[4].word == slot1[4].word) {
                        continue;
                    }
                    u32 xor0 = slot0->word ^slot1->word;
                    u32 xorbucketid = htobe32(xor0) >> 2 & BUCKMASK;
                    const u32 xorslot = getslot1(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = heaps.heap1[xorbucketid][xorslot];
                    xs++->word = xor0;
                    u64 *x = (u64 *) xs, *x0 = (u64 *) (slot0 + 1), *x1 = (u64 *) (slot1 + 1);
                    *x++ = x0[0] ^ x1[0];
                    *x++ = x0[1] ^ x1[1];
                    ((htunit *) x)->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit4() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot1 *buck = heaps.heap1[bucketid];
            u32 bsize = getnslots1(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, (slot1->bytes[3] & 0x3) << 8 | slot1->bytes[4]);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (slot0[4].word == slot1[4].word) {
                        continue;
                    }
                    u32 xorbucketid = htobe32(slot0[1].word ^ slot1[1].word) >> 14 & BUCKMASK;
                    const u32 xorslot = getslot0(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    u64 *x = (u64 *) heaps.heap0[xorbucketid][xorslot];
                    u64 *x0 = (u64 *) (slot0 + 1), *x1 = (u64 *) (slot1 + 1);
                    *x++ = x0[0] ^ x1[0];
                    *x++ = x0[1] ^ x1[1];
                    ((htunit *) x)->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit5() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot0 *buck = heaps.heap0[bucketid];
            u32 bsize = getnslots0(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htobe32(slot1->word) >> 4 & 0x3ff);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (slot0[3].word == slot1[3].word) {
                        continue;
                    }
                    u32 xor1 = slot0[1].word ^slot1[1].word;
                    u32 xorbucketid = (((u32) (slot0->bytes[3] ^ slot1->bytes[3]) & 0xf)
                            << 6) | (xor1 >> 2 & 0x3f);
                    const u32 xorslot = getslot1(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = heaps.heap1[xorbucketid][xorslot];
                    xs++->word = xor1;
                    u64 *x = (u64 *) xs, *x0 = (u64 *) slot0, *x1 = (u64 *) slot1;
                    *x++ = x0[1] ^ x1[1];
                    ((htunit *) x)->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit6() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot1 *buck = heaps.heap1[bucketid];
            u32 bsize = getnslots1(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htobe32(slot1->word) >> 16 & 0x3ff);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    if (slot0[2].word == slot1[2].word) {
                        continue;
                    }
                    u32 xor0 = slot0->word ^slot1->word;
                    u32 xorbucketid = htobe32(xor0) >> 6 & BUCKMASK;
                    const u32 xorslot = getslot0(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = heaps.heap0[xorbucketid][xorslot];
                    xs++->word = xor0;
                    u64 *x = (u64 *) xs, *x0 = (u64 *) (slot0 + 1), *x1 = (u64 *) (slot1 + 1);
                    *x++ = x0[0] ^ x1[0];
                    ((htunit *) x)->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit7() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot0 *buck = heaps.heap0[bucketid];
            u32 bsize = getnslots0(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, (slot1->bytes[3] & 0x3f) << 4 | slot1->bytes[4] >> 4);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    u32 xor2 = slot0[2].word ^slot1[2].word;
                    if (!xor2) {
                        continue;
                    }
                    u32 xor1 = slot0[1].word ^slot1[1].word;
                    u32 xorbucketid = htobe32(xor1) >> 18 & BUCKMASK;
                    const u32 xorslot = getslot1(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = heaps.heap1[xorbucketid][xorslot];
                    xs++->word = xor1;
                    xs++->word = xor2;
                    xs->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

    void digit8() {
        htalloc heaps = hta;
        collisiondata cd;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot1 *buck = heaps.heap1[bucketid];
            u32 bsize = getnslots1(bucketid);
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htobe32(slot1->word) >> 8 & 0x3ff);
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
                    const htunit *slot0 = buck[s0];
                    u32 xor1 = slot0[1].word ^slot1[1].word;
                    if (!xor1) {
                        continue;
                    }
                    u32 xorbucketid = ((u32) (slot0->bytes[3] ^ slot1->bytes[3]) << 2)
                                      | (xor1 >> 6 & 0x3);
                    const u32 xorslot = getslot0(xorbucketid);
                    if (xorslot >= NSLOTS) {
                        continue;
                    }
                    htunit *xs = heaps.heap0[xorbucketid][xorslot];
                    xs++->word = xor1;
                    xs->tag = tree(bucketid, s0, s1);
                }
            }
        }
    }

#endif

    // final round looks simpler
    void digitK() {
        collisiondata cd;
        htlayout htl(this, WK);
        u32 nc = 0;
        for (u32 bucketid = 0; bucketid < NBUCKETS; bucketid++) {
            cd.clear();
            slot0 *buck = htl.hta.heap0[bucketid];   // assume WK odd
            u32 bsize = getnslots0(bucketid);      // assume WK odd
            for (u32 s1 = 0; s1 < bsize; s1++) {
                const htunit *slot1 = buck[s1];
                cd.addslot(s1, htl.getxhash0(slot1));  // assume WK odd
                for (; cd.nextcollision();) {
                    const u32 s0 = cd.slot();
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
        digit0();
        if (equihashProxy(user_data, 0))
            return 0;
#if WN == 200 && WK == 9 && RESTBITS == 10
        digit1();
        if (equihashProxy(user_data, 0))
            return 0;
        digit2();
        if (equihashProxy(user_data, 0))
            return 0;
        digit3();
        if (equihashProxy(user_data, 0))
            return 0;
        digit4();
        if (equihashProxy(user_data, 0))
            return 0;
        digit5();
        if (equihashProxy(user_data, 0))
            return 0;
        digit6();
        if (equihashProxy(user_data, 0))
            return 0;
        digit7();
        if (equihashProxy(user_data, 0))
            return 0;
        digit8();
        if (equihashProxy(user_data, 0))
            return 0;
#else
        for (u32 r = 1; r < WK; r++) {
            r&1 ? digitodd(r) : digiteven(r);
            if (equihashProxy(user_data, 0))
                return 0;
        }
#endif
        digitK();
        if (equihashProxy(user_data, 0))
            return 0;
        return 1;
    }
};

#define COMPRESSED_SOL_SIZE (PROOFSIZE * (DIGITBITS + 1) / 8)

void compress_solution(const u32* sol, uchar *csol) {
    uchar b;
    for (u32 i = 0, j = 0, bits_left = DIGITBITS + 1; j < COMPRESSED_SOL_SIZE; csol[j++] = b) {
        if (bits_left >= 8) {
            // Read next 8 bits, stay at same sol index
            b = sol[i] >> (bits_left -= 8);
        } else { // less than 8 bits to read
            // Read remaining bits and shift left to make space for next sol index
            b = sol[i];
            b <<= (8 - bits_left); // may also set b=0 if bits_left was 0, which is fine
            // Go to next sol index and read remaining bits
            bits_left += DIGITBITS + 1 - 8;
            b |= sol[++i] >> bits_left;
        }
    }
}

static int verifyrec(const blake2b_state *ctx, u32 *indices, uchar *hash, int r) {
    if (r == 0) {
        genhash(ctx, *indices, hash);
        return POW_OK;
    }
    u32 *indices1 = indices + (1 << (r - 1));
    if (*indices >= *indices1)
        return POW_OUT_OF_ORDER;

    uchar hash0[WN / 8], hash1[WN / 8];
    int vrf0 = verifyrec(ctx, indices, hash0, r - 1);
    if (vrf0 != POW_OK)
        return vrf0;

    int vrf1 = verifyrec(ctx, indices1, hash1, r - 1);
    if (vrf1 != POW_OK)
        return vrf1;

    for (int i = 0; i < WN / 8; i++)
        hash[i] = hash0[i] ^ hash1[i];

    int i, b = r < WK ? r * DIGITBITS : WN;

    for (i = 0; i < b / 8; i++)
        if (hash[i])
            return POW_NONZERO_XOR;

    if ((b % 8) && hash[i] >> (8 - (b % 8)))
        return POW_NONZERO_XOR;
    return POW_OK;
}

int verify(u32* indices, u32 proofsize, const unsigned char *input, const u32 input_len, int64_t nonce) {
    if (input_len > HEADERNONCELEN)
        return POW_INVALID_HEADER_LENGTH;
    if (proofsize != PROOFSIZE)
        return POW_SOL_SIZE_MISMATCH;
    if (duped(indices))
        return POW_DUPLICATE;
    blake2b_state ctx;
    setheader(&ctx, input, input_len, nonce);
    uchar hash[WN / 8];
    return verifyrec(&ctx, indices, hash, WK);
}

int solve(const unsigned char* input, u32 input_len, int64_t nonce, const void* userData) {
    equi eq(userData);
    eq.setheadernonce(input, input_len, nonce);
    eq.worker();

    u32 maxsols = std::min(MAXSOLS, eq.nsols);
    uchar csol[COMPRESSED_SOL_SIZE];

    for (u32 nsols = 0; nsols < maxsols; nsols++) {
        if (verify(eq.sols[nsols], PROOFSIZE, input, input_len, nonce) != POW_OK) {
            continue;
        }

        compress_solution(eq.sols[nsols], csol);

        if (equihashProxy(eq.user_data, csol)) {
            return 1;
        }
    }

    return eq.nsols;
}

#ifdef __EQUI_MAIN__
extern "C" int equihashProxy(const void* blockData, void* solution) {
    return 0;
}

static int hextobyte(const char *x) {
    u32 b = 0;
    for (int i = 0; i < 2; i++) {
        uchar c = tolower(x[i]);
        assert(isxdigit(c));
        b = (b << 4) | (c - (c >= '0' && c <= '9' ? '0' : ('a' - 10)));
    }
    return b;
}

int main(int argc, char **argv) {
    int nonce = 0;
    int range = 1;
    bool showsol = false;
    bool compress_sol = false;
    bool print_decimal = false;
    const char *header = "";
    const char *hex = "";
    int c;
    while ((c = getopt(argc, argv, "h:n:r:t:x:dsc")) != -1) {
        switch (c) {
            case 'h':
                header = optarg;
                break;
            case 'n':
                nonce = atoi(optarg);
                break;
            case 'r':
                range = atoi(optarg);
                break;
            case 's':
                showsol = true;
                break;
            case 'x':
                hex = optarg;
                break;
            case 'c':
                compress_sol = true;
                break;
            case 'd':
                print_decimal = true;
                break;
        }
    }
#ifndef XWITHASH
    if (sizeof(tree) > 4)
        printf("WARNING: please compile with -DXWITHASH to shrink tree!\n");
#endif

    printf("Looking for wagner-tree on (\"%s\",%d", hex ? "0x..." : header, nonce);
    if (range > 1)
        printf("-%d", nonce + range - 1);
    printf(") with %d %d-bit digits\n", NDIGITS, DIGITBITS);
    equi eq(nullptr);
    printf("Using 2^%d buckets, %dMB of memory, and %d-way blake2b, print decimal %d\n", BUCKBITS, 1 + eq.hta.alloced / 0x100000,
           NBLAKES, print_decimal);

    u32 sumnsols = 0;
//    unsigned char headernonce[HEADERNONCELEN];
//    u32 hdrlen = strlen(header);
//    if (*hex) {
//        assert(strlen(hex) == 2 * HEADERNONCELEN);
//        for (u32 i = 0; i < HEADERNONCELEN; i++)
//            headernonce[i] = hextobyte(&hex[2 * i]);
//    } else {
//        memcpy(headernonce, header, hdrlen);
//        memset(headernonce + hdrlen, 0, sizeof(headernonce) - hdrlen);
//    }
    for (int r = 0; r < range; r++) {
//        ((u32 *) headernonce)[27] = htole32(nonce + r);
//        eq.setheadernonce(headernonce, sizeof(headernonce), nonce + r);
        eq.setheadernonce((const u8*)header, strlen(header), nonce + r);
        eq.worker();

        u32 nsols, maxsols = min(MAXSOLS, eq.nsols);
        for (nsols = 0; nsols < maxsols; nsols++) {
            if (showsol) {
                printf("Solution");
                if (compress_sol) {
                    printf(" ");
                    uchar csol[COMPRESSED_SOL_SIZE];
                    compress_solution(eq.sols[nsols], csol);
                    for (u32 i = 0; i < COMPRESSED_SOL_SIZE; ++i) {
                        printf("%02hhx", csol[i]);
                    }
                } else {
                    for (u32 i = 0; i < PROOFSIZE; i++)
                        printf((print_decimal ? " %ld" : " %jx"), (uintmax_t) eq.sols[nsols][i]);
                }
                printf("\n");
            }
        }
        printf("%d solutions... ", nsols);
        sumnsols += nsols;
    }
    printf("\n%d total solutions\n", sumnsols);
    return 0;
}

#endif
