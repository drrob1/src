package main // dellist

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
	"regexp"
	"runtime"
	"src/list"
	"strings"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 22 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                 This is going to take a while.
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                 And how am I going to handle collisions?
  22 Dec 22 -- I'm going to add a display like dsrt, using color to show sizes.  And I'll display the timestamp.  This means that I changed NewList to return []FileInfoExType.
                 So I'm propagating that change thru.
  25 Dec 22 -- Moving the file selection stuff to list.go
                 Now called dellist.go
  29 Dec 22 -- Adding check for an empty list, and the list package code was enhanced to include '.' as a sentinel.
   1 Jan 23 -- Now uses list.New instead of list.NewList.
   6 Jan 23 -- list package functions now return an error.  This allows better error handling and a stop code.
   7 Jan 23 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag
  24 Jan 23 -- And added list.ReverseFlag and list.SizeFlag.
  23 Mar 23 -- Now based on list2, so I can use a regexp on the input files.
   4 Apr 23 -- Now back to list, as I think I've sorted out my issues on the bash command line.  So compiling this will replace the older version based on list2 in GoBin.
   5 Apr 23 -- Updated the usage message.
*/

const LastAltered = "5 Apr 2023" //

const defaultHeight = 40
const minWidth = 90

var autoWidth, autoHeight int
var err error

//var rexStr, inputStr string
//var rex *regexp.Regexp

func main() {
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), LastAltered)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information: %s [glob pattern]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
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

	//if rexStr != "" {
	//	rex, err = regexp.Compile(rexStr)
	//	if err != nil {
	//		fmt.Printf(" Input regular expression error is %s.  Ignoring\n", err)
	//	}
	//}
	//list.InputDir = inputStr
	//list.IncludeRex = rex
	list.FilterFlag = filterFlag
	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.ReverseFlag = revFlag
	list.SizeFlag = sizeFlag
	list.ExcludeRex = excludeRegex
	list.DelListFlag = true

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", ExecFI.Name(), ExecTimeStamp, execName)
		fmt.Println()
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

	fileList, err := list.New(excludeRegex, sizeFlag, revFlag) // fileList used to be []string, but now it's []FileInfoExType.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.New is %s\n", err)
		os.Exit(1)
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

	for i, f := range fileList {
		fmt.Printf(" to be deleted fileList[%d] = %s\n", i, f.RelPath)
	}
	fmt.Println()
	fmt.Printf(" There are %d files in the file list.\n", len(fileList))
	fmt.Print(" Continue (y/N)? ")
	var ans string
	n, err := fmt.Scanln(&ans)
	if n == 0 || err != nil {
		fmt.Printf("\n n = %d, err = %s.  Aborting.\n", n, err)
		os.Exit(1)
	}
	ans = strings.ToLower(ans)
	if !strings.HasPrefix(ans, "y") { // ans doesn't begin w/ y, so abort
		fmt.Printf("\n ans = %s, which does not begin with y.  Aborting.", ans)
		os.Exit(1)
	}

	// time to delete the files

	onWin := runtime.GOOS == "windows"
	for _, f := range fileList {
		err = os.Remove(f.RelPath)
		ctfmt.Printf(ct.Green, onWin, " Deleting %s\n", f.RelPath)
		if err != nil {
			//fmt.Fprintf(os.Stderr, " ERROR while copying %s -> %s is %#v.  Skipping to next file.\n", f.RelPath, destDir, err)
			ctfmt.Printf(ct.Red, onWin, " ERROR: %s\n", err)
			continue
		}
	}
} // end main
