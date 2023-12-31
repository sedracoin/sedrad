package secp256k1

// #include "./depend/secp256k1/include/secp256k1.h"
import "C"
import (
	"crypto/rand"
	"github.com/pkg/errors"
	"unsafe"
)

// ECDSAPrivateKey is a type representing a Secp256k1 ECDSA private key.
type ECDSAPrivateKey struct {
	privateKey [SerializedPrivateKeySize]byte
	init       bool
}

// ECDSAPublicKey generates a PublicKey for the corresponding private key.
func (key *ECDSAPrivateKey) ECDSAPublicKey() (*ECDSAPublicKey, error) {
	if !key.init {
		return nil, errors.WithStack(errNonInitializedKey)
	}
	pubkey := ECDSAPublicKey{init: true}
	cPtrPrivateKey := (*C.uchar)(&key.privateKey[0])
	ret := C.secp256k1_ec_pubkey_create(context, &pubkey.pubkey, cPtrPrivateKey)
	if ret != 1 {
		return nil, errors.New("failed Generating an ECDSAPublicKey. You should call `DeserializeECDSAPrivateKey` before calling this")
	}
	return &pubkey, nil
}

// ECDSASign creates an ECDSA signature using the private key and the input hashed message.
// Notice: the [32] byte array *MUST* be a hash of a message.
func (key *ECDSAPrivateKey) ECDSASign(hash *Hash) (*ECDSASignature, error) {
	var auxilaryRand [32]byte
	n, err := rand.Read(auxilaryRand[:])
	if err != nil || n != len(auxilaryRand) {
		return nil, err
	}
	return key.ecdsaSignInternal(hash, &auxilaryRand)
}

func (key *ECDSAPrivateKey) ecdsaSignInternal(hash *Hash, auxiliaryRand *[32]byte) (*ECDSASignature, error) {
	if !key.init {
		return nil, errNonInitializedKey
	}
	signature := ECDSASignature{}
	cPtrHash := (*C.uchar)(&hash[0])
	cPtrPrivKey := (*C.uchar)(&key.privateKey[0])
	cPtrAux := unsafe.Pointer(auxiliaryRand)
	ret := C.secp256k1_ecdsa_sign(context, &signature.signature, cPtrHash, cPtrPrivKey, C.secp256k1_nonce_function_rfc6979, cPtrAux)
	if ret != 1 {
		return nil, errors.New("failed Signing. You should call `DeserializeECDSAPrivateKey` before calling this")
	}
	return &signature, nil
}

// String returns the ECDSAPrivateKey as the hexadecimal string
func (key ECDSAPrivateKey) String() string {
	return key.Serialize().String()
}

// DeserializeECDSAPrivateKey returns a ECDSAPrivateKey type from a 32 byte private key.
// will verify it's a valid private key(Group Order > key > 0)
func DeserializeECDSAPrivateKey(data *SerializedPrivateKey) (key *ECDSAPrivateKey, err error) {
	cPtr := (*C.uchar)(&data[0])

	ret := C.secp256k1_ec_seckey_verify(C.secp256k1_context_no_precomp, cPtr)
	if ret != 1 {
		return nil, errors.New("invalid ECDSAPrivateKey (zero or bigger than the group order)")
	}

	return &ECDSAPrivateKey{privateKey: *data, init: true}, nil
}

// DeserializeECDSAPrivateKeyFromSlice returns a ECDSAPrivateKey type from a serialized private key slice.
// will verify that it's 32 byte and it's a valid private key(Group Order > key > 0)
func DeserializeECDSAPrivateKeyFromSlice(data []byte) (key *ECDSAPrivateKey, err error) {
	if len(data) != SerializedPrivateKeySize {
		return nil, errors.Errorf("invalid ECDSA private key length got %d, expected %d", len(data),
			SerializedPrivateKeySize)
	}

	serializedKey := &SerializedPrivateKey{}
	copy(serializedKey[:], data)
	return DeserializeECDSAPrivateKey(serializedKey)
}

// GenerateECDSAPrivateKey generates a random valid private key from `crypto/rand`
func GenerateECDSAPrivateKey() (key *ECDSAPrivateKey, err error) {
	key = &ECDSAPrivateKey{init: true}
	cPtr := (*C.uchar)(&key.privateKey[0])
	for {
		n, tmpErr := rand.Read(key.privateKey[:])
		if tmpErr != nil {
			return nil, tmpErr
		}
		if n != len(key.privateKey) {
			panic("The standard library promises that this should never happen")
		}
		ret := C.secp256k1_ec_seckey_verify(C.secp256k1_context_no_precomp, cPtr)
		if ret == 1 {
			return key, nil
		}
	}
}

// Serialize a private key
func (key *ECDSAPrivateKey) Serialize() *SerializedPrivateKey {
	ret := SerializedPrivateKey(key.privateKey)
	return &ret
}

// Negate a private key in place.
func (key *ECDSAPrivateKey) Negate() error {
	if !key.init {
		return errNonInitializedKey
	}
	cPtr := (*C.uchar)(&key.privateKey[0])
	ret := C.secp256k1_ec_privkey_negate(C.secp256k1_context_no_precomp, cPtr)
	if ret != 1 {
		panic("Failed Negating the private key. Should never happen")
	}
	return nil
}

// Add a tweak to the private key by doing `key + tweak % Group Order`. this adds it in place.
// This is meant for creating BIP-32(HD) wallets
func (key *ECDSAPrivateKey) Add(tweak [32]byte) error {
	if !key.init {
		return errNonInitializedKey
	}
	cPtrKey := (*C.uchar)(&key.privateKey[0])
	cPtrTweak := (*C.uchar)(&tweak[0])
	ret := C.secp256k1_ec_privkey_tweak_add(C.secp256k1_context_no_precomp, cPtrKey, cPtrTweak)
	if ret != 1 {
		return errors.New("failed Adding to private key. Tweak is bigger than the order or the complement of the private key")
	}
	return nil
}

// ToSchnorr converts an ECDSA private key to a schnorr keypair
// Note: You shouldn't sign using the same key in both ECDSA and Schnorr signatures.
// this function is for convenience when using BIP-32
func (key *ECDSAPrivateKey) ToSchnorr() (*SchnorrKeyPair, error) {
	if !key.init {
		return nil, errors.WithStack(errNonInitializedKey)
	}
	return DeserializeSchnorrPrivateKey((*SerializedPrivateKey)(&key.privateKey))
}
