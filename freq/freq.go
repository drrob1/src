// (C) 1990-2021.  Robert W Solomon.  All rights reserved.
// freq.go
package main

import (
	"bytes"
	"flag"
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
11 Apr 21 -- Adding flag package and test flag to streamline the output.  And adding <CR> and <LF> counts to output.
*/

const lastCompiled = "11 Apr 2021"
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
	flag.Parse()

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	if *testFlag {
		fmt.Println(ExecFI.Name(), "was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
		fmt.Println(" Full name of executable file is", execname)
	}
	fmt.Println()

	args := flag.Args()
	// os.Arg[1] could be the -test flag and I don't want that confusion.
	if flag.NArg() > 0 {
		ans = args[0]
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
			if *testFlag {
				fmt.Println(infilename, " size =", fi.Size())
			}
		}
	} else {
		infilename = BaseFilename + extDefault
		fi, err := os.Stat(infilename)
		if err == nil {
			InFileExists = true
			if *testFlag {
				fmt.Println(infilename, "size is", fi.Size())
			}
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
	if *testFlag {
		fmt.Println(" Size of filecontents is", len(filecontents), "and length of filebuffer is", filebuffer.Len(), "and cap of buffer is", filebuffer.Cap())
	}

	for {
		r, size, err := filebuffer.ReadRune()
		if err != nil {
			break
		}
		if size > 1 || r > 126 { // there is a character that looks like a reverse video '?', has value of 63355 but a size of 1, that I want to trap here.
			// fmt.Fprintf(os.Stderr, "Size of read rune is %d.  Skipping %d, %q \n", size, r, r)  Don't need to see these anymore.
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

	if *testFlag {
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

	sortfcn := func(i, j int) bool {
		return letters[i].count > letters[j].count // I want the most often letter to sort in front.
	}
	sort.Slice(letters, sortfcn)

	if *testFlag {
		fmt.Println()
		fmt.Println(" letters and counts after sort:")
		for _, ltr := range letters {
			fmt.Printf("%c:%d ", ltr.r, ltr.count)
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Print(" Just sorted letters: ")
	for i := 0; i < len(letters); i++ {
		fmt.Printf("%c", letters[i].r)
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
