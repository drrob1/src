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
	"github.com/spf13/pflag"
)

/*
   3 Feb 26 -- First started writing this.
   4 Feb 26 -- GUI is still not quite right, but it's working.  And there should only be 1 button to check both spelling and the schedule.
				I'll have to proceed with the other tasks tomorrow.
				Turns out that I can't automatically check the schedule file yet, because of dependencies.  IE, need to pick a schedule file first, before checking it.
				I figured it out.  It was a matter of making sure the correct variables were defined near the top, so they would be defined before they were used, even if they were not yet set.
   6 Feb 26 -- Removed the subslice operation that limited to 26 filenames for display in the select box.
   7 Feb 26 -- Entire list is now sorted by date stamp.  And I got monthsThresholdEntry working.
   8 Feb 26 -- Adding verboseFlag and veryVerboseFlag.  The shortcut doesn't have these, but if the pgm is started on the command line, then these are available.
  13 Feb 26 -- Removed redundant call to FindAndReadConfIni
   8 Mar 26 -- Added keyboard shortcuts for quit.
  10 Mar 26 -- Added test against empty string for monthsThresholdEntry.OnChanged
------------------------------------------------------------------------------------------------------------------------------------------------------
  14 Mar 26 -- Today is Pi day, but that's not important now.  Now called lingui2, and I want to see if I can use a fyne.list instead of the select entry.
*/

const lastModified = "14 March 2026"

//go:embed schedule.png
var scheduleIcon []byte

func main() {
	var verboseFlag, veryVerboseFlag bool
	var pickedFilename string

	pflag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode")
	pflag.BoolVarP(&veryVerboseFlag, "veryverbose", "V", false, "very verbose mode")
	pflag.Parse()

	if veryVerboseFlag {
		verboseFlag = true
	}

	_, startDirFromConfigFile, _ := lint.FindAndReadConfIni() // I don't care if the file isn't there.
	lint.StartDirFromConfigFile = startDirFromConfigFile
	lint.VerboseFlag = verboseFlag
	lint.VeryVerboseFlag = veryVerboseFlag

	scheduleIconRes := fyne.NewStaticResource("schedule.png", scheduleIcon)
	a := app.NewWithID("com.example.lintgui")
	a.SetIcon(scheduleIconRes)
	s := fmt.Sprintf("Lint GUI2, last modified %s, compiled with %s", lastModified, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(900, 700))

	typedKey := func(ev *fyne.KeyEvent) {
		key := string(ev.Name)
		switch key { // these are all synonyms, but I'm doing this to see if it works.
		case "Q":
			a.Quit()
		case "Escape", "X":
			w.Close()
		}
	}
	w.Canvas().SetOnTypedKey(typedKey)

	monthsThresholdLabel := widget.NewLabel("Months Threshold:")
	monthsThresholdEntry := widget.NewEntry()
	monthsThresholdEntry.SetText("1") // default value
	lint.MonthsThreshold = 1
	filenames, err := lint.GetScheduleFilenames()
	if err != nil {
		dialog.ShowError(err, w)
	}
	selectFilenameList := widget.NewList(
		func() int {
			return len(filenames)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("weekly schedule filename") // placeholder text
		},
		func(i int, o fyne.CanvasObject) {
			filename := o.(*widget.Label)
			filename.SetText(filenames[i])
		},
	)
	//selectFilenameList.Resize(fyne.NewSize(150, 400))  ignored, that's why I need the scrollContainer
	scrollContainer := container.NewVScroll(selectFilenameList)
	scrollContainer.SetMinSize(fyne.NewSize(150, 400))
	monthsThresholdEntry.OnChanged = func(s string) {
		if s != "" {
			s = strings.TrimSpace(s)
			lint.MonthsThreshold, err = strconv.Atoi(s)
			if err != nil {
				lint.MonthsThreshold = 1
				dialog.ShowError(err, w)
			}
			if verboseFlag {
				fmt.Printf("line ~92: monthsThresholdEntry.OnChanged called with %s, and lint.MonthsThreshold is %d\n", s, lint.MonthsThreshold)
			}
			filenames, err = lint.GetScheduleFilenames()
			if err != nil {
				dialog.ShowError(err, w)
			}
			//selectFilenameList.SetOptions(filenames)  This, or an equivalent, may not be needed
			selectFilenameList.Refresh()
		}
	}
	monthContainer := container.NewHBox(monthsThresholdLabel, monthsThresholdEntry)

	spellingErrorsLabel := widget.NewLabel("Spelling Errors go here.") // defined here but used in section below that says check spelling
	spellingErrorsLabel.Wrapping = fyne.TextWrapWord

	messagesLabel := widget.NewLabel("Messages go here.")
	messagesLabel.TextStyle.Bold = true

	spellingFcn := func() { // move this here so it can be called from the scheduleCheckFcn
		soundx := lint.GetSoundex(lint.Names)
		spellingErrors := lint.ShowSpellingErrors(soundx)
		//fmt.Printf("line ~129: spellingErrors is %#v\n", spellingErrors)
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
	selectFilenameList.OnSelected = func(id int) {
		pickedFilename = filenames[id]
		pickedFilenameLabel.SetText(filepath.Base(pickedFilename))
		pickedFilenameLabel.Resize(fyne.Size{Width: 250, Height: 300})
		pickedFilenameLabel.Refresh()
		spellingErrorsLabel.SetText("")
		spellingErrorsLabel.Refresh()
		messagesLabel.SetText("")
		messagesLabel.Refresh()
		//fmt.Printf("line ~174: selectFilenameList.OnSelected called with %d, and pickedFilename is %s\n", id, pickedFilename)
		docNames, err := lint.GetDocNames(pickedFilename)
		if err != nil {
			//fmt.Printf("line ~78: Error from GetDocNames is %v\n", err)
			dialog.ShowError(err, w)
		}
		lint.Names = docNames

		spellingFcn()
		scheduleCheckFcn()
	}
	selectFilenameList.Show()

	if len(filenames) > 0 {
		selectFilenameList.Select(0)
	}

	leftHandColumn := container.NewVBox(monthContainer, pickedFilenameLabel, scrollContainer)
	selectFilenameList.Refresh()
	rightHandColumn := container.NewVBox(spellingBtn, spellingErrorsLabel, scheduleBtn, messagesLabel, quitBtn)
	grid := container.NewAdaptiveGrid(2, leftHandColumn, rightHandColumn)
	grid.Resize(fyne.NewSize(800, 800))

	w.SetContent(
		grid,
	)
	w.CenterOnScreen() // added 3/14/26 as I'm playing w/ a list instead of a select box.
	w.ShowAndRun()

}
