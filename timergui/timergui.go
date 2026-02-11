package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*
  10 Feb 26 -- Created this program to test the timer.  The example doesn't include a refresh, or fyne.Do.  It works so it doesn't need the refresh().
				It can be easily modified to allow user-settable timer duration.
*/

func main() {

	a := app.NewWithID("")
	w := a.NewWindow("Simple Timer")
	w.Resize(fyne.NewSize(400, 400))

	timerLabel := widget.NewLabel("Time remaining: 10 seconds")

	go func() {
		remaining := 10
		for remaining > 0 {
			time.Sleep(1 * time.Second)
			remaining--
			fyne.Do(func() {
				timerLabel.SetText(fmt.Sprintf("Time remaining: %d seconds", remaining))
				// timerLabel.Refresh()  this isn't in the example, and it works.
			})
		}
		fyne.Do(func() {
			timerLabel.SetText("Time's up")
		})
	}()

	c := container.NewVBox(timerLabel)
	w.SetContent(c)
	w.ShowAndRun()
}
