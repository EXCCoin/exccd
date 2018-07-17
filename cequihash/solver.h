#ifndef __SOLVER_H
#define __SOLVER_H

#include "miner.h"

// int verify_485(uint32_t* indices, uint32_t proofsize, const unsigned char *input, const uint32_t input_len, int64_t nonce);
// int verify_965(uint32_t* indices, uint32_t proofsize, const unsigned char *input, const uint32_t input_len, int64_t nonce);
// int verify_1445(uint32_t* indices, uint32_t proofsize, const unsigned char *input, const uint32_t input_len, int64_t nonce);
// int verify_2009(uint32_t* indices, uint32_t proofsize, const unsigned char *input, const uint32_t input_len, int64_t nonce);
//
// int solve_485(const unsigned char* input, uint32_t input_len, int64_t nonce, const void* userData);
// int solve_965(const unsigned char* input, uint32_t input_len, int64_t nonce, const void* userData);
// int solve_1445(const unsigned char* input, uint32_t input_len, int64_t nonce, const void* userData);
// int solve_2009(const unsigned char* input, uint32_t input_len, int64_t nonce, const void* userData);
//
// void compress_solution_485(const uint32_t* sol, uint8_t *csol);
// void compress_solution_965(const uint32_t* sol, uint8_t *csol);
// void compress_solution_1445(const uint32_t* sol, uint8_t *csol);
// void compress_solution_2009(const uint32_t* sol, uint8_t *csol);

inline constexpr size_t equihash_solution_size(unsigned int N, unsigned int K) {
    return (1 << K) * (N / (K + 1) + 1) / 8;
}
#endif  //__SOLVER_H
