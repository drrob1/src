package main // lint.go, from lint2.go from lint.go

import (
	"bytes"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/stoewer/go-strcase"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"src/filepicker"
	"src/misc"
	"src/tknptr"
	"src/whichexec"
	"strconv"
	"strings"
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

*/

const lastModified = "15 Mar 2025"
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

type dayType [23]string // there are a few unused entries here.  This goes from 0 .. 22.  Indices 0..2 are not used.

var categoryNamesList = []string{"0", "1", "2", "weekday On Call", "Neuro", "Body", "ER/Xrays", "IR", "Nuclear Medicine", "US", "Peds", "Fluoro JH", "Fluoro FH",
	"MSK", "Mammo", "Bone Density", "late", "weekend moonlighters", "weekend JH", "weekend FH", "weekend IR", "MD's Off"} // 0, 1 and 2 are unused

var dayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var workingDir string
var names = make([]string, 0, 25)  // a list of all the doc's last names as read from the config file.
var dayOff = make(map[string]bool) // only used in findAndReadConfIni when verboseFlag is set

var verboseFlag = flag.Bool("v", false, "Verbose mode")

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
	if *verboseFlag {
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
	flag.Parse() // need this because of use of flag.NArg() below

	filepicker.VerboseFlag = *verboseFlag

	var filename, ans string

	fmt.Printf(" lint for the weekly schedule last modified %s\n", lastModified)

	//sectionNames[4] = "neuro"
	//sectionNames[5] = "body"
	//sectionNames[6] = "ER/Xrays"
	//sectionNames[7] = "IR"
	//sectionNames[8] = "Nuclear"
	//sectionNames[9] = "US"
	//sectionNames[10] = "Peds"
	//sectionNames[11] = "FluoroJH"
	//sectionNames[12] = "FluoroFH"
	//sectionNames[13] = "MSk"
	//sectionNames[14] = "Mammo"
	//sectionNames[15] = "BoneDensity"
	//sectionNames[16] = "Late"
	//sectionNames[17] = "Moonlighters"
	//sectionNames[18] = "WeekendJH"
	//sectionNames[19] = "WeekendFH"
	//sectionNames[20] = "WeekendIR"
	//sectionNames[21] = "MDOff"

	err := findAndReadConfIni()
	if err != nil {
		fmt.Printf(" Error from findAndReadConfINI: %s\n", err)
		fmt.Printf(" Continue? (Y/n)")
		fmt.Scanln(&ans)
		ans = strings.ToLower(ans)
		if strings.Contains(ans, "n") {
			return
		}
	}

	// filepicker stuff.

	if flag.NArg() == 0 {
		filenames, err := filepicker.GetRegexFullFilenames("w:\\week.*xlsx$")
		if err != nil {
			ctfmt.Printf(ct.Red, false, " Error from filepicker is %s.  Exiting \n", err)
			return
		}
		if *verboseFlag {
			fmt.Printf(" Filenames length: %d\n", len(filenames))
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf(" Error from os.UserHomeDir: %s\n", err)
			return
		}
		docs := filepath.Join(filepath.Join(homeDir, "Documents"), "week.*xlsx$")
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
			fmt.Println(" Stop code entered.  Exiting.")
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

	scanXLSfile(filename)

	ctfmt.Printf(ct.Green, true, "\n\n Finished scanning %s\n", filename)
}

// scanXLSfile -- takes a filename and checks for 3 errors; vacation people assigned to work, fluoro also late person, and fluoro also remote person.
func scanXLSfile(filename string) {

	workBook, err := xlsx.OpenFile(filename)
	if err != nil {
		fmt.Printf("Error opening excel file %s in directory %s: %s\n", filename, workingDir, err)
		return
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

	if *verboseFlag {
		for i, day := range week {
			fmt.Printf("Day %d: %#v \n", i, day)
		}
	}

	// Who's on vacation for each day, and then check the rest of that day to see if any of these names exist in any other row.
	for dayCol := 1; dayCol < len(week); dayCol++ { // col 0 is empty and does not represent a day, dayCol 1 is Monday, ..., dayCol 5 is Friday
		mdsOffToday := whosOnVacationToday(week, dayCol)
		lateDocsToday := whosLateToday(week, dayCol)

		if *verboseFlag {
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
				return
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
			if *verboseFlag {
				fmt.Printf(" Remote doc for today: %s, FluoroJH: %s, FluoroFH: %s\n", name, week[dayCol][fluoroJH], week[dayCol][fluoroFH])
			}
			if lower := strings.ToLower(week[dayCol][fluoroJH]); strings.Contains(lower, name) {
				ctfmt.Printf(ct.Yellow, true, " %s is remote on %s, but is on fluoro JH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol][fluoroFH]); strings.Contains(lower, name) {
				ctfmt.Printf(ct.Yellow, true, " %s is remote on %s, but is on fluoro FH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
		}

		if *verboseFlag {
			if pause() {
				return
			}
		}
	}
}

func pause() bool {
	var ans string
	fmt.Printf(" Pausing.  Stop [y/N]: ")
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return strings.HasPrefix(ans, "y") // suggested by staticcheck.
}
