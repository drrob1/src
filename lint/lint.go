package main // lint.go

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/tealeg/xlsx/v3"
	"os"
	"src/filepicker"
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
*/

const lastModified = "30 Sep 2024"
const conf = "lint.conf"
const ini = "ini.conf"

type list struct {
	category string
	docs     []string
}

var dict map[string]list // dictionary of categories and doc names that belong in the list of that category.

type dayType struct {
	neuro       string
	body        string
	er          string
	xrays       string
	ir          string
	nuclear     string
	us          string
	peds        string
	fluoroJH    string
	fluoroFH    string
	msk         string
	mammo       string
	boneDensity string
	late        string
}

var week []dayType
var categoryNamesList = []string{"md's off", "neuro", "body", "er", "xrays", "ir", "nuclear medicine", "us", "fluoro jh", "fluoro fh", "msk", "mammo", "bone density", "late"}
var day dayType
var verboseFlag = flag.Bool("v", false, "Verbose mode")
var home string
var config string
var err error
var workingDir string

// Next I will code the check against the vacation people to make sure they're not assigned to anything else.  I'll need a vacationMap = map[string]bool where the string will
// be the names of everyone, and true/false for on vacation.  I'll need a doctor names list, I think.

func findAndReadConfIni() error {
	// will search first for conf and then for ini file in this order of directories: current, home, config.
	// It will populate the dictionary, dict.

	fmt.Printf("findAndReadConfIni not done yet\n")
	return nil
}

func readDay(idx int) error {
	panic("not done yet")
}

func main() {
	flag.Parse()

	var filename string

	fmt.Printf(" lint for the weekly schedule last modified %s\n", lastModified)

	home, err = os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %s\n", err)
		return
	}
	config, err = os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error getting user config dir: %s\n", err)
		return
	}
	workingDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %s\n", err)
		return
	}

	err = findAndReadConfIni()
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return
	}

	// filepicker stuff.

	var ans string
	if flag.NArg() == 0 {
		filenames, err := filepicker.GetFilenames("*.xlsx")
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

	// this is for demo purposes.  I need to understand this better.
	fmt.Println("Sheets in this file:")
	for i, sh := range workBook.Sheets {
		fmt.Println(i, sh.Name)
	}
	fmt.Println("----")

	sheets := workBook.Sheets
	fmt.Printf(" sheet contains %d sheets, and len(sheets) = %d\n", len(workBook.Sheets), len(sheets))
	row, err := sheets[0].Row(21)
	if err != nil {
		fmt.Printf("Error getting row 0: %s\n", err)
		return
	}
	cellr21c0 := row.GetCell(0)
	cellr21c1 := row.GetCell(1)
	cellr21c2 := row.GetCell(2)
	fmt.Printf(" row 21 c0 = %q, maxrow = %d, row 21 c1 = %q, row 21 c 2 = %q\n", cellr21c0, sheets[0].MaxRow, cellr21c1, cellr21c2)
	cell021, _ := sheets[0].Cell(0, 21)
	cell121, _ := sheets[0].Cell(1, 21)
	cell210, _ := sheets[0].Cell(21, 0)
	fmt.Printf(" Cell r0 c21 = %q, cell r1 c21 = %q, cell r21 c0 = %q\n", cell021, cell121, cell210)

	irCellr7c0, _ := sheets[0].Cell(7, 0)
	irCellr7c0lower := strings.ToLower(irCellr7c0.String())
	irCellr7c1, _ := sheets[0].Cell(7, 1)
	irCellr7c1lower := strings.ToLower(irCellr7c1.String())
	fmt.Printf(" IR Cell r7 c0 = %q, IR Cell r7 c1 = %q \n r7 c0 lower = %q, r7 c1 lower = %q\n", irCellr7c0, irCellr7c1, irCellr7c0lower, irCellr7c1lower)
}
