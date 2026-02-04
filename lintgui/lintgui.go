package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"src/lint"
	"strconv"
	"strings"

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

	spellingErrorsLabel := widget.NewLabel("Spelling Errors go here.") // defined here but used in section below that says check spelling
	spellingErrorsLabel.Wrapping = fyne.TextWrapWord

	// Pick a weekly schedule file
	var pickedFilename string
	pickedFilenameLabel := widget.NewLabel("Pick a filename:")
	selectFilename := widget.NewSelectEntry(filenames)
	selectFilename.Resize(fyne.Size{Width: 150, Height: 300})
	selectFilename.OnChanged = func(s string) {
		pickedFilename = s
		pickedFilenameLabel.SetText(filepath.Base(s))
		pickedFilenameLabel.Resize(fyne.Size{Width: 150, Height: 300})
		pickedFilenameLabel.Refresh()
		spellingErrorsLabel.SetText("")
		spellingErrorsLabel.Refresh()
		fmt.Printf("line ~75: selectFilename.OnChanged called with %s, and pickedFilename is %s\n", s, pickedFilename)
		docNames, err := lint.GetDocNames(pickedFilename)
		if err != nil {
			fmt.Printf("line ~78: Error from GetDocNames is %v\n", err)
			dialog.ShowError(err, w)
		}
		lint.Names = docNames
		fmt.Printf("line ~82: DocNames is %#v\n", lint.Names)
	}
	selectFilename.Show()

	// check spelling
	spellingBtn := widget.NewButton("Check Spelling", func() {
		soundx := lint.GetSoundex(lint.Names)
		spellingErrors := lint.ShowSpellingErrors(soundx)
		fmt.Printf("line ~82: spellingErrors is %#v\n", spellingErrors)
		if len(spellingErrors) > 0 {
			spellingErrorsLabel.SetText(strings.Join(spellingErrors, "\n"))
			spellingErrorsLabel.Resize(fyne.NewSize(50, 300))
			spellingErrorsLabel.Refresh()
		}
	})

	// check the weekly schedule Excel file
	scheduleBtn := widget.NewButton("Check Schedule", func() {
		err = lint.ScanXLSfile(pickedFilename)
		if err != nil {
			fmt.Printf("line ~91: Error from ScanXLSfile is %v\n", err)
			dialog.ShowError(err, w)
		}
	})

	quitBtn := widget.NewButton("Quit", func() { a.Quit() })

	leftHandColumn := container.NewVBox(monthContainer, pickedFilenameLabel, selectFilename)
	rightHandColumn := container.NewVBox(spellingErrorsLabel, spellingBtn, scheduleBtn, quitBtn)
	combinedColumn := container.NewHBox(leftHandColumn, rightHandColumn)

	w.SetContent(
		combinedColumn,
	)
	w.ShowAndRun()

}
