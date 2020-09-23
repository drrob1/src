// detox.go
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const lastModified = "22 Sep 20"

/*
  REVISION HISTORY
  ----------------
  14 Sep 20 -- First version, based on code from fromfx.go.
  16 Sep 20 -- Fixed when it changed '-' to '_'.  And I added IsPunct so it would change comma and other punctuations.
  18 Sep 20 -- I had it ignore '~' and not change it, as it was included as a punctuation mark by IsPunct.
  22 Sep 20 -- Fixed case issue by converting pattern to all lower case also.  I forgot that before.  And I will allow no pattern to be entered.
*/

func main() {
	//var e error
	var globPattern string

	fmt.Println()

	if len(os.Args) <= 1 { // this means no arguments on line, as the program name is always first argument passed in os.Args
		globPattern = "*"
	} else {
		globPattern = strings.ToLower(os.Args[1])
	}

	startDirectory, _ := os.Getwd() // startDirectory is a string
	fmt.Println()
	fmt.Printf(" detox.go lastModified is %s, will use globbing pattern of %q and will start in %s. \n", lastModified, globPattern, startDirectory)
	fmt.Println()

	files := myReadDirNames(startDirectory)
	ctr := 0
	for _, fn := range files {
		name := strings.ToLower(fn)
		if BOOL, _ := filepath.Match(globPattern, name); BOOL {
			detoxedName, toxic := detoxFilename(fn)
			if toxic {
				err := os.Rename(fn, detoxedName)
				if err != nil {
					//fmt.Fprintf(os.Stderr, " Error from rename function for name %s -> %s: %v \n", fn, detoxedName, err)
					fmt.Fprintln(os.Stderr, err)
				}
				ctr++
				fmt.Printf(" filename %q -> %q \n", fn, detoxedName)
			}
		}
	}
	if ctr > 1 {
		fmt.Printf("\n Total of %d files were renamed. \n", ctr)
	} else if ctr == 1 {
		fmt.Printf("\n One file was renamed. \n")
	} else {
		fmt.Println(" No files were renamed.")
	}
} // end main

//---------------------------------------------------------------------------------------------------
func detoxFilename(fname string) (string, bool) {
	var toxic bool

	buf := bytes.NewBufferString(fname)

	byteslice := make([]byte, 0, 100)

	for {
		r, size, err := buf.ReadRune()
		if err == io.EOF { // only valid exit from this loop
			name := string(byteslice)
			return name, toxic
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return "", false // returning toxic as false to not do anything with this name as it got an error of some type.
		}
		if size > 1 {
			toxic = true
			byteslice = append(byteslice, '_')
		} else if unicode.IsSpace(r) {
			toxic = true
			byteslice = append(byteslice, '_')
		} else if unicode.IsControl(r) {
			toxic = true
			byteslice = append(byteslice, '_')
		} else if r == '.' || r == '_' || r == '-' || r == '~' {
			byteslice = append(byteslice, byte(r))
		} else if unicode.IsSymbol(r) || unicode.IsPunct(r) {
			toxic = true
			byteslice = append(byteslice, '_')
		} else {
			byteslice = append(byteslice, byte(r))
		}
	}
} // end detoxFilename

// ------------------------------- myReadDirNames -----------------------------------
func myReadDirNames(dir string) []string { // based on the code from dsrt and descendents

	dirname, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return nil
	}
	dirname.Close()
	return names
} // myReadDirNames

// END detox.go
