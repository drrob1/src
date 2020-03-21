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

  20 Mar 20 -- Now a separate module called gogrep.  I am going to try to use dirmap but run gogrep on each walked dir.

 */
package gogrep

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const LastAltered  = "20 Mar 2020"

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

func (job Job) Do(lineRx *regexp.Regexp) {
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

		if lineRx.Match(linelowercase) {
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

// Gogrep is the library version of cgrep3.  pattern and files are obvious.  timeoutOpt of 0 becomes 10 minutes, the max.
func Gogrep(pattern string, files []string, timeoutOpt int) error {

	var timeoutopt int64 = int64(timeoutOpt)
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores

	if timeoutOpt < 0 || timeoutOpt > 240 {
		timeoutOpt = 240
	}

	if len(pattern) < 1 {
//		log.Fatalln("a regexp to match must be specified")
		e := errors.New("regexp must be specified")
		return e
	}

	if len(files) < 1 {
		var e error = errors.New("files must be defined") // I'll leave it to the calling rtn to expand the directory.  to be continued ...
		return e
	}

	if lineRx, err := regexp.Compile(pattern); err != nil {
		return err
	} else {
		var timeout int64 = 1e9 * 60 * 10 // 10 minutes!
		if timeoutOpt != 0 {
			timeout = timeoutopt * 1e9
		}
		grep(timeout, lineRx, commandLineFiles(files))  // this fails vet because it's in the platform specific code files.
	}
	return nil
}

func grep(timeout int64, lineRx *regexp.Regexp, filenames []string) {
	jobs := make(chan Job, workers)
	results := make(chan Result, minimum(1000, len(filenames)))
	done := make(chan struct{}, workers)

	go addJobs(jobs, filenames, results)
	for i := 0; i < workers; i++ {
		go doJobs(done, lineRx, jobs)
	}
	waitAndProcessResults(timeout, done, results)
}

func addJobs(jobs chan<- Job, filenames []string, results chan<- Result) {
	for _, filename := range filenames {
		jobs <- Job{filename, results}
	}
	close(jobs)
}

func doJobs(done chan<- struct{}, lineRx *regexp.Regexp, jobs <-chan Job) {
	for job := range jobs {
		job.Do(lineRx)
	}
	done <- struct{}{}
}

func waitAndProcessResults(timeout int64, done <-chan struct{},
	results <-chan Result) {
	finish := time.After(time.Duration(timeout))
	for working := workers; working > 0; {
		select { // Blocking
		case result := <-results:
			fmt.Printf("%s:%d:%s\n", result.filename, result.lino,
				result.line)
		case <-finish:
			fmt.Println("timed out")
			return // Time's up so finish with what results there were
		case <-done:
			working--
		}
	}
	for {
		select { // Nonblocking
		case result := <-results:
			fmt.Printf("%s:%d:%s\n", result.filename, result.lino,
				result.line)
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
