package main // cf3, from cf2, from cf, for copy fanout.  This one is truly a fanout pattern, and now adds viper.

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/spf13/pflag" // docs say that pflag is a drop in replacement for the standard library flag package
	"github.com/spf13/viper"
	"golang.org/x/term"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/few"
	"src/list"
	"strings"
	"sync"
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
                 Verify is checked in the copyAFile routine.
  13 Feb 23 -- Adding timestamp on the exec binary.
  20 Feb 23 -- Modified the verification failed message.
  23 Feb 23 -- Added verFlag.
  13 Mar 23 -- Will only create the lesser of number of files selected vs NumCPU() go routines for the copy operation.  And made the timeFudgeFactor = 10 ms.
                 And fixed a bug in how the verify operation works.
  17 Mar 23 -- Changed error message when verify returns an error.
  21 Mar 23 -- Completed the usage message, which was never completed.
  24 Mar 23 -- listutil_linux fixed case of when bash populates multiple files on command line.  And cleaned up the code.
  28 Mar 23 -- Added message saying how many files to be copied.
  31 Mar 23 -- StaticCheck found a few issues.
   5 Apr 23 -- Fixed list.CheckDest.
   8 Apr 23 -- Changed list.New signature.
  10 Apr 23 -- Moved copyAFile to its own separate file.  This will make maintenance easier.  Scratch that.  I forgot that the copyAFile routines are not all identical.
                 I'm moving it back to be here now.
  23 Apr 23 -- Seems to not have worked, the deletion of the copy when there's an error.  The case I saw was an error from Sync() did not erase the copy.  I'm adding a printf statement.
                 I think I found the problem.  I have to not return the message, as when the message is returned, the wait group is decremented so the main pgm exits
                 before the os.Erase is called.  I can only return 1 message per call to CopyAFile.
  27 Apr 23 -- At work I got an error that O: was full, and when this pgm tried to delete whatever did copy, I saw the error that the file is in use by another process.  So I'll try
                 closing the output file before calling os.Remove and see if that will work.
   6 May 23 -- Finally was able to test the error handling code here, on leox.  The Sync() step failed for 2 files.  Both were successfully deleted automatically.  Then I
                 ran the pgm again, and these were copied in the 2nd try.  Hooray!
  25 May 23 -- Changed the final message to be multicolored.
   8 Jul 23 -- I fixed part where dest dir is tested.
  26 Aug 23 -- I'm going to change the final message to suppress when zero files were copied or not copied.
  10 Feb 24 -- Making the timeFudgeFactor 1 ms
  11 Feb 24 -- I removed the min func, so the code will use the built-in func of min.  This was new in Go 1.21.  As I write this, I'm  now compiling w/ Go 1.22.
   6 Apr 24 -- Shortened the destination file is same or older message.
   8 Apr 24 -- Now shows the last altered dated for list.go
   9 Apr 24 -- Found an error in CopyAFile, in that I don't check for an error when I close the file.
               Listening to Miki Tebeka from Ardan Labs, he said that for I/O bound, you can spin up more goroutines than runtime.NumCPU() indicates.
               But for CPU bound, there's no advantage to exceeding that number.
------------------------------------------------------------------------------------------------------------------------------------------------------
  10 Apr 24 -- Now called cf, for copy fanout.  I'll use a multiplier, default 10, and set by a param in flag package.  I'm going to see if more is better for this I/O bound task.
  15 Jun 24 -- Changed completion message.
------------------------------------------------------------------------------------------------------------------------------------------------------
  27 Jun 24 -- Now called cf2, and will truly be a fanout pattern.  If it tries to copy > 900 files, it will complain but only copy 1st 900 files.
				And I changed how it tallies hits and misses, removing the atomic add as unnecessary.
   6 July 24-- Changed the startup message
  26 July 24-- Adding timing info from the individual goroutines that do the copying.  I have to expand the message sent on the channel to include a duration.
  28 July 24-- Race detector found a data race because ErrNotNew was global and being written to by multiple goroutines.  I fixed it by not making it global.
  17 Aug 24 -- Changed start message to make it clearer that globbing is used here, not regexp to match included files.  A regexp is used to exclude files.
  22 Oct 24 -- Will now check to make sure params are present.
  30 Nov 24 -- Noticed that sometimes the colors on tcc get confused.  I'm going to output on Windows a color reset message.
------------------------------------------------------------------------------------------------------------------------------------------------------
  14 Jan 25 -- Now called cf3, based on cf2 but adds viper for config stuff.  And adds filterStr code from dsrt family of rtns.
				Because list uses flag.NArgs, I had to get creative in the list code to determine which flag package is in use.  They conflict, so both can't be used in the same package.
				See list.go and listutil_windows.go and listutil_linux.go for more details.
  20 Jan 25 -- Added set-up timing display.
   3 May 25 -- When there are errors in the dest directory, that needs to be more obvious.
*/

const LastAltered = "3 May 2025" //

const defaultHeight = 40
const minWidth = 90
const sepString = string(filepath.Separator)
const timeFudgeFactor = 1 * time.Millisecond
const fanoutMax = 900
const configFilename = "cf3.yaml"

type msgType struct {
	s        string
	e        error
	color    ct.Color
	elapsed  time.Duration
	success  bool
	verified bool
}

var autoWidth, autoHeight int
var onWin = runtime.GOOS == "windows"

// Week of Feb 2024, Miki Tebeka gave an ultimate Go class.  In it he says that I/O bound work is not limited by runtime.NumCPU(), only cpu bound work is.
var msgChan chan msgType
var wg sync.WaitGroup
var succeeded, failed int64
var verifyFlag, verFlag bool
var multiplier int

func main() {
	t1 := time.Now()
	winflag := runtime.GOOS == "windows" // this is needed because I use it in the color statements, so the colors are bolded only on windows.
	execName, err := os.Executable()
	if err != nil {
		fmt.Printf(" Error from os.Executable() is: %s.  This will be ignored.\n", err)
	}
	execFI, err := os.Lstat(execName)
	if err != nil {
		fmt.Printf(" Error from os.Lstat(%s) is: %s.  This will be ignored\n", execName, err)
	}
	execTimeStamp := execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	fmt.Printf("%s using globbing, a true fanout pattern and viper.  Compiled w/ %s, last altered %s, list.go last altered %s, exec binary timestamp is %s\n",
		os.Args[0], runtime.Version(), LastAltered, list.LastAltered, execTimeStamp)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		ctfmt.Printf(ct.Red, winflag, "Error getting user home directory is %s\n", err.Error())
		return
	}
	fullConfigFileName := filepath.Join(homeDir, configFilename)
	pflag.Usage = func() {
		fmt.Printf(" %s last altered %s, compiled with %s and exec binary timestamp is %s. \n", os.Args[0], LastAltered, runtime.Version(), execTimeStamp)
		fmt.Printf(" Usage information: %s [flags] src-files dest-dir\n", os.Args[0])
		fmt.Printf(" AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		fmt.Printf(" Now uses viper config file of cf3.yaml\n")
		fmt.Printf(" Reads from diraliases environment variable if needed on Windows.\n")
		pflag.PrintDefaults()
	}

	var revFlag bool
	pflag.BoolVarP(&revFlag, "reverse", "r", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var sizeFlag bool
	pflag.BoolVarP(&sizeFlag, "size", "s", false, "sort by size instead of by date")

	var verboseFlag, veryVerboseFlag bool

	pflag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode, which is same as test mode.")
	pflag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

	//var excludeFlag bool  not used anymore, now I can test exclude regex against nil.
	//pflag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")
	var excludeRegex *regexp.Regexp
	var excludeRegexPattern string
	pflag.StringVarP(&excludeRegexPattern, "exclude", "x", "", "regex to be excluded from output.") // var, not a ptr.

	var filterFlag, noFilterFlag bool
	var filterStr string
	pflag.StringVar(&filterStr, "filterstr", "", "individual size filter value below which listing is suppressed.")
	pflag.BoolVarP(&filterFlag, "filter", "f", false, "filter value to suppress listing individual size below 1 MB.")
	pflag.BoolVarP(&noFilterFlag, "nofilter", "F", false, "Flag to undo an environment var with f set.")

	var globFlag bool
	pflag.BoolVar(&globFlag, "g", false, "glob flag to use globbing on file matching.")

	pflag.BoolVar(&verifyFlag, "verify", false, "Verify that destination is same as source.")
	pflag.BoolVar(&verFlag, "ver", false, "Verify copy operation")
	pflag.IntVarP(&multiplier, "multiplier", "m", 10, "Multiplier of NumCPU() for the worker pool pattern, or limited fanout.  Default is 10.")

	//flag.Parse()  //  this is here because list checkDest needs it, so I'm initializing both.  Nevermind, as these seem to be clashing.
	pflag.Parse()

	if pflag.NArg() < 2 {
		ctfmt.Printf(ct.Red, true, " Not enough params on command line.  Two needed, but found %d\n", pflag.NArg())
		return
	}

	// viper stuff
	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		ctfmt.Printf(ct.Red, winflag, "Error binding flags is %s.  Binding is ignored.\n", err.Error())
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(fullConfigFileName)

	//AutomaticEnv makes Viper check if environment variables match any of the existing keys (config, default or flags). If matching env vars are found, they are loaded into Viper.
	// This seems to be working.
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	verboseFlag = viper.GetBool("verbose")
	veryVerboseFlag = viper.GetBool("vv")
	if veryVerboseFlag { // setting veryVerbose flag will automatically set verboseFlag
		verboseFlag = true
		list.VeryVerboseFlag, list.VerboseFlag = true, true
	}

	if err != nil { // err from ReadInConfig is checked now that verboseFlag is init'd
		if verboseFlag {
			ctfmt.Printf(ct.Red, winflag, " %s.  Ignored\n", err.Error())
		}
	}

	if verboseFlag {
		fmt.Printf("Config file name: %s\n", fullConfigFileName)
	}
	revFlag = viper.GetBool("reverse")
	sizeFlag = viper.GetBool("size")

	filterStr = viper.GetString("filterstr")
	filterFlag = viper.GetBool("filter")
	noFilterFlag = viper.GetBool("nofilter") // the noFilterFlag takes priority.
	if noFilterFlag {
		filterFlag = false
	}

	globFlag = viper.GetBool("glob")

	verifyFlag = viper.GetBool("verify")
	verFlag = viper.GetBool("ver")
	verifyFlag = verifyFlag || verFlag

	multiplier = viper.GetInt("multiplier")

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", ExecFI.Name(), ExecTimeStamp, execName)
		fmt.Println()
		fmt.Printf(" globFlag=%t, verifyFlag=%t, verFlag=%t, multiplier=%d, revFlag=%t, sizeFlag=%t, filterStr=%q, filterFlag=%t, noFilterFlag=%t \n",
			globFlag, verifyFlag, verFlag, multiplier, revFlag, sizeFlag, filterStr, filterFlag, noFilterFlag)
	}

	excludeRegexPattern = viper.GetString("exclude")
	if len(excludeRegexPattern) > 0 {
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err = regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
		}
		if verboseFlag {
			fmt.Printf(" excludeRegexPattern found and is %d runes. ", len(excludeRegexPattern))
			fmt.Printf(" excludeRegexPattern = %q, excludeRegex.String = %q (these should be equal)\n", excludeRegexPattern, excludeRegex.String())
		}
	}

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.ReverseFlag = revFlag
	list.FilterFlag = filterFlag
	list.FilterStr = filterStr
	list.GlobFlag = globFlag
	list.ExcludeRex = excludeRegex
	list.SizeFlag = sizeFlag

	ctfmt.Printf(ct.Green, winflag, "%50s Set up time is %s \n", " ", time.Since(t1))

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
			//                                                         directoryAliasesMap := list.GetDirectoryAliases()
			destDir = list.ProcessDirectoryAliases(destDir)
		} else if strings.Contains(destDir, "~") { // this can only contain a ~ on Windows.
			homeDirStr, _ := os.UserHomeDir()
			destDir = strings.Replace(destDir, "~", homeDirStr, 1)
		}
		if !strings.HasSuffix(destDir, sepString) {
			destDir = destDir + sepString
		}
	}
	fmt.Printf("\n destDir = %#v\n", destDir)
	//                                                  fi, err := os.Lstat(destDir)  this was giving errors sometimes.
	d, err := os.Open(destDir)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " os.Open(%s) failed w/ error %s.  Exiting\n", destDir, err)
		return
	}
	fi, err := d.Stat()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " %s.Stat() failed w/ error %s.  Exiting\n", d.Name(), err)
		return
	}
	if !fi.IsDir() {
		ctfmt.Printf(ct.Red, true, " %s is supposed to be the destination directory, but stat(%s) not c/w a directory.  Exiting\n", destDir, destDir)
		return
	}
	d.Close()

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
		fmt.Printf(" FileList is empty.  Exiting.\n\n")
		os.Exit(1)
	}
	if len(fileList) > 10 {
		fmt.Printf(" There are %d files to be copied.", len(fileList))
	}
	fmt.Printf("\n\n")

	// Time to check if have too many files to copy.

	numOfFiles := len(fileList)
	if numOfFiles > fanoutMax {
		fmt.Printf(" There are %d files to be copied, which exceeds the max of %d.  Bye-bye.\n", len(fileList), fanoutMax)
		fileList = fileList[0:fanoutMax]
	}

	// Start the message channel
	msgChan = make(chan msgType, numOfFiles)
	go func() {
		for msg := range msgChan {
			if msg.success {
				ctfmt.Printf(msg.color, onWin, " %s\n", msg.s)
				succeeded++ // there is only 1 go routine, so this is not a data race
			} else {
				ctfmt.Printf(msg.color, onWin, " %s\n", msg.e)
				failed++ // there is only 1 go routine, so this is not a data race
			}
			wg.Done()
		}

	}()

	// time to copy the files

	start := time.Now()
	wg.Add(numOfFiles)
	for _, f := range fileList {
		go copyAFile(f.AbsPath, destDir) // this is the line for the true fanout pattern.
	}
	goRtnsNum := runtime.NumGoroutine()

	wg.Wait()
	close(msgChan)
	fmt.Printf(" \n %s:", os.Args[0])
	if succeeded > 0 {
		ctfmt.Printf(ct.Green, onWin, " Total files copied is %d, ", succeeded)
	}
	if failed > 0 {
		ctfmt.Printf(ct.Red, onWin, " Total files NOT copied is %d, ", failed)
	}
	ctfmt.Printf(ct.Cyan, onWin, " total elapsed time is %s using %d go routines for %s.\n", time.Since(start), goRtnsNum, os.Args[0])

	if onWin { // because tcc can get confused about it's color scheme sometimes.  Probably a bug.  But I'm glad that tcmd+tcc finally works w/ vim.
		ctfmt.Printf(ct.White, onWin, " Total files processed is %d\n", succeeded+failed)
	}
} // end main

//	------------------------------------ CopyAFile ----------------------------------------------
//
// CopyAFile(srcFile, destDir string) where src is a regular file.  destDir is a directory
func copyAFile(srcFile, destDir string) {
	// I have to open the file and write it to copy it.
	// Here, src is a regular file, and dest is a directory.  I have to construct the dest filename using the src filename.
	// This routine adds the time fudge factor to the copied file, because I discovered on linux that if I don't do this, the routine will not detect the copy timestamp is the same as the source timestamp.
	// I think this is because of the monotonic clock.  I found that by adding a small amount of time to the copied file, the copy is detected as later than the source, which is what I want.

	t0 := time.Now()
	in, err := os.Open(srcFile)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       fmt.Errorf("elapsed %s: %s", time.Since(t0), err),
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
			e:       fmt.Errorf("elapsed %s: %s", time.Since(t0), err),
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
			ErrNotNew := fmt.Errorf("elapsed %s: %s is not newer than %s", time.Since(t0), baseFile, destDir) // now this is not a data race.
			msg := msgType{
				s:       "",
				e:       ErrNotNew,
				elapsed: time.Since(t0),
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

	t0 = time.Now() // redefine t0 now that it's going to try to copy the file.
	_, err = io.Copy(out, in)

	if err != nil {
		var msg msgType

		e := out.Close() // close it so I can delete it and not get the error that the file is in use by another process.
		if e != nil {
			msg = msgType{
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
			msg = msgType{
				s: "",
				e: fmt.Errorf("ERROR from io.Copy was %s, so it was closed w/ error of %v, and %s was deleted.  There was no error returned from os.Remove(%s)",
					err, e, outName, outName),
				color:    ct.Yellow, // so I see it
				success:  false,
				verified: false,
			}
			msgChan <- msg
		} else {
			msg = msgType{
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
				e: fmt.Errorf("elapsed %s: ERROR from Sync() was %s, so it was closed w/ error of %v, and %s was deleted.  There was no error from os.Remove(%s)",
					time.Since(t0), err, e, outName, outName),
				color:    ct.Yellow, // yellow to make sure I see it.
				elapsed:  time.Since(t0),
				success:  false,
				verified: false,
			}
			msgChan <- msg
		} else {
			msg = msgType{
				s: "",
				e: fmt.Errorf("elapsed %s: ERROR from Sync() was %s, so it was closed w/ error of %v, and os.Remove(%s) was called.  The error from os.Remove was %s",
					time.Since(t0), err, e, outName, er),
				color:    ct.Yellow, // yellow to make sure I see it.
				elapsed:  time.Since(t0),
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
			elapsed: time.Since(t0),
			success: false,
		}
		msgChan <- msg
		return
	}
	t := inFI.ModTime()
	if runtime.GOOS == "linux" {
		t = t.Add(timeFudgeFactor)
	}

	err = os.Chtimes(outName, t, t) // name string, atime time.Time, mtime time.Time
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			elapsed: time.Since(t0),
			success: false,
		}
		msgChan <- msg
		return
	}

	if verifyFlag {
		result, err := few.Feq32withNames(srcFile, outName)
		if err != nil {
			msg := msgType{
				s:        "",
				e:        fmt.Errorf("elapsed %s: ERROR from verify operation is %s", time.Since(t0), err),
				color:    ct.Red,
				elapsed:  time.Since(t0),
				success:  false,
				verified: false,
			}
			msgChan <- msg
			return
		}
		if result {
			msg := msgType{
				s:        fmt.Sprintf("elapsed %s: %s copied to %s and is VERIFIED", time.Since(t0), srcFile, destDir),
				e:        nil,
				color:    ct.Green,
				elapsed:  time.Since(t0),
				success:  true,
				verified: true,
			}
			msgChan <- msg
			return
		} else {
			msg := msgType{
				s:        fmt.Sprintf("elapsed %s: %s copied to %s but failed VERIFICATION", time.Since(t0), srcFile, destDir),
				e:        nil,
				color:    ct.Red,
				elapsed:  time.Since(t0),
				success:  false,
				verified: false,
			}
			msgChan <- msg
			return
		}
	}

	elapsed := time.Since(t0)
	msg := msgType{
		s:        fmt.Sprintf("elapsed %s: %s copied to %s", elapsed, srcFile, destDir),
		e:        nil,
		color:    ct.Green,
		elapsed:  elapsed,
		success:  true,
		verified: verifyFlag, // I already know that this flag is false if get here.
	}
	msgChan <- msg
	// return  this is redundant.
} // end CopyAFile
