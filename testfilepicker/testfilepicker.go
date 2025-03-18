package main

/*
  REVISION HISTORY
  ================
  19 Aug 16 -- First Go version completed to test all parts of tokenize.go package
  21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
  13 Oct 17 -- Testing the inclusion of horizontal tab as a delim, needed for comparehashes.
  18 Oct 17 -- Now called testfilepicker, derived from testtoken.go
  10 Jan 22 -- Converted to modules, and will test both GetFilenames and GetRegexFilenames.
  17 Mar 25 -- After I added more routines, I'll benchmark them now.
				On leox, the results for a local directory are ~15-20 ms, and the FullFilename routine is always the fastest.
				On thelio, the results for a local directory are ~2 ms, and the FullFilename routine is still always the fastest.
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"src/filepicker"
	"strings"
	"time"
)

const maxChoices = 20

func testGetFilenames(pattern string) ([]string, error) {
	t0 := time.Now()
	filenames, err := filepicker.GetFilenames(pattern)
	fmt.Printf("test getFilenames: duration %v, finding %d files\n", time.Since(t0), len(filenames))
	return filenames, err
}

func testGetRegexFilenames(regex string) ([]string, error) {
	t0 := time.Now()
	filenames, err := filepicker.GetRegexFilenames(regex)
	fmt.Printf("test getRegexFilenames: duration %v, finding %d files\n", time.Since(t0), len(filenames))
	return filenames, err
}

func testGetRegexFullFilenames(regex string) ([]string, error) {
	t0 := time.Now()
	filenames, err := filepicker.GetRegexFullFilenames(regex)
	fmt.Printf("test getRegexFullFiles: duration %v, finding %d files\n", time.Since(t0), len(filenames))
	return filenames, err
}

func main() {

	globPattern := "*.mp4"
	regexPattern := "mp4$"
	globFiles, err := testGetFilenames(globPattern) // in directory w/ ~2300 files, this routine was ~19.5 ms.  In dir w/ ~13,800 files, all 3 of these rtn's were ~8.3 s.
	if err != nil {
		fmt.Println("testGetFilenames: ", err)
	}

	t0 := time.Now()
	glob2Files, err := getFilenames(globPattern)
	if err != nil {
		fmt.Printf("in main: getFilenames err %v\n ", err)
	}
	fmt.Printf(" in main, getFilenames took %v to match %d entries\n", time.Since(t0), len(glob2Files)) // same directory, slightly slower than library GetFilenames.

	regexFiles, err := testGetRegexFilenames(regexPattern) // same directory, this routine was ~23.7 ms
	if err != nil {
		fmt.Println("testGetRegexFilenames: ", err)
	}
	regexFullFiles, err := testGetRegexFullFilenames(regexPattern) // sam directory, this routine was ~14.8 ms
	if err != nil {
		fmt.Println("testGetRegexFullFilenames: ", err)
	}
	fmt.Printf(" Len of globFiles = %d, len(regexFiles)=%d, len(regexFullFiles)=%d	\n",
		len(globFiles), len(regexFiles), len(regexFullFiles))

	fmt.Printf(" As a test of the main copy of getFilenames\n")
	fmt.Printf("%v\n", glob2Files[:10])
	fmt.Printf(" library copy of GetFilenames\n")
	fmt.Printf("%v\n\n", globFiles[:10])
	equal := slices.Compare(globFiles, glob2Files)
	if equal == 0 {
		fmt.Println("slices are equal")
	} else {
		fmt.Println("slices are not equal, which sucks!")
	}

	fmt.Printf("\n\n")
	/*
		var ans, commandline, regex string // So I can test on linux.  Bash globs.
		fmt.Print(" commandline = ")
		_, err := fmt.Scanln(&commandline)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from Scanln is %v, will assume '*' \n", err)
			commandline = "*"
		}

		filenames, err := filepicker.GetFilenames(commandline)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepicker is %v \n", err)
		}
		fmt.Println(" Number of filenames in the string slice are", len(filenames))

		maxchoices := min(maxChoices, len(filenames))
		for i := 0; i < maxchoices; i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		_, err = fmt.Scanln(&ans)
		if err != nil {
			ans = "0"
		}
		fmt.Printf(" ans is string %s, hex %x \n", ans, ans)
		i, err := strconv.Atoi(ans)
		if err == nil {
			fmt.Println(" ans as int is", i)
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A') // may need byte(s0) - 'A' or byte(s0-'A') or some other permutation
			fmt.Println(" string ans as int is", i, " referenced filename is", filenames[i])
		}

		if i < len(filenames) {
			fmt.Println(" Picked filename is", filenames[i])
			fmt.Println()
		}

		a := 'a'
		b := a ^ 32  // exclusive or
		c := a &^ 32 // and not, where not means 1's complement
		d := a | 32  // or
		fmt.Printf(" Bit fiddling:  a,b,c,d = %d %c, %d %c, %d %c, %d  %c \n", a, a, b, b, c, c, d, d)

		fmt.Println()

	*/

	/*
		floatflag := false
		if (commandline == "REAL") || (commandline == "FLOAT") {
			floatflag = true
			testingstate = 1
		} else if commandline == "STRING" {
			testingstate = 2
		} else if commandline == "EOL" {
			testingstate = 3
		} else if commandline == "STR" {
			testingstate = 4
		} else if commandline == "LOWER" {
			testingstate = 5
		}

		fmt.Print(" Will be testing GetTknReal? ", floatflag, ", testingstate is ", testingstate)
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print(" Input test text: ")
			scanner.Scan()
			inputline := scanner.Text()
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				os.Exit(1)
			}
			if len(inputline) == 0 {
				os.Exit(0)
			}
			fmt.Println(" After the call to Scan(), before TrimSpace: ", inputline, ".")
			inputline = strings.TrimSpace(inputline)
			fmt.Println(" After call to TrimSpace: ", inputline)
			if strings.ToUpper(inputline) == "QUIT" {
				log.Println(" Test Token finished.")
				os.Exit(0)
			}
			tkn := tknptr.NewToken(inputline)
			EOL := false
			token := tknptr.TokenType{}
			for !EOL {
				if floatflag || testingstate == 1 {
					token, EOL = tkn.GETTKNREAL()
				} else if testingstate == 2 {
					token, EOL = tkn.GETTKNSTR()
				} else if testingstate == 3 {
					token, EOL = tkn.GETTKNEOL()
				} else if testingstate == 4 {
					token, EOL = tkn.GetTokenString(false)
				} else if testingstate == 5 {
					token, EOL = tkn.GetToken(false)
				} else {
					token, EOL = tkn.GETTKN()
				}

				fmt.Printf(" Token : %#v \n", token)
				//      if floatflag {     I think this is just an error.  I missed something in testing
				//        fmt.Println(" R = ",token.Rsum);
				//      }
				fmt.Println(" EOL : ", EOL)
				if EOL {
					break
				} // I don't want it to ask about ungettkn if there is an EOL condition.
				fmt.Print(" call UnGetTkn? (Y/N) ")
				//			scanner.Scan()
				//			ans := scanner.Text()
				//			if err := scanner.Err(); err != nil {
				//				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				//				os.Exit(1)
				//			}
				fmt.Scan(&ans)
				ans = strings.TrimSpace(ans)
				ans = strings.ToUpper(ans)
				if strings.HasPrefix(ans, "Y") {
					tkn.UNGETTKN()
				}
			}
			fmt.Println()
			log.Println(" Finished processing the inputline.")
		}

	*/

	/*
		fmt.Print(" regex is: ")
		_, err = fmt.Scanln(&regex)
		if regex == "" {
			fmt.Fprintf(os.Stderr, " Error from Scanln is %v, will assume '.' \n", err)
			regex = "."
		}

		filenames, err = filepicker.GetRegexFilenames(regex)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepicker is %v \n", err)
		}
		fmt.Println(" Number of filenames in the string slice are", len(filenames))

		maxchoices = min(maxChoices, len(filenames))
		for i := 0; i < maxchoices; i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		_, err = fmt.Scanln(&ans)
		fmt.Printf(" ans is string %s, hex %x \n", ans, ans)
		if err != nil {
			ans = "0"
		}
		i, err = strconv.Atoi(ans)
		if err == nil {
			fmt.Println(" ans as int is", i)
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A') // may need byte(s0) - 'A' or byte(s0-'A') or some other permutation
			fmt.Println(" string ans as int is", i, " referenced filename is", filenames[i])
		}

		if i < len(filenames) {
			fmt.Println(" Picked filename is", filenames[i])
			fmt.Println()
		}

	*/
}

/*  from the web documentation at golang.org
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
          fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
        if err := scanner.Err(); err != nil {
          fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
*/

type FISliceDate []os.FileInfo // used by sort.Sort in GetFilenames.

func (f FISliceDate) Less(i, j int) bool {
	return f[i].ModTime().UnixNano() > f[j].ModTime().UnixNano() // I want a reverse sort, newest first
}

func (f FISliceDate) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FISliceDate) Len() int {
	return len(f)
}

// GetFilenames -- pattern uses filepath.Match to see if there's a match.  IE, a glob type match.  And I'm writing it as I would now, using DirEntry
func getFilenames(pattern string) ([]string, error) { // This routine sorts using sort.Sort
	CleanDirName, CleanFileName := filepath.Split(pattern)
	if len(CleanDirName) == 0 {
		CleanDirName = "." + string(filepath.Separator)
	}

	if len(CleanFileName) == 0 {
		CleanFileName = "*"
	}

	dirname, err := os.Open(CleanDirName)
	if err != nil {
		return nil, err
	}
	defer dirname.Close()

	dirEntries, err := dirname.ReadDir(0)
	if err != nil {
		return nil, err
	}

	CleanFileName = strings.ToLower(CleanFileName)
	fileInfos := make(FISliceDate, 0, len(dirEntries))
	for _, entry := range dirEntries {
		lowerName := strings.ToLower(entry.Name())
		if matched, _ := filepath.Match(CleanFileName, lowerName); matched {
			if entry.IsDir() { // It was showing directories until I did this on 10/12/24.
				continue
			}
			L, err := entry.Info()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from os.Lstat is %v \n", err)
				continue
			}
			fileInfos = append(fileInfos, L)
		}
	}

	sort.Sort(fileInfos)

	stringSlice := make([]string, 0, 50) // hard coded this magic number.

	var count int
	for _, f := range fileInfos {
		stringSlice = append(stringSlice, f.Name()) // needs to preserve case of filename for linux
		count++
		if count >= 50 { // I'm hard coding this magic number, as that's what it is in filepicker.
			break
		}
	}
	return stringSlice, nil
} // end getFilenames
