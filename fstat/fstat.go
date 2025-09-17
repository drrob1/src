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
*/

const lastAltered = "17 Sep 25"

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

	symFlag := IsSymlink(fi)

	fmt.Printf("name: %s, fullname: %s\n   symFlag: %t, isDir: %t, isRegularFile: %t, modebits: %b, size: %d\n",
		name, fullname, symFlag, fi.IsDir(), fi.Mode().IsRegular(), fi.Mode(), fi.Size())

	symfi, err := os.Lstat(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Lstat(%s) is %s.  Exiting. \n", name, err)
		os.Exit(1)
	}
	symflag2 := IsSymlink(symfi)
	fmt.Printf("Using Lstat(%s):  fullname: %s\n   symFlag2: %t, isDir: %t, isRegularFile: %t, modebits: %b, size: %d\n",
		name, fullname, symflag2, symfi.IsDir(), symfi.Mode().IsRegular(), symfi.Mode(), symfi.Size())

	if symflag2 {
		link, err := os.Readlink(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from os.Readlink(%s) is %s.  Exiting. \n", name, err)
			os.Exit(1)
		}
		fmt.Printf("Target of symlink: %q\n", link)

	}

}
