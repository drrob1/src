// Copyright (C) 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and limitations under the License.

// The approach taken here was inspired by an example on the golang-nuts mailing list by Roger Peppe.
// I'm not sure the comment about Roger Peppe is for this version; it's likely for the concurrent version I have in multack.go.
/*
  REVISION HISTORY
  ----------------
  20 Mar 20 -- Made comparisons case-insensitive.  And decided to make this cgrepi.go.
                 And then I figured I could not improve performance by using more packages.
                 But I can change the side effect of displaying altered case.
  21 Mar 20 -- Another ack name change.  My plan is to reproduce the function of ack, but on windows not require
                 the complex installation that I cannot do at work.
                 I'll use multiple processes for the grep work.  For the dir walking I'll just do that in main.
  30 Mar 20 -- Started work on extracting the extensions from a slice of input filenames.  And will assume .txt extension if none is provided.
   1 Apr 20 -- Moved the regexp compile line out of main loop.
   7 Dec 21 -- All of the changes since Apr 2020 have been in multack.  I'm backporting a change to not track which dir have been entered, as the library will do that.
                 And I redid the walk closure to remove test for regular file.  The walk does not follow symlinks so this is not needed, either.
                 Starting w/ Go 1.16, there is a new walk function, that does not use a FiloInfo but a dirEntry, which they claim is faster.  I'll try it.
   8 Dec 21 -- Removing the test for .git.  It seems that the walk function knows not to enter .git.
  10 Dec 21 -- Nevermind.  I'm testing for .git and will skipdir if found.  And will simply return on IsDir
  13 Dec 21 -- Adding a total number of files scanned, and number of matches found.
  13 Nov 22 -- Adding some code that I developed for since and is now in multack.  Removed extensions stuff.  Only multack has smart case.
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
	"strings"
	"time"
)

const lastAltered = "13 Nov 2022"

const nullRune = 0

type devID uint64

var totFilesScanned, totMatchesFound int
var verboseFlag bool

func main() {
	log.SetFlags(0)
	var timeoutOpt *int = flag.Int("timeout", 0, "seconds < 240, where 0 means max timeout of 300 sec.")
	flag.BoolVar(&verboseFlag, "v", false, "verbose flag")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 300 {
		log.Fatalln("timeout must be in the range [0,300] seconds")
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = 300
	}

	if flag.NArg() < 1 {
		log.Fatalln("a regexp to match must be specified")
	}

	startDirectory, _ := os.Getwd() // startDirectory is a string

	pattern := flag.Arg(0)
	pattern = strings.ToLower(pattern)
	var lineRegex *regexp.Regexp
	var err error
	if lineRegex, err = regexp.Compile(pattern); err != nil {
		log.Fatalf("invalid regexp: %s\n", err)
	}
	/*
		extensions := make([]string, 0, 100)
		if flag.NArg() < 2 {
			extensions = append(extensions, ".txt")
		} else if runtime.GOOS == "linux" {
			files := flag.Args()
			files = files[1:]
			extensions = extractExtensions(files)
		} else { // windows branch
			extensions = args[1:]
			for i := range extensions {
				extensions[i] = strings.ReplaceAll(extensions[i], "*", "")
			}
		}
	*/
	fmt.Println()
	fmt.Printf(" Another ack, written in Go.  Last altered %s, and will start in %s, pattern=%s, extensions are gone. \n\n\n ",
		lastAltered, startDirectory, pattern)

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)
	rootFileInfo, err := os.Stat(startDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Exiting\n", startDirectory, err)
		os.Exit(1)
	}
	rootDeviceID := getDeviceID(rootFileInfo)

	// walkfunc closures.  Only the last one is being used.
	/*
		var filepathwalkfunction filepath.WalkFunc = func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf(" Error from walk is %v. \n ", err)
				return nil
			}

			if fi.IsDir() {
				//if DirAlreadyWalked[fpath] { return filepath.SkipDir } else { DirAlreadyWalked[fpath] = true }
				if dirToSkip[fpath] {
					return filepath.SkipDir
				}
			} else if fi.Mode().IsRegular() {
				for _, ext := range extensions {
					if strings.HasSuffix(fpath, ext) { // only search thru indicated extensions.  Especially not thru binary or swap files.
						grepFile(lineRegex, fpath)

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
	*/
	/*
		var filepathwalkfunction filepath.WalkFunc = func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf(" Error from walk is %v. \n ", err)
				return nil
			}

			if dirToSkip[fpath] {
				return filepath.SkipDir
			}

			for _, ext := range extensions {
				if strings.HasSuffix(fpath, ext) { // only search thru indicated extensions.  Especially not thru binary or swap files.
					grepFile(lineRegex, fpath)
				}
			}

			//log.Println(" Need to debug this.  Filepath is", fpath, ", fi is", fi.Name(), fi.IsDir())
			now := time.Now()
			if now.After(tfinal) {
				log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
			}
			return nil
		}

		err = filepath.Walk(startDirectory, filepathwalkfunction)
	*/

	walkDirFunction := func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			if verboseFlag {
				fmt.Printf(" Error from walk is %v.  Skipping \n ", err)
			}
			return filepath.SkipDir
		}

		if d.IsDir() {
			if filepath.Ext(fpath) == ".git" || strings.Contains(fpath, "vmware") {
				return filepath.SkipDir
			}
			return nil
		}
		/*
			for _, ext := range extensions {
				if strings.HasSuffix(fpath, ext) { // only search thru indicated extensions.  Especially not thru binary or swap files.
					grepFile(lineRegex, fpath)
				}
			}
		*/

		info, err := d.Info()
		if err != nil {
			return err
		}
		deviceID := getDeviceID(info)
		if rootDeviceID != deviceID { // don't follow symlinks onto other devices like bigbkupG or DSM.
			return filepath.SkipDir
		}
		grepFile(lineRegex, fpath)
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

	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed, "and total matches found is", totMatchesFound, "in", totFilesScanned, "files scanned.")
	fmt.Println()
} // end main

func grepFile(lineRegex *regexp.Regexp, fpath string) {
	file, err := os.Open(fpath)
	if err != nil {
		log.Printf("grepFile os.Open error : %s\n", err)
		return
	}
	defer file.Close()
	totFilesScanned++
	reader := bufio.NewReader(file)
	for lino := 1; ; lino++ {
		line, er := reader.ReadString('\n')
		// line = strings.TrimSpace(line)  I'm going to try without this.

		if strings.ContainsRune(line, nullRune) { // skip binary files
			return
		}

		// this is the change I made to make every comparison case-insensitive.
		lineStrLower := strings.ToLower(line)

		if lineRegex.MatchString(lineStrLower) {
			fmt.Printf("%s:%d:%s", fpath, lino, line)
			totMatchesFound++
		}
		if er != nil {
			//if er != io.EOF {  This became messy, so I'm removing it.
			//	log.Printf("error from reader.ReadString in grepfile %s line %d: %s\n", fpath, lino, err)
			//}
			break // just exit when hit any error condition.
		}
	}
} // end grepFile

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
			if strings.EqualFold(extensions[i-1], extensions[i]) {
				extensions[i-1] = "" // This needs to be [i-1] because when it was [i] it interferred w/ the next iteration.
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
*/

