package main // adduplist.go

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/list"
	"strings"

	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
)

/*
28 July 25 -- Now that heic.go works to convert a single heic -> jpg, I'll write this as a list converter
29 July 25 -- Now uses os.Create, instead of os.OpenFile.
----------------------------------------------------------------------------------------------------
31 Aug 25  -- Now called adduplist.go, and will add up the size of all the files in the list.  Used heiclist as the base code.
*/

const lastAltered = "31 Aug 2025"

func main() {
	fmt.Printf("%s is compiled w/ %s, last altered %s\n", os.Args[0], runtime.Version(), lastAltered)

	flag.Usage = func() {
		fmt.Printf(" %s last altered %s, and compiled with %s. \n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf(" Usage information: %s [glob pattern]\n", os.Args[0])
		fmt.Printf(" Converts image in heic format to jpg format.\n")
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

	var workingDir, fullPattern string
	var fileList []list.FileInfoExType
	var err error

	workingDir, err = os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from os.Getwd is %s\n", err)
		os.Exit(1)
	}
	if flag.NArg() == 0 {
		fullPattern = filepath.Join(workingDir, "*")
	} else {
		fullPattern = flag.Arg(0)
	}

	fileList, err = list.NewFromGlob(fullPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.NewFromGlob is %s\n", err)
		fmt.Printf(" flag.NArg = %d, len(os.Args) = %d\n", flag.NArg(), len(os.Args))
		fmt.Print(" Continue? [yN] ")
		var ans string
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			fmt.Printf(" No input detected.  Exiting.\n")
			os.Exit(1)
		}
		ans = strings.ToLower(ans)
		if strings.Contains(ans, "n") {
			os.Exit(1)
		}
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
	if len(fileList) == 0 {
		fmt.Printf(" Length of the fileList is zero.  Aborting \n")
		os.Exit(1)
	}

	fileList, err = list.FileSelection(fileList)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from list.FileSelection is %s\n", err)
		os.Exit(1)
	}

	if len(fileList) == 0 {
		fmt.Printf(" The selected list of files is empty.  Exiting.\n")
		os.Exit(1)
	}

	fmt.Printf("\n\n")

	// now have the fileList.

	fmt.Printf(" There are %d files in the file list.\n\n", len(fileList))

	var sumSize int64
	for _, f := range fileList { // I want symlinks to be included in the sum as if they were the files they point to.
		if f.FI.IsDir() { // skip directories
			continue
		}
		amount := f.FI.Size()
		if !f.FI.Mode().IsRegular() { // not a dir name and not a regular file, so must be a symlink.
			fi, err := os.Stat(f.AbsPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Skipping\n", f.FullPath, err)
				continue
			}
			amount = fi.Size()
		}
		sumSize += amount
	}

	fmt.Printf(" The sum of the sizes of the files in the list is %d bytes.\n\n", sumSize)

	magStr, color := list.GetMagnitudeString(sumSize)
	ctfmt.Printf(color, true, " The sum of the sizes of the files in the list is %s.\n\n", magStr)
}
