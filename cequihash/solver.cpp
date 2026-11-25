#include "solver.h"
#include <array>
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
    uint32_t     n;
    uint32_t     k;
    uint32_t     solution_size;
    uint32_t     proof_size;
    verify_ptr   vfn;
    solve_ptr    sfn;
    compress_ptr cfn;
};

template<uint32_t N, uint32_t K>
static constexpr solver_record solver_for() {
    return {N, K, equihash_solution_size(N, K), 1U << K,
            verify<N, K>, solve<N, K>, compress_solution<N, K>};
}

static constexpr solver_record solvers[] = {
    solver_for<48, 5>(), solver_for<96, 5>(), solver_for<144, 5>(), solver_for<200, 9>()
};

static const solver_record *find_solver(uint32_t n, uint32_t k) {
    for (const auto &solver : solvers) {
        if (solver.n == n && solver.k == k) return &solver;
    }
    return nullptr;
}

int EquihashPackedSize(int n, int k) {
    auto solver = find_solver(n, k);
    return solver ? solver->solution_size : 0;
}

// The proof packs fixed-width indices most significant bit first.
static bool to_indices(const solver_record &solver, const uint8_t *input, uint32_t *out) {
    const unsigned width = solver.n / (solver.k + 1) + 1;
    if (!input || !out || width < 8 || width > 25 ||
        solver.solution_size * 8 != solver.proof_size * width) return false;

    const uint32_t mask = (1U << width) - 1;
    uint32_t acc = 0;
    unsigned bits = 0;
    for (size_t i = 0; i < solver.solution_size; ++i) {
        acc = (acc << 8) | input[i];
        bits += 8;
        if (bits >= width) {
            bits -= width;
            *out++ = (acc >> bits) & mask;
        }
    }
    return true;
}

int EquihashValidate(int n, int k, const void *input, int len, const void *soln) {
    auto solver = find_solver(n, k);
    if (!solver || !input || len <= 0) return static_cast<int>(verify_code::POW_UNKNOWN_PARAMS);
    std::array<uint32_t, 1U << 9> indices;
    if (solver->proof_size > indices.size() ||
        !to_indices(*solver, static_cast<const uint8_t *>(soln), indices.data()))
        return static_cast<int>(verify_code::POW_SOL_SIZE_MISMATCH);

    return static_cast<int>(solver->vfn(indices.data(), solver->proof_size,
                                        static_cast<const uint8_t *>(input), len));
}

void EquihashSolve(int n, int k, const void *input, int len, uint32_t nonce, uint8_t algo_version, const void *validBlockData) {
    auto solver = find_solver(n, k);
    if (solver && input && len > 0 && (algo_version == 0 || len >= 144)) {
        solver->sfn(static_cast<const uint8_t *>(input), len, nonce, algo_version, validBlockData);
    }
}

void *IndicesFromSolution(int n, int k, void *soln) {
    auto solver = find_solver(n, k);
    if (!solver || !soln) return nullptr;
    auto result = static_cast<uint32_t *>(malloc(solver->proof_size * sizeof(uint32_t)));
    if (!result) return nullptr;
    if (!to_indices(*solver, static_cast<const uint8_t *>(soln), result)) {
        free(result);
        return nullptr;
    }
    return result;
}

void *SolutionFromIndices(int n, int k, const void *indices, uint32_t numIndices) {
    auto solver = find_solver(n, k);

    if (!solver) {
        return nullptr;
    }

    if (!indices || numIndices != solver->proof_size) {
        return nullptr;
    }

    void *result = malloc(solver->solution_size);

    if (!result) {
        return nullptr;
    }

    solver->cfn((const uint32_t *)indices, (uint8_t *)result);
    return result;
}
