cmake_minimum_required(VERSION 3.10)
project(cequihash_tromp)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_COMPILER g++)
set(CMAKE_CXX_FLAGS "-O3 -march=x86-64 -mtune=generic -std=c++17 -Wall -Wno-strict-aliasing -Wno-shift-count-overflow -Werror -Og")

set(MAIN_SOURCES
        ../blake2b.cpp
        ../portable_endian.h
        ../miner.h
        ../tester.cpp)

add_executable(equi ${MAIN_SOURCES} ../solver.cpp ../solver.h)
set_target_properties(equi PROPERTIES COMPILE_FLAGS "-D__TEST_MAIN__")
