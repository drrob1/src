package main // mattest2 from mattest.  Both test mat.  Duh!

/**********************************************************)
  (*                                                      *)
  (*              Test of Matrices module                 *)
  (*                                                      *)
  (*  Programmer:         P. Moylan                       *)
  (*  Last edited:        15 August 1996                  *)
  (*  Status:             Working                         *)
  (*                                                      *)
  (********************************************************/

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
18 Mar 24 -- Back home.  Playing some more.
*/

import (
	"bufio"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	gomat "gonum.org/v1/gonum/mat"
	"os"
	"src/mat"
	"src/misc"
	"strings"
)

const aRows = 3
const aCols = aRows
const bRows = aRows
const bCols = 1 // represents a column vector

func solveTest2() {

	var A, B, X mat.Matrix2D

	A = mat.NewMatrix(aRows, aCols)
	B = mat.NewMatrix(bRows, bCols)
	X = mat.NewMatrix(bRows, bCols)

	fmt.Println("Solving linear algebraic equations of form AX = B, solve for X")

	// Give a value to the A matrix.
	// I want these values to be whole positive numbers.  I need to determine the coefficient matrix, A, and values for the column vector, B.

	initialVal := misc.RandRange(1, 50)
	increment := misc.RandRange(1, 50)

	X[0][0] = float64(initialVal)
	X[1][0] = X[0][0] + float64(increment)
	X[2][0] = X[1][0] + float64(increment)

	// Now need to assign coefficients in matrix A
	for i := range A {
		for j := range A[0] {
			A[i][j] = float64(misc.RandRange(1, 20))
		}
	}

	fmt.Printf(" Coefficient matrix A is:\n")
	ss := mat.Write(A, 3)
	printString(ss)
	fmt.Println()

	//fmt.Printf(" x = %g, y = %g, z = %g\n\n", X[0][0], X[1][0], X[2][0])

	// Now do the calculation to determine what the V column vector needs to be for this to work.
	for i := range A {
		for j := range A[i] {
			product := A[i][j] * X[j][0]
			B[i][0] += product
			//fmt.Printf(" i=%d, j=%d, A[%d,%d] is %g, X[%d,0] is %g, product is %g, B[%d,0] is %g\n", i, j, i, j, A[i][j], i, X[j][0], product, i, B[i][0])
			//newPause()
		}
	}

	fmt.Printf("\n Column vectors X and B are:\n")
	ss = mat.WriteZeroPair(X, B, 4)
	printString(ss)
	fmt.Println()
	fmt.Printf("\n\n")

	// Another way to find the B column vector is to just do A*X.  It works.
	newB := mat.Mul(A, X)
	fmt.Printf(" Column vector newB is:\n")
	mat.WriteZeroln(newB, 6)

	solveSoln := mat.Solve(A, B)
	gaussSoln := mat.GaussJ(A, B)

	fmt.Printf("The solution X to AX = B\n using Solve       and then      GaussJ are:\n")
	ss = mat.WriteZeroPair(solveSoln, gaussSoln, 3)
	printString(ss)
	fmt.Println()

	if mat.EqualApprox(solveSoln, gaussSoln) {
		ctfmt.Printf(ct.Green, true, " The Solve and GaussJ methods returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and GaussJ methods DID NOT returned equal results.\n")
	}
	fmt.Println()

	if mat.EqualApprox(solveSoln, X) {
		ctfmt.Printf(ct.Green, true, " The Solve and X column vector returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and X column vector DID NOT returned equal results.\n")
	}
	fmt.Println()

	// Does AX - B = 0
	C := mat.Mul(A, solveSoln)
	D := mat.Sub(B, C)

	fmt.Println("As a check, AX-B should be 0, and evaluates to")
	ss = mat.Write(D, 3)
	printString(ss)
	ss = mat.WriteZero(D, 3)
	printString(ss)

	fmt.Printf("\n Will now use matrix inversion as a solution method.  Result is:\n")
	inverseA := mat.Invert(A)
	inverseSoln := mat.Mul(inverseA, B)
	ss = mat.WriteZero(inverseSoln, 3)
	printString(ss)

	solveInvert := mat.SolveInvert(A, B)

	if mat.EqualApprox(solveSoln, inverseSoln) {
		ctfmt.Printf(ct.Green, true, " The Solve and matrix inversion methods returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and matrix inversion methods DID NOT returned equal results.\n")
	}

	if mat.EqualApprox(solveInvert, inverseSoln) {
		ctfmt.Printf(ct.Green, true, " The SolveInvert and matrix inversion methods returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The SolveInvert and matrix inversion methods DID NOT returned equal results.\n")
	}

	fmt.Println()
	fmt.Println()

} // end SolveTest2

func printString(s []string) {
	for _, line := range s {
		ctfmt.Print(ct.Yellow, true, line)
	}

}

func goNumMatTest() {
	// Will look to solve AX = B, for X

	initialVal := float64(misc.RandRange(1, 50))
	increment := float64(misc.RandRange(1, 50))

	initX := make([]float64, aCols)
	initX[0] = initialVal
	initX[1] = initialVal + increment
	initX[2] = initialVal + 2*increment

	X := gomat.NewVecDense(bRows, initX)
	fmt.Printf(" X:\n%v\n\n", gomat.Formatted(X))

	// Now need to assign coefficients in matrix A
	initA := make([]float64, aRows*aCols) // 3 x 3 = 9, as of this writing.

	for i := range initA {
		initA[i] = float64(misc.RandRange(1, 20))
	}

	A := gomat.NewDense(aRows, aCols, initA)
	fmt.Printf(" A:\n%v\n\n", gomat.Formatted(A))

	initB := make([]float64, bRows) // col vec
	for i := range aRows {
		for j := range aCols {
			product := A.At(i, j) * X.At(j, 0)
			initB[i] += product
		}
	}
	Bvec := gomat.NewVecDense(bRows, initB)
	fmt.Printf(" Bvec:\n%v\n\n", gomat.Formatted(Bvec))

	// Will try w/ inversion
	var inverseA, invSoln, invSolnVec gomat.Dense
	err := inverseA.Inverse(A)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from inverting A: %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	invSolnVec.Mul(&inverseA, Bvec) // this works.  So far, it's the only method that does work.
	fmt.Printf(" Solution by GoNum inversion and Bvec is:\n%.5g\n\n", gomat.Formatted(&invSolnVec))
	//            fmt.Printf(" Solution by GoNum inversion is:\n%.5g\n\n", gomat.Formatted(&invSoln, gomat.Prefix("   "), gomat.Squeeze()))

	B := gomat.NewDense(bRows, bCols, initB)
	fmt.Printf(" B:\n%v\n\n", gomat.Formatted(B))

	invSoln.Mul(&inverseA, B)
	fmt.Printf(" Solution by GoNum inversion and B is:\n%.5g\n\n", gomat.Formatted(&invSoln))

	// Try LU stuff
	var lu gomat.LU
	var luSoln *gomat.Dense

	lu.Factorize(A)
	err = lu.SolveTo(luSoln, false, B)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from lu Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Soluton by gonum LU factorization is:\n %v\n\n", gomat.Formatted(luSoln))

	// try w/ QR stuff
	var qr gomat.QR
	var qrSoln *gomat.Dense
	qr.Factorize(A)
	err = qr.SolveTo(qrSoln, false, Bvec)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from qr Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Soluton by gonum QR factorization is:\n %v\n\n", gomat.Formatted(qrSoln))

	// Try Solve stuff
	var solvSoln *gomat.Dense
	err = solvSoln.Solve(A, B)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from Solve is %s.  Bye-bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Solution by gonum Solve is:\n %v\n\n", gomat.Formatted(solvSoln))

} // end gonummatTest

// -----------------------------------------------------------------------
//                              MAIN PROGRAM
// -----------------------------------------------------------------------

func main() {
	solveTest2()
	goNumMatTest()
}

func pause() { // written a long time ago, probably my first stab at this.
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" pausing ... hit <enter>")
	scanner.Scan()
	answer := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	ans := strings.TrimSpace(answer)
	ans = strings.ToUpper(ans)
	fmt.Println(ans)
}

func newPause() {
	fmt.Print(" pausing ... hit <enter>  x to stop ")
	var ans string
	fmt.Scanln(&ans)
	if strings.ToLower(ans) == "x" {
		os.Exit(1)
	}
}

// END MatTest.
