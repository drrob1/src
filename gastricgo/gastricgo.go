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

const LastAltered = "12 Aug 2017"



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
const MaxEle = 100 // for LR stuff
const IterMax = 20 // for regression algorithm
const tolerance = 1.0E-4 // for regression algorithm

  var sigfig,N,M,max,length int

  type Point struct {   // ax is an error factor defined by the author I reference below.
    x,y,ExpectedX,ExpectedY,ErrorX,ErrorY2,wt,ax,stdev float64
  }
//  type rowvec []Point  I think this is not needed.
//  type matrix []rowvec  I think this is not needed, because rowvec is made of points, not indiv elements
  type inputrow []float64
  type inputmatrix []inputrow
  type SummaryOfData struct {
	  SumWt,SumWtX,SumWtY,SumWtX2,SumWtXY,SumWtY2,SumWtAx,SumWtAxX,SumWtAxY,R2,ExpectedX,ExpectedY,
      ErrorX,ErrorY2,DENOM,StDevS,StDevI,Slope,Intercept,R2
      lambda0,intercept0,Thalf0,lambda1,intercept1,ln2,Thalf1,lambda2,intercept2,Thalf2 float64
  }


//  WEIGHT,AX : ARRAY[0..MAXELE] OF LONGREAL;
//  PREVSLOPE,PREVINTRCPT : LONGREAL;
//  C,K,ITERCTR : CARDINAL;


//************************************************************************
//*                              MAIN PROGRAM                            *
//************************************************************************

func main() {
  var point Point
  ln2 = math.Log(2)
  rows := make([]Point,0,MaxCol)
  ir := make(inputrow,2)  // input row
  im := make(inputmatrix,0,MaxN)  // input matrix
  var stats SummaryOfData

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
    fmt.Println(" Error",err," from iotuil.ReadFile when reading",Filename,".  Exiting.")
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
    rows = append(rows,point)
  }

  fmt.Println(" N = ", len(row));
  fmt.Println();
  fmt.Println(" X is time(min) and Y is kcounts");
  fmt.Println();
  for p := range rows {
    fmt.Println(p.x,p.y)
  }
  fmt.Println()
//------------------------------------------
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

func SQR(R float64) float64 {
  return R*R
} // END SQR;

//  -------------------------------SimpleSumUp ----------------------------------------
func SimpleSumUp(rows []Points) SummaryOfData {
/*
  Does the simple (unweighted) sums required of the std formula.  This is
  used as a first guess for the iterative solution performed by the other routines.
*/
  var Stats SummaryOfData

  Stats.SumWt   := float64(N)
  for _,p := range rows {                       // FOR c := 0 TO N-1 DO
      Stats.SumWtX  += p.x;
      Stats.SumWtY  += p.y;
      Stats.SumWtXY += p.x * p.y;
      Stats.SumWtX2 += SQR(p.x);
      Stats.SumWtY2 += SQR(p.y);
  }END(*FOR*);
  return Stats
}// END SimpleSumUp;

//  ------------------------------ StdLR ----------------------------------
func StdLR(Stats SummaryOfData) (SummaryOfData) {
/*
  This routine does the standard, unweighted, computation of the slope and
  intercept, using the formulas that are built into many pocket calculators,
  including mine.  This computation serves as an initial guess for the
  iterative solution used by this program as described by Dr. Zanter.
*/

  Stats := SimpleSumUp()
  SlopeNumerator := Stats.SumWt*Stats.SumWtXY - Stats.SumWtX*Stats.SumWtY;
  SlopeDenominator := Stats.SumWt*Stats.SumWtX2 - SQR(Stats.SumWtX);
  Stats.Slope := SlopeNumerator/SlopeDenominator;
  Stats.Intercept := (Stats.SumWtY - Slope*Stats.SumWtX)/Stats.SumWt;
  Stats.R2 := SQR(SlopeNumerator)/SlopeDenominator/(Stats.SumWt*Stats.SumWtY2 - SQR(Stats.SumWtY));
  return Stats
}//  END StdLR;

//  ------------------------ GetWts ------------------------------------
func GetWts(rows []Point, Stats SummaryOfData) SummaryOfData {
/*
  GET WEIGHTS.
  This routine computes the weights and the AX quantities as given by the referenced formulas.

*/

  var MinError float64 // MINIMUM ERROR ALLOWED.

  for c, p := range rows {       //    FOR c := 0 TO N-1 DO
    rows[c].ExpectedX = math.Abs((p.y - Stats.Intercept) / Stats.Slope)
    rows[c].ExpectedY = Stats.Slope * p.x + Stats.Intercept;
    rows[c].ErrorX = math.Abs(p.x - rows[c].ExpectedX)/math.Sqrt(rows[c].ExpectedX)
    MinError = tolerance*rows[c].ExpectedX // Don't need ABS call because that is now stored.
    if rows.ErrorX < MinError { ERRORX := MINERROR}
      ERRORY2 := SQR(Y[c] - EXPECTEDY)/ABS(EXPECTEDY);
      MINERROR := TOLERANCE*ABS(EXPECTEDY);
      IF ERRORY2 < MINERROR THEN ERRORY2 := MINERROR; END(*IF*);
      WEIGHT[c] := 1./(ERRORY2 + SQR(Slope*ERRORX));
      AX[c] := X[c] - WEIGHT[c]*(Slope*X[c] + Intercept - Y[c])*
                                                          Slope*SQR(ERRORX);
  }//  END FOR
  return Stats
}//  END GETWTS;

//  -------------------------------- WTSUMUP ---------------------------------
func WtSumUp(){
  (*
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
}//  END WTSUMUP;

func WTLR() {
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
}//  END WTLR;

func GETCORR() float64 {
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
	return R2
}//  END GETCORR;

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

// Sums are passed back globally
//  ---------------------------- SIMPLESUMUP --------------------------------
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

// -------------------------- fit -----------------------------
func fit(rows []Point, Stats SummaryOfData) SummaryOfData {
  
/*
  Based on Numerical Recipies code of same name on p 508-9 in Fortran,
  and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
  William H. Press, Brian P. Flannery, Saul A. Teukolsky, William T. Vettering.
  (C) 1986, Cambridge University Press.
*/
  var i int
  var wt,t,ay,ay,sxoss,sx,st2,ss,sigdata,chi2 float64

  if mwt != 0 { // means that there are variances for the data points
	for i,p := range rows {
        wt = 1/SQR(stdev[i])
		ss += wt
		sx += x[i]*wt
		sy += y[i]*wt
    }
  }else{
	for i,p := range rows {
        sx += x[i]
		sy += y[i]
	}
	ss = len(rows)
  }
  sxoss = sx/ss
  if mwt != 0 {
    for i,p := range rows {
      t = (x[i] - sxoss)/stdev[i]
	  st2 += SQR(t)
	  b += t*y[i]/stdev[i]
    }
  }else{
	for i,p := range rows {
      t = x[i] - sxoss
	  st2 =+ SQR(t)
	  b += t*y[i]
	}
  }
  b = b/st2
  a = (sy-sx*b)/ss 
  stdeva = math.Sqrt((1+SQR(sx)/(ss*st2))/ss)
  stdevb = math.Sqrt(1/st2)

  // Now to sum up Chi squared 
  if mwt == 0 {
    for i,p := range rows {
      chi2 += SQR(y[i]-a-b*x[i])
	}
	q = 1
	stdevdata = math.Sqrt(chi2/(len(rows)-2))
	stdeva *= stdevdata
	stdevb *= stdevdata
  }else{
    for i,p := range rows {
      chi2 += SQR((y[i] - a - b*x[i])/stdev[i])
	}
	q = gammq(0.5*(len(rows)-2),0.5*chi2)
  }
}

// -------------------------------- gamq ----------------------------------
func gammq(a,x float64) float64 {
// Incomplete Gamma Function, is what "Numerical Recipies" says.
// gln is the ln of the gamma function result

  if x < 0 || A <= 0 {  // error condition, but I don't want to return an error result
    return 0
  }
  if x < a+1 { // use the series representation
	gamser,gln := gser(a,x)
	return 1-gamser
  }else{
	gammcf,gln := gcf(a,x)
	return gammcf
  }
}

// ------------------------------- gser ----------------------------------
func gser(a,x float64) (float64,float64) {
  const ITERMAX = 100
  const tolerance = 3e-7

  gln := gammln(a)
  if x = 0 {
	  return 0,0
  }else if x < 0 { // invalid argument, but I'll just return 0 to not return an error result
      return 0,0
  }
  ap := a
  sum := 1/a
  del := sum
  for n := 1; n < ITERMAX; n++ {
    ap = ap + 1
	del *= x/ap
	sum += del
	if math.Abs(del) < math.Abs(sum)*tolerance {
      break	
	}
  }
  return sum*math.Exp(-x + a*math.Log(x)-gln)
}

func gammln(xx float64) float64 {
  const stp = 2.50662827465
  const half = 0.5
  const one = 1
  const fpf = 5.5

  var cof = [...]float64 {76.18009173,-86.50532033,24.01409822,-1.231739516,0.120858003e-2,-0.536382e-5}

  x := xx-one
  tmp := x+fpf
  tmp = (x + half)*math.Log(tmp)-tmp
  ser := one
  for j := range cof {
    x += one
	ser += cof[j]/x
  }
  return tmp + math.Log(stp*ser)
}

func gcf(a,x float64) (float64,float64) {
  const ITERMAX = 100
  const tolerance = 3e-7
  var gln,gold,g,fac,b1,b0,anf,ana,an,a1,a0 float64

  gln = gammln(a)
  a0 = 1.0
  a1 = x
  b1 = 1.
  fac = 1.

  for n := 1; n < ITERMAX; n++ {
	  an = float64(n)
	  ana = an - a
	  a0 = (a1 + a0*ana)*fac
	  b0 = (b1 + b0*ana)*fac
	  anf = an*fac
	  a1 = x*a0 + anf*a1
	  b1 = x*b0 + anf*b1
      if a1 != 0 {
        fac = 1/a1
		g = b1*fac
		if math.Abs((g-gold)/g) < tolerance {
          break
		}
		gold = g
	  }
  gammcf = math.Exp(-x + a*math.Log(x) - gln)*g
  return gammcf,gln
  }
}

