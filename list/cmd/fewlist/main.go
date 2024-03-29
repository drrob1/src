package main // fewlist

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"src/few"
	"time"

	"os"
	"path/filepath"
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
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.  And how am I going to handle collisions?
  22 Dec 22 -- I'm going to add a display like dsrt, using color to show sizes.  And I'll display the timestamp.  This means that I changed NewList to return []FileInfoExType.
                 So I'm propagating that change thru.
  25 Dec 22 -- Moving the file selection stuff to list.go.
  26 Dec 22 -- Shortened the messages.  And added a timer.
  29 Dec 22 -- Added check for an empty filelist.  And list package code was enhanced to include a sentinel of '.'
   1 Jan 23 -- Now uses list.New instead of list.NewList
   5 Jan 23 -- Adding stats to the output.
   6 Jan 23 -- Now that it clears the screen each time thru the selection loop, I'll print the version message at the end also.
                 Added a stop code of zero.
   7 Jan 23 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag.
  22 Jan 23 -- Added Sync call.
  23 Jan 23 -- Added changing destination file(s) timestamp to match the respective source file(s).  And fixed date comparison for replacement copies.
  25 Jan 23 -- Adding verify.
  28 Jan 23 -- Adding verify success message.
  30 Jan 23 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).  Added timeFudgeFactor
  31 Jan 23 -- timeFudgeFactor is now a Duration.
  28 Feb 23 -- Now called fewlist, based on copylist.  I'm going to use a list to run few 32 on each of them.  I'm not going to make that a param, yet.
  26 Mar 23 -- Completed the usage info.  And added list.CheckDest.
  31 Mar 23 -- StaticCheck found a few issues.
   5 Apr 23 -- Refactored list.ProcessDirectoryAliases
   8 Apr 23 -- Changed list.New signature.
*/

const LastAltered = "8 Apr 2023" //

const sepString = string(filepath.Separator)

var err error
var verifyFlag bool

func main() {
	execName, _ := os.Executable()
	execFI, _ := os.Stat(execName)
	execTimeStamp := execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	fmt.Printf("%s is compiled w/ %s, last altered %s, timestamp on binary is %s\n", os.Args[0], runtime.Version(), LastAltered, execTimeStamp)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s, timestamp on binary is %s. \n", os.Args[0], LastAltered, runtime.Version(), execTimeStamp)
		fmt.Fprintf(flag.CommandLine.Output(), " %s Usage information: src-dir-or-glob dest-dir; only using IEEE32 algorithm.\n", os.Args[0])
		//fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		//fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")
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

	flag.BoolVar(&verifyFlag, "verify", false, "Verify copy operation")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	//Reverse := revFlag

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

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.FilterFlag = filterFlag
	list.ReverseFlag = revFlag
	list.ExcludeRex = excludeRegex
	list.SizeFlag = sizeFlag

	//fileList, err := list.New(excludeRegex, sizeFlag, Reverse) // fileList used to be []string, but now it's []FileInfoExType.
	fileList, err := list.New() // fileList used to be []string, but now it's []FileInfoExType.
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
		fmt.Printf(" Length of the filelist is zero.  Aborting\n")
		os.Exit(1)
	}

	// now have the initial fileList.  Need to check the destination directory.

	destDir := list.CheckDest()
	if destDir == "" {
		fmt.Print(" Destination directory ? ")
		_, err = fmt.Scanln(&destDir)
		if err != nil {
			destDir = "." + sepString
		}
		if strings.ContainsRune(destDir, ':') {
			//directoryAliasesMap := list.GetDirectoryAliases()
			destDir = list.ProcessDirectoryAliases(destDir)
		} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			destDir = strings.Replace(destDir, "~", homeDirStr, 1)
		}
		if !strings.HasSuffix(destDir, sepString) {
			destDir = destDir + sepString
		}
	}
	//else {
	//	if strings.ContainsRune(destDir, ':') {
	//		//directoryAliasesMap := list.GetDirectoryAliases()
	//		destDir = list.ProcessDirectoryAliases(destDir)
	//	} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
	//		homeDirStr, _ := os.UserHomeDir()
	//		destDir = strings.Replace(destDir, "~", homeDirStr, 1)
	//	}
	//	if !strings.HasSuffix(destDir, sepString) {
	//		destDir = destDir + sepString
	//	}
	//}
	fmt.Printf("\n destDir = %#v\n", destDir)
	fi, err := os.Lstat(destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %s is supposed to be the destination directory, but os.Lstat(%s) = %#v.  Exiting\n", destDir, destDir, err)
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, " %s is supposed to be the distination directory, but os.Lstat(%s) not c/w a directory.  Exiting\n", destDir, destDir)
		os.Exit(1)
	}

	fileList, err = list.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}
	if verboseFlag {
		for i, f := range fileList {
			fmt.Printf(" second fileList[%d] = %s and also %s\n", i, f.RelPath, f.AbsPath)
		}
		fmt.Println()
		fmt.Printf(" There are %d files in the file list.\n", len(fileList))
	}
	fmt.Printf("\n\n")

	// time to check the files
	start := time.Now()

	var success, fail int
	onWin := runtime.GOOS == "windows"
	for _, f := range fileList {
		destF, err := os.Open(destDir)
		if err != nil {
			ctfmt.Printf(ct.Red, onWin, " os.Open(%s) error is: %s.  Skipping\n", destDir, err)
			continue
		}

		destFI, err := destF.Stat()
		if err != nil {
			ctfmt.Printf(ct.Red, onWin, " destF.Stat(%s) error is: %s.  Skipping\n", destDir, err)
			destF.Close()
			continue
		}
		if !destFI.IsDir() {
			s := fmt.Sprintf("os.Stat(%s) must show a directory, but it's not c/w a directory.  Skipping\n", destDir)
			ctfmt.Printf(ct.Red, onWin, "%s\n", s)
			destF.Close()
			continue
		}
		destF.Close()

		targetName := filepath.Join(destDir, f.FI.Name())
		if verboseFlag {
			fmt.Printf("Just before Feq32withNames.  f.Dir=%s, f.RelPath=%s, f.AbsPath=%s, f.FullPath=%s, targetname=%s\n", f.Dir, f.RelPath, f.AbsPath, f.FullPath, targetName)
		}
		result, err := few.Feq32withNames(f.AbsPath, targetName)
		if err != nil {
			s := fmt.Sprintf(" ERROR from Feq32withNames(%s, %s) is: %s", f.AbsPath, targetName, err)
			ctfmt.Printf(ct.Red, onWin, "%s\n", s)
			fail++
		}
		if result {
			s := fmt.Sprintf(" IEEE 32 matched for %s and in %s", f.AbsPath, destDir)
			ctfmt.Printf(ct.Green, onWin, " %s\n", s)
			success++
		} else {
			s := fmt.Sprintf(" IEEE 32 failed for %s and in %s", f.AbsPath, destDir)
			ctfmt.Printf(ct.Red, onWin, " %s\n", s)
			fail++
		}
	}
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), LastAltered)
	fmt.Printf("\n Successfully IEEE 32 matched %d files, and FAILED to match %d files; elapsed time is %s\n\n", success, fail, time.Since(start))
} // end main
