// GastricEmptying in Go.  (C) 2017.  Based on GastricEmtpying2.mod code.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	//
	"getcommandline"
	"tknptr"
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
*/

const LastAltered = "16 Aug 2017"

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

// stdev relates to counts, or the ordinate.  This algorithm has it apply to the abscissa
// also.  I'll run with this and see what happens.
type Point struct {
	x, y, lny, stdev float64
}

type inputrow []float64
type inputmatrix []inputrow
type FittedData struct {
	Slope, Intercept, StDevSlope, StDevIntercept, GoodnessOfFit float64
}

//  WEIGHT,AX : ARRAY[0..MAXELE] OF LONGREAL;
//  PREVSLOPE,PREVINTRCPT : LONGREAL;
//  C,K,ITERCTR : CARDINAL;

//************************************************************************
//*                              MAIN PROGRAM                            *
//************************************************************************

func main() {
	var point Point
	ln2 := math.Log(2)
	rows := make([]Point, 0, MaxCol)
	im := make(inputmatrix, 0, MaxN) // input matrix

	fmt.Println()
	fmt.Println(" Gastric Emtpying program written in Go.  Last modified", LastAltered)
	fmt.Println()
	if len(os.Args) <= 1 {
		fmt.Println(" Usage: GastricEmptying <filename>")
		os.Exit(0)
	}
	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Println(ExecFI.Name(), " was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
	fmt.Println(" Full name of executable file is", execname)
	fmt.Println()

	InExtDefault := ".txt"
	OutExtDefault := ".out"

	ns := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(ns)
	Filename := ""
	FileExists := false

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

	byteslice := make([]byte, 0, 5000)
	byteslice, err := ioutil.ReadFile(Filename)
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
		//		fmt.Println(" input row is ", ir, ", col is", col)  for debugging
		if col >= 1 { // process line as it has 2 numbers
			im = append(im, ir)
			N++
		}
	} // END main input reading loop

	// output im for debugging
	// fmt.Println(" inputmatrix IM is", im)

	// Now need to populate the Time And Counts Table
	for c := range im {
		point.x = im[c][0]
		point.y = im[c][1]
		point.lny = math.Log(point.y)
		point.stdev = 10
		rows = append(rows, point)
	}

	// fmt.Println(" Date and Time in default format:", date)
	s := fmt.Sprintf(" Date and Time in basic format: %s \n", datestring)
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	fmt.Println(" N = ", len(rows))
	fmt.Println()
	fmt.Println(" X is time(min)  Y is kcounts  Ln(y)     stdev")
	fmt.Println()
	for _, p := range rows {
		s := fmt.Sprintf("%11.0f %13.2f %10.4f %10.4f\n", p.x, p.y, p.lny, p.stdev)
		fmt.Print(s)
		_, err = OutBufioWriter.WriteString(s)
		check(err)
	}
	fmt.Println()
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	stdslope, stdintercept, stdr2 := StdLR(rows)
	stdhalflife := -ln2 / stdslope

	//	fmt.Println(" Original standard unweighted Slope", stdslope, ", standard Intercept is", stdintercept)
	//	fmt.Println(" standard R-squared Correlation Coefficient", stdr2)
	s = fmt.Sprintf(" Original T-1/2 of Gastric Emptying is %.2f minutes.  Original std unweighted slope is %.6f and std intercept is %.6f and R-squared is %.6f \n", stdhalflife, stdslope, stdintercept, stdr2)
	fmt.Print(s)
	fmt.Println()
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	WeightedResults := fit(rows)
	weightedhalflife := -ln2 / WeightedResults.Slope
	//	fmt.Println("Weighted Slope is", WeightedResults.Slope, ", Weighted Intercept is", WeightedResults.Intercept)
	//	fmt.Println("stdev Slope is", WeightedResults.StDevSlope, ", stdev Intercept is", WeightedResults.StDevIntercept)
	//	fmt.Println("GoodnessOfFit is", WeightedResults.GoodnessOfFit)
	s = fmt.Sprintf(" Weighted halflife of Gastric Emptying is %.2f minutes.  Slope=%.6f, intercept=%.6f, StDevSlope=%.6f, StDevIntercept=%.6f, GoodnessOfFit=%.6f \n",
		weightedhalflife, WeightedResults.Slope, WeightedResults.Intercept, WeightedResults.StDevSlope, WeightedResults.StDevIntercept, WeightedResults.GoodnessOfFit)
	fmt.Println(s)
	fmt.Println()
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	UnWeightedResults := fitfull(rows, false)
	Unweightedhalflife := -ln2 / UnWeightedResults.Slope
	//	fmt.Println("unWeighted Slope is", UnWeightedResults.Slope, ", Intercept is", UnWeightedResults.Intercept)
	//	fmt.Println("stdev Slope is", UnWeightedResults.StDevSlope, ", stdev Intercept is", UnWeightedResults.StDevIntercept)
	//	fmt.Println("GoodnessOfFit is", UnWeightedResults.GoodnessOfFit)
	s = fmt.Sprintf(" unweighted halflife of Gastric Emptying is %.2f minutes.  Slope=%.6f, Intercept=%.6f, StDevSlope=%.6f, StDevIntercept=%.6f. \n",
		Unweightedhalflife, UnWeightedResults.Slope, UnWeightedResults.Intercept, UnWeightedResults.StDevSlope, UnWeightedResults.StDevIntercept)
	fmt.Println(s)
	fmt.Println()
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	/* This code works but is redundant.  So I'll remove it.
	WeightedResults2 := fitfull(rows, true)
	weightedhalflife2 := -ln2 / WeightedResults2.Slope
	fmt.Println("Weighted Slope is", WeightedResults2.Slope, ", Weighted Intercept is", WeightedResults2.Intercept)
	fmt.Println("stdev Slope is", WeightedResults2.StDevSlope, ", stdev Intercept is", WeightedResults2.StDevIntercept)
	fmt.Println("GoodnessOfFit is", WeightedResults2.GoodnessOfFit)
	s = fmt.Sprintf(" halflife of Gastric Emptying using Weights2 is %.2f minutes. \n", weightedhalflife2)
	fmt.Println(s)
	fmt.Println()
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	*/
	// The files will close themselves because of the defer statements.
}

// -------------------------------------- SQR ---------------------------------------------
func SQR(R float64) float64 {
	return R * R
} // END SQR;

//  ------------------------------------- StdLR ----------------------------------
func StdLR(rows []Point) (float64, float64, float64) {
	/*
	   This routine does the standard, unweighted, computation of the slope and
	   intercept, using the formulas that are built into many pocket calculators,
	   including mine.  This computation serves as an initial guess for the
	   iterative solution used by this program as described by Dr.  Zanter.
	*/

	/*
	   First does the simple (unweighted) sums required of the std formula.  And remember that I'm linearizing
	   the data by using ln(y), not y itself.
	*/
	var sumx, sumy, sumx2, sumy2, sumxy float64

	N := float64(len(rows))
	for _, p := range rows { // FOR c := 0 TO N-1 DO
		sumx += p.x
		sumy += p.lny
		sumxy += p.x * p.lny
		sumx2 += SQR(p.x)
		sumy2 += SQR(p.lny)
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
	   Based on Numerical Recipies code of same name on p 508-9 in Fortran,
	   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
	   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering.
	   (C) 1986, Cambridge University Press.
	   y = ax + b.  And it returns stdeva, stdevb, and goodness of fit param q.
	   I think the docs are wrong.  The equation is y = a + bx, ie, b is Slope.  I'll make that switch now.
	*/
	var SumXoverSumWt, SumX, SumY, Sumt2, SumWt, chi2, a, b float64
	var result FittedData

	for _, p := range rows {
		wt := 1 / SQR(p.stdev)
		SumWt += wt
		SumX += p.x * wt
		SumY += p.lny * wt
	}
	SumXoverSumWt = SumX / SumWt
	for _, p := range rows {
		t := (p.x - SumXoverSumWt) / p.stdev
		Sumt2 += SQR(t)
		b += t * p.lny / p.stdev
	}
	b = b / Sumt2
	a = (SumY - SumX*b) / SumWt
	stdeva := math.Sqrt((1 + SQR(SumX)/(SumWt*Sumt2)) / SumWt)
	stdevb := math.Sqrt(1 / Sumt2)

	// Now to sum up Chi squared
	for _, p := range rows {
		chi2 += SQR((p.lny - a - b*p.x) / p.stdev)
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
	   William H. Press, Brian P. Flannery, Saul A. Teukolsky, William T. Vettering.
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
			sy += row[i].lny * wt
		}
	} else { // perform an unweighted sum
		for i := range row {
			sx += row[i].x
			sy += row[i].lny
		}
		ss = float64(len(row))
	}
	sxoss = sx / ss
	if weighted {
		for i := range row {
			t = (row[i].x - sxoss) / row[i].stdev
			st2 += SQR(t)
			b += t * row[i].lny / row[i].stdev
		}
	} else {
		for i := range row {
			t = row[i].x - sxoss
			st2 += SQR(t)
			b += t * row[i].lny
		}
	}
	b = b / st2
	a = (sy - sx*b) / ss
	siga := math.Sqrt((1 + SQR(sx)/(ss*st2)) / ss)
	sigb := math.Sqrt(1 / st2)

	// Now to sum up Chi squared
	if !weighted {
		for i := range row {
			chi2 += SQR(row[i].lny - a - b*row[i].x)
		}
		q = 1
		sigdat = math.Sqrt(chi2 / float64(len(row)-2))
		siga *= sigdat
		sigb *= sigdat
	} else {
		for i := range row {
			chi2 += SQR((row[i].lny - a - b*row[i].x) / row[i].stdev)
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
}

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
}

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
}
