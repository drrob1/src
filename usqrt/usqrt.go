// usqrt.go
package main

import (
	"bufio"
	"fmt"
	"getcommandline"
	"os"
	"strconv"
)

const LastAlteredDate = "25 Feb 2018"

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
}

//----------------------------------------------- usqrt ---------------------------
func usqrt(u uint) uint {

	sqrt := u / 3

	for i := 0; i < 20; i++ {
		guess := u / sqrt
		sqrt = (guess + sqrt) / 2
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
		fmt.Println(" i=", i, ", guess=", guess, ", sqrt=", sqrt)
	}
	return sqrt
}
