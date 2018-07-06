#ifndef __CEQUIHASH_H
#define __CEQUIHASH_H

#include <stdint.h>
#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif
    int EquihashValidate(int n, int k, const void *input, int len, int64_t nonce, const void *soln);
    int EquihashSolve(const void *input, int len, int64_t nonce, const void *validBlockData, int n, int k);
    void *SolutionFromIndices(int n, int k, const void *indices, uint32_t numIndices);
    void *IndicesFromSolution(int n, int k, void *soln);
#ifdef __cplusplus
}
#endif
#endif //__CEQUIHASH_H