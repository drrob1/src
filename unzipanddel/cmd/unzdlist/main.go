package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"src/list"
	"src/unzipanddel"
	"strings"
)

/*
REVISION HISTORY
-------- -------
14 Nov 23 -- Started working on the first version of this pgm.
*/

const lastModified = "14 Nov 2023"

var err error

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" unzip and delete list, last modified %s, binary is %s, timstamp of binary is %s\n", lastModified, execName, LastLinkedTimeStamp)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], lastModified, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information: %s [glob pattern]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	var revFlag bool
	flag.BoolVar(&revFlag, "r", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var sizeFlag bool
	flag.BoolVar(&sizeFlag, "s", false, "sort by size instead of by date")

	var verboseFlag, veryVerboseFlag bool

	flag.BoolVar(&verboseFlag, "v", false, "verbose mode, which is same as test mode.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

	var excludeFlag bool
	var excludeRegex *regexp.Regexp
	var excludeRegexPattern string
	flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")
	flag.StringVar(&excludeRegexPattern, "x", "", "regex to be excluded from output.") // var, not a ptr.

	var filterFlag, noFilterFlag bool
	var filterStr string
	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVar(&filterFlag, "f", false, "filter value to suppress listing individual size below 1 MB.")
	flag.BoolVar(&noFilterFlag, "F", false, "Flag to undo an environment var with f set.")
	//flag.StringVar(&inputStr, "i", "", "Input source directory which can be a symlink.")
	//flag.StringVar(&rexStr, "rex", "", "Regular expression inclusion pattern for input files")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	if len(excludeRegexPattern) > 0 {
		if verboseFlag {
			fmt.Printf(" excludeRegexPattern found and is %d runes. \n", len(excludeRegexPattern))
		}
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			excludeFlag = false
		}
		excludeFlag = true
		fmt.Printf(" excludeRegexPattern = %q, excludeRegex.String = %q\n", excludeRegexPattern, excludeRegex.String())
	}

	list.FilterFlag = filterFlag
	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.ReverseFlag = revFlag
	list.SizeFlag = sizeFlag
	list.ExcludeRex = excludeRegex
	list.DelListFlag = true

	fileList, err := list.New() // fileList used to be []string, but now it's []FileInfoExType.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.New is %s\n", err)
		fmt.Printf(" flag.NArg = %d, len(os.Args) = %d\n", flag.NArg(), len(os.Args))
		fmt.Print(" Continue? [yN] ")
		var ans string
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			fmt.Printf(" No input detected.  Exiting.\n")
			os.Exit(1)
		}
		ans = strings.ToLower(ans)
		if strings.Contains(ans, "n") {
			os.Exit(1)
		}
	}
	if verboseFlag {
		fmt.Printf(" len(fileList) = %d\n", len(fileList))
	}
	if veryVerboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %#v\n", i, f)
		}
		fmt.Println()
	}
	if len(fileList) == 0 {
		fmt.Printf(" Length of the fileList is zero.  Aborting \n")
		os.Exit(1)
	}

	fileList, err = list.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n\n")

	// now have the fileList.

	if len(fileList) == 0 {
		fmt.Printf(" The selected list of files is empty.  Exiting.\n")
		os.Exit(1)
	}
	if verboseFlag {
		fmt.Printf(" \nLength of filelist after the selection is %d.  The filelist is %+v\n", len(fileList), fileList)
	}
	for _, f := range fileList {
		filenames, er := unzipanddel.UnzipAndDel(f.FullPath)
		if er == nil {
			fmt.Printf(" \n%s successfully unzipped and deleted: %+v\n", f.FullPath, filenames)
		} else {
			fmt.Printf(" \nUnsuccessfully unzipped or deleted %s with error of %s\n", f.FullPath, er)
		}
	}
	fmt.Println()
}
