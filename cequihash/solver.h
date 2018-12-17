#ifndef __SOLVER_H
#define __SOLVER_H

#include "miner.h"

inline constexpr size_t equihash_solution_size(unsigned int N, unsigned int K) {
    return (1 << K) * (N / (K + 1) + 1) / 8;
}

#endif  //__SOLVER_H
