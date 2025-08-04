package main

/*
 1 Aug 25 -- Started working on a routine to scan the weekly schedule and create a list of doc names on it, instead of having to provide one.
			It works.  Now to see if one other varient works.  Yep, all routines here work.
 2 Aug 25 -- Playing some more.  Now that I remember about slice.Compact, I want to test if it needs a sorted slice to work.  Turns out that it does need a sorted slice.
				Turns out that it does need a sorted slice, as it only removes consecutive occurances of duplicate strings.
 4 Aug 25 -- I discovered that typos in the doc names are not rare.  I want to notify the user that there may be a typo.  I'll use soundex for this.
*/

import (
	"bytes"
	"errors"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"github.com/tealeg/xlsx/v3"
	"github.com/umahmood/soundex"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"src/filepicker"
	"src/misc"
	"src/whichexec"
	"strconv"
	"strings"
	"time"
)

const lastModified = "4 Aug 2025"
const maxDimensions = 200

const conf = "docnames.conf"
const ini = "docnames.ini"

const (
	weekdayOncall = iota + 3 // and code is a 0-origin, while Excel is 1-origin for rows.
	neuro
	body
	erXrays
	ir
	nuclear
	us
	peds
	fluoroJH
	fluoroFH
	msk // includes Xray In/Outpatient and off-sites
	mammo
	boneDensity
	late
	moonlighters
	weekendJH
	weekendFH
	weekendIR // includes On-Call Radiologist for diagnostic
	mdOff
	totalAmt
)

type fileDataType struct {
	ffname    string // full filename
	timestamp time.Time
}

type soundexSlice struct {
	s      string
	soundx string
}

var conlyFlag = flag.BoolP("conly", "c", false, "Conly mode, ie, search only Documents on c: for current user.")
var numLines = 15 // I don't expect to need more than these, as I display only the first 26 elements (a-z) so far.
var veryVerboseFlag bool
var startDirectory string
var verboseFlag = flag.BoolP("verbose", "v", false, "verbose debugging output")
var monthsThreshold int

func findAndReadConfIni() error { // Only is used to get startDirectory
	// will search first for conf and then for ini file in this order of directories: current, home, config.

	fullFile, found := whichexec.FindConfig(conf)
	if !found {
		fullFile, found = whichexec.FindConfig(ini)
		if !found {
			return fmt.Errorf("%s or %s not found", conf, ini)
		}
	}

	// now need to process the config file using code from fansha.
	fileByteSlice, err := os.ReadFile(fullFile)
	if err != nil {
		return err
	}
	bytesReader := bytes.NewReader(fileByteSlice)
	inputLine, err := misc.ReadLine(bytesReader)
	if err != nil {
		return err
	}

	trimmedInputLine, ok := strings.CutPrefix(inputLine, "startdirectory") // CutPrefix became available as of Go 1.20
	if ok {
		startDirectory = trimmedInputLine
		startDirectory = strings.TrimSpace(startDirectory)
	}

	return nil
}

func excludeMe(s string) bool {
	dgtRegexp := regexp.MustCompile(`\d`) // any digit character will match this exprn.
	if strings.Contains(s, "fh") || strings.Contains(s, "dr.") || strings.Contains(s, "(") || strings.Contains(s, ")") || strings.Contains(s, "/") ||
		strings.Contains(s, "jh") || strings.Contains(s, "plain") || strings.Contains(s, "please") || strings.Contains(s, "sat") ||
		strings.Contains(s, "see") || strings.Contains(s, "sun") || strings.Contains(s, "thu") || strings.Contains(s, "modality") ||
		strings.Contains(s, ":") || strings.Contains(s, "*") || strings.Contains(s, "@") || dgtRegexp.MatchString(s) {
		return true
	}
	return false
}

func uniqueStrings(s []string) []string { // AI wrote this, and then I changed how it defined the list string slice.
	keys := make(map[string]bool)
	list := make([]string, 0, len(s))
	for _, entry := range s {
		if _, value := keys[entry]; !value { // this relies on fact that if key is not present, returned result will be false.
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// After writing this, I was reminded that there is now slices.Compact(), which does this as part of the std library, as of ~ Go 1.20.
func anotherUniqueStrings(s []string) []string {
	sort.Strings(s)
	list := make([]string, 0, len(s))
	list = append(list, s[0]) // first element is copied to the list
	for i := 1; i < len(s); i++ {
		if s[i] != s[i-1] {
			list = append(list, s[i])
		}
	}
	return list
}

// func readScheduleRowsRtnStrSlice(wb *xlsx.File)  I decided to not use this form.
func readScheduleRowsRtnStrSlice(fn string) ([]string, error) {
	docNamesSlice := make([]string, 0, maxDimensions)
	workBook, err := xlsx.OpenFile(fn)
	if err != nil {
		return nil, err
	}
	sheet := workBook.Sheets[0]

	for row := weekdayOncall; row < totalAmt; row++ {
		for col := 1; col < 6; col++ {
			cell, err := sheet.Cell(row, col)
			if err != nil {
				return nil, err
			}
			s := cell.String()
			fields := strings.Fields(s)
			for _, field := range fields {
				field = strings.ToLower(field)
				if excludeMe(field) {
					continue
				}
				docNamesSlice = append(docNamesSlice, field)
			}
		}
	}
	return docNamesSlice, nil
}

func readScheduleRowsMapRtn(fn string) ([]string, error) {
	docNamesSlice := make([]string, 0, maxDimensions)
	keysMap := make(map[string]bool)
	workBook, err := xlsx.OpenFile(fn)
	if err != nil {
		return nil, err
	}
	sheet := workBook.Sheets[0]

	for row := weekdayOncall; row < totalAmt; row++ {
		for col := 1; col < 6; col++ {
			cell, err := sheet.Cell(row, col)
			if err != nil {
				return nil, err
			}
			s := cell.String()
			fields := strings.Fields(s)
			for _, field := range fields {
				field = strings.ToLower(field)
				if excludeMe(field) {
					continue
				}
				if _, value := keysMap[field]; !value { // this relies on fact that if key is not present, returned result will be false.
					keysMap[field] = true
					docNamesSlice = append(docNamesSlice, field)
				}
			}
		}
	}
	return docNamesSlice, nil
}

func getSoundex(input []string) []soundexSlice {
	out := make([]soundexSlice, 0, len(input))
	var sound soundexSlice
	for _, inp := range input {
		if inp == "choi" { // choi and chiu have the same result.  So only 1 can be used.
			continue
		}
		sound.s = inp
		sound.soundx = soundex.Code(inp)
		out = append(out, sound)
	}
	return out
}

func main() {

	flag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "very verbose debugging output")
	flag.IntVarP(&monthsThreshold, "months", "m", 1, "months threshold for schedule files")
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], lastModified, runtime.Version())
		fmt.Printf(" Usage: %s [weekly xlsx file] \n", os.Args[0])
		fmt.Printf(" Needs lint.conf or lint.ini, and looks in current, home and config directories.\n")
		fmt.Printf(" first line must begin with off, and 2nd line, if present, must begin with startdirectory.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if veryVerboseFlag {
		*verboseFlag = true
	}

	fmt.Printf(" Get Doc Names from weekly schedule, last modified %s, compiled with %s, using pflag.\n", lastModified, runtime.Version())

	err := findAndReadConfIni()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from findAndReadConfINI: %s.  Exiting.\n", err)
		return
	}

	if *verboseFlag {
		fmt.Printf(" After findAndReadConfIni, Start Directory: %s\n", startDirectory)
	}

	filepicker.VerboseFlag = *verboseFlag

	var filename, ans string

	// filepicker stuff.

	includeODrive := !*conlyFlag // a comvenience flag
	if flag.NArg() == 0 {
		var filenames []string
		if includeODrive {
			filenames, err = walkRegexFullFilenames() // function is below.  "o:\\week.*xls.?$"
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from walkRegexFullFilenames is %s.  Exiting \n", err)
				return
			}
			if *verboseFlag {
				fmt.Printf(" Filenames length from o drive: %d\n", len(filenames))
			}
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf(" Error from os.UserHomeDir: %s\n", err)
			return
		}
		docs := filepath.Join(filepath.Join(homeDir, "Documents"), "week.*xls.?$")
		if *verboseFlag {
			fmt.Printf(" homedir=%q, Joined Documents: %q\n", homeDir, docs)
		}
		//                                                  filenamesDocs, err := filepicker.GetRegexFullFilenames(docs)
		filenamesDocs, err := filepicker.GetRegexFullFilenamesNotLocked(docs)
		if err != nil {
			fmt.Printf(" Error from filepicker is %s.  Exiting \n", err)
			return
		}
		if *verboseFlag {
			fmt.Printf(" FilenamesDocs length: %d\n", len(filenamesDocs))
		}

		filenames = append(filenames, filenamesDocs...)
		if *verboseFlag {
			fmt.Printf(" Filenames length after append operation: %d\n", len(filenames))
		}

		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == ";" {
			fmt.Println(" No files entered.  Exiting.")
			return
		}
		i, e := strconv.Atoi(ans)
		if e == nil {
			filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			filename = filenames[i]
		}
		fmt.Println(" Picked spreadsheet is", filename)
	} else { // will use filename entered on commandline
		filename = flag.Arg(0)
	}

	docnamesSlice, err := readScheduleRowsRtnStrSlice(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from readScheduleRowsRtnStrSlice is %s.  Exiting \n", err)
		return
	}
	if *verboseFlag {
		fmt.Printf(" docnamesSlice length: %d\n", len(docnamesSlice))
		fmt.Printf(" docnamesSlice: %#v\n\n", docnamesSlice)
	}

	uniqueDocnamesSlice := uniqueStrings(docnamesSlice)
	anotherUniqueDocnamesSlice := anotherUniqueStrings(docnamesSlice)
	if *verboseFlag {
		fmt.Printf(" After calling uniqueStrings -- length: %d\n", len(uniqueDocnamesSlice))
		fmt.Printf(" uniqueDocnamesSlice: %#v\n\n", uniqueDocnamesSlice)
		fmt.Printf(" AntherUniquedocnamesSlice: %#v\n\n", anotherUniqueDocnamesSlice)
	}

	uniqueDocnamesFromMap, err := readScheduleRowsMapRtn(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from readScheduleRowsMapRtn is %s.  Exiting \n", err)
		return
	}
	if *verboseFlag {
		fmt.Printf(" After calling readScheduleRowsMapRtn -- length: %d\n", len(uniqueDocnamesFromMap))
		fmt.Printf(" uniqueDocnamesMap: %#v\n\n", uniqueDocnamesFromMap)
	}

	uniqueNames, err := readScheduleRowsRtnStrSlice(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error from readScheduleRowsRtnStrSlice is %s.  Exiting \n", err)
		return
	}
	uniqueNames = slices.Compact(uniqueNames)
	if *verboseFlag {
		fmt.Printf(" After calling slices.Compact(uniqueNames) -- length: %d\n", len(uniqueNames))
		fmt.Printf(" uniqueNames: %#v\n\n", uniqueNames)
	}

	soundexStrings := getSoundex(anotherUniqueDocnamesSlice)
	if *verboseFlag {
		fmt.Printf(" After calling getSoundex -- length: %d\n", len(soundexStrings))
		for i, s := range soundexStrings {
			fmt.Printf(" %d: %s -- %s\n", i, s.s, s.soundx)
		}
	}

}

// Needs a walk function to find what it is looking for.  See top comments.
func walkRegexFullFilenames() ([]string, error) { // This rtn sorts using sort.Slice

	// validate the regular expression
	regex, err := regexp.Compile("week.*xls.?$")
	if err != nil {
		return nil, err
	}

	// define the timestamp constraint of >= this monthyear.  No, >= 1 month ago.
	t0 := time.Now()
	threshold := t0.AddDate(0, -monthsThreshold, 0) // threshhold is months ago set by a commandline flag.  Default is 1 month.
	timeout := t0.Add(5 * time.Minute)

	// set up channel to receive FDSlices and append them to a master file data slice
	boolChan := make(chan bool)                 // unbuffered
	fileDataChan := make(chan fileDataType, 10) // a magic number I pulled out of the air
	FDSlice := make([]fileDataType, 0, 100)     // a magic number here, for now
	receiverFunc := func() {
		for fd := range fileDataChan { // fd is of type filedatatype
			FDSlice = append(FDSlice, fd)
		}
		boolChan <- true // pause until this go routine completes its work
	}
	go receiverFunc() // keep receiving on the receiver chan and appending to the master slice

	// Put walk func here.  It has to check the directory entry it gets, then search for all filenames that meet the regex and timestamp constraints.
	walkDirFunction := func(fpath string, de os.DirEntry, err error) error {
		if *verboseFlag {
			if err != nil {
				fmt.Printf(" WalkDir fpath %s, de.name invalid, err %v \n", fpath, err.Error())
			} else {
				fmt.Printf(" WalkDir fpath %s\n", fpath)
			}
		}
		if err != nil {
			return filepath.SkipDir
		}
		if de.IsDir() {
			return nil // allow walk function to drill down itself
		}

		if time.Now().After(timeout) {
			return errors.New("timeout occurred")
		}

		// Not a directory, and timeout has not happened.  Only process regular files, and skip symlinks.
		if !de.Type().IsRegular() {
			return nil
		}

		lowerName := strings.ToLower(de.Name())
		if regex.MatchString(lowerName) {
			if strings.HasPrefix(de.Name(), "~") { // ignore files that Excel has locked.
				return nil
			}
			fi, err := de.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from dirEntry(%s).Info() is %v \n", de.Name(), err)
				return nil
			}

			filedata := fileDataType{
				ffname:    fpath, // filepath.Join(fpath, de.Name()) doesn't work, because it turns out that fpath is the full path and filename
				timestamp: fi.ModTime(),
			}
			fileDataChan <- filedata
		}

		// I now have a walk function that has collected all matching names in this fpath into a slice of fileInfos.
		// Now what do I do now?  I want, after the call to the walk function, to have a slice of full filenames that match the constraints of name and timestamp.
		// Maybe not, I want, after the call to the walk function, to have a slice of matching full file infos, that then have to be tested to see if thay match the timestamp constraint.
		// I may need a go routine to collect all these slices into 1 slice to be sorted.  And I don't have a full filename in this slice.  I still have to
		// construct that.
		// Maybe I need a struct that has full filename and the timestamp, ie, fileInfo.ModTime(), which is of type time.Time.  I did this, and I call it fileDataType.
		// Then I made FileDataSliceType I call FDSliceType.  Then I made a masterFDSlice and a FDSlice channel so the walk function sends a slice of filedatas to the
		// goroutine that collects these and appends them to a masterFileDataSlice.  I use 2 channels for this, one to send the local filedata in the walk function,
		// and another to signal when all of these local slices have been appended to the master filedata slice.
		// Then I sort the master filedata slice and then only take at most the top 15 to be returned to the caller.
		// While debugging this code, I came across the fact that I've misunderstood what the walk fcn does.  It doesn't just produce directory names, it produces any file entries
		// with each iteration.  So opening a directory here and fetching all the entries is not correct.  I have to change this to work on individual entries.
		// Now it works.  The walk function gets the needed filenames, checks against the regex and sends down the channel if the file matches.
		// This routine then sorts the []FileDataType and checks against the date threshold constraint when converting to []string.

		return nil
	}

	if startDirectory == "" { // if not set by the config file.
		if runtime.GOOS == "windows" { // Variable is defined globally.
			startDirectory = "o:\\Nikyla's File\\RADIOLOGY MD Schedule\\"
		} else { // this is so I can debug on leox, too.  Variable is defined globally.
			startDirectory = "/home/rob/bigbkupG/Nikyla's File/RADIOLOGY MD Schedule"
		}
	}

	if *verboseFlag {
		fmt.Printf(" WalkDir startDirectory: %s\n", startDirectory)
	}

	err = filepath.WalkDir(startDirectory, walkDirFunction)
	if err != nil {
		return nil, err
	}
	close(fileDataChan)
	<-boolChan // pause until FDSlice is filled.  boolChan is an unbuffered channel

	lessFunc := func(i, j int) bool {
		//return masterFDSlice[i].timestamp.UnixNano() > masterFDSlice[j].timestamp.UnixNano() // this works
		return FDSlice[i].timestamp.After(FDSlice[j].timestamp) // I want to see if this will work.  It does, and TRUE means time[i] is after time[j].
	}
	sort.Slice(FDSlice, lessFunc)

	if *verboseFlag {
		fmt.Printf(" in walkRegexFullFilenames after sort.Slice.  Len=%d\n  FDSlice %#v\n", len(FDSlice), FDSlice)
	}

	stringSlice := make([]string, 0, len(FDSlice))
	var count int
	for _, f := range FDSlice {
		if f.timestamp.After(threshold) {
			stringSlice = append(stringSlice, f.ffname) // needs to preserve case of filename for linux
			count++
			if count >= numLines {
				break
			}
		} else {
			break // Made observation that in a sorted list the first file before the threshold timestamp ends the search.
		}
	}
	if *verboseFlag {
		fmt.Printf(" In walkRegexFullFilenames.  len(stringSlice) = %d\n stringSlice=%v\n", len(stringSlice), stringSlice)
	}
	return stringSlice, nil
} // end walkRegexFullFilenames
