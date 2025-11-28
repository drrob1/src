package main // rai.go
import (
	"fmt"
	"math"
	"runtime"
	"src/timlibg"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

/*
  REVISION HISTORY
  -------- -------
  26 Nov 25 -- First started writing this.  It is to calculate how much RAI to order when we have to decay the pill on site.
				It needs the desired date and dose, and the ordering date which is the calibration date.  I will confirm that the dates are within a month of each other,
				and make sure that the calibration date is before the treating date.
    I did some of my own algrebra, and I came up w/ two different calculations to get the same answer.  I'll use both here to compare.

                       A                      (lambda * time)
         A   =   -----------------    =  A * e                   where lambda = ln(2) / halflife.  For I-131, halflife = 8.0249 days.
          0      (-lambda * time)
                e

   Notice that the sign of the exponent in the factor is different btwn the 2 expressions

  28 Nov 25 -- Decided to use my own TimLibg routines as they're more flexible.  The time package cannot handle either a 4-digit year or a 2-digit year.  My routines can.
				And if no year is entered, then the current year is assumed.
				Nevermind, I decided to keep using the time package but allow the user to enter a 4-digit year, and if it's not 4 digits, then I'll use the current year.
*/

const lastModified = "28 Nov 25"
const shortLayout = "01/02/06"
const longLayout = "01/02/2006"

var lambda = math.Log(2) / 8.0249

func main() {
	fmt.Printf("  rai is last modified %s, compiled with %s \n", lastModified, runtime.Version())

	var verboseFlag bool
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output for debugging")
	flag.Parse()

	var mm, dd, yy int

	fmt.Printf(" Enter date for patient to be treated as mm dd yy : ")
	n, err := fmt.Scanf("%d %d %d\n", &mm, &dd, &yy)
	if err != nil {
		if n == 3 { // then year is not entered, so use current year.
			fmt.Printf(" Error scanning for treatingDateStr %s\n", err)
			fmt.Printf(" Exiting\n")
			return
		}
	}
	if yy == 0 {
		yy = time.Now().Year()
	}
	treatingDateStr := fmt.Sprintf("%02d/%02d/%02d", mm, dd, yy)
	treatingDate, err := time.Parse(shortLayout, treatingDateStr)
	if err != nil {
		//  fmt.Printf(" Error from parsing treatingDateStr %s, using layout %s is %s\n", treatingDateStr, shortLayout, err)
		treatingDate, err = time.Parse(longLayout, treatingDateStr)
		if err != nil {
			fmt.Printf(" Error from parsing treatingDateStr %s, using layout %s is %s\n", treatingDateStr, longLayout, err)
			fmt.Printf(" Exiting\n")
			return
		}
	}
	treatingJulian := timlibg.JULIAN(mm, dd, yy)

	fmt.Printf(" Enter date of calibration as mm dd yy : ")
	n, err = fmt.Scanln(&mm, &dd, &yy) // doing this differently to make sure it works.  It does.
	if err != nil {
		if n == 3 { // then year is not entered, so use current year.
			fmt.Printf(" Error from scanning for calibrationDateStr %s\n", err)
			fmt.Printf(" Exiting\n")
			return
		}
	}

	if yy == 0 {
		yy = time.Now().Year()
	}
	calibrationDateStr := fmt.Sprintf("%02d/%02d/%02d", mm, dd, yy)
	calibrationDate, err := time.Parse(shortLayout, calibrationDateStr)
	if err != nil {
		calibrationDate, err = time.Parse(longLayout, calibrationDateStr)
		if err != nil {
			fmt.Printf(" Error from parsing calibrationDateStr %s, using layout %s is %s\n", calibrationDateStr, longLayout, err)
			fmt.Printf(" Exiting\n")
			return
		}
	}

	if calibrationDate.After(treatingDate) {
		fmt.Printf(" Calibration date is after the treating date.  Should I switch the dates? ")
		var switchDatesAns string
		fmt.Scanln(&switchDatesAns)
		switchDatesAns = strings.ToLower(switchDatesAns)
		if strings.Contains(switchDatesAns, "n") { // if not, then exit.
			return
		}
		calibrationDate, treatingDate = treatingDate, calibrationDate
	}

	deltaT := treatingDate.Sub(calibrationDate).Hours() / 24.0               // calculating it this way makes it a float and not a duration.
	deltaT2 := float64(treatingDate.Sub(calibrationDate) / 24.0 / time.Hour) // need to explicitly convert to a float from a duration.

	calibrationJulian := timlibg.JULIAN(mm, dd, yy)
	ΔT := treatingJulian - calibrationJulian

	if verboseFlag {
		fmt.Printf(" Treating Date = %s, Calibration Date = %s\n", treatingDate.Format(shortLayout), calibrationDate.Format(shortLayout))
		fmt.Printf(" deltaT1 = %g, deltaT2 = %g, ΔT = %d, treatingJulian = %d, calibrationJulian = %d \n",
			deltaT, deltaT2, ΔT, treatingJulian, calibrationJulian)
	}

	fmt.Printf(" Enter desired dose (mCi) : ")
	var dose float64
	fmt.Scanf("%g\n", &dose)
	if dose <= 0 {
		fmt.Printf(" Error: dose must be positive.  Exiting\n")
		return
	}
	if dose > 33 {
		fmt.Printf(" Error: dose must be less than 33.  Exiting\n")
		return
	}

	raiToOrder1 := dose * math.Exp(lambda*deltaT)
	raiToOrder2 := dose / math.Exp(-lambda*deltaT)

	fmt.Printf(" RAI to order = %.2f\n", raiToOrder1) // I don't need to show both values, now that I've confirmed that they are equal.

	if verboseFlag {
		fmt.Printf(" RAI to order 1 = %.4g, RAI to order 2 = %.4g\n", raiToOrder1, raiToOrder2)
	}

	fmt.Printf(" Done\n\n")
}
