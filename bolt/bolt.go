package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"src/filepicker"
	"src/misc"
	"src/whichexec"
	"strconv"
	"strings"
	"time"
)

/*
30 May 25 -- Starting to think about how I would import and use the csv package to process the lightning-bolt csv files.
31 May 25 -- It's working to read the csv and write the processed csv.  Now I want to add writing in Excel format.  Nevermind.
			I think I'll create 2 slices, one byAssignment and the other byDate.  So the data can be more easily retrievable by either.
			Basically, this is by column and by row.  I don't think I need by row after all, only by column, i.e., by date.
			Maybe I just need to populate the table so it can be viewed.  I may not need to do anything, just teach them to download this file and read it into Excel.
			I decided to use the xlsx package to write an Excel file.
			It works as intended.
------------------------------------------------------------------------------------------------------------------------------------------------------
31 May 25 -- Now called bolt.go, and I'll add the file selection stuff from lint.go
*/

const LastAltered = "31 May 25"
const csvext = ".csv"
const conf = "bolt.conf"
const ini = "bolt.ini"

type fileDataType struct {
	ffname    string // full filename
	timestamp time.Time
}

var verboseFlag bool
var veryVerboseFlag bool
var numLines = 15 // I don't expect to need more than these, as I display only the first 26 elements (a-z) so far.
var monthsThreshold int

// Used for the walk function to find the csv files in the directory and it's subdirectories.
var startDirectory string

func findAndReadConfIni() error {
	// will search first for conf and then for ini file in this order of directories: current, home, config.
	fullFile, found := whichexec.FindConfig(conf)
	if !found {
		fullFile, found = whichexec.FindConfig(ini)
		if !found {
			return fmt.Errorf("%s or %s not found", conf, ini)
		}
	}

	// now need to process the config file using code from fansha.
	fileByteSlice, err := os.ReadFile(fullFile)
	if err != nil {
		return err
	}
	bytesReader := bytes.NewReader(fileByteSlice)
	inputLine, err := misc.ReadLine(bytesReader)
	if err != nil {
		return err
	}

	if verboseFlag {
		fmt.Printf(" Loaded config file: %s, containing %s\n", fullFile, inputLine)
		fmt.Println()
	}

	// Now to process setting startdirectory

	if verboseFlag {
		fmt.Printf(" In findAndReadConfIni after ReadLine BytesReader.Len=%d, and .size=%d, inputline=%q\n",
			bytesReader.Len(), bytesReader.Size(), inputLine)
	}
	trimmedInputLine, ok := strings.CutPrefix(inputLine, "startdirectory") // CutPrefix became available as of Go 1.20
	if ok {
		startDirectory = trimmedInputLine
		startDirectory = strings.TrimSpace(startDirectory)
	}

	return nil
}

func writeXLSX(fullFilename string, table [][]string) (string, error) {
	// I decided to populate an Excel type table, and then write it out.

	//workBook, err := xlsx.OpenFile(templateName)
	//if err != nil {
	//	return err
	//}

	baseFilename := filepath.Base(fullFilename)
	workbook := xlsx.NewFile()
	comment := removeExt(baseFilename)
	if len(comment) > 31 { // this limit is set by Excel
		comment = comment[:30]
	}

	sheet, err := workbook.AddSheet(comment)
	if err != nil {
		return "", err
	}

	_, _ = sheet.Cell(0, 1) // just to allow this to compile, for now.

	for i, row := range table { // remember that xl is 1-based, but the xlsx routines handle this correctly
		for j, field := range row {
			cell, err := sheet.Cell(i, j)
			if err != nil {
				fmt.Println(" Error from fmt.Fprintln: ", err, ".  Exiting.")
				return "", err
			}
			if isDate(field) {
				timedate, err := time.Parse("1/2/2006", field)
				if err != nil {
					return "", err
				}
				cell.SetDate(timedate)
				continue
			}
			cell.SetString(field)
		}
	}

	// construct output file name.
	number := misc.RandRange(1, 1000)
	numStr := strconv.Itoa(number)
	name := removeExt(fullFilename)
	fn := name + "_" + numStr + ".xlsx"
	err = workbook.Save(fn) // I don't want to clobber anything while I'm testing.

	return fn, err
}

func main() {

	fmt.Println(" bolt.go lastModified", LastAltered)
	var InFilename, fullFilename string
	var InFileExists bool
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode")
	flag.BoolVarP(&veryVerboseFlag, "veryverbose", "w", false, "very verbose mode")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "very verbose mode")
	flag.IntVarP(&monthsThreshold, "months", "m", 1, "months threshold for schedule files")
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], LastAltered, runtime.Version())
		fmt.Printf(" Usage: %s [downloaded csv file] \n", os.Args[0])
		fmt.Printf(" Needs bolt.conf or bolt.ini, and looks in current, home and config directories.\n")
		fmt.Printf(" first line must begin with startdirectory.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if veryVerboseFlag {
		verboseFlag = true
	}
	filepicker.VerboseFlag = verboseFlag

	err := findAndReadConfIni()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from findAndReadConfIni is %v, exiting\n", err)
		return
	}
	if verboseFlag {
		fmt.Printf(" startDirectory is %s\n", startDirectory)
	}

	if flag.NArg() == 0 {
		var filenames []string
		filenames, err = walkRegexFullFilenames() // function is below.  Previously, filepicker.GetRegexFullFilenamesNotLocked("csv$") // $ matches end of line
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Error from walkRegexFullFilenames is %v, exiting\n", err)
			return
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s \n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code= 999 . , / ;) : ")
		var ans string
		var n int
		n, err = fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil || n == 0 { // these are redundant.  I'm playing now.
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == "/" || ans == ";" {
			fmt.Println(" Stop code entered.")
			return
		}

		var i int
		i, err = strconv.Atoi(ans)
		if err == nil {
			InFilename = filenames[i]
		} else { // allow entering 'a' .. 'z' for 0 to 25.
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			if i > 25 {
				fmt.Printf(" Index out of bounds.  It is %d.\n", i)
				return
			}
			InFilename = filenames[i]
		}
		fmt.Println(" Picked filename is", InFilename)
		fullFilename = InFilename
	} else {
		inBuf := flag.Arg(0)
		fullFilename = filepath.Clean(inBuf)

		if strings.Contains(fullFilename, ".") { // there is an extension here
			InFilename = fullFilename
			_, err = os.Stat(InFilename)
			if err == nil {
				InFileExists = true
			}
		} else {
			InFilename = fullFilename + csvext
			_, err = os.Stat(InFilename)
			if err == nil {
				InFileExists = true
			}
		}

		if !InFileExists {
			fmt.Println(" File ", fullFilename, fullFilename+csvext, " or ", InFilename, " do not exist.  Exiting.")
			return
		}
		fmt.Println(" input filename is ", InFilename)
	}

	dir, rawBase := filepath.Split(InFilename)
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			fmt.Println(" Error from os.Getwd: ", err, ".  Exiting.")
			return
		}
	}
	base := removeExt(rawBase)
	if verboseFlag {
		fmt.Printf(" dir is %s, base is %s, raw base is %s\n", dir, base, rawBase)
	}

	// Open the file for reading.
	f, err := os.ReadFile(InFilename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from os.ReadFile: %v.  Exiting. ", err)
		return
	}

	// read in file using csv package.
	r := csv.NewReader(strings.NewReader(string(f)))
	r.Comment = '#'
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println(" Error from r.ReadAll: ", err, ".  Exiting.")
		return
	}
	fmt.Println(" Finished reading ", len(records), " records from ", InFilename)

	// construct output file name.

	fullFilename = base
	OutFilename := fullFilename + "_processed.out"
	f2, err := os.OpenFile(OutFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error from os.Create: ", err, ".  Exiting.")
		os.Exit(1)
	}
	defer f2.Close()

	// write processed records to file.

	var sum int
	var recCount int
	for _, record := range records {
		for _, field := range record {
			n, err := fmt.Fprintf(f2, "%10s |", field)
			if err != nil {
				fmt.Println(" Error from fmt.Fprintln: ", err, ".  Exiting.")
				return
			}
			sum += n
		}
		recCount++
		fmt.Fprintln(f2)
	}

	fmt.Printf(" Finished writing %d bytes and %d records to %s. \n", sum, recCount, OutFilename)
	fmt.Printf(" \n\n\n Getting ready to populate the byAssignment and byDate slices. \n\n\n")

	// write out Excel

	outName, err := writeXLSX(fullFilename, records)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from writeXLSX is %s\n", err)
	} else {
		ctfmt.Printf(ct.Green, true, " Finished writing Excel file %s\n", outName)
	}

}

func removeExt(filename string) string {
	if !strings.HasSuffix(filename, csvext) {
		return filename
	}
	return filename[:len(filename)-len(csvext)]
}

func isDate(instr string) bool {
	regexStr := `^[0-3]?[0-9]/[0-3]?[0-9]/(?:[0-9]{2})?[0-9]{2}$`
	regex := regexp.MustCompile(regexStr)
	isdate := regex.MatchString(instr)
	return isdate
}

func walkRegexFullFilenames() ([]string, error) { // This rtn sorts using sort.Slice

	// compile the regular expression
	regex := regexp.MustCompile("csv$")

	// define the timestamp constraint of >= this monthyear.  No, >= 1 month ago.
	t0 := time.Now()
	threshold := t0.AddDate(0, -monthsThreshold, 0) // threshhold is months ago set by a commandline flag.  Default is 1 month, and is a positive number.
	timeout := t0.Add(5 * time.Minute)

	// set up channel to receive FDSlices and append them to a master file data slice
	boolChan := make(chan bool)                 // unbuffered
	fileDataChan := make(chan fileDataType, 10) // a magic number I pulled out of the air
	FDSlice := make([]fileDataType, 0, 100)     // a magic number here, for now
	receiverFunc := func() {
		for fd := range fileDataChan { // fd is of type filedatatype
			FDSlice = append(FDSlice, fd)
		}
		boolChan <- true // pause until this go routine completes its work
	}
	go receiverFunc() // keep receiving on the receiver chan and appending to the master slice

	// Put walk func here.  It has to check the directory entry it gets, then search for all filenames that meet the regex and timestamp constraints.
	walkDirFunction := func(fpath string, de os.DirEntry, er error) error {
		if verboseFlag {
			if er != nil {
				fmt.Printf(" WalkDir fpath %s, de.name invalid, err %v \n", fpath, er.Error())
			} else {
				fmt.Printf(" WalkDir fpath %s\n", fpath)
			}
		}
		if er != nil {
			return filepath.SkipDir
		}
		if de.IsDir() {
			return nil // allow walk function to drill down itself
		}

		if time.Now().After(timeout) {
			return errors.New("timeout occurred")
		}

		// Not a directory, and timeout has not happened.  Only process regular files, and skip symlinks.
		if !de.Type().IsRegular() {
			return nil
		}

		lowerName := strings.ToLower(de.Name())
		if regex.MatchString(lowerName) {
			if strings.HasPrefix(de.Name(), "~") { // ignore files that Excel has locked.
				return nil
			}
			fi, err := de.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from dirEntry(%s).Info() is %v \n", de.Name(), err)
				return nil
			}

			filedata := fileDataType{
				ffname:    fpath, // filepath.Join(fpath, de.Name()) isn't needed because it turns out that fpath is the full path and filename
				timestamp: fi.ModTime(),
			}
			fileDataChan <- filedata
		}

		// I now have a walk function that has collected all matching names in this fpath into a slice of fileInfos.
		// Now what do I do now?  I want, after the call to the walk function, to have a slice of full filenames that match the constraints of name and timestamp.
		// Maybe not, I want, after the call to the walk function, to have a slice of matching full file infos, that then have to be tested to see if thay match the timestamp constraint.
		// I may need a go routine to collect all these slices into 1 slice to be sorted.  And I don't have a full filename in this slice.  I still have to
		// construct that.
		// Maybe I need a struct that has full filename and the timestamp, ie, fileInfo.ModTime(), which is of type time.Time.  I did this, and I call it fileDataType.
		// Then I made FileDataSliceType I call FDSliceType.  Then I made a masterFDSlice and a FDSlice channel so the walk function sends a slice of filedatas to the
		// goroutine that collects these and appends them to a masterFileDataSlice.  I use 2 channels for this, one to send the local filedata in the walk function,
		// and another to signal when all of these local slices have been appended to the master filedata slice.
		// Then I sort the master filedata slice and then only take at most the top 15 to be returned to the caller.
		// While debugging this code, I came across the fact that I've misunderstood what the walk fcn does.  It doesn't just produce directory names, it produces any file entries
		// with each iteration.  So opening a directory here and fetching all the entries is not correct.  I have to change this to work on individual entries.
		// Now it works.  The walk function gets the needed filenames, checks against the regex and sends down the channel if the file matches.
		// This routine then sorts the []FileDataType and checks against the date threshold constraint when converting to []string.

		return nil
	}

	if startDirectory == "" { // if not set by the config file.
		if runtime.GOOS == "windows" { // Variable is defined globally.
			startDirectory = "o:\\Nikyla's File\\RADIOLOGY MD Schedule\\"
		} else { // this is so I can debug on leox, too.  Variable is defined globally.
			startDirectory = "/home/rob/bigbkupG/Nikyla's File/RADIOLOGY MD Schedule"
		}
	}

	if verboseFlag {
		fmt.Printf(" WalkDir startDirectory: %s\n", startDirectory)
	}

	err := filepath.WalkDir(startDirectory, walkDirFunction)
	if err != nil {
		return nil, err
	}
	close(fileDataChan)
	<-boolChan // pause until FDSlice is filled.  boolChan is an unbuffered channel

	lessFunc := func(i, j int) bool {
		//return masterFDSlice[i].timestamp.UnixNano() > masterFDSlice[j].timestamp.UnixNano() // this works
		return FDSlice[i].timestamp.After(FDSlice[j].timestamp) // I want to see if this will work.  It does, and TRUE means time[i] is after time[j].
	}
	sort.Slice(FDSlice, lessFunc)

	if verboseFlag {
		fmt.Printf(" in walkRegexFullFilenames after sort.Slice.  Len=%d\n  FDSlice %#v\n", len(FDSlice), FDSlice)
	}

	stringSlice := make([]string, 0, len(FDSlice))
	var count int
	for _, f := range FDSlice {
		if f.timestamp.After(threshold) {
			stringSlice = append(stringSlice, f.ffname) // needs to preserve case of filename for linux
			count++
			if count >= numLines {
				break
			}
		} else {
			break // Made observation that in a sorted list the first file before the threshold timestamp ends the search.
		}
	}
	if verboseFlag {
		fmt.Printf(" In walkRegexFullFilenames.  len(stringSlice) = %d\n stringSlice=%v\n", len(stringSlice), stringSlice)
	}
	return stringSlice, nil
} // end walkRegexFullFilenames
