package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetUserGroupStr(fi os.FileInfo) (usernameStr, groupnameStr string) {
	return "", ""
}

// processCommandLine will return a slice of FileInfos after the filter and exclude expression are processed, and that match a pattern if given.
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
func getFileInfosFromCommandLine() []os.FileInfo {
	var fileInfos []os.FileInfo
	//var workingDir string
	//var er error

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + string(filepath.Separator)

	if flag.NArg() == 0 {
		//workingDir, er := os.Getwd()
		//if er != nil {
		//	//fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
		//	os.Exit(1)
		//}
		//fileInfos = MyReadDir(workingDir)
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
		if dirName == "" {
			dirName = "."
		}
		if testFlag {
			fmt.Printf(" dirName=%s, fileName=%s \n", dirName, fileName)
		}

		var filenames []string
		if globFlag {
			// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
			// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
			// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
			filenames, err = filepath.Glob(pattern)
			if testFlag {
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
				//fileInfos = MyReadDir(dirName)
			}

		}

		fileInfos = make([]os.FileInfo, 0, len(filenames))
		for _, f := range filenames {
			fi, err := os.Lstat(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			match, er := filepath.Match(strings.ToLower(pattern), strings.ToLower(f)) // redundant if glob is used, but I'm ignoring this.
			if er != nil {
				fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern is %v.\n", pattern, er)
				continue
			}

			if includeThis(fi) && match { // has to match pattern, size criteria and not match an exclude pattern.
				fileInfos = append(fileInfos, fi)
			}
			if fi.Mode().IsRegular() && showGrandTotal {
				grandTotal += fi.Size()
				grandTotalCount++
			}
		} // for f ranges over filenames
	} // if flag.NArgs()

	return fileInfos

} // end getFileInfosFromCommandLine

//displayFileInfos only as to display.  The matching, filtering and excluding was already done by getFileInfosFromCommandLine
func displayFileInfos(fiSlice []os.FileInfo) {
	var lnCount int
	for _, f := range fiSlice {
		s := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizestr := ""
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			sizeTotal += f.Size()
			if longFileSizeListFlag {
				sizestr = strconv.FormatInt(f.Size(), 10)
				if f.Size() > 100_000 {
					sizestr = AddCommas(sizestr)
				}
				fmt.Printf("%17s %s %s\n", sizestr, s, f.Name())

			} else {
				var color ct.Color
				sizestr, color = getMagnitudeString(f.Size())
				ctfmt.Printf(color, true, "%-17s %s %s\n", sizestr, s, f.Name())
			}
			lnCount++
		} else if IsSymlink(f.Mode()) {
			fmt.Printf("%17s %s <%s>\n", sizestr, s, f.Name())
			lnCount++
		} else if dirList && f.IsDir() {
			fmt.Printf("%17s %s (%s)\n", sizestr, s, f.Name())
			lnCount++
		}
		if lnCount >= numOfLines {
			break
		}

	}

}
