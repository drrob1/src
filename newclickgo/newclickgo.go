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
  8 Jun 22 -- Started playing w/ this.  This will take a while, as I have SIR in Boston soon.
  10 June 22 -- Seems to be mostly working.  Tomorrow going to Boston.

*/

const lastModified = "June 11, 2022"

var verboseFlag bool
var pid int
var target string

type pet struct {
	pid   int32
	exec  string
	title string
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

	for i := range processes {
		//fmt.Printf("i = %d, name = %q, PID = %d, PPID = %d.\n", i, processes[i].Executable(), processes[i].Pid(), processes[i].PPid())
		processNameLower := strings.ToLower(processes[i].Executable())
		if target != "" && strings.Contains(processNameLower, target) {
			pid = processes[i].Pid()
			fmt.Printf(" Matching process index = %d, pid = %d, PID() = %d, name = %q\n",
				i, pid, processes[i].Pid(), processes[i].Executable())
			break
		}
	}

	fmt.Printf(" Target is %q, matched pid = %d.\n", target, pid)
	pause()

	if pid != 0 { // pid == 0 when target is not found.  Don't want to activate process 0.
		err2 := fg.Activate(pid)
		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}

	fmt.Printf(" There are %d processes found by go-ps.\n", len(processes))

	pause()

	pets := make([]pet, 0, len(processes))
	for i := range processes {
		piD := int32(processes[i].Pid())
		apet := pet{ // meaning a pet
			pid:   piD,
			exec:  strings.ToLower(processes[i].Executable()),
			title: robotgo.GetTitle(piD),
		}
		pets = append(pets, apet)
	}

	fmt.Println(robotgo.GetTitle())
	ids, er := robotgo.FindIds("")
	if er != nil {
		fmt.Printf(" Error from robotgo FindIds is %v.  Exiting\n")
		os.Exit(1)
	}
	name, _ := robotgo.FindName(ids[100])
	fmt.Printf(" robotgo GetTitle for id[%d], title is %q, and name is %q\n", ids[100], robotgo.GetTitle(ids[100]), name)

	fmt.Printf(" Will now show you my pets.\n")
	pause()

	for i, peT := range pets {
		fmt.Printf(" i=%d; pet: PID=%d, exe=%q, Title=%q; processes pid = %d, name = %q\n",
			i, peT.pid, peT.exec, peT.title, processes[i].Pid(), processes[i].Executable())
		if i%40 == 0 {
			pause()
		}
	}
	fmt.Printf(" There are %d pets and %d processes.\n", len(pets), len(processes))

	pause()

	var piD int32
	var index int
	for i, peT := range pets {
		if target != "" && (strings.Contains(peT.title, target) || strings.Contains(peT.exec, target)) {
			piD = peT.pid
			index = i
			fmt.Printf(" index = %d, target = %q matches pet PID of %d.  Corresponding processes PID = %d, title = %q, name = %q\n",
				index, target, piD, processes[i].Pid(), peT.title, peT.exec)
			break
		}
	}

	if piD != 0 { // piD == 0 when target is not found.  Don't want to activate process 0.
		err2 := fg.Activate(int(piD))
		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}
	pause()
}

// --------------------------------------------------------------------------------------------

func pause() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing.  Hit <enter> to continue  ")
	scanner.Scan()
	_ = scanner.Text()
}
