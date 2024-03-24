// (C) 1990-2021.  Robert W Solomon.  All rights reserved.
// freq2.go based on freq.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

/*
REVISION HISTORY
----------------
 9 Apr 21 -- Just started working on the to populate and sort a letter frequency table using my .txt files as source material
11 Apr 21 -- Adding flag package and test flag to streamline the output.  And adding <CR> and <LF> counts to output.
13 Apr 21 -- Added verbose flag as synonym for test mode.  Verbose is more consistent w/ most utils.
22 Oct 21 -- Removing the depracated (as of Go 1.16) ioutil.
24 Mar 24 -- Now called freq2, based on freq.  But now I'm going to use a map.  Turns out that I've always used a map.  So now I'm just cleaning up my code from 3 yrs ago.
*/

const lastCompiled = "24 Mar 2024"
const extDefault = ".txt"

type letter struct {
	r     rune
	count int
}

func main() {
	var infilename, ans string

	rawRuneMap := make(map[rune]int, 255)

	fmt.Printf(" freq, a letter frequency program written in Go.  Last altered %s, compiled with %s. \n", lastCompiled, runtime.Version())
	var testFlag = flag.Bool("test", false, "enter a testing mode to println more variables")
	var verboseFlag = flag.Bool("v", false, "verbose mode to println more variables and messages.")
	flag.Parse()

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	verboseMode := *testFlag || *verboseFlag

	if verboseMode {
		fmt.Println(ExecFI.Name(), "was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingDir, ".")
		fmt.Println(" Full name of executable file is", execName)
	}
	fmt.Println()

	// os.Arg[1] could be the -test flag and I don't want that confusion.

	// construct a filename.
	if flag.NArg() > 0 {
		ans = flag.Arg(0)
	} else {
		fmt.Print(" Enter a filename to process: ")
		_, err := fmt.Scanln(&ans)
		if err != nil {
			fmt.Println(err, " Exiting.")
			os.Exit(1)
		}
	}
	BaseFilename := filepath.Clean(ans)
	InFileExists := false
	if strings.Contains(BaseFilename, ".") {
		infilename = BaseFilename
		fi, err := os.Stat(infilename)
		if err == nil {
			InFileExists = true
			if verboseMode {
				fmt.Println(infilename, " size =", fi.Size())
			}
		}
	} else {
		infilename = BaseFilename + extDefault
		fi, err := os.Stat(infilename)
		if err == nil {
			InFileExists = true
			if verboseMode {
				fmt.Println(infilename, "size is", fi.Size())
			}
		}
	}

	if !InFileExists {
		fmt.Println(" File ", BaseFilename, " or ", infilename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	// read in the file as a slice of bytes.
	fileContents, err := os.ReadFile(infilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err, ".  Exiting")
		os.Exit(1)
	}

	fileBuffer := bytes.NewBuffer(fileContents)
	if verboseMode {
		fmt.Println(" Size of file contents is", len(fileContents), "and length of file buffer is", fileBuffer.Len(), "and cap of buffer is", fileBuffer.Cap())
	}

	for { // determine the freq of each rune.
		r, size, err := fileBuffer.ReadRune()
		if err != nil { // err is not nil at EOF condition.
			break
		}
		if size > 1 || r > 126 { // there is a character that looks like a reverse video '?', has value of 63355 but a size of 1, that I want to trap here.
			continue
		}
		r = toLower(r) // my toLower take a rune as a param, the std lib takes a string.
		rawRuneMap[r]++
	}

	// The slice of letters is what will get sorted.
	letters := make([]letter, 0, 255)
	for i := 'a'; i <= 'z'; i++ {
		rn := rune(i)
		ltr := letter{r: rn, count: rawRuneMap[rn]}
		letters = append(letters, ltr)
	}

	if verboseMode {
		fmt.Println(" The length of the rawRuneMap is", len(rawRuneMap), ".  The length of the letters slice is", len(letters))
		fmt.Println()
		fmt.Println(" Unsorted rawRuneMap:")
		for i, rm := range rawRuneMap {
			//if i < ' ' { continue }  // skip control characters like <LF> or <CR>
			fmt.Printf("%q:%d:%d ", i, i, rm)
		}
		fmt.Println()
		fmt.Println()

		fmt.Println(" letters before sort:")
		for _, ltr := range letters {
			fmt.Printf("%c:%d ", ltr.r, ltr.count)
		}
		fmt.Println()
		fmt.Println()
	}

	sortFcn := func(i, j int) bool {
		return letters[i].count > letters[j].count // I want the most often letter to sort in front.
	}
	sort.Slice(letters, sortFcn)

	if verboseMode {
		fmt.Println()
		fmt.Println(" letters and counts after sort:")
		for _, ltr := range letters {
			fmt.Printf("%c:%d ", ltr.r, ltr.count)
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Print(" Just sorted letters: ")
	// for i := 0; i < len(letters); i++ {
	for _, letter := range letters {
		fmt.Printf("%c ", letter.r)
	}
	fmt.Println()
	fmt.Printf(" CR: %d, LF: %d \n", rawRuneMap[13], rawRuneMap[10])
	fmt.Println()
} // main in freq.go

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		r += 32 // convert from upper case to lower case
	}
	return r
}
