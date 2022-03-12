package main // for feq.go

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/crc32"
	"hash/crc64"
	"io"
	"os"
	"runtime"
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
		fmt.Fprintf(flag.CommandLine.Output(), " file equal tester, last modified %s.\n", LastCompiled)
		fmt.Fprintf(flag.CommandLine.Output(), " Filenames to hash and compare are given on the command line.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, workingDir, execName)
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	var verboseFlag bool
	flag.BoolVar(&verboseFlag, "v", false, " verbose mode.")
	flag.Parse()

	if verboseFlag {
		fmt.Printf("\n feq File equal determination, last modified %s, compiled by %s\n\n", LastCompiled, runtime.Version())
	}

	if flag.NArg() == 0 { // need to use filepicker, or not.
		fmt.Printf("\n Need two files on the command line to determine if they're equal.  Exiting. \n\n")
		os.Exit(1)
		/*
			filenames, err := filepicker.GetFilenames("*.sha*")
			if err != nil {
				fmt.Fprintln(os.Stderr, " Error from filepicker is", err)
				os.Exit(1)
			}
			for i := 0; i < min(len(filenames), 20); i++ {
				fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
			}
			fmt.Print(" Enter filename choice : ")
			n, err := fmt.Scanln(&ans)
			if n == 0 || err != nil {
				ans = "0"
			} else if ans == "999" {
				fmt.Println(" Stop code entered.  Exiting.")
				os.Exit(0)
			}
			i, err := strconv.Atoi(ans)
			if err == nil && i < len(filenames) {
				filename1 = filenames[i]
			} else {
				s := strings.ToUpper(ans)
				s = strings.TrimSpace(s)
				s0 := s[0]
				i = int(s0 - 'A')
				if i < len(filenames) {
					filename1 = filenames[i]
				}
			}
			if len(filename1) == 0 { // if entered choice is out of range, switch to use 0.  It's inelegant to panic.
				filename1 = filenames[0]
			}
			fmt.Println(" Picked filename is", filename1)

		*/
	} else { // will use filename entered on commandline
		filename1 = flag.Arg(0)
		filename2 = flag.Arg(1)
		if len(filename1) == 0 || len(filename2) == 0 {
			fmt.Printf("\n Need two files on the command line to determine if they're equal.  Exiting. ")
			os.Exit(1)
		}
	}

	fmt.Println()

	file1ByteSlice, err1 := os.ReadFile(filename1)
	file1ByteReader := bytes.NewReader(file1ByteSlice)
	file2ByteSlice, err2 := os.ReadFile(filename2)
	file2ByteReader := bytes.NewReader(file2ByteSlice)

	if err1 != nil || err2 != nil {
		fmt.Fprintf(os.Stderr, "\nos.ReadFile: %s error is %v, and %s error is %v\n\n", filename1, err1, filename2, err2)
		os.Exit(1)
	}

	//   now to compute the hashes,  compare them, and output results

	// crc32 IEEE section.  1a uses Sum32, 1b uses Checksum and 1c uses Sum32.
	crc32ieeehash1 := crc32.NewIEEE()
	size1, _ := io.Copy(crc32ieeehash1, file1ByteReader)
	crc32IEEEval1a := crc32ieeehash1.Sum32()
	crc32IEEEVal1b := crc32.ChecksumIEEE(file1ByteSlice)
	if crc32IEEEval1a == crc32IEEEVal1b {
		if verboseFlag {
			fmt.Printf(" crc32 IEEE code val1a == val1b.  crc32ieeeVal1b = %x\n\n", crc32IEEEVal1b)
		}
	} else {
		fmt.Printf("\n crc32 IEEE code: val1a = %x does not equal ieeeval1b = %x\n\n", crc32IEEEval1a, crc32IEEEVal1b)
	}

	// crc32 IEEE section.  2a uses Sum32, 2b uses Checksum.
	crc32ieeehash2 := crc32.NewIEEE()
	size2, _ := io.Copy(crc32ieeehash2, file2ByteReader)
	if verboseFlag {
		fmt.Printf(" io.Copy size1 = %d, size2 = %d, type of 1a is %T, type of 1b is %T \n", size1, size2, crc32IEEEval1a, crc32IEEEVal1b)
	}

	crc32ieeeVal2a := crc32ieeehash2.Sum32()
	crc32IEEEval2b := crc32.ChecksumIEEE(file2ByteSlice)
	if crc32ieeeVal2a == crc32IEEEval2b {
		if verboseFlag {
			fmt.Printf(" crc32 IEEE code val2a == val2b.  crc32ieeeVal2b = %x\n\n", crc32IEEEval2b)
		}
	} else {
		fmt.Printf("\n crc32 IEEE code: crc32IEEEval2a = %x does not equal crc32IEEEval2b = %x  \n\n", crc32ieeeVal2a, crc32IEEEval2b)
	}

	if crc32IEEEVal1b == crc32IEEEval2b {
		fmt.Printf(" IEEE CRC32 Checksum for %s and %s are equal.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" IEEE CRC32 CheckSums are not equal.  %s = %x and %s = %x are not equal.\n\n", filename1, crc32IEEEVal1b, filename2, crc32IEEEval2b)
	}

	// crc32 Castagnoli polynomial section
	crc32TableCastagnoli := crc32.MakeTable(crc32.Castagnoli)
	crc32CastVal1b := crc32.Checksum(file1ByteSlice, crc32TableCastagnoli)
	crc32Casthash1 := crc32.New(crc32TableCastagnoli)
	file1ByteReader.Reset(file1ByteSlice)
	size, err := io.Copy(crc32Casthash1, file1ByteReader)
	check(err, " io.Copy crc32Casthash1 error is")
	crc32CastVal1a := crc32Casthash1.Sum32()
	if crc32CastVal1a == crc32CastVal1b {
		if verboseFlag {
			fmt.Printf(" crc32 Castagnoli 1 a and b are equal.  The value is %x.\n\n", crc32CastVal1a)
		}
	} else {
		fmt.Printf(" crc32 Castagnoli 1 a and b are not equal.  a = %x, b = %x and size = %d\n\n", crc32CastVal1a, crc32CastVal1b, size)
	}

	crc32CastVal2b := crc32.Checksum(file2ByteSlice, crc32TableCastagnoli)
	crc32CastHash2 := crc32.New(crc32TableCastagnoli)
	file2ByteReader.Reset(file2ByteSlice)
	size, err = io.Copy(crc32CastHash2, file2ByteReader)
	check(err, "io.Copy crc32CastHash2 error is")
	crc32CastVal2a := crc32CastHash2.Sum32()
	if crc32CastVal2a == crc32CastVal2b {
		if verboseFlag {
			fmt.Printf(" crc32 Castagnoli 2 a and b are equal.  The value is %x.\n\n", crc32CastVal2a)
		}
	} else {
		fmt.Printf(" crc32 Castagnoli 2 a and b are not equal.  a = %x, b = %x and size = %d\n\n", crc32CastVal2a, crc32CastVal2b, size)
	}

	if crc32CastVal1b == crc32CastVal2b {
		fmt.Printf(" crc32 Castagnoli for %s and %s are equal.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" crc32 Castagnoli for %s and %s are not equal.\n\n", filename1, filename2)
	}
	if verboseFlag {
		fmt.Printf(" crc32 Castagnoli 1 is %x for %s; Castagnoli 2 is %x for %s\n\n", crc32CastVal1b, filename1, crc32CastVal2b, filename2)
	}

	// crc32 Koopman polynomial section
	crc32TableKoopman := crc32.MakeTable(crc32.Koopman)
	crc32KoopVal1 := crc32.Checksum(file1ByteSlice, crc32TableKoopman)
	crc32KoopVal2 := crc32.Checksum(file2ByteSlice, crc32TableKoopman)
	if crc32KoopVal1 == crc32KoopVal2 {
		fmt.Printf(" crc32Koopman for %s and %s are equal.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" crc32 Koopman for %s and %s are not equal.\n\n", filename1, filename2)
	}
	if verboseFlag {
		fmt.Printf(" crc32 Koopman 1 is %x for %s; Koopman 2 is %x for %s\n\n", crc32KoopVal1, filename1, crc32KoopVal2, filename2)
	}

	// crc64 ECMA section
	crc64TableECMA := crc64.MakeTable(crc64.ECMA)
	crc64ECMAval1 := crc64.Checksum(file1ByteSlice, crc64TableECMA)
	crc64ECMAval2 := crc64.Checksum(file2ByteSlice, crc64TableECMA)
	if crc64ECMAval1 == crc64ECMAval2 {
		fmt.Printf(" crc64ECMA values for %s and %s are equal.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" crc64 ECMA for the files are not equal.  %s = %x, %s = %x.\n\n", filename1, filename2)
	}
	if verboseFlag {
		fmt.Printf(" crc64 ECMA 1 is %x for %s; ECMA 2 is %x for %s\n\n", crc64ECMAval1, filename1, crc64ECMAval2, filename2)
	}

	// md5 section
	md5hash1 := md5.New()
	md5hash2 := md5.New()

	file1ByteReader.Reset(file1ByteSlice)
	file2ByteReader.Reset(file2ByteSlice)
	fileSize1, err2 := io.Copy(md5hash1, file1ByteReader)
	check(err2, "md5 hash1 io.copy err is ")
	fileSize2, err3 := io.Copy(md5hash2, file2ByteReader)
	check(err3, " md5 hash2 io.copy error is ")
	hashValueComputedStr1 := hex.EncodeToString(md5hash1.Sum(nil))
	hashValueComputedStr2 := hex.EncodeToString(md5hash2.Sum(nil))
	if hashValueComputedStr1 != hashValueComputedStr2 {
		fmt.Printf(" md5 results: %s does not equal %s.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" md5 results: %s equals %s.\n\n", filename1, filename2)
	}
	if verboseFlag {
		fmt.Printf(" md5 for %s is %s; for %s is %s.  FileSize1 = %d, filesize2 = %d\n\n", filename1, hashValueComputedStr1, filename2, hashValueComputedStr2, fileSize1, fileSize2)
	}

	// sha1 section
	sha1hash1 := sha1.New()
	sha1hash2 := sha1.New()

	file1ByteReader.Reset(file1ByteSlice)
	file2ByteReader.Reset(file2ByteSlice)
	fileSize1, err = io.Copy(sha1hash1, file1ByteReader)
	check(err, "sha1 hash1 io.copy err is ")
	fileSize2, err3 = io.Copy(sha1hash2, file2ByteReader)
	check(err3, " sha1 hash2 io.copy err is ")
	hashValueComputedStr1 = hex.EncodeToString(sha1hash1.Sum(nil))
	hashValueComputedStr2 = hex.EncodeToString(sha1hash2.Sum(nil))
	if hashValueComputedStr1 != hashValueComputedStr2 {
		fmt.Printf(" sha1 results: %s does not equal %s.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" sha1 results: %s equals %s.\n\n", filename1, filename2)
	}
	if verboseFlag {
		fmt.Printf(" sha1 results: for %s is %s; for %s is %s; filesize1= %d, filesize2= %d.\n\n", filename1, hashValueComputedStr1, filename2, hashValueComputedStr2,
			fileSize1, fileSize2)
	}

	// sha256 section
	sha256hash1 := sha256.New()
	file1ByteReader.Reset(file1ByteSlice)
	fileSize1, err = io.Copy(sha256hash1, file1ByteReader)
	check(err, "sha256 hash1 io.copy err is")
	hashValueComputedStr1 = hex.EncodeToString(sha256hash1.Sum(nil))
	sha256hash2 := sha256.New()
	file2ByteReader.Reset(file2ByteSlice)
	fileSize2, err = io.Copy(sha256hash2, file2ByteReader)
	check(err, "sha256 hash2 io.copy err is")
	hashValueComputedStr2 = hex.EncodeToString(sha256hash2.Sum(nil))
	if hashValueComputedStr1 != hashValueComputedStr2 {
		fmt.Printf(" sha256 result: %s NOT equal %s.\n\n", filename1, filename2)
	} else {
		fmt.Printf(" sha256 result: %s equal to %s.\n\n", filename1, filename2)
	}
	if verboseFlag {
		fmt.Printf(" sha256 result: %s = %s; %s = %s; filesize1 = %d, filesize2 = %d.\n\n", filename1, hashValueComputedStr1, filename2, hashValueComputedStr2,
			fileSize1, fileSize2)
	}

	// sha512 section
	sha512hash1 := sha512.New()
	file1ByteReader.Reset(file1ByteSlice)
	fileSize1, err = io.Copy(sha512hash1, file1ByteReader)
	check(err, "sha512 hash1 io.copy err is")
	hashValueComputedStr1 = hex.EncodeToString(sha512hash1.Sum(nil))
	sha512hash2 := sha512.New()
	file2ByteReader.Reset(file2ByteSlice)
	fileSize2, err = io.Copy(sha512hash2, file2ByteReader)
	check(err, "sha512 hash2 io.copy err is")
	hashValueComputedStr2 = hex.EncodeToString(sha512hash2.Sum(nil))
	if hashValueComputedStr1 != hashValueComputedStr2 {
		fmt.Printf("NOT equal to hash string 2 of %s\n\n", hashValueComputedStr2)
	} else {
		fmt.Printf("equal to hash string 2.\n\n")
	}
	if verboseFlag {
		fmt.Printf(" sha512 results: %s is %s; %s is %s; filesize1= %d; filesize2= %d\n\n", filename1, hashValueComputedStr1, filename2, hashValueComputedStr2,
			fileSize1, fileSize2)
	}

	// byte-by-byte section
	var matched bool
	if len(file1ByteSlice) == len(file2ByteSlice) {
		for i := range file1ByteSlice {
			if file1ByteSlice[i] != file2ByteSlice[i] {
				matched = false
				break
			}
			matched = true
		}
	}
	fmt.Printf(" Byte-by-byte method: %s and %s ", filename1, filename2)
	if matched {
		fmt.Printf("are equal.\n")
	} else {
		fmt.Printf("are not equal.\n")
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

/*
// ------------------------------------------------------- min ---------------------------------
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}


*/
