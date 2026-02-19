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
29 Sep 24 -- Added FindConfig
18 Feb 26 -- Debugging case where upgradelint.exe is in the working directory, and it won't start via exec cmd.
*/

const LastAltered = "18 Feb 2026"

var onWin = runtime.GOOS == "windows"

var VerboseFlag bool

func Find(file, morePath string) string {
	if VerboseFlag {
		fmt.Printf("In Find: Finding file %s\n", file)
	}
	path := os.Getenv("PATH")

	if onWin && !strings.HasSuffix(file, ".exe") {
		file += ".exe"
	}

	if VerboseFlag {
		fmt.Printf("In whichexec.Find.  file = %s, before split list: PATH is %s\n", file, path)
	}

	if morePath != "" {
		path = path + string(filepath.ListSeparator) + morePath
	}
	pathSplit := filepath.SplitList(path)

	if VerboseFlag {
		fmt.Printf("\n In Find after split list: there are %d path strings; PATH is %v\n", len(pathSplit), pathSplit)
	}

	for _, directory := range pathSplit {
		fullPath := filepath.Join(directory, file)

		// Does it exist?
		fileInfo, err := os.Stat(fullPath)
		if err != nil { // file not found in this directory
			continue
		}

		// So it exists.

		if VerboseFlag {
			fmt.Printf("\n In Find loop: Found file %s in directory %q, fullpath %s\n", file, directory, fullPath)
		}

		if !strings.Contains(fullPath, string(filepath.Separator)) { // then it doesn't have the full path.  I don't know why.
			//  fullPath = "./" + fullPath
			fullPath, err = filepath.Abs(fullPath) // I would rather return this than just a dot reference.
			if err != nil {
				panic(err)
			}
		}

		if onWin {
			return fullPath
		}

		// not windows if get here
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

// FindConfig searches current working directory, homedir, homedir/.config/, configdir in that order, and returns first match.
func FindConfig(file string) (string, bool) {
	// build the search path
	path := make([]string, 0, 4)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fullPath := filepath.Join(workingDir, file)
	path = append(path, fullPath)

	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	fullPath = filepath.Join(homedir, file)
	path = append(path, fullPath)

	fullPath = filepath.Join(homedir, ".config", file)
	path = append(path, fullPath)

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	fullPath = filepath.Join(configDir, file)
	path = append(path, fullPath)

	// find it
	for _, dir := range path {
		fileinfo, err := os.Stat(dir)
		if err != nil {
			continue
		}

		mode := fileinfo.Mode()
		if mode.IsRegular() {
			return dir, true
		}
	}
	return "", false
}
