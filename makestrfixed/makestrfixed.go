package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
26 July 21 -- I have a thought that I want to test.  Context is the multicolumn dsrt routines.  I want a routine that maxes a string of fixed length
                that I can specify as a param, and it either pads the input string or truncates it.


 */

func makeStrFixed(s string, size int) string {
	var built strings.Builder

	if len(s) > size { // need to truncate the string
		return s[:size]
	} else if len(s) == size {
		return s
	} else if len(s) < size { // need to pad the string
		needSpaces := size - len(s)
		built.Grow(size)
		built.WriteString(s)
		spaces := strings.Repeat(" ", needSpaces)
		built.WriteString(spaces)
		return built.String()
	} else {
		fmt.Fprintln(os.Stderr, " makeStrFixed input string length is strange.  It is", len(s))
		return s
	}
} // end makeStrFixed


func main() {

  source := "abcdefghijklmnopqrstuvwxyz"

  for {
  	ans := ""
  	fmt.Print(" Enter trim length: ")
  	n, err := fmt.Scanln(&ans)
  	if err != nil {
  		fmt.Println(" error from Scanln is", err, "exiting.")
  		break
	}
	if n == 0 {
		break
	}
	i, err := strconv.Atoi(ans)
	if err != nil {
		fmt.Println(" Error from ans conversion Atoi is", err, "exiting.")
		break
	}
	trimmedstr := makeStrFixed(source, i)
	fmt.Printf(" trimmed string is %q, length is %d.\n", trimmedstr, len(trimmedstr))
  }




}
