// GastricEmptying in Go.  (C) 2017.  Based on GastricEmtpying2.mod code.
// Includes LR code here.  No point in making it separate.

package main

/*
  REVISION HISTORY
  ----------------
  24 Feb 06 -- Added logic to excluded lines w/o numbers.
  23 Feb 06 -- LR.mod Converted to SBM2 for Win.  And since it is not allowed for this module to
                 write errors, slope and intercept will be set zero in case of an error.
  25 Sep 07 -- Decided to write to std out for output only
   3 Oct 08 -- Added decay correction for Tc-99m
  12 Sep 11 -- Added my new FilePicker module.
  26 Dec 14 -- Will change to the FilePickerBase module as I now find it easier to use.
                 But these use terminal mode.
  27 Dec 14 -- Removing unused modules so can compile in ADW.
  21 Jun 15 -- Will write out to file, and close all files before program closes.
  10 Aug 17 -- Converting to Go
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	//
	"getcommandline"
	//  "bufio"
	//  "strconv"
	//  "timlibg"
	//  "tokenize"
)

const LastAltered = "11 Aug 2017"



/*
  Normal values from source that I don't remember anymore.
  1 hr should have 90% of activity remaining in stomach = 6.6 hr halflife = 395 min halflife
  2 hr should have 60% of activity remaining in stomach = 2.7 hr halflife = 163 min halflife
  3 hr should have 30% of activity remaining in stomach = 1.7 hr halflife = 104 min halflife
  4 hr should have 10% of activity remaining in stomach = 1.2 hr halflife = 72 min halflife

*/


//----------------------------------------------------------------------------

const MaxN = 500;
const MaxCol = 10;
const SigFig = 2;
const GastricPattern = "gastric*.txt";
const MenuSep = '|';
const blankline = "                                                              ";
const sepline = "--------------------------------------------------------------";
  var sigfig,N,M,max,length int

  type Point struct {
    x,y,wt,ax float64  // ax is an error factor defined by the author I reference below.
  }
  type rowvec []Point
//  type matrix []rowvec  I think this is not needed, because rowvec is made of points, not indiv elements
  type inputrow []float64
  type inputmatrix []inputrow

  var ra1,ra2,ra3,ra4,IM,ans matrix
  var X,Y,DecayCorY rowvec
  var lambda0,intercept0,Thalf0,lambda1,intercept1,ln2,Thalf1,lambda2,intercept2,Thalf2 float64

  var SumWt,SumWtX,SumWtY,SumWtX2,SumWtXY,SumWtY2,SumWtAx,SumWtAxX,SumWtAxY,R2,ExpectedX,ExpectedY,
      ErrorX,ErrorY2,DENOM,StDevS,StDevI  float64
//  WEIGHT,AX : ARRAY[0..MAXELE] OF LONGREAL;
//  PREVSLOPE,PREVINTRCPT : LONGREAL;
//  C,K,ITERCTR : CARDINAL;


func CheckPattern(ns string) string {

  s := ns;
  FoundStar := false;
  if len(ns) == 0 {
    s = GastricPattern;
  }else if ns[0] == '*' {
    // do nothing
  }else if ! strings.Contains(ns,"*") { // asterisk not found
      s = s + "*"
  } // END if len == 0
  return s
} // END CheckPattern;

//************************************************************************
//*                              MAIN PROGRAM                            *
//************************************************************************

func main() {
  var point Point
  ln2 = math.Log(2)
  row := make(rowvec,0,MaxCol)
  TimeAndCountsTable := make(matrix,0,MaxN)
  ir := make(inputrow,2)  // input row
  im := make(inputmatrix,0,MaxN)  // input matrix

  fmt.Println(" Gastric Emtpying program written in Go.  Last modified",LastAltered)
  if len(os.Args) <= 1 {
    fmt.Println(" Usage: GastricEmptying <filename>")
    os.Exit(0)
  }

  InExtDefault := ".txt"
  OutExtDefault := ".out"

  ns := getcommandline.GetCommandLineString(ns);
  BaseFilename := filepath.Clean(ns)
  Filename := ""
  FileExists := false

  if strings.Contains(BaseFilename,".") {
    Filename = BaseFilename
    FI, err := os.Stat(Filename)
    if err == nil {
      FileExists = true
    }
  } else {
    Filename = BaseFilename + InExtDefault
    FI, err := os.Stat(Filename)
    if err == nil {
      FileExists = true
    }
  }

  if ! FileExists {
    fmt.Println(" File",BaseFilename," or ",Filename," do not exist.  Exiting.")
    os.Exit(1)
  }

  byteslice := make([]byte,0,5000)
  byteslice,err := ioutil.ReadFile(Filename)
  if err != nil {
    fmt.Println(" Error",err," from iotuil.ReadFile when reading",Filename,".  Exiting."
    os.Exit(1)
  }

  bytesbuffer := bytes.NewBuffer(byteslice)


  outfilename := BaseFilename + OutExtDefault
  OutputFile, err := os.OpenFile(OutFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

  if err != nil {
    fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
    os.Exit(1)
  }
  defer OutputFile.Close()
  OutBufioWriter := bufio.NewWriter(OutputFile)
  defer OutBufioWriter.Flush()
  _, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
  check(err)

  EOL := false;
  N := 0
  for N < MaxN { // Main input reading loop to read all lines
    line, err := bytesbuffer.ReadString('\n')
    if err != nil {
      break
    }
    token := tknptr.NewToken(line);
    col := 0
    // loop to process first 2 digit tokens on this line and ignore the rest
    for col < 2 {
        token,EOL =  tknptr.GETTKNREAL()
	if EOL {
          break
	}
        if token.State == DGT { // ignore non-digit tokens, which allows for comments
          ir[col] = token.Rsum  // ir is input row
          col++
        }
    } // UNTIL EOL OR have 2 numbers
    if col > 1 { // process line as it has 2 numbers
      im = append(im,ir)
      N++
    }
  }// END main input reading loop

// Now need to populate the Time And Counts Table
  for c := range im {
    point.x = im[c][0]
    point.y = im[c][1]
    row = append(row,point)
  }

  fmt.Println(" N = ", len(row));
  fmt.Println();
  fmt.Println(" X is time(min) and Y is kcounts");
  fmt.Println();
  for p := range row {
    fmt.Println(p.x,p.y)
  }
  fmt.Println()
------------------------------------------------------------------------
  SEMILOGLR(N,X,Y,lambda1,intercept1);
  Thalf1 := -ln2/lambda1;
  LongStr.RealToFixed(Thalf1,SigFig,Thalf1strFixed);
  LongStr.RealToStr(Thalf1,Thalf1gen);

  SEMILOGLR(N,X,DecayCorY,lambda2,intercept2);
  Thalf2 := -ln2/lambda2;
  LongStr.RealToFixed(Thalf2,SigFig,Thalf2strFixed);
  LongStr.RealToStr(Thalf2,Thalf2gen);


  WriteString(' Uncorrected T-1/2 is ');
  WriteString(Thalf1strFixed);
(*                                               WriteLongReal(Thalf1,4); *)
  WriteString(' minutes.  Corrected T-1/2 is ');
  WriteString(Thalf2strFixed);
(*                                               WriteLongReal(Thalf2,4); *)
  WriteString(' minutes.');
  WriteLn;

  FWRSTR(OutFile," Uncorrected T-1/2 is ");
  FWRSTR(OutFile,Thalf1strFixed);
  FWRSTR(OutFile," (");
  FWRSTR(OutFile,Thalf1gen);
  FWRSTR(OutFile," ) minutes.  Corrected T-1/2 is ");
  FWRSTR(OutFile,Thalf2strFixed);
  FWRSTR(OutFile," (");
  FWRSTR(OutFile,Thalf2gen);
  FWRSTR(OutFile," ) minutes.");
  FWRLN(OutFile);
(*
  WriteString(' uncorrected T-1/2 is ');
  WriteLongReal(Thalf1,4);
  WriteString(' minutes, lambda is ');
  WriteLongReal(-lambda1,6);
  WriteString(' reciprocal minutes.');
  WriteLn;
  WriteString(' intercept is ');
  WriteLongReal(intercept1,6);
  WriteString(' kcounts.');
  WriteLn;
  WriteString(' Decay corrected T-1/2 is ');
  WriteLongReal(Thalf2,4);
  WriteString(' minutes, lambda is ');
  WriteLongReal(-lambda2,6);
  WriteString(' reciprocal minutes.');
  WriteLn;
  WriteString(' intercept is ');
  WriteLongReal(intercept2,6);
  WriteString(' kcounts.');
  WriteLn;
*)
  IF tpv1 # NIL THEN DISPOSE(tpv1); END;
  FCLOSE(InFile);
  FCLOSE(OutFile);
  PressAnyKey;
} // END main GastricEmptying.go



/* From LR.mod
  REVISION HISTORY
  ----------------
  22 Oct 88 -- 1) Added the GETCORR Procedure, but on testing the R**2 value did not change with each 
                   iteration, casting doubt on the validity of my weighting modification.
               2) Fixed omission in STDLR in that SUMWTXY was not computed.
  23 Feb 06 -- Converted to SBM2 for Win.  And since it is not allowed for this module to
                 write errors, slope and intercept will be set zero in case of an error.
*/
  IMPORT WholeStr, LongStr, LongMath;
  FROM LongMath IMPORT sqrt, exp, ln;

  const MaxEle = 100;
        IterMax = 20;
        tolerance = 1.0E-4;

func SQR(R float64) float64 {
  return R*R
} // END SQR;

func DOLR(N:CARDINAL; X,Y:ARRAY OF LONGREAL; VAR Slope,Intercept:LONGREAL){
/*
---------------------------------- DOLR ---------------------------------
Do Linear Regression Routine.
  This routine does the actual linear regression using a weighted algorithm
that is described in Zanter, Jean-Paul, "Comparison of Linear Regression
Algortims," JPAM Jan/Feb '86, 5:1, p 14-22.  This algorithm is used instead
of the std one because the std one assumes that the errors on the independent
variable are negligible and the errors on the dependent variable are
constant.  This is definitely not the case in a clinical situation where,
for example, both the time of a blood sample and the counts per minute in
that blood sample, are subject to variances.


(*
                                      1
  WEIGHT OF A POINT = ---------------------------------------
                      (error_on_y)**2 + (slope*error_on_x)**2

  AX -- a quantity defined by the author, used in the iterative solution of
        the system of equations used in computing the weights, and hence the
        errors on the data, and on the slopes and intercepts.  Like weights,
        it applies to each point.

     =  X - Weight * (Intercept + Slope*X - Y) * Slope * Error_on_X**2

It is clear that the weights and the AX quantity depend on the errors
on the data points.  My modification is a variation on the chi square
analysis to determine the errors on the data points.  This, too is
iterated.

                     (Observed value - Expected value)**2
 CHI SQUARE = sum of ------------------------------------
                             Expected value

*/

(*
  The following procedures are local to this outer procedure, and all of the
  I/O for these are passed indirectly (globally).
*)

  PROCEDURE SIMPLESUMUP;
  (*
  **************************** SIMPLESUMUP ********************************
    Does the simple (unweighted) sums required of the std formula.  This is
  used as a first guess for the iterative solution performed by the
  other routines.
  *)
  VAR c : CARDINAL;
  BEGIN
    SUMWT   := FLOAT(N);
    SUMWTX  := 0.;
    SUMWTY  := 0.;
    SUMWTX2 := 0.;
    SUMWTXY := 0.;
    SUMWTY2 := 0.;

    FOR c := 0 TO N-1 DO
      SUMWTX  := SUMWTX  + X[c];
      SUMWTY  := SUMWTY  + Y[c];
      SUMWTXY := SUMWTXY + X[c]*Y[c];
      SUMWTX2 := SUMWTX2 + SQR(X[c]);
      SUMWTY2 := SUMWTY2 + SQR(Y[c]);
    END(*FOR*);
  END SIMPLESUMUP;

  PROCEDURE STDLR;
  (*
  ****************************** STDLR **********************************
  This routine does the standard, unweighted, computation of the slope and
  intercept, using the formulas that are built into many pocket calculators,
  including mine.  This computation serves as an initial guess for the
  iterative solution used by this program as described by Dr. Zanter.
  *)

  VAR SLOPENUMERATOR,SLOPEDENOMINATOR : LONGREAL;

  BEGIN
    SIMPLESUMUP;
    SLOPENUMERATOR := SUMWT*SUMWTXY - SUMWTX*SUMWTY;
    SLOPEDENOMINATOR := SUMWT*SUMWTX2 - SQR(SUMWTX);
    Slope := SLOPENUMERATOR/SLOPEDENOMINATOR;
    Intercept := (SUMWTY - Slope*SUMWTX)/SUMWT;
    R2 := SQR(SLOPENUMERATOR)/SLOPEDENOMINATOR/(SUMWT*SUMWTY2 - SQR(SUMWTY));
  END STDLR;

  PROCEDURE GETWTS;
  (*
  ************************ GETWTS ************************************
  GET WEIGHTS.
  This routine computes the weights and the AX quantities as given by the
  above formulas.

  *)

  VAR MINERROR : LONGREAL; (* MINIMUM ERROR ALLOWED. *)
          c        : CARDINAL;

  BEGIN
    FOR c := 0 TO N-1 DO
      EXPECTEDX := (Y[c] - Intercept) / Slope;
      EXPECTEDY := Slope * X[c] + Intercept;
      ERRORX := ABS(X[c] - EXPECTEDX)/sqrt(ABS(EXPECTEDX));
      MINERROR := TOLERANCE*ABS(EXPECTEDX);
      IF ERRORX < MINERROR THEN ERRORX := MINERROR; END(*IF*);
      ERRORY2 := SQR(Y[c] - EXPECTEDY)/ABS(EXPECTEDY);
      MINERROR := TOLERANCE*ABS(EXPECTEDY);
      IF ERRORY2 < MINERROR THEN ERRORY2 := MINERROR; END(*IF*);
      WEIGHT[c] := 1./(ERRORY2 + SQR(Slope*ERRORX));
      AX[c] := X[c] - WEIGHT[c]*(Slope*X[c] + Intercept - Y[c])*
                                                          Slope*SQR(ERRORX);
    END(*FOR*);
  END GETWTS;

  PROCEDURE WTSUMUP;
  (*
  ******************************** WTSUMUP *********************************
  Weighted Sum Up.
  This procedure sums the variables using the weights and AX quantaties as
  described (and computed) above.

  *)
  VAR W : LONGREAL;
          c : CARDINAL;

  BEGIN
    SUMWT    := 0.;
    SUMWTX   := 0.;
    SUMWTY   := 0.;
    SUMWTX2  := 0.;
    SUMWTXY  := 0.;
    SUMWTAX  := 0.;
    SUMWTAXX := 0.;
    SUMWTAXY := 0.;

    GETWTS;
    FOR c := 0 TO N-1 DO
      W := WEIGHT[c];
      SUMWT := SUMWT + W;
      SUMWTX := SUMWTX + W*X[c];
      SUMWTX2 := SUMWTX2 + W*SQR(X[c]);
      SUMWTY := SUMWTY + W*Y[c];
      SUMWTAX := SUMWTAX + W*AX[c];
      SUMWTAXX := SUMWTAXX + W*AX[c]*X[c];
      SUMWTAXY := SUMWTAXY + W*AX[c]*Y[c];
    END(*FOR*);
  END WTSUMUP;

  PROCEDURE WTLR;
  (*
  ******************************** WTLR ********************************
  Weighted Linear Regression.
  This procedure computes one iteration of the weighted*AX computation of
  the Slope and Intercept.

  *)
  BEGIN
    WTSUMUP;
    Slope := (SUMWTAX*SUMWTY - SUMWTAXY*SUMWT)/
                                          (SUMWTX*SUMWTAX - SUMWTAXX*SUMWT);
    Intercept := (SUMWTY - Slope*SUMWTX)/SUMWT;
  END WTLR;

  PROCEDURE GETCORR(VAR R2 : LONGREAL);
  (*
  ********************************* GETCORR ****************************
  Get Correlation Coefficient
  This uses a variant of the std formula, taking the weights into acnt.
  *)

  VAR W : LONGREAL;
          c : CARDINAL;

  BEGIN
    SUMWT    := 0.;
    SUMWTX   := 0.;
    SUMWTY   := 0.;
    SUMWTXY  := 0.;
    SUMWTX2  := 0.;
    SUMWTY2  := 0.;

    GETWTS;
    FOR c := 0 TO N-1 DO
      W := WEIGHT[c];
      SUMWT := SUMWT     + W;
      SUMWTX := SUMWTX   + W*X[c];
      SUMWTY := SUMWTY   + W*Y[c];
      SUMWTXY := SUMWTXY + W*X[c]*Y[c];
      SUMWTX2 := SUMWTX2 + W*SQR(X[c]);
      SUMWTY2 := SUMWTY2 + W*SQR(Y[c]);
    END(*FOR*);
    R2 := SQR(SUMWT*SUMWTXY - SUMWTX*SUMWTY)/
                (SUMWT*SUMWTX2 - SQR(SUMWTX))/(SUMWT*SUMWTY2 - SQR(SUMWTY));
  END GETCORR;

BEGIN (* BODY OF DOLR PROCEDURE *)
  STDLR;
  PREVSLOPE := Slope;
  PREVINTRCPT := Intercept;
(*
  GETCORR(R2);
  WriteString(' Slope = ');
  WriteReal(SLOPE,0);
  WriteString(';  Intercept = ');
  WriteReal(INTRCPT,0);
  WriteString(';  R**2 = ');
  WriteReal(R2,0);
  WriteLn;
*)

  LOOP
    FOR ITERCTR := 1 TO ITERMAX DO
      WTLR;
(*
      WriteString(' Slope = ');
      WriteReal(SLOPE,0);
      WriteString(';  Intercept = ');
      WriteReal(INTRCPT,0);
      WriteString(';  R**2 = ');
      WriteReal(R2,0);
      WriteLn;
      WriteString(' Weights =');
      FOR C := 0 TO N-1 DO
        WriteReal(WEIGHT[c],0);
        IF (C+1) MOD 6 = 0 THEN WriteLn; END(*IF*);
      END(*FOR*);
      WriteLn;
*)
      IF (ABS(Slope - PREVSLOPE) < TOLERANCE*ABS(Slope)) AND
                   (ABS(Intercept - PREVINTRCPT) < TOLERANCE*ABS(Intercept)) THEN
        EXIT;
      ELSE
        PREVSLOPE := Slope;
        PREVINTRCPT := Intercept;
      END(*IF*);
    END(*FOR*);
    EXIT;
  END(*LOOP*);
(*
  DENOM := SUMWT*SUMWTX2 - SQR(SUMWTX);
  StDevS := sqrt(ABS(SUMWT/DENOM));
  StDevI := sqrt(ABS(SUMWTX2/DENOM));
  WriteString(' St Dev on Slope is ');
  WriteReal(StDevS,0);
  WriteString(';  St Dev on Intercept is ');
  WriteReal(StDevI,0);
  WriteLn;
*)
} // END DOLR;

PROCEDURE SIMPLELR(N:CARDINAL; X,Y:ARRAY OF LONGREAL; VAR Slope,Intercept:LONGREAL);
(*
******************************* SIMPLELR *******************************
SIMPLE Linear Regression.
This routine is the entry point for client modules to use the linear
regression algorithm.  This separation allows for more complex
linearizations to be performed before the regression is computed.
*)
BEGIN
  IF (N > HIGH(X)+1) OR (N > HIGH(Y)+1) OR (N > MAXELE) OR (N < 3) THEN
    Slope := 0.;
    Intercept := 0.;
    RETURN;
  END;
(*
  IF (N > HIGH(X)+1) OR (N > HIGH(Y)+1) THEN
    WriteString(' *ERROR*  Value of N is greater than the size of the ');
    WriteString('X or Y arrays.');
    WriteLn;
    WriteString(' N = ');
    WriteCard(N,0);
    WriteString(';  Size of the X array is ');
    WriteCard(HIGH(X),0);
    WriteString(';  Size of the Y array is ');
    WriteCard(HIGH(Y),0);
    WriteString('.');
    WriteLn;
    WriteString(' Program results may be unreliable.');
    WriteLn;
    N := HIGH(X) + 1;
    IF N > HIGH(Y) THEN N := HIGH(Y) + 1; END/*IF*/;
  ELSIF N > MAXELE THEN
    WriteString(' Too Many Elements.  Maximum # is ');
    WriteCard(MAXELE,0);
    WriteString('.  Requested # of elements is ');
    WriteCard(N,0);
    WriteString(' Trailing elements truncated.');
    N := MAXELE;
  ELSIF N < 3 THEN
    WriteString(' Need at least 3 points.  Linear Regression not done.');
    WriteLn;
    SLOPE := 0.;
    INTRCPT := 0.;
    RETURN;
  END/*IF*/;
*)


  DOLR(N,X,Y,Slope,Intercept);
END SIMPLELR;

PROCEDURE SEMILOGLR(N:CARDINAL; X,Y:ARRAY OF LONGREAL; VAR Slope,Intercept:LONGREAL);
(*
******************************** SEMILOGLR *******************************
SemiLogarithmic Linear Regression.
This entry point first performs a log on the ordinate (Y values) before
calling the linear regression routine.

*)

VAR LOGY : ARRAY[0..MAXELE] OF LONGREAL;
    C    : CARDINAL;

BEGIN
  IF (N > HIGH(X)+1) OR (N > HIGH(Y)+1) OR (N > MAXELE) OR (N < 3) THEN
    Slope := 0.;
    Intercept := 0.;
    RETURN;
  END;
(*
  IF (N > HIGH(X)+1) OR (N > HIGH(Y)+1) THEN
    WriteString(' *ERROR*  Value of N is greater than the size of the ');
    WriteString('X or Y arrays.');
    WriteLn;
    WriteString(' N = ');
    WriteCard(N,0);
    WriteString(';  Size of the X array is ');
    WriteCard(HIGH(X),0);
    WriteString(';  Size of the Y array is ');
    WriteCard(HIGH(Y),0);
    WriteString('.');
    WriteLn;
    WriteString(' Program results may be unreliable.');
    WriteLn;
    N := HIGH(X) + 1;
    IF N > HIGH(Y) THEN N := HIGH(Y) + 1; END/*IF*/;
  ELSIF N > MAXELE THEN
    WriteString(' Too Many Elements.  Maximum # is ');
    WriteCard(MAXELE,0);
    WriteString('.  Requested # of elements is ');
    WriteCard(N,0);
    WriteString(' Trailing elements truncated.');
    N := MAXELE;
  ELSIF N < 3 THEN
    WriteString(' Need at least 3 points.  Linear Regression not done.');
    WriteLn;
    SLOPE := 0.;
    INTRCPT := 0.;
    RETURN;
  END/*IF*/;
*)
  FOR C := 0 TO N-1 DO
(*
    IF Y[c] < 0. THEN
      WriteString(' Cannot take the log of a negative number.  The Y');
      WriteString(' value of');
      WriteLn;
      WriteReal(Y[c],0);
      WriteString(' was made positive.  Results may be invalid.');
      WriteLn;
    END/*IF*/;
*)
    LOGY[C] := ln(ABS(Y[C]));
  END(*FOR*);
  DOLR(N,X,LOGY,Slope,Intercept);
  Intercept := exp(Intercept);
END SEMILOGLR;

END LR.
//===========================================================
func check(e error) {
	if e != nil {
		panic(e)
	}
}
