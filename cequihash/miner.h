// Equihash solver
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
#include <stdlib.h> // for function qsort

#include "blake2.h"

typedef blake2b_state blake_state;
typedef uint8_t u8;
typedef uint32_t u32;
typedef uint16_t u16;
typedef uint64_t u64;
typedef unsigned char uchar;

#ifndef HEADERNONCELEN
#define HEADERNONCELEN ((u32)180)
#endif

#define JOIN0(a,b)   a##b
#define JOIN(a,b)   JOIN0(a,b)

#if !(defined(CONF_48_5) || defined(CONF_96_5) || defined(CONF_144_5) || defined(CONF_200_9))
#define CONF_200_9
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

#endif // __EQUI_MINER_H