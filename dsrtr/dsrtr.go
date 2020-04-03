/*
dsrtr.go
  REVISION HISTORY
  ----------------
   1 Apr 20 -- dsrt recursive, named dsrtr.go.
   2 Apr 20 -- Tracking down bug of not finding .pdf files, and probably also not finding .epub or .mobi
                 Turned out to be case sensitivity in the comparisons.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"runtime"
)

const lastAltered = "2 Apr 2020"

func main() {
	//	runtime.GOMAXPROCS(runtime.NumCPU()) // Use all the machine's cores
	log.SetFlags(0)
	var timeoutOpt *int = flag.Int("timeout", 0, "seconds < 240, where 0 means max timeout of 240 sec.")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 240 {
		log.Fatalln("timeout must be in the range [0,240] seconds")
	}
	if *timeoutOpt == 0 {
		*timeoutOpt = 240
	}

	args := flag.Args()

	if len(args) < 1 {
		log.Fatalln("a pattern to match must be specified")
	} else if len(args) == 1 {
		//pattern = strings.ToLower(pattern)
		//fmt.Println(" pattern=", pattern)
	} else {
		// I cannot think of anything to put here at the moment.  I'll say that args must be a slice of strings of filenames, and on linux.
	}

	pattern := strings.ToLower(args[0])

	startDirectory, _ := os.Getwd() // startDirectory is a string
	fmt.Println()
	fmt.Printf(" dsrtr (recursive), written in Go.  Last altered %s, will use pattern of %q and will start in %s. \n", lastAltered, pattern, startDirectory)
	fmt.Println()
	fmt.Println()
	DirAlreadyWalked := make(map[string]bool, 500)
	DirAlreadyWalked[".git"] = true // ignore .git and its subdir's

	t0 := time.Now()
	tfinal := t0.Add(time.Duration(*timeoutOpt) * time.Second)

	// walkfunc closure
	filepathwalkfunction := func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf(" Error from walk is %v. \n ", err)
			return nil
		}

		if fi.IsDir() {
			if DirAlreadyWalked[fpath] {
				return filepath.SkipDir
			} else {
				DirAlreadyWalked[fpath] = true
			}
		} else  /* if fi.Mode().IsRegular()  */ {
			if runtime.GOOS == "linux" {
				for _, fp := range args {
					fp = strings.ToLower(fp)
					NAME := strings.ToLower(fi.Name())
					if BOOL, _ := filepath.Match(fp, NAME); BOOL {
						s := fi.ModTime().Format("Jan-02-2006_15:04:05")
						sizeint := int(fi.Size())
						sizestr := strconv.Itoa(sizeint)
						if sizeint > 100000 {
							sizestr = AddCommas(sizestr)
						}
						usernameStr, groupnameStr := GetUserGroupStr(fi) // util function in platform specific removed Oct 4, 2019 and then unremoved.
						fmt.Printf("%10v %s:%s %15s %s %s\n", fi.Mode(), usernameStr, groupnameStr, sizestr, s, fpath)
					}
				}
			} else if runtime.GOOS == "windows" {
				NAME := strings.ToLower(fi.Name()) // Despite windows not being case sensitive, filepath.Match is case sensitive.  Who new?copy
				if BOOL, _ := filepath.Match(pattern, NAME); BOOL {

					s := fi.ModTime().Format("Jan-02-2006_15:04:05")
					sizeint := int(fi.Size())
					sizestr := strconv.Itoa(sizeint)
					if sizeint > 100000 {
						sizestr = AddCommas(sizestr)
					}
					fmt.Printf("%15s %s %s\n", sizestr, s, fpath)
				}
			}
			now := time.Now()
			if now.After(tfinal) {
				log.Fatalln(" Time up.  Elapsed is", time.Since(t0))
			}
		}
		return nil
	}

	err := filepath.Walk(startDirectory, filepathwalkfunction)

	if err != nil {
		log.Fatalln(" Error from filepath.walk is", err, ".  Elapsed time is", time.Since(t0))
	}

	elapsed := time.Since(t0)
	fmt.Println(" Elapsed time is", elapsed)
	fmt.Println()
} // end main

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
} // InsertIntoByteSlice

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
	if linuxflag {
		for _, f := range files {
			s := f.ModTime().Format("Jan-02-2006_15:04:05")
			sizeint := 0
			sizestr := ""
			usernameStr, groupnameStr := GetUserGroupStr(f) // util function in platform specific removed Oct 4, 2019 and then unremoved.
			if FilenameList && f.Mode().IsRegular() {
				SizeTotal += f.Size()
				sizeint = int(f.Size())
				sizestr = strconv.Itoa(sizeint)
				if sizeint > 100000 {
					sizestr = AddCommas(sizestr)
				}
				fmt.Printf("%10v %s:%s %15s %s %s\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				count++
			} else if IsSymlink(f.Mode()) {
				fmt.Printf("%10v %s:%s %15s %s <%s>\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				count++
			} else if Dirlist && f.IsDir() {
				fmt.Printf("%10v %s:%s %15s %s (%s)\n", f.Mode(), usernameStr, groupnameStr, sizestr, s, f.Name())
				count++
			}
			if count >= NumLines {
				break
			}
		}
	} else if winflag {
		for _, f := range files {
			NAME := strings.ToUpper(f.Name())
			if BOOL, _ := filepath.Match(CleanFileName, NAME); BOOL {
				s := f.ModTime().Format("Jan-02-2006_15:04:05")
				sizeint := 0
				sizestr := ""
				if FilenameList && f.Mode().IsRegular() {
					SizeTotal += f.Size()
					sizeint = int(f.Size())
					sizestr = strconv.Itoa(sizeint)
					if sizeint > 100000 {
						sizestr = AddCommas(sizestr)
					}
					fmt.Printf("%15s %s %s\n", sizestr, s, f.Name())
					count++
				} else if IsSymlink(f.Mode()) {
					fmt.Printf("%15s %s <%s>\n", sizestr, s, f.Name())
					count++
				} else if Dirlist && f.IsDir() {
					fmt.Printf("%15s %s (%s)\n", sizestr, s, f.Name())
					count++
				}
				if count >= NumLines {
					break
				}
			}
		}
	}

}}}
*/
/*
{{{
// ------------------------------ IsSymlink ---------------------------
func IsSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink
}}}
*/
