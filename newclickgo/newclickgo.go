package main

import (
	"bufio"
	"flag"
	"fmt"
	fg "github.com/audrenbdb/goforeground"
	"github.com/go-vgo/robotgo"
	ps "github.com/mitchellh/go-ps"
	"os"
	"runtime"
	"strings"
	//"github.com/lxn/win"  I can't get this to be useful.
	//w32 "github.com/gonutz/w32/v2"  I also can't get this to be useful.
)

/*
  HISTORY
  -------
   8 June 22 -- Started playing w/ this.  This will take a while, as I have SIR in Boston soon.
  10 June 22 -- Seems to be mostly working.  Tomorrow going to Boston.

*/

const lastModified = "June 12, 2022"

var verboseFlag bool
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
	fmt.Printf("newclickgo to use Go to activate a process so can be clicked on the screen.  Last modified %s.  Compiled by %s \n",
		lastModified, runtime.Version())

	flag.BoolVar(&verboseFlag, "v", false, " Verbose flag")
	flag.StringVar(&target, "target", "", " Process name search target")
	flag.Parse()

	target = strings.ToLower(target)

	processes, err := ps.Processes()
	if err != nil {
		fmt.Printf(" Error from ps.Processes is %v.  Exiting \n", err)
		os.Exit(1)
	}

	fmt.Printf(" There are %d processes found by go-ps.\n", len(processes))
	pause(1)

	for i := range processes {
		//fmt.Printf("i = %d, name = %q, PID = %d, PPID = %d.\n", i, processes[i].Executable(), processes[i].Pid(), processes[i].PPid())
		processNameLower := strings.ToLower(processes[i].Executable())
		if target != "" && strings.Contains(processNameLower, target) {
			pidProcess = processes[i].Pid()
			fmt.Printf(" Matching process index = %d, pid = %d, PID() = %d, name = %q\n",
				i, pidProcess, processes[i].Pid(), processes[i].Executable())
			break
		}
	}

	fmt.Printf(" Target is %q, matched pid = %d.\n", target, pidProcess)
	pause(2)

	if pidProcess != 0 { // pid == 0 when target is not found.  Don't want to activate process 0.
		err2 := fg.Activate(pidProcess)
		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}

	fmt.Printf(" There are %d processes found by go-ps.\n", len(processes))

	fmt.Printf(" before constructing pets.\n")
	pause(3)

	ids, er := robotgo.FindIds("")
	if er != nil {
		fmt.Printf(" Error from robotgo FindIds is %v.  Exiting\n")
		os.Exit(1)
	}
	fmt.Printf(" robotgo.FindIDs found %d of them.\n", len(ids))
	pause(0)

	var title string
	pets := make([]pet, 0, len(processes))
	for i := range processes {
		pid := processes[i].Pid()
		pid32 := int32(pid)
		fmt.Printf(" after piD = processes[i].Pid(), before robotgo.GetTitle(pid32)\n")
		//title = robotgo.GetTitle(pid32)  this errored out on linux.
		//title = robotgo.GetTitle(ids[i])

		//if pause(4) {
		//	os.Exit(1)
		//}

		if i >= len(ids) {
			fmt.Printf(" i = %d, len(ids) = %d, so will break out of this look.\n")
			break
		}
		apet := pet{ // meaning a pet
			pid:        pid,
			pid32:      pid32,
			id:         ids[i],
			exec:       processes[i].Executable(),
			execLower:  strings.ToLower(processes[i].Executable()),
			title:      title,                  //robotgo.GetTitle(pid32),
			titleLower: strings.ToLower(title), // strings.ToLower(robotgo.GetTitle(pid32)),
		}
		pets = append(pets, apet)
		fmt.Printf(" after construction of apet %d.\n", i)
		//if pause(5) {
		//	os.Exit(1)
		//}
	}

	fmt.Printf(" After construction of PETs.\n")
	//	fmt.Println("robotgo.GetTitle() =", robotgo.GetTitle())

	//name, _ := robotgo.FindName(ids[100])  random test I made up, but don't need anymore.
	//fmt.Printf(" robotgo GetTitle for id[%d], title is %q, and name is %q\n", ids[100], robotgo.GetTitle(ids[100]), name)

	fmt.Printf(" Will now show you my pets.\n")
	pause(6)

	for i, peT := range pets {
		fmt.Printf(" PETs are i=%d; pet: pid=%d, exe=%q, Title=%q; id = %d, name = %q\n",
			i, peT.pid, peT.exec, peT.title, peT.id, processes[i].Executable())
		if i%40 == 0 && i > 0 {
			pause(7)
		}
	}
	fmt.Printf(" There are %d pets and %d processes.\n", len(pets), len(processes))

	pause(8)

	var piD32 int32
	var index int
	for i, peT := range pets {
		//if target != "" && (strings.Contains(peT.title, target) || strings.Contains(peT.exec, target)) {
		if target != "" && strings.Contains(peT.titleLower, target) { // only want to compare against the title, not exec name.
			piD32 = peT.pid32
			index = i
			fmt.Printf(" index = %d, target = %q matches pet PID of %d.  Corresponding processes PID = %d, title = %q, name = %q\n",
				index, target, piD32, processes[i].Pid(), peT.title, peT.exec)
			break
		}
	}

	if piD32 != 0 { // piD == 0 when target is not found.  Don't want to activate process 0.
		err2 := fg.Activate(int(piD32))
		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}

	fmt.Printf(" before robotgo.findids\n")
	pause(9)

	fmt.Printf(" before robotgo.process.\n")
	pause(10)

	nps, e := robotgo.Process()
	if e != nil {
		fmt.Printf(" Error from robotgo process is %v\n", e)
		os.Exit(1)
	}
	fmt.Printf(" robotgo.Process found %d of nps.\n", len(nps))
	for i, np := range nps {
		fmt.Printf(" i = %d, np.pid = %d, np.name = %q \n",
			i, np.Pid, np.Name)
		if strings.Contains(strings.ToLower(np.Name), "filezilla") || strings.Contains(np.Name, "vlc") {
			fmt.Printf(" maybe title is %q\n", robotgo.GetTitle(np.Pid)) // this really doesn't work on linux.
		}
		if i%40 == 0 && i > 0 {
			if pause(11) {
				os.Exit(1)
			}
		}
	}

}

// --------------------------------------------------------------------------------------------

func pause(n int) bool {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing ", n, ".  Hit <enter> to continue.  Or 'y' to exit  ")
	scanner.Scan()
	if strings.ToLower(scanner.Text()) == "y" {
		return true
	}
	return false
}

/*
My notes from going over robotgo docs.  It uses go doc which extracts the documentation from the code.  Perhaps I can do
that w/ fyne

For this to compile on linux, I had to install libxkbcommon-dev package.

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













*/
