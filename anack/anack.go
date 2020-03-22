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
  21 Mar 20 -- Another ack name change.  My plan is to reproduce the function of ack, but on windows not require
                 the complex installation that I cannot do at work.
                 I'll use multiple processes for the grep work.  For the dir walking I'll just do that in main.
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
	"strings"
	"time"
)

const LastAltered = "21 Mar 2020"

type Result struct {
	filename string
	lino     int
	line     string
}

func main() {
	//	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
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
	extensions := args[1:]
	if len(extensions) < 1 {
		log.Fatalln("must provide at least one filename extension")
	}
	for i, ext := range extensions { // validate extensions, as this is likely forgotten to be needed.
		if ! strings.ContainsAny(ext, ".") {
			extensions[i] = "." + ext
			fmt.Println(" Added dot to extension to give", extensions[i])
		}
	}

	for _, ext := range extensions {
		if len(ext) != 4 {
			fmt.Println(" Need dotted extensions only.  Not filenames, not wildcards.  A missing dot will be prepended.  Is", ext,"an extension?")
			fmt.Print(" Proceed? ")
			ans := ""
			_, err := fmt.Scanln(&ans)
			if err != nil {
				log.Fatalln(" Error from ScanLn.  It figures.", err)
			}
			ans = strings.ToUpper(ans)
			if ! strings.Contains(ans,"Y") {
				os.Exit(1)
			}
		}
	}

//	for _, ext := range extensions {   It works, so I can remove this.
//		fmt.Println(" debug for dot ext.  Ext is ", ext)
//	}

	startDirectory, _ := os.Getwd() // startDirectory is a string
	fmt.Println()
	fmt.Printf(" Another ack, written in Go.  Last altered %s, and will start in %s.", LastAltered, startDirectory)
	fmt.Println()
	fmt.Println()
	DirAlreadyWalked := make(map[string]bool, 500)
	DirAlreadyWalked[".git"] = true // ignore .git and its subdir's

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)
	// walkfunc closure
	var filepathwalkfunction filepath.WalkFunc = func(fpath string, fi os.FileInfo, err error) error {
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
					if lineRx, err := regexp.Compile(pattern); err != nil { // this is the regex compile line.
						log.Fatalf("invalid regexp: %s\n", err)
					} else {
						//fullname := fpath + string(filepath.Separator) + fi.Name()  Turns out that fpath is the full file name path.
						grepFile(lineRx, fpath)
					}
				}
			}
		}
		//log.Println(" Need to debug this.  Filepath is", fpath, ", fi is", fi.Name(), fi.IsDir())
		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}
		return nil
	}

	err := filepath.Walk(startDirectory, filepathwalkfunction)

	if err != nil {
		log.Fatalln(" Error from filepath.walk is", err, ".  Elapsed time is", time.Since(t0))
	}

	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed)
	fmt.Println()
} // end main

func grepFile(lineRx *regexp.Regexp, fpath string) {
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

		if lineRx.Match(linelowercase) {
			fmt.Printf("%s:%d:%s \n", fpath, lino, string(line))
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("error from reader.ReadBytes in grepfile:%d: %s\n", lino, err)
			}
			break // just exit when hit EOF condition.
		}
	}
} // end grepFile
