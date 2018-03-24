// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// primes2.go
package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"getcommandline"
	"os"
	"strconv"
)

const LastCompiled = "8 Mar 2018"

func main() {
	/*
	   	This module tests my thoughts on prime factoring, derived from rpn.go
	   	REVISION HISTORY
	   	----------------
	   	24 Feb 17 -- Primes.go is derived from rpn.go
	   	17 Feb 18 -- Made prime divisors a slice instead of an array.  Addressing syntax is the same.
	   	25 Feb 18 -- 736711 is trouble.  Will print out a factor.  And use uint.
	   	27 Feb 18 -- Fixing a bug about even numbers and correct number of factors.
	     7 Mar 18 -- Fixed another bug about 2 not being prime.  It is prime.
	   	 8 Mar 18 -- Added the PrimesMap and PrimesSlice
	*/

	const PrimeMapFilename = "primemap.gob"
	const PrimeSliceFilename = "primeslice.gob"
	var INBUF string
	var PrimeMap map[uint]bool
	var PrimeNumbersSlice []uint

	fmt.Println(" Prime Factoring Program.  Last compiled ", LastCompiled)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
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
	} // if command tail exists

	U, err := strconv.ParseUint(INBUF, 10, 64)
	if err != nil {
		fmt.Println(" Conversion to number failed.  Exiting")
		os.Exit(1)
	}

	fac, primeflag := IsPrimeInt64(U)
	if primeflag {
		fmt.Println(U, " is prime so it has no factors.")
		fmt.Println()
		//	os.Exit(0)  Nedd to fall thru to other routines.
	}

	N := int(U)
	PrimeFactors := PrimeFactorization(N)

	fmt.Print(" Prime factors for ", N, " are : ")
	for _, pf := range PrimeFactors {
		fmt.Print(pf, "  ")
	}
	fmt.Println()
	fmt.Println()

	if fac == 0 {
		fmt.Println(U, " is prime so there are no other factors.")
	} else {
		fmt.Println(U, " has ", fac, " as its first factor")
	}
	fmt.Println()
	fmt.Println()

	PrimeUfactors := PrimeFactorMemoized(U)
	fmt.Print(" Memoized Prime factors for ", U, " are : ")
	for _, pf := range PrimeUfactors {
		fmt.Print(pf, "  ")
	}
	fmt.Println()
	fmt.Println()

	// Now for the use of PrimeMap and PrimeNumbersSlice
	_, err1 := os.Stat(PrimeMapFilename)
	_, err2 := os.Stat(PrimeSliceFilename)
	if err1 != nil || err2 != nil {
		fmt.Println(" Cannot find the Prime files.  Ignoring.")
		fmt.Println()
		os.Exit(1)
	}

	var u uint
	u = uint(U)
	PrimeNumbersSlice = make([]uint, 1e6)
	theslicefile, err := os.Open(PrimeSliceFilename)
	check(err)
	defer theslicefile.Close()
	decoder := gob.NewDecoder(theslicefile)
	err = decoder.Decode(&PrimeNumbersSlice) // this did not work without &
	check(err)
	theslicefile.Close()

	PrimeMap = make(map[uint]bool)
	themapfile, err := os.Open(PrimeMapFilename)
	check(err)
	defer themapfile.Close()
	decoder = gob.NewDecoder(themapfile)
	err = decoder.Decode(&PrimeMap) // this did not work without &
	check(err)
	themapfile.Close()

	fmt.Println(" Now using the PrimeMap and PrimeNumbersSlice algorithm")
	fmt.Println(" Len of slice is ", len(PrimeNumbersSlice))
	fmt.Println(" Len of PrimeMap is ", len(PrimeMap))
	/*
		fmt.Println(" PrimeMap is:")
		fmt.Print(PrimeMap)
		fmt.Println()
		fmt.Println()

		fmt.Println(" Pausing ")
		ans := ""
		fmt.Scanln(&ans)
		fmt.Println(" PrimeNumbersSlice is:")
		fmt.Print(PrimeNumbersSlice[0:10])
		fmt.Println()
		fmt.Println()

		fmt.Println(" Pausing ")
		ans := ""
		fmt.Scanln(&ans)
	*/

	PrimeFlag := PrimeMap[u] // Note that if a key is not in a map, that returns the zero value for that type.

	if PrimeFlag {
		fmt.Println(u, " is prime.")
	} else {
		fmt.Println(u, " is NOT prime.")
	}
	//--------------------------------
	PrimeSliceFactorization := func(U uint) []uint { // Hey, a closure anonymous function
		PrimeFacSlice := make([]uint, 0, 100)
		if U == 0 {
			return nil
		}
		if PrimeFlag {
			PrimeFacSlice = append(PrimeFacSlice, U)
			return PrimeFacSlice
		}

		n := U
		for i := 0; i < len(PrimeNumbersSlice); i++ {
			//			fmt.Println(" in closure.  i=", i, " and n=", n)
			for n > 0 && n%PrimeNumbersSlice[i] == 0 {
				PrimeFacSlice = append(PrimeFacSlice, PrimeNumbersSlice[i])
				n = n / PrimeNumbersSlice[i]
			}
			if PrimeMap[n] {
				PrimeFacSlice = append(PrimeFacSlice, n)
				break
			}
			if n == 1 {
				break
			}
		}
		//	fmt.Println(" Length of the prime factors slice using the table lookup is", len(PrimeFacSlice))
		return PrimeFacSlice
	} // end anonymous function closure called PrimeSlaceFactorization
	//--------------------------------
	PrimeFacSlice := PrimeSliceFactorization(u)

	//	fmt.Println(" Length of the prime factors slice using the table lookup is", len(PrimeFacSlice))
	fmt.Print(" Prime factors for ", u, " using PrimeMap and Slice are : ")
	for _, pf := range PrimeFacSlice {
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
		if n == 1 {
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
	} else if Uint == 2 || Uint == 3 {
		return 0, true
	} else if Uint%2 == 0 {
		return 2, false
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
	} else if Uint == 2 || Uint == 3 {
		return 0, true
	} else if Uint%2 == 0 {
		return 2, false
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

// ---------------------------------------------- check --------------------------
func check(e error) {
	if e != nil {
		panic(e)
	}
}
