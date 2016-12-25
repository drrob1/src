package main;

/*
MODULE Solve;

  REVISION
  --------
   2 Mar 05 -- Added prompts to remind me of the file format.
   3 Mar 05 -- Made version 2 write lines like eqn w/o =.
   4 Mar 05 -- Don't need N as 1st line now.
  26 Feb 06 -- Will reject non-numeric entries and allows <tab> as delim.
  24 Dec 16 -- Converted to Go.
*/

import (
  "os"
  "fmt"
  "io"
  "bufio"
  "path/filepath"
  "strings"
  "getcommandline"
  "mat"
  "tokenize"
)

const LastCompiled = "25 Dec 16";
const MaxN = 9;

//                          MaxRealArray is not square because the B column vector is in last column of IM
//                                             TYPE MaxRealArray = ARRAY [1..MaxN],[1..MaxN+1] OF LONGREAL;
type Matrix2D [][]float64;



func main () {
/************************************************************************)
(*                              MAIN PROGRAM                            *)
(************************************************************************/

  fmt.Println(" Equation solver written in Go.  Last compiled ",LastCompiled);
  fmt.Println();

  if len(os.Args) <= 1 {
    fmt.Println(" Usage: solve <filename>");
    fmt.Println(" Solves vector equation A*X = B; A is a square coef matrix.");
    fmt.Println(" N is determined by number of rows and B value is last on each line.");
    os.Exit(0);
  }

  commandline := getcommandline.GetCommandLineString();
  cleancommandline := filepath.Clean(commandline);
  fmt.Println(" filename on command line is ",cleancommandline);

//  cleancommandline = "xy1.txt";

  infile,err := os.Open(cleancommandline);
  if err != nil {
    fmt.Println(" Cannot open input file.  Does it exist?  Error is ",err);
    os.Exit(1);
  }

  defer infile.Close();
  scanner := bufio.NewScanner(infile);

  IM := mat.NewMatrix(MaxN,MaxN+1);  // IM is input matrix
  IM = mat.Zero(IM);

  lines := 0;
  CountLinesLoop: for { // read, count and process lines
    for n := 0; n < MaxN; n++ {         // WHILE N < MaxN DO
      readSuccess := scanner.Scan();                      //   FRDTXLN(InFile,inputbuf,80,bool);
      if readSuccess {
        // do nothing for now,  I thought N=n made sense until I saw the need to not process short
        // lines, assuming that they are a comment line.  And lines without numbers are comment lines.
      }else{
        break CountLinesLoop;
      } // if readSuccess
      inputline := scanner.Text();
      if  readErr := scanner.Err(); readErr != nil {
        if readErr == io.EOF {
          break CountLinesLoop;
        }else{  // this may be redundant because of the readSuccess test
          break CountLinesLoop;
        }
      }

      tokenize.INITKN(inputline);
      col := 0;
      EOL := false;
      for (EOL == false) && (n <= MaxN) {                     // REPEAT
        token,EOL := tokenize.GETTKNREAL();
        if EOL { break }
        if (token.State == tokenize.DGT) {
          IM[lines][col] = token.Rsum;  // remember that IM is Input Matrix
          col++
        } // ENDIF token.state=DGT
      }                                                      //  UNTIL (retcod > 0) OR (col > MaxN);
      if col > 0 { // short line like if text only or null do not increment the row counter.
        lines++
      }
    } // END for n 
  } // END reading loop
  N := lines; // Note: lines is 0 origin

// Now need to create A and B matrices

  A := mat.NewMatrix(N,N);  // ra1 in Modula-2 code
  B := mat.NewMatrix(N,1);  // ra2 in Modula-2 code
  for row := range A {                                        // FOR row :=  1 TO N DO
    for col := range A[0] {                                   //   FOR col := 1 TO N DO
      A[row][col] = IM [row][col];
    } // END FOR col
    B[row][0] = IM[row][N];  // I have to keep remembering that [0,0] is the first row and col.
  } // END FOR row

  fmt.Println(" coef matrix A is:");
  ss := mat.Write(A,5);
  for _,s := range ss {
    fmt.Print(s);
  }
  fmt.Println();

  fmt.Println(" Right hand side vector matrix B is:");
  ss = mat.Write(B,5);
  for _,s := range ss {
    fmt.Print(s);
  }
  fmt.Println();

  ans := mat.NewMatrix(N,N);
  ans = mat.Solve(A,B);                                            // Solve (ra1, ra2, ans, N, 1);
  fmt.Println("The solution X to AX = B using Solve is");
  ss = mat.Write(ans,5);
  for _,s := range ss {
    fmt.Print(s);
  }

  ans2 := mat.NewMatrix(N,N);
  ans2 = mat.GaussJ(A,B);                                            // Solve (ra1, ra2, ans, N, 1);
  fmt.Println("The solution X to AX = B using GaussJ is");
  ss = mat.Write(ans2,5);
  for _,s := range ss {
    fmt.Print(s);
  }
  fmt.Println();

//  pause();

// Check that the solution looks right.

  C := mat.NewMatrix(N,1);
  D := mat.NewMatrix(N,1);
  C = mat.Mul(A,ans);                                          // Mul (ra1, ans, N, N, 1, ra3);
  D = mat.Sub(B,C);                                           //  Sub (ra3, ra2, N, 1, ra4);

  fmt.Println("As a check, AX-B should be 0, and evaluates to");
  ss = mat.Write(D,5);                                      //    Write (ra4, N, 1, 4);
  for _,s := range ss {
    fmt.Print(s);
  }

  D = mat.BelowSmallMakeZero(D);

  fmt.Println("As a check, AX-B should be all zeros after calling BelowSmall.  It evaluates to");
  ss = mat.Write(D,5);
  for _,s := range ss {
    fmt.Print(s);
  }
  fmt.Println();
  fmt.Println();

}// END Solve.



func pause() {
  scnr := bufio.NewScanner(os.Stdin)
  fmt.Print(" pausing ... hit <enter>");
  scnr.Scan();
  answer := scnr.Text();
  if err := scnr.Err(); err != nil {
    fmt.Fprintln(os.Stderr, "reading standard input:", err)
    os.Exit(1);
  }
  ans := strings.TrimSpace(answer);
  ans = strings.ToUpper(ans);
  fmt.Println(ans);
}

