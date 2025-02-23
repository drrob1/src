package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func getAvailablePrinters() ([]string, error) {
	cmd := exec.Command("lpstat", "-le")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	printers := strings.Split(strings.TrimSpace(string(output)), "\n")
	return printers, nil
}

func main() {
	fmt.Printf(" Show Linux Printers\n")
	onLinux := runtime.GOOS == "linux"
	if onLinux {
		fmt.Printf(" only works on linux computers.  This is a linux computer so this should work.\n")
	} else {
		fmt.Printf(" only works on linux computers.  This is NOT a linux computer so this won't work.\n")
		return
	}

	printers, err := getAvailablePrinters()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Available printers:")
	for _, printer := range printers {
		fmt.Println(printer)
	}

	fmt.Printf("\nipp:\n")
	for _, printer := range printers {
		field := strings.Split(printer, " ")
		fmt.Printf("  %s\n", field[3]) // need 4th field
	}
	fmt.Printf("\n\n")
}
