// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// nocr2.go

package main

import (
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
	 5 Oct 21 -- Now called nocr2.go, and I want to use the strings.replacer ability.
	 8 Oct 21 -- made verboseMode default by adding noVerboseFlag
	10 Oct 21 -- Timing for inner time on qpid.txt is ~20 ms, and outer time is ~96 ms on leox.
*/

	fmt.Println(" nocr2 removes all <CR> from a file using a strings.replacer.  Last compiled", lastCompiled, "by", runtime.Version())
	fmt.Println()

	var verboseFlag, noVerboseFlag, norenameFlag, noRenameFlag bool
	//flag.BoolVar(&verboseFlag, "v", false, "verbose switch")  desire is to make verbose default.  Need to use -V to turn it off.  But why would you.
	flag.BoolVar(&noVerboseFlag, "V", false, "no verbose switch default is true.")
	flag.BoolVar(&norenameFlag, "n", false, "norename output files switch.")
	flag.BoolVar(&noRenameFlag, "no", false, "noremane output files switch, another form.")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println(" Usage: nocr2 <filename> ")
		os.Exit(1)
	}

	renameflag := !(norenameFlag || noRenameFlag) // convenience variable
	verboseFlag = !noVerboseFlag

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

	outerT := time.Now() // timer outer
	InputFileData, err := os.ReadFile(InFilename)
	if err != nil {
		fmt.Println(" Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	inputString := string(InputFileData)

	innerT := time.Now() // timer inner
	replaced := strings.NewReplacer("\r", "")
	outputString := replaced.Replace(inputString)
	elapsedInner := time.Since(innerT)

	OutFilename := BaseFilename + OutFileSuffix
	outputByteSlice := []byte(outputString)
	err = os.WriteFile(OutFilename, outputByteSlice, 0666)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}


	filename1 := InFilename
	filename2 := OutFilename
	if renameflag {
		TempFilename := InFilename + OutFilename + ".tmp"
		os.Rename(InFilename, TempFilename)
		os.Rename(OutFilename, InFilename)
		os.Rename(TempFilename, OutFilename)
		filename1, filename2 = filename2, filename1
	}

	FI, err := os.Stat(filename1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	FileSize1 := FI.Size()

	FI, err = os.Stat(filename2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	FileSize2 := FI.Size()

	fmt.Println(" Input file is", filename1, " and size is", FileSize1)

	// inputString := string(InputFileData)
	// outputString := replaced.Replace(inputString)
	CRtot := strings.Count(inputString, "\r")
	LFtot := strings.Count(inputString, "\n")
	if verboseFlag {
		fmt.Println(" It has", CRtot, "CR and", LFtot, "LF.")
	}

	fmt.Println(" Output File is", filename2, " and size is", FileSize2)
	CRtot = strings.Count(outputString, "\r")
	LFtot = strings.Count(outputString, "\n")
	if verboseFlag {
		fmt.Println(" It has", CRtot, "CR and", LFtot, "LF.")
	}
	fmt.Println()
	elapsedOuter := time.Since(outerT)

	if verboseFlag {
		fmt.Printf(" Inner timer is %s, outer timer is %s.\n\n", elapsedInner.String(), elapsedOuter.String())
	}
} // main in nocr2.go

/*
func eols(fn string) (int, int) {
	byteSlice, err := os.ReadFile(fn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Reading", fn, err)
	}
	CRtot := strings.Count(string(byteSlice), "\r")
	LFtot := strings.Count(string(byteSlice), "\n")
	return CRtot, LFtot
}
*/
