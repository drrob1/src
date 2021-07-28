// rex.go -- directory sort using a regular expression pattern

package main

import (
	"bytes"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	//"time"
	"unicode"
)

const LastAltered = "Jul 28, 2021"

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
18 Jul 19 -- When there is an error from ioutil.ReadDir, I cannot change its behavior of not reading any more.  Just do dsrt * in bash as a work around.
19 Jul 19 -- Wrote MyReadDir
22 Jul 19 -- Added a winflag check so don't scan commandline on linux looking for : or ~.
 9 Sep 19 -- From Israel: Fixing issue on linux when entering a directory param.  And added test flag.  And added sortfcn.
22 Sep 19 -- Changed the error message under linux and have only 1 item on command line.  Error condition is likely file not found.
 4 Oct 19 -- No longer need platform specific code.  So I added GetUserGroupStrLinux.  And then learned that it won't compile on Windows.
               So as long as I want the exact same code for both platforms, I do need platform specific code.
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

const defaultlineswin = 50
const defaultlineslinux = 40
const maxWidth = 300
const minWidth = 90

//const dateSize = 30  // space the filesize and date occupy.

var directoryAliasesMap dirAliasMapType

func main() {
	var dsrtparam DsrtParamType
	var numoflines int
	var files []os.FileInfo
	var err error
	var count int
	var SizeTotal, GrandTotal int64
	var GrandTotalCount, autoheight, autowidth int
	var systemStr string
	var excludeRegexPattern string
	colorStringSlice := make([]colorizedStr, 0, 200)

	// environment variable processing.  If present, these will be the defaults.
	dsrtparam = ProcessEnvironString() // This is a function below.

	autoDefaults := term.IsTerminal(int(os.Stdout.Fd()))
	linuxflag := runtime.GOOS == "linux"
	winflag := runtime.GOOS == "windows"

	if !autoDefaults {
		if winflag {
			comspec, ok := os.LookupEnv("ComSpec")
			if ok {
				bytesbuf := bytes.NewBuffer([]byte{}) // from Go Standard Library Cookbook by Radomir Sohlich (C) 2018 Packtpub
				tcc := exec.Command(comspec, "-C", "echo", "%_columns")
				tcc.Stdout = bytesbuf
				tcc.Run()
				colstr := bytesbuf.String()
				lines := strings.Split(colstr, "\n")
				trimmedLine := strings.TrimSpace(lines[1]) // 2nd line of the output is what I want trimmed
				autowidth, err = strconv.Atoi(trimmedLine)
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
				autoheight, err = strconv.Atoi(trimmedLine)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error from rows conversion is", err, "Value ignored.")
				}

			} else {
				fmt.Fprintln(os.Stderr, "comspec expected but not found.  Using environment params settings only.")
			}
		} else {
			fmt.Fprintln(os.Stderr, "Expected a windows computer, but winflag is false.  WTF?")
			autowidth = minWidth
			autoheight = defaultlineslinux
		}
	} else {
		autowidth, autoheight, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			autoDefaults = false
			fmt.Fprintln(os.Stderr, " From term.Getsize:", err)
			autowidth = minWidth
			autoheight = defaultlineslinux
		}
	}

	if linuxflag {
		systemStr = "Linux"
		files = make([]os.FileInfo, 0, 500)
		if dsrtparam.numlines > 0 { // setting this on command line take priority over defaults
			numoflines = dsrtparam.numlines
		} else if autoheight > 0 {
			numoflines = autoheight - 5
		} else {
			numoflines = defaultlineslinux
		}
	} else if winflag {
		systemStr = "Windows"
		if dsrtparam.numlines > 0 {
			numoflines = dsrtparam.numlines
		} else if autoheight > 0 {
			numoflines = autoheight - 5
		} else {
			numoflines = defaultlineswin
		}
	} else {
		systemStr = "Unknown"
		numoflines = defaultlineslinux
	}

	sepstring := string(filepath.Separator)
	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.  Function avail as of go 1.12.
	if err != nil {
		HomeDirStr = ""
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr.")
	}
	HomeDirStr = HomeDirStr + sepstring

	// flag definitions and processing
	var revflag = flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
	var RevFlag bool
	flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nscreens = flag.Int("n", 1, "number of screens to display") // Ptr
	var NLines int
	flag.IntVar(&NLines, "N", 0, "number of lines to display") // Value

	var helpflag = flag.Bool("h", false, "print help message") // pointer
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "help", false, "print help message")

	var sizeflag = flag.Bool("s", false, "sort by size instead of by date") // pointer
	var SizeFlag bool
	flag.BoolVar(&SizeFlag, "S", false, "sort by size instead of by date")

	var DirListFlag = flag.Bool("d", false, "include directories in the output listing") // pointer

	var FilenameListFlag bool
	flag.BoolVar(&FilenameListFlag, "D", false, "Directories only in the output listing")

	var TotalFlag = flag.Bool("t", false, "include grand total of directory")

	var testFlag bool
	flag.BoolVar(&testFlag, "test", false, "enter a testing mode to println more variables")
	flag.BoolVar(&testFlag, "v", false, "enter a verbose (testing) mode to println more variables")

	var longflag = flag.Bool("l", false, "long file size format.") // Ptr

	var excludeFlag = flag.Bool("x", false, "exclude regex entered after prompt")            // pointer
	flag.StringVar(&excludeRegexPattern, "exclude", "", "regex to be excluded from output.") // var, not a ptr.

	var w int
	flag.IntVar(&w, "w", 0, " width of full displayed screen.")

	flag.Parse()

	fmt.Print(" rex will display sorted by date or size in 1 column.  LastAltered ", LastAltered, ", compiled using ", runtime.Version(), ".")
	fmt.Println()

	if testFlag {
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println()
		fmt.Println("winflag:", winflag, ", linuxflag:", linuxflag, "systemStr:", systemStr)
		fmt.Println()
		fmt.Println(" After flag.Parse(); option switches w=", w, "nscreens=", *nscreens, "Nlines=", NLines)
	}

	if *helpflag || HelpFlag {
		fmt.Println()
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
		os.Exit(0)
	}

	Reverse := *revflag || RevFlag || dsrtparam.reverseflag
	Forward := !Reverse // convenience variable

	SizeSort := *sizeflag || SizeFlag || dsrtparam.sizeflag
	DateSort := !SizeSort // convenience variable

	NumLines := numoflines
	if NLines > 0 {
		NumLines = NLines
	}
	NumLines *= *nscreens

	var excludeRegex *regexp.Regexp
	if len(excludeRegexPattern) > 0 {
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			*excludeFlag = false
		}
		*excludeFlag = true
	} else if *excludeFlag {
		ctfmt.Print(ct.Yellow, winflag, " Enter regex pattern to be excluded: ")
		fmt.Scanln(&excludeRegexPattern)
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			*excludeFlag = false
		}
	}

	Dirlist := *DirListFlag || FilenameListFlag || dsrtparam.dirlistflag || dsrtparam.filenamelistflag // if -D entered then this expression also needs to be true.
	FilenameList := !(FilenameListFlag || dsrtparam.filenamelistflag)                                  // need to reverse the flag because D means suppress the output of filenames.
	ShowGrandTotal := *TotalFlag || dsrtparam.totalflag                                                // added 09/12/2018 12:32:23 PM
	LongFileSizeList := *longflag

	inputRegEx := ""
	workingdir, _ := os.Getwd()
	startdir := workingdir

	if w == 0 { // w not set by command line flag
		w = dsrtparam.w
	}
	if autowidth > 0 {
		if w <= 0 || w > maxWidth {
			w = autowidth
		}
	} else {
		if w <= 0 || w > maxWidth {
			w = minWidth
		}
	}

	commandlineSlice := flag.Args()
	if len(commandlineSlice) == 0 {
		inputRegEx = "." // no regex entered on command line, default is everything, esp useful for testing.
	} else if len(commandlineSlice) == 1 {
		inputRegEx = commandlineSlice[0]
	} else if len(commandlineSlice) == 2 {
		inputRegEx = commandlineSlice[0]
		workingdir = commandlineSlice[1]

		if winflag { // added the winflag check so don't have to scan commandline on linux, which would be wasteful.
			if strings.ContainsRune(commandlineSlice[1], ':') {
				workingdir = ProcessDirectoryAliases(directoryAliasesMap, workingdir)
			} else if strings.Contains(commandlineSlice[1], "~") { // this can only contain a ~ on Windows.
				workingdir = strings.Replace(workingdir, "~", HomeDirStr, 1)
			}
		}
		fi, err := os.Lstat(workingdir)
		if err != nil || !fi.Mode().IsDir() {
			fmt.Println(workingdir, "is an invalid directory name.  Will use", startdir, "instead.")
			workingdir = startdir
		}
	} else {
		log.Fatalln("too many params on line.  Usage: regex pattern directory")
	}
	if testFlag {
		fmt.Println("inputRegEx=", inputRegEx, ", and workingdir =", workingdir)
	}
	inputRegEx = strings.ToLower(inputRegEx)

	// set which sort function will be in the sortfcn var
	sortfcn := func(i, j int) bool { return false }
	if SizeSort && Forward { // set the value of sortfcn so only a single line is needed to execute the sort.
		sortfcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return files[i].Size() > files[j].Size() // I want a largest first sort
		}
		if testFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return files[i].ModTime().After(files[j].ModTime()) // I want a newest-first sort
		}
		if testFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return files[i].Size() < files[j].Size() // I want a smallest-first sort
		}
		if testFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return files[i].ModTime().Before(files[j].ModTime()) // I want an oldest-first sort
		}
		if testFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	// files, err = ioutil.ReadDir(workingdir) depracated as of Go 1.16
	openedDir, err := os.Open(workingdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Open directory err is", err)
	}

	files, err = openedDir.Readdir(0) // zero means return all filemode entries into the returned slice
	if err != nil {                   // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
		log.Println(err, "so calling my own MyReadDir.")
		files = MyReadDir(workingdir)
	}
	if ShowGrandTotal {
		for _, f := range files {
			if f.Mode().IsRegular() {
				GrandTotal += f.Size()
				GrandTotalCount++
			}
		}
	}
	sort.Slice(files, sortfcn)

	if testFlag {
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println()
		fmt.Printf(" Autodefault=%v, autoheight=%d, autowidth=%d, w=%d, numlines=%d. \n", autoDefaults, autoheight, autowidth, w, NumLines)
		fmt.Printf(" dsrtparam numlines=%d, w=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelist=%t, totalflag=%t\n",
			dsrtparam.numlines, dsrtparam.w, dsrtparam.reverseflag, dsrtparam.sizeflag, dsrtparam.dirlistflag, dsrtparam.filenamelistflag,
			dsrtparam.totalflag)
		fmt.Println(" Dirname is", workingdir)
	}
	fmt.Println()

	// I need to add a description of how this code works, because I forgot.
	// The entire contents of the directory is read in.  Then the slice of fileinfo's is sorted, and finally only the matching filenames are displayed.

	regex, err := regexp.Compile(inputRegEx)
	if err != nil {
		log.Fatalln(" error from regex compile function is ", err)
		fmt.Println()
		fmt.Println()
		os.Exit(1)
	}

	for _, f := range files {
		NAME := strings.ToLower(f.Name())
		nameStr := f.Name() // truncStr(f.Name(), w)
		if BOOL := regex.MatchString(NAME); BOOL {
			if *excludeFlag {
				if flag := excludeRegex.MatchString(strings.ToLower(f.Name())); flag {
					continue
				}
			}
			modtimeStr := f.ModTime().Format("Jan-02-2006_15:04:05")
			sizestr := ""
			if FilenameList && f.Mode().IsRegular() {
				SizeTotal += f.Size()
				if LongFileSizeList {
					sizestr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
					if f.Size() > 100000 {
						sizestr = AddCommas(sizestr)
					}
					s := fmt.Sprintf("%8s %s %s", sizestr, modtimeStr, nameStr)
					_, color := getMagnitudeString(f.Size())
					colorized := colorizedStr{color: color, str: s}
					colorStringSlice = append(colorStringSlice, colorized)
				} else {
					var color ct.Color
					sizestr, color = getMagnitudeString(f.Size())
					s := fmt.Sprintf("%-8s %s %s", sizestr, modtimeStr, nameStr)
					colorized := colorizedStr{color: color, str: s}
					colorStringSlice = append(colorStringSlice, colorized)
				}
				count++
			} else if IsSymlink(f.Mode()) {
				s := fmt.Sprintf("%6s %s <%s>", sizestr, modtimeStr, nameStr)
				colorized := colorizedStr{color: ct.White, str: s}
				colorStringSlice = append(colorStringSlice, colorized)
				count++
			} else if Dirlist && f.IsDir() {
				s := fmt.Sprintf("%6s %s (%s)", sizestr, modtimeStr, nameStr)
				colorized := colorizedStr{color: ct.White, str: s}
				colorStringSlice = append(colorStringSlice, colorized)
				count++
			}
			if count >= NumLines {
				break
			}
		}
	}

	// Output the colorStringSlice, 1 items per line for this version of the code.
	columnWidth := w - 1
	for _, css := range colorStringSlice {
		c0 := css.color
		s0 := fixedStringLen(css.str, columnWidth)
		ctfmt.Printf(c0, winflag, "%s\n", s0)
	}

	s := fmt.Sprintf("%d", SizeTotal)
	if SizeTotal > 100000 {
		s = AddCommas(s)
	}
	s0 := fmt.Sprintf("%d", GrandTotal)
	if GrandTotal > 100000 {
		s0 = AddCommas(s0)
	}
	fmt.Println()
	fmt.Print(" File Size total = ", s)
	if ShowGrandTotal {
		s1, color := getMagnitudeString(GrandTotal)
		ctfmt.Println(color, winflag, ", Directory grand total is", s0, "or approx", s1, "in", GrandTotalCount, "files.")
	} else {
		fmt.Println(".")
	}
	fmt.Println()
} // end main regex

//-------------------------------------------------------------------- InsertByteSlice
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

// ---------------------------- GetIDname -----------------------------------------------------------
func GetIDname(uidStr string) string {

	if len(uidStr) == 0 {
		return ""
	}
	ptrToUser, err := user.LookupId(uidStr)
	if err != nil {
		panic("uid not found")
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
			fmt.Fprintf(os.Stderr, " dsw environ var not a valid number.  dswStr= %q, %v.  Ignored.", dswStr, err)
			dsrtparam.w = 0
		}
	} else { // dswStr not in environ, ie not ok
		dsrtparam.w = 0
	}

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

// ------------------------------- MyReadDir -----------------------------------

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

	fi := make([]os.FileInfo, 0, len(names))
	for _, s := range names {
		L, err := os.Lstat(s)
		if err != nil {
			log.Println(" Error from os.Lstat ", err)
			continue
		}
		fi = append(fi, L)
	}
	return fi
} // MyReadDir

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

/*
{{{
func truncStr(s string, w int) string {
	if w <= 0 || len(s) < w {
		return s
	}
	return s[:w]
}}}
*/

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
