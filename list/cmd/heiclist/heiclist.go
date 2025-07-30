package main

import (
	"bufio"
	"fmt"
	"github.com/jdeng/goheif"
	flag "github.com/spf13/pflag"
	"image/jpeg"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/list"
	"strings"
)

/*
28 July 25 -- Now that heic.go works to convert a single heic -> jpg, I'll write this as a list converter
29 July 25 -- Now uses os.Create, instead of os.OpenFile.
*/

const lastAltered = "29 July 2025"
const jpgExt = ".jpg"

func writeHeicToJpg(heic, jpg string, quality int) error {
	fi, err := os.Open(heic)
	if err != nil {
		return err
	}

	img, err := goheif.Decode(fi)
	if err != nil {
		return err
	}
	fi.Close()

	//fo, err := os.OpenFile(jpg, os.O_RDWR|os.O_CREATE, 0644)  I don't know why the sample code used OpenFile instead of Create.
	fo, err := os.Create(jpg)
	if err != nil {
		return err
	}
	defer fo.Close()

	w := bufio.NewWriter(fo)
	defer w.Flush()
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	return err
}

func processFilename(fn string, quality int) error {
	baseFilename := filepath.Base(fn)
	ext := filepath.Ext(baseFilename)
	baseFilename = strings.TrimSuffix(baseFilename, ext)
	fmt.Printf(" %s -> %s\n", fn, baseFilename+jpgExt)
	err := writeHeicToJpg(fn, baseFilename+jpgExt, quality)
	return err
}

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

	fileList, err := list.NewFromGlob("*.heic")
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

	for _, f := range fileList {
		err = processFilename(f.FullPath, quality)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from processFilename is %s\n", err)
		}
	}
	fmt.Println()
}
