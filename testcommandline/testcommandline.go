package main

import (
	"fmt"
	"runtime"
	"src/getcommandline" // converted to using modules 6/18/21
)

func main() {
	fmt.Println(" Test GetCommandLine routines, compiled with", runtime.Version())
	TheString := getcommandline.GetCommandLineString()
	fmt.Println(" Input commandline is : ", TheString)
	fmt.Printf(" Input command line is : %#v       of type %T\n", TheString, TheString)

	TheByteSlice := getcommandline.GetCommandLineByteSlice()
	fmt.Printf(" Input command line as byteslice : %#v    , of type %T\n", TheByteSlice, TheByteSlice)
}
