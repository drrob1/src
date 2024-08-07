package main // fewc from fewlist from copylist

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"src/few"
	"sync"
	"sync/atomic"
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
------------------------------------------------------------------------------------------------------------------------------------------------------
  28 Feb 23 -- Now called fewlist, based on copylist.  I'm going to use a list to run few 32 on each of them.  I'm not going to make that a param, yet.
------------------------------------------------------------------------------------------------------------------------------------------------------
   1 Mar 23 -- Now called fewc, based on fewlist, based on copylist.  I'm going to use a worker go routine pattern here.  And I'll use Bill Kennedy's more recent examples as reference.
   2 Mar 23 -- Abbreviated the output, as I did for the copy routines.
  26 Mar 23 -- Completed the usage info.  And added list.CheckDest.
  31 Mar 23 -- StaticCheck found a few issues.
   5 Apr 23 -- Refactored list.ProcessDirectoryAliases
   8 Apr 23 -- Changed list.New signature.
  30 Jul 24 -- Added a multiplier, based on comments by Miki Tebeka, and implemented in the other routines like cf and cf2.
                 I could make a true fanout version, but I don't think I need it, as I rarely use this routine anyway.  Not like cf2, which I use very often.
*/

const LastAltered = "30 July 2024" //

const sepString = string(filepath.Separator)

var err error
var verifyFlag bool
var multiplier int

type workersType struct {
	fName1, fName2, destDir string
}

func main() {
	execName, _ := os.Executable()
	execFI, _ := os.Stat(execName)
	execTimeStamp := execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	fmt.Printf("%s is compiled w/ %s, last altered %s, timestamp on binary is %s\n", os.Args[0], runtime.Version(), LastAltered, execTimeStamp)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s, timestamp on binary is %s. \n", os.Args[0], LastAltered, runtime.Version(), execTimeStamp)
		fmt.Fprintf(flag.CommandLine.Output(), " Concurrent few list Usage information: src-dir-or-glob dest-dir; only using IEEE32 algorithm. \n")
		fmt.Fprintf(flag.CommandLine.Output(), " It compares files w/ the same name in source and destination directories; only using IEEE32 algorithm. \n")
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

	flag.IntVar(&multiplier, "mult", 10, "Multiplier for goroutines, default currently is 10.")

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
			fmt.Printf(" second fileList[%d] = %s\n", i, f.RelPath)
		}
		fmt.Println()
		fmt.Printf(" There are %d files in the file list.\n", len(fileList))
	}
	fmt.Printf("\n\n")

	// time to check the files for equivalency.
	g := runtime.NumCPU() * multiplier
	num := min(g, len(fileList))
	var success, fail int64
	onWin := runtime.GOOS == "windows"
	workCh := make(chan workersType, num)
	var wg sync.WaitGroup

	for i := 0; i < num; i++ { // start the lesser of NumCPU() or the number of files waiting to be processed.
		go func() {
			defer wg.Done() // since this line is not in the for loop, I can't use wg.Add(len(fileList))
			for w := range workCh {
				result, err := few.Feq32withNames(w.fName1, w.fName2)
				if err != nil {
					s := fmt.Sprintf(" ERROR from Feq32withNames(%s, %s) is: %s", w.fName1, w.fName2, err)
					ctfmt.Printf(ct.Red, onWin, "%s\n", s)
					atomic.AddInt64(&fail, 1)
					continue
				}
				if result {
					s := fmt.Sprintf(" IEEE32 match succeeded for %s and in %s", w.fName1, w.destDir)
					ctfmt.Printf(ct.Green, onWin, " %s\n", s)
					atomic.AddInt64(&success, 1)
				} else {
					s := fmt.Sprintf(" IEEE32 failed for %s and in %s", w.fName1, w.fName2)
					ctfmt.Printf(ct.Red, onWin, " %s\n", s)
					atomic.AddInt64(&fail, 1)
				}
			}
		}()
	}

	start := time.Now()

	wg.Add(num)
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
		work := workersType{
			fName1:  f.AbsPath,
			fName2:  targetName,
			destDir: destDir,
		}
		workCh <- work
	}
	numGoRoutines := runtime.NumGoroutine()
	close(workCh)
	wg.Wait()
	fmt.Printf("%s is compiled w/ %s, last altered %s, elapsed time is %s using %d go routines.\n", os.Args[0],
		runtime.Version(), LastAltered, time.Since(start), numGoRoutines)
	ctfmt.Printf(ct.Green, onWin, "\n IEEE32 successfully matched %d files, ", success)
	ctfmt.Printf(ct.Red, onWin, "and FAILED to match %d files; elapsed time is %s\n\n", fail, time.Since(start))
} // end main
