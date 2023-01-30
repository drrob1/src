package main

import (
	"bufio"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"hash/crc32"
	"io"
	"src/list2"
	"time"

	//ct "github.com/daviddengcn/go-colortext"
	//ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
                   So I'm propagating that change through.
  25 Dec 2022 -- Moving the file selection stuff to list.go
  26 Dec 2022 -- Shortened the messages.  And added a timer.
  29 Dec 2022 -- Added check for an empty filelist.  And list package code was enhanced to include a sentinel of '.'
   1 Jan 2023 -- Now uses list.New instead of list.NewList
   5 Jan 2023 -- Adding stats to the output.
   6 Jan 2023 -- Now that it clears the screen each time through the selection loop, I'll print the version message at the end also.
                   Added a stop code of zero.
   7 Jan 2023 -- Now called copiesfiles.go, and is intended to have multiple targets.  If there is a target on the command line, then there will be only 1 target.
                   If this pgm prompts for a target, it will accept multiple targets.  It will have to validate each of them and will only send to the validated targets.
  10 Jan 2023 -- I've settled into calling this pgm copying.  But I'll do that w/ aliases on Windows and symlinks on linux.
  14 Jan 2023 -- Now really called copying (I had to remove the aliases and symlinks).  It will allow multiple input files and output directories.
                   To do this, I'll need flags like 'i' and 'o'.  I'll have to work on this some more.  I may get more mileage out of a GitHub flags package rather than
                   the std library one.  This will take a while, like maybe a week.
                   Kingpin looks interesting, as does go-flags.
  15 Jan 2023 -- I've decided that I only need 'i' flag for include regexp.  The command line will have 1 or more output destinations.  I don't need or want a flag for that.
                   But I'm going to continue looking at go-flags more closely.
                   I posted a message for help on golang-nuts as go get isn't working for this one.
                   I'll use the std flag package for now.
                   Now called list2.go, as the change to have 'i' inputDir is big enough that all routines need to be changed.
  17 Jan 2023 -- Uses i and rex flags.  And today I'm adding a check for zero results from the fileSelection routine.
  18 Jan 2023 -- Changing completion stats to be colorized.
  21 Jan 2023 -- I need to build in a hash check for the source and destination files.  If the hashes don't match, delete the destination and copy until the hashes match.
                   I'll use the crc32 hash.  Maybe not yet.  I'll compare the number of bytes copied w/ the size of the src file.  Let's see if that's useful enough.
  22 Jan 2023 -- I named 2 of the errors, so I can test for them.  Based on tests w/ copyc and copyc2, I'm not sure the comparison of bytes works.  So I added a call to out.Sync()
  23 Jan 2023 -- Will change time of destination file to time of source file.  Before this change, the destination has the time I ran the pgm.
  25 Jan 2023 -- Adding a verify option that uses crc32 IEEE.
  27 Jan 2023 -- Removed comparisons of number of bytes written.  The issue was OS buffering which was fixed by calling Sync(), so comparing bytes didn't work anyway.
  28 Jan 2023 -- Added a verify success message.
  30 Jan 2023 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).
                    I added timeFudgeFactor
*/

const LastAltered = "30 Jan 2023" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)
const timeFudgeFactor = 1 // seconds

// const minHeight = 26  not used here, but used in FileSelection.

var autoWidth, autoHeight int

var verboseFlag, veryVerboseFlag bool
var rex *regexp.Regexp
var rexStr, inputStr string
var ErrNotNew error
var ErrByteCountMismatch error
var verifyFlag bool

func main() {
	var err error
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), LastAltered)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Fprintf(flag.CommandLine.Output(), " Needs i flag for input.  Command line params will all be output params.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	var revFlag bool
	flag.BoolVar(&revFlag, "r", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var sizeFlag bool
	flag.BoolVar(&sizeFlag, "s", false, "sort by size instead of by date")

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

	flag.StringVar(&inputStr, "i", "", "Input source directory which can be a symlink.")
	flag.StringVar(&rexStr, "rex", "", "Regular expression inclusion pattern for input files")

	flag.BoolVar(&verifyFlag, "verify", false, "Verify copy operation")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

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
	}

	if rexStr != "" {
		rex, err = regexp.Compile(rexStr)
		if err != nil {
			fmt.Printf(" Input regular expression error is %s.  Ignoring\n", err)
		}
	}
	list2.InputDir = inputStr
	list2.FilterFlag = filterFlag
	list2.VerboseFlag = verboseFlag
	list2.VeryVerboseFlag = veryVerboseFlag
	list2.ReverseFlag = revFlag
	list2.SizeFlag = sizeFlag
	list2.ExcludeRex = excludeRegex
	list2.IncludeRex = rex

	// Finished processing the input flags and assigned list2 variables.  Now can get the fileList.

	fileList, err := list2.New() // fileList used to be []string, but now it's []FileInfoExType.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list2.New is %s\n", err)
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

	// now have the fileList.  Need to check the destination directory or directories.

	destDirs := flag.Args()
	var targetDirs []string
	if len(destDirs) == 0 { // now to process directories, if needed.
		fmt.Print(" Destination directories delimited by spaces? ")
		scanner := bufio.NewReader(os.Stdin) // need this to read the entire line and then parse it myself.
		ans, err := scanner.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR: %s.  Assuming '.'\n", err)
			ans = "." + sepString
		}
		destDir := strings.TrimSpace(ans)
		if len(destDir) == 0 { // if destDir empty, default is '.'
			destDir = "." + sepString
		}
		targetsRaw := strings.Split(destDir, " ")
		//                                            fmt.Printf("destDir: %#v, targetsRaw: %#v\n", destDir, targetsRaw)
		for _, target := range targetsRaw {
			td, err := validateTarget(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from validateTarget(%s) is %s\n", target, err)
				continue
			}
			targetDirs = append(targetDirs, td)
			if veryVerboseFlag {
				fmt.Printf(" in target and targetsRaw for loop.  target=%s,  td=%s, targetDirs=%#v\n", target, td, targetDirs)
			}
		}
	} else { // a single directory is a param on the command line
		for _, target := range destDirs {
			td, err := validateTarget(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from validateTarget(%s) is %s\n", target, err)
				targetDirs = append(targetDirs, "")
				continue
			}
			targetDirs = append(targetDirs, td)
			if veryVerboseFlag {
				fmt.Printf(" in target and targetsRaw for loop.  target=%s,  td=%s, targetDirs=%#v\n", target, td, targetDirs)
			}
		}
	}

	// By here, targetDirs is a slice which may be of length one that contains the target directories for copy operations.  I will copy the full list to each target.
	//                                                                      fmt.Printf(" targetDirs: %#v\n", targetDirs)

	fileList, err = list2.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from list2.FileSelection is %s\n", err)
		os.Exit(1)
	}
	if verboseFlag {
		for i, f := range fileList {
			fmt.Printf(" second fileList[%d] = %s\n", i, f.RelPath)
		}
		fmt.Println()
		fmt.Printf(" There are %d files in the file list to be copied to %d targets.\n", len(fileList), len(targetDirs))

		fmt.Println("Target Directories for the copy:")
		for i, d := range targetDirs {
			fmt.Printf("[%d] %q\n", i, d)
		}
		fmt.Println()
	}
	if len(fileList) == 0 {
		fmt.Printf(" Length of the fileList after calling FileSelection is zero.  Aborting.\n")
		os.Exit(1)
	}
	fmt.Printf("\n\n")

	// time to copy the files
	start := time.Now()

	var success, fail int
	onWin := runtime.GOOS == "windows"
	for _, td := range targetDirs {
		for _, f := range fileList {
			err = CopyAFile(f.RelPath, td)
			if err == nil {
				if verifyFlag {
					ctfmt.Printf(ct.Green, onWin, " %s copied to %s, and then VERIFIED.\n", f.RelPath, td) // if got here, verification succeeded.
				} else {
					ctfmt.Printf(ct.Green, onWin, " Copied %s -> %s\n", f.RelPath, td)
				}
				success++
			} else {
				ctfmt.Printf(ct.Red, onWin, " ERROR: %s\n", err)
				fail++
			}
		}
	}

	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), LastAltered)
	ctfmt.Printf(ct.Green, onWin, "\n Successfully copied %d files,", success)
	ctfmt.Printf(ct.Red, onWin, " and failed to copy %d files; ", fail)
	ctfmt.Printf(ct.Yellow, onWin, "elapsed time is %s\n\n", time.Since(start))
} // end main

// ------------------------------------ Copy ----------------------------------------------

func CopyAFile(srcFile, destDir string) error {
	// I'm surprised that there is no os.Copy.  I have to open the file and write it to copy it.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	//fmt.Printf(" CopyFile: src = %#v, destDir = %#v\n", srcFile, destDir)

	onWin := runtime.GOOS == "windows"
	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		//fmt.Printf(" CopyFile after os.Open(%s): src = %#v, destDir = %#v\n", srcFile, srcFile, destDir)
		return err
	}
	srcFI, err := in.Stat()
	//srcSize := srcFI.Size()

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
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.
		if !outFI.ModTime().Before(srcFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			ErrNotNew = fmt.Errorf(" %s is same or older than destination %s.  Skipping to next file", baseFile, destDir)
			return ErrNotNew
		}
	}
	out, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Sync()
	if err != nil {
		return err
	}

	//if srcSize != n {
	//	ErrByteCountMismatch = fmt.Errorf("Sizes are different.  Src size=%d, dest size=%d", srcSize, n)
	//	return ErrByteCountMismatch
	//}

	err = out.Close() // Needed to close destination file before I could change its timestamp.
	if err != nil {
		return err
	}

	t := srcFI.ModTime()
	if !onWin {
		t = t.Add(timeFudgeFactor * time.Second)
	}
	err = os.Chtimes(outName, t, t)
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

		if verifyFiles(in, out) {
			//ctfmt.Printf(ct.Green, onWin, " %s and its copy are verified.      ", srcFile) // No newline here on purpose.
		} else {
			return fmt.Errorf("%s and %s failed the verification process by crc32 IEEE", srcFile, outName)
		}
		if verboseFlag {
			ctfmt.Printf(ct.Green, onWin, "%s and %s pass the crc32 IEEE verification\n", srcFile, outName)
		}
	}

	return nil
} // end CopyAFile

// --------------------------------------------- validateTarget -----------------------------------------------------

func validateTarget(dir string) (string, error) {
	outDir := dir

	if veryVerboseFlag {
		fmt.Printf(" in validateTarget.  dir is %s   ", dir)
	}

	if strings.ContainsRune(dir, ':') {
		outDir = list2.ProcessDirectoryAliases(dir)
	} else if strings.Contains(dir, "~") { // this can only contain a ~ on Windows.
		homeDirStr, _ := os.UserHomeDir()
		outDir = strings.Replace(dir, "~", homeDirStr, 1)
	}

	if !strings.HasSuffix(outDir, sepString) {
		outDir = outDir + sepString
	}

	if list2.VeryVerboseFlag {
		fmt.Printf(" before call to os.Lstat(%s).  outDir is %s\n", dir, outDir)
	}

	fHandle, err := os.Open(outDir)
	defer fHandle.Close()
	if err != nil {
		return "", err
	}
	fi, er := fHandle.Stat()
	if er != nil {
		return "", er
	}
	if !fi.IsDir() {
		e := fmt.Errorf("os.Lstat(%s) is not a directory", outDir)
		return "", e
	}

	if list2.VeryVerboseFlag {
		fmt.Printf(" and exiting validateTarget.  outDir is %s\n", outDir)
	}

	return outDir, nil
} // validateTarget

// ----------------------------------------- verifyFiles ------------------------------------------------------

func verifyFiles(r1, r2 io.Reader) bool {
	return crc32IEEE(r1) == crc32IEEE(r2)
}

func crc32IEEE(r io.Reader) uint32 { // using IEEE Polynomial
	crc32Hash := crc32.NewIEEE()
	io.Copy(crc32Hash, r)
	crc32Val := crc32Hash.Sum32()
	//fmt.Printf(" crc32 value returned to caller is %d\n", crc32Val)  it works.
	return crc32Val
}
