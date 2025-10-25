package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

func main() {
	a := app.New()

	quitapp := func() { a.Quit() }

	w := a.NewWindow("Hello")
	w.SetContent(widget.NewVBox(widget.NewLabel("Hello Fyne"), widget.NewButton("Quit", quitapp)))
	w.ShowAndRun()
}
