// eols.go
// Copyright (C) 1987-2017  Robert Solomon MD.  All rights reserved.

package main

/*
REVISION HISTORY
----------------
16 Apr 17 -- Started coding first version of eols, based on cal.go
18 Apr 17 -- Tweaked output message text.
 9 May 17 -- Will AddCommas on filesize for output
11 Mar 21 -- Updated code based on my add'l experience.  And added method 2 as a test of concept.
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const lastCompiled = "11 Mar 2021"

// CR is the ASCII carriage return value
const CR = 13
const CRstr = "\r"

// LF is the ASCII line feed value
const LF = 10
const LFstr = "\n"

/*
   --------------------- MAIN ---------------------------------------------
*/
func main() {
	var filesize int64

	fmt.Println()
	fmt.Println("End Of Line Counting Program.  Last altered", lastCompiled, "and compiled by", runtime.Version())
	fmt.Println()

	if len(os.Args) < 2 {
		fmt.Println(" Usage: eols <filename>")
		os.Exit(1)
	}

	Ext1Default := ".txt"
	Ext2Default := ".out"

	BaseFilename := filepath.Clean(os.Args[1])
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

	byteSlice, err := ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(err, ".  Error from ioutil.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	var CRtotal, LFtotal uint

	for _, b := range byteSlice {
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

	fmt.Println(" File", Filename, "has", CRtotal, "CR and", LFtotal, "LF by method 1,")

	CRtot := strings.Count(string(byteSlice), CRstr)
	LFtot := strings.Count(string(byteSlice), LFstr)
	fmt.Println(" and has", CRtot, "CR and", LFtot, "LF by method 2, ie, strings.Count().")

	fmt.Println(" FileSize is ", FileSizeStr)
	fmt.Println()
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
