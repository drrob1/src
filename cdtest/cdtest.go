package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// still doesn't change directory upon exiting
func main() {
	if runtime.GOOS == "linux" {
		HomeDir := os.Getenv("HOME")
		fmt.Println(" change directory to", HomeDir, " on linux")
		//		cmd := exec.Command("cd", HomeDir)
		//		cmd.Stdout = os.Stdout
		//		cmd.Run()
		fmt.Fprintf(os.Stdout, "%s", HomeDir)
	} else {
		HomeDir := os.Getenv("userprofile")
		fmt.Println(" change directory to", HomeDir, " on windows")
		cmd := exec.Command("cd", HomeDir)
		cmd.Stdout = os.Stdout
		cmd.Run()
		fmt.Fprint(os.Stdout, HomeDir)
	}
}
