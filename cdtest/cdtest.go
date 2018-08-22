package main

import (
	"io"
	"os"
	"runtime"
)

// still doesn't change directory upon exiting
func main() {
	if runtime.GOOS == "linux" {
		HomeDir := os.Getenv("HOME") + "/"
		//		fmt.Println(" change directory to", HomeDir, " on linux")
		//		cmd := exec.Command("cd", HomeDir)
		//		cmd.Stdout = os.Stdout
		//		cmd.Run()
		io.WriteString(os.Stdout, HomeDir)
	} else {
		HomeDir := os.Getenv("userprofile") + "\\"
		//		fmt.Println(" change directory to", HomeDir, " on windows")
		//		cmd := exec.Command("cd", HomeDir)
		//		cmd.Stdout = os.Stdout
		//		cmd.Run()
		io.WriteString(os.Stdout, HomeDir)
	}
}
