package gcd

import (
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"math/rand"
	"testing"
	"time"
)

// No matter what I do, I don't get the randomness I hoped for from the math/rand or math/rand/v2 packages.
// I'll try something from the crypto/rand package.

var rn *rand.Rand

func TestGCD(t *testing.T) {
	for range 25 {
		i := randRange(1, 1000)
		j := randRange(1, 1000)

		if GCD(i, j) != HCF(j, i) {
			t.Errorf(" i = %d, j = %d.  These should be equal but are not.\n", i, j)
		} else {
			ctfmt.Printf(ct.Green, false, "GCD(%d, %d) = %d\n", i, j, GCD(i, j))
		}
	}
}

func randRange(minP, maxP int) int { // note that this is not cryptographically secure.  Would need crypto/rand for that.
	if maxP < minP {
		minP, maxP = maxP, minP
	}
	return minP + rn.Intn(maxP-minP)
}

func TestHCF(t *testing.T) {
	for i := 0; i < 1000; i++ {
		j := randRange(1, 1000)
		k := randRange(1, 1000)
		if HCF(j, k) != HCF(k, j) {
			t.Errorf(" permutation isn't working.  j=%d, k=%d, HCF(j,k)=%d, HCF(k,j)=%d.  These should be equal but are not.\n", j, k, HCF(j, k), HCF(k, j))
		}
		if HCF(j, k) != GCD(k, j) {
			t.Errorf(" HCF(%d,%d)=%d is not equal to GCD()=%d\n", j, k, HCF(j, j), GCD(k, j))
		}
	}
}

func init() {
	t := time.Now().UnixNano()
	rn = rand.New(rand.NewSource(t))
}
