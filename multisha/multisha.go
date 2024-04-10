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
	"src/misc"
	"sync"
	"sync/atomic"
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
  13 Dec 22 -- On the testing.sha, sequential routine (sha) took 12.3963 sec, and this rtn took 6.0804 sec, ratio is 2.04.  So the concurrent code is 2X faster than non-concurrent.
                 The first wait group, wg1 below, still had results print after wg1.Wait().  I'll leave it in as the result is interesting to me.
                 I had to add another wait group that gets decremented after a result is printed.  That one, called wg2 below, does what I need.
  15 Dec 22 -- I'm going to add a post counter that has to be atomically added and see how that affects the timings.
                 Doesn't seem to have increased the timings.  This rtn is still slightly faster (6.07 vs 6.1 sec) than conSha, ~ 0.5%.  Interesting.
  15 Feb 23 -- Seeing if changing the buffering of the channels makes a different.  And making the numOfWorkers = runtime.NumCPU(), as Bill Kennedy seems to love.
  25 Apr 23 -- Adding some enhancements I first developed w/ the copyc family
  30 Jun 23 -- Fixed the bug when there's no newline character to end the last (or only) line, that line's not processed because err != nil.
   4 Jul 23 -- The other day I created the misc package that has the code from makesubst, and I added the fixed readLine(*bytes.Reader) to it.  There should only be 1 version
                 of the code that I have to maintain.
   8 Aug 23 -- tknptr.NewToken -> tknptr.New
  10 Apr 24 -- Now that I learned that I/O bound work can benefit from many more goroutines than CPU bound work, and this here is I/O bound work, I'll increase the workers.
                 But I have to remember that linux only has 1000 or so file handles; this number cannot be exceeded.
*/

const LastCompiled = "10 Apr 2024"

const (
	undetermined = iota
	md5hash
	sha1hash
	sha256hash
	sha384hash
	sha512hash
)

const multiplier = 10

var numOfWorkers = runtime.NumCPU()

type hashType struct {
	fName     string
	hashValIn string
}

type resultMatchType struct {
	fname   string
	hashNum int
	match   bool
	err     error
}

var hashChan chan hashType
var resultChan chan resultMatchType
var wg1, wg2 sync.WaitGroup

var preCounter, postCounter int64 // atomic add requires int64, not int.  I might as well make both of them int64.  And staticcheck wrongly says preCounter is unused.

func matchOrNoMatch(hashIn hashType) { // returning filename, hash number, matched, error.  Input and output via a channel
	targetFile, err := os.Open(hashIn.fName)
	defer wg1.Done() // this has to be here to make sure that wg1.Done() is called, even if there's an error.
	if err != nil {
		result := resultMatchType{
			fname: hashIn.fName,
			err:   err,
		}
		resultChan <- result
		return
	}

	defer targetFile.Close() // I could do this w/ one defer func() as is done in cgrepi.  Turned out that StaticCheck wants this line after if err != nil because TargetFile would be nil then.

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
			err:   fmt.Errorf("indeterminate hash type"),
		}
		resultChan <- result
		return
	}

	_, err = io.Copy(hashFunc, targetFile)
	if err != nil {
		result := resultMatchType{
			fname: hashIn.fName,
			err:   err,
		}
		resultChan <- result
		return
	}

	computedHashValStr := hex.EncodeToString(hashFunc.Sum(nil))

	if strings.EqualFold(computedHashValStr, hashIn.hashValIn) { // golangci-lint found this optimization.
		result := resultMatchType{
			fname:   hashIn.fName,
			hashNum: hashInt,
			match:   true,
		}
		resultChan <- result
	} else {
		result := resultMatchType{
			fname:   hashIn.fName,
			hashNum: hashInt,
			match:   false,
		}
		resultChan <- result
	}
	// don't need a return statement, as results are sent thru a channel.
} // end matchOrNoMatch

var hashName = [...]string{"undetermined", "md5", "sha1", "sha256", "sha384", "sha512"}

// --------------------------------------- MAIN ----------------------------------------------------
func main() {
	var ans, Filename string
	var TargetFilename, HashValueReadFromFile string
	var h hashType
	//var counter int

	//if numOfWorkers < 1 {
	//	numOfWorkers = 1
	//}

	workingDir, _ := os.Getwd()
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" multisha.go, last altered %s, compiled with %s, and timestamp is %s\n", LastCompiled, runtime.Version(), LastLinkedTimeStamp)
	fmt.Printf("Working directory is %s.  Full name of executable is %s.\n", workingDir, execName)
	fmt.Println()

	// starting the worker go routines before the result goroutine.  This is not a fan out pattern, it's a worker pool pattern.
	//hashChan = make(chan hashType, numOfWorkers)  Turns out that this is not faster than making an unbuffered channel, and may even be slower.  Interesting.
	//resultChan = make(chan resultMatchType)  // This is slightly slower
	//resultChan = make(chan resultMatchType, numOfWorkers)
	//hashChan = make(chan hashType, 1)          // this may be slightly faster.
	hashChan = make(chan hashType)             // Bill Kennedy says in the worker pool pattern, making this synchronous is recommended.  And it seems to be the fastest of all I tested today.
	resultChan = make(chan resultMatchType, 1) // This is slightly faster than when the buffer is numOfWorkers
	for range numOfWorkers * multiplier {
		go func() {
			for h := range hashChan {
				matchOrNoMatch(h)
			}
		}()
	}

	// Start the results go routine.  There is only 1 of these.
	go func() {
		onWin := runtime.GOOS == "windows"
		for result := range resultChan {
			if result.err != nil {
				ctfmt.Printf(ct.Red, onWin, " Error from matchOrNoMatch is %s\n", result.err)
				atomic.AddInt64(&postCounter, 1)
				wg2.Done()
				continue // Using return was bad here; it's working using continue.
			}
			if result.match {
				ctfmt.Printf(ct.Green, onWin, " %s matched using %s hash\n", result.fname, hashName[result.hashNum])
			} else {
				ctfmt.Printf(ct.Red, onWin, " %s did not match using %s hash\n", result.fname, hashName[result.hashNum])
			}
			postCounter++ // I can't see why this needs to be atomic, as there's only 1 of these goroutines.
			wg2.Done()
		}
	}()

	// filepicker stuff.

	if len(os.Args) <= 1 {
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
	//bytesBuffer := bytes.NewBuffer(fileByteSlice)
	bytesReader := bytes.NewReader(fileByteSlice)

	onWin := runtime.GOOS == "windows" // for color output
	t0 := time.Now()

	for { // to read multiple lines
		//                                                      inputLine, er := bytesBuffer.ReadString('\n')
		inputLine, err := misc.ReadLine(bytesReader)
		//                                                      inputLine = strings.TrimSpace(inputLine) // probably not needed as I tokenize this, but I want to see if this works.  Yeah, it works.
		//                                                      fmt.Printf(" after ReadString and line is: %#v\n", inputLine)

		if err == io.EOF /* && inputLine == "" */ { // reached EOF condition, there are no more lines to read, and no line.  If there is a line, process it and error out the next time thru.
			break
		} else if len(inputLine) == 0 {
			continue
		} else if len(inputLine) < 10 || strings.HasPrefix(inputLine, ";") || strings.HasPrefix(inputLine, "#") {
			continue
		} else if err != nil /* && inputLine == "" */ {
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

		// Create Hash Section
		h = hashType{
			fName:     TargetFilename,
			hashValIn: HashValueReadFromFile,
		}
		wg1.Add(1)
		wg2.Add(1)
		preCounter++
		hashChan <- h // one sender (here), many receivers which are the workers in the pool.
	}

	// Sent all work into the matchOrNoMatch, so I'll close the hashChan
	close(hashChan)
	//                                                               ctfmt.Printf(ct.Green, true, " Just closed the hashChan.  There are %d goroutines, preCounter is %d and postCounter is %d.\n\n", runtime.NumGoroutine(), preCounter, postCounter) // counter = 24 is correct.

	wg1.Wait() // wg1.Done() is called in matchOrNoMatch.
	//                                                               fmt.Printf(" After wg1.Wait.  PostCounter = %d.\n", postCounter)
	wg2.Wait() // wg2.Done() is called in the goroutine that receives the results, after processing results so 2 branches in that goroutine call wg2.Done().
	//                                                               fmt.Printf(" After wg2.Wait.  PostCounter = %d.\n", postCounter)
	close(resultChan) // all work is done, so I can close the resultChan.

	ctfmt.Printf(ct.Yellow, onWin, "\n Elapsed time for everything was %s.\n\n\n", time.Since(t0))
} // Main for sha.go.

// ------------------------------------------------------- min ---------------------------------
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// ----------------------------------------------------- readLine ------------------------------------------------------
// Needed as a bytes reader does not have a readString method.

/*
func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byte, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if sb.Len() > 0 {
					return sb.String(), nil
				}
				// Error here is not EOF.
				return strings.TrimSpace(sb.String()), err
			}
		}
		if byte == '\n' {
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byte)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine

func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byte, err := r.ReadByte()
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
		if byte == '\n' {
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byte)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine
*/
