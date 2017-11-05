// DirWalk written in go.  (C) 2017.  All rights reserved
// dirwalk.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	//	"getcommandline"
)

const LastAltered = " 5 Nov 2017"

/*
  REVISION HISTORY
  -------- -------
   5 Nov 2017 -- First version, based on code I got from a book on Go.


*/

func main() {
	var dirTotal uint64
	fmt.Println()
	fmt.Println(" dirwalk sums the directories it walks.  Written in Go.  Last altered ", LastAltered)

	startDirectory := os.Args[1]
	start, err := os.Stat(startDirectory)
	if err != nil || !start.IsDir() {
		fmt.Println(" usage: diskwalk <directoryname>")
		os.Exit(1)
	}

	var filesList []string
	filepath.Walk(startDirectory, func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		filesList = append(filesList, fpath)
		dirTotal += uint64(fi.Size())
		return nil
	})

	DirTotalString := strconv.FormatUint(dirTotal, 10)
	DirTotalString = AddCommas(DirTotalString)
	fmt.Print(" start dir is ", startDirectory, ".  Found ", len(filesList), " files in this tree. ")
	fmt.Println(" Total Size of walked tree is", DirTotalString)
} // main

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
//---------------------------------------------------------------------------------------------------
