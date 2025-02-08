// detox.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

/*
  REVISION HISTORY
  ----------------
  14 Sep 20 -- First version, based on code from fromfx.go.
  16 Sep 20 -- Fixed when it changed '-' to '_'.  And I added IsPunct so it would change comma and other punctuations.
  18 Sep 20 -- I had it ignore '~' and not change it, as it was included as a punctuation mark by IsPunct.
  22 Sep 20 -- Fixed case issue by converting pattern to all lower case also.  I forgot that before.  And I will allow no pattern to be entered.
  28 Apr 23 -- I want to only have one '.' char in the filename.  So I'll replace all but the last one w/ a '-'.
  30 Apr 23 -- I added !IsGraphic to the tests, which may be redundant, but I'll try it and see.  And I'm combining the dot substitutions into detoxFilenameNewWay
   2 May 23 -- Will limit removing dots to not do so when extensions are .gpg, .gz, .xz
   4 May 23 -- Decided to add a flag to enable removing the dots.  This will be off by default.  I would always want to remove difficult characters, but not always remove dots.
                 go test isn't working by me passing -dots to it.  I'm going to try what happens when I use noDotsFlagPtr.  Nope, no difference.
   6 May 23 -- I posted on golang-nuts@googlegroups.com for help on how to use the go test system.  I finally got it working.
  11 May 23 -- Fixing bug in use of flag.NArg().
  23 Jan 24 -- Added noWorkFlag, which means do not actually do anything.  Just print what would be done.  And changed the name of the other option to dot.  It was too hard to type detox -dots
  28 Jan 24 -- noWorkFlag now really does work.
   2 Mar 24 -- Added back -dots as an alternative, so both -dot and -dots do the same thing.
   7 Feb 25 -- Changing the flag values.  I don't need viper to do this.  I'll use -n to be noWorkFlag, and -d to mean noDotsFlag.
*/

const lastModified = "7 Feb 25"

var noDotsFlag, noDotFlag bool
var noWorkFlag bool

func init() {
	flag.BoolVar(&noDotsFlag, "dots", false, "Enable removing excess dots from filenames.")
	flag.BoolVar(&noDotFlag, "dot", false, "Enable removing excess dots from filenames.") // I've done this before, and it works.
	flag.BoolVar(&noDotFlag, "d", false, "Enable removing excess dots from filenames.")   // IE, having 2 options set the same Flag bool variable.
	flag.BoolVar(&noWorkFlag, "n", false, "No work is to be done.  Just show what would be done.")
}

func main() {
	var globPattern string

	flag.Parse()
	if noDotFlag { // so if -d is used, it will also set -noDotsFlag.  If noDotsFlag is set by the -dots option, it doesn't matter.
		noDotsFlag = true
	}

	fmt.Println()

	if flag.NArg() == 0 {
		globPattern = "*"
	} else {
		globPattern = strings.ToLower(flag.Arg(0)) // first argument on command line.
	}

	startDirectory, _ := os.Getwd() // startDirectory is a string
	fmt.Println()
	fmt.Printf(" detox.go lastModified is %s, will use globbing pattern of %q and will start in %s.  noDotsFlag=%t, noDotFlag=%t. \n",
		lastModified, globPattern, startDirectory, noDotsFlag, noDotFlag)
	fmt.Println()

	files := myReadDirNames(startDirectory)
	ctr := 0
	for _, fn := range files {
		name := strings.ToLower(fn)
		if BOOL, _ := filepath.Match(globPattern, name); BOOL {
			detoxedName, toxic := detoxFilenameNewWay(fn)
			if toxic {
				ctr++
				if noWorkFlag {
					fmt.Printf(" filename %q would be detoxed\n", fn)
				} else {
					err := os.Rename(fn, detoxedName)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					fmt.Printf(" filename %q -> %q \n", fn, detoxedName)
				}
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

// detoxFilenameNewWay ------------------------------------------------------------------------------------------------
func detoxFilenameNewWay(fName string) (string, bool) {
	const dotReplacementRune = '-'

	var changed bool
	var sb strings.Builder
	var counter, targetNumOfDots int

	if noDotsFlag {
		targetNumOfDots = strings.Count(fName, ".") - 1 // because I want to keep the last dot.
		ext := filepath.Ext(fName)
		if ext == ".gpg" || ext == ".gz" || ext == ".xz" { // keep the last 2 dots.
			targetNumOfDots--
		}
	}

	for _, r := range fName {
		size := utf8.RuneLen(r)
		if size > 1 {
			changed = true
			sb.WriteRune('_')
		} else if r == '.' && counter < targetNumOfDots {
			sb.WriteRune(dotReplacementRune)
			counter++
			changed = true
		} else if unicode.IsSpace(r) {
			changed = true
			sb.WriteRune('_')
		} else if unicode.IsControl(r) {
			changed = true
			sb.WriteRune('_')
		} else if r == '.' || r == '_' || r == '-' || r == '~' {
			sb.WriteRune(r)
		} else if unicode.IsSymbol(r) || unicode.IsPunct(r) || !unicode.IsGraphic(r) {
			changed = true
			sb.WriteRune('_')
		} else {
			sb.WriteRune(r)
		}
	}
	f := sb.String()
	//f, changed := tooManyDots(f)
	return f, changed
} // end detoxFilenameNewWay

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

/*
// tooManyDots(fName string) string -------------------------------------------------------------------
func tooManyDots(fName string) (string, bool) { // this has 2 different ways to achieve the same goal.  I know I'm playing around.  Now that it works, I'm including the
	//                                             dot logic in the detoxFilenameNewWay routine.  The string split stuff works, but I don't need it.
	const replacementRune = '-'
	const replacementStr = string(replacementRune)

	targetNumOfDots := strings.Count(fName, ".") - 1 // because I want to keep the last dot.
	//fmt.Printf(" in tooManyDots: fName is %q, targetNumOfDots is %d\n", fName, targetNumOfDots)
	if targetNumOfDots < 1 {
		return fName, false
	}

	var sb strings.Builder
	var counter int
	for _, r := range fName {
		if r == '.' && counter < targetNumOfDots {
			r = replacementRune
			counter++
		}
		sb.WriteRune(r)
	}

	s1 := sb.String()

	splitStr := strings.Split(fName, ".") // this does not include dot
	//fmt.Printf(" in tooManyDots: fName is %q, s1 is %q, targetNumOfDots = %d, splitStr = %#v\n", fName, s1, targetNumOfDots, splitStr)
	j1 := strings.Join(splitStr[:len(splitStr)-1], replacementStr) // this line uses a subrange, so after the ':' so len-1 really means len-2
	j2 := j1 + "." + splitStr[len(splitStr)-1]                     // is using the expr as a subscript index, so len-1 means len-1.
	if s1 != j2 {
		fmt.Printf(" s1 = %q, j2 = %q and these are not the same.  This needs more work\n", s1, j2)
	}

	return s1, true
}

*/
// ---------------------------------------------------------------------------------------------------
/*
func detoxFilenameOldWay(fname string) (string, bool) {
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
} // end detoxFilenameOldWay

*/
