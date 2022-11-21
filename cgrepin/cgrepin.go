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

               First reported to me by Matthew Zimmerman.
               Looks like the error was the order of the defer and if err statements.  The way I first had it, defer was after the if err, so if there was a file error
                 (like the three access is denied errors I'm seeing from "My Videos", "My Music", and "MY Pictures") then wg.Done() would not be called.
                 So the wait group count would not go down to zero.  How subtle, and I needed help from someone else to notice that.

               Andrew Harris noticed that the condition for closing the channel could be when all work is sent into it.  I was closing the channel after all work was done.
                 So I changed that and noticed that it's still possible for the main routine to finish before some of the last grepFile calls.  I still need the WaitGroup.
   5 Oct 22 -- Based on output from ripgrep, I want all the matches from the same file to be displayed near one another.  So I have to output them to the same slice and then sort that.
   7 Oct 22 -- Added color to output.
  26 Oct 22 -- Now called cgrepin.go, as it will grep and sort from Stdin.  I expect that I don't need go routines for this, as it's only one stream of input.
                 I added my own method to a system type by embedding it into my own type.  IE, myBytesReader absorbed *bytes.Reader.
                 Interestingly, this is ~5% faster w/ the concurrent code in it, when compared to cgrepin2 in which I removed all the current stuff.  So I guess
                 the current stuff is slightly faster after all.
  20 Nov 22 -- static linter found an issue w/ const null, so I removed it.
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

const LastAltered = "20 Nov 2022"
const maxSecondsToTimeout = 300

const minMatches = 100

//const null = 0 // null rune to be used for strings.ContainsRune in GrepFile below.  But not in this cgrepin version, only in cgrepi.  So I'll remove it.

var caseSensitiveFlag bool // default is false.
//var grepChan chan grepType
//var matchChan chan matchType
var totMatchesFound int64
var t0, tfinal time.Time
var sliceOfStrings []string

type myBytesReader struct {
	*bytes.Reader // embedded system field so I can add my own method to it.
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)

	// flag definitions and processing
	verboseFlag := flag.Bool("v", false, "Verbose flag")
	var timeoutOpt = flag.Int64("timeout", maxSecondsToTimeout, "seconds (0 means no timeout)")
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

	t0 = time.Now()
	tfinal = t0.Add(time.Duration(*timeoutOpt) * time.Second)
	lineRegex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}

	fmt.Println()
	gotWin := runtime.GOOS == "windows"
	ctfmt.Printf(ct.Yellow, gotWin, " Concurrent grep Stdin last altered %s, using pattern of %q, compiled with %s. \n",
		LastAltered, pattern, runtime.Version())
	fmt.Println()

	//workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	if *verboseFlag {
		fmt.Printf(" %s was last linked %s.\n\n", execName, LastLinkedTimeStamp)
	}

	results := make(chan string, minMatches)
	done := make(chan bool)                        // unbuffered to be a blocking channel
	sliceOfStrings = make([]string, 0, minMatches) // this uses an anonymous type.
	go func() {                                    // start the receiving operation before the sending starts
		for result := range results {
			fmt.Printf(" %s\n", result)                     // print the raw matching strings; note that I removed the trailing \n characters.
			sliceOfStrings = append(sliceOfStrings, result) // then sort them.  I'll decide which to show soon enough.
		}
		sort.Strings(sliceOfStrings)
		done <- true
	}()

	grepStdin(lineRegex, results)
	close(results) // must close the channel so the for range results will stop.
	<-done         // this is a blocking read from channel operation.

	elapsed := time.Since(t0)

	ctfmt.Printf(ct.Green, gotWin, "\n Elapsed time is %s to find and sort %d matches in os.Stdin\n", elapsed, totMatchesFound)
	fmt.Println()

	// Time to sort and show
	// sort.Strings(sliceOfStrings)  already done in the go rtn

	if *verboseFlag {
		fmt.Printf("\n Sorted output:\n")
		for _, s := range sliceOfStrings {
			fmt.Printf("%s\n", s)
		}
		fmt.Println()
	}

	ctfmt.Printf(ct.Yellow, gotWin, "\n There were %d matches were found in os.Stdin, that took %s.\n", totMatchesFound, elapsed)
}

func grepStdin(lineRegex *regexp.Regexp, result chan string) {
	var localMatches int64
	var lineString string // either case sensitive or case insensitive string, depending on value of caseSensitiveFlag, which itself depends on case sensitivity of input pattern.
	fileContents, err := io.ReadAll(os.Stdin)
	defer func() { // gonuts group: Matthew Zimmerman noticed that if there's a file error, wg.Done() isn't called.  I just fixed that.
		//	wg.Done()
		//	file.Close()
		//	atomic.AddInt64(&totFilesScanned, 1)
		atomic.AddInt64(&totMatchesFound, localMatches)
	}()
	if err != nil {
		log.Printf("grepStdin io.ReadAll error is: %s\n", err)
		return
	}

	reader := myBytesReader{bytes.NewReader(fileContents)} // by using the structured literal syntax it looks like this is working.  I added a new method to system type.
	for lino := 1; ; lino++ {
		lineStr, er := reader.readLine() // lineStr is terminated w/ the \n character.  I called a trim function and removed it.
		if er != nil {                   // when can't read any more bytes, break.  If any bytes were read, er == nil.
			break // just exit when hit EOF condition.
		}
		if caseSensitiveFlag {
			lineString = lineStr
		} else {
			lineString = strings.ToLower(lineStr) // this is the change I made to make every comparison case insensitive.
		}

		if lineRegex.MatchString(lineString) { // this is now either case sensitive or not, depending on whether the input pattern has upper case letters.
			//fmt.Printf("%s:%d:%s", fpath, lino, lineStr)  Will now only see the sorted output.
			localMatches++
			result <- lineString
		}
		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}
	}
} // end grepStdin

//func (r *bytes.Reader) readLine() (string, error) {  I think I just tripped over the fact that I can't add methods to an established system type.  I can only add methods to my own type, which can embed a system type.
// readLine will behave like it's similarly named functions on an io.Reader.  Turns out that there is no such function for a bytes.Buffer or bytes.Reader.
func (r *myBytesReader) readLine() (string, error) {
	var strBuf strings.Builder

	for {
		byt, err := r.ReadByte()
		if err != nil {
			return strings.TrimSpace(strBuf.String()), err
		}
		if byt == '\n' {
			return strings.TrimSpace(strBuf.String()), nil
		}
		if byt == '\r' {
			continue
		}
		err = strBuf.WriteByte(byt)
		if err != nil {
			return strings.TrimSpace(strBuf.String()), err
		}
	}
}
