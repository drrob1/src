package main

import "fmt"

// factorials.go for the factorials project part of the recursion liveproject
// 21! is too big for even a uint64.  It is correctly computed by using a float64.

func factorial(n int64) int64 {
	if n < 0 {
		return 0
	} else if n < 2 {
		return 1
	}
	return n * factorial(n-1)
}

func ufactorial(n uint64) uint64 {
	if n < 2 {
		return 1
	}
	return n * ufactorial(n-1)
}

func rfactorial(r float64) float64 { // real factorial
	if r < 2 {
		return 1
	}
	return r * rfactorial(r-1)
}

func main() {
	var n int64
	var u uint64

	for n = 0; n <= 21; n++ {
		fmt.Printf(" %03d! = %20d\n", n, factorial(n))
	}
	for u = 0; u <= 21; u++ {
		fmt.Printf(" %03d! = %20d\n", u, ufactorial(u))
	}
	fmt.Println()
	fmt.Printf(" %03d! = %20.0f\n", 21, rfactorial(21))
	fmt.Println()
}
