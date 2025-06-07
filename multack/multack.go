// multack.go
package main

import (
	"bufio"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/*
  REVISION HISTORY
  ----------------
   1 Apr 20 -- Making it multithreaded by using go routines by copying cgrepi.go and multimap.go.
               Now created multack.go, derived from anack.go.
               With a ResultType buffer of   1,024 items, it's  <1% faster than anack, if that much.
               With a ResultType buffer of  10,000 items, it's  ~5% faster than anack.
               With a ResultType buffer of  50,000 items, it's ~15% faster than anack.
               With a ResultType buffer of 100,000 items, it's ~24% faster than anack.
               I'll stop at 100,000 items.  It's great it works.
               (4/10/24 but it's obvious I got the code wrong)
   2 Apr 20 -- Updated its start string to declare its correct name.  I forgot to change that yesterday.
  23 Apr 20 -- 2 edge cases don't work on linux.  If there is a filepattern but no matching files in the start directory,
                and if there is only 1 matching file in the start directory.
                And also if there appears to be more than one extension, like gastric.txt.out.
   5 Sep 20 -- Will not search thru symlinked directories
  27 Mar 21 -- making sure that the filename matches are case insensitive
   6 Dec 21 -- Maybe something I'm learning from Bill Kennedy applies here.  I only made the doneChan buffered.
   7 Dec 21 -- Extensions like .xls will also match .xlsm, .xlsx, etc.  And I don't think I have to track which directories I've visited, as the library func does that.
                 So I'll just use the map as a list of known directories to skip.  So far, only ".git" is skipped.
                 And I don't have to check for IsDir() or IsRegular(), so I removed that, also.
               Starting w/ Go 1.16, there is a new walk function, that does not use a FiloInfo but a dirEntry, which they claim is faster.  I'll try it.
   8 Dec 21 -- Will output when .git gets skipped, and will use the pattern of signaling without data, as I learned from Bill Kennedy.
  10 Dec 21 -- I'm testing for .git and will skipdir if found.  And will simply return on IsDir.
                 I'm going to restructure this to use waitgroups.  I'll see how that goes.
                 I think I was having a shadowing problem w/ err.  When I made that er, the code started working.
  11 Dec 21 -- Now I got the error that too many files were open.  So I need a worker pool.
  12 Dec 21 -- Added test for ".git" to SkipDir, and will measure responsiveness w/ different values for workerPoolSize.
                 I decided to base the workerPoolSize on a multiplier from runtime.NumCPU.  And to display NumGoroutine at the end.
  16 Dec 21 -- Need a waitgroup after all.  The sleeping at the end is a kludge.
   1 Oct 22 -- Adding smart case as I did yesterday for cgrepi.  If input pattern is lower case, search is case insensitive.  If input pattern is upper case, the search
                 is case sensitive.  And adding using a null byte as a marker for a binary file and then aborting that file.  Both ideas came from ripgrep.
                 Adding a count of matches and files, copied from cgrepi.go.
   2 Oct 22 -- Now that I've learned to abort a binary file as one that has null bytes, I don't need the extension system anymore.
                 And I corrected the order of defer vs if err in the grepFile routine.
   6 Oct 22 -- Will sort output of this routine, so all file matches are output together.  First debugged for cgrepi.
   7 Oct 22 -- Will add color to the output messages.
  21 Oct 22 -- Ran golangci-lint and made the changes it recommended.
  26 Oct 22 -- If I pattern this after since, which was essentially written by Michael T. Jones, I can eliminate the need for a waitgroup here.
                 This is because when the walk function returns, the work has all been sent to the workers and the work channel can be closed.
                 The new pattern to replace a wait group uses a done channel.
                 It doesn't work.  The only optimization I can make is that the sort.Strings is in the go routine instead of the main routine.
                 I can't close the results channel when all work is sent to the workers because of the processing time needed.  I'll restore the wait group.
   5 Nov 22 -- Walk function now returns SkipDir on errors, as I recently figured out when updating since.go.  And now allows a start dir after the regexp on command line.
   8 Nov 22 -- Fixed error as to when to return SkipDir.  I had it depend on verboseFlag, and that was an obvious error.
  13 Nov 22 -- Adding ability to optionally specify a start directory other than the current one.  Nevermind, it already has this.
  14 Nov 22 -- Adding a usage message.  I never did that before.  And adding processing for '~' which only applies to Windows.
  21 Nov 22 -- static linter found an error w/ a format verb.  Now fixed.
  24 Feb 23 -- I'm changing the multiplier to = 1, based on what Bill Kennedy said, ie, that NumCPU() is sort of a sweet spot.  And Evan is 31 today, but that's not relevant here.
  25 Feb 23 -- Optimizing walkDir as I did in since.go.  Run os.Stat only after directory check for the special directories and only call deviceID on a dir entry.
  10 Apr 24 -- I/O bound jobs benefit from having more workers than what NumCPU() says.
                 But I have to remember that linux only has 1000 or so file handles; this number cannot be exceeded.
   6 May 24 -- Wait groups are for the goroutines themselves, not the items processed by the goroutines.  I'm making that change now.
  10 May 24 -- Made sliceSize 50_000, as this can return ~20K matches when run in src directory.
  20 Nov 24 -- Will now exclude OneDrive.  This crashes Windows, so I have to exclude it.  And I'm excluding AppData.  Excluding AppData sped up the code a lot, from 3 min to 10 sec on Win11.
   2 Mar 25 -- Clarified help message that the pattern is a regexp, not a glob.  And I changed to using pflag as a drop-in replacement for flag.
   6 Jun 25 -- I got the idea to add an exclude expresssion, after I tried to use one and found that I never implemented that here.  Copied code I write in cgrepi to here.
				Turns out that none of the std grep versions have a way to do this.
*/

const lastAltered = "6 June 2025"
const maxSecondsToTimeout = 300
const null = 0 // null rune to be used for strings.ContainsRune in GrepFile below.

// I started w/ 1000 workers, which works very well on the Ryzen 9 5950X system, where it's runtime is ~10% of anack.
// On leox, value of 100 gives runtime is ~30% of anack.  Value of 50 is worse, value of 200 is slightly better than 100.
// Now it will be a multiplier of number of logical CPUs.

const multiplier = 10 // default value for the worker pool multiplier
var workerPoolMultiplier int

const sliceSize = 50_000 // a magic number I plucked out of the air.

type devID uint64

type grepType struct {
	regex    *regexp.Regexp
	excluded *regexp.Regexp // added 6/6/25
	filename string
	// goRtnNum int
}

type matchType struct {
	fpath        string
	lino         int
	lineContents string
}

type matchesSliceType []matchType // this is a named type.  I need a named type for the sort functions to work.  An anonymous type won't cut it.

func (m matchesSliceType) Len() int {
	return len(m)
}
func (m matchesSliceType) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m matchesSliceType) Less(i, j int) bool {
	return strings.ToLower(m[i].fpath) < strings.ToLower(m[j].fpath)
}

var grepChan chan grepType
var matchChan chan matchType
var caseSensitiveFlag bool // default is false.
var totFilesScanned, totMatchesFound int64
var sliceOfStrings []string // based on an anonymous type.
var wg sync.WaitGroup
var verboseFlag, veryverboseFlag bool
var excludeStr string

func main() {

	helpFcn := func() { // this wasn't working until I moved this to above flag.Parse()
		fmt.Printf(" %s last altered %s, compiled with %s\n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf("\n Usage: multack [option flags] regexp [start-Directory]\n")
		flag.PrintDefaults()
	}
	flag.Usage = helpFcn

	timeoutOpt := flag.Int("timeout", 0, "seconds < maxSeconds, where 0 means max timeout currently of 300 sec.")
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "Verbose flag")
	flag.BoolVarP(&veryverboseFlag, "vv", "w", false, "Very Verbose flag")
	flag.BoolVar(&veryverboseFlag, "veryverbose", false, "Very Verbose flag synonym")
	flag.IntVarP(&workerPoolMultiplier, "multiplier", "m", multiplier, "Multiplier for the number of goroutines in the worker pool.")
	flag.StringVarP(&excludeStr, "exclude", "x", "", "Exclude regular expression")
	flag.Parse()

	workerPoolSize := runtime.NumCPU() * workerPoolMultiplier
	if veryverboseFlag {
		verboseFlag = true
	}

	if *timeoutOpt < 0 || *timeoutOpt > maxSecondsToTimeout {
		fmt.Printf(" %s last altered %s, compiled with %s\n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf(" Usage: multack [option flags] regexp [start Directory]\n")
		log.Fatalln("timeout must be in the range [0,300] seconds")
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = maxSecondsToTimeout
	}

	if flag.NArg() < 1 {
		fmt.Printf("\n %s last altered %s, compiled with %s\n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf(" Usage: multack regexp [start Directory]\n")
		fmt.Printf("\t a regexp to match must be specified\n")
		return
	}
	pattern := flag.Arg(0)
	testCaseSensitivity, _ := regexp.Compile("[A-Z]") // If this matches then there is an upper case character in the input pattern.  And I'm ignoring errors, of course.
	caseSensitiveFlag = testCaseSensitivity.MatchString(pattern)
	if verboseFlag {
		fmt.Printf(" grep pattern is %s and caseSensitive flag is %t\n", pattern, caseSensitiveFlag)
	}
	if !caseSensitiveFlag {
		pattern = strings.ToLower(pattern) // this is the change for the pattern.
	}
	if verboseFlag {
		fmt.Printf(" after possible force to lower case, pattern is %s\n", pattern)
	}

	var lineRegex, excludeRegex *regexp.Regexp
	var err error
	if lineRegex, err = regexp.Compile(pattern); err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}

	if excludeStr != "" {
		excludeRegex, err = regexp.Compile(excludeStr)
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Exclude regexp.Compile(%q) is invalid, error is %s\n", excludeStr, err.Error())
		}
	}

	startDirectory, errr := os.Getwd() // startDirectory is a string
	if errr != nil {
		fmt.Printf(" Error from os.Getwd() is %s\n", errr)
		fmt.Printf(" Usage: multack regexp [start Directory]\n")
		os.Exit(1)
	}
	if flag.NArg() >= 2 { // will use 2nd arg as start dir and will ignore any others
		startDirectory = flag.Arg(1)
		home, er := os.UserHomeDir()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from os.UserHomeDir() is %s.  Exiting. \n", er)
			os.Exit(1)
		}
		startDirectory = strings.ReplaceAll(startDirectory, "~", home)
	}
	startInfo, err := os.Stat(startDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Aborting\n.", startDirectory, err)
		fmt.Printf(" Usage: multack regexp [start Directory]\n")
		os.Exit(1)
	}
	startDeviceID := getDeviceID(startInfo)

	fmt.Println()
	fmt.Printf(" Multi-threaded ack, written in Go.  Last altered %s, compiled using %s,\n will start in %q, pattern=%q, workerPoolSize=%d. \n [Extensions are obsolete]\n\n",
		lastAltered, runtime.Version(), startDirectory, pattern, workerPoolSize)

	if verboseFlag {
		execDir, _ := os.Getwd()
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
		fmt.Printf(" Current working Directory is %s; %s timestamp is %s.\n\n", execDir, execName, LastLinkedTimeStamp)
	}

	matchChan = make(chan matchType, sliceSize)               // this is a buffered channel.
	sliceOfAllMatches := make(matchesSliceType, 0, sliceSize) // this uses a named type, needed to satisfy the sort interface.
	sliceOfStrings = make([]string, 0, sliceSize)             // this uses an anonymous type.
	doneChan := make(chan bool)
	go func() { // start the receiving operation before the sending starts
		for match := range matchChan {
			sliceOfAllMatches = append(sliceOfAllMatches, match)
			s := fmt.Sprintf("%s:%d:%s", match.fpath, match.lino, match.lineContents)
			sliceOfStrings = append(sliceOfStrings, s)
		}
		sort.Stable(sliceOfAllMatches)
		sort.Strings(sliceOfStrings) // the sort operation is now done here in the go routine, instead of the main function body.
		close(doneChan)
	}()

	// start the worker pool
	wg.Add(workerPoolSize)
	grepChan = make(chan grepType, workerPoolSize) // buffered channel
	for range workerPoolSize {                     // don't need w:=0;w<workerPoolSize;w++ anymore.
		go func() {
			defer wg.Done()
			for g := range grepChan { // These are channel reads that are only stopped when the channel is closed.
				grepFile(g.regex, g.excluded, g.filename)
			}
		}()
	}

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	// walkfunc closure.

	walkDirFunction := func(fPath string, d os.DirEntry, err error) error { // this doesn't follow symlinks
		if err != nil {
			fmt.Printf(" Error from walkdirFunction is %v. \n ", err)
			return filepath.SkipDir
		}

		if d.IsDir() {
			if filepath.Ext(fPath) == ".git" || strings.Contains(fPath, ".config") || strings.Contains(fPath, ".local") ||
				strings.Contains(fPath, "vmware") || strings.Contains(fPath, ".cache") {
				return filepath.SkipDir
			}

			lower := strings.ToLower(fPath)
			if strings.Contains(lower, "onedrive") || strings.Contains(lower, "appdata") { // skipping appdata sped up this code a lot.  From 3 minutes to 10 sec on Win11.
				return filepath.SkipDir
			}

			info, _ := d.Info()
			deviceID := getDeviceID(info)
			if startDeviceID != deviceID {
				if verboseFlag {
					fmt.Printf(" DeviceID for %s is %d which is different than %d for %s.  Skipping\n", startDirectory, startDeviceID, deviceID, fPath) // fixed a format verb here.
				}
				return filepath.SkipDir
			}

			return nil
		}

		grepChan <- grepType{ // send this to a worker go routine.
			regex:    lineRegex,
			excluded: excludeRegex,
			filename: fPath,
		}

		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}
		return nil
	}

	err = filepath.WalkDir(startDirectory, walkDirFunction)
	if err != nil {
		log.Fatalln(" Error from filepath.walk is", err, ".  Elapsed time is", time.Since(t0))
	}

	goRtns := runtime.NumGoroutine() // must capture this before we sleep for a second.
	close(grepChan)                  // must close the channel so the worker go routines know to stop.  When get here, all work is sent to the workers.

	wg.Wait()        // all grep routines are finished when this is allowed to continue.
	close(matchChan) // must close the channel so the matchChan for loop will end.  And I have to do this after all the work is done.
	<-doneChan

	elapsed := time.Since(t0)

	gotWin := runtime.GOOS == "windows"

	// Time to show

	for _, m := range sliceOfAllMatches { //This is the only output that will be seen.
		fmt.Printf("%s:%d:%s", m.fpath, m.lino, m.lineContents) // remember that lineContents includes a \n at the end of each string.
	}

	ctfmt.Printf(ct.Yellow, gotWin, " Elapsed time is %s and there are %d go routines that found %d matches in %d files\n", elapsed, goRtns, totMatchesFound, totFilesScanned)
	fmt.Println()
} // end main

func grepFile(lineRegex, excludeRegex *regexp.Regexp, fpath string) {
	var lineStrng string // either case sensitive or case insensitive string, depending on value of caseSensitiveFlag, which itself depends on case sensitivity of input pattern.
	var localMatches int64
	file, err := os.Open(fpath)
	if err != nil {
		log.Printf("grepFile os.Open error : %s\n", err)
		return
	}

	defer func() {
		file.Close()
		atomic.AddInt64(&totFilesScanned, 1)
		atomic.AddInt64(&totMatchesFound, localMatches)
	}()
	reader := bufio.NewReader(file)
	for lino := 1; ; lino++ {
		lineStr, er := reader.ReadString('\n')
		if er != nil { // when can't read any more bytes, break.  The test for er is here so line fragments are processed, too.
			//if err != io.EOF { // this became messy, so I'm removing it
			//	log.Printf("error from reader.ReadString in grepfile %s line %d: %s\n", fpath, lino, err)
			//}
			break // just exit when hit EOF condition.
		}

		if strings.ContainsRune(lineStr, null) {
			return // the defer func()	 will take care of the cleanup here.
		}
		if caseSensitiveFlag {
			lineStrng = lineStr
		} else {
			lineStrng = strings.ToLower(lineStr) // this is the change I made to make every comparison case insensitive.
		}

		// lineStr = strings.TrimSpace(line)  Try this without the TrimSpace.

		if lineRegex.MatchString(lineStrng) {
			if veryverboseFlag {
				fmt.Printf("%s:%d:%s", fpath, lino, lineStr)
			}
			if excludeRegex == nil { // if there is no excludeRegex, then this matcch is enough to send it down the channel.
				localMatches++
				matchChan <- matchType{
					fpath:        fpath,
					lino:         lino,
					lineContents: lineStr,
				}
			} else if !excludeRegex.MatchString(lineStrng) { // if there is an excludeRegex, then need to test it.
				localMatches++
				matchChan <- matchType{
					fpath:        fpath,
					lino:         lino,
					lineContents: lineStr,
				}
			}
		}
	}
} // end grepFile

/*  Made obsolete by null detection.
func extractExtensions(files []string) []string {
	var extensions sort.StringSlice
	extensions = make([]string, 0, 100)
	for _, file := range files {
		ext := filepath.Ext(file)
		extensions = append(extensions, ext)
	}
	if len(extensions) > 1 {
		extensions.Sort()
		for i := range extensions {
			if i == 0 {
				continue
			}
			if extensions[i-1] == extensions[i] {
				extensions[i-1] = "" // This needs to be [i-1] because when it was [i] it interferred w/ the next iteration.
			}
		}
		//fmt.Println(" in extractExtensions before 2nd sort:", extensions)
		sort.Sort(sort.Reverse(extensions))

		trimmedExtensions := make([]string, 0, len(extensions))
		for _, ext := range extensions {
			if ext != "" {
				trimmedExtensions = append(trimmedExtensions, ext)
			}
		}
		//fmt.Println(" in extractExtensions after sort trimmedExtensions:", trimmedExtensions)
		//fmt.Println()
		return trimmedExtensions
	}
	//fmt.Println(" in extractExtensions without a sort:", extensions)
	//fmt.Println()
	return extensions
} // end extractExtensions
*/
