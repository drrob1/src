// Copyright (C) 2011-12 Qtrac Ltd.
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
  21 Mar 20 -- Another ack name change.  My plan is to reproduce the function of ack, but on windows not require
                 the complex installation that I cannot do at work.
                 I'll use multiple processes for the grep work.  For the dir walking I'll just do that in main.
  30 Mar 20 -- Started work on extracting the extensions from a slice of input filenames.  And will assume .txt extension if none is provided.
   1 Apr 20 -- Making it multi-threaded by using go routines by copying cgrepi.go and multimap.go.

               Now created multack.go, derived from anack.go.  It works, but is not faster than anack.  I need more go routines.
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
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

const lastAltered = "1 Apr 2020"

var workers = runtime.NumCPU()

type ResultType struct {
	filename string
	lino     int
	line     string
}

type Job struct {
	filename string
	results  chan<- ResultType
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

		linestr := string(line)
		linestr = strings.ToLower(linestr)
		linelowercase := []byte(linestr)

		if lineRegex.Match(linelowercase) {
			job.results <- ResultType{job.filename, lino, string(line)}
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
	var timeoutOpt *int = flag.Int("timeout", 0, "seconds < 240, where 0 means max timeout of 240 sec.")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 240 {
		log.Fatalln("timeout must be in the range [0,240] seconds")
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = 240
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

	extensions := make([]string, 0, 100)
	if flag.NArg() < 2 {
		extensions = append(extensions, ".txt")
	} else if runtime.GOOS == "linux" {
		files := args[1:]
		if len(files) > 1 {
			extensions = extractExtensions(files)
		}
	} else {
		extensions = args[1:]
		for i := range extensions {
			extensions[i] = strings.ReplaceAll(extensions[i], "*", "")
		}
	}

	startDirectory, _ := os.Getwd() // startDirectory is a string

	fmt.Println()
	fmt.Printf(" Another ack, written in Go.  Last altered %s, and will start in %s, pattern-%s, extensions=%v. \n\n\n ",
		lastAltered, startDirectory,pattern, extensions)

	DirAlreadyWalked := make(map[string]bool, 500)
	DirAlreadyWalked[".git"] = true // ignore .git and its subdir's

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	// goroutine to collect results from resultsChan
	doneChan := make(chan bool)
	resultsChan := make(chan ResultType, 1024)
	go func() {
		for r := range resultsChan {
			 fmt.Printf(" %s:%d:%s\n",r.filename, r.lino, r.line)
		}
		doneChan <- true
	}()


	// walkfunc closure that I hope is parallel.  I stopped here
	filepathwalkfunction := func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf(" Error from walk is %v. \n ", err)
			return nil
		}

		if fi.IsDir() {
			if DirAlreadyWalked[fpath] {
				return filepath.SkipDir
			} else {
				DirAlreadyWalked[fpath] = true
			}
		} else if fi.Mode().IsRegular() {
			for _, ext := range extensions {
				if strings.HasSuffix(fpath, ext) { // only search thru indicated extensions.  Especially not thru binary or swap files.
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

	err = filepath.Walk(startDirectory, filepathwalkfunction)
	close(resultsChan)

	if err != nil {
		log.Fatalln(" Error from filepath.walk is", err, ".  Elapsed time is", time.Since(t0))
	}

	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed)
	fmt.Println()
} // end main

// resultsChan := make(chan ResultType, 1024)
func grepFile(lineRegex *regexp.Regexp, fpath string, resultChan chan ResultType) {
	file, err := os.Open(fpath)
	if err != nil {
		log.Printf("grepFile os.Open error : %s\n", err)
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
			var r ResultType
			r.filename = fpath
			r.lino = lino
			r.line = string(line)
			resultChan <- r  // I think this is what makes this a concurrent walk function.
			// fmt.Printf("%s:%d:%s \n", fpath, lino, string(line)) from orig code
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("error from reader.ReadBytes in grepfile:%d: %s\n", lino, err)
			}
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
			if strings.EqualFold(extensions[i-1], extensions[i]) {
				extensions[i-1] = ""  // This needs to be [i-1] because when it was [i] it interferred w/ the next iteration.
			}
		}
		//fmt.Println(" in extractExtensions before sort:", extensions)
		sort.Sort(sort.Reverse(extensions))
		// sort.Sort(sort.Reverse(sort.IntSlice(s)))
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
