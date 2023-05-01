package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

/*
REVISION HISTORY
-------- -------
 8 Feb 22 -- Found a bug in glob flag when I tested with "z:*.TXT" that didn't happen w/ the separate glob.go, fast.go or dsrt z:*.txt
               I got an error from Lstat in that it tried to Lstat z:\z:filename.TXT.
15 Feb 22 -- Replaced testFlag w/ verboseFlag.  Finally.
24 Feb 22 -- Fixed a bug in the glob option.  And Evan's 30 today.  Wow.
12 Apr 23 -- Fixed a bug in GetIDName, which is now called idName to be more idiomatic for Go.  But that is not called here in Windows code, so nevermind.
 1 May 23 -- Now called rmvdotsutil_windows.go
*/

// getFileNamesFromCommandLine() will return a slice of FileInfos after the filter and exclude expression are processed, and that match a pattern if given.
// It handles if there are no files populated by bash or file not found by bash.

func getFileNamesFromCommandLine() []string {

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = "."
	}
	HomeDirStr = HomeDirStr + string(filepath.Separator)
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" ERROR from os.Getwd() is %s.  Exiting\n", err)
	}

	if flag.NArg() == 0 {
		fileNames := myReadDirNames(workingDir)
		return fileNames
	}

	pattern := flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.

	if strings.ContainsRune(pattern, ':') {
		pattern = ProcessDirectoryAliases(pattern)
	}
	pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
	dirName, fileNamePattern := filepath.Split(pattern)
	fileNamePattern = strings.ToLower(fileNamePattern)
	if dirName != "" && fileNamePattern == "" { // then have a dir pattern without a filename pattern
		fileNames := myReadDirNames(dirName)
		return fileNames
	}
	if dirName == "" {
		dirName = workingDir
	}
	if fileNamePattern == "" { // need this to not be blank because of the call to Match below.
		fileNamePattern = "*"
	}

	d, err := os.Open(dirName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from Windows processCommandLine directory os.Open is %v\n", err)
		os.Exit(1)
	}
	defer d.Close()
	rawFileNames, err := d.Readdirnames(0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
		rawFileNames = myReadDirNames(dirName)
		return rawFileNames
	}

	fileNames := make([]string, 0, len(rawFileNames))
	const sepStr = string(os.PathSeparator)

	for _, f := range rawFileNames {
		var fPath string
		if strings.Contains(f, sepStr) || strings.Contains(f, ":") {
			fPath = f
		} else {
			fPath = dirName + sepStr + f
		}

		match, err := filepath.Match(fileNamePattern, strings.ToLower(f))
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern is %v.\n", pattern, err)
			continue
		}
		if match {
			fileNames = append(fileNames, fPath)
		}
	}

	return fileNames

} // end getFileNamesFromCommandLine
