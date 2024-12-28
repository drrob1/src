package list

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"path/filepath"
	"regexp"
)

/*
  REVISION HISTORY
  -------- -------
  18 Dec 22 -- First got idea for this routine.  It will be based on the linux scripts I wrote years ago, makelist, copylist, movelist, runlist and renlist.
                 This is going to take a while.
  20 Dec 22 -- It's working.  But now I'll take out all the crap that came over from dsrtutils.  I'll have to do that tomorrow, as it's too late now.
                 I decided to only copy files if the new one is newer than the old one.
  22 Dec 22 -- Now I want to colorize the output, so I have to return the os.FileInfo also.  So I changed MakeList and NewList to not return []string, but return []FileInfoExType.
                 And myReadDir creates the relPath field that I added to FileInfoExType.
  22 Dec 22 -- I'm writing and testing listutil_linux.go.  It's too late to test the code, so I'll do that tomorrow.
  29 Dec 22 -- Adding the '.' to be a sentinel marker for the 1st param that's ignored.  This change is made in the platform specific code.
   6 Jan 23 -- Improving error handling.  Routines now return an error.
  14 Jan 23 -- I completely rewrote the section of getFileInfosFromCommandLine where there is only 1 identifier on the command line.  This was based on what I learned
                 from args.go.  Let's see if it works.  Basically, I relied too much on os.Lstat or os.Stat.  Now I'm relying on os.Open.
   1 Feb 23 -- Fixing how command line arguments are opened when there are > 1 on the line, ie, a source dir and destination dir.
  24 Mar 23 -- While in Florida I figured out how to handle a glob pattern on the bash command line.  I have to use the length of os.Args or equivalent.
   4 Apr 23 -- Added use of list.DelListFlag
  22 Apr 23 -- Found bug.  I again used flag.NFlag where I meant to use flag.NArg.  I HATE WHEN THAT HAPPENS.
  27 May 23 -- Added getFileInfoXSkipFirstOnCommandLine, for use of runlist.
  28 Dec 24 -- Adding concurrency to the reading of directory entries, like I did in fdsrt and rex.  Nevermind, it's already here.
*/

// getFileInfoXFromCommandLine will return a slice of FileInfoExType after the filter and exclude expression are processed.
// It handles if there are no files populated by bash or file not found by bash, but does not sort the slice before returning it, due to difficulty in passing the sort function.
// The returned slice of FileInfoExType will then be passed to the display rtn to colorize only the needed number of file infos.
// Prior to the refactoring, I first retrieved a slice of all file infos, sorted these, and then only displayed those that met the criteria to be displayed.

const sep = string(filepath.Separator)

func GetFileInfoXFromCommandLine(excludeMe *regexp.Regexp) ([]FileInfoExType, error) {
	var fileInfoX []FileInfoExType

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	ExcludeRex = excludeMe
	if flag.NArg() == 0 || flag.Arg(0) == "." { // the "." is the sentinel to be ignored.
		if VerboseFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		// fileInfoX, err = MyReadDir(workingDir, excludeMe) // excluding by regex, filesize or having an ext is done by MyReadDir.
		fileInfoX, err = myReadDirConcurrent(workingDir) // excluding by regex, filesize or having an ext is done by MyReadDirConcurrent.
		if err != nil {
			return nil, err
		}
		if VerboseFlag {
			fmt.Printf(" after call to Myreaddir.  Len(fileInfoX)=%d\n", len(fileInfoX))
		}

	} else if flag.NArg() == 1 || flag.NArg() == 2 { // First param should be a directory.  If there's a 2nd param, that would also have to be a directory, but that's not handled here.
		fileInfoX = make([]FileInfoExType, 0, 1)
		loneFilename := flag.Arg(0)

		fHandle, err := os.Open(loneFilename) // just try to open it, as it may be a symlink.
		if err == nil {
			stat, _ := fHandle.Stat()
			if stat.IsDir() { // either a direct or symlinked directory name
				fHandle.Close()
				//fileInfoX, err = MyReadDir(loneFilename, nil) // nil exclude regex
				fileInfoX, err = myReadDirConcurrent(loneFilename) // exclude regex passed by the global variable, and is now allowed to not be nil.
				return fileInfoX, err
			}

		} else { // err must not be nil after attempting to open loneFilename.
			loneFilename = workingDir + sep + loneFilename
			loneFilename = filepath.Clean(loneFilename)
			fHandle.Close()
		}

		fHandle, err = os.Open(loneFilename)
		if err != nil {
			return nil, err
		}
		defer fHandle.Close()

		fi, _ := fHandle.Stat()

		if fi.IsDir() {
			//fHandle.Close()
			//fileInfoX, err = MyReadDir(loneFilename, nil)
			fileInfoX, err = myReadDirConcurrent(loneFilename) // the excludeMe regexp is passed globally above.
			if err != nil {
				return nil, err
			}
			return fileInfoX, nil
		} else { // loneFilename is not a directory, but opening it did not return an error.  So just return a variable of fileInfoExType fields.
			joinedFilename := filepath.Join(workingDir, loneFilename)
			fix := FileInfoExType{
				FI:       fi,
				Dir:      workingDir,
				RelPath:  joinedFilename,
				AbsPath:  joinedFilename,
				FullPath: joinedFilename,
			}
			fileInfoX = append(fileInfoX, fix)
			return fileInfoX, nil
		}

	} else { // bash must have populated sources on command line.  Will process all but the last, which would be a destination directory.
		fileInfoX = make([]FileInfoExType, 0, flag.NArg())
		for i := 0; i < flag.NArg()-1; i++ { // don't process the last command line item, as that would be the destination directory.
			fn := flag.Arg(i)
			fHandle, err := os.Open(fn)
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from os.Open(%s) is %s\n", fn, err)
				return nil, err
			}
			stat, _ := fHandle.Stat()
			if VerboseFlag {
				fmt.Printf(" listutil_linux.go command line loop: fn=%s, fHandle.Name=%s, IsDir=%t\n", fn, fHandle.Name(), stat.IsDir())
			}
			fHandle.Close()
			fix := FileInfoExType{
				FI:       stat,
				Dir:      workingDir,
				RelPath:  filepath.Join(workingDir, fn),
				AbsPath:  filepath.Join(workingDir, fn),
				FullPath: filepath.Join(workingDir, fn),
			}
			fileInfoX = append(fileInfoX, fix)
		}
		if DelListFlag { // If this is dellist, don't forget about the last item on the list, which is intentionally not included in the for loop above.
			if VerboseFlag {
				fmt.Printf("In DelListFlag section before processing last item.  len(fileInfoX) = %d\n", len(fileInfoX))
			}
			fn := flag.Arg(flag.NArg() - 1) // last item
			fHandle, err := os.Open(fn)
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from os.Open(%s) is %s\n", fn, err)
				return nil, err
			}
			stat, _ := fHandle.Stat()
			fHandle.Close()
			fix := FileInfoExType{
				FI:       stat,
				Dir:      workingDir,
				RelPath:  filepath.Join(workingDir, fn),
				AbsPath:  filepath.Join(workingDir, fn),
				FullPath: filepath.Join(workingDir, fn),
			}
			fileInfoX = append(fileInfoX, fix)
		}
		if VerboseFlag {
			fmt.Printf("Length of fileInfoX slice after processing last item is %d\n", len(fileInfoX))
		}
		return fileInfoX, nil
	}
	if VerboseFlag {
		fmt.Printf(" Leaving getFileInfoXFromCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n", flag.NArg(), len(flag.Args()), len(fileInfoX))
	}
	return fileInfoX, nil
}

func getFileInfoXSkipFirstOnCommandLine() ([]FileInfoExType, error) {
	var fileInfoX []FileInfoExType

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %#v\n", err)
		os.Exit(1)
	}

	if flag.NArg() < 2 { // First param is the command to be run.  This condition when TRUE there are no files on the command line.
		if VerboseFlag {
			fmt.Printf(" workingDir=%s\n", workingDir)
		}

		fileInfoX, err = MyReadDir(workingDir, ExcludeRex) // excluding by regex, filesize or having an ext is done by MyReadDir.
		if err != nil {
			return nil, err
		}
		if VerboseFlag {
			fmt.Printf(" after call to Myreaddir.  Len(fileInfoX)=%d\n", len(fileInfoX))
		}

	} else if flag.NArg() == 2 { // First param would be the cmd to run.
		fileInfoX = make([]FileInfoExType, 0, 1)
		loneFilename := flag.Arg(1)

		fHandle, err := os.Open(loneFilename) // just try to open it, as it may be a symlink or a directory.
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
			// getting ready for another attempt of opening loneFilename.
		}

		fHandle, err = os.Open(loneFilename)
		if err != nil {
			fmt.Println(err)
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
				FI:       fi,
				Dir:      workingDir,
				RelPath:  filepath.Join(workingDir, loneFilename),
				AbsPath:  filepath.Join(workingDir, loneFilename),
				FullPath: filepath.Join(workingDir, loneFilename),
			}
			fileInfoX = append(fileInfoX, fix)
			return fileInfoX, nil
		}

	} else { // bash must have populated sources on command line.  Will process all but the last, which would be a destination directory.
		fileInfoX = make([]FileInfoExType, 0, flag.NArg())
		for i := 1; i < flag.NArg(); i++ { // don't process the first command line item, as that would be the command to be run.
			fn := flag.Arg(i)
			fHandle, err := os.Open(fn)
			if err != nil {
				ctfmt.Printf(ct.Red, false, " Error from os.Open(%s) is %s\n", fn, err)
				return nil, err
			}
			stat, _ := fHandle.Stat()
			if VerboseFlag {
				fmt.Printf(" listutil_linux.go command line loop: fn=%s, fHandle.Name=%s, IsDir=%t\n", fn, fHandle.Name(), stat.IsDir())
			}
			fHandle.Close()
			fix := FileInfoExType{
				FI:       stat,
				Dir:      workingDir,
				RelPath:  filepath.Join(workingDir, fn),
				AbsPath:  filepath.Join(workingDir, fn),
				FullPath: filepath.Join(workingDir, fn),
			}
			fileInfoX = append(fileInfoX, fix)
		}
		if VerboseFlag {
			fmt.Printf("Length of fileInfoX slice after processing last item is %d\n", len(fileInfoX))
		}
		return fileInfoX, nil
	}
	if VerboseFlag {
		fmt.Printf(" Leaving getFileInfoXSkipFirstOnCommandLine.  flag.Nargs=%d, len(flag.Args)=%d, len(fileinfos)=%d\n",
			flag.NArg(), len(flag.Args()), len(fileInfoX))
	}
	return fileInfoX, nil
}
