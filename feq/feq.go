package main // for feq.go

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
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

	// crc32 IEEE section.  1 and 3 are derived from file1, while 2 and 4 are derived from file2
	crc32ieeehash1 := crc32.NewIEEE()
	crc32ieeehash2 := crc32.NewIEEE()
	io.Copy(crc32ieeehash1, file1ByteReader)
	crc32ieeeVal1 := crc32ieeehash1.Sum(nil) // this should be []byte, but I don't know which endian it is.
	val1, _ := binary.Uvarint(crc32ieeeVal1)
	crc32ieeeVal3 := crc32.ChecksumIEEE(file1ByteSlice) // this should be uint32
	if val1 != uint64(crc32ieeeVal3) {
		fmt.Printf("\n crc32 IEEE code: val1=%x does not equal ieeeval3=%x\n\n", val1, crc32ieeeVal3)
	} else {
		fmt.Printf(" crc32 IEEE code val1 == val3.  crc32ieeeVal3=%x\n", crc32ieeeVal3)
	}

	io.Copy(crc32ieeehash2, file2ByteReader)
	crc32ieeeVal2 := crc32ieeehash2.Sum(nil) // this should be []byte, but I don't know which endian it is.
	val2, _ := binary.Uvarint(crc32ieeeVal2)
	crc32ieeeVal4 := crc32.ChecksumIEEE(file2ByteSlice) // this should be uint32
	if val2 == uint64(crc32ieeeVal4) {
		fmt.Printf(" crc32 IEEE code val2 == val4.  crc32ieeeVal4=%x\n")
	} else {
		fmt.Printf("\n crc32 IEEE code: val2=%x does not equal ieeeval4=%x  \n\n")
	}

	// I forgot to compare val1 and val2, which is the whole purpose of this exercise, after all.

	// crc32 Castagnoli polynomial section
	crc32TableCastagnoli := crc32.MakeTable(crc32.Castagnoli)
	crc32CastVal1 := crc32.Checksum(file1ByteSlice, crc32TableCastagnoli)
	crc32Casthash1 := crc32.New(crc32TableCastagnoli)
	io.Copy(crc32Casthash1, file1ByteReader)
	crc32CastVal3 := crc32Casthash1.Sum(nil)
	CastVal3, _ := binary.Uvarint(crc32CastVal3)

	crc32CastVal2 := crc32.Checksum(file2ByteSlice, crc32TableCastagnoli)
	crc32CastHash2 := crc32.New(crc32TableCastagnoli)
	io.Copy(crc32CastHash2, file2ByteReader)
	crc32CastVal4 := crc32CastHash2.Sum(nil)
	CastVal4, _ := binary.Uvarint(crc32CastVal4)
	if



	// crc32 Koopman polynomial section
	crc32TableKoopman := crc32.MakeTable(crc32.Koopman)
	crc32Koopman1 := crc32.New(crc32TableKoopman)
	crc32Koopman2 := crc32.New(crc32TableKoopman)

	// crc64 ECMA section
	crc64TableECMA := crc64.MakeTable(crc64.ECMA)
	crc64Hash1 := crc64.New(crc64TableECMA)
	crc64Hash2 := crc64.New(crc64TableECMA)

	// md5 section
	md5hash1 := md5.New()
	md5hash2 := md5.New()

	fileSize1, err := io.Copy(md5hash1, file1ByteReader)
	check(err, "md5 hash1 io.copy err is ")
	fileSize2, err3 := io.Copy(md5hash2, file2ByteReader)
	check(err3, " md5 hash2 io.copy error is ")
	hashValueComputedStr1 := hex.EncodeToString(md5hash1.Sum(nil))
	hashValueComputedStr2 := hex.EncodeToString(md5hash2.Sum(nil))
	fmt.Printf(" md5 results: filesizes are %d = %d; hash strings are %s:%s, and are", fileSize1, fileSize2, hashValueComputedStr1, hashValueComputedStr2)
	if hashValueComputedStr1 != hashValueComputedStr2 {
		fmt.Printf(" NOT")
	}
	fmt.Printf(" equal.\n\n")

	// sha1 section
	sha1hash1 := sha1.New()
	sha1hash2 := sha1.New()

	fileSize1, err = io.Copy(sha1hash1, file1ByteReader)
	check(err, "sha1 hash1 io.copy err is ")
	fileSize2, err3 = io.Copy(sha1hash2, file2ByteReader)
	check(err3, " sha1 hash2 io.copy err is ")
	hashValueComputedStr1 = hex.EncodeToString(sha1hash1.Sum(nil))
	hashValueComputedStr2 = hex.EncodeToString(sha1hash2.Sum(nil))
	fmt.Printf(" sha1 results: filesizes are %d = %d; hash strings are %s:%s, and are", fileSize1, fileSize2, hashValueComputedStr1, hashValueComputedStr2)
	if hashValueComputedStr1 != hashValueComputedStr2 {
		fmt.Printf(" NOT")
	}
	fmt.Printf(" equal.\n\n")

	// sha256 section
	sha256hash1 := sha256.New()
	sha256hash2 := sha256.New()

	// sha512 section
	sha512hash1 := sha512.New()
	sha512hash2 := sha512.New()

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