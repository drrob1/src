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

const LastAltered = "10 Aug 2017"



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
  var InputFile,OutputFile os.File;
  var InputReader *bufio.Reader;
  var OutputWriter *bufio.Writer;
  sigfig,N,M,max,length int
  infilename,outfilename string

  type Point struct {
    x,y,wt,ax float64  // ax is an error factor defined by the author I reference below.
  }
  type rowvec []Point
  type matrix []rowvec

  var ra1,ra2,ra3,ra4,IM,ans matrix
  var X,Y,DecayCorY rowvec
  var lambda0,intercept0,Thalf0,lambda1,intercept1,ln2,Thalf1,lambda2,intercept2,Thalf2 float64

  var SumWt,SumWtX,SumWtY,SumWtX2,SumWtXY,SumWtY2,SumWtAx,SumWtAxX,SumWtAxY,R2,ExpectedX,ExpectedY,
      ErrorX,ErrorY2,DENOM,StDevS,StDevI  float64
//  WEIGHT,AX : ARRAY[0..MAXELE] OF LONGREAL;
//  PREVSLOPE,PREVINTRCPT : LONGREAL;
//  C,K,ITERCTR : CARDINAL;

VAR
  ch,ch1,ch2,ch3 :  CHAR;
  bool,inputprocessed :  BOOLEAN;
  sigfig,c1,c2,c3,N,M,max,ctr,len :  CARDINAL;
  inputline,OpenFileName,str1,str2,str3,str4,str5,str6,str7,str8,filter,str0,
  Thalf1strFixed,Thalf2strFixed,Thalf1gen,Thalf2gen : STRTYP;
  ns, infilename, outfilename, DirEntry : NameString;
  longstr     : ARRAY [0..5120] OF CHAR;
  InputPromptLen, LastModLen : CARDINAL;
  inputbuf    : BUFTYP;
  mybuf,token : BUFTYP;
  r           : LONGREAL;
  L           : LONGINT;
  LC          : LONGCARD;
  InFile,OutFile : MYFILTYP;
  tknstate    : FSATYP;
  c,retcod,row,col    : CARDINAL;
  i           : INTEGER;
  ra1,ra2,ra3,ra4,IM,ans : MaxRealArray;
  tpv1        : TKNPTRTYP;
  X,Y,DecayCorY : aRealArray;
  lambda1,intercept1,ln2,Thalf1,lambda2,intercept2,Thalf2 : LONGREAL;

PROCEDURE CheckPattern(VAR ns: NameString);
VAR
  c,len : CARDINAL;
  FoundStar : BOOLEAN;

BEGIN
  FoundStar := FALSE;
  len := LENGTH(ns);
  c := len;
  IF len = 0 THEN
    ns := GastricPattern;
  ELSIF ns[0] = '*' THEN
    (* do nothing *)
  ELSE
    WHILE (c > 0) AND (ns[c] <> '*') DO
      DEC(c);
    END (* WHILE *);
    IF c = 0 THEN (* asterisk not found *)
      Strings.Append("*",ns);
    END (* if c=0 *);
  END (* if len=0 *);
END CheckPattern;

(************************************************************************)
(*                              MAIN PROGRAM                            *)
(************************************************************************)

func main() {
  ln2 = math.Log(2)

  WriteLn;
  GetCommandLine(ns);
  CheckPattern(ns);  (* if ns is empty, makes pattern gastric*.txt *)
  SetFilenamePattern(ns);

  max := CountOfEntries;
  IF max > 15 THEN max := 15 END;
  IF max > 1 THEN
    FOR ctr := 1 TO max DO
        GetNextFilename(ns,DirEntry);
        WriteString(DirEntry);
        WriteLn;
    END; (* for loop to get and display directory entries *)

    FOR ctr := 1 TO max DO
        GetPrevFilename(ns,DirEntry);
    END; (* for loop to get back to first directory entry *)
  END; (* for max not 0 or 1 *)
  WriteLn;
  WriteLn;
  Position(0,20);
  WriteString(blankline);
  WriteLn;
  Position(0,20);
  WriteString(sepline);
  WriteLn;
  WriteString(DirEntry);
  WriteLn;
  LOOP
    IF CountOfEntries = 0 THEN
      WriteString(' No valid filenames found.  Need New pattern. ');
      WriteLn;
    END;
    WriteString( '<enter> or <space> to select, n for new pattern ');
    MiscM2.ReadChar(ch);

    CASE ch OF
        Terminal.CursorUp:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetPrevFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.Enter :
        EXIT; (* ns and DirEntry are already set *)
    | Terminal.CursorDown:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetNextFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.PageUp:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetPrevFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.PageDown:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetNextFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.CursorLeft:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetPrevFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.CursorRight:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetNextFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.Tab:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetNextFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | Terminal.BackSpace:
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        GetPrevFilename(ns,DirEntry);
        WriteString(blankline);
        Position(0,21);
        WriteString(DirEntry);
        WriteLn;

    | ' ':
        EXIT; (* ns and DirEntry are already set *)
    | 'n','N':
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        WriteString(blankline);
        Position(0,21);
        WriteString(' Enter new pattern: ');
        ReadString(ns);
        WriteLn;
        CheckPattern(ns);
        SetFilenamePattern(ns);
        Terminal.Reset;
        max := CountOfEntries;
        IF max > 15 THEN max := 15 END;
        IF max > 0 THEN
          FOR ctr := 1 TO max DO
            GetNextFilename(ns,DirEntry);
            WriteString(DirEntry);
            WriteLn;
          END; (* for loop to get and display directory entries *)
          FOR ctr := 1 TO max DO
            GetPrevFilename(ns,DirEntry);
          END; (* for loop to get back to first directory entry *)
        END; (* for max not 0 or 1 *)
        WriteLn;
        WriteLn;
        Position(0,20);
        WriteString(blankline);
        WriteLn;
        Position(0,20);
        WriteString(sepline);
        WriteLn;
        WriteString(DirEntry);
        WriteLn;

    | Terminal.Escape:
        HALT;

    ELSE
        (* ignore the character.  *)
    END; (* case ch *)

  END; (* loop to read and process a char *)
  WriteLn;
  WriteLn;
  WriteString(' Picked File Name is ');
  WriteString(ns);
  WriteLn;

  infilename := ns;
  IF NOT FileExists(infilename) THEN
    MiscM2.Error(' Could not find input file.  Does it exist?');
    HALT;
  END(*if*);

  outfilename := infilename;
  Strings.Append(".out",outfilename);

  ASSIGN2BUF(infilename,mybuf);
  FOPEN(InFile,mybuf,RD);

  ASSIGN2BUF(outfilename,mybuf);
  FOPEN(OutFile,mybuf,APND);
  N := 0;
  LOOP   (* read, count and process lines *)
    WHILE N < MaxN DO
        FRDTXLN(InFile,inputbuf,80,bool);
        IF bool THEN
          EXIT;
        END;
        INI1TKN(tpv1,inputbuf);
        INC(N);
        col := 1;
        REPEAT
          GETTKNREAL(tpv1,token,tknstate,i,r,retcod);
          IF (retcod = 0) AND (tknstate = DGT) THEN
            IM[N,col] := r;  (* IM is Input Matrix *)
            INC(col);
          END;
        UNTIL (retcod > 0) OR (col > MaxCol);
        IF col <= 2 THEN (* not enough numbers found on line, like if line is text only *)
                DEC(N);
        END;
    END (*while N *);
  END; (* reading loop *)
(* Now need to create A and B matrices *)
  FOR c := 1 TO N DO
        X[c] := IM[c,1];
        Y[c] := IM[c,2];
  END;

  FOR c := 1 TO N DO
        DecayCorY[c] := Y[c]/(exp(-X[c]/360.6))  (* halflife Tc-99m in minutes *)
  END;

  WriteString(' N = ');
  WriteCard(N);
  WriteLn;
  WriteString(' X is time(min) and Y is kcounts :');
  WriteLn;
  FOR c := 1 TO N DO
        WriteLongReal(X[c],5);
        WriteString('         ');
        WriteLongReal(Y[c],5);
        WriteString('         ');
        WriteLongReal(DecayCorY[c],5);
        WriteLn;
  END;
  WriteLn;

(*  PressAnyKey; *)
(*  CLS;         *)
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
