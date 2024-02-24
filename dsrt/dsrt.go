// dsrt.go -- directory sort

package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

/*
REVISION HISTORY
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
20 Dec 20 -- For date sorting, I changed away from using NanoSeconds and I'm now using the time.Before(time) and time.After(time) functions.
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
26 Aug 21 -- Back porting autoHeight and autoWidth
22 Oct 21 -- Updating the idiom that uses bytes.buffer.
16 Jan 22 -- Updating how the help message is created, learned from "Powerful Command-Line Applications in Go" by Ricardo Gerardi
26 Jan 22 -- Adding a verbose flag
27 Jan 22 -- Full refactoring to use a lot more platform specific code instead of all the if windows or if linux stuff.
29 Jan 22 -- Refactoring is done.  Now to add -g option which is ignored on linux but on Windows it means to use the Glob function.
 1 Feb 22 -- Added veryVerboseFlag, and optimized includeThis.
 3 Feb 22 -- Finally reversed the -x and -exclude options, so now -x means I enter the exclude regex on the command line.  Whew!
               Current logic has the getFileInfos routine process the command line options and params, determines which files match
               the provided pattern, which don't match the exclude regex, which are filtered out by size, and returns what's left.
 8 Feb 22 -- Fixing a bug w/ the -g globbing option on Windows.
10 Feb 22 -- Fixing a bug in MyReadDir when an error occurs.
14 Feb 22 -- Fix bug of not treating an absolute path one that begins w/ the filepath.Separator character.  Actual fix is in _linux.go file.
15 Feb 22 -- Really replaced testFlag w/ VerboseFlag, because as I maintain the code, I forget if this has verboseFlag.  Now it does and doesn't have testFlag.
16 Feb 22 -- Time to remove the upper case flags that I don't use.
24 Feb 22 -- Fixed a bug in the glob option.  And Evan's 30 today.  Wow.
25 Apr 22 -- Added the -1 flag and it's halfFlag variable.  For displaying half the number of lines the screen allows.
14 Oct 22 -- Adding an undo option for the -1 flag, as I want to make it default thru the dsrt env var.  Or something like that.  I'm still thinking.
15 Oct 22 -- I noticed that the environment string can't process f, for filterFlag.  Now it can.
               Now I need an option, -F, to undo the filterflag set in an environment var.
11 Nov 22 -- Will output environ var settings on header.  They're easy to forget :-)
21 Nov 22 -- static linter found an issue that I'm going to fix.
14 Jan 23 -- I wrote args to learn more about how arguments are handled.  I think I got it wrong in dsrtutil_linux.  I'm going to fix it.
 6 Feb 23 -- Directory aliases still don't work perfectly.  I'm going to debug this a bit more.  Nevermind, the problem was a bad directory alias that mapped to the wrong drive letter.
 7 Feb 23 -- I found an area in dsrtutil_linux.go where I didn't close a file that needed to be closed.  I fixed that.
15 Feb 23 -- Added showing the timestamp of the binary.  And then removed it when I saw that it's already there in the verbose output.  And I tweaked the beginning verbose output.
12 Apr 23 -- Fixing a panic when run in a docker image.  Issue is in GetIDName.  I'll fix the bug and change this name to be more idiomatic in Go, ie, idName.
14 Apr 23 -- Tweaked output of the error in idName, formerly GetIDName.
18 Apr 23 -- Removed the error message in idName
28 Jun 23 -- Starting to think about a scroll switch, which will allow screen by screen output until I stop it, probably by hitting <ESC> or 'q'.
                I would need some way to tracking lines to be displayed, possibly using a slice like I do in the list based routines.  Or adding a beg, end value to show.
                Or just a beginning value.
                Right now, I can use the scroll back buffer to achieve the same thing when I use a large value of n (number of screens).
                The more I think about this, using the scroll back buffer is probably best because this affects ds and rex, and maybe others I can't think of right now.
30 Jun 23 -- I'm adding -a flag, to mean all.  It will be equivalent to 100 screens, or setting nscreen to 100.  For now.  Changed to default of 50 a few days later.
 2 Jul 23 -- I'm going to change the environ var number to mean number of screens when the all flag is used.  Not today, it's too late.  I'll start this tomorrow.
                I'll leave -N to mean lines/screen, as there's no point in having a command line switch to change that;  I could just use the nscreen option directly.
                I have to change dsrtparam.numlines, and the processing section for numlines.  I think I'll create a var along the lines of allScreens, defaulting to 50 or so.
 3 Jul 23 -- Added environment var h to mean halfFlag.
 4 Jul 23 -- Improved ProcessEnvironString.
18 Feb 24 -- Changed a message to make it clear that this sorts on mod date, and removed the unused sizeFlag.
*/

const LastAltered = "18 Feb 2024"

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, thru use of OS specific code.  On Windows it will get a pattern from the command line.
// but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

type dirAliasMapType map[string]string

type DsrtParamType struct {
	numscreens                                                                            int // set by dsrt environ var.
	reverseflag, sizeflag, dirlistflag, filenamelistflag, totalflag, filterflag, halfFlag bool
}

const defaultHeight = 40
const minWidth = 90

var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, verboseFlag bool
var filterFlag, globFlag, veryVerboseFlag, halfFlag, maxDimFlag bool
var filterAmt, numLines, numOfLines, grandTotalCount int
var sizeTotal, grandTotal int64
var filterStr string
var excludeRegex *regexp.Regexp

// allScreens is the number of screens to be used for the allFlag switch.  This can be set by the environ var dsrt.
var allScreens = 50

// this is to be equivalent to allScreens screens, by default same as n=50.
var allFlag bool

//var directoryAliasesMap dirAliasMapType // this was unused after I removed a redundant statement in dsrtutil_windows

func main() {
	var dsrtParam DsrtParamType
	var userPtr *user.User // from os/user
	var err error
	var autoWidth, autoHeight int
	var excludeRegexPattern string
	var fileInfos []os.FileInfo

	uid := 0
	gid := 0
	systemStr := ""

	//execName, _ := os.Executable()
	//ExecFI, _ := os.Stat(execName)
	//LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	// environment variable processing.  If present, these will be the defaults.  Processed before the flags so the flags will override these, if provided on the command line.
	dsrtEnviron := os.Getenv("dsrt")
	dsrtParam = ProcessEnvironString(dsrtEnviron) // This is a function below.
	winflag := runtime.GOOS == "windows"          // this is needed because I use it in the color statements, so the colors are bolded only on windows.
	ctfmt.Printf(ct.Magenta, winflag, "dsrt will display Directory SoRTed by date or size.  LastAltered %s, compiled with %s",
		LastAltered, runtime.Version())
	if dsrtEnviron == "" {
		ctfmt.Printf(ct.Magenta, winflag, "\n")
	} else {
		ctfmt.Printf(ct.Green, winflag, ", dsrt env = %s \n", dsrtEnviron)
	}

	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

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
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " dsrt environ values are: numscreens=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelistflag=%t, totalflag=%t, halfFlag=%t \n",
			dsrtParam.numscreens, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag, dsrtParam.totalflag, dsrtParam.halfFlag)

		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	revflag := flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
	var RevFlag bool                                                                      // this will always be false.  I can leave it this way for now.
	// flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nscreens = flag.Int("n", 1, "number of screens to display, ie, a multiplier for numOfLines") // Ptr
	var NLines int
	flag.IntVar(&NLines, "N", numOfLines, "number of lines to display, and takes priority over the auto settings.") // Value

	var sizeflag = flag.Bool("s", false, "sort by size instead of by mod date") // pointer

	var DirListFlag = flag.Bool("d", false, "include directories in the output listing") // pointer

	var FilenameListFlag bool
	flag.BoolVar(&FilenameListFlag, "D", false, "Directories only in the output listing")

	var TotalFlag = flag.Bool("t", false, "include grand total of directory, makes most sense when no pattern is given on command line.")

	// var testFlag bool  Now set globally
	flag.BoolVar(&verboseFlag, "test", false, "enter a testing mode to println more variables")
	flag.BoolVar(&verboseFlag, "v", false, "verbose mode, which is same as test mode.")

	var longflag = flag.Bool("l", false, "long file size format.") // Ptr

	var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")
	flag.StringVar(&excludeRegexPattern, "x", "", "regex to be excluded from output.")

	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVar(&filterFlag, "f", false, "filter value to suppress listing individual size below 1 MB.")
	noFilterFlag := flag.Bool("F", false, "Flag to undo an environment var with f set.")

	flag.BoolVar(&globFlag, "g", false, "Use glob function on Windows.")

	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose option for when I really want it.")
	flag.BoolVar(&halfFlag, "1", false, "display 1/2 of the screen.")

	mFlag := flag.Bool("m", false, "Set maximum height, usually 50 lines")
	maxFlag := flag.Bool("max", false, "Set max height, usually 50 lines, alternative flag")

	flag.BoolVar(&allFlag, "a", false, "Equivalent to 50 screens by default.  Intended to be used w/ the scroll back buffer.")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerbose flag will also set verbose flag, ie testFlag.
		verboseFlag = true
	}

	maxDimFlag = *mFlag || *maxFlag // either m or max options will set this flag and suppress use of halfFlag.

	if NLines > 0 { // priority
		numOfLines = NLines
		//                       } else if dsrtParam.numlines > 0 && !maxDimFlag { // then check this, but only if maxDimFlag is not set.
		//                 	         numOfLines = dsrtParam.numlines  These lines removed when I added allFlag and dsrtparam.numScreens
	} else if autoHeight > 0 { // finally use autoHeight.
		numOfLines = autoHeight - 7
	} else { // intended if autoHeight fails, like of the output is being redirected.
		numOfLines = defaultHeight
	}

	if dsrtParam.numscreens > 0 {
		allScreens = dsrtParam.numscreens
	}
	if allFlag { // if both nscreens and allScreens are used, allFlag takes precedence.
		*nscreens = allScreens
	}
	numOfLines *= *nscreens // Doesn't matter if *nscreens = 1

	if (halfFlag || dsrtParam.halfFlag) && !maxDimFlag { // halfFlag could be set by environment var, but overridden by use of maxDimFlag.
		numOfLines /= 2
	}

	if dsrtParam.filterflag && !*noFilterFlag {
		filterFlag = true
	}

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execName)
		fmt.Println()
		if runtime.GOARCH == "amd64" && runtime.GOOS == "linux" {
			fmt.Printf("uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
				uid, gid, systemStr, userPtr.Uid, userPtr.Gid, userPtr.Username, userPtr.Name, userPtr.HomeDir)
		}
		fmt.Printf(" dsrtparam numscreens=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelist=%t, totalflag=%t, halfFlag=%t\n",
			dsrtParam.numscreens, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag,
			dsrtParam.totalflag, dsrtParam.halfFlag)
		fmt.Printf(" autoheight=%d, autowidth=%d, excludeFlag=%t. \n", autoHeight, autoWidth, excludeFlag)
	}

	Reverse := *revflag || RevFlag || dsrtParam.reverseflag
	Forward := !Reverse // convenience variable

	SizeSort := *sizeflag || dsrtParam.sizeflag
	DateSort := !SizeSort // convenience variable

	if verboseFlag {
		fmt.Printf(" dsrtParam.numscreens=%d, NLines=%d, autoheight=%d, defaultHeight=%d, and finally numOfLines = %d  \n",
			dsrtParam.numscreens, NLines, autoHeight, defaultHeight, numOfLines)
	}
	noExtensionFlag = *extensionflag || *extflag

	if len(excludeRegexPattern) > 0 {
		if verboseFlag {
			fmt.Printf(" excludeRegexPattern is longer than 0 runes.  It is %d runes. \n", len(excludeRegexPattern))
		}
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			excludeFlag = false
		}
		excludeFlag = true
		if verboseFlag {
			fmt.Printf(" Regex condition: excludeFlag=%t, excludeRegex=%v\n", excludeFlag, excludeRegex.String())
		}
	} else if excludeFlag {
		ctfmt.Print(ct.Yellow, winflag, " Enter regex pattern to be excluded: ")
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
	showGrandTotal = *TotalFlag || dsrtParam.totalflag // added 09/12/2018 12:32:23 PM

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
		fmt.Println(" *** Here I am ***")
		fmt.Println(" FilterFlag =", filterFlag, ".  filterStr =", filterStr, ". filterAmt =", filterAmt, "excludeFlag =", excludeFlag)
		fmt.Printf(" nscreens=%d, numLines=%d, flag.NArgs=%d, dirList=%t, Filenametobelistedflag=%t, longfilesizelistflag=%t, showgrandtotal=%t\n",
			*nscreens, numLines, flag.NArg(), dirList, filenameToBeListedFlag, longFileSizeListFlag, showGrandTotal)
	}

	fileInfos = getFileInfosFromCommandLine()
	if verboseFlag {
		fmt.Printf(" After call to getFileInfosFromCommandLine.  flag.NArg=%d, len(fileinfos)=%d, numOfLines=%d\n", flag.NArg(), len(fileInfos), numOfLines)
	}
	if len(fileInfos) > 1 {
		sort.Slice(fileInfos, sortfcn) // must be sorted here for sortfcn to work correctly, because the slice name it uses must be correct.  Better if that name is not global.
	}

	displayFileInfos(fileInfos)

	s := fmt.Sprintf("%d", sizeTotal)
	if sizeTotal > 100000 {
		s = AddCommas(s)
	}
	s0 := fmt.Sprintf("%d", grandTotal)
	if grandTotal > 100000 {
		s0 = AddCommas(s0)
	}
	fmt.Print(" File Size total = ", s)
	if showGrandTotal {
		s1, color := getMagnitudeString(grandTotal)
		ctfmt.Println(color, true, ", Directory grand total is", s0, "or approx", s1, "in", grandTotalCount, "files.")
	} else {
		fmt.Println(".")
	}
} // end main dsrt

//-------------------------------------------------------------------- InsertByteSlice

func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

//---------------------------------------------------------------------- AddCommas

func AddCommas(instr string) string {
	//var Comma []byte = []byte{','}  Getting error that type can be omitted
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

// ---------------------------- GetIDname -----------------------------------------------------------

func idName(uidStr string) string {
	if len(uidStr) == 0 {
		return ""
	}
	ptrToUser, err := user.LookupId(uidStr)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "%s:  ", err)  too noisy in a docker container.
		return "----" // this line fixes the bug if user.LookupId failed, as it does in a docker image.
	}
	return ptrToUser.Username
} // idName, formerly GetIDname

// ------------------------------------ ProcessEnvironString ---------------------------------------

func ProcessEnvironString(envStr string) DsrtParamType { // use system utils when can because they tend to be faster
	// 4 Jul 23 -- the use of strings.Split is redundant.  I removed it here.  When I wrote that code, I probably forgot
	//             that I can iterate over a string and will get out individual runes.
	//             I may have done that so I can look at the next digit.  Not needed now.

	var dsrtParam DsrtParamType

	if envStr == "" {
		return dsrtParam
	} // empty dsrtparam is returned

	// The strings.Split creates slices of individual character strings.  But it's redundant now that I look at it 7/4/23.

	for _, s := range envStr {
		if s == 'r' || s == 'R' {
			dsrtParam.reverseflag = true
		} else if s == 's' || s == 'S' {
			dsrtParam.sizeflag = true
		} else if s == 'd' {
			dsrtParam.dirlistflag = true
		} else if s == 'D' {
			dsrtParam.filenamelistflag = true
		} else if s == 't' { // added 09/12/2018 12:26:01 PM
			dsrtParam.totalflag = true // for the grand total operation
		} else if s == 'f' {
			dsrtParam.filterflag = true
		} else if s == 'h' {
			dsrtParam.halfFlag = true
		} else if unicode.IsDigit(s) {
			d := s - '0'
			dsrtParam.numscreens = 10*dsrtParam.numscreens + int(d)
		}
	}
	return dsrtParam
}

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

// ------------------------------ GetDirectoryAliases ----------------------------------------
func getDirectoryAliases() dirAliasMapType { // Env variable is diraliases.
	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = MakeSubst(s, '_', ' ') // substitute the underscore, _, for a space
	directoryAliasesMap := make(dirAliasMapType, 10)

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
}

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

// ------------------------------- myReadDir -----------------------------------

func myReadDir(dir string) []os.FileInfo { // The entire change including use of []DirEntry happens here.  Who knew?
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil
	}

	fileInfos := make([]os.FileInfo, 0, len(dirEntries))
	for _, d := range dirEntries {
		fi, e := d.Info()
		if e != nil {
			fmt.Fprintf(os.Stderr, " Error from %s.Info() is %v\n", d.Name(), e)
			continue
		}
		if includeThis(fi) {
			fileInfos = append(fileInfos, fi)
		}
		if fi.Mode().IsRegular() && showGrandTotal {
			grandTotal += fi.Size()
			grandTotalCount++
		}
	}
	return fileInfos
} // myReadDir

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
	case j > 100_000_000_000: // 100 billion
		f = float64(j) / 1_000_000_000
		s1 = fmt.Sprintf(" %.4g GB", f)
		color = ct.White
	case j > 10_000_000_000: // 10 billion
		f = float64(j) / 1_000_000_000
		s1 = fmt.Sprintf("  %.4g GB", f)
		color = ct.White
	case j > 1_000_000_000: // 1 billion, or GB
		f = float64(j) / 1000000000
		s1 = fmt.Sprintf("   %.4g GB", f)
		color = ct.White
	case j > 100_000_000: // 100 million
		f = float64(j) / 1_000_000
		s1 = fmt.Sprintf("    %.4g mb", f)
		color = ct.Yellow
	case j > 10_000_000: // 10 million
		f = float64(j) / 1_000_000
		s1 = fmt.Sprintf("     %.4g mb", f)
		color = ct.Yellow
	case j > 1_000_000: // 1 million, or MB
		f = float64(j) / 1000000
		s1 = fmt.Sprintf("      %.4g mb", f)
		color = ct.Yellow
	case j > 100_000: // 100 thousand
		f = float64(j) / 1000
		s1 = fmt.Sprintf("       %.4g kb", f)
		color = ct.Cyan
	case j > 10_000: // 10 thousand
		f = float64(j) / 1000
		s1 = fmt.Sprintf("        %.4g kb", f)
		color = ct.Cyan
	case j > 1000: // KB
		f = float64(j) / 1000
		s1 = fmt.Sprintf("         %.3g kb", f)
		color = ct.Cyan
	default:
		s1 = fmt.Sprintf("%3d bytes", j)
		color = ct.Green
	}
	return s1, color
}

// --------------------------------------------- includeThis ----------------------------------------------------------

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
	if excludeFlag {
		if BOOL := excludeRegex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	return true
}
