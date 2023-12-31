package muhash

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
)

func TestNum3072_GetInverse(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(0))
	var element num3072
	for i := 0; i < 5; i++ {
		for i := range element.limbs {
			element.limbs[i] = word(r.Uint64())
		}
		inv := element.GetInverse()
		again := inv.GetInverse()

		if *again != element {
			t.Fatalf("Expected double inverting to be equal, found: %v != %v", again, element)
		}
	}
}

func num3072equalToWord(a *num3072, b word) bool {
	if a.limbs[0] != b {
		return false
	}
	for j := 1; j < len(a.limbs); j++ {
		if a.limbs[j] != 0 {
			return false
		}
	}
	return true
}

func TestNum3072_IsOverflow(t *testing.T) {
	var n num3072
	if n.IsOverflow() {
		t.Fatal("zeroed Num3072 isn't overflown")
	}
	n.limbs[0] = maxLimb
	if n.IsOverflow() {
		t.Fatalf("a %d num3072 isn't overflown", maxLimb)
	}
	for i := range n.limbs {
		n.limbs[i] = maxLimb
	}
	if !n.IsOverflow() {
		t.Fatal("maxed out num3072 is defenitely an overflow")
	}
	n.limbs[0] -= primeDiff - 1
	if !n.IsOverflow() {
		t.Fatal("The prime itself is considered an overflow(=0)")
	}
}

func TestNum3072_DivOverflow(t *testing.T) {
	tests := make([]byte, primeDiff)
	var max num3072
	for i := range max.limbs {
		max.limbs[i] = maxLimb
	}
	regularOne := oneNum3072()
	var wg sync.WaitGroup
	step := word(primeDiff / runtime.NumCPU())
	for c := word(0); c < word(runtime.NumCPU()); c++ {
		wg.Add(1)
		go func(c word) {
			defer wg.Done()
			start := c * step
			end := start + step
			if end > (primeDiff - step) {
				end = primeDiff
			}
			for i := end; i > start; i-- {
				expected := word(primeDiff - i)
				overflown := max
				overflown.limbs[0] = maxLimb - i + 1
				overflownCopy := overflown
				overflownCopy.Divide(&regularOne)
				tests[expected]++
				if !num3072equalToWord(&overflownCopy, expected) {
					t.Errorf("Expected %v to be %d", overflownCopy, expected)
					return
				}
				// Zero doesn't have a modular inverse
				if i != primeDiff {
					lhs := overflown
					rhs := overflown
					lhs.Divide(&rhs)
					if !num3072equalToWord(&lhs, 1) {
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

func TestNum3072_MulMax(t *testing.T) {
	t.Parallel()
	var max num3072
	for i := range max.limbs {
		max.limbs[i] = word(maxUint)
	}
	max.limbs[0] -= primeDiff
	copyMax := max
	max.Mul(&copyMax)
	if !num3072equalToWord(&max, 1) {
		t.Fatalf("(p-1)*(p-1) mod p should equal 1, instead got: %v", max)
	}
}

func TestNum3072MulDiv(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	var list [loopsN]num3072
	start := oneNum3072()
	for i := 0; i < loopsN; i++ {
		for n := range list[i].limbs {
			list[i].limbs[n] = word(r.Uint64())
		}
		start.Mul(&list[i])
	}
	if start == oneNum3072() {
		t.Errorf("start is 1 even though it shouldn't be: start '%x', one: %x\n", start, one())
	}

	for i := 0; i < loopsN; i++ {
		start.Divide(&list[i])
	}
	if start != oneNum3072() {
		t.Errorf("start should be 1 but it isn't: start: '%x', one: '%x'\n", start, one())
	}
}

// This specifically tests the zeroing loop at the end of num3072.GetInverse.
func TestNum3072_GetInverse_EdgeCase(t *testing.T) {
	orig := num3072{limbs: [limbs]word{7122228832992001076, 984226626229791276, 7630161757215403889, 6284986028532537849, 8045609952094061025, 11960578682873843289, 13746438324198032094, 13918942278011779234, 17733507388171786846, 10563242470999117317, 17037155475664456442, 17937456968131788544, 12599342294785769540, 13386260146859547870, 2817582499516127913, 652557987984108933, 9669847560665129471, 17711760030167214508, 5376140856964249866, 18051557786492143716, 2482926987284881227, 8605482545261324676, 7878786448874819977, 1266815984192471985, 2678516262590404672, 14004775981272003760, 10357003870690124643, 2730710396948079405, 4635754375072562978, 13656184258619915136, 803512205739688286, 11844116904145642840, 5760653310472302601, 15069027324939031326, 14913021043324743434, 17567013163360751106, 6302557725767759643, 17458497366820989801, 3410551217786514778, 14182717432968305815, 12471950523812677269, 2294197765573979691, 3220941588656114052, 605606616684921311, 1440136155000853957, 16361481774333736133, 11385241783616172231, 13968855456762740410}}
	inverse := orig.GetInverse()
	if *inverse.GetInverse() != orig {
		t.Fatalf("Double inverting resulted in different varaible than the original: %v", orig)
	}
}
