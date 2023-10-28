package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// This is the pgm that will demo the RSA algorithm.  RSA is in the public domain, hence, it can be used as a basis for these projects.
// Rivest Shamir Adleman is how the algorithm is named.  These are the mathematicians from MIT that came up w/ the algorithm in 1977.  A patent was granted in 1983, and expired in 2000.
// I used the digraph ^k l* to enter the greek lower case lambda.

const numOfTests = 30 // Will set this at top of package so it's easy to change later.  // 20 gives a similar chance of being struck by lightning/yr, 30 gives better odds of winning powerball than the numbers not being prime.

func totient(p, q int) int { //  totient is λ(n) where n = p x q.  p and q are primes.  λ(n) is the smallest value, m, where a**m == (1 mod n) for all 'a' that are relatively prime to n.
	// This sounds like Fermat's little theorem.
	// n = pq, λ(n) = lcm(λ(p), λ(q)). Since p and q are prime, (skipping math) λ(n) = lcm(p-1, q-1).  λ(n) is kept secret.
	return lcm(p-1, q-1)
}

func randomExponent(λN int) int { // picks random public exponent, e, btwn 2 and λ(n), such that gcd(e, λ(n)) = 1.  IE, e and λ(n) are mutually prime.
	var g, e int
	for {
		e = randRange(3, λN)
		g = gcd(e, λN)
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

	for newR != 0 {
		quotient := r / newR
		t, newT = newT, t-quotient*newT
		r, newR = newR, r-quotient*newR
	}

	if r > 1 { // then 'a' is not invertible
		return -1
	}
	if t < 0 {
		t += n
	}

	return t
}

func main() {
	p := findPrime(10_000, 50_000, numOfTests) // using larger numbers here would run into integer overflow issues.
	q := findPrime(10_000, 50_000, numOfTests) // would need hundreds of digits, like 6 hundred digits.
	n := p * q                                 // this is the public key modulus
	λN := totient(p, q)
	e := randomExponent(λN)
	d := inverseMod(e, λN)

	fmt.Printf("*** Public ***\n public key modulus:  %d\n public key exponent: %d\n\n", n, e)
	fmt.Printf("*** Private ***\n Primes:    %d, %d \n λ(n): %d \n d:        %d\n", p, q, λN, d)

	if d < 1 {
		fmt.Printf(" inverseMod failed.  Aborting.\n")
		os.Exit(1)
	}

	var m int
	for {
		fmt.Printf(" Enter a number > 1 and < public key modulus, %d: ", n)
		numChars, err := fmt.Scanln(&m)
		if numChars == 0 || err != nil || m < 1 {
			break
		}
		if m >= n {
			fmt.Printf(" Number too large.  Try again.\n")
			continue
		}
		cipherText := fastExpMod(m, e, n)
		fmt.Printf(" cipherText is %d.\n", cipherText)
		plainText := fastExpMod(cipherText, d, n)
		fmt.Printf(" plainText is %d.\n", plainText)
	}

	var ans string
	var cipherText []int
	var plainText strings.Builder
	for {
		fmt.Printf(" Enter a word: ")
		numChars, err := fmt.Scanln(&ans)
		if numChars == 0 || err != nil {
			break
		}
		for _, ch := range ans {
			cipher := fastExpMod(int(ch), e, n)
			cipherText = append(cipherText, cipher)
		}
		fmt.Printf(" CipherText: %+v\n", cipherText)

		for _, i := range cipherText {
			ch := fastExpMod(i, d, n)
			plainText.WriteRune(rune(ch))
		}
		fmt.Printf(" PlainText: %s\n", plainText.String())
		cipherText = []int{}
		plainText.Reset()
	}
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
		n := randRange(1, p)
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
		if isProbablyPrime(p, numTests) { // test for being even is in isProbablyPrime.
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

func lcm(a, b int) int { // lowest common multiple
	intermed := b / gcd(a, b) // to not cause an overflow.  Intermed will not have a remainder as this is a constraint of gcd.
	return a * intermed
}
