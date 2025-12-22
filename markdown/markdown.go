package main

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const width = 800
const height = 680
const minRowsVisible = 30

/*
  21 Dec 25 -- Andy Williams gave a talk at GopherCon UK in 2024 that I heard yesterday.  I created this from his talk, after expanding the minimal code he gave there.
*/

func main() {
	var path, basenameSearchStr string
	a := app.New()
	w := a.NewWindow("Markdown Editor")

	editWidget := widget.NewMultiLineEntry()
	editWidget.SetMinRowsVisible(minRowsVisible)        // got this from perplexity
	previewWidget := widget.NewRichTextFromMarkdown("") // empty string just to initialize it
	editWidget.OnChanged = previewWidget.ParseMarkdown

	typedKey := func(ev *fyne.KeyEvent) { // I separated this out so I can more easily understand it.
		key := string(ev.Name)
		switch key {
		case "Q", "Escape", "X":
			os.Exit(0)
		}
	}
	w.Canvas().SetOnTypedKey(typedKey)

	workingDir, err := os.Getwd()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	curURI, err := listableFromPath(workingDir)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	fileOpenFunc := func(reader fyne.URIReadCloser, err error) { // this closure gets called AFTER the user has selected a file from the fyne dialog.
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		path = reader.URI().Path()
		ext := filepath.Ext(path)
		basenameSearchStr = filepath.Base(path)
		basenameSearchStr = strings.TrimSuffix(basenameSearchStr, ext)
	}

	openBtnFunc := func() { // I want to specify starting directory 1st
		openDialog := dialog.NewFileOpen(fileOpenFunc, w)
		openDialog.SetLocation(curURI)
		openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown"}))
		openDialog.Show()
	}
	openBtn := widget.NewButton("Open markdown file", openBtnFunc)

	showSaveFunc := func(win fyne.Window, data []byte) {
		writeCallback := func(wr fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if wr == nil { // user cancelled
				return
			}
			defer wr.Close()

			if _, err := wr.Write(data); err != nil {
				dialog.ShowError(err, win)
				return
			}
		}
		wrDialog := dialog.NewFileSave(writeCallback, win)
		wrDialog.SetLocation(curURI)
		wrDialog.SetFilter(storage.NewExtensionFileFilter([]string{".md", ".markdown"}))
		wrDialog.Show()
	}

	saveBtnFunc := func() {
		showSaveFunc(w, []byte(editWidget.Text))
	}
	saveBtn := widget.NewButton("Save markdown file", saveBtnFunc)

	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	buttons := container.NewHBox(openBtn, saveBtn, quitBtn)
	editWidget.Resize(fyne.NewSize(width/2, height-50)) // AI wrote these params.  I'll see what it does.
	grid := container.NewAdaptiveGrid(2, editWidget, previewWidget)
	vbox := container.NewVBox(buttons, grid)
	//vbox := container.NewVBox(grid, buttons) // didn't matter

	w.SetContent(vbox)

	w.Resize(fyne.NewSize(width, height))
	w.Canvas().Focus(editWidget)
	w.ShowAndRun()

}

func listableFromPath(path string) (fyne.ListableURI, error) {
	u := storage.NewFileURI(path)
	listerURI, err := storage.ListerForURI(u)
	return listerURI, err
}
