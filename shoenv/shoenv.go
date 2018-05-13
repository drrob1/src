// shoenv.go -- show environment testing code.

package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const LastAltered = "May 11, 2018"

type DsrtParamType struct {
	numlines                                             int
	reverseflag, sizeflag, dirlistflag, filenamelistflag bool
}

func main() {
	var dsrtparam DsrtParamType
	fmt.Println(" shoenv.go.  Last altered", LastAltered)
	fmt.Println()

	EnvironSlice := os.Environ()

	for i, e := range EnvironSlice {
		if strings.HasPrefix(e, "dsrt") {
			fmt.Printf(" Environ[%d]: %s\n", i, e)
			dsrtslice := strings.SplitAfter(e, "=")
			fmt.Println(" dsrtparam: ")
			for _, d := range dsrtslice {
				fmt.Println(d)
			}
			indiv := strings.Split(dsrtslice[1], "") // characters after dsrt=
			fmt.Println(" indiv len=:", len(indiv), ", indiv is:", indiv)
			for j, str := range indiv {
				s := str[0]
				if s == 'r' || s == 'R' {
					dsrtparam.reverseflag = true
				} else if s == 's' || s == 'S' {
					dsrtparam.sizeflag = true
				} else if s == 'd' {
					dsrtparam.dirlistflag = true
					dsrtparam.filenamelistflag = true
				} else if s == 'D' {
					dsrtparam.dirlistflag = true
				} else if unicode.IsDigit(rune(s)) {
					dsrtparam.numlines = int(s) - int('0')
					if j+1 < len(indiv) && unicode.IsDigit(rune(indiv[j+1][0])) {
						dsrtparam.numlines = 10*dsrtparam.numlines + int(indiv[j+1][0]) - int('0')
						break // if have a 2 digit number, it ends processing of the indiv string
					}
				}
			}
		}
	}
	fmt.Println()
	fmt.Println("numlines,reverseflag,sizeflag,dirlistflag,filenamelistflag")
	fmt.Println(" dsrtparam:", dsrtparam)
	fmt.Println()
	fmt.Println()
}
