// since.go

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	// jwalk "github.com/MichaelTJones/walk"  It made no difference vs the std lib, once I figured out that I had to return SkipDir on any errors.
	// "github.com/whosonfirst/walk"  Not designed for windows, and I doesn't do what I want, so I'll delete it.
)

/*
   REVISION HISTORY
   -------- -------
   21 Oct 2018 -- First started playing w/ MichaelTJones' code.  I added a help flag
   21 Oct 2022 -- In the code again, after running golangci-lint.  Changed how help flag contents is output.  If an absolute time is given, use that, else use duration.
   26 Oct 2022 -- Added '~' processing.  And output timing info.
   28 Oct 2022 -- On linux this takes an hour to run when I invoke it using my home directory.  I'm looking into why.  I think because it's following a symlink to bigbkupG.
                  Naw, it's also following symlinks to DSM.
                  I posted on golang-nuts for help.  I'm adding the DevID that was recommended, and removing multiple start directories as the pre-processing was complex to work out.
   29 Oct 2022 -- jwalk doesn't work, as it exits too early.  filepath/walk takes ~2 min here on leox.  I'm adding a wait group, and now it works, taking ~7 sec on leox.
                  I'll leave in the done channel, as a model of something that's supposed to work but doesn't.  At least for now.
                  Turns out that the syscall used by GetDeviceID won't compile on Windows, so I have to use platform specific code for it.  I'll do that now.
   31 Oct 2022 -- Happy Halloween.  Turns out that I didn't need to add a wait group after all.  I'll confirm that here by removing it.  Confirmed.
    2 Nov 2022 -- Time to clean out the crap that accumulated while I sorted out the code.  The GitHub repo was last modified 8 yrs ago, which is Go 1.3 or 1.4.  A lot has
                    changed in that time, now that Go 1.19 is current.  I did not find a difference btwn the std library walk vs Michael T Jones' code I called jwalk.
                    I'll remove that stuff now.
   16 Feb 2023 -- I'll change to using WalkDir instead of Walk.  This essentially changes from a FileInfo to a DirEntry.  The docs say that WalkDir is slightly faster.
   17 Feb 2023 -- Timing info:  Here on Win10 desktop, the Nov 2022 version took 10.8 sec, and the latest version took 2.85 sec when running "since ~", which is 1/4 of orig time.
                                On work win10 computer, the Nov 22 version took 4.7 sec, and the latest version took 1.4 sec to run "since ~", which is ~30% of orig time.
                                This is a big drop.  Wow.
*/

var LastAlteredDate = "Feb 16, 2023"

var duration = flag.Duration("dur", 10*time.Minute, "find files modified within this duration")
var format = flag.String("f", "2006-01-02 03:04:05", "time format")
var instant = flag.String("t", "", "find files modified since TIME")
var quiet = flag.Bool("q", false, "do not print filenames")
var verbose = flag.Bool("v", false, "print summary statistics")
var days = flag.Int("d", 0, "days duration")
var weeks = flag.Int("w", 0, "weeks duration")

//var wg sync.WaitGroup  I'm using a channel to signal

type devID uint64

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last modified %s, compiled with %s, last linked %s.\n", os.Args[0], LastAlteredDate, runtime.Version(), ExecTimeStamp)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: since <options> <start-dir-list> \n")
		fmt.Fprintf(flag.CommandLine.Output(), " Valid time units for duration are ns, us, ms, s, m, h. \n")
		fmt.Fprintf(flag.CommandLine.Output(), " since -dur 5m -- show all files changed within last 5 minutes starting at current directory \n")
		fmt.Fprintf(flag.CommandLine.Output(), " since -dur 5m ~ or $HOME or %%userprofile -- show all files changed within last 5 minutes starting at home directory \n")
		flag.PrintDefaults()
	}
	flag.Parse()

	fmt.Printf(" since written in Go.  LastAltered %s, compiled with %s, binary timestamp is %s.\n", LastAlteredDate, runtime.Version(), ExecTimeStamp)

	now := time.Now()
	when := now
	switch {
	case *instant != "":
		t, err := time.Parse(*format, *instant)
		if err != nil {
			fmt.Printf("error parsing time %q, %s\n", *instant, err)
			os.Exit(1)
		}
		when = t
	default:
		//d, err := time.ParseDuration(*duration)
		//if err != nil {
		//	fmt.Printf("error parsing duration %q, %s\n", *duration, err)
		//	os.Exit(2)
		//}
		*duration = *duration + time.Duration(*weeks)*7*24*time.Hour + time.Duration(*days)*24*time.Hour
		when = now.Add(-*duration) // subtract duration from now.
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from UserHomeDir is %v.\n", err)
	}

	if *verbose {
		fmt.Printf(" weeks = %d, days = %d, duration = %s\n", *weeks, *days, *duration)
		fmt.Printf(" when = %s, home directory is %s\n", when, home)
	}

	// goroutine to collect names of recently-modified files
	var result []string
	done := make(chan bool)
	results := make(chan string, 1024)
	go func() {
		for r := range results {
			result = append(result, r)
		}
		sort.Strings(result) // simulate ordered traversal
		done <- true
	}()

	// walkDir to find recently-modified files
	//var lock sync.Mutex    // I took out this mutex pattern here as a reference; I like the atomic add stuff better as I believe it to be cleaner w/ less code.
	var tFiles int64         // total files and bytes
	var rFiles, rBytes int64 // recent files and bytes
	var rootDeviceID devID
	var dir string

	if len(flag.Args()) < 1 {
		dir, err = os.Getwd()
		if err != nil {
			log.Fatalln(" error from Getwd is", err)
		}
	} else {
		dir = flag.Arg(0) // will only use the first argument, which is all I use anyway.
		dir = strings.Replace(dir, "~", home, 1)
	}
	fi, er := os.Stat(dir)
	if er != nil {
		log.Fatalf(" error from os.Stat(%s) is %v\n", dir, er)
	}
	rootDeviceID = getDeviceID(fi)

	//sizeVisitor := func(path string, info os.FileInfo, err error) error {
	walkDirFunc := func(path string, d os.DirEntry, err error) error {
		if *verbose {
			fmt.Printf(" Path = %s, err = %s, isdir = %t, d.name = %s, d.type = %v  \n ", path, err, d.IsDir(), d.Name(), d.Type())
		}

		var info os.FileInfo
		if err != nil {
			if *verbose {
				fmt.Printf(" Trying to enter %s, got error %s.  Skipping it.\n", path, err)
			}
			return filepath.SkipDir
		}

		atomic.AddInt64(&tFiles, 1)
		//lock.Lock()
		//tFiles += 1
		//tBytes += int(info.Size())
		//lock.Unlock()

		if d.IsDir() {
			if filepath.Ext(path) == ".git" || strings.Contains(path, "vmware") || strings.Contains(path, ".cache") { // adding extra skipDir's saved from 13 min -> 9 sec runtime.
				if *verbose {
					fmt.Printf(" skipping %s.\n", path)
				}
				return filepath.SkipDir
				//} else if isSymlink(info.Mode()) { // skip all symlinked directories.  I intend this to catch bigbkupG and DSM.  It doesn't follow symlinks, so I'm removing this.
				//	if *verbose {
				//		fmt.Printf(" skipping symlink %s\n", path)
				//	}
				//	return filepath.SkipDir
			}

			info, err = d.Info()
			if err != nil {
				return filepath.SkipDir
			}

			id := getDeviceID(info)
			if rootDeviceID != id {
				if *verbose {
					fmt.Printf(" root device id is %d for %q, path device id is %d for %q.  Skipping %s.\n", rootDeviceID, dir, id, path, path)
				}
				return filepath.SkipDir
			}
			return nil // it's a directory.  Need to compare against files and not directories.  So leave.
		}

		info, err = d.Info()
		if err != nil {
			return filepath.SkipDir
		}

		if info.ModTime().After(when) {
			atomic.AddInt64(&rFiles, 1)
			atomic.AddInt64(&rBytes, info.Size())
			//lock.Lock()
			//rFiles += 1
			//rBytes += info.Size()
			//lock.Unlock()

			if !*quiet {
				results <- path // allows sorting into "normal" order
			}
		}
		return nil
	}

	//err = filepath.Walk(dir, sizeVisitor)
	err = filepath.WalkDir(dir, walkDirFunc)

	if err != nil {
		log.Printf(" error from walk.Walk is %v\n", err)
	}

	// wait for traversal results and print
	close(results) // no more results
	<-done         // blocking channel receive, to wait for final results and sorting
	//wg.Wait()
	//𝛥t := float64(time.Since(now)) / 1e9 // duration unit is essentially nanosec's.  So by dividing by nn/s it converts to sec, and is reported that way below.
	elapsed := time.Since(now)

	for _, r := range result {
		fmt.Printf("%s\n", r)
	}

	fmt.Printf(" since ran for %s\n", elapsed)

	// print optional verbose summary report
	if *verbose {
		//log.Printf("     total: %8d files (%7.2f%%), %13d bytes (%7.2f%%)\n", tFiles, 100.0, tBytes, 100.0)
		log.Printf("     total: %8d files (%7.2f%%)\n", tFiles, 100.0)

		rfp := 100 * float64(rFiles) / float64(tFiles)
		//rbp := 100 * float64(rBytes) / float64(tBytes)
		//log.Printf("    recent: %8d files (%7.2f%%), %13d bytes (%7.2f%%) in %.4f seconds\n", rFiles, rfp, rBytes, rbp, 𝛥t)
		log.Printf("    recent: %8d files (%7.2f%%), %13d bytes in %s \n", rFiles, rfp, rBytes, elapsed)
	}
}

/* Not used as of Feb 16, 2023.
func isSymlink(fm os.FileMode) bool {
	intermed := fm & os.ModeSymlink
	return intermed != 0
}

*/
