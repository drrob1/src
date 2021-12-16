/*
  REVISION HISTORY
  ----------------
  20 Mar 20 -- Made comparisons case insensitive.  And decided to make this cgrepi.go.
                 And then I figured I could not improve performance by using more packages.
                 But I can change the side effect of displaying altered case.
  22 Mar 20 -- Will add timing code that I wrote for anack.
  27 Mar 21 -- Changed commandLineFiles in platform specific code, and added the -g flag to force globbing.
  14 Dec 21 -- I'm porting the changed I wrote to multack here.  Also, I noticed that this is mure complex than it
                 needs to be.  I'm going to take a crack at writing a simpler version myself.
                 It takes a list of files from the command line (or on windows, a globbing pattern) and iterates
                 thru all of the files in the list.  Then it exits.  But this is using 2 channels.  I have to understand
                 this better.  It seems much too complex.  I'm going to simplify it.
  16 Dec 21 -- Adding a waitgroup, as the sleep at the end is a kludge.  And will only start number of worker go routines to match number of files.
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
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const LastAltered = "16 Dec 2021"
const maxSecondsToTimeout = 300
const workerPoolMultiplier = 20

var workers = runtime.NumCPU() * workerPoolMultiplier // this works very well in multack

type grepType struct {
	regex    *regexp.Regexp
	filename string
	goRtnNum int
}

var grepChan chan grepType
var totFilesScanned, totMatchesFound int64
var t0, tfinal time.Time
var wg sync.WaitGroup

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)

	// flag definitions and processing
	globflag := flag.Bool("g", false, "force use of globbing, only makes sense on Windows.") // Ptr
	var timeoutOpt = flag.Int64("timeout", 0, "seconds (0 means no timeout)")
	flag.Parse()

	if *timeoutOpt < 1 || *timeoutOpt > maxSecondsToTimeout {
		fmt.Fprintln(os.Stderr, "timeout is", *timeoutOpt, ", and is out of range of [0,300] seconds.  Set to", maxSecondsToTimeout)
		*timeoutOpt = maxSecondsToTimeout
	}
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("a regexp to match must be specified")
	}
	pattern := args[0]
	pattern = strings.ToLower(pattern) // this is the change for the pattern.
	files := args[1:]
	if len(files) < 1 { // no files or globbing pattern on command line.
		if runtime.GOOS == "windows" {
			files = []string{"*.txt"}
		} else {
			files = txtFiles() // intended only for use on linux.
		}
	}

	t0 = time.Now()
	tfinal = t0.Add(time.Duration(*timeoutOpt) * time.Second)
	lineRegex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}

	fmt.Println()
	fmt.Printf(" Concurrent grep insensitive case last altered %s, using pattern of %q, %d worker rtns, compiled with %s. \n",
		LastAltered, pattern, workers, runtime.Version())
	fmt.Println()

	if *globflag && runtime.GOOS == "windows" { // glob function only makes sense on Windows.
		files = globCommandLineFiles(files) // this fails vet because it's in the platform specific code file.
	} else {
		files = commandLineFiles(files)
	}

	minGoRtns := min(len(files), workers)
	// start the worker pool
	grepChan = make(chan grepType, workers) // buffered channel
	for w := 0; w < minGoRtns; w++ {
		go func() {
			for g := range grepChan { // These are channel reads that are only stopped when the channel is closed.
				grepFile(g.regex, g.filename)
			}
		}()
	}

	for _, file := range files {
		wg.Add(1)
		grepChan <- grepType{regex: lineRegex, filename: file}
	}

	goRtns := runtime.NumGoroutine() // must capture this before we sleep for a second.
	wg.Wait()
	close(grepChan) // must close the channel so the worker go routines know to stop.

	elapsed := time.Since(t0)
	//time.Sleep(time.Second) // I've noticed that sometimes main exits before everything can be displayed.  This sleep line fixes that.

	fmt.Println()
	fmt.Println()

	fmt.Printf(" Elapsed time is %s and there are %d go routines that found %d matches in %d files\n", elapsed.String(), goRtns, totMatchesFound, totFilesScanned)
	fmt.Println()
}

func grepFile(lineRegex *regexp.Regexp, fpath string) {
	//fmt.Println(" in grepFile and file is", fpath)
	file, err := os.Open(fpath)
	if err != nil {
		log.Printf("grepFile os.Open error : %s\n", err)
		return
	}
	defer file.Close()
	atomic.AddInt64(&totFilesScanned, 1)
	reader := bufio.NewReader(file)
	for lino := 1; ; lino++ {
		lineStr, er := reader.ReadString('\n')
		lineStrLower := strings.ToLower(lineStr) // this is the change I made to make every comparison case insensitive.
		if lineRegex.MatchString(lineStrLower) {
			fmt.Printf("%s:%d:%s", fpath, lino, lineStr)
			atomic.AddInt64(&totMatchesFound, 1)
		}
		if er != nil { // when can't read any more bytes, break.  The test for er is here so line fragments are processed, too.
			break // just exit when hit EOF condition.
		}
		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}
	}
	wg.Done()
} // end grepFile

func txtFiles() []string { // intended to be needed on linux.
	workingDirname, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "from commandlinefiles:", err)
		return nil
	}
	direntries, err := os.ReadDir(workingDirname) // became available as of Go 1.16
	if err != nil {
		fmt.Fprintln(os.Stderr, "While using os.ReadDir got:", err)
		os.Exit(1)
	}

	pattern := "*.txt"
	matchingNames := make([]string, 0, len(direntries))
	for _, d := range direntries {
		if d.IsDir() {
			continue // skip it
		}
		bool, er := filepath.Match(pattern, strings.ToLower(d.Name()))
		if er != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if bool {
			matchingNames = append(matchingNames, d.Name())
		}
	}
	return matchingNames
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
