package main

/********************************************************)
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
10 Mar 24 -- Added test of ScalarMult
15 Oct 24 -- Updated the code so that it would compile after I changed the API of my mat package.
*/

import (
	"bufio"
	"fmt"
	"os"
	"src/mat"
	"strings"
)

func BasicTest() {

	// Checks some simple matrix operations.

	const aRows = 2
	const aCols = 3
	const bRows = 3
	const bCols = 2

	var A, B, C, D, E, F [][]float64

	A = make([][]float64, aRows)

	B = make([][]float64, bRows)

	D = make([][]float64, aRows)
	E = make([][]float64, aRows)

	for i := range A {
		A[i] = make([]float64, aCols)
		D[i] = make([]float64, aCols)
		E[i] = make([]float64, aCols)
	}

	for i := range B {
		B[i] = make([]float64, bCols)
	}
	F = mat.New(aRows, aCols)        //  testing NewMatrix, not in original code
	G := mat.NewMatrix(bRows, bCols) // testing NewMatrix, not in original code
	fmt.Printf("ARows = %d, ACols = %d, bRows = %d and bCols = %d\n", aRows, aCols, bRows, bCols)
	fmt.Printf(" NewMatrix F has %d rows and %d columns.  NewMatrix G has %d rows and %d columns.\n", len(F), len(F[0]), len(G), len(G[0]))

	fmt.Println("Test of simple matrix operations.")
	fmt.Println()
	fmt.Println()

	//      Give a value to the A matrix.

	// A = mat.Random(A)
	mat.Random(A)
	fmt.Println(" Matrix A should be random numbers from 0..100, it is:")
	mat.Writeln(A, 5)
	//ss := mat.Write(A, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	//      Give a value to the B matrix.

	//B = mat.Random(B) // Random (B, Brows, Bcols);
	mat.Random(B) // Random (B, Brows, Bcols);
	fmt.Println(" Matrix B should be random numbers from 0..100, it is:")
	mat.Writeln(B, 5)
	//ss = mat.Write(B, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	// Test scalar multiply
	A5 := mat.ScalarMul(5, A)
	fmt.Printf(" 5 dot A, or scalar multiply of A, is:\n")
	mat.Writeln(A5, 4)
	//ss = mat.Write(A5, 3)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	//      Try an addition (it will fail).
	C = mat.Add(A, B)
	if C == nil {
		fmt.Println("We can't compute A+B because the dimensions don't match, as expected.")
	} else {
		fmt.Println(" Trying to add A+B, which should have failed.  It seems to have worked.  C is:")
		mat.Writeln(C, 5)
		//ss = mat.Write(C, 5)
		//for _, s := range ss {
		//	fmt.Print(s)
		//}
		fmt.Println()
	}

	// Try a multiplication (it should work).

	C = mat.Mul(A, B)
	fmt.Println("C = A*B is")
	mat.Writeln(C, 5)
	//ss = mat.Write(C, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	// Give a value to the D matrix.

	//D = mat.Random(D)
	mat.Random(D)
	fmt.Println("Matrix D should be random containing 0..100, it is")
	mat.Writeln(D, 5)
	//ss = mat.Write(D, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	// Try another addition (this one should work).

	E = mat.Add(A, D)
	fmt.Println("E = A+D works because their dimensions match.  Result is")
	mat.Writeln(E, 5)
	//ss = mat.Write(E, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	// My new test code
	F = mat.Add(D, E)
	fmt.Println(" F = D + E should succeed, which is same as F = D + A + D.")
	if F != nil {
		mat.Writeln(F, 5)
		//ss = mat.Write(F, 5)
		//for _, s := range ss {
		//	fmt.Print(s)
		//}
		fmt.Println()
	} else {
		fmt.Println(" F = D + E failed because the dimensions don't match.")
		//F = mat.Random(F)
		mat.Random(F)
	}

	G = mat.Sub(F, E) //   should fail
	fmt.Println(" G = F - E failed because the dimensions don't match, as expected.")
	if G != nil {
		//ss = mat.Write(G, 5)
		mat.Writeln(G, 5)
		//for _, s := range ss {
		//	fmt.Print(s)
		//}
		fmt.Println()
	} else {
		fmt.Printf(" E - F should have failed because the dimensions don't match, but it succeeded for some reason.  \n")
		//G = mat.Random(G)
		mat.Random(G)
		//ss = mat.Write(G, 4)
		fmt.Println(" Random G after E - F failed.")
		mat.Writeln(G, 4)
		//for _, s := range ss {
		//	fmt.Print(s)
		//}
		fmt.Println()
	}

	fmt.Println(" Matrix D is                     Matrix E:")
	mat.WriteZeroPairln(D, E, 4)
	//ss = mat.WriteZeroPair(D, E, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	//fmt.Println()

	newPause()

	fmt.Println(" Matrix E is:")
	//ss = mat.Write(E, 4)
	mat.Writeln(E, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}

	H := mat.Mul(D, B) // should work
	fmt.Println("H =  D*B:")

	if H != nil {
		ss := mat.Write(H, 6)
		for _, s := range ss {
			fmt.Print(s)
		}
		fmt.Println()
	} else {
		fmt.Println(" H = D*B did not work but should have.")
	}

	Q := mat.Sub(A, A)
	fmt.Println(" Q = A - A")
	if Q != nil {
		mat.Writeln(Q, 5)
		//ss = mat.Write(Q, 4)
		//for _, s := range ss {
		//	fmt.Print(s)
		//}
		fmt.Println()
	} else {
		fmt.Println(" Q = A - A did not work but should have.")
	}

	K := mat.NewMatrix(2, 2)
	//K = mat.Random(K)
	mat.Random(K)
	L := mat.NewMatrix(2, 2)
	mat.Random(L)
	fmt.Println()
	fmt.Println(" random K and then L, and then K*L")
	mat.Writeln(K, 6)
	mat.Writeln(L, 6)
	//ss = mat.Write(K, 6)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	//ss = mat.Write(L, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}

	L = mat.Mul(K, L)
	mat.Writeln(L, 6)

	//ss = mat.Write(L, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
} //    END BasicTest;

//************************************************************************

func SolveTest() {

	// Solution of a linear equation.

	const aRows = 4
	const aCols = 4
	const bRows = 4
	const bCols = 2

	var B, C, D, X [][]float64

	A := make([][]float64, aRows) // testing if create and assign works here.
	for i := range A {
		A[i] = make([]float64, aCols)
	}

	B = make([][]float64, bRows)
	C = make([][]float64, bRows)
	D = make([][]float64, bRows)
	X = make([][]float64, bRows)
	for i := range B {
		B[i] = make([]float64, bCols)
		C[i] = make([]float64, bCols)
		D[i] = make([]float64, bCols)
		X[i] = make([]float64, bCols)
	}

	fmt.Println("Solving linear algebraic equations")

	// Give a value to the A matrix.

	//A = mat.Random(A)
	mat.Random(A)
	fmt.Println("Matrix A is random:")
	ss := mat.Write(A, 4)
	for _, s := range ss {
		fmt.Print(s)
	}

	// Give a value to the B matrix.

	//B = mat.Random(B)
	mat.Random(B)
	fmt.Println("Matrix B is random:")
	ss = mat.Write(B, 4)
	for _, s := range ss {
		fmt.Print(s)
	}

	// Solve the equation A * X = B.

	X = mat.Solve(A, B)
	Y := mat.GaussJ(A, B)
	z := mat.SolveInvert(A, B)

	// Write the solution.

	fmt.Println("Using mat.Solve, the solution X to A * X = B is         Using mat.GaussJ:")
	mat.WriteZeroPairln(X, Y, 5)
	//ss = mat.Write(X, 4)
	//ss = mat.WriteZeroPair(X, Y, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}

	fmt.Println("Using mat.GaussJ, the solution X to A * X = B is:          Using mat.SolveInvert")
	mat.WriteZeroPairln(Y, z, 4)
	//ss = mat.Write(Y, 4)
	//ss = mat.WriteZeroPair(Y, z, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}

	// Check that the solution looks right from mat.Solve.

	C = mat.Mul(A, X)
	D = mat.Sub(B, C)
	d := mat.BelowSmallMakeZero(D)
	fmt.Println("As a check, A * X - B evaluates to zero from mat.Solve, before and after mat.BelowSmallMakeZero")
	mat.WriteZeroPairln(D, d, 4)
	//ss = mat.WriteZeroPair(D, d, 4) // Write (D, Brows, Bcols, 4);
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	//fmt.Println()
	//fmt.Println("As a check, AX-B evaluates to zero after running mat.BelowSmallMakeZero")
	//ss = mat.Write(D, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()

	// Check that the solution looks right from mat.GaussJ.

	C = mat.Mul(A, Y)
	D = mat.Sub(B, C)
	d = mat.Mul(A, z)
	e := mat.Sub(B, d)
	fmt.Println("As a check, AX-B evaluates to zero from mat.GaussJ and mat.SolveInvertafter.")
	mat.WriteZeroPairln(D, e, 5)
	//ss = mat.WriteZeroPair(D, e, 5)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()
} //    END SolveTest;

// -------------------------------------------------------

func SingularTest() {

	// Linear equation with singular coefficient matrix.

	const aRows = 2
	const aCols = 2
	const bRows = 2
	const bCols = 1

	A := mat.NewMatrix(aRows, aCols)
	B := mat.NewMatrix(bRows, bCols)
	X := mat.NewMatrix(bRows, bCols)

	if A == nil || B == nil || X == nil {
		fmt.Println(" Singular test failed in that a matrix came back nil from NewMatrix call.")
		return
	}

	fmt.Println("A singularity problem, which can't be solved.")

	// Give a value to the A matrix.

	A[0][0] = 1.0
	A[0][1] = 2.0
	A[1][0] = 2.0
	A[1][1] = 4.0
	fmt.Println("Matrix A is not random; it is singular :")
	ss := mat.Write(A, 4)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	// Give a value to the B matrix.

	//B = mat.Random(B)
	mat.Random(B)
	fmt.Println("Matrix B is random:")
	ss = mat.Write(B, 4)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	// Try to solve the equation AX = B.

	X = mat.Solve(A, B)

	if X == nil { // it should be nil, as A is singular
		fmt.Println("The equation AX = B could not be solved, which is correct.")
	}

} //    END SingularTest;

// ------------------------------------------------------------ InversionTest ------------------------

func InversionTest() {

	// Inverting a matrix, also an eigenvalue calculation.

	const N = 5

	A := mat.NewMatrix(N, N)

	fmt.Println("INVERTING A SQUARE MATRIX")

	// Give a random value to the A matrix.

	//A = mat.Random(A) // Random (A, N, N);
	mat.Random(A) // Random (A, N, N);
	fmt.Println(" Random Matrix A is")
	ss := mat.Write(A, 4)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	// Invert it.

	X := mat.Invert(A) //  X = mat.Invert(A, N);

	// Write the solution.

	fmt.Println()
	fmt.Println("The inverse of A is")
	ss = mat.Write(X, 4)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	// Check that the solution looks right.

	B := mat.Mul(A, X) // Mul(A, X, N, N, N, B);
	fmt.Println()
	fmt.Println("As a check, the product evaluates to the identity matrix")
	ss = mat.Write(B, 4)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("As a check, the product evaluates to the identity matrix after using mat.WriteZero.")
	mat.WriteZeroln(B, 4, mat.Small)
	//ss = mat.WriteZero(B, 4)
	//for _, s := range ss {
	//	fmt.Print(s)
	//}
	fmt.Println()
	fmt.Println()

	pause()

	fmt.Println()
	fmt.Println("EIGENVALUES")
	fmt.Println()
	fmt.Println("The eigenvalues of A are")
	W := mat.Eigenvalues(A)
	for j := range W {
		fmt.Print("    ")
		fmt.Print(W[j])
		fmt.Println()
	}
	fmt.Println()
	for _, w := range W { // just to see if this also works
		fmt.Printf("  %5G\n", w)
	}
	fmt.Println()

	fmt.Println("The eigenvalues of its inverse are")
	W = mat.Eigenvalues(X)
	for _, w := range W {
		fmt.Printf("  %5G\n", w)
	}
	fmt.Println()

} //    END InversionTest;

// -----------------------------------------------------------------------
//                              MAIN PROGRAM
// -----------------------------------------------------------------------

func main() {
	BasicTest()
	pause()
	SolveTest()
	newPause()
	//SingularTest()
	//newPause()
	InversionTest()
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
	fmt.Print(" pausing ... hit <enter>")
	var ans string
	fmt.Scanln(&ans)
}

// END MatTest.
