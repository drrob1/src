package main

import (
	"bufio"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/gonutz/w32/v2"
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
  27 Jun 22 -- The testing version I called w32 now works, based on the help I got from Howard C. Shaw III.  Now I have to fine tune it.
                 I'm going to remove all the PID based stuff as that didn't work anyway.  It's saved as oldgoclick.
                 But this will take a little while.
  28 Jun 22 -- TARGET can use tilde squiggle character to represent a space in the title.  I use strings.ReplaceAll " " to "~" on the titles.
                 Added noFlag to do a trial run of what matches before trying to activate it.
*/

const lastModified = "June 28, 2022"

var verboseFlag, suppressFlag, skipFlag, noFlag bool
var target string

type htext struct {
	h         w32.HWND
	title     string // this is the title of the window as returned by w32.GetWindowText
	isWindow  bool
	isEnabled bool
	isVisible bool
	className string
}

func main() {
	fmt.Printf("goclick to use Go to activate a process so can be clicked on the screen.  Last modified %s.  Compiled by %s\n",
		lastModified, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, " Verbose flag")
	flag.BoolVar(&suppressFlag, "suppress", false, " Suppress output of non-blank titles.  Now default behavior, so this is ignored.")
	flag.BoolVar(&skipFlag, "skip", false, "Skip output of all hwnd's found")
	flag.BoolVar(&noFlag, "no", false, "No activating any windows.  IE, do a trial run.")
	flag.Parse()

	target = os.Getenv("TARGET")
	target = strings.ToLower(target)
	//replaced := strings.NewReplacer("~", " ") // this will allow me to use ~ as a space in the target.
	//replaced.Replace(target)

	if verboseFlag {
		fmt.Printf(" Target is %q after calling Getenv for TARGET\n", target)
	}

	// w32 section

	fmt.Printf("\n w32 section\n")
	if pause0() {
		os.Exit(0)
	}

	fmt.Printf("\n done.\n")

	fmt.Printf(" About to start displaying hwnd retrieved by EnumWindows.\n")
	hwndText := make([]htext, 0, 1000) // magic number I expect will be large enough.

	enumCallBack := func(hwnd w32.HWND) bool { // this fcn is what is called by EnumWindows.  Maybe that shouldn't activate a window.  I'm moving this to its own loop.
		if hwnd == 0 {
			return false
		}

		ht := htext{
			h:         hwnd,
			title:     strings.ToLower(w32.GetWindowText(hwnd)),
			isWindow:  w32.IsWindow(hwnd),
			isEnabled: w32.IsWindowEnabled(hwnd),
			isVisible: w32.IsWindowVisible(hwnd),
		}
		ht.title = strings.ReplaceAll(ht.title, " ", "~") // this will allow me to use ~ as a space in the target.  I hope.
		ht.className, _ = w32.GetClassName(hwnd)
		hwndText = append(hwndText, ht)
		return true
	}
	w32.EnumWindows(enumCallBack)
	ctfmt.Printf(ct.Green, true, " \n Found %d elements in the hwnd text slice. \n Now will find the target of %q.\n", len(hwndText), target)

	pause0()

	var ctr int
	var found bool

	for i, ht := range hwndText {
		if ht.title == "" {
			continue // skip the Printf and search
		}

		ctr++

		if !skipFlag {
			fmt.Printf(" i:%d; hwnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
				i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className) // className is of type string.

			if ctr%40 == 0 && ctr > 0 {
				if pause0() {
					os.Exit(0)
				}
			}
		}

		if target != "" && strings.Contains(ht.title, target) {
			found = true
		} else {
			found = false
		}

		if found {
			ctfmt.Printf(ct.Yellow, true, " window is found.\n")
			ctfmt.Printf(ct.Cyan, true, " i:%d; hWnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
				i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className)

			if noFlag {
				// I might think of something to put here.  Nothing comes to mind yet.
			} else {
				hWnd := ht.h

				var uFlags, param uint
				if ht.isVisible {
					param = w32.SWP_NOACTIVATE
				}
				uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_SHOWWINDOW | param
				w32.SetWindowPos(hWnd, w32.HWND_TOP, 0, 0, 0, 0, uFlags)
				w32.SetForegroundWindow(hWnd)
				fmt.Printf(" I did setWindowPos and then SetForegroundWindow.  I hope it works.\n") // and it worked!!!!.
			}
			if pause0() {
				os.Exit(0)
			}
		}
	}
}

// --------------------------------------------------------------------------------------------

//func pause(n int) bool {
//	scanner := bufio.NewScanner(os.Stdin)
//	fmt.Print(" Pausing ", n, ".  Hit <enter> to continue.  Or 'n' to exit  ")
//	scanner.Scan()
//	if strings.ToLower(scanner.Text()) == "n" {
//		return true
//	}
//	return false
//}
func pause(b bool) bool {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing.  Hit <enter> to continue.  ")
	if b {
		fmt.Printf(" or 'y' to allow the action.  ")
	}
	scanner.Scan()
	if b && strings.ToLower(scanner.Text()) == "y" { // the boolean means to return true on "y"
		return true
	} else if !b && strings.ToLower(scanner.Text()) == "n" { // here it returns true on "n"
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
