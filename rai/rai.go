package main // rai.go
import (
	"fmt"
	"runtime"
	"strings"
	"time"
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

func main() {
	fmt.Printf("  rai is last modified %s, compiled with %s \n", lastModified, runtime.Version())

	fmt.Printf(" Enter date for patient to be treated as mm/dd/yy : ")
	var treatingDateStr string
	fmt.Scanln(&treatingDateStr)
	treatingDate, err := time.Parse(layout, treatingDateStr)
	if err != nil {
		fmt.Printf(" Error from parsing treatingDateStr %s is %s\n", treatingDateStr, err)
		fmt.Printf(" Exiting\n")
		return
	}

	fmt.Printf(" Enter date of calibration as mm/dd/yy : ")
	var calibrationDateStr string
	fmt.Scanln(&calibrationDateStr)
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
		if strings.Contains(switchDatesAns, "y") {
			calibrationDate, treatingDate = treatingDate, calibrationDate
		}
	}

	fmt.Printf(" Enter desired dose : ")
	var dose float64
	fmt.Scanln("%g", &dose)
	if dose <= 0 {
		fmt.Printf(" Error: dose must be positive.  Exiting\n")
		return
	}
	if dose > 33 {
		fmt.Printf(" Error: dose must be less than 33.  Exiting\n")
		return
	}
}
