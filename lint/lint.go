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
				MD's Off (vacation)
				neuro
				body
				ER
				Xrays
				IR
				Nuclear
				US
				Pediatrics
				FLUORO JH
				FLUORO FH
				MSK
				MAMMO
				BONE (DENSITY)
				LATE
				if the line begins w/ # ; / then it's a comment.  If a line doesn't begin w/ a keyword, then it's an error and the pgm exits.
				I think I'll just check the vacation rule first.  Then expand it to the other rules.

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

var day dayType
var verboseFlag = flag.Bool("v", false, "Verbose mode")
var home string
var config string
var err error
var workingDir string

func findAndReadConfInit() error {
	// will search first for conf and then for ini file in this order of directories: current, home, config.
	// It will populate the dictionary, dict.

}

func readDay(idx int) error {

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

	err = findAndReadConfInit()
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

	sheet, err := xlsx.OpenFile(filename)
	if err != nil {
		fmt.Printf("Error opening excel file %s in directory %s: %s\n", filename, workingDir, err)
		return
	}

	// this is for demo purposes.  I need to understand this better.
	fmt.Println("Sheets in this file:")
	for i, sh := range sheet.Sheets {
		fmt.Println(i, sh.Name)
	}
	fmt.Println("----")
}
