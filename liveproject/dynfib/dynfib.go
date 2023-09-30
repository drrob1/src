package main

import (
	"fmt"
	"strconv"
)

// This uses dynamic fibonacci, or uses memoization to speed up the routines.
// The author describes 3 methods, one is top-down, and 2 are bottom-up.

const maxFib = 93

var fibonacciValues []int64
var fibonacciFlyValues []int64

func fibonacciOnTheFly(n int) int64 {
	if n < 2 {
		return int64(n)
	}

	if len(fibonacciFlyValues)-1 < n { // this value of n has not yet been memoized.  So memoize it.
		fibI := fibonacciOnTheFly(n-2) + fibonacciOnTheFly(n-1)
		fibonacciFlyValues = append(fibonacciFlyValues, fibI)
	}

	return fibonacciFlyValues[n]
}

func initializeSlice() {
	fibonacciFlyValues = make([]int64, 0, maxFib)
	fibonacciFlyValues = append(fibonacciFlyValues, 0)
	fibonacciFlyValues = append(fibonacciFlyValues, 1)

	fibonacciValues = make([]int64, 0, maxFib)
	fibonacciValues = append(fibonacciValues, 0)
	fibonacciValues = append(fibonacciValues, 1)

	for i := 2; i <= maxFib; i++ {
		fibonacciValues = append(fibonacciValues, fibonacciValues[i-1]+fibonacciValues[i-2])
	}
}

func fibonacciPreFilled(n int) int64 {
	if n < 0 {
		panic("fibonacci n < 0 not allowed.")
	}
	if n > maxFib {
		s := fmt.Sprintf("Fibonacci > %d not allowed.  n = %d", maxFib, n)
		panic(s)
	}
	return fibonacciValues[n]
}

func fibonacciBottomUp(n int) int64 {
	if n < 0 {
		panic("fibonacci n < 0 not allowed.")
	}
	if n < 2 {
		return int64(n)
	}

	var fibMinus1, fibMinus2, fibI int64
	fibMinus1 = 1
	fibMinus2 = 0
	for i := 1; i < n; i++ {
		fibI = fibMinus1 + fibMinus2

		fibMinus2 = fibMinus1
		fibMinus1 = fibI
	}
	return fibI
}

func main() {
	// Fill-on-the-fly.
	//fibonacciValues = make([]int64, 2)
	//fibonacciValues[0] = 0
	//fibonacciValues[1] = 1

	// Prefilled.
	initializeSlice()

	for {
		// Get n as a string.
		var nString string
		fmt.Printf("N: ")
		fmt.Scanln(&nString)

		// If the n string is blank, break out of the loop.
		if len(nString) == 0 {
			break
		}

		// Convert to int and calculate the Fibonacci number.
		//n, _ := strconv.ParseInt(nString, 10, 64)
		n, err := strconv.Atoi(nString)
		if err != nil {
			fmt.Printf(" Converting %q to an int returned ERROR %s.  Aborting.\n", nString, err)
			break
		}

		// Uncomment one of the following.
		fmt.Printf("fibonacciOnTheFly(%d) = %d\n", n, fibonacciOnTheFly(n))
		fmt.Printf("fibonacciPrefilled(%d)  = %d\n", n, fibonacciPreFilled(n))
		fmt.Printf("fibonacciBottomUp(%d)  = %d\n", n, fibonacciBottomUp(n))
	}

	// Print out all memoized values just so we can see them.
	fmt.Printf(" Len(fibonacciValues) = %d.  Len(FibonacciFlyValues) = %d", len(fibonacciValues), len(fibonacciFlyValues))
	for i := 0; i < len(fibonacciValues); i++ {
		fmt.Printf("%d: %d; %d\n", i, fibonacciValues[i], fibonacciFlyValues[i])
	}
}
