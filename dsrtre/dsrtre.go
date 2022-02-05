/*
dsrtre.go
  REVISION HISTORY
  ----------------
   1 Apr 20 -- dsrt recursive, named dsrtr.go.
   2 Apr 20 -- Tracking down bug of not finding .pdf files, and probably also not finding .epub or .mobi
                 Turned out to be case sensitivity in the comparisons.
  17 Aug 20 -- I'm using this way more than I expected.  And it's slower than I expected.  I'm going to take a stab at
                 multitasking here.
  19 Aug 20 -- Made timeout 15 min by default, max of 30 min.  4 min was too short on win10 machine.
                 This forked from dsrtr and now called dsrtre as it takes a regular expression.
                 Changed option to -t instead of -timeout, as I never remembered its name.
  20 Aug 20 -- Will write errors to os.Stderr.  Changed how default timeout is set.
  23 Aug 20 -- Make sure a newline is displayed after the error message.
   5 Sep 20 -- Don't follow symlinked directories
   4 Feb 22 -- Updated code, removing the concurrency pattern as it's not needed.  And removing the tracking of directories visited.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const lastAltered = "4 Feb 2022"

func main() {
	var timeoutOpt *int = flag.Int("t", 900, "seconds < 1800, where 0 means timeout of 900 sec.")
	var verboseFlag = flag.Bool("v", false, "enter a verbose testing mode to println more variables")
	var inputRegexPattern, startDir string
	var inputRegex *regexp.Regexp
	var err error

	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 1800 {
		log.Println("timeout must be in the range [0..1800] seconds.  Set to 900")
		*timeoutOpt = 900
	}

	if flag.NArg() == 0 {
		fmt.Print(" Enter regex: ")
		fmt.Scanln(&inputRegexPattern)
		startDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, " Getwd returned this error: %v\n", err)
			os.Exit(1)
		}

	} else if flag.NArg() == 1 {
		inputRegexPattern = flag.Arg(0)
		startDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, " Getwd returned this error: %v\n", err)
			os.Exit(1)
		}
	} else {
		inputRegexPattern = flag.Arg(0)
		startDir = flag.Arg(1)
	}

	inputRegexPattern = strings.ToLower(inputRegexPattern)
	inputRegex, err = regexp.Compile(inputRegexPattern)
	if err != nil {
		log.Fatalln(" error from regex compile function is ", err)
	}

	fmt.Println()
	fmt.Printf(" dsrtre (recursive), written in Go.  Last altered %s, will use regex of %q and will start in %s. \n", lastAltered, inputRegex.String(), startDir)
	fmt.Println()
	if *verboseFlag { // I don't really have anything for verbose mode yet.  I'll have to think of something.
		fmt.Println()
	}

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	/*	// walkfunc closure
		filepathwalkfunction := func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from walk is %v. \n", err)
				return nil
			}

			if fi.IsDir() {
				if DirAlreadyWalked[fpath] {
					return filepath.SkipDir
				} else {
					DirAlreadyWalked[fpath] = true
				}
			} else if isSymlink(fi.Mode()) && fi.IsDir() {
				if runtime.GOOS == "linux" {
					for _, fp := range args {
						fp = strings.ToLower(fp)
						NAME := strings.ToLower(fi.Name())
						if BOOL := pattern.MatchString(NAME); BOOL {
							var r ResultType
							s := fi.ModTime().Format("Jan-02-2006_15:04:05")
							r.path = fpath
							r.datestamp = s
							r.sizeint = int(fi.Size()) // fi.Size() is an int64
							resultsChan <- r
						}
					}
				} else if runtime.GOOS == "windows" {
					NAME := strings.ToLower(fi.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?
					if BOOL := pattern.MatchString(NAME); BOOL {
						var r ResultType
						s := fi.ModTime().Format("Jan-02-2006_15:04:05")
						r.path = fpath
						r.datestamp = s
						r.sizeint = int(fi.Size())
						resultsChan <- r
					}
				}
				now := time.Now()
				if now.After(tfinal) {
					log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
				}
			}
			return nil
		}

		er := filepath.Walk(startDirectory, filepathwalkfunction)
	*/
	filepathWalkDirEntry := func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from walk is %v. \n ", err)
			return nil
		}

		if d.IsDir() && fpath == ".git" {
			return filepath.SkipDir
		} else if isSymlink(d.Type()) {
			fmt.Printf(" %s is a symlink, name is %s. \n", fpath, d.Name())
			//return filepath.SkipDir
		}

		// Must be a regular file
		NAME := strings.ToLower(d.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?
		if BOOL := inputRegex.MatchString(NAME); BOOL {
			fi, er := d.Info()
			if er != nil {
				fmt.Fprintf(os.Stderr, " %s.Info() call error is %v\n", d.Name())
				return er
			}
			t := fi.ModTime().Format("Jan-02-2006_15:04:05")
			sizeStr := strconv.Itoa(int(fi.Size()))
			if fi.Size() > 100_000 {
				sizeStr = AddCommas(sizeStr)
			}

			fmt.Printf("%15s %s %s\n", sizeStr, t, fpath)
		}

		now := time.Now()
		if now.After(tfinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}

		return nil
	}

	err = filepath.WalkDir(startDir, filepathWalkDirEntry)
	if err != nil {
		log.Fatalln(" Error from filepath.walk is", err, ".  Elapsed time is", time.Since(t0))
	}

	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed)
	fmt.Println()
} // end main

//-------------------------------------------------------------------- InsertByteSlice

func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas

// ---------------------------- GetIDname -----------------------------------------------------------
func GetIDname(uidStr string) string {

	if len(uidStr) == 0 {
		return ""
	}
	ptrToUser, err := user.LookupId(uidStr)
	if err != nil {
		panic("uid not found")
	}

	idname := ptrToUser.Username
	return idname

} // GetIDname

// ------------------------------ isSymlink ---------------------------
func isSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink
