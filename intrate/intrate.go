package main // intrate
import (
	"fmt"
	"math"
	"os"
)

/*
  7 Sep 2024 -- First writing of intrate based on the compound amount formulas I first read in 1975 from the HP-25 calculator book.  These are:
        n = number of periods.  i = periodic interest rate.  PV, FV are present and future values, I = total accrued interest, in dollars, like total finance charges.
        Today, I want to know i given PV, FV and n.
  9 Sep 2024 -- Confirming that 1/n = -n.  Actually, I confirmed that they are not the same.  I forgot that a negative exponent causes the reciprocal of the entire
        expression, not just the exponent.

                         1/n
    ln(FV/PV)        (FV)                         -n               n                     n
n = ---------    i = (--)    - 1      PV = FV(1+i)     FV = PV(1+i)         I = PV [(1+i)  - 1]
    ln(1 + i)        (PV)

Any of the unknowns can be calculated as long as the other quantities are known.
*/

const lastModified = "Sep 9, 2024"

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
	fmt.Printf(" PV = %.0f, FV = %.0f, n = %.0f, and annual interest rate is %.2f %%\n\n", PV, FV, n, i*12*100)

	//i2 := math.Pow(valueRatio, -n) - 1
	//iRound := math.Round(i*100*100) / 100 / 100
	//i2Round := math.Round(i2*100*100) / 100 / 100
	//if iRound == i2Round {
	//	ctfmt.Printf(ct.Green, true, " -n and 1/n compute to the same value, which is %f\n\n", iRound*1200)
	//} else {
	//	ctfmt.Printf(ct.Red, true, " 1/n and -n DO NOT compute to the same value, which are %.4f and %.4f, and iround = %f, i2round = %f\n\n",
	//		i*100*12, i2*100*12, iRound*1200, i2Round*1200)
	//}
}
