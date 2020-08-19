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
	"runtime"
	"strconv"
	"strings"
	"time"
)

const lastAltered = "19 Aug 2020"

type ResultType struct {
	// filename  string  Not needed, AFAICT (as far as I can tell)
	path      string
	datestamp string
	sizeint   int
	// fileinfo  os.FileInfo  Not needed, AFAICT
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	//var timeoutOpt *int = flag.Int("timeout", 0, "seconds < 1800, where 0 means timeout of 900 sec.")
	var timeoutOpt *int = flag.Int("t", 0, "seconds < 1800, where 0 means timeout of 900 sec.")
	//var testFlag = flag.Bool("test", false, "enter a testing mode to println more variables")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 1800 {
		log.Println("timeout must be in the range [0..1800] seconds")
		*timeoutOpt = 0
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = 900
	}

	args := flag.Args()

	if len(args) < 1 {
		log.Fatalln("a regex to match must be specified")
	}

	inputpattern := strings.ToLower(args[0])
	pattern, err := regexp.Compile(inputpattern)
	if err != nil {
		log.Fatalln(" error from regex compile function is ", err)
	}

	startDirectory, _ := os.Getwd() // startDirectory is a string
	fmt.Println()
	fmt.Printf(" dsrtre (recursive), written in Go.  Last altered %s, will use regex of %q and will start in %s. \n", lastAltered, pattern, startDirectory)
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
			fmt.Printf(" Error from walk is %v. \n ", err)
			return nil
		}

		if fi.IsDir() {
			if DirAlreadyWalked[fpath] {
				return filepath.SkipDir
			} else {
				DirAlreadyWalked[fpath] = true
			}
		} else /* if fi.Mode().IsRegular()  */ {
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
				NAME := strings.ToLower(fi.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?copy
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
	if er != nil {
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
/*
{{{
	if linuxflag {
		for _, f := range files {
			s := f.ModTime().Format("Jan-02-2006_15:04:05")
			sizeint := 0
			sizestr := ""
			usernameStr, groupnameStr := GetUserGroupStr(f) // util function in platform specific removed Oct 4, 2019 and then unremoved.
			if FilenameList && f.Mode().IsRegular() {
				SizeTotal += f.Size()
				sizeint = int(f.Size())
				sizestr = strconv.Itoa(sizeint)
				if sizeint > 100000 {
					sizestr = AddCommas(sizestr)
				}
				fmt.Printf("%10v %s:%s %15s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				count++
			} else if IsSymlink(f.Mode()) {
				fmt.Printf("%10v %s:%s %15s %s <%s>\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				count++
			} else if Dirlist && f.IsDir() {
				fmt.Printf("%10v %s:%s %15s %s (%s)\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				count++
			}
			if count >= NumLines {
				break
			}
		}
	} else if winflag {
		for _, f := range files {
			NAME := strings.ToUpper(f.Name())
			if BOOL, _ := filepath.Match(CleanFileName, NAME); BOOL {
				s := f.ModTime().Format("Jan-02-2006_15:04:05")
				sizeint := 0
				sizestr := ""
				if FilenameList && f.Mode().IsRegular() {
					SizeTotal += f.Size()
					sizeint = int(f.Size())
					sizestr = strconv.Itoa(sizeint)
					if sizeint > 100000 {
						sizestr = AddCommas(sizestr)
					}
					fmt.Printf("%15s %s %s\n", sizestr, s, f.Name())
					count++
				} else if IsSymlink(f.Mode()) {
					fmt.Printf("%15s %s <%s>\n", sizestr, s, f.Name())
					count++
				} else if Dirlist && f.IsDir() {
					fmt.Printf("%15s %s (%s)\n", sizestr, s, f.Name())
					count++
				}
				if count >= NumLines {
					break
				}
			}
		}
	}

}}}
*/
/*
{{{
// ------------------------------ IsSymlink ---------------------------
func IsSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink
}}}
*/
