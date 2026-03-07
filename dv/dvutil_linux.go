package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/spf13/pflag"
)

/*
 12 Apr 23 -- Fixed a bug in GetIDName, which is now called idName to be more idiomatic for Go.
 21 Jun 25 -- DisplayFileInfos now knows when output is redirected.
 22 Jun 25 -- myPrintf now used.
 21 Aug 25 -- Now gets more info for symlinks by using a separate call to lstat.  I don't yet know if this is needed on linux too.
                Lstat makes no attempt to follow the symlink.  I think Stat does follow the symlink.
                To really be able to do that, I need to return the dirname from the getFileInfosFromCommandLine call.  And then pass that into the displayFileInfos call.
                This doesn't work the same as it does on Windows.  The symlink shows up as an error if called from another directory.  It does work from the same directory.
                I don't yet know what to do about this.
 17 Sep 25 -- In the case of a symlink, will now display what the symlink points to.  Doesn't seem to be working here on linux.  I don't yet know why.  Maybe it's because
                using DirEntry does follow symlinks on linux but not windows?
  7 Mar 26 -- Copied routines from dvutil_windows.go, so this would compile on linux also.  It seems that I broke that compatibility when I was chasing down problems I saw at work.

*/

func GetUserGroupStr(fi os.FileInfo) (usernameStr, groupnameStr string) {
	if runtime.GOARCH != "amd64" { // 06/20/2019 11:23:40 AM made condition not equal, and will remove conditional from dsrt.go
		return "", ""
	}
	sysUID := int(fi.Sys().(*syscall.Stat_t).Uid) // Stat_t is a uint32
	uidStr := strconv.Itoa(sysUID)
	sysGID := int(fi.Sys().(*syscall.Stat_t).Gid) // Stat_t is a uint32
	gidStr := strconv.Itoa(sysGID)
	usernameStr = idName(uidStr)
	groupnameStr = idName(gidStr)
	return usernameStr, groupnameStr
} // end GetUserGroupStr

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, but the sorting will be done in main, as passing the sort fcn was a problem.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.

// on Jan 14, 2023 I completely rewrote the section of getFileInfosFromCommandLine where there is only 1 identifier on the command line.  This was based on what I learned
// from args.go.  Let's see if it works.  Basically, I relied too much on os.Lstat or os.Stat.  Now I'm relying on os.Open.

func getFileInfosFromCommandLine() ([]os.FileInfo, string) {
	var fileInfos []os.FileInfo
	var narg int
	var args []string
	var arg0 string

	if flag.Parsed() {
		args = flag.Args()
		narg = flag.NArg()
		arg0 = flag.Arg(0)
	} else if pflag.Parsed() {
		args = pflag.Args()
		narg = pflag.NArg()
		arg0 = pflag.Arg(0)
	} else {
		myPrintf(ct.Red, false, " Neither flag.Parsed() nor pflag.Parsed() is true.  WTF?\n")
		return nil, ""
	}
	if verboseFlag {
		fmt.Printf(" Entering getFileInfosFromCommandLine.  Nargs=%d, len(Args)=%d, len(fileinfos)=%d\n", narg, len(args), len(fileInfos))
	}

	workingDir, er := os.Getwd()
	if er != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
		os.Exit(1)
	}

	if narg == 0 {
		if verboseFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfos = myReadDir(workingDir) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if verboseFlag {
			fmt.Printf(" after call to myreaddir.  Len(fileInfos)=%d\n", len(fileInfos))
		}
		return fileInfos, workingDir

	} else if narg == 1 { // a lone name may either mean file not found or it's a directory which could be a symlink.
		const sep = string(filepath.Separator)
		fileInfos = make([]os.FileInfo, 0, 1)

		loneFilename := arg0
		fHandle, err := os.Open(loneFilename) // just try to open it, as it may be a symlink.
		if err == nil {
			stat, _ := fHandle.Stat()
			if stat.IsDir() { // either a direct or symlinked directory name
				fHandle.Close()
				fileInfos = myReadDir(loneFilename)
				return fileInfos, loneFilename // return the directory name, as we know it's a directory.
			}

		} else { // err must not be nil after attempting to open loneFilename.
			fHandle.Close()
			loneFilename = workingDir + sep + loneFilename
			loneFilename = filepath.Clean(loneFilename)
		}

		fHandle, err = os.Open(loneFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Println()
			os.Exit(1)
		}

		fi, _ := fHandle.Stat()

		if fi.IsDir() {
			fHandle.Close()
			fileInfos = myReadDir(loneFilename)
			dir := filepath.Dir(loneFilename)
			return fileInfos, dir
		} else { // loneFilename is not a directory, but opening it did not return an error.  So just return its fileInfo.
			fileInfos = append(fileInfos, fi)
			dir := filepath.Dir(loneFilename)
			return fileInfos, dir
		}
	} else { // must have more than one filename on the command line, populated by bash.
		fileInfos = make([]os.FileInfo, 0, narg)
		for _, f := range args {
			fi, err := os.Lstat(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if verboseFlag {
				fmt.Printf(" in loop: fi.Name=%s, fi.Size=%d, fi.IsDir=%t\n", fi.Name(), fi.Size(), fi.IsDir())
			}
			if includeThis(fi) {
				fileInfos = append(fileInfos, fi)
			}
			if fi.Mode().IsRegular() && showGrandTotal {
				grandTotal += fi.Size()
				grandTotalCount++
			}
		}
	}
	if verboseFlag {
		fmt.Printf(" Leaving getFileInfosFromCommandLine.  narg=%d, len(args)=%d, len(fileinfos)=%d\n", narg, len(args), len(fileInfos))
	}
	dir := filepath.Dir(args[0])
	return fileInfos, dir
}

func displayFileInfos(fiSlice []os.FileInfo, dirName string) {
	var lnCount int
	for _, f := range fiSlice {
		s := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizeStr := ""
		usernameStr, groupnameStr := GetUserGroupStr(f)
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			sizeTotal += f.Size()
			if longFileSizeListFlag {
				sizeStr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
				if f.Size() > 100000 {
					sizeStr = AddCommas(sizeStr)
				}
				myPrintf(ct.Yellow, false, "%10v %s:%s %16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizeStr, s, f.Name())
			} else {
				var color ct.Color
				sizeStr, color = getMagnitudeString(f.Size())
				if termDisplayOut {
					myPrintf(color, false, "%10v %s:%s %-16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizeStr, s, f.Name())
				} else {
					fmt.Printf("%10v %s:%s %-16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizeStr, s, f.Name())
				}
			}
			lnCount++

		} else if IsSymlink(f.Mode()) {
			fullFileName := filepath.Join(dirName, f.Name())
			fInfo, err := os.Stat(fullFileName)
			if err != nil {
				fmt.Printf(" Error from os.Stat(%s) is %v\n", f.Name(), err)
				continue
			}
			link, err := os.Readlink(fullFileName)
			if err != nil {
				fmt.Printf(" Error from os.Readlink(%s) is %v\n", fullFileName, err)
				continue
			}
			var color ct.Color
			sizeStr, color = getMagnitudeString(fInfo.Size())
			myPrintf(color, true, "%10v %s:%s %16s %s <%s> [%s]\n", f.Mode(), usernameStr, groupnameStr, sizeStr, s, f.Name(), link)
			lnCount++
		} else if dirList && f.IsDir() {
			fmt.Printf("%10v %s:%s %16s %s (%s)\n", f.Mode(), usernameStr, groupnameStr, sizeStr, s, f.Name())
			lnCount++
		}
		if lnCount >= numOfLines {
			break
		}
	}
}

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

func ThirdFileInfosFromCommandLine() ([]os.FileInfo, string) {
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
			fmt.Fprintf(os.Stderr, " Error from ThirdFileInfosFromCommandLine Getwd is %v\n", er)
			os.Exit(1)
		}
		fileInfos = StdLinearReadDir(workingDir)
	} else { // Must have a pattern on the command line, ie, NArg > 0
		pattern := pflag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.

		if strings.ContainsRune(pattern, ':') {
			pattern = ProcessDirectoryAliases(pattern)
		}
		pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		dirName, fileName = filepath.Split(pattern)
		fileName = strings.ToLower(fileName)
		if dirName != "" && fileName == "" { // then have a dir pattern without a filename pattern
			fileInfos = StdLinearReadDir(dirName)
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
				path = filepath.Join(dirName, f)
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

} // end ThirdFileInfosFromCommandLine
