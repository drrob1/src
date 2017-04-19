// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// nocr.go

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	//
	"getcommandline"
)

const lastCompiled = "18 Apr 17"

func main() {
	/*
	   REVISION HISTORY
	   ----------------
	   17 Apr 17 -- Started writing nocr, based on rpn.go
	   18 Apr 17 -- It worked yesterday.  Now I'll rename files as in Modula-2.
	*/

	var inoutline string
	//	var err error

	fmt.Println(" nocr removes all <CR> from a file.  Last compiled ", lastCompiled)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: nocr <filename> ")
		os.Exit(1)
	}

	commandline := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(commandline)
	InFilename := ""
	InFileExists := false
	Ext1Default := ".txt"
	OutFileSuffix := ".out"

	if strings.Contains(BaseFilename, ".") {
		InFilename = BaseFilename
		_, err := os.Stat(InFilename)
		if err == nil {
			InFileExists = true
		}
	} else {
		InFilename = BaseFilename + Ext1Default
		_, err := os.Stat(InFilename)
		if err == nil {
			InFileExists = true
		}
	}

	if !InFileExists {
		fmt.Println(" File ", BaseFilename, " or ", InFilename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	InputFile, err := os.Open(InFilename)
	if err != nil {
		fmt.Println(" Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer InputFile.Close()

	OutFilename := BaseFilename + OutFileSuffix
	OutputFile, err := os.Create(OutFilename)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()

	InBufioScanner := bufio.NewScanner(InputFile)
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()

	for InBufioScanner.Scan() {
		inoutline = InBufioScanner.Text() // does not include the trailing EOL char
		_, err := OutBufioWriter.WriteString(inoutline)
		check(err)
		_, err = OutBufioWriter.WriteRune('\n')
		check(err)
	}

	InputFile.Close()
	OutBufioWriter.Flush() // code did not work without this line.
	OutputFile.Close()

	// Make the processed file the same name as the input file.  IE, swap in and
	// out files.
	TempFilename := InFilename + OutFilename + ".tmp"
	os.Rename(InFilename, TempFilename)
	os.Rename(OutFilename, InFilename)
	os.Rename(TempFilename, OutFilename)

	FI, err := os.Stat(InFilename)
	InputFileSize := FI.Size()

	FI, err = os.Stat(OutFilename)
	OutputFileSize := FI.Size()

	fmt.Println(" Original file is now ", OutFilename, " and size is ", OutputFileSize)
	fmt.Println(" Output File is now ", InFilename, " and size is ", InputFileSize)
	fmt.Println()

} // main in nocr.go

func check(e error) {
	if e != nil {
		panic(e)
	}
}
