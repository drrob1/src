// getdir.go -- get directory to sort

package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
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

const LastAltered = "27 Feb 2021"

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
26 Feb 21 -- Now called getdir, as in get directory to sort.  And will use go1.16 new way of getting a directory
*/

// FIS is a FileInfo slice, as in os.FileInfo
type FISlice []os.FileInfo
type dirAliasMapType map[string]string

type DsrtParamType struct {
	numlines                                                        int
	reverseflag, sizeflag, dirlistflag, filenamelistflag, totalflag bool
}

func main() {
	const defaultlineswin = 50
	const defaultlineslinux = 40
	var dsrtparam DsrtParamType
	var numoflines int
	var userptr *user.User // from os/user
	//var files FISlice  I'm deleting this for now so I can fix the code to use the new DirEntry.  I may need to put it back.
	var direntries []os.DirEntry
	var err error
	var count int
	var SizeTotal, GrandTotal int64
	var GrandTotalCount int
	var havefiles bool
	var commandline string
	var directoryAliasesMap dirAliasMapType

	uid := 0
	gid := 0
	systemStr := ""

	// environment variable processing.  If present, these will be the defaults.
	dsrtparam = ProcessEnvironString() // This is a function below.

	linuxflag := runtime.GOOS == "linux"
	winflag := runtime.GOOS == "windows"
	if linuxflag {
		systemStr = "Linux"
		direntries = make([]os.DirEntry, 0, 500)
		//files = make([]os.FileInfo, 0, 500)
		if dsrtparam.numlines > 0 {
			numoflines = dsrtparam.numlines
		} else {
			numoflines = defaultlineslinux
		}
	} else if winflag {
		systemStr = "Windows"
		if dsrtparam.numlines > 0 {
			numoflines = dsrtparam.numlines
		} else {
			numoflines = defaultlineswin
		}
	} else {
		systemStr = "Unknown"
		numoflines = defaultlineslinux
	}

	sepstring := string(filepath.Separator)
	HomeDirStr, e := os.UserHomeDir() // HomeDir code used for processing ~ symbol meaning home directory.
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + sepstring

	if runtime.GOARCH == "amd64" {
		uid = os.Getuid() // int
		gid = os.Getgid() // int
		userptr, err = user.Current()
		if err != nil {
			fmt.Println(" user.Current error is ", err, "Exiting.")
			os.Exit(1)
		}
		HomeDirStr = userptr.HomeDir + sepstring
	}

	// flag definitions and processing
	revflag := flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr
	var RevFlag bool
	flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nlines = flag.Int("n", numoflines, "number of lines to display") // Ptr
	var NLines int
	flag.IntVar(&NLines, "N", numoflines, "number of lines to display") // Value

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

	var testFlag = flag.Bool("test", false, "enter a testing mode to println more variables")

	var longflag = flag.Bool("l", false, "long file size format.") // Ptr

	var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	var excludeFlag = flag.Bool("x", false, "exclude regex entered after prompt")

	flag.Parse()

	ctfmt.Println(ct.Blue, winflag, " dsrt will display Directory SoRTed by date or size.  Written in Go.  LastAltered ", LastAltered)
	if *testFlag {
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println()
		if runtime.GOARCH == "amd64" {
			fmt.Printf("uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
				uid, gid, systemStr, userptr.Uid, userptr.Gid, userptr.Username, userptr.Name, userptr.HomeDir)
		}
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
	if *nlines != numoflines {
		NumLines = *nlines
	} else if NLines != numoflines {
		NumLines = NLines
	}

	noExtensionFlag := *extensionflag || *extflag
	excludeRegexPattern := ""
	var excludeRegex *regexp.Regexp

	if *excludeFlag {
		ctfmt.Print(ct.Cyan, winflag, " Enter regex pattern to be excluded: ")
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

	//	CleanDirName := "." + string(filepath.Separator)  commented out 9/9/19
	CleanDirName := ""
	CleanFileName := ""
	filenamesStringSlice := flag.Args() // Intended to process linux command line filenames.
	//	fmt.Println(" filenames on command line",filenamesStringSlice)
	//  fmt.Println(" linuxflag =",linuxflag,", length filenamesstringslice =", len(filenamesStringSlice))

	// set which sort function will be in the sortfcn var
	sortfcn := func(i, j int) bool { return false }
	if SizeSort && Forward { // set the value of sortfcn so only a single line is needed to execute the sort.
		sortfcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method
			fi, e1 := direntries[i].Info()

			fj, e2 := direntries[j].Info()
			if e1 != nil || e2 != nil {
				fmt.Fprintln(os.Stderr, e)
				fmt.Fprintln(os.Stderr, "Cannot get size from fileinfo.  Aborting.")
				os.Exit(1)
			}

			return fi.Size() > fj.Size() // I want a largest first sort
		}
		if *testFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			fi, e1 := direntries[i].Info()

			fj, e2 := direntries[j].Info()
			if e1 != nil || e2 != nil {
				fmt.Fprintln(os.Stderr, e)
				fmt.Fprintln(os.Stderr, "Cannot get size from fileinfo.  Aborting.")
				os.Exit(1)
			}

			return fi.ModTime().After(fj.ModTime()) // I want a newest-first sort.  Changed 12/20/20
		}
		if *testFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			fi, e1 := direntries[i].Info()

			fj, e2 := direntries[j].Info()
			if e1 != nil || e2 != nil {
				fmt.Fprintln(os.Stderr, e)
				fmt.Fprintln(os.Stderr, "Cannot get size from fileinfo.  Aborting.")
				os.Exit(1)
			}

			return fi.Size() < fj.Size() // I want an smallest-first sort
		}
		if *testFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortfcn = func(i, j int) bool { // this is a closure anonymous function
			fi, e1 := direntries[i].Info()

			fj, e2 := direntries[j].Info()
			if e1 != nil || e2 != nil {
				fmt.Fprintln(os.Stderr, e)
				fmt.Fprintln(os.Stderr, "Cannot get size from fileinfo.  Aborting.")
				os.Exit(1)
			}

			return fi.ModTime().Before(fj.ModTime()) // I want an oldest-first sort
		}
		if *testFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if linuxflag && len(filenamesStringSlice) > 0 { // linux command line processing for filenames.  This condition had to be fixed July 4, 2019.
		paramIsDir := false
		if len(filenamesStringSlice) == 1 {
			// need to determine if the 1 param on command line is a directory
			//fi, err := os.Lstat(filenamesStringSlice[0])
			direntry, err := os.ReadDir(filenamesStringSlice[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, err, "; after ReadDir call for only one param.")
				os.Exit(1)
			}
			if len(direntry) > 1 {
				fmt.Fprintln(os.Stderr, " expecting only 1 direntry, but len(direntry) is", len(direntry), ".  Don't know what this means yet.")
			}
			paramIsDir = direntry[0].IsDir()
			if *testFlag {
				fmt.Println(" have only 1 param on line. filenameStringSlice=", filenamesStringSlice[0], "paramIsDir=", paramIsDir)
				fmt.Println()
			}
			if paramIsDir {
				CleanDirName = direntry[0].Name()
			} else { // not a directory so this one file needs to be displayed.
				direntries = append(direntries, direntry[0])
				havefiles = true
			}
		} else { // bash has placed more than one file in the command line, ie, len(filenameStringSlice) > 1
			for _, fn := range filenamesStringSlice { // fill a slice of DirEntry type given filename.
				direntry, err := os.ReadDir(fn)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				if len(direntry) > 1 {
					fmt.Fprintln(os.Stderr, " expecting only 1 direntry, but len(direntry) is", len(direntry), ".  Don't know what this means yet.")
				}
				direntries = append(direntries, direntry[0])
				if direntry[0].Type().IsRegular() && ShowGrandTotal {
					fi, err := direntry[0].Info()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					GrandTotal += fi.Size()
					GrandTotalCount++
				}
			}
			sort.Slice(direntries, sortfcn)
			havefiles = true
		} // end if filenamestringslice == 1

	} else { // either no params were present on the command line or this is running under Windows and may have a command line param.
		// commandline = filenamesStringSlice[0] -- this panics if there are no params on the line.
		commandline = flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.
	} // end if linuxflag and len(filenamestringslice > 0

	//	fmt.Println(" havefiles = ", havefiles)
	if winflag && len(commandline) > 0 { // added the winflag check so don't have to scan commandline on linux, which would be wasteful.
		if strings.ContainsRune(commandline, ':') {
			commandline = ProcessDirectoryAliases(directoryAliasesMap, commandline)
		} else if strings.Contains(commandline, "~") { // this can only contain a ~ on Windows.
			commandline = strings.Replace(commandline, "~", HomeDirStr, 1)
		}
		CleanDirName, CleanFileName = filepath.Split(commandline)
		CleanDirName = filepath.Clean(CleanDirName)
		CleanFileName = strings.ToUpper(CleanFileName)
	}

	if len(CleanDirName) == 0 {
		workingdir, _ := os.Getwd()
		if *testFlag {
			fmt.Println(" CleanDirName is empty, and will be ", workingdir)
		}
		CleanDirName = workingdir
	}

	if len(CleanFileName) == 0 {
		CleanFileName = "*"
	}

	if !havefiles {
		//files, err = ioutil.ReadDir(CleanDirName)  Go 1.16 is deprecating ioutil functions.
		direntries, err = os.ReadDir(CleanDirName)
		if err != nil { // It seems that ioutil.ReadDir itself stops when it gets an error of any kind, and I cannot change that.  Don't yet know about os.ReadDir.
			fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
			direntries = MyReadDir(CleanDirName)
		}
		if ShowGrandTotal {
			for _, d := range direntries {
				fi, err := d.Info()
				if err != nil {
					fmt.Fprintln(os.Stderr, err, ".  No idea why this error occurred.")
				}
				if fi.Mode().IsRegular() {
					GrandTotal += fi.Size()
					GrandTotalCount++
				}
			}
		}

		sort.Slice(direntries, sortfcn)
	}

	fmt.Println(" Dirname is", CleanDirName)

	// I need to add a description of how this code works, because I forgot.
	// The entire contents of the directory is read in by either ioutil.ReadDir or MyReadDir.  Then the slice of fileinfo's is sorted, and finally only the matching filenames are displayed.
	// This is still the way it works for Windows.
	// On linux, bash populated the command line by globbing, or no command line params were entered
	if linuxflag {
		for _, d := range direntries {
			f, err := d.Info()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			s := f.ModTime().Format("Jan-02-2006_15:04:05")
			sizestr := ""
			usernameStr, groupnameStr := GetUserGroupStr(f) // platform specific code
			if FilenameList && d.Type().IsRegular() {       //  && f.Mode().IsRegular() was condition here in dsrt code.
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
				if showthis {
					if LongFileSizeList {
						sizestr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
						if f.Size() > 100000 {
							sizestr = AddCommas(sizestr)
						}
						ctfmt.Printf(ct.Yellow, false, "%10v %s:%s %16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
					} else {
						var color ct.Color
						sizestr, color = getMagnitudeString(f.Size())
						ctfmt.Printf(color, false, "%10v %s:%s %-16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
					}
					count++
				}
			} else if IsSymlink(d.Type()) { // f.Mode() was condition here in dsrt code
				fmt.Printf("%10v %s:%s %16s %s <%s>\n", d.Type(), usernameStr, groupnameStr, sizestr, s, f.Name()) // f.Mode() was first exprn in dsrt code.
				count++
			} else if Dirlist && d.IsDir() { // f.IsDir() was condition here in dsrt code
				fmt.Printf("%10v %s:%s %16s %s (%s)\n", d.Type(), usernameStr, groupnameStr, sizestr, s, f.Name()) // f.Mode() was first exprn in dsrt code.
				count++
			}
			if count >= NumLines {
				break
			}
		} // end for range direntries
	} else if winflag {
		for _, d := range direntries {
			showthis := false
			NAME := strings.ToUpper(d.Name())
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
			}

			f, err := d.Info()
			if err != nil {
				fmt.Fprintln(os.Stderr, err, ".  No idea why this caused an error.")
			}

			if showthis {
				s := f.ModTime().Format("Jan-02-2006_15:04:05")
				sizestr := ""
				if FilenameList && d.Type().IsRegular() { // f.Mode().IsRegular() was condition here in dsrt code.
					SizeTotal += f.Size()
					if LongFileSizeList {
						sizestr = strconv.FormatInt(f.Size(), 10)
						if f.Size() > 100000 {
							sizestr = AddCommas(sizestr)
						}
						fmt.Printf("%17s %s %s\n", sizestr, s, d.Name()) // f.Name() was last exprn here in dsrt code.
					} else {
						var color ct.Color
						sizestr, color = getMagnitudeString(f.Size())
						ctfmt.Printf(color, true, "%-17s %s %s\n", sizestr, s, d.Name()) // f.Name() was last exprn here in dsrt code.
					}
					count++
				} else if IsSymlink(f.Mode()) {
					fmt.Printf("%17s %s <%s>\n", sizestr, s, d.Name()) // f.Name() was last exprn here in dsrt code.
					count++
				} else if Dirlist && f.IsDir() {
					fmt.Printf("%17s %s (%s)\n", sizestr, s, d.Name()) // f.Name() was last exprn here in dsrt code.
					count++
				}
				if count >= NumLines {
					break
				}
			}
		} // end for range direntries
	} // end if linuxflag else if winflag

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
} // end main getdir

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
		panic("uid not found")
	}

	idname := ptrToUser.Username
	return idname

} // GetIDname

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

	/* non-idiomatic code
	s := os.Getenv("diraliases")
	if len(s) == 0 {
		return nil
	}
	*/

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
func MyReadDir(dir string) []os.DirEntry {

	dirname, err := os.Open(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}

	direntries := make([]os.DirEntry, 0, len(names))
	for _, name := range names {
		d, err := os.ReadDir(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, " Error from os.ReadDir. ", err)
			continue
		}
		if len(d) > 1 {
			fmt.Fprintln(os.Stderr, " expected len(d) == 1, but it's", len(d), ", which I don't yet understand.")
		}
		direntries = append(direntries, d[0])
	}
	return direntries
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
