package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
)

/*
  19 Apr 25 -- Got the idea to check this based on code in Chapter 14 of Mastering Go, 4th ed.
				The buffer for both reading and writing is 4K.
*/

const lastModified = "19 Apr 2025"

func main() {
	fmt.Printf(" Finding what the default buffer sizes are for the bufio package.  Last modified %s.  Compiled with %s\n", lastModified, runtime.Version())
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(" reader buffer size from os.Stdin is %d\n", reader.Size())
	writer := bufio.NewWriter(os.Stdout)
	fmt.Printf(" writer buffer size from os.Stdout is %d\n", writer.Size())
	winFile, err := os.Open("MyWin.txt")
	if err != nil {
		fmt.Printf("Error opening MyWin.txt: %v\n", err)
		return
	}
	defer winFile.Close()
	winrdr := bufio.NewReader(winFile)
	fmt.Printf(" reader buffer size from opening MyWin is %d\n", winrdr.Size())

	f, err := os.CreateTemp(".", "bufiodefaults")
	if err != nil {
		fmt.Printf("Error creating TempFile: %v\n", err)
		return
	}
	fmt.Printf(" TempFile created TempFile: %s\n", f.Name())
	tmpwriter := bufio.NewWriter(f)
	fmt.Printf(" Writer buffer size from TempFile is %d\n", tmpwriter.Size())
	tempname := f.Name()
	f.Close()
	err = os.Remove(tempname)
	if err != nil {
		fmt.Printf("Error removing TempFile: %v\n", err)
	} else {
		fmt.Printf(" TempFile removed TempFile: %s\n", tempname)
	}

}
