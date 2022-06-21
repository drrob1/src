package main

import (
	"bufio"
	"flag"
	"fmt"
	w32a "github.com/JamesHovious/w32"
	fg "github.com/audrenbdb/goforeground"
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
*/

const lastModified = "June 20, 2022"

var verboseFlag, suppressFlag bool
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

func main() {
	fmt.Printf("goclick to use Go to activate a process so can be clicked on the screen.  Last modified %s.  Compiled by %s\n",
		lastModified, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, " Verbose flag")
	flag.BoolVar(&suppressFlag, "suppress", false, " Suppress output of non-blank titles")
	//flag.StringVar(&target, "target", "", " Process name search target")  Will use the environment string TARGET so this doesn't appear in the title.
	flag.Parse()

	target = os.Getenv("TARGET")
	target = strings.ToLower(target)

	fmt.Printf(" Target is %q after calling Getenv for TARGET\n", target)

	processes, err := ps.Processes()
	if err != nil {
		fmt.Printf(" Error from ps.Processes is %v.  Exiting \n", err)
		os.Exit(1)
	}

	//fmt.Printf(" There are %d processes found by go-ps.\n", len(processes))
	//pause(1)

	var indx int
	if target != "" { // only look to match a target if there is one.
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

	fmt.Printf(" Target is %q, processes index = %d, Exec name matched pid to activate = %d.\n", target, indx, pidProcess)
	if pause(2) {
		os.Exit(0)
	}

	if pidProcess != 0 { // pid == 0 when target is not found or target not set.  Don't want to activate process 0.
		err2 := fg.Activate(pidProcess)
		time.Sleep(2 * time.Second)
		fg.Activate(pidProcess)
		time.Sleep(2 * time.Second)
		robotgo.MaxWindow(int32(pidProcess))
		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}

	//fmt.Printf(" There are %d processes found by go-ps.\n", len(processes))

	//fmt.Printf(" before constructing pets.\n")
	//pause(3)

	ids, er := robotgo.FindIds("")
	if er != nil {
		fmt.Printf(" Error from robotgo FindIds is %v.  Exiting\n")
		os.Exit(1)
	}
	fmt.Printf(" There are %d processes found by go-ps.  And robotgo.FindIDs found %d of them.\n", len(processes), len(ids))
	if pause(0) {
		os.Exit(0)
	}

	var title string
	pspets := make([]pet, 0, len(processes))
	for i := range processes {
		pid := processes[i].Pid()
		pid32 := int32(pid)
		//fmt.Printf(" after piD = processes[i].Pid(), before robotgo.GetTitle(pid32)\n")
		title = robotgo.GetTitle(pid32) //this errored out on linux.
		//title = robotgo.GetTitle(ids[i])  Not sure yet which one to use.
		//if pause(4) {
		//	os.Exit(1)
		//}
		//if i >= len(ids) {
		//	fmt.Printf(" i = %d, len(ids) = %d, so will break out of this loop.\n", i, len(ids))
		//	break
		//}

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
		//fmt.Printf(" after construction of apet %d.\n", i)
		//if pause(5) {
		//	os.Exit(1)
		//}
	}

	fmt.Printf(" After construction of PETs.\n")
	//	fmt.Println("robotgo.GetTitle() =", robotgo.GetTitle())

	//name, _ := robotgo.FindName(ids[100])  random test I made up, but don't need anymore.
	//fmt.Printf(" robotgo GetTitle for id[%d], title is %q, and name is %q\n", ids[100], robotgo.GetTitle(ids[100]), name)

	fmt.Printf(" Will now show you my pets.\n")
	if pause(6) {
		os.Exit(0)
	}

	for i, peT := range pspets {
		if suppressFlag && peT.title == "" { // skip empty titles
			continue
		}
		fmt.Printf(" PETs are i=%d; pet: pid=%d, exe=%q, Title=%q, name = %q\n",
			i, peT.pid, peT.exec, peT.title, processes[i].Executable())
		if i%40 == 0 && i > 0 {
			if pause(7) {
				os.Exit(0)
			}
		}
	}
	fmt.Printf(" There are %d pets and %d processes.\n  About to show based on ids slice.\n", len(pspets), len(processes))

	if pause(8) {
		os.Exit(0)
	}

	for i, id := range ids {
		if suppressFlag && robotgo.GetTitle(id) == "" {
			continue
		}
		name, err := robotgo.FindName(id)
		if err != nil {
			fmt.Printf(" error from robotgo.FindName(%d) is %v\n", id, err)
		}
		fmt.Printf(" ids PET is i=%d, pid=%d, name=%q, title=%q\n", i, id, name, robotgo.GetTitle(id))
		if i%40 == 0 && i > 0 {
			if pause(0) {
				os.Exit(0)
			}
		}
	}
	fmt.Printf(" After for range ids.  About to activate a process based on ids.\n")
	if pause(9) {
		os.Exit(0)
	}

	var piD32 int32
	var index int
	for i, peT := range pspets {
		//if target != "" && (strings.Contains(peT.title, target) || strings.Contains(peT.exec, target)) {
		//if target != "" && strings.Contains(peT.titleLower, target) { // only wanted to compare against the title, but this doesn't work as hoped.
		if target != "" && strings.Contains(peT.execLower, target) {
			piD32 = peT.pid32
			index = i
			fmt.Printf(" index = %d, target = %q matches pet PID of %d.  Corresponding processes PID = %d, title = %q, name = %q\n",
				index, target, piD32, processes[i].Pid(), peT.title, peT.exec)
			break
		}
	}

	if piD32 != 0 { // piD == 0 when target is not found.  Don't want to activate process 0.
		err2 := fg.Activate(int(piD32))
		robotgo.MaxWindow(piD32)

		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}

	fmt.Printf(" before robotgo.process returning a slice of nps.\n")
	pause(10)

	nps, e := robotgo.Process()
	if e != nil {
		fmt.Printf(" Error from robotgo process is %v\n", e)
		os.Exit(1)
	}
	fmt.Printf(" robotgo.Process found %d of nps.\n", len(nps))
	for i, np := range nps {
		if suppressFlag && robotgo.GetTitle(np.Pid) == "" {
			continue
		}
		hwnd := robotgo.FindWindow(np.Name)
		fmt.Printf(" i = %d, np.pid = %d, np.name = %q, title = %q, hwnd = %d \n",
			i, np.Pid, np.Name, robotgo.GetTitle(np.Pid), hwnd)
		//if strings.Contains(strings.ToLower(np.Name), "filezilla") || strings.Contains(np.Name, "vlc") {
		//	fmt.Printf(" maybe title is %q\n", robotgo.GetTitle(np.Pid)) // this really doesn't work on linux.
		//}
		if i%40 == 0 && i > 0 {
			if pause(11) {
				os.Exit(1)
			}
		}
	}

	fmt.Printf(" Found %d processes, %d id and %d nps\n", len(processes), len(ids), len(nps))

	// w32 section

	fmt.Printf("\n w32 section\n")
	if pause0() {
		os.Exit(0)
	}
	fmt.Printf(" Now to use w32.FindWindow\n")

	target = "*" + target + "*"
	//hwnd := w32.FindWindow("", processes[indx].Executable())  doesn't work
	hwnd := w32.FindWindow("MDIClient", target)
	fmt.Printf(" indx=%d, processes[%d].pid=%d, ppid=%d, exec name=%q, target=%q, MDIClient hwnd=%d\n", indx, indx,
		processes[indx].Pid(), processes[indx].PPid(), processes[indx].Executable(), target, hwnd)

	if hwnd > 0 {
		rslt := w32.SetFocus(hwnd)
		fmt.Printf(" after w32.SetFocus(%d), rslt = %d\n", hwnd, rslt)
	}

	hwnd = w32.FindWindow("", target)
	fmt.Printf(" target=%q, empty class hwnd=%d\n", target, hwnd)

	if hwnd > 0 {
		rslt := w32.SetFocus(hwnd)
		fmt.Printf(" after w32.SetFocus(%d), rslt = %d\n", hwnd, rslt)
	}

	hwnd = w32.FindWindow("*", target)
	fmt.Printf(" target=%q, * hwnd=%d\n", target, hwnd)

	if hwnd > 0 {
		rslt := w32.SetFocus(hwnd)
		fmt.Printf(" after w32.SetFocus(%d), rslt = %d\n", hwnd, rslt)
	}

	hwnd = w32.FindWindow("*lient*", target) // covers Client and client
	fmt.Printf(" target=%q, *lient* hwnd=%d\n", target, hwnd)

	if hwnd > 0 {
		rslt := w32.SetFocus(hwnd)
		fmt.Printf(" after w32.SetFocus(%d), rslt = %d\n", hwnd, rslt)
	}

	if pause0() {
		os.Exit(0)
	}

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

	if pause0() {
		os.Exit(0)
	}

	foreground := w32.GetForegroundWindow()
	focus := w32.GetFocus()
	fmt.Printf(" ForegroundWindow()=%v, Getfocus() = %v\n", foreground, focus)
	fmt.Printf(" if focus > 0, About to setfocus on %d\n", focus)
	if pause(14) {
		os.Exit(0)
	}
	if focus > 0 {
		result := w32.SetFocus(focus)
		fmt.Printf(" result from setfocus(%v) is %v\n", focus, result)
	}

	fmt.Printf(" if > 0, about to setfocus on foregroundwindow of %v\n", foreground)
	if pause0() {
		os.Exit(0)
	}
	if foreground > 0 {
		result := w32.SetFocus(foreground)
		fmt.Printf(" result after setfocus on %v is %v\n", foreground, result)
	}

	fmt.Printf(" about to use hardcoded string firefox\n")
	if pause0() {
		os.Exit(0)
	}
	hwnd = w32.FindWindow("MDIClient", "*firefox*")
	fmt.Printf(" After FindWindow on firefox.  hwnd = %v\n", hwnd)
	if hwnd > 0 {
		rslt := w32.SetFocus(hwnd)
		fmt.Printf(" after setfocus on firefox.  Rslt = %v\n", rslt)
	}
	fmt.Printf(" after possible attempt on setfocus firefox.  Will now try vlc\n")
	hwnd = w32.FindWindow("MDIClient", "*vlc*")
	fmt.Printf(" After FindWindow on vlc, hwnd = %v\n", hwnd)
	if hwnd > 0 {
		result := w32.SetFocus(hwnd)
		fmt.Printf(" After setfocus on vlc.  Result = %v\n", result)
	}
	fmt.Printf(" done.\n")
	// hardcoded "firefox" returned 0, but "hardcoded" vlc returned hwnd=1049536.  Maybe I have to use more asterisks.
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
