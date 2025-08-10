package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"os"
	"time"

	"github.com/cavaliergopher/grab/v3"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
)

/*
  9 Aug 25 -- This was originally copied from "Black Hat Go", something about poetry.  I'm going to change it to implement my idea towards lint updating itself.
               First, I have to see if I can get it to list directory contents.
               Turns out I did do this, in digest.go.  It uses a GitHub package called grab.  I'll use that, so I don't have to write my own code to do this.

               So, I need lint.info and lint.sha, and pgms that will create these files that will be read and processed by upgradelint.go.  I'll need to use some code from my sha
               routines like fsha.go.

               Lint.info only needs the current timestamp.  Or it could read lint.exe and use that in this file.  I'll see how it goes as I write it.
               The verbose flag will be needed to show all relevant stuff to debug this.

               I don't yet know if I should print a message to the terminal saying when it's been automatically upgraded.

               A running program can't update itself, so this has to be a separate program that will download the latest version of lint.exe and upgrade it.
*/

const lastAltered = "10 Aug 2025"
const urlRwsNet = "http://drrws.net/"               // from 1and1, which is now ionos.
const urlRobSolomonName = "http://robsolomon.name/" // hostgator
const urlRwsCom = "http://drrws.com"                // from SimpleNetHosting
const lintExe = "lint.exe"
const lintInfo = "lint.info"

var verboseFlag = flag.BoolP("verbose", "v", false, "verbose flag")

func main() {
	fmt.Printf(" %s to test downloading lint.info and upgrading lint.exe if appropriate.  Last altered %s, %s last linked %s\n", os.Args[0], lastAltered)

	fullLintInfoName := urlRwsNet + lintInfo
	fullLintExeName := urlRwsNet + lintExe

	if *verboseFlag {
		fmt.Printf(" fullLintInfoName is %s, fullLintExeName is %s\n\n", fullLintInfoName, fullLintExeName)
	}

	resp, err := grab.Get(".", fullLintInfoName)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from grab.Get(%s): %q.  \n", fullLintInfoName, err)
		os.Exit(1)
	}
	if *verboseFlag {
		fmt.Printf(" resp.Filename is %s\n\n", resp.Filename)
	}

}

func readInfoFile(fn string) (time.Time, hash.Hash, hash.Hash, error) {
	var t0 time.Time
	var sha1hash hash.Hash
	var sha256hash hash.Hash

	inputBytes, err := os.ReadFile(fn)
	if err != nil {
		return time.Time{}, nil, nil, err
	}

	buf := bytes.NewBuffer(inputBytes)
	err = binary.Read(buf, binary.LittleEndian, &t0)
	if err != nil {
		return time.Time{}, nil, nil, err
	}

	err = binary.Read(buf, binary.LittleEndian, sha1hash.Sum(nil))
	if err != nil {
		return time.Time{}, nil, nil, err
	}

	err = binary.Read(buf, binary.LittleEndian, sha256hash.Sum(nil))

	return t0, sha1hash, sha256hash, err
}
