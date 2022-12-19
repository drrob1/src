package main

import (
	"flag"
	"fmt"
	"io"

	//ct "github.com/daviddengcn/go-colortext"
	//ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/list"
	"strings"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 2022 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                   This is going to take a while.

*/

const LastAltered = "19 Dec 2022" //

const defaultHeight = 40
const minWidth = 90
const minHeight = 22
const sepString = string(filepath.Separator)

type dirAliasMapType map[string]string

var autoWidth, autoHeight int
var err error
var fileInfos []os.FileInfo
var maxDimFlag bool

func main() {
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), LastAltered)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		//autoDefaults = false
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")
		//fmt.Fprintf(flag.CommandLine.Output(), " dsrt environ values are: numlines=%d, reverseflag=%t, sizeflag=%t, dirlistflag=%t, filenamelistflag=%t, totalflag=%t \n",
		//	dsrtParam.numlines, dsrtParam.reverseflag, dsrtParam.sizeflag, dsrtParam.dirlistflag, dsrtParam.filenamelistflag, dsrtParam.totalflag)

		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	var revFlag bool
	flag.BoolVar(&revFlag, "r", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	//var nscreens = flag.Int("n", 1, "number of screens to display, ie, a multiplier") // Ptr
	//var NLines int
	//flag.IntVar(&NLines, "N", 0, "number of lines to display") // Value

	var sizeFlag bool
	flag.BoolVar(&sizeFlag, "s", false, "sort by size instead of by date")

	var verboseFlag, veryVerboseFlag bool

	flag.BoolVar(&verboseFlag, "v", false, "verbose mode, which is same as test mode.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

	//var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	//var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	var excludeFlag bool
	var excludeRegex *regexp.Regexp
	var excludeRegexPattern string
	flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")
	flag.StringVar(&excludeRegexPattern, "x", "", "regex to be excluded from output.") // var, not a ptr.

	var filterFlag, noFilterFlag bool
	var filterStr string
	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVar(&filterFlag, "f", false, "filter value to suppress listing individual size below 1 MB.")
	flag.BoolVar(&noFilterFlag, "F", false, "Flag to undo an environment var with f set.")

	mFlag := flag.Bool("m", false, "Set maximum height, usually 50 lines")
	maxFlag := flag.Bool("max", false, "Set max height, usually 50 lines, alternative flag")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	maxDimFlag = *mFlag || *maxFlag // either m or max options will set this flag and suppress use of halfFlag.

	Reverse := revFlag
	Forward := !Reverse // convenience variable

	SizeSort := sizeFlag
	DateSort := !SizeSort // convenience variable

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execName)
		fmt.Println()
		//fmt.Println("winFlag:", winFlag)
		//fmt.Println()
		//fmt.Printf(" After flag.Parse(); option switches w=%d, nscreens=%d, Nlines=%d, numOfCols=%d\n", w, *nscreens, NLines, numOfCols)
	}

	if len(excludeRegexPattern) > 0 {
		if verboseFlag {
			fmt.Printf(" excludeRegexPattern found and is %d runes. \n", len(excludeRegexPattern))
		}
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
			excludeFlag = false
		}
		excludeFlag = true
	} else { // there is not excludeRegexPattern
		excludeRegex, _ = regexp.Compile("") // this will be detected by includeThis as an empty expression and will be ignored.  But if I don't do this, referencing it will panic.
	}

	// set which sort function will be in the sortfcn var
	sortFcn := func(i, j int) bool { return false } // became available as of Go 1.8
	if SizeSort && Forward {                        // set the value of sortfcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfos[i].Size() > fileInfos[j].Size() // I want a largest first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfos[i].ModTime().After(fileInfos[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
		}
		if verboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if SizeSort && Reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfos[i].Size() < fileInfos[j].Size() // I want an smallest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && Reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime()) // I want an oldest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	fileList := list.MakeList(sortFcn, excludeRegex)
	if verboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %s\n", i, f)
		}
		fmt.Println()
	}

	// now have the filelist.  Need to check the destination directory.

	destDir := flag.Arg(1) // this means the 2nd param on the command line.  The first must exist, even if it's just a "."
	fi, err := os.Lstat(destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %s is supposed to be the destionation directory, but os.Lstat(%s) = %#v.  Exiting\n", destDir, destDir, err)
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, " %s is supposed to be a directory, but os.Lstat(%s) not c/w a directory.  Exiting\n", destDir, destDir)
		os.Exit(1)
	}

	fileList = fileSelection(fileList)
	if verboseFlag {
		for i, f := range fileList {
			fmt.Printf(" second fileList[%d] = %s\n", i, f)
		}
		fmt.Println()
	}

	// time to copy the files

	for _, f := range fileList {
		err = CopyList(destDir, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR while copying %s -> %s is %#v.  Skipping to next file.\n", f, destDir, err)
			continue
		}
	}
}

// ------------------------------------ Copy ----------------------------------------------

func CopyList(src, destDir string) error {
	// I'm surprised that there is no os.Copy.  I have to open the file and write it to copy it.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	in, err := os.Open(src)
	defer in.Close()
	if err != nil {
		return err
	}

	destFI, err := os.Stat(destDir)
	if err != nil {
		return err
	}
	if !destFI.IsDir() {
		return fmt.Errorf("%s must be a directory.  Stat is not c/w it being a directory", destDir)
	}

	if !strings.HasSuffix(destDir, sepString) {
		destDir = destDir + sepString
	}

	outName := destDir + src
	out, err := os.Open(outName)
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

func fileSelection(inList []string) []string {
	outList := make([]string, 0, len(inList))
	numOfLines := min(autoHeight, minHeight)
	numOfLines = min(numOfLines, len(inList))
	var beg int
	end := len(inList)
	lenList := end
	var ans string

	for {
		if lenList < end {
			break
		}
		fList := inList[beg:end]

		for i, f := range fList {
			fmt.Printf(" %c: %s\n", i+'a', f)
		}
		fmt.Print(" Enter selections: ")
		_, err := fmt.Scanln(&ans)
		if err != nil { // usually means that there was no entry at the Scanln prompt.
			break
		}
		beg = end
		end = beg + numOfLines
		for _, c := range ans { // parse the answer character by character.  Well, really rune by rune but I'm ignoring that.
			f := fList[c-'a']
			outList = append(outList, f)
		}
	}

	return outList

}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
