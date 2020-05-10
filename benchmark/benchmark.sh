#!/bin/bash
cd $(dirname "$0")

mkdir results 2> /dev/null
mv results/latest.txt results/previous.txt 2> /dev/null

go test -c -o results/test

#export GODEBUG=allocfreetrace=1,gctrace=1

BENCHMARK=${1:-.}
PREV_BENCHMARK=${2:-results/previous.txt}

./results/test  -test.bench=$BENCHMARK | tee results/latest.txt
./results/test  -test.bench=$BENCHMARK | tee -a results/latest.txt
./results/test  -test.bench=$BENCHMARK | tee -a results/latest.txt
benchcmp "$PREV_BENCHMARK" results/latest.txt
benchstat "$PREV_BENCHMARK" results/latest.txt

cp "results/latest.txt" "results/$(date).txt"

