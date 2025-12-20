package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const width = 800
const height = 680

func main() {
	a := app.New()
	w := a.NewWindow("Markdown Editor")

	editWidget := widget.NewMultiLineEntry()
	previewWidget := widget.NewRichTextFromMarkdown("") // empty string just to initialize it
	editWidget.OnChanged = previewWidget.ParseMarkdown

	w.SetContent(container.NewAdaptiveGrid(2, editWidget, previewWidget))

	w.Resize(fyne.NewSize(width, height))
	w.Canvas().Focus(editWidget)
	w.ShowAndRun()

}
