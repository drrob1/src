// multack.go
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
   8 Dec 21 -- Will output when .git gets skipped, and will use the pattern of signalling without data, as I learned from Bill Kennedy.
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
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
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

const lastAltered = "2 Oct 2022"
const maxSecondsToTimeout = 300
const null = 0 // null rune to be used for strings.ContainsRune in GrepFile below.

// I started w/ 1000, which works very well on the Ryzen 9 5950X system, where it's runtime is ~10% of anack.
// Here on leox, value of 100 gives runtime is ~30% of anack.  Value of 50 is worse, value of 200 is slightly better than 100.
// Now it will be a multiplier of number of logical CPUs.
const workerPoolMultiplier = 20

type grepType struct {
	regex    *regexp.Regexp
	filename string
	goRtnNum int
}

var grepChan chan grepType
var caseSensitiveFlag bool // default is false.
var totFilesScanned, totMatchesFound int64
var wg sync.WaitGroup

func main() {
	workerPoolSize := runtime.NumCPU() * workerPoolMultiplier
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	var timeoutOpt *int = flag.Int("timeout", 0, "seconds < maxSeconds, where 0 means max timeout currently of 300 sec.")
	verboseFlag := flag.Bool("v", false, "Verbose flag")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > maxSecondsToTimeout {
		log.Fatalln("timeout must be in the range [0,300] seconds")
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = maxSecondsToTimeout
	}

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("a regexp to match must be specified")
	}
	pattern := args[0]
	testCaseSensitivity, _ := regexp.Compile("[A-Z]") // If this matches then there is an upper case character in the input pattern.  And I'm ignoring errors, of course.
	caseSensitiveFlag = testCaseSensitivity.MatchString(pattern)
	if *verboseFlag {
		fmt.Printf(" grep pattern is %s and caseSensitive flag is %t\n", pattern, caseSensitiveFlag)
	}
	if !caseSensitiveFlag {
		pattern = strings.ToLower(pattern) // this is the change for the pattern.
	}
	if *verboseFlag {
		fmt.Printf(" after possible force to lower case, pattern is %s\n", pattern)
	}

	var lineRegex *regexp.Regexp
	var err error
	if lineRegex, err = regexp.Compile(pattern); err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}

	/*  Made obsolete by realization of meaning of null bytes.  Commented out Oct 2, 2022.
	extensions := make([]string, 0, 100)
	if flag.NArg() < 2 {
		//extensions = append(extensions, ".txt")
		extensions = append(extensions, "*")
	} else if runtime.GOOS == "linux" {
		files := args[1:]
		extensions = extractExtensions(files)
	} else { // on windows
		extensions = args[1:]
		for i := range extensions {
			extensions[i] = strings.ToLower(strings.ReplaceAll(extensions[i], "*", ""))
		}
	}
	*/

	startDirectory, _ := os.Getwd() // startDirectory is a string

	fmt.Println()
	fmt.Printf(" Multi-threaded ack, written in Go.  Last altered %s, compiled using %s, and will start in %s, pattern=%s, workerPoolSize=%d.  [Extensions are obsolete]\n\n\n",
		lastAltered, runtime.Version(), startDirectory, pattern, workerPoolSize)

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	if *verboseFlag {
		fmt.Printf(" Current working Directory is %s; %s was last linked %s.\n\n", workingDir, execName, LastLinkedTimeStamp)
	}

	//DirAlreadyWalked := make(map[string]bool, 500)  // now only for directories to be skipped.
	//DirAlreadyWalked[".git"] = true // ignore .git and its subdir's
	// dirToSkip := make(map[string]bool, 5)  This didn't get triggered in a directory I know has a .git.  I'm removing the overhead.
	//dirToSkip[".git"] = true

	// start the worker pool
	grepChan = make(chan grepType, workerPoolSize) // buffered channel
	for w := 0; w < workerPoolSize; w++ {
		go func() {
			for g := range grepChan { // These are channel reads that are only stopped when the channel is closed.
				grepFile(g.regex, g.filename)
			}
		}()
	}

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	// walkfunc closures.  Only the last one is being used now.
	/*
		filepathwalkfunction := func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf(" Error from walk is %v. \n ", err)
				return nil
			}

			if fi.IsDir() {
				//	if DirAlreadyWalked[fpath] { return filepath.SkipDir
				//	} else {  I don't think I have to track the directories visited myself.  So I'm taking this out.
				//		DirAlreadyWalked[fpath] = true
				//	}

				if dirToSkip[fpath] {
					return filepath.SkipDir
				}
				//} else if isSymlink(fi.Mode()) && fi.IsDir() {  // also not needed, because the docs say that walk does not follow symlinks.
				//	return filepath.SkipDir
			} else if fi.Mode().IsRegular() {
				for _, ext := range extensions {
					fpathlower := strings.ToLower(fpath)
					fpathext := filepath.Ext(fpathlower)
					//if strings.HasSuffix(fpathlower, ext) { // only search thru indicated extensions.  Especially not thru binary or swap files.
					if strings.HasPrefix(fpathext, ext) { // added Dec 7, 2021.  So .doc will match .docx, etc.
						grepFile(lineRegex, fpath, resultsChan)
					}
				}
			}

			now := time.Now()
			if now.After(tfinal) {
				log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
			}
			return nil
		}
	*/
	/*
		filepathwalkfunction := func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf(" Error from walk is %v. \n ", err)
				return nil
			}

			if dirToSkip[fpath] {
				return filepath.SkipDir
			}

			for _, ext := range extensions {
				fpathlower := strings.ToLower(fpath)
				fpathext := filepath.Ext(fpathlower)
				//if strings.HasSuffix(fpathlower, ext) { // only search thru indicated extensions.  Especially not thru binary or swap files.
				if strings.HasPrefix(fpathext, ext) { // added Dec 7, 2021.  So .doc will match .docx, etc.
					grepFile(lineRegex, fpath, resultsChan)
				}
			}

			now := time.Now()
			if now.After(tfinal) {
				log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
			}
			return nil
		}

		err = filepath.Walk(startDirectory, filepathwalkfunction)
	*/

	walkDirFunction := func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf(" Error from walk is %v. \n ", err)
			return nil
		}

		if d.IsDir() {
			if filepath.Ext(fpath) == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// only search thru indicated extensions, especially not thru binary or swap files.  Made obsolete by recognition of role of null bytes in files.
		/*  commented out 10/2/22.
		for _, ext := range extensions {
			fpathLower := strings.ToLower(fpath)
			fpathExt := filepath.Ext(fpathLower)

			if strings.HasPrefix(fpathExt, ext) { // added Dec 7, 2021.  So .doc will match .docx, etc.
				wg.Add(1)
				grepChan <- grepType{ // send this to a worker go routine.
					regex:    lineRegex,
					filename: fpath,
				}
			}
		}
		*/

		wg.Add(1)
		grepChan <- grepType{ // send this to a worker go routine.
			regex:    lineRegex,
			filename: fpath,
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
	wg.Wait()
	close(grepChan) // must close the channel so the worker go routines know to stop.

	elapsed := time.Since(t0)

	//time.Sleep(time.Second) // I've noticed that sometimes main exits before everything can be displayed.  I'm trying to stop that.
	//fmt.Println(" Elapsed time is", elapsed, "and number of Go routines is", goRtns)
	fmt.Printf(" Elapsed time is %s and there are %d go routines that found %d matches in %d files\n", elapsed.String(), goRtns, totMatchesFound, totFilesScanned)
	fmt.Println()
} // end main

func grepFile(lineRegex *regexp.Regexp, fpath string) {
	var lineStrng string // either case sensitive or case insensitive string, depending on value of caseSensitiveFlag, which itself depends on case sensitivity of input pattern.
	var localMatches int64
	file, err := os.Open(fpath)
	defer func() {
		file.Close()
		atomic.AddInt64(&totFilesScanned, 1)
		atomic.AddInt64(&totMatchesFound, localMatches)
		wg.Done()
	}()
	if err != nil {
		log.Printf("grepFile os.Open error : %s\n", err)
		return
	}

	reader := bufio.NewReader(file)
	for lino := 1; ; lino++ {
		lineStr, er := reader.ReadString('\n')

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
			fmt.Printf("%s:%d:%s", fpath, lino, lineStr)
			localMatches++
		}
		if er != nil { // when can't read any more bytes, break.  The test for er is here so line fragments are processed, too.
			//if err != io.EOF { // this became messy, so I'm removing it
			//	log.Printf("error from reader.ReadString in grepfile %s line %d: %s\n", fpath, lino, err)
			//}
			break // just exit when hit EOF condition.
		}
	}
} // end grepFile

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
/*
// ------------------------------ isSymlink ---------------------------
func isSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink

*/
