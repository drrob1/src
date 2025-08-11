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
*/

const lastAltered = "11 Aug 2025"

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("Usage: hardlink source target, also called hardlink something somewhere, or hardlink old new")
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

	_, err = os.Stat(somewhere)
	if err == nil {
		fmt.Printf(" %s exists. Should I continue? (y/N) ", somewhere)
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
