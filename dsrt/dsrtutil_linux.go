package main

import (
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
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
}

// processCommandLine will return a slice of FileInfos after the filter and exclude expression are processed
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.
func processCommandLine(sortfcn func(i, j int) bool) FISliceType {
	var GrandTotal int64
	var GrandTotalCount int
	var fileInfos FISliceType
	var workingDir string
	var er error

	if flag.NFlag() == 0 {
		workingDir, er = os.Getwd()
		if er != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			os.Exit(1)
		}
		d, err := os.Open(workingDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine os.Open is %v\n", err)
			os.Exit(1)
		}

		fileInfos, err = d.Readdir(0) // I don't know if I have to make this slice first.  I'm going to assume not for now.
		if err != nil {               // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
			fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
			fileInfos = MyReadDir(workingDir)
		}
		if ShowGrandTotal {
			for _, f := range fileInfos {
				if f.Mode().IsRegular() {
					GrandTotal += f.Size()
					GrandTotalCount++
				}
			}
		}
		sort.Slice(fileInfos, sortfcn)
		return fileInfos
	} else if flag.NFlag() == 1 { // a lone name may mean file not found, as bash will populate what it finds.
		loneFilename := flag.Arg(0)
		fi, err := os.Lstat(loneFilename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "%s does not exist.\n", fi.Name())
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, err)
			fmt.Println()
			os.Exit(1)
		}

		// even if this is a directory it must be returned.  A different routine will decide whether and how to display this.
		fileInfos = append(fileInfos, fi)
		return fileInfos

	} else { // must have more than one filename on the command line, populated by bash.
		filenames := flag.Args()
		fileInfos = make(FISliceType, 0, len(filenames))
		for _, f := range filenames {
			fi, err := os.Lstat(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			if includeThis(fi) {
				fileInfos = append(fileInfos, fi)
			}
			if fi.Mode().IsRegular() && ShowGrandTotal {
				GrandTotal += fi.Size()
				GrandTotalCount++
			}
			sort.Slice(fileInfos, sortfcn)
			return fileInfos
		}
	}
	panic(" Linux processCommandLine and should never have gotten here.")
}

func includeThis(fi os.FileInfo) bool {
	showThis := true
	if noExtensionFlag && strings.ContainsRune(fi.Name(), '.') {
		showThis = false
	}
	if excludeFlag {
		if BOOL := excludeRegex.MatchString(strings.ToLower(fi.Name())); BOOL {
			showThis = false
		}
	}
	if filterAmt > 0 {
		if fi.Size() < int64(filterAmt) {
			showThis = false
		}
	}
	return showThis
}

func displayFileInfos(fiSlice FISliceType) {
	var lncount int
	for _, f := range fiSlice {
		s := f.ModTime().Format("Jan-02-2006_15:04:05")
		sizestr := ""
		usernameStr, groupnameStr := GetUserGroupStr(f) // util function in platform specific removed Oct 4, 2019 and then unremoved.
		if FilenameToBeListed && f.Mode().IsRegular() {
			if LongFileSizeList {
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
			lncount++

		}

	}

}
