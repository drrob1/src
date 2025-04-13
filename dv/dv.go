// dv.go adding viper, based on fdsrt.go -- fast directory sort

package main

import (
	"errors"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/metrics"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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
------------------------------------------------------------------------------------------------------------------------------------------------------
 1 May 24 -- Now called fdsrt, for fast directory sort.  I'm going to see if I can use a worker pool here for collecting the []FileInfos to be returned.  I'm playing on Windows now.
             First I have to create the sending and receiving go routines, and then I write the code to send the data into the first channel.  This occurs all within MyReadDir, for now.
             That's done.  Now I have to check the other routines.
 2 May 24 -- I decided to debug what I have before writing more.  Just goes to show you that not everything can be sped up by concurrency.
             The coordination with a wait group and the done channel works, but is slightly slower than dsrt.  This is w/ fetch = 1000.  It's worse when fetch=100.  And maybe also
             slightly worse when fetch = 2000.  I'll leave it at 1000, and stop working on this.
             At least I got it to work.
             When run in jpg1 w/ 23000 files, and jpg2 w/ 13000 files, this rtn is slightly faster.  From 13 ms w/ dsrt, to 12 ms w/ fdsrt.
             When run on thelio and logged into the /mnt/misc dir w/ 23000 files, dsrt is ~6 sec and fdsrt is ~1 sec, so here it's much faster.  In a directory w/ only 300 files, this rtn is ~2x slower.
             So this is more complicated after all.
 3 May 24 -- I moved the IncludeThis test into the worker goroutines.  That's where it belongs.  Now the timing may be slightly faster than dsrt, here on Win11
             I'm going to continue development after all.  On windows, I'll write a handler that allows a match to occur.  This does not make sense on bash, so this will only apply to Windows.
             And glob option is removed.  It's too complex to add and I never use it.  It will stay in dsrt.go.
 7 Jun 24 -- Edited some comments and changed the final message.
 8 Jan 25 -- There's a bug in how the dsrt environ variable is processed.  It sets the variable that's now interpretted as nscreens instead of nlines (off the top of my head).
				nscreens can only be set on the command line, not by environ var.  The environ var is used to set lines to display on screen.
----------------------------------------------------------------------
*/
/*
16 Jan 25 -- Now called dv, for dsrt viper.
             Will use viper to replace all my own logic.  Viper handles priorities among command line, environment, config file and default value.
16 Jan 25 -- Looks like it's working, including the dv.yaml config file, environ flags that match option names, and command line option.
18 Jan 25 -- It will now check the current directory for a config file.  If not found, it will check the home directory.  This allows different directories to have different defaults.
19 Jan 25 -- Added -w as a single letter meaning --vv.  The use of -w comes from a play on its name.
20 Jan 25 -- Added timing info for startup code.
22 Jan 25 -- Added runtime.Compiler to the startup message.
17 Feb 25 -- Removed code regarding the fullConfigFile for the viper config file.
 9 Apr 25 -- Modified help message.
13 Apr 25 -- Added metrics as covered by Mastering Go, 4th ed.
*/

const LastAltered = "13 Apr 2025"

// Outline
// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, thru use of OS specific code.  On Windows it will get a pattern from the command line.
// but does not sort the slice before returning it, due to difficulty in passing the sort function.  Instead, the slice is sorted in main, on or about line #503.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

type dirAliasMapType map[string]string

//type DsrtParamType struct { replaced by viper
//	paramNum                                                                              int // set by dsrt environ var.
//	reverseFlag, sizeFlag, dirListFlag, filenameListFlag, totalFlag, filterflag, halfFlag bool
//}

const defaultHeight = 40
const minWidth = 90
const multiplier = 10 // used for the worker pool pattern in MyReadDir
const fetch = 1000    // used for the concurrency pattern in MyReadDir
var numWorkers = runtime.NumCPU() * multiplier

const configFilename = "dv.yaml"
const configShortName = "dv"

var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, verboseFlag bool
var filterFlag, globFlag, veryVerboseFlag, halfFlag, maxDimFlag, fastFlag bool
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
	var userPtr *user.User // from os/user
	var err error
	var autoWidth, autoHeight int
	var excludeRegexPattern string
	var fileInfos []os.FileInfo

	var uid, gid int
	var systemStr string

	t1 := time.Now()
	winflag := runtime.GOOS == "windows" // this is needed because I use it in the color statements, so the colors are bolded only on windows.
	ctfmt.Printf(ct.Magenta, winflag, "dv will display Directory SoRTed by date or size, using concurrent code.  LastAltered %s, compiled with %s version %s\n",
		LastAltered, runtime.Compiler, runtime.Version())
	homeDir, err := os.UserHomeDir()
	if err != nil {
		ctfmt.Printf(ct.Red, winflag, "Error getting user home directory is %s\n", err.Error())
		return
	}
	//fullConfigFileName := filepath.Join(homeDir, configFilename)
	//_ = fullConfigFileName // this is a kludge so this var is marked as needed when it's not

	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}
	autoHeight -= 7 // an empirically determined fudge factor.

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
	pflag.Usage = func() {
		fmt.Printf(" %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Printf(" Usage: %s [directory]glob\n", os.Args[0])
		fmt.Printf(" AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Printf(" Config file is dv.yaml.\n")
		fmt.Printf(" Reads from diraliases environment variable if needed on Windows.\n")
		pflag.PrintDefaults()
	}

	revflag := pflag.BoolP("reverse", "r", false, "reverse the sort, ie, oldest or smallest is first")          // Ptr
	nscreens := pflag.IntP("nscreens", "n", 1, "number of screens to display, ie, a multiplier for numOfLines") // Ptr

	var NLines int
	pflag.IntVarP(&NLines, "NLines", "N", numOfLines, "number of lines to display, and takes priority over the auto settings.") // Value
	viper.SetDefault("NLines", autoHeight)

	sizeflag := pflag.BoolP("size", "s", false, "sort by size instead of by mod date")             // pointer
	DirListFlag := pflag.BoolP("dirlist", "d", false, "include directories in the output listing") // pointer

	var FilenameListFlag bool
	pflag.BoolVarP(&FilenameListFlag, "onlydir", "D", false, "Directories only in the output listing")

	TotalFlag := pflag.BoolP("totals", "t", false, "include grand total of directory, makes most sense when no pattern is given on command line.")

	pflag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode, which is same as test mode.")
	pflag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "Very verbose option for when I really want it.") // short for vv is w, kind of a play on letter names.

	longflag := pflag.BoolP("long", "l", false, "long file size format.") // Ptr

	extflag := pflag.BoolP("noext", "e", false, "only print if there is no extension, like a binary file")
	extensionflag := pflag.BoolP("binary", "b", false, "only print if there is no extension, like a binary file")

	pflag.StringVarP(&excludeRegexPattern, "exclude", "x", "", "regex to be excluded from output.")

	pflag.StringVar(&filterStr, "filterstr", "", "individual size filter value below which listing is suppressed.  k, m, g are currently implemented.")
	pflag.BoolVarP(&filterFlag, "filter", "f", false, "filter value to suppress listing individual size below 1 MB.")
	noFilterFlag := pflag.BoolP("nofilter", "F", false, "Flag to undo an environment var with f set.")

	pflag.BoolVarP(&globFlag, "glob", "g", false, "Use glob function on Windows.")

	pflag.BoolVarP(&halfFlag, "half", "1", false, "display 1/2 of the screen.")

	mFlag := pflag.BoolP("max", "m", false, "Set maximum height, usually 50 lines")

	pflag.BoolVarP(&allFlag, "all", "a", false, "Equivalent to 50 screens by default.  Intended to be used w/ the scroll back buffer.")
	pflag.IntVar(&allScreens, "allscreens", allScreens, "Number of screens to display when all option is selected.")

	pflag.BoolVar(&fastFlag, "fast", false, "Fast debugging flag.  Used (so far) in MyReadDir.")

	pflag.Parse()

	// viper stuff
	err = viper.BindPFlags(pflag.CommandLine) // this statement passes control of all the flags to viper from the pflag package.  Remember, verbose and veryverbose flags are not init'd yet
	if err != nil {
		ctfmt.Printf(ct.Red, winflag, "Error binding flags is %s.  Binding is ignored.\n", err.Error())
	}

	viper.SetConfigType("yaml")
	//viper.SetConfigFile(fullConfigFileName) // This works but I'm experimenting.
	viper.SetConfigName(configShortName) // From an online source.  This works too.  Great.
	viper.AddConfigPath(".")

	//AutomaticEnv makes Viper check if environment variables match any of the existing keys (config, default or flags). If matching env vars are found, they are loaded into Viper.
	// This seems to be working.  But I don't intend to use it much.  I like having directory specific config files.  I removed the config file from homeDir and put it in Documents on Win11.
	viper.AutomaticEnv()

	var errconfig1, errconfig2 error
	errconfig1 = viper.ReadInConfig()
	if errconfig1 != nil {
		viper.AddConfigPath(homeDir)
		errconfig2 = viper.ReadInConfig()
	}

	verboseFlag = viper.GetBool("verbose")
	veryVerboseFlag = viper.GetBool("vv")
	if veryVerboseFlag { // setting veryVerbose flag will also set verbose flag
		verboseFlag = true
	}
	if verboseFlag {
		if errconfig1 != nil {
			ctfmt.Printf(ct.Red, winflag, "Error reading config file 1, from current directory  Err: %s. \n", errconfig1.Error())
		}
		if errconfig2 != nil {
			ctfmt.Printf(ct.Red, winflag, "Error reading config file 2, from current directory  Err: %s. \n", errconfig2.Error())
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
	numOfLines *= *nscreens // Doesn't matter if *nscreens = 1 which is the default

	halfFlag = viper.GetBool("half")
	if halfFlag && !maxDimFlag { // halfFlag could be set by environment var, but overridden by use of maxDimFlag.
		numOfLines /= 2
	}

	filterStr = viper.GetString("filterstr")
	filterFlag = viper.GetBool("filter")
	*noFilterFlag = viper.GetBool("nofilter") // the noFilterFlag takes priority.
	if *noFilterFlag {
		filterFlag = false
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

	globFlag = viper.GetBool("glob")
	if globFlag {
		fmt.Printf(" Glob flag has been removed from fdsrt.  This flag is now ignored.\n")
		globFlag = false
	}

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

	*DirListFlag = viper.GetBool("dirlist")
	FilenameListFlag = viper.GetBool("onlydir")
	*TotalFlag = viper.GetBool("totals")
	*longflag = viper.GetBool("long")
	dirList = *DirListFlag || FilenameListFlag // if -D entered then this expression also needs to be true.
	filenameToBeListedFlag = !FilenameListFlag // need to reverse the flag because D means suppress the output of filenames.
	longFileSizeListFlag = *longflag
	showGrandTotal = *TotalFlag // added 09/12/2018 12:32:23 PM

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
		fmt.Printf(" autoheight=%d, autowidth=%d, excludeFlag=%t, halfFlag=%t. \n", autoHeight, autoWidth, excludeFlag, halfFlag)
		fmt.Printf(" NLines=%d, Reverse=%t, ncreens=%d, sizeflag=%t, DirListFlag=%t, FilenameListFlag=%t, TotalFlag=%t, longflag=%t \n",
			NLines, Reverse, *nscreens, *sizeflag, *DirListFlag, FilenameListFlag, *TotalFlag, *longflag)
		fmt.Printf(" extflag=%t, extensionflag=%t, filterFlag=%t, noFilterFlag=%t, globFlag=%t, mFlag=%t, allFlag=%t \n",
			*extflag, *extensionflag, filterFlag, *noFilterFlag, globFlag, *mFlag, allFlag)
	}

	ctfmt.Printf(ct.Green, winflag, "%50s Finished startup code, which took %s\n", " ", time.Since(t1))
	// from here down, the code is essentially the same as before.  Config section is finished.

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
		fmt.Println(" *** Here I am at or about line 467 ***")
		fmt.Println(" FilterFlag =", filterFlag, ".  filterStr =", filterStr, ". filterAmt =", filterAmt, "excludeFlag =", excludeFlag)
		fmt.Printf(" nscreens=%d, numLines=%d, pflag.NArgs=%d, dirList=%t, Filenametobelistedflag=%t, longfilesizelistflag=%t, showgrandtotal=%t\n",
			*nscreens, numLines, pflag.NArg(), dirList, filenameToBeListedFlag, longFileSizeListFlag, showGrandTotal)
	}

	t0 := time.Now()

	fileInfos = getFileInfosFromCommandLine()
	if verboseFlag {
		fmt.Printf(" After call to getFileInfosFromCommandLine.  pflag.NArg=%d, len(fileinfos)=%d, numOfLines=%d\n", pflag.NArg(), len(fileInfos), numOfLines)
	}
	if len(fileInfos) > 1 {
		sort.Slice(fileInfos, sortfcn) // must be sorted here for sortfcn to work correctly, because the slice name it uses must be correct.  Better if that name is not global.
	}

	elapsed := time.Since(t0)

	displayFileInfos(fileInfos)

	s := fmt.Sprintf("%d", sizeTotal)
	if sizeTotal > 100000 {
		s = AddCommas(s)
	}
	s0 := fmt.Sprintf("%d", grandTotal)
	if grandTotal > 100000 {
		s0 = AddCommas(s0)
	}
	fmt.Printf(" %s: Elapsed time = %s, File Size total = %s, len(fileInfos)=%d", os.Args[0], elapsed, s, len(fileInfos))
	if showGrandTotal {
		s1, color := getMagnitudeString(grandTotal)
		ctfmt.Println(color, true, ", Directory grand total is", s0, "or approx", s1, "in", grandTotalCount, "files.")
	} else {
		fmt.Println(".")
	}

	// metrics
	metricSlice := make([]metrics.Sample, 2) // I want to get total cpu time, and user cpu time
	metricSlice[0] = metrics.Sample{Name: "/cpu/classes/total:cpu-seconds"}
	metricSlice[1] = metrics.Sample{Name: "/cpu/classes/user:cpu-seconds"}
	metrics.Read(metricSlice)
	if metricSlice[0].Value.Kind() == metrics.KindBad {
		fmt.Printf("metric %q no longer supported\n", metricSlice[0].Name)
	}
	if metricSlice[1].Value.Kind() == metrics.KindBad {
		fmt.Printf("metric %q no longer supported\n", metricSlice[1].Name)
	}
	totalCPUseconds := metricSlice[0].Value.Float64()
	userCPUseconds := metricSlice[1].Value.Float64()
	fmt.Printf(" User CPU seconds: %.4f;  total CPU Seconds: %.4f.\n", userCPUseconds, totalCPUseconds)

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

// MakeSubst -- substitutes characters.  Written before I knew about strings.replacer.
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

// GetDirectoryAliases -- used on Windows systems
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

func myReadDir(dir string) []os.FileInfo { // The entire change including use of []DirEntry happens here.  Concurrent code here is what makes this fdsrt.
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

	// reading from deChan to get the slices of DirEntry's
	for range numWorkers {
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

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
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
	// This routine adds a call to filepath.Match

	var wg sync.WaitGroup

	if verboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	deChan := make(chan []os.DirEntry, numWorkers) // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fiChan := make(chan os.FileInfo, numWorkers)   // of individual file infos to be collected and returned to the caller of this routine.
	doneChan := make(chan bool)                    // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fiSlice := make([]os.FileInfo, 0, fetch*multiplier*multiplier)
	wg.Add(numWorkers)

	// reading from deChan to get the slices of DirEntry's
	for range numWorkers {
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

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
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
