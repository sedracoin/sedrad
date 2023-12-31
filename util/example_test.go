package util_test

import (
	"fmt"
	"github.com/sedracoin/sedrad/util/difficulty"
	"math"
	"math/big"

	"github.com/sedracoin/sedrad/util"
)

func ExampleAmount() {

	a := util.Amount(0)
	fmt.Println("Zero Seep:", a)

	a = util.Amount(1e8)
	fmt.Println("100,000,000 Seep:", a)

	a = util.Amount(1e5)
	fmt.Println("100,000 Seep:", a)
	// Output:
	// Zero Seep: 0 SDR
	// 100,000,000 Seep: 1 SDR
	// 100,000 Seep: 0.001 SDR
}

func ExampleNewAmount() {
	amountOne, err := util.NewAmount(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountOne) //Output 1

	amountFraction, err := util.NewAmount(0.01234567)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountFraction) //Output 2

	amountZero, err := util.NewAmount(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountZero) //Output 3

	amountNaN, err := util.NewAmount(math.NaN())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountNaN) //Output 4

	// Output: 1 SDR
	// 0.01234567 SDR
	// 0 SDR
	// invalid sedra amount
}

func ExampleAmount_unitConversions() {
	amount := util.Amount(44433322211100)

	fmt.Println("Seep to kSDR:", amount.Format(util.AmountKiloSDR))
	fmt.Println("Seep to SDR:", amount)
	fmt.Println("Seep to MilliSDR:", amount.Format(util.AmountMilliSDR))
	fmt.Println("Seep to MicroSDR:", amount.Format(util.AmountMicroSDR))
	fmt.Println("Seep to Seep:", amount.Format(util.AmountSeep))

	// Output:
	// Seep to kSDR: 444.333222111 kSDR
	// Seep to SDR: 444333.222111 SDR
	// Seep to MilliSDR: 444333222.111 mSDR
	// Seep to MicroSDR: 444333222111 μSDR
	// Seep to Seep: 44433322211100 Seep
}

// This example demonstrates how to convert the compact "bits" in a block header
// which represent the target difficulty to a big integer and display it using
// the typical hex notation.
func ExampleCompactToBig() {
	bits := uint32(419465580)
	targetDifficulty := difficulty.CompactToBig(bits)

	// Display it in hex.
	fmt.Printf("%064x\n", targetDifficulty.Bytes())

	// Output:
	// 0000000000000000896c00000000000000000000000000000000000000000000
}

// This example demonstrates how to convert a target difficulty into the compact
// "bits" in a block header which represent that target difficulty .
func ExampleBigToCompact() {
	// Convert the target difficulty from block 300000 in the bitcoin
	// main chain to compact form.
	t := "0000000000000000896c00000000000000000000000000000000000000000000"
	targetDifficulty, success := new(big.Int).SetString(t, 16)
	if !success {
		fmt.Println("invalid target difficulty")
		return
	}
	bits := difficulty.BigToCompact(targetDifficulty)

	fmt.Println(bits)

	// Output:
	// 419465580
}
