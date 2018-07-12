// Copyright (c) 2018 The ExchangeCoin team

#ifndef _SOLVER_H
#define _SOLVER_H

#include <vector>

template <typename T>
std::vector<T> toVector(const T* src, size_t len) {
    std::vector<T> result;
    T* ptr = const_cast<T*> (src);

    for(size_t i = 0; i < len; i++, ptr++) {
        result.push_back(*ptr);
    }
    return result;
}

template <typename T>
void* make_copy(std::vector<T> input) {
    unsigned long size = input.size() * sizeof(T);
    void* result = malloc(size);
    memcpy(result, input.data(), size);
    return result;
}

#endif //_SOLVER_H
