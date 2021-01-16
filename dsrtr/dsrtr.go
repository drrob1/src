/*
dsrtr.go
  REVISION HISTORY
  ----------------
   1 Apr 20 -- dsrt recursive, named dsrtr.go.
   2 Apr 20 -- Tracking down bug of not finding .pdf files, and probably also not finding .epub or .mobi
                 Turned out to be case sensitivity in the comparisons.
  17 Aug 20 -- I'm using this way more than I expected.  And it's slower than I expected.  I'm going to take a stab at
                 multitasking here.
  19 Aug 20 -- Made timeout 15 min by default, max of 30 min.  4 min was too short on win10 machine.
                 And made t as an option name for timeout.
  20 Aug 20 -- Will write errors to os.Stderr.  And changed how the default timeout is set.
   5 Sep 20 -- Will look to not follow symlinks
  20 Dec 20 -- Looking to change sort functions based on time to be idiomatic, but there aren't any here.  Go figure.
                 I did remove some dead comments, though.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const lastAltered = "20 Dec 2020"

type ResultType struct {
	path      string
	datestamp string
	sizeint   int
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	//var timeoutOpt *int = flag.Int("timeout", 0, "seconds < 1800, where 0 means timeout of 900 sec.")
	var timeoutOpt *int = flag.Int("t", 900, "seconds < 1800, where 0 means timeout of 900 sec.")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 1800 {
		log.Println("timeout must be in the range [0..1800] seconds.  Making default of 900")
		*timeoutOpt = 900
	}

	args := flag.Args()

	if len(args) < 1 {
		log.Fatalln("a globbing pattern to match must be specified")
	} else if len(args) == 1 {
		//pattern = strings.ToLower(pattern)
		//fmt.Println(" pattern=", pattern)
	} else {
		// I cannot think of anything to put here at the moment.  I'll say that args must be a slice of strings of filenames, and on linux.
	}

	pattern := strings.ToLower(args[0])

	startDirectory, _ := os.Getwd() // startDirectory is a string
	fmt.Println()
	fmt.Printf(" dsrtr (recursive), written in Go.  Last altered %s, will use globbing pattern of %q and will start in %s. \n", lastAltered, pattern, startDirectory)
	fmt.Println()
	fmt.Println()
	DirAlreadyWalked := make(map[string]bool, 500)
	DirAlreadyWalked[".git"] = true // ignore .git and its subdir's

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	// goroutine to collect results from resultsChan
	doneChan := make(chan bool)
	resultsChan := make(chan ResultType, 100_000)
	go func() {
		for r := range resultsChan {
			sizestr := strconv.Itoa(r.sizeint)
			if r.sizeint > 100000 {
				sizestr = AddCommas(sizestr)
			}
			fmt.Printf("%15s %s %s\n", sizestr, r.datestamp, r.path)
		}
		doneChan <- true
	}()

	// walkfunc closure
	filepathwalkfunction := func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from walk is %v. \n ", err)
			return nil
		}

		if fi.IsDir() {
			if DirAlreadyWalked[fpath] {
				return filepath.SkipDir
			} else {
				DirAlreadyWalked[fpath] = true
			}
		} else if isSymlink(fi.Mode()) && fi.IsDir() { // don't follow symlinked directories
			return filepath.SkipDir
		} else /* if fi.Mode().IsRegular()  */ {
			if runtime.GOOS == "linux" {
				for _, fp := range args {
					fp = strings.ToLower(fp)
					NAME := strings.ToLower(fi.Name())
					if BOOL, _ := filepath.Match(fp, NAME); BOOL {
						var r ResultType
						s := fi.ModTime().Format("Jan-02-2006_15:04:05")
						//r.filename = NAME
						r.path = fpath
						r.datestamp = s
						r.sizeint = int(fi.Size()) // fi.Size() is an int64
						//r.fileinfo = fi
						resultsChan <- r
					}
				}
			} else if runtime.GOOS == "windows" {
				NAME := strings.ToLower(fi.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?
				if BOOL, _ := filepath.Match(pattern, NAME); BOOL {
					var r ResultType
					s := fi.ModTime().Format("Jan-02-2006_15:04:05")
					//r.filename = NAME
					r.path = fpath
					r.datestamp = s
					r.sizeint = int(fi.Size())
					//r.fileinfo = fi
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

	err := filepath.Walk(startDirectory, filepathwalkfunction)
	if err != nil {
		log.Fatalln(" Error from filepath.walk is", err, ".  Elapsed time is", time.Since(t0))
	}

	close(resultsChan)
	<-doneChan

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
