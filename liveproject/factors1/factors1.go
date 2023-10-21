package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"time"
)

// For the liveproject.  Starting to write routines to factor numbers.
// Don't need to check all numbers, as really looking for prime factors.  Pgm ends up w/ prime factors by starting small, so non-prime factors are removed.
// Next steps of this project is to use a sieve routine to generate primes up to 20 million.

var primes []int

func findFactors(num int) []int {
	if num <= 0 {
		return nil
	}

	factors := make([]int, 0, 10) // I just plucked 10 out of the air.
	for num%2 == 0 && num > 0 {   // This loop stops when num is odd.
		if num%2 == 0 { // this if statement is probably redundant.
			factors = append(factors, 2)
			num = num / 2
		}
	}
	if num < 2 { // this means that the number is even.
		return factors
	}
	//fmt.Printf(" in findFactors after check for factors of 2:  num=%d, factors so far is/are: %v\n", num, factors)

	// Now all factors of 2 have been handled.  Handle larger factors.  I don't understand the project instructions at all.

	factor := 3
	numSQRT := iSqrt(num)
	//fmt.Printf(" in findFactors before loop: factor=%d, num = %d, numSQRT=%d\n", factor, num, numSQRT)
	for factor <= numSQRT && num > 1 { // only need to check factors < sqrt(num)
		if num%factor == 0 {
			factors = append(factors, factor)
			num = num / factor
			//fmt.Printf(" in found Factors loop: factor=%d, num = %d\n", factor, num)
			continue
		}
		// factor is not a factor for num, so increment factor and try again.
		factor += 2
		//fmt.Printf(" in findFactors loop: factor=%d, num = %d\n", factor, num)
	}

	// If there's a factor that > sqrt(num), include it.
	if num > 1 {
		factors = append(factors, num)
	}

	return factors
}

func findFactorsSieve(num int) []int {
	if num <= 0 {
		return nil
	}

	factors := make([]int, 0, 10) // I just plucked 10 out of the air.
	for num%2 == 0 && num > 0 {   // This loop stops when num is odd.
		if num%2 == 0 {
			factors = append(factors, 2)
			num = num / 2
		}
	}
	if num < 2 { // this means that the number is even.
		return factors
	}

	// Now all factors of 2 have been handled.  Handle larger factors.

	factor := 3 // the primes slice[0] = 2, so primes slice[1] = 3
	numSQRT := iSqrt(num)
	lenPrimes := len(primes)
	for facIdx := 1; factor <= numSQRT && num > 1 && facIdx < lenPrimes; facIdx++ { // only need to check factors < sqrt(num), and for very big numbers I have to make sure we don't panic.
		factor = primes[facIdx]
		//if num%factor == 0 { // this is not right.  This will only check a factor once.
		//	factors = append(factors, factor)
		//	num = num / factor
		//}
		for num%factor == 0 {
			factors = append(factors, factor)
			num = num / factor
		}
	}

	// If there's a factor that > sqrt(num), include it.
	if num > 1 {
		factors = append(factors, num)
	}

	return factors
}

func multiplyFactors(factors []int) int {
	product := 1
	for _, fac := range factors {
		product *= fac
	}
	return product
}

func main() {
	var num int

	sieveTime := time.Now()
	//sieve := eulerSieve(20_000_000)
	sieve := eulerSieve(1_600_000_000)
	elapsedSieveTime := time.Since(sieveTime)
	primes = sieveToPrimes(sieve)
	fmt.Printf(" Time to create a sieve up to 1.6 billion is %s\n", elapsedSieveTime.String()) // turned out to be ~7.5 sec for 1.6 billion, and 8.4 sec for 1.8 billion on win11 system.

	for {
		fmt.Printf(" number: ")
		n, err := fmt.Scanln(&num)
		if err != nil || n == 0 {
			break
		}

		start := time.Now()

		factors := findFactors(num)
		elapsed := time.Since(start)
		check := multiplyFactors(factors)
		if num-check == 0 {
			ctfmt.Printf(ct.Green, false, " Factors = [%v], in %s, check = %d\n", factors, elapsed.String(), check)
		} else {
			ctfmt.Printf(ct.Red, true, " Factors = [%v], in %s, check = %d\n", factors, elapsed.String(), check)
		}

		start = time.Now()
		primeFactors := findFactorsSieve(num)
		elapsed = time.Since(start)
		check = multiplyFactors(primeFactors)
		if num-check == 0 {
			ctfmt.Printf(ct.Green, false, " Factors = [%v], in %s, check = %d\n\n", primeFactors, elapsed.String(), check)
		} else {
			ctfmt.Printf(ct.Red, true, " Factors = [%v], in %s, check = %d\n\n", primeFactors, elapsed.String(), check)
		}
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

func eulerSieve(mx int) []bool {
	if mx < 2 {
		return nil
	}

	sieve := make([]bool, mx+1) // part of author's solution.  I need this here because when I tested w/ 999, I got an index out of bounds on line 77.

	sieve[2] = true
	for i := 3; i < mx; i += 2 {
		sieve[i] = true
	}

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
			if sieve[q] { // I misunderstood the algorithm.  I first made this if sieve[p*q], but the author uses this expression so I changed my code.
				sieve[p*q] = false
			}
		}
	}

	return sieve
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
