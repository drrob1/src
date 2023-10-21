// usqrt.go
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"src/getcommandline"
	"strconv"
)

/*
  REVISION HISTORY
  ================
  26 Feb 18 -- First version
  14 Jun 18 -- Added Newton's method as estimate factor.
  21 Oct 23 -- Updated code to compile w/ modules.
*/

const LastAlteredDate = "21 Oct 23"

func main() {

	var INBUF string

	fmt.Println(" usqrt Program.  Last altered ", LastAlteredDate)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
	} else {
		fmt.Print(" Enter number to factor : ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(INBUF) == 0 {
			os.Exit(0)
		}
	} // if command tail exists

	//	N, err := strconv.Atoi(INBUF)
	U, err := strconv.ParseUint(INBUF, 10, 64)
	if err != nil {
		fmt.Println(" Conversion to number failed.  Exiting")
		os.Exit(1)
	}
	sqrt := usqrt(uint(U))
	fmt.Println(" sqrt of ", U, " is ", sqrt)
	x := float64(U)
	fmt.Println()
	fmt.Println(" using divfloat=", divfloat(x))
	fmt.Println()
	fmt.Println(" using newtonfloat=", newtonfloat(x))
	fmt.Println()
	fmt.Println(" using newtonint=", newtonint(int(U)))
}

// ----------------------------------------------- usqrt ---------------------------
func usqrt(u uint) uint {

	sqrt := u / 2

	for i := 0; i < 20; i++ {
		guess := u / sqrt
		sqrt = (guess + sqrt) / 2
		fmt.Println(" i=", i, ", guess=", guess, ", sqrt=", sqrt)
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}
	return sqrt
}

// ---------------------------------------------- divfloat ------------------------
func divfloat(x float64) float64 {
	sqrt := x / 2
	for i := 0; i < 30; i++ {
		guess := x / sqrt
		sqrt = (guess + sqrt) / 2
		fmt.Println(" i=", i, ", guess=", guess, ", sqrt=", sqrt)
		if math.Abs(sqrt-guess) <= 1.e-4*sqrt { // recall that this is not floating math.
			break
		}
	}
	return sqrt
}

// -------------------------------------------- Newton float ---------------------
func newtonfloat(x float64) float64 {
	z := 1.0
	for i := 0; i < 30; i++ {
		z0 := z
		z -= (z*z - x) / (2 * z)
		if math.Abs(z0-z) < 1.e-4*z {
			break
		}
		fmt.Println(" i=", i, ", z=", z)
	}
	return z
}

// ------------------------------------------- Newton int ----------------------
func newtonint(x int) int {
	z := 1
	for i := 0; i < 30; i++ {
		z0 := z
		z -= (z*z - x) / (2 * z)
		fmt.Println(" i=", i, ", z=", z)
		if iAbs(z0-z) <= 1 {
			break
		}
	}
	return z - 1
}

// ------------------------------------------- iabs ---------------------------
func iAbs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}
