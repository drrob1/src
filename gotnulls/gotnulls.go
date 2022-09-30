package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
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
                 Some of this code will likely be based on feqbbb.go.
*/

const lastModified = "Sep 30, 2022"
const zero = 0 // I want this to be a null byte.
const K = 1024
const M = K * K

func main() {
	fmt.Printf("GotNulls last modified %s.  This stops when it finds one null, but handles very large files as it only reads 1 MB at a time. \n\n", lastModified)

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
		fmt.Printf("\n Need a file on the command line to determine if it's Got Nulls.  Exiting. \n\n")
		os.Exit(1)
	} else if flag.NArg() >= 1 { // will use first filename entered on commandline
		filename1 = flag.Arg(0)
	}

	openedFile1, err := os.Open(filename1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading the file is %v.  Exiting \n", err)
		os.Exit(1)
	}
	defer openedFile1.Close()

	fileReader1 := bufio.NewReader(openedFile1)
	buf1 := make([]byte, 1*M) // I initially wrote this as ([]byte,0,M) so the buffer was 0 bytes long and this code didn't work.

	var counter int
	var gotNull bool

	for {
		n1, er1 := fileReader1.Read(buf1)
		if n1 == 0 { // no more bytes to process.
			if verboseFlag {
				fmt.Printf(" No more bytes to process.  Read operation from %s returned %v, counter = %d.  Finished.\n", filename1, er1, counter)
			}
			break
		}
		if n1 < M {
			if verboseFlag {
				fmt.Printf(" Read %d bytes, so this must be the last time around this loop.  counter = %d, er1 = %v.\n", n1, counter, er1) // partially read buffer still has er1 = nil.
			}
		}
		if bytes.ContainsRune(buf1[:n1], zero) { // before I used the subslice syntax, this function was finding the nulls in the buffer past the file contents.
			fmt.Printf("ContainsRune of zero is true.  And Contains is %t.\n", bytes.Contains(buf1[:n1], []byte{0}))
			gotNull = true
			i := bytes.IndexByte(buf1[:n1], 0)
			j := bytes.LastIndexByte(buf1[:n1], 0)
			offset := counter*M + i
			lastOffset := counter*M + j
			fmt.Printf(" File %s does contain at least one null byte, counter=%d, i=%d, offset=%d, last index=%d and last offset=%d. \n",
				filename1, counter, i, offset, j, lastOffset)
			break
		}

		counter++
		if er1 == io.EOF { // as earlier runs of this code demonstrated, the read operation only returns io.EOF when there are no more bytes to read, not when a buffer is partially filled.
			fmt.Printf(" Error is io.EOF.  I didn't think I would ever get here.\n")
			break
		}
	}
	fmt.Printf("\n gotnull is %t.\n", gotNull)
}
