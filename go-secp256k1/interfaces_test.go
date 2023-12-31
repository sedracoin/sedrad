package secp256k1

import (
	"fmt"
	"math/rand"
	"testing"
)

type alogirthmInterface interface {
	EmptyPrivKey() privateKeyInterface
	EmptyPubKey() publicKeyInterface
	EmptySignature() signatureInterface
	fmt.Stringer
}

type privateKeyInterface interface {
	GenerateNew(testing.TB, *rand.Rand) privateKeyInterface
	Sign(*Hash) (signatureInterface, error)
	Add([32]byte) error
	PublicKey() (publicKeyInterface, bool, error)
	Serialize() *[32]byte
	DeserializeNew([]byte) (privateKeyInterface, error)
	SetBytes([]byte)
	IsEqual(privateKeyInterface) bool
	Clone() privateKeyInterface
	fmt.Stringer
}
type publicKeyInterface interface {
	Verify(*Hash, signatureInterface) bool
	Add([32]byte) (isNegated bool, err error)
	Serialize() ([]byte, error)
	DeserializeNew([]byte) (publicKeyInterface, error)
	IsEqual(other publicKeyInterface) bool
	Clone() publicKeyInterface
	fmt.Stringer
}
type signatureInterface interface {
	Serialize() *[64]byte
	DeserializeNew([]byte) (signatureInterface, error)
	IsEqual(signatureInterface) bool
	Clone() signatureInterface
	fmt.Stringer
}

type schnorr struct{}

func (s schnorr) EmptyPrivKey() privateKeyInterface  { return new(schnorrPrivKey) }
func (s schnorr) EmptyPubKey() publicKeyInterface    { return new(schnorrPubkey) }
func (s schnorr) EmptySignature() signatureInterface { return new(schnorrSignature) }
func (s schnorr) String() string                     { return "schnorr" }

type ecdsa struct{}

func (s ecdsa) EmptyPrivKey() privateKeyInterface  { return new(ecdsaPrivKey) }
func (s ecdsa) EmptyPubKey() publicKeyInterface    { return new(ecdsaPubkey) }
func (s ecdsa) EmptySignature() signatureInterface { return new(ecdsaSignature) }
func (s ecdsa) String() string                     { return "ecdsa" }

type schnorrPrivKey SchnorrKeyPair

func (s *schnorrPrivKey) GenerateNew(t testing.TB, r *rand.Rand) privateKeyInterface {
	buf := fastGenerateTweak(t, r)
	keypair, err := DeserializeSchnorrPrivateKey((*SerializedPrivateKey)(buf))
	if err != nil {
		t.Fatalf("A valid tweak should be a valid private key: '%s'", err)
	}
	return (*schnorrPrivKey)(keypair)
}
func (s *schnorrPrivKey) Sign(hash *Hash) (signatureInterface, error) {
	sig, err := (*SchnorrKeyPair)(s).SchnorrSign(hash)
	return (*schnorrSignature)(sig), err
}
func (s *schnorrPrivKey) Add(tweak [32]byte) error {
	return (*SchnorrKeyPair)(s).Add(tweak)
}
func (s *schnorrPrivKey) PublicKey() (publicKeyInterface, bool, error) {
	pubkey, isNegated, err := (*SchnorrKeyPair)(s).schnorrPublicKeyInternal()
	return (*schnorrPubkey)(pubkey), isNegated, err
}
func (s *schnorrPrivKey) Serialize() *[32]byte {
	ret := (*SchnorrKeyPair)(s).SerializePrivateKey()
	return (*[32]byte)(ret)
}
func (s *schnorrPrivKey) DeserializeNew(bytes []byte) (privateKeyInterface, error) {
	key, err := DeserializeSchnorrPrivateKeyFromSlice(bytes)
	return (*schnorrPrivKey)(key), err
}
func (s *schnorrPrivKey) SetBytes(bytes []byte) {
	for i := 0; i < len(bytes); i++ {
		(*SchnorrKeyPair)(s).keypair.data[i] = _Ctype_uchar(bytes[i])
	}
}
func (s *schnorrPrivKey) IsEqual(other privateKeyInterface) bool {
	otherKey, ok := other.(*schnorrPrivKey)
	if !ok {
		return false
	}
	return *s == *otherKey
}
func (s schnorrPrivKey) Clone() privateKeyInterface {
	return &s
}
func (s schnorrPrivKey) String() string {
	return (SchnorrKeyPair)(s).String()
}

type ecdsaPrivKey ECDSAPrivateKey

func (s *ecdsaPrivKey) GenerateNew(t testing.TB, r *rand.Rand) privateKeyInterface {
	buf := fastGenerateTweak(t, r)
	privkey, err := DeserializeECDSAPrivateKey((*SerializedPrivateKey)(buf))
	if err != nil {
		t.Fatalf("A valid tweak should be a valid private key: '%s'", err)
	}
	return (*ecdsaPrivKey)(privkey)
}
func (s *ecdsaPrivKey) Sign(hash *Hash) (signatureInterface, error) {
	sig, err := (*ECDSAPrivateKey)(s).ECDSASign(hash)
	return (*ecdsaSignature)(sig), err
}
func (s *ecdsaPrivKey) Add(tweak [32]byte) error {
	return (*ECDSAPrivateKey)(s).Add(tweak)
}
func (s *ecdsaPrivKey) PublicKey() (publicKeyInterface, bool, error) {
	pubkey, err := (*ECDSAPrivateKey)(s).ECDSAPublicKey()
	// ECDSA is never negated
	return (*ecdsaPubkey)(pubkey), false, err
}

func (s *ecdsaPrivKey) Serialize() *[32]byte {
	ret := (*ECDSAPrivateKey)(s).Serialize()
	return (*[32]byte)(ret)
}
func (s *ecdsaPrivKey) DeserializeNew(bytes []byte) (privateKeyInterface, error) {
	key, err := DeserializeECDSAPrivateKeyFromSlice(bytes)
	return (*ecdsaPrivKey)(key), err
}
func (s *ecdsaPrivKey) SetBytes(bytes []byte) {
	copy((*ECDSAPrivateKey)(s).privateKey[:], bytes)
}
func (s *ecdsaPrivKey) IsEqual(other privateKeyInterface) bool {
	otherKey, ok := other.(*ecdsaPrivKey)
	if !ok {
		return false
	}
	return *s == *otherKey
}
func (s ecdsaPrivKey) Clone() privateKeyInterface {
	return &s
}
func (s ecdsaPrivKey) String() string {
	return (ECDSAPrivateKey)(s).String()
}

type schnorrPubkey SchnorrPublicKey

func (s *schnorrPubkey) Verify(hash *Hash, sig signatureInterface) bool {
	return (*SchnorrPublicKey)(s).SchnorrVerify(hash, (*SchnorrSignature)(sig.(*schnorrSignature)))
}
func (s *schnorrPubkey) Add(tweak [32]byte) (isNegated bool, err error) {
	return (*SchnorrPublicKey)(s).addInternal(tweak)
}
func (s *schnorrPubkey) Serialize() ([]byte, error) {
	serialized, err := (*SchnorrPublicKey)(s).Serialize()
	if err != nil {
		return nil, err
	}
	return serialized[:], nil
}
func (s *schnorrPubkey) DeserializeNew(bytes []byte) (publicKeyInterface, error) {
	key, err := DeserializeSchnorrPubKey(bytes)
	return (*schnorrPubkey)(key), err
}
func (s *schnorrPubkey) IsEqual(other publicKeyInterface) bool {
	if other == nil {
		other = (*schnorrPubkey)(nil)
	}
	otherKey, ok := other.(*schnorrPubkey)
	if !ok {
		return false
	}
	return (*SchnorrPublicKey)(s).IsEqual((*SchnorrPublicKey)(otherKey))

}
func (s schnorrPubkey) Clone() publicKeyInterface {
	return &s
}
func (s schnorrPubkey) String() string {
	return (SchnorrPublicKey)(s).String()
}

type ecdsaPubkey ECDSAPublicKey

func (s *ecdsaPubkey) Verify(hash *Hash, sig signatureInterface) bool {
	return (*ECDSAPublicKey)(s).ECDSAVerify(hash, (*ECDSASignature)(sig.(*ecdsaSignature)))
}
func (s *ecdsaPubkey) Add(tweak [32]byte) (isNegated bool, err error) {
	// ECDSA Add never negates.
	return false, (*ECDSAPublicKey)(s).Add(tweak)
}
func (s *ecdsaPubkey) Serialize() ([]byte, error) {
	serialized, err := (*ECDSAPublicKey)(s).Serialize()
	if err != nil {
		return nil, err
	}
	return serialized[:], nil
}
func (s *ecdsaPubkey) DeserializeNew(bytes []byte) (publicKeyInterface, error) {
	key, err := DeserializeECDSAPubKey(bytes)
	return (*ecdsaPubkey)(key), err
}
func (s *ecdsaPubkey) IsEqual(other publicKeyInterface) bool {
	if other == nil {
		other = (*ecdsaPubkey)(nil)
	}
	otherKey, ok := other.(*ecdsaPubkey)
	if !ok {
		return false
	}
	return (*ECDSAPublicKey)(s).IsEqual((*ECDSAPublicKey)(otherKey))
}
func (s ecdsaPubkey) Clone() publicKeyInterface {
	return &s
}
func (s ecdsaPubkey) String() string {
	return (ECDSAPublicKey)(s).String()
}

type ecdsaSignature ECDSASignature

func (s *ecdsaSignature) Serialize() *[64]byte {
	return (*[64]byte)((*ECDSASignature)(s).Serialize())
}
func (s *ecdsaSignature) DeserializeNew(bytes []byte) (signatureInterface, error) {
	key, err := DeserializeECDSASignatureFromSlice(bytes)
	return (*ecdsaSignature)(key), err
}
func (s *ecdsaSignature) IsEqual(other signatureInterface) bool {
	if other == nil {
		other = (*ecdsaSignature)(nil)
	}
	otherKey, ok := other.(*ecdsaSignature)
	if !ok {
		return false
	}
	return (*ECDSASignature)(s).IsEqual((*ECDSASignature)(otherKey))

}
func (s ecdsaSignature) Clone() signatureInterface {
	return &s
}
func (s ecdsaSignature) String() string {
	return (ECDSASignature)(s).String()
}

type schnorrSignature SchnorrSignature

func (s *schnorrSignature) Serialize() *[64]byte {
	return (*[64]byte)((*SchnorrSignature)(s).Serialize())
}
func (s *schnorrSignature) DeserializeNew(bytes []byte) (signatureInterface, error) {
	key, err := DeserializeSchnorrSignatureFromSlice(bytes)
	return (*schnorrSignature)(key), err
}
func (s *schnorrSignature) IsEqual(other signatureInterface) bool {
	if other == nil {
		other = (*schnorrSignature)(nil)
	}
	otherKey, ok := other.(*schnorrSignature)
	if !ok {
		return false
	}
	return (*SchnorrSignature)(s).IsEqual((*SchnorrSignature)(otherKey))

}
func (s schnorrSignature) Clone() signatureInterface {
	return &s
}
func (s schnorrSignature) String() string {
	return (SchnorrSignature)(s).String()
}
