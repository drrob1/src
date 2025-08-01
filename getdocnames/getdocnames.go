package main

/*
 1 Aug 25 -- Started working on a routine to scan the weekly schedule and create a list of doc names on it, instead of having to provide one.
*/

import (
	"bytes"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
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

const lastModified = "1 Aug 2025"
const maxDimensions = 1000

const conf = "docnames.conf"
const ini = "docnames.ini"

const (
	weekdayOncall = iota + 3 // and code is a 0-origin, while Excel is 1-origin for rows.
	neuro
	body
	erXrays
	ir
	nuclear
	us
	peds
	fluoroJH
	fluoroFH
	msk // includes Xray In/Outpatient and off-sites
	mammo
	boneDensity
	late
	moonlighters
	weekendJH
	weekendFH
	weekendIR // includes On-Call Radiologist for diagnostic
	mdOff
	totalAmt
)

type dayType [23]string // there are a few unused entries here.  This goes from 0..22.  Indices 0..2 are not used.

type fileDataType struct {
	ffname    string // full filename
	timestamp time.Time
}

var categoryNamesList = []string{"0", "1", "2", "weekday On Call", "Neuro", "Body", "ER/Xrays", "IR", "Nuclear Medicine", "US", "Peds", "Fluoro JH", "Fluoro FH",
	"MSK", "Mammo", "Bone Density", "late", "weekend moonlighters", "weekend JH", "weekend FH", "weekend IR", "MD's Off"} // 0, 1 and 2 are unused

var dayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
var docNames = make([]string, 0, 25) // a list of all the doc's last names as read from the config file.
var dayOff = make(map[string]bool)   // only used in findAndReadConfIni when verboseFlag is set

var conlyFlag = flag.BoolP("conly", "c", false, "Conly mode, ie, search only Documents on c: for current user.")
var numLines = 15 // I don't expect to need more than these, as I display only the first 26 elements (a-z) so far.
var veryVerboseFlag bool
var startDirectory string
var verboseFlag = flag.BoolP("v", "V", false, "verbose debugging output")
var monthsThreshold int

func findAndReadConfIni() error { // Only is used to get startDirectory
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

	trimmedInputLine, ok := strings.CutPrefix(inputLine, "startdirectory") // CutPrefix became available as of Go 1.20
	if ok {
		startDirectory = trimmedInputLine
		startDirectory = strings.TrimSpace(startDirectory)
	}

	return nil
}

func main() {

	flag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "very verbose debugging output")
	flag.IntVarP(&monthsThreshold, "months", "m", 1, "months threshold for schedule files")
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], lastModified, runtime.Version())
		fmt.Printf(" Usage: %s [weekly xlsx file] \n", os.Args[0])
		fmt.Printf(" Needs lint.conf or lint.ini, and looks in current, home and config directories.\n")
		fmt.Printf(" first line must begin with off, and 2nd line, if present, must begin with startdirectory.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if veryVerboseFlag {
		*verboseFlag = true
	}

	fmt.Printf(" Get Doc Names from weekly schedule, last modified %s, compiled with %s, using pflag.\n", lastModified, runtime.Version())

	err := findAndReadConfIni()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from findAndReadConfINI: %s.  Exiting.\n", err)
		return
	}

	if *verboseFlag {
		fmt.Printf(" After findAndReadConfIni, Start Directory: %s\n", startDirectory)
	}

	filepicker.VerboseFlag = *verboseFlag

	var filename, ans string

	// filepicker stuff.

	includeODrive := !*conlyFlag // a comvenience flag
	if flag.NArg() == 0 {
		var filenames []string
		if includeODrive {
			filenames, err = walkRegexFullFilenames() // function is below.  "o:\\week.*xls.?$"
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from walkRegexFullFilenames is %s.  Exiting \n", err)
				return
			}
			if *verboseFlag {
				fmt.Printf(" Filenames length from o drive: %d\n", len(filenames))
			}
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf(" Error from os.UserHomeDir: %s\n", err)
			return
		}
		docs := filepath.Join(filepath.Join(homeDir, "Documents"), "week.*xls.?$")
		if *verboseFlag {
			fmt.Printf(" homedir=%q, Joined Documents: %q\n", homeDir, docs)
		}
		//                                                  filenamesDocs, err := filepicker.GetRegexFullFilenames(docs)
		filenamesDocs, err := filepicker.GetRegexFullFilenamesNotLocked(docs)
		if err != nil {
			fmt.Printf(" Error from filepicker is %s.  Exiting \n", err)
			return
		}
		if *verboseFlag {
			fmt.Printf(" FilenamesDocs length: %d\n", len(filenamesDocs))
		}

		filenames = append(filenames, filenamesDocs...)
		if *verboseFlag {
			fmt.Printf(" Filenames length after append operation: %d\n", len(filenames))
		}

		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == ";" {
			fmt.Println(" No files entered.  Exiting.")
			return
		}
		i, e := strconv.Atoi(ans)
		if e == nil {
			filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			filename = filenames[i]
		}
		fmt.Println(" Picked spreadsheet is", filename)
	} else { // will use filename entered on commandline
		filename = flag.Arg(0)
	}

}

//	Needs a walk function to find what it is looking for.  See top comments.
func walkRegexFullFilenames() ([]string, error) { // This rtn sorts using sort.Slice

	// validate the regular expression
	regex, err := regexp.Compile("week.*xls.?$")
	if err != nil {
		return nil, err
	}

	// define the timestamp constraint of >= this monthyear.  No, >= 1 month ago.
	t0 := time.Now()
	threshold := t0.AddDate(0, -monthsThreshold, 0) // threshhold is months ago set by a commandline flag.  Default is 1 month.
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
	walkDirFunction := func(fpath string, de os.DirEntry, err error) error {
		if *verboseFlag {
			if err != nil {
				fmt.Printf(" WalkDir fpath %s, de.name invalid, err %v \n", fpath, err.Error())
			} else {
				fmt.Printf(" WalkDir fpath %s\n", fpath)
			}
		}
		if err != nil {
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
				ffname:    fpath, // filepath.Join(fpath, de.Name()) doesn't work, because it turns out that fpath is the full path and filename
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

	if *verboseFlag {
		fmt.Printf(" WalkDir startDirectory: %s\n", startDirectory)
	}

	err = filepath.WalkDir(startDirectory, walkDirFunction)
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

	if *verboseFlag {
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
	if *verboseFlag {
		fmt.Printf(" In walkRegexFullFilenames.  len(stringSlice) = %d\n stringSlice=%v\n", len(stringSlice), stringSlice)
	}
	return stringSlice, nil
} // end walkRegexFullFilenames
