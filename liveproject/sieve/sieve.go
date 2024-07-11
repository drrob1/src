package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

/*
  Sieve of Eratosthenes for prime numbers.
  15 Oct 2023 -- First version written
   9 Jul 2024 -- I'm going to try to benchmark this according to what Dave of Dave's Garage did.  IE, run for 5 sec and count how many runs it does per sec for the sieve of 1 million elements.
                   Result is ~997/sec.
  10 Jul 2024 -- Removed a superfluous loop.
  11 Jul 2024 -- Added simple code to be a baseline on the speed issue.
				Nevermind.  I wrote code that doesn't work because the first occurrence of each factor doesn't make a number not be prime.
				It's the next occurrence of each factor that does.  Hence, it starts w/ i^2 for the inner loop, as i^2 is the next occurrence of that factor.
				I took the code out.
*/

const LastModified = "11 July 2024"
const iter = 2 * time.Second

func sieveOfEratosthenes(mx int) []bool {
	if mx < 2 {
		return nil
	}

	sieve := make([]bool, mx)

	for i := range sieve {
		sieve[i] = true
	}
	sieve[0] = false
	sieve[1] = true
	sieve[2] = true

	if mx < 4 {
		return sieve
	}

	root := iSqrt(mx) // I have an off by one error, handled in the iSqrt routine.

	for i := 4; i < mx; i += 2 {
		sieve[i] = false
	}

	for i := 3; i < root; i += 2 { // i < sqrt(mx) works because of the next loop where j := i**2
		if sieve[i] {
			for j := i * i; j < mx; j += i { // the square of an odd number is always odd.  (2n+1)^2 = 4n^2 + 4n + 1
				sieve[j] = false
			}
		}
	}

	return sieve
}

func basicSieveWithIf(mx int) []bool {
	if mx < 2 {
		return nil
	}

	siev := make([]bool, mx)

	siev[0] = false
	siev[2] = true
	for i := 3; i < mx; i += 2 {
		siev[i] = true
	}
	if mx <= 100 {
		fmt.Printf(" in basicSieveWithIf, all odd #s should be true: Sieve: %v\n", siev)
	}

	for i := 3; i < mx; i += 2 {
		if siev[i] {
			for j := i; j < mx; j += i {
				siev[j] = false
			}
			if mx <= 100 {
				fmt.Printf(" in basicSieveWithIf, some should be true: Sieve: %v\n", siev)
			}
			if pause() {
				os.Exit(1)
			}
		}
	}
	return siev
}

func basicSieveWithOutIf(mx int) []bool {
	if mx < 2 {
		return nil
	}

	siev := make([]bool, mx)

	for i := 3; i < mx; i += 2 {
		siev[i] = true
	}
	siev[0] = false
	siev[2] = true

	for i := 3; i < mx; i += 2 {

		for j := i; j < mx; j += i {
			siev[j] = false
		}

	}
	return siev
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
	var mx = 1_000_000
	var sieve []bool
	fmt.Printf(" Sieve of Eratosthenes, last modified %s, Enter Max: ", LastModified)
	fmt.Scanln(&mx)
	fmt.Printf(" Using max of %d.\n", mx)

	t0 := time.Now()
	tFinal := t0.Add(iter)
	for i := 0; ; i++ {
		now := time.Now()
		dur := now.Sub(t0)
		sec := dur.Seconds()
		if now.After(tFinal) {
			fmt.Printf("\n Sieve of E.  Elapsed time: %s, i = %d, dur=%v type = %T, rate = %.2f per second.\n", now.Sub(t0), i, dur, dur, float64(i)/sec)
			break
		}
		sieve = sieveOfEratosthenes(mx)
	}
	elapsed := time.Since(t0)
	fmt.Printf("Elapsed Sieve of E : %f seconds, %s \n", elapsed.Seconds(), elapsed.String())
	primes := sieveToPrimes(sieve)
	fmt.Printf(" Found %d primes less than %d.\n", len(primes), mx)
	if mx <= 1000 {
		printSieve(sieve)
		fmt.Println()

		fmt.Println(primes)
	}

	t0 = time.Now()
	tFinal = t0.Add(iter)
	for i := 0; ; i++ {
		now := time.Now()
		dur := now.Sub(t0)
		sec := dur.Seconds()
		if now.After(tFinal) {
			fmt.Printf("\n BasicSieveWithIf.  Elapsed time: %s, i = %d, dur=%v type = %T, rate = %.2f per second.\n", now.Sub(t0), i, dur, dur, float64(i)/sec)
			break
		}
		sieve = basicSieveWithIf(mx)
		if pause() {
			os.Exit(1)
		}
	}
	elapsed = time.Since(t0)
	fmt.Printf("Elapsed BasicWith: %f seconds, %s \n", elapsed.Seconds(), elapsed.String())

	primes = sieveToPrimes(sieve)
	fmt.Printf(" Found %d primes less than %d.\n", len(primes), mx)
	if mx <= 1000 {
		printSieve(sieve)
		fmt.Println()

		fmt.Println(primes)
	}

	t0 = time.Now()
	tFinal = t0.Add(iter)
	for i := 0; ; i++ {
		now := time.Now()
		dur := now.Sub(t0)
		sec := dur.Seconds()
		if now.After(tFinal) {
			fmt.Printf("\n BasicSieveWithOutIf.  Elapsed time: %s, i = %d, dur=%v type = %T, rate = %.2f per second.\n", now.Sub(t0), i, dur, dur, float64(i)/sec)
			break
		}
		sieve = basicSieveWithOutIf(mx)
	}
	elapsed = time.Since(t0)
	fmt.Printf("Elapsed BasicWithout: %f seconds, %s \n", elapsed.Seconds(), elapsed.String())

	primes = sieveToPrimes(sieve)
	fmt.Printf(" Found %d primes less than %d.\n", len(primes), mx)
	if mx <= 1000 {
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

// ------------------------------ pause -----------------------------------------

func pause() bool {
	fmt.Print(" Pausing the loop.  Hit <enter> to continue; 'n' or 'x' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "x") {
		return true
	}
	return false
}
