package main

import (
	"fmt"
	"os"
	"src/unzipanddel"
)

const lastModified = "14 Nov 2023"

/*
REVISION HISTORY
-------- -------
14 Nov 23 -- Started working on the first version of this pgm.
*/

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" unzip and delete, last modified %s, binary is %s, timstamp of binary is %s\n", lastModified, execName, LastLinkedTimeStamp)

	if len(os.Args) < 2 {
		fmt.Printf(" Need name of zip file.  Exiting.")
		os.Exit(1)
	}
	fn := os.Args[1] + ".zip"
	filenames, err := unzipanddel.UnzipAndDel(fn)
	if err == nil {
		fmt.Printf(" Successfully unzipped and deleted %+v\n", filenames)
	} else {
		fmt.Printf(" Unsuccessfully unzipped or deleted %s with error of %s\n", fn, err)
	}
}
