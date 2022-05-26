package main // for feqbbb.go

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

/*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- M2:  First modified version of module.  I will use VLI to compare all digits of the hashes.
  23 Apr 13 -- Fixed problem of a single line in the hashes file, that does not contain an EOL character, causes
                an immediate return without processing of the characters just read in.
  24 Apr 13 -- Added output of which file either matches or does not match.
  19 Sep 16 -- Finished conversion to Go, that was started 13 Sep 16.  Added the removal of '*' which is part of a std linux formatted hash file.  And I forgot that
                 the routine allowed either order in the file.  If the token has a '.' I assume it is a filename, else it is a hash value.
  21 Sep 16 -- Fixed the case issue in tokenize.GetToken.  Edited code here to correspond to this fix.
  25 Nov 16 -- Need to not panic when target file is not found, only panic when hash file is not found.
                 And added a LastCompiled message and string.
  13 Oct 17 -- No changes here, but tokenize was changed so that horizontal tab char is now a delim.
  14 Oct 17 -- Tweaked output a bit.  And added executable timestamp code.
  19 Oct 17 -- Added ability to ignore the * that standard hash files for linux use.
  22 Oct 17 -- Added filepicker.
  21 Jan 18 -- Really ignore *.  Before method did not work.
  26 Jan 18 -- Changed tokenize so that SetMapDelim change sticks and actually works.
  13 Nov 18 -- Will use "-" and "_" also to detect a filename token.
  10 Nov 19 -- Now uses ToLower to compare the string hashes, to ignore case.
  15 Jul 20 -- Decided to make better guesses.  Sha1 has 40 digits, Sha256 has 64 digits and Sha512 has 128 digits.
  27 Sep 20 -- From help file of TakeCommand: MD-5 has 32 digits, SHA384 has 96 digits, and the above hash lengths are correct.
                 And I'm going to change from tokenize to tknptr.  Just to see if it works.
  25 Feb 21 -- Added 999 as a stop code.
   3 Mar 21 -- Now called sha.go, which will always use hash length, while ignoring file extension.
                 Errors now go to Stderr.  Uses bytes buffer to read sha file using os.ReadAll, using go 1.16.
   7 Mar 21 -- Added use of strings.TrimSpace()
   8 Apr 21 -- Converting to module version of ~/go/src.
  24 Jan 22 -- Adding a help message using the flag package.  And since I recently changed the interface for filepicker, I have to fix that here too.
   9 Mar 22 -- Using package constants instead of my magic numbers.
                  Now called feq for File Equal, that is, it will determine if 2 files are equal by computing a bunch of hashes.
                  And as this can apply to non-text files as well as text, I won't assume an extension.  Binaries on linux don't have one, anyway.
                  Turns out that crc is much more complex than I expected.  I tried each method just to see if I could get it to work.  But only once.
  12 Mar 22 -- Adding timing info.  For a 3.5 GB file, the results on leox are:
                  Castognoli 2.4s, IEEE 3.6s, byte-by-byte 4.7s, crc64 ECMA 11.5s, sha1 20.3s, md5 28.2s, sha512 35.7s, Koopman 46.8s, sha256 52.8s.
               Now called feqlarge.go, intended for files that are quite large.  I will only open one file at a time, calculate some of the hashes, and
                  compare them afterwards.  When I tested feq.go on a large file (22GB, IIRC), the OS shut it down.
                  I'll only use Castognoli, crc64 ECMA and sha512.  I can't do byte-by-byte because this is intended for very large files that can't both be in memory.
                  I forgot that some of the above timings include 2 methods of computation, checksum and sum32.
               Now called feq64.go, and will only compute crc64 ECMA checksum, in a way that only opens 1 file at a time.  And it doesn't need a bytes.Reader.
  21 May 22 -- Since I want to be able to process huge files > 3 GB, I can't read in the entire file at once.  I'll switch to a file reader algorithm.
                  Now called feq32, to play w/ the crc32 functions.  Since I intend this for huge files (> 3 GB), I'll use a flag to determine which crc32 to use.
  24 May 22 -- Now called feqbbb, for byte by byte.  I'll read in chunks and compare them, so very large files can be handled, too.
                 On leox, a 2.3 GB file comparison is ~1.6 s, about the same as feq32 -IEEE, and slightly slower than feq32 -cast which is ~1.4 s.
                 Using a 10 MB buffer instead of a 1 MB buffer is slower, at ~1.9 s.  Imagine that.
*/

const LastCompiled = "26 May 2022"
const K = 1024
const M = K * K

//* ************************* MAIN ***************************************************************
func main() {

	var filename1, filename2 string

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	// flag help message
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " file equal byte by byte tester, last modified %s, compiled with %s.\n", LastCompiled, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Filenames to compare are given on the command line.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "v", false, " verbose mode.")
	flag.Parse()

	if verboseFlag {
		fmt.Printf("\n feqbbb File equal using byte by byte comparisons of chunks, last modified %s, compiled by %s\n\n", LastCompiled, runtime.Version())
	}

	if flag.NArg() == 0 {
		fmt.Printf("\n Need two files on the command line to determine if they're equal.  Exiting. \n\n")
		os.Exit(1)
	} else if flag.NArg() >= 2 { // will use first 2 filenames entered on commandline
		filename1 = flag.Arg(0)
		filename2 = flag.Arg(1)
	} else {
		fmt.Printf("\n Need two files on the command line to determine if they're equal.  Exiting. ")
		os.Exit(1)
	}

	openedFile1, err1 := os.Open(filename1)
	check(err1, " Reading first file error is")
	defer openedFile1.Close()
	fi1, e1 := openedFile1.Stat()
	if e1 != nil {
		fmt.Fprintf(os.Stderr, " Opening %s and error is: %v.  Exiting. \n", filename1, e1)
		os.Exit(1)
	}
	fileReader1 := bufio.NewReader(openedFile1)

	openedFile2, err2 := os.Open(filename2)
	check(err2, " Reading 2nd file error is")
	defer openedFile2.Close()
	fi2, e2 := openedFile2.Stat()
	if e2 != nil {
		fmt.Fprintf(os.Stderr, " Opening %s and error is %v.  Exiting.\n", filename2, e2)
		os.Exit(1)
	}
	fileReader2 := bufio.NewReader(openedFile2)

	if fi1.Size() != fi2.Size() {
		fmt.Printf(" Files not equal as their sizes are not equal.  %s size is %d, and %s size is %d.\n", filename1, fi1.Size(), filename2, fi2.Size())
		os.Exit(1)
	}

	buf1 := make([]byte, 1*M) // I initially wrote this as ([]byte,0,M) so the slice was 0 bytes long.  This is what I need when using append() as in other code.
	buf2 := make([]byte, 1*M) // But it's completely wrong here.  The backing array of size M is irrelevant; the slice behaves as a buffer of length 0.

	t0 := time.Now()
	matched := true

	var counter int

outerLoop:
	for { //outer loop to refill the buffers
		n1, er1 := fileReader1.Read(buf1)
		n2, er2 := fileReader2.Read(buf2)
		counter++
		if er1 != nil || er2 != nil {
			if er1 == io.EOF || er2 == io.EOF {
				break outerLoop
			}
			fmt.Fprintf(os.Stderr, " File read errors.  %s err is %v, %s err is %v\n\n", filename1, er1, filename2, er2)
			break outerLoop
		}
		if n1 != n2 { // should never happen.  If file sizes are different should be caught above by the stat calls.
			matched = false
			fmt.Printf(" n1 is %d, n2 is %d\n", n1, n2)
			break outerLoop
		}
		if n1 == 0 || n2 == 0 { // I don't know if this will ever happen.
			break outerLoop
		}

		for i := range buf1 {
			if buf1[i] != buf2[i] {
				matched = false
				fmt.Printf(" First mismatching character is at position %d in loop %d.\n", i, counter)
				if verboseFlag {
					fmt.Printf(" The mismatched characters are %c in %s, and %c in %s.\n", buf1[i], filename1, buf2[i], filename2)
				}
				break outerLoop
			}
		}
	}

	if verboseFlag {
		fmt.Printf(" Outer loop counter is %d.  %s and %s are ", counter, filename1, filename2)
	} else {
		fmt.Print(" Files are ")
	}

	if matched {
		fmt.Print("equal")
	} else {
		fmt.Print("NOT equal")
	}
	fmt.Printf(".  Elapsed time is %s\n\n", time.Since(t0))

} // Main for feqbbb.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Fprintln(os.Stderr, msg, e)
	}
}
