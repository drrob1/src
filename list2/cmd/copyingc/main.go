package main // copyingC

import (
	"bufio"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/few"
	"src/list2"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 22 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                 This is going to take a while.
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                 And how am I going to handle collisions?
  22 Dec 22 -- I'm going to add a display like dsrt, using color to show sizes.  And I'll display the timestamp.  This means that I changed NewList to return []FileInfoExType.
                 So I'm propagating that change through.
  25 Dec 22 -- Moving the file selection stuff to list.go
  26 Dec 22 -- Shortened the messages.  And added a timer.
  29 Dec 22 -- Added check for an empty filelist.  And list package code was enhanced to include a sentinel of '.'
   1 Jan 23 -- Now uses list.New instead of list.NewList
   5 Jan 23 -- Adding stats to the output.
   6 Jan 23 -- Now that it clears the screen each time through the selection loop, I'll print the version message at the end also.
                 Added a stop code of zero.
   7 Jan 23 -- Now called copiesfiles.go, and is intended to have multiple targets.  If there is a target on the command line, then there will be only 1 target.
                 If this pgm prompts for a target, it will accept multiple targets.  It will have to validate each of them and will only send to the validated targets.
  10 Jan 23 -- I've settled into calling this pgm copying.  But I'll do that w/ aliases on Windows and symlinks on linux.
  14 Jan 23 -- Now really called copying (I had to remove the aliases and symlinks).  It will allow multiple input files and output directories.
                 To do this, I'll need flags like 'i' and 'o'.  I'll have to work on this some more.  I may get more mileage out of a GitHub flags package rather than
                 the std library one.  This will take a while, like maybe a week.
                 Kingpin looks interesting, as does go-flags.
  15 Jan 23 -- I've decided that I only need 'i' flag for include regexp.  The command line will have 1 or more output destinations.  I don't need or want a flag for that.
                 But I'm going to continue looking at go-flags more closely.
                 I posted a message for help on golang-nuts as go get isn't working for this one.
                 I'll use the std flag package for now.
                 Now called list2.go, as the change to have 'i' inputDir is big enough that all routines need to be changed.
  17 Jan 23 -- Uses i and rex flags.  And today I'm adding a check for zero results from the fileSelection routine.
  18 Jan 23 -- Changing completion stats to be colorized.
  21 Jan 23 -- I need to build in a hash check for the source and destination files.  If the hashes don't match, delete the destination and copy until the hashes match.
                 I'll use the crc32 hash.  Maybe not yet.  I'll compare the number of bytes copied w/ the size of the src file.  Let's see if that's useful enough.
  22 Jan 23 -- I named 2 of the errors, so I can test for them.  Based on tests w/ copyc and copyc2, I'm not sure the comparison of bytes works.  So I added a call to out.Sync()
  23 Jan 23 -- Will change time of destination file to time of source file.  Before this change, the destination has the time I ran the pgm.
  25 Jan 23 -- Adding a verify option that uses crc32 IEEE.
  27 Jan 23 -- Removed comparisons of number of bytes written.  The issue was OS buffering which was fixed by calling Sync(), so comparing bytes didn't work anyway.
  28 Jan 23 -- Added a verify success message.
  30 Jan 23 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).
                  I added timeFudgeFactor
  31 Jan 23 -- timeFudgeFactor is now a Duration.
  20 Feb 23 -- Minor edit in verification messages.
  22 Feb 23 -- Now called copyingC, as I intend to write a concurrent version of the copying logic, based on the copyC family of routines.
                 And timeFudgeFactor is now 10 ms, down from 100 ms.
  23 Feb 23 -- Fixed an obvious bug that's rarely encountered in validating the output destDirs.  And added verFlag as an abbreviation for verify
  27 Feb 23 -- Fixed a bug first discovered in copyc1, in the verifyChannel.  And also a bug in the verify logic.
  14 Mar 23 -- Removed some comments.  And changed number of go routines to be the lesser of NumCPU() and len(fileList)
  15 Mar 23 -- Number of go routines should be the lesser of NumCPU() and the product of len(fileList) * len(targetDirs).
                 Will only start the verify go routine if needed.
  17 Mar 23 -- Changed error from verify operation
  19 Mar 23 -- Will adjust pooling if verifyFlag is off.
  21 Mar 23 -- Completed the usage message.
  28 Mar 23 -- Added message saying how many files will be copied.
  31 Mar 23 -- StaticCheck found a few issues.
   6 Sep 23 -- Changed the final message.
   7 Jan 24 -- Edited a comment.  Main sets up the channels and both the file list and destination list.  The actual work of copying is done by CopyAFile.  I'm putting this here as I forgot it.
   7 Apr 24 -- Shortened the destination file is same or older message
   8 Apr 24 -- Now shows the last altered date for list2.go.
   9 Apr 24 -- Looks like I never added the code here to del a failed file transfer.  Added now.
               Listening to Miki Tebeka from Ardan Labs, he said that for I/O bound, you can spin up more goroutines than runtime.NumCPU() indicates.
               But for CPU bound, there's no advantage to exceeding that number.
  11 Apr 24 -- Adding a multiplier, as I already did in cf.
  28 Jul 24 -- Fixed a data race by ErrNotNew not being global.  It should never have been global.
   3 May 25 -- Changed how dest errors are displayed, to make them more obvious.
*/

const LastAltered = "May 3, 2025" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)
const timeFudgeFactor = 1 * time.Millisecond

type cfType struct { // copy file type
	srcFile string
	destDir string
}

type msgType struct {
	s        string
	e        error
	color    ct.Color
	success  bool
	verified bool
}

type verifyType struct {
	srcFile, destFile, destDir string
}

// var pooling = runtime.NumCPU() - 3 // account for main, msgChan and verifyChan routines.  Bill Kennedy says that NumCPU() is near the sweet spot.  It's a worker pool pattern.
// But Miki Tebeka in Feb 2024 said that only applies to CPU bound work.  For I/O bound work, there is an advantage to exceeding this number.
var pooling = runtime.NumCPU()
var multiplier int
var cfChan chan cfType
var msgChan chan msgType
var verifyChan chan verifyType
var wg sync.WaitGroup
var succeeded, failed int64

var autoWidth, autoHeight int

var verboseFlag, veryVerboseFlag bool
var rex *regexp.Regexp
var rexStr, inputStr string

// var ErrNotNew error This should never have been global, as this creates a data race.  Now fixed.
var verifyFlag, verFlag bool

func main() {
	execName, err := os.Executable()
	if err != nil {
		fmt.Printf(" Error from os.Executable() is: %s.  This will be ignored.\n", err)
	}
	execFI, err := os.Lstat(execName)
	if err != nil {
		fmt.Printf(" Error from os.Lstat(%s) is: %s.  This will be ignored\n", execName, err)
	}
	execTimeStamp := execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	fmt.Printf("%s is compiled w/ %s, last altered %s, list2.go last altered %s, binary timestamp is %s\n",
		os.Args[0], runtime.Version(), LastAltered, list2.LastAltered, execTimeStamp)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, compiled with %s and binary timestamp is %s.\n", os.Args[0],
			LastAltered, runtime.Version(), execTimeStamp)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information: %s [flags] dest-dir ...\n", os.Args[0])
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
	flag.BoolVar(&verFlag, "ver", false, "Verify copy operation")

	flag.IntVar(&multiplier, "m", 10, "Multiplier for worker pool.  Default is 10.")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	if verboseFlag {
		fmt.Printf("%s timestamp is %s, full exec is %s\n", execFI.Name(), execTimeStamp, execName)
		fmt.Println()
	}

	verifyFlag = verifyFlag || verFlag

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

	onWin := runtime.GOOS == "windows"

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

	// Now have the fileList.  Need to check the destination directory or directories.

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
				ctfmt.Printf(ct.Red, true, " Error from validateTarget(%s) is %s\n", target, err)
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
		ctfmt.Printf(ct.Red, true, "Error from list2.FileSelection is %s\n", err)
		return
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
		fmt.Printf("\n Length of the fileList after calling FileSelection is zero.  Aborting.\n")
		return
	}
	fmt.Printf(" There are %d files to be copied to each of %d directories.\n\n", len(fileList), len(targetDirs))

	// time to set up the channels for the concurrent parts.  I'm going to base this on copyC1 as I got that working the other day.

	num := min(pooling*multiplier, len(fileList)*len(targetDirs))
	if verifyFlag {
		verifyChan = make(chan verifyType, num)
		go func() { // a single verify go routine.
			for v := range verifyChan {
				result, err := few.Feq32withNames(v.srcFile, v.destFile)
				if err != nil {
					msg := msgType{
						s:        "",
						e:        fmt.Errorf("ERROR from verify operation is %s", err),
						color:    ct.Red,
						success:  false,
						verified: false,
					}
					msgChan <- msg
					continue
				}

				if result {
					msg := msgType{
						s:        fmt.Sprintf("%s copied to %s and is VERIFIED", v.srcFile, v.destDir),
						e:        nil,
						color:    ct.Green,
						success:  true,
						verified: true,
					}
					msgChan <- msg
				} else {
					msg := msgType{
						s:        fmt.Sprintf("%s copied to %s but FAILED VERIFICATION", v.srcFile, v.destDir),
						e:        nil,
						color:    ct.Red,
						success:  false,
						verified: false,
					}
					msgChan <- msg
				}
				//fmt.Printf(" after msg sent to msgChan, and about to return")
				// I just learned that I can't have a return inside of the channel receive loop.  That stops the message receiving loop.  I need to use "continue" instead.
				// None of the message receiving go routines here have a return statement inside them.
				// I think I've gotten caught by this before.  Hopefully, I'll remember for the next time!
			}
		}()
	}

	cfChan = make(chan cfType, num)
	for i := 0; i < num; i++ {
		go func() { // set up a pool of worker routines, all waiting for work from the same channel.
			for c := range cfChan {
				CopyAFile(c.srcFile, c.destDir)
			}
		}()
	}

	msgChan = make(chan msgType, num)
	go func() {
		for msg := range msgChan {
			if msg.success {
				ctfmt.Printf(msg.color, onWin, " %s\n", msg.s)
				atomic.AddInt64(&succeeded, 1)
			} else {
				ctfmt.Printf(msg.color, onWin, " %s\n", msg.e)
				atomic.AddInt64(&failed, 1)
			}
			wg.Done()
		}
	}()

	// time to copy the files, now using concurrent code.

	start := time.Now()

	for _, td := range targetDirs {
		wg.Add(len(fileList))
		for _, f := range fileList {
			cf := cfType{
				srcFile: f.RelPath,
				destDir: td,
			}

			cfChan <- cf // this sends work into the worker pool.
		}
	}

	goRtns := runtime.NumGoroutine()
	close(cfChan)
	wg.Wait()
	close(msgChan)
	if verifyChan != nil {
		close(verifyChan)
	}

	fmt.Printf("%s is compiled w/ %s, last altered %s, binary timestamp of %s, using %d go routines.\n",
		os.Args[0], runtime.Version(), LastAltered, execTimeStamp, goRtns)
	if succeeded > 0 {
		ctfmt.Printf(ct.Green, onWin, "\n Successfully copied %d files.", succeeded)
	}
	if failed > 0 {
		ctfmt.Printf(ct.Red, onWin, "  Failed to copy %d files.", failed)
	}
	ctfmt.Printf(ct.Yellow, onWin, "  Elapsed time is %s.\n\n", time.Since(start))
} // end main

// ------------------------------------ Copy ----------------------------------------------

func CopyAFile(srcFile, destDir string) {
	// This is the concurrent version of this routine, that I got from copyC1.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	//fmt.Printf(" CopyFile: src = %#v, destDir = %#v\n", srcFile, destDir)

	in, err := os.Open(srcFile)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       fmt.Errorf("%s", err),
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	defer in.Close()

	destFI, err := os.Stat(destDir)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	if !destFI.IsDir() {
		msg := msgType{
			s:       "",
			e:       fmt.Errorf("os.Stat(%s) must be a directory, but it's not c/w a directory", destDir),
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}

	baseFile := filepath.Base(srcFile)
	outName := filepath.Join(destDir, baseFile)
	inFI, _ := in.Stat()
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.
		if !outFI.ModTime().Before(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			ErrNotNew := fmt.Errorf(" %s is not newer %s", baseFile, destDir)
			msg := msgType{
				s:       "",
				e:       ErrNotNew,
				color:   ct.Red,
				success: false,
			}
			msgChan <- msg
			return
		}
	}
	out, err := os.Create(outName)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	defer out.Close()

	_, err = io.Copy(out, in)

	if err != nil {
		e := out.Close() // close it so I can delete it and not get the error that the file is in use by another process.
		if e != nil {
			msg := msgType{
				s:        "",
				e:        e,
				color:    ct.Yellow, // so I see it.
				success:  false,
				verified: false,
			}
			msgChan <- msg
			return
		}
		er := os.Remove(outName)
		if er == nil {
			msg := msgType{
				s: "",
				e: fmt.Errorf("ERROR from io.Copy was %s, so it was closed w/ error of %v, and %s was deleted.  There was no error returned from os.Remove(%s)",
					err, e, outName, outName),
				color:    ct.Yellow, // so I see it
				success:  false,
				verified: false,
			}
			msgChan <- msg
		} else {
			msg := msgType{
				s: "",
				e: fmt.Errorf("ERROR from io.Copy was %s, so it was closed w/ error of %v, and os.Remove(%s) was called.  The error from os.Remove was %s",
					err, e, outName, er),
				color:    ct.Yellow, // so I see it
				success:  false,
				verified: false,
			}
			msgChan <- msg
		}
		return
	}

	err = out.Sync()
	if err != nil {
		var msg msgType

		e := out.Close() // close it so I can delete it and not get the error that the file is in use by another process.
		er := os.Remove(outName)
		if er == nil {
			msg = msgType{
				s: "",
				e: fmt.Errorf("ERROR from Sync() was %s, so it was closed w/ error of %v, and %s was deleted.  There was no error from os.Remove(%s)",
					err, e, outName, outName),
				color:    ct.Yellow, // yellow to make sure I see it.
				success:  false,
				verified: false,
			}
			msgChan <- msg
		} else {
			msg = msgType{
				s: "",
				e: fmt.Errorf("ERROR from Sync() was %s, so it was closed w/ error of %v, and os.Remove(%s) was called.  The error from os.Remove was %s",
					err, e, outName, er),
				color:    ct.Yellow, // yellow to make sure I see it.
				success:  false,
				verified: false,
			}
			msgChan <- msg
		}
		return

	}

	err = out.Close()
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	t := inFI.ModTime()
	if runtime.GOOS == "linux" {
		t = t.Add(timeFudgeFactor)
	}

	err = os.Chtimes(outName, t, t)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}

	if verifyFlag {
		vmsg := verifyType{
			srcFile:  srcFile,
			destFile: outName,
			destDir:  destDir, // this is here so the messages can be shorter.
		}
		verifyChan <- vmsg
		return
	}

	msg := msgType{
		s:        fmt.Sprintf("%s copied to %s", srcFile, destDir),
		e:        nil,
		color:    ct.Green,
		success:  true,
		verified: verifyFlag, // this flag must be false by now.
	}
	msgChan <- msg
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
	if err != nil {
		return "", err
	}
	defer fHandle.Close()
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

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}
