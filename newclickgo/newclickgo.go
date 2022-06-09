package main

import (
	"flag"
	"fmt"
	fg "github.com/audrenbdb/goforeground"
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

*/

const lastModified = "June 9, 2022"

var verboseFlag bool
var pid int
var target string

func main() {
	fmt.Printf("newclickgo is my attempt to use Go to activate a process so I can click on the screen.  Last modified %s.  Compiled by %s \n",
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
		fmt.Printf("i = %d, name = %q, PID = %d, PPID = %d.\n", i, processes[i].Executable(), processes[i].Pid(), processes[i].PPid())
		processNameLower := strings.ToLower(processes[i].Executable())
		if target != "" && strings.Contains(processNameLower, target) {
			pid = processes[i].Pid()
			if !verboseFlag { // if verbose, show all processes even after find a match w/ target.
				break
			}
		}
	}

	fmt.Printf(" Target is %q, matched pid = %d.\n", target, pid)

	if pid != 0 { // pid == 0 when target is not found.  Don't want to activate process 0.
		err2 := fg.Activate(pid)
		if err2 != nil {
			fmt.Printf(" Error from fg.Activate is %v.  Exiting \n", err2)
			os.Exit(1)
		}
	}

	fmt.Printf(" There are %d processes found by go-ps.\n", len(processes))

}
