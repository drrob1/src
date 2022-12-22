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
  20 Dec 2022 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                   And how am I going to handle collisions?

*/

const LastAltered = "21 Dec 2022" //

const defaultHeight = 40
const minWidth = 90
const minHeight = 26
const sepString = string(filepath.Separator)

type dirAliasMapType map[string]string

var autoWidth, autoHeight int
var err error

//var fileInfos []os.FileInfo
//var maxDimFlag bool

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

	//mFlag := flag.Bool("m", false, "Set maximum height, usually 50 lines")
	//maxFlag := flag.Bool("max", false, "Set max height, usually 50 lines, alternative flag")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	Reverse := revFlag

	//maxDimFlag = *mFlag || *maxFlag // either m or max options will set this flag and suppress use of halfFlag.
	//Forward := !Reverse // convenience variable
	//SizeSort := sizeFlag
	//DateSort := !SizeSort // convenience variable

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", ExecFI.Name(), ExecTimeStamp, execName)
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
		fmt.Printf(" excludeRegexPattern = %q, excludeRegex.String = %q\n", excludeRegexPattern, excludeRegex.String())
	} else { // there is not excludeRegexPattern
		excludeRegex, _ = regexp.Compile("") // this will be detected by includeThis as an empty expression and will be ignored.  But if I don't do this, referencing it will panic.
		fmt.Printf(" excludeRegex.String = %q\n", excludeRegex.String())
	}

	fileList := list.NewList(excludeRegex, sizeFlag, Reverse)
	if verboseFlag {
		fmt.Printf(" len(fileList) = %d\n", len(fileList))
	}
	if veryVerboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %s\n", i, f)
		}
		fmt.Println()
	}

	// now have the fileList.  Need to check the destination directory.

	destDir := flag.Arg(1) // this means the 2nd param on the command line, if present.
	if destDir == "" {
		fmt.Print(" Destination directory ? ")
		_, err = fmt.Scanln(&destDir)
		if err != nil {
			destDir = "." + sepString
		}
		if strings.ContainsRune(destDir, ':') {
			directoryAliasesMap := list.GetDirectoryAliases()
			destDir = list.ProcessDirectoryAliases(directoryAliasesMap, destDir)
		} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			destDir = strings.Replace(destDir, "~", homeDirStr, 1)
		}
	} else {
		if strings.ContainsRune(destDir, ':') {
			directoryAliasesMap := list.GetDirectoryAliases()
			destDir = list.ProcessDirectoryAliases(directoryAliasesMap, destDir)
		} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			destDir = strings.Replace(destDir, "~", homeDirStr, 1)
		}
	}
	fmt.Printf("\n destDir = %#v\n", destDir)
	fi, err := os.Lstat(destDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, " %s is supposed to be the destination directory, but os.Lstat(%s) = %#v.  Exiting\n", destDir, destDir, err)
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Fprintf(os.Stderr, " %s is supposed to be the distination directory, but os.Lstat(%s) not c/w a directory.  Exiting\n", destDir, destDir)
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
		err = CopyAFile(f, destDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR while copying %s -> %s is %#v.  Skipping to next file.\n", f, destDir, err)
			continue
		}
	}
} // end main

// ------------------------------------ Copy ----------------------------------------------

func CopyAFile(srcFile, destDir string) error {
	// I'm surprised that there is no os.Copy.  I have to open the file and write it to copy it.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	//fmt.Printf(" CopyFile: src = %#v, destDir = %#v\n", srcFile, destDir)

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		//fmt.Printf(" CopyFile after os.Open(%s): src = %#v, destDir = %#v\n", srcFile, srcFile, destDir)
		return err
	}

	destFI, err := os.Stat(destDir)
	if err != nil {
		//fmt.Printf(" CopyFile after os.Stat(%s): src = %#v, destDir = %#v, err = %#v\n", destDir, srcFile, destDir, err)
		return err
	}
	if !destFI.IsDir() {
		return fmt.Errorf("os.Stat(%s) must be a directory.  Stat is not c/w it being a directory", destDir)
	}

	baseFile := filepath.Base(srcFile)
	outName := filepath.Join(destDir, baseFile)
	//fmt.Printf(" CopyFile after Join: src = %#v, destDir = %#v, outName = %#v\n", srcFile, destDir, outName)
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.
		inFI, _ := in.Stat()
		if outFI.ModTime().After(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			return fmt.Errorf(" Source %s is same or older than destination %s.  Skipping\n", srcFile, outName)
		}
	}
	out, err := os.Create(outName)
	defer out.Close()
	if err != nil {
		//fmt.Printf(" CopyFile after os.Create(%s): src = %#v, destDir = %#v, outName = %#v, err = %#v\n", outName, srcFile, destDir, outName, err)
		return err
	}
	_, err = io.Copy(out, in)
	if err != nil {
		//fmt.Printf(" CopyFile after io.Copy(%s, %s): src = %#v, destDir = %#v, outName = %#v, err = %#v\n", outName, srcFile, destDir, outName, err)
		return err
	}
	return nil
} // end CopyAFile

// --------------------------------------------fileSelection -------------------------------------------------------

func fileSelection(inList []string) []string {
	outList := make([]string, 0, len(inList))
	numOfLines := min(autoHeight, minHeight)
	numOfLines = min(numOfLines, len(inList))
	var beg, end int
	lenList := len(inList)
	var ans string

outerLoop:
	for {
		if lenList-beg >= numOfLines {
			end = beg + numOfLines
		} else {
			end = lenList
		}

		fList := inList[beg:end]

		for i, f := range fList {
			fmt.Printf(" %c: %s\n", i+'a', f)
		}
		fmt.Print(" Enter selections: ")
		_, err := fmt.Scanln(&ans)
		if err != nil || len(ans) == 0 { // usually means that there was no entry at the Scanln prompt.
			continue
		}
		// here is where I can scan the ans string looking for a-z or a.z or a,z and replace that with all the letters so indicated before passing it onto the processing loop.
		// ans = strings.ToLower(ans)  Upper case letter will mean something, not sure what yet.
		for _, c := range ans { // parse the answer character by character.  Well, really rune by rune but I'm ignoring that.
			idx := int(c - 'a')
			if idx < 0 || idx > minHeight || idx > (end-beg-1) { // entered character out of range, so complete.  IE, if enter a digit, xyz or a non-alphabetic character routine will return.
				break outerLoop
			}
			f := fList[c-'a']
			outList = append(outList, f)
		}
		if end >= lenList {
			break
		}
		beg = end
	}

	return outList
} // end fileSelection

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}
