package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/jonhadfield/findexec"
	"os"
	"runtime"
	"src/whichexec"
)

/*
REVISION HISTORY
-------- -------
29 Apr 24 -- Added option to search more directories.  IE, more directories option is appended to the system path for the search.
             This is different from findExec, which only searches the system path if no search path is provided.
             And findExec cares which slash is used; my code doesn't.

*/

// var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
var vlcPath = "C:/Program Files/VideoLAN/VLC"
var onWin = runtime.GOOS == "windows"

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
	execBin := whichexec.Find(file, vlcPath)
	oldWayExec := findexec.Find(file, vlcPath)
	if execBin == "" {
		ctfmt.Printf(ct.Red, false, "%s is not found using whichexec\n", file)
	} else {
		ctfmt.Printf(ct.Green, false, "%s is found using whichexec to be at %s\n", file, execBin)
	}
	if oldWayExec == "" {
		ctfmt.Printf(ct.Red, false, "%s is not found using findexec\n", file)
	} else {
		ctfmt.Printf(ct.Green, false, "%s is found using findexec to be at %s\n", file, oldWayExec)
	}
}
