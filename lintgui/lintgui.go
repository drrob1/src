package main

import (
	"fmt"
	"runtime"
	"src/lint"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

/*
   3 Feb 26 -- First started writing this.
*/

const lastModified = "4 Feb 2026"

func main() {
	a := app.NewWithID("com.example.lintgui")
	s := fmt.Sprintf("Lint GUI, last modified %s, compiled with %s", lastModified, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(600, 600))

	_, startDirFromConfigFile, err := lint.FindAndReadConfIni()
	if err != nil {
		fmt.Printf("line ~30: Error from FindAndReadConfIni is %v\n", err)
		dialog.ShowError(err, w)
	}
	lint.StartDirFromConfigFile = startDirFromConfigFile

	monthsThresholdLabel := widget.NewLabel("Months Threshold:")
	monthsThresholdEntry := widget.NewEntry()
	monthsThresholdEntry.SetText("1") // default value
	lint.MonthsThreshold = 1
	monthsThresholdEntry.OnChanged = func(s string) {
		lint.MonthsThreshold, err = strconv.Atoi(s)
		if err != nil {
			fmt.Printf("line ~42: Message converting months threshold to int is %v\n", err)
			//dialog.ShowError(err, w)  I'm getting spurious errors here while I'm changing the text.
			lint.MonthsThreshold = 1
		}
		fmt.Printf("line ~46: monthsThresholdEntry changed to %d\n", lint.MonthsThreshold)
	}
	monthContainer := container.NewHBox(monthsThresholdLabel, monthsThresholdEntry)

	filenames, err := lint.GetFilenames()
	if err != nil {
		fmt.Printf("line ~52: Error from GetFilenames is %v\n", err)
		dialog.ShowError(err, w)
	}
	if len(filenames) > 26 {
		filenames = filenames[:26] // these are to be displayed in a select box
	}

	var pickedFilename string
	pickedFilenameLabel := widget.NewLabel("Pick a filename:")
	selectFilename := widget.NewSelectEntry(filenames)
	selectFilename.Resize(fyne.Size{Width: 30, Height: 300})
	selectFilename.OnChanged = func(s string) {
		pickedFilename = s
		fmt.Printf("line ~65: selectFilename.OnChanged called with %s, and pickedFilename is %s\n", s, pickedFilename)
		docNames, err := lint.GetDocNames(pickedFilename)
		if err != nil {
			fmt.Printf("line ~70: Error from GetDocNames is %v\n", err)
			dialog.ShowError(err, w)
		}
		lint.Names = docNames
		fmt.Printf("line ~74: DocNames is %#v\n", lint.Names)
	}
	selectFilename.Show()

	w.SetContent(
		container.NewVBox(
			monthContainer,
			pickedFilenameLabel,
			selectFilename,
		),
	)
	w.ShowAndRun()

}
