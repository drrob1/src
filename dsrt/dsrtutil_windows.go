package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
)

func GetUserGroupStr(fi os.FileInfo) (usernameStr, groupnameStr string) {
	return "", ""
}

// processCommandLine will return a slice of FileInfos after the filter and exclude expression are processed, and that match a pattern if given.
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
func getFileInfosFromCommandLine(sortfcn func(i, j int) bool) FISliceType {
	var GrandTotal int64
	var GrandTotalCount int
	var fileInfos FISliceType
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
		workingDir, er := os.Getwd()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			os.Exit(1)
		}
		fileInfos = MyReadDir(workingDir)
	} else { // Must have a pattern on the command line, ie, NArg > 0
		pattern := flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.

		if strings.ContainsRune(pattern, ':') {
			directoryAliasesMap = getDirectoryAliases()
			pattern = ProcessDirectoryAliases(directoryAliasesMap, pattern)
		} else if strings.Contains(commandLine, "~") { // this can only contain a ~ on Windows.
			pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		}
		dirName, fileName := filepath.Split(pattern)
		fileName = strings.ToLower(fileName)
		d, err := os.Open(dirName)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine os.Open is %v\n", err)
			os.Exit(1)
		}
		filenames, e := d.Readdirnames(0) // I don't know if I have to make this slice first.  I'm going to assume not for now.
		if e != nil {                     // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
			fmt.Fprintln(os.Stderr, e, "so calling my own MyReadDir.")
			fileInfos = MyReadDir(dirName)
		}

		fileInfos = make(FISliceType, 0, len(filenames))
		for _, f := range filenames {
			fi, err := os.Lstat(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			if includeThis(fi) && filepath.Match(pattern, strings.ToLower(f)) { // has to match pattern, size criteria and not match an exclude pattern.
				fileInfos = append(fileInfos, fi)
			}
			if fi.Mode().IsRegular() && ShowGrandTotal {
				GrandTotal += fi.Size()
				GrandTotalCount++
			}
		}
	}
	sort.Slice(fileInfos, sortfcn)
	return fileInfos

} // end getFileInfosFromCommandLine

//displayFileInfos only as to display.  The matching, filtering and excluding was already done by getFileInfosFromCommandLine
func displayFileInfos(fiSlice FISliceType) {
	var lnCount int
	for _, f := range fiSlice {
		s := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizestr := ""
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			SizeTotal += f.Size()
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
			count++
		} else if IsSymlink(f.Mode()) {
			fmt.Printf("%17s %s <%s>\n", sizestr, s, f.Name())
			count++
		} else if Dirlist && f.IsDir() {
			fmt.Printf("%17s %s (%s)\n", sizestr, s, f.Name())
			count++
		}
		if count >= NumLines {
			break
		}

	}

}
