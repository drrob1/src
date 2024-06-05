package main // csha.go from consha.go

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"hash"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"runtime"
	"src/filepicker"
	"src/misc"
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
------------------------------------------------------------------------------------------------------------------------------------------------------
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
  13 Dec 22 -- On the testing.sha, sequential routine (sha) took 12.3963 sec, and this rtn took 6.0804 sec, ratio is 2.04.  So the concurrent code is 2X faster than non-concurrent.
                 The first wait group, wg1 below, still had results print after wg1.Wait().  I'll leave it in as the result is interesting to me.
                 I had to add another wait group that gets decremented after a result is printed.  That one, called wg2 below, does what I need.
------------------------------------------------------------------------------------------------------------------------------------------------------
  14 Dec 22 -- Now called conSha.go, and I want to simplify the code.  I don't need a receiving go routine; I'll have matchOrNoMatch print the results, too.
                 Timing results on same testing.sha show that this is slightly slower than multiSha.  IE, having separate go routines to collect the results and then show them is slightly
                 faster than doing both in the same routine.  Go figure.  Wait, scratch that.  This routine also has a post counter that is incremented atomically.  This is not in
                 multiSha.  That could also account for the differences in speed.  So I'll say it's a tie.  The difference on win10 desktop is 6.1 sec here vs 6.07 sec from multiSha.
                 And this difference persists even after I added the atomic adds to multiSha.
  15 Feb 23 -- Following Bill Kennedy's advice and making the channel either synchronous or maybe buffer of 1.  This is still slower than multisha, by ~0.04 sec here on Win10 desktop.
   7 Apr 23 -- StaticCheck reports that resultMatchType is unused.  So I commented it out.  And another error in matchOrNoMatch, which I fixed.
------------------------------------------------------------------------------------------------------------------------------------------------------
  26 Apr 23 -- I'm going to use some of the enhancements I developed for the copyc family of routines here.
                 Now called csha, and I'll change to putting the hashes in a slice so I know how many there are, and then only start those go routines.
                 On win10 desktop, this sha file took 10.1 sec, but on leox it took 2 min 2.9 sec.  And now csha took 7.17 sec on win10 desktop, and 1 min 51.7 sec on leox.
                 Win10_22H2_English_x64.iso	F41BA37AA02DCB552DC61CEF5C644E55B5D35A8EBDFAC346E70F80321343B506
                 Win10_22H2_English_x32.iso	7CB5E0E18B0066396A48235D6E56D48475225E027C17C0702EA53E05B7409807
  28 Apr 23 -- At work, I tested again, and this routine was slower than either multisha or consha.  Consha was slightly faster than multisha at work.  I'm increasing the size of the channel.
                 This has only 1 channel, for passing hashing work to matchOrNoMatch.  There is no message channel here.
  30 Jun 23 -- Fixed the bug when there's no newline character to end the last (or only) line, that line's not processed because err != nil.
                 And now moved the improved ReadLine code to the new package misc, which started life as makesubst.
   8 Aug 23 -- Since I changed tknptr and removed NewToken, I had make the change here, so NewToken becomes New.  This is more idiomatic for Go, anyway.
   8 Sep 23 -- There's a bug here in the processing of the errors, ie, when one file is found and another is not found.  I found the error, it was using a var that was not assigned.
                 I tried removing the hashSlice, but the code ran slightly slower on Win11 desktop.  ~35 ms without the slice, vs ~33 ms w/ the slice.  Go figure.
                 I used git to restore the prev code.
  10 Apr 24 -- I/O bound work, as in here, benefits from more workers than NumCPU()
                 But I have to remember that linux only has 1024 file handles; this number cannot be exceeded.
   3 May 24 -- Learned that wait groups are not intended to increment and decrement for each individual file; they cover the goroutines themselves.
   4 Jun 24 -- Removed dead code that was commented out long ago.  This routine first determines how many hashes need to be checked, and spins up that number of goroutines, up to a max.
                 A results channel is not used as matchOrNoMatch prints the result as soon as it's determined.
------------------------------------------------------------------------------------------------------------------------------------------------------
   5 Jun 24 -- Now called fansha, as I want to implement a fanout pattern instead of a worker pool pattern.  As Miki Tebeka said, know your data.  I usually use this routine
                for a few files, almost never more than 5.  So a fanout pattern should work fine.
*/

const LastCompiled = "5 June 2024"

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

var postCounter int64
var wg1 sync.WaitGroup

var hashName = [...]string{"undetermined", "md5", "sha1", "sha256", "sha384", "sha512"}
var onWin bool

func matchOrNoMatch(hashIn hashType) { // returning filename, hash number, matched, error.  Input and output via a channel
	TargetFile, err := os.Open(hashIn.fName)
	defer wg1.Done()
	defer atomic.AddInt64(&postCounter, 1)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ctfmt.Printf(ct.Red, onWin, " ERROR from matchOrNoMatch: %s not found, skipping. \n", hashIn.fName)
		} else {
			ctfmt.Printf(ct.Red, onWin, " Error from matchOrNoMatch is %s\n", err)
		}
		return
	}
	defer TargetFile.Close()

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
		ctfmt.Printf(ct.Red, onWin, " unknown hash type for %s\n", hashIn.fName)
		return
	}

	onWin := runtime.GOOS == "windows"
	_, er := io.Copy(hashFunc, TargetFile)
	if er != nil {
		ctfmt.Printf(ct.Red, onWin, " Error from io.Copy in matchOrNoMatch is %s\n", er)
		return
	}

	computedHashValStr := hex.EncodeToString(hashFunc.Sum(nil))

	if strings.EqualFold(computedHashValStr, hashIn.hashValIn) { // golangci-lint found this optimization.
		ctfmt.Printf(ct.Green, onWin, "                                       %s matched using %s hash\n", hashIn.fName, hashName[hashInt])
	} else {
		ctfmt.Printf(ct.Red, onWin, "                                       %s did not match using %s hash\n", hashIn.fName, hashName[hashInt])
	}
} // end matchOrNoMatch

// --------------------------------------- MAIN ----------------------------------------------------
func main() {
	var ans, Filename string
	var TargetFilename, HashValueReadFromFile string
	var h hashType
	var preCounter int
	var verboseFlag bool

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" %s is last altered %s, compiled with %s, and timestamp is %s\n", os.Args[0], LastCompiled, runtime.Version(), LastLinkedTimeStamp)
	fmt.Printf("Working directory is %s.  Full name of executable is %s.\n", workingDir, execName)
	fmt.Println()
	flag.BoolVar(&verboseFlag, "v", false, "verbose mode, mostly for debugging.")
	flag.Parse()

	onWin = runtime.GOOS == "windows"

	// filepicker stuff.

	if flag.NArg() == 0 {
		filenames, err := filepicker.GetFilenames("*.sha*")
		if err != nil {
			ctfmt.Printf(ct.Red, false, " Error from filepicker is %v.  Exiting \n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		n, er := fmt.Scanln(&ans)
		if n == 0 || er != nil {
			ans = "0"
		} else if ans == "999" {
			fmt.Println(" Stop code entered.  Exiting.")
			os.Exit(0)
		}
		i, e := strconv.Atoi(ans)
		if e == nil {
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
		ctfmt.Println(ct.Red, false, os.Stderr, err)
		os.Exit(1)
	}
	bytesReader := bytes.NewReader(fileByteSlice)

	t0 := time.Now()

	for { // to read multiple lines
		inputLine, err := misc.ReadLine(bytesReader)
		if errors.Is(err, io.EOF) { // reached EOF condition; there are no more lines to read, and no line here to be processed, as that condition is established by misc.ReadLine().
			break
		} else if len(inputLine) == 0 {
			continue
		} else if len(inputLine) < 10 || strings.HasPrefix(inputLine, ";") || strings.HasPrefix(inputLine, "#") {
			continue
		} else if err != nil {
			ctfmt.Println(ct.Red, false, "While reading from the HashesFile:", err)
			continue
		}

		tokenPtr := tknptr.New(inputLine)
		tokenPtr.SetMapDelim('*')
		FirstToken, EOL := tokenPtr.GetTokenString(false)
		if EOL {
			ctfmt.Println(ct.Red, false, " EOL while getting 1st token in the hashing file.  Skipping to next line.")
			continue
		}

		if strings.ContainsRune(FirstToken.Str, '.') || strings.ContainsRune(FirstToken.Str, '-') ||
			strings.ContainsRune(FirstToken.Str, '_') { // have filename first on line
			TargetFilename = FirstToken.Str
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get hash string from the line in the file
			if EOL {
				ctfmt.Println(ct.Red, false, " Got EOL while getting HashValue (2nd) token in the hashing file.  Skipping")
				continue
			}
			HashValueReadFromFile = SecondToken.Str

		} else { // have hash first on line
			HashValueReadFromFile = FirstToken.Str
			SecondToken, EOL := tokenPtr.GetTokenString(false) // Get name of file on which to compute the hash
			if EOL {
				ctfmt.Println(ct.Red, false, " EOL while gatting TargetFilename token in the hashing file.  Skipping")
				continue
			}

			if strings.HasPrefix(SecondToken.Str, "*") { // If it contains a *, it will be the first position.
				SecondToken.Str = SecondToken.Str[1:]
				if strings.ContainsRune(SecondToken.Str, '*') { // this should not happen
					ctfmt.Println(ct.Red, false, " Filename token still contains a * character.  Str:", SecondToken.Str, "  Skipping.")
					continue
				}
			}
			TargetFilename = SecondToken.Str
		} // endif have filename first or hash value first

		if verboseFlag {
			fmt.Printf(" TargetFilename = %s, hash value read from file = %s\n", TargetFilename, HashValueReadFromFile)
		}

		// Create Hash Section and send to matchOrNoMatch
		h = hashType{
			fName:     TargetFilename,
			hashValIn: HashValueReadFromFile,
		}
		wg1.Add(1)
		preCounter++
		go matchOrNoMatch(h) // this is a fanout pattern
	}

	runningGoRoutines := runtime.NumGoroutine()

	if verboseFlag {
		ctfmt.Printf(ct.Cyan, true, " Before wg1.Wait().  There are %d goroutines, pre counter is %d and post counter is %d.\n\n",
			runningGoRoutines, preCounter, postCounter) // counter = 25 is correct.
	}

	wg1.Wait() // wg1.Done() used to be called in matchOrNoMatch, now it's deferred in the worker goroutines.

	if verboseFlag {
		ctfmt.Printf(ct.Cyan, true, "\n After wg1.  There are %d goroutines, pre counter is %d and post counter is %d.\n\n",
			runtime.NumGoroutine(), preCounter, postCounter) // counter = 25 is correct.
	}

	ctfmt.Printf(ct.Yellow, onWin, "\n\n %s: Elapsed time for everything was %s using %d goroutines.\n\n",
		os.Args[0], time.Since(t0), runningGoRoutines)
} // Main for sha.go.
