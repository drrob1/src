// rmvdots.go from detox.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
   1 May 23 -- Now called rmvdots.go.  And will only, well, remove excess dots.  But for .gpg files, it keeps the last 2 dots.
   2 May 23 -- And for ".gz" and ".xz" it will also keep the last 2 dots.
  20 Aug 23 -- Stopped using ToLower on filename string.
*/

const lastModified = "20 Aug 23"

type dirAliasMapType map[string]string

func main() {
	flag.Parse()

	fmt.Println()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" rmvdots.go lastModified is %s, last linked %s, and compiled w/ %s. \n", lastModified, LastLinkedTimeStamp, runtime.Version())
	fmt.Println()

	files := getFileNamesFromCommandLine()
	var ctr int
	for _, fn := range files {
		//ext := filepath.Ext(fn)
		//fmt.Printf(" Filename of %s has ext of %s\n", fn, ext)
		//name := strings.ToLower(fn)
		newName, changed := tooManyDots(fn)
		if changed {
			err := os.Rename(fn, newName)
			if err != nil {
				fmt.Fprintf(os.Stderr, " ERROR from os.Rename(%s,%s) is %s\n", fn, newName, err)
			}
			ctr++
			fmt.Printf(" filename %q -> %q \n", fn, newName)
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

// tooManyDots(fName string) string -------------------------------------------------------------------

func tooManyDots(fName string) (string, bool) { // this no longer has 2 different ways to achieve the same goal, because it could not accommodate special ext's like .gpg
	const replacementRune = '-'

	targetNumOfDots := strings.Count(fName, ".") - 1 // because I want to keep the last dot.
	//fmt.Printf(" in tooManyDots: fName is %q, targetNumOfDots is %d\n", fName, targetNumOfDots)
	ext := filepath.Ext(fName)
	if ext == ".gpg" || ext == ".gz" || ext == ".xz" {
		targetNumOfDots-- // so don't change the last 2 dots, ie, .docx.gpg remains that.
	}
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

	return sb.String(), true
}

// ------------------------------ GetDirectoryAliases ----------------------------------------
func getDirectoryAliases() dirAliasMapType { // Env variable is diraliases.
	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = MakeSubst(s, '_', ' ') // substitute the underscore, _, for a space
	directoryAliasesMap := make(dirAliasMapType, 10)

	dirAliasSlice := strings.Fields(s)

	for _, aliasPair := range dirAliasSlice {
		if string(aliasPair[len(aliasPair)-1]) != "\\" {
			aliasPair = aliasPair + "\\"
		}
		aliasPair = MakeSubst(aliasPair, '-', ' ') // substitute a dash,-, for a space
		splitAlias := strings.Fields(aliasPair)
		directoryAliasesMap[splitAlias[0]] = splitAlias[1]
	}
	return directoryAliasesMap
}

// ------------------------------ ProcessDirectoryAliases ---------------------------

func ProcessDirectoryAliases(cmdline string) string {
	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap := getDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := aliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathnFile := cmdline[idx+1:]
	completeValue := aliasValue + PathnFile
	fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
	return completeValue
} // ProcessDirectoryAliases

// --------------------------- MakeSubst -------------------------------------------

func MakeSubst(instr string, r1, r2 rune) string {
	inRune := make([]rune, len(instr))
	if !strings.ContainsRune(instr, r1) {
		return instr
	}

	for i, s := range instr {
		if s == r1 {
			s = r2
		}
		inRune[i] = s // was byte(s) before I made this a slice of runes.
	}
	return string(inRune)
} // makesubst

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

/*
func detoxFilenameNewWay(fName string) (string, bool) {
	const dotReplacementRune = '-'

	var changed bool
	var sb strings.Builder
	var counter int

	targetNumOfDots := strings.Count(fName, ".") - 1 // because I want to keep the last dot.

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

*/

/*

func tooManyDots(fName string) (string, bool) { // this has 2 different ways to achieve the same goal.  I know I'm playing around.  Now that it works, I'm including the
	//                                             dot logic in the detoxFilenameNewWay routine.  The string split stuff works, but I don't need it.
	const replacementRune = '-'
	const replacementStr = string(replacementRune)

	targetNumOfDots := strings.Count(fName, ".") - 1 // because I want to keep the last dot.
	//fmt.Printf(" in tooManyDots: fName is %q, targetNumOfDots is %d\n", fName, targetNumOfDots)
	if targetNumOfDots < 1 {
		return fName, false
	}

	ext := filepath.Ext(fName)
	if ext == ".gpg" || ext == ".gz" || ext == ".xz" {
		targetNumOfDots-- // so don't change the last 2 dots, ie, .docx.gpg remains that.
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
