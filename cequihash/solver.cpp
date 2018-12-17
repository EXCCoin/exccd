#include "solver.h"
#include <vector>
#include "cequihash.h"

typedef verify_code (*verify_ptr)(uint32_t *indices, uint32_t proofsize, const unsigned char *input, const uint32_t input_len);
typedef void (*solve_ptr)(const unsigned char *input, uint32_t input_len, uint32_t nonce, uint8_t algo_version, const void *userData);
typedef void (*compress_ptr)(const uint32_t *sol, uint8_t *csol);

template verify_code verify<48, 5>(uint32_t *indices, uint32_t proofsize, const unsigned char *input, uint32_t input_len);
template verify_code verify<96, 5>(uint32_t *indices, uint32_t proofsize, const unsigned char *input, uint32_t input_len);
template verify_code verify<144, 5>(uint32_t *indices, uint32_t proofsize, const unsigned char *input, uint32_t input_len);
template verify_code verify<200, 9>(uint32_t *indices, uint32_t proofsize, const unsigned char *input, uint32_t input_len);
template void solve<48, 5>(const unsigned char *input, uint32_t input_len, uint32_t nonce, uint8_t algo_version, const void *userData);
template void solve<96, 5>(const unsigned char *input, uint32_t input_len, uint32_t nonce, uint8_t algo_version, const void *userData);
template void solve<144, 5>(const unsigned char *input, uint32_t input_len, uint32_t nonce, uint8_t algo_version, const void *userData);
template void solve<200, 9>(const unsigned char *input, uint32_t input_len, uint32_t nonce, uint8_t algo_version, const void *userData);
template void compress_solution<48, 5>(const uint32_t *sol, uint8_t *csol);
template void compress_solution<96, 5>(const uint32_t *sol, uint8_t *csol);
template void compress_solution<144, 5>(const uint32_t *sol, uint8_t *csol);
template void compress_solution<200, 9>(const uint32_t *sol, uint8_t *csol);

struct solver_record {
    uint32_t     N;
    uint32_t     solution_size;
    uint32_t     proof_size;
    verify_ptr   vfn;
    solve_ptr    sfn;
    compress_ptr cfn;

    solver_record(uint32_t n, uint32_t size, uint32_t psize, verify_ptr vptr, solve_ptr sptr, compress_ptr cptr)
        : N(n), solution_size(size), proof_size(psize), vfn(vptr), sfn(sptr), cfn(cptr) {}
};

std::vector<solver_record> solvers = {
    {48, equihash_solution_size(48, 5), 1 << 5, verify<48, 5>, solve<48, 5>, compress_solution<48, 5>},
    {96, equihash_solution_size(96, 5), 1 << 5, verify<96, 5>, solve<96, 5>, compress_solution<96, 5>},
    {144, equihash_solution_size(144, 5), 1 << 5, verify<144, 5>, solve<144, 5>, compress_solution<144, 5>},
    {200, equihash_solution_size(200, 9), 1 << 9, verify<200, 9>, solve<200, 9>, compress_solution<200, 9>},
};

static solver_record *find_solver(uint32_t n) {
    for (uint32_t i = 0; i < solvers.size(); i++) {
        if (solvers[i].N == (uint32_t)n) {
            return &solvers[i];
        }
    }
    return nullptr;
}

static uint32_t array_to_index(const unsigned char *array) {
    uint32_t bei;
    memcpy(&bei, array, sizeof(uint32_t));
    return be32toh(bei);
}

static void expand_array(const unsigned char *in, size_t in_len, unsigned char *out, size_t out_len, size_t bit_len, size_t byte_pad) {
    assert(bit_len >= 8);
    assert(8 * sizeof(uint32_t) >= 7 + bit_len);

    size_t out_width{(bit_len + 7) / 8 + byte_pad};
    assert(out_len == 8 * out_width * in_len / bit_len);

    uint32_t bit_len_mask{((uint32_t)1 << bit_len) - 1};

    // The acc_bits least-significant bits of acc_value represent a bit sequence
    // in big-endian order.
    size_t   acc_bits  = 0;
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
                                 acc_value >> (acc_bits + (8 * (out_width - x - 1)))) &
                             (
                                 // Apply bit_len_mask across byte boundaries
                                 (bit_len_mask >> (8 * (out_width - x - 1))) & 0xFF);
            }
            j += out_width;
        }
    }
}

std::vector<uint32_t> to_indices(const unsigned char *minimal, uint32_t sol_size, size_t cBitLen) {
    assert(((cBitLen + 1) + 7) / 8 <= sizeof(uint32_t));

    size_t lenIndices{8 * sizeof(uint32_t) * sol_size / (cBitLen + 1)};
    size_t bytePad{sizeof(uint32_t) - ((cBitLen + 1) + 7) / 8};

    std::vector<unsigned char> array(lenIndices);
    expand_array(minimal, sol_size, array.data(), lenIndices, cBitLen + 1, bytePad);
    std::vector<uint32_t> result;
    result.reserve(lenIndices);

    for (size_t i = 0; i < lenIndices; i += sizeof(uint32_t)) {
        result.push_back(array_to_index(array.data() + i));
    }

    return result;
}

int EquihashValidate(int n, int k, const void *input, int len, const void *soln) {
    if (n <= 0 || k <= 0) {
        return static_cast<std::underlying_type<verify_code>::type>(verify_code::POW_UNKNOWN_PARAMS);
    }

    auto solver = find_solver(n);

    if (!solver) {
        return static_cast<std::underlying_type<verify_code>::type>(verify_code::POW_UNKNOWN_PARAMS);
    }

    size_t collision_bit_length = n / (k + 1);

    auto indices = to_indices((uint8_t *)soln, solver->solution_size, collision_bit_length);
    auto result  = solver->vfn(indices.data(), indices.size(), (uint8_t *)input, len);

    return static_cast<std::underlying_type<verify_code>::type>(result);
}

void EquihashSolve(int n, int k, const void *input, int len, uint32_t nonce, uint8_t algo_version, const void *validBlockData) {
    auto solver = find_solver(n);
    if (solver) {
        solver->sfn((uint8_t *)input, len, nonce, algo_version, validBlockData);
    }
}

void *IndicesFromSolution(int n, int k, void *soln) {
    size_t collision_bit_length = n / (k + 1);
    size_t solution_width       = (1 << k) * (collision_bit_length + 1) / 8;

    auto indices = to_indices((uint8_t *)soln, solution_width, collision_bit_length);

    uint32_t lenIndices = indices.size() * sizeof(uint32_t);
    void *   result     = malloc(lenIndices);
    memcpy(result, indices.data(), lenIndices);

    return result;
}

void *SolutionFromIndices(int n, int k, const void *indices, uint32_t numIndices) {
    auto solver = find_solver(n);

    if (!solver) {
        return nullptr;
    }

    if (numIndices != solver->proof_size) {
        return nullptr;
    }

    void *result = malloc(solver->solution_size);

    if (!result) {
        return nullptr;
    }

    solver->cfn((const uint32_t *)indices, (uint8_t *)result);
    return result;
}
