package main // solve2.go, from solve.go

/*
  REVISION
  --------
   2 Mar 05 -- Added prompts to remind me of the file format.
   3 Mar 05 -- Made version 2 write lines like eqn w/o =.
   4 Mar 05 -- Don't need N as 1st line now.
  26 Feb 06 -- Will reject non-numeric entries and allows <tab> as delim.
  24 Dec 16 -- Converted to Go.
  13 Feb 21 -- Updated to modules.  And added filePicker and flag package.
  21 Mar 24 -- Adding use of gonum routines.  And removing min procedure as that's part of the std lib as of Go 1.22.
  23 Mar 24 -- Increased MaxN
  26 Mar 24 -- Added checks on input matrix size, so it won't panic.
  27 Mar 24 -- Now called Solve2.  I intend to build the IM by appending to a slice so I don't need a maxN size.  And I won't display the matrix symbols on Windows.
               I'm amazed that this worked the first time.  I based the code on cal2 and cal3, and this seemed to have worked.  Wow!
  28 Mar 24 -- Adding AX -B = 0 to the gonum.org part.
*/

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	gomat "gonum.org/v1/gonum/mat"
	"io"
	"os"
	"runtime"
	"src/filepicker"
	"src/mat"
	"src/tknptr"
	"strconv"
	"strings"
)

const LastAltered = "28 Mar 2024"
const small = 1e-10

type rows []float64

var verboseFlag = flag.Bool("v", false, "Verbose mode.")
var onWin = runtime.GOOS == "windows"

//                          InputMatrix (IM) is not square because the B column vector is in last column of IM
//                          type Matrix2D [][]float64  It is defined in and used by mat.

func makeDense(matrix mat.Matrix2D) *gomat.Dense {
	var idx int
	r := len(matrix)
	c := len(matrix[0])
	initDense := make([]float64, r*c)
	for i := range matrix {
		for j := range matrix[i] {
			initDense[idx] = matrix[i][j]
			idx++
		}
	}
	dense := gomat.NewDense(r, c, initDense)
	return dense
}

func makeDense2(matrix mat.Matrix2D) *gomat.Dense {
	// Just to see if this works too.  It does.
	var idx int
	r := len(matrix)
	c := len(matrix[0])
	dense := gomat.NewDense(r, c, nil)
	for i := range matrix {
		for j := range matrix[i] {
			dense.Set(i, j, matrix[i][j])
			idx++
		}
	}

	return dense
}

func outputDense(m *gomat.Dense) {
	s := fmt.Sprintf("%.5g\n", gomat.Formatted(m, gomat.Squeeze()))
	if onWin {
		s = cleanString(s)
	}
	fmt.Printf("%s", s)
	fmt.Println()
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

/************************************************************************
 *                              MAIN PROGRAM                            *
 ************************************************************************/

func main() {
	fmt.Printf(" Equation solver v2 written in Go.  Last altered %s on %s, last altered mat on %s, compiled with %s\n",
		os.Args[0], LastAltered, mat.LastAltered, runtime.Version())
	fmt.Println()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Solves vector equation A*X = B; A is a square coef matrix.")
		fmt.Fprintf(flag.CommandLine.Output(), " Input text file has each row being the coefficients.")
		fmt.Fprintf(flag.CommandLine.Output(), " N is determined by number of rows and B value is last on each line.")

		flag.PrintDefaults()
	}
	flag.Parse()

	var filename string
	if flag.NArg() == 0 {
		filenames, err := filepicker.GetFilenames("*") // Not sure what the default ext should be.  For now, any file is allowed.
		if err != nil {
			fmt.Fprintf(os.Stderr, " filepicker returned error %v\n.  Exiting.", err)
			os.Exit(1)
		}
		if len(filenames) == 0 {
			fmt.Fprintln(os.Stderr, " No filenames found that match pattern.  Exiting")
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ { // goes 0 .. 25, or a .. z
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		var ans string
		fmt.Print(" Enter filename choice : ")
		_, err = fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil {
			ans = "0"
		}
		i, er := strconv.Atoi(ans)
		if er == nil {
			filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			filename = filenames[i]
		}
		if *verboseFlag {
			fmt.Println(" Picked filename is", filename)
		}
	} else {
		filename = flag.Arg(0)
		if *verboseFlag {
			fmt.Println(" filename on command line is ", filename)
		}
	}

	infile, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, " %s does not exit.  Exiting.", filename)
		} else {
			fmt.Fprintf(os.Stderr, " Error while opening %s is %v.  Exiting.\n ", filename, err)
		}
		os.Exit(1)
	}

	defer infile.Close()
	if *verboseFlag {
		fmt.Printf(" Opened filename is %s\n", infile.Name())
	}
	scanner := bufio.NewScanner(infile)

	IM := make([]rows, 0, 20)

	for { // read and process all the lines in the input file.
		readSuccess := scanner.Scan()
		if !readSuccess {
			if readErr := scanner.Err(); readErr != nil {
				if *verboseFlag {
					fmt.Printf(" readErr is %s, len(IM) = %d\n", readErr, len(IM))
				}
				if readErr == io.EOF {
					break
				} else { // this may be redundant because of the readSuccess test
					ctfmt.Printf(ct.Red, true, " ERROR while reading from %s at line %d is %s.\n", filename, len(IM), readErr)
					break
				}
			}
			break
		}
		inputLine := scanner.Text()

		token := tknptr.TokenRealSlice(inputLine)
		if token[0].State != tknptr.DGT { // treat this as a comment line if it doesn't begin w/ a number.
			continue
		}

		row := make(rows, 0, 20)
		// append all numbers to a row
		for _, t := range token {
			if t.State == tknptr.DGT { // ignore non numerical tokens on an individual line
				row = append(row, t.Rsum)
			}
		}
		IM = append(IM, row)
		if *verboseFlag {
			fmt.Printf(" at bottom of line reading loop.  lines so far = %d, len(row) = %d, len(token) = %d\n", len(IM), len(row), len(token))
		}
	} // END file reading loop, ie, all lines in the file are to have been read by now.
	N := len(IM) // Note: lines is 0 origin, but N is length.

	if *verboseFlag {
		fmt.Printf(" Read in all lines in %s file.  Number of lines read is %d\n", filename, N)
	}

	// Now need to create A and B matrices

	A := mat.NewMatrix(N, N)
	B := mat.NewMatrix(N, 1)
	for row := range A {
		for col := range A[row] {
			A[row][col] = IM[row][col]
		}
		B[row][0] = IM[row][N]
	}

	fmt.Println(" coef matrix A is:")
	mat.Writeln(A, 5)

	fmt.Println(" Right hand side vector matrix B is:")
	mat.Writeln(B, 5)
	fmt.Println()

	X := mat.Solve(A, B)
	fmt.Println("The solution X to AX = B using Solve is")
	mat.Writeln(X, 5)

	ans2 := mat.GaussJ(A, B)
	fmt.Println("The solution X to AX = B using GaussJ is")
	mat.Writeln(ans2, 5)
	fmt.Println()

	// Check that the solution looks right.

	C := mat.Mul(A, X)
	D := mat.Sub(B, C)

	fmt.Println("As a check, AX-B should be 0, and evaluates to")
	mat.Writeln(D, 5)

	if mat.IsZeroApprox(D) {
		ctfmt.Printf(ct.Green, false, " Result of AX-B is approx zero.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " Result of AX-B is NOT approx zero.\n")
	}
	fmt.Println("After calling WriteZeroln:")
	mat.WriteZeroln(D, 5)
	fmt.Println()
	fmt.Println()

	newPause()

	// New for the gonum.org code.

	fmt.Printf("---------------------------------------------------------------------------")
	fmt.Printf(" gonum Test ---------------------------------------------------------------------------\n\n")

	denseA := makeDense(A)
	denseB := makeDense(B)
	denseX := makeDense(X) // used below for validation checks.
	fmt.Printf("A:\n")
	outputDense(denseA)
	fmt.Println()
	fmt.Printf("B:\n")
	outputDense(denseB)

	// Will try w/ inversion
	var inverseA, invSoln gomat.Dense
	err = inverseA.Inverse(denseA)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from inverting A: %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	invSoln.Mul(&inverseA, denseB)
	fmt.Printf(" Solution by GoNum inversion and B is:\n")
	outputDense(&invSoln)

	// Try LU stuff
	var lu gomat.LU
	luSoln := gomat.NewDense(N, 1, nil)
	lu.Factorize(denseA)
	err = lu.SolveTo(luSoln, false, denseB)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from lu Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Soluton by gonum LU factorization is:\n")
	outputDense(luSoln)

	// try w/ QR stuff
	var qr gomat.QR
	qrSoln := gomat.NewDense(N, 1, nil)
	qr.Factorize(denseA)
	err = qr.SolveTo(qrSoln, false, denseB)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from qr Solve To is %s.  Bye-Bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Soluton by gonum QR factorization is:\n")
	outputDense(qrSoln)

	// Try Solve stuff
	solvSoln := gomat.NewDense(N, 1, nil) // just to see if this works.
	err = solvSoln.Solve(denseA, denseB)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from Solve is %s.  Bye-bye\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Solution by gonum Solve is:\n")
	outputDense(solvSoln)

	if gomat.EqualApprox(denseX, &invSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and inversion solution are equal.\n")
	} else {
		ctfmt.Printf(ct.Red, true, " X and inversion solution are not equal.\n")
	}
	if gomat.EqualApprox(denseX, luSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and LU solution are equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and LU solution are not equal.\n")
	}
	if gomat.EqualApprox(denseX, qrSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and QR solution are equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and QR solution are not equal.\n")
	}
	if gomat.EqualApprox(denseX, solvSoln, small) {
		ctfmt.Printf(ct.Green, false, " X and Solve solution are equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " X and Solve solution are not equal.\n")
	}

	denseA2 := makeDense2(A)
	denseB2 := makeDense2(B)
	denseX2 := makeDense2(X) // used below for validation checks.
	if gomat.Equal(denseX, denseX2) && gomat.Equal(denseA, denseA2) && gomat.Equal(denseB2, denseB) {
		ctfmt.Printf(ct.Green, false, " makeDense and makeDense2 matrices are exactly equal.\n")
	} else {
		ctfmt.Printf(ct.Red, false, " makeDense and makeDense2 matrices are NOT exactly equal.\n")
	}

	rA, _ := denseA.Dims()
	_, cB := denseB.Dims()
	shouldBeZeroMatrix := gomat.NewDense(rA, cB, nil)
	intermResult := gomat.NewDense(rA, cB, nil)
	intermResult.Mul(denseA, denseX)
	shouldBeZeroMatrix.Sub(intermResult, denseB)
	fmt.Printf("\n AX - B should be Zero matrix:\n")
	outputDense(shouldBeZeroMatrix)
	fmt.Println()
	allZeros := gomat.NewDense(rA, cB, nil)
	allZeros.Zero()
	if gomat.EqualApprox(shouldBeZeroMatrix, allZeros, small) {
		ctfmt.Printf(ct.Green, false, " shouldbeZeroMatrix and allZeros matrix are approximately equal.\n\n")
	} else {
		ctfmt.Printf(ct.Red, true, " shouldbeZeroMatrix and allZeros matrices are NOT approximately equal.\n")
	}

} // END Solve.

func newPause() {
	fmt.Print(" pausing ... hit <enter>  x to stop ")
	var ans string
	fmt.Scanln(&ans)
	if strings.ToLower(ans) == "x" {
		os.Exit(1)
	}
}
