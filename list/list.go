package list

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 22 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                 This is going to take a while.
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                 I decided to only copy files if the new one is newer than the old one.
  22 Dec 22 -- Now I want to colorize the output, so I have to return the os.FileInfo also.  So I changed MakeList and NewList to not return []string, but return []FileInfoExType.
                 And myReadDir creates the relPath field that I added to FileInfoExType.
  25 Dec 22 -- Moved FileSection here.
  26 Dec 22 -- Changed test against the regexp to be nil instead of "".
  29 Dec 22 -- Adding the '.' to be a sentinel marker for the 1st param that's ignored.  This change is made in the platform-specific code.
  30 Dec 22 -- I'm thinking about being able to use environment strings to pass around flag values.  ListFilter, ListVerbose, ListVeryVerbose, ListReverse.
                 Nevermind.  I'll pass the variables globally, exported from here.  And I added a procedure New to not stutter, as in list.NewList.  But I kept the old NewList, for now.
   1 Jan 23 -- I changed the display colors for the list.  The line is not all the same color now.
   4 Jan 23 -- Adding screen clearing between screen displays.  Copied from rpng.
   6 Jan 23 -- Improving error handling, by having these functions here return an error variable.  This was needed to better handle the newly added stop code.
  15 Jan 23 -- Split off list2, which will have the code that takes an input regexp, etc, for copying.go.
   8 Feb 23 -- Combined the 2 init functions into one.  It was a mistake to have 2 of them.
  28 Feb 23 -- The field name called RelPath is a misnomer, as it's an absolute path.  I added a field name to reflect what it really is.  I'll leave the misnomer, for now.
  18 Mar 23 -- Thought I experienced a bug, but then I figured it out.  There's no bug here. :-)
  24 Mar 23 -- Added CheckDest after fixing issue in listutil_linux.go.  More details in listutil_linux.go
  31 Mar 23 -- StaticCheck found a few issues.
   1 Apr 23 -- delList not using the pattern on the commandline on Windows.  Investigating why.  Nevermind, I forgot that dellist is now under list2 and works differently.
   4 Apr 23 -- delList moved back here under list.go.  I added a flag, DelListFlag, so the last item on the linux command line will be included.
                 And I added FileSelectionString, which returns a string instead of the FileInfoExType.
   5 Apr 23 -- Fixed CheckDest(), ProcessDirectoryAliases and an issue in listutil_windows found by staticCheck.
   8 Apr 23 -- New now does not need params.  NewList will be the format that needs params.
  11 May 23 -- Adding replacement of digits 1..9 to mean a..i.
   1 Jun 23 -- Added getFileInfoXFromGlob, which behaves the same on Windows and linux.
   2 Jul 23 -- Made the FileSelection routines use "newlines" between iterations.  This way, I can use the scroll back buffer if I want to.
   8 Jul 23 -- In _windows part, I changed how the first param is tested for being a directory.
  12 Jul 23 -- Globbing isn't working.  Nevermind, I forgot about first param must be a dot if I'm going to use globbing.
  14 Jul 23 -- Now I'm exporting GetFileInfoXFromCommandLine, from platform-specific code.
  16 Jul 23 -- I'm thinking about adding GetFileInfoXFromRegexp.  And I'll need the corresponding rex flag for it.  And I'll need NewFromRex.
  25 Sep 23 -- There's a bug in runlist and runx in which the beginning of line anchor is not processed correctly.  I'm tracking this down now.
                 I found the bug.  I was matching against RelPath which includes the path dir info, so the ^ anchor is meaningless.
   7 Apr 24 -- Adding color to the display of choices, ie, alternating white and yellow.
                 Nevermind.  I already do something like this; the color of the filename is determined by the size of the file.
                 They only use the same color if they are all in the same size magnitude. I think I'm going to test if the color is yellow, then ... nevermind.
                 I'm going to alternate brightness, ie, bright is true or false, and see what happens.
                 I like it, so I'll keep it for now.  And I added it to list2.go.
  25 May 24 -- Adding doc comments for go doc.
  15 June 24-- On linux, searching /mnt/misc takes ~8 sec, but on Windows it only takes ~800 ms.  That's a huge difference.  It sounds like Windows is caching it but linux is not.
                 I want to use the new concurrent directory code that's in fdsrt, but only default to that on linux.  I don't yet know how to do that.
  18 June 24-- Made the change that has the concurrent routines used by FileInfoXFromGlob() and FileInfoXFromRegexp(), which are called by NewFromGlob() and NewFromRegexp(), respectively.
                 Fixed error message texts that have yet to be needed.
  19 June 24-- Clarifying how these routines are intended to work.  Client packages are to use any routine that begins w/ New.  IE, New(), NewFromGlob() or NewFromRegexp().
                 NewFromGlob() no longer uses filepath.Glob() routine.  That only persists in my glob and dsrt routines, as I also removed it from fdsrt and ds.  Rex never had it.
                 Currently, only runlist, runx and runlst use NewFromGlob() and NewFromRegexp().
  25 Oct 24 -- Fixed bug in CheckDest so that it will catch if no params are on the line.
  28 Nov 24 -- Will now display the size in the same color as the filename.
  28 Dec 24 -- While investigating to make changes to use concurrency, I see that I've already done this to populate the slice of dirEntries, mostly on linux.
				My testing did not find it to be faster on Windows.  Concurrency was slightly faster when a pattern had to be matched, but otherwise not.
				Hence, the concurrent routines are primarily used on linux.
  15 Jan 25 -- Adding filterStr here by copying the code already in the dsrt family of routines (dsrt, ds, rex, dv).
				Since I'm starting to use pflag and viper, I'm going to have to remove the use of flag.NArgs here.  I'll make it global?  Then I'll have to change lots of other code.
				I'll have to use both flag and pflag.
				The only way for flag and pflag to not clobber each other is to use the Parsed() function to determine which is active.
  19 Feb 25 -- Making "0" a synonym for "1" instead of a stop code.
   9 Mar 25 -- Suppressed display of error in FileSelection().  I don't need to display errors twice.
  18 May 25 -- Added unique function.
  20 Aug 25 -- I decided to include symlinks to be copied.  And I'll add a flag to only copy symlinks.
*/

var LastAltered = "Aug 20, 2025"

type DirAliasMapType map[string]string

// FileInfoExType is what is returned by all the routines here.  Fields are file info, Dir, RelPath, AbsPath and FullPath.  Some may be redundant, but this is what it is.
type FileInfoExType struct {
	FI       os.FileInfo
	Dir      string
	RelPath  string // this is a misnomer, but to not have to propagate the correction thru my code, I'll leave this here.
	AbsPath  string
	FullPath string // probably not needed, but I really do want to be complete.
}

var filterAmt int64 // not exported.  Don't remember why this is an int64 instead of just an int.
var VerboseFlag bool
var VeryVerboseFlag bool
var FilterFlag bool
var FilterStr string
var ReverseFlag bool
var GlobFlag bool
var SymFlag bool // only copy symlinks.

// FastDebugFlag is used for debugging the concurrent routines.
var FastDebugFlag bool
var fileInfoX []FileInfoExType
var clearMapOfFunc map[string]func()
var ExcludeRex *regexp.Regexp
var IncludeRex *regexp.Regexp
var InputDir string
var SizeFlag bool
var DelListFlag bool
var brightFlag bool // intended for use in the file selection routines so the brightness will alternate on and off.

//var directoryAliasesMap DirAliasMapType  Not needed anymore.

const defaultHeight = 40
const minWidth = 90
const minHeight = 26
const stopCode = "0"
const sepString = string(filepath.Separator)
const fetch = 1000    // used for the concurrency pattern in MyReadDirConccurent
const multiplier = 10 // used for the worker pool pattern in MyReadDirConcurrent
var numWorkers = runtime.NumCPU() * multiplier

var autoWidth, autoHeight int

func init() {
	var err error
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}
	_ = autoWidth
	clearMapOfFunc = make(map[string]func(), 2)
	clearMapOfFunc["linux"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("clearMapOfFunc")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clearMapOfFunc["windows"] = func() { // this is a closure, or an anonymous function
		comspec := os.Getenv("ComSpec")
		cmd := exec.Command(comspec, "/c", "cls") // this was calling cmd, but I'm trying to preserve the scrollback buffer.
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clearMapOfFunc["newlines"] = func() {
		fmt.Printf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
	}

}

// New does not need params and returns a slice of FileInfoExType and an error.  This is the idiomatic way to call the routine here.
func New() ([]FileInfoExType, error) {
	lst, err := MakeList(ExcludeRex, SizeFlag, ReverseFlag)
	return lst, err
}

// NewList needs these params (excludeMe *regexp.Regexp, sizeSort, reverse bool) and returns a slice of FileInfoExType, and error.
func NewList(excludeMe *regexp.Regexp, sizeSort, reverse bool) ([]FileInfoExType, error) {
	lst, err := MakeList(excludeMe, sizeSort, reverse)
	return lst, err
}

// MakeList needs the excludeRegex, sizeSort and reverse params, and returns a slice of FileInfoExType and error.  After writing this, I decided to use the idiomatic wrapper functions above.
func MakeList(excludeRegex *regexp.Regexp, sizeSort, reverse bool) ([]FileInfoExType, error) {
	var err error

	if FilterFlag {
		filterAmt = 1_000_000
	} else if FilterStr != "" {
		if len(FilterStr) > 1 { // If the character is a letter, it has to be k, m or g.  Or it's a number, but not both.  For now.
			amt, err := strconv.Atoi(FilterStr)
			filterAmt = int64(amt)
			if err != nil {
				fmt.Fprintln(os.Stderr, "converting filterStr:", err)
			}
		} else if unicode.IsLetter(rune(FilterStr[0])) {
			FilterStr = strings.ToLower(FilterStr)
			if FilterStr == "k" {
				filterAmt = 1000
			} else if FilterStr == "m" {
				filterAmt = 1_000_000
			} else if FilterStr == "g" {
				filterAmt = 1_000_000_000
			} else {
				fmt.Fprintln(os.Stderr, "filterStr is not valid and was ignored.  filterStr=", FilterStr)
			}
			FilterFlag = true
		} else {
			fmt.Fprintln(os.Stderr, "filterStr not valid.  filterStr =", FilterStr)
		}
	}

	if VeryVerboseFlag {
		VerboseFlag = true
	}

	if VerboseFlag {
		fmt.Printf(" in MakeList.  FilterFlag=%t, FilterStr=%s, filterAmt=%d\n", FilterFlag, FilterStr, filterAmt)
	}

	fileInfoX, err = GetFileInfoXFromCommandLine(excludeRegex)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from getFileInfoXFromCommandLine is %s.\n", err)
		return nil, err
	}
	fmt.Printf(" length of fileInfoX = %d\n", len(fileInfoX))

	// set which sort function will be in the sortfcn var
	forward := !(reverse || ReverseFlag)
	dateSort := !sizeSort
	sortFcn := func(i, j int) bool { return false }
	if sizeSort && forward { // set the value of sortFcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfoX[i].FI.Size() > fileInfoX[j].FI.Size() // I want a largest first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if dateSort && forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//       return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfoX[i].FI.ModTime().After(fileInfoX[j].FI.ModTime()) // I want a newest-first sort.
		}
		if VerboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if sizeSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfoX[i].FI.Size() < fileInfoX[j].FI.Size() // I want a smallest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if dateSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfoX[i].FI.ModTime().Before(fileInfoX[j].FI.ModTime()) // I want an oldest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if len(fileInfoX) > 1 {
		sort.Slice(fileInfoX, sortFcn) // sort functions became available as of Go 1.8
	}

	return fileInfoX, nil
} // end MakeList

// SkipFirstNewList will return a slice of FileInfoExType and an error.  I don't remember why I coded this.
func SkipFirstNewList() ([]FileInfoExType, error) {
	var err error

	sizeSort := SizeFlag   // passed globally
	reverse := ReverseFlag // passed globally

	if FilterFlag {
		filterAmt = 1_000_000
	}
	if VeryVerboseFlag {
		VerboseFlag = true
	}

	fileInfoX, err = getFileInfoXSkipFirstOnCommandLine() // this needs ExcludeRex, passed globally.
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from getFileInfoXFromCommandLine is %s.\n", err)
		return nil, err
	}
	fmt.Printf(" length of fileInfoX = %d\n", len(fileInfoX))

	// set which sort function will be in the sortfcn var
	forward := !(reverse || ReverseFlag)
	dateSort := !sizeSort
	sortFcn := func(i, j int) bool { return false }
	if sizeSort && forward { // set the value of sortFcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfoX[i].FI.Size() > fileInfoX[j].FI.Size() // I want a largest first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if dateSort && forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//       return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfoX[i].FI.ModTime().After(fileInfoX[j].FI.ModTime()) // I want a newest-first sort.
		}
		if VerboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if sizeSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfoX[i].FI.Size() < fileInfoX[j].FI.Size() // I want a smallest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if dateSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfoX[i].FI.ModTime().Before(fileInfoX[j].FI.ModTime()) // I want an oldest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if len(fileInfoX) > 1 {
		sort.Slice(fileInfoX, sortFcn) // sort functions became available as of Go 1.8
	}

	return fileInfoX, nil
} // end SkipFirstNewList

// NewFromGlob takes a glob expression and returns a slice of FileInfoExType, and an error.
func NewFromGlob(globExpr string) ([]FileInfoExType, error) {
	var err error

	sizeSort := SizeFlag   // passed globally
	reverse := ReverseFlag // passed globally

	if FilterFlag {
		filterAmt = 1_000_000
	}
	if VeryVerboseFlag {
		VerboseFlag = true
	}

	fileInfoX, err = FileInfoXFromGlob(globExpr)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from getFileInfoXFromGlob is %s.\n", err)
		return nil, err
	}
	fmt.Printf(" length of fileInfoX = %d\n", len(fileInfoX))

	// set which sort function will be in the sortfcn var
	forward := !(reverse || ReverseFlag)
	dateSort := !sizeSort
	sortFcn := func(i, j int) bool { return false }
	if sizeSort && forward { // set the value of sortFcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfoX[i].FI.Size() > fileInfoX[j].FI.Size() // I want a largest first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if dateSort && forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//       return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfoX[i].FI.ModTime().After(fileInfoX[j].FI.ModTime()) // I want a newest-first sort.
		}
		if VerboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if sizeSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfoX[i].FI.Size() < fileInfoX[j].FI.Size() // I want a smallest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if dateSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfoX[i].FI.ModTime().Before(fileInfoX[j].FI.ModTime()) // I want an oldest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if len(fileInfoX) > 1 {
		sort.Slice(fileInfoX, sortFcn) // sort functions became available as of Go 1.8
	}

	return fileInfoX, nil
} // end NewFromGlob

// NewFromRegexp takes a regexp and returns a slice of FileInfoExType and an error.
func NewFromRegexp(rex *regexp.Regexp) ([]FileInfoExType, error) { // remember that the caller must call regexp.Compile
	var err error

	sizeSort := SizeFlag   // passed globally
	reverse := ReverseFlag // passed globally

	if FilterFlag {
		filterAmt = 1_000_000
	}
	if VeryVerboseFlag {
		VerboseFlag = true
	}

	fileInfoX, err = FileInfoXFromRegexp(rex)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from getFileInfoXFromRegexp is %s.\n", err)
		return nil, err
	}
	fmt.Printf(" length of fileInfoX = %d\n", len(fileInfoX))

	// set which sort function will be in the sortfcn var
	forward := !(reverse || ReverseFlag)
	dateSort := !sizeSort
	sortFcn := func(i, j int) bool { return false }
	if sizeSort && forward { // set the value of sortFcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfoX[i].FI.Size() > fileInfoX[j].FI.Size() // I want a largest first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if dateSort && forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//       return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfoX[i].FI.ModTime().After(fileInfoX[j].FI.ModTime()) // I want a newest-first sort.
		}
		if VerboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if sizeSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfoX[i].FI.Size() < fileInfoX[j].FI.Size() // I want a smallest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if dateSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfoX[i].FI.ModTime().Before(fileInfoX[j].FI.ModTime()) // I want an oldest-first sort
		}
		if VerboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if len(fileInfoX) > 1 {
		sort.Slice(fileInfoX, sortFcn) // sort functions became available as of Go 1.8
	}

	return fileInfoX, nil
} // end NewFromRex

// MyReadDir -- single routine version
func MyReadDir(dir string, excludeMe *regexp.Regexp) ([]FileInfoExType, error) { // The entire change including use of []DirEntry happens here.  Not concurrent.
	dirEntries, err := os.ReadDir(dir) // this function doesn't need to be closed.
	if err != nil {
		return nil, err
	}

	fileInfoExs := make([]FileInfoExType, 0, len(dirEntries))
	for _, d := range dirEntries {
		fi, e := d.Info()
		if e != nil {
			fmt.Fprintf(os.Stderr, " Error from %s.Info() is %v\n", d.Name(), e)
			continue
		}
		if includeThis(fi, excludeMe) {
			joinedFilename := filepath.Join(dir, fi.Name())
			fix := FileInfoExType{ // fix is a file info extended var
				FI:       fi,
				Dir:      dir,
				RelPath:  joinedFilename, // this is a misnomer, but to not have to propagate the correction thru my code, I'll leave this here.
				AbsPath:  joinedFilename,
				FullPath: joinedFilename,
			}
			fileInfoExs = append(fileInfoExs, fix)
		}
	}
	return fileInfoExs, nil
} // myReadDir

// includeThis -- Using the excludeRex, FilterFlag and if a regular file, determines of this FileInfo is to be included in the slice that's being built.
func includeThis(fi os.FileInfo, excludeRex *regexp.Regexp) bool { // this already has matched the include expression
	if VeryVerboseFlag {
		fmt.Printf(" includeThis.  FI=%#v, FilterFlag=%t\n", fi, FilterFlag)
	}
	//if !fi.Mode().IsRegular() {  removed Aug 20, 2025 because I want to include symlinks.
	//	return false
	//}
	if fi.IsDir() {
		return false
	}

	// symflag means to only include symlinks.
	if fi.Mode().IsRegular() && SymFlag { // skip regular files if symflag is set.
		return false
	}

	// if get here, then it's either a regular file or a symlink.  I want to include both.

	if excludeRex != nil {
		if BOOL := excludeRex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	if FilterFlag {
		if fi.Size() < filterAmt {
			return false
		}
	}
	return true
}

//------------------------------ GetDirectoryAliases ----------------------------------------

func GetDirectoryAliases() DirAliasMapType { // Env variable is diraliases.

	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = strings.ReplaceAll(s, "_", " ") // substitute the underscore, _, for a space so strings.Fields works correctly
	directoryAliasesMap := make(DirAliasMapType, 10)

	dirAliasSlice := strings.Fields(s)

	for _, aliasPair := range dirAliasSlice {
		if string(aliasPair[len(aliasPair)-1]) != "\\" {
			aliasPair = aliasPair + "\\"
		}
		aliasPair = strings.ReplaceAll(aliasPair, "-", " ") // substitute a dash,-, for a space
		splitAlias := strings.Fields(aliasPair)
		directoryAliasesMap[splitAlias[0]] = splitAlias[1]
	}
	return directoryAliasesMap
} // end getDirectoryAliases

// ------------------------------ ProcessDirectoryAliases ---------------------------

func ProcessDirectoryAliases(cmdline string) string {

	idx := strings.IndexRune(cmdline, ':')
	if VerboseFlag {
		fmt.Printf("In ProcessDirectoryAliases.  colon idx=%d\n", idx)
	}
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap := GetDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := aliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathNFile := cmdline[idx+1:]
	completeValue := aliasValue + PathNFile
	if VerboseFlag {
		fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
	}
	return completeValue
} // ProcessDirectoryAliases

// ReplaceDigits -- replaces the digits 1..9 with the letters a..z.  This is because I have a habit of hitting "1" for the first item I see, instead of "a".
func ReplaceDigits(in string) string {
	const fudgefactor = 'a' - '1'
	var sb strings.Builder
	for _, ch := range in {
		if ch >= '1' && ch <= '9' {
			ch = ch + fudgefactor
		} else if ch == '0' {
			ch += fudgefactor + 1 // as this is '0' and not '1' but I want it to behave as if a '1' was entered.
		}
		sb.WriteRune(ch)
	}
	return sb.String()
}

// ExpandADash -- expands the first instance of a dash that it finds
func ExpandADash(in string) (string, error) {

	if !strings.Contains(in, "-") {
		return in, nil
	}

	in = strings.ToLower(in)
	idx := strings.IndexRune(in, '-')
	begChar := in[idx-1]
	if idx+1 >= len(in) {
		return in, fmt.Errorf("no ending character found for substitution at position %d", idx)
	}
	endChar := in[idx+1]
	c := endChar - 'a'
	begPart := in[:idx-1]
	endPart := in[idx+2:]
	if c > 26 { // byte value can't be < 0
		return in, fmt.Errorf("invalid index found, idx=%d, endChar=%c", idx, endChar)
	}
	var sb strings.Builder
	for i := begChar - 'a'; i < endChar-'a'+1; i++ { // must include the endChar in the expanded string.
		err := sb.WriteByte(i + 'a')
		if err != nil {
			return in, err
		}
	}

	result := begPart + sb.String() + endPart
	return result, nil
}

// ExpandAllDashes -- expands all instances of a dash that it finds by calling ExpandADash as often as it needs to.
func ExpandAllDashes(in string) (string, error) {
	var workingStr = in
	var err error

	for strings.Contains(workingStr, "-") {
		workingStr, err = ExpandADash(workingStr)
		//fmt.Printf(" in ExpandAllDashes: out = %#v, err = %#v\n", workingStr, err)
		if err != nil {
			return workingStr, err
		}
	}

	return workingStr, nil
}

// unique -- makes sure that the slice only contains unique values.
func unique(in string) string {
	var sb strings.Builder
	for _, ch := range in {
		if !strings.ContainsRune(sb.String(), ch) {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// FileSelection -- This displays the files on screen and creates the list from what's entered by the user.
func FileSelection(inList []FileInfoExType) ([]FileInfoExType, error) {
	outList := make([]FileInfoExType, 0, len(inList))
	numOfLines := min(autoHeight, minHeight)
	numOfLines = min(numOfLines, len(inList))
	var beg, end int
	lenList := len(inList)
	var ans string

outerLoop:
	for {
		if lenList-beg >= numOfLines {
			end = beg + numOfLines
		} else {
			end = lenList
		}

		fList := inList[beg:end]

		for i, f := range fList {
			t := f.FI.ModTime().Format("Jan-02-2006_15:04:05") // t is a timestamp string.
			s, colr := GetMagnitudeString(f.FI.Size())
			brightFlag = i%2 == 0
			ctfmt.Printf(colr, brightFlag, " %c: %s ", i+'a', f.RelPath)
			clr := ct.White
			if clr == colr { // don't use same color as rest of the displayed string.
				clr = ct.Yellow
			}
			ctfmt.Printf(clr, brightFlag, " -- %s", t)
			//             ctfmt.Printf(ct.Cyan, brightFlag, " %s\n", s)  Removed 11/28/24
			ctfmt.Printf(colr, brightFlag, " %s\n", s) // added 11/28/24
		}

		fmt.Print(" Enter selections: ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "" // it seems that if I don't do this, the prev contents are not changed when I just hit <enter>
		}

		// Check for the stop code anywhere in the input.
		//if strings.Contains(ans, stopCode) { // this is a "0" at time of writing this comment.  I decided to remove the zero stop code.  I never used it.  A punctuation char is easier.
		//	e := fmt.Errorf("stopcode of %q found in input.  Stopping", stopCode)
		//	return nil, e
		//}

		// Here is where I can scan the ans string first replacing digits 1..9, and then looking for a-z and replace that with all the letters so indicated before
		// passing it onto the processing loop.
		// Upper case letter will mean something, not sure what yet.
		ans = ReplaceDigits(ans)
		expandedAns, err := ExpandAllDashes(ans)
		if err != nil {
			//fmt.Fprintf(os.Stderr, " ERROR from ExpandAllDashes(%s): %q\n", ans, err)  Don't need to display this twice.
			return nil, err
		}

		uniqueAns := unique(expandedAns) // added May 18, 2025

		for _, c := range uniqueAns { // parse the answer character by character.  Well, really rune by rune but I'm ignoring that.
			idx := int(c - 'a')
			if idx < 0 || idx > minHeight || idx > (end-beg-1) { // entered character out of range, so complete.  IE, if enter a digit, xyz or a non-alphabetic character routine will return.
				break outerLoop
			}
			f := fList[c-'a']
			outList = append(outList, f)
		}
		if end >= lenList {
			break
		}
		beg = end

		//clearFunc := clearMapOfFunc[runtime.GOOS]
		clearFunc := clearMapOfFunc["newlines"]
		clearFunc()
	}

	return outList, nil
} // end FileSelection

// -------------------------------------------- FileSelectionString -------------------------------------------------------

func FileSelectionString(inList []FileInfoExType) ([]string, error) {
	outStrList := make([]string, 0, len(inList))
	numOfLines := min(autoHeight, minHeight)
	numOfLines = min(numOfLines, len(inList))
	var beg, end int
	lenList := len(inList)
	var ans string

outerLoop:
	for {
		if lenList-beg >= numOfLines {
			end = beg + numOfLines
		} else {
			end = lenList
		}

		fList := inList[beg:end]

		for i, f := range fList {
			t := f.FI.ModTime().Format("Jan-02-2006_15:04:05") // t is a timestamp string.
			s, colr := GetMagnitudeString(f.FI.Size())
			brightFlag = i%2 == 0
			ctfmt.Printf(colr, brightFlag, " %c: %s ", i+'a', f.RelPath)
			clr := ct.White
			if clr == colr { // don't use same color as rest of the displayed string.
				clr = ct.Yellow
			}
			ctfmt.Printf(clr, brightFlag, " -- %s", t)
			ctfmt.Printf(ct.Cyan, brightFlag, " %s\n", s)
		}

		fmt.Print(" Enter selections: ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "" // it seems that if I don't do this, the prev contents are not changed when I just hit <enter>
		}

		// Check for the stop code anywhere in the input.
		if strings.Contains(ans, stopCode) {
			e := fmt.Errorf("stopcode of %q found in input.  Stopping", stopCode)
			return nil, e
		}

		// here is where I can scan the ans string looking for a-z and replace that with all the letters so indicated before passing it onto the processing loop.
		// ans = strings.ToLower(ans)  Upper case letter will mean something, not sure what yet.
		processedAns, err := ExpandAllDashes(ans)
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR from ExpandAllDashes(%s): %q\n", ans, err)
			return nil, err
		}
		for _, c := range processedAns { // parse the answer character by character.  Well, really rune by rune but I'm ignoring that.
			idx := int(c - 'a')
			if idx < 0 || idx > minHeight || idx > (end-beg-1) { // entered character out of range, so complete.  IE, if enter a digit, xyz or a non-alphabetic character routine will return.
				break outerLoop
			}
			f := fList[c-'a']
			outStrList = append(outStrList, f.AbsPath)
		}
		if end >= lenList {
			break
		}
		beg = end

		//clearFunc := clearMapOfFunc[runtime.GOOS]
		clearFunc := clearMapOfFunc["newlines"]
		clearFunc()
	}

	return outStrList, nil
} // end FileSelectionString

// ----------------------------- GetMagnitudeString -------------------------------

func GetMagnitudeString(j int64) (string, ct.Color) {
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

// ------------------------------------------------ CheckDest ------------------------------------------------------

func CheckDest() string {
	var nargs int
	var args []string
	if flag.Parsed() {
		nargs = flag.NArg()
		args = flag.Args()
	} else if pflag.Parsed() {
		nargs = pflag.NArg()
		args = pflag.Args()
	} else {
		fmt.Printf(" Error: neither flag.Parsed nor pflag.Parsed are true.  WTF?\n")
		return ""
	}

	if len(args) <= 1 {
		return ""
	}
	//d := flag.Arg(flag.NArg() - 1)
	d := args[nargs-1] // this is the last command line argument, which is intended as the destination for the commands that need this.
	if runtime.GOOS == "windows" {
		if strings.ContainsRune(d, ':') {
			//directoryAliasesMap := GetDirectoryAliases()  Doesn't belong here.  It's initialized in ProcessDirectoryAliases where it belongs.
			d = ProcessDirectoryAliases(d)
		} else if strings.Contains(d, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			d = strings.Replace(d, "~", homeDirStr, 1)
		}
	}

	if !strings.HasSuffix(d, sepString) {
		d = d + sepString
	}

	f, err := os.Open(d)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " ERROR from opening %s is %s\n", d, err)
		return ""
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		ctfmt.Printf(ct.Red, false, " ERROR from %s.Stat is %s\n", d, err)
		return ""
	}
	if !fi.IsDir() {
		fmt.Printf(" Last item on command line is %s which is not a directory.  Ignoring.\n", d)
		return ""
	}
	return d
}

// ----------------------------------------------------------------------------------------------------

// FileInfoXFromGlob behaves the same on linux and Windows, so it's here and not in platform specific code file.  Uses concurrent code to read the disk.
func FileInfoXFromGlob(globStr string) ([]FileInfoExType, error) { // Uses list.ExcludeRex, does NOT use filepath.Glob
	var fileInfoX []FileInfoExType
	//excludeMe := ExcludeRex

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = "."
	}
	HomeDirStr = HomeDirStr + sepString

	pattern := globStr
	if VerboseFlag {
		fmt.Printf(" file pattern is %s\n", pattern)
	}
	if pattern == "" {
		workingDir, er := os.Getwd()
		if er != nil {
			return nil, er
			//fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			//os.Exit(1)
		}
		fileInfoX, err = myReadDirConcurrent(workingDir)
		if err != nil {
			return nil, err
		}
	} else { // pattern is not blank
		if strings.ContainsRune(pattern, ':') {
			pattern = ProcessDirectoryAliases(pattern)
		}

		pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		dirNamePattern, fileNamePattern := filepath.Split(pattern)
		fileNamePattern = strings.ToLower(fileNamePattern)
		if dirNamePattern != "" && fileNamePattern == "" { // then have a dir pattern without a filename pattern
			fileInfoX, err = myReadDirConcurrent(dirNamePattern)
			return fileInfoX, err
		}
		if dirNamePattern == "" {
			dirNamePattern = "."
		}
		if fileNamePattern == "" { // need this to not be blank because of the call to Match below.
			fileNamePattern = "*"
		}

		if VerboseFlag {
			fmt.Printf(" dirName=%s, fileName=%s, pattern=%s \n", dirNamePattern, fileNamePattern, pattern)
		}

		//var filenames []string  removed when the concurrent code was developed for here.  That is, I removed the use of filepath.Glob().  I never used it anyway.
		//if GlobFlag {
		//	// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
		//	// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
		//	// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
		//	filenames, err = filepath.Glob(pattern)
		//	if VerboseFlag {
		//		fmt.Printf(" after glob: len(filenames)=%d, filenames=%v \n\n", len(filenames), filenames)
		//	}
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//} else {
		//	d, err := os.Open(dirName)
		//	if err != nil {
		//		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine os.Open is %v\n", err)
		//		os.Exit(1)
		//	}
		//	defer d.Close()
		//	filenames, err = d.Readdirnames(0) // I don't know if I have to make this slice first.  I'm going to assume not for now.
		//	if err != nil {                    // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
		//		fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
		//		fileInfoX, err = MyReadDir(dirName, excludeMe)
		//		return fileInfoX, err
		//	}
		//}

		//fileInfoX = make([]FileInfoExType, 0, len(filenames))
		//for _, f := range filenames { // basically I do this here because of a pattern to be matched.
		//	var path string
		//	if strings.Contains(f, sepString) {
		//		path = f
		//	} else {
		//		path = filepath.Join(dirName, f)
		//	}
		//
		//	fi, err := os.Lstat(path)
		//	if err != nil {
		//		fmt.Fprintf(os.Stderr, " Error from Lstat call on %s is %v\n", path, err)
		//		continue
		//	}
		//
		//	match, er := filepath.Match(strings.ToLower(fileName), strings.ToLower(f)) // redundant if glob is used, and glob is always used in this routine.
		//	if er != nil {
		//		fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern is %v.\n", pattern, er)
		//		continue
		//	}
		//
		//	if includeThis(fi, excludeMe) && match { // has to match pattern, size criteria and not match an exclude pattern.
		//		joinedFilename := filepath.Join(dirName, f)
		//		fix := FileInfoExType{
		//			FI:       fi,
		//			Dir:      dirName,
		//			RelPath:  joinedFilename,
		//			AbsPath:  joinedFilename,
		//			FullPath: joinedFilename,
		//		}
		//		fileInfoX = append(fileInfoX, fix)
		//	}
		//} // for f ranges over filenames
		fileInfoX, err = myReadDirConcurrentWithMatch(dirNamePattern, fileNamePattern)
	} // if flag.NArgs()

	return fileInfoX, err

} // end FileInfoXFromGlob

func FileInfoXFromRegexp(rex *regexp.Regexp) ([]FileInfoExType, error) { // Uses list.ExcludeRex, and can only be used on the current directory.  Uses concurrent code to read the disk.
	var fileInfoX []FileInfoExType
	//excludeMe := ExcludeRex

	if VerboseFlag {
		if rex != nil {
			fmt.Printf(" file regex is %s\n", rex.String())
		} else {
			fmt.Printf(" regexp is nil.\n")
		}
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fileInfoX, err = myReadDirConcurrentWithRex(workingDir, rex) // this already calls includeThis.
	//if err != nil {
	//	return nil, err
	//}

	//if rex == nil {
	//	return fileInfoX, nil
	//}

	//fileInfoX2 := make([]FileInfoExType, 0, len(fileInfoX))
	//for _, f := range fileInfoX { // The exclude expression has already been processed.  Now I have to process the include regexp.
	//	lower := strings.ToLower(f.FI.Name()) // I first used relname here, but that includes the directory, so the ^ anchor was meaningless.  Not what I want.
	//	if VeryVerboseFlag {
	//		fmt.Printf(" FileInfoXFromRegexp: lower = %q, regex = %q, rex.MatchString(lower) = %t\n", lower, rex.String(), rex.MatchString(lower))
	//	}
	//	if rex.MatchString(lower) {
	//		fileInfoX2 = append(fileInfoX2, f)
	//	}
	//}

	//return fileInfoX2, nil
	return fileInfoX, err

} // end FileInfoXFromRegexp

func myReadDirConcurrent(dir string) ([]FileInfoExType, error) { // The entire change including use of []DirEntry happens here.  Concurrent code here is what makes this fdsrt.
	// Adding concurrency in returning []os.FileInfo

	var wg sync.WaitGroup

	if VerboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	deChan := make(chan []os.DirEntry, numWorkers)   // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fixChan := make(chan FileInfoExType, numWorkers) // of individual file infos to be collected and returned to the caller of this routine.
	doneChan := make(chan bool)                      // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fixSlice := make([]FileInfoExType, 0, fetch*multiplier*multiplier)
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
					if includeThisForConcurrent(fi) {
						joinedFilename := filepath.Join(dir, fi.Name())
						fix := FileInfoExType{
							FI:       fi,
							Dir:      dir,
							RelPath:  joinedFilename, // this is a misnomer, but to not have to propagate the correction thru my code, I'll leave this here.
							AbsPath:  joinedFilename,
							FullPath: joinedFilename,
						}
						fixChan <- fix
					}
				}
			}
		}()
	}

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
		for fix := range fixChan {
			fixSlice = append(fixSlice, fix)
			//if fi.Mode().IsRegular() && showGrandTotal {
			//	grandTotal += fi.Size()
			//	grandTotalCount++
			//}
		}
		close(doneChan)
	}()

	d, err := os.Open(dir)
	if err != nil {
		return nil, err
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

	wg.Wait()      // for the deChan
	close(fixChan) // This way I only close the channel once.  I think if I close the channel from within a worker, and there are multiple workers, closing an already closed channel panics.

	<-doneChan // block until channel is freed

	if VerboseFlag {
		fmt.Printf("Found %d files in directory %s.\n", len(fixSlice), dir)
	}

	if FastDebugFlag {
		fmt.Printf("Found %d files in directory %s, first few entries is %v.\n", len(fixSlice), dir, fixSlice[:5])
		if pause() {
			os.Exit(1)
		}
	}

	return fixSlice, nil
} // myReadDirConcurrent

func includeThisForConcurrent(fi os.FileInfo) bool {
	if VeryVerboseFlag {
		fmt.Printf(" includeThisForConcurrent, filterAmt=%d \n", filterAmt)
	}
	if fi.Size() < int64(filterAmt) { // don't need to first check against 0.
		return false
	}
	if fi.Mode().IsRegular() && SymFlag { // Don't include regular files if SymFlag is set.  Only include symlinks.
		return false
	}

	// if get here, file is either a regular file or a symlink.  I want both.

	if ExcludeRex != nil {
		if BOOL := ExcludeRex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	return true
}

func myReadDirConcurrentWithMatch(dir, matchPat string) ([]FileInfoExType, error) { // The entire change including use of []DirEntry happens here, and now concurrent code.
	// Adding concurrency in returning []os.FileInfo
	// This routine adds a call to filepath.Match

	var wg sync.WaitGroup

	if VerboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	deChan := make(chan []os.DirEntry, numWorkers)   // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fixChan := make(chan FileInfoExType, numWorkers) // of individual file infos to be collected and returned to the caller of this routine.
	doneChan := make(chan bool)                      // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fixSlice := make([]FileInfoExType, 0, fetch*multiplier*multiplier)
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
					if includeThisWithMatchForConcurrent(fi, matchPat) {
						joinedFilename := filepath.Join(dir, fi.Name())
						fix := FileInfoExType{
							FI:       fi,
							Dir:      dir,
							RelPath:  joinedFilename, // this is a misnomer, but to not have to propagate the correction thru my code, I'll leave this here.
							AbsPath:  joinedFilename,
							FullPath: joinedFilename,
						}
						fixChan <- fix
					}
				}
			}
		}()
	}

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
		for fix := range fixChan {
			fixSlice = append(fixSlice, fix)
		}
		close(doneChan)
	}()

	d, err := os.Open(dir)
	if err != nil {
		return nil, err
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

	wg.Wait()      // for the closing of the deChan to stop all worker goroutines.
	close(fixChan) // This way I only close the channel once.  I think if I close the channel from within a worker, and there are multiple workers, closing an already closed channel panics.

	<-doneChan // block until channel is freed

	if VerboseFlag {
		fmt.Printf("Found %d files in directory %s.\n", len(fixSlice), dir)
	}

	if FastDebugFlag {
		fmt.Printf("Found %d files in directory %s, first few entries is %v.\n", len(fixSlice), dir, fixSlice[:5])
		if pause() {
			os.Exit(1)
		}
	}

	return fixSlice, nil
} // myReadDirConcurrentWithMatch

func includeThisWithMatchForConcurrent(fi os.FileInfo, matchPat string) bool {
	if VeryVerboseFlag {
		fmt.Printf(" includeThis: filterAmt=%d, match pattern=%s \n", filterAmt, matchPat)
	}
	if fi.Size() < int64(filterAmt) {
		return false
	}
	if fi.Mode().IsRegular() && SymFlag { // Don't include regular files if SymFlag is set.  Only include symlinks.
		return false
	}

	// if get here, file is either a regular file or a symlink.  I want both.

	if ExcludeRex != nil {
		if ExcludeRex.MatchString(strings.ToLower(fi.Name())) {
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
} // end includeThisWithMatchForConcurrent

func myReadDirConcurrentWithRex(dir string, regx *regexp.Regexp) ([]FileInfoExType, error) { // The entire change including use of []DirEntry happens here, and now concurrent code.
	// Adding concurrency in returning []os.FileInfo
	// This routine adds a call to filepath.Match

	var wg sync.WaitGroup

	if VerboseFlag {
		fmt.Printf("Reading directory %s, numworkers = %d\n", dir, numWorkers)
	}
	deChan := make(chan []os.DirEntry, numWorkers)   // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fixChan := make(chan FileInfoExType, numWorkers) // of individual file infos to be collected and returned to the caller of this routine.
	doneChan := make(chan bool)                      // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fixSlice := make([]FileInfoExType, 0, fetch*multiplier*multiplier)
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
					if includeThisWithRexForConcurrent(fi, regx) {
						joinedFilename := filepath.Join(dir, fi.Name())
						fix := FileInfoExType{
							FI:       fi,
							Dir:      dir,
							RelPath:  joinedFilename, // this is a misnomer, but to not have to propagate the correction thru my code, I'll leave this here.
							AbsPath:  joinedFilename,
							FullPath: joinedFilename,
						}
						fixChan <- fix
					}
				}
			}
		}()
	}

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
		for fix := range fixChan {
			fixSlice = append(fixSlice, fix)
		}
		close(doneChan)
	}()

	d, err := os.Open(dir)
	if err != nil {
		return nil, err
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

	wg.Wait()      // for the closing of the deChan to stop all worker goroutines.
	close(fixChan) // This way I only close the channel once.  I think if I close the channel from within a worker, and there are multiple workers, closing an already closed channel panics.

	<-doneChan // block until channel is freed

	if VerboseFlag {
		fmt.Printf("Found %d files in directory %s.\n", len(fixSlice), dir)
	}

	if FastDebugFlag {
		fmt.Printf("Found %d files in directory %s, first few entries is %v.\n", len(fixSlice), dir, fixSlice[:5])
		if pause() {
			os.Exit(1)
		}
	}

	return fixSlice, nil
} // myReadDirConcurrentWithRex

func includeThisWithRexForConcurrent(fi os.FileInfo, rex *regexp.Regexp) bool {
	if VeryVerboseFlag {
		fmt.Printf(" includeThis: filterAmt=%d, match pattern=%v \n", filterAmt, rex)
	}
	if fi.Size() < int64(filterAmt) {
		return false
	}
	if fi.Mode().IsRegular() && SymFlag { // Don't include regular files if SymFlag is set.  Only include symlinks.
		return false
	}

	// if get here, file is either a regular file or a symlink.  I want both.

	if ExcludeRex != nil {
		if ExcludeRex.MatchString(strings.ToLower(fi.Name())) {
			return false
		}
	}

	f := strings.ToLower(fi.Name())
	match := rex.MatchString(f)
	return match

} // end includeThisWithRexForConcurrent

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
