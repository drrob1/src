package list

import (
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
   6 Jan 2023 -- Improving error handling.  Routines now return an error.
  14 Jan 2023 -- I completely rewrote the section of getFileInfosFromCommandLine where there is only 1 identifier on the command line.  This was based on what I learned
                   from args.go.  Let's see if it works.  Basically, I relied too much on os.Lstat or os.Stat.  Now I'm relying on os.Open.
   1 Feb 2023 -- Fixing how command line arguments are opened when there are > 1 on the line, ie, a source dir and destination dir.
*/

// getFileInfoXFromCommandLine will return a slice of FileInfoExType after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfoExType will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

func getFileInfoXFromCommandLine(excludeMe *regexp.Regexp) ([]FileInfoExType, error) {
	var fileInfoX []FileInfoExType

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %#v\n", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 || flag.Arg(0) == "." { // the "." is the sentinel to be ignored.
		if VerboseFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfoX, err = MyReadDir(workingDir, excludeMe) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if err != nil {
			return nil, err
		}
		if VerboseFlag {
			fmt.Printf(" after call to Myreaddir.  Len(fileInfoX)=%d\n", len(fileInfoX))
		}

	} else if flag.NArg() == 1 { // a lone name may either mean file not found or it's a directory which could be a symlink.
		const sep = string(filepath.Separator)
		fileInfoX = make([]FileInfoExType, 0, 1)
		loneFilename := flag.Arg(0)

		fHandle, err := os.Open(loneFilename) // just try to open it, as it may be a symlink.
		if err == nil {
			stat, _ := fHandle.Stat()
			if stat.IsDir() { // either a direct or symlinked directory name
				fHandle.Close()
				fileInfoX, err = MyReadDir(loneFilename, nil) // nil exclude regex
				return fileInfoX, err
			}

		} else { // err must not be nil after attempting to open loneFilename.
			loneFilename = workingDir + sep + loneFilename
			loneFilename = filepath.Clean(loneFilename)
		}

		fHandle, err = os.Open(loneFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Println()
			os.Exit(1)
		}

		fi, _ := fHandle.Stat()

		if fi.IsDir() {
			fHandle.Close()
			fileInfoX, err = MyReadDir(loneFilename, nil) // ExcludeMe regex is nil
			if err != nil {
				return nil, err
			}
			return fileInfoX, nil
		} else { // loneFilename is not a directory, but opening it did not return an error.  So just return a variable of fileInfoExType fields.
			fix := FileInfoExType{
				FI:      fi,
				Dir:     workingDir,
				RelPath: filepath.Join(workingDir, loneFilename),
			}
			fileInfoX = append(fileInfoX, fix)
			return fileInfoX, nil
		}

	} else { // must have source and destination directories on command line.  Will only process the first param and hope for the best.
		fileInfoX = make([]FileInfoExType, 0, flag.NArg())
		f := flag.Arg(0)
		fHandle, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		stat, _ := fHandle.Stat()
		if VerboseFlag {
			fmt.Printf(" in loop: fHandle.Name=%s, IsDir=%t\n", fHandle.Name(), stat.IsDir())
		}
		fHandle.Close()
		fileInfoX, err = MyReadDir(f, nil)
		return fileInfoX, nil
	}
	if VerboseFlag {
		fmt.Printf(" Leaving getFileInfoXFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfoX))
	}
	return fileInfoX, nil
}
