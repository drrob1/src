// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// primes.go
package main

import (
	"bufio"
	"fmt"
	"getcommandline"
	"makesubst"
	"os"
	"strconv"
)

const LastCompiled = "27 Feb 2018"

func main() {
	/*
	   This module tests my thoughts on prime factoring, derived from rpn.go
	   REVISION HISTORY
	   ----------------
	   24 Feb 17 -- Primes.go is derived from rpn.go
	   17 Feb 18 -- Made prime divisors a slice instead of an array.  Addressing syntax is the same.
	   25 Feb 18 -- 736711 is trouble.  Will print out a factor.  And use uint.
	   27 Feb 18 -- Fixing a bug about even numbers and correct number of factors.
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
		fmt.Println(U, " is prime so it has no factors.")
		fmt.Println()
		os.Exit(0)
	}

	N := int(U)
	PrimeFactors := PrimeFactorization(N)

	fmt.Print(" Prime factors for ", N, " are : ")
	for _, pf := range PrimeFactors {
		fmt.Print(pf, "  ")
	}
	fmt.Println()
	fmt.Println()

	fmt.Println(U, " is NOT prime, and ", fac, " is its first factor")
	fmt.Println()
	fmt.Println()

	PrimeUfactors := PrimeFactorMemoized(U)
	fmt.Print(" Memoized Prime factors for ", U, " are : ")
	for _, pf := range PrimeUfactors {
		fmt.Print(pf, "  ")
	}

	fmt.Println()
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

// --------------------------------------- PrimeFactorMemoized -------------------
func PrimeFactorMemoized(U uint64) []uint64 {

	if U == 0 {
		return nil
	}

	var val uint64 = 2

	PrimeUfactors := make([]uint64, 0, 20)

	//fmt.Print("u, fac, val, primeflag : ")
	for u := U; u > 1; {
		fac, facflag := NextPrimeFac(u, val)
		//	fmt.Print(u, " ", fac, " ", val, " ", primeflag, ", ")
		if facflag {
			PrimeUfactors = append(PrimeUfactors, fac)
			u = u / fac
			val = fac
		} else { // no more factors found
			PrimeUfactors = append(PrimeUfactors, u)
			break
		}
	}
	//fmt.Println()
	return PrimeUfactors
}

// ------------------------------------------------- IsPrimeInt64 -----------------
func IsPrimeInt64(n uint64) (uint64, bool) {

	var t uint64 = 3

	Uint := n

	if Uint == 0 || Uint == 1 {
		return Uint, false
	} else if Uint%2 == 0 {
		return 2, false
	} else if Uint == 2 || Uint == 3 {
		return 0, true
	}

	//	sqrt := math.Sqrt(float64(Uint))
	//	UintSqrt := uint64(sqrt)
	UintSqrt := usqrt(n)

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

	var t uint64 = 3

	Uint := uint64(n)

	if Uint == 0 || Uint == 1 {
		return uint(Uint), false
	} else if Uint%2 == 0 {
		return 2, false
	} else if Uint == 2 || Uint == 3 {
		return 0, true
	}

	//	sqrt := math.Sqrt(float64(Uint))
	//	UintSqrt := uint(sqrt)
	UintSqrt := usqrt(uint64(n))

	for t <= UintSqrt {
		if Uint%t == 0 {
			return uint(t), false
		}
		t += 2
	}
	return 0, true
} // IsPrime

// ------------------------------------------------- NextPrimeFac -----------------
func NextPrimeFac(n, startfac uint64) (uint64, bool) { // note that this is the reverse of IsPrime

	var t = startfac

	UintSqrt := usqrt(n)

	for t <= UintSqrt {
		if n%t == 0 {
			return t, true
		}
		if t == 2 {
			t = 3
		} else {
			t += 2
		}
	}
	return 0, false
} // NextPrimeFac

//----------------------------------------------- usqrt ---------------------------
func usqrt(u uint64) uint64 {

	sqrt := u / 2

	for i := 0; i < 30; i++ {
		guess := u / sqrt
		sqrt = (guess + sqrt) / 2
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}
	return sqrt
} // usqrt
