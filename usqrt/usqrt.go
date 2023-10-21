// usqrt.go
package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"src/getcommandline"
	"strconv"
)

/*
  REVISION HISTORY
  ================
  26 Feb 18 -- First version
  14 Jun 18 -- Added Newton's method as estimate factor.
  21 Oct 23 -- Updated code to compile w/ modules.  And added timestamp display code.
*/

const LastAlteredDate = "21 Oct 23"

func main() {

	var INBUF string

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	fmt.Printf(" usqrt Program.  Last altered %s, compiled w/ %s, binary %s linked %s\n", LastAlteredDate, runtime.Version(), execName, LastLinkedTimeStamp)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
	} else {
		fmt.Print(" Enter number: ")
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

	fmt.Println(" using usqrt [divint], sqrt of ", U, " is ", sqrt)
	x := float64(U)
	fmt.Println()
	fmt.Printf(" using divfloat=%.6f, using math.Sqrt=%.6f\n", divfloat(x), math.Sqrt(x))
	fmt.Println()
	fmt.Printf(" using newtonfloat=%.6f\n", newtonfloat(x))
	fmt.Println()
	fmt.Printf(" using newtonint=%d\n", newtonint(int(U)))
}

// ----------------------------------------------- usqrt ---------------------------
func usqrt(u uint) uint {

	sqrt := u / 2

	for i := 0; i < 20; i++ {
		guess := u / sqrt
		sqrt = (guess + sqrt) / 2
		//fmt.Print("   i=", i, ", guess=", guess, ", sqrt=", sqrt)
		fmt.Print("   i=", i, ", sqrt=", sqrt)
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}
	fmt.Println()
	fmt.Println()
	return sqrt
}

// ---------------------------------------------- divfloat ------------------------
func divfloat(x float64) float64 {
	sqrt := x / 2
	for i := 0; i < 30; i++ {
		guess := x / sqrt
		sqrt = (guess + sqrt) / 2
		fmt.Printf("   i=%d, sqrt=%.4f", i, sqrt)
		if math.Abs(sqrt-guess) <= 1.e-6*sqrt {
			break
		}
	}
	fmt.Println()
	fmt.Println()
	return sqrt
}

// -------------------------------------------- Newton float ---------------------
func newtonfloat(x float64) float64 {
	z := 1.0
	for i := 0; i < 30; i++ {
		z0 := z
		z -= (z*z - x) / (2 * z)
		if math.Abs(z0-z) < 1.e-6*z {
			break
		}
		fmt.Printf("   i=%d, z0=%.4f, z=%.4f", i, z0, z)
	}
	fmt.Println()
	fmt.Println()
	return z
}

// ------------------------------------------- Newton int ----------------------
func newtonint(x int) int {
	z := 1
	for i := 0; i < 30; i++ {
		z0 := z
		z -= (z*z - x) / (2 * z)
		fmt.Print("   i=", i, ", z=", z)
		if iAbs(z0-z) <= 1 {
			break
		}
	}
	fmt.Println()
	fmt.Println()
	return z //- 1  I think I'll remove the fudge factor.
}

// ------------------------------------------- iabs ---------------------------
func iAbs(i int) int {
	if i < 0 {
		i = -i
	}
	return i
}
