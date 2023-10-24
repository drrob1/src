package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"math"
	"math/rand"
)

// The only way to determine if a huge number (w/ 600 decimal digits) is prime is to use probability.  IE, determine if it's probably prime.
// Fermat little theorem: if P is prime and 1 <= n < P, then n**(P-1) mod P = 1.  IE, raise n to the P-1 power, modulo P, the result is 1.
// So to test if P is prime, pick a random n meeting the constraint 1 <= n < P, and if P really is prime, the result will be 1, probably.
// If you do this, and the modulo P != 1, then P is definitely not prime.
// So the algorithm is to pick a bunch of random n, see if they obey this theorem.  If any such tests fail the test, then P is definitely not prime.
// If you pick "k" values for n and the theorem succeeds, the probability that n is not prime (the algorithm is wrong) is 2**(-k).

// var random = rand.New(rand.NewSource(time.Now().UnixNano()))  But Go 1.20+ doesn't need this, so I'm not going to use it.

const numOfTests = 30 // Will set this at top of package so it's easy to change later.  // 20 gives a similar chance of being struck by lightning/yr, 30 gives better odds of winning powerball than the numbers not being prime.

func isProbablyPrime(p int, numTests int) bool {
	if p%2 == 0 {
		return false
	}
	if p == 1 {
		return false
	}
	// Run numTests number of Fermat's little theorem.  For any that fail, return false, if all succeed return true.  These are fast to check, so having numTests of 30, 50 or 100 will be fast.
	for i := 0; i < numTests; i++ {
		n := randRange(p/3, p)
		expMod := fastExpMod(n, p-1, p)
		//fmt.Printf(" n=%d, p=%d, ExpMod = %d\n", n, p, expMod)
		if expMod != 1 {
			return false
		}
	}
	return true
}

func randRange(minP, maxP int) int { // note that this is not cryptographically secure.  Writing a cryptographically secure pseudorandom number generator (CSPRNG) is beyond the scope of this exercise.
	return minP + rand.Intn(maxP-minP)
}

func findPrime(minP, maxP, numTests int) int { // I don't want to clobber the new built-in functions called min and max.
	// Will need an infinite for loop to keep calling isProbablyPrime until it succeeds.
	for {
		p := randRange(minP, maxP)
		fmt.Printf(" random prime candidate = %d\n", p)
		if isProbablyPrime(p, numOfTests) { // test for being even is in isProbablyPrime.
			return p
		}
	}
}

func testKnownValues() {
	primes := []int{
		10009, 11113, 11699, 12809, 14149,
		15643, 17107, 17881, 19301, 19793,
	}
	composites := []int{
		10323, 11397, 12212, 13503, 14599,
		16113, 17547, 17549, 18893, 19999,
	}
	for i := range primes { // will test both primes and composites here, using the same loop as the slices have the same number of elements.
		if isProbablyPrime(primes[i], numOfTests) {
			ctfmt.Printf(ct.Green, false, " %d is prime, and is probably Prime, which is correct.\n", primes[i])
		} else {
			ctfmt.Printf(ct.Red, true, " %d is prime, and is not probably Prime which is not correct.\n", primes[i])
		}
		if isProbablyPrime(composites[i], numOfTests) {
			ctfmt.Printf(ct.Red, true, " %d is composite and is probably prime, which is not correct.\n", composites[i])
		} else {
			ctfmt.Printf(ct.Green, false, " %d is composite and is not probably prime, which is correct.\n", composites[i])
		}
	}
}

func main() {
	prob := math.Pow(2, float64(-numOfTests))
	percentProb := prob * 100
	fmt.Printf(" Main pgm for findPrime for largish numbers.  will use %d tests of Fermat's little theorem, giving a probability of %.5g %% of being wrong.\n",
		numOfTests, percentProb)
	testKnownValues()

	// Now need to ask for # of digits, and then find a prime containing that number of digits.
	for {
		var numDigits int
		fmt.Printf(" Enter number of digits for the random prime: ")
		n, err := fmt.Scanln(&numDigits)
		if n == 0 || err != nil {
			break
		}
		if numDigits <= 0 || numDigits > 9 { // w/ > 9 digits, the pgm has problems w/ integer overflow.
			fmt.Printf(" %d is out of range.  Can't have more than 9 digits, or be zero or less.  Try again.\n", numDigits)
			continue
		}
		minF := math.Pow(10.0, float64(numDigits-1))
		maxF := 10 * minF
		minP := int(minF)
		maxP := int(maxF)
		fmt.Printf(" MinF = %.4g, maxF = %.4g, minP = %d, maxP = %d\n", minF, maxF, minP, maxP)
		p := findPrime(minP, maxP, numOfTests)
		fmt.Printf(" Random prime having %d digits is %d.\n", numDigits, p)
	}
}

func fastExpMod(num, pow, mod int) int { // pow can't be negative, or else it will panic.
	Z := 1
	if pow < 0 || mod < 0 {
		s := fmt.Sprintf("fastExpMod pow or mod cannot be negative.  pow = %d, mod = %d", pow, mod)
		panic(s)
	}
	for pow > 0 {
		if pow%2 == 1 { // ie, if pow is odd
			Z = (Z * num) % mod // Z = Z * R
		}
		num = (num * num) % mod // R = R squared
		pow /= 2                // I = half I
	}
	return Z //% mod
}
