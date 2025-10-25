package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"src/getcommandline" // this ones mine.
	"hash"
	"io"
	"os"
	"runtime"
)

/*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- First modified version of module.  I will use VLI to compare all digits of the hashes.
  23 Apr 13 -- Fixed problem of a single line in the hashes file, that does not contain an EOL character, causes
                an immediate return without processing of the characters just read in.
  24 Apr 13 -- Added output of which file either matches or does not match.
  13 Sep 16 -- Started conversion to Go.  Added the removed of '*' which is part of a std linux formated hash file.  And I forgot that
                 the routine allowed either order in the file.  If the token has a '.' I assume it is a filename, else it is a hash value.
*/

//* ************************* MAIN ***************************************************************
func main() {

	const K = 1024
	const M = 1024 * 1024

	const (
		md5hash = iota
		sha1hash
		sha256hash
		sha384hash
		sha512hash
		HashType
	)

	const ReadBufferSize = M
	//  const ReadBufferSize = 10 * M;

	var HashName = [...]string{"md5", "sha1", "sha256", "sha384", "sha512"}
	var WhichHash int
	var hasher hash.Hash
	var FileSize int64

	if len(os.Args) <= 1 {
		fmt.Println(" Need input filename as a param. ")
		os.Exit(0)
	}
	FileToHash := getcommandline.GetCommandLineString()

	fmt.Println()
	fmt.Print(" GOOS =", runtime.GOOS, ".  ARCH=", runtime.GOARCH)
	fmt.Println("  WhichHash = ", HashName[WhichHash])
	fmt.Println()
	fmt.Println()

	for {
		FileSize = 0

		/* Create Hash Section */
		TargetFile, readErr := os.Open(FileToHash)
		check(readErr, " Error opening FileToHash.")
		defer TargetFile.Close()

		switch WhichHash { // Initialing case switch on WhichHash
		case md5hash:
			hasher = md5.New()
		case sha1hash:
			hasher = sha1.New()
		case sha256hash:
			hasher = sha256.New()
		case sha384hash:
			hasher = sha512.New384()
		case sha512hash:
			hasher = sha512.New()
		default:
			hasher = sha256.New()
		} /* initializing case on WhichHash */

		//    FileReadBuffer := make([]byte,ReadBufferSize);
		/*
		   for {   // Repeat Until eof loop.
		     n,err := TargetFile.Read(FileReadBuffer);
		     if n == 0 || err == io.EOF { break }
		     check(err," Unexpected error while reading the target file on which to compute the hash,");
		     hasher.Write(FileReadBuffer[:n]);
		     FileSize += int64(n);
		   } // Repeat Until TargetFile.eof loop;
		*/
		// If I understand a response to my post for help correctly, I can do this instead of the for loop
		n, err := io.Copy(hasher, TargetFile)
		check(err, " Unexpected error from io.Copy")
		FileSize += int64(n)

		HashValueComputedStr := hex.EncodeToString(hasher.Sum(nil))

		fmt.Println(" Filename  = ", FileToHash, ", FileSize = ", FileSize, ", ", HashName[WhichHash])
		fmt.Println(" Computed hash hex encoded:", HashValueComputedStr)

		TargetFile.Close() // Close the handle to allow opening a target from the next line, if there is one.
		fmt.Println()
		fmt.Println()

		WhichHash++
		if WhichHash > sha512hash {
			break
		}
	} /* outer LOOP */

	fmt.Println()
} // Main for comparehashes.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Errorf("%s : ", msg)
		panic(e)
	}
}
