package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

func main() {
	app := app.New()

	quitapp := func() { app.Quit() }

	w := app.NewWindow("Hello")
	w.SetContent(widget.NewVBox(widget.NewLabel("Hello Fyne"), widget.NewButton("Quit", quitapp)))
	w.ShowAndRun()
}
