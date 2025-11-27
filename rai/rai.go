package main // rai.go
import (
	"fmt"
	"math"
	"runtime"
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

*/

const lastModified = "26 Nov 25"
const layout = "01/02/06"

var lambda = math.Log(2) / 8.0249

func main() {
	fmt.Printf("  rai is last modified %s, compiled with %s \n", lastModified, runtime.Version())

	var verboseFlag bool
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output for debugging")
	flag.Parse()

	var mm, dd, yy int

	fmt.Printf(" Enter date for patient to be treated as mm dd yy : ")
	_, err := fmt.Scanf("%d %d %d\n", &mm, &dd, &yy)
	if err != nil {
		fmt.Printf(" Error scanning for treatingDateStr %s\n", err)
		fmt.Printf(" Exiting\n")
		return
	}
	treatingDateStr := fmt.Sprintf("%02d/%02d/%02d", mm, dd, yy)
	treatingDate, err := time.Parse(layout, treatingDateStr)
	if err != nil {
		fmt.Printf(" Error from parsing treatingDateStr %s, using layout %s is %s\n", treatingDateStr, layout, err)
		fmt.Printf(" Exiting\n")
		return
	}

	fmt.Printf(" Enter date of calibration as mm dd yy : ")
	_, err = fmt.Scanf("%d %d %d\n", &mm, &dd, &yy)
	if err != nil {
		fmt.Printf(" Error from scanning for calibrationDateStr %s\n", err)
		fmt.Printf(" Exiting\n")
		return
	}

	calibrationDateStr := fmt.Sprintf("%02d/%02d/%02d", mm, dd, yy)
	calibrationDate, err := time.Parse(layout, calibrationDateStr)
	if err != nil {
		fmt.Printf(" Error from parsing calibrationDateStr %s is %s\n", calibrationDateStr, err)
		fmt.Printf(" Exiting\n")
		return
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
	deltaT2 := float64(treatingDate.Sub(calibrationDate) / 24.0 / time.Hour) // calculating it this way changes it from a duration to a float.

	if verboseFlag {
		fmt.Printf(" Treating Date = %s, Calibration Date = %s\n", treatingDate.Format(layout), calibrationDate.Format(layout))
		fmt.Printf(" deltaT1 = %g, deltaT2 = %g \n", deltaT, deltaT2)
	}

	fmt.Printf(" Enter desired dose : ")
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

	if verboseFlag {
		fmt.Printf(" RAI to order 1 = %.2f, RAI to order 2 = %.2f\n", raiToOrder1, raiToOrder2)
	}

	fmt.Printf(" RAI to order 1 = %.2f, RAI to order 2 = %.2f\n", raiToOrder1, raiToOrder2)

	fmt.Printf(" Done\n\n")
}
