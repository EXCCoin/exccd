#include <vector>
#include "solver.h"
#include "cequihash.h"

typedef int (*verify_ptr)(u32* indices, u32 proofsize, const unsigned char *input, const u32 input_len, int64_t nonce);
typedef int (*solve_ptr)(unsigned char* input, u32 input_len, uint32_t nonce, void* userData);
typedef void (*compress_ptr)(const u32* sol, uchar *csol);

inline constexpr size_t equihash_solution_size(unsigned int N, unsigned int K) {
    return (1 << K)*(N/(K+1)+1)/8;
}


struct solver_record {
    uint32_t N;
    uint32_t solution_size;
    uint32_t proof_size;
    verify_ptr vfn;
    solve_ptr sfn;
    compress_ptr cfn;

    solver_record(uint32_t n, uint32_t size, uint32_t psize, verify_ptr vptr, solve_ptr sptr, compress_ptr cptr)
            :N(n), solution_size(size), proof_size(psize), vfn(vptr), sfn(sptr), cfn(cptr) {}
};

solver_record solvers[] = {
        { 48, equihash_solution_size(48, 5) , 1 << 5, verify_485 , solve_485 , compress_solution_485},
        { 96, equihash_solution_size(96, 5) , 1 << 5, verify_965 , solve_965 , compress_solution_965},
        {144, equihash_solution_size(144, 5), 1 << 5, verify_1445, solve_1445, compress_solution_1445},
        {200, equihash_solution_size(200, 9), 1 << 9, verify_2009, solve_2009, compress_solution_2009},
};

#define ARRAY_LEN(arr)  (sizeof(arr)/sizeof(arr[0]))

static solver_record* find_solver(u32 n) {
    for(uint32_t i = 0; i < ARRAY_LEN(solvers); i++) {
        if(solvers[i].N == (u32)n) {
            return &solvers[i];
        }
    }
    return nullptr;
}

static uint32_t array_to_index(const unsigned char *array)
{
    uint32_t bei;
    memcpy(&bei, array, sizeof(uint32_t));
    return be32toh(bei);
}

static void expand_array(const unsigned char *in, size_t in_len,
                         unsigned char *out, size_t out_len,
                         size_t bit_len, size_t byte_pad)
{
    assert(bit_len >= 8);
    assert(8*sizeof(uint32_t) >= 7+bit_len);

    size_t out_width { (bit_len+7)/8 + byte_pad };
    assert(out_len == 8*out_width*in_len/bit_len);

    uint32_t bit_len_mask { ((uint32_t)1 << bit_len) - 1 };

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
                out[j+x] = 0;
            }
            for (size_t x = byte_pad; x < out_width; x++) {
                out[j+x] = (
                                   // Big-endian
                                   acc_value >> (acc_bits+(8*(out_width-x-1)))
                           ) & (
                                   // Apply bit_len_mask across byte boundaries
                                   (bit_len_mask >> (8*(out_width-x-1))) & 0xFF
                           );
            }
            j += out_width;
        }
    }
}

static std::vector<uint32_t> to_indices(const unsigned char *minimal, uint32_t sol_size, size_t cBitLen)
{
    assert(((cBitLen+1)+7)/8 <= sizeof(uint32_t));

    size_t lenIndices { 8*sizeof(uint32_t) * sol_size/(cBitLen+1) };
    size_t bytePad { sizeof(uint32_t) - ((cBitLen+1)+7)/8 };

    std::vector<unsigned char> array(lenIndices);
    expand_array(minimal, sol_size, array.data(), lenIndices, cBitLen + 1, bytePad);
    std::vector<uint32_t> result;
    for (size_t i = 0; i < lenIndices; i += sizeof(uint32_t)) {
        result.push_back(array_to_index(array.data() + i));
    }
    return result;
}

int EquihashValidate(int n, int k, const void *input, int len, int64_t nonce, const void *soln) {
    if (n <= 0 || k <= 0) {
        return POW_UNKNOWN_PARAMS;
    }

    auto solver = find_solver(n);

    if (!solver) {
        return POW_UNKNOWN_PARAMS;
    }

    size_t collision_bit_length = n/(k+1);

    auto indices = to_indices((u8 *) soln, solver->solution_size, collision_bit_length);

    return solver->vfn(indices.data(), indices.size(), (u8*)input, len, nonce) ? 0:1;
}

int EquihashSolve(void *input, int len, int64_t nonce, void *validBlockData, int n, int k) {
    if (n <= 0 || k <= 0) {
        return 0;
    }

    auto solver = find_solver(n);
    return (solver) ? solver->sfn((u8*)input, len, nonce, validBlockData) : 0;
}

void *IndicesFromSolution(int n, int k, void *soln) {
    size_t collision_bit_length = n/(k+1);
    size_t solution_width = (1 << k)*(collision_bit_length + 1)/8;
    auto indices = to_indices((u8 *) soln, solution_width, collision_bit_length);
    u32 lenIndices = indices.size() * sizeof(uint32_t);
    void *result = malloc(lenIndices);
    memcpy(result, indices.data(), lenIndices);

    return result;
}

void *SolutionFromIndices(int n, int k, const void *indices, u32 numIndices) {
    auto solver = find_solver(n);

    if (!solver) {
        return nullptr;
    }

    if (numIndices != solver->proof_size) {
        return nullptr;
    }

    void* result = malloc(solver->solution_size);

    if (!result) {
        return nullptr;
    }

    solver->cfn((const u32*)indices, (u8*)result);
    return result;
}
