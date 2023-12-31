package muhash

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

func Test_mul(t *testing.T) {
	t.Parallel()
	type Test struct {
		a            uint
		b            uint
		expectedLow  uint
		expectedHigh uint
	}
	tests := []Test{
		{
			a:            ^uint(0),
			b:            ^uint(0),
			expectedLow:  1,
			expectedHigh: 18446744073709551614,
		},
		{
			a:            ^uint(0) - 100,
			b:            ^uint(0) - 30,
			expectedLow:  3131,
			expectedHigh: 18446744073709551484,
		},
	}
	for i, test := range tests {
		var low, high uint
		mul(&low, &high, test.a, test.b)
		if low != test.expectedLow {
			t.Fatalf("Test: %d. Expected: %d, found: %d", i, test.expectedLow, low)
		}
		if high != test.expectedHigh {
			t.Fatalf("Test: %d. Expected: %d, found: %d", i, test.expectedHigh, high)
		}
	}
}

func Test_mulnadd3(t *testing.T) {
	t.Parallel()
	type Test struct {
		c0         uint
		c1         uint
		c2         uint
		d0         uint
		d1         uint
		d2         uint
		n          uint
		expectedC0 uint
		expectedC1 uint
		expectedC2 uint
	}
	tests := []Test{
		{
			c0:         ^uint(0) - 99,
			c1:         ^uint(0) - 75,
			c2:         ^uint(0) - 100,
			d0:         ^uint(0) - 30,
			d1:         ^uint(0) - 3452,
			d2:         ^uint(0) - 321,
			n:          ^uint(0) - 543,
			expectedC0: 16764,
			expectedC1: 1877782,
			expectedC2: 171173,
		},
		{
			c0:         0,
			c1:         ^uint(0) - 32432432,
			c2:         ^uint(0) - 534532431432423,
			d0:         ^uint(0) - 534543534534,
			d1:         1,
			d2:         ^uint(0) - 3242353456341,
			n:          ^uint(0) - 546546456543,
			expectedC0: 11788773271371804448,
			expectedC1: 18446742446040687397,
			expectedC2: 10322986003028211010,
		},
	}
	for i, test := range tests {
		mulnadd3(&test.c0, &test.c1, &test.c2, test.d0, test.d1, test.d2, test.n)
		if test.c0 != test.expectedC0 {
			t.Fatalf("Test: %d. Expected c0: %d, found: %d", i, test.expectedC0, test.c0)
		}
		if test.c1 != test.expectedC1 {
			t.Fatalf("Test: %d. Expected c1: %d, found: %d", i, test.expectedC1, test.c1)
		}
		if test.c2 != test.expectedC2 {
			t.Fatalf("Test: %d. Expected c2: %d, found: %d", i, test.expectedC2, test.c2)
		}
	}
}

func Test_muln2(t *testing.T) {
	t.Parallel()
	type Test struct {
		low          uint
		high         uint
		n            uint
		expectedLow  uint
		expectedHigh uint
	}
	tests := []Test{
		{
			low:          ^uint(0) - 99,
			high:         ^uint(0) - 75,
			n:            ^uint(0) - 543,
			expectedLow:  54400,
			expectedHigh: 40700,
		},
		{
			low:          0,
			high:         ^uint(0) - 32432432,
			n:            ^uint(0) - 546546456543,
			expectedLow:  0,
			expectedHigh: 17725831333250691552,
		},
	}
	for i, test := range tests {
		muln2(&test.low, &test.high, test.n)
		if test.low != test.expectedLow {
			t.Fatalf("Test: %d. Expected low: %d, found: %d", i, test.expectedLow, test.low)
		}
		if test.high != test.expectedHigh {
			t.Fatalf("Test: %d. Expected high: %d, found: %d", i, test.expectedHigh, test.high)
		}
	}
}

func Test_muladd3(t *testing.T) {
	t.Parallel()
	type Test struct {
		low           uint
		high          uint
		carry         uint
		a             uint
		b             uint
		expectedLow   uint
		expectedHigh  uint
		expectedCarry uint
	}
	tests := []Test{
		{
			low:           ^uint(0) - 99,
			high:          ^uint(0) - 75,
			carry:         ^uint(0) - 100,
			a:             ^uint(0) - 30,
			b:             ^uint(0) - 3452,
			expectedLow:   106943,
			expectedHigh:  18446744073709548057,
			expectedCarry: 18446744073709551516,
		},
		{
			low:           0,
			high:          ^uint(0) - 32432432,
			carry:         ^uint(0) - 534532431432423,
			a:             ^uint(0) - 534543534534,
			b:             1,
			expectedLow:   18446743539166017081,
			expectedHigh:  18446744073677119183,
			expectedCarry: 18446209541278119192,
		},
	}
	for i, test := range tests {
		muladd3(&test.low, &test.high, &test.carry, test.a, test.b)
		if test.low != test.expectedLow {
			t.Fatalf("Test: %d. %#v: %d, found: %d", i, test.expectedLow, test.expectedLow, test.low)
		}
		if test.high != test.expectedHigh {
			t.Fatalf("Test: %d. %#v: %d, found: %d", i, test.expectedHigh, test.expectedHigh, test.high)
		}
		if test.carry != test.expectedCarry {
			t.Fatalf("Test: %d. %#v: %d, found: %d", i, test.expectedCarry, test.expectedCarry, test.carry)
		}
	}
}

func Test_muldbladd3(t *testing.T) {
	t.Parallel()
	type Test struct {
		low           uint
		high          uint
		carry         uint
		a             uint
		b             uint
		expectedLow   uint
		expectedHigh  uint
		expectedCarry uint
	}
	tests := []Test{
		{
			low:           ^uint(0) - 99,
			high:          ^uint(0) - 75,
			carry:         ^uint(0) - 100,
			a:             ^uint(0) - 30,
			b:             ^uint(0) - 3452,
			expectedLow:   213986,
			expectedHigh:  18446744073709544573,
			expectedCarry: 18446744073709551517,
		},
		{
			low:           0,
			high:          ^uint(0) - 32432432,
			carry:         ^uint(0) - 534532431432423,
			a:             ^uint(0) - 534543534534,
			b:             1,
			expectedLow:   18446743004622482546,
			expectedHigh:  18446744073677119184,
			expectedCarry: 18446209541278119192,
		},
		{
			low:           0,
			high:          0,
			carry:         0,
			a:             1,
			b:             1,
			expectedLow:   2,
			expectedHigh:  0,
			expectedCarry: 0,
		},
	}
	for i, test := range tests {
		muldbladd3(&test.low, &test.high, &test.carry, test.a, test.b)
		if test.low != test.expectedLow {
			t.Fatalf("Test: %d. %#v: %d, found: %d", i, test.expectedLow, test.expectedLow, test.low)
		}
		if test.high != test.expectedHigh {
			t.Fatalf("Test: %d. %#v: %d, found: %d", i, test.expectedHigh, test.expectedHigh, test.high)
		}
		if test.carry != test.expectedCarry {
			t.Fatalf("Test: %d. %#v: %d, found: %d", i, test.expectedCarry, test.expectedCarry, test.carry)
		}
	}
}

func TestUint3072_GetInverse(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(0))
	var element uint3072
	for i := 0; i < 5; i++ {
		for i := range element {
			element[i] = uint(r.Uint64())
		}
		inv := element.GetInverse()
		again := inv.GetInverse()

		if again != element {
			t.Fatalf("Expected double inverting to be equal, found: %v != %v", again, element)
		}
	}
}

func uint3072equalToUint(a *uint3072, b uint) bool {
	if a[0] != b {
		return false
	}
	for j := 1; j < len(a); j++ {
		if a[j] != 0 {
			return false
		}
	}
	return true
}

func TestUint3072_DivOverflow(t *testing.T) {
	tests := make([]byte, primeDiff)
	var max uint3072
	for i := range max {
		max[i] = maxUint
	}
	regularOne := one()
	var wg sync.WaitGroup
	step := primeDiff / runtime.NumCPU()
	for c := 0; c < runtime.NumCPU(); c++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			start := c * step
			end := start + step
			if end > (primeDiff - step) {
				end = primeDiff
			}
			for i := end; i > start; i-- {
				expected := uint(primeDiff - i)
				overflown := max
				overflown[0] = maxUint - uint(i) + 1
				overflownCopy := overflown
				overflownCopy.Divide(&regularOne)
				tests[expected]++
				if !uint3072equalToUint(&overflownCopy, expected) {
					t.Errorf("Expected %v to be %d", overflownCopy, expected)
					return
				}
				// Zero doesn't have a modular inverse
				if i != primeDiff {
					lhs := overflown
					rhs := overflown
					lhs.Divide(&rhs)
					if !uint3072equalToUint(&lhs, 1) {
						t.Errorf("Expected %v to be %d", overflownCopy, 1)
						return
					}
				}
			}
		}(c)
	}
	wg.Wait()
	for i, n := range tests {
		if n != 1 {
			t.Fatalf("Expected all the integers 0..%d to be checked once, but %d was checked %d times", primeDiff, i, n)
		}
	}
}

func TestUint3072_MulMax(t *testing.T) {
	t.Parallel()
	var max uint3072
	for i := range max {
		max[i] = maxUint
	}
	max[0] -= primeDiff
	copyMax := max
	max.Mul(&copyMax)
	if !uint3072equalToUint(&max, 1) {
		t.Fatalf("(p-1)*(p-1) mod p should equal 1, instead got: %v", max)
	}
}

func TestUint3072MulDiv(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	var list [loopsN]uint3072
	start := one()
	for i := 0; i < loopsN; i++ {
		for n := range list[i] {
			list[i][n] = uint(r.Uint64())
		}
		start.Mul(&list[i])
	}
	if start == one() {
		t.Errorf("start is 1 even though it shouldn't be: start '%x', one: %x\n", start, one())
	}

	for i := 0; i < loopsN; i++ {
		start.Divide(&list[i])
	}
	if start != one() {
		t.Errorf("start should be 1 but it isn't: start: '%x', one: '%x'\n", start, one())
	}
}
