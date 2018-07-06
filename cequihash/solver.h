#ifndef __SOLVER_H
#define __SOLVER_H

#include "miner.h"
int verify_485(u32* indices, u32 proofsize, const unsigned char *input, const u32 input_len, int64_t nonce);
int verify_965(u32* indices, u32 proofsize, const unsigned char *input, const u32 input_len, int64_t nonce);
int verify_1445(u32* indices, u32 proofsize, const unsigned char *input, const u32 input_len, int64_t nonce);
int verify_2009(u32* indices, u32 proofsize, const unsigned char *input, const u32 input_len, int64_t nonce);

int solve_485(const unsigned char* input, u32 input_len, uint32_t nonce, const void* userData);
int solve_965(const unsigned char* input, u32 input_len, uint32_t nonce, const void* userData);
int solve_1445(const unsigned char* input, u32 input_len, uint32_t nonce, const void* userData);
int solve_2009(const unsigned char* input, u32 input_len, uint32_t nonce, const void* userData);

void compress_solution_485(const u32* sol, uchar *csol);
void compress_solution_965(const u32* sol, uchar *csol);
void compress_solution_1445(const u32* sol, uchar *csol);
void compress_solution_2009(const u32* sol, uchar *csol);

#endif //__SOLVER_H
