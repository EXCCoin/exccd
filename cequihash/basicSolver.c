/*
* Copyright (c) 2018 The ExchangeCoin team
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

#define _DEFAULT_SOURCE
#define _GNU_SOURCE

#include <stdint.h>

#ifdef __INCLUDE_MAIN__
#include <stdio.h>
#endif

#include <stdlib.h>
#include <stdbool.h>
#include <fcntl.h>
#include <errno.h>
#include <assert.h>

#include "cequihash.h"
#include "blake2.h"
#include "blake2-impl.h"

void swap_impl(void *x, void *y, const int size) {
    unsigned char temp[size];
    memcpy(temp, y, size);
    memcpy(y, x, size);
    memcpy(x, temp, size);
}

#define swap(x, y) swap_impl(&x, &y, sizeof(x) == sizeof(y) ? sizeof(x) : -1)


int equihashProxy(void *, void *);

/* Writes Zcash personalization string. */
static void zcashPerson(uint8_t *person, uint32_t n, uint32_t k) {
    memcpy(person, "ZcashPoW", 8);
    *(uint32_t *) (person + 8) = htole32(n);
    *(uint32_t *) (person + 12) = htole32(k);
}

static void digestInit(blake2b_state *S, uint32_t n, uint32_t k) {
    blake2b_param P[1];

    memset(P, 0, sizeof(blake2b_param));
    P->fanout = 1;
    P->depth = 1;
    P->digest_length = (uint8_t) ((512 / n) * n / 8);
    zcashPerson(P->personal, n, k);
    blake2b_init_param(S, P);
}

static void ehIndexToArray(const uint32_t i, uint8_t *array) {
    const uint32_t be_i = htobe32(i);

    memcpy(array, &be_i, sizeof(be_i));
}

#ifdef __INCLUDE_MAIN__
static uint32_t arrayToEhIndex(const uint8_t *array) {
    return be32toh(*(uint32_t *) array);
}
#endif

static void generateHash(blake2b_state *S, const uint32_t g, uint8_t *hash, const uint8_t hashLen) {
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
                out[j + x] = (
                                     // Big-endian
                                     acc_value >> (acc_bits + (8 * (out_width - x - 1)))
                             ) & (
                                     // Apply bit_len_mask across byte boundaries
                                     (bit_len_mask >> (8 * (out_width - x - 1))) & 0xFF
                             );
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
                acc_value = acc_value | (
                        (
                                // Apply bit_len_mask across byte boundaries
                                in[j + x] & ((bit_len_mask >> (8 * (in_width - x - 1))) & 0xFF)
                        ) << (8 * (in_width - x - 1))); // Big-endian
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

// Checks if the intersection of a.indices and b.indices is empty
static int distinctIndices(const uint8_t *a, const uint8_t *b, const size_t len, const size_t lenIndices) {
    for (size_t i = 0; i < lenIndices; i += sizeof(uint32_t)) {
        for (size_t j = 0; j < lenIndices; j += sizeof(uint32_t)) {
            if (memcmp(a + len + i, b + len + j, sizeof(uint32_t)) == 0) {
                return 0;
            }
        }
    }
    return 1;
}

static int hasCollision(const uint8_t *a, const uint8_t *b, const size_t len) {
    return memcmp(a, b, len) == 0;
}

static uint32_t getIndices(const uint8_t *hash, size_t len, size_t lenIndices, size_t cBitLen,
                      uint8_t *data, size_t maxLen) {
    assert(((cBitLen + 1) + 7) / 8 <= sizeof(uint32_t));
    size_t minLen = (cBitLen + 1) * lenIndices / (8 * sizeof(uint32_t));
    size_t bytePad = sizeof(uint32_t) - ((cBitLen + 1) + 7) / 8;
    assert(maxLen >= minLen);
    if (data) {
        compressArray(hash + len, lenIndices, data, minLen, cBitLen + 1, bytePad);
    }
    return minLen;
}

static int indicesBefore(const uint8_t *a, const uint8_t *b, const size_t len, const size_t lenIndices) {
    return memcmp(a + len, b + len, lenIndices) < 0;
}

static void combineRows(uint8_t *hash, const uint8_t *a, const uint8_t *b,
                        const size_t len, const size_t lenIndices, const size_t trim) {
    for (size_t i = trim; i < len; i++) {
        hash[i - trim] = a[i] ^ b[i];
    }
    if (indicesBefore(a, b, len, lenIndices)) {
        memcpy(hash + len - trim, a + len, lenIndices);
        memcpy(hash + len - trim + lenIndices, b + len, lenIndices);
    } else {
        memcpy(hash + len - trim, b + len, lenIndices);
        memcpy(hash + len - trim + lenIndices, a + len, lenIndices);
    }
}

static int isZero(const uint8_t *hash, size_t len) {
    // This doesn't need to be constant time.
    for (size_t i = 0; i < len; i++) {
        if (hash[i] != 0)
            return 0;
    }
    return 1;
}

#define X(y)  (x  + sizeof(hash) * (y))
#define Xc(y) (xc + sizeof(hash) * (y))

static int basicSolve(blake2b_state *digest,
                      const int n, const int k,
                      void *validBlockData) {
    const uint32_t collisionBitLength = n / (k + 1);
    const uint32_t collisionByteLength = (collisionBitLength + 7) / 8;
    const uint32_t hashLength = (k + 1) * collisionByteLength;
    const uint32_t indicesPerHashOutput = 512 / n;
    const uint32_t hashOutput = indicesPerHashOutput * n / 8;
    const uint32_t fullWidth = 2 * collisionByteLength + sizeof(uint32_t) * (1 << (k - 1));
    const uint32_t initSize = 1 << (collisionBitLength + 1);
    const uint32_t equihashSolutionSize = (1 << k) * (n / (k + 1) + 1) / 8;

    uint8_t hash[fullWidth];
    size_t x_room = initSize;
    size_t xc_room = initSize;
    uint8_t *x = malloc(x_room * sizeof(hash));
    uint8_t *xc = malloc(xc_room * sizeof(hash)); // merge list
    assert(x);
    assert(xc);

    uint8_t tmpHash[hashOutput];
    uint32_t x_size = 0, xc_size = 0;

    for (uint32_t g = 0; x_size < initSize; g++) {
        generateHash(digest, g, tmpHash, hashOutput);

        if (validBlockData && equihashProxy(validBlockData, 0)) {
            goto cleanup;
        }

        for (uint32_t i = 0; i < indicesPerHashOutput && x_size < initSize; i++) {
            expandArray(tmpHash + (i * n / 8), n / 8,
                        hash, hashLength,
                        collisionBitLength, 0);
            ehIndexToArray(g * indicesPerHashOutput + i, hash + hashLength);
            memcpy(X(x_size), hash, hashLength + sizeof(uint32_t));
            ++x_size;
        }
    }

    size_t hashLen = hashLength;       /* Offset of indices array;
					     shortens linearly by collisionByteLength. */
    size_t lenIndices = sizeof(uint32_t); /* Byte length of indices array;
					     doubles with every round. */
    for (int r = 1; r < k && x_size > 0; r++) {
        qsort_r(x, x_size, sizeof(hash), compareSR, (int *) &collisionByteLength);

        for (uint32_t i = 0; i < x_size - 1;) {
            // 2b) Find next set of unordered pairs with collisions on the next n/(k+1) bits
            int j = 1;
            while (i + j < x_size && hasCollision(X(i), X(i + j), collisionByteLength)) {
                j++;
            }

            if (validBlockData && equihashProxy(validBlockData, 0)) {
                goto cleanup;
            }

            /* Found partially collided values range between i and i+j. */

            // 2c) Calculate tuples (X_i ^ X_j, (i, j))
            for (int l = 0; l < j - 1; l++) {
                for (int m = l + 1; m < j; m++) {
                    if (distinctIndices(X(i + l), X(i + m), hashLen, lenIndices)) {
                        combineRows(Xc(xc_size), X(i + l), X(i + m), hashLen, lenIndices, collisionByteLength);
                        ++xc_size;
                        if (xc_size >= xc_room) {
                            xc_room *= 2;
                            xc = realloc(xc, xc_room * sizeof(hash));
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
        qsort_r(x, x_size, sizeof(hash), compareSR, (int *) &hashLen);
        for (uint32_t i = 0; i < x_size - 1;) {
            uint32_t j = 1;
            while (i + j < x_size && hasCollision(X(i), X(i + j), hashLen)) {
                j++;
            }

            for (uint32_t l = 0; l < j - 1; l++) {
                if (validBlockData && equihashProxy(validBlockData, 0)) {
                    goto cleanup;
                }

                for (uint32_t m = l + 1; m < j; m++) {
                    combineRows(Xc(xc_size), X(i + l), X(i + m), hashLen, lenIndices, 0);

                    if (isZero(Xc(xc_size), hashLen) &&
                        distinctIndices(X(i + l), X(i + m), hashLen, lenIndices)) {
                        uint8_t soln[equihashSolutionSize];

                        uint32_t ssize = getIndices(Xc(xc_size), hashLen, 2 * lenIndices, collisionBitLength,
                                               soln, sizeof(soln));
                        ++solnr;

                        assert(equihashSolutionSize == ssize);

#ifdef __INCLUDE_MAIN__
                        fprintf(stderr, "+ collision of size %d (%d)\n", equihashSolutionSize, ssize);
                        for (int y = 0; y < 2 * lenIndices; y += sizeof(uint32_t))
                          fprintf(stderr, " %u", arrayToEhIndex(Xc(xc_size) + hashLen + y));
                        fprintf(stderr, "\n");
#endif

                        if (validBlockData && equihashProxy(validBlockData, soln)) {
                            goto cleanup;
                        }
                    } else if (validBlockData && equihashProxy(validBlockData, 0)) {
                        goto cleanup;
                    }
                    ++xc_size;
                    assert(xc_size < xc_room);
                }
            }
            i += j;
        }
    }

cleanup:

    free(x);
    free(xc);
    return solnr;
}

static int basicValidate(int n, int k, blake2b_state *digest, void *soln) {
    const int collisionBitLength = n / (k + 1);
    const int collisionByteLength = (collisionBitLength + 7) / 8;
    const int HashLength = (k + 1) * collisionByteLength;
    const int indicesPerHashOutput = 512 / n;
    const int hashOutput = indicesPerHashOutput * n / 8;
    const int equihashSolutionSize = (1 << k) * (n / (k + 1) + 1) / 8;
    const int solnr = 1 << k;
    const int fullWidth = 2 * collisionByteLength + sizeof(uint32_t) * (1 << (k - 1));

    uint8_t hash[fullWidth];
    uint8_t *x = malloc(solnr * sizeof(hash));
    uint8_t *xc = malloc(solnr * sizeof(hash)); // merge list
    assert(x);
    assert(xc);

    uint32_t x_size = 0, xc_size = 0;
    uint32_t indices[solnr];

    expandArray(soln, equihashSolutionSize, (unsigned char *) &indices, sizeof(indices), collisionBitLength + 1, 1);

    uint8_t vHash[HashLength];
    memset(vHash, 0, sizeof(vHash));
    for (int j = 0; j < solnr; j++) {
        uint8_t tmpHash[hashOutput];
        uint8_t buf[HashLength + sizeof(uint32_t)];
        int i = be32toh(indices[j]);

        generateHash(digest, i / indicesPerHashOutput, tmpHash, hashOutput);
        expandArray(tmpHash + (i % indicesPerHashOutput * n / 8), n / 8, buf, HashLength, collisionBitLength, 0);

        ehIndexToArray(i, buf + HashLength);
        memcpy(X(x_size), buf, HashLength + sizeof(uint32_t));

        ++x_size;
    }

    size_t hashLength = HashLength;
    size_t lenIndices = sizeof(uint32_t);
    bool result = false;

    while (x_size > 1) {
        for (uint32_t i = 0; i < x_size; i += 2) {
            if (!hasCollision(X(i), X(i + 1), collisionByteLength)) {
                goto cleanup;
            }

            if (indicesBefore(X(i + 1), X(i), hashLength, lenIndices)) {
                goto cleanup;
            }

            if (!distinctIndices(X(i), X(i + 1), hashLength, lenIndices)) {
                goto cleanup;
            }
            combineRows(Xc(xc_size), X(i), X(i + 1), hashLength, lenIndices, collisionByteLength);
            ++xc_size;
        }
        hashLength -= collisionByteLength;
        lenIndices *= 2;

        /* swap arrays */
        swap(x, xc);
        x_size = xc_size;
        xc_size = 0;
    }

    result = isZero(X(0), hashLength);

cleanup:
    free(x);
    free(xc);

    return result;
}

static void hashNonce(blake2b_state *S, uint32_t nonce) {
    uint32_t expandedNonce[8] = {0};
    expandedNonce[0] = htole32(nonce);

    blake2b_update(S, (uint8_t *) &expandedNonce, sizeof(expandedNonce));
}

int EquihashValidate(int n, int k, void *input, int len, int64_t nonce, void *soln) {
    blake2b_state digest[1];
    digestInit(digest, n, k);
    blake2b_update(digest, (uint8_t *) input, len);

    if (nonce >= 0) {
        hashNonce(digest, (uint32_t) nonce);
    }

    return basicValidate(n, k, digest, soln);
}

int EquihashSolve(void *input, int len, int64_t nonce, void *validBlockData, int n, int k) {
    blake2b_state digest[1];
    digestInit(digest, n, k);
    blake2b_update(digest, (const uint8_t *) input, len);

    if (nonce >= 0) {
        hashNonce(digest, (uint32_t) nonce);
    }

    return basicSolve(digest, n, k, validBlockData);
}

void *GetIndices(int n, int k, void *soln) {
    const int equihashSolutionSize = (1 << k) * (n / (k + 1) + 1) / 8;
    const int collisionBitLength = n / (k + 1);
    const int lenIndices = 8 * sizeof(uint32_t) * equihashSolutionSize / (collisionBitLength + 1);
    const int bytePad = sizeof(uint32_t) - ((collisionBitLength + 1) + 7) / 8;

    unsigned char indices_array[lenIndices];
    uint32_t *indices = (uint32_t *) indices_array;

    expandArray(soln, equihashSolutionSize, indices_array, lenIndices, collisionBitLength + 1, bytePad);

    void *result = malloc(lenIndices);
    memcpy(result, indices, lenIndices);

    return result;
}

void *getMinimalFromIndices(uint32_t *solutionIndices, size_t len, size_t cBitLen) {
    const size_t lenIndices = len * sizeof(uint32_t);
    const size_t minLen = (cBitLen + 1) * lenIndices / (8 * sizeof(uint32_t));
    const size_t bytePad = sizeof(uint32_t) - ((cBitLen + 1) + 7) / 8;

    unsigned char *array = (unsigned char *) malloc(lenIndices);

    for (size_t i = 0; i < len; i++) {
        uint32_t value = solutionIndices[i];
        uint8_t *ptr = array + (i * sizeof(uint32_t));
        ehIndexToArray(value, ptr);
    }
    unsigned char *ret = (unsigned char *) malloc(minLen);

    compressArray(array, lenIndices, ret, minLen, cBitLen + 1, bytePad);
    free(array);

    return ret;
}

void *PutIndices(int n, int k, void *input, int inputLen, uint32_t nonce, void *indices, int numIndices) {
    const size_t cBitLen = n / (k + 1);
    blake2b_state digest[1];
    digestInit(digest, n, k);
    blake2b_update(digest, (const uint8_t *) input, inputLen);

    hashNonce(digest, nonce);

    return getMinimalFromIndices((uint32_t *) indices, numIndices, cBitLen);
}

#ifdef __INCLUDE_MAIN__

struct validData {
    int n;
    int k;
    blake2b_state *digest;

    int (*validator)(struct validData *v, void *);
};

int equihash_proxy(void *userData, void *soln) {
    struct validData *data = (struct validData *) userData;

    if (data && data->validator) {
        return data->validator(data, soln);
    } else {
        return 0;
    }
}

int basicValidator(struct validData *v, void *soln) {
    return basicValidate(v->n, v->k, v->digest, soln);
}

typedef struct {
    int n;
    int k;
    bool result;
    char *input;
    int nonce;
    uint32_t *indices;
} ValidatorTest;

uint32_t validIndices[] = {2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080, 45858,
                           116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460,
                           49807, 52426, 80391, 69567, 114474, 104973, 122568};
uint32_t invalidIndices1[] = {2262, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080,
                              45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132,
                              23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568};
uint32_t invalidIndices2[] = {45858, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080,
                              2261, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132,
                              23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568};
uint32_t invalidIndices3[] = {15185, 2261, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080,
                              45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132,
                              23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568};
uint32_t invalidIndices4[] = {36112, 104243, 2261, 15185, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080,
                              45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132,
                              23460, 49807, 52426, 80391, 69567, 114474, 104973, 122568};
uint32_t invalidIndices5[] = {2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080,
                              45858, 116805, 92842, 111026, 15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132,
                              23460, 49807, 52426, 80391, 104973, 122568, 69567, 114474};
uint32_t invalidIndices6[] = {15972, 115059, 85191, 90330, 68190, 122819, 81830, 91132, 23460, 49807, 52426, 80391,
                              69567, 114474, 104973, 122568, 2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041,
                              32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026};
uint32_t invalidIndices7[] = {2261, 15185, 15972, 23460, 23779, 32642, 36112, 45858, 49807, 52426, 68190, 69567, 69878,
                              76925, 80080, 80391, 81830, 85191, 90330, 91132, 92842, 104243, 104973, 111026, 114474,
                              115059, 116805, 118332, 118390, 122568, 122819, 130041};
uint32_t invalidIndices8[] = {2261, 2261, 15185, 15185, 36112, 36112, 104243, 104243, 23779, 23779, 118390, 118390,
                              118332, 118332, 130041, 130041, 32642, 32642, 69878, 69878, 76925, 76925, 80080, 80080,
                              45858, 45858, 116805, 116805, 92842, 92842, 111026, 111026};
uint32_t invalidIndices9[] = {2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041, 32642, 69878, 76925, 80080,
                              45858, 116805, 92842, 111026, 2261, 15185, 36112, 104243, 23779, 118390, 118332, 130041,
                              32642, 69878, 76925, 80080, 45858, 116805, 92842, 111026};

ValidatorTest validatorTests[] = {
// Original valid solution
        {96, 5, true,  "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                validIndices,
        },
// Change one index
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices1,
        },
// Swap two arbitrary indices
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices2,
        },
// Reverse the first pair of indices
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices3,
        },
// Swap the first and second pairs of indices
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices4,
        },
// Swap the second-to-last and last pairs of indices
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices5,
        },
// Swap the first half and second half
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices6,
        },
// Sort the indices
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices7,
        },
// Duplicate indices
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices8,
        },
// Duplicate first half
        {96, 5, false, "Equihash is an asymmetric PoW based on the Generalised Birthday problem.", 1,
                invalidIndices9,
        },
};

bool singleTest(ValidatorTest *test) {
    size_t inputLen = strlen(test->input);
    void *equihashSolution = PutIndices(test->n, test->k, test->input, inputLen, test->nonce, test->indices, 68);
    bool result = EquihashValidate(test->n, test->k, test->input, inputLen, test->nonce, equihashSolution);

    free(equihashSolution);

    return result == test->result;
}

int main(int argc, char **argv) {
    int nTests = sizeof(validatorTests) / sizeof(ValidatorTest);

    for (int i = 0; i < nTests; i++) {
        fprintf(stderr, "Test (%d) ...", i);
        fprintf(stderr, "%s\n", singleTest(&validatorTests[i]) ? "PASS" : "FAIL");
    }
}

#endif
