package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/spf13/pflag"
)

/*
REVISION HISTORY
-------- -------
 8 Feb 22 -- Found a bug in glob flag when I tested with "z:*.TXT" that didn't happen w/ the separate glob.go, fast.go or dsrt z:*.txt
               I got an error from Lstat in that it tried to Lstat z:\z:filename.TXT.
15 Feb 22 -- Replaced testFlag w/ verboseFlag.  Finally.
24 Feb 22 -- Fixed a bug in the glob option.  And Evan's 30 today.  Wow.
12 Apr 23 -- Fixed a bug in GetIDName, which is now called idName to be more idiomatic for Go.  But that is not called here in Windows code, so nevermind.
21 Jun 25 -- DisplayFileInfos now knows when output is redirected.
22 Jun 25 -- myPrintf now used.
21 Aug 25 -- Now gets more info for symlinks by using a separate call to lstat.  I don't yet know if this is needed on linux too.
				Lstat makes no attempt to follow the symlink.  I think Stat does follow the symlink.
				To really be able to do that, I need to return the dirname from the getFileInfosFromCommandLine call.  And then pass that into the displayFileInfos call.
17 Sep 25 -- In the case of a symlink, will now display what the symlink points to.
 3 Oct 25 -- Adding non-concurrent routine to getFileInfosFromCommandLine.
*/

// getFileInfosFromCommandLine() will return a slice of FileInfos after the filter and exclude expression are processed, and that match a pattern if given.
// It handles if there are no files populated by bash or file not found by bash, but doesn't sort the slice before returning it, because of difficulty passing
// the sortfcn.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
// And now it also returns the directory name, which is needed for the symlink case.  It turns out that I can use filepath.Abs() to get the full path.  Nope, that doesn't work after all.
func getFileInfosFromCommandLine() ([]os.FileInfo, string) {
	var fileInfos []os.FileInfo
	var dirName, fileName string

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + string(filepath.Separator)

	if pflag.NArg() == 0 {
		workingDir, er := os.Getwd()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			os.Exit(1)
		}
		fileInfos = myReadDir(workingDir)
	} else { // Must have a pattern on the command line, ie, NArg > 0
		pattern := pflag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.

		if strings.ContainsRune(pattern, ':') {
			pattern = ProcessDirectoryAliases(pattern)
		}
		pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		dirName, fileName = filepath.Split(pattern)
		fileName = strings.ToLower(fileName)
		if dirName != "" && fileName == "" { // then have a dir pattern without a filename pattern
			fileInfos = myReadDir(dirName)
			return fileInfos, dirName
		}
		if dirName == "" {
			dirName = "."
		}
		if fileName == "" { // need this to not be blank because of the call to Match below.
			fileName = "*"
		}
		if verboseFlag {
			fmt.Printf(" In getFileInfosFromComandLine: dirName=%s, fileName=%s \n", dirName, fileName)
		}

		var filenames []string
		if globFlag {
			// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
			// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
			// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
			filenames, err = filepath.Glob(pattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, " In getFileInfosFromCommandLine: error from Glob is %v.\n", err)
				return nil, ""
			}
			dirName = "" // make this an empty string because the name returned by glob includes the dir info.
			if verboseFlag {
				fmt.Printf(" after glob: len(filenames)=%d, filenames=%v \n\n", len(filenames), filenames)
			}

		} else {
			d, err := os.Open(dirName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error from Windows processCommandLine directory os.Open is %v\n", err)
				os.Exit(1)
			}
			defer d.Close()
			filenames, err = d.Readdirnames(0) // I don't have to make filenames slice first.
			if err != nil {
				fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
				fileInfos = myReadDir(dirName)
				return fileInfos, dirName
			}
		} // if globFlag

		if veryVerboseFlag {
			fmt.Printf(" dirName=%s, len(filenames)=%d, filenames=%v \n\n", dirName, len(filenames), filenames)
		}
		fileInfos = make([]os.FileInfo, 0, len(filenames))
		const sepStr = string(os.PathSeparator)
		for _, f := range filenames { // basically I do this here because of a pattern to be matched.
			var path string
			if strings.Contains(f, sepStr) || strings.Contains(f, ":") || globFlag {
				path = f
			} else {
				path = dirName + sepStr + f
			}

			fi, err := os.Lstat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from Lstat call on %s is %v\n", path, err)
				continue
			}

			match, er := filepath.Match(strings.ToLower(fileName), strings.ToLower(f)) // redundant if glob is used, but I'm ignoring this.
			if er != nil {
				fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern and %s dirName is %v.\n", pattern, dirName, er)
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
	} // if pflag.NArgs()

	return fileInfos, dirName

} // end getFileInfosFromCommandLine

func nonConcurrentFileInfosFromCommandLine() ([]os.FileInfo, string) {
	var fileInfos []os.FileInfo
	var dirName, fileName string

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = ""
	}
	HomeDirStr = HomeDirStr + string(filepath.Separator)

	workingDir, er := os.Getwd()
	if er != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
		os.Exit(1)
	}
	pattern := pflag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.
	if pattern == "" {
		dirName = workingDir
		fileName = "*"
	} else {
		if strings.ContainsRune(pattern, ':') {
			pattern = ProcessDirectoryAliases(pattern)
		}
		pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		dirName, fileName = filepath.Split(pattern)
		fileName = strings.ToLower(fileName)

		if dirName == "" {
			dirName = "."
		}
		if fileName == "" { // need this to not be blank because of the call to Match below.
			fileName = "*"
		}
	}
	if verboseFlag {
		fmt.Printf(" In nonConcurrentFileInfosFromComandLine: dirName=%s, fileName=%s \n", dirName, fileName)
	}

	var filenames []string
	if globFlag {
		// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
		// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
		// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
		filenames, err = filepath.Glob(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, " In getFileInfosFromCommandLine: error from Glob is %v.\n", err)
			return nil, ""
		}
		dirName = "" // make this an empty string because the name returned by glob includes the dir info.
		if verboseFlag {
			fmt.Printf(" after glob: len(filenames)=%d, filenames=%v \n\n", len(filenames), filenames)
		}

	} else {
		d, err := os.Open(dirName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error from Windows processCommandLine directory os.Open is %v\n", err)
			os.Exit(1)
		}
		defer d.Close()
		filenames, err = d.Readdirnames(0) // I don't have to make filenames slice first.
		if err != nil {
			fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
			fileInfos = myReadDir(dirName)
			return fileInfos, dirName
		}
	} // if globFlag

	if veryVerboseFlag {
		fmt.Printf(" dirName=%s, len(filenames)=%d, filenames=%v \n\n", dirName, len(filenames), filenames)
	}
	fileInfos = make([]os.FileInfo, 0, len(filenames))
	const sepStr = string(os.PathSeparator)
	for _, f := range filenames { // basically I do this here because of a pattern to be matched.
		var path string
		if strings.Contains(f, sepStr) || strings.Contains(f, ":") || globFlag {
			path = f
		} else {
			path = dirName + sepStr + f
		}

		fi, err := os.Stat(path) // Lstat is not used here as it doesn't follow symlinks.  I want to follow symlinks.  So use Stat.
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from Lstat call on %s is %v\n", path, err)
			continue
		}

		match, er := filepath.Match(strings.ToLower(fileName), strings.ToLower(f)) // redundant if glob is used, but I'm ignoring this.
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern and %s dirName is %v.\n", pattern, dirName, er)
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

	return fileInfos, dirName

} // end nonConcurrentFileInfosFromCommandLine

// displayFileInfos only has to display.  The matching, filtering and excluding was already done by getFileInfosFromCommandLine.
// This is platform specific because of lack of uid:gid on Windows.
// And now it needs the directory name, which is needed for the symlink case.
func displayFileInfos(fiSlice []os.FileInfo, dirName string) {
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
				myPrintf(color, true, "%-17s %s %s\n", sizestr, s, f.Name())
			}
			lnCount++
		} else if IsSymlink(f.Mode()) {
			fullFileName := filepath.Join(dirName, f.Name()) // the full path is needed for the symlink case.
			finfo, err := os.Stat(fullFileName)              //  Without this, the date is that of the symlink, and the size is zero.  It still sorts by the date of the symlink.
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from os.Stat on %s is %v\n", f.Name(), err)
				continue
			}
			link, err := os.Readlink(fullFileName)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from os.Readlink on %s is %v\n", f.Name(), err)
				continue
			}
			var color ct.Color
			sizestr, color = getMagnitudeString(finfo.Size())
			s = finfo.ModTime().Format("Jan-02-2006_15:04:05")
			myPrintf(color, true, "%-17s %s <%s> [%s]\n", sizestr, s, f.Name(), link)
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
