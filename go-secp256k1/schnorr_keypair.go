package secp256k1

// #include "./depend/secp256k1/include/secp256k1_extrakeys.h"
// #include "./depend/secp256k1/include/secp256k1_schnorrsig.h"
import "C"
import "C"
import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"unsafe"
)

// SchnorrKeyPair is a type representing a pair of Secp256k1 private and public keys.
// This can be used to create Schnorr signatures
type SchnorrKeyPair struct {
	keypair C.secp256k1_keypair
	init    bool
}

// SerializedPrivateKey is a byte array representing the storage representation of a SchnorrKeyPair
type SerializedPrivateKey [SerializedPrivateKeySize]byte

// String returns the SchnorrKeyPair as the hexadecimal string
func (key SerializedPrivateKey) String() string {
	return hex.EncodeToString(key[:])
}

// String returns the SchnorrKeyPair as the hexadecimal string
func (key SchnorrKeyPair) String() string {
	return key.SerializePrivateKey().String()
}

// DeserializeSchnorrPrivateKey returns a SchnorrKeyPair type from a 32 byte private key.
// will verify it's a valid private key(Group Order > key > 0)
func DeserializeSchnorrPrivateKey(data *SerializedPrivateKey) (*SchnorrKeyPair, error) {
	key := SchnorrKeyPair{init: true}
	cPtrPrivateKey := (*C.uchar)(&data[0])

	ret := C.secp256k1_keypair_create(context, &key.keypair, cPtrPrivateKey)
	if ret != 1 {
		return nil, errors.New("invalid SchnorrKeyPair (zero or bigger than the group order)")
	}
	return &key, nil
}

// DeserializeSchnorrPrivateKeyFromSlice returns a SchnorrKeyPair type from a serialized private key slice.
// will verify that it's 32 byte and it's a valid private key(Group Order > key > 0)
func DeserializeSchnorrPrivateKeyFromSlice(data []byte) (key *SchnorrKeyPair, err error) {
	if len(data) != SerializedPrivateKeySize {
		return nil, errors.Errorf("invalid private key length got %d, expected %d", len(data),
			SerializedPrivateKeySize)
	}

	serializedKey := &SerializedPrivateKey{}
	copy(serializedKey[:], data)
	return DeserializeSchnorrPrivateKey(serializedKey)
}

// GenerateSchnorrKeyPair generates a random valid private key from `crypto/rand`
func GenerateSchnorrKeyPair() (key *SchnorrKeyPair, err error) {
	rawKey := SerializedPrivateKey{}
	for {
		n, err := rand.Read(rawKey[:])
		if err != nil {
			return nil, err
		}
		if n != len(rawKey) {
			panic("The standard library promises that this should never happen")
		}
		key, err = DeserializeSchnorrPrivateKey(&rawKey)
		if err == nil {
			return key, nil
		}
	}
}

// SerializePrivateKey returns the private key in the keypair.
func (key *SchnorrKeyPair) SerializePrivateKey() *SerializedPrivateKey {
	serialized := SerializedPrivateKey{}
	cPtr := (*C.uchar)(&serialized[0])
	ret := C.secp256k1_keypair_sec(C.secp256k1_context_no_precomp, cPtr, &key.keypair)
	if ret != 1 {
		panic("failed serializing the secret key. Should never happen (upstream promise to return 1)")
	}
	return &serialized
}

// Add a tweak to the public key by doing `key + tweak % Group Order` and adjust the pub/priv keys according to parity. this adds it in place.
// This is meant for creating BIP-32(HD) wallets
func (key *SchnorrKeyPair) Add(tweak [32]byte) error {
	if !key.init {
		return errors.WithStack(errNonInitializedKey)
	}
	cPtrTweak := (*C.uchar)(&tweak[0])
	ret := C.secp256k1_keypair_xonly_tweak_add(context, &key.keypair, cPtrTweak)
	if ret != 1 {
		return errors.New("failed Adding to private key. Tweak is bigger than the order or the complement of the private key")
	}
	return nil
}

// SchnorrPublicKey generates a PublicKey for the corresponding private key.
func (key *SchnorrKeyPair) SchnorrPublicKey() (*SchnorrPublicKey, error) {
	pubkey, _, err := key.schnorrPublicKeyInternal()
	return pubkey, err
}

func (key *SchnorrKeyPair) schnorrPublicKeyInternal() (pubkey *SchnorrPublicKey, wasOdd bool, err error) {
	if !key.init {
		return nil, false, errors.WithStack(errNonInitializedKey)
	}
	pubkey = &SchnorrPublicKey{init: true}
	cParity := C.int(42)
	ret := C.secp256k1_keypair_xonly_pub(context, &pubkey.pubkey, &cParity, &key.keypair)
	if ret != 1 {
		return nil, false, errors.New("the keypair contains invalid data")
	}

	return pubkey, parityBitToBool(cParity), nil

}

// SchnorrSign creates a schnorr signature using the private key and the input hashed message.
// Notice: the [32] byte array *MUST* be a hash of a message.
func (key *SchnorrKeyPair) SchnorrSign(hash *Hash) (*SchnorrSignature, error) {
	var auxilaryRand [32]byte
	n, err := rand.Read(auxilaryRand[:])
	if err != nil || n != len(auxilaryRand) {
		return nil, err
	}
	return key.schnorrSignInternal(hash, &auxilaryRand)
}

func (key *SchnorrKeyPair) schnorrSignInternal(hash *Hash, auxiliaryRand *[32]byte) (*SchnorrSignature, error) {
	if !key.init {
		return nil, errors.WithStack(errNonInitializedKey)
	}
	signature := SchnorrSignature{}
	cPtrSig := (*C.uchar)(&signature.signature[0])
	cPtrHash := (*C.uchar)(&hash[0])
	cPtrAux := unsafe.Pointer(auxiliaryRand)
	ret := C.secp256k1_schnorrsig_sign(context, cPtrSig, cPtrHash, &key.keypair, C.secp256k1_nonce_function_bip340, cPtrAux)
	if ret != 1 {
		return nil, errors.New("failed Signing. You should call `DeserializeSchnorrPrivateKey` before calling this")
	}
	return &signature, nil
}

func parityBitToBool(parity C.int) bool {
	switch parity {
	case 0:
		return false
	case 1:
		return true
	default:
		panic(fmt.Sprintf("should never happen, parity should always be 1 or 0, instead got: %d", int(parity)))
	}
}
