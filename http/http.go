package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

/*
  9 Aug 25 -- This was originally written from "Black Hat Go".  I don't need this.  I'm going to change it to implement my idea towards lint updating itself.
               First, I have to see if I can get it to list directory contents.
               Turns out I did do this, in digest.go.  It uses a GitHub package called grab.  I'll use that, so I don't have to write my own code to do this.

               So, I need lint.info and lint.sha, and pgms that will create these files that will be read and processed by upgradelint.go.  I'll need to use some code from my sha
               routines like fsha.go.

               Lint.info only needs the current timestamp.  Or it could read lint.exe and use that in this file.  I'll see how it goes as I write it.
               The verbose flag will be needed to show all relevant stuff to debug this.

               I don't yet know if I should print a message to the terminal saying when it's been automatically upgraded.
*/

const lastAltered = "9 Aug 25"
const urlRwsNet = "http://drrws.net/"
const urlRobSolomonName = "http://robsolomon.name/"
const urlHostGator = "http://drrws.com"

var verboseFlag = flag.BoolP("verbose", "v", false, "verbose flag")

func main() {
	fmt.Printf(" %s to test downloading lint.info and upgrading lint.exe if appropriate.  Last altered %s, %s last linked %s\n", os.Args[0], lastAltered)

}
