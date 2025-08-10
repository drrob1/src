package main

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

	t0 := time.Now()
	targetFilename := lintInfo

	// Create Hash Section
	targetFile, err := os.Open(targetFilename)
	if os.IsNotExist(err) {
		ctfmt.Println(ct.Red, true, targetFilename, " does not exist.  Skipping.")
		return
	} else if err != nil { // we know that the file exists
		ctfmt.Println(ct.Red, true, " Error opening ", targetFilename, ".  Exiting.")
		os.Exit(1)
	}

	sha1hash = sha1.New()
	sha256hash = sha256.New()
	_, err = io.Copy(sha1hash, targetFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error copying %s is %s.  Exiting.\n\n ", targetFilename, err)
		os.Exit(1)
	}

	_, err = io.Copy(sha256hash, targetFile)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error copying %s is %s.  Exiting.\n\n ", targetFilename, err)
		os.Exit(1)
	}

	if verboseFlag {
		fmt.Printf(" sha1hash is %x\n", sha1hash.Sum(nil))
		fmt.Printf(" sha256hash is %x\n", sha256hash.Sum(nil))
	}

	// write binary data section
	var buf bytes.Buffer

	err = binary.Write(&buf, binary.LittleEndian, t0)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error writing timestamp to binary file is %s.  Exiting.\n\n ", err)
		os.Exit(1)
	}
	err = binary.Write(&buf, binary.LittleEndian, sha1hash.Sum(nil))
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error writing sha1 hash to binary file is %s.  Exiting.\n\n ", err)
		os.Exit(1)
	}
	err = binary.Write(&buf, binary.LittleEndian, sha256hash.Sum(nil))
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error writing sha256 hash to binary file is %s.  Exiting.\n\n ", err)
		os.Exit(1)
	}

	//write to file section

	err = os.WriteFile(targetFilename, buf.Bytes(), os.ModePerm) // os.ModePerm = 0777
	if err != nil {
		ctfmt.Printf(ct.Red, true, " os.WriteFile failed with error %v \n", err)
	}
}
