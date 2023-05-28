package main // runlist

import (
	"flag"
	"fmt"
	"os/exec"
	//ct "github.com/daviddengcn/go-colortext"
	//ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"golang.org/x/term"
	"os"
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
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.  And how am I going to handle collisions?
  22 Dec 22 -- I'm going to add a display like dsrt, using color to show sizes.  And I'll display the timestamp.  This means that I changed NewList to return []FileInfoExType.
                 So I'm propagating that change thru.
  25 Dec 22 -- Moving the file selection stuff to list.go.
  26 Dec 22 -- Shortened the messages.  And added a timer.
  29 Dec 22 -- Added check for an empty filelist.  And list package code was enhanced to include a sentinel of '.'
   1 Jan 23 -- Now uses list.New instead of list.NewList
   5 Jan 23 -- Adding stats to the output.
   6 Jan 23 -- Now that it clears the screen each time thru the selection loop, I'll print the version message at the end also.
                 Added a stop code of zero.
   7 Jan 23 -- Forgot to init the list.VerboseFlag and list.VeryVerboseFlag.
  22 Jan 23 -- Added Sync call.
  23 Jan 23 -- Added changing destination file(s) timestamp to match the respective source file(s).  And fixed date comparison for replacement copies.
  25 Jan 23 -- Adding verify.
  28 Jan 23 -- Adding verify success message.
  30 Jan 23 -- Will add 1 sec to file timestamp on linux.  This is to prevent recopying the same file over itself (I hope).  Added timeFudgeFactor
  31 Jan 23 -- timeFudgeFactor is now a Duration.
  31 Mar 23 -- StaticCheck found a few issues.
   5 Apr 23 -- Refactored list.ProcessDirectoryAliases
   8 Apr 23 -- Changed list.New signature.
  26 May 23 -- Now called runlist, based on copylist.  I intend this to be like executable extensions on Windows.  The command is the first param, and the list follows.

type FileInfoExType struct {
	FI       os.FileInfo
	Dir      string
	RelPath  string
	AbsPath  string
	FullPath string // probably not needed, but I really do want to be complete.
}

*/

const LastAltered = "27 May 2023" //

const defaultHeight = 40
const minWidth = 90

var autoWidth, autoHeight int
var err error
var verifyFlag bool

func main() {
	fmt.Printf("%s is compiled w/ %s, last altered %s.\n", os.Args[0], runtime.Version(), LastAltered)
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

	flag.BoolVar(&verifyFlag, "verify", false, "Verify copy operation")

	var globFlag bool
	flag.BoolVar(&globFlag, "g", false, "glob flag to use globbing on file matching.")

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

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.FilterFlag = filterFlag
	list.ReverseFlag = revFlag
	list.ExcludeRex = excludeRegex
	list.SizeFlag = sizeFlag
	list.GlobFlag = globFlag

	fileList, err := list.SkipFirstNewList()
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
		fmt.Printf("\n\n")
	}

	// Convert from []FileInfoX to []string
	fileNameStr := make([]string, 0, len(fileList))
	for _, f := range fileList {
		fileNameStr = append(fileNameStr, f.FI.Name())
	}

	// now have the fileListStr.  Need to get the cmdStr.

	cmdStr := flag.Arg(0) // this means the first param on the command line, if present.  If not present, that's ok and will mean the empty command, like an executable extension on Windows.
	if cmdStr == "." {
		cmdStr = "" // this is for windows and executable extensions.
	}

	if verboseFlag {
		fmt.Printf("\n cmdStr = %q\n", cmdStr)
	}

	// Time to run the cmd.

	var cmdPath string
	var execCmd *exec.Cmd
	variadicParam := make([]string, 0, len(fileNameStr))

	//variadicParam := []string{"-C", "vlc"} // This isn't really needed anymore.  I'll leave it here anyway, as a model in case I ever need to do this again.
	//variadicParam = append(variadicParam, fileNameStr...)

	// For me to be able to pass a variadic param here, I must match the definition of the function, not pass some and then try the variadic syntax.
	// I got this answer from stack overflow.

	cmdPath = cmdStr
	if cmdStr == "" {
		if runtime.GOOS == "linux" {
			cmdPath = "/bin/bash"
			variadicParam = append(variadicParam, fileNameStr...)
		} else { // must be on Windows.
			cmdPath = os.Getenv("COMSPEC")
			variadicParam = append(variadicParam, "-C")
			variadicParam = append(variadicParam, fileNameStr...)
		}
	}
	execCmd = exec.Command(cmdPath, variadicParam...)

	if verboseFlag {
		fmt.Printf(" cmdStr = %s, cmdPath = %s, len of fileNameStr = %d, and filenames in fileNameStr are %v\n",
			cmdStr, cmdPath, len(fileNameStr), fileNameStr)
		fmt.Printf(" Len(variadiacParam) = %d, variadiacParam = %#v\n", len(variadicParam), variadicParam)
	}

	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	e := execCmd.Start()
	if e != nil {
		fmt.Printf(" Error returned by running %s %s is %v\n", cmdStr, fileNameStr, e)
	}
} // end main
