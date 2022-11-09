// multack3.go
/*
  REVISION HISTORY
  ----------------
   1 Apr 20 -- Making it concurrent by using go routines by copying cgrepi.go and multimap.go.
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
  16 Dec 21 -- Putting back wait group.  The sleep at the end is a kludge.
  18 Dec 21 -- Now called multack3.go.  I'm going to add the optimizations that Bill Kennedy uses in the last example of his class.
                 The main difference here is thet each file is read entirely at once, not line by line.
                 I think this code is slightly faster than multack2, but it's hard for me to really know.  I may have to wait to test
                 until after I'm finished using VMware workstation.
   5 Nov 22 -- Updating code based on what I've learned when writing since.go, and adding nullRune detection instead of needing extensions.
                 I'm not adding smartcase.  Will skip vmware directories as well as .git ones.
   6 Nov 22 -- I've been struggling here for 2 days, and the bug was that I forgot to initialize resultsChan w/ a make call.  I finally tripped over that error
                 when I removed the wait() call and tried to close(resultsChan).  I got an error saying that I tried to close a nil channel.
                 Now this pgm works.
   7 Nov 22 -- Well, it doesn't work on Windows.  If there's a file error, then the wait group count goes below zero and panics.  I found the extra call to wg.Done(), and removed it.
   8 Nov 22 -- I'll add a stat check to make sure I skip files > 100 MB.  And checked for device ID which only makes sense on linux.  Interestingly, this pgm is faster
                 than multack on linux, but much slower on Windows.
   9 Nov 22 -- I'm going to remove the call to os.Stat().  No difference in the timing.  It's still worse than multack, and sometimes much worse.
                 Here uses a DirEntry callback func.  And it's slightly slower than multack2 which now only differs by using os.FileInfo callback func.
*/

package main

import (
	"bytes"
	"errors"
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

const lastAltered = "9 Nov 2022"
const maxSecondsToTimeout = 300
const nullByte = 0 // null rune to be used for strings.ContainsRune in GrepFile below.
const K = 1024
const M = K * K
const fileTooBig = 100 * M

var totFilesScanned, totMatchesFound, totalMatchesFound int64

// I started w/ 1000, which works very well on the Ryzen 9 5950X system, where it's runtime is ~10% of anack.
// Here on leox, value of 100 gives runtime is ~30% of anack.  Value of 50 is worse, value of 200 is slightly better than 100.
// Now it will be a multiplier of number of logical CPUs.
// I'll see how this works.  Bill Kennedy keeps saying that less is more regarding go routines, but the channel buffer can take it all.
const workerPoolMultiplier = 20

type devID uint64

type grepType struct {
	regex    *regexp.Regexp
	filename string
	goRtnNum int
}

var grepChan chan grepType
var resultsChan chan string
var verboseFlag, veryverboseFlag bool

var wg sync.WaitGroup

func main() {
	workerPoolSize := runtime.NumCPU() * workerPoolMultiplier
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	var timeoutOpt = flag.Int("timeout", 0, "seconds < maxSeconds, where 0 means max timeout currently of 300 sec.")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose flag")
	flag.BoolVar(&veryverboseFlag, "vv", false, "Very Verbose flag")
	flag.Parse()
	if veryverboseFlag {
		verboseFlag = true
	}

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
	pattern = strings.ToLower(pattern)
	var lineRegex *regexp.Regexp
	var err error
	if lineRegex, err = regexp.Compile(pattern); err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}

	startDirectory, _ := os.Getwd() // startDirectory is a string
	startInfo, err := os.Stat(startDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Aborting\n.", startDirectory, err)
		os.Exit(1)
	}
	startDeviceID := getDeviceID(startDirectory, startInfo)

	fmt.Println()
	fmt.Printf(" Multi-threaded ack, %s, written in Go.  Last altered %s, compiled using %s, and will start in %s, pattern=%s, extensions were deleted, workerPoolSize=%d.\n\n\n",
		os.Args[0], lastAltered, runtime.Version(), startDirectory, pattern, workerPoolSize)

	// read resultschan so can sort the results and simulate ordered traversal.

	results := make([]string, 0, workerPoolSize)
	resultsChan = make(chan string, workerPoolSize)
	go func() { // start the receiving operation before the sending starts
		for r := range resultsChan {
			results = append(results, r)
		}
		sort.Strings(results) // the sort operation is now done here in the go routine, instead of the main function body.
	}()

	grepChan = make(chan grepType, workerPoolSize) // buffered channel
	// start the worker pool until the channel closes.
	for w := 0; w < workerPoolSize; w++ {
		go func() {
			for g := range grepChan { // These are channel reads that are only stopped when the channel is closed.
				grepFile(g.regex, g.filename)
			}
		}()
	}

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	// walkfunc closure.
	walkDirFunction := func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf(" Error from walkDirFunc is %v. \n ", err)
			return filepath.SkipDir
		}

		if d.IsDir() {
			if filepath.Ext(fpath) == ".git" || strings.Contains(fpath, "vmware") {
				return filepath.SkipDir
			}
			return nil
		}

		// I don't want to follow links on linux to other devices like DSM or bigbkupG.
		info, _ := d.Info()
		deviceID := getDeviceID(fpath, info)
		if startDeviceID != deviceID {
			if verboseFlag {
				fmt.Printf(" DeviceID for %s is %d which is different than %d for %d.  Skipping\n", startDirectory, startDeviceID, deviceID, fpath)
			}
			return filepath.SkipDir
		}
		if info.Size() > fileTooBig {
			if verboseFlag {
				fmt.Printf(" Size for %s is %d which is > than %d.  Skipping\n", fpath, info.Size(), fileTooBig)
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			if verboseFlag {
				fmt.Printf(" %s is not a regular file.  Skipping.\n", fpath)
			}
			return nil
		}

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
		log.Printf(" after walkdir and error from filepath.walk is %s.  Elapsed time is %s.", err, time.Since(t0))
	}
	goRtnsNum := runtime.NumGoroutine() // must capture this before they shutdown.

	close(grepChan) // must close the channel so the worker go routines know to stop.  IE, closing the channel after all work is sent to the workers.
	if verboseFlag {
		fmt.Printf("\n The grepChan is closed now.  Scanned %d files, and found (%d, %d) matches using %d go routines.\n\n",
			totFilesScanned, totMatchesFound, totalMatchesFound, goRtnsNum)
	}

	wg.Wait() // wait until all the go routines stop.

	close(resultsChan) // Here I'm closing the channel after all the work is done.
	if verboseFlag {
		fmt.Printf("\n The resultsChan is closed now.\n\n")
	}

	elapsed := time.Since(t0)

	fmt.Printf("\n\n Will now output the results after being sorted.\n")
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	fmt.Printf("\n\n Elapsed time is %s, number of Go routines is %d, and %d matches were found in %d files scanned\n", elapsed.String(), goRtnsNum, totMatchesFound, totFilesScanned)
	fmt.Println()
} // end main

func grepFile(lineRegex *regexp.Regexp, fpath string) {
	var localMatches int64
	defer func() {
		//file.Close()  Not needed now that the entire file is being read in at once.
		wg.Done()
		atomic.AddInt64(&totFilesScanned, 1)
		atomic.AddInt64(&totMatchesFound, localMatches) // only need one atomic instruction per go routine
	}()
	//fi, e := os.Stat(fpath)
	//if e != nil {
	//	log.Printf(" os.Stat(%s) returns error of %s\n", fpath, e)
	//	return
	//}
	//if !fi.Mode().IsRegular() {
	//	log.Printf(" os.Stat(%s) is not a regular file.  Returning.\n", fpath)
	//	return
	//}
	//if fi.Size() > 100*M {
	//	log.Printf(" %s is too big and was skipped.  It's size is %d.\n", fpath, fi.Size())
	//	return
	//}

	file, err := os.ReadFile(fpath) // changing to read entire file in at once.
	if err != nil {
		log.Printf("grepFile os.ReadFile error : %s\n", err)
		//wg.Done()  too many wg.Done() calls here.  That's why it's panicking.
		return
	}
	//file = append(file, byte('\n')) // sentinel marker

	//reader := bufio.NewReader(file) // don't need this now that entire file is read in at once.
	reader := bytes.NewReader(file)
	for lino := 1; ; lino++ {
		lineStr, er := readLine(reader)
		if er != nil { // when can't read any more bytes, break.
			break // just exit when hit any error, which will mostly be either end of file or binary file containing a null byte.
		}

		// this is the change I made to make every comparison case insensitive.
		// lineStr = strings.TrimSpace(line)  Try this without the TrimSpace.
		lineStrLower := strings.ToLower(lineStr)

		if lineRegex.MatchString(lineStrLower) {
			localMatches++ // a local variable does not need to be atomically incremented.
			atomic.AddInt64(&totalMatchesFound, 1)
			s := fmt.Sprintf("%s:%d:%s", fpath, lino, lineStr)
			if veryverboseFlag {
				fmt.Printf("%s\n", s)
			}
			resultsChan <- s
		}
	}
} // end grepFile

// ----------------------------------------------------------
// readLine

func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		rn, siz, err := r.ReadRune() // byte and rune are reserved words for a variable type.
		/*		if verboseFlag {
					fmt.Printf(" %c %v ", byt, err)
					pause()
				}
		*/ //if err == io.EOF {  I have to return io.EOF so the EOF will be properly detected as such.
		//	return strings.TrimSpace(sb.String()), nil
		//} else
		if err != nil || siz == 0 {
			//return strings.TrimSpace(sb.String()), err
			return sb.String(), err
		}
		if siz > 1 {
			continue
		}
		if rn == '\n' { // will stop scanning a line after seeing these characters like in bash or C-ish.
			//return strings.TrimSpace(sb.String()), nil
			break
		}
		if rn == '\r' {
			continue
		}
		if rn == nullByte {
			e := errors.New("null byte found interpretted as a file to be skipped")
			return "", e
		}
		_, err = sb.WriteRune(rn)
		if err != nil {
			//return strings.TrimSpace(sb.String()), err
			return sb.String(), err
		}
	}
	//return strings.TrimSpace(sb.String()), nil
	return sb.String(), nil
} // readLine

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

/*
// ------------------------------ isSymlink ---------------------------
func isSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink

*/
