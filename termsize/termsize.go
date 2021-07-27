package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	//"src/github.com/nathan-fiscaletti/consolesize.go"
//	"github.com/nathan-fiscaletti/consolesize.go"

	//"src/golang.org/x/term"
	"golang.org/x/term"
	//"src/github.com/olekukonko/ts"
//	"github.com/olekukonko/ts"

	// src/github.com/kopoli/go-terminal-size
//	tsize "github.com/kopoli/go-terminal-size"
)

type terminalSize struct {
	Row, Col, Xpixel, Ypixel uint16
}

func main() {
//	rows, cols := consolesize.GetConsoleSize()
//	fmt.Println(" GetConsoleSize says", rows, "rows and", cols, "columns.")

	if term.IsTerminal(int(os.Stdout.Fd())) { // os.Stdout should be fd = 1.
		fmt.Println(" in a terminal according to term.IsTerminal")
	} else {
		fmt.Println(" Not in a terminal according to term.IsTerminal.")
	}
	width, height, err := term.GetSize(int(os.Stdout.Fd())) // this works on linux and Windows
	if err == nil {
		fmt.Println(" term.GetSize says", height, "rows and", width, "columns.")
	} else {
		fmt.Println(" error from golang.org/x/term.GetSize is", err)
	}

	termsize := &terminalSize{}
	retCode, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(termsize)))            // only works on linux

	if int(retCode) < 0 {
		fmt.Fprintln(os.Stderr, "retCode from the syscall operation is", retCode)
	}
	fmt.Println(" after syscall stuff.  Row =", termsize.Row, ", Col =", termsize.Col, ", Xpix =", termsize.Xpixel, ", Ypix =", termsize.Ypixel, ".")
	fmt.Println()


//	size, _ := ts.GetSize()
//	fmt.Println(" ts.GetSize says", size.Rows(), "rows and", size.Cols(), "columns.")

//	var sz tsize.Size
//	s, err = tsize.GetSize()
//	if err == nil {
//		fmt.Println(" tsize.GetSize says", s.Height," rows and", s.Width, "columns.")
//	}








}
