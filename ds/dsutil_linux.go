package main

import (
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
)

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

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfos will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

func getFileInfosFromCommandLine() []os.FileInfo {
	var fileInfos []os.FileInfo
	if testFlag {
		fmt.Printf(" Entering getFileInfosFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfos))
	}

	workingDir, er := os.Getwd()
	if er != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		if testFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfos = myReadDir(workingDir) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if testFlag {
			fmt.Printf(" after call to myreaddir.  Len(fileInfos)=%d\n", len(fileInfos))
		}

	} else if flag.NArg() == 1 { // a lone name may mean file not found, as bash will populate what it finds.
		var loneFilename string
		const sep = filepath.Separator
		fileInfos = make([]os.FileInfo, 0, 1)
		firstChar := rune(flag.Arg(0)[0])
		if firstChar == sep { // have an absolute path, so don't prepend anything
			loneFilename = flag.Arg(0)
		} else {
			loneFilename = workingDir + string(sep) + flag.Arg(0)
			loneFilename = filepath.Clean(loneFilename)
		}
		fi, err := os.Lstat(loneFilename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "%s is a lone filepath and does not exist.  Exiting\n\n", fi.Name())
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, err)
			fmt.Println()
			os.Exit(1)
		}

		if testFlag {
			fmt.Printf(" in getFileInfosFromCommandLine: loneFilename=%s, fi.Name=%s, IsDir=%t\n", loneFilename, fi.Name(), fi.IsDir())
		}

		if fi.IsDir() {
			fileInfos = myReadDir(loneFilename)
		} else {
			fileInfos = append(fileInfos, fi)
		}

	} else { // must have more than one filename on the command line, populated by bash.
		fileInfos = make([]os.FileInfo, 0, flag.NArg())
		for _, f := range flag.Args() {
			fi, err := os.Lstat(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if testFlag {
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
	if testFlag {
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
	if testFlag {
		fmt.Printf(" In getColorizedString.  len(fiSlice)=%d, len(cs)=%d, numofLines=%d, cols=%d\n", len(fiSlice), len(cs), numOfLines, cols)
	}
	return cs
}
