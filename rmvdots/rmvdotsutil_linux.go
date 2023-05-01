package main

import (
	"flag"
	"fmt"
	"os"
)

//  14 Jan 23 -- I completely rewrote the section of getFileInfosFromCommandLine where there is only 1 identifier on the command line.  This was based on what I learned
//               from args.go.  Let's see if it works.  Basically, I relied too much on os.Lstat or os.Stat.  Now I'm relying on os.Open.
//  12 Apr 23 -- Fixed a bug in GetIDName, which is now called idName to be more idiomatic for Go.
//   1 May 23 -- Now called rmvdotsutil_linux.go, based on dsrtutil_linux.go

// It handles if there are no files populated by bash or file not found by bash.

func getFileNamesFromCommandLine() []string {

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" ERROR from os.Getwd() is %S\n", err)
		fileNames := myReadDirNames(workingDir)
		return fileNames
	}

	if flag.NArg() == 0 {
		fileNames := myReadDirNames(workingDir)
		return fileNames
	} else if flag.NArg() == 1 {
		pattern := flag.Arg(0)
		fHandle, err := os.Open(pattern) // just try to open it, as it may be a symlink.
		if err == nil {
			stat, _ := fHandle.Stat()
			if stat.IsDir() { // either a direct or symlinked directory name
				fileNames, err := fHandle.Readdirnames(0)
				if err != nil {
					fmt.Printf(" ERROR: fHandle opened %s,  err from Readdirnames is %s.  Will use myReadDirNames(%s).\n", fHandle.Name(), err, pattern)
					return myReadDirNames(pattern)
				}
				fHandle.Close()
				return fileNames
			}
		}
	}

	return flag.Args()
}
