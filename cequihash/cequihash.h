#pragma once
#include <stdint.h>
#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

int EquihashValidate(int n, int k, const void *input, int len, const void *soln);
void EquihashSolve(int n, int k, const void *input, int len, uint32_t nonce, uint8_t algo_version, const void *validBlockData);
void *SolutionFromIndices(int n, int k, const void *indices, uint32_t numIndices);
void *IndicesFromSolution(int n, int k, void *soln);

#ifdef __cplusplus
}
#endif