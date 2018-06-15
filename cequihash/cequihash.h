// Copyright (c) 2018 The ExchangeCoin team
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

#pragma once

#include <stdint.h>
#include <stdlib.h>

void* GetIndices(int, int, void* soln);
void *PutIndices(int, int, void *, int, uint32_t, void *, int);
int EquihashSolve(void*, int, int64_t , void *, int, int);
int EquihashValidate(int, int, void *, int, int64_t, void *soln);

