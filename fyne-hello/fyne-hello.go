package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*
  17 Dec 25 -- Edited and changed the depracated widget.NewVBox to container.NewVBox.
*/

func main() {
	a := app.New()

	quitapp := func() { a.Quit() }

	w := a.NewWindow("Hello")
	// w.SetContent(widget.NewVBox(widget.NewLabel("Hello Fyne"), widget.NewButton("Quit", quitapp)))  widget.NewVBox is depracated.
	w.SetContent(container.NewVBox(widget.NewLabel("Hello Fyne"), widget.NewButton("Quit", quitapp)))
	w.ShowAndRun()
}
