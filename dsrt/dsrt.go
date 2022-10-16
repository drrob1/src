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
	"path/filepath"
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
*/

const LastAltered = "15 Oct 2022"

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, thru use of OS specific code.  On Windows it will get a pattern from the command line.
// but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

type dirAliasMapType map[string]string

type DsrtParamType struct {
	numlines                                                        int
	reverseflag, sizeflag, dirlistflag, filenamelistflag, totalflag bool
}

const defaultHeight = 40
const minWidth = 90

var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, verboseFlag bool
var filterFlag, globFlag, veryVerboseFlag, halfFlag, maxDimFlag bool
var filterAmt, numLines, numOfLines, grandTotalCount int
var sizeTotal, grandTotal int64
var filterStr string
var excludeRegex *regexp.Regexp
var directoryAliasesMap dirAliasMapType

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

	// environment variable processing.  If present, these will be the defaults.  Processed before the flags so the flags will override these, if provided on the command line.
	dsrtParam = ProcessEnvironString() // This is a function below.

	winflag := runtime.GOOS == "windows" // this is needed because I use it in the color statements, so the colors are bolded only on windows.
	ctfmt.Print(ct.Magenta, winflag, "dsrt will display Directory SoRTed by date or size.  LastAltered ", LastAltered, ", compiled using ",
		runtime.Version(), ".")
	fmt.Println()

	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}
	/*
		if !autoDefaults {
			if winflag {
				comspec, ok := os.LookupEnv("ComSpec")
				if ok {
					//bytesbuf := bytes.NewBuffer([]byte{}) // from Go Standard Library Cookbook by Radomir Sohlich (C) 2018 Packtpub
					bytesbuf := bytes.NewBuffer(make([]byte, 0, 200))
					tcc := exec.Command(comspec, "-C", "echo", "%_columns")
					tcc.Stdout = bytesbuf
					tcc.Run()
					colstr := bytesbuf.String()
					lines := strings.Split(colstr, "\n")
					trimmedLine := strings.TrimSpace(lines[1]) // 2nd line of the output is what I want trimmed
					autoWidth, err = strconv.Atoi(trimmedLine)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Error from cols conversion is", err, "Value ignored.")
					}

					bytesbuf.Reset()
					tcc = exec.Command(comspec, "-C", "echo", "%_rows")
					tcc.Stdout = bytesbuf
					tcc.Run()
					rowstr := bytesbuf.String()
					lines = strings.Split(rowstr, "\n")
					trimmedLine = strings.TrimSpace(lines[1]) // 2nd line of the output is what I need trimmed
					autoHeight, err = strconv.Atoi(trimmedLine)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Error from rows conversion is", err, "Value ignored.")
					}
				} else {
					fmt.Fprintln(os.Stderr, "comspec expected but not found.  Using environment params settings only.")
				}
			} else {
				fmt.Fprintln(os.Stderr, "Expected a windows computer, but winflag is false.  WTF?")
			}
		}
	*/

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + string(filepath.Separator)

	if runtime.GOARCH == "amd64" {
		uid = os.Getuid() // int
		gid = os.Getgid() // int
		userPtr, err = user.Current()
		if err != nil {
			fmt.Println(" user.Current error is ", err, "Exiting.")
			os.Exit(1)
		}
		// HomeDirStr = userptr.HomeDir + sepstring
	}

	// flag definitions and processing
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " dsrt environ values are: numlines=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelistflag=%t, totalflag=%t \n",
			dsrtParam.numlines, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag, dsrtParam.totalflag)

		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	revflag := flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
	var RevFlag bool                                                                      // this will always be false.  I can leave it this way for now.
	// flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nscreens = flag.Int("n", 1, "number of screens to display, ie, a multiplier for numOfLines") // Ptr
	var NLines int
	flag.IntVar(&NLines, "N", numOfLines, "number of lines to display, and takes priority over the auto settings.") // Value

	var sizeflag = flag.Bool("s", false, "sort by size instead of by date") // pointer
	var SizeFlag bool                                                       // will always be false.  I can leave it this way, for now.
	//flag.BoolVar(&SizeFlag, "S", false, "sort by size instead of by date")

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

	flag.BoolVar(&globFlag, "g", false, "Use glob function on Windows.")

	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose option for when I really want it.")
	flag.BoolVar(&halfFlag, "1", false, "display 1/2 of the screen.")

	mFlag := flag.Bool("m", false, "Set maximum height, usually 50 lines")
	maxFlag := flag.Bool("max", false, "Set max height, usually 50 lines, alternative flag")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerbose flag will also set verbose flag, ie testFlag.
		verboseFlag = true
	}

	maxDimFlag = *mFlag || *maxFlag // either m or max options will set this flag and suppress use of halfFlag.

	if NLines > 0 { // priority
		numOfLines = NLines
	} else if dsrtParam.numlines > 0 && !maxDimFlag { // then check this, but only if maxDimFlag is not set.
		numOfLines = dsrtParam.numlines
	} else if autoHeight > 0 { // finally use autoHeight.
		numOfLines = autoHeight - 7
	} else { // intended if autoHeight fails, like of the output is being redirected.
		numOfLines = defaultHeight
	}

	numOfLines *= *nscreens // Doesn't matter if *nscreens = 1

	if halfFlag {
		numOfLines /= 2
	}

	if verboseFlag {
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println()
		if runtime.GOARCH == "amd64" {
			fmt.Printf("uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
				uid, gid, systemStr, userPtr.Uid, userPtr.Gid, userPtr.Username, userPtr.Name, userPtr.HomeDir)
			fmt.Printf(" dsrtparam numlines=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelist=%t, totalflag=%t\n",
				dsrtParam.numlines, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag,
				dsrtParam.totalflag)
			fmt.Printf(" autoheight=%d, autowidth=%d, excludeFlag=%t. \n", autoHeight, autoWidth, excludeFlag)

		}
	}

	Reverse := *revflag || RevFlag || dsrtParam.reverseflag
	Forward := !Reverse // convenience variable

	SizeSort := *sizeflag || SizeFlag || dsrtParam.sizeflag
	DateSort := !SizeSort // convenience variable

	/*
	       if NLines > 0 && numOfLines == 0 { // then the -N option flag.  dsrtParam is tested above.
	   		numOfLines = NLines
	   	} else if autoHeight > 0 && numOfLines == 0 { // then autoheight
	   		numOfLines = autoHeight - 7
	   	} else { // finally the default height
	   		numOfLines = defaultHeight
	   	}
	*/
	if verboseFlag {
		fmt.Printf(" dsrtParam.numlines=%d, NLines=%d, autoheight=%d, defaultHeight=%d, and finally numOfLines = %d  \n",
			dsrtParam.numlines, NLines, autoHeight, defaultHeight, numOfLines)
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

func GetIDname(uidStr string) string {
	if len(uidStr) == 0 {
		return ""
	}
	ptrToUser, err := user.LookupId(uidStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	idname := ptrToUser.Username
	return idname
} // GetIDname

/*
// ------------------------------- GetEnviron ------------------------------------------------
func GetEnviron() DsrtParamType { // first solution to my environ var need.  Obsolete now but not gone.
	var dsrtparam DsrtParamType

	EnvironSlice := os.Environ()

	for _, e := range EnvironSlice {
		if strings.HasPrefix(e, "dsrt") {
			dsrtslice := strings.SplitAfter(e, "=")
			indiv := strings.Split(dsrtslice[1], "") // all characters after dsrt=
			for j, str := range indiv {
				s := str[0]
				if s == 'r' || s == 'R' {
					dsrtparam.reverseflag = true
				} else if s == 's' || s == 'S' {
					dsrtparam.sizeflag = true
				} else if s == 'd' {
					dsrtparam.dirlistflag = true
				} else if s == 'D' {
					dsrtparam.filenamelistflag = true
				} else if unicode.IsDigit(rune(s)) {
					dsrtparam.numlines = int(s) - int('0')
					if j+1 < len(indiv) && unicode.IsDigit(rune(indiv[j+1][0])) {
						dsrtparam.numlines = 10*dsrtparam.numlines + int(indiv[j+1][0]) - int('0')
						break // if have a 2 digit number, it ends processing of the indiv string
					}
				}
			}
		}
	}
	return dsrtparam
} // GetEnviron

*/

// ------------------------------------ ProcessEnvironString ---------------------------------------

func ProcessEnvironString() DsrtParamType { // use system utils when can because they tend to be faster
	var dsrtparam DsrtParamType

	s, ok := os.LookupEnv("dsrt")

	if !ok {
		return dsrtparam
	} // empty dsrtparam is returned

	indiv := strings.Split(s, "")

	for j, str := range indiv {
		s := str[0]
		if s == 'r' || s == 'R' {
			dsrtparam.reverseflag = true
		} else if s == 's' || s == 'S' {
			dsrtparam.sizeflag = true
		} else if s == 'd' {
			dsrtparam.dirlistflag = true
		} else if s == 'D' {
			dsrtparam.filenamelistflag = true
		} else if s == 't' { // added 09/12/2018 12:26:01 PM
			dsrtparam.totalflag = true // for the grand total operation
		} else if s == 'f' {
			filterFlag = true
		} else if unicode.IsDigit(rune(s)) {
			dsrtparam.numlines = int(s) - int('0')
			if j+1 < len(indiv) && unicode.IsDigit(rune(indiv[j+1][0])) {
				dsrtparam.numlines = 10*dsrtparam.numlines + int(indiv[j+1][0]) - int('0')
				break // if have a 2 digit number, it ends processing of the indiv string
			}
		}
	}
	return dsrtparam
}

//------------------------------ GetDirectoryAliases ----------------------------------------
func getDirectoryAliases() dirAliasMapType { // Env variable is diraliases.
	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = MakeSubst(s, '_', ' ') // substitute the underscore, _, or a space
	directoryAliasesMap = make(dirAliasMapType, 10)

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

func ProcessDirectoryAliases(aliasesMap dirAliasMapType, cmdline string) string {
	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap = getDirectoryAliases()
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

/*
func MyReadDir(dir string) []os.FileInfo {
	dirname, err := os.Open(dir)
	//	dirname, err := os.OpenFile(dir, os.O_RDONLY,0777)
	if err != nil {
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return nil
	}

	fileInfs := make([]os.FileInfo, 0, len(names))
	for _, s := range names {
		path := dir + string(os.PathSeparator) + s
		fi, err := os.Lstat(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, " Error from os.Lstat ", err)
			continue
		}
		if includeThis(fi) {
			fileInfs = append(fileInfs, fi)
		}
		if fi.Mode().IsRegular() && showGrandTotal {
			grandTotal += fi.Size()
			grandTotalCount++
		}
	}
	return fileInfs
} // MyReadDir
*/
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

/*
func getMagnitudeString(j int64) string {

	var s1 string
	var i int64
	switch {
	case j > 1_000_000_000_000: // 1 trillion, or TB
		i = j / 1000000000000               // I'm forcing an integer division.
		if j%1000000000000 > 500000000000 { // rounding up
			i++
		}
		s1 = fmt.Sprintf("%3d TB", i)
	case j > 1_000_000_000: // 1 billion, or GB
		i = j / 1000000000
		if j%1000000000 > 500000000 { // rounding up
			i++
		}
		s1 = fmt.Sprintf("%6d GB", i)
	case j > 1_000_000: // 1 million, or MB
		i = j / 1000000
		if j%1000000 > 500000 {
			i++
		}
		s1 = fmt.Sprintf("%9d MB", i)
	case j > 1000: // KB
		i = j / 1000
		if j%1000 > 500 {
			i++
		}
		s1 = fmt.Sprintf("%12d kb", i)
	default:
		s1 = fmt.Sprintf("%3d bytes", j)
	}
	return s1
}
*/

/*
 {{{
package strings
func Contains
func Contains(s, substr string) bool
Contains reports whether substr is within s.



func ContainsAny
func ContainsAny(s, chars string) bool
ContainsAny reports whether any Unicode code points in chars are within s.



func ContainsRune
func ContainsRune(s string, r rune) bool
ContainsRune reports whether the Unicode code point r is within s.

func Count
func Count(s, substr string) int
Count counts the number of non-overlapping instances of substr in s. If substr is an empty string, Count returns 1 + the number of Unicode code points in s.

func Fields
func Fields(s string) []string
Fields splits the string s around each instance of one or more consecutive white space characters, as defined by unicode.IsSpace, returning a slice of substrings of s or an empty slice if s contains only white space.

package path
func Match

func Match(pattern, name string) (matched bool, err error)

Match reports whether name matches the shell file name pattern.  The pattern syntax is:

pattern:
	{ term }
term:
	'*'         matches any sequence of non-/ characters
	'?'         matches any single non-/ character
	'[' [ '^' ] { character-range } ']'
	            character class (must be non-empty)
	c           matches character c (c != '*', '?', '\\', '[')
	'\\' c      matches character c

character-range:
	c           matches character c (c != '\\', '-', ']')
	'\\' c      matches character c
	lo '-' hi   matches character c for lo <= c <= hi

Match requires pattern to match all of name, not just a substring.  The only possible returned error is ErrBadPattern, when pattern is malformed.


package os
type FileInfo

type FileInfo interface {
        Name() string       // base name of the file
        Size() int64        // length in bytes for regular files; system-dependent for others
        Mode() FileMode     // file mode bits
        ModTime() time.Time // modification time
        IsDir() bool        // abbreviation for Mode().IsDir()
        Sys() interface{}   // underlying data source (can return nil)
}

A FileInfo describes a file and is returned by Stat and Lstat.

func Lstat

func Lstat(name string) (FileInfo, error)

Lstat returns a FileInfo describing the named file.  If the file is a symbolic link, the returned FileInfo describes the symbolic link.  Lstat makes no attempt to follow the link.
If there is an error, it will be of type *PathError.

func Stat

func Stat(name string) (FileInfo, error)

Stat returns a FileInfo describing the named file.  If there is an error, it will be of type *PathError.


The insight I had with my append troubles that the 1 slice entries were empty, is that when I used append, it would do just that to the end of the slice, and ignore the empty slices.
I needed to make the slice as empty for this to work.  So I am directly assigning the DirEntries slice, and appending the FileNames slice, to make sure that these both are doing what I want.
This code is now doing exactly what I want.  I guess there is no substitute for playing with myself.  Wait, that didn't come out right.  Or did it.


package os/user
type User struct {
  Uid string
  Gid string
  Username string // login name
  Name string     // full or display name.  It may be blank.
  HomeDir string
}

package os
func Getenv
func Getenv(key string) string
Getenv retrieves the value of the environment variable named by the key.  It returns the value, which will be empty if the variable is not present.  To distinguish between an empty value and an unset
value, use LookupEnv.

Example
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("%s lives in %s.\n", os.Getenv("USER"), os.Getenv("HOME"))

}

func Getwd
func Getwd() (dir string, err error)
Getwd returns a rooted path name corresponding to the current directory.  If the current directory can be reached via multiple paths (due to symbolic links), Getwd may return any one of them.

func Environ
func Environ() []string
Environ returns a copy of strings representing the environment, in the form "key=value".


func LookupEnv
func LookupEnv(key string) (string, bool)
LookupEnv retrieves the value of the environment variable named by the key.  If the variable is present in the environment the value (which may be empty) is returned and the boolean is true.  Otherwise the returned value will be empty and the boolean will be false.

Example
package main

import (
	"fmt"
	"os"
)

func main() {
	show := func(key string) {
		val, ok := os.LookupEnv(key)
		if !ok {
			fmt.Printf("%s not set\n", key)
		} else {
			fmt.Printf("%s=%s\n", key, val)
		}
	}

	show("USER")
	show("GOPATH")

}

}}}
*/
