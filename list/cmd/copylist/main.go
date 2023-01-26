package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"hash/crc32"
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
  29 Dec 2022 -- Added check for an empty filelist.  And list package code was enhanced to include a sentinel of '.'
   1 Jan 2023 -- Now uses list.New instead of list.NewList
   5 Jan 2023 -- Adding stats to the output.
   6 Jan 2023 -- Now that it clears the screen each time thru the selection loop, I'll print the version message at the end also.
                   Added a stop code of zero.
   7 Jan 2023 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag
  22 Jan 2023 -- Added Sync call.
  23 Jan 2023 -- Added changing destination file(s) timestamp to match the respective source file(s).  And fixed date comparison for replacement copies.
  25 Jan 2023 -- Adding verify
*/

const LastAltered = "26 Jan 2023" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)

// const minHeight = 26  not used here, but used in FileSelection.

var autoWidth, autoHeight int
var err error
var verifyFlag bool

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
	//var extflag = flag.Bool("e", false, "only print if there is no extension, like a binary file")
	//var extensionflag = flag.Bool("ext", false, "only print if there is no extension, like a binary file")

	var sizeFlag bool
	flag.BoolVar(&sizeFlag, "s", false, "sort by size instead of by date")

	var verboseFlag, veryVerboseFlag bool

	flag.BoolVar(&verboseFlag, "v", false, "verbose mode, which is same as test mode.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

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

	flag.BoolVar(&verifyFlag, "verify", false, "Verify copy operation")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	Reverse := revFlag

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.FilterFlag = filterFlag
	list.ReverseFlag = revFlag

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

	fileList, err := list.New(excludeRegex, sizeFlag, Reverse) // fileList used to be []string, but now it's []FileInfoExType.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.New is %s\n", err)
		os.Exit(1)
	}
	if verboseFlag {
		fmt.Printf(" len(fileList) = %d\n", len(fileList))
	}
	if veryVerboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %#v\n", i, f)
		}
		fmt.Println()
	}
	if len(fileList) == 0 {
		fmt.Printf(" Length of the filelist is zero.  Aborting\n")
		os.Exit(1)
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

	fileList, err = list.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}
	if verboseFlag {
		for i, f := range fileList {
			fmt.Printf(" second fileList[%d] = %s\n", i, f.RelPath)
		}
		fmt.Println()
		fmt.Printf(" There are %d files in the file list.\n", len(fileList))
	}
	fmt.Printf("\n\n")

	// time to copy the files
	start := time.Now()

	var success, fail int
	onWin := runtime.GOOS == "windows"
	for _, f := range fileList {
		err = CopyAFile(f.RelPath, destDir)
		if err == nil {
			ctfmt.Printf(ct.Green, onWin, " Copied %s -> %s\n", f.RelPath, destDir)
			success++
		} else {
			ctfmt.Printf(ct.Red, onWin, " ERROR: %s\n", err)
			fail++
		}
	}
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), LastAltered)
	fmt.Printf("\n Successfully copied %d files, and FAILED to copy %d files; elapsed time is %s\n\n", success, fail, time.Since(start))
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
	inFI, _ := in.Stat()
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.
		if !outFI.ModTime().Before(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
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
	err = out.Sync()
	if err != nil {
		return err
	}

	err = in.Close()
	if err != nil {
		return err
	}
	err = out.Close()
	if err != nil {
		return err
	}
	err = os.Chtimes(outName, inFI.ModTime(), inFI.ModTime())
	if err != nil {
		return err
	}

	if verifyFlag {
		in, err = os.Open(srcFile)
		if err != nil {
			return err
		}
		out, err = os.Open(outName)
		if err != nil {
			return err
		}

		if !verifyFiles(in, out) {
			return fmt.Errorf("%s and %s failed the verification process by crc32 IEEE", srcFile, outName)
		}
		if list.VerboseFlag { // this is made global by assigning to list above.
			onWin := runtime.GOOS == "windows"
			ctfmt.Printf(ct.Green, onWin, "%s and %s pass the crc32 IEEE verification\n", srcFile, outName)
		}
	}

	return nil
} // end CopyAFile

func verifyFiles(r1, r2 io.Reader) bool {
	return crc32IEEE(r1) == crc32IEEE(r2)
}

func crc32IEEE(r io.Reader) uint32 { // using IEEE Polynomial
	crc32Hash := crc32.NewIEEE()
	io.Copy(crc32Hash, r)
	crc32Val := crc32Hash.Sum32()
	//fmt.Printf(" crc32 value returned to caller is %d\n", crc32Val)  It works.0
	return crc32Val
}
