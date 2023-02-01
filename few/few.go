package few // for few.go, to be used by main in ./cmd/few/main.go

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"hash/crc64"
	"io"
	"os"
)

/*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- M2:  First modified version of module.  I will use VLI to compare all digits of the hashes.
  23 Apr 13 -- Fixed problem of a single line in the hashes file, that does not contain an EOL character, causes
                an immediate return without processing of the characters just read in.
  24 Apr 13 -- Added output of which file either matches or does not match.
  13 Sep 16 -- Started conversion to Go.
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
  26 May 22 -- Now called feq1 and will only use the sha1 algorithm
  22 Jun 22 -- Adding color to the output.
  27 Jan 23 -- Now called few, because that's much easier than typing feq each time I want it.  This will consist of one top level routine and a sub cmd main pgm that
                 will take the hashes to run as params.  The API will use io.Readers.  The hash types are []byte, which cannot be directly compared.  So I convert the hashes
                 to strings, which can be directly compared.
   1 Feb 23 -- Adding API that takes filenames, not io.Reader's.
*/

const LastCompiled = "1 Feb 2023"

func Feq1withNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}
	buf1 := bufio.NewReader(fHandle1)
	buf2 := bufio.NewReader(fHandle2)

	return Feq1(buf1, buf2), nil
}

func Feq1(r1, r2 io.Reader) bool { // sha1
	sha1Hash1 := sha1.New()
	io.Copy(sha1Hash1, r1)
	sha1val1 := sha1Hash1.Sum(nil)
	sha1Str1 := hex.EncodeToString(sha1val1)

	sha1Hash2 := sha1.New()
	io.Copy(sha1Hash2, r2)
	sha1val2 := sha1Hash2.Sum(nil)
	sha1Str2 := hex.EncodeToString(sha1val2)

	return sha1Str1 == sha1Str2
} // feq1

func Feq2withNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}

	buf1 := bufio.NewReader(fHandle1)
	buf2 := bufio.NewReader(fHandle2)

	return Feq2(buf1, buf2), nil
}

func Feq2(r1, r2 io.Reader) bool { // sha256
	sha256Hash1 := sha256.New()
	io.Copy(sha256Hash1, r1)
	sha1val1 := sha256Hash1.Sum(nil)
	sha256Str1 := hex.EncodeToString(sha1val1)

	sha256Hash2 := sha256.New()
	io.Copy(sha256Hash2, r2)
	sha256val2 := sha256Hash2.Sum(nil)
	sha256Str2 := hex.EncodeToString(sha256val2)

	return sha256Str1 == sha256Str2
} // feq2

func Feq32withNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}

	buf1 := bufio.NewReader(fHandle1)
	buf2 := bufio.NewReader(fHandle2)

	return Feq32(buf1, buf2), nil
}

func Feq32(r1, r2 io.Reader) bool { // crc32 IEEE
	crc32Hash1 := crc32.NewIEEE()
	io.Copy(crc32Hash1, r1)
	crc32Val1 := crc32Hash1.Sum32() // This type is a uint32

	crc32Hash2 := crc32.NewIEEE()
	io.Copy(crc32Hash2, r2)
	crc32Val2 := crc32Hash2.Sum32()

	return crc32Val1 == crc32Val2
} // feq32

func Feq3withNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}

	buf1 := bufio.NewReader(fHandle1)
	buf2 := bufio.NewReader(fHandle2)

	return Feq3(buf1, buf2), nil
}

func Feq3(r1, r2 io.Reader) bool { // sha384
	sha384Hash1 := sha512.New384()
	io.Copy(sha384Hash1, r1)
	sha384val1 := sha384Hash1.Sum(nil)
	sha384Str1 := hex.EncodeToString(sha384val1)

	sha384Hash2 := sha512.New384()
	io.Copy(sha384Hash2, r2)
	sha384val2 := sha384Hash2.Sum(nil)
	sha384Str2 := hex.EncodeToString(sha384val2)

	return sha384Str1 == sha384Str2
} // feq3

func Feq5withNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}

	buf1 := bufio.NewReader(fHandle1)
	buf2 := bufio.NewReader(fHandle2)

	return Feq5(buf1, buf2), nil
}

func Feq5(r1, r2 io.Reader) bool { // sha512
	sha512Hash1 := sha512.New()
	io.Copy(sha512Hash1, r1)
	sha512val1 := sha512Hash1.Sum(nil)
	sha512Str1 := hex.EncodeToString(sha512val1)

	sha512Hash2 := sha512.New()
	io.Copy(sha512Hash2, r2)
	sha512val2 := sha512Hash2.Sum(nil)
	sha512Str2 := hex.EncodeToString(sha512val2)

	return sha512Str1 == sha512Str2
} // feq5

func Feq64withNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}

	buf1 := bufio.NewReader(fHandle1)
	buf2 := bufio.NewReader(fHandle2)

	return Feq64(buf1, buf2), nil
}

func Feq64(r1, r2 io.Reader) bool { // crc64
	// now to compute the ECMA 64-bit hash first for file 1, then for file 2, then compare them and output results.  The final hash is of type uint64
	crc64TableECMA := crc64.MakeTable(crc64.ECMA)
	crc64Hash1 := crc64.New(crc64TableECMA)
	io.Copy(crc64Hash1, r1)
	crc64ECMAVal1 := crc64Hash1.Sum64()

	crc64Hash2 := crc64.New(crc64TableECMA)
	io.Copy(crc64Hash2, r2)
	crc64ECMAVal2 := crc64Hash2.Sum64()

	return crc64ECMAVal1 == crc64ECMAVal2
} // feq64

func FeqbbbwithNames(fn1, fn2 string) (bool, error) {
	fHandle1, err := os.Open(fn1)
	defer fHandle1.Close()
	if err != nil {
		return false, err
	}
	fHandle2, err := os.Open(fn2)
	defer fHandle2.Close()
	if err != nil {
		return false, err
	}

	BOOL, err := Feqbbb(fHandle1, fHandle2)
	return BOOL, err
}

func Feqbbb(r1, r2 io.Reader) (bool, error) {
	const M = 1024 * 1024
	matched := true

	buf1 := make([]byte, 1*M) // in feqbbb, I initially wrote this as ([]byte,0,M) so the slice was 0 bytes long and this code didn't work because the slice behaved as a buffer of length 0.
	buf2 := make([]byte, 1*M) // Of course, the code is working w/ length = capacity = M.

outerForLoop:
	for { //outer loop to refill the buffers
		n1, er1 := r1.Read(buf1)
		n2, er2 := r2.Read(buf2)

		if er1 != nil || er2 != nil {
			if er1 == io.EOF || er2 == io.EOF {
				break outerForLoop
			}
			err := fmt.Errorf(" File read errors.  file-1 err is %s, file-2 err is %s", er1, er2)
			return false, err
		}
		if n1 != n2 {
			return false, nil
		}
		if n1 == 0 || n2 == 0 { // I don't know if this will ever happen.
			break outerForLoop
		}

		for i := range buf1 {
			if buf1[i] != buf2[i] {
				matched = false
				break outerForLoop
			}
		}
	}

	return matched, nil
} // feqbbb
