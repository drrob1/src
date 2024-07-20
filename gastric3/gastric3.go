// GastricEmptying in Go, V 3.  V1 was the first version.
// V2 added calculating errors in X and Y which were then used to determine the reported T-1/2.
// V3 added the old iterative solution I learned about in the 80's.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"gonum.org/v1/gonum/stat"
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
------------------------------------------------------------------------------------------------------------------------
10 Aug 17 -- Converting to Go
12 Aug 17 -- Added "Numerical Recipies" code, and removed old iterative algorithm.
15 Aug 17 -- Added ability to show timestamp of the executable file, as a way of showing last linking timestamp.
             To be used in addition of the LastAltered string.  But I can recompile without altering, as when
             a new version of the Go toolchain is released.
 9 Sep 17 -- Changing how bufio errors are checked, based on a posting from Rob Pike.
11 Sep 17 -- Made the stdev a % factor instead of a constant factor, and tweaked the output in other ways.
12 Sep 17 -- Added heading to output file.
------------------------------------------------------------------------------------------------------------------------
13 Sep 17 -- Adding code from Numerical Recipes for errors in X and Y.  And now called gastric2.
24 Sep 17 -- To make the new code work, I'll remove lny and use OrigY and y.
27 Sep 17 -- It works after I fixed some typos.  And I changed the order of the output values.
 2 Oct 17 -- Discovered that in very normal patients, counts are low enough for stdev to be neg.  I can't allow that!
             And error in X (time) will be same %-age as in Y.  And StDevY cannot be larger than y.
 3 Oct 17 -- Added separate StDevTime constant, and set a minimum of 2.  Output optimized for Windows since that
             is where it will be used in "production."
18 Oct 17 -- Added filepicker
-------------------------------------------------------------------------------------------------------------------------------
15 Jun 18 -- Now at version 3, and I'm revisiting the old iterative solution that I used for a long time.  Is there a bug here?
18 Jun 18 -- Added some comments regarding my thoughts and observations, and decided to return FittedData3 so I can get R2.
             And screen output will be shorter than file output.
20 Jun 18 -- Added comments regarding reference normal.
21 Jun 18 -- Added first non digit token on line will skip that line.  So entire line can be easily commented out.
 4 Nov 18 -- Realized that the intercept is log(counts), so exp(counts) is worth displaying.
 6 Nov 18 -- Will also account for a lag time.
19 Nov 19 -- Will start coding an automatic detection for the lag period by looking for a local peak in counts.
31 Jul 20 -- Coding use of gonum.org, just for experience.
 2 Aug 20 -- Got idea to vary uncertainty in counts as a function of i, so there is more uncertainty in later points.
             And removed the check for error > number itself.  I think the results are better.
             Nevermind.  I'm reverting this change because the T-1/2 was higher than the orig in many runs of this.
             So instead of changing the error values, I'll change the weights for the gonum computation.
             And I added an unweighted run of gonum's LinearRegression.
29 Dec 20 -- Now that new Siemens camera is installed, I want to be able to use both heads and the geometric mean.
             So I have to read both count points on the line, assuming first is anterior and 2nd is posterior.
             And changed to more idiomatic Go for the slices.
20 Jan 21 -- Will issue a warning if a line does not have at least 2 points.  And it's inauguration day.  But that's not important now.
10 Aug 21 -- Fixed minor output bug in format statement for a line only output to the file and not to screen.  Forgot to use exp() and newline char.
               and converted to modules.
22 Oct 21 -- Removed ioutil, which was deprecated as of Go 1.16.
23 Oct 21 -- Stopped pre-allocating the slice of file contents.
11 Apr 22 -- Modernizing the filepicker output.  Changed from bytes.buffer to bytes.reader.  Added comment characters of # and / like for BJ strategy files.
19 Jul 24 -- Adding use of MultiWriter for both file and screen output.  And modified code to comply w/ new API for tknptr, ie, New instead of NewToken.
			Decided to not add MultiWriter, as I prefer having different output to file and screen.  To the screen I'm using %.0f to simplify the output.
			Adding color fmt routines.
20 Jul 24 -- Writing to the file is now using %.0f instead of %.2f.  It does not make sense to report fractional minutes, not even in the results file.
*/

const LastAltered = "July 20, 2024"

/*
  Normal values from source that I don't remember anymore.
  1 hr should have 90% of activity remaining in stomach = 6.6 hr halflife = 395 min halflife
  2 hr should have 60% of activity remaining in stomach = 2.7 hr halflife = 163 min halflife
  3 hr should have 30% of activity remaining in stomach = 1.7 hr halflife = 104 min halflife
  4 hr should have 10% of activity remaining in stomach = 1.2 hr halflife = 72 min halflife
  Using these 4 points as data for gastric3, I get T-1/2 of Orig std unweighted = 72.91,
  fit weighted= 60.65, fitexy weighted= 60.76, and iterated weighted= 75.73
*/

//----------------------------------------------------------------------------

const MaxN = 500
const MaxCol = 10
const StDevFac = 25 // treated as a percentage for weights used for GastricGo v1 code.
const StDevTime = 10
const POTN = 1.571000         // used for fitexy and related routines
const BIG = 1e30              // this too
const ACC = 1e-3              // this too
const ITERMAX = 100           // this too, and now also for the old ressurrected code.
const ToleranceFactor = 1.e-5 // used for old code that's been ressurrected.

// stdev relates to counts, or the ordinate.  This algorithm used to apply it to the abscissa also.
// Version 2 handles these separately in fitexy.

type Point struct {
	x, y, OrigY, stdev,
	sigx, sigy, xx, yy, sx, sy, ww, // added when I added fitexy in Sep 2017, v 2.
	weight, ax float64 // added when I added old code in June 2018, v 3.
}

type inputrow []float64
type inputmatrix []inputrow
type FittedData struct {
	Slope, Intercept, StDevSlope, StDevIntercept, GoodnessOfFit float64
}

type FittedData2 struct {
	Slope, Intercept, StDevSlope, StDevIntercept, chi2, q, scale float64
}

type FittedData3 struct {
	Slope, Intercept, SumWt, SumWtX, SumWtX2, SumWtXY, SumWtY, SumWtY2, SumWtAx, SumWtAxX, SumWtAxY, R2 float64
}

var aa, offs, nn float64 // used by the Numerical Recipies book.

//------------------------------------------------------------------------
//-                              MAIN PROGRAM                            -
//------------------------------------------------------------------------

func main() {
	var point Point
	var err, bufioErr error
	var BaseFilename string

	ln2 := math.Log(2)
	rows := make([]Point, 0, MaxCol)
	im := make(inputmatrix, 0, MaxN) // type inputmatrix []inputrow

	fmt.Println()
	fmt.Println(" GastricGo v 3, Gastric Emtpying program written in Go.  Last modified", LastAltered)
	fmt.Println()
	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf("%s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n", ExecFI.Name(), LastLinkedTimeStamp, workingdir, execname)
	fmt.Println()

	InExtDefault := ".txt"
	OutExtDefault := ".out"
	ans := ""
	Filename := ""
	FileExists := false

	if len(os.Args) <= 1 { // need to use filepicker
		filenames, err := filepicker.GetFilenames("gastric*.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepicker using pattern of gastric*.txt is %v.  Exiting.\n", err)
			os.Exit(1)
		}

		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		_, err = fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil {
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

	byteSlice, err := os.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error", err, " from iotuil.ReadFile when reading", Filename, ".  Exiting.")
		os.Exit(1)
	}

	bytesReader := bytes.NewReader(byteSlice)

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

	for { // Main input reading loop to read all lines
		line, e := readLine(bytesReader) // bytesbuffer.ReadString('\n')  A modernization.
		if e != nil {
			break
		}
		tokenReader := tknptr.New(line)
		ir := make(inputrow, 0, 3) // type inputrow []float64; input row for time, ant counts, post counts.

		// loop to process first 2-digit tokens on this line and ignore the rest
		for {
			token, EOL := tokenReader.GETTKNREAL()
			if EOL || token.State != tknptr.DGT { // better way to handle non digit tokens.
				break
			}
			if token.State == tknptr.DGT { // ignore non-digit tokens, which allows for comments
				ir = append(ir, token.Rsum)
				//ir[col] = token.Rsum // ir is input row
				//col++
			}
		} // UNTIL EOL OR token is not a digit token
		if len(ir) > 1 {
			im = append(im, ir)
		} else if len(ir) == 1 {
			fmt.Println(" Data line only has one entry, likely time but no counts.  Line ignored.")
		}
	} // END main input reading loop

	fmt.Printf(" Input Matrix by rows.  Number of rows = len(im) = %d \n", len(im))
	for i := range im {
		fmt.Println(im[i])
	}
	fmt.Println()
	fmt.Println()

	// Now need to populate the Time And Counts Table called rows.
	// And now to construct xvector, yvector and wtvector for gonum.org stats package.
	lenvector := len(im)
	xvector := make([]float64, 0, lenvector)
	yvector := make([]float64, 0, lenvector)
	wtvector := make([]float64, 0, lenvector)
	unwtvector := make([]float64, 0, lenvector)
	for c := range im {
		point.x = im[c][0]
		point.OrigY = im[c][1] // OrigY will have either the anterior counts, or the geometric mean of anterior and posterior counts, if available.
		if len(im[c]) > 2 {
			point.OrigY = math.Sqrt(point.OrigY * im[c][2])
		}

		point.y = math.Log(point.OrigY)
		point.stdev = math.Abs(math.Log(point.OrigY * StDevFac / 100)) // treating StDevFac as a %-age.
		point.sigx = point.x * StDevTime / 100
		if point.sigx < 2 {
			point.sigx = 2
		}
		point.sigy = point.stdev
		rows = append(rows, point)

		// added for gonum stats package
		xvector = append(xvector, point.x)
		yvector = append(yvector, point.y)
		weight := 1 + float64(c)/10 // these are weights, not uncertainties, so I'll try something so that the later points have more weight.
		wtvector = append(wtvector, weight)
		unwtvector = append(unwtvector, 1.0)
	}

	// this is a closure.
	writestr := func(s string) {
		if bufioErr != nil {
			return
		}
		_, bufioErr = OutBufioWriter.WriteString(s)
	}

	writerune := func() { // this is also a closure.
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

	if len(rows) == 0 {
		fmt.Println(" No valid rows found.  Exiting")
		os.Exit(1)
	} else if len(rows) < 2 {
		fmt.Println(" Only 1 valid line found.  Exiting.")
		os.Exit(1)
	}

	fmt.Println(" N = ", len(rows))
	fmt.Println()
	fmt.Println(" i  X is time(min)  Y is kcounts  Ln(y)     X StDev   Y stdev   Wt Vector ")
	fmt.Println(" =  ==============  ============  =====     =======   =======   =========")
	writestr(" i  X is time(min)  Y is kcounts  Ln(y)     X StDev   Y stdev   Wt Vector") // the closure from above
	writerune()
	for i, p := range rows {
		// The last format verb has a smaller width to look more balanced when displayed, as only 1 decimal place is shown.
		s := fmt.Sprintf("%2d %11.0f %13.2f %10.4f %10.4f %10.4f %8.1f\n",
			i, p.x, p.OrigY, p.y, p.sigx, p.stdev, wtvector[i])
		fmt.Print(s)
		writestr(s) // the closure from above
	}
	fmt.Println()
	fmt.Println()
	writerune()
	check(bufioErr)

	// Now to calculate regressions

	stdslope, stdintercept, stdr2 := StdLR(rows)
	stdhalflife := -ln2 / stdslope
	ctfmt.Printf(ct.Yellow, true, " Original unweighted: halflife is %.0f minutes, and exp(intercept) is %.0f counts.\n", stdhalflife, math.Exp(stdintercept))
	s = fmt.Sprintf(" Original std unweighted: T-1/2 of Gastric Emptying is %.0f minutes, slope is %.6f cnts/min, exp(intercept) is %.0f cnts and R-squared is %.6f.",
		stdhalflife, stdslope, math.Exp(stdintercept), stdr2)
	writestr(s) // using the write closure
	writerune()
	check(bufioErr)

	UnWeightedResults := fitfull(rows, false)
	UnweightedHalfLife := -ln2 / UnWeightedResults.Slope
	// not writing this output to screen as its redundant.
	s = fmt.Sprintf(" fitful unweighted: halflife of Gastric Emptying is %.0f minutes.  Slope= %.6f cnts/min, exp(intercept)= %.0f, StDevSlope= %.6f cnts.",
		UnweightedHalfLife, UnWeightedResults.Slope, math.Exp(UnWeightedResults.Intercept), UnWeightedResults.StDevSlope)
	writestr(s) // using the write closure, I hope
	writerune()
	check(bufioErr)

	WeightedResults := fit(rows)
	weightedHalfLife := -ln2 / WeightedResults.Slope
	// not writing this output to screen as its redundant.
	s = fmt.Sprintf(" fit weighted: halflife of Gastric Emptying is %.0f minutes.  Slope= %.6f, exp(Intercept)= %.0f, StDevSlope= %.6f, GoodnessOfFit= %.6f.",
		weightedHalfLife, WeightedResults.Slope, exp(WeightedResults.Intercept), WeightedResults.StDevSlope, WeightedResults.GoodnessOfFit)
	writestr(s) // using the write closure
	writerune()
	check(bufioErr)

	WeightedResults2 := fitfull(rows, true)
	weightedHalfLife2 := -ln2 / WeightedResults2.Slope
	ctfmt.Printf(ct.Green, true, " fitful weighted: halflife is %.0f minutes, exp(intercept) is %.0f counts.\n", weightedHalfLife2, exp(WeightedResults2.Intercept))
	s = fmt.Sprintf(" fitful weighted: halflife of Gastric Emptying is %.0f minutes, slope is %.6f, exp(intercept) = %.0f, fit is %.6f.",
		weightedHalfLife2, WeightedResults2.Slope, exp(WeightedResults2.Intercept), WeightedResults2.GoodnessOfFit)
	writestr(s)
	writerune()
	check(bufioErr)

	WeightedResults3 := fitexy(rows)
	weightedHalfLife3 := -ln2 / WeightedResults3.Slope
	ctfmt.Printf(ct.Yellow, false, " fitexy weighted: halflife is %.0f minutes, Y-intercept is %.0f counts.\n ", weightedHalfLife3, exp(WeightedResults3.Intercept))
	ctfmt.Printf(ct.Cyan, true, "fitexy uses errors in X and Y to compute results.\n")
	s = fmt.Sprintf(" fitexy (errors in X and Y used to refine results) weighted: halflife of Gastric Emptying is %.0f minutes, slope is %.6f, Y-intercept = %.0f counts, stdev is %.6f, chi2= %.6f, q= %.6f.",
		weightedHalfLife3, WeightedResults3.Slope, exp(WeightedResults3.Intercept), WeightedResults3.StDevSlope, WeightedResults3.chi2, WeightedResults3.q)
	writestr(s)
	writerune()
	check(bufioErr)

	// Version 3 code, adding the orig weighting function
	IteratedResults := DoOldWeightedLR(rows, stdslope, stdintercept)
	IteratedHalfLife := -ln2 / IteratedResults.Slope
	ctfmt.Printf(ct.Green, true, " Old iterative method: halflife is %.0f minutes, exp(intercept) is %.0f.\n", IteratedHalfLife, exp(IteratedResults.Intercept))
	s = fmt.Sprintf(" Old iterative method: halflife of Gastric Emptying is %.0f minutes, slope is %.6f, Y-intercept = %.0f counts, R^2 = %.6f.",
		IteratedHalfLife, IteratedResults.Slope, exp(IteratedResults.Intercept), IteratedResults.R2)
	writestr(s)
	writerune()
	check(bufioErr)

	interceptStats, slopeStats := stat.LinearRegression(xvector, yvector, wtvector, false)
	halfLifeStats := -ln2 / slopeStats
	s0 := fmt.Sprintf(" gonum.org LinearRegression: halflife is %.0f minutes, Y-intercept is %.0f counts. \n", halfLifeStats, exp(interceptStats))
	fmt.Print(s0)
	s1 := fmt.Sprintf(" gonum.org LinearRegression: halflife is %.0f minutes, Y-intercept is %.0f counts. \n", halfLifeStats, exp(interceptStats))
	writestr(s1)
	check(bufioErr)

	interceptUnWtStats, slopeUnWtStats := stat.LinearRegression(xvector, yvector, unwtvector, false)
	halfLifeUnWtStats := -ln2 / slopeUnWtStats
	s = fmt.Sprintf(" gonum.org unweighted LinearRegression halflife is %.0f minutes, Y-intercpt is %.0f counts. \n", halfLifeUnWtStats, exp(interceptUnWtStats))
	// fmt.Print(s)  same as original unweighted slope and intercept.  So I won't display it but I will write it to the file.
	writestr(s)
	writerune()
	check(bufioErr)
	fmt.Println()
	fmt.Println()

	// Separating output from peak, so it's easier to read.
	// ask me about lag time for this patient.

	proposedPeak := FindLocalCountsPeak(rows)
	fmt.Print(" Enter point number to use as peak.  Default is [", proposedPeak, "]  ")

	// Will try a new way to scan and process input
	peakPt := 0
	n, err := fmt.Scanf("%d\n", &peakPt) // fmt.Scan functions read from os.Stdin
	if n < 1 || err != nil {
		peakPt = proposedPeak
	}
	fmt.Print(" Will use point [", peakPt, "] as peak point.")
	fmt.Println()

	peakrows := rows[peakPt:] // peakrows covers the specified point to the end, usually n = 9
	PeakNonZero := peakPt > 0
	fmt.Println()

	stdPeakSlope, stdPeakIntercept, stdPeakR2 := StdLR(peakrows)
	stdPeakHalfLife := -ln2 / stdPeakSlope
	unweightedPeakResults := fitfull(peakrows, false)
	unweightedPeakHalfLife := -ln2 / unweightedPeakResults.Slope
	WeightedPeakResults := fit(peakrows)
	WeightedPeakHalfLife := -ln2 / WeightedPeakResults.Slope
	WeightedPeakResults2 := fitfull(peakrows, true)
	WeightedPeakHalfLife2 := -ln2 / WeightedPeakResults2.Slope
	WeightedPeakResults3 := fitexy(peakrows)
	WeightedPeakHalfLife3 := -ln2 / WeightedPeakResults3.Slope
	IteratedPeakResults := DoOldWeightedLR(peakrows, stdPeakSlope, stdPeakIntercept)
	IteratedPeakHalfLife := -ln2 / IteratedPeakResults.Slope

	peakXVector := xvector[peakPt:]
	peakYVector := yvector[peakPt:]
	peakWtVector := wtvector[peakPt:]
	peakUnwtVector := unwtvector[peakPt:]
	interceptPeakStats, slopePeakStats := stat.LinearRegression(peakXVector, peakYVector, peakWtVector, false)
	halfLifePeakStats := -ln2 / slopePeakStats

	interceptPeakUnWtStats, slopePeakUnWtStats := stat.LinearRegression(peakXVector, peakYVector, peakUnwtVector, false)
	halfLifePeakUnWtStats := -ln2 / slopePeakUnWtStats

	fmt.Println()
	fmt.Println()

	if PeakNonZero {
		writerune()
		writerune()
		s = fmt.Sprintf(" Peak point is at %.3g minutes.\n", peakrows[0].x)
		fmt.Print(s)
		writestr(s)
		ctfmt.Printf(ct.Yellow, true, " Original unweighted peak halflife is %.0f minutes, and Y-intercept is %.0f counts.\n",
			stdPeakHalfLife, exp(stdPeakIntercept))
		s := fmt.Sprintf(" Original std peak unweighted: T-1/2 of Gastric Emptying is %.0f minutes, slope is %.6f cnts/min, Y-intercept is %.0f cnts and R-squared is %.6f.",
			stdPeakHalfLife, stdPeakSlope, exp(stdPeakIntercept), stdPeakR2)
		writestr(s)
		writerune()
		check(bufioErr)
		s = fmt.Sprintf(" fitful peak unweighted: halflife of Gastric Emptying is %.0f minutes.  Slope= %.6f cnts/min, Y-intercept= %.0f, StDevSlope= %.6f cnts.",
			unweightedPeakHalfLife, unweightedPeakResults.Slope, exp(unweightedPeakResults.Intercept), unweightedPeakResults.StDevSlope)
		writestr(s)
		writerune()
		s = fmt.Sprintf(" fit peak weighted: halflife of Gastric Emptying is %.0f minutes.  Slope= %.6f, Y-intercept= %.0f, StDevSlope= %.6f, GoodnessOfFit= %.6f.",
			WeightedPeakHalfLife, WeightedPeakResults.Slope, exp(WeightedPeakResults.Intercept), WeightedPeakResults.StDevSlope, WeightedPeakResults.GoodnessOfFit)
		writestr(s)
		writerune()
		ctfmt.Printf(ct.Yellow, false, " fitful peak weighted: halflife is %.0f minutes, Y-intercept is %.0f counts.\n", WeightedPeakHalfLife2, exp(WeightedPeakResults2.Intercept))
		s = fmt.Sprintf(" fitful weighted: halflife of Gastric Emptying is %.0f minutes, slope is %.6f, Y-intercept = %.0f, fit is %.6f.",
			WeightedPeakHalfLife2, WeightedPeakResults2.Slope, exp(WeightedPeakResults2.Intercept), WeightedPeakResults2.GoodnessOfFit)
		writestr(s)
		writerune()
		ctfmt.Printf(ct.Yellow, true, " fitexy peak weighted: halflife is %.0f minutes, exp(intercept) is %.0f counts.\n ", WeightedPeakHalfLife3, exp(WeightedPeakResults3.Intercept))
		s = fmt.Sprintf(" fitexy peak weighted: halflife of Gastric Emptying is %.0f minutes, slope is %.6f, Y-intercept = %.0f counts, stdev is %.6f, chi2= %.6f, q= %.6f.",
			WeightedPeakHalfLife3, WeightedPeakResults3.Slope, exp(WeightedPeakResults3.Intercept), WeightedPeakResults3.StDevSlope, WeightedPeakResults3.chi2, WeightedPeakResults3.q)
		writestr(s)
		writerune()
		ctfmt.Printf(ct.Yellow, false, " Old iterative method: from peak halflife is %.0f minutes, exp(intercept) is %.0f.\n", IteratedPeakHalfLife, exp(IteratedPeakResults.Intercept))
		s = fmt.Sprintf(" Old iterative method: from peak halflife of Gastric Emptying is %.2f minutes, slope is %.6f, exp(intercept) = %.2f, R^2 = %.6f.",
			IteratedPeakHalfLife, IteratedPeakResults.Slope, exp(IteratedPeakResults.Intercept), IteratedPeakResults.R2)
		writestr(s)
		writerune()
		s1 := fmt.Sprintf(" gonum.org LinearRegression: peak halflife is %.0f minutes, exp(intercept) is %.0f. \n", halfLifePeakStats, exp(interceptPeakStats))
		ctfmt.Print(ct.Green, true, s1)
		s2 := fmt.Sprintf(" gonum.org LinearRegression: peak halflife is %.0f minutes, Y-intercept is %.0f. \n", halfLifePeakStats, exp(interceptPeakStats))
		writestr(s2)
		writerune()

		s = fmt.Sprintf(" gonum.org UnWeighted LR peak halflife is %.0f minutes, Y-intrcpt is %.0f. \n",
			halfLifePeakUnWtStats, exp(interceptPeakUnWtStats))
		// fmt.Print(s)  same as original unweighted slope and intercept.  So I won't display it but I will write it to the file.
		writestr(s)
		writerune()

		check(bufioErr)
	}

	// The files will flush and close themselves because of the defer statements.
	fmt.Println()
	fmt.Println()
} // end main

// SQR is the squared function
func SQR(R float64) float64 {
	return R * R
} // END SQR;

// StdLR is the standard linear regression routine.  It returns the slope, intercept and correlation coefficient.
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
	}
	SlopeNumerator := N*sumxy - sumx*sumy
	SlopeDenominator := N*sumx2 - SQR(sumx)
	Slope := SlopeNumerator / SlopeDenominator
	Intercept := (sumy - Slope*sumx) / N
	R2 := SQR(SlopeNumerator) / SlopeDenominator / (N*sumy2 - SQR(sumy))
	return Slope, Intercept, R2
} //  END StdLR;

// check is an error check.  It panics if there's an error.
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// fit does assume that X is independent and only Y is dependent.  Does a weighted fit.
func fit(rows []Point) FittedData {
	/*
		subroutine fit(x,y,ndata,sig,mwt,a,b,siga,sigb,chi2,q) is the Fortran signature.
		   Based on Numerical Recipies code of same name on p 508-9 in Fortran, 1st ed,
		   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
		   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering. (C) 1986, Cambridge University Press.
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

// fitfull returns slope and intercept data in the FitedData return param.
func fitfull(row []Point, weighted bool) FittedData {

	/*
	   Based on Numerical Recipies code of same name on p 508-9 in Fortran,
	   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
	   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering. (C) 1986, Cambridge University Press.
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

// gammq is an incomplete Gamma function.
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

// ------------------------------------------------------------------------------------
// -------------------------------------------- Version 2 -----------------------------
// ------------------------------------------------------------------------------------
//
// Section 15.3 -- Straight line data with errors in both coordinates.  P 660 ff.

// subroutine fitexy(x, y, sigx, sigy, a, b, siga, sigb, chi2, q float64, ndat int) is fortran signature.  Calculates errors in X and Y.
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

	bmx := BIG // find standard errors for b as points where
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
	// Returns the value of Chi squared - offsets, for the slope b = tan(bang).
	// scaled data and offsets are communicated via the common block fitxyc in the fortran code.
	var Chixy, avex, avey, sumw, b float64
	// COMMON /fitxyc/ xx,yy,sx,sy,ww,aa,offs,nn    from the fortran code.

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
	// Given a function, chixy, and given a bracketing triplet of abscissas ax, bx, cx such that
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

// fit2 is extended version of fit, now including errors in X as well as Y.
func fit2(rows []Point) FittedData2 {
	/*
		subroutine fit(x,y,ndata,sig,mwt,a,b,siga,sigb,chi2,q) is the Fortran signature.
		   Based on Numerical Recipies code of same name on p 508-9 in Fortran, 1st ed,
		   and p 771 in Pascal.  "Numerical Recipies: The Art of Scientific Computing",
		   William H.  Press, Brian P.  Flannery, Saul A.  Teukolsky, William T.  Vettering. (C) 1986, Cambridge University Press.
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

// ============================================================================================
//               Version 3
// ============================================================================================
// I don't remember what was wrong with the old iterative solution.  I'm revisiting it
// and will document what I see wrong, so I don't forget again.
// Turns out that I don't see anything wrong, and the code is working.  Sometimes the iterated
// result is smaller and sometimes larger than the original unweighted estimate.  That seems to
// be as it should be.
// Basic algorithm is that the variances are used to calculate a weight for each point, and then
// those weights are used to calculate a new slope and intercept.  Then new weights are computed
// using the new slope and intercept, and a new slope and intercept is computed by the new weights.
// The iteration stops at either 100 iterations or if the change is below the tolerance factor,
// currently 1.e-5 * slope.

/*
   Do Old Linear Regression Routine.
     This routine does the actual linear regression using a weighted algorithm
   that is described in Zanter, Jean-Paul, "Comparison of Linear Regression
   Algortims," JPAM Jan/Feb '86, 5:1, p 14-22.  This algorithm is used instead
   of the std one because the std one assumes that the errors on the independent
   variable are negligible and the errors on the dependent variable are
   constant.  This is definitely not the case in a clinical situation where,
   for example, both the time of a blood sample and the counts per minute in
   that blood sample, are subject to variances.



                                         1
     Weight of a point = ---------------------------------------
                         (error_on_y)**2 + (slope*error_on_x)**2

     ax -- a quantity defined by the author, used in the iterative solution of
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

// DoOldWeightedLR returns the results in the FittedData3 struct.
func DoOldWeightedLR(rows []Point, slope, intercept float64) FittedData3 {
	PrevSlope := slope
	PrevIntrcpt := intercept
	var result FittedData3

	for ITERCTR := 0; ITERCTR < ITERMAX; ITERCTR++ {
		rows = GetWeights(rows, slope, intercept)
		result = WeightedLR(rows)
		slope = result.Slope
		intercept = result.Intercept
		if (math.Abs(slope-PrevSlope) < ToleranceFactor*math.Abs(slope)) &&
			(math.Abs(intercept-PrevIntrcpt) < ToleranceFactor*math.Abs(intercept)) {
			break
		} else {
			PrevSlope = slope
			PrevIntrcpt = intercept
		} //END(*IF*);
	} // ENDFOR
	/*
	   DENOM := result.SumWt * result.SumWtX2 - SQR(result.SumWtX);
	   StDevS := sqrt(ABS(result.SumWt/DENOM));
	   StDevI := sqrt(ABS(result.SumWtX2/DENOM));
	   fmt.Println(" St Dev on Slope is ",StDevS,", St Dev on Intercept is ", StDevI)
	   fmt.Println()
	*/
	return result
} // END DoOldWeightedLR

func GetWeights(rows []Point, slope, intercept float64) []Point {
	/*
	  ************************ GetWeights ************************************
	  Get Weights.
	  This routine computes the weights and the AX quantities as given by the above formulas.
	*/

	for c, p := range rows {
		ExpectedX := math.Abs((p.y - intercept) / slope)
		ExpectedY := math.Abs(slope*p.x + intercept)
		ErrorX := math.Abs(p.x-ExpectedX) / math.Sqrt(ExpectedX)
		MinError := ToleranceFactor * ExpectedX
		if ErrorX < MinError {
			ErrorX = MinError
		}

		ErrorY2 := SQR(p.y-ExpectedY) / ExpectedY
		MinError = ToleranceFactor * ExpectedY
		if ErrorY2 < MinError {
			ErrorY2 = MinError
		}
		rows[c].weight = 1 / (ErrorY2 + SQR(slope*ErrorX))
		rows[c].ax = p.x - rows[c].weight*(slope*p.x+intercept-p.y)*slope*SQR(ErrorX)
	} //  ENDFOR
	return rows
} //END GetWeights

func WeightedLR(rows []Point) FittedData3 {
	/*
	  ******************************** WeightedLR *********************************
	  Weighted Sum Up.
	  This procedure sums the variables using the weights and AX quantaties as
	  described (and computed) above.

	*/

	var result3 FittedData3

	for _, p := range rows {
		result3.SumWt += p.weight
		result3.SumWtX += p.weight * p.x
		result3.SumWtX2 += p.weight * SQR(p.x)
		result3.SumWtXY += p.weight * p.x * p.y
		result3.SumWtY += p.weight * p.y
		result3.SumWtY2 += p.weight * SQR(p.y)
		result3.SumWtAx += p.weight * p.ax
		result3.SumWtAxX += p.weight * p.ax * p.x
		result3.SumWtAxY += p.weight * p.ax * p.y
	} //  ENDFOR

	result3.Slope = (result3.SumWtAx*result3.SumWtY - result3.SumWtAxY*result3.SumWt) /
		(result3.SumWtX*result3.SumWtAx - result3.SumWtAxX*result3.SumWt)
	result3.Intercept = (result3.SumWtY - result3.Slope*result3.SumWtX) / result3.SumWt
	result3.R2 = SQR(result3.SumWt*result3.SumWtXY-result3.SumWtX*result3.SumWtY) /
		(result3.SumWt*result3.SumWtX2 - SQR(result3.SumWtX)) / (result3.SumWt*result3.SumWtY2 -
		SQR(result3.SumWtY))
	return result3
} //  END WeightedLR

// ------------------------------------ exp -----------------------------------
func exp(f float64) float64 {
	return math.Exp(f)
}

// FindLocalCountsPeak  does what its name says.  I don't remember why I need this.
func FindLocalCountsPeak(rows []Point) int {
	var maxcounts float64
	var pointindex int

	for i, point := range rows {
		if point.OrigY > maxcounts {
			maxcounts = point.OrigY
			pointindex = i
		}
	}
	return pointindex
}

// readLine is needed because a bytes.Reader does not have a readLine method, so I have to write one.
func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byt, err := r.ReadByte() // byte is a reserved word for a variable type.
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
		if byt == '\n' { // will stop scanning a line after seeing these characters like in bash or C-ish.
			return strings.TrimSpace(sb.String()), nil
		}
		if byt == '\r' {
			continue
		}
		if byt == '#' || byt == '/' { // a single / is enough to mark a comment, but I still use // in the data files.
			discardRestOfLine(r)
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byt)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine

// ----------------------------------------------------------------------

func discardRestOfLine(r *bytes.Reader) { // To allow comments on a line, I have to discard rest of line from the bytes.Reader
	for { // keep swallowing characters until EOL or an error.
		rn, _, err := r.ReadRune()
		if err != nil {
			return
		}
		if rn == '\n' {
			return
		}
	}
}
