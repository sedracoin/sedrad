package muhash

// #include "muhash.h"
import "C"
import (
	"math/big"
	"math/bits"
	"unsafe"
)

type word = C.limb_t

const (
	wordSizeInBytes = int(unsafe.Sizeof(word(0)))
	wordSize        = wordSizeInBytes * 8
	elementWordSize = elementByteSize / wordSizeInBytes
	maxLimb         = ^word(0)
)

func init() {
	// Some sanity asserts
	assert(C.LIMBS == elementWordSize)
	assert(bits.UintSize == unsafe.Sizeof(word(0))*8)
	assert(unsafe.Sizeof(num3072{}.limbs) == elementByteSize)

	assert(unsafe.Sizeof(uint(0)) == unsafe.Sizeof(num3072{}.limbs[0]))
	assert(unsafe.Alignof(uint(0)) == unsafe.Alignof(num3072{}.limbs[0]))

	assert(unsafe.Sizeof(uint(0)) == unsafe.Sizeof(big.Word(0)))
	assert(unsafe.Alignof(uint(0)) == unsafe.Alignof(big.Word(0)))

	assert(unsafe.Sizeof([elementWordSize]big.Word{}) == unsafe.Sizeof(num3072{}.limbs))
	assert(unsafe.Alignof([elementWordSize]big.Word{}) == unsafe.Alignof(num3072{}.limbs))
}

func oneNum3072() num3072 {
	return num3072{limbs: [C.LIMBS]word{1}}
}

type num3072 C.Num3072

func (lhs *num3072) SetToOne() {
	*lhs = num3072{limbs: [C.LIMBS]word{1}}
}

func (lhs *num3072) Mul(rhs *num3072) {
	C.Num3072_Multiply((*C.Num3072)(lhs), (*C.Num3072)(rhs))
}

func (lhs *num3072) Divide(rhs *num3072) {
	if lhs.IsOverflow() {
		lhs.FullReduce()
	}
	inv := rhs.GetInverse()
	lhs.Mul(inv)
	if lhs.IsOverflow() {
		lhs.FullReduce()
	}
}

func (lhs *num3072) IsOverflow() bool {
	if lhs.limbs[0] <= (maxLimb - primeDiff) {
		return false
	}
	for i := 1; i < len(lhs.limbs); i++ {
		if lhs.limbs[i] != maxLimb {
			return false
		}
	}
	return true
}

func (lhs *num3072) FullReduce() {
	C.Num3072_FullReduce((*C.Num3072)(lhs))
}

func (lhs *num3072) GetInverse() *num3072 {
	if lhs.IsOverflow() {
		lhs.FullReduce()
	}
	inv := *lhs
	words := (*[elementWordSize]big.Word)(unsafe.Pointer(&inv.limbs))
	var bigInt big.Int
	bigInt.SetBits(words[:])
	bigInt.ModInverse(&bigInt, prime)
	for i := len(bigInt.Bits()); i < len(inv.limbs); i++ {
		inv.limbs[i] = 0
	}
	return &inv
}
