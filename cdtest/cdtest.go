package main

import (
	"fmt"
	"os"
	"runtime"
)

// this does not exit in HomeDir.  So I'm dropping it.
func main() {
	if runtime.GOOS == "linux" {
		HomeDir := os.Getenv("HOME")
		err := os.Chdir(HomeDir)
		fmt.Println(" change directory to", HomeDir, " on linux, with error", err)
	} else {
		HomeDir := os.Getenv("userprofile")
		err := os.Chdir(HomeDir)
		fmt.Println(" change directory to", HomeDir, " on windows, with error", err)
	}
}
