package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
)

/*
  REVISION HISTORY
  -------- -------
  29 Sep 22 -- First version.  Inspired by reading docs of ripgrep, which determines whether a file is a binary if it finds a null byte.
                 My plan is to open a file, attach a bytes buffer, than use a form of bytes.Contains to see if a null byte matches.
                 I won't be able to determine where the match is doing this.  Maybe I'll then switch to determine where the first match occurs.
                 Or try to use bytes.IndexByte, maybe IndexRune, maybe Count.  I primarily want to see if any office docs of any kind have a null byte.
                 And maybe find out how many null bytes exist, if I'm curious.  I may need a byte slice containing one element which is zero for some
                 of these functions to work.
                 Some of this code will likely be based on feqbbb.go and eols.go.
*/

const lastModified = "Sep 30, 2022"
const zero = 0 // I want this to be a null byte.
const K = 1024
const M = K * K
const G = M * M
const maxSize = G

func main() {
	fmt.Printf("Nulls last modified %s.  This pgm reads the entire file at once to count nulls. \n\n", lastModified)

	var filename1 string

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "v", false, " verbose mode.")
	flag.Parse()

	if verboseFlag {
		fmt.Printf(" WorkingDir = %s, execName is %s, which was last linked %s.\n\n", workingDir, execName, LastLinkedTimeStamp)
	}

	if flag.NArg() == 0 {
		fmt.Printf("\n Need a file on the command line to count it's nulls.  Exiting. \n\n")
		os.Exit(1)
	} else if flag.NArg() >= 1 { // will use first filename entered on commandline
		filename1 = flag.Arg(0)
	}

	fi1, e1 := os.Stat(filename1)
	if e1 != nil {
		fmt.Fprintf(os.Stderr, " Stat operation on %s gives error of %v.  Exiting. \n", filename1, e1)
		os.Exit(1)
	}

	if fi1.Size() > maxSize {
		fmt.Fprintf(os.Stderr, " Size of %s is %d, which exceeds the max size allowed of %d.  Exiting \n", filename1, fi1.Size(), maxSize)
		os.Exit(1)
	}

	fileBytes, err := os.ReadFile(filename1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading the file %s is %#v.  Exiting \n", filename1, err)
		os.Exit(1)
	}
	if verboseFlag {
		fmt.Printf(" Size of %s is %d, and length of the read bytes slice is %d.\n", filename1, fi1.Size(), len(fileBytes))
	}

	i := bytes.IndexByte(fileBytes, 0)
	j := bytes.IndexRune(fileBytes, zero)
	cnt := bytes.Count(fileBytes, []byte{0})
	lastNull := bytes.LastIndexByte(fileBytes, 0)

	if i < 0 {
		fmt.Printf(" No null bytes found in %s.\n", filename1)
	} else {
		fmt.Printf(" Found first null byte at %d index and first null rune at %d index.  Total of %d null bytes were found.  Last null byte found at index of %d\n",
			i, j, cnt, lastNull)
	}
	fmt.Println()
}
