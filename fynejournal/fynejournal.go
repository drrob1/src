package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"time"
)

/*
12 Aug 24 -- From LXF 317 Aug 2024.  Part 1/2 written by Andrew Williams, principal of Fyne.io.
             This is a diary or journal keeping app, but it's entirely in memory until next month.
*/

const dateFormat = "2 Jan 2006"

var entries = make(map[string]string)

func main() {
	a := app.New()
	w := a.NewWindow("Fyne Journal")
	var date time.Time

	entry := widget.NewMultiLineEntry()
	entry.OnChanged = func(s string) {
		dateStr := date.Format(dateFormat)
		entries[dateStr] = s
	}

	title := widget.NewLabel("Today")
	title.Alignment = fyne.TextAlignCenter

	setDate := func(d time.Time) {
		date = d
		dateStr := date.Format(dateFormat)
		title.SetText(dateStr)
		entry.SetText(entries[dateStr])
	}
	setDate(time.Now())

	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		setDate(date.Add(time.Hour * -24))
	})
	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		setDate(date.Add(time.Hour * 24))
	})
	bar := container.NewBorder(nil, nil, prev, next, title)

	w.SetContent(container.NewBorder(bar, nil, nil, nil, entry))
	w.Resize(fyne.NewSize(250, 220))
	w.ShowAndRun()

}
