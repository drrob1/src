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
   1 Apr 20 -- Change some variable names to include where they are channels, and lineRx to LineRegex
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const LastAltered = "1 Apr 2020"

var workers = runtime.NumCPU()

type Result struct {
	filename string
	lino     int
	line     string
}

type Job struct {
	filename string
	results  chan<- Result
}

func (job Job) Do(lineRegex *regexp.Regexp) {
	file, err := os.Open(job.filename)
	if err != nil {
		log.Printf("error: %s\n", err)
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for lino := 1; ; lino++ {
		line, err := reader.ReadBytes('\n')
		line = bytes.TrimRight(line, "\n\r")

		// this is the change I made to make every comparison case insensitive.  Side effect of output is not original case.
		linestr := string(line)
		linestr = strings.ToLower(linestr)
		linelowercase := []byte(linestr)

		if lineRegex.Match(linelowercase) {
			job.results <- Result{job.filename, lino, string(line)}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("error:%d: %s\n", lino, err)
			}
			break
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	var timeoutOpt *int64 = flag.Int64("timeout", 0, "seconds (0 means no timeout)")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 240 {
		log.Fatalln("timeout must be in the range [0,240] seconds")
	}
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("a regexp to match must be specified")
	}
	pattern := args[0]
	pattern = strings.ToLower(pattern) // this is the change for the pattern.
	files := args[1:]
	if len(files) < 1 {
		log.Fatalln("must provide at least one filename")
	}
	t0 := time.Now()
	if lineRegex, err := regexp.Compile(pattern); err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	} else {
		fmt.Println()
		fmt.Printf(" Concurrent grep insensitive case last altered %s. \n", LastAltered)
		fmt.Println()
		var timeout int64 = 1e9 * 60 * 10 // 10 minutes!
		if *timeoutOpt != 0 {
			timeout = *timeoutOpt * 1e9
		}
		grep(timeout, lineRegex, commandLineFiles(files)) // this fails vet because it's in the platform specific code files.
	}
	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed)
	fmt.Println()
}

func grep(timeout int64, lineRegex *regexp.Regexp, filenames []string) {
	jobsChan := make(chan Job, workers)
	resultsChan := make(chan Result, minimum(1000, len(filenames)))
	doneChan := make(chan struct{}, workers)

	go addJobs(jobsChan, filenames, resultsChan)
	for i := 0; i < workers; i++ {
		go doJobs(doneChan, lineRegex, jobsChan)
	}
	waitAndProcessResults(timeout, doneChan, resultsChan)
}

func addJobs(jobsChan chan<- Job, filenames []string, resultsChan chan<- Result) {
	for _, filename := range filenames {
		jobsChan <- Job{filename, resultsChan}
	}
	close(jobsChan)
}

func doJobs(doneChan chan<- struct{}, lineRegex *regexp.Regexp, jobsChan <-chan Job) {
	for job := range jobsChan {
		job.Do(lineRegex)
	}
	doneChan <- struct{}{}
}

func waitAndProcessResults(timeout int64, doneChan <-chan struct{}, resultsChan <-chan Result) {
	finish := time.After(time.Duration(timeout))
	for working := workers; working > 0; {
		select { // Blocking
		case result := <-resultsChan:
			fmt.Printf("%s:%d:%s\n", result.filename, result.lino,
				result.line)
		case <-finish:
			fmt.Println("timed out")
			return // Time's up so finish with what results there were
		case <-doneChan:
			working--
		}
	}
	for {
		select { // Nonblocking
		case result := <-resultsChan:
			fmt.Printf("%s:%d:%s\n", result.filename, result.lino, result.line)
		case <-finish:
			fmt.Println("timed out")
			return // Time's up so finish with what results there were
		default:
			return
		}
	}
}

func minimum(x int, ys ...int) int {
	for _, y := range ys {
		if y < x {
			x = y
		}
	}
	return x
}
