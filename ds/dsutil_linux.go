package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"os"
	"path/filepath"
	"strconv"
)

/*  Not used, so I'll comment it out.
func GetUserGroupStr(fi os.FileInfo) (usernameStr, groupnameStr string) {
	if runtime.GOARCH != "amd64" { // 06/20/2019 11:23:40 AM made condition not equal, and will remove conditional from dsrt.go
		return "", ""
	}
	sysUID := int(fi.Sys().(*syscall.Stat_t).Uid) // Stat_t is a uint32
	uidStr := strconv.Itoa(sysUID)
	sysGID := int(fi.Sys().(*syscall.Stat_t).Gid) // Stat_t is a uint32
	gidStr := strconv.Itoa(sysGID)
	usernameStr = GetIDname(uidStr)
	groupnameStr = GetIDname(gidStr)
	return usernameStr, groupnameStr
} // end GetUserGroupStr
*/

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

// on Jan 14, 2023 I completely rewrote the section of getFileInfosFromCommandLine where there is only 1 identifier on the command line.  This was based on what I learned
// from args.go.  It worked in dsrt, so now I'm porting that code here.  Basically, I relied too much on os.Lstat or os.Stat.  Now I'm relying on os.Open.

func getFileInfosFromCommandLine() []os.FileInfo {
	var fileInfos []os.FileInfo
	if verboseFlag {
		fmt.Printf(" Entering getFileInfosFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfos))
	}

	workingDir, er := os.Getwd()
	if er != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		if verboseFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfos = myReadDir(workingDir) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if verboseFlag {
			fmt.Printf(" after call to myreaddir.  Len(fileInfos)=%d\n", len(fileInfos))
		}

	} else if flag.NArg() == 1 { // a lone name may either mean file not found or it's a directory which could be a symlink.
		const sep = string(filepath.Separator)
		fileInfos = make([]os.FileInfo, 0, 1)
		loneFilename := flag.Arg(0)
		//firstChar := rune(flag.Arg(0)[0])  bad idea.  Deleted 1/14/23.
		fHandle, err := os.Open(loneFilename) // just try to open it, as it may be a symlink.
		if err == nil {
			stat, _ := fHandle.Stat()
			if stat.IsDir() { // either a direct or symlinked directory name
				fHandle.Close()
				fileInfos = myReadDir(loneFilename)
				return fileInfos
			}

		} else { // err must not be nil after attempting to open loneFilename.
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
			return fileInfos
		} else { // loneFilename is not a directory, but opening it did not return an error.  So just return its fileInfo.
			fileInfos = append(fileInfos, fi)
			return fileInfos
		}

	} else { // must have more than one filename on the command line, populated by bash.
		fileInfos = make([]os.FileInfo, 0, flag.NArg())
		for _, f := range flag.Args() {
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
		fmt.Printf(" Leaving getFileInfosFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfos))
	}
	return fileInfos
}

func getColorizedStrings(fiSlice []os.FileInfo, cols int) []colorizedStr {

	cs := make([]colorizedStr, 0, len(fiSlice))

	for i, f := range fiSlice {
		t := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizeStr := ""
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			sizeTotal += f.Size()
			if longFileSizeListFlag { // changed 5 Feb 22.  All digits of length can only be seen by dsrt now.
				sizeStr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
				if f.Size() > 100000 {
					sizeStr = AddCommas(sizeStr)
				}
				var colr ct.Color
				sizeStr, colr = getMagnitudeString(f.Size())
				strng := fmt.Sprintf("%10v %-10s %s %s", f.Mode(), sizeStr, t, f.Name())
				colorized := colorizedStr{color: colr, str: strng}
				cs = append(cs, colorized)

			} else { // by default, the mode bits will not be shown.  Need longFileSizeListFlag to see the mode bits.
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
		fmt.Printf(" In getColorizedString.  len(fiSlice)=%d, len(cs)=%d, numofLines=%d, cols=%d\n", len(fiSlice), len(cs), numOfLines, cols)
	}
	return cs
}
