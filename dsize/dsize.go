// dsize.go -- sort in reverse size order.  IE, biggest is first.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	//
	"getcommandline"
)

const lastCompiled = "21 Apr 17"

/*
Revision History
----------------
20 Apr 17 -- Started writing this rtn, based on dirlist.go
21 Apr 17 -- Now tweaking the output format.
*/

// FIS is a FileInfo slice, as in os.FileInfo
type FISlice []os.FileInfo

func (f FISlice) Less(i, j int) bool {
	return f[i].Size() > f[j].Size() // I want a reverse sort
}

func (f FISlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FISlice) Len() int {
	return len(f)
}

func main() {
	var files FISlice
	var err error

	CleanDirName := ""

	if len(os.Args) > 1 {
		commandline := getcommandline.GetCommandLineString()
		CleanDirName = filepath.Clean(commandline)
	} else {
		CleanDirName = "." + string(filepath.Separator)
	}
	fmt.Println(" dsize will display a directory by size.  Written in Go.  lastCompiled ", lastCompiled)
	fmt.Println()
	fmt.Println(" Dirname is", CleanDirName)

	files, err = ioutil.ReadDir(CleanDirName)
	if err != nil {
		log.Fatal(err)
	}

	sort.Sort(files)

	for _, f := range files {
		fmt.Printf("%10v %11d %s %s\n", f.Mode(), f.Size(), f.ModTime().String(), f.Name())
		//		fmt.Println(f.Mode(), "  ", f.Size(), "  ", f.ModTime().String(), "  ", f.Name())
	}

}

/*
package path
func Match

func Match(pattern, name string) (matched bool, err error)

Match reports whether name matches the shell file name pattern.  The pattern syntax is:

pattern:
	{ term }
term:
	'*'         matches any sequence of non-/ characters
	'?'         matches any single non-/ character
	'[' [ '^' ] { character-range } ']'
	            character class (must be non-empty)
	c           matches character c (c != '*', '?', '\\', '[')
	'\\' c      matches character c

character-range:
	c           matches character c (c != '\\', '-', ']')
	'\\' c      matches character c
	lo '-' hi   matches character c for lo <= c <= hi

Match requires pattern to match all of name, not just a substring.  The only possible returned error is ErrBadPattern, when pattern is malformed.


package os
type FileInfo

type FileInfo interface {
        Name() string       // base name of the file
        Size() int64        // length in bytes for regular files; system-dependent for others
        Mode() FileMode     // file mode bits
        ModTime() time.Time // modification time
        IsDir() bool        // abbreviation for Mode().IsDir()
        Sys() interface{}   // underlying data source (can return nil)
}

A FileInfo describes a file and is returned by Stat and Lstat.

func Lstat

func Lstat(name string) (FileInfo, error)

Lstat returns a FileInfo describing the named file.  If the file is a symbolic link, the returned FileInfo describes the symbolic link.  Lstat makes no attempt to follow the link.
If there is an error, it will be of type *PathError.

func Stat

func Stat(name string) (FileInfo, error)

Stat returns a FileInfo describing the named file.  If there is an error, it will be of type *PathError.


The insight I had with my append troubles that the 1 slice entries were empty, is that when I used append, it would do just that to the end of the slice, and ignore the empty slices.
I needed to make the slice as empty for this to work.  So I am directly assigning the DirEntries slice, and appending the FileNames slice, to make sure that these both are doing what I want.
This code is now doing exactly what I want.  I guess there is no substitute for playing with myself.  Wait, that didn't come out right.  Or did it.

*/
