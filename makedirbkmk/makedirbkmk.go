package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
)

const LastAltered = "Dec 24, 2020"

/*
REVISION HISTORY
================
 2 Dec 19 -- Started development of make directory bookmark, using dirbkmk.go as a start.  Will focus on Windows, as bash already has dirb.
			   This routine will write the map of bookmarks as a gob file, that makedirbkmk.go will read.  This is copying code from
			   primes2.go and makeprimeslice.go.
			   It will need these commands:
				 s -- save bookmark
				 a -- about
				 d -- delete bookmark
				 p -- print bookmarks
				 h -- help
 4 Dec 19 -- Will make both directoryAliasesMap and bookmark global.  I'm not doing concurrency anyway.
			 os.Args = makedirbkmk.exe cmd bkmk-name target-directory
		   subscript =      0           1      2           3
			  len    =      1           2      3           4
28 Dec 19 -- changed cd to cdd, so it will change drive and directory at same time.
14 Jan 20 -- my dirsave() also needed to change cd to cdd.  Just done.
16 Sep 20 -- Added sl cmd, as in dirb, as synynom for p.  And will prompt if trying to overwrite a bookmark.
24 Dec 20 -- Will now sort output from Print command.
*/

const bookmarkfilename = "bookmarkfile.gob"

var HomeDir string             // global because it is also needed in my dirsave rtn.
var bookmark map[string]string // used in all routines.

type bkmkslicetype struct {
	key, value string
}

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
		fmt.Println(" p, sl -- print out bookmark list.")
		fmt.Println(" h -- help.")
		fmt.Println()
		fmt.Println()
	}

	ch := ""
	if len(os.Args) < 2 {
		fmt.Println(" usage: makedirbkmk [s|a|d|p|h].  Note that there is no dash preceeding the commands, as these are not options.")
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

	case "p", "sl": // print out bookmark list
		bkmkslice := make([]bkmkslicetype, 0, len(bookmark))
		for idx, valu := range bookmark {
			bkmk := bkmkslicetype{idx, valu}  // structured literal syntax
			bkmkslice = append(bkmkslice, bkmk)
		}
		sortless := func (i,j int) bool {
			return bkmkslice[i].key < bkmkslice[j].key
		}
		sort.Slice(bkmkslice, sortless)
		for _, bkmk := range bkmkslice {
			fmt.Printf(" bookmark[%s] = %s\n", bkmk.key, bkmk.value)
		}
		fmt.Println()
		fmt.Println()

	case "h": // help
		help()

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
func dirsave() { // implement s (save) command
	//    os.Args = makedirbkmk.exe cmd bkmk-name target-directory
	//    len    =         1         2      3           4
	//    subscript =      0         1      2           3

	workingdir, err := os.Getwd()
	if err != nil {
		log.Fatalln(" could not get working directory because", err)
	}

	if len(os.Args) <= 2 { // only have cmd without a bookmark name to save
		log.Fatalln(" need bookmark name on command line to use s command.")
	} else if len(os.Args) == 3 { // no directory target name on command line.  Use current directory
		bkmkname := os.Args[2]
		_, ok := bookmark[bkmkname]
		if ok { // ok will be true if name already exists in the map.
			fmt.Print(" bookmark named ", bkmkname, " already exists.  Overwrite? [y/N] ")
			ans := ""
			fmt.Scanln(&ans)
			ans = strings.ToLower(ans)
			if strings.HasPrefix(ans, "y") {
				bookmark[bkmkname] = "cdd " + workingdir
				fmt.Printf(" created bookmark[%s] = %s \n", os.Args[2], bookmark[os.Args[2]])
			} else {
				fmt.Println(" save bookmark command ignored.")
			}
		} else { // bookmark name is not already in the map.
			bookmark[bkmkname] = "cdd " + workingdir
			fmt.Printf(" created bookmark[%s] = %s \n", os.Args[2], bookmark[os.Args[2]])
		}
	} else if len(os.Args) == 4 { // have potential directory target on command line
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
			bkmkname := os.Args[2]
			_, ok := bookmark[bkmkname]
			if ok { // ok will be true if name already exists in the map
				fmt.Print(" bookmark named ", bkmkname, " already exists.  Overwrite? [y/N] ")
				ans := ""
				fmt.Scanln(&ans)
				ans = strings.ToLower(ans)
				if strings.HasPrefix(ans, "y") {
					bookmark[bkmkname] = "cdd " + target
					fmt.Printf(" created bookmark[%s] = %s \n", os.Args[2], bookmark[os.Args[2]])
				} else {
					fmt.Println(" save bookmark command ignored.")
				}
			} else { // bookmark name is not already in the map
				bookmark[bkmkname] = "cdd " + target
				fmt.Printf(" created bookmark[%s] = %s \n", os.Args[2], bookmark[os.Args[2]])
			}
		} else {
			log.Fatalln(" Lstat call for", target, "failed with error of", err)
		}
	}

} // end dirsave

// ------------------------------------------ direntrydel ----------------------------------
func direntrydel() { // implement d (delete) command.
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
		log.Println(os.Args[2], " not in bookmark map.")
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

	s = MakeSubst(s, '_', ' ') // substitute the underscore, _, for a space
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
