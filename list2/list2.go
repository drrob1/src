package list2

import (
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
	"strings"
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
  29 Dec 22 -- Adding the '.' to be a sentinel marker for the 1st param that's ignored.  This change is made in the platform specific code.
  30 Dec 22 -- I'm thinking about being able to use environment strings to pass around flag values.  ListFilter, ListVerbose, ListVeryVerbose, ListReverse.
                 Nevermind.  I'll pass the variables globally, exported from here.  And I added a procedure New to not stutter, as in list.NewList.  But I kept the old NewList, for now.
   1 Jan 23 -- I changed the display colors for the list.  The line is not all the same color now.
   4 Jan 23 -- Adding screen clearing between screen displays.  Copied from rpng.
   6 Jan 23 -- Improving error handling, by having these functions here return an error variable.  This was needed to better handle the newly added stop code.
  15 Jan 23 -- Now called list2 which will use globals and will use InputRex, so I don't need platform specific code.  All command line params will be output directories,
                 including symlinks.
  18 Jan 23 -- Adding SmartCaseFlag
   8 Feb 23 -- Combined the 2 init functions into one.  It was a mistake to have 2 of them.
  31 Mar 23 -- StaticCheck found a minor issue, about byte values can't be < 0.
*/

type DirAliasMapType map[string]string

type FileInfoExType struct {
	FI      os.FileInfo
	Dir     string
	RelPath string
}

var filterAmt int64 // not exported.  Only the FilterFlag is exported.
var VerboseFlag bool
var VeryVerboseFlag bool
var FilterFlag bool
var ReverseFlag bool
var GlobFlag bool
var DirectoryAliasesMap DirAliasMapType
var fileInfoX []FileInfoExType
var clear map[string]func()
var ExcludeRex *regexp.Regexp
var IncludeRex *regexp.Regexp
var InputDir string
var SizeFlag bool
var SmartCaseFlag bool

const defaultHeight = 40
const minWidth = 90
const minHeight = 26
const stopCode = "0"

var autoWidth, autoHeight int

// ------------------------------------------------------- init -----------------------------------
func init() {
	var err error
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}
	_ = autoWidth

	clear = make(map[string]func(), 3)
	clear["linux"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clear["windows"] = func() { // this is a closure, or an anonymous function
		comspec := os.Getenv("ComSpec")
		cmd := exec.Command(comspec, "/c", "cls") // this was calling cmd, but I'm trying to preserve the scrollback buffer.
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clear["newlines"] = func() {
		fmt.Printf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
	}

	if runtime.GOOS == "windows" {
		DirectoryAliasesMap = GetDirectoryAliases()
	}
}

func New() ([]FileInfoExType, error) {
	lst, err := MakeList(ExcludeRex, SizeFlag, ReverseFlag)
	return lst, err
}

// MakeList will return a slice of strings that contain a full filename including dir
func MakeList(excludeRegex *regexp.Regexp, sizeSort, reverse bool) ([]FileInfoExType, error) {
	var err error

	if FilterFlag {
		filterAmt = 1_000_000
	}
	if VeryVerboseFlag {
		VerboseFlag = true
	}

	fileInfoX, err = getFileInfoXWithGlobals()
	if err != nil {
		return nil, err
	}
	fmt.Printf(" length of fileInfoX = %d\n", len(fileInfoX))

	// set which sort function will be in the sortfcn var
	forward := !ReverseFlag
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

// ------------------------------- MyReadDir -----------------------------------

func MyReadDir(dir string) ([]FileInfoExType, error) { // The entire change including use of []DirEntry happens here.  Who knew?
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
		if includeThis(fi) {
			fix := FileInfoExType{ // fix is a file info extended var
				FI:      fi,
				Dir:     dir,
				RelPath: filepath.Join(dir, fi.Name()),
			}
			fileInfoExs = append(fileInfoExs, fix)
		}
	}
	return fileInfoExs, nil
} // myReadDir

// ---------------------------------------------------- includeThis ----------------------------------------

func includeThis(fi os.FileInfo) bool {
	if VeryVerboseFlag {
		fmt.Printf(" includeThis.  FI=%#v, FilterFlag=%t\n", fi, FilterFlag)
	}
	if !fi.Mode().IsRegular() {
		return false
	}

	if ExcludeRex != nil {
		if BOOL := ExcludeRex.MatchString(strings.ToLower(fi.Name())); BOOL { // If does match the Exclude Regexp
			return false
		}
	}
	if FilterFlag {
		if fi.Size() < filterAmt {
			return false
		}
	}
	if IncludeRex != nil {
		if SmartCaseFlag {
			if BOOL := IncludeRex.MatchString(fi.Name()); !BOOL { // if does not match the Include Regexp
				return false
			}
		} else {
			if BOOL := IncludeRex.MatchString(strings.ToLower(fi.Name())); !BOOL { // if does not match the Include Regexp
				return false
			}
		}
	}
	return true
} // end includeThis

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

func ProcessDirectoryAliases(cmdline string) string { // the directory aliases map is initialized in init()

	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	//aliasesMap = GetDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := DirectoryAliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathNFile := cmdline[idx+1:]
	completeValue := aliasValue + PathNFile
	return completeValue
} // ProcessDirectoryAliases

// ----------------------------- ExpandADash ---------------------------------------

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
	if c > 26 { // a byte value can't be < 0.
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

// ------------------------------------ ExpandAllDashes --------------------------------------------

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

// -------------------------------------------- FileSelection -------------------------------------------------------

func FileSelection(inList []FileInfoExType) ([]FileInfoExType, error) {
	outList := make([]FileInfoExType, 0, len(inList))
	numOfLines := min(autoHeight, minHeight)
	numOfLines = min(numOfLines, len(inList))
	var beg, end int
	lenList := len(inList)
	var ans string
	onWin := runtime.GOOS == "windows"

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
			//ctfmt.Printf(colr, onWin, " %c: %s -- %s  %s\n", i+'a', f.RelPath, s, t)
			ctfmt.Printf(colr, onWin, " %c: %s ", i+'a', f.RelPath)
			clr := ct.White
			if clr == colr { // don't use same color as rest of the displayed string.
				clr = ct.Yellow
			}
			ctfmt.Printf(clr, onWin, " -- %s", t)
			ctfmt.Printf(ct.Cyan, onWin, " %s\n", s)
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
			outList = append(outList, f)
		}
		if end >= lenList {
			break
		}
		beg = end

		//clearFunc := clear[runtime.GOOS]  // I'm playing w/ an alternative to blanking the screen, so the scroll back buffer is preserved.
		clearFunc := clear["newlines"]
		clearFunc()
	}

	return outList, nil
} // end FileSelection

// ------------------------------------- min ----------------------------------------------------------

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

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
} // end GetMagnitudeString

// --------------------------------------------- getFileInfoXWithGlobals ------------------------------------------

func getFileInfoXWithGlobals() ([]FileInfoExType, error) {
	var fileInfoX []FileInfoExType
	// From globals: InputDir, IncludeRex, ExcludeRex, ReverseFlag, SizeFlag

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from getFileInfoXWithGlobals os.Getwd is %#v\n", err)
		os.Exit(1)
	}

	if InputDir != "" {
		workingDir = InputDir
		if strings.Contains(InputDir, ":") {
			workingDir = ProcessDirectoryAliases(InputDir)
		}
	}

	if VerboseFlag {
		fmt.Printf(" workingDir=%s\n", workingDir)
	}

	fileInfoX, err = MyReadDir(workingDir) // excluding by regex, filesize or having an ext is done by MyReadDir.
	if err != nil {
		return nil, err
	}
	if VerboseFlag {
		fmt.Printf(" after call to MyReadDir.  Len(fileInfoX)=%d\n", len(fileInfoX))
	}

	if VerboseFlag {
		fmt.Printf(" Leaving GetFileInfoXWithGlobals.  len(fileinfos)=%d\n", len(fileInfoX))
	}
	return fileInfoX, nil
} // end getFileInfoXWithGlobals
