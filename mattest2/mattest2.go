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

	solveSoln := mat.Solve(A, B)
	gaussSoln := mat.GaussJ(A, B)

	fmt.Printf("The solution X to AX = B\n using Solve       and then      GaussJ are:\n")
	ss = mat.WriteZeroPair(solveSoln, gaussSoln, 3)
	printString(ss)
	fmt.Println()

	if mat.Equal(solveSoln, gaussSoln) {
		ctfmt.Printf(ct.Green, true, " The Solve and GaussJ methods returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and GaussJ methods DID NOT returned equal results.\n")
	}
	fmt.Println()

	if mat.Equal(solveSoln, X) {
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

	fmt.Printf("\n Will now use matrix inversion as a solution method.  Result is:\n")
	inverseA := mat.Invert(A)
	inverseSoln := mat.Mul(inverseA, B)
	ss = mat.WriteZero(inverseSoln, 3)
	printString(ss)

	solveInvert := mat.SolveInvert(A, B)

	if mat.Equal(solveSoln, inverseSoln) {
		ctfmt.Printf(ct.Green, true, " The Solve and matrix inversion methods returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and matrix inversion methods DID NOT returned equal results.\n")
	}

	if mat.Equal(solveInvert, inverseSoln) {
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

	initialVal := float64(misc.RandRange(1, 50))
	increment := float64(misc.RandRange(1, 50))

	initX := make([]float64, aCols)
	initX[0] = initialVal
	initX[1] = initialVal + increment
	initX[2] = initialVal + 2*increment

	X := gomat.NewDense(aRows, bCols, initX)

	// Now need to assign coefficients in matrix A
	initA := make([]float64, aRows*aCols) // 3 x 3 = 9, as of this writing.

	for i := range initA {
		initA[i] = float64(misc.RandRange(1, 20))
	}

	A := gomat.NewDense(aRows, aCols, initA)

	initB := make([]float64, bRows*bCols)
	// Now do the calculation to determine what the B column vector needs to be for this to work.
	// I have to get the proper way to access a matrix element, and use it in the loop below.
	for i := range initB {
		product := A[i][j] * X[j][0]
		initB[i] += product
		//fmt.Printf(" i=%d, j=%d, A[%d,%d] is %g, X[%d,0] is %g, product is %g, B[%d,0] is %g\n", i, j, i, j, A[i][j], i, X[j][0], product, i, B[i][0])
		//newPause()
	}

}

// -----------------------------------------------------------------------
//                              MAIN PROGRAM
// -----------------------------------------------------------------------

func main() {
	solveTest2()
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
