package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"io"
	"time"

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
  22 Dec 2022 -- I'm going to add a display like dsrt, using color to show sizes.  And I'll display the timestamp.  This means that I changed NewList to return []FileInfoExType.
                   So I'm propagating that change thru.
  25 Dec 2022 -- Moving the file selection stuff to list.go
  26 Dec 2022 -- Shortened the messages.  And added a timer.
*/

const LastAltered = "26 Dec 2022" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)

// const minHeight = 26  not used here, but used in FileSelection.

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

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	Reverse := revFlag

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", ExecFI.Name(), ExecTimeStamp, execName)
		fmt.Println()
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
	} //else { // there is not excludeRegexPattern
	//excludeRegex, _ = regexp.Compile("") // this will be detected by includeThis as an empty expression and will be ignored.  But if I don't do this, referencing it will panic.
	//  but now I test against nil, and it works
	//fmt.Printf(" excludeRegex.String = %q\n", excludeRegex.String())
	//}

	fileList := list.NewList(excludeRegex, sizeFlag, Reverse) // fileList used to be []string, but now it's []FileInfoExType.
	if verboseFlag {
		fmt.Printf(" len(fileList) = %d\n", len(fileList))
	}
	if veryVerboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %#v\n", i, f)
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
		if !strings.HasSuffix(destDir, sepString) {
			destDir = destDir + sepString
		}
	} else {
		if strings.ContainsRune(destDir, ':') {
			directoryAliasesMap := list.GetDirectoryAliases()
			destDir = list.ProcessDirectoryAliases(directoryAliasesMap, destDir)
		} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			destDir = strings.Replace(destDir, "~", homeDirStr, 1)
		}
		if !strings.HasSuffix(destDir, sepString) {
			destDir = destDir + sepString
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

	fileList = list.FileSelection(fileList)
	if verboseFlag {
		for i, f := range fileList {
			fmt.Printf(" second fileList[%d] = %s\n", i, f.RelPath)
		}
		fmt.Println()
		fmt.Printf(" There are %d files in the file list.\n", len(fileList))
	}

	// time to copy the files
	start := time.Now()

	onWin := runtime.GOOS == "windows"
	for _, f := range fileList {
		err = CopyAFile(f.RelPath, destDir)
		if err == nil {
			ctfmt.Printf(ct.Green, onWin, " Copied %s -> %s\n", f.RelPath, destDir)
		} else {
			ctfmt.Printf(ct.Red, onWin, " ERROR: %s\n", err)
		}
	}
	fmt.Printf("\n Elapsed time is %s\n\n", time.Since(start))
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
		return fmt.Errorf("os.Stat(%s) must be a directory, but it's not c/w a directory", destDir)
	}

	baseFile := filepath.Base(srcFile)
	outName := filepath.Join(destDir, baseFile)
	//fmt.Printf(" CopyFile after Join: src = %#v, destDir = %#v, outName = %#v\n", srcFile, destDir, outName)
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.
		inFI, _ := in.Stat()
		if outFI.ModTime().After(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			return fmt.Errorf(" %s is same or older than destination %s.  Skipping to next file", baseFile, destDir)
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
