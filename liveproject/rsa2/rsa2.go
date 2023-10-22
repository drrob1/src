package main

import (
	"fmt"
	"math"
)

// This is the 2nd step in the RSA live project.  Here, the goals are fast exponentiation and fast modulo exponentiation
// It's now called exponentiation by squaring (I first read about it in Programming in Modula-2, and is covered on multiple websites
// referenced by Stephens, on Math Less Traveled and Wikipedia.
// Fast modular exponentiation is covered by Khan Academy.
//  2
// A mod C = (A * A) mod C = ((A mod C) * (A mod C)) mod C

func fastExp(num, pow int) int { // pow can't be negative here, or else it will panic.
	Z := 1
	if pow < 0 {
		s := fmt.Sprintf("fastExp pow cannot be negative.  It is %d", pow)
		panic(s)
	}
	for pow > 0 {
		if pow%2 == 1 { // ie, if pow is odd
			Z *= num // Z = Z * R
		}
		num *= num // R = R squared
		pow /= 2   // I = half I
	}
	return Z
}

func fastExpMod(num, pow, mod int) int { // pow can't be negative, or else it will panic.
	Z := 1
	if pow < 0 || mod < 0 {
		s := fmt.Sprintf("fastExpMod pow or mod cannot be negative.  pow = %d, mod = %d", pow, mod)
		panic(s)
	}
	for pow > 0 {
		if pow%2 == 1 { // ie, if pow is odd
			Z = (Z * num) % mod // Z = (Z * R) % mod.  Can't use *= operator here as the mod operator won't be correctly applied.
		}
		num = (num * num) % mod // R = (R squared) % mod.  Can't use *= operator here as the mod operator won't be correctly applied.
		pow /= 2                // I = half I
	}
	return Z //% mod Author's solution does not have the mod operator here.  So I guess it's not needed?
}

func main() { // I know these work because of the unit testing I've done, code is in rsa2_test.go
	var a, b, mod int
	for {
		fmt.Print(" Enter a b mod: ")
		//n, err := fmt.Scanf("%d %d %d\n", &a, &b, &mod)
		n, err := fmt.Scanln(&a, &b, &mod)
		if n == 0 || err != nil {
			fmt.Printf(" n=%d, a=%d, b=%d, mod=%d, err = %s\n", n, a, b, mod, err)
			break
		}
		fastResult := fastExp(a, b)
		powResult := math.Pow(float64(a), float64(b))
		intPowResult := int(powResult)

		fastResultMod := fastExpMod(a, b, mod)
		powResultModF := math.Pow(float64(a), float64(b))
		powResultModI := int(powResultModF) % mod
		fmt.Printf(" a=%3d, b=%3d, mod=%4d, a^b mod=%4d, a^b = %d; matched mod=%t, matched exp=%t\n", a, b, mod, fastResultMod, fastResult,
			fastResultMod == powResultModI, fastResult == intPowResult)
	}
}

// ---------------------------------------- pwrI -------------------------------------------------
// Power of I.
// This is a power function with a real base and integer exponent, using the optimized algorithm as discussed in PIM-2, V 2.
/*
func pwrI(R float64, I int) float64 {
	Z := 1.0
	NEGFLAG := false
	if I < 0 {
		NEGFLAG = true
		I = -I
	}
	for I > 0 {
		if I%2 == 1 {
			Z *= R // Z = Z * R
		}
		R *= R // R = R squared
		I /= 2 // I = half I
	}
	if NEGFLAG {
		Z = 1 / Z
	}
	return Z
} // pwrI

*/
