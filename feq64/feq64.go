package main // for feq64.go

import (
	"flag"
	"fmt"
	"hash/crc64"
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
*/

const LastCompiled = "12 Mar 2022"

//* ************************* MAIN ***************************************************************
func main() {

	var filename1, filename2 string

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	// flag help message
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " file equal tester only using crc64 ECMA, last modified %s, compiled with %s.\n", LastCompiled, runtime.Version())
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
		fmt.Printf("\n feq64 File equal only using crc64 ECMA, last modified %s, compiled by %s\n\n", LastCompiled, runtime.Version())
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

	fileByteSlice, err := os.ReadFile(filename1)
	check(err, " Reading first file error is")

	// now to compute the Castognoli, ECMA and sha512 hashes first for file 1, then for file 2, then compare them and output results.

	// first file's first.
	t0 := time.Now()
	crc64TableECMA := crc64.MakeTable(crc64.ECMA)
	crc64ECMAval1 := crc64.Checksum(fileByteSlice, crc64TableECMA)

	if verboseFlag {
		fmt.Printf(" file 1 %s: crc64 ECMA = %x, elapsed time = %s\n\n", filename1, crc64ECMAval1, time.Since(t0))
	}

	// second file's second, and then comparing the values.
	fileByteSlice, err = os.ReadFile(filename2)
	check(err, " Reading 2nd file error is")

	// crc64 ECMA section
	crc64ECMAval2 := crc64.Checksum(fileByteSlice, crc64TableECMA)
	if crc64ECMAval1 == crc64ECMAval2 {
		if verboseFlag {
			fmt.Printf(" crc64ECMA values for %s and %s are equal.\n\n", filename1, filename2)
		} else {
			fmt.Printf(" crc64 ECMA values are equal.\n")
		}
	} else {
		fmt.Printf(" crc64 ECMA for the files are not equal.  %s = %x, %s = %x.\n\n", filename1, crc64ECMAval1, filename2, crc64ECMAval2)
	}

	if verboseFlag {
		fmt.Printf(" file 2 %s: crc64 ECMA = %x, total elapsed time = %s \n\n",
			filename2, crc64ECMAval2, time.Since(t0))
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