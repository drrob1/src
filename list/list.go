package list

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 2022 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                   This is going to take a while.
  20 Dec 2022 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                   I decided to only copy files if the new one is newer than the old one.
  22 Dec 2022 -- Now I want to colorize the output, so I have to return the os.FileInfo also.  So I changed MakeList and NewList to not return []string, but return []FileInfoExType.
                   And myReadDir creates the relPath field that I added to FileInfoExType.
  25 Dec 2022 -- Moved FileSection here.
  26 Dec 2022 -- Changed test against the regexp to be nil instead of "".
*/

type dirAliasMapType map[string]string

type FileInfoExType struct {
	FI      os.FileInfo
	Dir     string
	RelPath string
}

//var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, verboseFlag bool
//var filterFlag, globFlag, veryVerboseFlag, halfFlag, maxDimFlag bool
//var filterAmt, numLines, numOfLines, grandTotalCount int
//var sizeTotal, grandTotal int64
//var filterStr string
//var excludeRegex *regexp.Regexp
//var excludeRegexPattern string
var noExtensionFlag, verboseFlag bool
var globFlag bool
var filterAmt int
var directoryAliasesMap dirAliasMapType
var fileInfoX []FileInfoExType

const defaultHeight = 40
const minWidth = 90
const minHeight = 26

//const sepString = string(filepath.Separator) not used, it seems

var autoWidth, autoHeight int
var err error

func init() {
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		//autoDefaults = false
		autoHeight = defaultHeight
		autoWidth = minWidth
	}
	_ = autoWidth
}

func NewList(excludeMe *regexp.Regexp, sizeSort, reverse bool) []FileInfoExType {
	return MakeList(excludeMe, sizeSort, reverse)
}

// MakeList will return a slice of strings that contain a full filename including dir
func MakeList(excludeRegex *regexp.Regexp, sizeSort, reverse bool) []FileInfoExType {

	fileInfoX = getFileInfoXFromCommandLine(excludeRegex)
	fmt.Printf(" length of fileInfoX = %d\n", len(fileInfoX))

	// set which sort function will be in the sortfcn var
	Forward := !reverse
	DateSort := !sizeSort
	sortFcn := func(i, j int) bool { return false } // became available as of Go 1.8
	if sizeSort && Forward {                        // set the value of sortfcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfoX[i].FI.Size() > fileInfoX[j].FI.Size() // I want a largest first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//       return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfoX[i].FI.ModTime().After(fileInfoX[j].FI.ModTime()) // I want a newest-first sort.
		}
		if verboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if sizeSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfoX[i].FI.Size() < fileInfoX[j].FI.Size() // I want a smallest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfoX[i].FI.ModTime().Before(fileInfoX[j].FI.ModTime()) // I want an oldest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if len(fileInfoX) > 1 {
		sort.Slice(fileInfoX, sortFcn)
	}

	//fileString := make([]string, 0, len(fileInfoX))
	//for _, fix := range fileInfoX {
	//	f := filepath.Join(fix.dir, fix.fi.Name())
	//	fileString = append(fileString, f)
	//}
	return fileInfoX
} // end MakeList

// ------------------------------- myReadDir -----------------------------------

func MyReadDir(dir string, excludeMe *regexp.Regexp) []FileInfoExType { // The entire change including use of []DirEntry happens here.  Who knew?
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	fileInfoExs := make([]FileInfoExType, 0, len(dirEntries))
	for _, d := range dirEntries {
		fi, e := d.Info()
		if e != nil {
			fmt.Fprintf(os.Stderr, " Error from %s.Info() is %v\n", d.Name(), e)
			continue
		}
		if includeThis(fi, excludeMe) {
			fix := FileInfoExType{ // fix is a file info extended var
				FI:      fi,
				Dir:     dir,
				RelPath: filepath.Join(dir, fi.Name()),
			}
			fileInfoExs = append(fileInfoExs, fix)
		}
	}
	return fileInfoExs
} // myReadDir

// ---------------------------------------------------- includeThis ----------------------------------------

func includeThis(fi os.FileInfo, excludeRex *regexp.Regexp) bool {
	//if veryVerboseFlag {
	//	fmt.Printf(" includeThis.  noExtensionFlag=%t, excludeFlag=%t, filterAmt=%d \n", noExtensionFlag, excludeFlag, filterAmt)
	//}
	if !fi.Mode().IsRegular() {
		return false
	}
	if noExtensionFlag && strings.ContainsRune(fi.Name(), '.') {
		return false
	} else if filterAmt > 0 {
		if fi.Size() < int64(filterAmt) {
			return false
		}
	}
	//if excludeRex.String() != "" {
	//	if BOOL := excludeRex.MatchString(strings.ToLower(fi.Name())); BOOL {
	//		return false
	//	}
	//}
	if excludeRex != nil {
		if BOOL := excludeRex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	return true
}

//------------------------------ GetDirectoryAliases ----------------------------------------

func GetDirectoryAliases() dirAliasMapType { // Env variable is diraliases.

	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = strings.ReplaceAll(s, "_", " ") // substitute the underscore, _, for a space so strings.Fields works correctly
	directoryAliasesMap := make(dirAliasMapType, 10)

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
	PathNFile := cmdline[idx+1:]
	completeValue := aliasValue + PathNFile
	//fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
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
	if c < 0 || c > 26 {
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

func FileSelection(inList []FileInfoExType) []FileInfoExType {
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
			ctfmt.Printf(colr, onWin, " %c: %s -- %s  %s\n", i+'a', f.RelPath, s, t)
		}

		fmt.Print(" Enter selections: ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "" // it seems that if I don't do this, the prev contents are not changed when I just hit <enter>
		}

		// here is where I can scan the ans string looking for a-z and replace that with all the letters so indicated before passing it onto the processing loop.
		// ans = strings.ToLower(ans)  Upper case letter will mean something, not sure what yet.
		processedAns, err := ExpandAllDashes(ans)
		//fmt.Printf(" ans = %#v, processedAns = %#v\n", ans, processedAns)

		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR from ExpandAllDashes(%s): %q\n", ans, err)
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
	}

	return outList
} // end FileSelection

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
}