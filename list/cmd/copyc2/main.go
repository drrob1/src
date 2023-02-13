package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"io"
	"src/few"
	"sync"
	"sync/atomic"
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
  26 Dec 2022 -- Shortened the messages.
                   Now called copyc, meaning copy concurrently.  I'm going for it.  I'll need a channel for cfType and the returned msg string for either success or failure message.
  29 Dec 2022 -- I'm back in the code.  I want to add ability to end the file selection loop on same pass as selections, make sure the slice index doesn't exceed its bounds,
                   and look into how to allow command line use of file completion since I can't do that here.  Maybe code a sentinel character that is a placeholder for 1st param
                   so that the 2nd param can have the command processor do the file completion.  And exit if there are no files that match the patterns.
  30 Dec 2022 -- I'm thinking about being able to set a filter like in dsrt routines.  It occurred to me that I can use environment strings to pass around flag values.
                   I have to think about this more.  Something like ListFilter, ListVerbose, ListVeryVerbose, ListReverse.  I can either always set them to true or false, or if set
                   then they are true, and test with LookupEnv instead of Getenv, or if use Getenv, an empty string means not set.  If filter is set, it can be set w/ the characters
                   K, M, G, etc.  Or just leave it as M as I do in dsrt.  I can combine filterFlag and filterStr so that the environment var is both.  I only really used the default which
                   I set to M, or skip files < 1 MB in size.  That worked for me and I never change that.  ListVerbose could be V or VV, ListReverse could be true only if set.
                   I'll have it ignore the dsrt environment variable so I have to explicitly set it here when I want it.
                   Nevermind.  I'll just pass the variables globally.  From the list package to here.  I'll redo the code.
  31 Dec 2022 -- Now called copyc2.  I'm removing the separate go routine that displays the messages, and including that in the primary go routine.  Then I won't need the kludge about sleeping.
                   And don't need the msgChan stuff.
   2 Jan 2023 -- I'm adding stats on how many were successfully copied and how many were not, probably because they were not newer versions of that file.  So I'm adding a return type to
                   copyAFile so I can track successes and failures.
                   All further development is here in copyc2 because I think it's smoother; it doesn't need a kludge of sleep for 10 ms.
   3 Jan 2023 -- Added output of number of go routines.
   6 Jan 2023 -- Better error handling now that all list routines return an error variable.  And a stop code was added.
   7 Jan 2023 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag
  22 Jan 2023 -- I'm going to backport the bytes copied comparison to here, and name the errors.  Hmmm, naming the errors doesn't apply here.
  23 Jan 2023 -- Changing time on destination file(s) to match the source file(s).  And fixed the date comparison for replacement copies.
  25 Jan 2023 -- Added verify.
  27 Jan 2023 -- Removed comparisons of number of bytes written.  The issue was OS buffering which was fixed by calling Sync(), so comparing bytes didn't work anyway.
  28 Jan 2023 -- Adding verify success message, which was refined the next day.
  30 Jan 2023 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).
                   Added timeFudgeFactor.
  31 Jan 2023 -- Adjusting fanOut variable to account for the main and GC goroutines.  And timeFudgeFactor is now of type Duration.
   2 Feb 2023 -- Will now use the few file equal routines.
  12 Feb 2023 -- Will make sync errors a different color, because I got today an error that said sync failed because host is down.
  13 Feb 2023 -- Adding timestamp on the exec binary.
*/

const LastAltered = "13 Feb 2023" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)
const timeFudgeFactor = 100 * time.Millisecond

type cfType struct { // copy file type
	srcFile string
	destDir string
}

//type msgType struct {
//	s     string
//	e     error
//	color ct.Color
//}
//var msgChan chan msgType
//var ErrNotNew error
//var ErrByteCountMismatch error

var autoWidth, autoHeight int
var err error

var onWin = runtime.GOOS == "windows"
var fanOut = runtime.NumCPU() - 2
var cfChan chan cfType
var wg sync.WaitGroup
var totalSucceeded, totalFailed int64

var verifyFlag bool

func main() {
	execName, err := os.Executable()
	if err != nil {
		fmt.Printf(" Error from os.Executable() is: %s, which will be ignored.\n", err)
	}
	execFI, err := os.Lstat(execName)
	if err != nil {
		fmt.Printf(" Error from os.Lstat(%s) is: %s, which will be ignored.\n", execName, err)
	}
	execTimeStamp := execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	fmt.Printf("%s is compiled w/ %s, last altered %s, and exec binary timpstamp is %s\n", os.Args[0], runtime.Version(), LastAltered, execTimeStamp)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {                                                // if output is redirected, for example.
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, compiled with %s and exec binary timestamp is %s. \n", os.Args[0], LastAltered, runtime.Version(), execTimeStamp)
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

	var globFlag bool
	flag.BoolVar(&globFlag, "g", false, "glob flag to use globbing on file matching.")

	flag.BoolVar(&verifyFlag, "verify", false, "Verify copy operation")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
		list.VeryVerboseFlag, list.VerboseFlag = true, true
	}

	Reverse := revFlag

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.ReverseFlag = revFlag
	list.FilterFlag = filterFlag

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", ExecFI.Name(), ExecTimeStamp, execName)
		fmt.Println()
		list.VerboseFlag = true
	}

	if filterFlag {
		list.FilterFlag = true
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

	if globFlag {
		list.GlobFlag = true
	}

	cfChan = make(chan cfType, fanOut)
	for i := 0; i < fanOut; i++ {
		go func() {
			for c := range cfChan {
				if BOOL := copyAFile(c.srcFile, c.destDir); BOOL {
					atomic.AddInt64(&totalSucceeded, 1)
				} else {
					atomic.AddInt64(&totalFailed, 1)
				}
			}
		}()
	}

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
		fmt.Printf(" Length of the fileList is zero.  Exiting\n")
		os.Exit(1)
	}

	// now have the fileList.  Need to check the destination directory.

	destDir := flag.Arg(1) // this means the 2nd param on the command line, if present.
	if destDir == "" {
		fmt.Print(" Destination directory ? ")
		n, err := fmt.Scanln(&destDir)
		if n == 0 || err != nil {
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
		fmt.Fprintf(os.Stderr, " Error from list.FileSelection is %s\n", err)
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

	fmt.Printf("\n\n")
	start := time.Now()
	wg.Add(len(fileList))
	for _, f := range fileList {
		cf := cfType{
			srcFile: f.RelPath,
			destDir: destDir,
		}
		//                             wg.Add(1)
		cfChan <- cf
	}
	gortns := runtime.NumGoroutine()
	close(cfChan)
	wg.Wait()

	ctfmt.Printf(ct.Cyan, onWin, " \nTotal succeeded = %d, total failed = %d, elapsed time is %s using %d go routines.\n",
		totalSucceeded, totalFailed, time.Since(start), gortns)
} // end main

// ------------------------------------ Copy ----------------------------------------------

func copyAFile(srcFile, destDir string) bool {
	// I'm surprised that there is no os.Copy.  I have to open the file and write it to copy it.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	//fmt.Printf(" CopyFile: src = %#v, destDir = %#v\n", srcFile, destDir)

	defer wg.Done()

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		return false
	}
	//srcStat, _ := in.Stat()
	//srcSize := srcStat.Size()

	destFI, err := os.Stat(destDir)
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, " CopyFile after os.Stat(%s): src = %#v, destDir = %#v, err = %#v\n", destDir, srcFile, destDir, err)
		return false
	}
	if !destFI.IsDir() {
		ctfmt.Printf(ct.Red, onWin, " os.Stat(%s) must be a directory, but it's not c/w a directory\n", destDir)
		return false
	}

	baseFile := filepath.Base(srcFile)
	outName := filepath.Join(destDir, baseFile)
	inFI, _ := in.Stat()
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.
		if !outFI.ModTime().Before(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			ctfmt.Printf(ct.Red, onWin, " %s is same or older than destination %s.  Skipping\n", baseFile, destDir)
			return false
		}
	}
	out, err := os.Create(outName)
	defer out.Close()
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		return false
	}
	_, err = io.Copy(out, in)
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		return false
	}
	err = out.Sync()
	if err != nil {
		ctfmt.Printf(ct.Magenta, onWin, "%s\n", err)
		return false
	}
	//if srcSize != n {
	//	ctfmt.Printf(ct.Red, onWin, "Sizes are different.  Src size=%d, dest size=%d\n", srcSize, n)
	//	return false
	//}
	err = in.Close()
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		return false
	}
	err = out.Close()
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		return false
	}
	t := inFI.ModTime()
	if !onWin {
		t = t.Add(timeFudgeFactor)
	}
	err = os.Chtimes(outName, t, t)
	if err != nil {
		ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		return false
	}

	if verifyFlag {
		result, err := few.Feq32withNames(srcFile, outName)
		//in, err = os.Open(srcFile)
		if err != nil {
			ctfmt.Printf(ct.Red, onWin, "%s\n", err)
			return false
		}
		//out, err = os.Open(outName)
		//if err != nil {
		//	ctfmt.Printf(ct.Red, onWin, "%s\n", err)
		//	return false
		//}

		if result {
			//                                  ctfmt.Printf(ct.Green, onWin, "%s and its copy are verified.       \n", srcFile)  This created too many output lines.
		} else {
			ctfmt.Printf(ct.Red, onWin, "%s and %s fail verification\n", srcFile, outName)
			return false
		}

		if list.VerboseFlag {
			onWin := runtime.GOOS == "windows"
			ctfmt.Printf(ct.Green, onWin, "%s and %s pass the crc32 IEEE verification\n", srcFile, outName)
		}
	}

	if verifyFlag {
		ctfmt.Printf(ct.Green, onWin, "%s copied to %s, and then VERIFIED. \n", srcFile, destDir) // If get here, the verification succeeded.
	} else {
		ctfmt.Printf(ct.Green, onWin, "%s copied to %s\n", srcFile, destDir)
	}

	return true
} // end CopyAFile

/*  Replaced by few (feq but easier to type) routine.
// ----------------------------------------- verifyFiles ------------------------------------------------------

func verifyFiles(r1, r2 io.Reader) bool {
	return crc32IEEE(r1) == crc32IEEE(r2)
}

func crc32IEEE(r io.Reader) uint32 { // using IEEE Polynomial
	crc32Hash := crc32.NewIEEE()
	io.Copy(crc32Hash, r)
	crc32Val := crc32Hash.Sum32()
	//fmt.Printf(" crc32 value returned to caller is %d\n", crc32Val) // it works.
	return crc32Val
}

*/
