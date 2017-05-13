// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// utf-8 to ascii, based on nocr.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
	//
)

const lastCompiled = "13 May 17"

//const openQuoteRune = 0xe2809c  \  These values are in the file itself seen by hexdump -C
//const closeQuoteRune = 0xe2809d  \ but are not the rune (unicode code point)
//const squoteRune = 0xe28099      / representing these characters.
//const emdashRune = 0xe28094     /  I didn't know these could be different.
const openQuoteRune = 8220
const closeQuoteRune = 8221
const squoteRune = 8217
const opensquoteRune = 8216
const emdashRune = 8212
const endashRune = 8211
const bulletpointRune = 8226
const threedotsRune = 8230
const hyphenRune = 8208
const diagraphFIrune = 64257
const diagraphFLrune = 64258

const quoteString = "\""
const squoteString = "'"
const emdashStr = " -- "
const bulletpointStr = "--"
const threedotsStr = " ... "
const hyphenStr = "-"
const diagraphFIstr = "fi"
const diagraphFLstr = "fl"

/*
   REVISION HISTORY
   ----------------
   17 Apr 17 -- Started writing nocr, based on rpn.go
   18 Apr 17 -- It worked yesterday.  Now I'll rename files as in Modula-2.
    5 May 17 -- Now will convert utf8 to ascii, based on nocr.go
    6 May 17 -- After I wrote ShowUtf8, I added more runes here and
                  added OS based line endings.
    8 May 17 -- Added the -n or -no switch meaning no renaming at end of substitutions.
   13 May 17 -- Changed the text of the final output message.
*/

func main() {
	var instr, outstr, str, lineEndings string

	fmt.Println(" utf8toascii converts utf8 to ascii.  Last compiled ", lastCompiled)
	fmt.Println()

	var norenameflag = flag.Bool("no", false, "norenameflag -- do not rename files at end.")
	var NoRenameFlag bool
	flag.BoolVar(&NoRenameFlag, "N", false, "NoRenameFlag -- do not rename files at end.")
	var helpflag = flag.Bool("h", false, "Print help message")
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "H", false, "Print help message")

	flag.Parse()

	commandline := flag.Arg(0)
	if len(os.Args) <= 1 || len(commandline) == 0 {
		fmt.Println(" Usage: utf8toascii <filename> ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *helpflag || HelpFlag {
		flag.PrintDefaults()
	}

	RenameFlag := !(*norenameflag || NoRenameFlag) // same as ~A && ~B, symbollically.  So this reads better in the code below.

	BaseFilename := filepath.Clean(commandline)
	InFilename := ""
	InFileExists := false
	Ext1Default := ".txt"
	OutFileSuffix := ".out"

	if runtime.GOOS == "linux" {
		lineEndings = "\n"
	} else if runtime.GOOS == "windows" {
		lineEndings = "\r\n"
	}

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

	OutFilename := BaseFilename + OutFileSuffix
	OutputFile, err := os.Create(OutFilename)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()

	InBufioScanner := bufio.NewScanner(InputFile)
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()

	for InBufioScanner.Scan() {
		instr = InBufioScanner.Text() // does not include the trailing EOL char
		runecount := utf8.RuneCountInString(instr)
		//		fmt.Println(" Len of instr is ", len(instr), ", runecount is ", runecount)
		if len(instr) == runecount {
			outstr = instr
		} else { // a mismatch btwn instr length and rune count means that a multibyte rune is in this instr
			stringslice := make([]string, 0, runecount)
			for dnctr := runecount; dnctr > 0; dnctr-- {
				r, siz := utf8.DecodeRuneInString(instr) // front rune in r
				instr = instr[siz:]                      // chop off the first rune
				//				fmt.Print(" r, siz: ", r, siz, ".  ")
				if r == openQuoteRune || r == closeQuoteRune {
					str = quoteString
				} else if r == squoteRune || r == opensquoteRune {
					str = squoteString
				} else if r == emdashRune || r == endashRune {
					str = emdashStr
				} else if r == bulletpointRune {
					str = bulletpointStr
				} else if r == threedotsRune {
					str = threedotsStr
				} else if r == hyphenRune {
					str = hyphenStr
				} else if r == diagraphFIrune {
					str = diagraphFIstr
				} else if r == diagraphFLrune {
					str = diagraphFLstr
				} else {
					str = string(r)
				}
				stringslice = append(stringslice, str)

			}
			outstr = strings.Join(stringslice, "")

		}
		_, err := OutBufioWriter.WriteString(outstr)
		check(err)
		_, err = OutBufioWriter.WriteString(lineEndings)
		check(err)
	}

	InputFile.Close()
	OutBufioWriter.Flush() // code did not work without this line.
	OutputFile.Close()

	// Make the processed file the same name as the input file.  IE, swap in and
	// out files, unless the norename flag was used on the command line.

	inputfilename := InFilename
	outputfilename := OutFilename

	if RenameFlag {
		TempFilename := InFilename + OutFilename + ".tmp"
		os.Rename(InFilename, TempFilename)
		os.Rename(OutFilename, InFilename)
		os.Rename(TempFilename, OutFilename)
		inputfilename = OutFilename
		outputfilename = InFilename
	}

	FI, err := os.Stat(inputfilename)
	InputFileSize := FI.Size()

	FI, err = os.Stat(outputfilename)
	OutputFileSize := FI.Size()

	fmt.Println(" UTF-8 File is ", inputfilename, " and size is ", InputFileSize)
	fmt.Println(" ASCII File is ", outputfilename, " and size is ", OutputFileSize)
	fmt.Println()

} // main in utf8toascii.go

func check(e error) {
	if e != nil {
		panic(e)
	}
}
