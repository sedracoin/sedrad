package secp256k1

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
)

const loopsN = 150

var Secp256k1Order = new(big.Int).SetBytes([]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 254, 186, 174, 220, 230, 175, 72, 160, 59, 191, 210, 94, 140, 208, 54, 65, 65})
var algorithms = []alogirthmInterface{new(schnorr), new(ecdsa)}

func ForAllAlgorithms(t *testing.T, testFunc func(*testing.T, *rand.Rand, alogirthmInterface)) {
	for _, alg := range algorithms {
		algCopy := alg
		t.Run(algCopy.String(), func(t *testing.T) {
			t.Parallel()
			testFunc(t, rand.New(rand.NewSource(42)), algCopy)
		})
	}
}

func intTo32Bytes(i *big.Int) [32]byte {
	res := [32]byte{}
	serialized := i.Bytes()
	copy(res[32-len(serialized):], serialized)
	return res
}

func negateSecp256k1Tweak(tweak []byte) {
	bigTweak := new(big.Int).SetBytes(tweak)
	bigTweak.Neg(bigTweak)
	bigTweak.Mod(bigTweak, Secp256k1Order)
	res := intTo32Bytes(bigTweak)
	copy(tweak, res[:])
}

func fastGenerateTweak(t testing.TB, r *rand.Rand) *[32]byte {
	buf := [32]byte{}
	for {
		n, err := r.Read(buf[:])
		if err != nil || n != len(buf) {
			t.Fatalf("Failed generating 32 random bytes '%s'", err)
		}
		_, err = DeserializeSchnorrPrivateKey((*SerializedPrivateKey)(&buf))
		if err == nil {
			return &buf
		}
	}
}

func TestECDSAPublicKey_ToSchnorr(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	serializedPrivKey := (*SerializedPrivateKey)(fastGenerateTweak(t, r))
	ecdsaPrivKey, err := DeserializeECDSAPrivateKey(serializedPrivKey)
	if err != nil {
		t.Fatalf("A valid tweak should be a valid private key: '%s'", err)
	}
	schnorrKeyPair, err := DeserializeSchnorrPrivateKey(serializedPrivKey)
	if err != nil {
		t.Fatalf("A valid tweak should be a valid private key: '%s'", err)
	}
	ecdsaPubKey, err := ecdsaPrivKey.ECDSAPublicKey()
	if err != nil {
		t.Fatalf("A valid privkey should convert to a pubkey: '%s'", err)
	}
	schnorrPublicKey, err := schnorrKeyPair.SchnorrPublicKey()
	if err != nil {
		t.Fatalf("A valid privkey should convert to a pubkey: '%s'", err)
	}
	convertedSchnorrPublicKey, err := ecdsaPubKey.ToSchnorr()
	if err != nil {
		t.Fatalf("A valid ECDSA public key should convert to a valid schnorr public key: '%s'", err)
	}
	if !schnorrPublicKey.IsEqual(convertedSchnorrPublicKey) {
		t.Fatalf("Expected %s == %s", schnorrPublicKey, convertedSchnorrPublicKey)
	}

	serializedECDSAPubKey, err := ecdsaPubKey.Serialize()
	if err != nil {
		t.Fatalf("A valid pubkey should serialize: '%s'", err)
	}
	serializedSchnorrPubKey, err := schnorrPublicKey.Serialize()
	if err != nil {
		t.Fatalf("A valid pubkey should serialize: '%s'", err)
	}

	// a schnorr pubkey is an ecdsa pubkey without the parity bit at the start.
	if !bytes.Equal(serializedECDSAPubKey[1:], serializedSchnorrPubKey[:]) {
		t.Fatalf("Expected %x == %x", serializedECDSAPubKey[1:], serializedSchnorrPubKey[:])
	}
}

func TestParseSerializePrivateKey(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		for i := 0; i < loopsN; i++ {
			privkey := alg.EmptyPrivKey().GenerateNew(t, r)
			serialized := privkey.Serialize()
			privkey2, err := alg.EmptyPrivKey().DeserializeNew(serialized[:])
			if err != nil {
				t.Errorf("Failed parsing privateKey '%s'", err)
			}
			if !privkey.IsEqual(privkey2) {
				t.Errorf("Privkeys aren't equal '%s' '%s'", privkey, privkey2)
			}
		}
	})

}

func TestGeneratePrivateKey(t *testing.T) {
	_, err := GenerateSchnorrKeyPair()
	if err != nil {
		t.Errorf("Failed generating a privatekey '%s'", err)
	}
	_, err = GenerateECDSAPrivateKey()
	if err != nil {
		t.Errorf("Failed generating a privatekey '%s'", err)
	}
}

func TestPrivateKey_Add_Fail(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		for i := 0; i < loopsN; i++ {
			privkey := alg.EmptyPrivKey().GenerateNew(t, r)
			privkeyInverse := privkey.Serialize()
			_, isNegated, err := privkey.PublicKey()
			if err != nil {
				t.Fatalf("Failed converting privkey to pubkey: %s", err)
			}
			// If the key is already being negated then the serialized one is already the inverse of the privkey.
			if !isNegated {
				negateSecp256k1Tweak(privkeyInverse[:])
			}
			err = privkey.Add(*privkeyInverse)
			if err == nil {
				t.Errorf("Adding the inverse of itself should fail, '%s', '%x', '%s'", privkey, privkeyInverse, err)
			}
			privkey = alg.EmptyPrivKey().GenerateNew(t, r)
			oufOfBounds := [32]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
			err = privkey.Add(oufOfBounds)
			if err == nil {
				t.Errorf("Adding a tweak bigger than the order should fail, '%s', '%x' '%s'", privkey, oufOfBounds, err)
			}
		}
	})
}

func TestPrivateKey_Add(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		privkey := alg.EmptyPrivKey().GenerateNew(t, r)
		pubkey, wasOdd, err := privkey.PublicKey()
		if err != nil {
			t.Fatal(err)
		}
		privkeyBig := new(big.Int).SetBytes(privkey.Serialize()[:])
		seedBig := big.Int{}

		for i := 0; i < loopsN; i++ {
			if wasOdd { // Schnorr secret keys are always even, so if libsecp negated the key then we should too.
				privkeyBig.Neg(privkeyBig)
			}
			seed := *fastGenerateTweak(t, r)
			seedBig.SetBytes(seed[:])

			privkeyBig.Add(privkeyBig, &seedBig)
			privkeyBig.Mod(privkeyBig, Secp256k1Order)
			err := privkey.Add(seed)
			if err != nil {
				t.Fatalf("failed adding seed: '%s' to key: '%s'", seed, privkey)
			}
			wasOdd, err = pubkey.Add(seed)
			if err != nil { // This shouldn't fail if the same operation for the privateKey didn't fail.
				t.Fatal(err)
			}

			tmpPubKey, _, err := privkey.PublicKey()
			if err != nil {
				t.Fatalf("Failed generating pubkey from '%s'. '%s'", privkey, err)
			}

			if intTo32Bytes(privkeyBig) != *privkey.Serialize() {
				t.Fatalf("Add operation failed, i=%d '%x' != '%x'", i, intTo32Bytes(privkeyBig), privkey.Serialize())
			}
			if !pubkey.IsEqual(tmpPubKey) {
				t.Fatalf("tweaked pubkey '%s' doesn't match tweaked privateKey '%s', '%s'", pubkey, tmpPubKey, privkey)
			}
		}
	})
}

func TestParsePublicKeyFail(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		pubkeyAlg := alg.EmptyPubKey()
		zeros := [32]byte{}
		max := [32]byte{}
		for i := range max {
			max[i] = 0xff
		}
		_, err := pubkeyAlg.DeserializeNew(zeros[:])
		if err == nil {
			t.Errorf("Shouldn't parse 32 zeros as a pubkey '%x'", zeros)
		}
		_, err = pubkeyAlg.DeserializeNew(zeros[:30])
		if err == nil {
			t.Errorf("Shouldn't parse 30 zeros as a pubkey '%x'", zeros[:30])
		}
		_, err = pubkeyAlg.DeserializeNew(max[:])
		if err == nil {
			t.Errorf("Shouldn't parse 32 0xFF as a pubkey '%x' (it's above the field order)", max)
		}
	})
}

func TestPublicKey_SerializeFail(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		pubkeyAlg := alg.EmptyPubKey()
		_, err := pubkeyAlg.Serialize()
		if err == nil {
			t.Errorf("Zeroed public key isn't serializable as compressed")
		}
	})
}

func TestBadPrivateKeyPublicKeyFail(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		goodPrivKey := alg.EmptyPrivKey().GenerateNew(t, r)
		goodPublicKey, _, err := goodPrivKey.PublicKey()
		if err != nil {
			t.Fatalf("Failed generating pubkey from: '%s'. '%s'", goodPrivKey, err)
		}
		goodPublicKeyBackup := goodPublicKey.Clone()
		goodPrivKeyBackup := goodPrivKey.Clone()
		msg := Hash(*fastGenerateTweak(t, r))
		privkey := alg.EmptyPrivKey()
		var zeros32 [32]byte

		_, _, err1 := privkey.PublicKey()
		_, err2 := privkey.Sign(&msg)
		err3 := privkey.Add(zeros32)
		_, err4 := alg.EmptyPrivKey().DeserializeNew(privkey.Serialize()[:])
		if err1 == nil || err2 == nil || err3 == nil || err4 == nil {
			t.Errorf("A zeroed key is invalid, err1: '%s', err2: '%s', err3: '%s', err4: '%s'", err1, err2, err3, err4)
		}

		err5 := goodPrivKey.Add(zeros32)
		_, err6 := goodPublicKey.Add(zeros32)
		if err5 != nil || err6 != nil {
			t.Errorf("It should be possible to add zero to a key, err4: '%s', err5: '%s'", err5, err6)
		}

		privkey.SetBytes(Secp256k1Order.Bytes())
		_, _, err1 = privkey.PublicKey()
		_, err2 = privkey.Sign(&msg)
		_, err3 = alg.EmptyPrivKey().DeserializeNew(privkey.Serialize()[:])
		goodPrivKey = goodPrivKeyBackup.Clone()
		goodPublicKey = goodPublicKeyBackup.Clone()
		err4 = goodPrivKey.Add(intTo32Bytes(Secp256k1Order))
		_, err5 = goodPublicKey.Add(intTo32Bytes(Secp256k1Order))
		if err1 == nil || err2 == nil || err3 == nil || err4 == nil || err5 == nil {
			t.Errorf("the group order isn't a valid key, err1: '%s', err2: '%s', err3: '%s', err4: '%s', err5: '%s'", err1, err2, err3, err4, err5)
		}
		privkey = alg.EmptyPrivKey().GenerateNew(t, r)
		orderPlusOne := new(big.Int).SetInt64(1)
		orderPlusOne.Add(orderPlusOne, Secp256k1Order)
		privkey.SetBytes(orderPlusOne.Bytes())
		orderPlusOneArray := intTo32Bytes(orderPlusOne)
		_, err = alg.EmptyPrivKey().DeserializeNew(orderPlusOneArray[:])
		if err1 == nil || err2 == nil || err3 == nil || err4 == nil {
			t.Errorf("A key bigger than the group order isn't a valid key, err: '%s'", err)
		}
		goodPrivKey = goodPrivKeyBackup.Clone()
		goodPublicKey = goodPublicKeyBackup.Clone()

		err1 = goodPrivKey.Add(intTo32Bytes(orderPlusOne))
		_, err2 = goodPublicKey.Add(intTo32Bytes(orderPlusOne))
		if err1 == nil || err2 == nil {
			t.Errorf("A tweak bigger than the group order isn't a valid tweak, err1: '%s', err2: '%s'", err1, err2)
		}
		orderMinusOne := new(big.Int).Sub(Secp256k1Order, new(big.Int).SetInt64(1))
		orderPlusOne = nil
		privkey.SetBytes(orderMinusOne.Bytes())
		_, _, err1 = privkey.PublicKey()
		_, err2 = privkey.Sign(&msg)
		_, err3 = alg.EmptyPrivKey().DeserializeNew(privkey.Serialize()[:])
		goodPrivKey = goodPrivKeyBackup.Clone()
		goodPublicKey = goodPublicKeyBackup.Clone()
		err4 = goodPrivKey.Add(intTo32Bytes(orderMinusOne))
		_, err5 = goodPublicKey.Add(intTo32Bytes(orderMinusOne))
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
			t.Errorf("Group order - 1 should be a valid key, err1: '%s', err2: '%s', err3: '%s', err4: '%s', err5: '%s'", err1, err2, err3, err4, err5)
		}
	})
}

func TestParsePubKey(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		for i := 0; i < loopsN; i++ {
			privkey := alg.EmptyPrivKey().GenerateNew(t, r)
			pubkey, _, err := privkey.PublicKey()
			if err != nil {
				t.Errorf("Failed Generating a pubkey from privateKey: '%s'. '%s'", privkey, err)
			}
			serializedPubkey, err := pubkey.Serialize()
			if err != nil {
				t.Errorf("Failed serializing the key: %s, error: '%s'", pubkey, err)
			}
			pubkeyNew1, err := pubkey.DeserializeNew(serializedPubkey[:])
			if err != nil {
				t.Errorf("Failed Parsing the compressed public key from keypair: '%s'. '%s'", pubkeyNew1, err)
			}
			if !pubkey.IsEqual(pubkeyNew1) {
				t.Errorf("Pubkeys aren't the same: '%s', '%s',", pubkey, pubkeyNew1)
			}
		}
	})
}

func TestSignVerifyParse(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		for i := 0; i < loopsN; i++ {
			privkey := alg.EmptyPrivKey().GenerateNew(t, r)
			pubkey, _, err := privkey.PublicKey()
			if err != nil {
				t.Errorf("Failed generating a pubkey, privateKey: '%s', error: %s", privkey, err)
			}
			msg := Hash{}
			n, err := r.Read(msg[:])
			if err != nil || n != 32 {
				t.Errorf("Failed generating a msg. read: '%d' bytes. .'%s'", n, err)
			}
			sig1, err := privkey.Sign(&msg)
			if err != nil {
				t.Errorf("Failed signing: key: '%s', msg: '%s', error: '%s'", privkey, msg, err)
			}
			sig2, err := privkey.Sign(&msg)
			if err != nil {
				t.Errorf("Failed signing: key: '%s', msg: '%s', error: '%s'", privkey, msg, err)
			}
			if sig1.IsEqual(sig2) {
				t.Errorf("Signing uses auxilary randomness, the odds of 2 signatures being the same is 1/2^128 '%s' '%s'", sig1, sig2)
			}
			serialized := sig1.Serialize()
			sigDeserialized, err := sig2.DeserializeNew(serialized[:])
			if err != nil {
				t.Errorf("Failed deserializing sig: '%s', error: '%s'", serialized, err)
			}
			if !sig1.IsEqual(sigDeserialized) {
				t.Errorf("Failed Deserializing signatureInterface '%s'", serialized)
			}
			if !pubkey.Verify(&msg, sig1) {
				t.Errorf("Failed verifying signatureInterface privateKey: '%s' pubkey: '%s' signatureInterface: '%s'", privkey, pubkey, sig1)
			}
			if !pubkey.Verify(&msg, sig2) {
				t.Errorf("Failed verifying signatureInterface privateKey: '%s' pubkey: '%s' signatureInterface: '%s'", privkey, pubkey, sig2)
			}
		}
	})
}

func TestPublicKey_IsEqual(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		goodPrivKey := alg.EmptyPrivKey().GenerateNew(t, r)
		goodPublicKey, _, err := goodPrivKey.PublicKey()
		if err != nil {
			t.Fatalf("Failed generating pubkey from: '%s'. '%s'", goodPrivKey, err)
		}
		badPublicKey := alg.EmptyPubKey()
		if badPublicKey.IsEqual(goodPublicKey) {
			t.Errorf("Empty publickey shouldn't be equal to good one")
		}
		if !badPublicKey.IsEqual(alg.EmptyPubKey()) {
			t.Errorf("Empty publickey should be equal to another empty pubkey")
		}
		ty := reflect.TypeOf(goodPublicKey)
		nilPubKey := reflect.Zero(ty).Interface().(publicKeyInterface)
		if nilPubKey.IsEqual(goodPublicKey) {
			t.Fatalf("nil publickey shouldn't be equal to good one")
		}

		if !nilPubKey.IsEqual(nil) {
			t.Fatalf("two nil pubkeys should be equal")
		}

		copyGoodPubkey := goodPublicKey.Clone()
		if !copyGoodPubkey.IsEqual(goodPublicKey) {
			t.Errorf("A pubkey and its copy should be the same")
		}
		goodPrivKey2 := alg.EmptyPrivKey().GenerateNew(t, r)
		goodPublicKey2, _, err := goodPrivKey2.PublicKey()
		if err != nil {
			t.Fatalf("Failed generating pubkey from: '%s'. '%s'", goodPrivKey2, err)
		}

		if goodPublicKey.IsEqual(goodPublicKey2) {
			t.Errorf("'%s' shouldn't be equal to %s", goodPublicKey, goodPublicKey2)
		}
	})
}

func TestSignature_IsEqual(t *testing.T) {
	ForAllAlgorithms(t, func(t *testing.T, r *rand.Rand, alg alogirthmInterface) {
		var serializedSig [64]byte
		n, err := r.Read(serializedSig[:])
		if err != nil || n != len(serializedSig) {
			t.Errorf("Failed generating a random signatureInterface. read: '%d' bytes.. '%s'", n, err)
		}
		goodSignature, err := alg.EmptySignature().DeserializeNew(serializedSig[:])
		if err != nil {
			t.Fatalf("Failed deserializing signatureInterface: %s", goodSignature)
		}

		emptySignature := alg.EmptySignature()
		if emptySignature.IsEqual(goodSignature) {
			t.Errorf("Empty signatureInterface shouldn't be equal to good one")
		}
		if !emptySignature.IsEqual(alg.EmptySignature()) {
			t.Errorf("Empty signatureInterface should be equal to another empty signatureInterface")
		}
		ty := reflect.TypeOf(alg.EmptySignature())
		nilSignature := reflect.Zero(ty).Interface().(signatureInterface)
		if nilSignature.IsEqual(goodSignature) {
			t.Errorf("nil signatureInterface shouldn't be equal to good one")
		}

		if !nilSignature.IsEqual(nil) {
			t.Errorf("two nil signatures should be equal")
		}

		copyGoodSignature := goodSignature.Clone()
		if !copyGoodSignature.IsEqual(goodSignature) {
			t.Errorf("A signatureInterface and its copy should be the same")
		}

		var serializedSig2 [64]byte
		n, err = r.Read(serializedSig2[:])
		if err != nil || n != len(serializedSig2) {
			t.Errorf("Failed generating a random signatureInterface. read: '%d' bytes.. '%s'", n, err)
		}

		goodSignature2, err := alg.EmptySignature().DeserializeNew(serializedSig2[:])
		if err != nil {
			t.Fatalf("Failed deserializing signature: %s", serializedSig)
		}
		if goodSignature.IsEqual(goodSignature2) {
			t.Errorf("'%s' shouldn't be equal to %s", goodSignature, goodSignature2)
		}
	})
}

func TestHash_IsEqual(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	goodHash := &Hash{}
	n, err := r.Read(goodHash[:])
	if err != nil || n != len(goodHash) {
		t.Errorf("Failed generating a random hash. read: '%d' bytes.. '%s'", n, err)
	}

	emptyHash := Hash{}
	if emptyHash.IsEqual(goodHash) {
		t.Errorf("Empty hash shouldn't be equal to filled one")
	}
	if !emptyHash.IsEqual(&Hash{}) {
		t.Errorf("Empty hash should be equal to another empty hash")
	}
	var nilHash *Hash = nil
	if nilHash.IsEqual(goodHash) {
		t.Errorf("nil hash shouldn't be equal to good one")
	}

	if !nilHash.IsEqual(nil) {
		t.Errorf("two nil hashes should be equal")
	}

	copyGoodHash := *goodHash
	if !copyGoodHash.IsEqual(goodHash) {
		t.Errorf("A hash and its copy should be the same")
	}

	goodHash2 := &Hash{}
	n, err = r.Read(goodHash2[:])
	if err != nil || n != len(goodHash2) {
		t.Errorf("Failed generating a random hash. read: '%d' bytes. .'%s'", n, err)
	}

	if goodHash.IsEqual(goodHash2) {
		t.Errorf("'%s' shouldn't be equal to %s", goodHash, goodHash2)
	}
}

func BenchmarkVerify(b *testing.B) {
	for _, alg := range algorithms {
		algCopy := alg
		b.Run(algCopy.String(), func(b *testing.B) {
			r := rand.New(rand.NewSource(1))
			sigs := make([]signatureInterface, loopsN)
			msgs := make([]Hash, loopsN)
			pubkeys := make([]publicKeyInterface, loopsN)
			for i := 0; i < loopsN; i++ {
				msg := Hash{}
				n, err := r.Read(msg[:])
				if err != nil || n != 32 {
					panic(fmt.Sprintf("benchmark failed: '%s', n: %d", err, n))
				}
				privkey := alg.EmptyPrivKey().GenerateNew(b, r)
				sigTmp, err := privkey.Sign(&msg)
				if err != nil {
					panic("benchmark failed: " + err.Error())
				}
				sigs[i] = sigTmp
				pubkeyTmp, _, err := privkey.PublicKey()
				if err != nil {
					panic("benchmark failed: " + err.Error())
				}
				pubkeys[i] = pubkeyTmp
				msgs[i] = msg
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pubkeys[i%loopsN].Verify(&msgs[i%loopsN], sigs[i%loopsN])
			}
		})
	}
}
