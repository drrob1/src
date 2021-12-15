// Copyright © 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the
// Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License
// at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The approach taken here was inspired by an example on the gonuts mailing
// list by Roger Peppe.
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
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

const LastAltered = "15 Dec 2021"
const maxSecondsToTimeout = 300
const workerPoolMultiplier = 20

var workers = runtime.NumCPU() * workerPoolMultiplier // this works very well in multack

type Result struct {
	filename string
	lino     int
	line     string
}

type Job struct {
	filename string
	results  chan<- Result
}

type grepType struct {
	regex    *regexp.Regexp
	filename string
	goRtnNum int
}

var grepChan chan grepType
var totFilesScanned, totMatchesFound int64

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
	if len(files) < 1 {
		if runtime.GOOS == "windows" {
			files = append(files, "*.txt")
		} else {
			log.Fatalln(" A file must be given on linux.")
		}

	}

	// start the worker pool
	grepChan = make(chan grepType, workers) // buffered channel
	for w := 0; w < workers; w++ {
		go func() {
			for g := range grepChan { // These are channel reads that are only stopped when the channel is closed.
				grepFile(g.regex, g.filename)
			}
		}()
	}

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)
	//fmt.Println(" t0 is", t0, "and tfinal is", tfinal, "and timeoutOpt is", *timeoutOpt)
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

	for _, file := range files {
		grepChan <- grepType{regex: lineRegex, filename: file}
		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}
	}

	goRtns := runtime.NumGoroutine() // must capture this before we sleep for a second.
	close(grepChan)                  // must close the channel so the worker go routines know to stop.

	elapsed := time.Since(t0)
	time.Sleep(time.Second) // I've noticed that sometimes main exits before everything can be displayed.  This sleep line fixes that.

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
	}
} // end grepFile
