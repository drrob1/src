package main // hideme.go

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/go-vgo/robotgo"
	"github.com/gonutz/w32/v2"
	"github.com/jonhadfield/findexec"
	"os"
	"runtime"
	"strings"
	"time"
	// ps "github.com/mitchellh/go-ps" // using pid doesn't work to activate a window
	//"github.com/lxn/win"  I can't get this to be useful.
	//w32a "github.com/JamesHovious/w32"
	//"github.com/jonhadfield/findexec"
)

/*
  HISTORY
  -------
   8 Jun 22 -- Started playing w/ this and I called it goclick.go.  This will take a while, as I have SIR in Boston soon.
  10 Jun 22 -- Seems to be mostly working.  Tomorrow going to Boston.
  17 Jun 22 -- Back from Boston.  It doesn't work very well on linux.  Time again to try on Win10.
  18 Jun 22 -- Will separate the different ways of getting the pid.  They're not compatible.  And will use TARGET environment variable because
                 the command line params appear in the title.
  27 Jun 22 -- The testing version I called w32 now works, based on the help I got from Howard C. Shaw III.  Now I have to fine tune it.
                 I'm going to remove all the PID based stuff as that didn't work anyway.  It's saved as oldgoclick.
                 But this will take a little while.
  28 Jun 22 -- TARGET can use tilde squiggle character to represent a space in the title.  I use strings.ReplaceAll " " to "~" on the titles.
                 Added noFlag to do a trial run of what matches before trying to activate it.
   1 Jul 22 -- Repeated calls to activateFirstMatchingWindow don't work, only the first call worked.  Now I have to figure out how to get repeated
                 calls to work.  Perhaps by having another routine move the target window to the bottom of the Z-stack.
   5 Jul 22 -- Now called hideme, and it's purpose is to hide aidoc at work.  This way I don't have to kill its task using the task manager.
                 I'll pass the title name the same way I did for goclick, ie, via a TARGET environment variable.
*/

const lastModified = "July 5, 2022"
const clickedX = 450 // default for Jamaica
const clickedY = 325 // default for Jamaica
const incrementY = 100
const fhX = 348
const fhY = 370

var verboseFlag, skipFlag, noFlag, allFlag, fhFlag, gofshowFlag bool

var timer, mouseX, mouseY int

type htext struct {
	h         w32.HWND
	title     string // this is the title of the window as returned by w32.GetWindowText
	isWindow  bool
	isEnabled bool
	isVisible bool
	className string
}

var hWinText []htext

func hideFirstMatchingWindow(target string) (int, htext) {
	// The hWinText slice is created in main().  This finds the first match of the target and activates it.  Then it returns.
	// This will return -1 for an error.  So far, errors are either target is empty or target not found.
	if target == "" {
		return -1, htext{}
	}

	for i, ht := range hWinText {
		if ht.title == "" {
			continue // skip the Printf and search
		}

		var found bool
		if target != "" && strings.Contains(ht.title, target) {
			found = true
		}

		if found {
			if noFlag {
				// I might think of something to put here.  Nothing comes to mind yet.
			} else {
				hWnd := ht.h

				var uFlags, param uint
				if ht.isVisible {
					param = w32.SWP_NOACTIVATE
				}
				//uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_SHOWWINDOW | param
				uFlags = w32.SWP_HIDEWINDOW | w32.SWP_NOSIZE | w32.SWP_NOMOVE | param
				w32.SetWindowPos(hWnd, w32.HWND_TOP, 0, 0, 0, 0, uFlags)
				w32.SetForegroundWindow(hWnd)
			}
			return i, ht
		}
	}
	return -1, htext{} // if not found, will return zero values for each.  For the htext struct that means it's an empty struct.
} // hideFirstMatchingWindow

func showAllTargetMatches(target string) { // the hWinText slice is created in main().  This finds all matchs of the target and shows it without activating them.
	if target == "" {
		return
	}
	for i, ht := range hWinText {
		if ht.title == "" {
			continue // skip the Printf and search
		}

		if target != "" && strings.Contains(ht.title, target) {
			ctfmt.Printf(ct.Yellow, true, " window is found.\n")
			ctfmt.Printf(ct.Cyan, true, " i:%d; hWnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
				i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className)
		}
	}
} // showAllTargetMatches

func minimizeTargetMatchedWindow(indx int) { // will just use prev'ly located index into the hWinText slice.
	if indx < 0 {
		return // and do nothing.
	}
	var uFlags uint
	//uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_HIDEWINDOW | w32.SWP_NOACTIVATE
	uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_NOACTIVATE
	hWnd := hWinText[indx].h
	w32.SetWindowPos(hWnd, w32.HWND_BOTTOM, 0, 0, 0, 0, uFlags)
	time.Sleep(1 * time.Second)
} // minimizeTargetMatchedWindow

func unhideTargetMatchedWindow(indx int) { // will just use prev'ly located index into the hWinText slice.
	if indx < 0 {
		return // and do nothing.
	}
	var uFlags uint
	//uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_HIDEWINDOW | w32.SWP_NOACTIVATE
	uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_NOACTIVATE | w32.SWP_SHOWWINDOW
	hWnd := hWinText[indx].h
	w32.SetWindowPos(hWnd, w32.HWND_BOTTOM, 0, 0, 0, 0, uFlags)
	time.Sleep(1 * time.Second)
} // minimizeTargetMatchedWindow

func main() {
	fmt.Printf("goclick to use Go to activate a process so can be clicked on the screen.  Last modified %s.  Compiled by %s\n",
		lastModified, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, "Verbose flag.")
	flag.BoolVar(&skipFlag, "skip", true, "Skip output of all hwnd's found.")
	flag.BoolVar(&noFlag, "no", false, "No activating any windows.  IE, do a trial run.")
	flag.IntVar(&timer, "t", 1, "Timer value for ShowTimer.")
	flag.BoolVar(&allFlag, "all", false, "Show all matches of the TARGET environment variable in the modified titles.")
	flag.IntVar(&mouseX, "x", clickedX, "x coordinate for mouse double clicking.")
	flag.IntVar(&mouseY, "y", clickedY, "y coordinate for mouse double clicking.")
	flag.BoolVar(&fhFlag, "fh", false, "FH defaults instead of JH defaults.")
	flag.BoolVar(&gofshowFlag, "g", false, "gofShowTimer to be used instead of ShowTimer written in Modula-2. ")

	flag.Parse()
	if allFlag { // if I want to show all matches of a TARGET, then I don't want to activate any of them.
		noFlag = true
	}

	if fhFlag {
		mouseX = fhX
		mouseY = fhY
	}
	if verboseFlag {
		fmt.Printf(" X = %d, y = %d\n", mouseX, mouseY)
	}

	if gofshowFlag {
		execStr := findexec.Find("gofshowtimer.exe", "")
		if verboseFlag {
			fmt.Printf(" Looking for gofshowtimer and exec string is %q\n", execStr)
		}

		//if execStr == "" {
		//	//dir, _ := os.Getwd()
		//	fmt.Printf(" gofshowtimer.exe not in path.  Exiting\n")
		//	os.Exit(1)
		//}
	} else {
		execStr := findexec.Find("showtimer.exe", "")
		if verboseFlag {
			fmt.Printf(" Looking for showtimer and exec string is %q\n", execStr)
		}

		//if execStr == "" {
		//	//dir, _ := os.Getwd()
		//	fmt.Printf(" showtimer.exe not in path.  Exiting\n")
		//	os.Exit(1)
		//}
	}

	target := os.Getenv("TARGET")
	target = strings.ToLower(target)
	//replaced := strings.NewReplacer("~", " ") // this will allow me to use ~ as a space in the target.
	//replaced.Replace(target) // but the match failed.  So I reversed it by making all spaces a '~' and can match against '~'

	if verboseFlag {
		fmt.Printf(" Target is %q after calling Getenv for TARGET\n", target)
	}

	// w32 section

	if !skipFlag {
		fmt.Printf("\n w32 section\n")

		if pause0() {
			os.Exit(0)
		}
	}

	//fmt.Printf("\n done.\n")

	hWinText = make([]htext, 0, 1000) // magic number I expect will be large enough.  Should be about 500 hwnd on a typical computer.

	enumCallBack := func(hwnd w32.HWND) bool { // this callback fcn is used by EnumWindows to capture the hwnd and related data, esp modified window title.
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
		hWinText = append(hWinText, ht)
		return true
	}
	w32.EnumWindows(enumCallBack)
	ctfmt.Printf(ct.Green, true, " \n Found %d elements in the handle to window text slice. \n Now will find the target of %q.\n", len(hWinText), target)

	if !skipFlag {
		if pause0() {
			os.Exit(0)
		}
	}

	for i, ht := range hWinText {
		if skipFlag { // this is the default.  Need -skip=false to change this
			break
		}
		var ctr int

		if ht.title == "" {
			continue // skip the Printf and search
		}
		fmt.Printf(" About to start displaying hwnd retrieved by EnumWindows.\n")

		ctr++

		fmt.Printf(" i:%d; hwnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
			i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className) // className is of type string.

		if ctr%40 == 0 && ctr > 0 {
			if pause0() {
				os.Exit(0)
			}
		}
	}

	if allFlag {
		showAllTargetMatches(target)
	}

	if !noFlag {
		i, _ := hideFirstMatchingWindow(target)
		if i < 0 {
			fmt.Printf(" TARGET of %q was not matched.  Exiting\n\n", target)
			os.Exit(1)
		} else {
			fmt.Printf(" TARGET of %q was matched with hWinText[%d]\n", target, i)
		}

		moveAndClickMouse(mouseX, mouseY) // this is here so the cmd window is clicked and thereby activated.  Without this the command window does not have focus.

		fmt.Printf(" Called hide me on %q.  About to unhide if allowed.", target)
		if pause0() {
			os.Exit(0)
		}

		if timer > 0 {
			time.Sleep(time.Duration(timer) * time.Second)
		}

		unhideTargetMatchedWindow(i) // just to test if this works.  It does, and if run again it will unhide what was hidden in a previous run.
	}
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

func moveAndClickMouse(x, y int) {
	oldX, oldY, ok := w32.GetCursorPos()
	if !ok {
		return
	}
	robotX, robotY := robotgo.GetMousePos()
	if oldX != robotX || oldY != robotY {
		fmt.Printf(" w32 and robotgo packages return different values.  w32 x = %d, robot x = %d, w32 y = %d, robot y = %d\n",
			oldX, robotX, oldY, robotY)
	}
	if verboseFlag {
		fmt.Printf(" w32 x = %d, robot x = %d, w32 y = %d, robot y = %d\n", oldX, robotX, oldY, robotY)
	}
	ok = w32.SetCursorPos(x, y)
	if !ok {
		return
	}
	robotgo.Click("left", true) // button string, double bool
	time.Sleep(500 * time.Millisecond)
	y += incrementY
	w32.SetCursorPos(x, y)
	robotgo.Click("left", true) // button string, double bool
	time.Sleep(400 * time.Millisecond)
	y += incrementY
	w32.SetCursorPos(x, y)
	robotgo.Click("left", true) // button string, double bool
	time.Sleep(300 * time.Millisecond)

	w32.SetCursorPos(oldX, oldY)
}
