// (C) 2021.  Robert W Solomon.  All rights reserved.
// readmouse.go
package main

import (
	"flag"
	"fmt"
	"github.com/go-vgo/robotgo"
	"os"
	"runtime"
	"time"
)

/*
REVISION HISTORY
----------------
25 Aug 21 -- Starting to play w/ this package
29 Nov 21 -- Now it will loop and keep reporting mouse position until I stop it.

*/

const lastCompiled = "30 Nov 2021"
const timeDelay = 5
const maxIterations = 5

type mousepoint struct {
	x, y int
}

func main() {

	fmt.Printf(" robot.go, based on robotgo to control mouse and keyboard event, etc. Last altered %s, compiled with %s. \n", lastCompiled, runtime.Version())
	var verboseFlag = flag.Bool("v", false, "verbose mode to println more variables and messages.")
	var timedelay = flag.Int("t", timeDelay, "time delay between mouse readings.")
	var maxIter = flag.Int("max", maxIterations, "maximum iterations to read mouse position.")
	flag.Parse()

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	verboseMode := *verboseFlag

	if verboseMode {
		fmt.Println(ExecFI.Name(), "was last linked on", LastLinkedTimeStamp, ".  Working directory is", workingdir, ".")
		fmt.Println(" Full name of executable file is", execname)
	}
	fmt.Println()

	mousex, mousey := robotgo.GetMousePos()
	fmt.Println(" mouseX (col) =", mousex, "and mousey (row) =", mousey)
	mouseSlice := make([]mousepoint, 0, 10)
	mouseSlice = append(mouseSlice, mousepoint{mousex, mousey})

	go func() {
		fmt.Print(" Hit <enter> to terminate.")
		ans := ""
		fmt.Scanln(&ans)
		if len(mouseSlice) > 0 {
			fmt.Println()
			fmt.Println()
			fmt.Println(" Summary of mouse points:")
			for i, mouse := range mouseSlice {
				fmt.Printf(" point %d: x = %d, y = %d \n", i, mouse.x, mouse.y)
			}
		}
		fmt.Println()
		os.Exit(0)
	}()

	for i := 0; i < *maxIter; i++ {
		fmt.Println()
		fmt.Println()
		fmt.Println(" Will read and report mouse position in", *timedelay, "seconds.")
		time.Sleep(time.Duration(*timedelay) * time.Second)
		finalMouseX, finalMouseY := robotgo.GetMousePos()
		fmt.Println(" Final mouse x =", finalMouseX, "and final mouse y =", finalMouseY)
		fmt.Println(" Final row =", finalMouseY, "and final column =", finalMouseX)
                mouseSlice = append(mouseSlice, mousepoint{finalMouseX, finalMouseY})

	}
	robotgo.MoveMouse(mousex, mousey)

	if len(mouseSlice) > 0 {
		fmt.Println()
		fmt.Println()
		fmt.Println(" Summary of mouse points:")
		for i, mouse := range mouseSlice {
			fmt.Printf(" point %d: x = %d, y = %d \n", i, mouse.x, mouse.y)
		}
	}
	fmt.Println()
	fmt.Println()
} // main in robot.go
