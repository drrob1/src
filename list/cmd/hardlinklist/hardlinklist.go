package main // hardlinklist

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/list"
	"strings"

	flag "github.com/spf13/pflag"
)

/*
28 July 25 -- Now that heic.go works to convert a single heic -> jpg, I'll write this as a list converter
29 July 25 -- Now uses os.Create, instead of os.OpenFile.
31 Aug 25  -- Now called linklist.go based on heiclist code, and copylist.   It will make symlinks in the destination directory.
----------------------------------------------------------------------------------------------------
 2 Sep 25 -- Now called hardlinklist.go, and will make hard links in the destination directory.
*/

const lastAltered = "2 Sep 2025"

func main() {
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), lastAltered)

	flag.Usage = func() {
		fmt.Printf(" %s last altered %s, and compiled with %s. \n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf(" Usage information: %s src-dir[glob pattern] dest-dir\n", os.Args[0])
		fmt.Printf(" Makes hardlinks in the destination directory instead of actually copying the files.  This does not need concurrency.\n")
		flag.PrintDefaults()
	}

	var revFlag bool
	flag.BoolVarP(&revFlag, "reverse", "r", false, "Reverse the sort, ie, oldest or smallest is first")

	var sizeFlag bool
	flag.BoolVarP(&sizeFlag, "size", "s", false, "sort by size instead of by date")

	var verboseFlag, veryVerboseFlag bool

	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode, which is same as test mode.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "Very verbose debugging option.")

	var excludeRegex *regexp.Regexp
	var excludeRegexPattern string
	flag.StringVarP(&excludeRegexPattern, "exclude", "x", "", "regex to be excluded from output.")

	var filterFlag, noFilterFlag bool
	var filterStr string
	flag.StringVar(&filterStr, "filter", "", "individual size filter value below which listing is suppressed.")
	flag.BoolVar(&filterFlag, "f", false, "filter value to suppress listing individual size below 1 MB.")
	flag.BoolVar(&noFilterFlag, "F", false, "Flag to undo an environment var with f set.")

	var quality int
	flag.IntVarP(&quality, "quality", "q", 100, "quality of the jpg file")

	flag.Parse()

	if veryVerboseFlag { // setting veryVerboseFlag also sets verbose flag, ie, verboseFlag
		verboseFlag = true
	}

	if verboseFlag {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", ExecFI.Name(), ExecTimeStamp, execName)
		fmt.Println()
	}

	if len(excludeRegexPattern) > 0 {
		if verboseFlag {
			fmt.Printf(" excludeRegexPattern found and is %d runes. \n", len(excludeRegexPattern))
		}
		excludeRegexPattern = strings.ToLower(excludeRegexPattern)
		excludeRegex, err := regexp.Compile(excludeRegexPattern)
		if err != nil {
			fmt.Println(err)
			fmt.Println(" ignoring exclude regular expression.")
		}
		fmt.Printf(" excludeRegexPattern = %q, excludeRegex.String = %q\n", excludeRegexPattern, excludeRegex.String())
	}

	list.FilterFlag = filterFlag
	list.VerboseFlag = verboseFlag
	list.VeryVerboseFlag = veryVerboseFlag
	list.ReverseFlag = revFlag
	list.SizeFlag = sizeFlag
	list.ExcludeRex = excludeRegex
	list.DelListFlag = true

	if flag.NArg() != 2 {
		fmt.Printf(" Need two arguments, source and destination directories.  %d found.  Exiting.\n", flag.NArg())
		os.Exit(1)
	}
	srcDir := flag.Arg(0)
	if srcDir == "." {
		workingDir, err := os.Getwd()
		if err != nil {
			fmt.Printf(" Error from os.Getwd is %s\n", err)
			os.Exit(1)
		}
		srcDir = filepath.Join(workingDir, "*")
	}
	destDir := flag.Arg(1)
	if verboseFlag {
		fmt.Printf(" srcDir = %q, destDir = %q\n", srcDir, destDir)
	}
	destFI, err := os.Stat(destDir)
	if err != nil {
		fmt.Printf(" os.Stat(%s) error is: %s.  Exiting.\n", destDir, err)
		os.Exit(1)
	}
	if !destFI.IsDir() {
		fmt.Printf(" os.Stat(%s) must be a directory as the destination, but it's not.  Exiting.\n", destDir)
		os.Exit(1)
	}

	fileList, err := list.NewFromGlob(srcDir)

	if err != nil {
		fmt.Printf(" Error from list.NewFromGlob is %s\n", err)
		fmt.Printf(" flag.NArg = %d, len(os.Args) = %d\n", flag.NArg(), len(os.Args))
		os.Exit(1)
	}

	if len(fileList) == 0 {
		fmt.Printf(" Length of the fileList is zero.  Aborting \n")
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf(" len(fileList) = %d\n", len(fileList))
	}

	if veryVerboseFlag {
		for i, f := range fileList {
			fmt.Printf(" first fileList[%d] = %#v\n", i, f)
		}
		fmt.Println()
	}

	fileList, err = list.FileSelection(fileList)
	if err != nil {
		fmt.Printf(" Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n\n")

	// now have the fileList.

	fmt.Printf(" There are %d files in the file list.\n\n", len(fileList))

	for _, f := range fileList {
		if f.FI.IsDir() { // skip directory names
			fmt.Printf(" Skipping directory %s\n", f.RelPath)
			continue
		}
		fullDestPath := filepath.Join(destDir, f.FI.Name())
		err = os.Link(f.FullPath, fullDestPath)
		if err != nil {
			fmt.Printf(" Error from os.Symlink(%s,%s) is %q.  Continuing to next item.\n", f.FullPath, fullDestPath, err)
		}
	}
	fmt.Println()
}
