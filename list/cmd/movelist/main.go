package main // copyc

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"src/list"
	"strings"
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
   4 Apr 23 -- Now called copyCP, which will pass the variadic list of files to either tcc copy or cp.  copyC and listVLC were used to write this routine.
   5 Apr 23 -- Fixed list.CheckDest.
   6 Apr 23 -- Will wait for the shell to finish, so I can time it and be clearer when this routine is finished.
   8 Apr 23 -- Changed list.New signature.
*/

const LastAltered = "8 Apr 2023" //

const sepString = string(filepath.Separator)

var onWin = runtime.GOOS == "windows"

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
	fmt.Printf("%s is compiled w/ %s, last altered %s, exec binary timestamp is %s\n", os.Args[0], runtime.Version(), LastAltered, execTimeStamp)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, compiled with %s and exec binary timestamp is %s. \n", os.Args[0], LastAltered, runtime.Version(), execTimeStamp)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information: %s [flags] src-files dest-dir -- which will get passed to the shell copy pgm.\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from dsrt environment variable before processing commandline switches.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " Reads from diraliases environment variable if needed on Windows.\n")
		//fmt.Fprintf(flag.CommandLine.Output(), " AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
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

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
		list.VeryVerboseFlag, list.VerboseFlag = true, true
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
	//else {  This code belongs in list.CheckDest, and now is where it belongs.
	//	if strings.ContainsRune(destDir, ':') {
	//		directoryAliasesMap := list.GetDirectoryAliases()
	//		destDir = list.ProcessDirectoryAliases(directoryAliasesMap, destDir)
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

	fileListStr, err := list.FileSelectionString(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}
	if verboseFlag {
		for i, f := range fileListStr {
			fmt.Printf(" second fileList[%d] = %s\n", i, f)
		}
		fmt.Println()
		fmt.Printf(" There are %d files in the file list.\n", len(fileListStr))
	}
	if len(fileListStr) == 0 {
		fmt.Printf(" FileListString is empty.  Exiting.\n\n")
		os.Exit(1)
	}
	if len(fileListStr) > 10 {
		fmt.Printf(" There are %d files to be copied.", len(fileListStr))
	}
	fmt.Printf("\n\n")

	// Time to find the shell copy or cp.

	var shellStr string
	var ok bool
	var execCmd *exec.Cmd
	variadicParam := make([]string, 0, len(fileListStr)+3) // since I'm adding 3 params to variadicParam on Win.

	if onWin {
		shellStr, ok = os.LookupEnv("ComSpec")
		if !ok {
			fmt.Printf(" After os.LookupEnv(ComSpec), got a return of not ok.  ShellStr = %s\n", shellStr)
			os.Exit(1)
		}
		variadicParam = []string{"/C", "*copy", "/u"}         // start the variadic param w/ these required params, the first one has tcc only run 1 cmd and then exit.
		variadicParam = append(variadicParam, fileListStr...) // now append all the files to be copied.
	} else if runtime.GOOS == "linux" { // just in case this ever gets attempted using macOS.
		shellStr = "cp"
		variadicParam = []string{"-u", "-v"}                  // start the variadic param w/ these required params
		variadicParam = append(variadicParam, fileListStr...) // now append all the files to be copied.
	}
	variadicParam = append(variadicParam, destDir)

	//fmt.Printf(" Debug: shellStr = %s\n variadicParam = %s\n", shellStr, variadicParam)

	// time to copy the files

	execCmd = exec.Command(shellStr, variadicParam...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	t0 := time.Now()
	//err = execCmd.Start() // this does not wait for it to finish, so I can't time it this way.
	err = execCmd.Run() // this does wait for it to finish, so I'll time it.
	if err != nil {
		fmt.Printf(" Error returned by running %s %s is %v\n", shellStr, variadicParam, err)
	}

	ctfmt.Printf(ct.Cyan, onWin, " Sent %d files to %s, which took %s.\n", len(fileListStr), shellStr, time.Since(t0))
} // end main
