// eols.go
// Copyright (C) 1987-2017  Robert Solomon MD.  All rights reserved.

package main

/*
  REVISION HISTORY
  ----------------
 16 Apr 17 -- Started coding first version of eols, based on cal.go
 18 Apr 17 -- Tweaked output message text.
  9 May 17 -- Will AddCommas on filesize for output
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	//
	"getcommandline"
	//  "bufio"
	//  "strconv"
	//  "timlibg"
	//  "tokenize"
)

const lastCompiled = "9 May 2017"
const k = 1024

// CR is the ASCII carriage return value
const CR = 13

// LF is the ASCII line feed value
const LF = 10

/*
   --------------------- MAIN ---------------------------------------------
*/
func main() {
	var filesize int64

	fmt.Println("End Of Line Counting Program.  ", lastCompiled)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: eols <filename>")
		os.Exit(0)
	}

	Ext1Default := ".txt"
	Ext2Default := ".out"

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

	byteslice := make([]byte, 0, k*k) // initial capacity of 1 MB
	byteslice, err := ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from ioutil.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	CRtotal := 0
	LFtotal := 0

	for _, b := range byteslice {
		if b == CR {
			CRtotal++
		} else if b == LF {
			LFtotal++
		}

	}

	FileSizeStr := strconv.FormatInt(filesize, 10)
	if filesize > 100000 {
		FileSizeStr = AddCommas(FileSizeStr)
	}

	fmt.Print(" File ", Filename, " has ", CRtotal, " CR and ", LFtotal, " LF.")
	fmt.Println("  FileSize is ", FileSizeStr)
	//	fmt.Println("  Length of byteslice is ", len(byteslice), ", FileSize is ",
	//	filesize)  Length of byteslice and filesize are equal.
	fmt.Println()

} // end main func for eols

// end eols.go

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//-----------------------------------------------------------------------------------------------------------------------------
