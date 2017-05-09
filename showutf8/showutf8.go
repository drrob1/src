// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// ShowUtf-8 codes.  Based on utf8toascii, based on nocr.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
	//
)

const lastCompiled = "9 May 17"

//const openQuoteRune = 0xe2809c
//const closeQuoteRune = 0xe2809d
//const squoteRune = 0xe28099
//const emdashRune = 0xe28094
const openQuoteRune = 8220
const closeQuoteRune = 8221
const squoteRune = 8217
const opensquoteRune = 8216
const emdashRune = 8212
const endashRune = 8211
const bulletpointRune = 8226
const threedotsRune = 8230
const hyphenRune = 8208

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
	6 May 17 -- Added a flag -a for after to see the rest of the string to give me context for a new rune.
	7 May 17 -- Added help flag, and a character position counter.
	9 May 17 -- Tweaked output message regarding the line and position counter.
*/

func main() {
	var instr string
	//	var err error

	fmt.Println(" ShowUtf8.  Last compiled ", lastCompiled)
	fmt.Println()

	var afterflag = flag.Bool("a", false, "afterflag -- show string after rune.")
	var AfterFlag bool
	flag.BoolVar(&AfterFlag, "A", false, "AfterFlag -- show string after rune.")
	var helpflag = flag.Bool("h", false, "print help message") // pointer
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "H", false, "print help message")

	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: showutf8 <filename> ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *helpflag || HelpFlag {
		flag.PrintDefaults()
	}

	After := *afterflag || AfterFlag

	commandline := flag.Arg(0)
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
			lineCtrStr := strconv.Itoa(linecounter)
			if linecounter > 100000 {
				lineCtrStr = AddCommas(lineCtrStr)
			}
			fmt.Print(" Line ", lineCtrStr, " : ")
			for dnctr := runecount; dnctr > 0; dnctr-- {
				r, siz := utf8.DecodeRuneInString(instr) // front rune in r
				instr = instr[siz:]                      // chop off the first rune
				if r > 128 {
					fmt.Print(" r ", r, ", siz ", siz, ", posn:", runecount-dnctr+1, ".")
					if r == openQuoteRune {
						fmt.Print(" rune is opening", quoteString, "; ")
					} else if r == closeQuoteRune {
						fmt.Print(" rune is closing", quoteString, "; ")
					} else if r == squoteRune || r == opensquoteRune {
						fmt.Print(" rune is ", squoteString, "; ")
					} else if r == emdashRune || r == endashRune {
						fmt.Print(" rune is ", emdashStr, "; ")
					} else if r == bulletpointRune {
						fmt.Print(" rune is bulletpoint; ")
					} else if r == threedotsRune {
						fmt.Print(" rune is ... ; ")
					} else if r == hyphenRune {
						fmt.Print(" rune is hyphen; ")
					} else {
						fmt.Print(" rune is new.")
						if After {
							fmt.Print("  Rest of input line is: ")
							fmt.Println(instr)
						}
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

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//-----------------------------------------------------------------------------------------------------------------------------
