package main

/*
MODULE Solve;

  REVISION
  --------
   2 Mar 05 -- Added prompts to remind me of the file format.
   3 Mar 05 -- Made version 2 write lines like eqn w/o =.
   4 Mar 05 -- Don't need N as 1st line now.
  26 Feb 06 -- Will reject non-numeric entries and allows <tab> as delim.
  24 Dec 16 -- Converted to Go.  Ignores non-numeric entries to allow for comments.  First non-numeric entry skips rest of line.
  31 Jul 20 -- Added gonum.org/mat code, and now called gonumsolve.go.  I learned a few things about pointers.  More details below.
*/

import (
	"bufio"
	"fmt"
	"getcommandline"
	"gonum.org/v1/gonum/mat" // from memory.  This needs to be confirmed.
	"io"
	mymat "mat"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tokenize"
)

const LastCompiled = "July 31, 2020"
const MaxN = 30
const toleranceFactor = 1e-6

// IM (inputMatrix) is not square because the B column vector is in last column of IM.
//                                        was TYPE MaxRealArray = ARRAY [1..MaxN],[1..MaxN+1] OF LONGREAL; in M-2 code.

// type Matrix2D [][]float64;  not needed here.  It's internal to mymat package.

func main() {
	/* ------------------------------------------------------------------------
	                           MAIN PROGRAM
	------------------------------------------------------------------------*/

	fmt.Println(" Equation solver written in Go.  Last compiled ", LastCompiled)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: solve <filename>")
		fmt.Println(" Solves vector equation A * X = B; and A is a square coef matrix.")
		fmt.Println(" N is determined by number of rows and B value is last on each line.")
		os.Exit(0)
	}

	commandline := getcommandline.GetCommandLineString()
	cleancommandline := filepath.Clean(commandline)
	fmt.Println(" filename on command line is ", cleancommandline)

	//  cleancommandline = "xy1.txt";

	infile, err := os.Open(cleancommandline)
	if err != nil {
		fmt.Println(" Cannot open input file.  Does it exist?  Error is ", err)
		os.Exit(1)
	}

	defer infile.Close()
	scanner := bufio.NewScanner(infile)

	IM := mymat.NewMatrix(MaxN, MaxN+1) // IM is input matrix
	IM = mymat.Zero(IM)

	lines := 0
CountLinesLoop:
	for { // read, count and process lines
		for n := 0; n < MaxN; n++ {
			readSuccess := scanner.Scan()
			if readSuccess {
				// Non-numeric tokens are comments, and lines not beginning with a number are comment lines.
			} else {
				break CountLinesLoop // break outer loop
			} // if readSuccess

			inputline := scanner.Text()
			if readErr := scanner.Err(); readErr != nil {
				if readErr == io.EOF {
					break CountLinesLoop // break outer loop
				} else { // this may be redundant because of the readSuccess test
					break CountLinesLoop // break outer loop
				}
			}

			tokenize.INITKN(inputline)
			col := 0
			EOL := false
			for (EOL == false) && (n <= MaxN) { // read numbers into IM input matrix.
				token, EOL := tokenize.GETTKNREAL()
				if EOL {
					break
				}
				if token.State != tokenize.DGT {
					break
				} // treat remainder of line as a comment
				IM[lines][col] = token.Rsum // remember that IM is Input Matrix
				col++
			}
			if col > 0 { // allow comments by incrementing the pointer only when have numbers on the line.
				lines++
			}
		} // END for n
	} // END outer reading loop

	N := lines // Note: lines is 0 origin, and therefore so is N

	// Now need to create A and B matrices, using my own code.

	A := mymat.NewMatrix(N, N)
	B := mymat.NewMatrix(N, 1)
	for row := range A {
		for col := range A[0] {
			A[row][col] = IM[row][col]
		} // END FOR col
		B[row][0] = IM[row][N] // I have to keep remembering that [0,0] is the first row and col.
	} // END FOR row

	fmt.Println(" coef matrix A is:")
	ss := mymat.Write(A, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	fmt.Println(" Right hand side vector matrix B is:")
	ss = mymat.Write(B, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	ans := mymat.NewMatrix(N, N)
	ans = mymat.Solve(A, B) // Solve (ra1, ra2, ans, N, 1);
	fmt.Println("The solution X to AX = B using Solve is")
	ss = mymat.Write(ans, 5)
	for _, s := range ss {
		fmt.Print(s)
	}

	ans2 := mymat.NewMatrix(N, N)
	ans2 = mymat.GaussJ(A, B) // Solve (ra1, ra2, ans, N, 1);
	fmt.Println("The solution X to AX = B using GaussJ is")
	ss = mymat.Write(ans2, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	pause()

	// Check that the solution looks right.

	C := mymat.NewMatrix(N, 1)
	D := mymat.NewMatrix(N, 1)
	C = mymat.Mul(A, ans)
	D = mymat.Sub(B, C)

	fmt.Println("As a check, AX-B should be 0, and evaluates to")
	ss = mymat.Write(D, 5)
	for _, s := range ss {
		fmt.Print(s)
	}

	D = mymat.BelowSmallMakeZero(D)

	fmt.Println("As a check, AX-B should be all zeros after calling BelowSmall.  It evaluates to")
	ss = mymat.Write(D, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()
	fmt.Println()

	pause()

	// Now for gonum.org mat code
	// Some of the routines take a pointer, some do not.  mat.NewDense makes a pointer.  var mat.Dense does not.  So some
	// of the routines needed the & adrof operator, and in one case, I had to dereference the pointer for it to work.
	// This IDE helped me to sort this out by suggesting fixes to these errors.

	fmt.Println(" Using gonum.org Solve")
	newA := mat.NewDense(N, N, nil)
	newB := mat.NewDense(N, 1, nil)
	for row := 0; row < N; row++ {
		for col := 0; col < N; col++ {
			newA.Set(row, col, A[row][col])
		} // END FOR col
		newB.Set(row, 0, B[row][0])
	} // END FOR row

	var newX mat.Dense // construct empty matrix
	mat.NewDense(N, N, nil)
	er := newX.Solve(newA, newB)
	if er != nil {
		fmt.Println(" Error from newX.Solve is", er)
	}
	ss = write(newX, 5)
	fmt.Println(" write newX:")
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	ss = writeZero(newX, 5)
	fmt.Println(" writeZero newX:")
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	fmt.Println()

	pause()

	// Now to make B a vector (well, a column vector, but still a vector) and then use vector routines.  Just to see if I can get it to work, too.

	vecB := mat.NewVecDense(N, nil)
	for i := 0; i < N; i++ {
		vecB.SetVec(i, newB.At(i, 0))
	}

	var vecX mat.VecDense
	er = vecX.SolveVec(newA, vecB)
	if er != nil {
		fmt.Println(" Error from vecX.Solve is", er)
	}

	fmt.Println(" B and X are now vectors instead of matricies")
	ss = vectorWrite(*vecB, 5)
	fmt.Println(" B: ")
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()
	fmt.Println()

	ss = vectorWrite(vecX, 5)
	fmt.Println(" X: ")
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()
	fmt.Println()

	var inverseA mat.Dense
	err = inverseA.Inverse(newA) // I think this method has a pointer receiver, but syntatic sugar does not require me to dereference inverseA
	// It compiles fine if I declare var inverseA *matDense and then remove the adrof operator in the call to Mul.
	if err != nil {
		fmt.Println(" Error from inverseA.Inverse is", err)
	}
	var newX2 mat.Dense // empty matrix is constructed
	newX2.Mul(&inverseA, newB)
	fmt.Println(" Using inverse of A to solve for X:")
	ss = write(newX2, 5)
	for _, s := range ss {
		fmt.Print(s)
	}

	fmt.Println()
	fmt.Println()

	pause()

	fmt.Println(" Do the gonum.org computations look right?  First newA * newX - newB by solve, then by inverse")

	var newE, newF mat.Dense
	newE.Mul(newA, &newX) // compute newE = newA * newX
	newE.Sub(newB, &newE) // compute newA * newX - newB which should be all zeros
	fmt.Println(" newA * newX - newB should be all zeros")
	ss = write(newE, 5)
	for _, s := range ss {
		fmt.Print(s)
	}

	fmt.Println(" should still be all zeros")
	ss = writeZero(newE, 5)
	for _, s := range ss {
		fmt.Print(s)
	}

	fmt.Println()
	fmt.Println()

	pause()

	newF.Mul(newA, &newX2) // compute newF = NewA * newX2
	newF.Sub(newB, &newF)  // compute newA * newX2 - newB which should also be all zeros
	ss = write(newF, 5)
	fmt.Println(" newA * newX2 - newB should be matrix of all zeros")
	for _, s := range ss {
		fmt.Print(s)
	}

	ss = writeZero(newF, 5)
	fmt.Println(" should still be matrix of all zeros")
	for _, s := range ss {
		fmt.Print(s)
	}

	fmt.Println()
	fmt.Println()

	pause()

	fmt.Println(" newA * vecX - vecB should be all zeros")
	var checkVec mat.VecDense
	checkVec.MulVec(newA, &vecX)
	checkVec.SubVec(vecB, &checkVec)

	for _, s := range ss {
		fmt.Print(s)
	}

	fmt.Println()
	fmt.Println()

} // END Solve.

func pause() {
	scnr := bufio.NewScanner(os.Stdin)
	fmt.Print(" pausing ... hit <enter>")
	scnr.Scan()
	answer := scnr.Text()
	if err := scnr.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	ans := strings.TrimSpace(answer)
	ans = strings.ToUpper(ans)
	fmt.Println(ans)
}

func write(M mat.Dense, places int) []string {

	// Writes the r x c matrix M to a string slice which is returned.  Each string represents a field "places" characters wide.

	rows, cols := M.Dims()
	OutputStringSlice := make([]string, 0, rows*cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			ss := strconv.FormatFloat(M.At(i, j), 'G', places, 64)
			OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("%10s", ss))
		} // END FOR j
		OutputStringSlice = append(OutputStringSlice, "\n")
	} // END FOR i
	OutputStringSlice = append(OutputStringSlice, "\n")
	return OutputStringSlice
} // END write

func writeZero(M mat.Dense, places int) []string {

	// Writes the r x c matrix M to a string slice which is returned.  Each string represents a field "places" characters wide.

	rows, cols := M.Dims()
	OutputStringSlice := make([]string, 0, rows*cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			v := M.At(i, j)
			ss := ""
			if math.Abs(v) < toleranceFactor {
				ss = "0    "
			} else {
				ss = strconv.FormatFloat(M.At(i, j), 'G', places, 64)
			}
			OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("%10s", ss))
		} // END FOR j
		OutputStringSlice = append(OutputStringSlice, "\n")
	} // END FOR i
	OutputStringSlice = append(OutputStringSlice, "\n")
	return OutputStringSlice
} // END write

func vectorWrite(M mat.VecDense, places int) []string {
	rows, cols := M.Dims()
	OutputStringSlice := make([]string, 0, rows*cols)

	for i := 0; i < rows; i++ {
		ss := strconv.FormatFloat(M.At(i, 0), 'G', places, 64)
		OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("%10s", ss))
	} // END FOR i
	OutputStringSlice = append(OutputStringSlice, "\n")
	return OutputStringSlice
} // END write
