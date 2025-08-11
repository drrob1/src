package main // symlink.go

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	flag "github.com/spf13/pflag"
)

const lastAltered = "11 Aug 2025"

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("Usage: symlink source target, also called symlink something somewhere, or symlink old new")
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

	err = os.Symlink(something, somewhere)
	if err != nil {
		fmt.Printf(" Error returned from os.Symlink(%s, %s): %q.  \n", something, somewhere, err)
	} else {
		fmt.Printf(" Symlink created from %s to %s\n", something, somewhere)
	}

}
