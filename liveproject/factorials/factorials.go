package main

import "fmt"

// factorials.go for the factorials project part of the recursion liveproject

func factorial(n int64) int64 {
	if n < 0 {
		return 0
	} else if n < 2 {
		return 1
	}
	return n * factorial(n-1)
}

func main() {
	var n int64

	for n = 0; n <= 21; n++ {
		fmt.Printf(" %03d! = %20d\n", n, factorial(n))
	}
	fmt.Println()
}
