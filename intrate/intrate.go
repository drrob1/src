package main // intrate
import (
	"fmt"
	"math"
	"os"
)

/*
  7 Sep 2024 -- First writing of intrate based on the compound amount formulas I first read in 1975 from the HP-25 calculator book.  These are:
                n = number of periods.  i = periodic interest rate.  PV, FV are present and future values, I = total accrued interest, in dollars.
                Today, I want to know i given PV, FV and n.

                         1/n
    ln(FV/PV)        (FV)                         -n               n                     n
n = ---------    i = (--)    - 1      PV = FV(1+i)     FV = PV(1+i)         I = PV [(1+i)  - 1]
    ln(1 + i)        (PV)

Any of the unknowns can be calculated as long as the other quantities are known.
*/

const lastModified = "Sep 7, 2024"

func main() {
	var PV, FV, n float64
	fmt.Printf(" %s, to calculate the overall effective interest rate given PV, FV and n.  Last modified %s\n", os.Args[0], lastModified)
	fmt.Print(" PV: ")
	fmt.Scanln(&PV)
	fmt.Print(" FV: ")
	fmt.Scanln(&FV)
	fmt.Print("  n: ")
	fmt.Scanln(&n)
	valueRatio := FV / PV
	inverseN := 1 / n
	i := math.Pow(valueRatio, inverseN) - 1
	fmt.Printf(" PV = %.0f, FV = %.0f, n = %.0f, and i = %.2f %%\n", PV, FV, n, i*12*100)
}
