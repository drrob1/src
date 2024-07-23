package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

/*
18 Oct 2023 -- Now called the Sieve of Euler, based on Sieve of Eratosthenes for prime numbers.
				The algorithm is to make p loop over odd numbers from 3 .. max.  For each loop, calculate the largest odd integer max / p, and then loop down from that maxQ to p.
				If p is marked prime, then p * q is marked not prime.
				Something I just noticed by his solution and mine.  In sieveOfEratosthenes, he does for i := 3; i < max; i += 2.  I don't go to the max, but instead I go to sqrt(max).
 9 Jul 2024 -- I'm going to try to benchmark this according to what Dave Plumber of Dave's Garage did, on Win11 desktop.
				IE, run for 5 sec and count how many runs it does per sec for the sieve of 1 million elements.
                   Result for Sieve of Eratosthenes is ~997/sec.
                   Result for Sieve of Euler is ~446/sec.
10 Jul 2024 -- Tuning the sieveOfEratosthenes
                   Result for Sieve of Eratosthenes is ~725/sec.
                   Result for Sieve of Eratosthenes is ~1310/sec when only go to sqrt(max).
11 Jul 2024 -- Minor tweaks
------------------------------------------------------------------------------------------------------------------------------------------------------
21 Jul 2024 -- Now called consieve, and I'm going to try to add concurrency to this routine to see how much faster I can get.  I removed the Euler Sieve, as it was slower.
				Result for Sieve of Eratosthenes is ~7200/sec, on Win11 desktop w/ a Ryzen 9 CPU, 5950X, and workers = NumCPU()-1
				Result for Sieve of Eratosthenes is ~7400/sec, on Win11 desktop w/ a Ryzen 9 CPU, 5950X, and workers = NumCPU()
				Result for Sieve of Eratosthenes is ~7450-7480/sec, on Win11 desktop w/ a Ryzen 9 CPU, 5950X, and workers = NumCPU()+1
				Result for Sieve of Eratosthenes is ~2300/sec, on leox desktop, and workers = NumCPU()+1
				Result for Sieve of Eratosthenes is ~8700/sec, on thelio desktop w/ a Ryzen 9 CPU, 5950X, and workers = NumCPU()+1
22 Jul 2024 -- Result for the Sieve from work is ~4500/sec, before I started everything else, and it didn't change after I did start everything else.
*/

const LastModified = "21 July 2024"
const timeForTesting = 5 * time.Second

var wg sync.WaitGroup
var workers = runtime.NumCPU() + 1
var total int64

func sieveOfEratosthenes(mx int) []bool {
	if mx < 2 {
		return nil
	}

	sieve := make([]bool, mx+1) // part of author's solution

	sieve[2] = true
	for i := 3; i < mx; i += 2 { // Here I set only the odd numbers to default to be prime.
		sieve[i] = true
	}

	if mx < 4 {
		return sieve
	}

	root := iSqrt(mx) // I have an off by one error, handled in the iSqrt routine.

	for i := 3; i < root; i += 2 { // this works because of the next loop that's j := i**2
		if sieve[i] {
			for j := i * i; j < mx; j += i { // turns out that the square of an odd number is always odd. (2n + 1) **2 -> 4n^2 + 4n + 1
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
	var sieve []bool

	var max = 1_000_000
	fmt.Printf(" Concurrent Sieve of Eratosthenes, last modified %s, Enter Max: ", LastModified)
	fmt.Scanln(&max)
	fmt.Printf(" Using max of %d.\n", max)

	t0 := time.Now()
	tFinal := t0.Add(timeForTesting)

	if workers < 1 {
		workers = 1
	}

	// spin up the workers
	wg.Add(workers)
	for j := range workers {
		go func() {
			var i int64
			defer wg.Done()
			for i = 0; ; i++ {
				now := time.Now()
				if now.After(tFinal) {
					dur := now.Sub(t0)
					sec := dur.Seconds()
					fmt.Printf("Elapsed time: %.5f for worker %02d, i = %d, dur=%.5f, rate = %.0f per second.\n",
						now.Sub(t0).Seconds(), j, i, dur.Seconds(), float64(i)/sec)
					break
				}
				sieve = sieveOfEratosthenes(max)
			}
			atomic.AddInt64(&total, i)
		}()
	}

	wg.Wait()

	elapsed := time.Since(t0)
	ctfmt.Printf(ct.Yellow, true, "\n\n Elapsed: %f seconds, %s, total runs=%d, rate = %.0f/sec\n",
		elapsed.Seconds(), elapsed.String(), total, float64(total)/elapsed.Seconds())

	primes := sieveToPrimes(sieve)
	fmt.Printf(" Found %d primes less than %d.\n", len(primes), max)
	if max <= 1000 {
		printSieve(sieve)
		fmt.Println()
		fmt.Println(primes)
	}
	fmt.Println()
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
