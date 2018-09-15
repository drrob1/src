// Dirmap written in go.  (C) 2017.  All rights reserved
// dirmap.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"timlibg"
)

const LastAltered = " 15 Sep 2018"

/*
  REVISION HISTORY
  -------- -------
   5 Nov 2017 -- First version, based on code dirwalk.
   8 Nov 2017 -- My first use of sort.Slice, which uses a closure as the less procedure.
  14 Sep 2018 -- Added map data structure to sort out why the subtotals are wrong, but the GrandTotal is right.
                   I think subdirectories are being entered more than once.  I need to sort the list by name and subtotal to find this.
				   I will remove the old way.  Then use the slices to sort and display results.
				   And either display the output or write to a file.
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
	var GrandTotalSize, TotalOfFiles int64 // this used to be a uint64.  I think making it an int64 is better as of 09/14/2018 2:46:12 PM
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

	dirList = make(dirslice, 0, 500)
	DirMap := make(map[string]int64, 500)
	filepathwalkfunc := func(fpath string, fi os.FileInfo, err error) error { // this is a closure
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() { // not a reg file, maybe a directory or symlink
			return nil
		}
		//  Now have a regular file.
		TotalOfFiles++
		GrandTotalSize += fi.Size()
		DirMap[filepath.Dir(fpath)] += fi.Size() // using a map so order of walk is not important

		return nil
	}

	filepath.Walk(startDirectory, filepathwalkfunc)

	// Prepare for output.

	GrandTotalString := strconv.FormatInt(GrandTotalSize, 10)
	GrandTotalString = AddCommas(GrandTotalString)
	fmt.Print(" start dir is ", startDirectory, "; found ", TotalOfFiles, " files in this tree. ")
	fmt.Println(" Total Size of walked tree is", GrandTotalString, ", and len of DirMap is", len(DirMap))

	fmt.Println()
	// Output map
	for n, m := range DirMap { // n is name as a string, m is map as a directory subtotal
		d := directory{} // this is a structured constant
		d.name = n
		d.subtotal = m
		dirList = append(dirList, d)
	}
	fmt.Println(" Length if dirList is", len(dirList))
	sort.Sort(dirList)

	datestr := MakeDateStr()
	outfilename := filepath.Base(startDirectory) + "_" + datestr
	outfile, err := os.Create(outfilename)
	defer outfile.Close()
	outputfile := bufio.NewWriter(outfile)
	defer outputfile.Flush()

	if err != nil {
		fmt.Println(" Cannot open outputfile ", outfilename, " with error ", err)
		// I'm going to assume this branch does not occur in the code below.  Else I would need a
		// stop flag of some kind to write to screen.
	}

	if len(dirList) < 30 {
		for _, d := range dirList {
			str := strconv.FormatInt(d.subtotal, 10)
			str = AddCommas(str)
			s := fmt.Sprintf("%s size is %s", d.name, str)
			fmt.Println(s)
		}
		fmt.Println()
	} else { // write output to a file.  First, build filename
		for _, d := range dirList {
			str := strconv.FormatInt(d.subtotal, 10)
			str = AddCommas(str)
			s := fmt.Sprintf("%s size is %s\n", d.name, str)
			outputfile.WriteString(s)
		}
		outputfile.WriteString("\n")
		outputfile.WriteString("\n")
		outputfile.Flush()
		outfile.Close()
	}
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

// ------------------------------------------- MakeDateStr ---------------------------------------------
func MakeDateStr() (datestr string) {

	const DateSepChar = "-"

	m, d, y := timlibg.TIME2MDY()
	timenow := timlibg.GetDateTime()

	MSTR := strconv.Itoa(m)
	DSTR := strconv.Itoa(d)
	YSTR := strconv.Itoa(y)
	Hr := strconv.Itoa(timenow.Hours)
	Min := strconv.Itoa(timenow.Minutes)
	Sec := strconv.Itoa(timenow.Seconds)

	datestr = "_" + MSTR + DateSepChar + DSTR + DateSepChar + YSTR + "_" + Hr + DateSepChar + Min + DateSepChar +
		Sec + "__" + timenow.DayOfWeekStr
	return datestr
} // MakeDateStr
