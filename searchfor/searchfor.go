package main

import (
	"fmt"
	"os"
	"path/filepath"
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
  28 Sep 25 -- Added ability to include directory name in the search.
*/

const lastAltered = "28 Sep 2025"

func main() {
	pflag.Parse()
	fmt.Printf(" searchfor.go last altered %s, compiled with %s\n", lastAltered, runtime.Version())

	if pflag.NArg() != 1 {
		fmt.Printf(" This pgm searches for the file given as its first parameter to see if it exists, and which os routine can find it.\n")
		fmt.Printf(" Usage: searchfor <file>\n")
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" Error from os.Getwd() is %s\n", err)
		os.Exit(1)
	}

	searchTarget := pflag.Arg(0)

	_, err = os.Stat(searchTarget)
	if err != nil {
		fmt.Printf(" Error from os.Stat(%s) is %s\n", searchTarget, err)
		os.Exit(1)
	}

	fmt.Printf(" Search target exists\n")

	dir, target := filepath.Split(searchTarget)
	if dir == "" {
		dir = workingDir
	}
	fmt.Printf(" Search directory is %s, search target is %s\n", dir, target)

	// os.ReadDir section dealing w/ DirEntry
	DirEntries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf(" Error from os.ReadDir(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	fmt.Printf(" os.ReadDir(%s) succeeded, finding %d dir entries.\n", dir, len(DirEntries))

	lessDirEntries := func(i, j int) bool {
		return DirEntries[i].Name() < DirEntries[j].Name()
	}
	sort.Slice(DirEntries, lessDirEntries)
	position, found := binarySearchDirEntries(DirEntries, target)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n\n", searchTarget)
	}

	// Now have to open the directory to explore the other functions
	d, err := os.Open(dir)
	if err != nil {
		fmt.Printf(" Error from os.Open(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" os.Open(%s) succeeded.\n", dir)

	// os.Readdir section dealing w/ FileInfo
	FileInfoSlice, err := d.Readdir(-1) // -1 means read all.  Zero would also mean read all.  I guess -1 is clearer.
	if err != nil {
		fmt.Printf(" Error from d.Readdir(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Readdir(-1) succeeded, finding %d FileInfos.\n", len(FileInfoSlice))
	lessFileInfo := func(i, j int) bool {
		return FileInfoSlice[i].Name() < FileInfoSlice[j].Name()
	}
	sort.Slice(FileInfoSlice, lessFileInfo)
	position, found = binarySearchFileInfos(FileInfoSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	d.Close() // have to close it after a successful search.
	fmt.Printf(" 1st d.Close() succeeded.\n\n")

	// os.Readdirnames section.  Have to reopen it
	d, err = os.Open(dir)
	if err != nil {
		fmt.Printf(" Error from 2nd os.Open(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" 2nd os.Open(%s) succeeded.\n", dir)
	dirNamesStringSlice, err := d.Readdirnames(-1)
	if err != nil {
		fmt.Printf(" Error from d.Readdirnames(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Readdirnames(-1) succeeded, finding %d names.\n", len(dirNamesStringSlice))
	sort.Strings(dirNamesStringSlice)
	position = sort.SearchStrings(dirNamesStringSlice, target)
	if position < len(dirNamesStringSlice) && dirNamesStringSlice[position] == target {
		ctfmt.Printf(ct.Green, true, "Using sort.SearchStrings found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	position, found = binarySearchStrings(dirNamesStringSlice, target)
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
	fmt.Printf(" 2nd d.Close() succeeded.\n\n")

	// os.ReadDir section dealing w/ DirEntry
	d, err = os.Open(dir)
	if err != nil {
		fmt.Printf(" Error from 3rd os.Open(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" 3rd os.Open(%s) succeeded.\n", dir)
	DirEntries, err = d.ReadDir(-1)
	if err != nil {
		fmt.Printf(" Error from d.ReadDir(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.ReadDir(-1) succeeded, finding %d dir entries.\n", len(DirEntries))
	sort.Slice(DirEntries, lessDirEntries)
	position, found = binarySearchDirEntries(DirEntries, target)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	fmt.Printf("\n")
}

func binarySearchDirEntries(slice []os.DirEntry, target string) (int, bool) {
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
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
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
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
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
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
