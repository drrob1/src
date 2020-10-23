// (C) 1990-2019.  Robert W.  Solomon.  All rights reserved.
// testncurses.go
package main

import (
	"github.com/rthornton128/goncurses"
	"log"
)

const LastAltered = "May 1, 2019"

var StartCol, StartRow, sigfig, MaxRow, MaxCol, TitleRow, StackRow, RegRow, OutputRow, DisplayCol, PromptRow, outputmode, n int

//var BrightYellow, BrightCyan, Black termbox.Attribute

/*
REVISION HISTORY
----------------
 1 May 19 -- Testing a module using ncurses.go
*/

func main() {

	stdscr, err := goncurses.Init()
	if err != nil {
		log.Fatal("init:", err)
	}
	defer goncurses.End()
	//	stdscr.Print("Press enter to continue...")
	//	stdscr.Refresh()

	maxrow, maxcol := stdscr.MaxYX()

	canchangecolor := goncurses.CanChangeColor()

	//	maxnumofcolorpairs := goncurses.ColorPairs()

	//	numofcolors := goncurses.Colors()

	cursesversion := goncurses.CursesVersion()

	hascolors := goncurses.HasColors()

	hasinsertchar := goncurses.HasInsertChar()

	stdscr.Printf(" Maxrow %d, maxcol%d, Can change color %t, NCurses version %s, Has colors %t, Has insert char %t\n",
		maxrow, maxcol, canchangecolor, cursesversion, hascolors, hasinsertchar)

	stdscr.Println()

	stdscr.Print(" pausing...")

	stdscr.Refresh()

	_ = stdscr.GetChar()

}
