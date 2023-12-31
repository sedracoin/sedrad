package secp256k1

// #include "./depend/secp256k1/include/secp256k1.h"
import "C"
import (
	"encoding/hex"
	"github.com/pkg/errors"
)

const (
	// SerializedECDSASignatureSize defines the length in bytes of SerializedECDSASignature
	SerializedECDSASignatureSize = 64
)

// ECDSASignature is a type representing a ECDSA Signature.
// The struct itself is an opaque data type that should only be created via the supplied methods.
type ECDSASignature struct {
	signature C.secp256k1_ecdsa_signature
}

// SerializedECDSASignature is a is a byte array representing the storage representation of a ECDSASignature
type SerializedECDSASignature [SerializedECDSASignatureSize]byte

// IsEqual returns true if target is the same as signature.
func (signature *ECDSASignature) IsEqual(target *ECDSASignature) bool {
	if signature == nil && target == nil {
		return true
	}
	if signature == nil || target == nil {
		return false
	}
	return *signature.Serialize() == *target.Serialize()
}

// String returns the SerializedECDSASignature as the hexadecimal string
func (serialized SerializedECDSASignature) String() string {
	return hex.EncodeToString(serialized[:])
}

// String returns the ECDSASignature as the hexadecimal string
func (signature ECDSASignature) String() string {
	return signature.Serialize().String()
}

// Serialize returns a 64 byte serialized signature
func (signature *ECDSASignature) Serialize() *SerializedECDSASignature {
	serialized := SerializedECDSASignature{}
	cPtr := (*C.uchar)(&serialized[0])
	ret := C.secp256k1_ecdsa_signature_serialize_compact(C.secp256k1_context_no_precomp, cPtr, &signature.signature)
	if ret != 1 {
		panic("failed serializing a signature. Should never happen (upstream promise to return 1)")
	}
	return &serialized
}

// DeserializeECDSASignature deserializes a 64 byte serialized ECDSA signature into a ECDSASignature type.
func DeserializeECDSASignature(serializedSignature *SerializedECDSASignature) (*ECDSASignature, error) {
	signature := ECDSASignature{}
	cPtr := (*C.uchar)(&serializedSignature[0])
	ret := C.secp256k1_ecdsa_signature_parse_compact(C.secp256k1_context_no_precomp, &signature.signature, cPtr)
	if ret != 1 {
		return nil, errors.New("failed parsing the ECDSA signature")
	}
	return &signature, nil
}

// DeserializeECDSASignatureFromSlice returns a ECDSASignature type from a serialized signature slice.
// will verify that it's SerializedECDSASignatureSize bytes long
func DeserializeECDSASignatureFromSlice(data []byte) (signature *ECDSASignature, err error) {
	if len(data) != SerializedECDSASignatureSize {
		return nil, errors.Errorf("invalid ECDSA signature length got %d, expected %d", len(data),
			SerializedECDSASignatureSize)
	}

	serialized := SerializedECDSASignature{}
	copy(serialized[:], data)
	return DeserializeECDSASignature(&serialized)
}
