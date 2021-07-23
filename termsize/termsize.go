package main

import (
	"fmt"

	//"src/github.com/nathan-fiscaletti/consolesize.go"
//	"github.com/nathan-fiscaletti/consolesize.go"

	//"src/golang.org/x/term"
	"golang.org/x/term"
	//"src/github.com/olekukonko/ts"
//	"github.com/olekukonko/ts"

	// src/github.com/kopoli/go-terminal-size
//	tsize "github.com/kopoli/go-terminal-size"
)

func main() {
//	rows, cols := consolesize.GetConsoleSize()
//	fmt.Println(" GetConsoleSize says", rows, "rows and", cols, "columns.")

	if term.IsTerminal(0) {
		fmt.Println(" in a terminal according to term.IsTerminal")
	} else {
		fmt.Println(" Not in a terminal according to term.IsTerminal.")
	}
	width, height, err := term.GetSize(0)
	if err == nil {
		fmt.Println(" term.GetSize says", height, "rows and", width, "columns.")
	} else {
		fmt.Println(" error from golang.org/x/term.GetSize is", err)
	}


//	size, _ := ts.GetSize()
//	fmt.Println(" ts.GetSize says", size.Rows(), "rows and", size.Cols(), "columns.")

//	var sz tsize.Size
//	s, err = tsize.GetSize()
//	if err == nil {
//		fmt.Println(" tsize.GetSize says", s.Height," rows and", s.Width, "columns.")
//	}








}
