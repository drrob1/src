// rexv.go from rex.go -- directory sort using a regular expression pattern on the filename

package main

import (
	"errors"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	//"unicode"
	//"bytes"
	//"os/exec"
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
               I changed the meanings so now use <symlink> and (dir) indicators, and fixed the oversight on Windows
               whereby symlinks could not be displayed.
20 Jun 19 -- Changed logic so that symlinks to files are always displayed, like files.
               That required writing a new function to detect a symlink.
23 Jun 19 -- Changed to use Lstat when there are multiple filenames on the command line.  This only happens on Linux.
 2 Jul 19 -- Changed the format pattern for displaying the executable timestamp.  And Lstat error processing changed.
 3 Jul 19 -- Removing a confusing comment, and removed need for a flag variable for issymlink
 4 Jul 19 -- Removed the pattern check code on linux.  And this revealed a bug on linux if only 1 file is globbed on command line.  Now fixed.
 5 Jul 19 -- Optimized order of printing file types.  I hope.
18 Jul 19 -- When there is an error from ioutil.ReadDir, I cannot change its behavior of not reading any more.  Just do dsrt * in bash as a workaround.
19 Jul 19 -- Wrote MyReadDir
22 Jul 19 -- Added a winflag check so don't scan commandline on linux looking for : or ~.
 9 Sep 19 -- From Israel: Fixing issue on linux when entering a directory param.  And added test flag.  And added sortfcn.
22 Sep 19 -- Changed the error message under linux and have only 1 item on command line.  Error condition is likely file not found.
 4 Oct 19 -- No longer need platform specific code.  So I added GetUserGroupStrLinux.  And then learned that it won't compile on Windows.
               So as long as I want the exact same code for both platforms, I do need platform specific code.
------------------------------------------------------------------------------------------------------------------------------------------------------
 5 Oct 19 -- Started writing this as regex.go.  Will not display uid:gid.  If need that, need to use dsrt.  And doesn't have -x flag to exclude.
 6 Oct 19 -- Added help as a flag, removed -H, and expanded help to include the basics of regex syntax.
 8 Oct 19 -- Decided to work like dsrt, in that if there is no pattern, just show all recent files.  And I removed dead code, that's still in dsrt.
               Adding new usage to allow 'pattern' 'directory'.  Directory can be null to mean current dir.
27 Oct 19 -- Lower casing the regular expression so it matchs the lower cased filenames.  And added to help message.
21 Nov 19 -- Added Println() statements to separate header from filename outputs.
25 Aug 20 -- File sizes to be displayed in up to 3 digits and a suffix of kb, mb, gb and tb.  Unless new -l for long flag is used.
 9 Nov 20 -- Now using correct idiom to read environment and check for absent variable.
20 Dec 20 -- For date sorting, I changed away from using NanoSeconds and I'm now using the time.Before(time) and time.After(time) functions.
                 I found these to be much faster when I changed dsrt.go.
15 Jan 21 -- Now uses same getMagnitudeString as I wrote for dsrt.
17 Jan 21 -- Adding -x flag, for an exclude pattern, ie, if this pattern matches, don't print.
31 Jan 21 -- Adding color.
13 Feb 21 -- Swapping white and cyan.
15 Feb 21 -- Swapping yellow and white so yellow is mb and white is gb.
 2 Mar 21 -- Adding runtime.Version(), which I read about in Go Standard Library Cookbook.
 9 Mar 21 -- Added use of os.UserHomeDir, which became available as of Go 1.12.
17 Mar 21 -- Porting some recent changes in dsrt about ShowGrandTotal to here.
               Adding exclude string to allow the exclude regex pattern on the command line.  Convenient for recalling commands.
13 Jul 21 -- Now called reg.go, and will display its output in 2 columns like dsc.  ioutil is depracated, so that's now gone.
25 Jul 21 -- Now called rex.go, as reg conflicted on Windows w/ a registry edit pgm.
               The colors are a good way to give me the magnitude of filesize, so I don't need the displacements here.
               But I'm keeping the display of 4 significant figures.
               I'm adding the code to determine the number of rows and columns itself.  I'll use golang.org/x/term for linux, and shelling out to tcc for Windows.
               Now that I know autoheight, I'll have n be a multiplier for the number of screens to display, each autolines - 5 in size.  N will remain as is.
28 Jul 21 -- I'm removing truncStr and will use fixedStringLen instead.
 3 Feb 22 -- Porting simpler code from dsrt and ds to here.  And reversed -x and -exclude options.  Now -x means input exclude regex on command line.
               And adding a column number param.
 4 Feb 22 -- Added c2 and c3 flags to set 2 and 3 column modes.
 9 Feb 22 -- Fixed bug on sorting line, sorting the wrong file.
15 Feb 22 -- Replaced testFlag w/ verboseFlag, finally.
16 Feb 22 -- Time to remove the upper case flags that I don't use.
25 Apr 22 -- Added the -1 flag, and it's halfFlag variable.  For displaying half the number of lines the screen allows.
15 Oct 22 -- Added max flags to undo the effect of environment var dsrt=20
               I removed the filter flag from this code when I wrote it.
21 Oct 22 -- Removed unused variable as caught by golangci-lint, and incorrect use of format verb.
11 Nov 22 -- Will show environment variables on startup message, if they're not blank.
21 Nov 22 -- Use of dirAlisesMap was not correct.  It is not used as a param to a func, so I removed that.
16 Jan 23 -- Added smart case
26 Feb 23 -- Fixed bug that effects opening symlinked directories on linux.
27 Aug 23 -- I want to make the -t switch report how many total matches there are to the RegExp.  Instead of how many total files and bytes in the directory.
               I don't need to know how many total bytes there are in the matches to the RegExp.  So I have to capture the len of the slice of matches.
               I may just always show that, as it seems it would be easy and only 1 line.  I'll place that line at the bottom.
               I removed the -t ShowGrandTotal flag as I removed the code that calculated it quite a while ago.
28 Aug 23 -- Added the all flag, currently equivalent to indicating 50 screens.  Mostly copied the code from dsrt.go.
20 Feb 24 -- Changed a message to make it clear that this sorts on mod date.  And nscreen correctly handles numOfCols.
 4 May 24 -- Adding concurrent code from fdsrt.
 3 Jun 24 -- Removed commented out code and edited a few comments.
 5 Jan 25 -- There's a bug in how the dsrt environ variable is processed.  It sets the variable that's now interpretted as nscreens instead of nlines (off the top of my head)
				nscreens can only be set on the command line, not by environ var.  The environ var is used to set lines to display on screen.
				I decided to separate the environ variables, so this now uses rex instead of dsrt as the environ var name it uses to set its defaults.
 6 Jan 25 -- Today's my birthday.  But that's not important now.  If I set nlines via the environment, and then use the halfFlag, the base amount is what dsrt is, not the full screen.
				I want the base amount to be the full screen.  I have to think about this for a bit.
				I decided to use the maxflag system, and set maxflag if halfflag or if nscreens > 1 or if allflag.
 8 Jan 25 -- Using maxFlag is not a good idea, as it just prevents halfFlag from ever working.  See top comments in dsrt.go.  Tagged as rex-v1.0
17 Feb 25 -- Adding pflag.  I don't think I need viper yet.  I added pflag by naming its import path to flag.
----------------------------------------------------------------------------------------------------
17 Feb 25 -- Now called rexv.go, as I've added viper.
*/

const LastAltered = "Feb 17, 2025"

type dirAliasMapType map[string]string

//type DsrtParamType struct {  replaced by viper
//	paramNum, w                                                               int
//	reverseflag, sizeflag, dirlistflag, filenamelistflag, totalFlag, halfFlag bool
//}

type colorizedStr struct {
	color ct.Color
	str   string
}

const defaultHeight = 40
const maxWidth = 300
const minWidth = 90
const min2Width = 170
const min3Width = 170
const configShortName = "rexv"

const multiplier = 10 // used for the worker pool pattern in MyReadDir
const fetch = 1000    // used for the concurrency pattern in MyReadDir
var numWorkers = runtime.NumCPU() * multiplier

var excludeRegex *regexp.Regexp

var dirListFlag, longFileSizeListFlag, filenameToBeListedFlag, showGrandTotal, verboseFlag, noExtensionFlag, excludeFlag, veryVerboseFlag, halfFlag bool

var maxDimFlag, fastFlag bool
var sizeTotal, grandTotal int64
var numOfLines, grandTotalCount int
var smartCase bool

// allScreens is the number of screens to be used for the allFlag switch.  This can be set by the environ var dsrt.
var allScreens = 50

// this is to be equivalent to allScreens screens, by default same as n=50.
var allFlag bool

func main() {
	//var dsrtParam DsrtParamType  replaced by viper
	var fileInfos []os.FileInfo
	var err error
	var autoHeight, autoWidth int
	var excludeRegexPattern string
	var numOfCols int

	// environment variable processing.  If present, these will be the defaults.
	//dsrtEnviron := os.Getenv("rex")
	//dswEnviron := os.Getenv("dsw")
	//dsrtParam = ProcessEnvironString(dsrtEnviron, dswEnviron) // This is a function below.  Nevermind, replaced by viper.

	autoDefaults := term.IsTerminal(int(os.Stdout.Fd()))
	winFlag := runtime.GOOS == "windows"

	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		autoDefaults = false
		fmt.Fprintln(os.Stderr, " From term.Getsize:", err)
		autoWidth = minWidth
		autoHeight = defaultHeight
	}
	autoHeight -= 7 // an empirically determined fudge factor.

	//if autoHeight > 0 {
	//	numOfLines = autoHeight - 7
	//} else {
	//	numOfLines = defaultHeight
	//}

	sepstring := string(filepath.Separator)
	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.  Function avail as of go 1.12.
	if err != nil {
		HomeDirStr = ""
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr.")
	}
	HomeDirStr = HomeDirStr + sepstring

	// flag definitions and processing.  Now using pflag package
	var revflag = flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr

	var nscreens = flag.IntP("nscreens", "n", 1, "number of screens to display") // Ptr
	var NLines int
	flag.IntVarP(&NLines, "nlines", "N", 0, "number of lines to display") // Value
	viper.SetDefault("NLines", autoHeight)

	//var helpflag = flag.Bool("h", false, "print help message") // pointer
	//var HelpFlag bool
	//flag.BoolVar(&HelpFlag, "help", false, "print help message")

	var sizeflag = flag.BoolP("size", "s", false, "sort by size instead of by mod date") // pointer

	flag.BoolVarP(&dirListFlag, "dirlist", "d", false, "include directories in the output listing")

	var FilenameListFlag bool
	flag.BoolVarP(&FilenameListFlag, "onlydir", "D", false, "Directories only in the output listing")

	var TotalFlag = flag.BoolP("total", "t", false, "include grand total of directory") // Removed 8/27/23, added back 5/4/24

	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "enter a verbose (testing) mode to println more variables")

	var longflag = flag.BoolP("long", "l", false, "long file size format.") // Ptr

	//flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex to be entered after prompt")  Never used this way anyways
	flag.StringVarP(&excludeRegexPattern, "exclude", "x", "", "regex entered on command line to be excluded from output.")

	var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	var w int
	flag.IntVarP(&w, "width", "w", 0, " width of full displayed screen.")
	viper.SetDefault("width", autoWidth)

	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose flag for noisy tests.")

	flag.IntVarP(&numOfCols, "cols", "c", 1, "Columns in the output.")
	flag.BoolVarP(&halfFlag, "half", "1", false, "display 1/2 of the screen.")

	mFlag := flag.BoolP("max", "m", false, "Set maximum height, usually 50 lines")

	c2 := flag.BoolP("c2", "2", false, "Flag to set 2 column display mode.")
	c3 := flag.BoolP("c3", "3", false, "Flag to set 3 column display mode.")

	flag.BoolVarP(&allFlag, "all", "a", false, "Equivalent to 50 screens by default.  Intended to be used w/ the scroll back buffer.")

	flag.BoolVar(&fastFlag, "fast", false, "Fast debugging flag.  Used (so far) in MyReadDir.")

	flag.Parse()

	// viper stuff
	err = viper.BindPFlags(flag.CommandLine) // this statement passes control of all the flags to viper from the pflag package.  Remember, verbose and veryverbose flags are not init'd yet
	if err != nil {
		ctfmt.Printf(ct.Red, winFlag, "Error binding flags is %s.  Binding is ignored.\n", err.Error())
	}

	viper.SetConfigType("yaml")
	viper.SetConfigName(configShortName) // From an online source.  This works too.  Great.
	viper.AddConfigPath(".")

	//AutomaticEnv makes Viper check if environment variables match any of the existing keys (config, default or flags). If matching env vars are found, they are loaded into Viper.
	// This seems to be working.  But I don't intend to use it much.  I like having directory specific config files.  I removed the config file from homeDir and put it in Documents on Win11.
	viper.AutomaticEnv()

	var errconfig1, errconfig2 error
	errconfig1 = viper.ReadInConfig()
	if errconfig1 != nil {
		viper.AddConfigPath(HomeDirStr)
		errconfig2 = viper.ReadInConfig()
	}

	verboseFlag = viper.GetBool("verbose")
	veryVerboseFlag = viper.GetBool("vv")
	if veryVerboseFlag { // setting very verbose will also set verbose.
		verboseFlag = true
	}

	if verboseFlag {
		if errconfig1 != nil {
			ctfmt.Printf(ct.Red, winFlag, "Error reading config file 1, from current directory  Err: %s. \n", errconfig1.Error())
		}
		if errconfig2 != nil {
			ctfmt.Printf(ct.Red, winFlag, "Error reading config file 2, from current directory  Err: %s. \n", errconfig2.Error())
		}
	}

	*mFlag = viper.GetBool("max")
	maxDimFlag = *mFlag //  I removed maxFlag for this version.

	NLines = viper.GetInt("NLines")
	numOfLines = NLines

	*nscreens = viper.GetInt("nscreens")
	allFlag = viper.GetBool("all")
	if allFlag { // if both nscreens and allScreens are used, allFlag takes precedence.
		*nscreens = allScreens // allScreens is defined above w/ a default, non-zero value of 50 as of this writing.
	}
	numOfLines *= *nscreens * numOfCols // Doesn't matter if *nscreens or numOfCols = 1 which is the default

	halfFlag = viper.GetBool("half")
	if halfFlag && !maxDimFlag { // halfFlag could be set by environment var, but overridden by use of maxDimFlag.
		numOfLines /= 2
	}

	*revflag = viper.GetBool("reverse")
	Reverse := *revflag
	Forward := !Reverse // convenience variable

	*sizeflag = viper.GetBool("size")
	SizeSort := *sizeflag
	DateSort := !SizeSort // convenience variable

	*extflag = viper.GetBool("noext")
	*extensionflag = viper.GetBool("binary")
	noExtensionFlag = *extensionflag || *extflag

	excludeRegexPattern = viper.GetString("exclude")
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
	}

	dirListFlag = viper.GetBool("dirlist")
	FilenameListFlag = viper.GetBool("onlydir")
	*TotalFlag = viper.GetBool("totals")
	*longflag = viper.GetBool("long")
	dirListFlag = dirListFlag || FilenameListFlag // if -D entered then this expression also needs to be true.
	filenameToBeListedFlag = !FilenameListFlag    // need to reverse the flag because D means suppress the output of filenames.
	longFileSizeListFlag = *longflag
	showGrandTotal = *TotalFlag // added 09/12/2018 12:32:23 PM

	if numOfCols < 1 {
		numOfCols = 1
	} else if numOfCols > 3 {
		numOfCols = 3
	}
	if numOfCols == 1 { // there are 2 ways to set numOfCols, by numOfCols flag, or the c2 or c3 flags.
		if *c2 {
			numOfCols = 2
		} else if *c3 {
			numOfCols = 3
		}
	}

	w = viper.GetInt("width")
	if w <= 0 || w > maxWidth {
		if numOfCols == 1 {
			w = autoWidth
		} else if numOfCols == 2 {
			w = min2Width
		} else { // numOfCols must be 3
			w = min3Width
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

	ctfmt.Println(ct.Magenta, winFlag, " rexv will display sorted by date or size in up to 3 columns.  Uses viper.  LastAltered ", LastAltered, ", compiled using ", runtime.Version())
	if verboseFlag {
		fmt.Printf(" width = %d\n", w)
	}

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execName)
		fmt.Println()
		fmt.Println("winFlag:", winFlag)
		fmt.Println()
		fmt.Printf(" After flag.Parse(); option switches w=%d, nscreens=%d, Nlines=%d and numofCols=%d\n", w, *nscreens, NLines, numOfCols)
		fmt.Printf(" After flag.Parse(); autowidth=%d, autoheight=%d, numOfLines=%d and numofCols=%d\n", autoWidth, autoHeight, numOfLines, numOfCols)
	}

	flag.Usage = func() {
		fmt.Printf("\n AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Printf(" Viper uses rexv.yaml as its config file name. \n")
		fmt.Println(" regex pattern [directory] -- pattern defaults to '.', directory defaults to current directory.")
		fmt.Println(" Reads from dsrt environment variable before processing commandline switches, using same syntax as dsrt.")
		fmt.Println(" Uses strings.ToLower on the regex and on the filenames it reads in to make the matchs case insensitive.")
		fmt.Println()
		fmt.Println(" Regex Perl syntax: ., \\d digit, \\D Not digit, \\w word, \\W not word")
		fmt.Println("                    * zero or more, + one or more, ? zero or one")
		fmt.Println("                    x{n,m} from n to m of x, x{n,} n or more of x ")
		fmt.Println("                    ^ at beginning of text or line.  $ at end of text or line.")
		fmt.Println(" More help on syntax by go doc regexp/syntax, on the golang.org site for regexp/syntax package.")
		fmt.Println()
		flag.PrintDefaults()
		//return flagged by staticcheck as redundant.  Interesting
	}

	inputRegExStr := ""
	workingDir, er := os.Getwd()
	if er != nil {
		fmt.Fprintf(os.Stderr, " Error from Getwd() is %v\n", er)
		os.Exit(1)
	}

	startDir := workingDir

	if flag.NArg() == 0 {
		inputRegExStr = "." // no regex entered on command line, default is everything, esp useful for testing.
		// workingDir is set above
	} else if flag.NArg() == 1 {
		inputRegExStr = flag.Arg(0)
		// workingDir is set above
	} else { // flag.NArg() >= 2 so I'll ignore any extra params.
		inputRegExStr = flag.Arg(0)
		workingDir = flag.Arg(1)

		if winFlag { // added the winflag check so don't have to scan commandline on linux, which would be wasteful.
			if strings.ContainsRune(workingDir, ':') {
				workingDir = ProcessDirectoryAliases(workingDir)
			} //else if strings.Contains(workingDir, "~") // this can only contain a ~ on Windows.	Static linter said just use the Replace func.
			workingDir = strings.Replace(workingDir, "~", HomeDirStr, 1)
		}
		f, err := os.Open(workingDir)
		if err != nil {
			ctfmt.Printf(ct.Red, winFlag, " Opening %s gave this error: %s.  Will use %s instead.\n", workingDir, err, startDir)
			workingDir = startDir // error from opening workingDir, so use orig startDir
		}
		fi, err := f.Stat()
		if err != nil {
			ctfmt.Printf(ct.Red, winFlag, " Stat(%s) gave this error: %s.  Will use %s instead.\n", workingDir, err, startDir)
			workingDir = startDir // error from opening workingDir, so use orig startDir
		}

		if !fi.Mode().IsDir() {
			ctfmt.Printf(ct.Red, winFlag, " %s is not a directory.  Will use %s instead.\n", workingDir, startDir)
			workingDir = startDir
		}

		f.Close()
	}
	if verboseFlag {
		fmt.Println("inputRegEx=", inputRegExStr, ", and workingdir =", workingDir)
	}

	smartCaseRegex := regexp.MustCompile("[A-Z]")
	smartCase = smartCaseRegex.MatchString(inputRegExStr)
	if !smartCase {
		inputRegExStr = strings.ToLower(inputRegExStr)
	}
	inputRegEx, err := regexp.Compile(inputRegExStr)
	if err != nil {
		log.Fatalln(" error from regex compile function is ", err)
		fmt.Println()
		fmt.Println()
		os.Exit(1)
	}

	// set which sort function will be in the sortfcn var
	sortfcn := func(i, j int) bool { return false }
	if SizeSort && Forward { // set the value of sortfcn so only a single line is needed to execute the sort.
		sortfcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfos[i].Size() > fileInfos[j].Size() // I want a largest first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfos[i].ModTime().After(fileInfos[j].ModTime()) // I want a newest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfos[i].Size() < fileInfos[j].Size() // I want a smallest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime()) // I want an oldest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execName)
		fmt.Println()
		fmt.Printf(" Autodefault=%v, autoheight=%d, autowidth=%d, w=%d, numlines=%d. \n", autoDefaults, autoHeight, autoWidth, w, numOfLines)
		fmt.Printf(" Dirname is %s, smartCase = %t\n", workingDir, smartCase)
		fmt.Println()
	}

	// I need to add a description of how this code works, because I forgot.
	// The entire contents of the directory is read in and then only matching files after the excluded ones are removed, are returned as the slice of file infos.
	// Then the slice of fileinfo's is sorted, and finally the file infos are colorized and displayed in columns

	t0 := time.Now()
	fileInfos = getFileInfos(workingDir, inputRegEx)
	elapsed := time.Since(t0)

	sort.Slice(fileInfos, sortfcn)
	totalMatches := len(fileInfos) // this is before the fileInfos is truncated to only what's to be output.
	cs := getColorizedStrings(fileInfos, numOfCols)

	// Output the colorized string slice
	columnWidth := w/numOfCols - 2
	for i := 0; i < len(cs); i += numOfCols {
		c0 := cs[i].color
		s0 := fixedStringLen(cs[i].str, columnWidth)
		ctfmt.Printf(c0, winFlag, "%s", s0)

		j := i + 1
		if numOfCols > 1 && j < len(cs) { // numOfCols of 2 or 3
			c1 := cs[j].color
			s1 := fixedStringLen(cs[j].str, columnWidth)
			ctfmt.Printf(c1, winFlag, "  %s", s1)
		}

		k := j + 1
		if numOfCols == 3 && k < len(cs) {
			c2 := cs[k].color
			s2 := fixedStringLen(cs[k].str, columnWidth)
			ctfmt.Printf(c2, winFlag, "  %s", s2)
		}
		fmt.Println()
	}
	fmt.Println()

	s := fmt.Sprintf("%d", sizeTotal)
	if sizeTotal > 100000 {
		s = AddCommas(s)
	}

	fmt.Printf(" Total Matches = %d, displayed file Size total = %s, took %s.", totalMatches, s, elapsed)
	fmt.Println()
	if showGrandTotal {
		fmt.Printf(" Grand total of %d files is %d\n", grandTotalCount, grandTotal)
	}
} // end main rex

//-------------------------------------------------------------------- InsertByteSlice --------------------------------

func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

//---------------------------------------------------------------------- AddCommas

func AddCommas(instr string) string {
	// var Comma []byte = []byte{','}  compiler flagged this as type not needed
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
		s1 = fmt.Sprintf("%.4g gb", f)
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

// --------------------------------------------------- fixedString ---------------------------------------

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
		fmt.Fprintln(os.Stderr, " makeStrFixed input string length is strange.  It is", len(s))
		return s
	}
} // end fixedStringLen

// ------------------------------------------------------ getFileInfos -------------------------------------------------

// getFileInfos will return a slice of FileInfos after the regexp, filter and exclude expression are processed
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
func getFileInfos(workingDir string, inputRegex *regexp.Regexp) []os.FileInfo {

	fileInfos := myReadDirWithMatch(workingDir, inputRegex) // excluding by regex, filesize or having an ext is done by MyReadDir.
	if verboseFlag {
		fmt.Printf(" Leaving getFileInfosFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfos))
	}
	if verboseFlag {
		fmt.Printf(" Entering getFileInfos.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfos))
	}

	return fileInfos
}

func myReadDirWithMatch(dir string, regex *regexp.Regexp) []os.FileInfo { // The entire change including use of []DirEntry happens here, and now concurrent code.
	// Adding concurrency in returning []os.FileInfo

	var wg sync.WaitGroup

	if verboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	// numWorkers is set globally, above.
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
					if includeThisWithRegex(de.Name(), regex) { // this optimization only calls Stat for those DirEntry's that we're keeping.
						fi, err := de.Info()
						if err != nil {
							fmt.Printf("Error getting file info for %s: %v, ignored\n", de.Name(), err)
							continue
						}
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
		// reading DirEntry's and sending the slices into the channel happens here.
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
} // myReadDir

// --------------------------------------------- includeThis ----------------------------------------------------------

func includeThisWithRegex(fn string, regex *regexp.Regexp) bool { // I removed the filter against file size, so the input param can now be a string.
	if veryVerboseFlag {
		//fmt.Printf(" includeThis.  noExtensionFlag=%t, excludeFlag=%t, filterAmt=%d \n", noExtensionFlag, excludeFlag, filterAmt)
		fmt.Printf(" includeThis.  noExtensionFlag=%t, excludeFlag=%t \n", noExtensionFlag, excludeFlag)
	}
	if noExtensionFlag && strings.ContainsRune(fn, '.') {
		return false
	}
	fnl := strings.ToLower(fn)
	if excludeFlag {
		if excludeRegex.MatchString(fnl) {
			return false
		}
	}

	if !smartCase && !regex.MatchString(fnl) {
		return false
	} else if smartCase && !regex.MatchString(fn) {
		return false
	}
	return true
}

// --------------------------------------------- getColorizedStrings --------------------------------------------------

func getColorizedStrings(fiSlice []os.FileInfo, cols int) []colorizedStr { // cols is the intended number of columns for the colorizedStr output slice.

	cs := make([]colorizedStr, 0, len(fiSlice))

	for i, f := range fiSlice {
		t := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizeStr := ""
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			sizeTotal += f.Size()
			if longFileSizeListFlag {
				sizeStr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
				if f.Size() > 100000 {
					sizeStr = AddCommas(sizeStr)
				}
				strng := fmt.Sprintf("%16s %s %s", sizeStr, t, f.Name())
				colorized := colorizedStr{color: ct.Yellow, str: strng}
				cs = append(cs, colorized)

			} else {
				var colr ct.Color
				sizeStr, colr = getMagnitudeString(f.Size())
				strng := fmt.Sprintf("%-10s %s %s", sizeStr, t, f.Name())
				colorized := colorizedStr{color: colr, str: strng}
				cs = append(cs, colorized)
			}

		} else if IsSymlink(f.Mode()) {
			s := fmt.Sprintf("%5s %s <%s>", sizeStr, t, f.Name())
			colorized := colorizedStr{color: ct.White, str: s}
			cs = append(cs, colorized)
		} else if dirListFlag && f.IsDir() {
			s := fmt.Sprintf("%5s %s (%s)", sizeStr, t, f.Name())
			colorized := colorizedStr{color: ct.White, str: s}
			cs = append(cs, colorized)
		}
		if i > numOfLines*cols {
			break
		}
	}
	return cs
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
