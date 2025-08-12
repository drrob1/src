package main // symlink.go

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	flag "github.com/spf13/pflag"
)

/*
  11 Aug 25 -- I got the idea for this program from scrolling thru the os package.
  12 Aug 25 -- I added help.
*/

const lastAltered = "12 Aug 2025"

func main() {
	flag.Usage = func() {
		fmt.Printf(" %s last modified %s, compiled with %s, using pflag.\n", os.Args[0], lastAltered, runtime.Version())
		fmt.Printf(" Usage: symlink source target, also called symlink something somewhere, or symlink old new\n")
		//flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("Usage: symlink source target, also called symlink something somewhere, or symlink old new\n")
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

	err = os.Symlink(something, somewhere)
	if err != nil {
		fmt.Printf(" Error returned from os.Symlink(%s, %s): %q.  \n", something, somewhere, err)
	} else {
		fmt.Printf(" Symlink created from %s to %s\n", something, somewhere)
	}

}
