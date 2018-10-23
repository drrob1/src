/*
  multimap.go.  multithreaded directoy mapping.  Was pmap but that conflicted with a system command.
  22 Oct 2018 -- Started coding this based on MichaelTJones' multithreaded code.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
	"timlibg"

	"github.com/MichaelTJones/walk"
)

const LastAlteredDate = "Oct 22, 2018"

var duration = flag.String("d", "", "find files modified within DURATION")
var format = flag.String("f", "2006-01-02 03:04:05", "time format")
var instant = flag.String("t", "", "find files modified since TIME")
var quiet = flag.Bool("q", false, "do not print filenames")
var verbose = flag.Bool("v", false, "print summary statistics")
var help = flag.Bool("h", false, "print help message")

type directory struct {
	name     string
	subtotal int64
}

type item struct {
	name string
	size int64
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
	fmt.Println("multimap is a multithreaded directory mapping written in Go.  Last Altered", LastAlteredDate)

	var GrandTotalSize, TotalOfFiles int64
	var startDirectory string
	var dirList dirslice
	var err error
	now := time.Now()

	fmt.Println()

	if len(os.Args) < 2 {
		startDirectory, err = os.Getwd()
		if err != nil {
			log.Fatalln(" error from Getwd is", err)
		}
	} else {
		startDirectory = os.Args[1]
	}
	start, err := os.Stat(startDirectory)
	if err != nil || !start.IsDir() {
		fmt.Println(" usage: dirmap <directoryname>")
		os.Exit(1)
	}

	dirList = make(dirslice, 0, 1024)
	DirMap := make(map[string]int64, 1024)

	// goroutine to collect items
	done := make(chan bool)
	results := make(chan item, 1024)
	var lock sync.Mutex
	go func() {
		for r := range results {
			lock.Lock()
			DirMap[r.name] += r.size
			lock.Unlock()
		}
		done <- true
	}()

	// parallel walker to find dirs
	var tFiles, tBytes int64 // total files and bytes
	sizeVisitor := func(path string, info os.FileInfo, err error) error {
		var d item
		if err == nil {
			lock.Lock()
			tFiles += 1
			tBytes += info.Size()
			lock.Unlock()

			if info.IsDir() {
				d.name = path
				d.size = info.Size()
			} else {
				d.name = filepath.Dir(path)
				d.size = info.Size()
			}
			results <- d
		} else {
			fmt.Printf(" Error from walk.  Grand total size is %d in %d number of files, error is %v. \n ",
				GrandTotalSize, TotalOfFiles, err)
		}
		return nil
	}
	walk.Walk(startDirectory, sizeVisitor)

	// wait for traversal results and prepare for output.
	close(results) // no more results
	<-done         // wait for final results

	s2 := ""
	var i int64 = tBytes
	switch {
	case tBytes > 1e12: // 1 trillion, or TB
		i = tBytes / 1e12       // I'm forcing an integer division.
		if tBytes%1e12 > 5e11 { // rounding up
			i++
		}
		s2 = fmt.Sprintf("%d TB", i)
	case tBytes > 1e9: // 1 billion, or GB
		i = tBytes / 1e9
		if tBytes%1e9 > 5e8 { // rounding up
			i++
		}
		s2 = fmt.Sprintf("%d GB", i)
	case tBytes > 1e6: // 1 million, or MB
		i = tBytes / 1e6
		if tBytes%1e6 > 5e5 {
			i++
		}
		s2 = fmt.Sprintf("%d MB", i)
	case tBytes > 1000: // KB
		i = tBytes / 1000
		if tBytes%1000 > 500 {
			i++
		}
		s2 = fmt.Sprintf("%d KB", i)
	default:
		s2 = fmt.Sprintf("%d", i)
	}

	GrandTotalString := strconv.FormatInt(tBytes, 10)
	GrandTotalString = AddCommas(GrandTotalString)

	// Construct output filename
	datestr := MakeDateStr()
	outfilename := "dirmap_" + filepath.Base(startDirectory) + datestr + ".txt"
	outfile, err := os.Create(outfilename)
	defer outfile.Close()
	var bufoutfile = bufio.NewWriter(outfile)
	defer bufoutfile.Flush()

	//	outputfile := bufio.NewWriter(outfile)  these may duplicate the "expert" code below.
	//	defer outputfile.Flush()
	if err != nil {
		fmt.Println(" Cannot open outputfile ", outfilename, " with error ", err)
		// I'm going to assume this branch does not occur in the code below.  Else I would need a
		// stop flag of some kind to write to screen.
	}

	// Construct output map
	for n, m := range DirMap { // n is name as a string, m is map as a directory subtotal
		d := directory{} // this is a structured constant
		d.name = n
		d.subtotal = m
		dirList = append(dirList, d)
	}
	sort.Sort(dirList)

	𝛥t := float64(time.Since(now)) / 1e9

	s0 := fmt.Sprintf("start dir is %s, found %d files in this tree.  GrandTotal is %s, or %s, and number of directories is %d, took %.4g s to generate.\n",
		startDirectory, tFiles, GrandTotalString, s2, len(DirMap), 𝛥t)
	fmt.Println(s0)
	_, err = bufoutfile.WriteString(s0)
	_, err = bufoutfile.WriteRune('\n')
	if err != nil {
		fmt.Println(" error from writing bufoutfile is", err)
		os.Exit(1)
	}

	for _, d := range dirList {
		var str = strconv.FormatInt(d.subtotal, 10)
		str = AddCommas(str)
		s5 := fmt.Sprintf("%s size is %s\n", d.name, str)
		_, err := bufoutfile.WriteString(s5)
		if err != nil {
			fmt.Println(" error from writing to bufoutfile while writing dirList.  Error is", err)
			os.Exit(1)
		}
	}
	fmt.Println(" List of", len(dirList), " (sub)directories written to", outfilename)
}

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

//------------------------------------------------------------------- min
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
