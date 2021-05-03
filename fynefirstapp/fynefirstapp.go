package main

import (
	"fyne.io/fyne/v2/app"
)

func main() {
	app := app.New()

	w := app.NewWindow("Hello")

	w.ShowAndRun()
}
