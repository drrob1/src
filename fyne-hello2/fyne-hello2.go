package main

// example says this is a more declarative style for the hello world pgm.
import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

func main() {
	app := app.New()

	quitapp := func() { app.Quit() }

	w := app.NewWindow("Hello")
	w.SetContent(&widget.Box{Children: []fyne.CanvasObject{
		&widget.Label{Text: "Hello Fyne"},
		&widget.Button{Text: "Quit", OnTapped: quitapp}}})

	//  w.SetContent(widget.NewVBox(widget.NewLabel("Hello Fyne"),	widget.NewButton("Quit", quitapp)))
	w.ShowAndRun()
}
