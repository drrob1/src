// filepicker.go -- directory sort in reverse date order.  IE, newest is first.

package filepicker

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const LastAltered = "2 Oct 20"

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
19 May 17 -- Will now show the uid:gid for linux.
20 May 17 -- Turns out that (*syscall.Stat_t) only compiles on linux.  Time for platform specific code.
21 May 17 -- Cross compiling to GOARCH=386, and the uid and User routines won't work.
 2 Sep 17 -- Added timestamp detection code I first wrote for gastricgo.
18 Oct 17 -- Now called filepicker, derived from dsrt.go.
18 Oct 18 -- Added folding markers
 5 Sep 20 -- Added use of regex
 2 Oct 20 -- Made regex use the case insensitive flag
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

func GetFilenames(pattern string) []string { // Not sure what I want this routine to return yet. []os.FileInfo
	const numlines = 50
	var files FISlice
	var filesDate FISliceDate
	var filesSize FISliceSize
	var err error
	var count int
	/*
		{{{
			var userptr *user.User
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

			fmt.Println(" dsrt will display sorted by date or size.  Written in Go.  LastAltered ", LastAltered)
			execname, _ := os.Executable()
			ExecFI, _ := os.Stat(execname)
			ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
			fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
			fmt.Println()

			uid := 0
			gid := 0
			systemStr := ""
			linuxflag := runtime.GOOS == "linux"
			if linuxflag {
				systemStr = "Linux"
			} else if runtime.GOOS == "windows" {
				systemStr = "Windows"
			} else {
				systemStr = "Mac, maybe"
			}

			if runtime.GOARCH == "amd64" {
				uid = os.Getuid() // int
				gid = os.Getgid() // int
				userptr, err = user.Current()
				if err != nil {
					fmt.Println(" user.Current error is ", err, "Exiting.")
					os.Exit(1)
				}
			}

			if *helpflag || HelpFlag {
			  flag.PrintDefaults()
			  if runtime.GOARCH == "amd64" {
				fmt.Printf("uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
				  uid, gid, systemStr, userptr.Uid, userptr.Gid, userptr.Username, userptr.Name, userptr.HomeDir)
				}

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
		}}}
	*/
	CleanDirName := "." + string(filepath.Separator)
	CleanFileName := ""
	CleanDirName, CleanFileName = filepath.Split(pattern)
	CleanFileName = strings.ToUpper(CleanFileName)
	if len(CleanDirName) == 0 {
		CleanDirName = "." + string(filepath.Separator)
	}

	if len(CleanFileName) == 0 {
		CleanFileName = "*"
	}

	SizeSort := false
	Reverse := false

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

	//	fmt.Println(" Dirname is", CleanDirName)

	stringslice := make([]string, 0)

	for _, f := range files {
		NAME := strings.ToUpper(f.Name())
		if BOOL, _ := filepath.Match(CleanFileName, NAME); BOOL && f.Mode().IsRegular() { // ignore directory names that happen to match the pattern
			stringslice = append(stringslice, f.Name()) // needs to preserve case of filename for linux
			/*
				{{{
				   //			s := f.ModTime().Format("Jan-02-2006 15:04:05")
				   //			sizeint := int(f.Size())
				   //			sizestr := strconv.Itoa(sizeint)
				   //			if sizeint > 100000 {
				   //				sizestr = AddCommas(sizestr)
				   //			}
				   //			usernameStr, groupnameStr := "", ""
				   //			if runtime.GOARCH == "amd64" {
				   //				usernameStr, groupnameStr = GetUserGroupStr(f)
				   //			}
				   //			if linuxflag {
				   //				if f.IsDir() {
				   //					fmt.Printf("%10v %s:%s %15s %s <%s>\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				   //				} else if f.Mode().IsRegular() {
				   //					fmt.Printf("%10v %s:%s %15s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				   //				} else { // it's a symlink
				   //					fmt.Printf("%10v %s:%s %15s %s (%s)\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				   //				}
				   //			} else { // must be windows because this won't compile on Mac.  And I'm ignoring symlinks
				   //				if f.IsDir() {
				   //					fmt.Printf("%15s %s <%s>\n", sizestr, s, f.Name())
				   //				} else {
				   //					fmt.Printf("%15s %s %s\n", sizestr, s, f.Name())
				   //				}
				   //			}
				}}}
			*/
			count++
			if count > numlines {
				break
			}
		}
	}
	return stringslice

} // end GetFilenames

func GetRegexFilenames(pattern string) []string { // Not sure what I want this routine to return yet. []os.FileInfo
	const numlines = 50
	var files FISlice
	var filesDate FISliceDate
	var filesSize FISliceSize
	var err error
	var count int
	/*
		{{{
			var userptr *user.User
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

			fmt.Println(" dsrt will display sorted by date or size.  Written in Go.  LastAltered ", LastAltered)
			execname, _ := os.Executable()
			ExecFI, _ := os.Stat(execname)
			ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
			fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
			fmt.Println()

			uid := 0
			gid := 0
			systemStr := ""
			linuxflag := runtime.GOOS == "linux"
			if linuxflag {
				systemStr = "Linux"
			} else if runtime.GOOS == "windows" {
				systemStr = "Windows"
			} else {
				systemStr = "Mac, maybe"
			}

			if runtime.GOARCH == "amd64" {
				uid = os.Getuid() // int
				gid = os.Getgid() // int
				userptr, err = user.Current()
				if err != nil {
					fmt.Println(" user.Current error is ", err, "Exiting.")
					os.Exit(1)
				}
			}

			if *helpflag || HelpFlag {
			  flag.PrintDefaults()
			  if runtime.GOARCH == "amd64" {
				fmt.Printf("uid=%d, gid=%d, on a computer running %s for %s:%s Username %s, Name %s, HomeDir %s \n",
				  uid, gid, systemStr, userptr.Uid, userptr.Gid, userptr.Username, userptr.Name, userptr.HomeDir)
				}

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
		}}}
	*/
	CleanDirName := "." + string(filepath.Separator)
	CleanPattern := ""
	CleanDirName, CleanPattern = filepath.Split(pattern)
	//CleanPattern = strings.ToUpper(CleanPattern)
	CleanPattern = "(?i)" + CleanPattern // use the case insensitive flag
	if len(CleanDirName) == 0 {
		CleanDirName = "." + string(filepath.Separator)
	}

	if len(CleanPattern) == 0 {
		CleanPattern = "."
	}

	SizeSort := false
	Reverse := false

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

	//	fmt.Println(" Dirname is", CleanDirName)

	stringslice := make([]string, 0)
	regex, err := regexp.Compile(CleanPattern)
	if err != nil {
		log.Fatalln(" Error from regex compile is ", err)
	}

	for _, f := range files {
		//NAME := strings.ToUpper(f.Name())
		NAME := f.Name()                                                   // don't need the ToUpper as I'm using a case insensitive regex flag
		if BOOL := regex.MatchString(NAME); BOOL && f.Mode().IsRegular() { // ignore directory names that happen to match the pattern
			stringslice = append(stringslice, f.Name()) // needs to preserve case of filename for linux
			/*
				{{{
				   //			s := f.ModTime().Format("Jan-02-2006 15:04:05")
				   //			sizeint := int(f.Size())
				   //			sizestr := strconv.Itoa(sizeint)
				   //			if sizeint > 100000 {
				   //				sizestr = AddCommas(sizestr)
				   //			}
				   //			usernameStr, groupnameStr := "", ""
				   //			if runtime.GOARCH == "amd64" {
				   //				usernameStr, groupnameStr = GetUserGroupStr(f)
				   //			}
				   //			if linuxflag {
				   //				if f.IsDir() {
				   //					fmt.Printf("%10v %s:%s %15s %s <%s>\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				   //				} else if f.Mode().IsRegular() {
				   //					fmt.Printf("%10v %s:%s %15s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				   //				} else { // it's a symlink
				   //					fmt.Printf("%10v %s:%s %15s %s (%s)\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				   //				}
				   //			} else { // must be windows because this won't compile on Mac.  And I'm ignoring symlinks
				   //				if f.IsDir() {
				   //					fmt.Printf("%15s %s <%s>\n", sizestr, s, f.Name())
				   //				} else {
				   //					fmt.Printf("%15s %s %s\n", sizestr, s, f.Name())
				   //				}
				   //			}
				}}}
			*/
			count++
			if count > numlines {
				break
			}
		}
	}
	return stringslice

} // end GetRegexFilenames

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

	if len(uidStr) == 0 {
		return ""
	}
	ptrToUser, err := user.LookupId(uidStr)
	if err != nil {
		panic("uid not found")
	}

	idname := ptrToUser.Username
	return idname

} // GetIDname

/*
{{{
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


package os/user
type User struct {
  Uid string
  Gid string
  Username string // login name
  Name string     // full or display name.  It may be blank.
  HomeDir string
}
}}}
*/
