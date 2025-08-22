package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/spf13/pflag"
)

//  12 Apr 23 -- Fixed a bug in GetIDName, which is now called idName to be more idiomatic for Go.
//  21 Jun 25 -- DisplayFileInfos now knows when output is redirected.
//  22 Jun 25 -- myPrintf now used.
//  21 Aug 25 -- Now gets more info for symlinks by using a separate call to lstat.  I don't yet know if this is needed on linux too.
//                Lstat makes no attempt to follow the symlink.  I think Stat does follow the symlink.
//                To really be able to do that, I need to return the dirname from the getFileInfosFromCommandLine call.  And then pass that into the displayFileInfos call.
//                This doesn't work the same as it does on Windows.  The symlink shows up as an error if called from another directory.  It does work from the same directory.
//                I don't yet know what to do about this.

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
				return fileInfos, ""
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
				fmt.Printf(" Error from os.Stat on %s is %v\n", f.Name(), err)
				continue
			}
			var color ct.Color
			sizeStr, color = getMagnitudeString(fInfo.Size())
			myPrintf(color, true, "%10v %s:%s %16s %s <%s>\n", f.Mode(), usernameStr, groupnameStr, sizeStr, s, f.Name())
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
