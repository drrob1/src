package main

import (
	"bufio"
	"flag"
	"fmt"
	w32a "github.com/JamesHovious/w32"
	"github.com/go-vgo/robotgo"
	w32 "github.com/gonutz/w32/v2"
	"github.com/mitchellh/go-ps"
	"os"
	"runtime"
	"strings"
	"time"
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
  23 Jun 22 -- Now called w32, because I'm going to focus just on those routines.  So far, I can't get the other routines to work.
*/

const lastModified = "June 23, 2022"

const (
	SW_HIDE       = iota // 0 = hide window and activate another one
	SW_ShowNormal        // 1 = activates and displays a window.  If the window is minimized or maximized, the system restores it to its original size and posn.
	//                          App should use this when first displaying a window.
	SW_ShowMinimized  // 2 = activate the window and display it minimized.
	SW_ShowMaximized  // 3 = activate the window and display it maximized.
	SW_ShowNoActivate // 4 = display window in its most recent size and position, but window is not activated.
	SW_Show           // 5 = activate window and display it in its current size and posn.
	SW_Minimize       // 6 = minimize window and activate the next top-level window in the Z-order.
	SW_MinNoActive    // 7 = display the window as a minimized window, but don't activate it.
	SW_ShowNA         // 8 = display the window in its current size and position, but don't activate it.
	SW_Restore        // 9 = activate and display the window, and if min or max restore it to its original size and posn.
	SW_ShowDefault    // 10 = sets the show state based on the SW_ value specified in the STARTUPINFO struct passed to the CreateProcess fcn by the pgm that started the app.
	SW_ForceMinimize  // 11 = minimize the window even if the thread that started it is not responding.  Should only be used for windows from a different thread.
)

var verboseFlag, suppressFlag, skipFlag bool
var pidProcess int
var target string

type pet struct {
	pid32      int32
	pid        int
	id         int32
	exec       string
	execLower  string
	title      string
	titleLower string
}

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

	flag.BoolVar(&verboseFlag, "v", false, " Verbose flag")
	flag.BoolVar(&suppressFlag, "suppress", false, " Suppress output of non-blank titles")
	//flag.StringVar(&target, "target", "", " Process name search target")  Will use the environment string TARGET so this doesn't appear in the title.
	flag.BoolVar(&skipFlag, "skip", false, "Skip to w32 section")
	flag.Parse()

	target = os.Getenv("TARGET")
	target = strings.ToLower(target)

	if verboseFlag {
		fmt.Printf(" Target is %q after calling Getenv for TARGET\n", target)
	}

	processes, err := ps.Processes()
	if err != nil {
		fmt.Printf(" Error from ps.Processes is %v.  Exiting \n", err)
		os.Exit(1)
	}

	var indx int
	if target != "" && !skipFlag { // only look to match a target if there is one.
		for i := range processes {
			//fmt.Printf("i = %d, name = %q, PID = %d, PPID = %d.\n", i, processes[i].Executable(), processes[i].Pid(), processes[i].PPid())
			processNameLower := strings.ToLower(processes[i].Executable())
			if target != "" && strings.Contains(processNameLower, target) {
				indx = i
				pidProcess = processes[i].Pid()
				fmt.Printf(" Matching process index against exe name = %d, pid = %d, PID() = %d, name = %q\n",
					i, pidProcess, processes[i].Pid(), processes[i].Executable())
				break
			}
		}
	}
	_ = indx

	ids, er := robotgo.FindIds("")
	if er != nil {
		fmt.Printf(" Error from robotgo FindIds is %v.  Exiting\n")
		os.Exit(1)
	}
	fmt.Printf(" There are %d processes found by go-ps.  And robotgo.FindIDs found %d of them.\n", len(processes), len(ids))

	var title string
	pspets := make([]pet, 0, len(processes))
	for i := range processes {
		pid := processes[i].Pid()
		pid32 := int32(pid)
		title = robotgo.GetTitle(pid32) //this errored out on linux.
		apet := pet{ // meaning a pet
			pid:   pid,
			pid32: pid32,
			//id:         ids[i],  This doesn't sync w/ processes.  I'm separating them out.
			exec:       processes[i].Executable(),
			execLower:  strings.ToLower(processes[i].Executable()),
			title:      title,                  //robotgo.GetTitle(pid32),
			titleLower: strings.ToLower(title), // strings.ToLower(robotgo.GetTitle(pid32)),
		}
		pspets = append(pspets, apet)
	}

	fmt.Printf(" There are %d pets and %d processes.\n  About to show based on ids slice.\n", len(pspets), len(processes))

	nps, e := robotgo.Process()
	if e != nil {
		fmt.Printf(" Error from robotgo process is %v\n", e)
		os.Exit(1)
	}

	//fmt.Printf(" Found %d processes, %d id and %d nps\n", len(processes), len(ids), len(nps))

	// w32 section

	fmt.Printf("\n w32 section\n")
	if pause0() {
		os.Exit(0)
	}

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
	fmt.Printf(" Found %d processes, %d id, %d nps and %d w32.processIDs\n", len(processes), len(ids), len(nps), len(w32ProcessIDs))
	fmt.Printf(" EnumAllProcesses returned ok of %t.\n\n", ok)

	computerName := w32.GetComputerName()
	version := w32.GetVersion()
	fmt.Printf(" ComputerName = %v, version = %v\n\n", computerName, version)

	x, y, okk := w32.GetCursorPos() // this means mouse position.
	fmt.Printf(" x = %d, y = %d, ok = %t\n\n", x, y, okk)

	if pause0() {
		os.Exit(0)
	}

	fmt.Printf(" About to start displaying hwnd retrieved by EnumWindows.\n")
	hwndText := make([]htext, 0, 1000) // magic number I expect will be large enough.
	var ctr int
	var found bool
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
		hwndText = append(hwndText, ht)
		if ht.title == "" {
			return true // skip the Printf statements
		}
		if target != "" && strings.Contains(ht.title, target) {
			found = true
			ht.className, _ = w32.GetClassName(hwnd)

		} else {
			found = false
		}

		fmt.Printf(" hwnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; found is %t; className = %q\n",
			ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, found, ht.className) // the type of the className is string.

		ctr++
		//if ctr%2 == 0 { // every other line output an extra line.
		//	fmt.Println()
		//}
		if ctr%10 == 0 && ctr > 0 {
			if pause0() {
				os.Exit(0)
			}
		}

		if found {
			fmt.Printf(" about to setforeground for the found window")
			if pause(true) {
				ok5 := w32.ShowWindow(hwnd, SW_HIDE)
				time.Sleep(10 * time.Millisecond)
				ok2 := w32.ShowWindow(hwnd, SW_ShowNormal)
				time.Sleep(10 * time.Millisecond)
				ok3 := w32.ShowWindow(hwnd, SW_Minimize)
				time.Sleep(10 * time.Millisecond)
				ok4 := w32.ShowWindow(hwnd, SW_Show)
				fmt.Printf(" ShowWindow hide returned %t, Normal returned %t, Min returned %t, and Show returned %t.\n", ok5, ok2, ok3, ok4)

				//ok2 := w32.ShowWindowAsync(hwnd, 1)
				//ookk := w32.SetForegroundWindow(hwnd) doesn't work, returns 0
				//enabled := w32.EnableWindow(hwnd, true)  doesn't work, returns false
				//result := w32.SetFocus(hwnd) doesn't work.  Always returns 0.
				//result := w32.SetCapture(hwnd) doesn't work, returns 0
				//fmt.Printf(" Result from SetForegroundWindow(%d) is %t, ShowWindow result is %t, enabled = %t and captured = %v\n", hwnd, ookk, ok2, enabled, result)
				//time.Sleep(5 * time.Second)
			}
		}

		return true
	}
	w32.EnumWindows(enumCallBack)
	fmt.Printf(" \n Found %d elements in the hwnd text slice, and %d of them have a title. \n", len(hwndText), ctr)

	pause0()

}

// --------------------------------------------------------------------------------------------

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
