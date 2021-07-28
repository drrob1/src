// ds3.go -- directory sort outputting 3 columns

package main

import (
	"bytes"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const LastAltered = "27 July 2021"

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
10 Jul 21 -- Now called dsc for directory sort columns.  I'm going to write the output to a slice that I can display after it's generated.
               It will not show the mode bits on linux; it never showed the mode bits on Windows.
               I'm going to start using DirEntry.  Nevermind, because the IsRegular function only applies to a FileInfo struct.  But I'm no
               longer using ioutil.ReadDir to get the slice of FileInfo's, as this function is depracated as of Go 1.16.
21 Jul 21 -- Now called dsc3, for a 3 column output, but showing the filesizes in a more compressed format.  And
             It will set "w" to be 2/3's of dsw environ var.  And use a file timestamp only including the date and not the time.
25 Jul 21 -- I'm adding the code to determine the number of rows and columns itself.  I'll use golang.org/x/term for linux, and shelling out to tcc for Windows.
               Now that I know autoheight, I'll have n be a multiplier for the number of screens to display, each autolines - 5 in size.  N will remain as is.
               And it's now called ds3.go
27 Jul 21 -- I'm removing truncStr and will use fixedStringLen instead.
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

func main() {
	const defaultlineswin = 50
	const defaultlineslinux = 40
	//	const maxTruncationValue = 45
	//	const minTruncationValue = 30
	const maxwidth = 300
	const minwidth = 210
	//	const datesize = 20 // amount of size filesize and date occupy in each column
	var dsrtparam DsrtParamType
	var numoflines int
	var files []os.FileInfo
	var err error
	var count int
	var SizeTotal, GrandTotal int64
	var GrandTotalCount int
	var autoheight, autowidth int
	var havefiles bool
	var commandline string
	var directoryAliasesMap dirAliasMapType
	var excludeRegexPattern string
	colorStringSlice := make([]colorizedStr, 0, 200) // the string slice to be displayed after generation.

	// environment variable processing.  If present, these will be the defaults.
	dsrtparam = ProcessEnvironString() // This is a function below.

	linuxflag := runtime.GOOS == "linux"
	winflag := runtime.GOOS == "windows"
	ctfmt.Print(ct.Magenta, winflag, "ds3 will display Directory SoRTed by date or size in 3 columns.  LastAltered ", LastAltered, ", compiled using ",
		runtime.Version(), ".")
	fmt.Println()

	autoDefaults := term.IsTerminal(int(os.Stdout.Fd())) // this works on Windows now that I use Stdout instead of Stdin=0.
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
			autoheight = defaultlineslinux
			autowidth = minwidth
		}
	} else {
		autowidth, autoheight, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			autoDefaults = false
		}
	}

	if linuxflag {
		files = make([]os.FileInfo, 0, 500)
		if dsrtparam.numlines > 0 {
			numoflines = dsrtparam.numlines
		} else if autoheight > 0 {
			numoflines = autoheight - 7
		} else {
			numoflines = defaultlineslinux
		}
	} else if winflag {
		if dsrtparam.numlines > 0 {
			numoflines = dsrtparam.numlines
		} else if autoheight > 0 {
			numoflines = autoheight - 7
		} else {
			numoflines = defaultlineswin
		}
	} else {
		numoflines = defaultlineslinux
	}

	sepstring := string(filepath.Separator)
	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + sepstring

	if runtime.GOARCH == "amd64" {
		if err != nil {
			fmt.Println(" user.Current error is ", err, "Exiting.")
			os.Exit(1)
		}
	}

	// flag definitions and processing
	revflag := flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
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
	flag.BoolVar(&testFlag, "v", false, "verbose mode, which is same as test mode.")

	var longflag = flag.Bool("l", false, "long file size format.") // Ptr

	var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	var excludeFlag = flag.Bool("x", false, "exclude regex entered after prompt")
	flag.StringVar(&excludeRegexPattern, "exclude", "", "regex to be excluded from output.") // var, not a ptr.

	var filterAmt int
	var filterStr string
	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	var filterFlag = flag.Bool("f", false, "filter value to suppress listing individual size below 1 MB.")

	var w int // width of screen, in pixels
	flag.IntVar(&w, "w", 0, "width for displayed screen output")

	flag.Parse()

	if testFlag {
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println()
		fmt.Println(" After flag.Parse(): option switches w=", w, "nscreens=", nscreens, "Nlines=", NLines)
		fmt.Println()
	}

	if *helpflag || HelpFlag {
		fmt.Println(" Reads from dsrt environment variable before processing commandline switches.")
		fmt.Println(" Reads from diraliases environment variable if needed on Windows.")
		flag.PrintDefaults()
		os.Exit(0)
	}

	Reverse := *revflag || RevFlag || dsrtparam.reverseflag
	Forward := !Reverse // convenience variable

	SizeSort := *sizeflag || SizeFlag || dsrtparam.sizeflag
	DateSort := !SizeSort // convenience variable

	NumLines := numoflines
	if NLines > 0 { // -N option switch has priority over autoheight
		NumLines = NLines
	}

	noExtensionFlag := *extensionflag || *extflag
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
	LongFileSizeList := *longflag

	ShowGrandTotal := *TotalFlag || dsrtparam.totalflag // added 09/12/2018 12:32:23 PM

	CleanDirName := ""
	CleanFileName := ""
	filenamesStringSlice := flag.Args() // Intended to process linux command line filenames.

	if w == 0 { // w not set by flag option
		w = dsrtparam.w // will still be zero if dsw environ var not set.
	}
	if autowidth > 0 {
		if w <= 0 || w > maxwidth { // w not set by flag.Parse or dsw environ var
			w = autowidth // I expect that this will always be used.
		}
	} else {
		if w <= 0 || w > maxwidth { // if w is zero then there is no dsw environment variable to set it.
			w = minwidth
		}
	}

	if *nscreens > 1 {
		NumLines *= *nscreens
	}

	// set which sort function will be in the sortfcn var
	sortfcn := func(i, j int) bool { return false } // became available as of Go 1.8
	if SizeSort && Forward {                        // set the value of sortfcn so only a single line is needed to execute the sort.
		sortfcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return files[i].Size() > files[j].Size() // I want a largest first sort
		}
		if testFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return files[i].ModTime().After(files[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
		}
		if testFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			return files[i].Size() < files[j].Size() // I want an smallest-first sort
		}
		if testFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return files[i].ModTime().Before(files[j].ModTime()) // I want an oldest-first sort
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
			fmt.Printf(" Autodefault=%v, autoheight=%d, autowidth=%d, w=%d, numlines=%d. \n", autoDefaults, autoheight, autowidth, w, NumLines)
			fmt.Printf(" dsrtparam numlines=%d, w=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelist=%t, totalflag=%t\n",
				dsrtparam.numlines, dsrtparam.w, dsrtparam.reverseflag, dsrtparam.sizeflag, dsrtparam.dirlistflag, dsrtparam.filenamelistflag,
				dsrtparam.totalflag)
		}
	}

	if linuxflag && len(filenamesStringSlice) > 0 { // linux command line processing for filenames.  This condition had to be fixed July 4, 2019.
		paramIsDir := false
		if len(filenamesStringSlice) == 1 {
			// need to determine if the 1 param on command line is a directory
			fi, err := os.Lstat(filenamesStringSlice[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				fmt.Println()
				fmt.Println()
				os.Exit(1)
			}

			paramIsDir = fi.Mode().IsDir()
			if testFlag {
				fmt.Println(" have only 1 param on line. filenameStringSlice=", filenamesStringSlice[0], "paramIsDir=", paramIsDir)
				fmt.Println()
			}
			if paramIsDir {
				CleanDirName = fi.Name()
			} else { // not a param so this one file needs to be displayed.
				files = append(files, fi)
				havefiles = true
			}
		} else { // bash has placed more than one file in the command line, ie, len(filenameStringSlice) > 1
			for _, s := range filenamesStringSlice { // fill a slice of fileinfo
				fi, err := os.Lstat(s)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}

				files = append(files, fi)
				if fi.Mode().IsRegular() && ShowGrandTotal {
					GrandTotal += fi.Size()
					GrandTotalCount++
				}
			}
			sort.Slice(files, sortfcn)
			havefiles = true
		}

	} else { // either no params were present on the command line or this is running under Windows and may have a command line param.
		// commandline = filenamesStringSlice[0] -- this panics if there are no params on the line.
		commandline = flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.
	}

	if winflag && len(commandline) > 0 { // added the winflag check so don't have to scan commandline on linux, which would be wasteful.
		if strings.ContainsRune(commandline, ':') {
			commandline = ProcessDirectoryAliases(directoryAliasesMap, commandline)
		} else if strings.Contains(commandline, "~") { // this can only contain a ~ on Windows.
			commandline = strings.Replace(commandline, "~", HomeDirStr, 1)
		}
		CleanDirName, CleanFileName = filepath.Split(commandline)
		CleanDirName = filepath.Clean(CleanDirName)
		CleanFileName = strings.ToLower(CleanFileName)
	}

	if len(CleanDirName) == 0 {
		workingdir, _ := os.Getwd()
		if testFlag {
			fmt.Println(" CleanDirName is empty, and will be ", workingdir)
		}
		CleanDirName = workingdir
	}

	if len(CleanFileName) == 0 {
		CleanFileName = "*"
	}

	if !havefiles {
		// files, err = ioutil.ReadDir(CleanDirName) depracated as of Go 1.16
		openedDir, err := os.Open(CleanDirName)
		if err != nil {
			fmt.Fprintln(os.Stderr, " Open directory error is", err)
			// Don't know what I'm going to do yet
		}

		files, err = openedDir.Readdir(0) // zero means read all FileMode entries into the returned slice of file modes.
		if err != nil {                   // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
			fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
			files = MyReadDir(CleanDirName)
		}
		if ShowGrandTotal { // this optimization added 2/27/21, and it made a big difference.
			for _, f := range files {
				if f.Mode().IsRegular() {
					GrandTotal += f.Size()
					GrandTotalCount++
				}
			}
		}
		sort.Slice(files, sortfcn)
	}

	fmt.Println(" Dirname is", CleanDirName)

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

	// I need to add a description of how this code works, because I forgot.
	// The entire contents of the directory is read in by either ioutil.ReadDir or MyReadDir.  Then the slice of fileinfo's is sorted, and finally only the matching filenames are displayed.
	// This is still the way it works for Windows.
	// On linux, bash populated the command line by globbing, or no command line params were entered

	if linuxflag {
		for _, f := range files {
			//modTimeStr := f.ModTime().Format("Jan-02-2006_15:04:05")
			modTimeStr := f.ModTime().Format("Jan-02-2006")
			nameStr := f.Name() //    truncStr(f.Name(), oneThirdWidth)
			sizestr := ""
			if FilenameList && f.Mode().IsRegular() {
				SizeTotal += f.Size()
				showthis := true
				if noExtensionFlag && strings.ContainsRune(f.Name(), '.') {
					showthis = false
				}
				if *excludeFlag {
					if BOOL := excludeRegex.MatchString(strings.ToLower(f.Name())); BOOL {
						showthis = false
					}
				}
				if filterAmt > 0 {
					if f.Size() < int64(filterAmt) {
						showthis = false
					}
				}
				if showthis {
					if LongFileSizeList {
						sizestr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
						if f.Size() > 100000 {
							sizestr = AddCommas(sizestr)
						}
						s := fmt.Sprintf("%7s %s %s", sizestr, modTimeStr, nameStr) // can't be colorized
						colorized := colorizedStr{color: ct.White, str: s}
						colorStringSlice = append(colorStringSlice, colorized)
					} else {
						var color ct.Color
						sizestr, color = getMagnitudeString(f.Size())
						s := fmt.Sprintf("%-7s %s %s", sizestr, modTimeStr, nameStr)
						colorized := colorizedStr{color: color, str: s}
						colorStringSlice = append(colorStringSlice, colorized)
					}
					count++
				}
			} else if IsSymlink(f.Mode()) {
				s := fmt.Sprintf("%5s %s <%s>", sizestr, modTimeStr, nameStr)
				colorized := colorizedStr{color: ct.White, str: s}
				colorStringSlice = append(colorStringSlice, colorized)
				count++
			} else if Dirlist && f.IsDir() {
				s := fmt.Sprintf("%5s %s (%s)", sizestr, modTimeStr, nameStr)
				colorized := colorizedStr{color: ct.White, str: s}
				colorStringSlice = append(colorStringSlice, colorized)
				count++
			}
			if count >= NumLines*3 {
				break
			}
		}
	} else if winflag {
		for _, f := range files {
			showthis := false
			NAME := strings.ToLower(f.Name())
			nameStr := f.Name() //       truncStr(f.Name(), oneThirdWidth)
			// trying to figure out how to implement the noextensionflag.  I'm thinking that I will create a flag that will
			// be true if this file is to be printed, ie, either the flag is off or the flag is on and there is a '.' in the filename.
			// This way, the condition below can be BOOL && thisNewFlag
			BOOL, _ := filepath.Match(CleanFileName, NAME)
			if BOOL {
				showthis = true
				if noExtensionFlag && strings.ContainsRune(NAME, '.') {
					showthis = false
				}
				if *excludeFlag {
					if flag := excludeRegex.MatchString(strings.ToLower(NAME)); flag {
						showthis = false
					}
				}
				if filterAmt > 0 {
					if f.Size() < int64(filterAmt) {
						showthis = false
					}
				}
			}

			//			if BOOL, _ := filepath.Match(CleanFileName, NAME); BOOL {
			if showthis {
				//modTimeStr := f.ModTime().Format("Jan-02-2006_15:04:05")
				modTimeStr := f.ModTime().Format("Jan-02-2006")
				sizestr := ""
				if FilenameList && f.Mode().IsRegular() {
					SizeTotal += f.Size()
					if LongFileSizeList {
						sizestr = strconv.FormatInt(f.Size(), 10)
						if f.Size() > 100000 {
							sizestr = AddCommas(sizestr)
						}
						s := fmt.Sprintf("%7s %s %s", sizestr, modTimeStr, nameStr)
						var color = ct.White
						colorized := colorizedStr{color: color, str: s}
						colorStringSlice = append(colorStringSlice, colorized)
					} else {
						var color ct.Color
						sizestr, color = getMagnitudeString(f.Size())
						s := fmt.Sprintf("%-7s %s %s", sizestr, modTimeStr, nameStr)
						colorized := colorizedStr{color: color, str: s}
						colorStringSlice = append(colorStringSlice, colorized)
					}
					count++
				} else if IsSymlink(f.Mode()) {
					var color = ct.White
					s := fmt.Sprintf("%5s %s <%s>", sizestr, modTimeStr, nameStr)
					colorized := colorizedStr{color: color, str: s}
					colorStringSlice = append(colorStringSlice, colorized)
					count++
				} else if Dirlist && f.IsDir() {
					var color = ct.White
					s := fmt.Sprintf("%5s %s (%s)", sizestr, modTimeStr, nameStr)
					colorized := colorizedStr{color: color, str: s}
					colorStringSlice = append(colorStringSlice, colorized)
					count++
				}
				if count >= NumLines*3 {
					break
				}
			}
		}
	}

	// Now to output the colorStringSlice, 3 items per line, but I want the sort to remain vertical
	onethirdpoint := len(colorStringSlice) / 3
	columnWidth := w/3 - 3 // accounting for the extra spaces added by the formatting.
	for i := 0; i < onethirdpoint; i++ {
		c0 := colorStringSlice[i].color
		s0 := fixedStringLen(colorStringSlice[i].str, columnWidth)
		ctfmt.Printf(c0, winflag, "%s  ", s0)

		if i+onethirdpoint < len(colorStringSlice) {
			c1 := colorStringSlice[i+onethirdpoint].color
			s1 := fixedStringLen(colorStringSlice[i+onethirdpoint].str, columnWidth)
			ctfmt.Printf(c1, winflag, "%s  ", s1)
		} else {
			fmt.Println()
		}

		if i+2*onethirdpoint < len(colorStringSlice) {
			c2 := colorStringSlice[i+2*onethirdpoint].color
			s2 := fixedStringLen(colorStringSlice[i+2*onethirdpoint].str, columnWidth)
			ctfmt.Printf(c2, winflag, "%s\n", s2)
		} else {
			fmt.Println()
		}
	}

	fmt.Println()

	s := fmt.Sprintf("%d", SizeTotal)
	if SizeTotal > 100000 {
		s = AddCommas(s)
	}
	fmt.Print(" File Size total = ", s)

	if ShowGrandTotal {
		s0 := fmt.Sprintf("%d", GrandTotal)
		if GrandTotal > 100000 {
			s0 = AddCommas(s0)
		}

		s1, color := getMagnitudeString(GrandTotal)
		ctfmt.Println(color, true, ", Directory grand total is", s0, "or approx", s1, "in", GrandTotalCount, "files.")
	} else {
		fmt.Println(".")
	}
	fmt.Println()
} // end main ds3

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

/*
// ---------------------------- GetIDname -----------------------------------------------------------
No longer used.
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
*/

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

	envStr, ok := os.LookupEnv("ds")
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
			fmt.Fprintln(os.Stderr, " Error from os.Lstat ", err)
			continue
		}
		fi = append(fi, L)
	}
	return fi
} // MyReadDir

// ----------------------------- getMagnitudeString -------------------------------

func getMagnitudeString(j int64) (string, ct.Color) {
	// For this version in the 3 column display, I'm removing the extra space from the format strings.
	var s1 string
	var f float64
	var color ct.Color
	switch {
	case j > 1_000_000_000_000: // 1 trillion, or TB
		f = float64(j) / 1000000000000
		s1 = fmt.Sprintf("%.3g TB", f)
		color = ct.Red
	case j > 1_000_000_000: // 1 billion, or GB
		f = float64(j) / 1000000000
		s1 = fmt.Sprintf("%.3g gb", f)
		color = ct.White
	case j > 1_000_000: // 1 million, or MB
		f = float64(j) / 1000000
		s1 = fmt.Sprintf("%.3g mb", f)
		color = ct.Yellow
	case j > 1000: // KB
		f = float64(j) / 1000
		s1 = fmt.Sprintf("%.3g kb", f)
		color = ct.Cyan
	default:
		s1 = fmt.Sprintf("%3d bytes", j)
		color = ct.Green
	}
	return s1, color
}

/*
{{{
// --------------------------------------------------- truncStr -------------------------------------------
func truncStr(s string, w int) string {
	if w <= 0 || len(s) < w {
		return s
	}
	return s[:w]
} // end truncStr
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
