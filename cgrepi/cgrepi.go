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
  19 Dec 21 -- Will add the more selective use of atomic instructions as I learned about from Bill Kennedy and is in cgrepi2.go.  But I will
                 keep reading the file line by line.  Can now time difference when number of atomic operations is reduced.
                 Cgrepi2 is still faster, so most of the slowness here is the line by line file reading.
  30 Sep 22 -- Got idea from ripgrep about smart case, where if input string is all lower case, then the search is  ase insensitive.
                 But if input string has an upper case character, then the search is case sensitive.
   1 Oct 22 -- Will not search further in a file if there's a null byte.  I also got this idea from ripgrep.  And I added more info to be displayed if verbose is set.
   2 Oct 22 -- The extension system is made mostly obsolete by null byte detection.  So the default will be *.  But I discovered when the files slice exceeds 1790 elements,
                 the go routines all deadlock, so the wait group is not exiting.

               Posted to gonuts using the go playground for the code: 10/2/22 @1:35 pm   go playground sharing link: https://go.dev/play/p/gIVVLsiTqod/
                 Moved location of the wait statement, as suggested by Jan Merci.  I guess both a waitgroup and a channel are used for the syncronization.
                 Nope, then I got a negative WaitGroup number panic.  I moved it back, for now.

               Looks like the error was the order of the defer and if err statements.  The way I first had it, defer was after the if err, so if there was a file error
                 (like the three access is denied errors I'm seeing from "My Videos", "My Music", and "MY Pictures") then wg.Done() would not be called.
                 So the wait group count would not go down to zero.  How subtle, and I needed help from someone else to notice that.
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

const LastAltered = "2 Oct 2022"
const maxSecondsToTimeout = 300

const workerPoolMultiplier = 20
const null = 0 // null rune to be used for strings.ContainsRune in GrepFile below.

var workers = runtime.NumCPU() * workerPoolMultiplier

type grepType struct {
	regex    *regexp.Regexp
	filename string
	goRtnNum int
}

var caseSensitiveFlag bool // default is false.
var grepChan chan grepType
var totFilesScanned, totMatchesFound int64
var t0, tfinal time.Time
var wg sync.WaitGroup

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)

	// flag definitions and processing
	globFlag := flag.Bool("g", false, "force use of globbing, only makes sense on Windows.") // Ptr
	verboseFlag := flag.Bool("v", false, "Verbose flag")
	var timeoutOpt = flag.Int64("timeout", maxSecondsToTimeout, "seconds (0 means no timeout)")
	//maxFiles := flag.Int64("max", 1000, "Maximum files to process.  Looking for why I'm getting a deadlock error.")
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
	files := args[1:]
	if len(files) < 1 { // no files or globbing pattern on command line.
		if runtime.GOOS == "windows" {
			//files = []string{"*.txt"}
			files = []string{"*"} // Now that files containing a null byte are skipped, I can default to every file in this directory.
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

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	if *verboseFlag {
		fmt.Printf(" Current working Directory is %s; %s was last linked %s.\n\n", workingDir, execName, LastLinkedTimeStamp)
	}

	if *globFlag && runtime.GOOS == "windows" { // glob function only makes sense on Windows.
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

	if *verboseFlag {
		fmt.Printf(" Length of files = %d, minGoRtns = %d.\n\n", len(files), minGoRtns)
	}
	//if len(files) > int(*maxFiles) {
	//	files = files[:*maxFiles]
	//	if *verboseFlag {
	//		fmt.Printf(" Length of files = %d.\n", len(files))
	//	}
	//}
	for _, file := range files {
		wg.Add(1)
		grepChan <- grepType{regex: lineRegex, filename: file}
	}

	goRtns := runtime.NumGoroutine()
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
	var localMatches int64
	var lineStrng string // either case sensitive or case insensitive string, depending on value of caseSensitiveFlag, which itself depends on case sensitivity of input pattern.
	file, err := os.Open(fpath)
	defer func() { // gonuts group: Matthew Zimmerman noticed that if there's a file error, wg.Done() isn't called.  I just fixed that.
		wg.Done()
		file.Close()
		atomic.AddInt64(&totFilesScanned, 1)
		atomic.AddInt64(&totMatchesFound, localMatches)
	}()
	if err != nil {
		log.Printf("grepFile os.Open error is: %s\n", err)
		return
	}

	reader := bufio.NewReader(file)
	for lino := 1; ; lino++ {
		lineStr, er := reader.ReadString('\n')
		if strings.ContainsRune(lineStr, null) {
			return // I guess break would do the same thing here, but using return is a clearer way to indicate my intent.  The wg.Done() is deferred so it doesn't matter.
		}
		if caseSensitiveFlag {
			lineStrng = lineStr
		} else {
			lineStrng = strings.ToLower(lineStr) // this is the change I made to make every comparison case insensitive.
		}

		if lineRegex.MatchString(lineStrng) { // this is now either case sensitive or not, depending on whether the input pattern has upper case letters.
			fmt.Printf("%s:%d:%s", fpath, lino, lineStr)
			localMatches++
		}
		if er != nil { // when can't read any more bytes, break.  The test for er is here so line fragments are processed, too.
			break // just exit when hit EOF condition.
		}
		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}
	}
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
		boolean, er := filepath.Match(pattern, strings.ToLower(d.Name()))
		if er != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if boolean {
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
