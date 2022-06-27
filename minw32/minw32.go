package main

import (
	"flag"
	"fmt"
	w32a "github.com/JamesHovious/w32"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	w32 "github.com/gonutz/w32/v2"
	"os"
	"runtime"
	"strings"
	"time"
	// ps "github.com/mitchellh/go-ps"
	//"github.com/lxn/win"  I can't get this to be useful.
	//w32 "github.com/gonutz/w32/v2"  I also can't get this to be useful.
	//w32a "github.com/JamesHovious/w32"
	//	ct "github.com/daviddengcn/go-colortext"
	//	ctfmt "github.com/daviddengcn/go-colortext/fmt"
)

/*
  HISTORY
  -------
   8 Jun 22 -- Started playing w/ this.  This will take a while, as I have SIR in Boston soon.
  10 Jun 22 -- Seems to be mostly working.  Tomorrow going to Boston.
  17 Jun 22 -- Back from Boston.  It doesn't work very well on linux.  Time again to try on Win10.
  18 Jun 22 -- Will separate the different ways of getting the pid.  They're not compatible.  And will use TARGET environment variable because
                 the command line params appear in the title.
  23 Jun 22 -- Now called w32, because I'm going to focus just on those routines.  So far, I can't get the other routines to work.
  26 Jun 22 -- Stripping out unneeded stuff, so I can post for help.
*/

const lastModified = "June 26, 2022"

const (
	SW_HIDE           = iota // 0 = hide window and activate another one
	SW_ShowNormal            // 1 = activates and displays a window.  If window is min'd or max'd, restores it to its original size and posn.  App should use this when 1st showing window.
	SW_ShowMinimized         // 2 = activate the window and display it minimized.
	SW_ShowMaximized         // 3 = activate the window and display it maximized.
	SW_ShowNoActivate        // 4 = display window in its most recent size and position, but window is not activated.
	SW_Show                  // 5 = activate window and display it in its current size and posn.
	SW_Minimize              // 6 = minimize window and activate the next top-level window in the Z-order.
	SW_MinNoActive           // 7 = display the window as a minimized window, but don't activate it.
	SW_ShowNA                // 8 = display the window in its current size and position, but don't activate it.
	SW_Restore               // 9 = activate and display the window, and if min or max restore it to its original size and posn.
	SW_ShowDefault           // 10 = sets the show state based on the SW_ value specified in the STARTUPINFO struct passed to the CreateProcess fcn by the pgm that started the app.
	SW_ForceMinimize         // 11 = minimize the window even if the thread that started it is not responding.  Should only be used for windows from a different thread.
)

var target = "w32" // this is a firefox window that's opened.

type htext struct {
	h         w32.HWND
	title     string // this is the title of the window as returned by w32.GetWindowText
	isWindow  bool
	isEnabled bool
	isVisible bool
	className string
}

func main() {
	fmt.Printf("w32 testing routine.  Last modified %s.  Compiled by %s\n",
		lastModified, runtime.Version())

	flag.Parse()

	// w32 section

	foreground := w32.GetForegroundWindow() // this is the same as forgroundWindowH below.
	focus := w32.GetFocus()
	fmt.Printf(" ForegroundWindow()=%v, Getfocus() = %v\n", foreground, focus)

	activeWindowH := w32.GetActiveWindow()            // these are of type hwnd.  This one is zero.
	consoleWindowH := w32.GetConsoleWindow()          // this one is 69412 from both w32 and w32a routines
	desktopWindowH := w32.GetDesktopWindow()          // this one is 65552
	foregroundWindowH := w32.GetForegroundWindow()    // this one is 131244
	consoleW32a := w32a.GetConsoleWindow()            // this one is same as from w32, and is 69412 this run.
	topWindowH := w32.GetTopWindow(foregroundWindowH) // this one is 68854.
	fmt.Printf(" HWND for ... ActiveWindow = %d, ConsoleWindow = %d and %d, DesktopWindow = %d, ForegrndWin = %d, prev foregroundwin = %d, topwin=%d\n",
		activeWindowH, consoleWindowH, consoleW32a, desktopWindowH, foregroundWindowH, foreground, topWindowH)

	fmt.Printf("\n--\n")

	w32ProcessIDs, ok := w32.EnumAllProcesses()
	fmt.Printf(" EnumAllProcesses returned ok of %t and %d processes.\n\n", ok, len(w32ProcessIDs))

	computerName := w32.GetComputerName()
	version := w32.GetVersion() // don't know what this means.
	fmt.Printf(" ComputerName = %v, version = %v\n\n", computerName, version)

	hwndText := make([]htext, 0, 1000) // magic number I expect will be large enough.

	enumCallBack := func(hwnd w32.HWND) bool {
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
		ht.className, _ = w32.GetClassName(hwnd)
		hwndText = append(hwndText, ht)
		return true
	}
	w32.EnumWindows(enumCallBack)
	ctfmt.Printf(ct.Green, true, " \n Found %d elements in the hwnd text slice. \n Now will find the target of %q.\n", len(hwndText), target)

	var ctr int
	var found bool

	for i, ht := range hwndText {
		if ht.title == "" {
			continue // skip the Printf and search
		}

		ctr++

		fmt.Printf(" i:%d; hwnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
			i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className) // className is of type string.

		if ctr%40 == 0 && ctr > 0 {
			if pause() { // allows exiting the test if 'n' is hit.
				os.Exit(0)
			}
		}

		if target != "" && strings.Contains(ht.title, target) {
			found = true
		} else {
			found = false
		}

		if found {
			ctfmt.Printf(ct.Yellow, true, " window is found.  Will now attempt to activate it.\n")
			hWnd := ht.h

			ctfmt.Printf(ct.Magenta, true, " hWnd = %d\n", hWnd)
			ok2 := w32.ShowWindow(hWnd, SW_ShowNormal)
			time.Sleep(10 * time.Millisecond)
			ok3 := w32.ShowWindow(hWnd, SW_Restore)
			time.Sleep(10 * time.Millisecond)
			ok4 := w32.ShowWindow(hWnd, SW_Show)
			time.Sleep(10 * time.Millisecond)
			ok6 := w32.ShowWindow(hWnd, SW_Restore)
			time.Sleep(10 * time.Millisecond)
			ok7 := w32.SetForegroundWindow(hWnd)
			fmt.Printf(" hwnd[%d]=%d, ShowWindow Normal = %t, Restore = %t, Show = %t, Restore = %t and setforegroundwindow = %t.\n",
				i, hWnd, ok2, ok3, ok4, ok6, ok7)
			pause()
		}
	}
}

// --------------------------------------------------------------------------------------------

func pause() bool {
	fmt.Print(" Pausing.  Hit <enter> to continue.  Or 'n' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.Contains(ans, "n") {
		return true
	}
	return false
}
