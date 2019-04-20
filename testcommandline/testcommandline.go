package main

import (
	"fmt"
	"getcommandline"
)

func main() {
	TheString := getcommandline.GetCommandLineString()
	fmt.Println(" Input commandline is : ", TheString)
	fmt.Printf(" Input command line is : %#v       of type %T\n", TheString, TheString)

	TheByteSlice := getcommandline.GetCommandLineByteSlice()
	fmt.Printf(" Input command line as byteslice : %#v    , of type %T\n", TheByteSlice, TheByteSlice)
}
