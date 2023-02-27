package main // copyc1

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
   3 Jan 2023 -- Fixed the wait group so all msg's get printed, backported the stats to display and I removed the sleep kludge.  And then I added displaying the number of go routines.
   6 Jan 2023 -- list now has a stop code, and all routines return an error.
   7 Jan 2023 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag
  22 Jan 2023 -- I'm going to backport the bytes copied comparison to here, and name the errors.  And I added a call to out.sync.  That may have been the trouble all along.
  23 Jan 2023 -- Changing time on destination file(s) to match the source file(s).  And fixing the date comparison for replacement copies, from .After() to not .Before().
  27 Jan 2023 -- Removed comparisons of number of bytes written.  The issue was OS buffering which was fixed by calling Sync(), so comparing bytes didn't work anyway.
  30 Jan 2023 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).
                    I added timeFudgeFactor.
  31 Jan 2023 -- Adjusting fanOut variable to account for the main and GC goroutines.  And timeFudgeFactor is now a Duration.
  12 Feb 2023 -- Adding verify option (finally).  In testing later in the day, I got a sync failed because host is down error.  I'm making sync errors a different color now.
  13 Feb 2023 -- Adding timestamp on the exec binary.
  20 Feb 2023 -- Based on copyc.go, now called copyc1.go.  I want to add the verify option to be its own go routine.  But I'm splitting this off as another pgm.
                   So I need another type around which to base a channel for this new go routine to get it's work.
                   And I made the timeFudgeFacter smaller, to 10 ms.
                   I have to really, really, remember that channel receiving for loops do not have a return statement.
  23 Feb 2023 -- Added verFlag.
  26 Feb 2023 -- I'm tracking down a bug here.  I'm getting %!s(<nil>) displayed, and I don't know why.  I changed a use of Stat to Open, since that's been an issue before on linux.
                   This may be an issue because I was trying to copy a running program.  Windows may not allow me to get a lock on an open file.  Hence, the error is not accurate.
                   The error says that it can't find the file, but this may be an inaccurate error message.
*/

const LastAltered = "26 Feb 2023" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)
const timeFudgeFactor = 10 * time.Millisecond

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

var autoWidth, autoHeight int
var err error

var onWin = runtime.GOOS == "windows"
var pooling = runtime.NumCPU() - 3 // account for main, msgChan and verifyChan routines.  Bill Kennedy says that NumCPU() is near the sweet spot.  It's a worker pool pattern.
var cfChan chan cfType
var msgChan chan msgType
var verifyChan chan verifyType
var wg sync.WaitGroup
var succeeded, failed int64
var ErrNotNew error
var verifyFlag, verFlag bool

//var ErrByteCountMismatch error

func main() {
	if pooling < 1 {
		pooling = 1
	}
	execName, err := os.Executable()
	if err != nil {
		fmt.Printf(" Error from os.Executable() is: %s.  This will be ignored.\n", err)
	}
	execFI, err := os.Lstat(execName)
	if err != nil {
		fmt.Printf(" Error from os.Lstat(%s) is: %s.  This will be ignored\n", execName, err)
	}
	execTimeStamp := execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	fmt.Printf("%s is compiled w/ %s, last altered %s, exec binary timestamp is %s\n", os.Args[0], runtime.Version(), LastAltered, execTimeStamp)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		//autoDefaults = false
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

	flag.BoolVar(&verifyFlag, "verify", false, "Verify that destination is same as source.")
	flag.BoolVar(&verFlag, "ver", false, "Verify copy operation")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
		list.VeryVerboseFlag, list.VerboseFlag = true, true
	}

	verifyFlag = verifyFlag || verFlag

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

	if globFlag {
		list.GlobFlag = true
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

	cfChan = make(chan cfType, pooling)
	for i := 0; i < pooling; i++ {
		go func() {
			for c := range cfChan {
				CopyAFile(c.srcFile, c.destDir)
			}
		}()
	}

	verifyChan = make(chan verifyType, pooling)
	go func() {
		for v := range verifyChan {
			result, err := few.Feq32withNames(v.srcFile, v.destFile)
			if err != nil {
				msg := msgType{
					s:        "",
					e:        err,
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
				continue
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
			// I just learned that I can't have a return inside of the channel receive loop.  That stops the message receiving loop.
			// None of the message receiving go routines here have a return statement inside them.
			// I think I've gotten caught by this before.  Hopefully, I'll remember for the next time!
		}
	}()

	msgChan = make(chan msgType, pooling)
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
	goRtns := runtime.NumGoroutine()
	close(cfChan)
	wg.Wait()
	close(verifyChan)
	close(msgChan)
	ctfmt.Printf(ct.Cyan, onWin, " Total files copied is %d, total files NOT copied is %d, elapsed time is %s using %d go routines.\n",
		succeeded, failed, time.Since(start), goRtns)
} // end main

// ------------------------------------ Copy ----------------------------------------------

func CopyAFile(srcFile, destDir string) {
	// I'm surprised that there is no os.Copy.  I have to open the file and write it to copy it.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	//fmt.Printf(" CopyFile: src = %#v, destDir = %#v\n", srcFile, destDir)

	if list.VerboseFlag {
		fmt.Printf(" In CopyAFile.  srcFile is %s, destDir %s.\n", srcFile, destDir)
	}

	in, err := os.Open(srcFile)
	defer in.Close()
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

	destD, err := os.Open(destDir)
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

	destFI, err := destD.Stat()
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
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.  I'm ignoring err != nil because that means that file's not already there.
		if !outFI.ModTime().Before(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			ErrNotNew = fmt.Errorf(" Skipping %s as it's same or older than destination %s.", baseFile, destDir)
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
	_, err = io.Copy(out, in)
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
	err = out.Sync()
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Magenta,
			success: false,
		}
		msgChan <- msg
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
			srcFile:  baseFile,
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
	//return  this is implied.
} // end CopyAFile
