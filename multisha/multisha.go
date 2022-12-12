package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"hash"
	"io"
	"os"
	"time"

	"runtime"
	"src/filepicker"
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
                 Errors now go to Stderr.  Uses bytes buffer to read sha file using io.ReadAll. and go 1.15.8
   7 Mar 21 -- added strings.TrimSpace
   8 Apr 21 -- Converted import list to module named src.  So essentially, very little has changed except for these import statements.
  13 Feb 22 -- filepicker API changed recently.  So I'm updating the code here that uses filepicker.
   9 Mar 22 -- Using package constants instead of my magic numbers.
  13 Jun 22 -- Cleaning up some comments, from Boston's SIR 2022.  And removed unused code.  And finally removed depracated ioutil.
  21 Oct 22 -- Now using strings.EqualFold as recommended by golangci-lint.
  11 Dec 22 -- Now called multisha to play w/ concurrent sha matching.  I decided the easiest name for this is multisha.  This will not be fast to code.
                 The result will be matched or not matched for each.
                 I would need to pass in the hash and filename for each, and let the matchOrNoMatch function determine which hash is in play.
                 And the result channel could be a bool for match or not matched, the filename and hash function used.
                 I guess I need to for range on the input channel which takes a hashType.
                 I'll first debug without multitasking code
*/

const LastCompiled = "12 Dec 2022"

const (
	undetermined = iota
	md5hash
	sha1hash
	sha256hash
	sha384hash
	sha512hash
)

type hashType struct {
	fName     string
	hashValIn string
}

type resultMatchType struct {
	fname   string
	hashNum int
	match   bool
}

func matchOrNoMatch(hashIn hashType) (resultMatchType, error) { // returning filename, hash number, matched, error

	TargetFile, err := os.Open(hashIn.fName)
	defer TargetFile.Close()
	if err != nil {
		result := resultMatchType{
			fname: hashIn.fName,
		}
		return result, err
	}

	hashLength := len(hashIn.hashValIn)
	var hashFunc hash.Hash
	var hashInt int

	if hashLength == 2*sha256.Size { // 64, and the Size constant is number of bytes, not number of digits.
		hashInt = sha256hash
		hashFunc = sha256.New()
	} else if hashLength == 2*sha512.Size { // 128
		hashInt = sha512hash
		hashFunc = sha512.New()
	} else if hashLength == 2*sha1.Size { // 40
		hashInt = sha1hash
		hashFunc = sha1.New()
	} else if hashLength == 2*sha512.Size384 { // 96
		hashInt = sha384hash
		hashFunc = sha512.New384()
	} else if hashLength == 2*md5.Size { // 32
		hashInt = md5hash
		hashFunc = md5.New()
	} else {
		result := resultMatchType{
			fname: hashIn.fName,
		}
		return result, fmt.Errorf("indeterminate hash type")
	}

	_, er := io.Copy(hashFunc, TargetFile)
	if er != nil {
		result := resultMatchType{
			fname: hashIn.fName,
		}
		return result, er
	}

	computedHashValStr := hex.EncodeToString(hashFunc.Sum(nil))

	if strings.EqualFold(computedHashValStr, hashIn.hashValIn) { // golangci-lint found this optimization.
		result := resultMatchType{
			fname:   hashIn.fName,
			hashNum: hashInt,
			match:   true,
		}
		return result, nil
	} else {
		result := resultMatchType{
			fname:   hashIn.fName,
			hashNum: hashInt,
			match:   false,
		}
		return result, nil
	}
}

var hashName = [...]string{"undetermined", "md5", "sha1", "sha256", "sha384", "sha512"}

// --------------------------------------- MAIN ----------------------------------------------------
func main() {

	var ans, Filename string
	//var WhichHash int
	//var TargetFilename, HashValueReadFromFile, HashValueComputedStr string
	var TargetFilename, HashValueReadFromFile string
	//var hasher hash.Hash
	//var FileSize int64
	var h hashType
	var resultMatch resultMatchType

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" multisha.go, last altered %s, compiled with %s, and timestamp is %s\n", LastCompiled, runtime.Version(), LastLinkedTimeStamp)
	fmt.Printf("Working directory is %s.  Full name of executable is %s.\n", workingDir, execName)
	fmt.Println()

	// filepicker stuff.

	if len(os.Args) <= 1 {
		filenames, err := filepicker.GetFilenames("*.sha*")
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepicker is %v.  Exiting \n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
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
		i, er := strconv.Atoi(ans)
		if er == nil {
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
		Filename = os.Args[1]
	}

	fmt.Println()

	// Read and parse the file listing the hashes.

	fileByteSlice, err := os.ReadFile(Filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bytesBuffer := bytes.NewBuffer(fileByteSlice)

	onWin := runtime.GOOS == "windows"
	t0 := time.Now()
	for { // to read multiple lines
		//FileSize = 0
		//WhichHash = undetermined // reset it for this next line, allowing multiple types of hashes in same file.

		inputLine, er := bytesBuffer.ReadString('\n')
		inputLine = strings.TrimSpace(inputLine) // probably not needed as I tokenize this, but I want to see if this works.
		if er == io.EOF && len(inputLine) == 0 { // reached EOF condition, there are no more lines to read, and no line
			break
		} else if len(inputLine) == 0 && err != nil {
			fmt.Fprintln(os.Stderr, "While reading from the HashesFile:", err)
			os.Exit(1)
		}

		if strings.HasPrefix(inputLine, ";") || strings.HasPrefix(inputLine, "#") || (len(inputLine) <= 10) {
			continue
		} // allow comments and essentially blank lines

		tokenPtr := tknptr.NewToken(inputLine)
		tokenPtr.SetMapDelim('*')
		FirstToken, EOL := tokenPtr.GetTokenString(false)
		if EOL {
			fmt.Fprintln(os.Stderr, " EOL while getting 1st token in the hashing file.  Skipping to next line.")
			continue
		}
		//hashlength := 0

		if strings.ContainsRune(FirstToken.Str, '.') || strings.ContainsRune(FirstToken.Str, '-') ||
			strings.ContainsRune(FirstToken.Str, '_') { // have filename first on line
			TargetFilename = FirstToken.Str
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get hash string from the line in the file
			if EOL {
				fmt.Fprintln(os.Stderr, " Got EOL while getting HashValue (2nd) token in the hashing file.  Skipping")
				continue
			}
			HashValueReadFromFile = SecondToken.Str
			//hashlength = len(SecondToken.Str)

		} else { // have hash first on line
			HashValueReadFromFile = FirstToken.Str
			//hashlength = len(FirstToken.Str)
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
			TargetFilename = SecondToken.Str
		} // endif have filename first or hash value first

		// Create Hash Section
		h = hashType{
			fName:     TargetFilename,
			hashValIn: HashValueReadFromFile,
		}

		resultMatch, err = matchOrNoMatch(h)
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, TargetFilename, " does not exist.  Skipping.")
			continue
		} else { // we know that the file exists
			msg := fmt.Sprintf("error opening %s", TargetFilename)
			check(err, msg)
		}

		if resultMatch.match {
			ctfmt.Printf(ct.Green, onWin, " %s matched using %s hash\n", resultMatch.fname, hashName[resultMatch.hashNum])
		} else {
			ctfmt.Printf(ct.Red, onWin, " %s did not match using %s hash\n", resultMatch.fname, hashName[resultMatch.hashNum])
		}
		ctfmt.Printf(ct.Yellow, onWin, " Elapsed time was %s.\n", time.Since(t0))
		fmt.Println()
		fmt.Println()
	} // outer LOOP to read multiple lines
} // Main for sha.go.

// ------------------------------------------------------- check -------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Fprintln(os.Stderr, msg)
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
