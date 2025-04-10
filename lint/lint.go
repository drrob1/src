package main // lint.go, from lint2.go from lint.go

import (
	"bytes"
	"errors"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"github.com/stoewer/go-strcase"
	"github.com/tealeg/xlsx/v3"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"src/filepicker"
	"src/misc"
	"src/tknptr"
	"src/whichexec"
	"strconv"
	"strings"
	"time"
	//"flag"
)

/*
  26 Sep 24 -- Started first version.  Intended as a linter for the weekly work schedule.  It will need a .conf or .ini file to establish the suitable params.
               It will need lists to specify who can be covering a specific area, and to make sure that if someone is on vacation, their name does not appear anywhere else
               for that day.  So I'll need categories in the .conf or .ini file, such as:
				weekdayOncall row 3
				neuro row 4
				body row 5
				ER row 6
				Xrays row 6
				IR row 7
				Nuclear row 8
				US row 9
				Pediatrics row 10
				FLUORO JH row 11
				FLUORO FH row 12
				MSK row 13
				MAMMO row 14
				BONE (DENSITY) row 15
				LATE row 16
				Moonlighters row 17
				Weekend JH row 18
				Weekend FH row 19
				Weekend IR row 20
				MD's Off (vacation) row 21
				Below row 21 are the MD phone #'s.

				if the line begins with any of [# ; /] then it's a comment.  If a line doesn't begin w/ a keyword, then it's an error and the pgm exits.
				I think I'll just check the vacation rule first.  Then expand it to the other rules.

                I have to read the weekly schedule into an appropriate data structure, as also the .conf/.ini file.

 xlsx (github.com/tealeg/xlsx/v3)
----------------------------------------------------------------------------------------------------
  28 Jan 25 -- Now called lint2, and added detection of having the late doc also be on fluoro.  That happened today at Flushing, and it was a mistake.
				I think I can do it without much difficulity.
  29 Jan 25 -- I'm going to make the week a 2D array, and use a map to get the names from the row #.  Just to see if it will work.  It does.
				Now I'm going to try to see if a remote doc is on fluoro, like there was today.
  30 Jan 25 -- And I shortened the main loop looking for vacation docs assigned to clinical work.
----------------------------------------------------------------------------------------------------
  31 Jan 25 -- Renamed back to lint.go
   2 Feb 25 -- Made dayNames an array instead of a slice.  It's fixed, so it doesn't need to be a slice.
  14 Mar 25 -- Today is Pi Day.  But that's not important now.
				I want to refactor this so it works in the environment it's needed.  It needs to get the files from o: drive and then homeDir/Documents, both.
				So I want to write the routine here as taking a param of a full filename and scanning that file.
				First I want to see if the xlsx.OpenFile takes a full file name as its param.  If so, that'll be easier for me to code.  It does.
  16 Mar 25 -- Changing colors that are displayed.
  18 Mar 25 -- Still doesn't work for Nikki, as it doesn't find the files on O-drive.  I'll broaden the expression to include all Excel files.
  20 Mar 25 -- 1.  I need a switch to only search c:, for my use.  I'll call this conly, apprev as c.
               2.  Nikki uses a much more complex directory structure on o-drive than I expected.  I think I'm going to need a walk function to search for all files
					timestamped this month or next month, and add their full name to the slice, sort the slice by date, newest first.
				And changed to use pflag.
  22 Mar 25 -- It now works as intended.  So now I want to add a flag to set the time interval that's valid, and a config file setting for the directory searched on o:
				And I'll add a veryverboseFlag, using vv and w.  The veryverbose setting will be for the Excel tests I don't need anymore.
  27 Mar 25 -- I ran this using the -race flag; it's clean.  No data races.  Just checking.
   9 Apr 25 -- Made observation that since the walk function sorts its list, the first file that doesn't meet the threshold date can stop the search since the rest are older.
  10 Apr 25 -- Made the change by adding and else clause to an if statement in the walk function.
*/

const lastModified = "10 Apr 2025"
const conf = "lint.conf"
const ini = "lint.ini"

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

type dayType [23]string // there are a few unused entries here.  This goes from 0..22.  Indices 0..2 are not used.

type fileDataType struct {
	ffname    string // full filename
	timestamp time.Time
}

var categoryNamesList = []string{"0", "1", "2", "weekday On Call", "Neuro", "Body", "ER/Xrays", "IR", "Nuclear Medicine", "US", "Peds", "Fluoro JH", "Fluoro FH",
	"MSK", "Mammo", "Bone Density", "late", "weekend moonlighters", "weekend JH", "weekend FH", "weekend IR", "MD's Off"} // 0, 1 and 2 are unused

var dayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var names = make([]string, 0, 25)  // a list of all the doc's last names as read from the config file.
var dayOff = make(map[string]bool) // only used in findAndReadConfIni when verboseFlag is set

var verboseFlag = flag.BoolP("verbose", "v", false, "Verbose mode")
var conlyFlag = flag.BoolP("conly", "c", false, "Conly mode, ie, search only Documents on c: for current user.")
var numLines = 15 // I don't expect to need more than these, as I display only the first 26 elements (a-z) so far.
var veryVerboseFlag bool
var startDirectory string
var monthsThreshold int

// Next I will code the check against the vacation people to make sure they're not assigned to anything else.

func findAndReadConfIni() error {
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

	// remove any commas
	inputLine = strings.ReplaceAll(inputLine, ",", "")

	// will use my tknptr stuff here.
	tokenslice := tknptr.TokenSlice(inputLine)
	lower := strings.ToLower(tokenslice[0].Str)
	if !strings.Contains(lower, "off") {
		return fmt.Errorf("%s is not off", tokenslice[0].Str)
	}
	for _, token := range tokenslice[1:] {
		lower = strings.ToLower(token.Str)
		names = append(names, lower)
		dayOff[lower] = false // this is a map[string]bool
	}
	if veryVerboseFlag {
		fmt.Printf(" Loaded config file: %s, containing %s\n", fullFile, inputLine)
		for doc, vacay := range dayOff {
			fmt.Printf(" dayOff[%s]: %t, ", doc, vacay)
		}
		fmt.Println()
		fmt.Printf(" Names unsorted: %#v\n", names)
	}

	sort.Strings(names)

	if *verboseFlag {
		fmt.Printf(" Sorted Names: %#v\n\n", names)
	}

	// Now to process setting startdirectory

	if *verboseFlag {
		fmt.Printf(" In findAndReadConfIni before 2nd ReadLine. BytesReader.Len=%d, and .size=%d\n", bytesReader.Len(), bytesReader.Size())
	}
	inputLine, err = misc.ReadLine(bytesReader)
	if *verboseFlag {
		fmt.Printf(" In findAndReadConfIni after 2nd ReadLine BytesReader.Len=%d, and .size=%d, inputline=%q\n",
			bytesReader.Len(), bytesReader.Size(), inputLine)
	}
	if err != nil {
		fmt.Printf(" Error reading 2nd config line: %s\n", err)
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	trimmedInputLine, ok := strings.CutPrefix(inputLine, "startdirectory") // CutPrefix became available as of Go 1.20
	if ok {
		startDirectory = trimmedInputLine
		startDirectory = strings.TrimSpace(startDirectory)
	}

	return nil
}

func readDay(wb *xlsx.File, col int) (dayType, error) {
	var day dayType
	sheets := wb.Sheets

	for i := weekdayOncall; i < totalAmt; i++ {
		cell, err := sheets[0].Cell(i, col) // always sheet[0]
		if err != nil {
			return dayType{}, err
		}

		day[i] = cell.String()

		//switch i {
		//case 3:
		//	day.weekdayOncall = cell.String()
		//case 4:
		//	day.neuro = cell.String()
		//case 5:
		//	day.body = cell.String()
		//case 6:
		//	day.er = cell.String()
		//	day.xrays = cell.String()
		//case 7:
		//	day.ir = cell.String()
		//case 8:
		//	day.nuclear = cell.String()
		//case 9:
		//	day.us = cell.String()
		//case 10:
		//	day.peds = cell.String()
		//case 11:
		//	day.fluoroJH = cell.String()
		//case 12:
		//	day.fluoroFH = cell.String()
		//case 13:
		//	day.msk = cell.String()
		//case 14:
		//	day.mammo = cell.String()
		//case 15:
		//	day.boneDensity = cell.String()
		//case 16:
		//	day.late = cell.String()
		//case 17:
		//	day.moonlighters = cell.String()
		//case 18:
		//	day.weekendJH = cell.String()
		//case 19:
		//	day.weekendFH = cell.String()
		//case 20:
		//	day.weekendIR = cell.String()
		//case 21:
		//	day.mdOff = cell.String()
		//
		//default:
		//	return dayType{}, fmt.Errorf("unknown day type %d", i)
		//
		//}
	}
	return day, nil
}

func whosOnVacationToday(week [6]dayType, dayCol int) []string { // week is an array, not a slice.  It doesn't need a slice.
	// this function is to return a slice of names that are on vacation for this day
	vacationString := strings.ToLower(week[dayCol][mdOff])

	mdsOff := make([]string, 0, 15) // Actually, never more than 10 off, but religious holidays can have a lot off.

	// search for matching names
	for _, vacationName := range names { // names is a global
		dayOff[vacationName] = false
		if strings.Contains(vacationString, vacationName) {
			dayOff[vacationName] = true
			mdsOff = append(mdsOff, vacationName)
		}
	}
	return mdsOff
}

func whosLateToday(week [6]dayType, dayCol int) []string { // week is an array, not a slice.  It doesn't need a slice.
	// this function is to return a slice of names that are on the late shift for today.  Only 2 per day are late.

	lateString := strings.ToLower(week[dayCol][late])

	lateDocs := make([]string, 0, 2)
	// search for matching names
	for _, lateName := range names { // names is a global
		if strings.Contains(lateString, lateName) {
			lateDocs = append(lateDocs, lateName)
		}
	}
	return lateDocs
}

func whosRemoteToday(week [6]dayType, dayCol int) []string { // week is an array, not a slice.  It doesn't need a slice.
	// this function is to return a slice of names that are working remotely today.
	const remoteMarkerString = "(*R)"

	remoteDocs := make([]string, 0, 10) // Never more than 5 are allowed, but names can be duplicated.
	// search for matching names
	for _, cell := range week[dayCol] {
		cell = strings.ReplaceAll(cell, "   ", " ")
		cell = strings.ReplaceAll(cell, "  ", " ")
		cell = strings.ReplaceAll(cell, "  ", " ")
		cell = strings.ReplaceAll(cell, "  ", " ") // really really really make sure that extra spaces are removed.
		fields := strings.Split(cell, " ")

		for i, field := range fields {
			field = strings.TrimSpace(field)
			if strings.Contains(field, remoteMarkerString) {
				j := i - 1
				prevField := strings.ToLower(fields[j])
				remoteDocs = append(remoteDocs, prevField)
			}
		}
	}
	sort.Strings(remoteDocs)
	remoteDocs = slices.Compact(remoteDocs) //  De-duplicating w/ the new slices package.  And make sure there are no empty strings.
	return remoteDocs
}

func main() {
	flag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "very verbose debugging output")
	flag.IntVarP(&monthsThreshold, "months", "m", 1, "months threshold for schedule files")
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s\n", os.Args[0], lastModified, runtime.Version())
		fmt.Printf(" Usage: %s [weelky xlsx file] \n", os.Args[0])
		fmt.Printf(" Needs lint.conf or lint.ini, and looks in current, home and config directories.\n")
		fmt.Printf(" first line must begin with off, and 2nd line, if present, must begin with startdirectory.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if veryVerboseFlag {
		*verboseFlag = true
	}

	filepicker.VerboseFlag = *verboseFlag

	var filename, ans string

	fmt.Printf(" lint for the weekly schedule last modified %s\n", lastModified)

	err := findAndReadConfIni()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from findAndReadConfINI: %s.  Exiting.\n", err)
		return
	}

	if *verboseFlag {
		fmt.Printf(" After findAndReadConfIni, Start Directory: %s\n", startDirectory)
	}

	// filepicker stuff.

	includeODrive := !*conlyFlag // a comvenience flag
	if *verboseFlag {
		fmt.Printf(" conlyFlag=%t, includeODrive=%t\n", *conlyFlag, includeODrive)
	}
	if flag.NArg() == 0 {
		var filenames []string
		if includeODrive {
			filenames, err = walkRegexFullFilenames() // function is below.  "o:\\week.*xls.?$"
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from filepicker is %s.  Exiting \n", err)
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
		filenamesDocs, err := filepicker.GetRegexFullFilenames(docs)
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

	if *verboseFlag {
		fmt.Printf(" spreadsheet picked is %s\n", filename)
	}
	fmt.Println()

	err = scanXLSfile(filename)

	if err == nil {
		ctfmt.Printf(ct.Green, true, "\n\n Finished scanning %s\n\n", filename)
	} else {
		ctfmt.Printf(ct.Red, true, "\n\n Error scanning %s\n\n", filename)
	}
}

// scanXLSfile -- takes a filename and checks for 3 errors; vacation people assigned to work, fluoro also late person, and fluoro also remote person.
func scanXLSfile(filename string) error {

	workBook, err := xlsx.OpenFile(filename)
	if err != nil {
		//fmt.Printf("Error opening excel file %s in directory %s: %s\n", filename, workingDir, err)
		return err
	}

	// Populate the week's schedule
	var week [6]dayType      // Only need 5 workdays.  Element 0 is not used.
	for i := 1; i < 6; i++ { // Monday = 1, Friday = 5
		week[i], err = readDay(workBook, i) // the subscripts are reversed, as a column represents a day.  Each row is a different subspeciality.
		if err != nil {
			fmt.Printf("Error reading day %d: %s, skipping\n", i, err)
			continue
		}
	}

	if veryVerboseFlag {
		for i, day := range week {
			fmt.Printf("Day %d: %#v \n", i, day)
		}
	}

	// Who's on vacation for each day, and then check the rest of that day to see if any of these names exist in any other row.
	for dayCol := 1; dayCol < len(week); dayCol++ { // col 0 is empty and does not represent a day, dayCol 1 is Monday, ..., dayCol 5 is Friday
		mdsOffToday := whosOnVacationToday(week, dayCol)
		lateDocsToday := whosLateToday(week, dayCol)

		if veryVerboseFlag {
			fmt.Printf(" mdsOffToday on day %d is/are %#v\n", dayCol, mdsOffToday)
			i := 0
			for doc, vacay := range dayOff {
				if i%10 == 9 {
					fmt.Printf("\n")
				}
				fmt.Printf(" dayOff[%s]: %t, ", doc, vacay)
				i++
			}
			fmt.Printf("\n")
			if pause() {
				return errors.New("exit from pause")
			}

			fmt.Printf("\n Late shift docs on day %d are %#v\n", dayCol, lateDocsToday)
		}

		// Now, mdsOffToday is a slice of several names of who is off today.

		for _, name := range mdsOffToday {
			for i := weekdayOncall; i < mdOff; i++ { // since mdoff is the last one, can test for < mdOff.  Don't test against MD off as we already know whose off that day.
				if lower := strings.ToLower(week[dayCol][i]); strings.Contains(lower, name) {
					fmt.Printf(" %s is off on %s, but is on %s\n", strcase.UpperCamelCase(name), dayNames[dayCol], categoryNamesList[i])
				}
			}
		}

		// Now, lateDocsToday is a slice of two names of who is covering the late shift today.  Only checks against fluoro, as that's not good scheduling
		for _, name := range lateDocsToday {
			if lower := strings.ToLower(week[dayCol][fluoroJH]); strings.Contains(lower, name) {
				ctfmt.Printf(ct.Cyan, true, " %s is late on %s, but is on fluoro JH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol][fluoroFH]); strings.Contains(lower, name) {
				ctfmt.Printf(ct.Cyan, true, " %s is late on %s, but is on fluoro FH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
		}

		// Determine if the fluoro doc for today is remote
		remoteNames := whosRemoteToday(week, dayCol)
		for _, name := range remoteNames {
			if veryVerboseFlag {
				fmt.Printf(" Remote doc for today: %s, FluoroJH: %s, FluoroFH: %s\n", name, week[dayCol][fluoroJH], week[dayCol][fluoroFH])
			}
			if lower := strings.ToLower(week[dayCol][fluoroJH]); strings.Contains(lower, name) {
				ctfmt.Printf(ct.Yellow, true, " %s is remote on %s, but is on fluoro JH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol][fluoroFH]); strings.Contains(lower, name) {
				ctfmt.Printf(ct.Yellow, true, " %s is remote on %s, but is on fluoro FH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
		}

		if veryVerboseFlag {
			if pause() {
				return errors.New("exit from pause")
			}
		}
	}
	return nil
}

// getRegexFullFilenames -- uses a regular expression to determine a match, by using regex.MatchString.  Processes directory info and uses dirEntry type.
//
//	Needs a walk function to find what it is looking for.  See top comments.
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

func pause() bool {
	var ans string
	fmt.Printf(" Pausing.  Stop [y/N]: ")
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return strings.HasPrefix(ans, "y") // suggested by staticcheck.
}
