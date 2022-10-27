// since.go

package main

import (
	"flag"
	"fmt"
	"github.com/MichaelTJones/walk"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

/* REVISION HISTORY
   21 Oct 2018 -- First started playing w/ MichaelTJones' code.  I added a help flag
   21 Oct 2022 -- In the code again, after running golangci-lint.  Changed how help flag contents is output.  If an absolute time is given, use that, else use duration.
   26 Oct 2022 -- Added '~' processing.  And output timing info.
*/

var LastAlteredDate = "Oct 26, 2022"

//var duration = flag.String("d", "", "find files modified within DURATION")
var duration = flag.Duration("dur", 5*time.Minute, "find files modified within this duration")
var format = flag.String("f", "2006-01-02 03:04:05", "time format")
var instant = flag.String("t", "", "find files modified since TIME")
var quiet = flag.Bool("q", false, "do not print filenames")
var verbose = flag.Bool("v", false, "print summary statistics")
var days = flag.Int("d", 0, "days duration")
var weeks = flag.Int("w", 0, "weeks duration")

//var help = flag.Bool("h", false, "print help message")

func main() {

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

	fmt.Printf(" since written in Go.  LastAltered %s, compiled with %s, last linked %s.\n", LastAlteredDate, runtime.Version(), ExecTimeStamp)

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

	// parallel walker and walk to find recently-modified files
	var lock sync.Mutex
	var tFiles, tBytes int // total files and bytes
	var rFiles, rBytes int // recent files and bytes
	sizeVisitor := func(path string, info os.FileInfo, err error) error {
		if err == nil {
			lock.Lock()
			tFiles += 1
			tBytes += int(info.Size())
			lock.Unlock()

			if info.ModTime().After(when) {
				lock.Lock()
				rFiles += 1
				rBytes += int(info.Size())
				lock.Unlock()

				if !*quiet {
					// fmt.Printf("%s %s\n", info.ModTime(), path) // simple
					results <- path // allows sorting into "normal" order
				}
			}
		}
		return nil
	}
	if len(flag.Args()) < 1 {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalln(" error from Getwd is", err)
		}
		err = walk.Walk(dir, sizeVisitor)
		if err != nil {
			log.Fatalln(" error from walk.Walk is", err)
		}
	} else {
		for _, root := range flag.Args() {
			dir := strings.Replace(root, "~", home, 1) // I decided to not test for windows or presence of ~.  This is pretty fast as it is.
			err := walk.Walk(dir, sizeVisitor)
			if err != nil {
				log.Fatalln(" error from walk.Walk is", err)
			}
		}
	}

	// wait for traversal results and print
	close(results) // no more results
	<-done         // wait for final results and sorting
	𝛥t := float64(time.Since(now)) / 1e9

	for _, r := range result {
		fmt.Printf("%s\n", r)
	}

	fmt.Printf(" since ran for %s\n", time.Since(now))

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
