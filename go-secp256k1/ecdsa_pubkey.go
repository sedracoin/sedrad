package secp256k1

// #include "./depend/secp256k1/include/secp256k1.h"
// #include "./depend/secp256k1/include/secp256k1_extrakeys.h"
import "C"
import (
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
)

// SerializedECDSAPublicKeySize defines the length in bytes of a SerializedECDSAPublicKey
const SerializedECDSAPublicKeySize = 33

// ECDSAPublicKey is a PublicKey type used to sign and verify ECDSA signatures.
// The struct itself is an opaque data type that should only be created via the supplied methods.
type ECDSAPublicKey struct {
	pubkey C.secp256k1_pubkey
	init   bool
}

// SerializedECDSAPublicKey is a is a byte array representing the storage representation of a compressed ECDSAPublicKey
type SerializedECDSAPublicKey [SerializedECDSAPublicKeySize]byte

// IsEqual returns true if target is the same as key.
func (key *ECDSAPublicKey) IsEqual(target *ECDSAPublicKey) bool {
	if key == nil && target == nil {
		return true
	}
	if key == nil || target == nil {
		return false
	}
	serializedKey, err1 := key.Serialize()
	if err1 != nil && !errors.Is(err1, errNonInitializedKey) {
		panic(errors.Wrap(err1, "Unexpected error when serrializing key"))
	}
	serializedTarget, err2 := target.Serialize()
	if err2 != nil && !errors.Is(err2, errNonInitializedKey) {
		panic(errors.Wrap(err1, "Unexpected error when serrializing key"))
	}

	if errors.Is(err1, errNonInitializedKey) && errors.Is(err2, errNonInitializedKey) { // They're both zeroed, shouldn't happen if a constructor is used.
		return true
	}
	if errors.Is(err1, errNonInitializedKey) || errors.Is(err2, errNonInitializedKey) { // Only one of them is zeroed
		return false
	}
	return *serializedKey == *serializedTarget
}

// String returns the SerializedECDSAPublicKey as a hexadecimal string
func (serialized SerializedECDSAPublicKey) String() string {
	return hex.EncodeToString(serialized[:])
}

// String returns the ECDSAPublicKey as the hexadecimal string
func (key ECDSAPublicKey) String() string {
	serialized, err := key.Serialize()
	if err != nil { // This can only happen if the user calls this function skipping a constructor. i.e. `ECDSAPublicKey{}.String()`
		return "<Invalid ECDSAPublicKey>"
	}
	return serialized.String()
}

// ECDSAVerify verifies a ECDSA signature using the public key and the input hashed message.
// Notice: the [32] byte array *MUST* be a hash of a message you hashed yourself.
func (key *ECDSAPublicKey) ECDSAVerify(hash *Hash, signature *ECDSASignature) bool {
	cPtrHash := (*C.uchar)(&hash[0])
	return C.secp256k1_ecdsa_verify(context, &signature.signature, cPtrHash, &key.pubkey) == 1
}

// DeserializeECDSAPubKey deserializes a serialized ECDSA public key, verifying it's valid.
// it supports only compressed(33 bytes) public keys.
// it does not support uncompressed(65 bytes) or hybrid(65 bytes) keys.
func DeserializeECDSAPubKey(serializedPubKey []byte) (*ECDSAPublicKey, error) {
	if len(serializedPubKey) != SerializedECDSAPublicKeySize {
		return nil, errors.New(fmt.Sprintf("serializedPubKey has to be %d bytes, instead got :%d", SerializedECDSAPublicKeySize, len(serializedPubKey)))
	}
	key := ECDSAPublicKey{init: true}
	cPtr := (*C.uchar)(&serializedPubKey[0])
	ret := C.secp256k1_ec_pubkey_parse(context, &key.pubkey, cPtr, SerializedECDSAPublicKeySize)
	if ret != 1 {
		return nil, errors.New("failed parsing the public key")
	}
	return &key, nil
}

// Serialize serializes a ECDSA public key
func (key *ECDSAPublicKey) Serialize() (*SerializedECDSAPublicKey, error) {
	if !key.init {
		return nil, errors.WithStack(errNonInitializedKey)
	}
	serialized := SerializedECDSAPublicKey{}
	cPtr := (*C.uchar)(&serialized[0])
	cLen := C.size_t(SerializedECDSAPublicKeySize)

	ret := C.secp256k1_ec_pubkey_serialize(context, cPtr, &cLen, &key.pubkey, C.SECP256K1_EC_COMPRESSED)
	if ret != 1 {
		panic("failed serializing a pubkey. Should never happen (upstream promise to return 1)")
	}
	if cLen != SerializedECDSAPublicKeySize {
		panic("Returned length should be 33 because we passed SECP256K1_EC_COMPRESSED")
	}
	return &serialized, nil
}

// Add a tweak to the public key by doing `key + tweak*Generator`. this adds it in place.
// This is meant for creating BIP-32(HD) wallets
func (key *ECDSAPublicKey) Add(tweak [32]byte) error {
	if !key.init {
		return errors.WithStack(errNonInitializedKey)
	}
	cPtrTweak := (*C.uchar)(&tweak[0])
	ret := C.secp256k1_ec_pubkey_tweak_add(context, &key.pubkey, cPtrTweak)
	if ret != 1 {
		return errors.New("failed adding to the public key. Tweak is bigger than the order or the complement of the private key")
	}
	return nil
}

// Negate a public key in place.
// Equivalent to negating the private key and then generating the public key.
func (key *ECDSAPublicKey) Negate() error {
	if !key.init {
		return errors.WithStack(errNonInitializedKey)
	}
	ret := C.secp256k1_ec_pubkey_negate(C.secp256k1_context_no_precomp, &key.pubkey)
	if ret != 1 {
		panic("failed Negating the public key. Should never happen")
	}
	return nil
}

// ToSchnorr converts an ECDSA public key to a schnorr public key
// Note: You shouldn't sign with the same key in ECDSA and Schnorr signatures.
// this function is for convenience when using BIP-32
func (key *ECDSAPublicKey) ToSchnorr() (*SchnorrPublicKey, error) {
	if !key.init {
		return nil, errors.WithStack(errNonInitializedKey)
	}
	schnorrPubKey := SchnorrPublicKey{init: true}
	ret := C.secp256k1_xonly_pubkey_from_pubkey(C.secp256k1_context_no_precomp, &schnorrPubKey.pubkey, nil, &key.pubkey)
	if ret != 1 {
		panic("failed converting an ECDSA key to a schnorr key. Should never happen")
	}
	return &schnorrPubKey, nil
}
