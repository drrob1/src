package main // mattestqr from mattest2 from mattest.  Both test mat.  Duh!

/*
REVISION HISTORY
================
21 Dec 16 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.
24 Dec 16 -- Seems to work.
29 Dec 16 -- Tweaked Write field size values
13 Feb 22 -- Converted to modules
21 Nov 22 -- static linter found issues.  Now addressed.
 1 Apr 23 -- Since I'm here because of StaticCheck, I'll fix some of the messages and update the code.
10 Mar 24 -- Now called mattest2, derived from mattest.  I'm updating to Go 1.22, and will generate test data if no input file is specified.
12 Mar 24 -- Playing w/ gonum.org mat package, from Miami
12 Mar 24 -- Now mattestQR, to try to simplify it to isolate why it's not working.
18 Mar 24 -- More playing, now that I'm homw again.
*/

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"gonum.org/v1/gonum/mat"
	"math/rand/v2"
	"os"
)

const aRows = 3
const aCols = aRows
const bRows = aRows
const bCols = 1

func goNumQRTest() {
	// Will look to solve AX = B, for X

	initialVal := float64(randRange(1, 10))
	increment := float64(randRange(1, 5))

	initX := make([]float64, aCols)
	initX[0] = initialVal
	initX[1] = initialVal + increment
	initX[2] = initialVal + 2*increment

	X := mat.NewVecDense(bRows, initX)
	fmt.Printf(" X:\n%.4g\n\n", mat.Formatted(X))

	// Now need to assign coefficients in matrix A
	initA := make([]float64, aRows*aCols) // 3 x 3 = 9, as of this writing.

	for i := range initA {
		initA[i] = float64(randRange(1, 20))
	}

	A := mat.NewDense(aRows, aCols, initA)
	fmt.Printf(" A:\n%.4g\n\n", mat.Formatted(A))

	var B mat.Dense
	B.Mul(A, X)
	fmt.Printf(" B:\n%.4g\n\n", mat.Formatted(&B))

	// try QR stuff
	var qr mat.QR
	qr.Factorize(A)
	qrSoln := mat.NewDense(bRows, bCols, nil)
	err := qr.SolveTo(qrSoln, false, &B) // this panics w/ invalid memory reference or nil pointer dereference.
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from qr Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Soluton by gonum QR factorization is:\n%.5g\n\n", mat.Formatted(qrSoln))

} // end gonumQRTest

//                              MAIN PROGRAM

func main() {
	goNumQRTest()

}

// ------------------------------------------------randRange -----------------------------------------------------------

func randRange(minP, maxP int) int { // note that this is not cryptographically secure.  Would need crypto/rand for that.
	if maxP < minP {
		minP, maxP = maxP, minP
	}
	return minP + rand.IntN(maxP-minP)
}
