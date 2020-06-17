package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

const LastAltered = "Jan 14, 2020"

/*
  REVISION HISTORY
  ================
   2 Dec 2019 -- Started development of make directory bookmark, using dirbkmk.go as a start.  Will focus on Windows, as bash already has dirb.
                   This routine will write the map of bookmarks as a gob file, that dirmkbk2.go will read.  This is copying code from
                   primes2.go and makeprimeslice.go.
                   It will need these commands:
                     s -- save bookmark
                     a -- about
                     d -- delete bookmark
                     p -- print bookmarks
                     h -- help
   4 Dec 2019 -- Will make both directoryAliasesMap and bookmark global.  I'm not doing concurrency anyway.
                 os.Args = makedirbkmk.exe cmd bkmk-name target-directory
               subscript =      0           1      2           3
                  len    =      1           2      3           4
  28 Dec 2019 -- changed cd to cdd, so it will change drive and directory at same time.
  14 Jan 2020 -- dirsave() also needed to change cd to cdd.  Just done.
*/

const bookmarkfilename = "bookmarkfile.gob"

var HomeDir string             // global because it is also needed in dirsave
var bookmark map[string]string // used in all routines.

func main() {

	fmt.Println(" makedirbkmk written in Go, last altered", LastAltered)
	sep := string(os.PathSeparator)
	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
	} else {
		fmt.Println(" not running on expected platform.  Will exit.  In fact, probably won't even compile.")
		os.Exit(1)
	}
	target := "cdd" + " " + HomeDir + sep
	fullbookmarkfilename := HomeDir + sep + bookmarkfilename

	// read or init directory bookmark file
	_, err := os.Stat(fullbookmarkfilename)
	if err == nil { // need to read in bookmarkfile
		bookmarkfile, err := os.Open(fullbookmarkfilename)
		if err != nil {
			log.Fatalln(" cannot open", fullbookmarkfilename, " as input bookmark file, because of", err)
		}
		defer bookmarkfile.Close()
		decoder := gob.NewDecoder(bookmarkfile)
		err = decoder.Decode(&bookmark)
		if err != nil {
			log.Fatalln(" cannot decode", fullbookmarkfilename, ", error is", err, ".  Aborting")
		}
		bookmarkfile.Close()
		fmt.Println(" Bookmarks read in from", fullbookmarkfilename)
		fmt.Println()

	} else { // need to init bookmarkfile
		bookmark = make(map[string]string, 10)

		bookmark["docs"] = target + "Documents"
		bookmark["doc"] = target + "Documents"
		bookmark["inet"] = target + "Downloads"
		bookmark["inetdnld"] = target + "Downloads"
		bookmark["vid"] = target + "Videos"
		bookmark["go"] = target + "go"
		bookmark["src"] = target + "go" + sep + "src"
		bookmark["pic"] = target + "Pictures"
		bookmark["pics"] = target + "Pictures"
		bookmark["winx"] = target + "Videos" + sep + "winxvideos"
		bookmark["bin"] = target + "go" + sep + "bin"

		fmt.Println("Bookmark's initialized.")
		fmt.Println()
	}

	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	help := func() {
		fmt.Println(" dirbkmk, a Directory Bookmark program written in Go.  Last altered", LastAltered)
		fmt.Println(" HomeDir is", HomeDir, ", ", ExecFI.Name(), "timestamp is", ExecTimeStamp, ". ")
		fmt.Println(" Full exec is", execname, ".  Fullbookmark is", fullbookmarkfilename)
		fmt.Println()
		fmt.Println(" s -- save current directory or entered directory name")
		fmt.Println(" a -- about.")
		fmt.Println(" d -- delete entry given on same line.")
		fmt.Println(" p -- print out bookmark list.")
		fmt.Println(" h -- help.")
		fmt.Println()
		fmt.Println()
	}

	ch := ""
	if len(os.Args) < 2 {
		fmt.Println(" usage: makedirbkmk [s|a|d|p|h]")
		os.Exit(0)
	} else {
		ch = strings.ToLower(os.Args[1])
	}
	fmt.Println(" os.Args is", os.Args)
	fmt.Println()

	switch ch {
	case "s": // save current directory or entered directory name
		dirsave()
		fmt.Println()
		fmt.Println()
		for idx, valu := range bookmark {
			fmt.Printf(" bookmark[%s] = %s \n", idx, valu)
		}
		fmt.Println()
		fmt.Println()

	case "a": // about.
		help()

	case "d": // delete entry given on same line
		direntrydel()

	case "p": // print out bookmark list
		for idx, valu := range bookmark {
			fmt.Printf(" bookmark[%s] = %s \n", idx, valu)
		}
		fmt.Println()
		fmt.Println()

	case "h": // help
		help()

	case "i": // initialize map w/ my 10 entries

	default:
		fmt.Println(" command not recognized.", ch, "was entered.")
		os.Exit(0)
	}

	// write out bookmarkfile
	bookmarkfile, er := os.Create(fullbookmarkfilename)
	if er != nil {
		log.Fatalln(" could not create bookmarkfile upon exiting, because", er)
	}
	defer bookmarkfile.Close()

	encoder := gob.NewEncoder(bookmarkfile)
	e := encoder.Encode(&bookmark)
	if e != nil {
		log.Println(" could not encode bookmarkfile upon exiting, because", e)
	}
	bookmarkfile.Close()
} // end main

// ------------------------------------------ dirsave -----------------------------------------------
func dirsave() {
	//    os.Args = makedirbkmk.exe cmd bkmk-name target-directory
	//    len    =         1         2      3           4
	//    subscript =      0         1      2           3

	workingdir, err := os.Getwd()

	if err != nil {
		log.Fatalln(" could not get working directory because", err)
	}
	if len(os.Args) <= 2 { // only have cmd without a bookmark name to save
		log.Fatalln(" need bookmark name on command line to use s command.")
	} else if len(os.Args) <= 3 { // no directory target name on command line.  Use current directory
		bookmark[os.Args[2]] = "cdd " + workingdir
	} else if len(os.Args) <= 4 { // have potential directory target on command line
		target := os.Args[3]
		if strings.ContainsRune(target, ':') {
			directoryAliasesMap := GetDirectoryAliases()
			target = ProcessDirectoryAliases(directoryAliasesMap, target)
		} else if strings.Contains(target, "~") {
			target = strings.Replace(target, "~", HomeDir, 1)
		}
		// verify that target is a valid directory or symlink name.
		_, err := os.Lstat(target)
		if err == nil {
			bookmark[os.Args[2]] = "cdd " + target
		} else {
			log.Fatalln(" Lstat call for", target, "failed with error of", err)
		}

	}
	fmt.Printf(" created bookmark[%s] = %s \n", os.Args[2], bookmark[os.Args[2]])
} // end dirsave

// ------------------------------------------ direntrydel ----------------------------------
func direntrydel() {
	//    os.Args = makedirbkmk.exe cmd bkmk-name target-directory
	//    len    =         1         2      3           4
	//    subscript =      0         1      2           3

	if len(os.Args) <= 2 {
		log.Fatalln(" need bookmark name on command line to be deleted.  os.Args is ", os.Args, ", len(os.Args) is", len(os.Args))
	}

	targt, ok := bookmark[os.Args[2]]
	if ok {
		delete(bookmark, os.Args[2])
	} else {
		log.Fatalln(os.Args[2], " not in bookmark map.")
	}
	fmt.Printf(" deleted bookmark[%s] which referenced %s", os.Args[2], targt)
} // end direntrydel

// --------------------------- MakeSubst -------------------------------------------
func MakeSubst(instr string, r1, r2 rune) string {

	inRune := make([]rune, len(instr))
	if !strings.ContainsRune(instr, r1) {
		return instr
	}

	for i, s := range instr {
		if s == r1 {
			s = r2
		}
		inRune[i] = s // was byte(s) before I made this a slice of runes.
	}
	return string(inRune)
} // makesubst

//------------------------------ GetDirectoryAliases ----------------------------------------
func GetDirectoryAliases() map[string]string { // Env variable is diraliases.

	s := os.Getenv("diraliases")
	if len(s) == 0 {
		return nil
	}

	s = MakeSubst(s, '_', ' ') // substitute the underscore, _, or a space
	directoryAliasesMap := make(map[string]string, 10)
	//anAliasMap := make(dirAliasMapType,1)

	dirAliasSlice := strings.Fields(s)

	for _, aliasPair := range dirAliasSlice {
		if string(aliasPair[len(aliasPair)-1]) != "\\" {
			aliasPair = aliasPair + "\\"
		}
		aliasPair = MakeSubst(aliasPair, '-', ' ') // substitute a dash,-, for a space
		splitAlias := strings.Fields(aliasPair)
		directoryAliasesMap[splitAlias[0]] = splitAlias[1]
	}
	return directoryAliasesMap
} // end GetDirectoryAliases

// ------------------------------ ProcessDirectoryAliases ---------------------------
func ProcessDirectoryAliases(aliasesMap map[string]string, cmdline string) string {

	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap = GetDirectoryAliases()
	aliasName := cmdline[:idx] // substring of directory alias not including the colon, :
	aliasValue, ok := aliasesMap[aliasName]
	if !ok {
		return cmdline
	}
	PathnFile := cmdline[idx+1:]
	completeValue := aliasValue + PathnFile
	fmt.Println("in ProcessDirectoryAliases and complete value is", completeValue)
	return completeValue
} // ProcessDirectoryAliases
