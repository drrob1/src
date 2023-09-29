package main

import (
	"fmt"
	"strconv"
)

func fibonacci(n int64) int64 {
	if n < 0 {
		panic("fibonacci n < 0 not allowed.")
	}
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
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
		n, err := strconv.ParseInt(nString, 10, 64)
		fmt.Printf("fibonacci(%d) = %d\n", n, fibonacci(n))
		if err != nil {
			break
		}
	}
}
