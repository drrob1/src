package main

import (
	"fmt"
	"os"
	"src/unzipanddel"
	"strings"
)

/*
REVISION HISTORY
-------- -------
14 Nov 23 -- Started working on the first version of this pgm.
*/

const lastModified = "14 Nov 2023"

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" unzip and delete, last modified %s, binary is %s, timstamp of binary is %s\n", lastModified, execName, LastLinkedTimeStamp)

	if len(os.Args) < 2 {
		fmt.Printf(" Name of zip file not found.  Exiting.")
		os.Exit(1)
	}

	fn := os.Args[1]
	lowerFN := strings.ToLower(fn)
	if !strings.HasSuffix(lowerFN, "zip") {
		fn += ".zip"
	}
	filenames, err := unzipanddel.UnzipAndDel(fn)
	if err == nil {
		fmt.Printf(" %s successfully unzipped and deleted: %+v\n", fn, filenames)
	} else {
		fmt.Printf(" Unsuccessfully unzipped or deleted %s with error of %s\n", fn, err)
	}
}
