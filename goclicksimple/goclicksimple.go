package main

import (
	"bufio"
	"fmt"
	w32a "github.com/JamesHovious/w32"
	w32 "github.com/gonutz/w32/v2"
	"os"
	"runtime"
	"strings"
	// ps "github.com/mitchellh/go-ps"
	//"github.com/lxn/win"  I can't get this to be useful.
	//w32 "github.com/gonutz/w32/v2"  I also can't get this to be useful.
	//w32a "github.com/JamesHovious/w32"
)

/*
  HISTORY
  -------
   8 Jun 22 -- Started playing w/ this.  This will take a while, as I have SIR in Boston soon.
  10 Jun 22 -- Seems to be mostly working.  Tomorrow going to Boston.
  17 Jun 22 -- Back from Boston.  It doesn't work very well on linux.  Time again to try on Win10.
  18 Jun 22 -- Will separate the different ways of getting the pid.  They're not compatible.  And will use TARGET environment variable because
                 the command line params appear in the title.
  20 Jun 22 -- Now called goclicksimple, to isolate the w32 and w32a stuff that I can use to post a question.
*/

const lastModified = "June 20, 2022"

func main() {
	fmt.Printf("goclicksimple to use Go to activate a process so can be clicked on the screen.  Last modified %s.  Compiled by %s\n",
		lastModified, runtime.Version())

	fmt.Printf(" Now to use w32.FindWindow\n")

	target := "*firefox*"
	hwnd := w32.FindWindow("MDIClient", target)
	fmt.Printf(" target=%q, MDIClient hwnd=%d\n", target, hwnd)

	hwnd = w32.FindWindow("", target)
	fmt.Printf(" target=%q, empty class hwnd=%d\n", target, hwnd)

	hwnd = w32.FindWindow("*", target)
	fmt.Printf(" target=%q, * hwnd=%d\n", target, hwnd)

	hwnd = w32.FindWindow("*lient*", target) // covers Client and client
	fmt.Printf(" target=%q, *lient* hwnd=%d\n", target, hwnd)

	var classString string
	hwnd2 := w32a.FindWindowS(&classString, &target)
	fmt.Printf(" w32a.FindWindowS empty class, target=%q, hwnd2=%v\n", target, hwnd2)

	classString = "*"
	hwnd2 = w32a.FindWindowS(&classString, &target)
	fmt.Printf(" w32a.FindWindowS '*' class, target=%q, hwnd2=%v\n", target, hwnd2)

	classString = "*lient*"
	hwnd2 = w32a.FindWindowS(&classString, &target)
	fmt.Printf(" w32a.FindWindowS '*lient*' class, target=%q, hwnd2=%v\n", target, hwnd2)

	classString = "*lass*"
	hwnd2 = w32a.FindWindowS(&classString, &target)
	fmt.Printf(" w32a.FindWindowS '*lass*' class, target=%q, hwnd2=%v\n", target, hwnd2)

	pause(0)
	pause0()
}

// --------------------------------------------------------------------------------------------

func pause(n int) bool {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing ", n, ".  Hit <enter> to continue.  Or 'n' to exit  ")
	scanner.Scan()
	if strings.ToLower(scanner.Text()) == "n" {
		return true
	}
	return false
}

func pause0() bool {
	fmt.Print(" Pausing.  Hit <enter> to continue.  Or 'n' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.Contains(ans, "n") {
		return true
	}
	return false
}

/*
My notes from going over robotgo docs.  It uses go doc which extracts the documentation from the code.  Perhaps I can do
that w/ fyne

For this to compile on linux, I had to install libxkbcommon-dev package.  robotgo.GetTitle really doesn't work on linux.

ActiveName(name string) error -- activate window by name

ActivePID(pid int32, args ...int) error -- activate window by PID

FindName(pid int32) (string, error)

GetTitle(args ...int32) string

MaxWindow(pid int32, args ...int)
MinWindow(pid int32, args ...int)

PidExists(pid int32) (bool, error)
Pids() ([]int32, error) -- get all pid's

ReadAll() (string, error) -- read string from clipboard
PasteStr(str string) string -- paste string, write to clipboard, and tap cmd-v

type Nps struct {
  Pid int32
  Name string
}

Process() ([]Nps, error)


go get github.com/atotto/clipboard
ReadAll() (string, error)  -- read from clipboard
WriteAll(text string) error -- write to clipboard


go get github.com/gonutz/w32/v2
func FindWindow(className, windowName string) HWND {
	var class, window uintptr
	if className != "" {
		class = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className)))
	}
	if windowName != "" {
		window = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName)))
	}
	ret, _, _ := findWindow.Call(class, window)
	return HWND(ret)
}










*/
