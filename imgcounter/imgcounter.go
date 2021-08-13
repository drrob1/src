/*
REVISION HISTORY
-------- -------
13 Aug 21 -- First version, came out of playing with fyne image viewer


*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const LastModified = "Friday, August 13, 2021"

// ----------------------------------isImage // ----------------------------------------------
func isImage(file string) bool {
	ext := filepath.Ext(file)
	ext = strings.ToLower(ext)
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif"
}

// ------------------------------- MyReadDirForImages -----------------------------------

/*
func MyReadDirForImages(dir string) []os.FileInfo { // works, but only for current working directory
	dirname, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return nil
	}

	fi := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		if isImage(name) {
			imgInfo, err := os.Lstat(name)
			if err != nil {
				fmt.Fprintln(os.Stderr, " Error from os.Lstat ", err)
				continue
			}
			fi = append(fi, imgInfo)
		}
	}

	return fi
} // MyReadDirForImages
*/

func MyReadDirForImages(dir string) int {
	dirname, err := os.Open(dir)
	if err != nil {
		return 0
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return 0
	}

	total := 0
	for _, name := range names {
		if isImage(name) {
			total++
		}
	}
	return total
}

func main() {
	var dir string

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from os.Getwd is", err)
	}

	if len(os.Args) > 1 { // use command line param
		dir = os.Args[1] // 2nd argument, ie, after pgm name
	} else {
		dir = cwd
	}

	totalimageFiles := MyReadDirForImages(dir)
	fmt.Println(" imgCounter written in Go.  Last modified", LastModified, "and compiled with", runtime.Version())

	fmt.Println()
	fmt.Println(" Total number of files contained an image extension is/are:", totalimageFiles)
	fmt.Println()
}
