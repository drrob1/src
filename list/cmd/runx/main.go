package main // runx

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/jonhadfield/findexec"
	flag "github.com/spf13/pflag"
	"golang.org/x/term"
	"os"
	"os/exec"
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
  29 May 23 -- Changed behavior on Windows.  Now I look to see if tcc or cmd is running; tcc uses the -C flag and uses .Start(), while cmd does not use the -C flag and uses .Run()
                 And will look for "xl" to change to "excel", and "w" to "winword".  I don't think I need to map "a" to msaccess or "p" to powerpnt.
  31 May 23 -- Expanding substitutions to p = powerpnt, a = msaccess, and l = libreoffice.  And I'm thinking about how to implement my own executable extensions.
                 That works.  Now I want to be able to enter the code for the office pgm, and it will just show me files that will open in that pgm.  But I still have to allow
                 executable extensions, like for pdf or txt files on Windows.
                 runlist
                 runlist p|l|a|x|w
                 runlist . glob -- behaves differently on linux and Windows.
   1 Jun 23 -- Uses the new list.FileInfoXFromGlob and list.NewFromGlob.
  12 Jul 23 -- Globbing doesn't work.  Nevermind.  I forgot that 1st param has to be a dot if I want to glob.  I added a check against an empty fileList.
                 I'm going to add a check to remind me if I forget again.
  14 Jul 23 -- I made the first param as a dot optional.
  15 Jul 23 -- I fucked up the automatic globbing by appropriate extension.  I have to put that back.  If there is no glob string on command line, use the default glob string.
                 Else, use the one provided.  Or, have the glob string a flag.  That is probably much easier to implement.  I'll make globStr global, and allow it to be set
                 as a param.  If it's not, use the default.  I already have a globFlag.  For this to work the same on Windows and linux, I have to have a separate glob string
                 as a param.  I'll do that.  So this will not use the globFlag.
                 On linux, this only works w/ LibreOffice.  So I'll automatically select that on linux.
  16 Jul 23 -- Now called runx.go.  I intend to have this also allow -rex specifier.  If regex is given, that will take priority over the default or given glob expression.
  24 Sep 23 -- I have to fix a bug.  It's not working as I expect on linux.  And Doug is 35 today.  But that's not why I'm here, in the code now.
   8 Jun 24 -- Fine-tuned the error message.  I forgot how this worked.
  14 Jun 25 -- Added output of execCmd to verboseFlag.
  21 Jun 25 -- Now using pflag as flag.
*/

const LastAltered = "21 June 2025" //

const defaultHeight = 40
const minWidth = 90

var autoWidth, autoHeight int
var err error
var verifyFlag bool
var officePath = "c:/Program Files/Microsoft Office/root/Office16;"
var globString string
var regex *regexp.Regexp

func main() {
	fmt.Printf("%s is compiled w/ %s, last altered %s.\n", os.Args[0], runtime.Version(), LastAltered)
	autoWidth, autoHeight, err = term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		//autoDefaults = false
		autoHeight = defaultHeight
		autoWidth = minWidth
	}

	flag.Usage = func() {
		fmt.Printf(" %s last altered %s, and compiled with %s.  Now uses pflag package. \n", os.Args[0], LastAltered, runtime.Version())
		fmt.Printf(" Usage information: [ x|w|p|a|l ].  Glob pattern is set by the -g flag, and -rex is implemented.\n")
		fmt.Printf(" If have both -rex and -g, -rex is followed and -g is ignored, as is the x|w|p|a pattern.\n")
		fmt.Printf(" In this way, this program works the same on both Windows and Linux.\n")
		fmt.Printf(" AutoHeight = %d and autoWidth = %d.\n", autoHeight, autoWidth)
		flag.PrintDefaults()
	}

	var revFlag bool
	flag.BoolVarP(&revFlag, "reverse", "r", false, "Reverse the sort, ie, oldest or smallest is first.") // Value

	var sizeFlag bool
	flag.BoolVarP(&sizeFlag, "size", "s", false, "sort by size instead of by date.")

	var verboseFlag, veryVerboseFlag bool

	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode, which is same as test mode.")
	flag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "Very verbose debugging option.")

	//var excludeFlag bool
	//flag.BoolVar(&excludeFlag, "exclude", false, "exclude regex entered after prompt")  Never used anyway.
	var excludeRegex *regexp.Regexp
	var excludeRegexPattern string
	flag.StringVarP(&excludeRegexPattern, "exclude", "x", "", "regex to be excluded from output.") // var, not a ptr.

	var filterFlag, noFilterFlag bool
	var filterStr string
	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVarP(&filterFlag, "filt", "f", false, "filter value to suppress listing individual size below 1 MB.")
	flag.BoolVarP(&noFilterFlag, "nofilter", "F", false, "Flag to undo an environment var with f set.")

	flag.BoolVarP(&verifyFlag, "verify", "V", false, "Verify copy operation (not used here).")

	var globFlag bool
	flag.BoolVar(&globFlag, "G", false, "glob flag to use globbing on file matching.") // essentially ignored.
	flag.StringVarP(&globString, "glob", "g", "", "Use this glob string pattern instead of the defaults.")

	var regexStr string
	flag.StringVar(&regexStr, "rex", "", "Regexp to use for creating the list.")

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
			// excludeFlag = false  this isn't used anywhere else
		}
		// excludeFlag = true  this isn't used anywhere else
		fmt.Printf(" excludeRegexPattern = %q, excludeRegex.String = %q\n", excludeRegexPattern, excludeRegex.String())
	}

	if regexStr != "" {
		regex, err = regexp.Compile(regexStr)
		if err != nil {
			fmt.Printf(" Regular Expression of %s returned error of %s.  Ignored.\n", regexStr, err)
		}
	}

	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.FilterFlag = filterFlag
	list.ReverseFlag = revFlag
	list.ExcludeRex = excludeRegex
	list.SizeFlag = sizeFlag
	list.GlobFlag = globFlag

	// Need to get the cmdStr.  cmd.exe behaves differently than tcc.exe

	var cmdStr, globStr string
	var fileList []list.FileInfoExType
	var err error

	if flag.NArg() == 0 {
		if runtime.GOOS == "windows" {
			cmdStr = ""
		} else {
			cmdStr = "libreoffice"
		}
		if regex != nil {
			fileList, err = list.NewFromRegexp(regex)
		} else {
			globStr = "*"
			if globString != "" {
				globStr = globString
			}
			fileList, err = list.NewFromGlob(globStr)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from either list.NewFromRegexp or list.NewFromGlob is %s.  Exiting\n", err)
			os.Exit(1)
		}
	} else if flag.NArg() == 1 { // use default glob string.  Or, I forgot to enter the first param letter.  I'll check.
		cmdStr = flag.Arg(0) // this means the first param on the command line, if present.  If not present, that's ok and will mean the empty command, like an executable extension on Windows.
		if cmdStr == "." {
			cmdStr = "" // this is for windows and executable extensions.
			globStr = "*"
			if globString != "" {
				globStr = globString
			}
		} else if strings.ToLower(cmdStr) == "xl" || strings.ToLower(cmdStr) == "x" { // These only apply to MS-Office on Windows.
			cmdStr = "excel"
			globStr = "*.xls*"
			if globString != "" {
				globStr = globString
			}
		} else if strings.ToLower(cmdStr) == "w" {
			cmdStr = "winword"
			globStr = "*.doc*"
			if globString != "" {
				globStr = globString
			}
		} else if strings.ToLower(cmdStr) == "p" {
			cmdStr = "powerpnt"
			globStr = "*.ppt*"
			if globString != "" {
				globStr = globString
			}
		} else if strings.ToLower(cmdStr) == "a" {
			cmdStr = "msaccess"
			globStr = "*.mdb"
			if globString != "" {
				globStr = globString
			}
		} else if strings.ToLower(cmdStr) == "l" || runtime.GOOS == "linux" {
			cmdStr = "libreoffice"
			globStr = "*"
			if globString != "" {
				globStr = globString
			}
		} else if cmdStr == "" {
			globStr = "*"
			if globString != "" {
				globStr = globString
			}
		} else { // Since I'm not allowing globbing on the command line, bash won't screw me up.  This is why I'm not allowing globbing on the command line.
			fmt.Printf(" First param is not .|xl|x|w|p|a|l.  Remember that glob pattern is now set by -g flag or regexp set by -rex flag.  And only 1 param is allowed.\n")
			fmt.Printf(" This is so this pgm works the same on both Windows and Linux.")
			fmt.Printf(" cmdStr = %q, globStr = %q, globString = %q, regexStr = %q, NArg = %d", cmdStr, globStr, globString, regexStr, flag.NArg())
			if regex != nil {
				fmt.Printf(" regex = %s\n", regex.String())
			}
			fmt.Println()
			os.Exit(1)
		}

		if regex != nil {
			fileList, err = list.NewFromRegexp(regex)
		} else {
			fileList, err = list.NewFromGlob(globStr)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from list.GetFileInfoXFromCommandLine or SkipFirstNewList is %s\n", err)
			os.Exit(1)
		}
	} else { // Since I'm not allowing globbing on the command line, bash won't screw me up.  This is why I'm not allowing globbing on the command line.
		fmt.Printf(" First param is not .|xl|x|w|p|a|l.  Glob pattern is now set by -g flag, and regex set by -rex.  Try again.\n")
		fmt.Printf(" cmdStr = %q, globStr = %q, globString = %q, regexStr = %q, NArg = %d", cmdStr, globStr, globString, regexStr, flag.NArg())
		if regex != nil {
			fmt.Printf(" regex = %s\n", regex.String())
		}
		fmt.Printf(" Remember that strings may have to be quoted to be recognized as params.\n")
		fmt.Println()
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf("\n cmdStr=%q, globStr=%q, globString=%q, regexStr = %q, len(fileList) = %d", cmdStr, globStr, globString, regexStr, len(fileList))
		if regex == nil {
			fmt.Printf(" regex is nil.\n")
		} else {
			fmt.Printf(" regex is %s.\n\n", regex.String())
		}
	}

	if veryVerboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %#v\n", i, f)
		}
		fmt.Println()
	}
	if len(fileList) == 0 {
		fmt.Printf(" Length of the filelist before selection stage is zero.  Aborting\n")
		os.Exit(1)
	}

	fileList, err = list.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}

	if len(fileList) == 0 {
		fmt.Printf(" Length of the selected filelist is zero.  Aborting\n")
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

	// Time to run the cmd.

	var cmdPath string
	var execCmd *exec.Cmd
	var cmd bool
	variadicParam := make([]string, 0, len(fileNameStr))

	//variadicParam := []string{"-C", "vlc"} // This isn't really needed anymore.  I'll leave it here anyway, as a model in case I ever need to do this again.

	// For me to be able to pass a variadic param here, I must match the definition of the function, not pass some and then try the variadic syntax.
	// I got this answer from stack overflow.

	cmdPath = cmdStr
	fmt.Printf(" cmdStr = %q, cmdPath = %q\n", cmdStr, cmdPath)
	if cmdPath == "" {
		if runtime.GOOS == "linux" {
			cmdPath = "/bin/bash"
		} else { // must be on Windows.
			cmdPath = strings.ToLower(os.Getenv("COMSPEC"))
			if strings.Contains(cmdPath, "tcc") {
				variadicParam = append(variadicParam, "-C")
			} else { // running cmd.exe, and likely at work.
				cmd = true // and variadicParam won't have the -C flag
			}
		}
	}
	variadicParam = append(variadicParam, fileNameStr...)

	if cmdStr == "excel" || cmdStr == "winword" || cmdStr == "powerpnt" || cmdStr == "msaccess" {
		searchPath := officePath + os.Getenv("PATH")
		execStr := findexec.Find(cmdStr, searchPath)
		if execStr == "" {
			ctfmt.Printf(ct.Red, true, " execStr is blank because could not find %s.  \nsearchPath = %s \n Exiting.\n", cmdStr, searchPath)
			os.Exit(1)
		}
		cmdPath = execStr
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
	if verboseFlag {
		if cmd {
			fmt.Printf(" cmdStr = %q, cmdPath = %s and will do execCmd.Run()\n", cmdStr, cmdPath)
		} else {
			fmt.Printf(" cmdStr = %q, cmdPath = %s and will do execCmd.Start()\n", cmdStr, cmdPath)
		}
	}
	if cmd {
		err = execCmd.Run() // will see if this works better when running cmd.exe, likely at work.
	} else {
		err = execCmd.Start()
	}
	if err != nil {
		fmt.Printf(" Error returned by running %s %s is %v\n", cmdStr, fileNameStr, err)
	}
} // end main
