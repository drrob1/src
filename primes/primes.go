// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// primes.go
package main

import (
	"bufio"
	"fmt"
	"getcommandline"
	"makesubst"
	"math"
	"os"
	"strconv"
)

const LastCompiled = "25 Feb 2018"

func main() {
	/*
	   This module tests my thoughts on prime factoring, derived from rpn.go
	   REVISION HISTORY
	   ----------------
	   24 Feb 17 -- Primes.go is derived from rpn.go
	   17 Feb 18 -- Made prime divisors a slice instead of an array.  Addressing syntax is the same.
	   25 Feb 18 -- 736711 is trouble.  Will print out a factor.  And use uint.
	*/

	var INBUF string

	fmt.Println(" Prime Factoring Program.  Last compiled ", LastCompiled)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
		INBUF = makesubst.MakeSubst(INBUF)
	} else {
		fmt.Print(" Enter number to factor : ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(INBUF) == 0 {
			os.Exit(0)
		}
		INBUF = makesubst.MakeSubst(INBUF)
	} // if command tail exists

	//	N, err := strconv.Atoi(INBUF)
	U, err := strconv.ParseUint(INBUF, 10, 64)
	if err != nil {
		fmt.Println(" Conversion to number failed.  Exiting")
		os.Exit(1)
	}

	fac, primeflag := IsPrimeInt64(U)
	if primeflag {
		fmt.Println(U, " is prime.")
		os.Exit(0)
	} else {
		fmt.Print(U, " is NOT prime.")
		if fac != 0 {
			fmt.Println("  ", fac, " is its first factor")
		}
	}
	fmt.Println()

	N := int(U)
	PrimeFactors := PrimeFactorization(N)

	fmt.Print(" Prime factors for ", N, " are : ")
	for _, pf := range PrimeFactors {
		fmt.Print(pf, "  ")
	}

	fmt.Println()
} // end of main

// -------------------------------------------- PrimeFactorization ------------------------------

func PrimeFactorization(N int) []int {

	var PD = []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47} // Prime divisors array

	if N == 0 {
		return nil
	}

	PrimeFactors := make([]int, 0, 10)

	_, flag := IsPrimeInt(uint(N))
	if flag {
		PrimeFactors = append(PrimeFactors, N)
		return PrimeFactors
	}

	n := N
	for i := 0; i < len(PD); i++ { // outer loop to sequentially test the prime divisors
		for n > 0 && n%PD[i] == 0 {
			PrimeFactors = append(PrimeFactors, PD[i])
			n = n / PD[i]
		}
		_, primeflag := IsPrimeInt(uint(n))
		if primeflag {
			PrimeFactors = append(PrimeFactors, n)
			break
		}
	}
	return PrimeFactors

} // PrimeFactorization

// ------------------------------------------------- IsPrimeInt64 -----------------
func IsPrimeInt64(n uint64) (uint64, bool) {

	var t uint64 = 3

	//	Uint := uint64(n)
	Uint := n

	if Uint == 0 || Uint == 1 || Uint%2 == 0 {
		return 0, false
	} else if Uint == 2 || Uint == 3 {
		return Uint, true
	}

	sqrt := math.Sqrt(float64(Uint))
	UintSqrt := uint64(sqrt)

	for t <= UintSqrt {
		if Uint%t == 0 {
			return t, false
		}
		t += 2
	}
	return 0, true
} // IsPrimeInt64

// ------------------------------------------------- IsPrimeInt -----------------
func IsPrimeInt(n uint) (uint, bool) {

	var t uint = 3

	//	Uint := uint(n)
	Uint := n

	if Uint == 0 || Uint == 1 || Uint%2 == 0 {
		return 0, false
	} else if Uint == 2 || Uint == 3 {
		return Uint, true
	}

	sqrt := math.Sqrt(float64(Uint))
	UintSqrt := uint(sqrt)

	for t <= UintSqrt {
		if Uint%t == 0 {
			return t, false
		}
		t += 2
	}
	return 0, true
} // IsPrime
