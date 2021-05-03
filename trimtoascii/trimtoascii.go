//   (C) 1990-2021.  Robert W Solomon.  All rights reserved.
// trimtoascii.go, based on toascii, based on utf8toascii, based on nocr.go
//   Note that this routine will preserve the line endings.  utf8toascii can change them.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

/*
REVISION HISTORY
----------------
17 Apr 17 -- Started writing nocr, based on rpn.go
18 Apr 17 -- It worked yesterday.  Now I'll rename files as in Modula-2.
 5 May 17 -- Now will convert utf8 to ascii, based on nocr.go
 6 May 17 -- After I wrote ShowUtf8, I added more runes here and added OS based line endings.
 8 May 17 -- Added the -n or -no switch meaning no renaming at end of substitutions.
13 May 17 -- Changed the text of the final output message.
15 May 17 -- Will now call this toascii.go.  io.Copy and encoding/ascii85.NewEncoder does not work.
10 Sep 17 -- Added code to show timestamp of execname.  And changed bufio error checking.
23 Dec 17 -- Added code to do what I also do in vim with the :%s/\%x91/ /g lines.
12 Apr 21 -- Used toascii.go as a base, and am now writing this as trimtoascii.go, and will use bytes.reader and bytes.buffer.
 3 May 21 -- Now handles case where input file does not have an extension, indicated by a terminating dot.
*/

const lastAltered = "3 May 2021"

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
const fileMode = 0777

// From the vim lines that change high ASCII characters to printable equivalents.
const highsquote91 = 0x91 // open squote
const highsquote92 = 0x92 // close squote
const highquote93 = 0x93  // open quote
const highquote94 = 0x94  // close quote
const emdash97 = 0x97     // emdash as ASCII character
const bullet96 = 0x96
const bullet95 = 0x95

func main() {
	var str string

	fmt.Println()
	fmt.Println(" runetoascii converts utf8 to ascii, removing unrecognized runes.  Last altered", lastAltered, "compiled with", runtime.Version())
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	var norenameflag = flag.Bool("no", false, "norenameflag -- do not rename files at end.")
	var noRenameFlag bool
	flag.BoolVar(&noRenameFlag, "N", false, "NoRenameFlag -- do not rename files at end.")
	var helpflag = flag.Bool("h", false, "Print help message")
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "H", false, "Print help message")
	var verboseFlag = flag.Bool("v", false, "verbose -- enable to be, well, verbose.")

	flag.Parse()

	if *helpflag || HelpFlag {
		fmt.Println(" Usage: trimtoascii <filename> ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *verboseFlag {
		fmt.Println(ExecFI.Name(), " was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
		fmt.Println(" Full name of executable file is", execname)
		fmt.Println()
	}

	RenameFlag := !(*norenameflag || noRenameFlag) // same as ~A && ~B, symbollically.  This reads better in the code below.

	if flag.NArg() == 0 {
		fmt.Println(" Usage: trimtoascii <filename> ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFilename := flag.Arg(0)

	BaseFilename := filepath.Clean(inputFilename)
	InFilename := ""
	InFileExists := false
	Ext1Default := ".txt"
	OutFileSuffix := ".out"

	if strings.Contains(BaseFilename, ".") {
		if BaseFilename[len(BaseFilename)-1] == '.' { // remove last char if it's a dot.
			BaseFilename = BaseFilename[:len(BaseFilename)-1]
		}
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

	inputFileContents, err := os.ReadFile(InFilename)
	if err != nil {
		fmt.Println(err, " Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	inputBuf := bytes.NewReader(inputFileContents)

	OutFilename := BaseFilename + OutFileSuffix
	outputSlice := make([]byte, 0, len(inputFileContents))
	outputFileBuf := bytes.NewBuffer(outputSlice)

	if *verboseFlag {
		fmt.Println(" Read filename is", InFilename, "and write filename is", OutFilename)
		fmt.Println()
	}

	for {
		r, siz, err := inputBuf.ReadRune()
		if err != nil {
			break
		}
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
		} else if r == highsquote91 || r == highsquote92 {
			str = "'"
		} else if r == highquote93 || r == highquote94 {
			str = "\""
		} else if r == emdash97 {
			str = " -- "
		} else if r == bullet95 || r == bullet96 {
			str = "--"
		} else if r > 127 || siz > 1 || r == unicode.ReplacementChar { // ReplacementChar represents invalid code points.
			continue // skip the WriteString step for this rune.
		} else {
			str = string(r)
		}

		_, err = outputFileBuf.WriteString(str)
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "from inputBuf.ReadRune() call.  Ignored.")
		}
	}

	// based on Rob Pike's posting.  Only need to check the error here.
	if err := os.WriteFile(OutFilename, outputFileBuf.Bytes(), fileMode); err != nil {
		fmt.Fprintln(os.Stderr, err, " Output file error from os.WriteFile.  Exiting.")
		os.Exit(1)
	}

	// Make the processed file the same name as the input file.  IE, swap in and out files,
	// unless the norename flag was used on the command line.

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

	fmt.Println(" UTF_8 File is ", inputfilename, " and size is ", InputFileSize)
	fmt.Println(" ASCII File is ", outputfilename, " and size is ", OutputFileSize)
	fmt.Println()

} // main in toascii.go

func check(e error) {
	if e != nil {
		panic(e)
	}
}
