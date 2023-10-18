package main

import (
	"fmt"
	"time"
)

// Now called the Sieve of Euler, based on Sieve of Eratosthenes for prime numbers.
// The algorithm is to make p loop over odd numbers from 3 .. max.  For each loop, calculate the largest odd integer max / p, and then loop down from that maxQ to p.
// If p is marked prime, then p * q is marked not prime.

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

func eulerSieve(mx int) []bool {
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

	for p := 3; p < mx; p += 2 {
		// need largest odd divisor of max/p.
		maxQ := mx / p
		if maxQ%2 == 0 { // then this is even
			maxQ-- // make the number odd
		}
		for q := maxQ; q >= p; q -= 2 {
			if sieve[p*q] {
				sieve[p*q] = false
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
	fmt.Printf(" Elapsed: %f seconds, %s \n", elapsed.Seconds(), elapsed.String())

	primes := sieveToPrimes(sieve)
	fmt.Printf(" Found %d primes less than %d.\n", len(primes), max)
	if max <= 1000 {
		printSieve(sieve)
		fmt.Println()

		fmt.Println(primes)
	}

	start = time.Now()
	euler := eulerSieve(max)
	elapsed = time.Since(start)
	fmt.Printf(" Elapsed for Euler: %f sec, %s\n", elapsed.Seconds(), elapsed.String())
	eulerPrimes := sieveToPrimes(euler)
	fmt.Printf(" Found %d primes less than %d.\n", len(eulerPrimes), max)
	if max <= 1000 {
		printSieve(euler)
		fmt.Println()
		fmt.Println(eulerPrimes)
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
