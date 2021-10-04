//   (C) 1990-2017.  Robert W Solomon.  All rights reserved.
// toascii.go, based on utf8toascii, based on nocr.go
//   Note that this routine will preserve the line endings.  utf8toascii can change them.

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
27 Apr 21 -- Added v flag, for verbose.
 3 May 21 -- Now handles case where inputfile does not have an extension, indicated by a terminating dot.
 3 Oct 21 -- Added elapsedTime output in verbose mode.  On a large text file, this routine took ~500 ms to process.  On leox.
 4 Oct 21 -- Now called toascii1, and read file into memory first.  And will time file I/O also.
               Here, the I/O was ~100 ms.  I don't know why the I/O is ~2x what it is for the other approaches.
*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"
	//
)

const lastAltered = "4 Oct 2021"

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
	fmt.Println(" toascii converts utf8 to ascii, without changing line endings.  Last altered ", lastAltered)
	fmt.Println()
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	var norenameflag = flag.Bool("no", false, "norenameflag -- do not rename files at end.")
	var NoRenameFlag bool
	flag.BoolVar(&NoRenameFlag, "N", false, "NoRenameFlag -- do not rename files at end.")
	var helpflag = flag.Bool("h", false, "Print help message")
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "H", false, "Print help message")
	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "v", false, "verbose messages")

	flag.Parse()

	if verboseFlag {
		fmt.Println(ExecFI.Name(), " was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
		fmt.Println(" Full name of executable file is", execname, "compiled using", runtime.Version())
	}

	if flag.NArg() == 0 {
		fmt.Println(" Usage: utf8toascii <filename> ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *helpflag || HelpFlag {
		flag.PrintDefaults()
	}

	RenameFlag := !(*norenameflag || NoRenameFlag) // same as ~A && ~B, symbollically.  This reads better in the code below.
	commandline := flag.Arg(0)
	BaseFilename := filepath.Clean(commandline)
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

	t1 := time.Now()
	InputFile, err := os.ReadFile(InFilename)
	if err != nil {
		fmt.Println(" Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	InBuf := bytes.NewReader(InputFile)

	OutFilename := BaseFilename + OutFileSuffix
	buf := make([]byte, 0, len(InputFile))
	OutBuf := bytes.NewBuffer(buf)

	//InBufioReader := bufio.NewReader(InputFile)
	//OutBufioWriter := bufio.NewWriter(OutputFile)
	//defer OutBufioWriter.Flush()

	t0 := time.Now()

	for {
		r, _, err := InBuf.ReadRune()
		if err == io.EOF || err != nil {
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
		} else if r == unicode.ReplacementChar {
			str = ""
		} else if r == highsquote91 || r == highsquote92 {
			str = "'"
		} else if r == highquote93 || r == highquote94 {
			str = "\""
		} else if r == emdash97 {
			str = " -- "
		} else if r == bullet95 || r == bullet96 {
			str = "--"
		} else {
			str = string(r)
		}

		_, _ = OutBuf.WriteString(str)
		//		_, err = OutBufioWriter.WriteString(str)
		//		check(err)
	}

	elapsedTime := time.Since(t0)

	err = os.WriteFile(OutFilename, OutBuf.Bytes(), 0666)
	elapsedT1 := time.Since(t1)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while writing output file is", err, ".  Exiting.")
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf(" Elapsed inner time is %s, elapsed outer time t1 is %s \n", elapsedTime, elapsedT1)
	}


	// Make the processed file the same name as the input file.  IE, swap in and out files, unless the norename flag was used on the command line.

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
