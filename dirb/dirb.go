// dirb.go, from code of dirbkmk.go
package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

const LastAltered = "Dec 6, 2019"

const bookmarkfilename = "bookmarkfile.gob"

/*
  REVISION HISTORY
  ================
   1 Dec 2019 -- Started development of this, using cdtest.go as a start.  Will focus on Windows, as bash already has dirb.
                   For it to work on windows, need %@execstr[] command in tcc.  And needed io.WriteString for both tcc and bash.
                   I'm hard coding the directories because I think it would be faster than loading from a file each time.  For now.
                   Aliasdef has to include  g %@execstr[dirbkmk %1], and it works like dirb on bash!
   2 Dec 2019 -- Added code to output more info under help || about
   6 Dec 2019 -- Now called dirb, based on code from dirbkmk.  Makedirbkmk works, so now I can use that map.  Or at least try it out.
*/

func main() {
	var bookmark map[string]string
	var HomeDir string
	sep := string(os.PathSeparator)

	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME") + sep
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile") + sep
	} else {
		fmt.Println(" not running on expected platform.  Will exit.  In fact, probably won't even compile.")
		os.Exit(1)
	}
	target := "cd" + " " + HomeDir
	fullbookmarkfilename := HomeDir + bookmarkfilename

	if len(os.Args) == 1 {
		io.WriteString(os.Stdout, target)
		os.Exit(0)
	}

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
		/*
			fmt.Println(" Bookmark's read in, and are:")
			for idx, valu := range bookmark {
				fmt.Printf(" bookmark[%s] = %s \n", idx, valu)
			}
			fmt.Println()
			fmt.Println()
		*/
	} else { // need to init bookmarkfile
		bookmark = make(map[string]string, 15)

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

		//		fmt.Println("Bookmark's initialized.")
	}

	if strings.ToLower(os.Args[1]) == "help" || os.Args[1] == "about" {
		fmt.Println(" dirb, a Directory Bookmark program written in Go.  Last altered", LastAltered)
		fmt.Println()
		execname, _ := os.Executable()
		ExecFI, _ := os.Stat(execname)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Println(" HomeDir is", HomeDir, ", ", ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
		fmt.Println(" bookmark file is", fullbookmarkfilename)
		fmt.Println()
		for idx, valu := range bookmark {
			fmt.Printf(" bookmark[%s] = %s \n", idx, valu)
		}
		fmt.Println()
		os.Exit(0)
	}

	target = bookmark[os.Args[1]]
	io.WriteString(os.Stdout, target)
}
