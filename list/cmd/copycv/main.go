package main // copycv

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
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
  18 Dec 22 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                 This is going to take a while.
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                 And how am I going to handle collisions?
  22 Dec 22 -- I'm going to add a display like dsrt, using color to show sizes.  And I'll display the timestamp.  This means that I changed NewList to return []FileInfoExType.
                 So I'm propagating that change thru.
  25 Dec 22 -- Moving the file selection stuff to list.go
  26 Dec 22 -- Shortened the messages.
                 Now called copyc, meaning copy concurrently.  I'm going for it.  I'll need a channel for cfType and the returned msg string for either success or failure message.
  29 Dec 22 -- I'm back in the code.  I want to add ability to end the file selection loop on same pass as selections, make sure the slice index doesn't exceed its bounds,
                 and look into how to allow command line use of file completion since I can't do that here.  Maybe code a sentinel character that is a placeholder for 1st param
                 so that the 2nd param can have the command processor do the file completion.  And exit if there are no files that match the patterns.
  30 Dec 22 -- I'm thinking about being able to set a filter like in dsrt routines.  It occurred to me that I can use environment strings to pass around flag values.
                 I have to think about this more.  Something like ListFilter, ListVerbose, ListVeryVerbose, ListReverse.  I can either always set them to true or false, or if set
                 then they are true, and test with LookupEnv instead of Getenv, or if use Getenv, an empty string means not set.  If filter is set, it can be set w/ the characters
                 K, M, G, etc.  Or just leave it as M as I do in dsrt.  I can combine filterFlag and filterStr so that the environment var is both.  I only really used the default which
                 I set to M, or skip files < 1 MB in size.  That worked for me and I never change that.  ListVerbose could be V or VV, ListReverse could be true only if set.
                 I'll have it ignore the dsrt environment variable so I have to explicitly set it here when I want it.
                 Nevermind.  I'll just pass the variables globally.  From the list package to here.  I'll redo the code.
   3 Jan 23 -- Fixed the wait group so all msg's get printed, backported the stats to display and I removed the sleep kludge.  And then I added displaying the number of go routines.
   6 Jan 23 -- list now has a stop code, and all routines return an error.
   7 Jan 23 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag
  22 Jan 23 -- I'm going to backport the bytes copied comparison to here, and name the errors.  And I added a call to out.sync.  That may have been the trouble all along.
  23 Jan 23 -- Changing time on destination file(s) to match the source file(s).  And fixing the date comparison for replacement copies, from .After() to not .Before().
  27 Jan 23 -- Removed comparisons of number of bytes written.  The issue was OS buffering which was fixed by calling Sync(), so comparing bytes didn't work anyway.
  30 Jan 23 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).
                  I added timeFudgeFactor.
  31 Jan 23 -- Adjusting fanOut variable to account for the main and GC goroutines.  And timeFudgeFactor is now a Duration.
  12 Feb 23 -- Adding verify option (finally).  In testing later in the day, I got a sync failed because host is down error.  I'm making sync errors a different color now.
  13 Feb 23 -- Adding timestamp on the exec binary.
  20 Feb 23 -- Based on copyc.go, now called copyc1.go.  I want to add the verify option to be its own go routine.  But I'm splitting this off as another pgm.
                 So I need another type around which to base a channel for this new go routine to get it's work.
                 And I made the timeFudgeFacter smaller, to 10 ms.
                 I have to really, really, remember that channel receiving for loops do not have a return statement.
  23 Feb 23 -- Added verFlag.
  26 Feb 23 -- I'm tracking down a bug here.  I'm getting %!s(<nil>) displayed, and I don't know why.  I changed a use of Stat to Open, since that's been an issue before on linux.
                 This may be an issue because I was trying to copy a running program.  Windows may not allow me to get a lock on an open file.  Hence, the error is not accurate.
                 The error says that it can't find the file, but this may be an inaccurate error message.
                 On further thought, the error is coming from the verify step.  A binary file in use can't be opened for the verify step.  But it does copy them.
                 So I can copy but not verify a file in use.
  27 Feb 23 -- Fixed a bug in the verify logic.
  13 Mar 23 -- Will limit the # of go routines started to match the # of selected files, if appropriate.
  15 Mar 23 -- Will only start the verify go routines if needed.
  17 Mar 23 -- Changed error from verify operation.
  19 Mar 23 -- Fiddled a bit w/ the number of go routines.
  21 Mar 23 -- Now called copycv, so that it defaults of verify on.
  24 Mar 23 -- listutil_linux fixed case of when bash populates multiple files on command line.  And cleaned up the code.
  28 Mar 23 -- Added message about how many files to be copied.
  31 Mar 23 -- StaticCheck found a few issues.
   5 Apr 23 -- list.ProcessdirectoryAliases was refactored, so I had to refactor here, too.
   8 Apr 23 -- Changed list.New signature.
  10 Apr 23 -- Moved copyAFile to its own separate file.  This will make maintenance easier.
  28 Apr 23 -- It didn't make maintenance easier.  I made a change there to see if I can delete a copied file that threw an error.
   8 Apr 24 -- Added help text to remind me that this is separate because it defaults to verify on.
  28 Jul 24 -- Corrected data race detected in cf2.  I ErrNotNew is no longer global.  It never should have been, anyway.
*/

const LastAltered = "28 July 2024" //

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
var onWin = runtime.GOOS == "windows"
var pooling = runtime.NumCPU() - 3 // account for main, msgChan and verifyChan routines.  Bill Kennedy says that NumCPU() is near the sweet spot.  It's a worker pool pattern.
var cfChan chan cfType
var msgChan chan msgType
var verifyChan chan verifyType
var wg sync.WaitGroup
var succeeded, failed int64

// var ErrNotNew error  This fixes a data race, by not making this global.
var verifyFlag, verFlag, noVerifyFlag bool

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
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, compiled with %s and exec binary timestamp is %s.  This defaults to verify on.\n",
			os.Args[0], LastAltered, runtime.Version(), execTimeStamp)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information: %s [flags] source-files destination-directory\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")

		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		flag.PrintDefaults()
	}

	var revFlag bool
	flag.BoolVar(&revFlag, "r", false, "Reverse the sort, ie, oldest or smallest is first") // Value

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

	flag.BoolVar(&verifyFlag, "verify", true, "Verify that destination is same as source.")
	flag.BoolVar(&verFlag, "ver", false, "Verify copy operation")
	flag.BoolVar(&noVerifyFlag, "no", false, "Turn off default of verify on.")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
		list.VeryVerboseFlag, list.VerboseFlag = true, true
	}

	verifyFlag = verifyFlag || verFlag
	if noVerifyFlag {
		verifyFlag = false
	}
	if verboseFlag {
		fmt.Printf(" VerifylFlag = %t, verFlag = %t, and noVerifyFlag = %t\n", verifyFlag, verFlag, noVerifyFlag)
	}

	//Reverse := revFlag

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

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.ReverseFlag = revFlag
	list.FilterFlag = filterFlag
	list.GlobFlag = globFlag
	list.ExcludeRex = excludeRegex
	list.SizeFlag = sizeFlag

	if verifyFlag {
		verifyChan = make(chan verifyType, pooling)
		go func() {
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
				// I just learned that I can't have a return inside of the channel receive loop.  That stops the message receiving loop.
				// None of the message receiving go routines here have a return statement inside them.
				// I think I've gotten caught by this before.  Hopefully, I'll remember for the next time!
			}
		}()
	} else {
		pooling++ // By doing this here, I don't have to check for pooling < 1.
	}

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

	//fileList, err := list.New(excludeRegex, sizeFlag, Reverse) // fileList used to be []string, but now it's []FileInfoExType.
	fileList, err := list.New() // fileList used to be []string, but now it's []FileInfoExType.
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

	destDir := list.CheckDest()
	if destDir == "" {
		fmt.Print(" Destination directory ? ")
		n, err := fmt.Scanln(&destDir)
		if n == 0 || err != nil {
			destDir = "." + sepString
		}
		if strings.ContainsRune(destDir, ':') {
			//directoryAliasesMap := list.GetDirectoryAliases()
			destDir = list.ProcessDirectoryAliases(destDir)
		} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			destDir = strings.Replace(destDir, "~", homeDirStr, 1)
		}
		if !strings.HasSuffix(destDir, sepString) {
			destDir = destDir + sepString
		}
	}
	//else {
	//	if strings.ContainsRune(destDir, ':') {
	//		//directoryAliasesMap := list.GetDirectoryAliases()
	//		destDir = list.ProcessDirectoryAliases(destDir)
	//	} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
	//		homeDirStr, _ := os.UserHomeDir()
	//		destDir = strings.Replace(destDir, "~", homeDirStr, 1)
	//	}
	//	if !strings.HasSuffix(destDir, sepString) {
	//		destDir = destDir + sepString
	//	}
	//}
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
	if len(fileList) == 0 {
		fmt.Printf(" FileList is empty.  Exiting.\n")
		os.Exit(1)
	}
	if len(fileList) > 10 {
		fmt.Printf(" There are %d files to be copied.", len(fileList))
	}
	fmt.Printf("\n\n")

	num := min(pooling, len(fileList))
	cfChan = make(chan cfType, num)
	for i := 0; i < num; i++ {
		go func() {
			for c := range cfChan {
				CopyAFile(c.srcFile, c.destDir)
			}
		}()
	}

	// time to copy the files

	start := time.Now()
	wg.Add(len(fileList))
	for _, f := range fileList {
		cf := cfType{
			srcFile: f.RelPath,
			destDir: destDir,
		}
		cfChan <- cf
	}
	goRtns := runtime.NumGoroutine()
	close(cfChan)
	wg.Wait()
	close(msgChan)
	if verifyChan != nil {
		close(verifyChan)
	}
	ctfmt.Printf(ct.Cyan, onWin, " Total files copied is %d, total files NOT copied is %d, elapsed time is %s using %d go routines.\n",
		succeeded, failed, time.Since(start), goRtns)
} // end main

//  Copy is now in a separate file as part of this package, ie, package main.

func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}
