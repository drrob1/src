package main // lint.go
import (
	"flag"
	"fmt"
	"os"
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
*/

const lastModified = "30 Sep 2024"
const conf = "lint.conf"
const ini = "ini.conf"

type list struct {
	category string
	docs     []string
}

var verboseFlag = flag.Bool("v", false, "Verbose mode")
var home string
var config string
var err error
var current string

func findAndReadConfInit() error {
	// will search first for conf and then for ini file in this order of directories: current, home, config.

}

func main() {
	flag.Parse()

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
	current, err = os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %s\n", err)
		return
	}

	err = findAndReadConfInit()
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return
	}

}
