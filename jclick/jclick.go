package main

import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/gen2brain/beeep"
	"github.com/go-vgo/robotgo"
	"github.com/gonutz/w32/v2"
	"github.com/jonhadfield/findexec"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"src/scanln"
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
   1 Jul 22 -- Repeated calls to activateFirstMatchingWindow don't work, only the first call worked.  Now I have to figure out how to get repeated
                 calls to work.  Perhaps by having another routine move the target window to the bottom of the Z-stack.
   6 Jul 22 -- No longer looking for an environment string called TARGET.  Search string will be in targetStr; interpretation of that string depends on value of useRegexFlag.
                 Will use -rex to set useRegexFlag.
   9 Jul 22 -- useRegexFlag now defaulted to true.  Would need -rex=false to unset the flag.
   3 Aug 22 -- Adding another fyne pgm to act as a 10 second warning before the clicks start, as I do in tcmd.  I just figured out that I don't need a separate routine, I can use
                 what I already have.  And I'll configure this to primarily use gofshowtimer instead of showtimer written in M-2.
                 For me to be able to use the 10 sec popup that can cancel the clicks this round, I must use gofshowtimer.  I'll change the code here and not check for showtimer.
   5 Aug 22 -- Will exclude system32 or cmd.exe, so it won't catch command line params from work that has to use cmd.exe
   8 Aug 22 -- Have to add a time delay for the timer loop in case it can't find gofshowtimer.
   9 Aug 22 -- Still trying to understand why gofshowtimer isn't being called if the binary is in the current directory instead of merely in the path.
  10 Aug 22 -- It's working as hoped at JH.  Now I want to add click function, so that I can click a point on the screen and that will become the new starting (x,y).
                 Wait, that's part of the functions offered by fyne.  This isn't using fyne.  I need to think a bit more.
                 I got it.  There is minTime (5 sec) count down timer that will read current mouse pointer and then ask to use these coordinates.  If not, it
                 displays the values of mouseX and mouseY that will be used.  I can escape out if I wish.
  13 Aug 22 -- I want better defaults.  Now the defaults will depend on allFlag and timer values.
  15 Aug 22 -- Now called jclick.  I will set up default title name for JH.  If there is a title target on the command line, that will be used, else the default will make sense.
  24 Sep 22 -- Now the current mouse pointer is accepted by default.  IE, I reversed the default case.  And I'm adding a timeout so that the default case can be set more quickly.
*/

const lastModified = "Sept 25, 2022"
const clickedX = 450 // default for Jamaica
const clickedY = 325 // default for Jamaica
const incrementY = 100
const fhX = 348 //  Supplanted by reading current mouse position and asking to use that as starting values
const fhY = 370
const beepDuration = 300 // in ms
const minTime = 5        // in sec
const defaultTimer = 870 // in sec
const defaultTarget = "hyperspace.-.jhmc"

var verboseFlag, skipFlag, noFlag, allFlag, fhFlag, gofShowFlag, useRegexFlag bool
var targetStr string // regexStr is in targetStr if useRegexFlag is true
var timer, mouseX, mouseY int
var compRex *regexp.Regexp

type htext struct {
	h         w32.HWND
	title     string // this is the title of the window as returned by w32.GetWindowText
	isWindow  bool
	isEnabled bool
	isVisible bool
	className string
}

var hWinText []htext

func activateFirstMatchingWindow() (int, htext) {
	// The hWinText slice is created in main().  This finds the first match of the target and activates it.  Then it returns.
	// This will return -1 for an error.  So far, error is target not found.  Can't be empty because I check for that first now.

	//if target == "" {
	//	return -1, htext{}
	//}

	for i, ht := range hWinText {
		if ht.title == "" {
			continue // skip the Printf and search
		}
		if strings.HasPrefix(ht.title, "tcc") || strings.Contains(ht.title, "system32") || strings.Contains(ht.title, "cmd.exe") {
			continue // ignore the tcc window itself, as the title will have the command it's currently executing, as will cmd.exe at work.
		}

		var found bool
		if useRegexFlag {
			if compRex.MatchString(ht.title) { // title still is devoid of its spaces
				found = true
			}
		} else {
			if strings.Contains(ht.title, targetStr) { // title still is devoid of its spaces
				found = true
			}
		}

		if found {
			if noFlag {
				if useRegexFlag {
					fmt.Printf(" matched regex of %q with title of %q in slice element [%d]\n", targetStr, ht.title, i)
				} else {
					fmt.Printf(" matched target of %q with title of %q in slice element [%d]", targetStr, ht.title, i)
				}
			} else {
				hWnd := ht.h

				var uFlags, param uint
				if ht.isVisible {
					param = w32.SWP_NOACTIVATE
				}
				uFlags = w32.SWP_NOMOVE | w32.SWP_NOSIZE | w32.SWP_SHOWWINDOW | param
				w32.SetWindowPos(hWnd, w32.HWND_TOP, 0, 0, 0, 0, uFlags)
				w32.SetForegroundWindow(hWnd)
			}
			return i, ht
		}
	}
	return -1, htext{} // if not found, will return zero values for each.  For the htext struct that means it's an empty struct.
} // activateFirstMatchingWindow

func showAllTargetMatches() { // the hWinText slice is created in main().  This finds all matches of the target and shows it without activating them.  Now target may be a regex.
	for i, ht := range hWinText {
		if ht.title == "" {
			continue // skip the Printf and search
		}
		if strings.HasPrefix(ht.title, "tcc") || strings.Contains(ht.title, "system32") || strings.Contains(ht.title, "cmd.exe") {
			continue // ignore the tcc window itself, as the title will have the command it's currently executing, as will cmd.exe at work.
		}

		if useRegexFlag {
			if compRex.MatchString(ht.title) {
				ctfmt.Printf(ct.Yellow, true, " window is found by regex.\n")
				ctfmt.Printf(ct.Cyan, true, " i:%d; hWnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
					i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className)
			}
		} else {
			if strings.Contains(ht.title, targetStr) {
				ctfmt.Printf(ct.Yellow, true, " window is found by simple match.\n")
				ctfmt.Printf(ct.Cyan, true, " i:%d; hWnd %d, title=%q, isWndw %t, isEnbld %t, isVis %t; className = %q\n",
					i, ht.h, ht.title, ht.isWindow, ht.isEnabled, ht.isVisible, ht.className)
			}
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

func main() {
	fmt.Printf("goclick to use Go to activate a process so can be clicked on the screen.  Last modified %s.  Compiled by %s.  For more info use -h \n",
		lastModified, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, "Verbose flag.")
	flag.BoolVar(&skipFlag, "skip", true, "Skip output of all hwnd's found.")
	flag.BoolVar(&noFlag, "no", false, "No activating any windows.  IE, do a trial run.")
	flag.IntVar(&timer, "t", 0, "Timer value for ShowTimer.")
	flag.BoolVar(&allFlag, "all", false, "Show all matches of the TARGET environment variable in the modified titles.")
	flag.IntVar(&mouseX, "x", clickedX, "x coordinate for mouse double clicking.")
	flag.IntVar(&mouseY, "y", clickedY, "y coordinate for mouse double clicking.")
	flag.BoolVar(&fhFlag, "fh", false, "FH defaults instead of JH defaults.")
	flag.BoolVar(&gofShowFlag, "g", false, "gofShowTimer to be used instead of ShowTimer written in Modula-2. ")
	flag.BoolVar(&useRegexFlag, "rex", true, " The command line expression is a regex (or not if false).")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " Search expression defaults to a regular expression.  Would need to set rex=false to change this.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " First test expression using -all flag, then with -no flag, then can set a timer value that is non-zero.  \n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.Arg(0) == "" {
		fmt.Printf(" No target provided.  Will use default of %q.\n", defaultTarget)
		targetStr = defaultTarget
	} else {
		targetStr = strings.ToLower(flag.Arg(0))
	}

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

	// I'm expanding the search directory here
	path := os.Getenv("PATH")
	home, e := os.UserHomeDir()
	if e != nil {
		fmt.Printf(" os.UserHomeDir returned error of %v\n", e)
	}
	searchpath := "." + string(filepath.Separator) + ";" + home + string(filepath.Separator) + ";" + path
	//                                                         execStr := findexec.Find("gofshowtimer.exe", searchpath)
	execStr := findexec.Find("gofshowtimer", searchpath)
	if verboseFlag {
		fmt.Printf(" Looking for gofshowtimer.  Searchpath is %q\n and exec string is %q and %s\n", searchpath, execStr, execStr)
	}

	if execStr == "" {
		fmt.Printf(" gofshowtimer.exe not in path.  Will no longer look for showtimer.exe, so exiting ...\n")
		os.Exit(1)
		/*
			gofshowFlag = false
			execStr = findexec.Find("showtimer.exe", "")
			if execStr == "" {
				fmt.Printf(" showtimer.exe also not in path.  Exiting \n")
				os.Exit(1)
			}
			if verboseFlag {
				fmt.Printf(" Looking for showtimer and exec string is %q\n", execStr)
			}

		*/
	} else {
		gofShowFlag = true
	}
	if !strings.Contains(execStr, ":") { // then it's not a full path
		execStr = ".\\" + execStr
		if verboseFlag {
			fmt.Printf(" Modified execStr is %s\n", execStr)
		}
	}

	if timer == 0 && !allFlag { // if allFlag is set, will leave timer alone.  If allFlag is not set and timer is at the default of 0, will make it defaultTimer
		timer = defaultTimer // value is 870 sec as of this writing.
	}

	// execStr now has gofshowtimer.exe.  I don't check for showtimer.exe as I don't want it anymore.

	// will now set desired start mouse position for the clicking.
	//fmt.Printf(" Counting down from 3 sec and will set starting mouse position for the clicking functions.\n")
	//for i := 3; i > 0; i-- {
	//	fmt.Printf(" %d \r", i)
	//	time.Sleep(1 * time.Second)
	//}

	var ans string
	currentX, currentY, ok := w32.GetCursorPos()
	if !ok {
		ctfmt.Printf(ct.Red, true, " w32.GetCursorPos() returned not ok.  This is odd.  Should I exit? ")
		fmt.Scanln(&ans)
		ans = strings.ToLower(ans)
		if strings.Contains(ans, "y") {
			os.Exit(1)
		}
	}
	fmt.Println()
	fmt.Printf(" Current X = %d, Current Y = %d.  Should I use these values to set X and Y: ", currentX, currentY)
	ans = scanln.WithTimeout(3)
	//n, e := fmt.Scanln(&ans)

	if ans == "" {
		mouseX, mouseY = currentX, currentY
	} else {
		ans = strings.ToLower(ans)
		if strings.Contains(ans, "n") || strings.Contains(ans, "x") { // default is to accept the currentX and currentY as the starting click position.
			// do nothing
		} else {
			mouseX, mouseY = currentX, currentY
		}
	}
	fmt.Printf("\n Will be using X = %d and Y = %d\n", mouseX, mouseY)

	var err error
	if useRegexFlag {
		compRex, err = regexp.Compile(targetStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from compiling the regex of %q is %v\n", targetStr, err)
			os.Exit(1)
		}
	}
	// targetStr = os.Getenv("TARGET")  Environment not used anymore.  targetStr now set by the command line.
	//replaced := strings.NewReplacer("~", " ") // not used anymore as I'm using a regular expression.
	//replaced.Replace(target)

	if !skipFlag {
		fmt.Printf("\n w32 section\n")

		if pause0() {
			os.Exit(0)
		}
	}

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
	ctfmt.Printf(ct.Green, true, " \n Found %d elements in the handle to window text slice.\n", len(hWinText))
	if useRegexFlag {
		ctfmt.Printf(ct.Green, true, " Now will search for the regex of %q.\n", targetStr)
	} else {
		ctfmt.Printf(ct.Green, true, " Now will search for the simple target of %q.\n", targetStr)
	}

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
		showAllTargetMatches() // now either the regex or simple target are passed globally.  The routine uses the useRegexFlag to determine which method is used.
	}

	if !noFlag {
		i, _ := activateFirstMatchingWindow() // now either the regex or simple target are passed globally.  The routine uses the useRegexFlag to determine which method is used.
		if i < 0 {
			fmt.Printf(" Regex or target of %q was not matched.  Exiting\n\n", targetStr)
			os.Exit(1)
		} else {
			fmt.Printf(" Regex or target of %q was matched with hWinText[%d]\n", targetStr, i)
		}
		time.Sleep(2 * time.Second) // need time for the activation (if successful) to occur, else the clicks don't make it onto the activated window.
		moveAndClickMouse(mouseX, mouseY)
		time.Sleep(1 * time.Second)
		minimizeTargetMatchedWindow(i) // this routine builds in a delay of 2 sec.
	}

	var totalIterations int

	if verboseFlag {
		fmt.Printf(" timer = %d, allFlag = %t\n", timer, allFlag)
	}
	if timer > 0 {
		rand.Seed(time.Now().Unix()) // I'm being cute here, randomly choosing version 1 or 2 of gShowTimer just to make sure both are correct.
		var ans string

		for {
			totalIterations++
			if gofShowFlag { // a flag to use gShowTimer instead of the ShowTimer written in Modula-2.
				n := rand.Intn(2) // so result should be 0 or 1.
				if n == 0 {
					ans = gShowTimer1(execStr, timer) // adding execStr because if gofshowtimer is in workingDir, it doesn't start correctly.
				} else {
					ans = gShowTimer2(execStr, timer) // adding execStr because if gofshowtimer is in workingDir, it doesn't start correctly.
				}
				if ans == "escaped" {
					if verboseFlag {
						fmt.Printf(" hit escaped.  n = %d\n", n)
					}

					break
				}
				if verboseFlag {
					fmt.Printf(" answer returned from n=%d is %q\n", n, ans)
				}

				t0 := time.Now()

				// will now beep
				er := beeep.Beep(beeep.DefaultFreq, beepDuration) // duration in ms
				if er != nil {
					fmt.Printf(" Error from beep is %v\n", er)
				}
				er = beeep.Notify("10 sec warning", "Clicks can be aborted w/ <esc>", "")
				if er != nil {
					fmt.Printf(" Error from notify is %v\n", er)
				}

				ans = gShowTimer1(execStr, 10) // adding execStr because if gofshowtimer is in workingDir, it doesn't start correctly.
				if verboseFlag {
					fmt.Printf(" answer returned from 10 sec timer is %q\n", ans)
				}
				if ans == "escaped" { // This is the 10 sec popup warning that the clicks are about to come.
					continue
				} else if ans == "cycled" { // this allows <tab> to stop the loop completely.  This is different from the tcmd version which can't be stopped from the 10 sec popup.
					break
				}

				// now to make sure enough time has elapsed before continuing w/ this loop

				for time.Now().Before(t0.Add(minTime * time.Second)) {
					fmt.Printf(" waiting  \n")
					time.Sleep(1 * time.Second)
				}

			} else { // this is a no-go branch because if gofshowtimer isn't found, the program will exit.
				showTimer(10)
				showTimer(timer) // in the file called showtimer_windows.go
				_, err := os.Stat("st.flg")

				if os.IsNotExist(err) {
					// st.flg is not supposed to exist during the running of this loop.  So keep looping.  But have to get to the if !noFlag sttmnt below.
				} else if err != nil {
					// err is not nil, and it's not because the file doesn't exist.  Time to exit and figure this out.  This isn't supposed to happen at all.
					fmt.Printf(" Not supposed to have an error here from os.Stat for st.flg.  Err is %v\n\n", err)
					break // exit and figure out why this errored.
				}
				if err == nil { // file exists, so it's time to leave.
					err = os.Remove("st.flg")
					if err != nil {
						fmt.Printf(" Error from os.Remove for st.flg is %v\n\n", err)
					}
					break // st.flg exists, so time to exit.\w
				}
			}

			if !noFlag { // this allows me to test the looping w/ the -all flag and nothing will activate.
				i, _ := activateFirstMatchingWindow() // now either the regex or simple target are passed globally.  The routine uses the useRegexFlag to determine which method is used.
				if i < 0 {
					fmt.Printf(" Regex or simple target of %q was not matched\n\n", targetStr)
				}

				time.Sleep(2 * time.Second) // need time for the activation (if successful) to occur, else the clicks don't make it onto the activated window.
				moveAndClickMouse(mouseX, mouseY)
				time.Sleep(time.Second)
				minimizeTargetMatchedWindow(i)
			}
		}
	}
	//fmt.Printf(" Simulating a countdown of the timer using sleep\n")
	//for i := timer; i > 0; i-- {
	//	fmt.Printf(" %d \r", i)
	//	time.Sleep(1 * time.Second)
	//}
	fmt.Printf(" Completed %d iterations of activating the window with the title matching %q.\n\n", totalIterations, targetStr)
} // end main()

// --------------------------------------------------------------------------------------------
/*  Not actually used at the moment.
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
*/
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
