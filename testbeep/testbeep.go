package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gen2brain/beeep"
	"github.com/spf13/pflag"
)

/*
  10 Feb 26 -- Created this program to test the timer.  The example doesn't include a refresh, or fyne.Do.  It works so it doesn't need the refresh().
				It can be easily modified to allow user-settable timer duration.
  11 Feb 26 -- Added beep-beep sound for timer completion.  And I added an optional command line param to mean seconds.
  14 Feb 26 -- Learned how to use a sound buffer to replay the sound.  And added beeep
  15 Feb 26 -- Added an exit beep sound, adding a lower note, and shortened the durations.  And in the evening I switched the entry field w/ the display label.
  16 Feb 26 -- Uses unicode.IsLetter to determine whether the "s" has to be appended to the duration string on the command line.  If not, it can already have a letter which may not be "s".
				And added a clock icon.
------------------------------------------------------------------------------------------------------------------------------------------------------
  27 Feb 26 -- Now called testbeep.  I want to see what's failing on linux.
*/

const lastAltered = "27 Feb 26"

//go:embed road-runner-beep-beep.mp3
var beepBeep []byte

//go:embed clock-clipart.png
var clockIcon []byte

func main() {
	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	pflag.Parse()
	a := app.NewWithID("")

	clockIconRes := fyne.NewStaticResource("clock-clipart.png", clockIcon)
	a.SetIcon(clockIconRes)

	s := fmt.Sprintf("Test Beep, Last altered: %s, compiled with %s", lastAltered, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(400, 400))

	f := io.NopCloser(bytes.NewReader(beepBeep))
	streamer, format, err = mp3.Decode(f)
	if err != nil {
		dialog.ShowError(err, w)
	}
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/2))
	if err != nil {
		fmt.Printf("Error from speaker.Init is %s\n", err)
		dialog.ShowError(err, w)
	}
	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	_ = streamer.Close()

	beepBeepFunc := func() {
		beepStreamer := buffer.Streamer(0, buffer.Len())
		speaker.Play(beepStreamer)
	}

	beeepFunc := func() {
		err = beeep.Beep(261.6256, 500) // frequency in Hz, duration in milliseconds.  Middle C, also called C4, or c' 1 line octave
		if err != nil {
			fmt.Printf("Error from beeep.Beep is %s\n", err)
			dialog.ShowError(err, w)
		}
		err = beeep.Beep(440, 500) // frequency in Hz, duration in milliseconds.  A4, a' or high A.
		if err != nil {
			fmt.Printf("Error from beeep.Beep is %s\n", err)
			dialog.ShowError(err, w)
		}
	}

	beepBeepBtn := widget.NewButton("Road Runner beep beep", beepBeepFunc)
	beeepBtn := widget.NewButton("Beep", beeepFunc)

	quitBtn := widget.NewButton("Quit", func() {
		w.Close()
	})

	c := container.NewVBox(beepBeepBtn, beeepBtn, quitBtn)
	w.SetContent(c)
	w.ShowAndRun()
}
