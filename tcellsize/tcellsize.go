package main

import (
	"fmt"
	"os"

	//"github.com/gdamore/tcell"
	//"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/v2"
)

func main() {
	scrn, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintln(os.Stderr," tcell NewScreen error is", err)
		os.Exit(1)
	}

	if err = scrn.Init(); err != nil {
		fmt.Fprintln(os.Stderr, " scrn init error is", err)
		os.Exit(1)
	}

	maxcol, maxrow := scrn.Size()
	scrn.Fini()

	fmt.Println(" after init and fini of tcell screen.  There are", maxrow, "rows and", maxcol, "columns.")









}
