package main // launchv.go

import (
	"flag"
	"fmt"
	"github.com/jonhadfield/findexec"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
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
20 Sep 22 -- After writing and debugging findme, I now know that my issue all along was that actual quote characters were in the VLCPATH environment string.
               After removing those, the code now works as originally intended.
21 Sep 22 -- The code now has a default value for the location of C:\Program Files\VideoLAN\VLC.  The environment value VLCPATH can be used to change this.
               The code will exit if the find function returns a blank string.  This would be an opportunity to use the environment var to help the pgm find vlc.
               And since linux and windows use different characters as subdir separators in the PATH, and the filesystem uses a different delimiter, I have to use a conditional
               based on runtime.GOOS.
23 Oct 22 -- On linux will call cvlc instead of vlc.  Nevermind, it's not better.  But I changed MyReadDir to skip directories.
14 Nov 22 -- Will use fact that an empty regexp always matches everything.  Turned out to be a bad thing, because therefore the exclude expression excluded everything.
               I undid it.
18 Jan 23 -- Adding smartCase
16 Feb 23 -- Added init() which accounts for change of behavior in rand.Seed() starting w/ Go 1.20.
26 Oct 23 -- Added hard coded regexp's.  And increased the default value for numNames.
*/

const lastModified = "Oct 26, 2023"

var includeRegex, excludeRegex *regexp.Regexp
var verboseFlag, veryverboseFlag, notccFlag, ok, smartCaseFlag bool
var includeRexString, excludeRexString, searchPath, path, vPath string
var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
var numNames int

//var preDefinedRegexp []string

func init() {
	goVersion := runtime.Version()
	goVersion = goVersion[4:6] // this should be a string of characters 4 and 5, or the numerical digits after Go1.  At the time of writing this, it will be 20.
	goVersionInt, err := strconv.Atoi(goVersion)
	if err == nil {
		fmt.Printf(" Go 1 version is %d\n", goVersionInt)
		if goVersionInt >= 20 { // starting w/ go1.20, rand.Seed() is deprecated.  It will auto-seed if I don't call it, and it wants to do that itself.
			return
		}
	} else {
		fmt.Printf(" ERROR from Atoi: %s\n", err)
	}
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var preBoolOne, preBoolTwo, domFlag, fuckFlag, numericFlag bool
	fmt.Printf(" %s last modified %s, compiled w/ %s\n\n", os.Args[0], lastModified, runtime.Version())

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	path = os.Getenv("PATH")
	vPath, ok = os.LookupEnv("VLCPATH")
	if ok {
		vlcPath = strings.ReplaceAll(vPath, `"`, "") // Here I use back quotes to insert a literal quote.
	}
	if runtime.GOOS == "windows" {
		searchPath = vlcPath + ";" + path
	} else if runtime.GOOS == "linux" && ok {
		searchPath = vlcPath + ":" + path
	} else { // on linux and not ok, meaning environment variable VLCPATH is empty.
		searchPath = path
	}
	preDefinedRegexp := []string{
		"femdom|tntu",
		"fuck.*dung|tiefuck|fuck.*bound|bound.*fuck|susp.*fuck|fuck.*susp|sexually|sas",
		"\b+\b",
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " This pgm will match an input regexp using smart case, against all filenames in the current directory\n")
		fmt.Fprintf(flag.CommandLine.Output(), " shuffle them, and then output 'n' of them on the command line to vlc.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s, working directory is %s, full name of executable is %s and vlcPath is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName, vlcPath)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: launchv <options> <input-regex> where <input-regex> cannot be empty. \n")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode flag.")
	flag.BoolVar(&veryverboseFlag, "vv", false, " Very Verbose mode flag.")
	flag.StringVar(&excludeRexString, "x", "", " Exclude file regexp string, which is usually empty.")
	flag.IntVar(&numNames, "n", 50, " Number of file names to output on the commandline to vlc.")
	flag.BoolVar(&notccFlag, "not", true, " Not using tcc flag.") // Since the default is true, to make it false requires -not=false syntax.
	flag.BoolVar(&preBoolOne, "1", false, "Use 1st predefined pattern of femdon|tntu")
	flag.BoolVar(&preBoolTwo, "2", false, "Use 2nd predefined pattern of fuck.*dung|tiefuck|fuck.*bound|bound.*fuck|susp.*fuck|fuck.*susp|sexually|sas")
	flag.BoolVar(&domFlag, "dom", false, "Use predefined pattern #1.")
	flag.BoolVar(&fuckFlag, "fuck", false, "Use predifined pattern #2.")
	flag.BoolVar(&numericFlag, "numeric", false, "Use predefined pattern ^\b+\b.")
	flag.Parse()

	if veryverboseFlag { // very verbose also turns on verbose flag.
		verboseFlag = true
	}

	if verboseFlag {
		fmt.Printf(" %s has timestamp of %s, working directory is %s, and full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
	}
	if verboseFlag {
		fmt.Printf(" vlcPath = %s, searchPath is: \n", vlcPath)
		listPath(searchPath)
	}

	if flag.NArg() < 1 && !preBoolOne && !preBoolTwo && !domFlag && !fuckFlag && !numericFlag { // if there are more than 1 arguments, the extra ones are ignored.
		fmt.Printf(" Usage: launchv <options> <input-regex> where <input-regex> cannot be empty.  Exiting\n")
		os.Exit(0)
	}

	if preBoolOne || domFlag {
		includeRexString = preDefinedRegexp[0]
	} else if preBoolTwo || fuckFlag {
		includeRexString = preDefinedRegexp[1]
	} else {
		includeRexString = flag.Arg(0) // this is the first argument on the command line that is not the program name.
	}

	var err error
	smartCase := regexp.MustCompile("[A-Z]")
	smartCaseFlag = smartCase.MatchString(includeRexString)
	if smartCaseFlag {
		includeRegex, err = regexp.Compile(includeRexString)
	} else {
		includeRegex, err = regexp.Compile(strings.ToLower(includeRexString))
	}
	if err != nil {
		fmt.Printf(" Error from compiling the regexp input string is %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(" using regular expression of: %s, %s\n", includeRexString, includeRegex.String())
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
	//                  rand.Seed(now.UnixNano())  Now handled by the init() function that knows Go 1.20+ doesn't want Seed called.
	shuffleAmount := now.Nanosecond()/1e6 + now.Second() + now.Minute() + now.Day() + now.Hour() + now.Year()
	swapFnt := func(i, j int) {
		fileNames[i], fileNames[j] = fileNames[j], fileNames[i]
	}
	for i := 0; i < shuffleAmount; i++ {
		rand.Shuffle(len(fileNames), swapFnt)
	}

	fmt.Printf(" Shuffled %d filenames %d times, which took %s.\n", len(fileNames), shuffleAmount, time.Since(now))

	// ready to start calling vlc

	// Turns out that the shell searches against the path on Windows, but just executing it here doesn't.  So I have to search the path myself.
	// Nope, I still have that wrong.  I need to start a command processor, too.  And vlc is not in the %PATH, but it does work when I just give it as a command without a path.

	var vlcStr, shellStr string
	if runtime.GOOS == "windows" {
		vlcStr = findexec.Find("vlc", searchPath) //Turns out that vlc was not in the path.  But it shows up when I use "which vlc".  So it seems that findexec doesn't find it on my win10 system.  So I added it to the path.
		//vlcStr = "vlc"
		//shellStr = os.Getenv("ComSpec") not needed anymore
	} else if runtime.GOOS == "linux" {
		vlcStr = findexec.Find("vlc", "") // calling vlc without a console.
		shellStr = "/bin/bash"            // not needed as I found out by some experimentation on leox.
	}

	if vlcStr == "" {
		fmt.Printf(" vlcStr is null.  Exiting ")
		os.Exit(1)
	}

	// Time to run vlc.

	var execCmd *exec.Cmd

	variadicParam := []string{"-C", "vlc"} // This isn't really needed anymore.  I'll leave it here anyway, as a model in case I ever need to do this again.
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
		} else { // this isn't needed anymore.  I'll leave it here because it does work, in case I ever need to do this again.
			execCmd = exec.Command(shellStr, variadicParam...)
		}
		//switch n { // just to see if this works.  Once I figured out the variadic syntax I don't need a switch case statement here.
		//case 1:
		//	execCmd = exec.Command(shellStr, "-C", vlcStr, fileNames[0])
		//case 2:
		//	execCmd = exec.Command(shellStr, "-C", vlcStr, fileNames[0], fileNames[1])
		//case 3:
		//	execCmd = exec.Command(shellStr, "-C", vlcStr, fileNames[0], fileNames[1], fileNames[2])
		//case 4:
		//	execCmd = exec.Command(shellStr, "-C", vlcStr, fileNames[0], fileNames[1], fileNames[2], fileNames[3])
		//case 5:
		//	execCmd = exec.Command(shellStr, "-C", vlcStr, fileNames[0], fileNames[1], fileNames[2], fileNames[3], fileNames[4])
		//default:
		//	execCmd = exec.Command(shellStr, variadicParam...)
		//}
	} else if runtime.GOOS == "linux" { // I'm ignoring this for now.  I'll come back to it after I get the Windows code working.
		execCmd = exec.Command(vlcStr, fileNames...)
	}

	if veryverboseFlag {
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

	if veryverboseFlag {
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
		maybeLower := d.Name()
		if !smartCaseFlag {
			maybeLower = strings.ToLower(maybeLower)
		}
		if !inputRegex.MatchString(maybeLower) { // skip dirEntries that do not match the input regex.
			continue
		}
		if d.IsDir() { // skip directories
			continue
		}

		if excludeRegex == nil {
			fileNames = append(fileNames, d.Name())
		} else if !excludeRegex.MatchString(strings.ToLower(d.Name())) { // excludeRegex is not empty, so using it won't panic.  And always use ToLower here.
			fileNames = append(fileNames, d.Name())
		}
	}
	return fileNames
} // myReadDir

// ------------------------------- minInt ----------------------------------------

func minInt(i, j int) int {
	if i <= j {
		return i
	}
	return j
}

// ------------------------------- listPath --------------------------------------

func listPath(path string) {
	splitEnv := strings.Split(path, ";")
	for _, s := range splitEnv {
		fmt.Printf(" %s\n", s)
	}
}

/* ------------------------------------------- MakeDateStr ---------------------------------------------------* */
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

// ------------------------------ pause -----------------------------------------

func pause() bool {
	fmt.Print(" Pausing the loop.  Hit <enter> to continue; 'n' or 'x' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.HasPrefix(ans, "n") || strings.HasPrefix(ans, "x") {
		return true
	}
	return false
}
*/
