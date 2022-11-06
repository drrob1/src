// multack2.go
/*
  REVISION HISTORY
  ----------------
   1 Apr 20 -- Making it multi-threaded by using go routines by copying cgrepi.go and multimap.go.
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
  13 Dec 21 -- Now called multack2.go.  I want to add atomic totFilesScanned and totMatchesFound
  16 Dec 21 -- Putting back waitgroup.  The sleep at the end is a kludge.
  19 Dec 21 -- Will add the more selective use of atomic instructions as I learned about from Bill Kennedy and is in multack3.go.  But I will
                 keep reading the file line by line.  Can now time difference when number of atomic operations is reduced.
                 Multack3 is still faster, so most of the slowness here is the line by line file reading.
   4 Nov 22 -- Will try to remove the wait group code and instead use the pattern of a done channel.  And included the pattern of returning SkipDir on err.
                 And also will SkipDir on vmware.
                 And added matchType so can sort the results by ranging over a channel.  This is needed for the done channel pattern to work.
                 I'm not going to bother w/ smart case.  But I did remove the extension detection code and will use null byte detection instead.
                 I'm stuck.  I can't close the resultsChan when all work is finished sending to grepFile, as I must account for the work actually taking some time.
                 Here I do need a wait group, AFAICT.  Since assumes that the time it takes for work to complete after being sent is negligible, so the done channel works there but not here.

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

const lastAltered = "5 Nov 2022"
const maxSecondsToTimeout = 300
const nullRune = 0 // null rune to be used for strings.ContainsRune in GrepFile below.

var totFilesScanned, totMatchesFound int64

// I started w/ 1000, which works very well on the Ryzen 9 5950X system, where it's runtime is ~10% of anack.
// Here on leox, value of 100 gives runtime is ~30% of anack.  Value of 50 is worse, value of 200 is slightly better than 100.
// Now it will be a multiplier of number of logical CPUs.
const workerPoolMultiplier = 20

type grepType struct {
	regex    *regexp.Regexp
	filename string
	//goRtnNum int  not used.  Not even usable.
}

var grepChan chan grepType

//type matchType struct {
//	fpath        string
//	lino         int
//	lineContents string
//}

var wg sync.WaitGroup
var resultsChan chan string

func main() {
	workerPoolSize := runtime.NumCPU() * workerPoolMultiplier
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	var timeoutOpt *int = flag.Int("timeout", 0, "seconds < maxSeconds, where 0 means max timeout currently of 300 sec.")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > maxSecondsToTimeout {
		log.Fatalln("timeout must be in the range [0,300] seconds")
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = maxSecondsToTimeout
	}

	if flag.NArg() < 1 {
		log.Fatalln("a regexp to match must be specified")
	}
	pattern := flag.Arg(0)
	pattern = strings.ToLower(pattern)
	var lineRegex *regexp.Regexp
	var err error
	if lineRegex, err = regexp.Compile(pattern); err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}
	/*
	   Made obsolete by detection of null runes in GrepFile.
	   	extensions := make([]string, 0, 100)
	   	if flag.NArg() < 2 {
	   		extensions = append(extensions, ".txt")
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
	fmt.Printf(" Multi-threaded ack, written in Go.  Last altered %s, compiled using %s, and will start in %s, pattern=%s, extensions= removed, workerPoolSize=%d.\n\n\n",
		lastAltered, runtime.Version(), startDirectory, pattern, workerPoolSize)

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

	var results []string
	done := make(chan bool)                         // unbuffered, so it's a blocking receive.  This is how the sync is done.
	resultsChan = make(chan string, workerPoolSize) // buffered channel
	go func() {
		for r := range resultsChan {
			results = append(results, r)
		}
		sort.Strings(results)
		done <- true
	}()

	// walkfunc closure.
	walkDirFunction := func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf(" Error from walk is %v. \n ", err)
			return filepath.SkipDir
		}

		if d.IsDir() {
			if filepath.Ext(fpath) == ".git" || strings.EqualFold(fpath, "vmware") {
				return filepath.SkipDir
			}
			return nil
		}

		/*		for _, ext := range extensions { // only search thru indicated extensions.  Especially not thru binary or swap files.
					fpathLower := strings.ToLower(fpath)
					fpathExt := filepath.Ext(fpathLower)

					if strings.HasPrefix(fpathExt, ext) { // added Dec 7, 2021.  So .doc will match .docx, etc.
						wg.Add(1)
					}
				}
		*/
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
	close(grepChan)                  // must close the channel so the worker go routines know to stop, and will do that after all work has been sent to the workers.

	<-done
	wg.Wait() // I wonder if there's some way I can know the value of the wait group count when I get here?

	elapsed := time.Since(t0)

	fmt.Printf(" Elapsed time is %s, number of Go routines is %d, and %d matches were found in %d files scanned\n", elapsed.String(), goRtns, totMatchesFound, totFilesScanned)
	fmt.Println()
} // end main

func grepFile(lineRegex *regexp.Regexp, fpath string) {
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
		wg.Done()
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

		if strings.ContainsRune(lineStr, nullRune) { // this way binary files are not searched.  Don't need extensions anymore.
			return
		}
		// this is the change I made to make every comparison case insensitive.
		// lineStr = strings.TrimSpace(line)  Try this without the TrimSpace.
		lineStrLower := strings.ToLower(lineStr)

		if lineRegex.MatchString(lineStrLower) {
			//matchChan <- matchType{
			//	fpath:        fpath,
			//	lino:         lino,
			//	lineContents: lineStr,
			//}
			fmt.Printf("%s:%d:%s", fpath, lino, lineStr)
			str := fmt.Sprintf("%s:%d:%s", fpath, lino, lineStr)
			resultsChan <- str
			//atomic.AddInt64(&totMatchesFound, 1)
			localMatches++
		}
	}
} // end grepFile

// ------------------------------ isSymlink ---------------------------
func isSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink

/*
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
