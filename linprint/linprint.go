package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

/*
  23 Feb 25 -- First written.  Going to try to see if I can at least get this to happen on Linux.
*/

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
		fmt.Printf("name: %s;      %s\n", field[0], field[3]) // need 1st and 4th fields
	}
	fmt.Printf("\n\n")

	fullString8620Printer := printers[2]
	printer8620Field := strings.Split(fullString8620Printer, " ")
	printer8620 := printer8620Field[0]
	printer8620IPP := printer8620Field[3]
	fmt.Printf("\nname: %q     %s    %q   \n", printer8620, printer8620IPP, printer8620IPP)
	fmt.Printf("\n\n")

}
