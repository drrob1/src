package main

import (
	"fmt"
	"time"
)

// Sieve of Eratosthenes for prime numbers.

func sieveOfEratosthenes(mx int) []bool {
	if mx < 2 {
		return nil
	}

	sieve := make([]bool, mx)

	for i := range sieve {
		sieve[i] = true
	}
	sieve[0] = false

	if mx < 4 {
		return sieve
	}

	root := iSqrt(mx) // I have an off by one error, handled in the iSqrt routine.

	for i := 4; i < mx; i += 2 { // handling the even numbers > 2
		sieve[i] = false
	}

	for i := 3; i < root; i += 2 {
		if sieve[i] {
			for j := i * i; j < mx; j += i {
				sieve[j] = false
			}
		}
	}

	return sieve
}

func printSieve(sieve []bool) {
	n := len(sieve)

	if n > 2 {
		fmt.Printf(" %d", 2)
	}

	if n < 3 {
		return
	}

	for i := 3; i < n; i += 2 {
		if sieve[i] {
			fmt.Printf(" %d", i)
		}
		if i%150 == 149 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func sieveToPrimes(sieve []bool) []int {
	n := len(sieve)
	sq := iSqrt(n)
	primes := make([]int, 0, sq)

	if n > 2 {
		primes = append(primes, 2)
	}

	for i := 3; i < n; i += 2 {
		if sieve[i] {
			primes = append(primes, i)
		}
	}
	return primes
}

func main() {
	var max int
	fmt.Printf("Max: ")
	fmt.Scan(&max)

	start := time.Now()
	sieve := sieveOfEratosthenes(max)
	elapsed := time.Since(start)
	fmt.Printf("Elapsed: %f seconds, %s \n", elapsed.Seconds(), elapsed.String())

	primes := sieveToPrimes(sieve)
	fmt.Printf(" Found %d primes less than %d.\n", len(primes), max)
	if max <= 1000 {
		printSieve(sieve)
		fmt.Println()

		fmt.Println(primes)
	}
}

func iSqrt(i int) int { // this uses dividing and averaging
	if i <= 0 {
		return 0
	}

	sqrt := i / 2

	for j := 0; j < 30; j++ {
		guess := i / sqrt
		sqrt = (guess + sqrt) / 2
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}

	return sqrt + 1 // to address an off by 1 problem.
}
