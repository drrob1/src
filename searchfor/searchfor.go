package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/spf13/pflag"
)

/*
  27 Sep 25 -- At work I've noticed that when the j.mdb file is a hard link, it doesn't always show up in dv list.  I'm exploring different methods of retrieving the directory list
                 and will match the retrived list against the input param file.
                 I'll need os.Getwd(), os.ReadDir() which returns a slice of dirEntry, and after opening a directory,
                 I can use Readdir() returning []FileInfo, Readdirnames() returning []string and ReadDir() returning []DirEntry.

*/

const lastAltered = "27 Sep 2025"

func main() {
	pflag.Parse()
	fmt.Printf(" searchfor.go last altered %s, compiled with %s\n", lastAltered, runtime.Version())

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" Error from os.Getwd() is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" Working directory is %s\n", workingDir)

	searchTarget := pflag.Arg(0)
	fmt.Printf(" Search target is %s\n", searchTarget)

	_, err = os.Stat(searchTarget)
	if err != nil {
		fmt.Printf(" Error from os.Stat(%s) is %s\n", searchTarget, err)
		os.Exit(1)
	}

	fmt.Printf(" Search target exists\n")

	// os.ReadDir section dealing w/ DirEntry
	DirEntries, err := os.ReadDir(workingDir)
	if err != nil {
		fmt.Printf(" Error from os.ReadDir(%s) is %s\n", workingDir, err)
		os.Exit(1)
	}
	fmt.Printf(" os.ReadDir(%s) succeeded, finding %d entries.\n", workingDir, len(DirEntries))

	lessDirEntries := func(i, j int) bool {
		return DirEntries[i].Name() < DirEntries[j].Name()
	}
	sort.Slice(DirEntries, lessDirEntries)
	position, found := binarySearchDirEntries(DirEntries, searchTarget)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}

	// Now have to open the directory to explore the other functions
	d, err := os.Open(workingDir)
	if err != nil {
		fmt.Printf(" Error from os.Open(%s) is %s\n", workingDir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" os.Open(%s) succeeded.\n", workingDir)

	// os.Readdir section dealing w/ FileInfo
	FileInfoSlice, err := d.Readdir(-1) // -1 means read all.  Zero would also mean read all.  I guess -1 is clearer.
	if err != nil {
		fmt.Printf(" Error from d.Readdir(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Readdir(-1) succeeded.\n")
	lessFileInfo := func(i, j int) bool {
		return FileInfoSlice[i].Name() < FileInfoSlice[j].Name()
	}
	sort.Slice(FileInfoSlice, lessFileInfo)
	position, found = binarySearchFileInfos(FileInfoSlice, searchTarget)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	d.Close() // have to close it after a successful search.

	// os.Readdirnames section.  Have to reopen it
	d, err = os.Open(workingDir)
	if err != nil {
		fmt.Printf(" Error from 2nd os.Open(%s) is %s\n", workingDir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" 2nd os.Open(%s) succeeded.\n", workingDir)
	dirNamesSlice, err := d.Readdirnames(-1)
	if err != nil {
		fmt.Printf(" Error from d.Readdirnames(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Readdirnames(-1) succeeded, finding %d names.\n", len(dirNamesSlice))
	sort.Strings(dirNamesSlice)
	position = sort.SearchStrings(dirNamesSlice, searchTarget)
	if position < len(dirNamesSlice) && dirNamesSlice[position] == searchTarget {
		ctfmt.Printf(ct.Green, true, "Using sort.SearchStrings found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	position, found = binarySearchStrings(dirNamesSlice, searchTarget)
	if found {
		ctfmt.Printf(ct.Green, true, "Using binarySearchStrings found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	err = d.Close()
	if err != nil {
		fmt.Printf(" Error from 2nd d.Close() is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Close() succeeded.\n\n")

	// os.ReadDir section dealing w/ DirEntry
	d, err = os.Open(workingDir)
	if err != nil {
		fmt.Printf(" Error from 3rd os.Open(%s) is %s\n", workingDir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" 3rd os.Open(%s) succeeded.\n", workingDir)
	DirEntries, err = d.ReadDir(-1)
	if err != nil {
		fmt.Printf(" Error from d.ReadDir(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.ReadDir(-1) succeeded, finding %d entries.\n", len(DirEntries))
	sort.Slice(DirEntries, lessDirEntries)
	position, found = binarySearchDirEntries(DirEntries, searchTarget)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
}

func binarySearchDirEntries(slice []os.DirEntry, target string) (int, bool) {
	var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		numTries++
		if slice[current].Name() < target {
			left = current + 1
		} else if slice[current].Name() > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}

func binarySearchFileInfos(slice []os.FileInfo, target string) (int, bool) {
	var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		numTries++
		if slice[current].Name() < target {
			left = current + 1
		} else if slice[current].Name() > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}

func binarySearchStrings(slice []string, target string) (int, bool) {
	var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		numTries++
		if slice[current] < target {
			left = current + 1
		} else if slice[current] > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}
