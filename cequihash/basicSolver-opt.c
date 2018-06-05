/*
 * Copyright (c) 2016 abc at openwall dot com
 * Copyright (c) 2016 Jack Grigg
 * Copyright (c) 2016 The Zcash developers
 *
 * Distributed under the MIT software license, see the accompanying
 * file COPYING or http://www.opensource.org/licenses/mit-license.php.
 *
 * Port to C of C++ implementation of the Equihash Proof-of-Work
 * algorithm from zcashd.
 */

#define _BSD_SOURCE
#define _GNU_SOURCE

#include <endian.h>
#include <stdint.h>
#ifdef __INCLUDE_MAIN__
#include <stdio.h>
#endif
#include <stdlib.h>
#include <stdbool.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>
#include <assert.h>

#include "cequihash.h"
#include "blake2.h"
#include "blake2-impl.h"

#define swap(a, b) \
    do { __typeof__(a) __tmp = (a); (a) = (b); (b) = __tmp; } while (0)

int equihash_proxy(void*, void*);

/* Writes Zcash personalization string. */
static void zcashPerson(uint8_t *person, const int n, const int k) {
    memcpy(person, "ZcashPoW", 8);
    *(uint32_t *) (person + 8) = htole32(n);
    *(uint32_t *) (person + 12) = htole32(k);
}

static void digestInit(blake2b_state *S, const int n, const int k) {
    blake2b_param P[1];

    memset(P, 0, sizeof(blake2b_param));
    P->fanout = 1;
    P->depth = 1;
    P->digest_length = (512 / n) * n / 8;
    zcashPerson(P->personal, n, k);
    blake2b_init_param(S, P);
}

static void ehIndexToArray(const uint32_t i, uint8_t *array) {
    const uint32_t be_i = htobe32(i);

    memcpy(array, &be_i, sizeof(be_i));
}

uint32_t arrayToEhIndex(const uint8_t *array) {
    return be32toh(*(uint32_t *) array);
}

static void generateHash(blake2b_state *S, const uint32_t g, uint8_t *hash, const size_t hashLen) {
    const uint32_t le_g = htole32(g);
    blake2b_state digest = *S; /* copy */

    blake2b_update(&digest, (uint8_t *) &le_g, sizeof(le_g));
    blake2b_final(&digest, hash, hashLen);
}

/* https://github.com/zcash/zcash/issues/1175 */
static void expandArray(const unsigned char *in, const size_t in_len,
                        unsigned char *out, const size_t out_len,
                        const size_t bit_len, const size_t byte_pad) {
    assert(bit_len >= 8);
    assert(8 * sizeof(uint32_t) >= 7 + bit_len);

    const size_t out_width = (bit_len + 7) / 8 + byte_pad;
    assert(out_len == 8 * out_width * in_len / bit_len);

    const uint32_t bit_len_mask = ((uint32_t) 1 << bit_len) - 1;

    // The acc_bits least-significant bits of acc_value represent a bit sequence
    // in big-endian order.
    size_t acc_bits = 0;
    uint32_t acc_value = 0;

    size_t j = 0;
    for (size_t i = 0; i < in_len; i++) {
        acc_value = (acc_value << 8) | in[i];
        acc_bits += 8;

        // When we have bit_len or more bits in the accumulator, write the next
        // output element.
        if (acc_bits >= bit_len) {
            acc_bits -= bit_len;
            for (size_t x = 0; x < byte_pad; x++) {
                out[j + x] = 0;
            }
            for (size_t x = byte_pad; x < out_width; x++) {
                out[j + x] = (acc_value >> (acc_bits + (8 * (out_width - x - 1))))
                        & ((bit_len_mask >> (8 * (out_width - x - 1))) & 0xFF);
            }
            j += out_width;
        }
    }
}

static void compressArray(const unsigned char *in, const size_t in_len,
                          unsigned char *out, const size_t out_len,
                          const size_t bit_len, const size_t byte_pad) {
    assert(bit_len >= 8);
    assert(8 * sizeof(uint32_t) >= 7 + bit_len);

    const size_t in_width = (bit_len + 7) / 8 + byte_pad;
    assert(out_len == bit_len * in_len / (8 * in_width));

    const uint32_t bit_len_mask = ((uint32_t) 1 << bit_len) - 1;

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
                acc_value = acc_value |
                        ((in[j + x] & ((bit_len_mask >> (8 * (in_width - x - 1))) & 0xFF)) << (8 * (in_width - x - 1)));
            }
            j += in_width;
            acc_bits += bit_len;
        }

        acc_bits -= 8;
        out[i] = (acc_value >> acc_bits) & 0xFF;
    }
}

static int compareSR(const void *p1, const void *p2, void *arg) {
    return memcmp(p1, p2, *(int *) arg) < 0;
}

static int distinctSortedArrays(const uint32_t *a, const uint32_t *b, const size_t len) {
    int i = len - 1, j = len - 1;
    uint32_t prev;

    prev = (a[i] >= b[j]) ? a[i--] : b[j--];
    while (j >= 0 && i >= 0) {
        uint32_t acc = (a[i] >= b[j]) ? a[i--] : b[j--];
        if (acc == prev)
            return 0;
        prev = acc;
    }
    return 1;
}

// Checks if the intersection of a.indices and b.indices is empty
static int distinctIndices(const uint8_t *a, const uint8_t *b, const size_t len, const size_t lenIndices) {
    return distinctSortedArrays((uint32_t *) (a + len), (uint32_t *) (b + len), lenIndices / sizeof(uint32_t));
}

static int hasCollision(const uint8_t *a, const uint8_t *b, const size_t len) {
    return memcmp(a, b, len) == 0;
}

static int getIndices(const uint8_t *hash, size_t len, size_t lenIndices, size_t cBitLen,
                      uint8_t *data, size_t maxLen) {
    assert(((cBitLen + 1) + 7) / 8 <= sizeof(uint32_t));
    size_t minLen = (cBitLen + 1) * lenIndices / (8 * sizeof(uint32_t));
    size_t bytePad = sizeof(uint32_t) - ((cBitLen + 1) + 7) / 8;
    if (minLen > maxLen)
        return -1;
    if (data)
        compressArray(hash + len, lenIndices, data, minLen, cBitLen + 1, bytePad);
    return minLen;
}

static void joinSortedArrays(uint32_t *dst, const uint32_t *a, const uint32_t *b, const size_t len) {
    int i = len - 1, j = len - 1, k = len * 2;

    while (k > 0)
        dst[--k] = (j < 0 || (i >= 0 && a[i] >= b[j])) ? a[i--] : b[j--];
}

static void combineRows(uint8_t *hash, const uint8_t *a, const uint8_t *b,
                        const size_t len, const size_t lenIndices, const int trim) {
    for (int i = trim; i < len; i++)
        hash[i - trim] = a[i] ^ b[i];

    joinSortedArrays((uint32_t *) (hash + len - trim),
                     (uint32_t *) (a + len), (uint32_t *) (b + len),
                     lenIndices / sizeof(uint32_t));
}

static int isZero(const uint8_t *hash, size_t len) {
    // This doesn't need to be constant time.
    for (int i = 0; i < len; i++) {
        if (hash[i] != 0)
            return 0;
    }
    return 1;
}

static int basicSolve(blake2b_state *digest,
                      const int n, const int k,
                      void *validBlockData) {
    const int collisionBitLength = n / (k + 1);
    const int collisionByteLength = (collisionBitLength + 7) / 8;
    const int hashLength = (k + 1) * collisionByteLength;
    const int indicesPerHashOutput = 512 / n;
    const int hashOutput = indicesPerHashOutput * n / 8;
    const int fullWidth = 2 * collisionByteLength + sizeof(uint32_t) * (1 << (k - 1));
    const int initSize = 1 << (collisionBitLength + 1);
    const int equihashSolutionSize = (1 << k) * (n / (k + 1) + 1) / 8;

    uint8_t hash[fullWidth];
    size_t x_room = initSize * sizeof(hash);
    size_t xc_room = initSize * sizeof(hash);
    uint8_t *x = malloc(x_room);
    uint8_t *xc = malloc(xc_room); // merge array
    assert(x);
    assert(xc);
#define X(y)  (x  + (hashLen                       + lenIndices)     * (y))
#define Xc(y) (xc + (hashLen - collisionByteLength + lenIndices * 2) * (y))

    uint8_t tmpHash[hashOutput];
    uint32_t x_size = 0, xc_size = 0;
    size_t hashLen = hashLength;       /* Offset of indices array;
					     shortens linearly by collisionByteLength. */
    size_t lenIndices = sizeof(uint32_t); /* Byte length of indices array;
					     doubles with every round. */

    for (uint32_t g = 0; x_size < initSize; g++) {
        generateHash(digest, g, tmpHash, hashOutput);
        //if (g == 0) dump_hex(tmpHash, hashOutput);
        for (uint32_t i = 0; i < indicesPerHashOutput && x_size < initSize; i++) {
            expandArray(tmpHash + (i * n / 8), n / 8,
                        hash, hashLength,
                        collisionBitLength, 0);
            ehIndexToArray(g * indicesPerHashOutput + i, hash + hashLength);
            memcpy(X(x_size), hash, hashLen + lenIndices);
            ++x_size;
        }
    }

    for (int r = 1; r < k && x_size > 0; r++) {
        qsort_r(x, x_size, hashLen + lenIndices, compareSR, (int *) &collisionByteLength);

        for (int i = 0; i < x_size - 1;) {
            // 2b) Find next set of unordered pairs with collisions on the next n/(k+1) bits
            int j = 1;
            while (i + j < x_size && hasCollision(X(i), X(i + j), collisionByteLength)) {
                j++;
            }
            /* Found partially collided values range between i and i+j. */

            // 2c) Calculate tuples (X_i ^ X_j, (i, j))
            for (int l = 0; l < j - 1; l++) {
                for (int m = l + 1; m < j; m++) {
                    if (distinctIndices(X(i + l), X(i + m), hashLen, lenIndices)) {
                        combineRows(Xc(xc_size), X(i + l), X(i + m), hashLen, lenIndices, collisionByteLength);
                        ++xc_size;
                        if (Xc(xc_size) >= (xc + xc_room)) {
                            xc_room += 100000000;
                            xc = realloc(xc, xc_room);
                            assert(xc);
                        }
                    }
                }
            }

            /* Skip processed block to the next. */
            i += j;
        }

        hashLen -= collisionByteLength;
        lenIndices *= 2;

        /* swap arrays */
        swap(x, xc);
        swap(x_room, xc_room);
        x_size = xc_size;
        xc_size = 0;
    } /* step 2 */

    // k+1) Find a collision on last 2n(k+1) bits
    int solnr = 0;
    if (x_size > 1) {
        qsort_r(x, x_size, (hashLen + lenIndices), compareSR, (int *) &hashLen);
        for (int i = 0; i < x_size - 1;) {
            int j = 1;
            while (i + j < x_size && hasCollision(X(i), X(i + j), hashLen)) {
                j++;
            }

            for (int l = 0; l < j - 1; l++) {
                for (int m = l + 1; m < j; m++) {
                    combineRows(Xc(xc_size), X(i + l), X(i + m), hashLen, lenIndices, 0);
                    if (isZero(Xc(xc_size), hashLen) && distinctIndices(X(i + l), X(i + m), hashLen, lenIndices)) {
                        uint8_t soln[equihashSolutionSize];
                        int ssize = getIndices(Xc(xc_size), hashLen, 2 * lenIndices, collisionBitLength, soln,
                                               sizeof(soln));
                        ++solnr;

                        assert(equihashSolutionSize == ssize);

                        if (validBlockData && equihash_proxy(validBlockData, soln)) {
                                return solnr;
                        }
                    }
                    ++xc_size;
                    assert(xc_size < xc_room);
                }
            }
            i += j;
        }
    }
    free(x);
    free(xc);
    return solnr;
}

struct validData {
    int n;
    int k;
    blake2b_state *digest;
};

int EquihashValidate(int n, int k, void* hash, void* soln) {
    blake2b_state *digest = (blake2b_state*) hash;
    const int collisionBitLength = n / (k + 1);
    const int collisionByteLength = (collisionBitLength + 7) / 8;
    const int hashLength = (k + 1) * collisionByteLength;
    const int indicesPerHashOutput = 512 / n;
    const int hashOutput = indicesPerHashOutput * n / 8;
    const int equihashSolutionSize = (1 << k) * (n / (k + 1) + 1) / 8;
    const int solnr = 1 << k;
    uint32_t indices[solnr];

    expandArray(soln, equihashSolutionSize, (unsigned char *) &indices, sizeof(indices), collisionBitLength + 1, 1);

    uint8_t vHash[hashLength];
    memset(vHash, 0, sizeof(vHash));
    for (int j = 0; j < solnr; j++) {
        uint8_t tmpHash[hashOutput];
        uint8_t hash[hashLength];
        int i = be32toh(indices[j]);

        generateHash(digest, i / indicesPerHashOutput, tmpHash, hashOutput);
        expandArray(tmpHash + (i % indicesPerHashOutput * n / 8), n / 8, hash, hashLength, collisionBitLength, 0);

        for (int k = 0; k < hashLength; ++k)
            vHash[k] ^= hash[k];
    }

    return isZero(vHash, sizeof(vHash));
}

void* ExpandArray(int n, int k, void* soln) {
    const int solnr = 1 << k;
    const int equihashSolutionSize = (1 << k) * (n / (k + 1) + 1) / 8;
    const int collisionBitLength = n / (k + 1);

    uint32_t indices[solnr];

    expandArray(soln, equihashSolutionSize, (unsigned char *) &indices, sizeof(indices), collisionBitLength + 1, 1);

    void* result = malloc(sizeof(indices));
    memcpy(result, indices, sizeof(indices));

    return result;
}

int basicValidator(void *data, void *soln) {
    const struct validData *v = data;
    const int n = v->n;
    const int k = v->k;
    return EquihashValidate(v->n, v->k, v->digest, soln);
}

int EquihashSolve(void *input, void *validBlockData, int n, int k) {
    //TODO: probably this should be moved somewhere?
    blake2b_state digest[1];
    digestInit(digest, n, k);
    blake2b_update(digest, (const uint8_t *)input, 140);

    return basicSolve(digest, n, k, validBlockData);
}

#ifdef __INCLUDE_MAIN__

static void hashNonce(blake2b_state *S, uint32_t nonce) {
    for (int i = 0; i < 8; i++) {
        uint32_t le = i == 0 ? htole32(nonce) : 0;
        blake2b_update(S, (uint8_t *) &le, sizeof(le));
    }
}

int main(int argc, char **argv) {
    int n = 96;
    int k = 5;
    char *ii = "block header";
    uint32_t nn = 0;
    char *input = NULL;
    int tFlags = 0;
    int opt;

    while ((opt = getopt(argc, argv, "n:k:N:I:t:i:h")) != -1) {
        switch (opt) {
            case 'n':
                n = atoi(optarg);
                break;
            case 'k':
                k = atoi(optarg);
                break;
            case 'N':
                nn = strtoul(optarg, NULL, 0);
                tFlags = 1;
                break;
            case 'I':
                ii = strdup(optarg);
                tFlags = 2;
                break;
            case 'i':
                input = strdup(optarg);
                break;

            case 'h':
            default:
                fprintf(stderr, "Solver CPI API mode:\n");
                fprintf(stderr, "  %s -i input -n N -k K\n", argv[0]);
                fprintf(stderr, "Test vector mode:\n");
                fprintf(stderr, "  %s [-n N] [-k K] [-I string] [-N nonce]\n", argv[0]);
                exit(1);
        }
    }
    if (tFlags && input) {
        fprintf(stderr, "Test vector parameters (-I, -N) cannot be used together with input (-i)\n");
        exit(1);
    }

    if (input) {
        uint8_t block_header[140];
        int fd = open(input, O_RDONLY);
        if (fd == -1) {
            fprintf(stderr, "open: %s: %s\n", input, strerror(errno));
            exit(1);
        }
        int i = read(fd, block_header, sizeof(block_header));
        if (i == -1) {
            fprintf(stderr, "read: %s: %s\n", input, strerror(errno));
            exit(1);
        } else if (i != sizeof(block_header)) {
            fprintf(stderr, "read: %s: Zcash block header is not full\n", input);
            exit(1);
        }
        close(fd);

        int ret = EquihashSolve(block_header, NULL, NULL, n, k);
        exit(ret < 0);
    } else {
        blake2b_state digest[1];
        struct validData valData = {.n = n, .k = k, .digest = digest};
        digestInit(digest, n, k);
        blake2b_update(digest, (uint8_t *) ii, strlen(ii));
        hashNonce(digest, nn);
        basicSolve(digest, n, k, basicValidator, &valData);
    }
}
#endif