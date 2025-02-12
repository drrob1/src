package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

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
		  flag.Arg() =                  0      1           2
28 Dec 19 -- changed cd to cdd, so it will change drive and directory at same time.
14 Jan 20 -- my dirsave() also needed to change cd to cdd.  Just done.
16 Sep 20 -- Added sl cmd, as in dirb, as synynom for p.  And will prompt if trying to overwrite a bookmark.
24 Dec 20 -- Will now sort output from Print command.
10 May 21 -- Wrote dirPrint() so will sort output when other commands display the map, esp the save cmd.
12 May 21 -- Remove the dash of an option, if I forget and use it anyway.  These are not options, as I'm not using the flag package.
21 Nov 22 -- I'm here because of static linter.  And there's an issue w/ dirAliasesMap that doesn't need to be a param.
10 Dec 22 -- I'm adding use of flag package, for now I'll just use -v and -h.
18 Feb 23 -- Changing from os.UserHomeDir to os.UserConfigDir.  This is %appdata% or $HOME/.config
14 May 24 -- Have both configDir and HomeDir, which are separate.  I changed to userConfigDir badly last year by blurring these two.
15 May 24 -- Changed all the references to os.Arg[] to flag.Arg() and flag.NArgs.  Not yet fully tested.
28 Nov 24 -- Fixed config entry.
29 Nov 24 -- Added a message to display when -v is used.  And I'm going to use filepath.Join more.
				It's become obvious to me that I should not store the "cdd " part in the map[string]string.  That will be added by dirb.
*/

const LastAltered = "Nov 29, 2024"

const bookmarkFilename = "bookmarkfile.gob"

var HomeDir string             // global because it is also needed in my dirsave rtn.
var bookmark map[string]string // used in all routines.

type bkmksliceType struct {
	key, value string
}

func main() {
	var err error
	fmt.Println(" makedirbkmk is a directory bookmark manager written in Go, last altered", LastAltered, "and compiled w/", runtime.Version())

	verboseFlag := flag.Bool("v", false, "verbose message and exit.")
	helpFlag := flag.Bool("h", false, "help message.")
	flag.Parse()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	if *verboseFlag {
		fmt.Printf(" %s last compiled %s by %s.  Full binary is %s with timestamp of %s.\n", os.Args[0], LastAltered, runtime.Version(), execName, ExecTimeStamp)
	}

	// sep := string(os.PathSeparator)  Not used anymore.  A single backslash on Windows was giving me a lot of trouble because it would not show up in the cdd command.
	HomeDir, err = os.UserHomeDir() // this routine became available in Go 1.12
	if err != nil {
		fmt.Println(err, "Exiting")
		os.Exit(1)
	}
	configDir, err := os.UserConfigDir() // this routine became available in Go 1.14
	if err != nil {
		fmt.Println(err, "Exiting")
		os.Exit(1)
	}
	// target := "cdd " + HomeDir + sep  removed 11/29/24, as I don't want to store the "cdd " part.  That's too complex because a single backslash is not being interpretted correctly.
	//fullBookmarkFilename := configDir + sep + bookmarkFilename // old way before I learned a better way
	fullBookmarkFilename := filepath.Join(configDir, bookmarkFilename)

	if *verboseFlag { // added 11/29/24
		fmt.Printf(" HomeDir: %q, fullBookmarkFilename: %q\n", HomeDir, fullBookmarkFilename)
	}

	help := func() {
		fmt.Println(" HomeDir is", HomeDir, ", ", ExecFI.Name(), "timestamp is", ExecTimeStamp, ". ")
		fmt.Println(" Full exec is", execName, ".  Fullbookmark is", fullBookmarkFilename)
		fmt.Println()
		fmt.Println(" s -- save current directory or entered directory name")
		fmt.Println(" a -- about.")
		fmt.Println(" d -- delete entry given on same line.")
		fmt.Println(" p, sl -- print out bookmark list.")
		fmt.Println(" h -- help.")
		fmt.Println()
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println()
	}
	if *helpFlag {
		help()
		return
	}

	// read or init directory bookmark file
	_, err = os.Stat(fullBookmarkFilename)
	if err == nil { // need to read in bookmarkfile
		bookmarkfile, err := os.Open(fullBookmarkFilename)
		if err != nil {
			log.Fatalln(" cannot open", fullBookmarkFilename, " as input bookmark file, because of", err)
		}
		defer bookmarkfile.Close()
		decoder := gob.NewDecoder(bookmarkfile)
		err = decoder.Decode(&bookmark)
		if err != nil {
			log.Println(" cannot decode", fullBookmarkFilename, ", error is", err, ".  Aborting")
			return
		}
		bookmarkfile.Close()
		fmt.Printf(" %d Bookmarks read in from %s\n", len(bookmark), fullBookmarkFilename)
		fmt.Println()

	} else { // need to init bookmarkfile
		bookmark = make(map[string]string, 15)

		bookmark["config"] = configDir                                  // Join was called above.
		bookmark["docs"] = filepath.Join(HomeDir, "Documents")          // target + "Documents"
		bookmark["doc"] = filepath.Join(HomeDir, "Documents")           // target + "Documents"
		bookmark["inet"] = filepath.Join(HomeDir, "Downloads")          // target + "Downloads"
		bookmark["inetdnld"] = filepath.Join(HomeDir, "Downloads")      // target + "Downloads"
		bookmark["vid"] = filepath.Join(HomeDir, "Videos")              // target + "Videos"
		bookmark["go"] = filepath.Join(HomeDir, "go")                   // target + "go"
		bookmark["src"] = filepath.Join(bookmark["go"], "src")          // target + "go" + sep + "src"
		bookmark["bin"] = filepath.Join(bookmark["go"], "bin")          // target + "go" + sep + "bin"
		bookmark["pic"] = filepath.Join(HomeDir, "Pictures")            // target + "Pictures"
		bookmark["pics"] = filepath.Join(HomeDir, "Pictures")           // target + "Pictures"
		bookmark["winx"] = filepath.Join(bookmark["vid"], "winxvideos") // target + "Videos" + sep + "winxvideos"

		fmt.Println("Bookmark's initialized.")
		fmt.Println()
	}

	ch := ""
	if flag.NArg() == 0 {
		fmt.Println(" usage: makedirbkmk [s|a|d|p|h].  Note that there is no dash preceeding the commands, as these are not options.")
		help()
		os.Exit(0)
	} else {
		ch = strings.ToLower(os.Args[1])
		ch = strings.TrimPrefix(ch, "-") // recommended by static linter
	}

	switch ch {
	case "s": // save current directory or entered directory name
		dirSave()
		fmt.Println()
		fmt.Println()
		dirPrint()
		fmt.Println()
		fmt.Println()

	case "a": // about.
		help()
		return
		//os.Exit(0)

	case "d": // delete entry given on same line
		dirEntryDel()

	case "p", "sl": // print out bookmark list
		dirPrint()
		fmt.Println()
		fmt.Println()
		return

	case "h": // help
		help()
		return

	default:
		fmt.Println(" command not recognized.", ch, "was entered.")
		return
	}

	// write out bookmarkFile
	bookmarkFile, er := os.Create(fullBookmarkFilename)
	if er != nil {
		log.Println(" could not create bookmarkfile upon exiting, because", er)
		return // better to return, because this allows the deferred functions to run.  os.Exit does not run deferred functions.
	}
	defer bookmarkFile.Close()

	encoder := gob.NewEncoder(bookmarkFile)
	e := encoder.Encode(&bookmark)
	if e != nil {
		log.Println(" could not encode bookmarkfile upon exiting, because", e)
	}
	bookmarkFile.Close()
} // end main

// ------------------------------------------ dirsave -----------------------------------------------
func dirSave() { // implement s (save) command
	//    os.Args = makedirbkmk.exe cmd bkmk-name target-directory
	//    len    =         1         2      3           4
	//    subscript =      0         1      2           3
	//   flag.Arg() =                0      1           2
	//   flag.NArgs =                1      2           3

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalln(" could not get working directory because", err)
	}

	if flag.NArg() < 2 { // len(os.Args) <= 2 { // only have cmd without a bookmark name to save
		log.Fatalln(" need bookmark name on command line to use s command.")
	} else if flag.NArg() == 2 { // no directory target name on command line.  Use current directory
		bkmkName := flag.Arg(1)
		_, ok := bookmark[bkmkName]
		if ok { // ok will be true if name already exists in the map.
			fmt.Print(" bookmark named ", bkmkName, " already exists.  Overwrite? [y/N] ")
			ans := ""
			fmt.Scanln(&ans)
			ans = strings.ToLower(ans)
			if strings.HasPrefix(ans, "y") {
				bookmark[bkmkName] = workingDir
				fmt.Printf(" created bookmark[%s] = %s \n", os.Args[2], bookmark[os.Args[2]])
			} else {
				fmt.Println(" save bookmark command ignored.")
			}
		} else { // bookmark name is not already in the map.
			// bookmark[bkmkName] = "cdd " + workingDir
			bookmark[bkmkName] = workingDir
			fmt.Printf(" created bookmark[%s] = %s \n", bkmkName, bookmark[bkmkName])
		}
	} else if flag.NArg() == 3 { // have potential directory target on command line
		target := flag.Arg(2)
		if strings.ContainsRune(target, ':') {
			target = ProcessDirectoryAliases(target)
		}
		target = strings.Replace(target, "~", HomeDir, 1)
		// verify that target is a valid directory or symlink name.
		_, err := os.Lstat(target)
		if err == nil {
			bkmkName := target
			_, ok := bookmark[bkmkName]
			if ok { // ok will be true if name already exists in the map
				fmt.Print(" bookmark named ", bkmkName, " already exists.  Overwrite? [y/N] ")
				ans := ""
				fmt.Scanln(&ans)
				ans = strings.ToLower(ans)
				if strings.HasPrefix(ans, "y") {
					bookmark[bkmkName] = target
					fmt.Printf(" created bookmark[%s] = %s \n", bkmkName, bookmark[bkmkName])
				} else {
					fmt.Println(" save bookmark command ignored.")
				}
			} else { // bookmark name is not already in the map
				bookmark[bkmkName] = target
				fmt.Printf(" created bookmark[%s] = %s \n", bkmkName, bookmark[bkmkName])
			}
		} else {
			log.Println(" Lstat call for", target, "failed with error of", err)
			return
		}
	}

} // end dirSave

// ------------------------------------------ dirEntryDel ----------------------------------
func dirEntryDel() { // implement d (delete) command.
	//    os.Args = makedirbkmk.exe cmd bkmk-name target-directory
	//    len    =         1         2      3           4
	//    subscript =      0         1      2           3
	//   flag.Arg() =                0      1           2
	//   flag.NArgs =                1      2           3

	if flag.NArg() < 2 {
		log.Fatalln(" need bookmark name on command line to be deleted.  flag.Args is ", flag.Args(), ", len is", flag.NArg())
	}

	name := flag.Arg(1)
	target, ok := bookmark[name]
	if ok {
		delete(bookmark, name)
	} else {
		log.Println(name, " not in bookmark map.")
	}
	fmt.Printf(" deleted bookmark[%s] which referenced %s", name, target)
} // end dirEntryDel

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

//func ProcessDirectoryAliases(aliasesMap map[string]string, cmdline string) string {

func ProcessDirectoryAliases(cmdline string) string {
	idx := strings.IndexRune(cmdline, ':')
	if idx < 2 { // note that if rune is not found, function returns -1.
		return cmdline
	}
	aliasesMap := GetDirectoryAliases()
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

func dirPrint() {
	bkmkslice := make([]bkmksliceType, 0, len(bookmark))
	for idx, valu := range bookmark {
		bkmk := bkmksliceType{idx, valu} // structured literal syntax
		bkmkslice = append(bkmkslice, bkmk)
	}
	sortless := func(i, j int) bool {
		return bkmkslice[i].key < bkmkslice[j].key
	}
	sort.Slice(bkmkslice, sortless)
	for _, bkmk := range bkmkslice {
		fmt.Printf(" bookmark[%s] = %s\n", bkmk.key, bkmk.value)
	}
}
