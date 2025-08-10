package main

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strconv"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/spf13/pflag"
)

/*
  10 Aug 25 -- Creates the info file for the lint program.  First line has to be the timestamp, and the 2nd line the sha1,
               and the 3rd line the sha256.
*/

const lintExe = "lint.exe"
const lintInfo = "lint.info" // this is a binary file

func main() {
	var sha1hash, sha256hash hash.Hash
	var verboseFlag bool
	pflag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose flag")
	pflag.Parse()

	fi, err := os.Stat(lintExe)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error returned from os.Stat(%s): %q.  \n", lintExe, err)
		os.Exit(1)
	}
	if fi.IsDir() {
		ctfmt.Printf(ct.Red, true, " Error: %s is a directory.  Exiting.\n\n ", lintExe)
		os.Exit(1)
	}
	t0 := fi.ModTime() // this is the timestamp of the lint.exe file.

	// Create Hash Section
	targetFile, err := os.Open(lintExe)
	if os.IsNotExist(err) {
		ctfmt.Println(ct.Red, true, lintExe, " does not exist.  Skipping.")
		return
	} else if err != nil { // we know that the file exists
		ctfmt.Println(ct.Red, true, " Error opening ", lintExe, ".  Exiting.")
		os.Exit(1)
	}

	sha1hash = sha1.New()
	sha256hash = sha256.New()
	_, err = io.Copy(sha1hash, targetFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error copying %s is %s.  Exiting.\n\n ", lintExe, err)
		os.Exit(1)
	}

	_, err = io.Copy(sha256hash, targetFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error copying %s is %s.  Exiting.\n\n ", lintExe, err)
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf(" sha1hash is %x\n", sha1hash.Sum(nil))
		fmt.Printf(" sha256hash is %x\n", sha256hash.Sum(nil))
	}

	// write data section as strings.  Binary i/o didn't work.
	f, err := os.Create(lintInfo)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error creating %s is %s.  Exiting.\n\n ", lintInfo, err)
		os.Exit(1)
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()

	micro := strconv.Itoa(int(t0.UnixMicro()))
	buf.WriteString(micro)
	_, err = buf.WriteString("\n")
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error writing timestamp to binary file is %s.  Exiting.\n\n ", err)
		return
	}

	s1 := hex.EncodeToString(sha1hash.Sum(nil))
	_, err = buf.WriteString(s1)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error writing sha1 hash to output file is %s.  Exiting.\n\n ", err)
		return
	}
	buf.WriteString("\n")

	s256 := hex.EncodeToString(sha256hash.Sum(nil))
	_, err = buf.WriteString(s256)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error writing sha256 hash to output file is %s.  Exiting.\n\n ", err)
		return
	}
	buf.WriteString("\n")

}
