package main

import "fmt"
import "os"

func main() {
	fmt.Println()
	wd, _ := os.Getwd()
	fmt.Println(" Working Directory is", wd)
	execname, _ := os.Executable()
	fmt.Println(" Executable name is", execname)
	fi, _ := os.Stat(execname)
	fmt.Println(" Name", fi.Name(), ", size", fi.Size(), ", date", fi.ModTime())
}
