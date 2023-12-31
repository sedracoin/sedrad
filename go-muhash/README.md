
go-muhash
====
Warning: This is pre-alpha software. The code has not been audited and/or reviewed by anyone other than the author.
====
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](https://choosealicense.com/licenses/isc/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org//go-muhash)

go-muhash implements a rolling hash function using a multiplicative hash. <br>
it is based on a multiplicative group over Z=2^3072-1103717 using multiplication and division for adding/removing 
elements from the hash function.<br>
the current code is heavily based on: https://github.com/bitcoin/bitcoin/blob/a1fcceac69097a8e6540a6fd8121a5d53022528f/src/crypto/muhash.cpp 
(written by Pieter Wuille, MIT licensed) <br>
But uses BLAKE2B as the hash function and Go's standard library big.Int for fast modular inversions(GCD) <br>

`MuHash` is the public interface implementing Add/Remove elements functions, and a Finalize function to return a 
final hash.

`uint3072.go` is a go implementation of the multiplicative group <br>
`num3072.c/h` is a C implementation of the multiplicative group <br>
`num3072.go` is go bindings for the C imlementation

Ideally we will add Go Assembly implementations using SSE2/SSE4.1/AVX and will choose the correct one in runtime, this 
should also remove the cgo overhead.


## Tests
`./build_and_test.sh` will run all the tests and checks in this library. <br>
`./fuzz.sh` will run the fuzzer and put new corpus in the `corpus` directory. by default, it will use [go-fuzz](https://github.com/dvyukov/go-fuzz)
But if you run with `LIBFUZZER=1 ./fuzz.sh` it will run it with [libfuzzer](https://llvm.org/docs/LibFuzzer.html) <br>
All the current corpus are checked in the unit test in `fuzz_corpuses_test.go` (requires `-tags=gofuzz`)