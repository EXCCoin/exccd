// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2016 Jack Grigg
// Copyright (c) 2016 The Zcash developers
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

#ifndef __EQUIHASH_H
#define __EQUIHASH_H

#include "blake2.h"

#include <cstring>
#include <exception>
#include <functional>
#include <memory>
#include <set>
#include <vector>
#include <algorithm>
#include <cassert>

typedef blake2b_state eh_HashState;
typedef uint32_t eh_index;
typedef uint8_t eh_trunc;

void ExpandArray(const unsigned char* in, size_t in_len,
                 unsigned char* out, size_t out_len,
                 size_t bit_len, size_t byte_pad=0);
void CompressArray(const unsigned char* in, size_t in_len,
                   unsigned char* out, size_t out_len,
                   size_t bit_len, size_t byte_pad=0);

eh_index ArrayToEhIndex(const unsigned char* array);
eh_trunc TruncateIndex(const eh_index i, const unsigned int ilen);

std::vector<eh_index> GetIndicesFromMinimal(std::vector<unsigned char> minimal,
                                            size_t cBitLen);
std::vector<unsigned char> GetMinimalFromIndices(std::vector<eh_index> indices,
                                                 size_t cBitLen);

template<size_t WIDTH>
class StepRow
{
    template<size_t W>
    friend class StepRow;
    friend class CompareSR;

protected:
    unsigned char hash[WIDTH];

public:
    StepRow(const unsigned char* hashIn, size_t hInLen,
            size_t hLen, size_t cBitLen);
    ~StepRow() { }

    template<size_t W>
    StepRow(const StepRow<W>& a);

    bool IsZero(size_t len);
    std::string GetHex(size_t len) { return HexStr(hash, hash+len); }

    template<size_t W>
    friend bool HasCollision(StepRow<W>& a, StepRow<W>& b, int l);
};

class CompareSR
{
private:
    size_t len;

public:
    CompareSR(size_t l) : len {l} { }

    template<size_t W>
    inline bool operator()(const StepRow<W>& a, const StepRow<W>& b) { return memcmp(a.hash, b.hash, len) < 0; }
};

template<size_t WIDTH>
bool HasCollision(StepRow<WIDTH>& a, StepRow<WIDTH>& b, int l);

template<size_t WIDTH>
class FullStepRow : public StepRow<WIDTH>
{
    template<size_t W>
    friend class FullStepRow;

    using StepRow<WIDTH>::hash;

public:
    FullStepRow(const unsigned char* hashIn, size_t hInLen,
                size_t hLen, size_t cBitLen, eh_index i);
    ~FullStepRow() { }

    FullStepRow(const FullStepRow<WIDTH>& a) : StepRow<WIDTH> {a} { }
    template<size_t W>
    FullStepRow(const FullStepRow<W>& a, const FullStepRow<W>& b, size_t len, size_t lenIndices, int trim);
    FullStepRow& operator=(const FullStepRow<WIDTH>& a);

    inline bool IndicesBefore(const FullStepRow<WIDTH>& a, size_t len, size_t lenIndices) const { return memcmp(hash+len, a.hash+len, lenIndices) < 0; }
    std::vector<unsigned char> GetIndices(size_t len, size_t lenIndices,
                                          size_t cBitLen) const;

    template<size_t W>
    friend bool DistinctIndices(const FullStepRow<W>& a, const FullStepRow<W>& b,
                                size_t len, size_t lenIndices);
    template<size_t W>
    friend bool IsValidBranch(const FullStepRow<W>& a, const size_t len, const unsigned int ilen, const eh_trunc t);
};

template<size_t WIDTH>
class TruncatedStepRow : public StepRow<WIDTH>
{
    template<size_t W>
    friend class TruncatedStepRow;

    using StepRow<WIDTH>::hash;

public:
    TruncatedStepRow(const unsigned char* hashIn, size_t hInLen,
                     size_t hLen, size_t cBitLen,
                     eh_index i, unsigned int ilen);
    ~TruncatedStepRow() { }

    TruncatedStepRow(const TruncatedStepRow<WIDTH>& a) : StepRow<WIDTH> {a} { }
    template<size_t W>
    TruncatedStepRow(const TruncatedStepRow<W>& a, const TruncatedStepRow<W>& b, size_t len, size_t lenIndices, int trim);
    TruncatedStepRow& operator=(const TruncatedStepRow<WIDTH>& a);

    inline bool IndicesBefore(const TruncatedStepRow<WIDTH>& a, size_t len, size_t lenIndices) const { return memcmp(hash+len, a.hash+len, lenIndices) < 0; }
    std::shared_ptr<eh_trunc> GetTruncatedIndices(size_t len, size_t lenIndices) const;
};

enum EhSolverCancelCheck
{
    ListGeneration,
    ListSorting,
    ListColliding,
    RoundEnd,
    FinalSorting,
    FinalColliding,
    PartialGeneration,
    PartialSorting,
    PartialSubtreeEnd,
    PartialIndexEnd,
    PartialEnd
};

class EhSolverCancelledException : public std::exception
{
    virtual const char* what() const throw() {
        return "Equihash solver was cancelled";
    }
};

inline constexpr const size_t max(const size_t A, const size_t B) { return A > B ? A : B; }

inline constexpr size_t equihash_solution_size(unsigned int N, unsigned int K) {
    return (1 << K)*(N/(K+1)+1)/8;
}

enum ValidationResult {
    POW_OK,
    POW_INVALID_COLLISION_LENGTH,
    POW_NON_ZERO_XOR,
    POW_INCORRECT_ORDER,
    POW_DUPLICATE_INDICES,
    POW_INVALID_SOLUTION_LEN,
};

template<unsigned int N, unsigned int K>
class Equihash
{
private:
    static_assert(K < N, "K must be less than N");
    static_assert(N % 8 == 0, "N must be divisible by 8");
    static_assert((N/(K+1)) + 1 < 8*sizeof(eh_index), "Must be true");
public:
    enum : size_t { IndicesPerHashOutput=512/N };
    enum : size_t { HashOutput=IndicesPerHashOutput*N/8 };
    enum : size_t { CollisionBitLength=N/(K+1) };
    enum : size_t { CollisionByteLength=(CollisionBitLength+7)/8 };
    enum : size_t { HashLength=(K+1)*CollisionByteLength };
    enum : size_t { FullWidth=2*CollisionByteLength+sizeof(eh_index)*(1 << (K-1)) };
    enum : size_t { FinalFullWidth=2*CollisionByteLength+sizeof(eh_index)*(1 << (K)) };
    enum : size_t { TruncatedWidth=max(HashLength+sizeof(eh_trunc), 2*CollisionByteLength+sizeof(eh_trunc)*(1 << (K-1))) };
    enum : size_t { FinalTruncatedWidth=max(HashLength+sizeof(eh_trunc), 2*CollisionByteLength+sizeof(eh_trunc)*(1 << (K))) };
    enum : size_t { SolutionWidth=(1 << K)*(CollisionBitLength+1)/8 };

    Equihash() { }

    void InitialiseState(eh_HashState& base_state, const void* input, uint32_t input_len, int64_t nonce);

    bool BasicSolve(const eh_HashState& base_state,
                    const std::function<bool(std::vector<unsigned char>)> validBlock,
                    const std::function<bool(EhSolverCancelCheck)> cancelled);
    bool OptimisedSolve(const eh_HashState& base_state,
                        const std::function<bool(std::vector<unsigned char>)> validBlock,
                        const std::function<bool(EhSolverCancelCheck)> cancelled);

    ValidationResult ValidateSolution(const eh_HashState &base_state, std::vector<unsigned char> soln);
};

inline size_t EquihashSolutionLen(unsigned int N, unsigned int K) {
    return (1 << K)*(N/(K+1)+1)/8;
}

template<size_t WIDTH>
bool DistinctIndices(const FullStepRow<WIDTH>& a, const FullStepRow<WIDTH>& b, size_t len, size_t lenIndices)
{
    for(size_t i = 0; i < lenIndices; i += sizeof(eh_index)) {
        for(size_t j = 0; j < lenIndices; j += sizeof(eh_index)) {
            if (memcmp(a.hash+len+i, b.hash+len+j, sizeof(eh_index)) == 0) {
                return false;
            }
        }
    }
    return true;
}

template<size_t MAX_INDICES>
bool IsProbablyDuplicate(std::shared_ptr<eh_trunc> indices, size_t lenIndices)
{
    assert(lenIndices <= MAX_INDICES);
    bool checked_index[MAX_INDICES] = {false};
    size_t count_checked = 0;
    for (size_t z = 0; z < lenIndices; z++) {
        // Skip over indices we have already paired
        if (!checked_index[z]) {
            for (size_t y = z+1; y < lenIndices; y++) {
                if (!checked_index[y] && indices.get()[z] == indices.get()[y]) {
                    // Pair found
                    checked_index[y] = true;
                    count_checked += 2;
                    break;
                }
            }
        }
    }
    return count_checked == lenIndices;
}

template<size_t WIDTH>
bool IsValidBranch(const FullStepRow<WIDTH>& a, const size_t len, const unsigned int ilen, const eh_trunc t)
{
    return TruncateIndex(ArrayToEhIndex(a.hash+len), ilen) == t;
}

static Equihash<144,5> Eh144_5;
static Equihash<200,9> Eh200_9;
static Equihash<96,5> Eh96_5;
static Equihash<48,5> Eh48_5;

#define EhInitialiseState(n, k, base_state, input, input_len, nonce)  \
    if (n == 144 && k == 5) {                 \
        Eh144_5.InitialiseState(base_state, input, input_len, nonce);  \
    } else if (n == 200 && k == 9) {         \
        Eh200_9.InitialiseState(base_state, input, input_len, nonce); \
    } else if (n == 96 && k == 5) {          \
        Eh96_5.InitialiseState(base_state, input, input_len, nonce);  \
    } else if (n == 48 && k == 5) {          \
        Eh48_5.InitialiseState(base_state, input, input_len, nonce);  \
    } else {                                 \
        throw std::invalid_argument("Unsupported Equihash parameters"); \
    }

inline bool EhBasicSolve(unsigned int n, unsigned int k, const eh_HashState& base_state,
                    const std::function<bool(std::vector<unsigned char>)> validBlock,
                    const std::function<bool(EhSolverCancelCheck)> cancelled)
{
    if (n == 144 && k == 5) {
        return Eh144_5.BasicSolve(base_state, validBlock, cancelled);
    } else if (n == 200 && k == 9) {
        return Eh200_9.BasicSolve(base_state, validBlock, cancelled);
    } else if (n == 96 && k == 5) {
        return Eh96_5.BasicSolve(base_state, validBlock, cancelled);
    } else if (n == 48 && k == 5) {
        return Eh48_5.BasicSolve(base_state, validBlock, cancelled);
    } else {
        throw std::invalid_argument("Unsupported Equihash parameters");
    }
}

inline bool EhOptimisedSolve(unsigned int n, unsigned int k, const eh_HashState& base_state,
                    const std::function<bool(std::vector<unsigned char>)> validBlock,
                    const std::function<bool(EhSolverCancelCheck)> cancelled)
{
    if (n == 144 && k == 5) {
        return Eh144_5.OptimisedSolve(base_state, validBlock, cancelled);
    } else if (n == 200 && k == 9) {
        return Eh200_9.OptimisedSolve(base_state, validBlock, cancelled);
    } else if (n == 96 && k == 5) {
        return Eh96_5.OptimisedSolve(base_state, validBlock, cancelled);
    } else if (n == 48 && k == 5) {
        return Eh48_5.OptimisedSolve(base_state, validBlock, cancelled);
    } else {
        throw std::invalid_argument("Unsupported Equihash parameters");
    }
}

#define EhValidateSolution(n, k, base_state, soln, ret)   \
    if (n == 144 && k == 5) {                             \
        ret = Eh144_5.ValidateSolution(base_state, soln);  \
    } else if (n == 200 && k == 9) {                     \
        ret = Eh200_9.ValidateSolution(base_state, soln); \
    } else if (n == 96 && k == 5) {                      \
        ret = Eh96_5.ValidateSolution(base_state, soln);  \
    } else if (n == 48 && k == 5) {                      \
        ret = Eh48_5.ValidateSolution(base_state, soln);  \
    } else {                                             \
        throw std::invalid_argument("Unsupported Equihash parameters"); \
    }

#endif // __EQUIHASH_H
