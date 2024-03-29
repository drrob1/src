package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"src/filepicker"
	"src/getcommandline"
	"src/tknptr"
	"strconv"
	"strings"
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
  22 May 22 -- Edited so it will now compile if needed.  But it's not needed.
*/

const LastCompiled = "22 May 2022"

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
	var inbuf, ans, Filename string
	var WhichHash int
	var readErr error
	var TargetFilename, HashValueReadFromFile, HashValueComputedStr string
	var hasher hash.Hash
	var FileSize int64

	fmt.Print(" comparehashes written in Go.  GOOS =", runtime.GOOS, ".  ARCH=", runtime.GOARCH)

	fmt.Println(".  Last altered", LastCompiled)
	//            fmt.Println(".  HashType = md5, sha1, sha256, sha384, sha512.  WhichHash = ",HashName[WhichHash]);
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf("%s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n", ExecFI.Name(), LastLinkedTimeStamp, workingdir, execname)
	fmt.Println()

	// filepicker stuff.

	if len(os.Args) <= 1 { // need to use filepicker
		filenames, err := filepicker.GetFilenames("*.sha*")
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepicker is %v.\n", err)
		}
		for i := 0; i < min(len(filenames), 10); i++ {
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
		Filename = getcommandline.GetCommandLineString()
	}

	fmt.Println()

	// process extension

	extension := filepath.Ext(Filename)
	extension = strings.ToLower(extension)
	switch extension {
	case ".md5":
		WhichHash = md5hash
	case ".sha1":
		WhichHash = sha1hash
	case ".sha256":
		WhichHash = sha256hash
	case ".sha384":
		WhichHash = sha384hash
	case ".sha512":
		WhichHash = sha512hash
	default:
		fmt.Println()
		fmt.Println()
		fmt.Println(" Not a recognized hash extension.  Will determine by hash length.")
	} // switch case on extension for HashType

	fmt.Println()

	// Read and parse the file with the hashes.

	HashesFile, err := os.Open(Filename)
	if os.IsNotExist(err) {
		fmt.Println(inbuf, " does not exist.")
		os.Exit(1)
	} else { // we know that the file exists
		check(err, " Error opening hashes file.")
	}
	defer HashesFile.Close()

	scanner := bufio.NewScanner(HashesFile)
	//	scanner.Split(bufio.ScanLines)  This is the default.  I may experiment to see if I need this line for my code to work, AFTER I debug it as it is.

	for { /* to read multiple lines */
		FileSize = 0
		readSuccess := scanner.Scan()
		if !readSuccess {
			break
		} // end main reading loop

		inputline := scanner.Text()
		if readErr = scanner.Err(); readErr != nil {
			if readErr == io.EOF {
				break
			} // reached EOF condition, so there are no more lines to read.
			fmt.Fprintln(os.Stderr, "Unknown error while reading from the HashesFile :", readErr)
			os.Exit(1)
		}

		if strings.HasPrefix(inputline, ";") || strings.HasPrefix(inputline, "#") || (len(inputline) <= 10) {
			continue
		} /* allow comments and essentially blank lines */

		//	inputline = strings.Replace(inputline, "*", " ", -1) // just blank out the * char
		//tokenize.INITKN(inputline)
		tokenPtr := tknptr.NewToken(inputline) // I need to declare a pointer receiver, not like the static tokenize rtn.  I'll use a Go idiom.

		//tokenize.SetMapDelim('*') // this should now work, as of 01/26/2018
		tokenPtr.SetMapDelim('*')

		FirstToken, EOL := tokenPtr.GetTokenString(false)

		if EOL {
			fmt.Println(" EOL or other Error while getting 1st token in the hashing file.  Skipping to next line.")
			continue
		}
		hashlength := 0

		if strings.ContainsRune(FirstToken.Str, '.') || strings.ContainsRune(FirstToken.Str, '-') ||
			strings.ContainsRune(FirstToken.Str, '_') { // have filename first on line
			TargetFilename = FirstToken.Str
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get hash string from the line in the file
			if EOL {
				fmt.Println(" Got EOL while getting HashValue (2nd) token in the hashing file.  Skipping")
				continue
			} /* if EOL */
			HashValueReadFromFile = SecondToken.Str
			hashlength = len(SecondToken.Str)

		} else { /* have hash first on line */
			HashValueReadFromFile = FirstToken.Str
			hashlength = len(FirstToken.Str)
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get name of file on which to compute the hash
			if EOL {
				fmt.Println(" Error while gatting TargetFilename token in the hashing file.  Skipping")
				continue
			} /* if EOL */

			if strings.ContainsRune(SecondToken.Str, '*') { // If it contains a *, it will be the first position.
				SecondToken.Str = SecondToken.Str[1:]
				if strings.ContainsRune(SecondToken.Str, '*') { // this should not happen
					fmt.Println(" Filename token still contains a * character.  Str:", SecondToken.Str, "  Skipping.")
					continue
				}
			}
			TargetFilename = (SecondToken.Str)
		} /* if have filename first or hash value first */

		/*
		   now to compute the hash, compare them, and output results
		*/
		/* Create Hash Section */
		TargetFile, err := os.Open(TargetFilename)
		//    exists := true;
		if os.IsNotExist(err) {
			fmt.Println(TargetFilename, " does not exist.")
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
				fmt.Fprintln(os.Stderr, " Could not determine hash type for file.  Exiting.")
				os.Exit(1)
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
		} /* initializing switch case on WhichHash */

		/*    This loop works, but there is a much shorter way.  I got this after asking for help on the mailing list.
		      FileReadBuffer := make([]byte,ReadBufferSize);
		      for {   // Repeat Until eof loop.
		        n,err := TargetFile.Read(FileReadBuffer);
		        if n == 0 || err == io.EOF { break }
		        check(err," Unexpected error while reading the target file on which to compute the hash,");
		        hasher.Write(FileReadBuffer[:n]);
		        FileSize += int64(n);
		      } // Repeat Until TargetFile.eof loop;
		*/

		FileSize, readErr = io.Copy(hasher, TargetFile)
		HashValueComputedStr = hex.EncodeToString(hasher.Sum(nil))

		// I got the idea to use the different base64 versions and my own hex converter code, just to see.
		// And I can also use sfprintf with the %x verb.  base64 versions are not useful as they use a larger
		// character set than hex.  I deleted all references to the base64 versions.  And the hex encoded and
		// sprintf using %x were the same, so I removed the sprintf code.
		//    HashValueComputedSprintf := fmt.Sprintf("%x",hasher.Sum(nil));

		//		fmt.Println(" Filename  = ", TargetFilename, ", FileSize = ", FileSize, ", ", HashName[WhichHash], " computed hash string -- ")
		fmt.Printf(" Filename  = %s, filesize = %d, using hash %s.\n", TargetFilename, FileSize, HashName[WhichHash])
		fmt.Println("       Read From File:", HashValueReadFromFile)
		fmt.Println(" Computed hex encoded:", HashValueComputedStr)
		//    fmt.Println(" Computed sprintf:",HashValueComputedSprintf);

		if strings.ToLower(HashValueReadFromFile) == strings.ToLower(HashValueComputedStr) {
			fmt.Print(" Matched.")
		} else {
			fmt.Print(" Not matched.")
		} /* if hashes */
		TargetFile.Close() // Close the handle to allow opening a target from the next line, if there is one.
		fmt.Println()
		fmt.Println()
		fmt.Println()
	} /* outer LOOP to read multiple lines */

	HashesFile.Close() // Don't really need this because of the defer statement.
	fmt.Println()
} // Main for comparehashes.go.

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
