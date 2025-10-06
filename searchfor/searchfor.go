package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/spf13/pflag"
)

/*
  27 Sep 25 -- At work I've noticed that when the j.mdb file is a hard link, it doesn't always show up in dv list.  I'm exploring different methods of retrieving the directory list
                 and will match the retrived list against the input param file.
                 I'll need os.Getwd(), os.ReadDir() which returns a slice of dirEntry, and after opening a directory,
                 I can use Readdir() returning []FileInfo, Readdirnames() returning []string and ReadDir() returning []DirEntry.
  28 Sep 25 -- Added ability to include directory name in the search.
  29 Sep 25 -- Adding ability to check the concurrent method of getting the directory list.
  30 Sep 25 -- The linear search is finding the target when the binary search is not.  This means I have to write out the entire slice of FileInfo's to a file to debug this.
   1 Oct 25 -- In the concurrent method, I'm removing the done channel as I don't think I need it.  Yep, it still works without it.
				Now I want to add a simpler concurrent method, one that just gets file infos without the dir entry intermediate step.
                Now I'll add timing info.
   2 Oct 25 -- Added use of filepath.Glob
   6 Oct 25 -- So far, all of these routines find the target, which is usually j.mdb.  But in dv, only the non-concurrent code finds it.  This is very puzzling.
                 I'm going to try to resort the files to the newest first like I do in dv, to see if that makes a difference.
*/

const lastAltered = "6 Oct 2025"
const multiplier = 10 // used for the worker pool pattern in MyReadDir
//  const debugName = "debug*.txt"

var fetchAmountofFiles int
var numWorkers = runtime.NumCPU() * multiplier
var verboseFlag bool

func main() {
	pflag.IntVarP(&fetchAmountofFiles, "fetch", "f", 1000, "number of files to fetch")
	pflag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose flag")
	pflag.Parse()
	fmt.Printf(" searchfor.go last altered %s, compiled with %s\n", lastAltered, runtime.Version())
	fmt.Printf(" fetchAmountofFiles: %d, numWorkers: %d\n", fetchAmountofFiles, numWorkers)

	if pflag.NArg() != 1 {
		fmt.Printf(" This pgm searches for the file given as its first parameter to see if it exists, and which os routine can find it.\n")
		fmt.Printf(" Usage: searchfor <file>\n")
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" Error from os.Getwd() is %s\n", err)
		os.Exit(1)
	}

	searchTarget := pflag.Arg(0)

	_, err = os.Stat(searchTarget)
	if err != nil {
		fmt.Printf(" Error from os.Stat(%s) is %s\n", searchTarget, err)
		os.Exit(1)
	}

	fmt.Printf(" Search target exists\n")

	dir, target := filepath.Split(searchTarget)
	if dir == "" {
		dir = workingDir
	}
	fmt.Printf(" Search directory is %s, search target is %s\n", dir, target)

	// os.ReadDir section dealing w/ DirEntry
	t0 := time.Now()
	DirEntries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf(" Error from os.ReadDir(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	fmt.Printf(" os.ReadDir(%s) succeeded, finding %d dir entries, which took %s.\n", dir, len(DirEntries), time.Since(t0))

	lessDirEntries := func(i, j int) bool {
		return DirEntries[i].Name() < DirEntries[j].Name()
	}
	sort.Slice(DirEntries, lessDirEntries)
	position, found := binarySearchDirEntries(DirEntries, target)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}

	// now to sort dir entries by ModTime, newest first.
	lessDirEntries = func(i, j int) bool {
		//return DirEntries[i].Info().ModTime().After(DirEntries[j].ModTime())
		FI1, err := DirEntries[i].Info()
		if err != nil {
			fmt.Printf(" Error from DirEntries[i].Info() is %s\n", err)
			os.Exit(1)
		}
		FI2, err := DirEntries[j].Info()
		if err != nil {
			fmt.Printf(" Error from DirEntries[j].Info() is %s\n", err)
			os.Exit(1)
		}
		return FI1.ModTime().After(FI2.ModTime())
	}
	t0a := time.Now()
	sort.Slice(DirEntries, lessDirEntries) // This now sorts by ModTime, newest first.
	position, found = linearSearchDirEntries(DirEntries, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Linear search of time sorted DirEntries found %s at position %d, taking %s\n\n", searchTarget, position, time.Since(t0a))
	} else {
		ctfmt.Printf(ct.Red, true, " Linear search of time sorted DirEntries did not find %s, taking %s\n\n", searchTarget, time.Since(t0a))
	}

	// Now have to open the directory to explore the other functions
	d, err := os.Open(dir)
	if err != nil {
		fmt.Printf(" Error from os.Open(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" os.Open(%s) succeeded.\n", dir)

	// os.Readdir section dealing w/ FileInfo
	t1 := time.Now()
	FileInfoSlice, err := d.Readdir(-1) // -1 means read all.  Zero would also mean read all.  I guess -1 is clearer.
	if err != nil {
		fmt.Printf(" Error from d.Readdir(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Readdir(-1) succeeded, finding %d FileInfos, which took %s.\n", len(FileInfoSlice), time.Since(t1))
	lessFileInfo := func(i, j int) bool {
		return FileInfoSlice[i].Name() < FileInfoSlice[j].Name()
	}
	sort.Slice(FileInfoSlice, lessFileInfo)
	position, found = binarySearchFileInfos(FileInfoSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n\n", searchTarget)
	}
	d.Close() // have to close it after a successful search.
	fmt.Printf(" 1st d.Close() succeeded.\n\n")

	// os.Readdirnames section.  Have to reopen it
	d, err = os.Open(dir)
	if err != nil {
		fmt.Printf(" Error from 2nd os.Open(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" 2nd os.Open(%s) succeeded.\n", dir)
	t2 := time.Now()
	dirNamesStringSlice, err := d.Readdirnames(-1)
	if err != nil {
		fmt.Printf(" Error from d.Readdirnames(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.Readdirnames(-1) succeeded, finding %d names, which took %s.\n", len(dirNamesStringSlice), time.Since(t2))
	sort.Strings(dirNamesStringSlice)
	position = sort.SearchStrings(dirNamesStringSlice, target)
	if position < len(dirNamesStringSlice) && dirNamesStringSlice[position] == target {
		ctfmt.Printf(ct.Green, true, "Using sort.SearchStrings found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	position, found = binarySearchStrings(dirNamesStringSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Using binarySearchStrings found %s at position %d\n\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n\n", searchTarget)
	}
	err = d.Close()
	if err != nil {
		fmt.Printf(" Error from 2nd d.Close() is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" 2nd d.Close() succeeded.\n\n")

	// os.ReadDir section dealing w/ DirEntry
	d, err = os.Open(dir)
	if err != nil {
		fmt.Printf(" Error from 3rd os.Open(%s) is %s\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()
	fmt.Printf(" 3rd os.Open(%s) succeeded.\n", dir)
	t3 := time.Now()
	DirEntries, err = d.ReadDir(-1)
	if err != nil {
		fmt.Printf(" Error from d.ReadDir(-1) is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" d.ReadDir(-1) succeeded, finding %d dir entries, which took %s.\n", len(DirEntries), time.Since(t3))
	lessDirEntries = func(i, j int) bool { // Need to define this again because I changed it above to sort by time instead of name.
		return DirEntries[i].Name() < DirEntries[j].Name()
	}
	sort.Slice(DirEntries, lessDirEntries)
	position, found = binarySearchDirEntries(DirEntries, target)
	if found {
		ctfmt.Printf(ct.Green, true, " Found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, " Did not find %s\n", searchTarget)
	}
	err = d.Close()
	if err != nil {
		fmt.Printf(" Error from 3rd d.Close() is %s\n", err)
		os.Exit(1)
	}
	fmt.Printf(" 3rd d.Close() succeeded.\n\n")

	// Now to try the original concurrent method of getting the directory list.

	t4 := time.Now()
	fiSlice := myReadDir(dir)
	fmt.Printf(" myReadDir(%s) succeeded, finding %d FileInfo's, which took %s.\n", dir, len(fiSlice), time.Since(t4))
	lessFileInfo = func(i, j int) bool { // Not having a correct less function here was causing the sort.Slice to fail.  I lost a day figuring this out.
		return fiSlice[i].Name() < fiSlice[j].Name()
	}
	sort.Slice(fiSlice, lessFileInfo)

	if verboseFlag {
		fmt.Printf(" Concurrent myReadDir(%s) finished reading %d files, after sort.Slice\n", dir, len(fiSlice))
		for i := 0; i < 20; i++ {
			fmt.Printf("i: %d, FI.Name(): %s\n", i, fiSlice[i].Name())
		}
		fmt.Printf(" fiSlice[%d].Name(): %q\n", len(fiSlice)-1, fiSlice[len(fiSlice)-1].Name())
		fmt.Printf("\n")
	}

	position, found = binarySearchFileInfos(fiSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Concurrent myReadDir FileInfo Binary search found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, "Concurrent myReadDir FileInfo did not find %s\n", searchTarget)
	}

	position, found = linearSearchFileInfos(fiSlice, target)

	if found {
		ctfmt.Printf(ct.Green, true, "Concurrent myReadDir FileInfo Linear search found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, "Concurrent myReadDir FileInfo did not find %s\n", searchTarget)
	}

	t4a := time.Now()
	lessFileInfo = func(i, j int) bool {
		return fiSlice[i].ModTime().After(fiSlice[j].ModTime())
	}
	sort.Slice(fiSlice, lessFileInfo) // This now sorts by ModTime, newest first.
	fmt.Printf("Concurrent myReadDir FileInfo sort.Slice(fiSlice, lessFileInfo) sorting on ModTime, newest first took %s\n", time.Since(t4a))
	position, found = linearSearchFileInfos(fiSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Linear search of time sorted FileInfos found %s at position %d, taking %s\n\n", searchTarget, position, time.Since(t4a))
	} else {
		ctfmt.Printf(ct.Red, true, " Linear search of time sorted FileInfos did not find %s, taking %s\n\n", searchTarget, time.Since(t4a))
	}

	// Now test the simpler version of myReadDir.

	t5 := time.Now()
	fiSimplerSlice := myReadDirSimpler(dir)
	fmt.Printf(" myReadDirSimpler(%s) succeeded, finding %d FileInfo's, which took %s.\n", dir, len(fiSimplerSlice), time.Since(t5))
	lessFileInfo = func(i, j int) bool {
		return fiSimplerSlice[i].Name() < fiSimplerSlice[j].Name()
	}
	sort.Slice(fiSimplerSlice, lessFileInfo)
	position, found = binarySearchFileInfos(fiSimplerSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Simpler: Binary search found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, "Simpler: Did not find %s\n", searchTarget)
	}

	position, found = linearSearchFileInfos(fiSimplerSlice, target)

	if found {
		ctfmt.Printf(ct.Green, true, "Simpler: Linear search found %s at position %d\n\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, "Simpler: Did not find %s\n\n", searchTarget)
	}

	// Now test using filepath.Glob

	t6 := time.Now()
	combinedPattern := filepath.Join(dir, "*")
	GlobStringSlice, err := filepath.Glob(combinedPattern)
	if err != nil {
		fmt.Printf(" Error from filepath.Glob(%s) is %s\n", combinedPattern, err)
		os.Exit(1)
	}
	fmt.Printf(" filepath.Glob(%s) succeeded, finding %d FileInfo's, which took %s.\n", combinedPattern, len(GlobStringSlice), time.Since(t6))
	sort.Strings(GlobStringSlice)
	position, found = binarySearchGlobStrings(GlobStringSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Glob: Binary search found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, "Glob: Binary search did not find %s\n", searchTarget)
	}
	position, found = linearSearchGlobStrings(GlobStringSlice, target)
	if found {
		ctfmt.Printf(ct.Green, true, "Glob: Linear search found %s at position %d\n", searchTarget, position)
	} else {
		ctfmt.Printf(ct.Red, true, "Glob: Linear search did not find %s\n", searchTarget)
	}

	// wrap up
	if verboseFlag {
		nowStr := time.Now().Format("2006-01-02_15_04_05")
		fn := filepath.Join(dir, fmt.Sprintf("searchfor_%s_%s.txt", nowStr, target))
		f, err := os.Create(fn)
		if err != nil {
			fmt.Printf(" Error from os.Create(%s) is %s\n", fn, err)
			os.Exit(1)
		}
		defer f.Close()
		fmt.Printf(" os.Create(%s) succeeded, file is %s\n", fn, f.Name())

		// Write out the fiSlice to the debug file.
		buf := bufio.NewWriter(f)
		buf.WriteString(" --------------------------------- fiSlice  --------------------------------- \n")
		for i, fi := range fiSlice {
			buf.WriteString(fmt.Sprintf("%d: %s    ", i, fi.Name()))
			if i%4 == 3 {
				buf.WriteString("\n")
			}
		}
		buf.WriteString("\n----------------------------------  end of fiSlice  ---------------------------------- \n\n")

		buf.WriteString("----------------------------------  start of FileInfoSlice  ---------------------------------- \n")
		for i, fi := range FileInfoSlice {
			buf.WriteString(fmt.Sprintf("%d: %s    ", i, fi.Name()))
			if i%4 == 3 {
				buf.WriteString("\n")
			}
		}
		buf.WriteString("\n----------------------------------  end of FileInfoSlice  ---------------------------------- \n\n")

		buf.WriteString("----------------------------------  start of Strings  ---------------------------------- \n")
		for i, s := range dirNamesStringSlice {
			buf.WriteString(fmt.Sprintf("%d: %s    ", i, s))
			if i%4 == 3 {
				buf.WriteString("\n")
			}
		}

		buf.WriteString(" --------------------------------- Start of fiSimplerSlice  --------------------------------- \n")
		for i, fi := range fiSimplerSlice {
			buf.WriteString(fmt.Sprintf("%d: %s    ", i, fi.Name()))
			if i%4 == 3 {
				buf.WriteString("\n")
			}
		}
		buf.WriteString("\n----------------------------------  end of fiSimplerSlice  ---------------------------------- \n\n")

		buf.WriteString(" --------------------------------- Start of GlobStringSlice  --------------------------------- \n")
		for i, s := range GlobStringSlice {
			buf.WriteString(fmt.Sprintf("%d: %s    ", i, s))
			if i%4 == 3 {
				buf.WriteString("\n")
			}
		}
		buf.WriteString("\n----------------------------------  end of GlobStringSlice  ---------------------------------- \n\n")

		buf.WriteString("\n\n")
		buf.Flush()

	}

	fmt.Printf("\n")
}

func binarySearchDirEntries(slice []os.DirEntry, target string) (int, bool) {
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
		if slice[current].Name() < target {
			left = current + 1
		} else if slice[current].Name() > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}

func binarySearchFileInfos(slice []os.FileInfo, target string) (int, bool) {
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
		if slice[current].Name() < target {
			left = current + 1
		} else if slice[current].Name() > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}

func binarySearchStrings(slice []string, target string) (int, bool) {
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
		if slice[current] < target {
			left = current + 1
		} else if slice[current] > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}

func binarySearchGlobStrings(slice []string, target string) (int, bool) {
	//var numTries int
	left := 0
	right := len(slice) - 1

	for left <= right {
		current := (left + right) / 2
		//numTries++
		if filepath.Base(slice[current]) < target {
			left = current + 1
		} else if filepath.Base(slice[current]) > target {
			right = current - 1
		} else { // found it
			return current, true
		}
	}
	return -1, false
}

func linearSearchFileInfos(slice []os.FileInfo, target string) (int, bool) {
	for i, fi := range slice {
		if fi.Name() == target {
			return i, true
		}
	}
	return -1, false
}

func linearSearchDirEntries(slice []os.DirEntry, target string) (int, bool) {
	for i, de := range slice {
		if de.Name() == target {
			return i, true
		}
	}
	return -1, false
}

func linearSearchGlobStrings(slice []string, target string) (int, bool) {
	for i, s := range slice {
		if filepath.Base(s) == target {
			return i, true
		}
	}
	return -1, false
}

func myReadDir(dir string) []os.FileInfo {
	// Adding concurrency in returning []os.FileInfo

	var wg sync.WaitGroup

	deChan := make(chan []os.DirEntry, numWorkers) // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fiChan := make(chan os.FileInfo, numWorkers)   // of individual file infos to be collected and returned to the caller of this routine.
	//doneChan := make(chan bool)                    // unbuffered channel to signal when it's time to get the resulting fiSlice and return it.
	fiSlice := make([]os.FileInfo, 0, fetchAmountofFiles*numWorkers)
	wg.Add(numWorkers)

	// reading from deChan to get the slices of DirEntry's
	for range numWorkers {
		go func() {
			defer wg.Done()
			for deSlice := range deChan {
				for _, de := range deSlice {
					fi, err := de.Info()
					if err != nil {
						fmt.Printf("Error getting file info for %s: %v, ignored\n", de.Name(), err)
						continue
					}
					if de.IsDir() {
						continue
					}
					//if !de.Type().IsRegular() {
					//	continue
					//}
					fiChan <- fi // the code in the other routines uses a function here, includeThis, which I don't need here.
				}
			}
		}()
	}

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
		for fi := range fiChan {
			fiSlice = append(fiSlice, fi)
		}
		//close(doneChan)
	}()

	d, err := os.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error os.open(%s) is %s.  exiting.\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()

	for {
		// reading DirEntry's and sending the slices into the channel needs to happen here.
		deSlice, err := d.ReadDir(fetchAmountofFiles) // the docs say that this way of getting a FileInfo does not follow symlinks.
		if errors.Is(err, io.EOF) {                   // finished.  So return the slice.
			close(deChan) // here is where I close the deChan channel.
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR from %s.ReadDir(%d) is %s.\n", dir, numWorkers, err)
			continue
		}
		deChan <- deSlice
		if verboseFlag {
			fmt.Printf(" myReadDir(%s) sent %d DirEntry's to deChan.\n", dir, len(deSlice))
			for i := 0; i < 10; i++ {
				fmt.Printf("deSlice[%d].Name(): %q\n", i, deSlice[i].Name())
			}
			fmt.Printf("\n")
		}
	}
	wg.Wait()     // for the deChan
	close(fiChan) // This way I only close the channel once.  I think if I close the channel from within a worker, and there are multiple workers, closing an already closed channel panics.

	if verboseFlag {
		fmt.Printf(" myReadDir(%s) finished reading %d files.\n", dir, len(fiSlice))
		for i := 0; i < 20; i++ {
			fmt.Printf("i: %d, FI.Name(): %q\n", i, fiSlice[i].Name())
		}
		fmt.Printf(" fiSlice[%d].Name(): %q\n", len(fiSlice)-1, fiSlice[len(fiSlice)-1].Name())
		fmt.Printf("\n")
	}

	return fiSlice
} // myReadDir

func myReadDirSimpler(dir string) []os.FileInfo {
	// Adding concurrency in returning []os.FileInfo, not using dir entries so this should be simpler.

	var wg sync.WaitGroup

	//deChan := make(chan []os.DirEntry, numWorkers) // a channel of a slice to a DirEntry, to be sent from calls to dir.ReadDir(n) returning a slice of n DirEntry's
	fiFetchChan := make(chan []os.FileInfo, numWorkers)
	fiChan := make(chan os.FileInfo, numWorkers) // of individual file infos to be collected and returned to the caller of this routine.
	fiSlice := make([]os.FileInfo, 0, fetchAmountofFiles*numWorkers)
	wg.Add(numWorkers)

	// reading from fiSliceChan to get the slices of FileInfos
	for range numWorkers {
		go func() {
			defer wg.Done()
			for fiSlice := range fiFetchChan {
				for _, fi := range fiSlice {
					if fi.IsDir() {
						continue
					}
					fiChan <- fi
				}
			}
		}()
	}

	// collecting all the individual file infos, putting them into a single slice, to be returned to the caller of this rtn.  How do I know when it's done?
	// I figured it out, by closing the channel after all work is sent to it.
	go func() {
		for fi := range fiChan {
			fiSlice = append(fiSlice, fi)
		}
	}()

	d, err := os.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error os.open(%s) is %s.  exiting.\n", dir, err)
		os.Exit(1)
	}
	defer d.Close()

	for {
		// reading FileInfos and sending the slices into the channel needs to happen here.
		fiSlice, err := d.Readdir(fetchAmountofFiles) // ? the docs say that this way of getting a FileInfo does not follow symlinks.
		if errors.Is(err, io.EOF) {                   // finished.  So return the slice.
			close(fiFetchChan) // here is where I close the deChan channel.
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, " ERROR from %s.Readdir(%d) is %s.\n", dir, numWorkers, err)
			continue
		}
		fiFetchChan <- fiSlice
		if verboseFlag {
			fmt.Printf(" myReadDir(%s) sent %d fileinfos to deChan.\n", dir, len(fiSlice))
			for i := 0; i < 10; i++ {
				fmt.Printf("deSlice[%d].Name(): %q\n", i, fiSlice[i].Name())
			}
			fmt.Printf("\n")
		}
	}
	wg.Wait() // for the fiSliceChan
	close(fiChan)

	if verboseFlag {
		fmt.Printf(" myReadDir(%s) finished reading %d files.\n", dir, len(fiSlice))
		for i := 0; i < 20; i++ {
			fmt.Printf("i: %d, FI.Name(): %q\n", i, fiSlice[i].Name())
		}
		fmt.Printf(" fiSlice[%d].Name(): %q\n", len(fiSlice)-1, fiSlice[len(fiSlice)-1].Name())
		fmt.Printf("\n")
	}

	return fiSlice
} // myReadDirSimpler
