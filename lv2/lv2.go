package main // lv2.go from launchV.go from vlcshuffle.go

import (
	"bufio"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"src/misc"
	"src/whichexec"
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
------------------------------------------------------------------------------------------------------------------------------------------------------
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
23 Jan 24 -- I'm going to remove code I haven't used in quite a while.
24 Jan 24 -- I'm having the default for the excludeRegex be xspf$.  And I'm adding maxNumOfTracks and numOfTracks.  Hey, it worked.  The patterns now all work.
                I'm now going to as a random number to the name of the file so similar regexp's don't clobber one other.  I read at xspf.org that <title>Playlist</title> is not required.
                So I deleted it.
26 Jan 24 -- Still trying to figure out what makes xspf files too large.  So I added output of total file string lengths.  And since my go.mod file says a min of Go 1.21 for compilation,
                I don't need the init() fcn here to possibly call rand.Seed().  I removed it.  And I'm colorizing the final output message.
                So far, the maximum string buffer for location strings is 13,937 < max 4 buffer < 13965.  Nope, that's not it.  See xspftotstrlen.go.
27 Jan 24 -- Found it.  Bad filenames containing characters that choked vlc.  In this case, !!! was the culprit.  When I removed those files, it worked.  Then I detoxed those files.
 5 Feb 24 -- Increased the shuffling number
 6 Feb 24 -- Added randRange.
 8 Feb 24 -- Added math/rand/v2, newly introduced w/ Go 1.22
30 Apr 24 -- Changed to use my own which exec code instead of someone else's code.
17 May 24 -- On linux, will pass "2>/dev/nul" to suppress garbage error messages.  Nevermind, not needed.  See comments.
20 May 24 -- Refactored the writeOutputFile routine.
22 May 24 -- Fixed some displayed messages.
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

const lastModified = "May 22, 2024"

const lineTooLong = 500    // essentially removing it
const maxNumOfTracks = 300 // I'm trying to track down why some xspf files work and others don't.  Found it, see comment above dated 27 Jan 24.

const extDefault = ".xspf" // XML Sharable Playlist Format

var includeRegex, excludeRegex *regexp.Regexp
var verboseFlag, veryVerboseFlag, ok, smartCaseFlag bool
var includeRexString, excludeRexString, vPath string
var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
var numOfTracks int

func main() {
	var preBoolOne, preBoolTwo, domFlag, fuckFlag, numericFlag, vibeFlag, spandexFlag, femdomFlag, forcedFlag bool
	fmt.Printf(" %s last modified %s, compiled w/ %s\n\n", os.Args[0], lastModified, runtime.Version())

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" Call to Getwd failed w/ ERROR: %s.  Bye-Bye.\n", err)
		os.Exit(1)
	}
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	vPath, ok = os.LookupEnv("VLCPATH")
	if ok {
		vlcPath = strings.ReplaceAll(vPath, `"`, "") // Here I use back quotes to insert a literal quote.
	}
	if !ok && runtime.GOOS == "linux" { // on a linux system vlcPath isn't needed, unless the VLCPATH environment variable is set by the user.
		vlcPath = ""
	}

	preDefinedRegexp := []string{
		"femdom|tntu",
		"fuck.*dung|tiefuck|fuck.*bound|bound.*fuck|susp.*fuck|fuck.*susp|sexually|sas|fit18",
		"^[0-9]+[0-9]",
		"wmbcv|^tbc|^fiterotic|^bjv|hardtied|vib|ethnick|chair|orgasmabuse",
		"spandex|camel|yoga|miamix|^amg|^sporty|balle|dancerb",
		"vib|forced|abuse|torture",
		//                                                                                  "^\b+\b",  This doesn't work
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " This pgm will match an input regexp using smart case, against all filenames in the current directory\n")
		fmt.Fprintf(flag.CommandLine.Output(), " shuffle them, write them to xspf file which is passed on the command line to vlc.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s, working directory is %s, full name of executable is %s and vlcPath is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName, vlcPath)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: lv2 <options> <input-regex> where <input-regex> should not be empty. \n")
		fmt.Fprintf(flag.CommandLine.Output(), " Environment variable VLCPATH can be used to specify the path to vlc binary. \n")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.BoolVar(&verboseFlag, "v", false, " Verbose mode flag.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, " Very Verbose mode flag.")
	flag.StringVar(&excludeRexString, "x", "xspf$", " Exclude file regexp string, which is usually empty.")
	flag.BoolVar(&preBoolOne, "1", false, "Use 1st predefined pattern of femdon|tntu")
	flag.BoolVar(&preBoolTwo, "2", false, "Use 2nd predefined pattern of fuck.*dung|tiefuck|fuck.*bound|bound.*fuck|susp.*fuck|fuck.*susp|sexually|sas")
	flag.BoolVar(&femdomFlag, "femdom", false, "Use predefined pattern for femdom")
	flag.BoolVar(&domFlag, "dom", false, "Use predefined pattern #1.")
	flag.BoolVar(&fuckFlag, "fuck", false, "Use predefined pattern #2.")
	flag.BoolVar(&numericFlag, "numeric", false, "Use predefined pattern ^[0-9]+[0-9]")
	numFlag := flag.Bool("num", false, "Alternate for numeric.")
	flag.BoolVar(&vibeFlag, "vibe", false, "Use predefined pattern: wmbcv|^tbc|^fiterotic|^bjv|hardtied|vib|ethnick|chair|orgasmabuse.")
	flag.BoolVar(&spandexFlag, "spandex", false, "Use spandex predefined pattern: spandex|camel|yoga|miamix|^amg|^sporty|balle|dancerb")
	flag.BoolVar(&forcedFlag, "forced", false, "Use predefined pattern: vib|forced|abuse|torture.")
	flag.IntVar(&numOfTracks, "n", maxNumOfTracks, "Max num of tracks in the output file.  Currently 300.")
	//flag.BoolVar(&notccFlag, "not", true, " Not using tcc flag.") // Since the default is true, to make it false requires -not=false syntax.
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
	} else if forcedFlag {
		includeRexString = preDefinedRegexp[5]
	} else {
		includeRexString = flag.Arg(0) // this is the first argument on the command line that is not the program name.
	}

	if includeRexString == "" { // if there are more than 1 arguments, the extra ones are ignored.
		fmt.Printf(" Usage: lv2 <options> <input-regex> where <input-regex> should not be empty.\n")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println()
		fmt.Printf(" No arguments specified.  Are you sure? ")
		var ans string
		fmt.Scanln(&ans)
		if strings.Contains(strings.ToLower(ans), "n") {
			os.Exit(0)
		}
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
	shuffleAmount := now.Nanosecond()/1e4 + now.Second() + now.Minute() + now.Day() + now.Hour() + now.Year() + len(fileNames) // incr'g # of shuffling loops.
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

	// The file names slice is ready.  Now to create the output file.  Part of the filename will be the regexp used to create this file.

	regexpStr := includeRegex.String()
	rplcPattern := strings.NewReplacer("=", "", "+", "", ".", "", "?", "", "*", "", "|", "_", " ", "", "[", "", "]", "", "^", "", "$", "")
	replacedStr := rplcPattern.Replace(regexpStr)
	outFilename := "vlc" + "_" + replacedStr + "_" + strconv.Itoa(len(fileNames)) + "-" + strconv.Itoa(shuffleAmount) + extDefault
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

	// Now to write out the xspf file

	totStrngLen, err := writeOutputFile(outputFile, fileNames)
	if err != nil {
		fmt.Printf(" Writing output file %s failed w/ ERROR: %s.  Bye-bye.\n", outFilename, err)
		return // this will allow the deferred functions to run.
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

	// Turns out on Windows that vlc was not in the path.  But it shows up when I use "which vlc" in tcc.  On linux only use VLCPATH if set by the user, on Windows always use it, either
	// using the default setting as defined above, or the one set by the user in the environment variable.  Above, vlcPath is blanked if it's not needed, ie, on linux or not set by user.
	vlcStr := whichexec.Find("vlc", vlcPath)

	if vlcStr == "" {
		fmt.Printf(" vlcStr is null.  Exiting ")
		os.Exit(1)
	}

	// Time to run vlc.

	var execCmd *exec.Cmd

	if runtime.GOOS == "windows" {
		execCmd = exec.Command(vlcStr, fullOutFilename)
	} else if runtime.GOOS == "linux" {
		execCmd = exec.Command(vlcStr, fullOutFilename)
	}

	if veryVerboseFlag {
		fmt.Printf(" vlcStr = %s, len of filenames = %d, regex = %s and excludeRegex = %q\n", vlcStr, len(fileNames), includeRegex.String(), excludeRexString)
	}

	execCmd.Stdin = os.Stdin
	//execCmd.Stdout = os.Stdout
	//execCmd.Stderr = os.Stderr //I don't have to assign this.  Let's see what happens if I leave it at nil.  It worked as I hoped.  No errors are displayed to the screen in linux.
	e := execCmd.Start()
	if e != nil {
		fmt.Printf(" Error returned by running vlc %s is %v\n", fullOutFilename, e)
	}

	ctfmt.Printf(ct.Green, false, "Full output filename is %s, ", fullOutFilename)
	ctfmt.Printf(ct.Yellow, false, "from regexp of %s, ", includeRegex.String())
	ctfmt.Printf(ct.Cyan, true, "exclude regexp of %q, ", excludeRexString)
	ctfmt.Printf(ct.Yellow, true, "and total string length= %d\n", totStrngLen)
} // end main()

// ------------------------------------------------------------------------ getFileNames -------------------------------------------------------

func getFileNames(workingDir string, inputRegex *regexp.Regexp) []string {

	fileNames := myReadDir(workingDir, inputRegex) // excluding by regex, filesize or having an ext is done by MyReadDir.

	if veryVerboseFlag {
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

// ------------------------------- writeOutputFile --------------------------------

func writeOutputFile(w io.Writer, fn []string) (int, error) {
	const header1 = `<?xml version="1.0" encoding="UTF-8"?>
<playlist xmlns="http://xspf.org/ns/0/" xmlns:vlc="http://www.videolan.org/vlc/playlist/ns/0/" version="1">
`

	// const titleOpen = "<title>"
	// const titleClose = "</title>"
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
	var totalStrLen int

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
		if i > 0 && i > numOfTracks { // allow i == 0 to mean unlimited.
			break
		}

		fullName, err := filepath.Abs(f)
		if err != nil {
			//fmt.Printf(" filepath.Abs(%s) returned ERROR: %s.  Bye-Bye.\n", f, err)
			return 0, err
		}
		if len(fullName) > lineTooLong {
			continue
		}

		fullName = strings.ReplaceAll(fullName, "\\", "/") // Change backslash to forward slash, if that makes a difference.
		totalStrLen += len(fullName)

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
			fmt.Printf(" Buffered write on track %d returned ERROR: %s", i, err)
			return 0, err
		}
	}

	buf.WriteRune('\t')
	buf.WriteString(trackListClose)
	buf.WriteRune('\n')

	buf.WriteString(playListClose)
	_, err := buf.WriteRune('\n')
	if err != nil {
		return totalStrLen, err // Flush() not called as there's no point after writing returned an error.  It's deferred anyway.
	}
	err = buf.Flush()
	return totalStrLen, err
}
