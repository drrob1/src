package main // upgradelint.go

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"src/misc"
	"strconv"
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
  11 Aug 25 -- Added sleep to give lint time to exit.
  16 Aug 25 -- Added flags to use the other websites as backup, which have to get passed into upgradelint.go from lint.go.
  18 Aug 25 -- Added code to reset the file pointer after computing sha1 and before computing sha256.  This fixes the bug in computing sha256.
  19 Aug 25 -- Changed the test for whether lint.exe is not older than lint.info to use the timestamp of the lint.exe file.  This is more robust as it branches the correct way on equality.
  22 Aug 25 -- Added -1 and -2 as shortcuts for -u 1 and -u 2.
				And now I learned that I have to erase the old lint.exe in the configDir before downlading the new one.
*/

const lastAltered = "22 Aug 2025"
const urlRwsNet = "http://drrws.net/"               // from 1and1, which is now ionos.
const urlRobSolomonName = "http://robsolomon.name/" // hostgator
const urlRwsCom = "http://drrws.com/"               // from SimpleNetHosting
const lintExe = "lint.exe"
const lintInfo = "lint.info"

var verboseFlag = flag.BoolP("verbose", "v", false, "verbose flag")
var whichURL int

func main() {
	flag.IntVarP(&whichURL, "url", "u", 0, "which URL to use.  0 is 1and1, 2 is hostgator, 3 is SimpleNetHosting")
	u1 := flag.BoolP("u1", "1", false, "Shortcut for -u 1")
	u2 := flag.BoolP("u2", "2", false, "Shortcut for -u 2")
	flag.Parse()
	if *u1 {
		whichURL = 1
	} else if *u2 {
		whichURL = 2
	}

	var fullLintInfoName, fullRemoteLintExeName string
	switch whichURL {
	case 1:
		fullLintInfoName = urlRobSolomonName + lintInfo
		fullRemoteLintExeName = urlRobSolomonName + lintExe
	case 2:
		fullLintInfoName = urlRwsCom + lintInfo
		fullRemoteLintExeName = urlRwsCom + lintExe
	default:
		fullLintInfoName = urlRwsNet + lintInfo
		fullRemoteLintExeName = urlRwsNet + lintExe
	}

	if *verboseFlag {
		fmt.Printf(" %s to test downloading lint.info and upgrading lint.exe if appropriate.  Last altered %s\n", os.Args[0], lastAltered)
		fmt.Printf(" fullLintInfoName is %s, fullLintExeName is %s\n\n", fullLintInfoName, fullRemoteLintExeName)
	}

	time.Sleep(time.Second * 1) // to give lint time to exit after calling upgradelint.go.

	_, err := os.Stat(lintInfo) // before I added this, the code seems to not download lint.info.  It's best if lint.info is not there.
	if err == nil {
		if *verboseFlag {
			fmt.Printf(" lint.info exists.  Got to delete it.\n")
		}
		err = os.Remove(lintInfo)
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Error returned from os.Remove(%s): %q.  \n", lintInfo, err)
			os.Exit(1)
		}
	}

	resp, err := grab.Get(".", fullLintInfoName)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from grab.Get(%s): %q.  \n", fullLintInfoName, err)
		os.Exit(1)
	}
	if *verboseFlag {
		fmt.Printf(" resp.Filename is %s\n\n", resp.Filename)
	}

	t0, sha1StrReadIn, sha256StrReadIn, err := readInfoFile(resp.Filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from readInfoFile(%s): %q.  \n", resp.Filename, err)
		os.Exit(1)
	}
	infoTimeStamp := t0.Format("Jan-02-2006_15:04:05")

	if *verboseFlag {
		fmt.Printf(" t0 read from %s is %s, sha1StrReadIn is %s, sha256StrReadIn is %s\n\n", lintInfo, infoTimeStamp, sha1StrReadIn, sha256StrReadIn)
	}

	execFI, err := os.Stat(lintExe)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.Stat(%s): %q.  \n", lintExe, err)
		goto downloadMe // bad, bad boy
	}

	if !execFI.ModTime().Before(t0) { // take care to branch the correct way on equality.
		if *verboseFlag {
			fmt.Printf(" lint.exe is newer than lint.info value.  Nothing to do.  I'm going home.\n")
			execTimeStamp := execFI.ModTime().Format("Jan-02-2006_15:04:05")
			fmt.Printf(" lint.exe timestamp is %s, lint.info timestamp is %s\n", execTimeStamp, infoTimeStamp)
		}
		fmt.Printf(" Hit <enter> \n\n")
		os.Exit(0)
	}

downloadMe:
	// Need to download the latest version of lint.exe and check its hashes.
	configDir, err := os.UserConfigDir() // os.TempDir returns a directory that is not guaranteed to exist or have write access.
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.UserConfigDir(): %q.  Exiting.\n", err)
		os.Exit(1)
	}

	joinedName := filepath.Join(configDir, lintExe)
	_, err = os.Stat(joinedName)
	if err == nil { // it exists, so delete it.
		err = os.Remove(joinedName)
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Error returned from os.Remove(%s): %q.  Aborting the upgrade.\n", joinedName, err)
		}
		return
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

	sha1HashComputed := sha1.New()
	sha256HashComputed := sha256.New()
	_, err = io.Copy(sha1HashComputed, lintExeFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from io.Copy(%s): %q.  \n", resp.Filename, err)
		return
	}
	sha1StrComputed := hex.EncodeToString(sha1HashComputed.Sum(nil))

	_, err = lintExeFile.Seek(0, 0) // reset the file pointer to the beginning of the file.
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from lintExeFile.Seek(0, 0): %q.  Call Rob Solomon. \n", err)
		return
	}
	_, err = io.Copy(sha256HashComputed, lintExeFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from io.Copy(%s): %q.  \n", resp.Filename, err)
		return
	}
	lintExeFile.Close() // can't rename a file that is open.
	sha256StrComputed := hex.EncodeToString(sha256HashComputed.Sum(nil))
	if *verboseFlag {
		fmt.Printf(" sha1StrComputed is %s, sha256StrComputed is %s\n\n", sha1StrComputed, sha256StrComputed)
	}

	if sha1StrComputed != sha1StrReadIn || sha256StrComputed != sha256StrReadIn {
		ctfmt.Printf(ct.Red, true, " Error: \n sha1StrReadIn is  %s \nsha1StrComputed is %s \n sha256StrReadIn is  %s \nsha256StrComputed is %s\n",
			sha1StrReadIn, sha1StrComputed, sha256StrReadIn, sha256StrComputed)
		ctfmt.Printf(ct.Red, true, " lint.exe not upgraded. \n")
		fmt.Printf(" Hit <enter> \n\n")
		return
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
	fmt.Printf(" lint.exe upgraded to the most recent version dated %s. \n\n Hit <enter>\n\n", infoTimeStamp)
}

func readInfoFile(fn string) (time.Time, string, string, error) {
	inputBytes, err := os.ReadFile(fn)
	if err != nil {
		return time.Time{}, "", "", err
	}

	buf := bytes.NewReader(inputBytes)

	microStr, err := misc.ReadLine(buf)
	if err != nil {
		return time.Time{}, "", "", err
	}
	microsecs, err := strconv.Atoi(microStr)
	if err != nil {
		return time.Time{}, "", "", err
	}
	timeStamp := time.UnixMicro(int64(microsecs))

	sha1Str, err := misc.ReadLine(buf)
	if err != nil {
		return timeStamp, "", "", err
	}

	sha256Str, err := misc.ReadLine(buf)
	if err != nil {
		return timeStamp, "", "", err
	}

	return timeStamp, sha1Str, sha256Str, err
}
