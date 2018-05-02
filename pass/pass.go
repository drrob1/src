package main

/*
  REVISION HISTORY
  21 Apr 18 -- First version, based on vlc.go
  22 Apr 18 -- Added ability to write the passwords to a file.
  28 Apr 18 -- Will exit on Windows.
*/

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

const LastAltered = "April 28, 2018"
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

	fmt.Println(" pass program written in Go.  Last altered", LastAltered)

	if runtime.GOOS == "windows" {
		fmt.Println("This program will not run on Windows.  Exiting.")
		os.Exit(0)
	}

	if len(os.Args) <= 1 {
		n = 10
	} else {
		n, err = strconv.Atoi(os.Args[1])
	}
	infile, err := os.Open("/dev/random")
	check(err, "after Open /dev/random")
	defer infile.Close()

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
		byteslice := make([]byte, 1)
		_, err := infile.Read(byteslice)
		check(err, "after infile.Read")
		//    fmt.Println(" byteslice=",byteslice)
		//    u := uint(byteslice[0])
		//    fmt.Println(" u=",u)
		r := rune(byteslice[0])
		if r > 32 && r <= '~' {
			//      s := strconv.QuoteRuneToASCII(r)
			//      fmt.Println(" r=",r, ".  s=",s)
			passwordslice = append(passwordslice, byteslice[0])
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
}
