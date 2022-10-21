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
   2 Feb 22 -- Refactoring -- removing the go routine pattern as it's not necessary.  And experimenting w/ Walk vs WalkDir
  21 Oct 22 -- Fixed a bad use of format verb on an error message.  Caught by golangci-lint
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const lastAltered = "21 Oct 2022"

func main() {
	var timeoutOpt *int = flag.Int("t", 900, "seconds < 1800, where 0 means timeout of 900 sec.")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 1800 {
		log.Println("timeout must be in the range [0..1800] seconds.  Making default of 900")
		*timeoutOpt = 900
	}

	var globPattern, startDir string
	var err error
	if flag.NArg() == 0 {
		fmt.Print(" Enter globbing pattern: ")
		fmt.Scanln(&globPattern)
		startDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, " Getwd returned this error: %v\n", err)
			os.Exit(1)
		}

	} else if flag.NArg() == 1 {
		globPattern = flag.Arg(0)
		startDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, " Getwd returned this error: %v\n", err)
			os.Exit(1)
		}
	} else {
		globPattern = flag.Arg(0)
		startDir = flag.Arg(1)
	}

	globPattern = strings.ToLower(globPattern)

	fmt.Println()
	fmt.Printf(" dsrtr (recursive), written in Go.  Last altered %s, will use globbing pattern of %q and will start in %s. \n", lastAltered, globPattern, startDir)
	fmt.Println()
	fmt.Println()
	//DirAlreadyWalked := make(map[string]bool, 500)
	//DirAlreadyWalked[".git"] = true // ignore .git and its subdir's

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)
	/*
		// walkfunc closure
		filepathWalkFunction := func(fpath string, fi os.FileInfo, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from walk is %v. \n ", err)
				return nil
			}

			if fi.IsDir() && fpath == ".git" {
				return filepath.SkipDir
			} else if isSymlink(fi.Mode()) {
				fmt.Printf(" %s is a symlink, mode is %v\n", fpath, fi.Mode())
				return filepath.SkipDir
			}

			// Must be a regular file
			NAME := strings.ToLower(fi.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?
			if BOOL, _ := filepath.Match(globPattern, NAME); BOOL {
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

		err = filepath.Walk(startDir, filepathWalkFunction)
	*/

	filepathWalkDirEntry := func(fpath string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from walk is %v. \n ", err)
			return nil
		}

		if d.IsDir() && fpath == ".git" {
			return filepath.SkipDir
		} else if isSymlink(d.Type()) {
			fmt.Printf(" %s is a symlink, name is %s, mode is %v\n", fpath, d.Name(), d.Type())
			//return filepath.SkipDir
		}

		// Must be a regular file
		NAME := strings.ToLower(d.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?
		if BOOL, _ := filepath.Match(globPattern, NAME); BOOL {
			fi, er := d.Info()
			if er != nil {
				fmt.Fprintf(os.Stderr, " %s.Info() call error is %v\n", d.Name(), er)
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
		fmt.Fprintf(os.Stderr, " Error from filepath.walk is %v.  Elapsed time is %s\n", err, time.Since(t0))
	}

	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed)
	fmt.Println()
} // end main

//-------------------------------------------------------------------- InsertByteSlice --------------------------------

func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

//---------------------------------------------------------------------- AddCommas ------------------------------------

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

// ----------------------------                   GetIDname -----------------------------------------------------------

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
