package main

import (
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
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

// getFileInfosFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
func getFileInfosFromCommandLine() FISliceType {
	var fileInfos FISliceType
	if testFlag {
		fmt.Printf(" Entering getFileInfosFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfos))
	}

	if flag.NArg() == 0 {
		workingDir, er := os.Getwd()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			os.Exit(1)
		}

		if testFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfos = MyReadDir(workingDir) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if testFlag {
			fmt.Printf(" after call to myreaddir.  Len(fileInfos)=%d\n", len(fileInfos))
		}

	} else if flag.NArg() == 1 { // a lone name may mean file not found, as bash will populate what it finds.
		fileInfos = make(FISliceType, 0, 1)
		loneFilename := flag.Arg(0)
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

		// even if this is a directory it must be returned.  A different routine will decide whether and how to display this.
		fileInfos = append(fileInfos, fi)

	} else { // must have more than one filename on the command line, populated by bash.
		//filenames := flag.Args()
		fileInfos = make(FISliceType, 0, flag.NArg())
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

func displayFileInfos(fiSlice FISliceType) {
	var lnCount int
	for _, f := range fiSlice {
		s := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizestr := ""
		usernameStr, groupnameStr := GetUserGroupStr(f) // util function in platform specific removed Oct 4, 2019 and then unremoved.
		if filenameToBeListedFlag && f.Mode().IsRegular() {
			sizeTotal += f.Size()
			if longFileSizeListFlag {
				sizestr = strconv.FormatInt(f.Size(), 10) // will convert int64.  Itoa only converts int.  This matters on 386 version.
				if f.Size() > 100000 {
					sizestr = AddCommas(sizestr)
				}
				ctfmt.Printf(ct.Yellow, false, "%10v %s:%s %16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
			} else {
				var color ct.Color
				sizestr, color = getMagnitudeString(f.Size())
				ctfmt.Printf(color, false, "%10v %s:%s %-16s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
			}
			lnCount++

		} else if IsSymlink(f.Mode()) {
			fmt.Printf("%10v %s:%s %16s %s <%s>\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
			lnCount++
		} else if dirList && f.IsDir() {
			fmt.Printf("%10v %s:%s %16s %s (%s)\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
			lnCount++
		}
		if lnCount >= numOfLines {
			break
		}
	}
}
