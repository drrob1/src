// Dirmap written in go.  (C) 2017.  All rights reserved
// dirmap.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

const LastAltered = " 28 Feb 2018"

/*
  REVISION HISTORY
  -------- -------
   5 Nov 2017 -- First version, based on code dirwalk.
   8 Nov 2017 -- My first use of sort.Slice, which uses a closure as the less procedure.
  28 Feb 2810 -- My use of alternate closure syntax seems to be working.  I can understand this more easily.
*/

type directory struct {
	name     string
	subtotal int64
}

type dirslice []directory

func (ds dirslice) Less(i, j int) bool {
	return ds[i].subtotal > ds[j].subtotal // I want a reverse sort, largest first
}

func (ds dirslice) Swap(i, j int) {
	ds[i], ds[j] = ds[j], ds[i]
}

func (ds dirslice) Len() int {
	return len(ds)
}

func main() {
	var GrandTotal uint64
	var startDirectory string
	var dirList dirslice

	fmt.Println()
	fmt.Println(" dirmap sums the directories it walks.  Written in Go.  Last altered ", LastAltered)

	if len(os.Args) < 2 {
		startDirectory, _ = os.Getwd()
	} else {
		startDirectory = os.Args[1]
	}
	start, err := os.Stat(startDirectory)
	if err != nil || !start.IsDir() {
		fmt.Println(" usage: diskwalk <directoryname>")
		os.Exit(1)
	}

	var filesList []string
	filesList = make([]string, 0, 5000)
	dirList = make(dirslice, 0, 5000)
	filepath.Walk(startDirectory, func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.Mode().IsDir() && len(dirList) == 0 { // This one is the first in the list.
			d := directory{} // null element to init.
			d.name = filepath.Dir(fpath)
			d.subtotal = 0
			dirList = append(dirList, d)
			return nil
		} else if fi.Mode().IsDir() && filepath.Dir(fpath) != dirList[len(dirList)-1].name {
			d := directory{} // null element to init.
			d.name = filepath.Dir(fpath)
			d.subtotal = 0
			dirList = append(dirList, d)
			return nil
		} else if !fi.Mode().IsRegular() { // not a dir or a reg file, maybe a symlink
			return nil
		}

		filesList = append(filesList, fpath)
		GrandTotal += uint64(fi.Size())
		lastDirList := len(dirList) - 1
		if filepath.Dir(fpath) == dirList[lastDirList].name { // if not already there.
			dirList[lastDirList].subtotal += fi.Size()
		}

		return nil
	})

	// Will now sort by name and attempt to remove duplicates by setting their subtotal to zero.
	//	sort.Slice(dirList, func(i, j int) bool { return dirList[i].name < dirList[j].name })
	// Attempting to us an alternate closure syntax, that I find easier to understand.  It seems
	// to work.
	f := func(i, j int) bool {
		return dirList[i].name < dirList[j].name
	}
	sort.Slice(dirList, f)

	NumOfDirs := len(dirList)
	for i := 0; i < NumOfDirs-1; i++ { // will compare current to prev list entry
		if dirList[i].name == dirList[i+1].name {
			for j := i + 1; j < NumOfDirs && dirList[i].name == dirList[j].name; j++ { // walk down list of == names
				dirList[i].subtotal += dirList[j].subtotal
				dirList[j].subtotal = 0 // zero it out so it will sort towards bottom and not display.
			}
		}

	}

	// Prepare for output.
	sort.Sort(dirList)

	for i := NumOfDirs - 1; i > 0 && dirList[i].subtotal == 0; i-- {
		NumOfDirs--
	}

	GrandTotalString := strconv.FormatUint(GrandTotal, 10)
	GrandTotalString = AddCommas(GrandTotalString)
	fmt.Print(" start dir is ", startDirectory, "; found ", len(filesList), " files in this tree. ")
	fmt.Println(" Total Size of walked tree is", GrandTotalString, ", and number of non-empty directories is", NumOfDirs)

	fmt.Println()
	for i, d := range dirList {
		subtotalstr := strconv.FormatInt(d.subtotal, 10)
		subtotalstr = AddCommas(subtotalstr)
		fmt.Printf(" %s subtotal is %s.\n", d.name, subtotalstr)
		if i > 40 || d.subtotal == 0 {
			break
		}
	}
	fmt.Println()
	fmt.Println()
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

func min(i, j int) int {
	if i < j {
		return i
	} else {
		return j
	}
} // min
//---------------------------------------------------------------------------------------------------
