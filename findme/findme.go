package main // findme.go

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
REVISION HISTORY
======== =======
19 Jul 22 -- First version.  I'm writing this as I go along, pulling code from other pgms as I need them.
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
20 Sep 22 -- Now called findme.go.  I want to explore how searching thru the path, and adding to the path, works (or doesn't work).  I tried to add the dir for vlc here in an
               environment variable.  That didn't work.  Why?
*/

const lastModified = "Sep 21, 2022"

var includeRegex, excludeRegex *regexp.Regexp
var verboseFlag, notccFlag bool
var includeRexString, excludeRexString string
var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
var numNames int

func main() {
	fmt.Printf(" launch vlc.go.  Last modified %s, compiled w/ %s\n\n", lastModified, runtime.Version())

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " This pgm will match an input regexp against all filenames in the current directory\n")
		fmt.Fprintf(flag.CommandLine.Output(), " shuffle them, and then output 'n' of them on the command line to vlc.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: launchv <options> <input-regex> where <input-regex> cannot be empty. \n")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode flag.")
	flag.StringVar(&excludeRexString, "x", "", " Exclude file regexp string, which is usually empty.")
	flag.IntVar(&numNames, "n", 20, " Number of file names to output on the commandline to vlc.")
	flag.BoolVar(&notccFlag, "not", true, " Not using tcc flag.") // Since the default is true, to make it false requires -not=false syntax.
	flag.Parse()
	numNames += 2 // account for 2 extra items I have to add to the slice, ie, the -C and vlc add'l params.

	if verboseFlag {
		fmt.Printf(" %s has timestamp of %s, working directory is %s, and full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
	}
	/*
		if flag.NArg() < 1 { // if there are more than 1 arguments, the extra ones are ignored.
			fmt.Printf(" Usage: findme <options> <input-regex> where <input-regex> cannot be empty.  Exiting\n")
			os.Exit(0)
		}

	*/

	includeRexString = flag.Arg(0) // this is the first argument on the command line that is not the program name.
	var err error
	includeRegex, err = regexp.Compile(strings.ToLower(includeRexString))
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
	}

	fileNames := getFileNames(workingDir, includeRegex) // this slice of filenames matches the includeRegexp and does not match the excludeRegexp, if given.
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
	rand.Seed(now.UnixNano())
	shuffleAmount := now.Nanosecond()/1e6 + now.Second() + now.Minute() + now.Day() + now.Hour() + now.Year()
	swapFnt := func(i, j int) {
		fileNames[i], fileNames[j] = fileNames[j], fileNames[i]
	}
	for i := 0; i < shuffleAmount; i++ {
		rand.Shuffle(len(fileNames), swapFnt)
	}
	//if verboseFlag { // now always shows this message
	fmt.Printf(" Shuffled %d filenames %d times, which took %s.\n", len(fileNames), shuffleAmount, time.Since(now))
	//}

	// ready to start calling vlc

	// Turns out that the shell searches against the path on Windows, but just executing it here doesn't.  So I have to search the path myself.
	// Nope, I still have that wrong.  I need to start a command processor, too.  And vlc is not in the %PATH, but it does work when I just give it as a command without a path.

	pathStr := os.Getenv("PATH")
	listEnv(pathStr)
	if pause() {
		os.Exit(0)
	}

	vPath, ok := os.LookupEnv("VLCPATH")
	if ok {
		fmt.Printf(" vPath is %s\n", vPath)
	} else {
		fmt.Printf(" vPath is not found.\n")
	}
	vlcPath = strings.ReplaceAll(vPath, `"`, "") // Here I use back quotes to insert a literal quote.
	fmt.Printf(" vlcPath is %s after removing the quote characters, if present.\n", vlcPath)
	if pause() {
		os.Exit(0)
	}

	searchPath := vlcPath + ";" + pathStr
	listEnv(searchPath)
	if pause() {
		os.Exit(0)
	}

	var vlcStr, shellStr string
	if runtime.GOOS == "windows" {
		//vlcStr = findexec.Find("vlc.exe", searchPath) //Turns out that vlc was not in the path.  But it shows up when I use "which vlc".  This now works.
		vlcStr = findexec.Find("vlc", searchPath) //Turns out that vlc was not in the path.  But it shows up when I use "which vlc".  This also works.
		//vlcStr = filepath.Join(vlcPath, "vlc.exe") // Maybe this will work?  Yes, but I figured out that the vlcPath includes quotes, and that messed everything up.
		//vlcStr = "vlc"
		shellStr = os.Getenv("ComSpec")
	} else if runtime.GOOS == "linux" {
		vlcStr = findexec.Find("vlc", "") // I'm not changing this line to use searchPath.  I'll leave it as is to use the system $PATH
		shellStr = "/bin/bash"            // not needed as I found out by some experimentation on leox.
	}
	fmt.Printf(" vlcStr is %s \n", vlcStr)
	if pause() {
		os.Exit(1)
	}

	// Time to run vlc.

	var execCmd *exec.Cmd

	variadicParam := []string{"-C", "vlc"}
	if notccFlag {
		variadicParam = []string{}
	}
	variadicParam = append(variadicParam, fileNames...)
	n := minInt(numNames, len(fileNames))
	if n > 0 {
		variadicParam = variadicParam[:n]
	}

	// For me to be able to pass a variadic param here, I must match the definition of the function, not pass some and then try the variadic syntax.
	// I got this answer from stack overflow.

	if runtime.GOOS == "windows" {
		if notccFlag {
			execCmd = exec.Command(vlcStr, variadicParam...)
		} else {
			execCmd = exec.Command(shellStr, variadicParam...)
		}
	} else if runtime.GOOS == "linux" { // I'm ignoring this for now.  I'll come back to it after I get the Windows code working.
		execCmd = exec.Command(vlcStr, fileNames...)
	}

	if verboseFlag {
		fmt.Printf(" vlcStr = %s, len of variadicParam = %d, and filenames in variadicParam are %v\n", vlcStr, len(variadicParam), variadicParam)
	}

	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	e := execCmd.Start()
	if e != nil {
		fmt.Printf(" Error returned by running vlc %s is %v\n", variadicParam, e)
	}
} // end main()

// ------------------------------------------------------------------------ getFileNames -------------------------------------------------------

func getFileNames(workingDir string, inputRegex *regexp.Regexp) []string {

	fileNames := myReadDir(workingDir, inputRegex) // excluding by regex, filesize or having an ext is done by MyReadDir.

	if verboseFlag {
		fmt.Printf(" Leaving getFileNames.  flag.Nargs=%d, len(flag.Args)=%d, len(fileNames)=%d\n", flag.NArg(), len(flag.Args()), len(fileNames))
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
		}

		if excludeRegex == nil {
			fileNames = append(fileNames, d.Name())
		} else if !excludeRegex.MatchString(lower) { // excludeRegex is not empty, so using it won't panic.
			fileNames = append(fileNames, d.Name())
		}
	}
	return fileNames
} // myReadDir

// ------------------------------ pause -----------------------------------------

func pause() bool {
	fmt.Print(" Pausing.  Hit <enter> to continue; 'n' or 'x' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "x") || strings.HasPrefix(ans, "q") {
		return true
	}
	return false
}

// ------------------------------- minInt ----------------------------------------

func minInt(i, j int) int {
	if i <= j {
		return i
	}
	return j
}

// ------------------------------- listEnv --------------------------------------

func listEnv(env string) {
	splitEnv := strings.Split(env, ";")
	for _, s := range splitEnv {
		fmt.Printf(" %s\n", s)
	}
}

// ------------------------------------------- MakeDateStr ---------------------------------------------------
/*
func MakeDateStr() string {

	const DateSepChar = "-"
	var dateStr string

	m, d, y := timlibg.TIME2MDY()
	timeNow := timlibg.GetDateTime()

	MSTR := strconv.Itoa(m)
	DSTR := strconv.Itoa(d)
	YSTR := strconv.Itoa(y)
	Hr := strconv.Itoa(timeNow.Hours)
	Min := strconv.Itoa(timeNow.Minutes)
	Sec := strconv.Itoa(timeNow.Seconds)

	dateStr = "_" + MSTR + DateSepChar + DSTR + DateSepChar + YSTR + "_" + Hr + DateSepChar + Min + DateSepChar + Sec + "__" + timeNow.DayOfWeekStr
	return dateStr
} // MakeDateStr
*/
