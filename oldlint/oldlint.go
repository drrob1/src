package main // oldlint.go from lint.go

import (
	"bytes"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/stoewer/go-strcase"
	"github.com/tealeg/xlsx/v3"
	"os"
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
				MD's Off (vacation) row 21
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
				if the line begins w/ # ; / then it's a comment.  If a line doesn't begin w/ a keyword, then it's an error and the pgm exits.
				I think I'll just check the vacation rule first.  Then expand it to the other rules.

                I have to read the weekly schedule into an appropriate data structure, as also the .conf/.ini file.

 xlsx (github.com/tealeg/xlsx/v3)

  28 Jan 25 -- I want to add detection of having the late doc also be on fluoro.  That happened today at Flushing, and it was a mistake.
				I think I can do it without much difficulity.
  29 Jan 25 -- Copied lint2 code onto lint.go.  Preserved old version of lint.go as lint-2Oct24.go
------------------------------------------------------------------------------------------------------------------------------------------------------
  31 Jan 25 -- Renamed to oldlint.go
*/

const lastModified = "31 Jan 2025"
const conf = "lint.conf"
const ini = "lint.ini"

type dayType struct {
	weekdayOncall string
	neuro         string
	body          string
	er            string
	xrays         string
	ir            string
	nuclear       string
	us            string
	peds          string
	fluoroJH      string
	fluoroFH      string
	msk           string
	mammo         string
	boneDensity   string
	late          string
	moonlighters  string
	weekendJH     string
	weekendFH     string
	weekendIR     string
	mdOff         string
}

var dayNames = []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var workingDir string
var dayOff = make(map[string]bool) // not sure if I need this, but here it is.
var names = make([]string, 0, 25)  // a list of all the doc's last names as read from the config file.

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
		dayOff[lower] = false
		names = append(names, lower)
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

	for i := 3; i < 22; i++ {
		cell, err := sheets[0].Cell(i, col) // always sheet[0]
		if err != nil {
			return dayType{}, err
		}
		switch i {
		case 3:
			day.weekdayOncall = cell.String()
		case 4:
			day.neuro = cell.String()
		case 5:
			day.body = cell.String()
		case 6:
			day.er = cell.String()
			day.xrays = cell.String()
		case 7:
			day.ir = cell.String()
		case 8:
			day.nuclear = cell.String()
		case 9:
			day.us = cell.String()
		case 10:
			day.peds = cell.String()
		case 11:
			day.fluoroJH = cell.String()
		case 12:
			day.fluoroFH = cell.String()
		case 13:
			day.msk = cell.String()
		case 14:
			day.mammo = cell.String()
		case 15:
			day.boneDensity = cell.String()
		case 16:
			day.late = cell.String()
		case 17:
			day.moonlighters = cell.String()
		case 18:
			day.weekendJH = cell.String()
		case 19:
			day.weekendFH = cell.String()
		case 20:
			day.weekendIR = cell.String()
		case 21:
			day.mdOff = cell.String()

		default:
			return dayType{}, fmt.Errorf("unknown day type %d", i)

		}
	}
	return day, nil
}

func whosOnVacationToday(week [6]dayType, dayCol int) []string { // week is an array, not a slice.  It doesn't need a slice.
	// this function is to return a slice of names that are on vacation for this day
	vacationCell := week[dayCol].mdOff // row 21 is the MD's Off row.  It is a string containing multiple names.  I don't need to divide them into separate strings
	vacationString := vacationCell
	vacationString = strings.ToLower(vacationString)

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
	lateCell := week[dayCol].late // row 16 is the late doc's row.  It is a string containing the 2 names that are late that day
	lateString := lateCell
	lateString = strings.ToLower(lateString)

	lateDocs := make([]string, 0, 2)
	// search for matching names
	for _, lateName := range names { // names is a global
		if strings.Contains(lateString, lateName) {
			lateDocs = append(lateDocs, lateName)
		}
	}
	return lateDocs
}

func main() {
	flag.Parse() // need this because of use of flag.NArg() below

	var filename, ans string

	fmt.Printf(" lint for the weekly schedule last modified %s\n", lastModified)

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
		filenames, err := filepicker.GetRegexFilenames("week.*xlsx$")
		if err != nil {
			ctfmt.Printf(ct.Red, false, " Error from filepicker is %s.  Exiting \n", err)
			return
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
			if lower := strings.ToLower(week[dayCol].weekdayOncall); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on weekday On call\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].neuro); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on neuro\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].body); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on body\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].er); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on ER Xrays\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].ir); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on IR\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].nuclear); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on Nuclear\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].us); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on US\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].peds); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on peds\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].fluoroJH); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on fluoro JH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].fluoroFH); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on fluoro FH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].msk); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on MSK\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].mammo); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on mammo\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].boneDensity); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on bone density\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].late); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on late\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].moonlighters); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on weekend moonlighters\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].weekendJH); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on weekend JH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].weekendFH); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on weekend FH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].weekendIR); strings.Contains(lower, name) {
				fmt.Printf(" %s is off on %s, but is on weekend IR\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
		}

		// Now, lateDocsToday is a slice of two names of who is covering the late shift today.  Only checks against fluoro, as that's not good scheduling
		for _, name := range lateDocsToday {
			if lower := strings.ToLower(week[dayCol].fluoroJH); strings.Contains(lower, name) {
				fmt.Printf(" %s is late on %s, but is on fluoro JH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
			}
			if lower := strings.ToLower(week[dayCol].fluoroFH); strings.Contains(lower, name) {
				fmt.Printf(" %s is late on %s, but is on fluoro FH\n", strcase.UpperCamelCase(name), dayNames[dayCol])
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
