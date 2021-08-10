package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"syscall"
	"unsafe"
)

//  Aug 9, 2021
//  I ran the code below that displays int(Stdin, Stdout and Stderr) on linux and Windows 10.  I got the expected result for linux of 0, 1, and 2.
//  However, the results on Windows 10 were very surprising: Stdin=80, Stdout=84 and Stderr=88.  No wonder why IsTerminal(0) only worked on linux!
//

type terminalSize struct {
	Row, Col, Xpixel, Ypixel uint16
}

func main() {

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

	fmt.Println(" int(Stdout) =", int(os.Stdout.Fd()), ", int(Stdin) =", int(os.Stdin.Fd()), "and int(Stderr) =", int(os.Stderr.Fd()))

//	size, _ := ts.GetSize()
//	fmt.Println(" ts.GetSize says", size.Rows(), "rows and", size.Cols(), "columns.")

//	var sz tsize.Size
//	s, err = tsize.GetSize()
//	if err == nil {
//		fmt.Println(" tsize.GetSize says", s.Height," rows and", s.Width, "columns.")
//	}








}
