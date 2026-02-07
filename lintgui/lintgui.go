package main

import (
	_ "embed"
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
   4 Feb 26 -- GUI is still not quite right, but it's working.  And there should only be 1 button to check both spelling and the schedule.
				I'll have to proceed with the other tasks tomorrow.
				Turns out that I can't automatically check the schedule file yet, because of dependencies.  IE, need to pick a schedule file first, before checking it.
				I figured it out.  It was a matter of making sure the correct variables were defined near the top, so they would be defined before they were used, even if they were not yet set.
   6 Feb 26 -- Removed the subslice operation that limited to 26 filenames for display in the select box.
   7 Feb 26 -- Entire list is now sorted by date stamp.  And I got monthsThresholdEntry working.
*/

const lastModified = "7 Feb 2026"

//go:embed schedule.png
var scheduleIcon []byte

func main() {
	var pickedFilename string

	// forgot to run the config init.  This picks up an additional start directory.
	_, startDirFromConfigFile, _ := lint.FindAndReadConfIni() // I don't care if the file isn't there.
	lint.StartDirFromConfigFile = startDirFromConfigFile

	scheduleIconRes := fyne.NewStaticResource("schedule.png", scheduleIcon)
	a := app.NewWithID("com.example.lintgui")
	a.SetIcon(scheduleIconRes)
	s := fmt.Sprintf("Lint GUI, last modified %s, compiled with %s", lastModified, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(900, 700))

	_, startDirFromConfigFile, err := lint.FindAndReadConfIni()
	if err != nil {
		//fmt.Printf("line ~30: Error from FindAndReadConfIni is %v\n", err)
		dialog.ShowError(err, w)
	}
	lint.StartDirFromConfigFile = startDirFromConfigFile

	monthsThresholdLabel := widget.NewLabel("Months Threshold:")
	monthsThresholdEntry := widget.NewEntry()
	monthsThresholdEntry.SetText("1") // default value
	lint.MonthsThreshold = 1
	filenames, err := lint.GetScheduleFilenames()
	if err != nil {
		dialog.ShowError(err, w)
	}
	selectFilename := widget.NewSelectEntry(filenames)
	selectFilename.Resize(fyne.Size{Width: 150, Height: 300})
	monthsThresholdEntry.OnChanged = func(s string) {
		lint.MonthsThreshold, err = strconv.Atoi(s)
		fmt.Printf("line ~63: monthsThresholdEntry.OnChanged called with %s, and lint.MonthsThreshold is %d\n", s, lint.MonthsThreshold)
		if err != nil {
			lint.MonthsThreshold = 1
		}
		filenames, err = lint.GetScheduleFilenames()
		if err != nil {
			dialog.ShowError(err, w)
		}
		selectFilename.SetOptions(filenames)
		selectFilename.Refresh()
	}
	monthContainer := container.NewHBox(monthsThresholdLabel, monthsThresholdEntry)

	spellingErrorsLabel := widget.NewLabel("Spelling Errors go here.") // defined here but used in section below that says check spelling
	spellingErrorsLabel.Wrapping = fyne.TextWrapWord

	messagesLabel := widget.NewLabel("Messages go here.")
	messagesLabel.TextStyle.Bold = true

	spellingFcn := func() { // move this here so it can be called from the scheduleCheckFcn
		soundx := lint.GetSoundex(lint.Names)
		spellingErrors := lint.ShowSpellingErrors(soundx)
		//fmt.Printf("line ~82: spellingErrors is %#v\n", spellingErrors)
		if len(spellingErrors) > 0 {
			spellingErrorsLabel.SetText(strings.Join(spellingErrors, "\n"))
			spellingErrorsLabel.Resize(fyne.NewSize(50, 300))
			spellingErrorsLabel.Refresh()
		}
	}

	// check spelling
	spellingBtn := widget.NewButton("Check Spelling", spellingFcn)

	// check the weekly schedule Excel file
	scheduleCheckFcn := func() {
		msg, err := lint.ScanXLSfile(pickedFilename)
		if err != nil {
			//fmt.Printf("line ~91: Error from ScanXLSfile is %v\n", err)
			dialog.ShowError(err, w)
		}
		if len(msg) > 0 {
			msgJoined := strings.Join(msg, "\n")
			messagesLabel.SetText(msgJoined)
			messagesLabel.Resize(fyne.NewSize(150, 300))
			messagesLabel.Refresh()
		} else {
			messagesLabel.SetText("No warnings found in schedule.")
			messagesLabel.Resize(fyne.NewSize(150, 300))
			messagesLabel.Refresh()
		}
	}
	scheduleBtn := widget.NewButton("Check Schedule", scheduleCheckFcn)

	quitBtn := widget.NewButton("Quit", func() { a.Quit() })

	// Pick a weekly schedule file
	pickedFilenameLabel := widget.NewLabel("Pick a filename:")
	pickedFilenameLabel.Resize(fyne.Size{Width: 150, Height: 300})
	selectFilename.SetMinRowsVisible(10) // no difference
	selectFilename.OnChanged = func(s string) {
		pickedFilename = s
		pickedFilenameLabel.SetText(filepath.Base(s))
		pickedFilenameLabel.Resize(fyne.Size{Width: 250, Height: 300})
		pickedFilenameLabel.Refresh()
		spellingErrorsLabel.SetText("")
		spellingErrorsLabel.Refresh()
		messagesLabel.SetText("")
		messagesLabel.Refresh()
		//fmt.Printf("line ~75: selectFilename.OnChanged called with %s, and pickedFilename is %s\n", s, pickedFilename)
		docNames, err := lint.GetDocNames(pickedFilename)
		if err != nil {
			//fmt.Printf("line ~78: Error from GetDocNames is %v\n", err)
			dialog.ShowError(err, w)
		}
		lint.Names = docNames

		spellingFcn()
		scheduleCheckFcn()
	}
	selectFilename.Show()

	leftHandColumn := container.NewVBox(monthContainer, pickedFilenameLabel, selectFilename)
	rightHandColumn := container.NewVBox(spellingBtn, spellingErrorsLabel, scheduleBtn, messagesLabel, quitBtn)
	//combinedColumn := container.NewHBox(leftHandColumn, rightHandColumn)
	grid := container.NewAdaptiveGrid(2, leftHandColumn, rightHandColumn)
	grid.Resize(fyne.NewSize(800, 800))

	w.SetContent(
		grid,
	)
	w.ShowAndRun()

}
