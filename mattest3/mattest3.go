package main // mattest3 from mattest2a from mattest2 from mattest.  Both test mat.  Duh!

import (
	"bufio"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	gonum "gonum.org/v1/gonum/mat"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"src/mat"
	"src/misc"
	"strconv"
	"strings"
	"time"
)

/**********************************************************)
  (*                                                      *)
  (*              Test of Matrices module                 *)
  (*                                                      *)
  (*  Programmer:         P. Moylan                       *)
  (*  Last edited:        15 August 1996                  *)
  (*  Status:             Working                         *)
  (*                                                      *)
  (********************************************************)
*/
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
19 Mar 22 -- Summary of what I've discovered today.
             The problem I was having w/ using the gonum matrix stuff was that I needed to initialize the solution result, as in
				qrSoln := mat.NewDense(bRows, bCols, nil)
             That left me w/ the formatting issue.  Turned out that the characters that are being output are not handled correctly by tcc.  Cmd does, and these are matrix symbols.
             I can clean the output by either using my clean string routine, or by converting to a mat.Matrix2D and outputting that.  I did not write the conversion routine to
             handle a VecDense type.  I could, but I won't bother now that I've figured it out.  I could either run the tests from cmd, or use cleanString before outputting them.
19 Mar 24 -- Now called mattest2a, and I'll increase X.
			 Last thing today I added was VecDense solution, to see if that also worked.  It does.
20 Mar 24 -- Will accept a param that will determine the matrix sizes, esp size of X.  I'll use the flag package for this.
21 Mar 24 -- Adding file output of A and B so that these can be read by solve.go
26 Mar 24 -- Enhancing the equality test.  And adding possibly negative numbers.
30 Mar 24 -- Added findMaxDiff for when the equality test fails.  Added lastAltered string, and added verboseFlag.
             And added call to mat.BelowSmallMakeZero() and belowTolMakeZero(), to be used as needed.
31 Mar 24 -- My first use of a type assertion in belowSmallMakeZero.  These are not called type checks, as type checks is something the compiler always does.
 8 Oct 24 -- Rewriting my use of a type assertion into a type switch.
 9 Oct 24 -- I'm going to add writing the A and B matrices to a file, and also the X solution vector.  I'm not going to use the multi-output because I want colorized screen output.
               I may need to write more routines for this.  I'll add routines to my mat package to take io.Writer, or maybe just add file writing routines.  I have to look into this.
				Debugging on leox is easier on the eyes because the matrix output symbols are displayed correctly.  On Windows, only when running on Win11 desktop on cmd does that.
                The mat-size-...txt file is meant to be fed into solve or solve2, as a debugging step for the solve routines.
                I want more info output now.
13 Oct 24 -- Changed code to match the change in the mat API.
*/

const lastAltered = "Oct 13, 2024"
const small = 1e-10
const outputName = "mattest3-output.txt"

var n int
var negFlag bool
var verboseFlag bool
var aRows int
var aCols int
var bRows int
var bCols int

func solveTest(fn string, outfilebuf *bufio.Writer) error {
	fmt.Printf("---------------------------------------------------------------------------")
	fmt.Printf(" my mat Test ---------------------------------------------------------------------------\n\n")
	s := "--------- my mat Test -----------------\n\n"
	_, err := outfilebuf.WriteString(s)
	if err != nil {
		return err
	}

	var A, B, X mat.Matrix2D

	A = mat.NewMatrix(aRows, aCols)
	B = mat.NewMatrix(bRows, bCols)
	X = mat.NewMatrix(bRows, bCols)

	fmt.Println("Solving linear algebraic equations of form AX = B, solve for X")

	// Give a value to the A matrix.
	// I want these values to be whole positive numbers.  I need to determine the coefficient matrix, A, and values for the column vector, B.

	// initialize X
	for i := range aRows {
		X[i][0] = float64(misc.RandRange(1, 50))
		if negFlag {
			X[i][0] -= float64(rand.N(50))
		}
	}
	_, err = outfilebuf.WriteString("Solution Vector X:\n")
	if err != nil {
		return err
	}
	ss := mat.Write(X, 3)
	for _, line := range ss {
		outfilebuf.WriteString(line)
	}
	_, err = outfilebuf.WriteString("\n\n")
	if err != nil {
		return err
	}

	// Now need to assign coefficients in matrix A
	for i := range A {
		for j := range A[i] {
			A[i][j] = float64(misc.RandRange(1, 40))
			if negFlag {
				A[i][j] -= float64(rand.N(40))
			}
		}
	}

	ss = mat.Write(A, 3)

	_, err = outfilebuf.WriteString("Coeffecient Matrix A:\n")
	if err != nil {
		return err
	}
	for _, line := range ss {
		outfilebuf.WriteString(line)
	}
	_, err = outfilebuf.WriteString("\n\n")
	if err != nil {
		return err
	}

	if verboseFlag {
		fmt.Printf(" Coefficient matrix A is:\n")
		printString(ss)
		fmt.Println()
	}

	// Now do the calculation to determine what the B column vector needs to be for this to work.
	for i := range A {
		for j := range A[i] {
			product := A[i][j] * X[j][0]
			B[i][0] += product
		}
	}

	_, err = outfilebuf.WriteString("RHS Vector B:\n")
	if err != nil {
		return err
	}
	ss = mat.Write(B, 4)
	for _, line := range ss {
		outfilebuf.WriteString(line)
	}
	_, err = outfilebuf.WriteString("\n\n")
	if err != nil {
		return err
	}

	fmt.Printf("\n Column vectors X and B are:\n")
	ss = mat.MakeZeroPair(X, B, 4, small)
	printString(ss)
	fmt.Println()
	fmt.Printf("\n\n")

	// Another way to find the B column vector is to just do A*X.  It works.
	newB := mat.Mul(A, X)
	fmt.Printf(" Column vector newB is:\n")
	mat.WriteZeroln(newB, 6, small)

	// Generate file to be read in by Solve(2).
	WriteMatrices(A, B, fn)

	solveSoln := mat.Solve(A, B)
	gaussSoln := mat.GaussJ(A, B)
	if negFlag {
		solveSoln = mat.BelowSmallMakeZero(solveSoln)
		gaussSoln = mat.BelowSmallMakeZero(gaussSoln)
	}

	fmt.Printf("The solution X to AX = B\n using Solve       and then      GaussJ are:\n")
	ss = mat.MakeZeroPair(solveSoln, gaussSoln, 3, small)
	printString(ss)
	fmt.Println()

	// Does AX - B = 0
	C := mat.Mul(A, solveSoln)
	D := mat.Sub(B, C)

	if mat.IsZeroApprox(D) {
		ctfmt.Printf(ct.Green, false, " As a check, AX-B is approx zero, and evaluates to:\n")
		mat.WriteZeroln(D, 3, small)
	} else {
		ctfmt.Printf(ct.Red, true, "AX-B is not approx zero, but is:\n")
	}
	ctfmt.Printf(ct.Yellow, false, "\n\n AX - B raw result:\n")
	mat.Writeln(D, 3)

	if verboseFlag {
		ctfmt.Printf(ct.Yellow, false, " After calling mat.WriteZeroln\n")
		mat.WriteZeroln(D, 3, small)
	}

	fmt.Printf("\n Will now use matrix inversion as a solution method.  Result is:\n")
	inverseA := mat.Invert(A)
	inverseSoln := mat.Mul(inverseA, B)

	if negFlag {
		inverseSoln = mat.BelowSmallMakeZero(inverseSoln)
	}

	if verboseFlag {
		mat.WriteZeroln(inverseSoln, 3, small)
	}

	solveInvert := mat.SolveInvert(A, B)
	if negFlag {
		solveInvert = mat.BelowSmallMakeZero(solveInvert)
	}

	if mat.EqualApprox(solveSoln, gaussSoln) {
		ctfmt.Printf(ct.Green, true, " The Solve and GaussJ methods returned approx equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and GaussJ methods DID NOT return approx equal results.\n")
		if mat.EqualApproximately(solveSoln, gaussSoln, mat.Small*10) {
			ctfmt.Printf(ct.Green, true, " Now the Solve and GaussJ methods returned approx equal results using Small*10 tolerance factor.\n")
		} else {
			f := findMaxDiff(solveSoln, gaussSoln)
			ctfmt.Printf(ct.Red, true, " The Solve and GaussJ methods DID NOT return approx equal results, even using Small*10 tol fac.  Diff=%.3g\n", f)
		}
	}
	fmt.Println()

	if mat.EqualApprox(solveSoln, X) {
		ctfmt.Printf(ct.Green, true, " The Solve and X column vector returned approx equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and X column vector DID NOT return approx equal results.\n")
		if mat.EqualApproximately(solveSoln, X, mat.Small*10) {
			ctfmt.Printf(ct.Green, true, " Now the Solve and X column vector returned approx equal results using Small*10 tolerance factor.\n")
		} else {
			f := findMaxDiff(solveSoln, X)
			ctfmt.Printf(ct.Red, true, " The Solve and X column vector DID NOT return equal results, even using Small*10 tol fac.  Diff=%.3g\n", f)
		}
	}
	fmt.Println()

	if mat.EqualApprox(solveSoln, inverseSoln) {
		ctfmt.Printf(ct.Green, true, " The Solve and matrix inversion methods returned approx equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The Solve and matrix inversion methods DID NOT returned approx equal results.\n")
		if mat.EqualApproximately(solveSoln, inverseSoln, mat.Small*10) {
			ctfmt.Printf(ct.Green, true, " Now the Solve and matrix inversion methods returned approx equal results using mat.Small*10 tolerance factor.\n")
		} else {
			f := findMaxDiff(solveSoln, inverseSoln)
			ctfmt.Printf(ct.Red, true, " The Solve and matrix inversion methods DID NOT return approx equal results, using mat.Small*10 tol fac.  Diff=%.3g\n", f)
		}
	}

	if mat.EqualApprox(solveInvert, inverseSoln) {
		ctfmt.Printf(ct.Green, true, " The SolveInvert and matrix inversion methods returned equal results.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " The SolveInvert and matrix inversion methods DID NOT returned equal results.\n")
	}

	fmt.Println()

	return nil
} // end SolveTest

func printString(s []string) {
	for _, line := range s {
		ctfmt.Print(ct.Yellow, true, line)
	}
}

func WriteMatrices(A, B mat.Matrix2D, name string) {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" ERROR from os.Getwd is %s.  Output file not written.\n", err)
		return
	}
	outputFile, err := os.CreateTemp(workingDir, name)
	if err != nil {
		fmt.Printf(" ERROR from os.CreateTemp is %s.  Output file not written.\n", err)
		return
	}
	defer outputFile.Close()

	//                                   fmt.Printf(" WriteMatrices outputFile is %s and %s\n", name, outputFile.Name())
	outputBuf := bufio.NewWriter(outputFile)
	defer outputBuf.Flush()
	//                                                                 fmt.Printf(" WriteMatrices outputBuf created.\n")

	for i := range A {
		for j := range A[i] { // write a row of A
			s := strconv.FormatFloat(A[i][j], 'g', 6, 64)
			//fmt.Printf(" value = %.6g; s = %s\n", A[i][j], s)
			_, err = outputBuf.WriteString(s)
			if err != nil {
				fmt.Printf(" ERROR from %s.WriteString(%s) is %s.  Aborting writing output file.\n", outputFile.Name(), s, err)
				return
			}
			_, err = outputBuf.WriteString("  ")
			if err != nil {
				fmt.Printf(" ERROR from %s.WriteString(%s) is %s.  Aborting writing output file.\n", outputFile.Name(), s, err)
				return
			}
		}
		s := strconv.FormatFloat(B[i][0], 'g', 6, 64)
		_, err = outputBuf.WriteString(s)
		if err != nil {
			fmt.Printf(" ERROR from %s.WriteString(%s) is %s.  Aborting writing output file.\n", outputFile.Name(), s, err)
			return
		}
		_, err = outputBuf.WriteRune('\n')
		if err != nil {
			fmt.Printf(" ERROR from %s.WriteString(%s) is %s.  Aborting writing output file.\n", outputFile.Name(), s, err)
			return
		}
	}
	//                                                                       fmt.Printf(" Finished writing matrices.\n")
}

func goNumMatTest(outputFileBuf *bufio.Writer) error {
	// Will look to solve AX = B, for X

	s := "----------- gonum Test ---------\n\n"
	_, err := outputFileBuf.WriteString(s)
	if err != nil {
		return err
	}

	fmt.Printf("---------------------------------------------------------------------------")
	fmt.Printf(" gonum Test ---------------------------------------------------------------------------\n\n")

	initX := make([]float64, aCols)
	for i := range aCols {
		initX[i] = float64(misc.RandRange(1, 50))
		if negFlag {
			initX[i] -= float64(rand.N(50))
		}
	}

	X := gonum.NewVecDense(bRows, initX)
	str := fmt.Sprintf("%.5g\n", gonum.Formatted(X, gonum.Squeeze()))
	strClean := cleanString(str)
	outputFileBuf.WriteString("Not cleaned X:\n")
	outputFileBuf.WriteString(str)
	outputFileBuf.WriteString("\nCleaned X:\n")
	_, err = outputFileBuf.WriteString(strClean)
	if err != nil {
		return err
	}
	outputFileBuf.WriteString("\n\n")

	if verboseFlag {
		fmt.Printf("not cleaned X=\n%s\n\n", str)
		fmt.Printf("cleaned X=\n%s\n\n", strClean)
	}

	// Now need to assign coefficients in matrix A
	initA := make([]float64, aRows*aCols)

	for i := range initA {
		initA[i] = float64(misc.RandRange(1, 20))
		if negFlag {
			initA[i] -= float64(rand.N(20))
		}
	}

	A := gonum.NewDense(aRows, aCols, initA)
	if verboseFlag {
		fmt.Printf(" A:\n%.5g\n", gonum.Formatted(A, gonum.Squeeze()))
		aMatrix := extractDense(A)
		mat.WriteZeroln(aMatrix, 5, small)
	}
	str = fmt.Sprintf("%.5g\n", gonum.Formatted(A, gonum.Squeeze()))
	strClean = cleanString(str)
	outputFileBuf.WriteString("Not cleaned A:\n")
	_, err = outputFileBuf.WriteString(str)
	if err != nil {
		return err
	}
	outputFileBuf.WriteString("\nCleaned A:\n")
	_, err = outputFileBuf.WriteString(strClean)
	if err != nil {
		return err
	}
	outputFileBuf.WriteString("\n\n")

	initB := make([]float64, bRows) // col vec
	for i := range aRows {
		for j := range aCols {
			product := A.At(i, j) * X.At(j, 0)
			initB[i] += product
		}
	}
	Bvec := gonum.NewVecDense(bRows, initB)

	str = fmt.Sprintf("%.5g\n", gonum.Formatted(Bvec, gonum.Squeeze()))
	strClean = cleanString(str)
	outputFileBuf.WriteString("Not cleaned B:\n")
	_, err = outputFileBuf.WriteString(str)
	if err != nil {
		return err
	}
	outputFileBuf.WriteString("\nCleaned B:\n")
	_, err = outputFileBuf.WriteString(strClean)
	if err != nil {
		return err
	}
	outputFileBuf.WriteString("\n\n")

	if verboseFlag {
		fmt.Printf(" Bvec:\n%.6g\n\n", gonum.Formatted(Bvec, gonum.Squeeze()))
	}

	// Will try w/ inversion
	var inverseA, invSoln, invSolnVec gonum.Dense
	err = inverseA.Inverse(A)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from inverting A: %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	invSolnVec.Mul(&inverseA, Bvec) // this works.  So far, it's the only method that does work.
	belowSmallMakeZero(&invSolnVec, small)
	if verboseFlag {
		fmt.Printf(" Solution by GoNum inversion and Bvec is (after calling belowSmallMakeZero on *Dense):\n%.5g\n\n", gonum.Formatted(&invSolnVec, gonum.Squeeze()))
	}

	B := gonum.NewDense(bRows, bCols, initB)
	if verboseFlag {
		fmt.Printf(" B:\n%.5g\n\n", gonum.Formatted(B, gonum.Squeeze()))
	}
	bMatrix := extractDense(B)
	if verboseFlag {
		mat.WriteZeroln(bMatrix, 4, small)
	}

	invSoln.Mul(&inverseA, B)
	belowSmallMakeZero(&invSoln, small)
	fmt.Printf(" Solution by GoNum inversion and B is (after calling belowSmallMakeZero on *Dense):\n%.5g\n\n", gonum.Formatted(&invSoln, gonum.Squeeze()))

	// Try LU stuff
	var lu gonum.LU
	luSoln := gonum.NewDense(bRows, bCols, nil)

	lu.Factorize(A)
	err = lu.SolveTo(luSoln, false, B)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from lu Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	belowSmallMakeZero(luSoln, small)
	if verboseFlag {
		fmt.Printf(" Soluton by gonum LU factorization is (after calling belowSmallMakeZero on *Dense):\n%.5g\n\n", gonum.Formatted(luSoln, gonum.Squeeze()))
	}

	// try w/ QR stuff
	var qr gonum.QR
	qrSoln := gonum.NewDense(bRows, bCols, nil)
	qr.Factorize(A)
	err = qr.SolveTo(qrSoln, false, Bvec)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from qr Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	belowSmallMakeZero(qrSoln, small)
	if verboseFlag {
		fmt.Printf(" Soluton by gonum QR factorization is (after calling belowSmallMakeZero on *Dense:\n%.5g\n\n", gonum.Formatted(qrSoln, gonum.Squeeze()))
	}

	// Try Solve stuff
	bR, bC := B.Dims()
	solvSoln := gonum.NewDense(bR, bC, nil) // just to see if this works.
	err = solvSoln.Solve(A, B)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from Solve is %s.  Bye-bye\n", err)
		os.Exit(1)
	}
	belowSmallMakeZero(solvSoln, small)
	if verboseFlag {
		fmt.Printf(" Solution by gonum Solve is (after calling belowSmallMakeZero on *Dense):\n%.5g\n\n", gonum.Formatted(solvSoln, gonum.Squeeze()))
	}

	// Try Vec Solve
	bRV, _ := Bvec.Dims() // just to see if this works.
	vecSolveSoln := gonum.NewVecDense(bRV, nil)
	err = vecSolveSoln.SolveVec(A, Bvec)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from VecSolve is %s.  Bye-bye\n", err)
		os.Exit(1)
	}
	belowSmallMakeZero(vecSolveSoln, small)
	if verboseFlag {
		fmt.Printf(" Solution by gonum VecSolve is (after calling belowSmallMakeZero on *VecDense):\n%.5g\n\n", gonum.Formatted(vecSolveSoln, gonum.Squeeze()))
	}

	if gonum.EqualApprox(X, &invSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and inversion solution are approx equal.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " X and inversion solution are not approx equal.\n")
	}
	if gonum.EqualApprox(X, luSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and LU solution are approx equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and LU solution are not approx equal.\n")
	}
	if gonum.EqualApprox(X, &invSolnVec, small) {
		ctfmt.Printf(ct.Green, false, " X and vector inversion solution are approx equal.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " X and vector inversion solution are not approx equal.\n")
	}
	if gonum.EqualApprox(X, qrSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and QR solution are approx equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and QR solution are not approx equal.\n")
	}
	if gonum.EqualApprox(X, solvSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and Solve solution are approx equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and Solve solution are not approx equal.\n")
	}
	if gonum.EqualApprox(X, vecSolveSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and Vec Solve solution are approx equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and Vec Solve solution are not approx equal.\n")
	}

	return nil

} // end goNumMatTest

func showRunes(s string) { // the unidentified runes turned out to be matrix symbols 0x23a1 .. 0x23a6, or 9121 .. 9126
	for _, r := range s {
		fmt.Printf(" %c, %x, %d, %s\n", r, r, r, string(r))
	}
}

func cleanString(s string) string {
	var sb strings.Builder

	for _, r := range s {
		if r < 128 {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func extractDense(m *gonum.Dense) [][]float64 {
	r, c := m.Dims()
	matrix := mat.NewMatrix(r, c)
	for i := range matrix { // different from in mattest2
		for j := range matrix[i] { // to see if this works, too.
			matrix[i][j] = m.At(i, j)
		}
	}
	return matrix
}

// -----------------------------------------------------------------------
//                              MAIN PROGRAM
// -----------------------------------------------------------------------

func main() {
	flag.IntVar(&n, "n", 3, "Size of X and other arrays.  Default is 3.")
	flag.BoolVar(&negFlag, "neg", false, "Allow creation of negative coefficients.")
	flag.BoolVar(&verboseFlag, "v", false, "Versose output flag.")

	flag.Parse()
	aRows = n
	aCols = aRows
	bRows = n
	bCols = 1

	fmt.Printf(" Linear Algebra Matrix Test routine 3.  Last altered %s, compiled w/ %s\n", lastAltered, runtime.Version())

	outFilename := "mat-" + strconv.Itoa(n) + "-*.txt"

	outputFile, outputFileBuf, err := misc.CreateOrAppendWithBuffer(outputName)
	if err != nil {
		fmt.Printf("Error creating output file %s: %s\n", outputName, err)
		return
	}
	outputFileBuf.WriteString("------------------------------------------------------------------------------------\n")
	nowStr := time.Now().Format(time.ANSIC)
	_, err = outputFileBuf.WriteString(nowStr)
	if err != nil {
		fmt.Printf("Error writing time to output file %s: %s\n", outputName, err)
		return
	}
	outputFileBuf.WriteRune('\n')
	outputFileBuf.WriteRune('\n')

	newPause()
	err = goNumMatTest(outputFileBuf)
	if err != nil {
		fmt.Printf("Error writing goNumMatTest to output file %s: %s\n", outputName, err)
	}
	pause()

	err = solveTest(outFilename, outputFileBuf)
	if err != nil {
		fmt.Printf("Error writing solveTest to output file %s: %s\n", outputName, err)
	}

	err = outputFileBuf.Flush()
	if err != nil {
		fmt.Printf("Error flushing output file %s: %s\n", outputName, err)
	}
	err = outputFile.Close()
	if err != nil {
		fmt.Printf("Error closing output file %s: %s\n", outputName, err)
	}

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

func findMaxDiff(a, b mat.Matrix2D) float64 {
	var maxVal float64
	if len(a) != len(b) {
		return 10 // this means the matrices are not the same size
	}
	if len(a[0]) != len(b[0]) {
		return 10 // this means the matrices are not the same size
	}
	for i := range a {
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) >= maxVal { // I can't compare floats using equal, it's too likely to fail due small differences in the numbers.
				maxVal = math.Abs(a[i][j] - b[i][j])
			}
		}
	}
	return maxVal
}

func belowSmallMakeZero(m gonum.Matrix, small float64) {
	//if matrx, ok := m.(*gonum.Dense); ok { // this is my first use of a type assertion.  I decided to rewrite it as a type switch, below.
	//	belowTolMakeZero(matrx, small)
	//} else if matrx, ok := m.(*gonum.VecDense); ok {
	//	belowTolMakeZeroVector(matrx, small)
	//} else {
	//	fmt.Printf(" Invalid type (%T) for use of belowSmallMakZero.  Skipped.\n", m)
	//}
	switch m := m.(type) { // this is my first use of a type switch
	case *gonum.Dense:
		belowTolMakeZero(m, small) // m is now an assigned type assertion so I don't need to use the type assertion in a switch case
		//belowTolMakeZero(m.(*gonum.Dense), small) // m is using an interface, so I have to define which concrete type to pass to the next function.
	case *gonum.VecDense:
		belowTolMakeZeroVector(m, small) // m is now an assigned type assertion so I don't need to use the type assertion in a switch case
		//belowTolMakeZeroVector(m.(*gonum.VecDense), small) // m is using an interface, so I have to define which concrete type to pass to the next function.
	default:
		fmt.Printf(" Invalid type (%T) for use of belowSmallMakZero.  Skipped.\n", m)
	}
}

func belowTolMakeZero(m *gonum.Dense, tol float64) {
	r, c := m.Dims()
	for i := range r {
		for j := range c {
			if math.Abs(m.At(i, j)) < tol {
				m.Set(i, j, 0)
			}
		}
	}
}

func belowTolMakeZeroVector(vec *gonum.VecDense, tol float64) {
	r, c := vec.Dims()
	for i := range r {
		for j := range c {
			if math.Abs(vec.At(i, j)) < tol {
				vec.SetVec(i, 0)
			}
		}
	}
}

// END MatTest3.
