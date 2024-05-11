package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

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
	21 Oct 22 -- Fixed bad format verb use caught by golangci-lint.
	12 Nov 22 -- Adding device ID code, and error handling code I developed for since and multack.  And I think I need a sync mechanism like a wait group or done channel.
	14 Nov 22 -- Added processing for "~".
	17 Feb 23 -- Based on what I learned by speeding up since.go, I'll port those optimizations here.  These are:
	                  I took out tests for symlink, run os.Stat only after directory check for the special directories, only call deviceID on a dir entry,
	                  and does an ordinary directory return without checking Modtime().After(when).
     9 May 24 -- Removed commented out code.  And added test for ".git".
    10 May 24 -- Made result slice to be size of 1000 instead of 0.
*/

const lastAltered = "10 May 2024"

type devID uint64

func main() {
	var timeoutOpt *int = flag.Int("t", 900, "seconds < 1800, where 0 means timeout of 900 sec.")
	var verboseFlag = flag.Bool("v", false, "enter a verbose testing mode to println more variables")
	var inputRegexPattern, startDir string
	var inputRegex *regexp.Regexp
	var err error
	var rootDeviceID devID

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
		home, er := os.UserHomeDir()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from os.UserHomeDir() is %s.  Exiting. \n", er)
			os.Exit(1)
		}
		startDir = strings.ReplaceAll(startDir, "~", home)
	}

	inputRegexPattern = strings.ToLower(inputRegexPattern)
	inputRegex, err = regexp.Compile(inputRegexPattern)
	if err != nil {
		log.Fatalf(" error from regex compile function is %s", err)
	}

	fmt.Println()
	fmt.Printf(" dsrtre (recursive), written in Go.  Last altered %s, will use regex of %q and will start in %s. \n", lastAltered, inputRegex.String(), startDir)
	fmt.Println()
	if *verboseFlag {
		execDir, _ := os.Getwd()
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
		fmt.Printf(" Current working Directory is %s; %s timestamp is %s.\n\n", execDir, execName, LastLinkedTimeStamp)
		fmt.Println()
	}

	t0 := time.Now()
	tFinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)
	rootFileInfo, err := os.Stat(startDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Exiting\n", startDir, err)
		os.Exit(1)
	}
	rootDeviceID = getDeviceID(rootFileInfo)

	// goroutine to collect names of matching files
	result := make([]string, 0, 1000)
	done := make(chan bool)
	results := make(chan string, 1024)
	go func() {
		for r := range results {
			result = append(result, r)
		}
		sort.Strings(result) // simulate ordered traversal
		done <- true
	}()

	filepathWalkDirEntry := func(fPath string, d os.DirEntry, err error) error {
		if err != nil {
			if *verboseFlag {
				fmt.Fprintf(os.Stderr, " Error from walk is %v. \n ", err)
			}
			return filepath.SkipDir
		}

		if fPath == ".git" {
			return filepath.SkipDir
		}

		if d.IsDir() { // if directory, either skip it or return, but don't process it.
			if strings.Contains(fPath, ".git") || strings.Contains(fPath, "vmware") || strings.Contains(fPath, ".cache") {
				return filepath.SkipDir
			}

			info, e := d.Info() // needed to feed into getDeviceID.
			if e != nil {
				if *verboseFlag {
					fmt.Fprintf(os.Stderr, " Error from %s is %s \n", fPath, e)
				}
				return filepath.SkipDir
			}

			pathDeviceID := getDeviceID(info)
			if rootDeviceID != pathDeviceID {
				if *verboseFlag {
					fmt.Fprintf(os.Stderr, " %s is on a difference device from %s,  Skipping\n", fPath, startDir)
				}
				return filepath.SkipDir
			}

			return nil
		}

		// Must be a regular file
		NAME := strings.ToLower(d.Name()) // Despite windows not being case-sensitive, filepath.Match is case-sensitive.  Who new?
		if BOOL := inputRegex.MatchString(NAME); BOOL {
			fi, er := d.Info()
			if er != nil {
				fmt.Fprintf(os.Stderr, " %s.Info() call error is %v\n", d.Name(), er)
				return filepath.SkipDir
			}
			t := fi.ModTime().Format("Jan-02-2006_15:04:05")
			sizeStr := strconv.Itoa(int(fi.Size()))
			if fi.Size() > 100_000 {
				sizeStr = AddCommas(sizeStr)
			}
			s := fmt.Sprintf("%15s : %s : %s", sizeStr, t, fPath)
			results <- s
		}

		now := time.Now()
		if now.After(tFinal) {
			log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
		}

		return nil
	}

	err = filepath.WalkDir(startDir, filepathWalkDirEntry)
	if err != nil {
		log.Printf(" Error from filepath.walk is %s.  Elapsed time is %s\n", err, time.Since(t0))
	}
	close(results)

	<-done // blocking until something is received from the done channel.  That something is then discarded.

	for _, r := range result {
		fmt.Printf(" %s\n", r)
	}
	fmt.Println()

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

/*
// ------------------------------ isSymlink ---------------------------
func isSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink

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
*/
