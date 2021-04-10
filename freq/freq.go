// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// freq.go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

*/

const lastCompiled = "10 Apr 2021"
const extDefault = ".txt"

type letter struct {
	r     rune
	count int
}

func main() {
	var infilename, ans string

	rawRuneMap := make(map[rune]int, 255)

	fmt.Printf(" freq, a letter frequency program written in Go.  Last altered %s, compiled with %s. \n", lastCompiled, runtime.Version())

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Println(ExecFI.Name(), "was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
	fmt.Println(" Full name of executable file is", execname)
	fmt.Println()

	if len(os.Args) > 1 {
		ans = os.Args[1]
	} else {
		fmt.Print(" Enter a filename to process: ")
		fmt.Scanln(&ans)
	}
	BaseFilename := filepath.Clean(ans)
	InFileExists := false
	if strings.Contains(BaseFilename, ".") {
		infilename = BaseFilename
		fi, err := os.Stat(infilename)
		if err == nil {
			InFileExists = true
			fmt.Println(infilename, " size =", fi.Size())
		}
	} else {
		infilename = BaseFilename + extDefault
		fi, err := os.Stat(infilename)
		if err == nil {
			InFileExists = true
			fmt.Println(infilename, "size is", fi.Size())
		}
	}

	if !InFileExists {
		fmt.Println(" File ", BaseFilename, " or ", infilename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	filecontents, e := ioutil.ReadFile(infilename)
	if e != nil {
		fmt.Fprintln(os.Stderr, e, ".  Exiting")
		os.Exit(1)
	}

	filebuffer := bytes.NewBuffer(filecontents)
	fmt.Println(" Size of filecontents is", len(filecontents), "and length of filebuffer is", filebuffer.Len(), "and cap of buffer is", filebuffer.Cap())

	for {
		r, size, err := filebuffer.ReadRune()
		if err != nil {
			break
		}
		if size > 1 {
			fmt.Fprintln(os.Stderr, "Size of read rune is", size, "skipping.")
			continue
		}
		r = toLower(r)
		rawRuneMap[r]++
	}

	letters := make([]letter, 0, 255)
	for i := 'a'; i <= 'z'; i++ {
		rn := rune(i)
		ltr := letter{r: rn, count: rawRuneMap[rn]}
		letters = append(letters, ltr)
	}

	fmt.Println(" The length of the rawRuneMap is", len(rawRuneMap), ".  The length of the letters slice is", len(letters))
	fmt.Println()

	fmt.Println(" Unsorted rawRuneMap:", rawRuneMap)
	fmt.Println(" letters before sort:", letters)

	sortfcn := func(i, j int) bool {
		return letters[i].count > letters[j].count  // I want the most often letter to sort in front.
	}
	sort.Slice(letters, sortfcn)

	fmt.Println()
	fmt.Print(" After sorting the letters slice: ")
	for i := 0; i < len(letters); i++ {
		fmt.Printf("%c", letters[i].r)
	}
	fmt.Println()
	fmt.Println()
} // main in freq.go

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		r += 32 // convert from upper case to lower case
	}
	return r
}
