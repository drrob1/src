package whichexec

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

/*
  From Mastring Go, 4th ed, by Mihalis Tsoukalos, published by Packtpub (c) 2024
    Author: Mihalis Tsoukalos
  This is a rewrite of findExec, but without its limitations.
  It will take as its first param the exec binary to search for, and the remaining are directories to search.  If there are none given, it will search PATH by default.
  I'll start w/ the provided code, which searches PATH, and then maybe add the option to search more directories.

REVISION HISTORY
-------- -------
28 Apr 24 -- First started writing this
29 Apr 24 -- Added option to search more directories.  IE, more directories option is appended to the system path for the search.  This is different from findExec.
             And the format of the more string is that it's parsed like a PATH string, so it can contain multiple directories to be searched, separated by the appropriate character for that OS.

*/

const LastAltered = "30 Apr 2024"

var onWin = runtime.GOOS == "windows"

var VerboseFlag bool

func Find(file, morePath string) string {

	if VerboseFlag {
		fmt.Printf("In Find: Finding file %s\n", file)
	}
	path := os.Getenv("PATH")
	pathSplit := filepath.SplitList(path)
	if morePath != "" {
		moreSplit := filepath.SplitList(morePath)
		pathSplit = append(pathSplit, moreSplit...)
	}

	for _, directory := range pathSplit {
		fullPath := filepath.Join(directory, file)
		if runtime.GOOS == "windows" && !strings.HasSuffix(fullPath, ".exe") {
			fullPath += ".exe"
		}

		// Does it exist?
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if onWin {
			//f := strings.ToLower(fullPath)  // don't need to compare against the filename, because Stat was already called on this name.
			//fn := strings.ToLower(file)
			//if strings.HasSuffix(fullPath, "exe") && strings.Contains(f, fn) {
			if strings.HasSuffix(fullPath, "exe") {
				return fullPath
			} else {
				continue
			}
		}
		// now not windows
		mode := fileInfo.Mode()
		// Is it a regular file?
		if !mode.IsRegular() {
			continue
		}

		// Is it executable?
		if mode&0111 != 0 {
			fmt.Println(fullPath)
			return fullPath
		}
	}
	return ""
}
