package main // launchVLC.go

import (
	"flag"
	"fmt"
	"github.com/jonhadfield/findexec"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

/*
   HISTORY
   =======
19 Jul 22 -- First version.  I'm writing this as I go along, pulling code from other pgms as I need them.
             I want this to take an input string on the command line.  This string will be a regexp used to match against a filename, like what rex.go does.
             From the resultant slice of matches of this regexp, I'll shuffle it and then feed them one at a time into vlc.
             So it looks like I'll need pieces of rex.go, shuffle code from bj.go, and then launching and external pgm code like I do in a few places now.
             The final launching loop will pause and exit if I want it to, like I did w/ the pid and windows title matching routines.  I'll let the import list auto-populate.
*/

const lastModified = "July 20, 2022"

var includeRegex, excludeRegex *regexp.Regexp
var verboseFlag, excludeStringEmpty bool
var includeRexString, excludeRexString string

func main() {
	fmt.Printf(" launch vlc.go.  Last modified %s, compiled w/ %s\n\n", lastModified, runtime.Version())

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " This pgm will match an input regexp against all filenames in the current directory\n")
		fmt.Fprintf(flag.CommandLine.Output(), " shuffle them, and then feed them one at a time into vlc.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: launchvlc <options> <input-regex> where <input-regex> cannot be empty. \n")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode flag.")
	flag.StringVar(&excludeRexString, "x", "", " Exclude file regexp string, which is usually empty.")
	flag.Parse()

	if verboseFlag {
		fmt.Printf(" %s has timestamp of %s, working directory is %s, and full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
	}

	if flag.NArg() < 1 { // if there are more than 1 arguments, the extra ones are ignored.
		fmt.Printf(" Usage: launchvlc <options> <input-regex> where <input-regex> cannot be empty.  Exiting\n")
		os.Exit(0)
	}

	includeRexString = flag.Arg(0) // this is the first argument on the command line that is not the program name.
	var err error
	includeRegex, err = regexp.Compile(strings.ToLower(includeRexString))
	if err != nil {
		fmt.Printf(" Error from compiling the regexp input string is %v\n", err)
		os.Exit(1)
	}
	if excludeRexString == "" {
		excludeStringEmpty = true
	} else {
		excludeRegex, err = regexp.Compile(strings.ToLower(excludeRexString))
		if err != nil {
			fmt.Printf(" Error from compiling the exclude regexp is %v\n", err)
			os.Exit(1)
		}
		excludeStringEmpty = false
	}

	fileNames := getFileNames(workingDir, includeRegex) // this slice of filenames matches the includeRegexp and does not match the excludeRegexp, if given.
	if verboseFlag {
		fmt.Printf(" There are %d filenames found using includeRexString = %q and %q, excludeStringNotEmpty = %t, and excludeRexString = %q\n",
			len(fileNames), includeRexString, includeRegex.String(), excludeStringEmpty, excludeRexString)
	}

	// Now to shuffle the file names slice.

	now := time.Now()
	rand.Seed(now.UnixNano())
	shuffleAmount := now.Nanosecond()/1e6 + now.Second() + now.Minute() + now.Day() + now.Hour() + now.Year()
	swapfnt := func(i, j int) {
		fileNames[i], fileNames[j] = fileNames[j], fileNames[i]
	}
	for i := 0; i < shuffleAmount; i++ {
		rand.Shuffle(len(fileNames), swapfnt)
	}
	if verboseFlag {
		fmt.Printf(" Shuffled %d filenames %d times, which took %s.\n", len(fileNames), shuffleAmount, time.Since(now))
	}

	// ready to start calling vlc

	if verboseFlag {
		fmt.Printf(" About to call vlc w/ each filename.\n")
	}

	if pause() {
		os.Exit(0)
	}

	// Turns out that the shell searches against the path on Windows, but just executing it here doesn't.  So I have to search the path myself.
	// Nope, I still have that wrong.  I need to start a command processor, too.

	var vlcStr, shellStr string
	if runtime.GOOS == "windows" {
		vlcStr = findexec.Find("vlc.exe", "")
		shellStr = os.Getenv("ComSpec")
	} else if runtime.GOOS == "linux" {
		vlcStr = findexec.Find("vlc", "")
		shellStr = "/bin/bash" // not needed as I found out by some experimentation on leox.
	}

	// Time to run vlc.  I didn't change the code yet for linux, but linux does not need to call bash first, and certainly not

	var execCmd *exec.Cmd
	for _, name := range fileNames {
		if runtime.GOOS == "windows" {
			execCmd = exec.Command(shellStr, "-C", vlcStr, name)
			//execCmd = exec.Command(shellStr, "-C", "vlc", name) // just to see if this works now.  It does, since I'm starting a tcc shell.
		} else if runtime.GOOS == "linux" {
			execCmd = exec.Command(vlcStr, name)
		}

		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		e := execCmd.Run()
		if e != nil {
			fmt.Printf(" Error returned by running vlc %s is %v\n", name, e)
		}
		if pause() {
			os.Exit(0)
		}
	}
} // end main()

// ------------------------------------------------------------------------ getFileNames -------------------------------------------------------

func getFileNames(workingDir string, inputRegex *regexp.Regexp) []string {

	fileNames := myReadDir(workingDir, inputRegex) // excluding by regex, filesize or having an ext is done by MyReadDir.

	if verboseFlag {
		fmt.Printf(" Leaving getFileInfosFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileNames)=%d\n", flag.NArg(), len(flag.Args()), len(fileNames))
	}

	return fileNames
}

// ------------------------------- myReadDir -----------------------------------

func myReadDir(dir string, inputRegex *regexp.Regexp) []string {

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	fileNames := make([]string, 0, len(dirEntries))
	for _, d := range dirEntries {
		lower := strings.ToLower(d.Name())
		if !inputRegex.MatchString(lower) { // skip dirEntries that do not match the input regex.
			continue
		} else if excludeStringEmpty {
			fileNames = append(fileNames, d.Name())
		} else { // excludeString is not empty, so must test against it
			if !excludeRegex.MatchString(lower) { // I have to guard against using an empty excludeRegex, or it will panic.
				fileNames = append(fileNames, d.Name())
			}
		}
	}
	return fileNames
} // myReadDir

// ------------------------------ pause -----------------------------------------

func pause() bool {
	fmt.Print(" Pausing.  Hit <enter> to continue.  Or 'n' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.HasPrefix(ans, "n") {
		return true
	}
	return false
}
