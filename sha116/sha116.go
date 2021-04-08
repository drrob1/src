package main   // for sha116.go

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"runtime"
	"strconv"
	"strings"
	"src/filepicker"
	"src/tknptr"
)

/*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- First modified version of module.  I will use VLI to compare all digits of the hashes.
  23 Apr 13 -- Fixed problem of a single line in the hashes file, that does not contain an EOL character, causes
                an immediate return without processing of the characters just read in.
  24 Apr 13 -- Added output of which file either matches or does not match.
  19 Sep 16 -- Finished conversion to Go, that was started 13 Sep 16.  Added the removed of '*' which is part of a std linux formated hash file.  And I forgot that
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
*/

const LastCompiled = "8 Apr 2021"

//* ************************* MAIN ***************************************************************
func main() {

	const K = 1024
	const M = 1024 * 1024

	const (
		undetermined = iota
		md5hash
		sha1hash
		sha256hash
		sha384hash
		sha512hash
	)

	const ReadBufferSize = M

	var HashName = [...]string{"undetermined", "md5", "sha1", "sha256", "sha384", "sha512"}
	var ans, Filename string
	var WhichHash int
	var TargetFilename, HashValueReadFromFile, HashValueComputedStr string
	var hasher hash.Hash
	var FileSize int64

	fmt.Print(" sha.go.  GOOS =", runtime.GOOS, ".  ARCH=", runtime.GOARCH)
	fmt.Println(".  Last altered", LastCompiled, ", compiled using", runtime.Version())
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf("%s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n", ExecFI.Name(), LastLinkedTimeStamp, workingdir, execname)
	fmt.Println()


	// filepicker stuff.

	if len(os.Args) <= 1 { // need to use filepicker
		filenames := filepicker.GetFilenames("*.sha*")
		for i := 0; i < min(len(filenames), 20); i++ {
			fmt.Println("filename[", i, "] is", filenames[i])
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
		if err == nil {
			Filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			Filename = filenames[i]
		}
		fmt.Println(" Picked filename is", Filename)
	} else { // will use filename entered on commandline
		//            Filename = getcommandline.GetCommandLineString()  removed 3/3/21, as os.Args is fine.
                Filename = os.Args[1]
	}

	fmt.Println()


	// Now ignores extension, always going by hash length.


	// Read and parse the file with the hashes.

        filebyteslice := make([]byte, 0, 2000)
        filebyteslice, err := os.ReadFile(Filename)
        if os.IsNotExist(err) {
            fmt.Println(Filename, " does not exist.  Exiting.")
            os.Exit(1)
        } else if err != nil {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
        bytesbuffer := bytes.NewBuffer(filebyteslice)

	for { /* to read multiple lines */
		FileSize = 0
		WhichHash = undetermined  // reset it for this next line, allowing multiple types of hashes in same file.

		inputline, err := bytesbuffer.ReadString('\n')
		if err == io.EOF && len(inputline) == 0 { // reached EOF condition, there are no more lines to read, and no line
			break
		} else if len(inputline) == 0 && err != nil {
			fmt.Fprintln(os.Stderr, "While reading from the HashesFile:", err)
			os.Exit(1)
		}
		inputline = strings.TrimSpace(inputline)  // trims off the trailing newline

		if strings.HasPrefix(inputline, ";") || strings.HasPrefix(inputline, "#") || (len(inputline) <= 10) {
			continue
		} // allow comments and essentially blank lines

		tokenPtr := tknptr.NewToken(inputline)
		tokenPtr.SetMapDelim('*')
		FirstToken, EOL := tokenPtr.GetTokenString(false)
		if EOL {
			fmt.Fprintln(os.Stderr," EOL while getting 1st token in the hashing file.  Skipping to next line.")
			continue
		}
		hashlength := 0

		if strings.ContainsRune(FirstToken.Str, '.') || strings.ContainsRune(FirstToken.Str, '-') ||
			strings.ContainsRune(FirstToken.Str, '_') { // have filename first on line
			TargetFilename = FirstToken.Str
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get hash string from the line in the file
			if EOL {
				fmt.Fprintln(os.Stderr, " Got EOL while getting HashValue (2nd) token in the hashing file.  Skipping")
				continue
			}
			HashValueReadFromFile = SecondToken.Str
			hashlength = len(SecondToken.Str)

		} else { // have hash first on line
			HashValueReadFromFile = FirstToken.Str
			hashlength = len(FirstToken.Str)
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get name of file on which to compute the hash
			if EOL {
				fmt.Fprintln(os.Stderr, " EOL while gatting TargetFilename token in the hashing file.  Skipping")
				continue
			}

			if strings.ContainsRune(SecondToken.Str, '*') { // If it contains a *, it will be the first position.
				SecondToken.Str = SecondToken.Str[1:]
				if strings.ContainsRune(SecondToken.Str, '*') { // this should not happen
					fmt.Println(" Filename token still contains a * character.  Str:", SecondToken.Str, "  Skipping.")
					continue
				}
			}
			TargetFilename = (SecondToken.Str)
		} /* if have filename first or hash value first */


		//   now to compute the hash, compare them, and output results

		// Create Hash Section 
		TargetFile, err := os.Open(TargetFilename)
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, TargetFilename, " does not exist.  Skipping.")
			continue
		} else { // we know that the file exists
			check(err, " Error opening TargetFilename.")
		}

		defer TargetFile.Close()

		if WhichHash == undetermined {
			if hashlength == 64 {
				WhichHash = sha256hash
			} else if hashlength == 128 {
				WhichHash = sha512hash
			} else if hashlength == 40 {
				WhichHash = sha1hash
			} else if hashlength == 96 {
				WhichHash = sha384hash
			} else if hashlength == 32 {
				WhichHash = md5hash
			} else {
				fmt.Fprintln(os.Stderr, " Could not determine hash type for file.  Skipping.")
				continue
			}
			fmt.Println(" hash determined by length to be", HashName[WhichHash])
			fmt.Println()
		}

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
		}

		FileSize, err = io.Copy(hasher, TargetFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "skipped.")
			continue
		}
		HashValueComputedStr = hex.EncodeToString(hasher.Sum(nil))

		// I got the idea to use the different base64 versions and my own hex converter code, just to see.
		// And I can also use Sfprintf with the %x verb.  base64 versions are not useful as they use a larger
		// character set than hex.  I deleted all references to the base64 versions.  And the hex encoded and
		// Sprintf using %x were the same, so I removed the sprintf code.
		//    HashValueComputedSprintf := fmt.Sprintf("%x",hasher.Sum(nil));

		fmt.Printf(" Filename  = %s, filesize = %d, using hash %s.\n", TargetFilename, FileSize, HashName[WhichHash])
		fmt.Println("       Read From File:", HashValueReadFromFile)
		fmt.Println(" Computed hex encoded:", HashValueComputedStr)

		if strings.ToLower(HashValueReadFromFile) == strings.ToLower(HashValueComputedStr) {
			fmt.Print(" Matched.")
		} else {
			fmt.Print(" Not matched.")
		} /* if hashes */
		TargetFile.Close() // Close the handle to allow opening a target from the next line, if there is one.
		fmt.Println()
		fmt.Println()
	} // outer LOOP to read multiple lines
} // Main for sha116.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Println(msg)
		panic(e)
	}
}

// ------------------------------------------------------- min ---------------------------------
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
