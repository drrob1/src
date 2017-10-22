package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"filepicker"
	"fmt"
	"getcommandline"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"tokenize"
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
*/

const LastCompiled = "22 Oct 2017"

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
	//  const ReadBufferSize = 80 * M;

	var HashName = [...]string{"md5", "sha1", "sha256", "sha384", "sha512"}
	var inbuf, ans, Filename string
	var WhichHash int
	var readErr error
	var TargetFilename, HashValueReadFromFile, HashValueComputedStr string
	var hasher hash.Hash
	var FileSize int64

	fmt.Print(" comparehashes written in Go.  GOOS =", runtime.GOOS, ".  ARCH=", runtime.GOARCH)

	fmt.Println(".  Last compiled ", LastCompiled)
	//            fmt.Println(".  HashType = md5, sha1, sha256, sha384, sha512.  WhichHash = ",HashName[WhichHash]);
	workingdir, _ := os.Getwd()
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf("%s has timestamp of %s.  Working directory is %s.  Full name of executable is %s.\n", ExecFI.Name(), LastLinkedTimeStamp, workingdir, execname)
	fmt.Println()

	// filepicker stuff.

	if len(os.Args) <= 1 { // need to use filepicker
		filenames := filepicker.GetFilenames("*.sha*")
		for i := 0; i < min(len(filenames), 10); i++ {
			fmt.Println("filename[", i, "] is", filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		fmt.Scanln(&ans)
		if len(ans) == 0 {
			ans = "0"
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
		fmt.Println(" Not a recognized hash extension.  Will assume sha1.")
		WhichHash = sha1hash
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
	//	scanner.Split(bufio.ScanLines) // I believe this is the default.  I may experiment to see if I need this line for my code to work, AFTER I debug it as it is.

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

		tokenize.INITKN(inputline)
		tokenize.SetMapDelim('*') // to ignore this character that begins the filename field.  Don't know why it's there.

		FirstToken, EOL := tokenize.GetTokenString(false)

		if EOL {
			fmt.Errorf(" Error while getting 1st token in the hashing file.  Skipping")
			continue
		}

		if strings.ContainsRune(FirstToken.Str, '.') { /* have filename first on line */
			TargetFilename = FirstToken.Str
			SecondToken, EOL := tokenize.GetTokenString(false) // Get hash string from the line in the file
			if EOL {
				fmt.Errorf(" Got EOL while getting HashValue (2nd) token in the hashing file.  Skipping \n")
				continue
			} /* if EOL */
			HashValueReadFromFile = SecondToken.Str

		} else { /* have hash first on line */
			HashValueReadFromFile = FirstToken.Str
			SecondToken, EOL := tokenize.GetTokenString(false) // Get name of file on which to compute the hash
			if EOL {
				fmt.Errorf(" Error while gatting TargetFilename token in the hashing file.  Skipping \n")
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

		if HashValueReadFromFile == HashValueComputedStr {
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
		fmt.Errorf("%s : ", msg)
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
