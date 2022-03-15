// (C) 1990-2021.  Robert W.  Solomon.  All rights reserved.
// makewordfile.go
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

/*
REVISION HISTORY
----------------
19 July 17 -- Started writing this sort testing routine.
14 Mar 22 -- Converting to Go 1.16 by removing ioutils, and updating w/ the stuff I've learned over the 5 yrs since I wrote this.
*/

const LastAlteredDate = "15 Mar 22"

func readWord(rdr *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		r, sz, err := rdr.ReadRune()
		if err != nil {
			return sb.String(), err
		}
		if unicode.IsSpace(r) {
			if sb.Len() > 0 {
				return sb.String(), nil
			} else {
				continue
			}
		}
		if sz == 1 && (unicode.IsDigit(r) || unicode.IsLetter(r)) { // skip -, =, /, ', " and anything else not an ASCII letter or number.  Dates won't be returned as such.
			if err := sb.WriteByte(byte(r)); err != nil {
				return sb.String(), err
			}
		} else {
			continue
		}
	}
}

func main() {
	var filesize int64
	fmt.Println(" Make a file of words, one per line, for testing of my sort routines, written in Go.  Last altered ", LastAlteredDate, ", compiled by", runtime.Version())
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: makewordfil <filename>")
		os.Exit(0)
	}

	Ext1Default := ".txt"
	Ext2Default := ".out"
	OutDefault := ".dat"

	commandline := os.Args[1]
	BaseFilename := filepath.Clean(commandline)
	Filename := ""
	FileExists := false

	if strings.Contains(BaseFilename, ".") {
		Filename = BaseFilename
		FI, err := os.Stat(Filename)
		if err == nil {
			FileExists = true
			filesize = FI.Size()
		}
	} else {
		Filename = BaseFilename + Ext1Default
		FI, err := os.Stat(Filename)
		if err == nil {
			FileExists = true
			filesize = FI.Size()
		} else {
			Filename = BaseFilename + Ext2Default
			FI, err := os.Stat(Filename)
			if err == nil {
				FileExists = true
				filesize = FI.Size()
			}
		}
	}

	if !FileExists {
		fmt.Println(" File ", BaseFilename, " or ", Filename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	fileContents, err := os.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from os.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	fileReader := bytes.NewReader(fileContents)
	fileWriteBuffer := bytes.NewBuffer(make([]byte, 0, len(fileContents)))

	totalwords := 0
	for { // Main processing loop
		word, err := readWord(fileReader)
		if err != nil {
			break
		}
		word = strings.TrimSpace(word)
		if len(word) < 4 {
			continue
		}

		_, err = fileWriteBuffer.WriteString(word)
		check(err)
		_, err = fileWriteBuffer.WriteRune('\n')
		check(err)
		totalwords++
	}

	OutFilename := BaseFilename + OutDefault
	if err := os.WriteFile(OutFilename, fileWriteBuffer.Bytes(), 0666); err != nil {
		fmt.Printf(" Error while writing %s is %v\n ", OutFilename, err)
		os.Exit(1)
	}

	fmt.Printf(" Found %d words of correct length to write to %s from %s of %d bytes read in.  Bytes to write was %d.\n",
		totalwords, OutFilename, Filename, filesize, fileWriteBuffer.Len())
} // main in rpng.go

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// ---------------------------------------------------- End makewordfile.go ------------------------------
