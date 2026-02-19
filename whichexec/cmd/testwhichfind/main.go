package main

import (
	"fmt"
	"os"
	"runtime"
	"src/whichexec"

	"github.com/spf13/pflag"
)

/*
  18 Feb 26 -- Starting to debug whichexec.Find.  It's not returning what I expect if the target is in the working directory.
*/

const lastAltered = "18 Feb 2026"

func main() {
	fmt.Printf(" %s last altered whichexec %s, last altered this pgm %s, compiled with %s\n", os.Args[0], whichexec.LastAltered, lastAltered, runtime.Version())

	var verboseFlag bool
	pflag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose flag")
	pflag.Parse()

	whichexec.VerboseFlag = verboseFlag

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	target := whichexec.Find("upgradelint.exe", workingDir)

	fmt.Printf("target=%s\n", target)
}
