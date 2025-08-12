package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	flag "github.com/spf13/pflag"
)

/*
  11 Aug 25 -- I got the idea for this program from scrolling thru the os package.
  12 Aug 25 -- I added help.  And I discovered that there is a linux command called hardlink, that's clashing w/ this name.  I'll use hlink.
*/

const lastAltered = "12 Aug 2025"

func main() {
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf(" Usage: hardlink source target, also called hardlink something somewhere, or hardlink old new\n")
		//flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("Usage: hardlink source target, also called hardlink something somewhere, or hardlink old new\n")
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], lastAltered, runtime.Version())
		os.Exit(1)
	}

	something := flag.Arg(0)
	somewhere := flag.Arg(1)

	_, err := os.Stat(something)
	if err != nil {
		fmt.Printf(" Error returned from os.Stat(%s): %q.  Exiting.\n", something, err)
		os.Exit(1)
	}

	fi, err := os.Stat(somewhere)
	if err == nil {
		fmt.Printf(" %s exists, IsDir=%t and isRegular=%t. Should I continue? (y/N) ", somewhere, fi.IsDir(), fi.Mode().IsRegular())
		var answer string
		_, er := fmt.Scanln(&answer)
		if er != nil {
			fmt.Printf(" Ok.  Bye\n")
			os.Exit(1)
		}
		answer = strings.ToLower(answer)
		if !strings.Contains(answer, "y") {
			fmt.Printf(" Ok.  Bye\n")
			os.Exit(1)
		}
	}

	err = os.Link(something, somewhere)
	if err != nil {
		fmt.Printf(" Error returned from os.Link(%s, %s): %q.  \n", something, somewhere, err)
	} else {
		fmt.Printf(" Hardlink created from %s to %s\n", something, somewhere)
	}
}
