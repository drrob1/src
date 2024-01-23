package main // lv2.go from launchV.go

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jonhadfield/findexec"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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
15 Nov 23 -- Added another hard coded regexp.
 6 Dec 23 -- Fixed the numeric pattern, and added alternate num option.
13 Dec 23 -- Really fixed the numeric pattern.
14 Dec 23 -- Going to try setting the StdErr to /dev/null and see what happens.  I don't want to see all the errors that show up on linux.
               No, that didn't work.  But I can not assign it to anything, and that works.
20 Jan 24 -- Adding femdom as a switch
21 Jan 24 -- Now called lv2, for launchvlc 2.  By trial and error, I discovered that the input file to vlc must have an .xspf extension, but it does not need the duration line.
               So now I want to write all the files matched w/ the regexp to temp .xspf file.  Since I don't need the duration line, I can do this now.  I'll hardcode the html stuff
               that I need to use for each file.  It must have the <location> line else it won't work.  But it may be faster to parse if I give it everything I have but for the duration.
               I expect this will take a bit of time.
               Header1 will be lines 1 and 2.  Title line will have the regexp as the title btwn the html tags.
               trackList opening is once per file.
               track has location line, and then the extension application stuff needs a counter starting from 0.  <track> <location> <extension application> </extension> </track>
               </trackList><playlist>
               So the input code is essentially the same as launchV.go.  It's the output code that has to be fairly different from launchV.go.
               First the pgm builds the filenames slice, then shuffles it, then writes the output file, and finally calls vlc w/ the .xspf file on the command line.
22 Jan 24 -- It works.  I got stuck for at least 2 hrs in that the file was created but vlc would not use it.  Then I finally saw the error, there was an extra space at the beginning of
               the first line.  When I took that out, it started working.  But I had already fixed some of the strings so that all had a closing angle bracket.  I left a few off at first.
*/

/*
<?xml version="1.0" encoding="UTF-8"?>
<playlist xmlns="http://xspf.org/ns/0/" xmlns:vlc="http://www.videolan.org/vlc/playlist/ns/0/" version="1">
	<title>Playlist</title>
	<trackList>
		<track>
			<location>file:///E:/Movie/Wooden-Horse-Bondage-Vibrator-Vol02-1_54_59-Pornhub.mp4</location>
			<extension application="http://www.videolan.org/vlc/playlist/0">
				<vlc:id>0</vlc:id>
			</extension>
		</track>
		<track>
			<location>file:///E:/Movie/asian-orgasm-in-box-Pornhub.com.mp4</location>
			<extension application="http://www.videolan.org/vlc/playlist/0">
				<vlc:id>1</vlc:id>
			</extension>
		</track>
        <track>
            ...
        </track>
	</trackList>
</playlist>
*/

const lastModified = "Jan 23, 2024"

const header1 = `<?xml version="1.0" encoding="UTF-8"?>
<playlist xmlns="http://xspf.org/ns/0/" xmlns:vlc="http://www.videolan.org/vlc/playlist/ns/0/" version="1">
`
const titleOpen = "<title>"
const titleClose = "</title>"
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
const extDefault = ".xspf" // XML Sharable Playlist Format

var includeRegex, excludeRegex *regexp.Regexp
var verboseFlag, veryverboseFlag, notccFlag, ok, smartCaseFlag bool
var includeRexString, excludeRexString, searchPath, path, vPath string
var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
var numNames int

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
		fmt.Printf(" ERROR from Atoi: %s.  Calling rand.Seed(time.Now().UnixNano())\n", err)
	}
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var preBoolOne, preBoolTwo, domFlag, fuckFlag, numericFlag, vibeFlag, spandexFlag, femdomFlag bool
	fmt.Printf(" %s last modified %s, compiled w/ %s\n\n", os.Args[0], lastModified, runtime.Version())

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" Call to Getwd failed w/ ERROR: %s.  Bye-Bye.\n", err)
		os.Exit(1)
	}
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
		"fuck.*dung|tiefuck|fuck.*bound|bound.*fuck|susp.*fuck|fuck.*susp|sexually|sas|fit18",
		//                                                                                  "^\b+\b",  This doesn't work
		"^[0-9]+[0-9]",
		"wmbcv|^tbc|^fiterotic|^bjv|hardtied|vib|ethnick|chair|orgasmabuse",
		"spandex|camel|yoga|miamix|^AMG|^sporty|balle|dancerb",
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
	flag.BoolVar(&fuckFlag, "fuck", false, "Use predefined pattern #2.")
	flag.BoolVar(&numericFlag, "numeric", false, "Use predefined pattern ^[0-9]+[0-9]")
	numFlag := flag.Bool("num", false, "Alternate for numeric.")
	flag.BoolVar(&vibeFlag, "vibe", false, "Use predefined pattern for vibrating.")
	flag.BoolVar(&spandexFlag, "spandex", false, "Use spandex predefined pattern")
	flag.BoolVar(&femdomFlag, "femdom", false, "Use predefined pattern for femdom")
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
		fmt.Printf(" Usage: launchv <options> <input-regex> where <input-regex> may not be empty.\n")
	}

	// Process predefined regular expressions.
	if preBoolOne || domFlag || femdomFlag {
		includeRexString = preDefinedRegexp[0]
	} else if preBoolTwo || fuckFlag {
		includeRexString = preDefinedRegexp[1]
	} else if numericFlag || *numFlag {
		includeRexString = preDefinedRegexp[2]
	} else if vibeFlag {
		includeRexString = preDefinedRegexp[3]
	} else if spandexFlag {
		includeRexString = preDefinedRegexp[4]
	} else {
		includeRexString = flag.Arg(0) // this is the first argument on the command line that is not the program name.
	}

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
	fmt.Printf(" using regular expression of: %s\n", includeRegex.String())
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

	// The file names slice is ready.  Now to create the output file.  Part of the filename will be the regexp used to create this file.

	regexpStr := includeRegex.String()
	rplcPattern := strings.NewReplacer("=", "", "+", "", ".", "", "?", "", "*", "", "|", "_", " ", "", "[", "", "]", "", "^", "", "$", "")
	replacedStr := rplcPattern.Replace(regexpStr)
	outFilename := "vlc" + "_" + replacedStr + "_" + strconv.Itoa(len(fileNames)) + extDefault
	outputFile, err := os.Create(outFilename)
	if err != nil {
		fmt.Printf(" Tried to create %s but got ERROR: %s.  Bye-bye.\n", outFilename, err)
		os.Exit(1)
	}
	defer outputFile.Close()
	defer outputFile.Sync()

	fullOutFilename, err := filepath.Abs(outputFile.Name())
	if err != nil {
		fmt.Printf(" filepath.Abs(%s) = ERROR is %s.  Exiting\n", outFilename, err)
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf(" Output filename is %s, workingDir is %s, abs() is %s \n", outFilename, workingDir, fullOutFilename)
	}
	outfileBuf := bufio.NewWriter(outputFile)
	defer outfileBuf.Flush()

	// Now to write out the xspf file

	err = writeOutputFile(outfileBuf, fileNames)
	if err != nil {
		fmt.Printf(" Writing output file %s failed w/ ERROR: %s.  Bye-bye.\n", outFilename, err)
		os.Exit(1)
	}

	err = outfileBuf.Flush()
	if err != nil {
		fmt.Printf(" Flushing %s buffer failed w/ ERROR: %s.  Bye-bye.\n", outFilename, err)
		os.Exit(1)
	}

	err = outputFile.Sync()
	if err != nil {
		fmt.Printf(" Sync-ing %s failed w/ ERROR: %s.  Bye-bye.\n", outFilename, err)
		os.Exit(1)
	}

	err = outputFile.Close()
	if err != nil {
		fmt.Printf(" Closing %s failed w/ ERROR: %s.  Bye-bye.\n", outFilename, err)
		os.Exit(1)
	}

	fmt.Printf("")

	// ready to start calling vlc

	// Turns out that the shell searches against the path on Windows, but just executing it here doesn't.  So I have to search the path myself.
	// Nope, I still have that wrong.  I need to start a command processor, too.  And vlc is not in the %PATH, but it does work when I just give it as a command without a path.

	var vlcStr, shellStr string
	if runtime.GOOS == "windows" {
		vlcStr = findexec.Find("vlc", searchPath) //Turns out that vlc was not in the path.  But it shows up when I use "which vlc".  So it seems that findexec doesn't find it on my win10 system.  So I added it to the path.
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
	//variadicParam = append(variadicParam, fileNames...)
	variadicParam = append(variadicParam, fullOutFilename)

	// For me to be able to pass a variadic param here, I must match the definition of the function, not pass some and then try the variadic syntax.
	// I got this answer from stack overflow.

	if runtime.GOOS == "windows" {
		if notccFlag {
			execCmd = exec.Command(vlcStr, variadicParam...)
		} else { // this isn't needed anymore.  I'll leave it here because it does work, in case I ever need to do this again.
			execCmd = exec.Command(shellStr, variadicParam...)
		}
	} else if runtime.GOOS == "linux" { // I'm ignoring this for now.  I'll come back to it after I get the Windows code working.
		execCmd = exec.Command(vlcStr, fullOutFilename)
	}

	if veryverboseFlag {
		fmt.Printf(" vlcStr = %s, len of variadicParam = %d, and filenames in variadicParam are %v\n", vlcStr, len(variadicParam), variadicParam)
	}

	execCmd.Stdin = os.Stdin
	//execCmd.Stdout = os.Stdout
	//execCmd.Stderr = os.Stderr //I don't have to assign this.  Let's see what happens if I leave it at nil.  It worked as I hoped.  No errors are displayed to the screen in linux.
	e := execCmd.Start()
	if e != nil {
		fmt.Printf(" Error returned by running vlc %s is %v\n", variadicParam, e)
	}
	fmt.Printf(" Full output file name is %s\n", fullOutFilename)
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

//func minInt(i, j int) int {
//	if i <= j {
//		return i
//	}
//	return j
//}

// ------------------------------- listPath --------------------------------------

func listPath(path string) {
	splitEnv := strings.Split(path, ";")
	for _, s := range splitEnv {
		fmt.Printf(" %s\n", s)
	}
}

// ------------------------------- writeOutputFile --------------------------------

func writeOutputFile(w *bufio.Writer, fn []string) error {
	w.WriteString(header1) // this includes a lineTerm

	w.WriteRune('\t')
	s := fmt.Sprintf("%s%s%s\n", titleOpen, "Playlist", titleClose)
	w.WriteString(s)

	w.WriteRune('\t')
	w.WriteString(trackListOpen)
	w.WriteRune('\n')

	for i, f := range fn {
		fullName, err := filepath.Abs(f)
		if err != nil {
			//fmt.Printf(" filepath.Abs(%s) returned ERROR: %s.  Bye-Bye.\n", f, err)
			return err
		}

		s2 := fmt.Sprintf("\t\t%s\n", trackOpen)
		w.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t%s%s%s\n", locationOpen, fullName, locationClose)
		w.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t%s\n", extensionApplication)
		w.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t\t%s%d%s\n", vlcIDOpen, i, vlcIDClose)
		w.WriteString(s2)

		s2 = fmt.Sprintf("\t\t\t%s\n", extensionClose)
		w.WriteString(s2)

		s2 = fmt.Sprintf("\t\t%s\n", trackClose)
		_, err = w.WriteString(s2)
		if err != nil {
			fmt.Printf(" Buffered write on track %d returnned ERROR: %s", i, err)
			return err
		}
	}

	w.WriteRune('\t')
	w.WriteString(trackListClose)
	w.WriteRune('\n')

	w.WriteString(playListClose)
	_, err := w.WriteRune('\n')
	return err
}

/*
func writeOutputFile(w *bufio.Writer, fn []string) {
	w.WriteString(header1) // this includes a lineTerm

	w.WriteRune('\t')
	w.WriteString(titleOpen)
	w.WriteString(includeRegex.String())
	if excludeRexString != "" {
		w.WriteRune('_')
		w.WriteString(excludeRegex.String())
	}
	w.WriteString(titleClose)
	w.WriteString(lineTerm)

	w.WriteRune('\t')
	w.WriteString(trackListOpen)
	w.WriteString(lineTerm)

	for i, f := range fn {
		fullName, err := filepath.Abs(f)
		if err != nil {
			fmt.Printf(" filepath.Abs(%s) returned ERROR: %s.  Bye-Bye.\n", f, err)
			os.Exit(1)
		}

		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteString(trackOpen)
		w.WriteString(lineTerm)

		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteString(locationOpen)
		w.WriteString(fullName)
		w.WriteString(locationClose)
		w.WriteString(lineTerm)

		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteString(extensionApplication)
		w.WriteString(lineTerm)

		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteString(vlcIDOpen)
		w.WriteString(strconv.Itoa(i))
		w.WriteString(vlcIDClose)
		w.WriteString(lineTerm)

		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteString(extensionClose)
		w.WriteString(lineTerm)

		w.WriteRune('\t')
		w.WriteRune('\t')
		w.WriteString(trackClose)
		w.WriteString(lineTerm)
	}

	w.WriteRune('\t')
	w.WriteString(trackListClose)
	w.WriteString(lineTerm)

	w.WriteString(playListClose)
	w.WriteString(lineTerm)
}


*/
