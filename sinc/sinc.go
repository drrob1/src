// sinc.go because I'm going to play a bit more w/ this code.

package main

import (
	"flag"
	"fmt"
	pwalk "github.com/peter-mount/go-kernel/util/walk"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
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
                  Now called sinc.go, so I can play a bit more w/ it.
*/

var LastAlteredDate = "Oct 29, 2022"

//var duration = flag.String("d", "", "find files modified within DURATION")
var duration = flag.Duration("dur", 10*time.Minute, "find files modified within this duration")
var format = flag.String("f", "2006-01-02 03:04:05", "time format")
var instant = flag.String("t", "", "find files modified since TIME")
var quiet = flag.Bool("q", false, "do not print filenames")
var verbose = flag.Bool("v", false, "print summary statistics")
var days = flag.Int("d", 0, "days duration")
var weeks = flag.Int("w", 0, "weeks duration")
var wg sync.WaitGroup

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
		fmt.Fprintf(flag.CommandLine.Output(), " since -d 5m -- show all files changed within last 5 minutes starting at current directory \n")
		fmt.Fprintf(flag.CommandLine.Output(), " since -d 5m $HOME or %%userprofile -- show all files changed within last 5 minutes starting at home directory \n")
		flag.PrintDefaults()
	}
	flag.Parse()

	fmt.Printf(" since written in Go.  LastAltered %s, compiled with %s, binary timestamp is %s.\n", LastAlteredDate, runtime.Version(), ExecTimeStamp)

	t0 := time.Now()
	when := t0
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
		when = t0.Add(-*duration) // subtract duration from now.
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

	// parallel walker and walk to find recently-modified files
	//var lock sync.Mutex
	var tFiles, tBytes int64 // total files and bytes
	var rFiles, rBytes int64 // recent files and bytes
	var rootDeviceID devID
	var rootDir string

	if len(flag.Args()) < 1 {
		rootDir, err = os.Getwd()
		if err != nil {
			log.Fatalln(" error from Getwd is", err)
		}
	} else {
		rootDir = flag.Arg(0) // will only use the first argument, which is all I use anyway.
		rootDir = strings.Replace(rootDir, "~", home, 1)
	}
	fi, er := os.Stat(rootDir)
	if er != nil {
		log.Fatalf(" error from os.Stat(%s) is %v\n", rootDir, er)
	}
	rootDeviceID = getDevID(rootDir, fi)

	devPredicate := pwalk.PathPredicate(func(path string, info os.FileInfo) bool { // I probably don't need to declare this func as a PathPredicate.
		deviceID := getDevID(path, info)
		if rootDeviceID == deviceID {
			return true
		}
		return false
	})

	gitPredicate := func(path string, info os.FileInfo) bool { // this has the type of pwalk.PathPredicate, I guess by having the correct signature.
		if strings.Contains(path, ".git") {
			return false // don't want to process anything in the .git tree
		}
		return true // process everything else, for now
	}

	dbusPredicate := func(path string, info os.FileInfo) bool { // this has the type of pwalk.PathPredicate, I guess by having the correct signature.
		if strings.Contains(path, ".dbus") { // I'm getting a permission denied error from /home/rob/.dbus, so I'll block the attempt.
			return false // don't want to process anything in the .git tree
		}
		return true // process everything else, for now
	}

	walker := pwalk.NewPathWalker()

	walker = func(path string, info os.FileInfo) error {
		wg.Add(1)
		defer func() {
			wg.Done()
			atomic.AddInt64(&tFiles, 1)
			atomic.AddInt64(&tBytes, info.Size())
		}()
		if info.ModTime().After(when) {
			atomic.AddInt64(&rFiles, 1)
			atomic.AddInt64(&rBytes, info.Size())
			results <- path // allows sorting into "normal" order
		}
		return nil
	}

	walker = walker.If(devPredicate)
	walker = walker.If(gitPredicate)
	walker = walker.If(dbusPredicate)
	walker = walker.PathHasNotSuffix(".git") // redundant, but I want to see if this will compile.  So far, so good.
	walker = walker.PathNotContain("vmware") // I don't want it to run thru any vmware stuff.
	walker = walker.PathNotContain("cache")
	walker = walker.IsFile()

	/*	sizeVisitor := func(path string, info os.FileInfo, err error) error {
			wg.Add(1)
			defer func() {
				wg.Done()
				atomic.AddInt64(&tFiles, 1)
				atomic.AddInt64(&tBytes, info.Size())
			}()
			//err = nil                                  // ignore all errors.
			//if err == nil || err == filepath.SkipDir { // maybe I've been doing this myself all along w/ my own foot gun.
			if info.IsDir() {
				if filepath.Ext(path) == ".git" {
					if *verbose {
						fmt.Printf(" skipping .git\n")
					}
					return filepath.SkipDir
				} else if strings.Contains(path, ".cache") {
					if *verbose {
						fmt.Printf(" skipping .cache\n")
					}
					return filepath.SkipDir
				} else if isSymlink(info.Mode()) { // skip all symlinked directories.  I intend this to catch bigbkupG and DSM.
					if *verbose {
						fmt.Printf(" skipping symlink %s\n", path)
					}
					return filepath.SkipDir
				} else {
					id := getDeviceID(path, info)
					if rootDeviceID != id {
						if *verbose {
							fmt.Printf(" root device id is %d for %q, path device id is %d for %q.  Skipping.\n", rootDeviceID, dir, id, path)
						}
						return filepath.SkipDir
					}
				}
			}

			if info.ModTime().After(when) {
				atomic.AddInt64(&rFiles, 1)
				atomic.AddInt64(&rBytes, info.Size())

				if !*quiet {
					// fmt.Printf("%s %s\n", info.ModTime(), path) // simple
					results <- path // allows sorting into "normal" order
				}
			}
			//}
			return nil
		}

		if *quiet { // just so compiler sees this can potentially still be executed.
			err = jwalk.Walk(dir, sizeVisitor) // at least this compiles on Windows.  It doesn't work, but it does compile.
			// err = walk.Walk(dir, sizeVisitor) // a fork of jwalk w/ some needed changes.  But it doesn't work, either.  It's not even designed to compile on Windows.
			err = filepath.Walk(dir, sizeVisitor) // this is the only one that works as expected.
			err = walk.WalkWithNFSKludge(dir, sizeVisitor)
			err = awalk.Walk(dir, sizeVisitor)
		} else {
			err = powerwalk.WalkLimit(dir, sizeVisitor, powerwalk.DefaultConcurrentWalks) // docs say that this routine does not follow symlinks.  Maybe that's what I need?
		}

	*/

	err = walker.Walk(rootDir)

	if err != nil {
		log.Printf(" error from walk.Walk is %v\n", err)
	}

	// wait for traversal results and print
	close(results) // no more results
	<-done         // wait for final results and sorting
	fmt.Printf(" done has returned, but before waitgroup.  Elapsed time is %s.\n\n", time.Since(t0))
	wg.Wait()
	𝛥t := float64(time.Since(t0)) / 1e9

	for _, r := range result {
		fmt.Printf("%s\n", r)
	}

	fmt.Printf(" since ran for %s\n", time.Since(t0))

	// print optional verbose summary report
	if *verbose {
		log.Printf("     total: %8d files (%7.2f%%), %13d bytes (%7.2f%%)\n",
			tFiles, 100.0, tBytes, 100.0)

		rfp := 100 * float64(rFiles) / float64(tFiles)
		rbp := 100 * float64(rBytes) / float64(tBytes)
		log.Printf("    recent: %8d files (%7.2f%%), %13d bytes (%7.2f%%) in %.4f seconds\n",
			rFiles, rfp, rBytes, rbp, 𝛥t)
	}
}

func getDevID(path string, fi os.FileInfo) devID {
	var stat = fi.Sys().(*syscall.Stat_t)
	return devID(stat.Dev)
}

/*func isSymlink(fm os.FileMode) bool {
	intermed := fm & os.ModeSymlink
	return intermed != 0
}
*/
