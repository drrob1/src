package main

import (
	"fmt"
	"github.com/godoes/printers"
	"log"
	"runtime"
)

func main() {
	fmt.Printf(" Show Windows Printers\n")
	onWin := runtime.GOOS == "windows"
	if onWin {
		fmt.Printf(" Only works on windows.  This is a Windows system so this should work\n")
	} else {
		fmt.Printf(" Only works on Windows.  This is not windows.  Bye-bye.\n")
		return
	}

	printerNames, err := printers.ReadNames()
	if err != nil {
		log.Fatalf("Error reading printer names: %v", err)
	}

	fmt.Println("Available printers:")
	for _, name := range printerNames {
		fmt.Println(name)
	}

	defaultPrinter, err := printers.GetDefault()
	if err != nil {
		fmt.Printf("Error getting default printer: %v", err)
		return
	}
	fmt.Printf(" Default Printer: %s\n", defaultPrinter)
}
