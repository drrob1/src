package main

/*
MODULE Solve;

  REVISION
  --------
   2 Mar 05 -- Added prompts to remind me of the file format.
   3 Mar 05 -- Made version 2 write lines like eqn w/o =.
   4 Mar 05 -- Don't need N as 1st line now.
  26 Feb 06 -- Will reject non-numeric entries and allows <tab> as delim.
  24 Dec 16 -- Converted to Go.
  13 Feb 21 -- Updated to modules.  And added filepicker and flag package.
*/

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"src/filepicker"
	"src/mat"
	"src/tokenize"
	"strconv"
	"strings"
)

const LastCompiled = "21 Nov 22"
const MaxN = 9

//                          MaxRealArray is not square because the B column vector is in last column of IM
//                                             TYPE MaxRealArray = ARRAY [1..MaxN],[1..MaxN+1] OF LONGREAL;

//type Matrix2D [][]float64  Not used here.  But it is defined in and used by mat.

var verboseFlag = flag.Bool("v", false, "Verbose mode.") // linter says that this is unused.  I'll have to use it, I guess.

func main() {
	/************************************************************************)
	  (*                              MAIN PROGRAM                            *)
	  (************************************************************************/

	fmt.Println(" Equation solver written in Go.  Last compiled ", LastCompiled)
	fmt.Println()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastCompiled, runtime.Version())
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
		fmt.Println(" Picked filename is", filename)
	} else {
		filename = flag.Arg(0)
		fmt.Println(" filename on command line is ", filename)
	}

	//  cleancommandline = "xy1.txt";

	infile, err := os.Open(filename)
	if err != nil {
		if err == os.ErrNotExist {
			fmt.Fprintf(os.Stderr, " %s does not exit.  Exiting.", filename)
		} else {
			fmt.Fprintf(os.Stderr, " Error while opening %s is %v.  Exiting.\n ", filename, err)
		}
		os.Exit(1)
	}

	defer infile.Close()
	scanner := bufio.NewScanner(infile)

	IM := mat.NewMatrix(MaxN, MaxN+1) // IM is input matrix
	IM = mat.Zero(IM)

	lines := 0
CountLinesLoop:
	for { // read, count and process lines
		for n := 0; n < MaxN; n++ { // WHILE N < MaxN DO
			readSuccess := scanner.Scan() //   FRDTXLN(InFile,inputbuf,80,bool);
			if readSuccess {
				// do nothing for now,  I thought N=n made sense until I saw the need to not process short
				// lines, assuming that they are a comment line.  And lines without numbers are comment lines.
			} else {
				break CountLinesLoop
			} // if readSuccess
			inputline := scanner.Text()
			if readErr := scanner.Err(); readErr != nil {
				if readErr == io.EOF {
					break CountLinesLoop
				} else { // this may be redundant because of the readSuccess test
					break CountLinesLoop
				}
			}

			tokenize.INITKN(inputline)
			col := 0
			var EOL bool
			var token tokenize.TokenType
			for !EOL && (n <= MaxN) { // linter says to not do (EOL == false), but to change it to what's there now.
				token, EOL = tokenize.GETTKNREAL() // if I use the gopher operator here, then EOL gets shadowed and is not the variable evaluated in the for condition.
				if EOL {
					break
				}
				if token.State == tokenize.DGT {
					IM[lines][col] = token.Rsum // remember that IM is Input Matrix
					col++
				} // ENDIF token.state=DGT
			} //  UNTIL (EOL is true) OR (col > MaxN);
			if col > 0 { // text only or null lines do not increment the row counter.
				lines++
			}
		} // END for n
	} // END reading loop
	N := lines // Note: lines is 0 origin

	// Now need to create A and B matrices

	A := mat.NewMatrix(N, N) // ra1 in Modula-2 code, ie, square matrix of coefficients to solve
	B := mat.NewMatrix(N, 1) // ra2 in Modula-2 code, ie, a column vector of coefficients on the RHS of each line.
	for row := range A {     // FOR row :=  1 TO N DO
		for col := range A[0] { //   FOR col := 1 TO N DO
			A[row][col] = IM[row][col]
		} // END FOR col
		B[row][0] = IM[row][N] // I have to keep remembering that [0,0] is the first row and col.
	} // END FOR row

	fmt.Println(" coef matrix A is:")
	ss := mat.Write(A, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	fmt.Println(" Right hand side vector matrix B is:")
	ss = mat.Write(B, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	//ans := mat.NewMatrix(N, N)  I don't have to do this.  Obviously, I learned this sometime after 2016 when this code was written.
	ans := mat.Solve(A, B) // Solve (ra1, ra2, ans, N, 1);
	fmt.Println("The solution X to AX = B using Solve is")
	ss = mat.Write(ans, 5)
	for _, s := range ss {
		fmt.Print(s)
	}

	//ans2 := mat.NewMatrix(N, N)
	ans2 := mat.GaussJ(A, B) // Solve (ra1, ra2, ans, N, 1);
	fmt.Println("The solution X to AX = B using GaussJ is")
	ss = mat.Write(ans2, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()

	//  pause();

	// Check that the solution looks right.

	//C := mat.NewMatrix(N, 1)  linter said this value of C is not used
	//D := mat.NewMatrix(N, 1)  linter said this value of D is not used
	C := mat.Mul(A, ans) // Mul (ra1, ans, N, N, 1, ra3);
	D := mat.Sub(B, C)   //  Sub (ra3, ra2, N, 1, ra4);

	fmt.Println("As a check, AX-B should be 0, and evaluates to")
	ss = mat.Write(D, 5) //    Write (ra4, N, 1, 4);
	for _, s := range ss {
		fmt.Print(s)
	}

	D = mat.BelowSmallMakeZero(D)

	fmt.Println("As a check, AX-B should be all zeros after calling BelowSmall.  It evaluates to")
	ss = mat.Write(D, 5)
	for _, s := range ss {
		fmt.Print(s)
	}
	fmt.Println()
	fmt.Println()

} // END Solve.

// -------------------------------------------- min ---------------------------------------------
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

/*
func pause() {  Written in Dec 2016.  It's not the way I would write this in 2022.  I would use fmt.Scanln(&ans)
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
*/
