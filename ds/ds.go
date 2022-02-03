// ds.go -- directory sort output in a single column

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

const LastAltered = "3 Feb 2022"

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
*/

type dirAliasMapType map[string]string

type DsrtParamType struct {
	numlines, w                                                     int
	reverseflag, sizeflag, dirlistflag, filenamelistflag, totalflag bool
}

type colorizedStr struct {
	color ct.Color
	str   string
}

const defaultHeight = 40
const minWidth = 90
const maxWidth = 300

var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, testFlag bool
var globFlag, veryVerboseFlag bool
var filterAmt, numLines, numOfLines, grandTotalCount int
var sizeTotal, grandTotal int64
var filterStr string
var excludeRegex *regexp.Regexp
var directoryAliasesMap dirAliasMapType
var autoWidth, autoHeight int

func main() {
	var dsrtParam DsrtParamType
	var userptr *user.User // from os/user
	var fileInfos []os.FileInfo
	var err error
	var SizeTotal, GrandTotal int64
	var GrandTotalCount int
	var excludeRegexPattern string

	uid := 0
	gid := 0
	systemStr := ""

	winFlag := runtime.GOOS == "windows"
	ctfmt.Print(ct.Magenta, winFlag, "ds -- Directory SoRTed w/ filename truncation.  LastAltered ", LastAltered, ", compiled using ", runtime.Version(), ".")
	fmt.Println()

	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		//autoDefaults = false
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	// environment variable processing.  If present, these will be the defaults.
	dsrtParam = ProcessEnvironString() // This is a function below.

	sepString := string(filepath.Separator)
	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.  Function avail as of Go 1.12
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + sepString

	if runtime.GOARCH == "amd64" {
		uid = os.Getuid() // int
		gid = os.Getgid() // int
		userptr, err = user.Current()
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
		fmt.Fprintf(flag.CommandLine.Output(), " dsrt values are: numlines=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelistflag=%t, totalflag=%t \n",
			dsrtParam.numlines, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag, dsrtParam.totalflag)

		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	revflag := flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
	var RevFlag bool
	flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nscreens = flag.Int("n", 1, "number of screens to display, ie, a multiplier") // Ptr
	var NLines int
	flag.IntVar(&NLines, "N", 0, "number of lines to display") // Value

	var sizeflag = flag.Bool("s", false, "sort by size instead of by date") // pointer
	var SizeFlag bool
	flag.BoolVar(&SizeFlag, "S", false, "sort by size instead of by date")

	var DirListFlag = flag.Bool("d", false, "include directories in the output listing") // pointer

	var FilenameListFlag bool
	flag.BoolVar(&FilenameListFlag, "D", false, "Directories only in the output listing")

	var TotalFlag = flag.Bool("t", false, "include grand total of directory")

	flag.BoolVar(&testFlag, "test", false, "enter a testing mode to println more variables")
	flag.BoolVar(&testFlag, "v", false, "verbose mode, which is same as test mode.")

	var longflag = flag.Bool("l", false, "long file size format.") // Ptr

	var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")
	flag.StringVar(&excludeRegexPattern, "x", "", "regex to be excluded from output.") // var, not a ptr.

	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	var filterFlag = flag.Bool("f", false, "filter value to suppress listing individual size below 1 MB.")

	var w int // width maximum of the filename string to be displayed
	flag.IntVar(&w, "w", 0, "width for displayed file name")

	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, testFlag
		testFlag = true
	}

	if testFlag {
		fmt.Println(" After flag.Parse(); option switches w=", w, "nscreens=", *nscreens, "Nlines=", NLines)
	}

	Reverse := *revflag || RevFlag || dsrtParam.reverseflag
	Forward := !Reverse // convenience variable

	SizeSort := *sizeflag || SizeFlag || dsrtParam.sizeflag
	DateSort := !SizeSort // convenience variable

	if NLines > 0 { // priority is -N option
		numOfLines = NLines
	} else if dsrtParam.numlines > 0 { // then check this
		numOfLines = dsrtParam.numlines
	} else if autoHeight > 0 { // finally use autoHeight.
		numOfLines = autoHeight - 7
	} else { // intended if autoHeight fails, just in case.
		numOfLines = defaultHeight
	}

	numOfLines *= *nscreens

	noExtensionFlag = *extensionflag || *extflag

	if len(excludeRegexPattern) > 0 {
		if testFlag {
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
		if w <= 0 || w > maxWidth { // if w is zero then there is no dsw environment variable to set it.
			w = minWidth
		}
	}

	// set which sort function will be in the sortfcn var
	sortfcn := func(i, j int) bool { return false } // became available as of Go 1.8
	if SizeSort && Forward {                        // set the value of sortfcn so only a single line is needed to execute the sort.
		sortfcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfos[i].Size() > fileInfos[j].Size() // I want a largest first sort
		}
		if testFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfos[i].ModTime().After(fileInfos[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
		}
		if testFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfos[i].Size() < fileInfos[j].Size() // I want an smallest-first sort
		}
		if testFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime()) // I want an oldest-first sort
		}
		if testFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if testFlag {
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println()
		if runtime.GOARCH == "amd64" {
			fmt.Printf(" uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s.\n",
				uid, gid, systemStr, userptr.Uid, userptr.Gid, userptr.Username, userptr.Name, userptr.HomeDir)
			fmt.Printf(" Autoheight=%d, autowidth=%d, w=%d, numOfLines=%d. \n", autoHeight, autoWidth, w, numOfLines)
			fmt.Printf(" dsrtparam numlines=%d, w=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelist=%t, totalflag=%t\n",
				dsrtParam.numlines, dsrtParam.w, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag,
				dsrtParam.totalflag)
		}
	}

	// If the character is a letter, it has to be k, m or g.  Or it's a number, but not both.  For now.
	if *filterFlag {
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

	if testFlag {
		fmt.Println(" FilterFlag =", *filterFlag, ".  filterStr =", filterStr, ". filterAmt =", filterAmt)
	}

	fileInfos = getFileInfosFromCommandLine()
	if len(fileInfos) > 1 {
		sort.Slice(fileInfos, sortfcn)
	}

	cs := getColorizedStrings(fileInfos)

	if testFlag {
		fmt.Printf(" Len(fileinfos)=%d, len(colorizedStrings)=%d, numOfLines=%d\n", len(fileInfos), len(cs), numOfLines)
	}

	// Now to output the colorStringSlice, 1 item per line
	columnWidth := w - 1 // a tolerance factor.
	for i, css := range cs {
		s0 := fixedStringLen(css.str, columnWidth)
		ctfmt.Printf(css.color, winFlag, "%s\n", s0)
		if i >= numOfLines {
			break
		}
	}

	fmt.Println()

	s := fmt.Sprintf("%d", SizeTotal)
	if SizeTotal > 100000 {
		s = AddCommas(s)
	}
	s0 := fmt.Sprintf("%d", GrandTotal)
	if GrandTotal > 100000 {
		s0 = AddCommas(s0)
	}
	fmt.Print(" File Size total = ", s)
	if ShowGrandTotal {
		s1, color := getMagnitudeString(GrandTotal)
		ctfmt.Println(color, true, ", Directory grand total is", s0, "or approx", s1, "in", GrandTotalCount, "files.")
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

// ------------------------------------ ProcessEnvironString ---------------------------------------

func ProcessEnvironString() DsrtParamType { // use system utils when can because they tend to be faster
	var dsrtparam DsrtParamType

	dswStr, ok := os.LookupEnv("dsw")
	if ok {
		n, err := strconv.Atoi(dswStr)
		if err == nil {
			dsrtparam.w = n
		} else {
			fmt.Fprintf(os.Stderr, " dsw environment variable not a valid number.  dswStr = %q, %v.  Ignored.", dswStr, err)
			dsrtparam.w = 0
		}
	} else { // not ok, ie, dsw variable not found in environment
		dsrtparam.w = 0
	}

	envStr, ok := os.LookupEnv("dsrt")
	if !ok {
		return dsrtparam
	}

	indiv := strings.Split(envStr, "") // this splits into individual characters

	for j, str := range indiv {
		envChar := str[0]
		if envChar == 'r' || envChar == 'R' {
			dsrtparam.reverseflag = true
		} else if envChar == 's' || envChar == 'S' {
			dsrtparam.sizeflag = true
		} else if envChar == 'd' {
			dsrtparam.dirlistflag = true
		} else if envChar == 'D' {
			dsrtparam.filenamelistflag = true
		} else if envChar == 't' { // added 09/12/2018 12:26:01 PM
			dsrtparam.totalflag = true // for the grand total operation
		} else if unicode.IsDigit(rune(envChar)) {
			dsrtparam.numlines = int(envChar) - int('0')
			if j+1 < len(indiv) && unicode.IsDigit(rune(indiv[j+1][0])) {
				dsrtparam.numlines = 10*dsrtparam.numlines + int(indiv[j+1][0]) - int('0')
				break // if have a 2 digit number, it ends processing of the indiv string
			}
		}
	}
	return dsrtparam
} // end ProcessEnvironString

//------------------------------ GetDirectoryAliases ----------------------------------------
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
func myReadDir(dir string) []os.FileInfo {
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
		fi, err := os.Lstat(s)
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
} // myReadDir
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
	if excludeFlag {
		if BOOL := excludeRegex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	return true
}
