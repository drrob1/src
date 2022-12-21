package list

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// processCommandLine will return a slice of FileInfos after the filter and exclude expression are processed, and that match a pattern if given.
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
func getFileInfoXFromCommandLine(excludeMe *regexp.Regexp) []FileInfoExType {
	var fileInfoX []FileInfoExType

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + string(filepath.Separator)

	if flag.NArg() == 0 {
		workingDir, er := os.Getwd()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			os.Exit(1)
		}
		fileInfoX = MyReadDir(workingDir, excludeMe)
	} else { // Must have a pattern on the command line, ie, NArg > 0
		pattern := flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.

		if strings.ContainsRune(pattern, ':') {
			directoryAliasesMap = getDirectoryAliases()
			pattern = ProcessDirectoryAliases(directoryAliasesMap, pattern)
		} else if strings.Contains(pattern, "~") { // this can only contain a ~ on Windows.
			pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		}
		dirName, fileName := filepath.Split(pattern)
		fileName = strings.ToLower(fileName)
		if dirName != "" && fileName == "" { // then have a dir pattern without a filename pattern
			fileInfoX = MyReadDir(dirName, excludeMe)
			return fileInfoX
		}
		if dirName == "" {
			dirName = "."
		}
		if fileName == "" { // need this to not be blank because of the call to Match below.
			fileName = "*"
		}
		if verboseFlag {
			fmt.Printf(" dirName=%s, fileName=%s \n", dirName, fileName)
		}

		if verboseFlag {
			fmt.Printf(" dirName=%s, fileName=%s \n", dirName, fileName)
		}

		var filenames []string
		if globFlag {
			// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
			// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
			// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
			filenames, err = filepath.Glob(pattern)
			if verboseFlag {
				fmt.Printf(" after glob: len(filenames)=%d, filenames=%v \n\n", len(filenames), filenames)
			}

		} else {
			d, err := os.Open(dirName)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine os.Open is %v\n", err)
				os.Exit(1)
			}
			defer d.Close()
			filenames, err = d.Readdirnames(0) // I don't know if I have to make this slice first.  I'm going to assume not for now.
			if err != nil {                    // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
				fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
				fileInfoX = MyReadDir(dirName, excludeMe)
			}

		}

		fileInfoX = make([]FileInfoExType, 0, len(filenames))
		const sepStr = string(os.PathSeparator)
		for _, f := range filenames { // basically I do this here because of a pattern to be matched.
			var path string
			if strings.Contains(f, sepStr) {
				path = f
			} else {
				path = filepath.Join(dirName, f)
			}

			fi, err := os.Lstat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from Lstat call on %s is %v\n", path, err)
				continue
			}
			//if !fi.Mode().IsRegular() { // skip directories and symlink.  IE, skip anything that is not a regular file.  Too bad it doesn't work.  It does work in IncludeThis, though
			//	continue
			//}

			match, er := filepath.Match(strings.ToLower(fileName), strings.ToLower(f)) // redundant if glob is used, but I'm ignoring this.
			if er != nil {
				fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern is %v.\n", pattern, er)
				continue
			}

			if includeThis(fi, excludeMe) && match { // has to match pattern, size criteria and not match an exclude pattern.
				fix := FileInfoExType{
					fi:  fi,
					dir: dirName,
				}
				fileInfoX = append(fileInfoX, fix)
			}
			if fi.Mode().IsRegular() && showGrandTotal {
				grandTotal += fi.Size()
				grandTotalCount++
			}
		} // for f ranges over filenames
	} // if flag.NArgs()

	return fileInfoX

} // end getFileInfoXFromCommandLine

/*
func getColorizedStrings(fiSlice []os.FileInfo, cols int) []colorizedStr {

	cs := make([]colorizedStr, 0, len(fiSlice))

	for i, f := range fiSlice {
		t := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizeStr := ""
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			sizeTotal += f.Size()
			if longFileSizeListFlag {
				sizeStr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
				if f.Size() > 100000 {
					sizeStr = AddCommas(sizeStr)
				}
				strng := fmt.Sprintf("%16s %s %s", sizeStr, t, f.Name())
				colorized := colorizedStr{color: ct.Yellow, str: strng}
				cs = append(cs, colorized)

			} else {
				var colr ct.Color
				sizeStr, colr = getMagnitudeString(f.Size())
				strng := fmt.Sprintf("%-10s %s %s", sizeStr, t, f.Name())
				colorized := colorizedStr{color: colr, str: strng}
				cs = append(cs, colorized)
			}

		} else if IsSymlink(f.Mode()) {
			s := fmt.Sprintf("%5s %s <%s>", sizeStr, t, f.Name())
			colorized := colorizedStr{color: ct.White, str: s}
			cs = append(cs, colorized)
		} else if dirList && f.IsDir() {
			s := fmt.Sprintf("%5s %s (%s)", sizeStr, t, f.Name())
			colorized := colorizedStr{color: ct.White, str: s}
			cs = append(cs, colorized)
		}
		if i > numOfLines*cols {
			break
		}
	}
	if verboseFlag {
		fmt.Printf(" In getColorizedString.  len(fiSlice)=%d, len(cs)=%d, numofLines=%d\n", len(fiSlice), len(cs), numOfLines)
	}
	return cs
}
*/
