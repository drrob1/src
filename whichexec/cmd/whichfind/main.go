package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"runtime"
	"src/whichexec"
)

func main() {
	fmt.Printf(" %s last altered %s, compiled with %s\n", os.Args[0], whichexec.LastAltered, runtime.Version())
	flag.BoolVar(&whichexec.VerboseFlag, "v", false, "Verbose output flag")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Please provide an argument!")
		return
	}
	file := flag.Arg(0)

	if whichexec.VerboseFlag {
		fmt.Printf("file=%s\n", file)
	}
	execBin := whichexec.Find(file)
	if execBin == "" {
		ctfmt.Printf(ct.Red, false, "%s is not found!", file)
	} else {
		ctfmt.Printf(ct.Green, false, "%s is found!", file)
	}
}
