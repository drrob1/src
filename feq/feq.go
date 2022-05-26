package main // for feq.go

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/crc32"
	"hash/crc64"
	"io"
	"os"
	"runtime"
	"time"
)

/*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- M2:  First modified version of module.  I will use VLI to compare all digits of the hashes.  And it's called CompareHashes.mod
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
  24 Mar 22 -- Adding b flag to do a byte by byte comparison but by using bufio to open both files.  Thinking this will work even on huge files.
  26 Mar 22 -- Tweaked output when the hashes don't match.
  20 May 22 -- Want to add timing info, but here it's not easy to time them all separately, so I'll have to do it all combined.
  21 May 22 -- Timing will now not be per file, but for both files.
  26 May 22 -- Adding check for filesize and if too big, will abort.
*/

const LastCompiled = "26 May 2022"

//const tooBig = 2_000_000_000
const tooBig = 2e9

//* ************************* MAIN ***************************************************************
func main() {

	var filename1, filename2 string

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	// flag help message
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " file equal tester for very large files, last modified %s, compiled with %s.\n", LastCompiled, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Filenames to compare are given on the command line.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	var verboseFlag, byteByByteFlag bool
	flag.BoolVar(&verboseFlag, "v", false, " verbose mode flag.")
	flag.BoolVar(&byteByByteFlag, "b", false, " byte by byte comparison flag.")
	flag.Parse()

	if verboseFlag {
		fmt.Printf("\n feqlarge File equal for LARGE files, last modified %s, compiled by %s\n\n", LastCompiled, runtime.Version())
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

	fi1, err1 := os.Stat(filename1)
	check(err1, " Error calling Stat on file1 is")
	fi2, err2 := os.Stat(filename2)
	check(err2, " Error calling Stat on file2 is")
	if fi1.Size() > tooBig || fi2.Size() > tooBig {
		fmt.Fprintf(os.Stderr, " Either size of %s or %s or both is > %2g.  Need feq1, feq32, feq64 or feqbbb.\n\n", filename1, filename2, tooBig)
		os.Exit(1)
	}

	fileByteSlice, err := os.ReadFile(filename1)
	check(err, " Reading first file error is")

	// now to compute the Castognoli, ECMA and sha512 hashes first for file 1, then for file 2, then compare them and output results.

	t1 := time.Now()
	// first file's first.
	crc32TableCastagnoli := crc32.MakeTable(crc32.Castagnoli)
	t0 := time.Now()
	crc32CastVal1 := crc32.Checksum(fileByteSlice, crc32TableCastagnoli)
	crc32Duration := time.Since(t0)

	crc64TableECMA := crc64.MakeTable(crc64.ECMA)
	t0 = time.Now()
	crc64ECMAval1 := crc64.Checksum(fileByteSlice, crc64TableECMA)
	crc64Duration := time.Since(t0)

	fileByteReader := bytes.NewReader(fileByteSlice)
	sha512hash1 := sha512.New()
	t0 = time.Now()
	fileSize1, er := io.Copy(sha512hash1, fileByteReader)
	sha512Duration := time.Since(t0)
	check(er, "sha512 hash1 io.copy err is")
	sha512ValueComputedStr1 := hex.EncodeToString(sha512hash1.Sum(nil))

	if verboseFlag {
		fmt.Printf(" file 1 %s: crc32 Cast = %x, crc64 ECMA = %x, file size = %d, elapsed time = %s, sha512 = %s\n",
			filename1, crc32CastVal1, crc64ECMAval1, fileSize1, time.Since(t0), sha512ValueComputedStr1)
	}

	// second file's next, and then comparing the values.
	fileByteSlice, er = os.ReadFile(filename2)
	check(er, " Reading 2nd file error is")

	t0 = time.Now()
	crc32CastVal2 := crc32.Checksum(fileByteSlice, crc32TableCastagnoli)
	crc32Duration += time.Since(t0)

	if crc32CastVal1 == crc32CastVal2 {
		if verboseFlag {
			fmt.Printf(" crc32 Castagnoli for %s and %s are equal.  Time for both files is %s.  \n\n", filename1, filename2, crc32Duration.String())
		} else {
			fmt.Printf(" crc32 Castagnoli hashes are equal.  Time for both files is %s.\n", crc32Duration.String())
		}
	} else {
		fmt.Printf(" crc32 Castagnoli values are not equal.  File 1: %s = %x;  File 2: %s = %x.  Time for both files is %s.\n\n",
			filename1, crc32CastVal1, filename2, crc32CastVal2, crc32Duration.String())
	}

	// crc64 ECMA section
	t0 = time.Now()
	crc64ECMAval2 := crc64.Checksum(fileByteSlice, crc64TableECMA)
	crc64Duration += time.Since(t0)
	if crc64ECMAval1 == crc64ECMAval2 {
		if verboseFlag {
			fmt.Printf(" crc64ECMA values for %s and %s are equal.  Time for both files is %s.\n\n", filename1, filename2, crc64Duration.String())
		} else {
			fmt.Printf(" crc64 ECMA values are equal.  Time for both files is %s.\n", crc64Duration.String())
		}
	} else {
		fmt.Printf(" crc64 ECMA for the files are not equal.  %s = %x, %s = %x.  Time for both files is %s.\n\n",
			filename1, crc64ECMAval1, filename2, crc64ECMAval2, crc64Duration.String())
	}

	// sha512 section
	sha512hash2 := sha512.New()
	fileByteReader = bytes.NewReader(fileByteSlice)
	t0 = time.Now()
	fileSize2, err := io.Copy(sha512hash2, fileByteReader)
	sha512Duration += time.Since(t0)
	check(err, "sha512 hash2 io.copy err is")
	sha512ValueComputedStr2 := hex.EncodeToString(sha512hash2.Sum(nil))
	if sha512ValueComputedStr1 == sha512ValueComputedStr2 {
		if verboseFlag {
			fmt.Printf("sha512 results: %s equal to %s.  Time for both files is %s.\n", filename1, filename2, sha512Duration.String())
		} else {
			fmt.Printf(" sha512 values are equal.  Time for both files is %s.\n", sha512Duration.String())
		}
	} else {
		fmt.Printf("sha512 results: %s NOT equal to %s.  Time for both files is %s.\n", filename1, filename2, sha512Duration.String())
	}
	if verboseFlag {
		fmt.Printf(" file 1 %s: crc32 Cast = %x, crc64 ECMA = %x, filesize = %d, \n sha512 = %s\n", filename1, crc32CastVal1, crc64ECMAval1, fileSize1, sha512ValueComputedStr2)
		fmt.Printf(" file 2 %s: crc32 Cast = %x, crc64 ECMA = %x, filesize = %d,  \n sha512 = %s\n\n",
			filename2, crc32CastVal2, crc64ECMAval2, fileSize2, sha512ValueComputedStr2)
	}
	fmt.Printf(" Entire run took %s\n", time.Since(t1))

	// Comparing byte by byte, if requested by the b flag.
	if byteByByteFlag {
		file1, e := os.Open(filename1)
		check(e, " byteByByte section and opening file1 error is")
		file2, errr := os.Open(filename2)
		check(errr, " bytebybyte section and opening file2 error is")

		fReader1 := bufio.NewReader(file1)
		fReader2 := bufio.NewReader(file2)

		var matched bool

		t2 := time.Now()

		for {
			b1, err1 := fReader1.ReadByte()
			b2, err2 := fReader2.ReadByte()

			if err1 != nil || err2 != nil { // should mostly be EOF condition.
				break
			}
			if b1 != b2 {
				matched = false
				break
			}
			matched = true
		}
		fmt.Printf(" Byte by byte comparison result is %t for %s and %s, taking %s\n", matched, filename1, filename2, time.Since(t2))
	}

	fmt.Println()
	fmt.Println()
} // Main for feq.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Fprintln(os.Stderr, msg, e)
	}
}
