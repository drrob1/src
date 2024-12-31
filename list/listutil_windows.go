package list

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

/*
   8 Jul 23 -- I changed how the first param is tested for being a directory.
  28 Dec 24 -- Going to add concurrency to the reading of directory entries, like I did in fdsrt and rex.
				Nevermind, it's already here.  I don't use it on Windows because my testing did not find it to be faster.  It is much faster on linux.
*/

const sepStr = string(os.PathSeparator)

// getFileInfoXFromCommandLine will return a slice of FileInfos after the filter and exclude expression are processed, and that match a pattern if given.
// It handles if there are no files populated by bash or file not found by bash, and sorts the slice before returning it.
// The returned slice of FileInfos will then be passed to the display rtn to determine how it will be displayed.

func GetFileInfoXFromCommandLine(excludeMe *regexp.Regexp) ([]FileInfoExType, error) {
	var fileInfoX []FileInfoExType

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = "."
	}
	HomeDirStr = HomeDirStr + sepStr

	pattern := flag.Arg(0) // this only gets the first non flag argument and is all I want on Windows.  And it doesn't panic if there are no arg's.
	if VerboseFlag {
		fmt.Printf(" file pattern is %s\n", pattern)
	}
	if flag.NArg() == 0 || pattern == "." {
		workingDir, er := os.Getwd()
		if er != nil {
			return nil, er
			//fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			//os.Exit(1)
		}
		fileInfoX, err = MyReadDir(workingDir, excludeMe) // this is not concurrent here on Windows.
		if err != nil {
			return nil, err
		}
	} else { // Must have a pattern on the command line, ie, NArg > 0
		if strings.ContainsRune(pattern, ':') {
			pattern = ProcessDirectoryAliases(pattern)
		}

		pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		dirName, fileName := filepath.Split(pattern)
		fileName = strings.ToLower(fileName)
		if dirName != "" && fileName == "" { // then have a dir pattern without a filename pattern
			fileInfoX, err = MyReadDir(dirName, excludeMe)
			return fileInfoX, err
		}
		if dirName == "" {
			dirName = "."
		} else if !strings.HasSuffix(dirName, sepStr) { // I was sometimes getting errors if a pathname didn't end w/ a slash or backslash.
			dirName = dirName + sepStr
		}
		if fileName == "" { // need this to not be blank because of the call to Match below.
			fileName = "*"
		}

		if VerboseFlag {
			fmt.Printf(" dirName=%s, fileName=%s, pattern=%s \n", dirName, fileName, pattern)
		}

		//var filenames []string  not used as I'm having the actual work done in my concurrent code.

		if GlobFlag {
			// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
			// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
			// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
			fileInfoX, err = FileInfoXFromGlob(pattern) // this is concurrent
			return fileInfoX, err

			//   removed Dec 28, 2024, as I'm adding more uses of my concurrent code.
			//filenames, err = filepath.Glob(pattern)
			//if VerboseFlag {
			//	fmt.Printf(" after glob: len(filenames)=%d, filenames=%v \n\n", len(filenames), filenames)
			//}
			//if err != nil {
			//	return nil, err
			//}

		} else {
			fileInfoX, err = myReadDirConcurrentWithMatch(dirName, fileName) // now uses concurrent code, and gets the excludeMe globally.
			return fileInfoX, err

			// Removed Dec 28, 2024 as I'm adding more uses of my concurrent code.
			//d, err := os.Open(dirName)
			//if err != nil {
			//	fmt.Fprintf(os.Stderr, " Error from os.Open is %v\n", err)
			//	return nil, err
			//}
			//defer d.Close()
			//filenames, err = d.Readdirnames(0) // I don't know if I have to make this slice first.  I'm going to assume not for now.
			//if err != nil {                    // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
			//	fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
			//	// fileInfoX, err = MyReadDir(dirName, excludeMe)
			//	fileInfoX, err = myReadDirConcurrent(dirName)  // now uses concurrent code, and gets the excludeMe globally.
			//	return fileInfoX, err
			//}
		}

		//fileInfoX = make([]FileInfoExType, 0, len(filenames))
		//const sepStr = string(os.PathSeparator)
		//for _, f := range filenames { // basically I do this here because of a pattern to be matched.
		//	var path string
		//	if strings.Contains(f, sepStr) {
		//		path = f
		//	} else {
		//		path = filepath.Join(dirName, f)
		//	}
		//
		//	fi, err := os.Lstat(path)
		//	if err != nil {
		//		fmt.Fprintf(os.Stderr, " Error from Lstat call on %s is %v\n", path, err)
		//		continue
		//	}
		//	//if !fi.Mode().IsRegular() { // skip anything that is not a regular file.  Too bad it doesn't work.  It does work in IncludeThis, though
		//	//	continue
		//	//}
		//
		//	match, er := filepath.Match(strings.ToLower(fileName), strings.ToLower(f)) // redundant if glob is used, but I'm ignoring this.
		//	if er != nil {
		//		fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern is %v.\n", pattern, er)
		//		continue
		//	}
		//
		//	if includeThis(fi, excludeMe) && match { // has to match pattern, size criteria and not match an exclude pattern.
		//		fix := FileInfoExType{
		//			FI:       fi,
		//			Dir:      dirName,
		//			RelPath:  filepath.Join(dirName, f),
		//			AbsPath:  filepath.Join(dirName, f),
		//			FullPath: filepath.Join(dirName, f),
		//		}
		//		fileInfoX = append(fileInfoX, fix)
		//	}
		//} // for f ranges over filenames
	} // if flag.NArgs()

	return fileInfoX, nil

} // end getFileInfoXFromCommandLine

func getFileInfoXSkipFirstOnCommandLine() ([]FileInfoExType, error) { // Uses list.ExcludeRex
	var fileInfoX []FileInfoExType
	excludeMe := ExcludeRex

	HomeDirStr, err := os.UserHomeDir() // used for processing ~ symbol meaning home directory.
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprintln(os.Stderr, ".  Ignoring HomeDirStr")
		HomeDirStr = "."
	}
	HomeDirStr = HomeDirStr + sepStr

	pattern := flag.Arg(1) // this gets the 2nd non flag argument and is all I want on Windows.  And it doesn't panic if it's not there.
	if VerboseFlag {
		fmt.Printf(" file pattern is %s\n", pattern)
	}
	if pattern == "" { // this means no pattern was given on the cmd line.
		workingDir, er := os.Getwd()
		if er != nil {
			return nil, er
			//fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine Getwd is %v\n", er)
			//os.Exit(1)
		}
		fileInfoX, err = MyReadDir(workingDir, excludeMe)
		if err != nil {
			return nil, err
		}
	} else { // Must have a pattern on the command line
		if strings.ContainsRune(pattern, ':') {
			pattern = ProcessDirectoryAliases(pattern)
		}

		pattern = strings.Replace(pattern, "~", HomeDirStr, 1)
		dirName, fileName := filepath.Split(pattern)
		fileName = strings.ToLower(fileName)
		if dirName != "" && fileName == "" { // then have a dir pattern without a filename pattern
			fileInfoX, err = MyReadDir(dirName, excludeMe)
			return fileInfoX, err
		}
		if dirName == "" {
			dirName = "."
		}
		if fileName == "" { // need this to not be blank because of the call to Match below.
			fileName = "*"
		}

		if VerboseFlag {
			fmt.Printf(" dirName=%s, fileName=%s \n", dirName, fileName)
		}

		var filenames []string
		if GlobFlag {
			// Glob returns the names of all files matching pattern or nil if there is no matching file. The syntax of patterns is the same as in Match.
			// The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').  Caveat: it's case sensitive.
			// Glob ignores file system errors such as I/O errors reading directories. The only possible returned error is ErrBadPattern, when pattern is malformed.
			filenames, err = filepath.Glob(pattern)
			if VerboseFlag {
				fmt.Printf(" after glob: len(filenames)=%d, filenames=%v \n\n", len(filenames), filenames)
			}
			if err != nil {
				return nil, err
			}

		} else {
			d, err := os.Open(dirName)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from Linux processCommandLine os.Open is %v\n", err)
				os.Exit(1)
			}
			defer d.Close()
			filenames, err = d.Readdirnames(0) // I don't know if I have to make this slice first.  I'm going to assume not for now.
			if err != nil {                    // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
				fmt.Fprintln(os.Stderr, err, "so calling my own MyReadDir.")
				fileInfoX, err = MyReadDir(dirName, excludeMe)
				return fileInfoX, err
			}
		}

		fileInfoX = make([]FileInfoExType, 0, len(filenames))
		for _, f := range filenames { // basically I do this here because of a pattern to be matched.
			var path string
			if strings.Contains(f, sepStr) {
				path = f
			} else {
				path = filepath.Join(dirName, f)
			}

			fi, err := os.Lstat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error from Lstat call on %s is %v\n", path, err)
				continue
			}

			match, er := filepath.Match(strings.ToLower(fileName), strings.ToLower(f)) // redundant if glob is used, but I'm ignoring this.
			if er != nil {
				fmt.Fprintf(os.Stderr, " Error from filepath.Match on %s pattern is %v.\n", pattern, er)
				continue
			}

			if includeThis(fi, excludeMe) && match { // has to match pattern, size criteria and not match an exclude pattern.
				fix := FileInfoExType{
					FI:       fi,
					Dir:      dirName,
					RelPath:  filepath.Join(dirName, f),
					AbsPath:  filepath.Join(dirName, f),
					FullPath: filepath.Join(dirName, f),
				}
				fileInfoX = append(fileInfoX, fix)
			}
		} // for f ranges over filenames
	} // if flag.NArgs()

	return fileInfoX, nil

} // end getFileInfoXSkipFirstOnCommandLine
