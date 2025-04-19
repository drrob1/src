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
}
