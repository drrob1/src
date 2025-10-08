package main // fstat.go
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/pflag"
)

/*
  17 Sep 25 -- Started writing this.  It will show the results of a file stat.  If a symlink, it will show the target.
                 I found that lstat() reports the target as a symlink if needed, stat() does not.
                 Lstat() reports a symlink to not be a regular file, but Stat() does report it as a regular file.
                 Hardlinks do not show up as symlinks.
                 I am going to rewrite this to only use lstat() to detect symlinks.
                 Using open(name) and then f.Stat() behaves the same as Stat().
                 I'm not going to test to see what a DirEntry does.
  21 Sep 25 -- Added display of UnixNano time stamps, file sizes, and now UnixMicro() and Unix().
   7 Oct 25 -- I'm adding a date format output that is human-readable.
*/

const lastAltered = "7 Oct 25"
const timeFormat = "2006-01-02 15:04:05"

// IsSymlink -- returns true if the file is a symlink.
func IsSymlink(m os.FileInfo) bool {
	intermed := m.Mode() & os.ModeSymlink
	return intermed != 0
} // IsSymlink

func main() {
	pflag.Usage = func() {
		fmt.Printf("Usage: %s file\n", os.Args[0])
	}
	pflag.Parse()

	if pflag.NArg() == 0 {
		pflag.Usage()
		os.Exit(1)
	}

	fmt.Printf("fstat last altered %s, compiled with: %s\n", lastAltered, runtime.Version())

	name := pflag.Arg(0)
	fi, err := os.Stat(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Exiting. \n", name, err)
		os.Exit(1)
	}

	fullname, err := filepath.Abs(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from filepath.Abs(%s) is %s.  Exiting. \n", name, err)
		os.Exit(1)
	}

	symfi, err := os.Lstat(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Lstat(%s) is %s.  Exiting. \n", name, err)
		os.Exit(1)
	}

	dirname := filepath.Dir(name)

	symFlag := IsSymlink(symfi)

	linkName, err := os.Readlink(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Readlink(%s) is %s.\n", name, err) // this is an error for a hardlink or regular file.  Showing it anyway.
	}

	fmt.Printf("name: %s, fullname: %s, dir: %q,\n link name: %s\n   symFlag: %t, isDir: %t, isRegularFile: %t, modebits: %b, size: %d\n",
		name, fullname, dirname, linkName, symFlag, symfi.IsDir(), symfi.Mode().IsRegular(), symfi.Mode(), symfi.Size())

	fmt.Printf("Using Lstat: NanoTime: %d, Size: %d, MicroTime: %d, UnixSec: %d, TimeStamp: %s\n",
		symfi.ModTime().UnixNano(), symfi.Size(), symfi.ModTime().UnixMicro(), symfi.ModTime().Unix(), symfi.ModTime().Format(timeFormat))
	fmt.Printf(" Using Stat: NanoTime: %d, Size: %d, MicroTime: %d, UnixSec: %d, TimeStamp: %s\n",
		fi.ModTime().UnixNano(), fi.Size(), fi.ModTime().UnixMicro(), fi.ModTime().Unix(), fi.ModTime().Format(timeFormat))

	if symFlag {
		fmt.Printf("Target of symlink: %q\n", linkName)
		fmt.Printf("Using Stat(%s):  fullname: %s, dir: %s\n   isDir: %t, isRegularFile: %t, modebits: %b, size: %d\n",
			name, fullname, dirname, fi.IsDir(), fi.Mode().IsRegular(), fi.Mode(), fi.Size())
	}

	f, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Open(%s) is %s.  Exiting. \n", name, err)
		os.Exit(1)
	}
	fi, err = f.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from f.Stat(%s) is %s.  Exiting. \n", name, err)
		os.Exit(1)
	}
	fullname, err = filepath.Abs(fi.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from filepath.Abs(%s) is %s.  Exiting. \n", fi.Name(), err)
		os.Exit(1)
	}
	fmt.Printf("Using open.Stat(%s):  fullname: %s, dir: %s\n   isDir: %t, isRegularFile: %t, modebits: %b, size: %d\n",
		fi.Name(), fullname, dirname, fi.IsDir(), fi.Mode().IsRegular(), fi.Mode(), fi.Size())

	//deSlice, err := os.ReadDir(name) // returns a slice of DirEntry's.  But it needs the name of a directory, not a file.  This won't work for a file.  Nevermind.
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, " Error from os.ReadDir(%s) is %s.  Exiting. \n", name, err)
	//	os.Exit(1)
	//}
	//de := deSlice[0]
	//fullname, err = filepath.Abs(de.Name())
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, " Error from filepath.Abs(%s) is %s.  Exiting. \n", fi.Name(), err)
	//	os.Exit(1)
	//}
	//dirName := filepath.Dir(name)
	//info, err := de.Info()
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, " Error from de.Info() is %s.  Exiting. \n", err)
	//}
	//fmt.Printf("Using os.ReadDir(%s):  fullname: %s, dir: %s\n   isDir: %t, isRegularFile: %t, modebits: %b, size: %d\n",
	//	name, fullname, dirName, de.IsDir(), info.Mode().IsRegular(), info.Mode(), info.Size())

}
