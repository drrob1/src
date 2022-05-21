package main // for feq32.go

import (
	"bufio"
	"flag"
	"fmt"
	"hash"
	"hash/crc32"
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
*/

const LastCompiled = "21 May 2022"

const (
	ieeePoly = iota
	castPoly
	koopPoly
)

var polyNames = []string{"IEEE", "Castagnoli", "Koopman"} // this is a slice of strings, not an array of strings.  But this is a fine point that doesn't really matter.

//* ************************* MAIN ***************************************************************
func main() {

	var filename1, filename2 string

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	// flag help message
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " file equal tester only using 1 crc32 selected by a switch, last modified %s, compiled with %s.\n", LastCompiled, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Filenames to compare are given on the command line.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	var verboseFlag, ieeeFlag, castFlag, koopFlag bool
	var polynomial int // used to determine which one of the three to use.  Default will be IEEE, as it's the fastest and most widely used because of this.
	flag.BoolVar(&verboseFlag, "v", false, " verbose mode.")
	flag.BoolVar(&ieeeFlag, "ie", false, " IEEE polynomial flag.")
	flag.BoolVar(&castFlag, "cast", false, " Castagnoli polynomial flag.")
	flag.BoolVar(&koopFlag, "koop", false, "Koopman polynomial flag.")
	flag.Parse()

	if verboseFlag {
		fmt.Printf("\n feq32 File equal using 1 of 3 crc32 polynomials, last modified %s, compiled by %s\n\n", LastCompiled, runtime.Version())
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

	if ieeeFlag { // notice that this will be default if no polynomial flags are used.
		polynomial = ieeePoly
	} else if castFlag {
		polynomial = castPoly
	} else if koopFlag {
		polynomial = koopPoly
	}

	//fileByteSlice, err := os.ReadFile(filename1)  Needs too much memory if file is huge, like when > 3 or 4 GB.
	openedFile, err := os.Open(filename1)
	check(err, " Reading first file error is")
	fileReader := bufio.NewReader(openedFile)

	// compute the hash first for file 1, then for file 2, then compare them and output results.

	var crc32Table *crc32.Table
	var crc32Hash1, crc32Hash2 hash.Hash32

	// first file's first.
	if polynomial == ieeePoly {
		crc32Hash1 = crc32.NewIEEE()
		crc32Hash2 = crc32.NewIEEE()
	} else if polynomial == castPoly {
		crc32Table = crc32.MakeTable(crc32.Castagnoli)
		crc32Hash1 = crc32.New(crc32Table)
		crc32Hash2 = crc32.New(crc32Table)
	} else if polynomial == koopPoly {
		crc32Table = crc32.MakeTable(crc32.Koopman)
		crc32Hash1 = crc32.New(crc32Table)
		crc32Hash2 = crc32.New(crc32Table)
	} else {
		fmt.Fprintf(os.Stderr, " polynomial variable value not valid.  It is %d.  Exiting\n", polynomial)
		os.Exit(1)
	}

	t0 := time.Now()
	io.Copy(crc32Hash1, fileReader)
	crc32Val1 := crc32Hash1.Sum32()

	if verboseFlag {
		fmt.Printf(" file 1 %s: %s polynomial, hash1 = %x, elapsed time = %s\n\n", filename1, polyNames[polynomial], crc32Val1, time.Since(t0))
	}

	// second file's second, and then comparing the values.
	openedFile, err = os.Open(filename2)
	check(err, " Reading 2nd file error is")
	fileReader = bufio.NewReader(openedFile)

	// crc32 section for file 2
	io.Copy(crc32Hash2, fileReader)
	crc32Val2 := crc32Hash2.Sum32()

	if crc32Val1 == crc32Val2 {
		if verboseFlag {
			fmt.Printf(" crc32 %s polynomial values for %s and %s are equal.  Total elapsed time is %s.\n\n", polyNames[polynomial], filename1, filename2, time.Since(t0))
		} else {
			fmt.Printf(" crc32 %s polynomial values are equal.  Total elapsed time is %s.\n", polyNames[polynomial], time.Since(t0))
		}
	} else {
		fmt.Printf(" crc32 %s polynomial for the files are not equal.  %s = %x, %s = %x.  Total elapsed time is %s.\n\n",
			polyNames[polynomial], filename1, crc32Val1, filename2, crc32Val2, time.Since(t0))
	}

	if verboseFlag {
		fmt.Printf(" file 2 %s: crc32 %s polynomial = %x, total elapsed time = %s. \n\n",
			polyNames[polynomial], filename2, crc32Val2, time.Since(t0))
	}

	fmt.Println()
	fmt.Println()
} // Main for feq64.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Fprintln(os.Stderr, msg, e)
	}
}
