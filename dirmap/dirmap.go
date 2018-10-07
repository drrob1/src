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

const LastAltered = "7 Oct 2018"

/*
  REVISION HISTORY
  -------- -------
   5 Nov 2017 -- First version, based on code dirwalk.
   8 Nov 2017 -- My first use of sort.Slice, which uses a closure as the less procedure.
  14 Sep 2018 -- Added map data structure to sort out why the subtotals are wrong, but the GrandTotal is right.
				   I will remove the old way.  Then use the slices to sort and display results.
				   And either display the output or write to a file.
  16 Sep 2018 -- Added code from dsrt that shows TB, GB, etc.
   4 Oct 2018 -- Will no longer ignore errors from the walk function, to try to track down why it sometimes fails.
                   It seems like checking for errors was enough for the program to work.
				   It reports an error and then continues without any more errors.
   5 Oct 2018 -- Still improving, based on a thread from go-nuts.
   7 Oct 2018 -- I posted on golang-nuts.  And I'll use their suggestions.  From the Masters, of course.
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
	DirAlreadyWalked := make(map[string]bool, 500)

	// walkfunc closure
	filepathwalkfunc := func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf(" Error from walk.  Grand total size is %d in %d number of files, error is %v. \n ", GrandTotalSize, TotalOfFiles, err)
			return nil
		}

		if !fi.Mode().IsRegular() { // not a reg file, maybe a directory or symlink
			if fi.IsDir() {
				if DirAlreadyWalked[fpath] {
					return nil
				} else {
					DirAlreadyWalked[fpath] = true
				}
			} else {
				return nil
			}
		}

		//  Now have a regular file.
		TotalOfFiles++
		GrandTotalSize += fi.Size()
		if fi.IsDir() {
			DirMap[fpath] += fi.Size()
		} else {
			DirMap[filepath.Dir(fpath)] += fi.Size() // using a map so order of walk is not important
		}
		fmt.Println(" fpath is", fpath, "dir of fpath is", filepath.Dir(fpath), " is dir", fi.IsDir())
		return nil
	}

	filepath.Walk(startDirectory, filepathwalkfunc)

	// Prepare for output.
	s2 := ""
	var i int64 = GrandTotalSize
	switch {
	case GrandTotalSize > 1e12: // 1 trillion, or TB
		i = GrandTotalSize / 1e12       // I'm forcing an integer division.
		if GrandTotalSize%1e12 > 5e11 { // rounding up
			i++
		}
		s2 = fmt.Sprintf("%d TB", i)
	case GrandTotalSize > 1e9: // 1 billion, or GB
		i = GrandTotalSize / 1e9
		if GrandTotalSize%1e9 > 5e8 { // rounding up
			i++
		}
		s2 = fmt.Sprintf("%d GB", i)
	case GrandTotalSize > 1e6: // 1 million, or MB
		i = GrandTotalSize / 1e6
		if GrandTotalSize%1e6 > 5e5 {
			i++
		}
		s2 = fmt.Sprintf("%d MB", i)
	case GrandTotalSize > 1000: // KB
		i = GrandTotalSize / 1000
		if GrandTotalSize%1000 > 500 {
			i++
		}
		s2 = fmt.Sprintf("%d KB", i)
	default:
		s2 = fmt.Sprintf("%d", i)
	}

	GrandTotalString := strconv.FormatInt(GrandTotalSize, 10)
	GrandTotalString = AddCommas(GrandTotalString)
	fmt.Print(" start dir is ", startDirectory, "; found ", TotalOfFiles, " files in this tree. ")
	fmt.Println(" Total Size of walked tree is", GrandTotalString, "or", s2, ", and len of DirMap is", len(DirMap))

	// Output map
	for n, m := range DirMap { // n is name as a string, m is map as a directory subtotal
		d := directory{} // this is a structured constant
		d.name = n
		d.subtotal = m
		dirList = append(dirList, d)
	}
	fmt.Println(" Length of sorted dirList is", len(dirList), ", length of DirAlreadyWalked is", len(DirAlreadyWalked))
	sort.Sort(dirList)

	datestr := MakeDateStr()
	outfilename := filepath.Base(startDirectory) + datestr + ".txt"
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
		s0 := fmt.Sprintf("start dir is %s, found %d files in this tree.  GrandTotal is %s, or %s, and number of directories is %d\n", startDirectory, TotalOfFiles, GrandTotalString, s2, len(DirMap))
		outputfile.WriteString(s0)
		outputfile.WriteString("\n")
		for _, d := range dirList {
			str := strconv.FormatInt(d.subtotal, 10)
			str = AddCommas(str)
			s1 := fmt.Sprintf("%s size is %s\n", d.name, str)
			outputfile.WriteString(s1)
		}
		outputfile.WriteString("\n")
		outputfile.WriteString("\n")
		outputfile.Flush()
		outfile.Close()
		fmt.Println(" List of subdirectories written to", outfilename)
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
