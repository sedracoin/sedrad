package secp256k1

import (
	"encoding/hex"
	"github.com/pkg/errors"
)

const (
	// SerializedSchnorrSignatureSize defines the length in bytes of SerializedSchnorrSignature
	SerializedSchnorrSignatureSize = 64
)

// SchnorrSignature is a type representing a Schnorr Signature.
// The struct itself is an opaque data type that should only be created via the supplied methods.
type SchnorrSignature struct {
	signature [SerializedSchnorrSignatureSize]byte
}

// SerializedSchnorrSignature is a is a byte array representing the storage representation of a SchnorrSignature
type SerializedSchnorrSignature [SerializedSchnorrSignatureSize]byte

// IsEqual returns true if target is the same as signature.
func (signature *SchnorrSignature) IsEqual(target *SchnorrSignature) bool {
	if signature == nil && target == nil {
		return true
	}
	if signature == nil || target == nil {
		return false
	}
	return *signature.Serialize() == *target.Serialize()
}

// String returns the SerializedSchnorrSignature as the hexadecimal string
func (serialized SerializedSchnorrSignature) String() string {
	return hex.EncodeToString(serialized[:])
}

// String returns the SchnorrSignature as the hexadecimal string
func (signature SchnorrSignature) String() string {
	return signature.Serialize().String()
}

// Serialize returns a 64 byte serialized signature
func (signature *SchnorrSignature) Serialize() *SerializedSchnorrSignature {
	ret := SerializedSchnorrSignature(signature.signature)
	return &ret
}

// DeserializeSchnorrSignature deserializes a 64 byte serialized schnorr signature into a SchnorrSignature type.
func DeserializeSchnorrSignature(serializedSignature *SerializedSchnorrSignature) *SchnorrSignature {
	return &SchnorrSignature{signature: *serializedSignature}
}

// DeserializeSchnorrSignatureFromSlice returns a SchnorrSignature type from a serialized signature slice.
// will verify that it's SerializedSchnorrSignatureSize bytes long
func DeserializeSchnorrSignatureFromSlice(data []byte) (signature *SchnorrSignature, err error) {
	if len(data) != SerializedSchnorrSignatureSize {
		return nil, errors.Errorf("invalid schnorr signature length got %d, expected %d", len(data),
			SerializedSchnorrSignatureSize)
	}
	signature = &SchnorrSignature{}
	copy(signature.signature[:], data)
	return
}
