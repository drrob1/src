// makeprimestable.go
package main

import (
	"encoding/gob"
	"fmt"
	"os"
	hpcalc "src/hpcalc2"
	"strconv"
	"time"
)

const LastAlteredDate = "8 Mar 2018"

/*
REVISION HISTORY
----------------
 7 Mar 18 -- First version of this routine make a slice of all 64 bit primes and write it out.
               This is to be read in by the prime checking and prime factoring routines.
	       A map of booleans works for prime checking, and a slice of the actual numbers
	       to be used for prime factoring.
	       Both are to be saved as gob files.
               And my first use of a ticker
 8 Mar 18 -- Tweaking output file process and when to write these files.
 3 Sep 21 -- Add a module so it would compile on Go 1.16
*/

const PrimeMapFilename = "primemap.gob"
const PrimeSliceFilename = "primeslice.gob"
const MaxPrime32 uint = 4_294_967_291

var StopVal uint = 1e7

func main() {
	fmt.Println(" MakePrimesTable written in Go.  Code was last altered ", LastAlteredDate)
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fmt.Println()
	//	StopVal = 100

	_, err1 := os.Stat(PrimeMapFilename)
	_, err2 := os.Stat(PrimeSliceFilename)
	if err1 == nil || err2 == nil { // file exists.  Don't need to make it here.
		fmt.Println(PrimeMapFilename, "or", PrimeSliceFilename, " already exist.  Exiting")
		os.Exit(0)
	}

	// make the boolean map and prime numbers slice
	PrimeMap := make(map[uint]bool)
	PrimeNumbersSlice := make([]uint, 0, 100000) // 100,000 size array
	PrimeNumbersSlice = append(PrimeNumbersSlice, 2)
	PrimeMap[2] = true
	PrimeMap[MaxPrime32] = true

	var u uint

	const updateinterval time.Duration = 1e10 // 1e10 ns = 10 sec
	ticker := time.NewTicker(updateinterval)
	defer ticker.Stop()

	for u = 3; u < StopVal; u += 2 {
		select {
		case <-ticker.C:
			s := strconv.FormatUint(uint64(u), 10)
			s = hpcalc.AddCommas(s)
			fmt.Println(" ticker: u =", u, ", s=", s)
		default: // no value ready to be received
			if IsPrimeInt(u) {
				PrimeMap[u] = true
				PrimeNumbersSlice = append(PrimeNumbersSlice, u)
			}
		}
	}

	ticker.Stop()

	fmt.Println(" Len of slice is ", len(PrimeNumbersSlice))
	fmt.Println(" Len of PrimeMap is ", len(PrimeMap))

	// Time to write files before exiting.

	themapfile, err := os.Create(PrimeMapFilename) // for writing
	check(err)                                     // This should not fail, so panic if it does fail.
	defer themapfile.Close()

	encoder := gob.NewEncoder(themapfile) // an encoder writes the file
	err = encoder.Encode(&PrimeMap)       // this worked with or without the &
	check(err)                            // Panic if there is an error
	themapfile.Close()

	theslicefile, err := os.Create(PrimeSliceFilename)
	check(err)
	defer theslicefile.Close()
	encoder = gob.NewEncoder(theslicefile)   // an encoder writes the file
	err = encoder.Encode(&PrimeNumbersSlice) // this worked with or without the &
	check(err)

	theslicefile.Close()
	/*
		fmt.Println(" PrimeMap:")
		fmt.Print(PrimeMap)
		fmt.Println()

		fmt.Println(" PrimeNumbersSlice:")
		fmt.Print(PrimeNumbersSlice[0:50])
		fmt.Println()
		fmt.Println()
	*/
}

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------- PrimeFactorization ------------------------------

func PrimeFactorization(N int) []int {

	var PD = []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47} // Prime divisors array

	if N == 0 {
		return nil
	}

	PrimeFactors := make([]int, 0, 10)

	flag := IsPrimeInt(uint(N))
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
		primeflag := IsPrimeInt(uint(n))
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
//                                                               func IsPrimeInt(n uint) (uint, bool) {
func IsPrimeInt(n uint) bool {

	var t uint64 = 3

	Uint := uint64(n)

	if Uint == 0 || Uint == 1 {
		return false
	} else if Uint == 2 || Uint == 3 {
		return true
	} else if Uint%2 == 0 {
		return false
	}

	//	sqrt := math.Sqrt(float64(Uint))
	//	UintSqrt := uint(sqrt)
	UintSqrt := usqrt(uint64(n))

	for t <= UintSqrt {
		if Uint%t == 0 {
			return false
		}
		t += 2
	}
	return true
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

/*

The time package has some interesting functionality to use in combination with channels.

It contains a struct time.Ticker which is an object that repeatedly sends a time value on a contained channel C at a specified time interval:

type Ticker struct {
   C <-chan Time // the channel on which the ticks are delivered.
   // contains filtered or unexported fields
   ..
}

The time interval ns is specified (in nanoseconds as an int64) is specified as a variable dur of type Duration in the factory function time.NewTicker:
    func NewTicker(dur) *Ticker

It can be very useful when during the execution of goroutines when something (logging of a status, a printout, a calculation, etc.) has to be done periodically at a certain time interval.

A Ticker is stopped with Stop(), use this in a defer statement.  All this fits nicely in a select statement:

ticker := time.NewTicker(updateInterval_in_ns_as_int64)
defer ticker.Stop()
..
select {
case u:= <- ch1:
        ..
case v:= <- ch2:
        ..
case <- ticker.C:
        logState(status) // call some logging function logState
default: // no value ready to be received
        ..
}

*/
