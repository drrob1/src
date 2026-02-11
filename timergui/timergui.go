package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

/*
  10 Feb 26 -- Created this program to test the timer.  The example doesn't include a refresh, or fyne.Do.  It works so it doesn't need the refresh().
				It can be easily modified to allow user-settable timer duration.
*/

const lastAltered = "11 Feb 26"

func main() {

	a := app.NewWithID("")
	w := a.NewWindow("Simple Timer")
	w.Resize(fyne.NewSize(400, 400))

	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder(" Enter a duration string")

	timerLabel := widget.NewLabel("...")

	startTimerFunc := func() {
		duration, err := time.ParseDuration(durationEntry.Text)
		if err != nil {
			dialog.ShowError(err, w)
		}
		remaining := int(duration.Seconds())
		for remaining > 0 {
			time.Sleep(1 * time.Second)
			remaining--
			s1 := fmt.Sprintf("%d", remaining)
			s2 := fmt.Sprintf("Time remaining: %d seconds", remaining)
			fyne.Do(func() {
				w.SetTitle(s1)
				timerLabel.SetText(s2)
				// timerLabel.Refresh()  this isn't in the example, and it works.
			})
		}
		fyne.Do(func() {
			timerLabel.SetText("Time's up")
		})
	}

	startTimerBtn := widget.NewButton("Start timer", func() {
		go startTimerFunc()
	})

	quitBtn := widget.NewButton("Quit", func() {
		w.Close()
	})

	c := container.NewVBox(durationEntry, timerLabel, startTimerBtn, quitBtn)
	w.SetContent(c)
	w.ShowAndRun()
}
