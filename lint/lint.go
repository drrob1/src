package main // lint.go, from lint2.go from lint.go

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"src/misc"
	"src/tknptr"
	"src/whichexec"
	"strconv"
	"strings"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"github.com/stoewer/go-strcase"
	"github.com/tealeg/xlsx/v3"
	"github.com/umahmood/soundex"
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
				I think I can do it without much difficulty.
  29 Jan 25 -- I'm going to make the week a 2D array and use a map to get the names from the row #.  Just to see if it will work.  It does.
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
  10 Apr 25 -- Made the change by adding an else clause to an if statement in the walk function.
   8 May 25 -- Fixed the help message.
  26 May 25 -- When I installed this on Caity's computer, I got the idea that I should filter out the files that Excel controls, i.e., those that begin w/ ~, tilda.
  31 May 25 -- Changed an error message.
   2 Aug 25 -- Completed getdocnames yesterday which extracts the names from the schedule itself.  This way I don't need to specify them in the config file, making this routine
				more robust.  I'm going to add it.  I have to ignore the line of the config file that begins w/ "off".  I have to check what the code does if there is no
				startdirectory line in the config file, or the line has invalid syntax (like not beginning w/ the correctly spelled keyword).
				Processing the config file used to do so by using global var's.  I'm changing that to use params.  This way, I can ignore a return param if I want to.
				I'm tagging this as lint v2.0
   3 Aug 25 -- walk function will skip .git
   4 Aug 25 -- Our 40th Anniversary.  But that's not important now.  I'm using soundex codes to report likely spelling errors so they can be fixed.
   6 Aug 25 -- I found out today that the hospital will retire the o: drive, in favor of OneDrive.  I'll need to change the code to use OneDrive.
               There's an environment varible called OneDrive that is set to the path of OneDrive.  And another one called OneDriveConsumer.
               At work, there's OneDriveCommercial, which is set to the same value as OneDrive.  This is also true at home in that OneDrive and OneDriveConsumer have the same value.
				I first coded this to use a filepicker function, but that doesn't exclude old files.  The walk function will skip files that are older than the threshold.
				I need to modify the walk function to take a param that is the start directory, and then combine the results of all the walk function calls.
                And O: drive is going away at work as of Aug 8, 2025.  I'll need to change the code to use OneDrive.  I'll remove the conly flag as it's not needed now.
  11 Aug 25 -- Time to add the code to autoupdate.
*/

const lastModified = "11 Aug 2025"
const conf = "lint.conf"
const ini = "lint.ini"
const numOfDocs = 40 // used to dimension a string slice.
const maxDimensions = 200

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

type soundexSlice struct {
	s      string
	soundx string
}

var names = make([]string, 0, numOfDocs)
var categoryNamesList = []string{"0", "1", "2", "weekday On Call", "Neuro", "Body", "ER/Xrays", "IR", "Nuclear Medicine", "US", "Peds", "Fluoro JH", "Fluoro FH",
	"MSK", "Mammo", "Bone Density", "late", "weekend moonlighters", "weekend JH", "weekend FH", "weekend IR", "MD's Off"} // 0, 1 and 2 are unused

var dayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var dayOff = make(map[string]bool) // only used in findAndReadConfIni when verboseFlag is set

var verboseFlag = flag.BoolP("verbose", "v", false, "Verbose mode")

// var conlyFlag = flag.BoolP("conly", "c", false, "Conly mode, ie, search only Documents on c: for current user.")  Not relevant as o: drive is gone.
var numLines = 15 // I don't expect to need more than these, as I display only the first 26 elements (a-z) so far.
var veryVerboseFlag bool
var monthsThreshold int
var startDirFromConfigFile string // this needs to be a global, esp for the walk function.

// findAndReadConfIni now returns a string slice of the docNames it found, a string representing the startdirectory, and an error.
func findAndReadConfIni() ([]string, string, error) {
	// will search first for conf and then for ini file in this order of directories: current, home, config.
	fullFile, found := whichexec.FindConfig(conf)
	if !found {
		fullFile, found = whichexec.FindConfig(ini)
		if !found {
			if *verboseFlag {
				return nil, "", fmt.Errorf("%s or %s not found", conf, ini)
			} else {
				return nil, "", nil
			}
		}
	}

	docNames := make([]string, 0, numOfDocs) // a list of all the doc's last names as read from the config file.
	var startDirectory string
	// now need to process the config file using code from fansha.
	fileByteSlice, err := os.ReadFile(fullFile)
	if err != nil {
		return nil, "", err
	}
	bytesReader := bytes.NewReader(fileByteSlice)
	inputLine, err := misc.ReadLine(bytesReader)
	if err != nil {
		return nil, "", err
	}

	// remove any commas
	inputLine = strings.ReplaceAll(inputLine, ",", "")

	// will use my tknptr stuff here.
	tokenslice := tknptr.TokenSlice(inputLine)
	paramName := strings.ToLower(tokenslice[0].Str)
	//if !strings.Contains(lower, "off") {
	//	return fmt.Errorf("%s is not off", tokenslice[0].Str)
	//}
	if paramName == "off" {
		if veryVerboseFlag {
			fmt.Printf(" Found config file: %s, containing %s\n", fullFile, inputLine)
		}
		for _, token := range tokenslice[1:] {
			lower := strings.ToLower(token.Str)
			docNames = append(docNames, lower)
			dayOff[lower] = false // this is a map[string]bool
		}
		if veryVerboseFlag {
			fmt.Printf(" Loaded config file: %s, containing %s\n", fullFile, inputLine)
			for doc, vacay := range dayOff {
				fmt.Printf(" dayOff[%s]: %t, ", doc, vacay)
			}
			fmt.Println()
			fmt.Printf(" Names unsorted: %#v\n", docNames)
		}

		sort.Strings(docNames)

		if *verboseFlag {
			fmt.Printf(" Sorted Names: %#v\n\n", docNames)
		}
	} else if paramName == "startdirectory" { // only have 1 line in the config file, and it's for startdirectory
		trimmedInputLine, ok := strings.CutPrefix(inputLine, "startdirectory") // CutPrefix became available as of Go 1.20
		if ok {
			startDirectory = trimmedInputLine
			startDirectory = strings.TrimSpace(startDirectory)
		}
		return docNames, startDirectory, nil
	} else {
		return nil, "", fmt.Errorf("first line of config file must be 'off' or 'startdirectory'")
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
			return docNames, "", nil
		}
		return docNames, "", err
	}

	trimmedInputLine, ok := strings.CutPrefix(inputLine, "startdirectory") // CutPrefix became available as of Go 1.20
	if ok {
		startDirectory = trimmedInputLine
		startDirectory = strings.TrimSpace(startDirectory)
	}

	return docNames, startDirectory, nil
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

	mdsOff := make([]string, 0, numOfDocs) // Actually, never more than 10 off, but religious holidays can have a lot off.

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
	var err error
	var noUpgradeLint bool
	flag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "very verbose debugging output")
	flag.IntVarP(&monthsThreshold, "months", "m", 1, "months threshold for schedule files")
	flag.BoolVarP(&noUpgradeLint, "noupgrade", "n", false, "do not upgrade schedule files")
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], lastModified, runtime.Version())
		fmt.Printf(" Usage: %s <weekly xlsx file> \n", os.Args[0])
		fmt.Printf(" Needs lint.conf or lint.ini, and looks in current, home and config directories.\n")
		fmt.Printf(" first line must begin with off, and 2nd line, if present, must begin with startdirectory.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if veryVerboseFlag {
		*verboseFlag = true
	}

	// filepicker.VerboseFlag = *verboseFlag  no longer using filepicker.  I'm now using walkRegexFullFilenames.

	var filename, ans string

	fmt.Printf(" lint V 2.1 for the weekly schedule, last modified %s\n", lastModified)

	_, startDirFromConfigFile, err = findAndReadConfIni() // ignore the doc names list from the config file, as that's now extracted from the schedule itself.
	if err != nil {
		if *verboseFlag { // only show this message if verbose flag is set.  Otherwise, it's too much.
			fmt.Printf(" Warning from findAndReadConfIni: %s.  Ignoring. \n", err)
			ctfmt.Printf(ct.Red, true, " Warning message from findAndReadConfINI: %s. \n", err)
			//   return  No longer need the names from the file.  And don't absolutely need startDirectory.
		}
	}
	if *verboseFlag {
		fmt.Printf(" After findAndReadConfIni, Start Directory: %s\n", startDirFromConfigFile)
	}

	// filepicker no longer used.  Now I'm using the walkRegexFullFilenames function.

	//filenames, err := filepicker.GetRegexFullFilenames(docs)
	//if err != nil {
	//	ctfmt.Printf(ct.Red, true, " Error from filepicker.GetRegexFullFilenames is %s.  Exiting \n", err)
	//	return
	//}

	//includeODrive := !*conlyFlag // a comvenience flag
	//if *verboseFlag {
	//	fmt.Printf(" conlyFlag=%t, includeODrive=%t\n", *conlyFlag, includeODrive)
	//}
	var filenames []string
	var homeDir string
	var docs string

	if flag.NArg() == 0 {
		//if includeODrive {  O: drive is gone as of 8/8/25.
		//	filenames, err = walkRegexFullFilenames() // function is below.  "o:\\week.*xls.?$"
		//	if err != nil {
		//		ctfmt.Printf(ct.Red, false, " Error from walkRegexFullFilenames is %s.  Exiting \n", err)
		//		return
		//	}
		//	if *verboseFlag {
		//		fmt.Printf(" Filenames length from o drive: %d\n", len(filenames))
		//	}
		//}

		if startDirFromConfigFile != "" {
			filenamesStartDir, err := walkRegexFullFilenames(startDirFromConfigFile)
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from walkRegexFullFilenames(%s) is %s.  Ignoring. \n",
					startDirFromConfigFile, err)
			}
			if len(filenamesStartDir) > 0 {
				filenames = append(filenames, filenamesStartDir...)
				if *verboseFlag {
					fmt.Printf(" Filenames length after append %s: %d\n", filenamesStartDir, len(filenames))
				}
			}
			if *verboseFlag {
				fmt.Printf(" Filenames length after append %s: %d\n", filenamesStartDir, len(filenames))
			}
		}

		homeDir, err = os.UserHomeDir()
		if err != nil {
			fmt.Printf(" Error from os.UserHomeDir: %s\n", err)
			return
		}
		// docs = filepath.Join(filepath.Join(homeDir, "Documents"), "week.*xls.?$")  Don't want the regex as part of this expression.
		docs = filepath.Join(homeDir, "Documents")
		if *verboseFlag {
			fmt.Printf(" homedir=%q, Joined Documents: %q\n", homeDir, docs)
		}
		//filenamesDocs, err := walkRegexFullFilenames(docs)
		//if err != nil {
		//	fmt.Printf(" Error from walkRegexFullFilenames(%s) is %s.  Ignored \n", docs, err)
		//} else {
		//	filenames = append(filenames, filenamesDocs...)
		//}
		//oneDrive := filepath.Join(filepath.Join(homeDir, "OneDrive"), "week.*xls.?$")
		//if *verboseFlag {
		//	fmt.Printf(" oneDrive=%q, OneDriveWork: %q\n", oneDrive, oneDriveWork)
		//}
		oneDriveString := os.Getenv("OneDrive")
		if *verboseFlag {
			fmt.Printf(" oneDriveString = %s  \n", oneDriveString)
		}
		//                                                  filenamesOneDrive, err := filepicker.GetRegexFullFilenames(docs)
		//                                         filenamesOneDrive, err := filepicker.GetRegexFullFilenamesNotLocked(docs)
		filenamesOneDrive, err := walkRegexFullFilenames(oneDriveString)
		if err != nil {
			fmt.Printf(" Error from walkRegexFullFilenames(%s) is %s.  Ignoring \n", oneDriveString, err)
		}
		if *verboseFlag {
			fmt.Printf(" FilenamesDocs length: %d\n", len(filenamesOneDrive))
		}

		filenames = append(filenames, filenamesOneDrive...)
		if *verboseFlag {
			fmt.Printf(" Filenames length after append %s: %d\n", filenamesOneDrive, len(filenames))
		}

		filenamesDocs, err := walkRegexFullFilenames(docs)
		if err != nil {
			fmt.Printf(" Error from walkRegesFullFilenames(%s) is %s.  Ignored \n", docs, err)
		} else {
			filenames = append(filenames, filenamesDocs...)
			if *verboseFlag {
				fmt.Printf(" Filenames length after append %s: %d\n", docs, len(filenames))
			}
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

	names, err = getDocNames(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from getDocNames: %s.  Exiting \n", err)
		return
	}
	if *verboseFlag {
		fmt.Printf(" doc names extracted from %s length: %d\n", filename, len(names))
		fmt.Printf(" names: %#v\n\n", names)
	}

	// detecting and reporting likely spelling errors based on the soundex algorithm

	soundx := getSoundex(names)
	spellingErrors := showSpellingErrors(soundx)
	if len(spellingErrors) > 0 {
		ctfmt.Printf(ct.Cyan, true, "\n\n %d spelling error(s) detected in %s: ", len(spellingErrors)/2, filename)
		for _, spell := range spellingErrors {
			ctfmt.Printf(ct.Red, true, " %s  ", spell)
		}
		fmt.Printf("\n\n\n")
	}

	// scan the xlsx schedule file

	err = scanXLSfile(filename)

	if err == nil {
		ctfmt.Printf(ct.Green, true, "\n\n Finished scanning %s\n\n", filename)
	} else {
		ctfmt.Printf(ct.Red, true, "\n\n Error scanning %s\n\n", filename)
	}

	if noUpgradeLint {
		return
	}

	// Time to run the updatelist cmd.

	execCmd := exec.Command("upgradelint", "")
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	err = execCmd.Start()
	if err != nil {
		ctfmt.Printf(ct.Red, true, "\n\n Error starting upgradelint: %s.  Contact Rob Solomon\n\n", err)
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
func walkRegexFullFilenames(startdirectory string) ([]string, error) { // This rtn sorts using sort.Slice

	if startdirectory == "" {
		return nil, errors.New("startdirectory is empty")
	}

	if *verboseFlag {
		fmt.Printf(" walkRegexFullFilenames, startdirectory: %s\n", startdirectory)
	}

	// compile the regular expression
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
				fmt.Printf(" WalkDir fpath %s, de.Name %s, err %v \n", fpath, de.Name(), err.Error())
			} else {
				fmt.Printf(" WalkDir fpath %s, de.Name %s\n", fpath, de.Name())
			}
		}
		if err != nil {
			if *verboseFlag {
				fmt.Printf(" Error from walkDirFunction: %s\n", err)
			}
			return filepath.SkipDir
		}
		if de.IsDir() {
			if *verboseFlag {
				fmt.Printf(" de.IsDir() is true, fpath = %q, de.Name=%s\n", fpath, de.Name())
			}
			return nil // allow walk function to drill down itself
		}
		if de.Name() == ".git" { // only if full directory name is .git, then skip this directory.  This is a hack.  I don't want to skip the entire directory.
			if *verboseFlag {
				fmt.Printf(" fpath contains .git, fpath = %q, de.Name=%s\n", fpath, de.Name())
			}
			return filepath.SkipDir
		}

		if time.Now().After(timeout) {
			return errors.New("timeout occurred")
		}

		// Not a directory, and timeout has not happened.  Only process regular files, and skip symlinks.  Is this my problem?  Yes, it is.  I do want symlinks, after all.
		//if !de.Type().IsRegular() {
		//	if *verboseFlag {
		//		fmt.Printf(" de.Type().IsRegular() is false, fpath = %q, de.Name=%s\n", fpath, de.Name())
		//	}
		//	return nil
		//}

		// now de.Name is a regular file.
		lowerName := strings.ToLower(de.Name())
		if regex.MatchString(lowerName) {
			if strings.HasPrefix(de.Name(), "~") { // ignore files that Excel has locked.
				return nil
			}
			fi, err := de.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from dirEntry(%s).Info() is %v \n", de.Name(), err)
				return err
			}

			filedata := fileDataType{
				ffname:    fpath, // filepath.Join(fpath, de.Name()) doesn't work, because it turns out that fpath is the full path and filename
				timestamp: fi.ModTime(),
			}
			if *verboseFlag {
				fmt.Printf(" matches the regexp: filedata.timestamp is %s\n", filedata.timestamp)
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

	//if startDirectory == "" { // if not set by the config file.  As of Aug 8, 2025, the O: drive is going away.  So I'm not going to use it.
	//	if runtime.GOOS == "windows" { // Variable is defined globally.
	//		startDirectory = "o:\\Nikyla's File\\RADIOLOGY MD Schedule\\"
	//	} else { // this is so I can debug on leox, too.  Variable is defined globally.
	//		startDirectory = "/home/rob/bigbkupG/Nikyla's File/RADIOLOGY MD Schedule"
	//	}
	//}

	if *verboseFlag {
		fmt.Printf(" WalkDir startDirectory: %s\n", startdirectory)
	}

	err = filepath.WalkDir(startdirectory, walkDirFunction)
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

func excludeMe(s string) bool { // used by getDocNames
	dgtRegexp := regexp.MustCompile(`\d`) // any digit character will match this exprn.
	if strings.Contains(s, "fh") || strings.Contains(s, "dr.") || strings.Contains(s, "(") || strings.Contains(s, ")") || strings.Contains(s, "/") ||
		strings.Contains(s, "jh") || strings.Contains(s, "plain") || strings.Contains(s, "please") || strings.Contains(s, "sat") ||
		strings.Contains(s, "see") || strings.Contains(s, "sun") || strings.Contains(s, "thu") || strings.Contains(s, "modality") ||
		strings.Contains(s, ":") || strings.Contains(s, "*") || strings.Contains(s, "@") || dgtRegexp.MatchString(s) {
		return true
	}
	return false
}

// getDocNames -- takes a filename and returns a slice of doc names extracted from the Excel weekly schedule file.  The slice is sorted.  The slice is sorted by the first word of the doc name.
func getDocNames(fn string) ([]string, error) {
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
	sort.Strings(docNamesSlice)
	docNamesSlice = slices.Compact(docNamesSlice) // slices package became available with Go 1.20-ish
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

func showSpellingErrors(in []soundexSlice) []string {
	out := make([]string, 0, len(in))

	for i := 1; i < len(in); i++ {
		if in[i].soundx == in[i-1].soundx { // one of these is likely a spelling error
			out = append(out, in[i-1].s)
			out = append(out, in[i].s)
		}
	}
	return out
}
