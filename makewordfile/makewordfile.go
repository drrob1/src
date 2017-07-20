// (C) 1990-2017.  Robert W.  Solomon.  All rights reserved.
// makewordfile.go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	//
	"getcommandline"
)

const LastAlteredDate = "19 July 17"
const k = 1024

/*
	     REVISION HISTORY
	     ----------------
		 19 July 17 -- Started writing this sort testing routine.
*/

func main() {
	var filesize int64
	fmt.Println(" Make a file of words, one per line, for testing of my sort routines, written in Go.  Last altered ", LastAlteredDate)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: makewordfil <filename>")
		os.Exit(0)
	}

	Ext1Default := ".txt"
	Ext2Default := ".out"
	OutDefault := ".dat"

	commandline := getcommandline.GetCommandLineString()
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

	byteslice := make([]byte, 0, filesize+50) // add 50 just in case
	byteslice, err := ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from ioutil.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(byteslice)

	OutFilename := BaseFilename + OutDefault
	OutputFile, err := os.Create(OutFilename)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()

	totalwords := 0
	for { // Main processing loop
		word, err := bytesbuffer.ReadString(' ')
		if err != nil {
			break
		}
		word = strings.TrimSpace(word)
		if len(word) < 4 {
			continue
		}

		// now have to write out this word to the file.  I didn't open the file yet.
		_, err = OutBufioWriter.WriteString(word)
		check(err)
		_, err = OutBufioWriter.WriteRune('\n')
		check(err)
		totalwords++
	}

	OutBufioWriter.Flush()
	defer OutputFile.Close()
	fmt.Println(" Wrote", totalwords, "words")
} // main in rpng.go

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// ---------------------------------------------------- End makewordfile.go ------------------------------
