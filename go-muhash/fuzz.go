// +build gofuzz

package muhash

import "C"
import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/bits"
	"unsafe"
)

var (
	mainInt = new(big.Int).SetBits(make([]big.Word, 0, 48))
	tmpInt  = new(big.Int).SetBits(make([]big.Word, 0, 48))
	slice   = make([]byte, 0, elementByteSize)
)

func Fuzz(data []byte) int {
	if len(data) < elementByteSize {
		replace := make([]byte, elementByteSize)
		copy(replace, data[:])
		data = replace
	}
	startNum := oneNum()
	startUint := oneUint3072()
	startBigInt := mainInt.SetUint64(1)
	for start := 0; start+elementByteSize <= len(data); start += elementByteSize {
		current := data[start : start+elementByteSize]
		currentNum := getNum3072(current)
		currentUint := getUint3072(current)
		currentInt := getBigInt(current)
		if (current[0] & 1) == 1 {
			startNum.Divide(currentNum)
			startUint.Divide(currentUint)
			currentInt.ModInverse(currentInt, prime)
			startBigInt.Mul(startBigInt, currentInt)
			startBigInt.Mod(startBigInt, prime)
		} else {
			startNum.Mul(currentNum)
			startUint.Mul(currentUint)
			startBigInt.Mul(startBigInt, currentInt)
			startBigInt.Mod(startBigInt, prime)
		}
	}

	if !areEqual(&startNum, &startUint) {
		panic(fmt.Sprintf("Expected %v == %v", startNum, startUint))
	}
	if !NumBigEqual(&startNum, startBigInt) {
		panic(fmt.Sprintf("Expected %v == %v", startNum, startBigInt.Bits()))
	}
	return 1
}

func areEqual(num *num3072, uin *uint3072) bool {
	for i := range uin {
		if uin[i] != uint(num.limbs[i]) {
			return false
		}
	}
	return true
}

func NumBigEqual(num *num3072, b *big.Int) bool {
	numBig := new(big.Int).SetBits((*[limbs]big.Word)(unsafe.Pointer(&num.limbs))[:])
	return numBig.Cmp(b) == 0
}

func oneUint3072() uint3072 {
	return uint3072{1}
}
func oneNum() num3072 {
	return num3072{limbs: [48]C.ulong{1}}
}

func getBigInt(data []byte) *big.Int {
	// Reverse the slice because big.Int is Big Endian.
	for i := len(data) - 1; i >= 0; i-- {
		slice = append(slice, data[i])
	}
	res := tmpInt.SetBytes(slice[:])
	slice = slice[:0]
	return res
}

func getNum3072(data []byte) *num3072 {
	var num num3072
	for i := range num.limbs {
		switch bits.UintSize {
		case 64:
			num.limbs[i] = C.ulong(binary.LittleEndian.Uint64(data[i*wordSizeInBytes:]))
		case 32:
			num.limbs[i] = C.ulong(binary.LittleEndian.Uint32(data[i*wordSizeInBytes:]))
		default:
			panic("Only 32/64 bits machines are supported")
		}
	}
	return &num
}

func getUint3072(data []byte) *uint3072 {
	var num uint3072
	for i := range num {
		switch bits.UintSize {
		case 64:
			num[i] = uint(binary.LittleEndian.Uint64(data[i*wordSizeInBytes:]))
		case 32:
			num[i] = uint(binary.LittleEndian.Uint32(data[i*wordSizeInBytes:]))
		default:
			panic("Only 32/64 bits machines are supported")
		}
	}
	return &num
}
