// dsrt.go -- directoy sort in reverse date order.  IE, newest is first.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

const lastCompiled = "20 May 17"

/*
Revision History
----------------
20 Apr 17 -- Started writing dsize rtn, based on dirlist.go
21 Apr 17 -- Now tweaking the output format.  And used flag package.  One as a pointer and one as a value, just to learn them.
22 Apr 17 -- Coded the use of the first non flag commandline param,  which is all I need.  Note that the flag must appear before the non-flag param, else the flag is ignored.
22 Apr 17 -- Now writing dsrt, to function similarly to dsort.
24 Apr 17 -- Now adding file matching, like "dir" or "ls" does.
25 Apr 17 -- Now adding sort by size as an option, like -s, and commas
26 Apr 17 -- Noticed that the match routine is case sensitive.  I don't like that.
27 Apr 17 -- commandline now allows a file spec.  I intend this for Windows.  I'll see how it goes.
19 May 19 -- Will now show the uid:gid for linux.
*/

// FIS is a FileInfo slice, as in os.FileInfo
type FISlice []os.FileInfo
type FISliceDate []os.FileInfo
type FISliceSize []os.FileInfo

func (f FISliceDate) Less(i, j int) bool {
	return f[i].ModTime().UnixNano() > f[j].ModTime().UnixNano() // I want a reverse sort, newest first
}

func (f FISliceDate) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FISliceDate) Len() int {
	return len(f)
}

func (f FISliceSize) Less(i, j int) bool {
	return f[i].Size() > f[j].Size() // I want a reverse sort, largest first
}

func (f FISliceSize) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f FISliceSize) Len() int {
	return len(f)
}

func main() {
	const numlines = 50
	var userptr *user.User
	var files FISlice
	var filesDate FISliceDate
	var filesSize FISliceSize
	var err error
	var count int

	var revflag = flag.Bool("r", false, "reverse the sort, ie, oldest or smallest is first") // Ptr

	var RevFlag bool
	flag.BoolVar(&RevFlag, "R", false, "Reverse the sort, ie, oldest or smallest is first") // Value

	var nlines = flag.Int("n", numlines, "number of lines to display") // Ptr

	var NLines int
	flag.IntVar(&NLines, "N", numlines, "number of lines to display") // Value

	var helpflag = flag.Bool("h", false, "print help message") // pointer
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "H", false, "print help message")

	var sizeflag = flag.Bool("s", false, "sort by size instead of by date") // pointer
	var SizeFlag bool
	flag.BoolVar(&SizeFlag, "S", false, "sort by size instead of by date")
	flag.Parse()

	fmt.Println(" dsrt will display a directory by date or size.  Written in Go.  LastCompiled ", lastCompiled)
	fmt.Println()

	if *helpflag || HelpFlag {
		flag.PrintDefaults()
	}

	Reverse := *revflag || RevFlag
	SizeSort := *sizeflag || SizeFlag

	NumLines := numlines
	if *nlines != numlines {
		NumLines = *nlines
	} else if NLines != numlines {
		NumLines = NLines
	}
	askforinput := true

	CleanDirName := "." + string(filepath.Separator)
	CleanFileName := ""
	commandline := flag.Arg(0) // this only gets the first non flag argument.  That's all I want
	if len(commandline) > 0 {
		//		CleanDirName = filepath.Clean(commandline)
		CleanDirName, CleanFileName = filepath.Split(commandline)
		CleanFileName = strings.ToUpper(CleanFileName)
		askforinput = false
	}

	if askforinput {
		// Asking for input so don't have to worry about command line globbing
		fmt.Print(" Enter input for globbing: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		newtext := scanner.Text()
		if err = scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, " reading std input: ", err)
			os.Exit(1)
		}
		if len(newtext) > 0 {
			// time to do the stuff I'm writing this pgm for
			CleanDirName, CleanFileName = filepath.Split(newtext)
			CleanFileName = strings.ToUpper(CleanFileName)
		}

	}

	if len(CleanDirName) == 0 {
		CleanDirName = "." + string(filepath.Separator)
	}

	if len(CleanFileName) == 0 {
		CleanFileName = "*"
	}

	if SizeSort {
		filesSize, err = ioutil.ReadDir(CleanDirName)
		if err != nil {
			log.Fatal(err)
		}
		if Reverse {
			sort.Sort(sort.Reverse(filesSize))
		} else {
			sort.Sort(filesSize)
		}
		files = FISlice(filesSize)
	} else {
		filesDate, err = ioutil.ReadDir(CleanDirName)
		if err != nil {
			log.Fatal(err)
		}
		if Reverse {
			sort.Sort(sort.Reverse(filesDate))
		} else {
			sort.Sort(filesDate)
		}
		files = FISlice(filesDate)
	}

	fmt.Println(" Dirname is", CleanDirName)

	uid := os.Getuid() // int
	gid := os.Getgid() // int
	systemStr := ""
	linuxflag := runtime.GOOS == "linux"
	if linuxflag {
		systemStr = "Linux"
	} else if runtime.GOOS == "windows" {
		systemStr = "Windows"
	} else {
		systemStr = "Mac, maybe"
	}

	userptr, err = user.Current()
	if err != nil {
		fmt.Println(" user.Current error is ", err, "Exiting.")
		os.Exit(1)
	}

	fmt.Printf("uid = %d, gid = %d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
		uid, gid, systemStr, userptr.Uid, userptr.Gid, userptr.Username, userptr.Name, userptr.HomeDir)
	for _, f := range files {
		NAME := strings.ToUpper(f.Name())
		if BOOL, _ := filepath.Match(CleanFileName, NAME); BOOL {
			sysUID := int(f.Sys().(*syscall.Stat_t).Uid) // Stat_t is a uint32
			uidStr := strconv.Itoa(sysUID)
			sysGID := int(f.Sys().(*syscall.Stat_t).Gid) // Stat_t is a uint32
			gidStr := strconv.Itoa(sysGID)

			usernameStr := GetIDname(uidStr)
			groupnameStr := GetIDname(gidStr)

			s := f.ModTime().Format("Jan-02-2006 15:04:05")
			sizeint := int(f.Size())
			sizestr := strconv.Itoa(sizeint)
			if sizeint > 100000 {
				sizestr = AddCommas(sizestr)
			}
			//	old way:		fmt.Printf("%10v %11d %s %s\n", f.Mode(), f.Size(), s, f.Name())
			fmt.Printf("%10v %s:%s %15s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
			count++
			if count > NumLines {
				break
			}
		}
	}

} // end main dsrt

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//---------------------------------------------------------------------------------------------------

// ---------------------------- GetIDname -----------------------------------------------------------
func GetIDname(uidStr string) string {

	ptrToUser, err := user.LookupId(uidStr)
	if err != nil {
		panic("uid not found")
	}

	idname := ptrToUser.Username
	return idname

} // GetIDname

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
