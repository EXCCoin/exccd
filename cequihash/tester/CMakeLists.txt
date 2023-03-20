cmake_minimum_required(VERSION 3.10)
project(cequihash_tromp)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_FLAGS "-O3 -std=c++17 -Wall -Wno-strict-aliasing -Wno-shift-count-overflow -Werror")

EXECUTE_PROCESS( COMMAND uname -m COMMAND tr -d '\n' OUTPUT_VARIABLE MACHINE )
EXECUTE_PROCESS( COMMAND uname -p COMMAND tr -d '\n' OUTPUT_VARIABLE PROCESSOR )

set(MAIN_SOURCES
        ../portable_endian.h
        ../miner.h
        ../tester.cpp)

if(${MACHINE} STREQUAL "x86_64" OR ${MACHINE} STREQUAL "amd64" OR ${PROCESSOR} STREQUAL "amd64")
        set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -march=x86-64 -mtune=generic")
        set(MAIN_SOURCES ${MAIN_SOURCES} ../sse-blake2b.cpp)
elseif(${PROCESSOR} MATCHES "^POWER(8|9).*")
        set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -mtune=power8 -DNO_WARN_X86_INTRINSICS")
        set(MAIN_SOURCES ${MAIN_SOURCES} ../sse-blake2b.cpp)
else()
        set(MAIN_SOURCES ${MAIN_SOURCES} ../blake2b-ref.c)
endif()

add_executable(equi ${MAIN_SOURCES} ../solver.cpp ../solver.h)
set_target_properties(equi PROPERTIES COMPILE_FLAGS "-D__TEST_MAIN__")
