package secp256k1

// // **This is CGO's build system. CGO parses the following comments as build instructions.**
// // Including the headers and code, and defining the default macros
// #cgo CFLAGS: -I./depend/secp256k1 -I./depend/secp256k1/src/
// #cgo CFLAGS: -DSECP256K1_BUILD=1 -DECMULT_WINDOW_SIZE=15 -DENABLE_MODULE_SCHNORRSIG=1 -DENABLE_MODULE_EXTRAKEYS=1
// #cgo CFLAGS: -DECMULT_GEN_PREC_BITS=4
// // x86_64 can use the Assembly implementation.
// #cgo amd64 CFLAGS: -DUSE_ASM_X86_64=1
// #include "./depend/secp256k1/include/secp256k1.h"
// #include "./depend/secp256k1/src/secp256k1.c"
import "C"

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/pkg/errors"
)

// A global context for using secp256k1. this is generated once and used to speed computation
// and aid resisting side channel attacks.
var context *C.secp256k1_context

// Initialize the context for both signing and verifying.
// and randomize it to help resist side channel attacks.
func init() {
	context = C.secp256k1_context_create(C.SECP256K1_CONTEXT_SIGN | C.SECP256K1_CONTEXT_VERIFY)
	seed := [32]byte{}
	n, err := rand.Read(seed[:])
	if err != nil || n != len(seed) {
		panic("Failed getting random values on initializing")
	}
	cPtr := (*C.uchar)(&seed[0])
	ret := C.secp256k1_context_randomize(context, cPtr)
	if ret != 1 {
		panic("Failed randomizing the context. Should never happen")
	}
}

// errNonInitializedKey is the error returned when using a zeroed pubkey
var errNonInitializedKey = errors.New("the key isn't initialized, you should use the generate/deserialize functions")

const (
	// HashSize of array used to store hashes. See Hash.
	HashSize = 32

	// SerializedPrivateKeySize defines the length in bytes of SerializedPrivateKey
	SerializedPrivateKeySize = 32
)

// Hash is a type encapsulating the result of hashing some unknown sized data.
// it typically represents Sha256 / Double Sha256.
type Hash [HashSize]byte

// IsEqual returns true if target is the same as hash.
func (hash *Hash) IsEqual(target *Hash) bool {
	if hash == nil && target == nil {
		return true
	}
	if hash == nil || target == nil {
		return false
	}
	return *hash == *target
}

// SetBytes sets the bytes which represent the hash. An error is returned if
// the number of bytes passed in is not HashSize.
func (hash *Hash) SetBytes(newHash []byte) error {
	if len(newHash) != HashSize {
		return errors.Errorf("invalid hash length got %d, expected %d", len(newHash),
			HashSize)
	}
	copy(hash[:], newHash)
	return nil
}

// String returns the Hash as the hexadecimal string
func (hash *Hash) String() string {
	return hex.EncodeToString(hash[:])
}
