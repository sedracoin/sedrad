#!/bin/sh -ex
go version
# This is to bypass a go bug: https://github.com/golang/go/issues/27643
GO111MODULE=off go get github.com/dvyukov/go-fuzz/go-fuzz \
                          github.com/dvyukov/go-fuzz/go-fuzz-build

if [ -z ${LIBFUZZER} ]; then
  go-fuzz-build
  go-fuzz -testoutput
else
  go-fuzz-build -libfuzzer -o fuzz_muhash.a
  clang -fsanitize=fuzzer fuzz_muhash.a -o fuzz_muhash
 ./fuzz_muhash -use_counters=1 -use_value_profile=1 corpus/
fi