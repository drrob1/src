// pass2.og
package main

/*
  REVISION HISTORY
  21 Apr 18 -- First version, based on vlc.go
  22 Apr 18 -- Added ability to write the passwords to a file.
  26 Apr 18 -- Now called pass2.  Now uses rand routines, init based on OS.
*/

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"
	//  "strings"
	//  "io"
	//  "time"
	//  "path/filepath"
)

const LastAltered = "April 26, 2018"
const passwordfilename = "pass.txt"

// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Errorf("%s : ", msg)
		panic(e)
	}
}

/* ---------------------------- MAIN -------------------------------- */
func main() {
	var n int
	var err error

	fmt.Println(" pass2 program written in Go.  Last altered", LastAltered)

	if runtime.GOOS == "linux" {
		infile, err := os.Open("/dev/random")
		check(err, "after Open /dev/random")
		defer infile.Close()

		byteslice := make([]byte, 8)
		infile.Read(byteslice)
		i64 := int64(byteslice[0])
		rand.Seed(i64)
		infile.Close()
	} else if runtime.GOOS == "windows" {
		t := time.Now()
		nano := t.UnixNano()
		unix := t.Unix()
		nsec := t.Nanosecond()
		fmt.Println("nano=", nano, ", unix=", unix, "nsec=", nsec)
		rand.Seed(nano - unix)
		for i := 0; i < int(unix)-nsec; i++ {
			_ = rand.Int()
		}
	}

	if len(os.Args) <= 1 {
		n = 10
	} else {
		n, err = strconv.Atoi(os.Args[1])
	}

	// how long is the password to be
	check(err, "param not an integer value")
	if n < 0 {
		fmt.Println(" n <= 0 which is out of range.  Exiting")
		os.Exit(1)
	} else if n == 0 {
		n = 10
	} else if n > 30 {
		fmt.Println(" n > 30.  n set to = 30.")
		n = 30
	}

	PasswordFile, err := os.OpenFile(passwordfilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	check(err, "Cannot open the output file.")
	defer PasswordFile.Close()

	PasswordFileWriter := bufio.NewWriter(PasswordFile)
	defer PasswordFileWriter.Flush()

	passwordslice := make([]byte, 0, 50)
	for i := 0; i < n; {
		char := rand.Intn(128)
		b := byte(char)
		if b > 32 && b <= '~' {
			// r := rune(b)
			// s := strconv.QuoteRuneToASCII(r)
			// fmt.Println(" b=", b, ", r=", r, ".  s=", s)
			passwordslice = append(passwordslice, b)
			i++
		}
	}
	password := string(passwordslice)
	fmt.Println(" Password= ", password)
	PasswordFileWriter.WriteString(password)
	_, err = PasswordFileWriter.WriteRune('\n')
	check(err, "bufio write failed")
	fmt.Println()
	//  fmt.Println()
} // END MAIN
