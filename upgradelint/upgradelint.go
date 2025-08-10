package main // upgradelint.go

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
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
------------------------------------------------------------------------------------------------------------------------------------------------------
  10 Aug 25 -- Now called upgradelint.go.
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
	fullRemoteLintExeName := urlRwsNet + lintExe

	if *verboseFlag {
		fmt.Printf(" fullLintInfoName is %s, fullLintExeName is %s\n\n", fullLintInfoName, fullRemoteLintExeName)
	}

	resp, err := grab.Get(".", fullLintInfoName)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from grab.Get(%s): %q.  \n", fullLintInfoName, err)
		os.Exit(1)
	}
	if *verboseFlag {
		fmt.Printf(" resp.Filename is %s\n\n", resp.Filename)
	}

	t0, sha1HashReadIn, sha256HashReadIn, err := readInfoFile(resp.Filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from readInfoFile(%s): %q.  \n", resp.Filename, err)
		os.Exit(1)
	}

	fi, err := os.Stat(lintExe)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.Stat(%s): %q.  \n", lintExe, err)
	}

	if fi.ModTime().After(t0) {
		if *verboseFlag {
			fmt.Printf(" lint.exe is newer than lint.info value.  Nothing to do.\n")
			fmt.Printf(" lint.exe timestamp is %s, lint.info timpstamp contains %s\n", fi.ModTime(), t0)
		}
		os.Exit(0)
	}

	// Need to download the latest version of lint.exe and check its hashes.
	configDir, err := os.UserConfigDir() // os.TempDir returns a directory that is not guaranteed to exist or have write access.
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.UserConfigDir(): %q.  Exiting.\n", err)
		os.Exit(1)
	}

	resp, err = grab.Get(configDir, fullRemoteLintExeName) // can't put the file into the current directory because it will overwrite the current one.
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from grab.Get(%s): %q.  \n", fullRemoteLintExeName, err)
		os.Exit(1)
	}
	if *verboseFlag {
		fmt.Printf(" resp.Filename is %s\n\n", resp.Filename) // I'm assuming that this is a full filename.
	}

	lintExeFile, err := os.Open(resp.Filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.Open(%s): %q.  \n", resp.Filename, err)
		os.Exit(1)
	}
	defer lintExeFile.Close()

	var sha1HashComputed, sha256HashComputed hash.Hash
	sha1HashComputed = sha1.New()
	sha256HashComputed = sha256.New()
	_, err = io.Copy(sha1HashComputed, lintExeFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from io.Copy(%s): %q.  \n", resp.Filename, err)
		return
	}
	_, err = io.Copy(sha256HashComputed, lintExeFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from io.Copy(%s): %q.  \n", resp.Filename, err)
		return
	}

	if sha1HashReadIn != sha1HashComputed || sha256HashReadIn != sha256HashComputed { // I don't yet know how to compare these.
		ctfmt.Printf(ct.Red, true, " Error: sha1HashReadIn is %x \n sha1HashComputed is %x \n sha256HashReadIn is %x \n sha256HashComputed is %x\n",
			sha1HashReadIn, sha1HashComputed, sha256HashReadIn, sha256HashComputed)
		ctfmt.Printf(ct.Red, true, " lint.exe not upgraded. \n")
		os.Exit(1)
	}

	if *verboseFlag {
		fmt.Printf(" lint.exe is to be upgraded. \n")
	}
	currentDir, err := os.Getwd()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.Getwd(): %q.  \n", err)
		os.Exit(1)
	}
	err = os.Rename(lintExe, currentDir+"/"+lintExe+".old")
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.rename(%s, %s): %q.  \n", lintExe, currentDir+"/"+lintExe+".old", err)
	}
	err = os.Rename(resp.Filename, currentDir+"/"+lintExe)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.Rename(%s, %s): %q.  \n", resp.Filename, currentDir+"/"+lintExe, err)
		os.Exit(1)
	}
}

func readInfoFile(fn string) (time.Time, hash.Hash, hash.Hash, error) {
	var t0 time.Time
	var microsecs int64
	var sha1hash hash.Hash
	var sha256hash hash.Hash

	inputBytes, err := os.ReadFile(fn)
	if err != nil {
		return time.Time{}, nil, nil, err
	}

	buf := bytes.NewBuffer(inputBytes)
	err = binary.Read(buf, binary.LittleEndian, &microsecs)
	if err != nil {
		return time.Time{}, nil, nil, err
	}
	t0 = time.UnixMicro(microsecs)
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
