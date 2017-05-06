// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// ShowUtf-8 codes.  Based on utf8toascii, based on nocr.go

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
	//
	"getcommandline"
)

const lastCompiled = "6 May 17"

//const openQuoteRune = 0xe2809c
//const closeQuoteRune = 0xe2809d
//const squoteRune = 0xe28099
//const emdashRune = 0xe28094
const openQuoteRune = 8220
const closeQuoteRune = 8221
const squoteRune = 8217
const emdashRune = 8212
const bulletpointRune = 8226
const quoteString = "\""
const squoteString = "'"
const emdashStr = " -- "
const bulletpointStr = "--"

/*
   REVISION HISTORY
   ----------------
   17 Apr 17 -- Started writing nocr, based on rpn.go
   18 Apr 17 -- It worked yesterday.  Now I'll rename files as in Modula-2.
    5 May 17 -- Now will convert utf8 to ascii, based on nocr.go
	6 May 17 -- Need to know the utf8 codes before I can convert 'em.
*/

func main() {
	var instr string
	//	var err error

	fmt.Println(" ShowUtf8.  Last compiled ", lastCompiled)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: utf8toascii <filename> ")
		os.Exit(1)
	}

	commandline := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(commandline)
	InFilename := ""
	InFileExists := false
	Ext1Default := ".txt"
	//	OutFileSuffix := ".out"

	if strings.Contains(BaseFilename, ".") {
		InFilename = BaseFilename
		_, err := os.Stat(InFilename)
		if err == nil {
			InFileExists = true
		}
	} else {
		InFilename = BaseFilename + Ext1Default
		_, err := os.Stat(InFilename)
		if err == nil {
			InFileExists = true
		}
	}

	if !InFileExists {
		fmt.Println(" File ", BaseFilename, " or ", InFilename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	InputFile, err := os.Open(InFilename)
	if err != nil {
		fmt.Println(" Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer InputFile.Close()
	InBufioScanner := bufio.NewScanner(InputFile)
	linecounter := 0
	for InBufioScanner.Scan() {
		instr = InBufioScanner.Text() // does not include the trailing EOL char
		linecounter++
		runecount := utf8.RuneCountInString(instr)
		if len(instr) == runecount {
			continue
		} else { // a mismatch btwn instr length and rune count means that a multibyte rune is in this instr
			fmt.Print(" Line ", linecounter, " : ")
			for dnctr := runecount; dnctr > 0; dnctr-- {
				r, siz := utf8.DecodeRuneInString(instr) // front rune in r
				instr = instr[siz:]                      // chop off the first rune
				if r > 128 {
					fmt.Print(" r: ", r, ", siz: ", siz, "; ")
					if r == openQuoteRune {
						fmt.Print(" rune is opening", quoteString, "; ")
					} else if r == closeQuoteRune {
						fmt.Print(" rune is closing", quoteString, "; ")
					} else if r == squoteRune {
						fmt.Print(" rune is ", squoteString, "; ")
					} else if r == emdashRune {
						fmt.Print(" rune is ", emdashStr, "; ")
					} else if r == bulletpointRune {
						fmt.Print(" rune is bulletpoint; ")
					} else {
						fmt.Print(" rune is new ")
					}
				}
			}
			fmt.Println()
		}
	}

	//	InputFile.Close()
	fmt.Println()

} // main in ShowUtf8.go
/*
func check(e error) {
	if e != nil {
		panic(e)
	}
}
*/
