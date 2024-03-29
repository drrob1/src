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
 2 Oct 21 -- Now called toascii2, based on toascii.  It will use strings.ReplaceAll instead of reading one rune at a time.  Just to see how this goes.
 3 Oct 21 -- Added elapsedTime output w/ verbose mode.  On same large file that took ~500 ms to process (ref pt), this took ~50 ms.  On leox.
 4 Oct 21 -- Added timing to include file I/O, to be fairer to "toascii", and added runtime.Version().  File renaming excluded from the timing.
               Turned out that the file I/O took ~ 50 ms.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"
	//
)

const lastAltered = "4 Oct 2021"

const openQuoteRune rune = 8220
const closeQuoteRune rune = 8221
const squoteRune rune = 8217
const opensquoteRune rune = 8216
const emdashRune rune = 8212
const endashRune rune = 8211
const bulletpointRune rune = 8226
const threedotsRune rune = 8230
const hyphenRune rune = 8208
const diagraphFIrune rune = 64257
const diagraphFLrune rune = 64258

const quoteString = "\""
const squoteString = "'"
const emdashStr = " -- "
const bulletpointStr = "--"
const threedotsStr = " ... "
const hyphenStr = "-"
const diagraphFIstr = "fi"
const diagraphFLstr = "fl"

// From the vim lines that change high ASCII characters to printable equivalents.
const highsquote91 rune = 0x91 // open squote
const highsquote92 rune = 0x92 // close squote
const highquote93 rune = 0x93  // open quote
const highquote94 rune = 0x94  // close quote
const emdash97 rune = 0x97     // emdash as ASCII character
const bullet96 rune = 0x96
const bullet95 rune = 0x95

func main() {
	fmt.Println()
	fmt.Println(" toascii2 converts utf8 to ascii, without changing line endings.  Last altered ", lastAltered, ", compiled by", runtime.Version())
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
	InputString := string(InputFile)

	t0 := time.Now()
	FileString := strings.ReplaceAll(InputString, string(unicode.ReplacementChar), "")
	FileString = strings.ReplaceAll(FileString, string(highsquote91), squoteString)
	FileString = strings.ReplaceAll(FileString, string(highsquote92), squoteString)
	FileString = strings.ReplaceAll(FileString, string(highquote93), quoteString)
	FileString = strings.ReplaceAll(FileString, string(highquote94), quoteString)
	FileString = strings.ReplaceAll(FileString, string(openQuoteRune), quoteString)
	FileString = strings.ReplaceAll(FileString, string(closeQuoteRune), quoteString)
	FileString = strings.ReplaceAll(FileString, string(squoteRune), squoteString)
	FileString = strings.ReplaceAll(FileString, string(opensquoteRune), squoteString)
	FileString = strings.ReplaceAll(FileString, string(emdashRune), emdashStr)
	FileString = strings.ReplaceAll(FileString, string(endashRune), emdashStr)
	FileString = strings.ReplaceAll(FileString, string(emdash97), emdashStr)
	FileString = strings.ReplaceAll(FileString, string(bulletpointRune), bulletpointStr)
	FileString = strings.ReplaceAll(FileString, string(bullet95), bulletpointStr)
	FileString = strings.ReplaceAll(FileString, string(bullet96), bulletpointStr)
	FileString = strings.ReplaceAll(FileString, string(threedotsRune), threedotsStr)
	FileString = strings.ReplaceAll(FileString, string(hyphenRune), hyphenStr)
	FileString = strings.ReplaceAll(FileString, string(diagraphFIrune), diagraphFIstr)
	FileString = strings.ReplaceAll(FileString, string(diagraphFLrune), diagraphFLstr)

	elapsedTime := time.Since(t0)

	OutputByteSlice := []byte(FileString)
	OutFilename := BaseFilename + OutFileSuffix
	err = os.WriteFile(OutFilename, OutputByteSlice, 0666)
	elapsedT1 := time.Since(t1)
	lengthMsg := fmt.Sprintf("Len of InputFile is %d, len of InputString is %d, len of FileString is %d, len of OutputByteSlice is %d. \n",
		len(InputFile), len(InputString), len(FileString), len(OutputByteSlice))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error while writing output file is", err)
		_, _ = fmt.Fprintln(os.Stderr, lengthMsg, "Exiting.")
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Println(lengthMsg)
		fmt.Printf(" Elapsed inner time is %s, and elapsed t1, outer time is %s. \n", elapsedTime, elapsedT1)
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

	InFI, _ := os.Stat(inputfilename)
	OutFI, _ := os.Stat(outputfilename)
	fmt.Println(" UTF_8 File is ", inputfilename, " and size is ", InFI.Size())
	fmt.Println(" ASCII File is ", outputfilename, " and size is ", OutFI.Size())
	fmt.Println()

} // main in toascii.go

func check(e error) {
	if e != nil {
		panic(e)
	}
}
