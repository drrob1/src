package main // listVLC

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"src/list2"
	"src/misc"
	"src/whichexec"
	"strings"
	"time"
)

/*
REVISION HISTORY
======== =======
19 Jul 22 -- First version of launchVLC.  I'm writing this as I go along, pulling code from other pgms as I need them.
             I want this to take an input string on the command line.  This string will be a regexp used to match against a filename, like what rex.go does.
             From the resultant slice of matches of this regexp, I'll shuffle it and then feed them one at a time into vlc.
             So it looks like I'll need pieces of rex.go, shuffle code from bj.go, and then launching and external pgm code like I do in a few places now.
             The final launching loop will pause and exit if I want it to, like I did w/ the pid and windows title matching routines.  I'll let the import list auto-populate.
20 Jul 22 -- Added verboseFlag being set will have it output the filename w/ each loop iteration.  And I added 'x' to the exit key behavior.
21 Jul 22 -- Now called lauv, it will output n files on the command line to vlc.  This way I can use 'n' from within vlc.
22 Jul 22 -- I can't get this to work by putting several filenames on the command line and it reading them all in.  Maybe I'll try redirection.
23 Jul 22 -- I finally figured out how to work w/ variadic params, after searching online.  An answer in stack overflow helped me a lot.  Now it works.
24 Jul 22 -- I allow n, or numNames, to be zero.  That means that there is no limit to what's passed into vlc.
30 Jul 22 -- Decided to always have it print out number of matches and shuffling time.
             Now called launchv, because the old name of lauv just wasn't working for me.  It was hard to type.
19 Sep 22 -- Trying to get it so I don't need tcc.  I added the notccFlag to test it.  When that started working, I made the default of true, so it's always on.
               I had to add vlc to the path for it to work.
20 Sep 22 -- After writing and debugging findme, I now know that my issue all along was that actual quote characters were in the VLCPATH environment string.
               After removing those, the code now works as originally intended.
21 Sep 22 -- The code now has a default value for the location of C:\Program Files\VideoLAN\VLC.  The environment value VLCPATH can be used to change this.
               The code will exit if the find function returns a blank string.  This would be an opportunity to use the environment var to help the pgm find vlc.
               And since linux and windows use different characters as subdir separators in the PATH, and the filesystem uses a different delimiter, I have to use a conditional
               based on runtime.GOOS.
23 Oct 22 -- On linux will call cvlc instead of vlc.  Nevermind, it's not better.  But I changed MyReadDir to skip directories.
14 Nov 22 -- Will use fact that an empty regexp always matches everything.  Turned out to be a bad thing, because therefore the exclude expression excluded everything.
               I undid it.
------------------------------------------------------------------------------------------------------------------------------------------------------
16 Jan 23 -- Now called listvlc.go in list2 tree.  It will use the list routines to make a list to shuffle and then include in the vlc call.
18 Jan 23 -- Adding smartCase
18 Feb 23 -- Added init() which accounts for change of behavior in rand.Seed() starting w/ Go 1.20.
21 Apr 23 -- Making spell checker happy, and removing dead code.
 9 Feb 24 -- Using Go 1.22 code for random numbers.
25 Mar 24 -- Changed last veryVerboseFlag to verboseFlag, as I forgot that I needed very verbose to see the variadic param for the vlc command.
29 Apr 24 -- I finished writing whichExec, based on code from Mastering Go, 4th ed.  I'm looking at adding it here.  While I'm here, I'm editing some comments and help text.
               Much of the setup code to find vlc is unnecessary now that I have my own whichExec.
17 May 24 -- On linux, passing "2>/dev/null" to suppress the garbage error messages.
18 May 24 -- Oops.  I removed what I added yesterday, and instead removed the assignment of os.Stderr.  That will also supporess the many error messages from being seen.
19 May 24 -- Before I made changes, this routine put the filenames on the command line, got truncated on Windows.  I changed that to create an xspf file, like lv2.
               The API for the routine that writes the new xspf file was changed.
22 May 24 -- Updated displayed messages.
 5 Aug 24 -- I'm going to add a regexp to match, like what I did for runlst.  Nevermind, it's already there, but I never documented it so I forgot.
*/

const lastModified = "Aug 6, 2024"

const extDefault = ".xspf" // XML Sharable Playlist Format
const outPattern = "vlc_"

var includeRegex, excludeRegex *regexp.Regexp
var verboseFlag, veryVerboseFlag, noTccFlag, ok bool

var includeRexString, excludeRexString, vPath string
var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
var numNames int

func main() {
	fmt.Printf(" listvlc.go.  Last modified %s, compiled w/ %s\n\n", lastModified, runtime.Version())

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	vPath, ok = os.LookupEnv("VLCPATH")
	if ok {
		vlcPath = strings.ReplaceAll(vPath, `"`, "") // Here I use back quotes to delete a literal quote.  And replace the default value of vlcPath defined globally.
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: listvlc <options> <input-regex> --> where <input-regex> may be empty. \n")
		fmt.Fprintf(flag.CommandLine.Output(), " This pgm will make a list of matching filenames in the current directory, supporting SmartCase in the regexp param,\n")
		fmt.Fprintf(flag.CommandLine.Output(), " shuffle them, and then write them to xspf file which is then fed to vlc on the command line.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s, working directory is %s, full name of executable is %s and vlcPath is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName, vlcPath)
		fmt.Fprintf(flag.CommandLine.Output(), " It checks environment variable VLCPATH to use instead of default path to VLC. \n")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode flag.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, " Very Verbose mode flag.")
	flag.StringVar(&excludeRexString, "x", "", " Exclude file regexp string, which is usually empty.")
	flag.IntVar(&numNames, "n", 50, " Number of file names to output on the commandline to vlc.  Now Ignored.")
	flag.BoolVar(&noTccFlag, "not", true, " Not using tcc flag.") // Since the default is true, to make it false requires -not=false syntax.

	var revFlag bool
	flag.BoolVar(&revFlag, "r", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var sizeFlag bool
	flag.BoolVar(&sizeFlag, "s", false, "sort by size instead of by date")

	var filterFlag, noFilterFlag bool
	var filterStr string
	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVar(&filterFlag, "f", false, "filter value to suppress listing individual size below 1 MB.")
	flag.BoolVar(&noFilterFlag, "F", false, "Flag to undo an environment var with f set.")

	flag.Parse()
	if veryVerboseFlag { // very verbose also turns on verbose flag.
		verboseFlag = true
	}

	if verboseFlag {
		fmt.Printf(" %s has timestamp of %s, working directory is %s, and full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
	}
	if verboseFlag {
		fmt.Printf(" vlcPath = %s \n", vlcPath)
		//listPath(searchPath)
	}

	includeRexString = flag.Arg(0) // this is the first argument on the command line that is not the program name.
	var err error
	smartCaseRegex := regexp.MustCompile("[A-Z]")
	smartCaseFlag := smartCaseRegex.MatchString(includeRexString)
	if smartCaseFlag {
		includeRegex, err = regexp.Compile(includeRexString) // an empty regex compiles and will include everything.
	} else {
		includeRegex, err = regexp.Compile(strings.ToLower(includeRexString)) // an empty regex compiles and will include everything.
	}
	if err != nil {
		fmt.Printf(" Error from compiling the regexp input string is %v\n", err)
		os.Exit(1)
	}
	if excludeRexString != "" {
		excludeRegex, err = regexp.Compile(strings.ToLower(excludeRexString))
		if err != nil {
			fmt.Printf(" Error from compiling the exclude regexp is %v\n", err)
			os.Exit(1)
		}
	} else { // predefined regexp to exclude xspf files
		excludeRegex = regexp.MustCompile("xspf$") // must compile panics if the expression fails to compile.  It's easier for me this way.
	}

	list2.VerboseFlag = verboseFlag
	list2.VeryVerboseFlag = veryVerboseFlag
	list2.ReverseFlag = revFlag
	list2.SizeFlag = sizeFlag
	list2.ExcludeRex = excludeRegex
	list2.IncludeRex = includeRegex
	list2.SmartCaseFlag = smartCaseFlag

	// Finished processing the input flags and assigned list2 variables.  Now can get the fileList.

	fileList, err := list2.New() // fileList used to be []string, but now it's []FileInfoExType.
	if err != nil {
		fmt.Fprintf(os.Stderr, " ERROR from list2 is: %s\n", err)
		os.Exit(1)
	}

	fileList, err = list2.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, " ERROR from list2 is: %s\n", err)
		os.Exit(1)
	}

	//fileNames := getFileNames(workingDir, includeRegex) // not used here, as the filenames are retrieved by the list2 package.

	fileNames := make([]string, 0, len(fileList))
	for _, f := range fileList {
		fileNames = append(fileNames, f.FI.Name())
	}

	if verboseFlag {
		fmt.Printf(" There are %d filenames found using includeRexString = %q and %q, and excludeRexString = %q\n",
			len(fileNames), includeRexString, includeRegex.String(), excludeRexString)
	}

	if len(fileNames) == 0 {
		fmt.Printf(" No filenames matched the regexp of %q and were excluded by %q.  Exiting  \n", includeRexString, excludeRexString)
		os.Exit(0)
	}

	// Now to shuffle the file names slice.

	now := time.Now()
	shuffleAmount := now.Nanosecond()/1e6 + now.Second() + now.Minute() + now.Day() + now.Hour() + now.Year()
	more := misc.RandRange(50_000, 100_000)
	sumShuffle := shuffleAmount + more
	swapFnt := func(i, j int) {
		fileNames[i], fileNames[j] = fileNames[j], fileNames[i]
	}
	fmt.Printf(" ShuffleAmount = %d, more = %d, sumShuffle = %d for %d files.  About to start the Shuffle.\n\n", shuffleAmount, more, sumShuffle, len(fileNames))
	for i := 0; i < sumShuffle; i++ {
		rand.Shuffle(len(fileNames), swapFnt)
	}

	fmt.Printf(" Shuffled %d filenames %d times, which took %s.\n", len(fileNames), sumShuffle, time.Since(now))

	// Create an xspf file

	outFile, err := os.CreateTemp(workingDir, outPattern) // outPattern is likely still vlc_
	if err != nil {
		fmt.Printf(" Tried to createTemp xspf file but got ERROR: %s.  Bye-bye.\n", err)
		os.Exit(1)
	}
	defer outFile.Close()
	defer outFile.Sync()

	err = writeOutputFile(outFile, fileNames)
	if err != nil {
		fmt.Printf(" ERROR from writing output file %s: %s\n", outFile.Name(), err)
	}

	err = outFile.Sync()
	if err != nil {
		fmt.Printf(" Sync-ing %s failed w/ ERROR: %s.  Bye-bye.\n", outFile.Name(), err)
		return
	}

	err = outFile.Close()
	if err != nil {
		fmt.Printf(" Closing %s failed w/ ERROR: %s.  Bye-bye.\n", outFile.Name(), err)
		return
	}

	newFilename := outFile.Name() + extDefault
	err = os.Rename(outFile.Name(), newFilename)
	if err != nil {
		fmt.Printf(" Rename to %s failed w/ ERROR: %s.  Bye.\n", extDefault, err)
		return
	}

	fullOutFilename, err := filepath.Abs(newFilename)
	if err != nil {
		fmt.Printf(" filepath.Abs(%s) = ERROR is %s.  Exiting\n", newFilename, err)
		return
	}
	if verboseFlag {
		fmt.Printf(" Will write output to %s\n", fullOutFilename)
	}

	// ready to start calling vlc

	// Turns out that the shell searches against the path on Windows, but just executing it here doesn't.  So I have to search the path myself.
	// Nope, I still have that wrong.  I need to start a command processor, too.  And vlc is not in the %PATH, but it does work when I just give it as a command without a path.

	var vlcStr, shellStr string
	if runtime.GOOS == "windows" {
		vlcStr = whichexec.Find("vlc", vlcPath) // My own whichExec.Find adds vlcPath to the system path automatically.  I don't have to build search path up anymore.
	} else if runtime.GOOS == "linux" {
		vlcStr = whichexec.Find("vlc", "") // calling vlc without a console.
		shellStr = "/bin/bash"             // not needed as I found out by some experimentation on leox.
	}

	if vlcStr == "" {
		fmt.Printf(" vlcStr is null.  Exiting ")
		return
	}
	if verboseFlag {
		fmt.Printf(" vlcStr is %s\n", vlcStr)
	}

	// Time to run vlc.

	var execCmd *exec.Cmd

	if runtime.GOOS == "windows" {
		if noTccFlag {
			execCmd = exec.Command(vlcStr, fullOutFilename) // instead of appending filenames to command line, now read in the xspf file.
			//execCmd = exec.Command(vlcStr, variadicParam...)
		} else { // this isn't needed anymore.  I'll leave it here because it does work, in case I ever need to do this again.
			execCmd = exec.Command(shellStr, fullOutFilename)
			//execCmd = exec.Command(shellStr, variadicParam...)
		}
	} else if runtime.GOOS == "linux" {
		execCmd = exec.Command(vlcStr, fullOutFilename)
		//execCmd = exec.Command(vlcStr, fileNames...)
	}

	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	//execCmd.Stderr = os.Stderr  commenting this out is enough to suppress the many error messages.
	e := execCmd.Start()
	if e != nil {
		fmt.Printf(" Error returned by running vlc %s is %v\n", fullOutFilename, e)
	}
} // end main()

// ------------------------------- writeOutputFile --------------------------------

func writeOutputFile(w io.Writer, fn []string) error {
	const header1 = `<?xml version="1.0" encoding="UTF-8"?>
<playlist xmlns="http://xspf.org/ns/0/" xmlns:vlc="http://www.videolan.org/vlc/playlist/ns/0/" version="1">
`
	const trackListOpen = "<trackList>"
	const trackListClose = "</trackList>"
	const trackOpen = "<track>"
	const trackClose = "</track>"
	const locationOpen = "<location>file:///"
	const locationClose = "</location>"
	const extensionApplication = "<extension application=\"http://www.videolan.org/vlc/playlist/0\">"
	const extensionClose = "</extension>"
	const vlcIDOpen = "<vlc:id>"
	const vlcIDClose = "</vlc:id>"
	const playListClose = "</playlist>"

	buf := bufio.NewWriter(w)
	defer buf.Flush()

	buf.WriteString(header1) // this includes a lineTerm

	//	w.WriteRune('\t')    don't need the title
	//s := fmt.Sprintf("%s%s%s\n", titleOpen, "Playlist", titleClose)
	//w.WriteString(s)

	buf.WriteRune('\t')
	buf.WriteString(trackListOpen)
	buf.WriteRune('\n')

	for i, f := range fn {
		fullName, err := filepath.Abs(f)
		if err != nil {
			//fmt.Printf(" filepath.Abs(%s) returned ERROR: %s.  Bye-Bye.\n", f, err)
			return err
		}

		fullName = strings.ReplaceAll(fullName, "\\", "/") // Change backslash to forward slash, if that makes a difference.

		s2 := fmt.Sprintf("\t\t%s\n", trackOpen)
		buf.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t%s%s%s\n", locationOpen, fullName, locationClose)
		buf.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t%s\n", extensionApplication)
		buf.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t\t%s%d%s\n", vlcIDOpen, i, vlcIDClose)
		buf.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t%s\n", extensionClose)
		buf.WriteString(s2)

		s2 = fmt.Sprintf("\t\t%s\n", trackClose)
		_, err = buf.WriteString(s2)
		if err != nil {
			fmt.Printf(" Buffered write on track %d returnned ERROR: %s", i, err)
			return err
		}
	}

	buf.WriteRune('\t')
	buf.WriteString(trackListClose)
	buf.WriteRune('\n')

	buf.WriteString(playListClose)
	_, err := buf.WriteRune('\n')
	if err != nil {
		return err // Flush() not called as there's no point after writing returned an error.  Defer will try anyway, I think.
	}
	err = buf.Flush()
	return err
}
