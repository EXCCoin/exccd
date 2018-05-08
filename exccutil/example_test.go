// Copyright (c) 2018 The ExchangeCoin team
package exccutil_test

import (
	"fmt"
	"math"

	"github.com/EXCCoin/exccd/exccutil"
)

func ExampleAmount() {

	a := exccutil.Amount(0)
	fmt.Println("Zero Atom:", a)

	a = exccutil.Amount(1e8)
	fmt.Println("100,000,000 Atoms:", a)

	a = exccutil.Amount(1e5)
	fmt.Println("100,000 Atoms:", a)
	// Output:
	// Zero Atom: 0 DCR
	// 100,000,000 Atoms: 1 DCR
	// 100,000 Atoms: 0.001 DCR
}

func ExampleNewAmount() {
	amountOne, err := exccutil.NewAmount(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountOne) //Output 1

	amountFraction, err := exccutil.NewAmount(0.01234567)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountFraction) //Output 2

	amountZero, err := exccutil.NewAmount(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountZero) //Output 3

	amountNaN, err := exccutil.NewAmount(math.NaN())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(amountNaN) //Output 4

	// Output: 1 DCR
	// 0.01234567 DCR
	// 0 DCR
	// invalid coin amount
}

func ExampleAmount_unitConversions() {
	amount := exccutil.Amount(44433322211100)

	fmt.Println("Atom to kCoin:", amount.Format(exccutil.AmountKiloCoin))
	fmt.Println("Atom to Coin:", amount)
	fmt.Println("Atom to MilliCoin:", amount.Format(exccutil.AmountMilliCoin))
	fmt.Println("Atom to MicroCoin:", amount.Format(exccutil.AmountMicroCoin))
	fmt.Println("Atom to Atom:", amount.Format(exccutil.AmountAtom))

	// Output:
	// Atom to kCoin: 444.333222111 kDCR
	// Atom to Coin: 444333.222111 DCR
	// Atom to MilliCoin: 444333222.111 mDCR
	// Atom to MicroCoin: 444333222111 μDCR
	// Atom to Atom: 44433322211100 Atom
}
