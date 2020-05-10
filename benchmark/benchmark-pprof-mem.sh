#!/bin/bash
cd $(dirname "$0")
mkdir results 2> /dev/null
go test -c -o results/test

#export GODEBUG=allocfreetrace=1,gctrace=1
#BENCHMARK=BenchmarkParallelParseStarwarsQuery
BENCHMARK=${1:-.}

./results/test -test.bench=$BENCHMARK -test.memprofile results/latest.mem  | tee b results/latest-memprofile.txt

# go tool pprof -alloc_objects -lines -top results/latest.mem | head -n 25
go tool pprof -alloc_space -lines -top results/latest.mem | head -n 25