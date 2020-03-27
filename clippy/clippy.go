package main

import (
	"bytes"
	"fmt"
	"strings"

	//	"io"
//	"log"

	//	"io"
//	"log"
	"os"
	"os/exec"
)

func main() {

	buf := []byte(" values written to stdin to go to clipboard ...  ")
	rdr := bytes.NewReader(buf)

	cmdclip := exec.Command("xclip")
	cmdclip.Stdin = rdr
	cmdclip.Stdout = os.Stdout
	cmdclip.Run()

	fmt.Println()
	fmt.Println(" written to stdin thru xclip")
	fmt.Println()



	// Both of these now work on z76.  It just depends on which clip pgm I like better.
/*
	buf := []byte(" string written to stdin to go to xsel ... ")
	rdr := bytes.NewReader(buf)
	cmdsel := exec.Command("xsel", "--clipboard", "-i")
	cmdsel.Stdin = rdr
	cmdsel.Stdout = os.Stdout
	cmdsel.Run()


	fmt.Println()
	fmt.Println(" written to stdin thru xsel")
	fmt.Println()


 */

    fmt.Println(" now to test from clip")
    var w strings.Builder
    cmdfromclip := exec.Command("xclip", "-o")
    cmdfromclip.Stdout = &w
    cmdfromclip.Run()

    fmt.Println(" fromclip is:", w.String())

}
