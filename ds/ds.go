// ds.go -- directory sort output in columns using a shortened output and truncation.

package main

import (
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

/*
Revision History
----------------
20 Apr 17 -- Started writing dsize rtn, based on dirlist.go
21 Apr 17 -- Now tweaking the output format.  And used flag package.  One as a pointer and one as a value, just to learn them.
22 Apr 17 -- Coded the use of the first non flag commandline param,  which is all I need.  Note that the flag must appear before the non-flag param, else the flag is ignored.
22 Apr 17 -- Now writing dsrt, to function similarly to dsort.
24 Apr 17 -- Now adding file matching, like "dir" or "ls" does.
25 Apr 17 -- Now adding sort by size as an option, like -s, and commas
26 Apr 17 -- Noticed that the match routine is case sensitive.  I don't like that.
27 Apr 17 -- commandline now allows a file spec.  I intend this for Windows.  I'll see how it goes.
19 May 17 -- Will now show the uid:gid for linux.
20 May 17 -- Turns out that (*syscall.Stat_t) only compiles on linux.  Time for platform specific code.
21 May 17 -- Cross compiling to GOARCH=386, and the uid and User routines won't work.
 2 Sep 17 -- Added timestamp detection code I first wrote for gastricgo.
18 Oct 17 -- Added filesize totals
22 Oct 17 -- Made default numlines of 40.
23 Oct 17 -- Broadened the defaults so that linux default is 40 and windows default is 50.
12 Dec 17 -- Added -d and -D flags to mean directory and nofilename output, respectively.
13 Dec 17 -- Changed how lines are counted.
10 Jan 18 -- Added correct processing of ~.
11 Jan 18 -- Switching to fmt.Scanln.
30 Jan 18 -- Will exit if use -h flag.
 8 Feb 18 -- Windows version will not pause to accept a pattern, as it's not necessary.
23 Feb 18 -- Fixing a bug when GOARCH=386 in that userptr causes a panic.
23 Apr 18 -- Linux version will properly process command line lists passed by the shell.
24 Apr 18 -- Improving comments, and removing prompt for a pattern, as it is no longer needed.
 2 May 18 -- More improving comments.
11 May 18 -- Adding use of dsrt environment variable.  Tested ideas in shoenv.go.
28 Jun 18 -- Refining my use of an environment variable.  I did not get it exactly right the first time around.
18 Jul 18 -- Fixed bug in processing of "d" and "D" in dsrt environment.  And removed askforinput completely.
21 Aug 18 -- Playing with folding.  So far, I only folded the block of commented code at the bottom of the file
11 Sep 18 -- Will total and display all filesizes in the files slice.
12 Sep 18 -- Adding a t flag to show the totals of the entire directory
13 Sep 18 -- Added GrandTotalCount.  And KB, MB, GB, TB.
16 Sep 18 -- Fixed small bug in code for default case of KB, MB, etc
20 Mar 19 -- Planning how to deal with directory aliases in take command, tcmd, tcc.  Environment variable, diraliases
19 Jun 19 -- Fixing bug that does not show symlinks on either windows or linux.
               I changed the meanings so now use <symlink> and (dir) indicators, and fixed the error on Windows
               whereby symlinks could not be displayed.
20 Jun 19 -- Changed logic so that symlinks to files are always displayed, like files.
               That required writing a new function to detect a symlink.
23 Jun 19 -- Changed to use Lstat when there are multiple filenames on the command line.  This only happens on Linux.
 2 Jul 19 -- Changed the format pattern for displaying the executable timestamp.  And Lstat error processing changed.
 3 Jul 19 -- Removing a confusing comment, and removed need for a flag variable for issymlink
 4 Jul 19 -- Removed the pattern check code on linux.  And this revealed a bug on linux if only 1 file is globbed on command line.  Now fixed.
 5 Jul 19 -- Optimized order of printing file types.  I hope.
18 Jul 19 -- When there is an error from ioutil.ReadDir, I cannot change its behavior of not reading any more.  Just do dsrt * in bash as a work around.
19 Jul 19 -- Wrote MyReadDir
22 Jul 19 -- Added a winflag check so don't scan commandline on linux looking for : or ~.
 9 Sep 19 -- From Israel: Fixing issue on linux when entering a directory param.  And added test flag.  And added sortfcn.
22 Sep 19 -- Changed the error message under linux and have only 1 item on command line.  Error condition is likely file not found.
 4 Oct 19 -- No longer need platform specific code.  So I added GetUserGroupStrLinux.  And then learned that it won't compile on Windows.
                 So as long as I want the exact same code for both platforms, I do need platform specific code.
 6 Oct 19 -- Removed -H and added -help flags
25 Aug 20 -- File sizes to be displayed in up to 3 digits and a suffix of kb, mb, gb and tb.  Unless new -l for long flag is used.
18 Sep 20 -- Added -e and -ext flags to only show files without extensions.
 7 Nov 20 -- Learned that the idiomatic way to test absence of environment variables is LookupEnv.  From the Go Standard Lib Cookbook.
20 Dec 20 -- For date sorting, I changed away from using NanoSeconds, and I'm now using the time.Before(time) and time.After(time) functions.
                 I hope these are faster.  I haven't used the sort interface in a long time.  It's still in file dated Dec-20-2020 as a demo.
                 I removed the demo code from here.
10 Jan 21 -- Adjusting alignment of decimal points
15 Jan 21 -- Adding -x flag, to exclude a regex.  When it works here, I'll add it to other pgms.
31 Jan 21 -- Adding color.
13 Feb 21 -- Switching cyan and white.
15 Feb 21 -- Switching yellow and white so yellow is mb and white is gb
27 Feb 21 -- Found an optimization when writing getdir about GrandTotals
 1 Mar 21 -- Made sure all error messages are written to Stderr.
 2 Mar 21 -- Added use of runtime.Version(), which I read about in Go Standard Library Cookbook.
 9 Mar 21 -- Added use of os.UserHomeDir, which became available as of Go 1.12.
12 Mar 21 -- Added an os.Exit call after what is essentially a file not found error.
16 Mar 21 -- Tweaked a file not found message on linux.  And changed from ToUpper -> ToLower on Windows.
17 Mar 21 -- Added exclude string flag to allow entering the exclude regex pattern on command line; convenient for recalling the command.
22 May 21 -- Adding filter option, to filter out smaller files from the display.  And v flag for verbose, which uses also uses testFlag.
------------------------------------------------------------------------------------------------------------------------------------------------------
 9 Jul 21 -- Now called ds, and I'll use limited lengths of the file name strings.  Uses environemnt variables ds and dsw, if present.
11 Jul 21 -- Decided to not show the mode bits.
23 Jul 21 -- The colors are a good way to give me the magnitude of filesize, so I don't need the displacements here.
               But I'm keeping the display of 4 significant figures, and increased defaultwidth to 70.
               I'm adding the code to determine the number of rows and columns itself.  I'll use golang.org/x/term for linux, and shelling out to tcc for Windows.
               Now that I know autoheight, I'll have n be a multiplier for the number of screens to display, each autolines - 5 in size.  N will remain as is.
28 Jul 21 -- Backporting the changes from ds2 and ds3, ie, autoheight, autowidth, and putting the output as strings in a slice struct.
22 Oct 21 -- Changed the code that uses bytes.NewBuffer()
29 Jan 22 -- Porting the code from dsrt.go to here.  I'm using more platform specific code now, and the code is much simpler.
31 Jan 22 -- Now that ds2 is working, I'm going to refactor this code.
 1 Feb 22 -- Added veryVerboseFlag, intended for only when I really need it.  And fixed environ var by making it dsrt instead of what it was, ds.
               And optimized includeThis.
 3 Feb 22 -- Finally reversed the -x and -exclude options, so now -x means I enter the exclude regex on the command line.  Whew!
 5 Feb 22 -- Now to add the numOfCols stuff that works in rex.go.  So this will also allow multi-column displays, too.
10 Feb 22 -- Fixing bug of when an error is returned to MyReadDir.
14 Feb 22 -- Fix bug of not treating an absolute path one that begins w/ the filepath.Separator character.  Actual fix is in _linux.go file.
15 Feb 22 -- Replaced testFlag w/ verboseFlag
16 Feb 22 -- Time to remove the upper case flags that I don't use.
25 Apr 22 -- Added the -1 flag and it's halfFlag variable.  For displaying half the number of lines the screen allows.
15 Oct 22 -- Added max flags to undo the effect of environment var dsrt=20
               I noticed that the environment string can't process f, for filterFlag.  Now it can.
               Now I need an option, -F, to undo the filterflag set in an environment var.
21 Oct 22 -- golangci-lint said I don't use global directoryAliasMap, so I'm removing it.
               Turned out that the linter was wrong, sort of.  I don't use it on linux, but I need it on Windows.  So I had to put it back for Windows.
11 Nov 22 -- Will output environ var settings on header.  They're easy to forget :-)
14 Jan 23 -- I wrote args to learn more about how arguments are handled.  I think I got it wrong in dsrtutil_linux.  I'm going to fix it.  Now that it works there
               I'll fix it here, too.
 7 Feb 23 -- Corrected an oversight of not closing a file in dsutil_linux.go
 9 Apr 23 -- StaticCheck reported several unused variables (numLines, grandTotalCount, sizeTotal, grandTotal) and not using a value of HomeDirStr here.
26 Apr 23 -- I noticed that this routine doesn't work when a command line pattern is given only on Windows.  Time to fix that now.  Turned out to be an errant "return fileInfos" at or near line 107 _windows
               And I added back -g for globFlag.  That got lost somehow.
 4 Jul 23 -- I'm back porting code from dsrt to here.  I added the -a flag here, changed the environ number to mean number of screens for the all option, and added environ var h to mean halfFlag.
               Then I improved ProcessEnvironString, as long as I was here.
18 Feb 24 -- Made it clear that this sorts by mod date.  And now *nscreens * numOfCols is the multiplier for num of lines.  Should have been this way all along.
22 Feb 24 -- Undid change about *nscreens.  Increasing the number of columns does not need a larger number of lines.  Oops.
 4 May 24 -- I was able to add concurrency to speed up dsrt by writing fast dsrt (fdsrt).  I'll add that code here too.
              I'm removing -g, glob switch.  I never used it, and it overly complicates the code.  It remains in dsrt, but not in fdsrt or here.
 5 Jan 25 -- There's a bug in how the dsrt environ variable is processed.  It sets the variable that's now interpretted as nscreens instead of nlines (off the top of my head)
				nscreens can only be set on the command line, not by environ var.  The environ var is used to set lines to display on screen.
 6 Jan 25 -- Today's my birthday.  But that's not important now.  If I set nlines via the environment, and then use the halfFlag, the base amount is what dsrt is, not the full screen.
				I want the base amount to be the full screen.  I have to think about this for a bit.
				I decided to use the maxflag system, and set maxflag if halfflag or if nscreens > 1 or if allflag.
*/

const LastAltered = "6 Jan 2025"

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, thru use of OS specific code.  On Windows it will get a pattern from the command line.
// but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.
// As of Feb 5, 2022, I can use this routine to handle the 2 and 3 column displays, so I don't need the separate routines.  I decided that the difficulity of maintaining
// all of these different files is too much.  From now on, all code improvements will happen in this pgm.

type dirAliasMapType map[string]string

type DsrtParamType struct {
	paramNum, w                                                                           int
	reverseflag, sizeflag, dirlistflag, filenamelistflag, totalflag, filterflag, halfFlag bool
}

type colorizedStr struct {
	color ct.Color
	str   string
}

const defaultHeight = 40
const minWidth = 90
const maxWidth = 300
const min2Width = 160
const min3Width = 170
const multiplier = 10 // used for the worker pool pattern in MyReadDir
const fetch = 1000    // used for the concurrency pattern in MyReadDir
var numWorkers = runtime.NumCPU() * multiplier

var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, verboseFlag bool
var filterFlag, globFlag, veryVerboseFlag, halfFlag, maxDimFlag, fastFlag bool
var filterAmt, numOfLines, grandTotalCount int

var sizeTotal, grandTotal int64
var filterStr string
var excludeRegex *regexp.Regexp

//var directoryAliasesMap dirAliasMapType //unused according to StaticCheck, and GoLand, too, in fact.  It was used dsutil_windows.go:43, but that's redundant so I took that out, also.

var autoWidth, autoHeight int

// allScreens is the number of screens to be used for the allFlag switch.  This can be set by the environ var dsrt.
var allScreens = 50

// this is to be equivalent to 100 screens.
var allFlag bool

func main() {
	var dsrtParam DsrtParamType
	var userPtr *user.User // from os/user
	var fileInfos []os.FileInfo
	var err error
	var GrandTotal int64
	var excludeRegexPattern string
	var numOfCols int

	uid := 0
	gid := 0
	systemStr := ""

	winFlag := runtime.GOOS == "windows"

	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		//autoDefaults = false
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	// environment variable processing.  If present, these will be the defaults.
	dsrtEnviron := os.Getenv("dsrt")
	dswEnviron := os.Getenv("dsw")
	dsrtParam = ProcessEnvironString(dsrtEnviron, dswEnviron) // This is a function below.

	ctfmt.Printf(ct.Magenta, winFlag, "ds -- Directory SoRTed w/ filename truncation.  LastAltered %s, compiled with %s", LastAltered, runtime.Version())
	if dsrtEnviron != "" {
		ctfmt.Printf(ct.Yellow, winFlag, ", dsrt env = %s", dsrtEnviron)
	}
	if dswEnviron != "" {
		ctfmt.Printf(ct.Yellow, winFlag, ", dsw env = %s", dswEnviron)
	}
	fmt.Println()

	if runtime.GOARCH == "amd64" {
		uid = os.Getuid() // int
		gid = os.Getgid() // int
		userPtr, err = user.Current()
		if err != nil {
			fmt.Println(" user.Current error is ", err, "Exiting.")
			os.Exit(1)
		}
	}

	// flag definitions and processing

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt and dsw environment variables before processing commandline switches.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " dsrt environ values are: paramNum=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelistflag=%t, totalflag=%t \n",
			dsrtParam.paramNum, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag, dsrtParam.totalflag)

		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	revflag := flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
	var RevFlag bool                                                                      // will always be false.  I'll leave it this way for now.
	//flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nscreens = flag.Int("n", 1, "number of screens to display, ie, a multiplier") // Ptr
	var NLines int
	flag.IntVar(&NLines, "N", 0, "number of lines to display") // Value

	var sizeflag = flag.Bool("s", false, "sort by size instead of by mod date") // pointer
	var SizeFlag bool                                                           // will always be false.  I'll leave it this way for now.
	//flag.BoolVar(&SizeFlag, "S", false, "sort by size instead of by date")

	var DirListFlag = flag.Bool("d", false, "include directories in the output listing") // pointer

	var FilenameListFlag bool
	flag.BoolVar(&FilenameListFlag, "D", false, "Directories only in the output listing")

	var TotalFlag = flag.Bool("t", false, "include grand total of directory")

	flag.BoolVar(&verboseFlag, "test", false, "enter a testing mode to println more variables")
	flag.BoolVar(&verboseFlag, "v", false, "verbose mode, which is same as test mode.")

	var longflag = flag.Bool("l", false, "long file size format.") // Ptr

	var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")
	flag.StringVar(&excludeRegexPattern, "x", "", "regex to be excluded from output.") // var, not a ptr.

	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVar(&filterFlag, "f", false, "filter value to suppress listing individual size below 1 MB.")
	noFilterFlag := flag.Bool("F", false, "Flag to undo an environment var with f set.")

	var w int // width maximum of the filename string to be displayed
	flag.IntVar(&w, "w", 0, "width for displayed file name")

	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

	flag.IntVar(&numOfCols, "c", 1, "Columns in the output.")
	flag.BoolVar(&halfFlag, "1", false, "display 1/2 of the screen.")

	flag.BoolVar(&globFlag, "g", false, "globbing flag, which on windows uses filepath.Glob.")

	mFlag := flag.Bool("m", false, "Set maximum height, usually 50 lines")
	maxFlag := flag.Bool("max", false, "Set max height, usually 50 lines, alternative flag")

	c2 := flag.Bool("2", false, "Flag to set 2 column display mode.")
	c3 := flag.Bool("3", false, "Flag to set 3 column display mode.")
	flag.BoolVar(&allFlag, "a", false, "Equivalent to 50 screens by default.  Intended to be used w/ the scroll back buffer.")

	flag.BoolVar(&fastFlag, "fast", false, "Fast debugging flag.  Used (so far) in MyReadDir.")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	maxDimFlag = *mFlag || *maxFlag           // either m or max options will set this flag and suppress use of halfFlag.
	if halfFlag || allFlag || *nscreens > 1 { // To make sure that a full screen of lines is the base for subsequent calculations when these conditions are met.
		maxDimFlag = true // The need arose for this when I'm using the environment to reduce the # of lines displayed routinesly.
	} // Added Jan 6, 2025.

	Reverse := *revflag || RevFlag || dsrtParam.reverseflag
	Forward := !Reverse // convenience variable

	SizeSort := *sizeflag || SizeFlag || dsrtParam.sizeflag
	DateSort := !SizeSort // convenience variable

	if NLines > 0 { // priority is -N option
		numOfLines = NLines
	} else if dsrtParam.paramNum > 0 && !maxDimFlag { // Use dsrt environ value if the -m or -M not used on command line.  Added Jan 5, 2025
		numOfLines = dsrtParam.paramNum
	} else if autoHeight > 0 { // finally use autoHeight.
		numOfLines = autoHeight - 7
	} else { // intended if autoHeight fails, like if the display is redirected
		numOfLines = defaultHeight
	}

	if numOfCols < 1 {
		numOfCols = 1
	} else if numOfCols > 3 {
		numOfCols = 3
	}
	if numOfCols == 1 {
		if *c2 {
			numOfCols = 2
		} else if *c3 {
			numOfCols = 3
		}
	}

	if allFlag { // if both nscreens and allScreens are used, allFlag takes precedence.
		*nscreens = allScreens // defined above to be a non-zero number, currently 50 as of this writing.
	}
	numOfLines *= *nscreens // updated 18 Feb 24.

	if (halfFlag || dsrtParam.halfFlag) && !maxDimFlag { // halfFlag could be set by environment var, but overridden by use of maxDimFlag.
		numOfLines /= 2
	}

	if dsrtParam.filterflag && !*noFilterFlag {
		filterFlag = true
	}

	noExtensionFlag = *extensionflag || *extflag

	if globFlag {
		fmt.Printf(" Glob flag has been removed.  This flag is now ignored here; it remains only in dsrt.\n")
		globFlag = false
	}

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execName)
		fmt.Println()
		fmt.Println("winFlag:", winFlag)
		fmt.Println()
		fmt.Printf(" After flag.Parse(); option switches w=%d, nscreens=%d, Nlines=%d, numOfCols=%d\n", w, *nscreens, NLines, numOfCols)
	}

	if len(excludeRegexPattern) > 0 {
		if verboseFlag {
			fmt.Printf(" excludeRegexPattern is longer than 0 runes.  It is %d runes. \n", len(excludeRegexPattern))
		}
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeFlag = true
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			excludeFlag = false
		}
	} else if excludeFlag {
		ctfmt.Print(ct.Yellow, winFlag, " Enter regex pattern to be excluded: ")
		fmt.Scanln(&excludeRegexPattern)
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			excludeFlag = false
		}
	}

	dirList = *DirListFlag || FilenameListFlag || dsrtParam.dirlistflag || dsrtParam.filenamelistflag // if -D entered then this expression also needs to be true.
	filenameToBeListedFlag = !(FilenameListFlag || dsrtParam.filenamelistflag)                        // need to reverse the flag because D means suppress the output of filenames.
	longFileSizeListFlag = *longflag
	ShowGrandTotal := *TotalFlag || dsrtParam.totalflag // added 09/12/2018 12:32:23 PM

	// set w, the width param, ie, number of columns available
	if w == 0 {
		w = dsrtParam.w
	}
	if autoWidth > 0 {
		if w <= 0 || w > maxWidth { // w not set by flag.Parse or dsw environ var
			w = autoWidth
		}
	} else {
		if w <= 0 || w > maxWidth {
			if numOfCols == 1 {
				w = minWidth
			} else if numOfCols == 2 {
				w = min2Width
			} else {
				w = min3Width
			}
		}
	}
	// check min widths
	if numOfCols == 3 && w < min3Width {
		fmt.Printf(" Width of %d is less than minimum of %d for %d column output.  Will make column = 1.\n", w, min3Width, numOfCols)
		numOfCols = 1
	} else if numOfCols == 2 && w < min2Width {
		fmt.Printf(" Width of %d is less than minimum of %d for %d column output.  Will make column = 1.\n", w, min2Width, numOfCols)
		numOfCols = 1
	} else if numOfCols == 1 && w < minWidth {
		fmt.Printf(" Width of %d is less than minimum of %d for %d column output.  Output may not look good.\n", w, minWidth, numOfCols)
	}

	// set which sort function will be in the sortfcn var
	sortfcn := func(i, j int) bool { return false } // became available as of Go 1.8
	if SizeSort && Forward {                        // set the value of sortfcn so only a single line is needed to execute the sort.
		sortfcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfos[i].Size() > fileInfos[j].Size() // I want a largest first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfos[i].ModTime().After(fileInfos[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
		}
		if verboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfos[i].Size() < fileInfos[j].Size() // I want an smallest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime()) // I want an oldest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if verboseFlag {
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			fmt.Printf(" uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s.\n",
				uid, gid, systemStr, userPtr.Uid, userPtr.Gid, userPtr.Username, userPtr.Name, userPtr.HomeDir)
		}
		fmt.Printf(" Autoheight=%d, autowidth=%d, w=%d, numOfLines=%d, numOfCols=%d. \n", autoHeight, autoWidth, w, numOfLines, numOfCols)
		fmt.Printf(" dsrtparam paramNum=%d, w=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelist=%t, totalflag=%t\n",
			dsrtParam.paramNum, dsrtParam.w, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag,
			dsrtParam.totalflag)
	}

	// If the character is a letter, it has to be k, m or g.  Or it's a number, but not both.  For now.
	if filterFlag {
		filterAmt = 1_000_000
	} else if filterStr != "" {
		if len(filterStr) > 1 {
			filterAmt, err = strconv.Atoi(filterStr)
			if err != nil {
				fmt.Fprintln(os.Stderr, "converting filterStr:", err)
			}
		} else if unicode.IsLetter(rune(filterStr[0])) {
			filterStr = strings.ToLower(filterStr)
			if filterStr == "k" {
				filterAmt = 1000
			} else if filterStr == "m" {
				filterAmt = 1_000_000
			} else if filterStr == "g" {
				filterAmt = 1_000_000_000
			} else {
				fmt.Fprintln(os.Stderr, "filterStr is not valid and was ignored.  filterStr=", filterStr)
			}
		} else {
			fmt.Fprintln(os.Stderr, "filterStr not valid.  filterStr =", filterStr)
		}
	}

	if verboseFlag {
		fmt.Println(" FilterFlag =", filterFlag, ", filterStr =", filterStr, ", filterAmt =", filterAmt, ", globFlag =", globFlag)
	}

	t0 := time.Now()

	fileInfos = getFileInfosFromCommandLine() // this rtn is in dsutil_windows.go and dsutil_linux.go.  So go vet gets this wrong.

	elapsed := time.Since(t0)
	if verboseFlag {
		fmt.Printf(" in main, after getFileInfosFromCommandLine(): Length(fileInfos) = %d, elapsed = %s\n", len(fileInfos), elapsed)
	}
	if len(fileInfos) > 1 {
		sort.Slice(fileInfos, sortfcn)
	}

	cs := getColorizedStrings(fileInfos, numOfCols)

	if verboseFlag {
		fmt.Printf(" Len(fileinfos)=%d, len(colorizedStrings)=%d, numOfLines=%d\n", len(fileInfos), len(cs), numOfLines)
	}

	// Output the colorized string slice
	columnWidth := w/numOfCols - 2
	for i := 0; i < len(cs); i += numOfCols {
		c0 := cs[i].color
		s0 := fixedStringLen(cs[i].str, columnWidth)
		ctfmt.Printf(c0, winFlag, "%s", s0)

		if numOfCols > 1 && (i+1) < len(cs) { // numOfCols of 2 or 3
			c1 := cs[i+1].color
			s1 := fixedStringLen(cs[i+1].str, columnWidth)
			ctfmt.Printf(c1, winFlag, "  %s", s1)
		}

		if numOfCols == 3 && (i+2) < len(cs) {
			c2 := cs[i+2].color
			s2 := fixedStringLen(cs[i+2].str, columnWidth)
			ctfmt.Printf(c2, winFlag, "  %s", s2)
		}
		fmt.Println()
	}
	fmt.Println()

	s := fmt.Sprintf("%d", sizeTotal)
	if sizeTotal > 100000 {
		s = AddCommas(s)
	}
	s0 := fmt.Sprintf("%d", GrandTotal)
	if GrandTotal > 100000 {
		s0 = AddCommas(s0)
	}
	fmt.Printf(" Elapsed = %s, len(fileInfos) = %d, File Size total = %s", elapsed, len(fileInfos), s)
	if ShowGrandTotal {
		s1, color := getMagnitudeString(GrandTotal)
		ctfmt.Println(color, true, ", Directory grand total is", s0, "or approx", s1, "in", grandTotalCount, "files.")
	} else {
		fmt.Println(".")
	}
} // end main ds

//-------------------------------------------------------------------- InsertByteSlice

func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

//---------------------------------------------------------------------- AddCommas

func AddCommas(instr string) string {
	Comma := []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas

// ------------------------------ IsSymlink ---------------------------

func IsSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink

// ------------------------------------ ProcessEnvironString ---------------------------------------

func ProcessEnvironString(dsrtEnv, dswEnv string) DsrtParamType { // use system utils when can because they tend to be faster
	// 4 Jul 23 -- the use of strings.Split is redundant.  I removed it here.

	var dsrtParam DsrtParamType

	if dswEnv == "" {
		dsrtParam.w = 0 // this is redundant because it's initialized to zero.
	} else {
		n, err := strconv.Atoi(dswEnv)
		if err == nil {
			dsrtParam.w = n
		} else {
			fmt.Fprintf(os.Stderr, " dsw environment variable not a valid number.  dswStr = %q, %v.  Ignored.", dswEnv, err)
			dsrtParam.w = 0
		}
	}

	if dsrtEnv == "" {
		return dsrtParam
	}

	// The strings.Split creates slices of individual character strings.  But it's redundant now that I look at it 7/4/23.

	for _, envChar := range dsrtEnv {
		if envChar == 'r' || envChar == 'R' {
			dsrtParam.reverseflag = true
		} else if envChar == 's' || envChar == 'S' {
			dsrtParam.sizeflag = true
		} else if envChar == 'd' {
			dsrtParam.dirlistflag = true
		} else if envChar == 'D' {
			dsrtParam.filenamelistflag = true
		} else if envChar == 't' { // added 09/12/2018 12:26:01 PM
			dsrtParam.totalflag = true // for the grand total operation
		} else if envChar == 'f' {
			dsrtParam.filterflag = true
		} else if envChar == 'h' {
			dsrtParam.halfFlag = true
		} else if unicode.IsDigit(rune(envChar)) {
			d := envChar - '0'
			dsrtParam.paramNum = 10*dsrtParam.paramNum + int(d)
		}
	}
	return dsrtParam
} // end ProcessEnvironString

// ------------------------------ GetDirectoryAliases ----------------------------------------
func getDirectoryAliases() dirAliasMapType { // Env variable is diraliases.

	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = MakeSubst(s, '_', ' ') // substitute the underscore, _, or a space
	directoryAliasesMap := make(dirAliasMapType, 10)
	//anAliasMap := make(dirAliasMapType,1)

	dirAliasSlice := strings.Fields(s)

	for _, aliasPair := range dirAliasSlice {
		if string(aliasPair[len(aliasPair)-1]) != "\\" {
			aliasPair = aliasPair + "\\"
		}
		aliasPair = MakeSubst(aliasPair, '-', ' ') // substitute a dash,-, for a space
		splitAlias := strings.Fields(aliasPair)
		directoryAliasesMap[splitAlias[0]] = splitAlias[1]
	}
	return directoryAliasesMap
} // end getDirectoryAliases

// --------------------------- MakeSubst -------------------------------------------

func MakeSubst(instr string, r1, r2 rune) string {

	inRune := make([]rune, len(instr))
	if !strings.ContainsRune(instr, r1) {
		return instr
	}

	for i, s := range instr {
		if s == r1 {
			s = r2
		}
		inRune[i] = s // was byte(s) before I made this a slice of runes.
	}
	return string(inRune)
} // makesubst

// ------------------------------ ProcessDirectoryAliases ---------------------------

func ProcessDirectoryAliases(cmdline string) string {

	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap := getDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := aliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathnFile := cmdline[idx+1:]
	completeValue := aliasValue + PathnFile
	fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
	return completeValue
} // ProcessDirectoryAliases

// ------------------------------- myReadDir (not concurrent one) -----------------------------------

//func myReadDir(dir string) []os.FileInfo { // The entire change including use of []DirEntry happens here.  Old one, not concurrent code.
//	dirEntries, err := os.ReadDir(dir)
//	if err != nil {
//		return nil
//	}
//
//	fileInfos := make([]os.FileInfo, 0, len(dirEntries))
//	for _, d := range dirEntries {
//		fi, e := d.Info()
//		if e != nil {
//			fmt.Fprintf(os.Stderr, " Error from %s.Info() is %v\n", d.Name(), e)
//			continue
//		}
//		if includeThis(fi) {
//			fileInfos = append(fileInfos, fi)
//		}
//		if fi.Mode().IsRegular() && showGrandTotal {
//			grandTotal += fi.Size()
//			grandTotalCount++
//		}
//	}
//	return fileInfos
//} // myReadDir

// ------------------------------- myReadDir concurrent code is here -----------------------------------

func myReadDir(dir string) []os.FileInfo { // The entire change including use of []DirEntry happens here.  With concurrent code.
	// Adding concurrency in returning []os.FileInfo

	var wg sync.WaitGroup

	if verboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	deChan := make(chan []os.DirEntry, numWorkers) // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fiChan := make(chan os.FileInfo, numWorkers)   // of individual file infos to be collected and returned to the caller of this routine.
	doneChan := make(chan bool)                    // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fiSlice := make([]os.FileInfo, 0, fetch*multiplier*multiplier)
	wg.Add(numWorkers)
	for range numWorkers { // reading from deChan to get the slices of DirEntry's
		go func() {
			defer wg.Done()
			for deSlice := range deChan {
				for _, de := range deSlice {
					fi, err := de.Info()
					if err != nil {
						fmt.Printf("Error getting file info for %s: %v, ignored\n", de.Name(), err)
						continue
					}
					if includeThis(fi) {
						fiChan <- fi
					}
				}
			}
		}()
	}

	go func() { // collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?  I figured it out, by closing the channel after all work is sent to it.
		for fi := range fiChan {
			fiSlice = append(fiSlice, fi)
			if fi.Mode().IsRegular() && showGrandTotal {
				grandTotal += fi.Size()
				grandTotalCount++
			}
		}
		close(doneChan)
	}()

	d, err := os.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error os.open(%s) is %s.  exiting.\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()

	for {
		// reading DirEntry's and sending the slices into the channel needs to happen here.
		deSlice, err := d.ReadDir(fetch)
		if errors.Is(err, io.EOF) { // finished.  So return the slice.
			close(deChan)
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR from %s.ReadDir(%d) is %s.\n", dir, numWorkers, err)
			continue
		}
		deChan <- deSlice
	}

	wg.Wait()     // for the deChan
	close(fiChan) // This way I only close the channel once.  I think if I close the channel from within a worker, and there are multiple workers, closing an already closed channel panics.

	<-doneChan // block until channel is freed

	if verboseFlag {
		fmt.Printf("Found %d files in directory %s.\n", len(fiSlice), dir)
	}

	if fastFlag {
		fmt.Printf("Found %d files in directory %s, first few entries is %v.\n", len(fiSlice), dir, fiSlice[:5])
		if pause() {
			os.Exit(1)
		}
	}

	return fiSlice
} // myReadDir

func myReadDirWithMatch(dir, matchPat string) []os.FileInfo { // The entire change including use of []DirEntry happens here, and now concurrent code.
	// Adding concurrency in returning []os.FileInfo
	// This routine add a call to filepath.Match

	var wg sync.WaitGroup

	if verboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	deChan := make(chan []os.DirEntry, numWorkers) // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fiChan := make(chan os.FileInfo, numWorkers)   // of individual file infos to be collected and returned to the caller of this routine.
	doneChan := make(chan bool)                    // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fiSlice := make([]os.FileInfo, 0, fetch*multiplier*multiplier)
	wg.Add(numWorkers)
	for range numWorkers { // reading from deChan to get the slices of DirEntry's
		go func() {
			defer wg.Done()
			for deSlice := range deChan {
				for _, de := range deSlice {
					fi, err := de.Info()
					if err != nil {
						fmt.Printf("Error getting file info for %s: %v, ignored\n", de.Name(), err)
						continue
					}
					if includeThisWithMatch(fi, matchPat) {
						fiChan <- fi
					}
				}
			}
		}()
	}

	go func() { // collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?  I figured it out, by closing the channel after all work is sent to it.
		for fi := range fiChan {
			fiSlice = append(fiSlice, fi)
			if fi.Mode().IsRegular() && showGrandTotal {
				grandTotal += fi.Size()
				grandTotalCount++
			}
		}
		close(doneChan)
	}()

	d, err := os.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error os.open(%s) is %s.  exiting.\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()

	for {
		// reading DirEntry's and sending the slices into the channel needs to happen here.
		deSlice, err := d.ReadDir(fetch)
		if errors.Is(err, io.EOF) { // finished.  So now can close the deChan.
			close(deChan)
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR from %s.ReadDir(%d) is %s.\n", dir, numWorkers, err)
			continue
		}
		deChan <- deSlice
	}

	wg.Wait()     // for the closing of the deChan to stop all worker goroutines.
	close(fiChan) // This way I only close the channel once.  I think if I close the channel from within a worker, and there are multiple workers, closing an already closed channel panics.

	<-doneChan // block until channel is freed

	if verboseFlag {
		fmt.Printf("Found %d files in directory %s.\n", len(fiSlice), dir)
	}

	if fastFlag {
		fmt.Printf("Found %d files in directory %s, first few entries is %v.\n", len(fiSlice), dir, fiSlice[:5])
		if pause() {
			os.Exit(1)
		}
	}

	return fiSlice
} // myReadDirWithMatch

// ----------------------------- getMagnitudeString -------------------------------

func getMagnitudeString(j int64) (string, ct.Color) {
	var s1 string
	var f float64
	var color ct.Color
	switch {
	case j > 1_000_000_000_000: // 1 trillion, or TB
		f = float64(j) / 1000000000000
		s1 = fmt.Sprintf("%.4g TB", f)
		color = ct.Red
	case j > 1_000_000_000: // 1 billion, or GB
		f = float64(j) / 1000000000
		s1 = fmt.Sprintf("%.4g GB", f)
		color = ct.White
	case j > 1_000_000: // 1 million, or MB
		f = float64(j) / 1000000
		s1 = fmt.Sprintf("%.4g mb", f)
		color = ct.Yellow
	case j > 1000: // KB
		f = float64(j) / 1000
		s1 = fmt.Sprintf("%.4g kb", f)
		color = ct.Cyan
	default:
		s1 = fmt.Sprintf("%3d bytes", j)
		color = ct.Green
	}
	return s1, color
}

// --------------------------------------------------- fixedStringlen ---------------------------------------

func fixedStringLen(s string, size int) string {
	var built strings.Builder

	if len(s) > size { // need to truncate the string
		return s[:size]
	} else if len(s) == size {
		return s
	} else if len(s) < size { // need to pad the string
		needSpaces := size - len(s)
		built.Grow(size)
		built.WriteString(s)
		spaces := strings.Repeat(" ", needSpaces)
		built.WriteString(spaces)
		return built.String()
	} else {
		fmt.Fprintln(os.Stderr, " fixedStringLen input string length is strange.  It is", len(s))
		return s
	}
} // end fixedStringLen

// ---------------------------------------------------- includeThis ----------------------------------------

func includeThis(fi os.FileInfo) bool {
	if veryVerboseFlag {
		fmt.Printf(" includeThis.  noExtensionFlag=%t, excludeFlag=%t, filterAmt=%d \n", noExtensionFlag, excludeFlag, filterAmt)
	}
	if noExtensionFlag && strings.ContainsRune(fi.Name(), '.') {
		return false
	} else if filterAmt > 0 {
		if fi.Size() < int64(filterAmt) {
			return false
		}
	}
	if excludeRegex != nil {
		if BOOL := excludeRegex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	return true
}

func includeThisWithMatch(fi os.FileInfo, matchPat string) bool {
	if veryVerboseFlag {
		fmt.Printf(" includeThis.  noExtensionFlag=%t, excludeFlag=%t, filterAmt=%d, match pattern=%s \n", noExtensionFlag, excludeFlag, filterAmt, matchPat)
	}
	if noExtensionFlag && strings.ContainsRune(fi.Name(), '.') {
		return false
	} else if filterAmt > 0 {
		if fi.Size() < int64(filterAmt) {
			return false
		}
	}
	if excludeFlag {
		if excludeRegex.MatchString(strings.ToLower(fi.Name())) {
			return false
		}
	}
	matchPat = strings.ToLower(matchPat)
	f := strings.ToLower(fi.Name())
	match, err := filepath.Match(matchPat, f)
	if err != nil {
		return false
	}
	if !match {
		return false
	}
	return true
}

// ------------------------------ pause -----------------------------------------

func pause() bool {
	fmt.Print(" Pausing the loop.  Hit <enter> to continue; 'n' or 'x' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "x") {
		return true
	}
	return false
}
