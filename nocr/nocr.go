// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// nocr.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const lastCompiled = "9 Oct 21"

func main() {
/*
REVISION HISTORY
----------------
17 Apr 17 -- Started writing nocr, based on rpn.go
18 Apr 17 -- It worked yesterday.  Now I'll rename files as in Modula-2.
27 Apr 21 -- Adding flags and checking err only once, as per Rob Pike's recommendation.
28 Apr 21 -- Added showing eols for both in and out files when verbose flag is set.
 2 May 21 -- If entered string ends in a dot, make sure outputfile does not have double dot.  This came up if file has no ext.
 5 Oct 21 -- Added timing reporting
 9 Oct 21 -- Timing includes reporting search for line endings on the files.
10 Oct 21 -- Using qpid.txt, timing of loop is ~50 ms and full timing ~115 ms, on leox.
*/

	var inoutline string
	//	var err error

	fmt.Println(" nocr removes all <CR> from a file.  Last compiled", lastCompiled, "by", runtime.Version())
	fmt.Println()

	var verboseFlag, norenameFlag, noRenameFlag bool
	flag.BoolVar(&verboseFlag, "v", false, "verbose switch.")
	flag.BoolVar(&norenameFlag, "n", false, "norename output files switch.")
	flag.BoolVar(&noRenameFlag, "no", false, "noremane output files switch, another form.")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println(" Usage: nocr <filename> ")
		os.Exit(1)
	}

	renameflag := !(norenameFlag || noRenameFlag) // convenience variable

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	if verboseFlag {
		fmt.Println(ExecFI.Name(), " was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
		fmt.Println(" Full name of executable file is", execname)
		fmt.Println()
	}

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

	t0 := time.Now()
	for InBufioScanner.Scan() {
		inoutline = InBufioScanner.Text() // does not include the trailing EOL char or chars
		OutBufioWriter.WriteString(inoutline)
		OutBufioWriter.WriteRune('\n')
		//	_, err := OutBufioWriter.WriteString(inoutline)
		//	check(err)
		//	_, err = OutBufioWriter.WriteRune('\n')
		//	check(err)
	}
	elapsedTime := time.Since(t0)

	_ = InputFile.Close()
	if err := OutBufioWriter.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, err, "Exiting.")
		os.Exit(1)
	}
	_ = OutputFile.Close()


	filename1 := InFilename
	filename2 := OutFilename
	if renameflag {
		TempFilename := InFilename + OutFilename + ".tmp"
		_ = os.Rename(InFilename, TempFilename)
		_ = os.Rename(OutFilename, InFilename)
		_ = os.Rename(TempFilename, OutFilename)
		filename1, filename2 = filename2, filename1
	}

	FI, err := os.Stat(filename1)
	if err != nil {
		_,_ = fmt.Fprintln(os.Stderr, err)
	}
	FileSize1 := FI.Size()

	FI, err = os.Stat(filename2)
	if err != nil {
		_,_ = fmt.Fprintln(os.Stderr, err)
	}
	FileSize2 := FI.Size()

	fmt.Println(" Input file is", filename1, " and size is", FileSize1)
	fmt.Println(" Output File is", filename2, " and size is", FileSize2)
	fmt.Println()

	if verboseFlag {
		fmt.Printf(" Elapsed time is %s \n", elapsedTime.String())
		CRtot, LFtot := eols(filename1)
		fmt.Println("File", filename1, "has", CRtot, "CR and", LFtot, "LF.")
		CRtot, LFtot = eols(filename2)
		fmt.Println("File", filename2, "has", CRtot, "CR and", LFtot, "LF.")
		fmt.Println(" Time incl'g this section is", time.Since(t0).String())
		fmt.Println()
	}

} // main in nocr.go

func eols(fn string) (int, int) {
	byteSlice, err := os.ReadFile(fn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reading", fn, err)
	}
	CRtot := strings.Count(string(byteSlice), "\r")
	LFtot := strings.Count(string(byteSlice), "\n")
	return CRtot, LFtot
}