// testfilemodes.go -- directoy sort

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const LastAltered = "24 June 2019"

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
  20 Jun 19 -- Changed logic so that symlinks to files are always displayed, like files.  Nevermind.
               It did not do what I wanted.  I need an IsSymlink function to do what I want.
  20 Jun 19 -- Based on dsrt code, now I want to test an IsSymlink function.
  21 Jun 19 -- And testing isSymlink based on extending a non-local type.
  24 Jun 19 -- Changed to use Lstat instead of Stat when command line has a globbed file list, which only happens in Linux.
*/

// FIS is a FileInfo slice, as in os.FileInfo
type FISlice []os.FileInfo
type FISliceDate []os.FileInfo // inexperienced way to sort on more than one criterion
type FISliceSize []os.FileInfo // having compatible types only differing in the sort criteria
type dirAliasMapType map[string]string

func (f FISliceDate) Less(i, j int) bool {
	return f[i].ModTime().UnixNano() > f[j].ModTime().UnixNano() // I want a reverse sort, newest first
}

func (f FISliceDate) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FISliceDate) Len() int {
	return len(f)
}

func (f FISliceSize) Less(i, j int) bool {
	return f[i].Size() > f[j].Size() // I want a reverse sort, largest first
}

func (f FISliceSize) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FISliceSize) Len() int {
	return len(f)
}

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
	var files FISlice
	var filesDate FISliceDate
	var filesSize FISliceSize
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
		files = make([]os.FileInfo, 0, 500)
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
	HomeDirStr := "" // HomeDir code used for processing ~ symbol meaning home directory.
	if runtime.GOARCH == "amd64" {
		uid = os.Getuid() // int
		gid = os.Getgid() // int
		userptr, err = user.Current()
		if err != nil {
			fmt.Println(" user.Current error is ", err, "Exiting.")
			os.Exit(1)
		}
		HomeDirStr = userptr.HomeDir + sepstring
	} else if linuxflag {
		HomeDirStr = os.Getenv("HOME") + sepstring
	} else if winflag {
		HomeDirStr = os.Getenv("HOMEPATH") + sepstring
	} else { // unknown system
		fmt.Println(" Program not designed for this architecture.  Maybe it will work, maybe not.  Good luck.")
	}

	// flag definitions and processing
	var revflag = flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr

	var RevFlag bool
	flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nlines = flag.Int("n", numoflines, "number of lines to display") // Ptr

	var NLines int
	flag.IntVar(&NLines, "N", numoflines, "number of lines to display") // Value

	var helpflag = flag.Bool("h", false, "print help message") // pointer
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "H", false, "print help message")

	var sizeflag = flag.Bool("s", false, "sort by size instead of by date") // pointer
	var SizeFlag bool
	flag.BoolVar(&SizeFlag, "S", false, "sort by size instead of by date")

	var DirListFlag = flag.Bool("d", false, "include directories in the output listing") // pointer
	var FilenameListFlag bool
	flag.BoolVar(&FilenameListFlag, "D", false, "Directories only in the output listing")

	var TotalFlag = flag.Bool("t", false, "include grand total of directory")

	flag.Parse()

	fmt.Println(" dsrt will display sorted by date or size.  Written in Go.  LastAltered ", LastAltered)
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006, 15:04:05 MST")
	fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fmt.Println()

	if *helpflag || HelpFlag {
		fmt.Println(" Reads from dsrt environment variable before processing commandline switches.")
		fmt.Println("reads from diraliases environment variable if needed on Windows.")
		flag.PrintDefaults()
		if runtime.GOARCH == "amd64" {
			fmt.Printf("uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
				uid, gid, systemStr, userptr.Uid, userptr.Gid, userptr.Username, userptr.Name, userptr.HomeDir)
		}
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

	Dirlist := *DirListFlag || FilenameListFlag || dsrtparam.dirlistflag || dsrtparam.filenamelistflag // if -D entered then this expression also needs to be true.
	FilenameList := !(FilenameListFlag || dsrtparam.filenamelistflag)                                  // need to reverse the flag because D means suppress the output of filenames.

	ShowGrandTotal := *TotalFlag || dsrtparam.totalflag // added 09/12/2018 12:32:23 PM

	CleanDirName := "." + string(filepath.Separator)
	CleanFileName := ""
	filenamesStringSlice := flag.Args() // Intended to process linux command line filenames.
	if len(filenamesStringSlice) > 1 {  // linux command line processing for filenames.
		for _, s := range filenamesStringSlice { // fill a slice of fileinfo
			fi, err := os.Lstat(s)
			if err != nil {
				log.Fatal(err)
			}
			files = append(files, fi)
			if fi.Mode().IsRegular() && ShowGrandTotal {
				GrandTotal += fi.Size()
				GrandTotalCount++
			}
		}
		if SizeSort && Forward {
			largestSize := func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
				return files[i].Size() > files[j].Size() // I want a largest first sort
			}
			sort.Slice(files, largestSize)
		} else if DateSort && Forward {
			newestDate := func(i, j int) bool { // this is a closure anonymous function
				return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest first sort
			}
			sort.Slice(files, newestDate)
		} else if SizeSort && Reverse {
			smallestSize := func(i, j int) bool { // this is a closure anonymous function
				return files[i].Size() < files[j].Size() // I want a smallest first sort
			}
			sort.Slice(files, smallestSize)
		} else if DateSort && Reverse {
			oldestDate := func(i, j int) bool { // this is a closure anonymous function
				return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest first sort
			}
			sort.Slice(files, oldestDate)
		}
		havefiles = true
	} else {
		// Inelegant after adding linux filenames on command line code.  Could have now used filenameStringSlice[0].  I chose to not change the use of flag.Arg(0).
		// commandline = filenamesStringSlice[0] // this panics if there are no params on the line.
		commandline = flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.
		if strings.ContainsRune(commandline, ':') {
			commandline = ProcessDirectoryAliases(directoryAliasesMap, commandline)
		}
	}
	if len(commandline) > 0 {
		if strings.Contains(commandline, "~") { // this can only contain a ~ on Windows.
			commandline = strings.Replace(commandline, "~", HomeDirStr, 1)
		}
		CleanDirName, CleanFileName = filepath.Split(commandline)
		CleanDirName = filepath.Clean(CleanDirName)
		CleanFileName = strings.ToUpper(CleanFileName)
	}

	if len(CleanDirName) == 0 {
		CleanDirName = "." + string(filepath.Separator)
	}

	if len(CleanFileName) == 0 {
		CleanFileName = "*"
	}

	if SizeSort && !havefiles {
		filesSize, err = ioutil.ReadDir(CleanDirName)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range filesSize {
			if f.Mode().IsRegular() && ShowGrandTotal {
				GrandTotal += f.Size()
				GrandTotalCount++
			}
		}
		if Reverse {
			sort.Sort(sort.Reverse(filesSize))
		} else {
			sort.Sort(filesSize)
		}
		files = FISlice(filesSize)
	} else if !havefiles {
		filesDate, err = ioutil.ReadDir(CleanDirName)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range filesDate {
			if f.Mode().IsRegular() && ShowGrandTotal {
				GrandTotal += f.Size()
				GrandTotalCount++
			}
		}
		if Reverse {
			sort.Sort(sort.Reverse(filesDate))
		} else {
			sort.Sort(filesDate)
		}
		files = FISlice(filesDate)
	}

	fmt.Println(" Dirname is", CleanDirName)

	// I need to add a description of how this code works, because I forgot.
	// The entire contents of the directory is read in by the ioutil.ReadDir.  Then the slice of fileinfo's is sorted, and finally only the matching filenames are displayed.
	// This is still the way it works for Windows.
	// On linux, the matching pattern is set to be a *, so all entries match, depending on regular file or directory options selected, unless bash populated the command line by globbing.
	for _, f := range files {
		NAME := strings.ToUpper(f.Name())
		if BOOL, _ := filepath.Match(CleanFileName, NAME); BOOL {
			s := f.ModTime().Format("Jan-02-2006, 15:04:05") // added a comma in the format string.
			sizeint := 0
			sizestr := ""
			if f.Mode().IsRegular() { // only sum regular files, not dir or symlink entries.
				SizeTotal += f.Size()
				sizeint = int(f.Size())
				sizestr = strconv.Itoa(sizeint)
				if sizeint > 100000 {
					sizestr = AddCommas(sizestr)
				}
			}
			usernameStr, groupnameStr := GetUserGroupStr(f) // util function in platform specific code, only for linux and windows.  Probably won't compile for foreign computer.
			symlinkflag := IsSymlink(f.Mode())
			irregflag := IsIrregular(f.Mode())
			var Field MyFileMode // extending a non-local type by adding a new method.
			Field.field = f.Mode()
			SymLinkFlag := Field.isSymlink()

			if linuxflag {
				if Dirlist && f.IsDir() {
					fmt.Printf("%10v %s %15s %s (%s)", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
					count++
				} else if FilenameList && f.Mode().IsRegular() {
					fmt.Printf("%10v %s %15s %s %s", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
					count++
				} else if Field.field = f.Mode(); Field.isSymlink() { // assume it's a symlink as it's not a directory and not a regular file
					fmt.Printf("%10v %s %15s %s <%s>", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
					count++
				}
			} else if winflag {
				if Dirlist && f.IsDir() {
					fmt.Printf("%10v %s %15s %s (%s)", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
					count++
				} else if FilenameList && f.Mode().IsRegular() {
					fmt.Printf("%10v %s %15s %s %s", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
					count++
				} else if Field.field = f.Mode(); Field.isSymlink() { // added 6/19/19.  Prior to then, this code could not show a symlink on Windows.
					fmt.Printf("%10v %s %15s %s <%s>", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
					count++
				}
			} else { // maybe darwin, maybe Rasberry Pi, or maybe computer from the future
				fmt.Printf("%10v %s %15s %s %s  unknown OS or architecture", f.Mode(), f.Mode().String(), sizestr, s, f.Name())
				count++
			}

			// turns out that none of these are irregular.  So I'm not going to use this in dsrt.
			fmt.Printf(", Symlink: %t %t, Irregular: %t,    %s %s \n", symlinkflag, SymLinkFlag, irregflag, usernameStr, groupnameStr)
			if count >= NumLines {
				break
			}
		}
	}

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
		s1 := ""
		var i int64
		switch {
		case GrandTotal > 1000000000000: // 1 trillion, or TB
			i = GrandTotal / 1000000000000               // I'm forcing an integer division.
			if GrandTotal%1000000000000 > 500000000000 { // rounding up
				i++
			}
			s1 = fmt.Sprintf("%d TB", i)
		case GrandTotal > 1000000000: // 1 billion, or GB
			i = GrandTotal / 1000000000
			if GrandTotal%1000000000 > 500000000 { // rounding up
				i++
			}
			s1 = fmt.Sprintf("%d GB", i)
		case GrandTotal > 1000000: // 1 million, or MB
			i = GrandTotal / 1000000
			if GrandTotal%1000000 > 500000 {
				i++
			}
			s1 = fmt.Sprintf("%d MB", i)
		case GrandTotal > 1000: // KB
			i = GrandTotal / 1000
			if GrandTotal%1000 > 500 {
				i++
			}
			s1 = fmt.Sprintf("%d KB", i)
		default:
			s1 = fmt.Sprintf("%d", GrandTotal)
		}
		fmt.Println(", Directory grand total is", s0, "or approx", s1, "in", GrandTotalCount, "files.")
	} else {
		fmt.Println(".")
	}
} // end main testfilemodes

// ------------------------------ MyFileMode -------------------------
type MyFileMode struct { // extending a non-local type by adding a new method/function.
	field os.FileMode
}

func (m MyFileMode) isSymlink() bool { // this is the new method.  I decided to not use it in dsrt.  But it works.
	intermed := m.field & os.ModeSymlink
	result := intermed != 0
	return result
}

// ------------------------------ IsSymlink ---------------------------
func IsSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
}

// ------------------------------ IsIrregular -------------------------
func IsIrregular(m os.FileMode) bool {
	intermed := m & os.ModeIrregular
	result := intermed != 0
	return result
}

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//---------------------------------------------------------------------------------------------------

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

// ------------------------------------ ProcessEnvironString ---------------------------------------
func ProcessEnvironString() DsrtParamType { // use system utils when can because they tend to be faster
	var dsrtparam DsrtParamType

	s := os.Getenv("dsrt")

	if len(s) < 1 {
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
func GetDirectoryAliases() dirAliasMapType { // Env variable is diraliases.

	s := os.Getenv("diraliases")
	if len(s) == 0 {
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
	aliasesMap = GetDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := aliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathnFile := cmdline[idx+1:]
	completeValue := aliasValue + PathnFile
	fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
	return completeValue
}

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
