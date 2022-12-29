package list

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 2022 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                   This is going to take a while.
  20 Dec 2022 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                   I decided to only copy files if the new one is newer than the old one.
  22 Dec 2022 -- Now I want to colorize the output, so I have to return the os.FileInfo also.  So I changed MakeList and NewList to not return []string, but return []FileInfoExType.
                   And myReadDir creates the relPath field that I added to FileInfoExType.
  22 Dec 2022 -- I'm writing and testing listutil_linux.go.  It's too late to test the code, so I'll do that tomorrow.
  29 Dec 2022 -- Adding the '.' to be a sentinel marker for the 1st param that's ignored.  This change is made in the platform specific code.
*/

// getFileInfoXFromCommandLine will return a slice of FileInfoExType after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfoExType will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

func getFileInfoXFromCommandLine(excludeMe *regexp.Regexp) []FileInfoExType {
	var fileInfoX []FileInfoExType

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %#v\n", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 || flag.Arg(0) == "." { // the "." is the sentinel to be ignored.
		if verboseFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfoX = MyReadDir(workingDir, excludeMe) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if verboseFlag {
			fmt.Printf(" after call to Myreaddir.  Len(fileInfoX)=%d\n", len(fileInfoX))
		}

	} else if flag.NArg() == 1 { // a lone name may mean file not found, as bash will populate what it finds.
		var loneFilename string
		const sep = filepath.Separator
		fileInfoX = make([]FileInfoExType, 0, 1)
		firstChar := rune(flag.Arg(0)[0])
		if firstChar == sep { // have an absolute path, so don't prepend anything
			loneFilename = flag.Arg(0)
		} else {
			//loneFilename = workingDir + string(sep) + flag.Arg(0)
			//loneFilename = filepath.Clean(loneFilename)
			loneFilename = filepath.Join(workingDir, flag.Arg(0))
		}
		fi, err := os.Lstat(loneFilename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "%s is a lone filepath and does not exist.  Exiting\n\n", fi.Name())
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, err)
			fmt.Println()
			os.Exit(1)
		}

		if verboseFlag {
			fmt.Printf(" in getFileInfoXFromCommandLine: loneFilename=%s, fi.Name=%s, IsDir=%t\n", loneFilename, fi.Name(), fi.IsDir())
		}

		if fi.IsDir() {
			fileInfoX = MyReadDir(loneFilename, excludeMe)
		} else {
			fix := FileInfoExType{
				FI:      fi,
				Dir:     workingDir,
				RelPath: filepath.Join(workingDir, loneFilename), // Not sure this is needed, but here it is.
			}
			fileInfoX = append(fileInfoX, fix)
		}

	} else { // must have more than one filename on the command line, populated by bash.
		fileInfoX = make([]FileInfoExType, 0, flag.NArg())
		for _, f := range flag.Args() {
			fi, err := os.Lstat(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if verboseFlag {
				fmt.Printf(" in loop: fi.Name=%s, fi.Size=%d, fi.IsDir=%t\n", fi.Name(), fi.Size(), fi.IsDir())
			}
			if includeThis(fi, excludeMe) {
				fix := FileInfoExType{
					FI:      fi,
					Dir:     workingDir,
					RelPath: filepath.Join(workingDir, f),
				}
				fileInfoX = append(fileInfoX, fix)
			}
		}
	}
	if verboseFlag {
		fmt.Printf(" Leaving getFileInfoXFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfoX))
	}
	return fileInfoX
}
