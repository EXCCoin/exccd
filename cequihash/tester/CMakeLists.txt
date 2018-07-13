cmake_minimum_required(VERSION 3.10)
project(cequihash_tromp)

#set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_COMPILER g++)
set(CMAKE_CXX_FLAGS "-std=c++17 -march=native -m64 -Wall -Wno-deprecated-declarations -D_POSIX_C_SOURCE=200112L -O3 -Og ")

set(MAIN_SOURCES
        blake2b.cpp
        portable_endian.h
        miner.h
        tester.cpp)

add_executable(equi ${MAIN_SOURCES} solver.cpp solver.h)
set_target_properties(equi PROPERTIES COMPILE_FLAGS "-D__TEST_MAIN__")
