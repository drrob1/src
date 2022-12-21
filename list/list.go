package list

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 2022 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                   This is going to take a while.
*/

type dirAliasMapType map[string]string

var showGrandTotal, noExtensionFlag, excludeFlag, longFileSizeListFlag, filenameToBeListedFlag, dirList, verboseFlag bool
var filterFlag, globFlag, veryVerboseFlag, halfFlag, maxDimFlag bool
var filterAmt, numLines, numOfLines, grandTotalCount int
var sizeTotal, grandTotal int64
var filterStr string
var excludeRegex *regexp.Regexp
var excludeRegexPattern string
var directoryAliasesMap dirAliasMapType
var fileInfos []os.FileInfo

func NewList(excludeMe *regexp.Regexp, sizeSort, reverse bool) []string {
	return MakeList(excludeMe, sizeSort, reverse)
}

func MakeList(excludeRegex *regexp.Regexp, sizeSort, reverse bool) []string {

	fileInfos = getFileInfosFromCommandLine(excludeRegex)
	fmt.Printf(" length of fileInfos = %d\n", len(fileInfos))

	// set which sort function will be in the sortfcn var
	Forward := !reverse
	DateSort := !sizeSort
	sortFcn := func(i, j int) bool { return false } // became available as of Go 1.8
	if sizeSort && Forward {                        // set the value of sortfcn so only a single line is needed to execute the sort.
		sortFcn = func(i, j int) bool { // closure anonymous function is my preferred way to vary the sort method.
			return fileInfos[i].Size() > fileInfos[j].Size() // I want a largest first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = largest size.")
		}
	} else if DateSort && Forward {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//       return files[i].ModTime().UnixNano() > files[j].ModTime().UnixNano() // I want a newest-first sort
			return fileInfos[i].ModTime().After(fileInfos[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
		}
		if verboseFlag {
			fmt.Println("sortfcn = newest date.")
		}
	} else if sizeSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			return fileInfos[i].Size() < fileInfos[j].Size() // I want an smallest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = smallest size.")
		}
	} else if DateSort && reverse {
		sortFcn = func(i, j int) bool { // this is a closure anonymous function
			//return files[i].ModTime().UnixNano() < files[j].ModTime().UnixNano() // I want an oldest-first sort
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime()) // I want an oldest-first sort
		}
		if verboseFlag {
			fmt.Println("sortfcn = oldest date.")
		}
	}

	if len(fileInfos) > 1 {
		sort.Slice(fileInfos, sortFcn)
	}

	fileString := make([]string, 0, len(fileInfos))
	for _, fi := range fileInfos {
		fileString = append(fileString, fi.Name())
	}
	return fileString
} // end MakeList

// ------------------------------- myReadDir -----------------------------------

func MyReadDir(dir string, excludeMe *regexp.Regexp) []os.FileInfo { // The entire change including use of []DirEntry happens here.  Who knew?
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	fileInfos := make([]os.FileInfo, 0, len(dirEntries))
	for _, d := range dirEntries {
		fi, e := d.Info()
		if e != nil {
			fmt.Fprintf(os.Stderr, " Error from %s.Info() is %v\n", d.Name(), e)
			continue
		}
		if includeThis(fi, excludeMe) {
			fileInfos = append(fileInfos, fi)
		}
	}
	return fileInfos
} // myReadDir

// ---------------------------------------------------- includeThis ----------------------------------------

func includeThis(fi os.FileInfo, excludeRex *regexp.Regexp) bool {
	//if veryVerboseFlag {
	//	fmt.Printf(" includeThis.  noExtensionFlag=%t, excludeFlag=%t, filterAmt=%d \n", noExtensionFlag, excludeFlag, filterAmt)
	//}
	if !fi.Mode().IsRegular() {
		return false
	}
	if noExtensionFlag && strings.ContainsRune(fi.Name(), '.') {
		return false
	} else if filterAmt > 0 {
		if fi.Size() < int64(filterAmt) {
			return false
		}
	}
	if excludeRex.String() != "" {
		if BOOL := excludeRex.MatchString(strings.ToLower(fi.Name())); BOOL {
			return false
		}
	}
	return true
}

//------------------------------ GetDirectoryAliases ----------------------------------------
func getDirectoryAliases() dirAliasMapType { // Env variable is diraliases.

	s, ok := os.LookupEnv("diraliases")
	if !ok {
		return nil
	}

	s = strings.ReplaceAll(s, "_", " ") // substitute the underscore, _, for a space so strings.Fields works correctly
	directoryAliasesMap := make(dirAliasMapType, 10)

	dirAliasSlice := strings.Fields(s)

	for _, aliasPair := range dirAliasSlice {
		if string(aliasPair[len(aliasPair)-1]) != "\\" {
			aliasPair = aliasPair + "\\"
		}
		aliasPair = strings.ReplaceAll(aliasPair, "-", " ") // substitute a dash,-, for a space
		splitAlias := strings.Fields(aliasPair)
		directoryAliasesMap[splitAlias[0]] = splitAlias[1]
	}
	return directoryAliasesMap
} // end getDirectoryAliases

// ------------------------------ ProcessDirectoryAliases ---------------------------

func ProcessDirectoryAliases(aliasesMap dirAliasMapType, cmdline string) string {

	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap = getDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := aliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathNFile := cmdline[idx+1:]
	completeValue := aliasValue + PathNFile
	//fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
	return completeValue
} // ProcessDirectoryAliases
