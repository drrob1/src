package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

/*
13 Jan 2023 -- On leox.  I had to change the name from cli which collided w/ another pgm here.  This is now called args.go.
      dot gives no error, isdir=true.  ds-docs gives no error, isdir=false, issymlink=true.
      Split on "." gives dir of empty string and filename is "."
      And when I call handle.Readdirnames(0) it works.

      Split on ds-docs gives dir of empty string and filename is ds-docs.  But if I use os.Open anyway, then isDir is true, isSymlink is false and isRegFile is false, which is correct.
      And when I call handle.Readdirnames(0) it works.

      ds-docs/ gives no error, isdir=true, issymlink=false even though it is a symlinked directory.  And isRegFile is false, which is correct.
      Split on ds-docs/ gives dir of ds-docs/ and filename is empty string
      And when I call handle.Readdirnames(0) it works.
      All the dsm symlinks give the same results.

      When I call Lstat on a non-existant file or try to os.Open() it, the errors.Is(err, os.ErrNotExist) is true.

      These test results show me that I've been doing it wrong in dsrt and derivative code.  I shouldn't be using Lstat or Stat, I should open it.
       Since ds-docs without the '/' or '\' characters does have isSymlink true, perhaps I can use this to my benefit.
*/

func main() {
	flag.Parse()
	fmt.Printf(" There were %d args on the command line.\n", flag.NArg())
	args := flag.Args()
	if len(args) < 15 {
		fmt.Printf(" The args were %#v\n", args)
	} else {
		args = args[:15]
		fmt.Printf(" The first 15 args were: %#v\n", args)
	}

	// now I want to parse using the filepath that splits off dir from filename.  That'll have to be tonight.
	for _, f := range args {
		dir, fileName := filepath.Split(f)
		fmt.Printf("\n Called filepath.Split(%s), and dir= %s, and filename= %s\n", f, dir, fileName)
		fi, err := os.Lstat(f)
		if err == nil {
			regFile := fi.Mode().IsRegular()
			typ := fi.Mode().String()
			fmt.Printf(" Error for %s is nil, name= %5s, isdir= %t, is symlink=%t, size= %5d, regfile= %t, type=%s   \n",
				f, fi.Name(), fi.IsDir(), isSymlink(fi.Mode()), fi.Size(), regFile, typ)
		} else {
			fmt.Printf(" err for %s is %q so os.FileInfo returned a nil value.\n", f, err)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf(" Error matches os.ErrNotExist\n")
			}
		}
	}

	fmt.Printf("\n\n\n")

	for _, f := range args {
		handle, err := os.Open(f)
		if err == nil {
			fmt.Printf(" After os.Open(%s) and err is nil, name=%q \n", f, handle.Name())
			stat, er := handle.Stat()
			if er == nil {
				fmt.Printf(" After handle.Stat and err is nil: name=%s, isdir=%t, issymlink=%t, size=%d, isRegfile=%t \n",
					stat.Name(), stat.IsDir(), isSymlink(stat.Mode()), stat.Size(), stat.Mode().IsRegular())
			} else {
				fmt.Printf(" Error is %s after handle.Stat\n", er)
				handle.Close()
				continue
			}
			dirnames, e := handle.Readdirnames(0)
			if e == nil {
				fmt.Printf(" After handle.Readdirnames.  Length of dirnames is %d\n", len(dirnames))
			}
			handle.Close()
		} else {
			fmt.Printf(" trying to run os.Open(%s) resulted in error of %s\n", f, err)
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf(" Error matches os.ErrNotExist\n")
			}
			handle.Close()
			continue
		}
	}
}

// ------------------------------ IsSymlink ---------------------------

func isSymlink(m os.FileMode) bool {
	intermed := m & os.ModeSymlink
	result := intermed != 0
	return result
} // IsSymlink
