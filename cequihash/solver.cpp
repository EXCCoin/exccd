// Copyright (c) 2018 The ExchangeCoin team

#include "equihash.h"
#include "cequihash.h"
#include "solver.h"

int EquihashValidate(int n, int k, const void *input, int len, int64_t nonce, const void *soln) {
    blake2b_state state;
    EhInitialiseState(n, k, state, input, len, nonce);

    ValidationResult ret;
    EhValidateSolution(n, k, state, toVector((unsigned char*)soln, EquihashSolutionLen(n, k)), ret);

    return ret;
}

int EquihashSolve(int n, int k, const void *input, int len, int64_t nonce, const void *validBlockData) {
    blake2b_state state;
    EhInitialiseState(n, k, state, input, len, nonce);

    std::function<bool(std::vector<unsigned char>)> validBlock = [&validBlockData](std::vector<unsigned char> solution) {
        return equihashProxy(const_cast<void *>(validBlockData), solution.data()) != 0;
    };

    std::function<bool(EhSolverCancelCheck)> cancelled = [&validBlockData](EhSolverCancelCheck check) {
        return equihashProxy(const_cast<void *>(validBlockData), 0);
    };

    try {
        EhOptimisedSolve(n, k, state, validBlock, cancelled);
        //EhBasicSolve(n, k, state, validBlock, cancelled);

    } catch (EhSolverCancelledException e) {
        // Do nothing, we were expecting it
    }

    return 0;
}

void *SolutionFromIndices(int n, int k, const void *indices, uint32_t lenIndices) {
    auto indicesVector { toVector((eh_index*)indices, lenIndices) };
    auto solution { GetMinimalFromIndices(indicesVector, n / (k + 1)) };

    return make_copy(solution);
}

void *IndicesFromSolution(int n, int k, void *soln) {
    auto minimal { toVector(static_cast<unsigned char*>(soln), EquihashSolutionLen(n, k)) };
    auto indices { GetIndicesFromMinimal(minimal, n / (k + 1)) };

    return make_copy(indices);
}
