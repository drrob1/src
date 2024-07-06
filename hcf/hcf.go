package main

import (
	crypt "crypto/rand"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"math/rand"
	"os"
	"time"
)

var rn *rand.Rand

func hcf(a, b int) int {
	// a = bt + r, then hcf(a,b) = hcf(b,r)
	var r, a1, b1 int

	if a < b {
		a1 = b
		b1 = a
	} else {
		a1 = a
		b1 = b
	}
	for {
		r = a1 % b1 // % is MOD operator
		a1 = b1
		b1 = r
		if r == 0 {
			break
		}
	}
	return a1
} // HCF

func gcd(a, b int) int {
	// a = bt + r, then hcf(a,b) = hcf(b,r)

	var r int

	if a < b {
		a, b = b, a
	}

	for {
		r = a % b
		if r == 0 {
			break
		}
		a = b
		b = r
	}
	return b
}

func randRange(minP, maxP int) int { // note that this is not cryptographically secure.  Would need crypto/rand for that.
	if maxP < minP {
		minP, maxP = maxP, minP
	}
	return minP + rn.Intn(maxP-minP)
}

func main() {
	c := 8
	b := make([]byte, c)
	_, err := crypt.Read(b)
	if err != nil {
		fmt.Printf(" Error from crypt.Read: %s\n", err)
		os.Exit(1)
	}

	var i64 int64
	for _, b := range b { // convert the random bytes to a single int64
		i64 = 256*i64 + int64(b)
	}
	t := time.Now().Unix()
	i64 += t
	fmt.Printf(" random int64 is %d, and t = %d\n", i64, t)
	rn = rand.New(rand.NewSource(i64))

	for range 25 {
		i := randRange(1, 1000)
		j := randRange(1, 1000)

		if gcd(i, j) == hcf(j, i) {
			ctfmt.Printf(ct.Green, false, "GCD(%d, %d) = %d\n", i, j, gcd(i, j))
		} else {
			ctfmt.Printf(ct.Red, true, " i = %d, j = %d.  GCD and HCF should be equal but are not.\n", i, j)
		}
	}

	for i := 0; i < 1000; i++ {
		j := randRange(1, 1000)
		k := randRange(1, 1000)
		if hcf(j, k) == hcf(k, j) {
			ctfmt.Printf(ct.Green, false, "GCD(%d, %d) = %d, and permutation is working.\n", k, j, gcd(k, j))
		} else {
			ctfmt.Printf(ct.Red, true, " permutation isn't working.  j=%d, k=%d, hcf(j,k)=%d, hcf(k,j)=%d.  These should be equal but are not.\n", j, k, hcf(j, k), hcf(k, j))

		}

		if hcf(j, k) != gcd(k, j) {
			ctfmt.Printf(ct.Red, true, " HCF(%d,%d)=%d is not equal to GCD()=%d\n", j, k, hcf(j, j), gcd(k, j))
		}
	}
}
