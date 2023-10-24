package main

import (
	"fmt"
	"math/rand"
)

// This is the pgm that will demo the RSA algorithm.  RSA is in the public domain, hence, it can be used as a basis for these projects.
// Rivest Shamir Adleman is how the algorithm is named.  These are the mathematicians from MIT that came up w/ the algorithm in 1977.  A patent was granted in 1983, and expired in 2000.

const numOfTests = 30 // Will set this at top of package so it's easy to change later.  // 20 gives a similar chance of being struck by lightning/yr, 30 gives better odds of winning powerball than the numbers not being prime.

func totient(p, q int) int { //  totient is lambda(n) where n = p x q.  p and q are primes.  lambda(n) is the smallest value, m, where a**m == (1 mod n) for all 'a' that are relatively prime to n.
	// This sounds like Fermat's little theorem.
	// n = pq, lambda(n) = lcm(lambda(p), lambda(q)). Since p and q are prime, (skipping math) lambda(n) = lcm(p-1, q-1).  Lambda(n) is kept secret.
	return lcm(p-1, q-1)
}

func randomExponent(lambdaN int) int { // picks random public exponent, e, btwn 2 and lambda(n), such that gcd(e, lambda(n)) = 1.
	var g, e int
	for {
		e = randRange(3, lambdaN)
		g = gcd(e, lambdaN)
		if g == 1 {
			break
		}
	}
	return e
}

func inverseMod(a, n int) int { // I don't get this at all.
	t := 0
	newT := 1
	r := n
	newR := a
	var ctr int

	for newR != 0 {
		ctr++
		quotient := r / newR
		t = newT
		newT = t - quotient*newT
		r = newR
		newR = r - quotient*newR
	}

	if r > 1 { // then 'a' is not invertable
		return -1
	}
	if t < 0 {
		t += n
	}

	return t
}

func main() {
	p := randRange(10_000, 50_000) // using larger numbers here would run into integer overflow issues.
	q := randRange(10_000, 50_000) // would need hundreds of digits, like 6 hundred digits.
	n := p * q                     // this is the public key modulus
	lambdaN := totient(p, q)
	e := randomExponent(lambdaN)
	d := inverseMod(e, lambdaN)

	fmt.Printf("*** Public ***\n public key modulus:  %d\n public key exponent: %d\n\n", n, e)
	fmt.Printf("*** Private ***\n Primes:    %d, %d \n lambda(n): %d \n d:        %d\n", p, q, lambdaN, d)
}

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

func gcd(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	if a < b { // swap 'em
		b, a = a, b
	}

	for {
		r := a % b
		//fmt.Printf(" a = %d, b = %d, r = %d\n", a, b, r)
		if r == 0 {
			return b
		}
		a = b
		b = r
	}
}

func hcf(a, b int) int {
	return gcd(a, b)
}

func lcm(a, b int) int { // lowest common multiple
	intermed := b / gcd(a, b) // to not cause an overflow.  Intermed will not have a remainder as this is a constraint of gcd.
	return a * intermed
}
