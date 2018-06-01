// Copyright (c) 2018 The ExchangeCoin team
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

#pragma once

void* GetIndices(int, int, void* soln);
void* PutIndices(int, int, const unsigned char*, int, int, void* , int solutionLen);
int EquihashSolve(void*, int, int, void *, int, int);
int EquihashValidate(int, int, void*, int, void*);

