// GastricEmptying in Go, V 2.  (C) 2017.  Originallly on GastricEmtpying2.mod code.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	//
	"src/filepicker"
	"src/getcommandline"
	"src/tknptr"
)

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
  12 Aug 17 -- Added "Numerical Recipies" code, and removed old iterative algorithm.
  15 Aug 17 -- Added ability to show timestamp of the executable file, as a way of showing last linking timestamp.
                 To be used in addition of the LastAltered string.  But I can recompile without altering, as when
		 a new version of the Go toolchain is released.
   9 Sep 17 -- Changing how bufio errors are checked, based on a posting from Rob Pike.
  11 Sep 17 -- Made the stdev a % factor instead of a constant factor, and tweaked the output in other ways.
  12 Sep 17 -- Added heading to output file.
  13 Sep 17 -- Adding code from Numerical Recipies for errors in x and y.
  24 Sep 17 -- To make the new code work, I'll remove lny and use OrigY and y.
  27 Sep 17 -- It works after I fixed some typos.  And I changed the order of the output values.
   2 Oct 17 -- Discovered that in very normal patients, counts are low enough for stdev to be neg.  I can't allow that!
                 And error in X (time) will be same %-age as in Y.  And StDevY cannot be larger than y.
   3 Oct 17 -- Added separate StDevTime constant, and set a minimum of 2.  Output optimized for Windows since that
                 is where it will be used in "production."
  18 Oct 17 -- Added filepicker
*/

const LastAltered = "18 Oct 2017"

/*
  Normal values from source that I don't remember anymore.
  1 hr should have 90% of activity remaining in stomach = 6.6 hr halflife = 395 min halflife
  2 hr should have 60% of activity remaining in stomach = 2.7 hr halflife = 163 min halflife
  3 hr should have 30% of activity remaining in stomach = 1.7 hr halflife = 104 min halflife
  4 hr should have 10% of activity remaining in stomach = 1.2 hr halflife = 72 min halflife

*/

//----------------------------------------------------------------------------

const MaxN = 500
const MaxCol = 10
const StDevFac = 25 // treated as a percentage for weights used for GastricGo v1 code.
const StDevTime = 10
const POTN = 1.571000 // used for fitexy and related routines
const BIG = 1e30      // this too
const ACC = 1e-3      // this too
const ITERMAX = 100   // this too

// stdev relates to counts, or the ordinate.  This algorithm has it apply to the abscissa also.
// Version 2 handles these separately in fitexy.
type Point struct {
	x, y, OrigY, stdev,
	sigx, sigy, xx, yy, sx, sy, ww float64 // added when I added fitexy in Sep 2017
}

type inputrow []float64
type inputmatrix []inputrow
type FittedData struct {
	Slope, Intercept, StDevSlope, StDevIntercept, GoodnessOfFit float64
}

type FittedData2 struct {
	Slope, Intercept, StDevSlope, StDevIntercept, chi2, q, scale float64
}

//	COMMON /fitxyc/ xx,yy,sx,sy,ww -- NOW PART OF Point -- ,aa,ofs,nn -- defined next.
var aa, offs, nn float64

//************************************************************************
//*                              MAIN PROGRAM                            *
//************************************************************************

func main() {
	var point Point
	var err, bufioErr error
	var BaseFilename string

	ln2 := math.Log(2)
	rows := make([]Point, 0, MaxCol)
	im := make(inputmatrix, 0, MaxN) // input matrix

	fmt.Println()
	fmt.Println(" GastricGo v 2, Gastric Emtpying program written in Go.  Last modified", LastAltered)
	fmt.Println()
	//	if len(os.Args) <= 1 {
	//		fmt.Println(" Usage: gastric2 <filename>.txt")
	//		os.Exit(0)
	//	}
	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf("%s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n", ExecFI.Name(), LastLinkedTimeStamp, workingdir, execname)
	//	fmt.Println(ExecFI.Name(), " has timestamp of", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
	//	fmt.Println(" Full name of executable file is", execname)
	fmt.Println()

	InExtDefault := ".txt"
	OutExtDefault := ".out"
	ans := ""
	Filename := ""
	FileExists := false

	if len(os.Args) <= 1 { // need to use filepicker
		filenames := filepicker.GetFilenames("gastric*.txt")
		for i := 0; i < min(len(filenames), 10); i++ {
			fmt.Println("filename[", i, "] is", filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		fmt.Scanln(&ans)
		if len(ans) == 0 {
			ans = "0"
		}
		i, err := strconv.Atoi(ans)
		if err == nil {
			Filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			Filename = filenames[i]
		}
		fmt.Println(" Picked filename is", Filename)
		BaseFilename = Filename
	} else { // will use filename entered on commandline
		ns := getcommandline.GetCommandLineString()
		BaseFilename = filepath.Clean(ns)

		if strings.Contains(BaseFilename, ".") {
			Filename = BaseFilename
			_, err := os.Stat(Filename)
			if err == nil {
				FileExists = true
			}
		} else {
			Filename = BaseFilename + InExtDefault
			_, err := os.Stat(Filename)
			if err == nil {
				FileExists = true
			}
		}

		if !FileExists {
			fmt.Println(" File", BaseFilename, " or ", Filename, " do not exist.  Exiting.")
			os.Exit(1)
		}
		fmt.Println(" Filename is", Filename)
	}

	byteslice := make([]byte, 0, 5000)
	byteslice, err = ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error", err, " from iotuil.ReadFile when reading", Filename, ".  Exiting.")
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(byteslice)

	outfilename := BaseFilename + OutExtDefault
	OutputFile, err := os.OpenFile(outfilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", outfilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()
	_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
	check(err)

	N := 0
	for N < MaxN { // Main input reading loop to read all lines
		line, err := bytesbuffer.ReadString('\n')
		if err != nil {
			break
		}
		tokenreader := tknptr.NewToken(line)
		col := 0
		ir := make(inputrow, 2) // input row
		// loop to process first 2 digit tokens on this line and ignore the rest
		for col < 2 {
			token, EOL := tokenreader.GETTKNREAL()
			if EOL {
				break
			}
			if token.State == tknptr.DGT { // ignore non-digit tokens, which allows for comments
				ir[col] = token.Rsum // ir is input row
				col++
			}
		} // UNTIL EOL OR have 2 numbers
		if col >= 1 { // process line as it has 2 numbers
			im = append(im, ir)
			N++
		}
	} // END main input reading loop

	// Now need to populate the Time And Counts Table
	for c := range im {
		point.x = im[c][0]
		point.OrigY = im[c][1]
		point.y = math.Log(point.OrigY)
		point.stdev = math.Abs(math.Log(point.OrigY * StDevFac / 100)) // treating StDevFac as a %-age.
		if math.Abs(point.y) < math.Abs(point.stdev) {
			point.stdev = math.Abs(point.y)
		}
		point.sigx = point.x * StDevTime / 100
		if point.sigx < 2 {
			point.sigx = 2
		}
		point.sigy = point.stdev
		rows = append(rows, point)
	}
	// this is a closure, I think my first one.
	writestr := func(s string) {
		if bufioErr != nil {
			return
		}
		_, bufioErr = OutBufioWriter.WriteString(s)
	}

	writerune := func() { // this is a closure.
		if bufioErr != nil {
			return
		}
		_, bufioErr = OutBufioWriter.WriteRune('\n')
	}

	// fmt.Println(" Date and Time in default format:", date)
	s := fmt.Sprintf(" Date and Time in basic format: %s \n", datestring)
	fmt.Println(s)
	writestr(s) // using the closure from above, I hope.
	writerune()
	check(bufioErr)

	fmt.Println(" N = ", len(rows))
	fmt.Println()
	fmt.Println(" X is time(min)  Y is kcounts  Ln(y)     X StDev   Y stdev")
	fmt.Println()
	writestr(" X is time(min)  Y is kcounts  Ln(y)     X StDev   Y stdev") // the closure from above
	writerune()
	for _, p := range rows {
		s := fmt.Sprintf("%11.0f %13.2f %10.4f %10.4f %10.4f\n", p.x, p.OrigY, p.y, p.sigx, p.stdev)
		fmt.Print(s)
		writestr(s) // the closure from above
	}
	fmt.Println()
	writerune()
	check(bufioErr)

	stdslope, stdintercept, stdr2 := StdLR(rows)
	stdhalflife := -ln2 / stdslope

	s = fmt.Sprintf(" Original T-1/2 of Gastric Emptying is %.2f minutes.  Original std unweighted slope is %.6f and std intercept is %.6f and R-squared is %.6f \n", stdhalflife, stdslope, stdintercept, stdr2)
	fmt.Print(s)
	fmt.Println()
	writestr(s) // using the write closure, I hope
	writerune()
	check(bufioErr)

	UnWeightedResults := fitfull(rows, false)
	Unweightedhalflife := -ln2 / UnWeightedResults.Slope
	s = fmt.Sprintf(" unweighted halflife of Gastric Emptying is %.2f minutes.  Slope= %.6f, StDevSlope= %.6f. \n",
		Unweightedhalflife, UnWeightedResults.Slope, UnWeightedResults.StDevSlope)
	fmt.Println(s)
	fmt.Println()
	writestr(s) // using the write closure, I hope
	writerune()
	check(bufioErr)

	WeightedResults := fit(rows)
	weightedhalflife := -ln2 / WeightedResults.Slope
	s = fmt.Sprintf(" Weighted halflife of Gastric Emptying is %.2f minutes.  Slope= %.6f, StDevSlope= %.6f, GoodnessOfFit= %.6f \n",
		weightedhalflife, WeightedResults.Slope, WeightedResults.StDevSlope, WeightedResults.GoodnessOfFit)
	fmt.Println(s)
	fmt.Println()
	writestr(s) // using the write closure
	writerune()
	check(bufioErr)

	WeightedResults2 := fitfull(rows, true)
	weightedhalflife2 := -ln2 / WeightedResults2.Slope
	s = fmt.Sprintf(" halflife of Gastric Emptying using Weights2 and fitfull is %.2f minutes, fit is %.6f. \n", weightedhalflife2, WeightedResults2.GoodnessOfFit)
	fmt.Println(s)
	writestr(s)
	writerune()
	check(bufioErr)

	WeightedResults3 := fitexy(rows)
	weightedhalflife3 := -ln2 / WeightedResults3.Slope
	s = fmt.Sprintf(" halflife of Gastric Emptying using Weights3 and fitexy is %.2f minutes, stdev is %.6f.", weightedhalflife3, WeightedResults3.StDevSlope)
	fmt.Print(s)
	writestr(s)
	s = fmt.Sprintf("  chi2= %.6f, q= %.6f \n", WeightedResults3.chi2, WeightedResults3.q)
	fmt.Println(s)
	writestr(s)
	writerune()
	check(bufioErr)

	// The files will flush and close themselves because of the defer statements.
	fmt.Println()
	fmt.Println()
} // end main

// -------------------------------------- SQR ---------------------------------------------
func SQR(R float64) float64 {
	return R * R
} // END SQR;

//  ------------------------------------- StdLR ----------------------------------
func StdLR(rows []Point) (float64, float64, float64) {
	/*
	   This routine does the standard, unweighted, computation of the slope and intercept,
	   using the formulas that are built into many pocket calculators, including mine.

	   Note that I'm linearizing the data by using ln(y), not y itself.
	*/
	var sumx, sumy, sumx2, sumy2, sumxy float64

	N := float64(len(rows))
	for _, p := range rows { // FOR c := 0 TO N-1 DO
		sumx += p.x
		sumy += p.y
		sumxy += p.x * p.y
		sumx2 += SQR(p.x)
		sumy2 += SQR(p.y)
	} // ENDFOR
	SlopeNumerator := N*sumxy - sumx*sumy
	SlopeDenominator := N*sumx2 - SQR(sumx)
	Slope := SlopeNumerator / SlopeDenominator
	Intercept := (sumy - Slope*sumx) / N
	R2 := SQR(SlopeNumerator) / SlopeDenominator / (N*sumy2 - SQR(sumy))
	return Slope, Intercept, R2
} //  END StdLR;

//===========================================================
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// -------------------------- fit -----------------------------
func fit(rows []Point) FittedData {

	/*
		subroutine fit(x,y,ndata,sig,mwt,a,b,siga,sigb,chi2,q) is the Fortran signature.
		   Based on Numerical Recipies code of same name on p 508-9 in Fortran, 1st ed,
		   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
		   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering.
		   (C) 1986, Cambridge University Press.
		   y = a + bx.  And it returns stdeva, stdevb, and goodness of fit param q.
	*/
	var SumXoverSumWt, SumX, SumY, Sumt2, SumWt, chi2, a, b float64
	var result FittedData

	for _, p := range rows {
		wt := 1 / SQR(p.stdev)
		SumWt += wt
		SumX += p.x * wt
		SumY += p.y * wt
	}
	SumXoverSumWt = SumX / SumWt
	for _, p := range rows {
		t := (p.x - SumXoverSumWt) / p.stdev
		Sumt2 += SQR(t)
		b += t * p.y / p.stdev
	}
	b = b / Sumt2
	a = (SumY - SumX*b) / SumWt
	stdeva := math.Sqrt((1 + SQR(SumX)/(SumWt*Sumt2)) / SumWt)
	stdevb := math.Sqrt(1 / Sumt2)

	// Now to sum up Chi squared
	for _, p := range rows {
		chi2 += SQR((p.y - a - b*p.x) / p.stdev)
	}
	q := gammq(0.5*float64(len(rows)-2), 0.5*chi2)
	result.Slope = b
	result.StDevSlope = stdevb
	result.Intercept = a
	result.StDevIntercept = stdeva
	result.GoodnessOfFit = q
	return result
} // END fit

// -------------------------------- fitunwt ---------------------------------

func fitfull(row []Point, weighted bool) FittedData {

	/*
	   Based on Numerical Recipies code of same name on p 508-9 in Fortran,
	   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
	   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering.
	   (C) 1986, Cambridge University Press.
	   I think the docs are wrong.  The equation is y = a + bx, ie, b is Slope.  I'll make that switch now.
	*/
	var wt, a, b, t, q, sxoss, sx, sy, st2, ss, sigdat, chi2 float64
	var result FittedData

	if weighted { // means that there are variances for the data points
		for i := range row {
			wt = 1 / SQR(row[i].stdev)
			ss += wt
			sx += row[i].x * wt
			sy += row[i].y * wt
		}
	} else { // perform an unweighted sum
		for i := range row {
			sx += row[i].x
			sy += row[i].y
		}
		ss = float64(len(row))
	}
	sxoss = sx / ss
	if weighted {
		for i := range row {
			t = (row[i].x - sxoss) / row[i].stdev
			st2 += SQR(t)
			b += t * row[i].y / row[i].stdev
		}
	} else {
		for i := range row {
			t = row[i].x - sxoss
			st2 += SQR(t)
			b += t * row[i].y
		}
	}
	b = b / st2
	a = (sy - sx*b) / ss
	siga := math.Sqrt((1 + SQR(sx)/(ss*st2)) / ss)
	sigb := math.Sqrt(1 / st2)

	// Now to sum up Chi squared
	if !weighted {
		for i := range row {
			chi2 += SQR(row[i].y - a - b*row[i].x)
		}
		q = 1
		sigdat = math.Sqrt(chi2 / float64(len(row)-2))
		siga *= sigdat
		sigb *= sigdat
	} else {
		for i := range row {
			chi2 += SQR((row[i].y - a - b*row[i].x) / row[i].stdev)
		}
		q = gammq(0.5*float64(len(row)-2), 0.5*chi2)
	}
	result.Slope = b
	result.StDevSlope = sigb
	result.Intercept = a
	result.StDevIntercept = siga
	result.GoodnessOfFit = q
	return result
} // END fitfull

// -------------------------------- gamq ----------------------------------
func gammq(a, x float64) float64 {
	// Incomplete Gamma Function, is what "Numerical Recipies" says.
	// gln is the ln of the gamma function result

	if x < 0 || a <= 0 { // error condition, but I don't want to return an error result
		return 0
	}
	if x < a+1 { // use the series representation
		gamser, _ := gser(a, x)
		return 1 - gamser
	} else {
		gammcf, _ := gcf(a, x)
		return gammcf
	}
}

// ------------------------------- gser ----------------------------------
func gser(a, x float64) (float64, float64) {
	const ITERMAX = 100
	const tolerance = 3e-7

	gln := gammln(a)
	if x == 0 {
		return 0, 0
	} else if x < 0 { // invalid argument, but I'll just return 0 to not return an error result
		return 0, 0
	}
	ap := a
	sum := 1 / a
	del := sum
	for n := 1; n < ITERMAX; n++ {
		ap = ap + 1
		del *= x / ap
		sum += del
		if math.Abs(del) < math.Abs(sum)*tolerance {
			break
		}
	}
	return sum * math.Exp(-x+a*math.Log(x)-gln), gln
} // end gser

func gammln(xx float64) float64 {
	const stp = 2.50662827465
	const half = 0.5
	const fpf = 5.5

	var cof = [...]float64{76.18009173, -86.50532033, 24.01409822, -1.231739516, 0.120858003e-2, -0.536382e-5}
	var x, ser float64

	x = xx - 1
	tmp := x + fpf
	tmp = (x+half)*math.Log(tmp) - tmp
	ser = 1
	for j := range cof {
		x += 1
		ser += cof[j] / x
	}
	return tmp + math.Log(stp*ser)
} // end gammln

func gcf(a, x float64) (float64, float64) {
	const ITERMAX = 100
	const tolerance = 3e-7
	var gammcf, gln, gold, g, fac, b1, b0, anf, ana, an, a1, a0 float64

	gln = gammln(a)
	a0 = 1.0
	a1 = x
	b1 = 1.
	fac = 1.

	for n := 1; n < ITERMAX; n++ {
		an = float64(n)
		ana = an - a
		a0 = (a1 + a0*ana) * fac
		b0 = (b1 + b0*ana) * fac
		anf = an * fac
		a1 = x*a0 + anf*a1
		b1 = x*b0 + anf*b1
		if a1 != 0 {
			fac = 1 / a1
			g = b1 * fac
			if math.Abs((g-gold)/g) < tolerance {
				break
			}
			gold = g
		}
		gammcf = math.Exp(-x+a*math.Log(x)-gln) * g
	}
	return gammcf, gln
} // end gcf

// -------------------------------------------- Version 2 -----------------------------
// -------------------------------------------- Version 2 -----------------------------
// -------------------------------------------- Version 2 -----------------------------
//
// Section 15.3 -- Straight line data with errors in both coordinates.  P 660 ff.

// subroutine fitexy(x, y, sigx, sigy, a, b, siga, sigb, chi2, q float64, ndat int) {
func fitexy(rows []Point) FittedData2 { // (a, b, siga, sigb, chi2, q float64) {

	var result2 FittedData2
	// REAL x(ndat),y(ndat),sigx(ndat),sigy(ndat),a,b,siga,sigb,chi2,q,POTN,PI,BIG,ACC
	//	const NMAX = 1000.  Will use MaxN that I defined above.
	var PI float64 = math.Pi
	var r2 float64

	// Uses avevar,brent,chixy,fit,gammq,mnbrak,zbrent
	// Straight line fit to input data x(1..ndat) and y(1..ndat) with errors in both
	// x and y.  The respective std deviations being the input quantities sigx(1..ndat) and
	// sigy(1..ndat).  Output quantaties are a and b such that y = a + bx, minimizing Chi^2,
	// whose value is returned as chi2.  The chi^2 probability is returned as q in which a small value
	// indicating a poor fit.  Standard errors on a and b are returned as siga and sigb.  If
	// siga and sigb are returned as BIG, then the data are c/w all values of b.

	ndat := len(rows)
	nn = float64(ndat)
	//	REAL xx, yy, sx, sy, ww [ndat]float64 // dim of NMAX, or same as row slice
	var amx, amn, varx, vary float64 // aa and offs are now global.
	var ang, ch [7]float64           // to emulate Fortran one origin arrays.  I'll ignore element [0]
	//REAL scale,bmn,bmx,d1,d2,r2,dum1,dum3,dum3,dum4,dum5,brent,chixy,gammq,zbrent
	//COMMON /fitxyc/ xx,yy,sx,sy,ww,aa,ofs,nn
	if ndat > MaxN {
		fmt.Println(" Too many data points.  N =", ndat, ".  MaxN =", MaxN)
		os.Exit(1)
	}
	//	call avevar(x,ndat,dum1,varx)       Find x and y variances, and scale the data into
	//	call avevar(lny,ndat,dum1,vary)     the common block for communication with chixy.
	_, varx, _, vary = avevar(rows)
	scale := math.Sqrt(varx / vary)

	for j := range rows { // for j := 0; j < ndat; j++ {
		rows[j].xx = rows[j].x
		rows[j].yy = rows[j].y * scale
		rows[j].sx = rows[j].sigx
		rows[j].sy = rows[j].sigy * scale
		rows[j].ww = math.Sqrt(SQR(rows[j].sx) + SQR(rows[j].sy)) // use both x and y weights in first trial fit.
	}
	// subroutine fit(x,y,ndata,sig,mwt,a,b,siga,sigb,chi2,q) is the Fortran signature.
	result2 = fit2(rows) // fit(xx, yy, nn, ww, 1, dum1, b, dum2, dum3, dum4, dum5) as a trial fit for b.
	result2.scale = scale
	offs = 0
	ang[1] = 0
	ang[2] = math.Atan(result2.Slope)
	ang[4] = 0
	ang[5] = ang[2]
	ang[6] = POTN

	for j := 4; j <= 6; j++ {
		ch[j] = chixy(ang[j], rows)
	}

	//	call mnbrak(ang[1],ang[2],ang[3],ch[1],ch[2],ch[3],chixy) // Bracket the Chi squared minumum
	//	chi2 = brent(ang[1],ang[2],ang[3],chixy,ACC,b)            // and then locate it with brent.
	ang[1], ang[2], ang[3], ch[1], ch[2], ch[3] = mnbrak(ang[1], ang[2], rows)
	result2.chi2, result2.Slope = brent(ang[1], ang[2], ang[3], rows)
	//	fmt.Println(" after brent call.  slope is", result2.Slope)
	result2.chi2 = chixy(result2.Slope, rows)
	result2.Intercept = aa // aa is passed in the common block globally

	result2.q = gammq(0.5*(nn-2), 0.5*result2.chi2) // compute the Chi squared probability.  nn also passed in the common blk

	for j := range rows { // for j := 0; j < nn; j++ {
		r2 += rows[j].ww // save the inverse sum of weights at the minumum
	}

	r2 = 1 / r2

	bmx := BIG // find sandard errors for b as points where
	bmn := BIG // delta chi squared = 1.
	offs = result2.chi2 + 1

	for j := 1; j <= 6; j++ {
		if ch[j] > offs { // Go thru saved values to bracket the desired roots.
			d1 := math.Mod(math.Abs(ang[j]-result2.Slope), PI) // Note periodicity in slope angles
			d2 := PI - d1
			if ang[j] < result2.Slope {
				d1, d2 = d2, d1
			}
			if d1 < bmx {
				bmx = d1
			}
			if d2 < bmn {
				bmn = d2
			}
		}
	}

	a := result2.Intercept
	b := result2.Slope
	if bmx < BIG {
		bmx = zbrent(b, b+bmx, rows) - b // bmx = zbrent(chixy, b, b+bmx, ACC) - b
		amx = aa - a
		bmn = zbrent(b, b-bmn, rows) - b // bnm = zbrent(chixy, b, b-bmn, ACC) - b
		amn = aa - a
		result2.StDevSlope = math.Sqrt(0.5*(SQR(bmx)+SQR(bmn))) / (scale * SQR(math.Cos(b)))
		result2.StDevIntercept = math.Sqrt(0.5*(SQR(amx)+SQR(amn))+r2) / scale // error in a has additional piece r2.
	} else {
		result2.StDevSlope = BIG
		result2.StDevIntercept = BIG
	}
	a = a / scale
	b = math.Tan(b) / scale
	result2.Slope = b
	result2.Intercept = a
	return result2
} // end fitexy

func chixy(bang float64, row []Point) float64 {
	// Returns the value of Chi squared - offs, for the slope b = tan(bang).
	// scaled data and offs are communicated via the common block fitxyc
	var Chixy, avex, avey, sumw, b float64
	// COMMON /fitxyc/ xx,yy,sx,sy,ww,aa,offs,nn

	b = math.Tan(bang)
	for j := range row { // for j = 0; j < nn; j++ {
		row[j].ww = SQR(b*row[j].sx) + SQR(row[j].sy)
		if row[j].ww < 1/BIG {
			row[j].ww = BIG
		} else {
			row[j].ww = 1 / row[j].ww
		}
		sumw += row[j].ww
		avex += row[j].ww * row[j].xx
		avey += row[j].ww * row[j].yy
	}
	avex = avex / sumw
	avey = avey / sumw
	aa = avey - b*avex
	Chixy = -offs
	for _, p := range row {
		Chixy += p.ww * SQR(p.yy-aa-b*p.xx)
	}
	return Chixy
} // end chixy

func avevar(points []Point) (float64, float64, float64, float64) { // return mean, variance of both x and y.
	var avex, avey, // ave of x and y which is lny
		variancex, variancey, // variance of x and y
		epx, epy float64 // ep of x and y

	for j := range points { // for j :=  0; j < n; j++ {
		avex += points[j].x
		avey += points[j].y
	}
	//	nn := len(points) but this is a float64 and is global.
	avex = avex / nn
	avey = avey / nn

	for j := range points { // for j := 0; j < n; j++ {
		points[j].sx = points[j].x - avex
		epx += points[j].sx
		variancex += SQR(points[j].sx)

		points[j].sy = points[j].y - avey
		epy += points[j].sy
		variancey += SQR(points[j].sy)
	}
	variancex = (variancex - SQR(epx)/nn) / (nn - 1)
	variancey = (variancey - SQR(epy)/nn) / (nn - 1)
	return avex, variancex, avey, variancey
} // end avevar

func SignTransfer(a, b float64) float64 {
	// Sign Transfer function from Fortran returns the first argument with the sign of the 2nd.
	neg := math.Signbit(b)
	if neg {
		return -math.Abs(a)
	} else {
		return math.Abs(a)
	}
}

func mnbrak(ax, bx float64, rows []Point) (float64, float64, float64, float64, float64, float64) {
	// given a function func, and given distinct initial points ax and bx, this routine
	// searches in the downhill direction and returns new points ax, bx, cx that bracket a
	// minimum of the function.  Also returned are the function values at the three points,
	// fa, fb, and fc.  Here, I use chixy as the function.

	const GOLD = 1.618034 // default ratio by which successive intervals are magnified
	const GLIMIT = 100    // maximum magnification allowed for a parabolic-fit step
	const TINY = 1e-20

	var cx, fa, fb, fc float64    // REAL ax,bx,cx,fa,fb,fc
	var fu, q, r, u, ulim float64 // REAL dum,fu,q,r,u,ulim

	fa = chixy(ax, rows)
	fb = chixy(bx, rows)
	if fb > fa { // Switch a and b so that we can go downhill from a to b.
		ax, bx = bx, ax
		fa, fb = fb, fa
	}
	cx = bx + GOLD*(bx-ax)
	fc = chixy(cx, rows)

	for { // repeat until we bracket
		if fb >= fc {
			r = (bx - ax) * (fb - fc) // compute u by parabolic extrapolation from a,b,c
			q = (bx - cx) * (fb - fa) // TINY is used to prevent any possible div by zero.
			u = bx - ((bx-cx)*q-(bx-ax)*r)/(2*SignTransfer(math.Max(math.Abs(q-r), TINY), q-r))
			ulim = bx + GLIMIT*(cx-bx) // Won't go farther than this.  Test possibilities
			if (bx-u)*(u-cx) > 0 {     // parabolic is btwn b and c.
				fu = chixy(u, rows)
				if fu < fc { // got a minimum btwn b and c
					ax = bx
					fa = fb
					bx = u
					fb = fu
					return ax, bx, cx, fa, fb, fc
				} else if fu > fb { // got a minimum btwn a and u
					cx = u
					fc = fu
					return ax, bx, cx, fa, fb, fc
				} // endif fu < fc or fu > fb
				u = cx + GOLD*(cx-bx)
				fu = chixy(u, rows)
			} else if (cx-u)*(u-ulim) > 0 { // parabolic fit is btwn c and its allowed limit
				fu = chixy(u, rows)
				if fu < fc {
					bx = cx
					cx = u
					u = cx + GOLD*(cx-bx)
					fb = fc
					fc = fu
					fu = chixy(u, rows)
				} //endif fu < fc
			} else if (u-ulim)*(ulim-cx) >= 0 { // limit parabolic u to maximum allowed value
				u = ulim
				fu = chixy(u, rows)
			} else { // reject parabolic u, use default magnification
				u = cx + GOLD*(cx-bx)
				fu = chixy(u, rows)
			} // endif (bx-u)*(u-cx) or (cx-u)*u-ulim) etc
			ax = bx
			bx = cx
			cx = u
			fa = fb
			fb = fc
			fc = fu
		} else { // I had to add this break to emulate the essence of the Fortran code
			break
		} // endif fb >= fc
	} // ENDFOR
	return ax, bx, cx, fa, fb, fc
} // end mnbrak

func brent(ax, bx, cx float64, rows []Point) (float64, float64) {
	// Given a function, chixy, and given a bracketing tiplet of abscissas ax,bx,cx such that
	// bx is btwn ax and cx, and chixy(bx) is less than both chixy(ax) and chixy(cx), this
	// routine isolates the minimum to a fractional precision of about tolerance using Brent's
	// method.  The abscissa of the minimum is returned as xmin, and the minimum function
	// value is returned as Brent.
	const tolerance = ACC
	const ITERMAX = 100
	const CGOLD = 0.3819660 // golden ratio
	const ZEPS = 1e-10      // small number that protects against trying to achieve fractional accuracy that happens to be exactly zero

	var a, b, d, e, etemp, fu, fv, fw, fx, p, q, r, u, v, w, x, xm, tol1, tol2 float64

	a = math.Min(ax, cx) // a and b must be in ascending order, though the input abscissas need not be.
	b = math.Max(ax, cx)
	v = bx
	w = v
	x = v
	e = 0 // distance moved on the step before last
	fx = chixy(x, rows)
	fv = fx
	fw = fx
	d = 1                                   // I think this omittion is a book printing error.  I'm assuming it should be 1 and not 0.
	for iter := 0; iter < ITERMAX; iter++ { // main routine loop
		xm = 0.5 * (a + b)
		tol1 = tolerance*math.Abs(x) + ZEPS
		tol2 = 2 * tol1
		if math.Abs(x-xm) <= (tol2 - 0.5*(b-a)) {
			break // goto 3 in orig Fortran code -- done
		}
		if math.Abs(e) > tol1 { // construct a trial parabolic fit
			r = (x - w) * (fx - fv)
			q = (x - v) * (fx - fw)
			p = (x-v)*q - (x-w)*r
			q = 2 * (q - r)
			if q > 0 {
				p = -p
			}
			q = math.Abs(q)
			etemp = e
			e = d
			if math.Abs(p) >= math.Abs(0.5*q*etemp) || p <= q*(a-x) || p >= q*(b-x) {
				goto Label1 // These conditions determine the acceptability of the parabolic fit.  It is ok to proceed.
			}
			d = p / q
			u = x + d
			if u-a < tol2 || b-u < tol2 {
				d = SignTransfer(tol1, xm-x)
			}
			goto Label2 // skip over the golden section step
		}
	Label1: // Arrive here for a golden section step, which we take into the larger of the 2 segments
		if x >= xm {
			e = a - x
		} else {
			e = b - x
		}
		d = CGOLD * e // take the golden section step

	Label2: // Arrive here with d computed either from parabolic fit or else from golden section
		if math.Abs(d) >= tol1 {
			u = x + d
		} else {
			u = x + SignTransfer(tol1, d)
		}
		fu = chixy(u, rows)
		if fu <= fx {
			if u >= x {
				a = x
			} else {
				b = x
			}
			v = w
			fv = fw
			w = x
			fw = fx
			x = u
			fx = fu
		} else {
			if u < x {
				a = u
			} else {
				b = u
			}
			if fu <= fw || w == x {
				v = w
				fv = fw
				w = u
				fw = fu
			} else if fu <= fv || v == x || v == w {
				v = u
				fv = fu
			}
		}
	}
	//	xmin = x
	//	Brent = fx
	return fx, x // looks like both of these never get a value.  That can't be right.
} // end brent

// Using brent's method find the root of a function know to lie btwn x1 and x2.  The root,
// returned as Zbrent, will be refined until its accuracy is ACC.
func zbrent(x1, x2 float64, rows []Point) float64 {
	const tolerance = ACC
	const EPS = 3e-8
	var d, e, p, q, r, s, xm float64

	a := x1
	b := x2
	fa := chixy(a, rows)
	fb := chixy(b, rows)
	if (fa > 0 && fb > 0) || (fa < 0 && fb < 0) {
		fmt.Println(" Root must be bracketed for zbrent, whatever that means.")
	}
	c := b
	fc := fb

	for iter := 0; iter < ITERMAX; iter++ {
		if (fb > 0 && fc > 0) || (fb < 0 && fc < 0) { // rename a,b,c and adjust bounding interval d
			c = a
			fc = fa
			d = b - a
			e = d
		}
		if math.Abs(fc) < math.Abs(fb) {
			a = b
			b = c
			c = a
			fa = fb
			fb = fc
			fc = fa
		}
		tol1 := 2*EPS*math.Abs(b) + 0.5*tolerance // convergence check
		xm = 0.5 * (c - b)
		if math.Abs(xm) <= tol1 || fb == 0 {
			return b
		}
		if math.Abs(e) >= tol1 && math.Abs(fa) > math.Abs(fb) {
			s = fb / fa // attempt inverse quadratic interpolation
			if a == c {
				p = 2 * xm * s
				q = 1 - s
			} else {
				q = fa / fc
				r = fb / fc
				p = s * (2*xm*q*(q-r) - (b-a)*(r-1))
				q = (q - 1) * (r - 1) * (s - 1)
			}
			if p > 0 { // check whether in bounds
				q = -q
			}
			p = math.Abs(p)
			if 2*p < math.Min(3*xm*q-math.Abs(tol1*q), math.Abs(e*q)) {
				e = d // accept interpolation
				d = p / q
			} else { // interpolation failed, use bisection
				d = xm
				e = d
			}
		} else { // bounds decresing too slowly, use bisection
			d = xm
			e = d
		}
		a = b
		fa = fb
		if math.Abs(d) > tol1 { // evaluate new trial root
			b += d
		} else {
			b += SignTransfer(tol1, xm)
		}
		fb = chixy(b, rows)
	}
	return b
} // end zbrent

func fit2(rows []Point) FittedData2 {
	/*
		subroutine fit(x,y,ndata,sig,mwt,a,b,siga,sigb,chi2,q) is the Fortran signature.
		   Based on Numerical Recipies code of same name on p 508-9 in Fortran, 1st ed,
		   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
		   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering.
		   (C) 1986, Cambridge University Press.
		   y = a + bx.  And it returns stdeva, stdevb, and goodness of fit param q.
		   fit2 uses the xx,yy,sx,sy,ww fields defined in fitexy.
	*/
	var SumXoverSumWt, SumX, SumY, Sumt2, SumWt, chi2, a, b float64
	var result2 FittedData2

	for _, p := range rows {
		wt := 1 / SQR(p.sy)
		SumWt += wt
		SumX += p.xx * wt
		SumY += p.yy * wt
	}
	SumXoverSumWt = SumX / SumWt
	for _, p := range rows {
		t := (p.xx - SumXoverSumWt) / p.sy
		Sumt2 += SQR(t)
		b += t * p.yy / p.sy
	}
	b = b / Sumt2
	a = (SumY - SumX*b) / SumWt
	stdeva := math.Sqrt((1 + SQR(SumX)/(SumWt*Sumt2)) / SumWt)
	stdevb := math.Sqrt(1 / Sumt2)

	// Now to sum up Chi squared
	for _, p := range rows {
		chi2 += SQR((p.yy - a - b*p.xx) / p.sy)
	}
	result2.q = gammq(0.5*float64(len(rows)-2), 0.5*chi2)
	result2.Slope = b
	result2.StDevSlope = stdevb
	result2.Intercept = a
	result2.StDevIntercept = stdeva
	return result2
} // END fit2

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
