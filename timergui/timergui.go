package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"runtime"
	"time"
	"unicode"

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
  28 Feb 26 -- This doesn't run on linux, it seems the init part fails.  I'll separate this into a windows and linux versions.
				I decided to ask perplexity for help.  It narrowed it down to possibly a bad window icon.  I fixed it by only loading the icon on Windows.
   1 Mar 26 -- I worked out yesterday, in testbeep.go, that I needed to make the clock png smaller.  I used GIMP to make it 64x64, and that worked.
                So today, I'm going to work on a button to stop the timer go routine.  I'll do it w/ a boolean channel.  A context may also do it, but I would have to research that a bit more.
*/

const lastAltered = "1 Mar 2026"

//go:embed road-runner-beep-beep.mp3
var beepBeep []byte

//go:embed clock-clipart.png
var clockIcon []byte

var stopTimerChan chan bool

func roadRunnerInit() (beep.StreamSeeker, error) {
	f := io.NopCloser(bytes.NewReader(beepBeep))
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/2))
	if err != nil {
		return nil, err
	}
	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	beepStreamer := buffer.Streamer(0, buffer.Len())
	return beepStreamer, nil
}

func beeepTones(frequency float64, duration int) error { // frequency in Hz, duration in milliseconds.
	err := beeep.Beep(frequency, duration)
	return err
}

func main() {
	pflag.Parse()
	a := app.NewWithID("")

	if runtime.GOOS == "windows" { // since it works on Windows, I'll only not use this on linux.
		clockIconRes := fyne.NewStaticResource("clock-clipart.png", clockIcon)
		a.SetIcon(clockIconRes)
	}

	s := fmt.Sprintf("Simple Timer, Last altered: %s, compiled with %s", lastAltered, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(400, 400))

	stopTimerChan = make(chan bool, 1)

	roadRunnerBufStreamer, err := roadRunnerInit()
	if err != nil {
		fmt.Printf("Error from roadRunnerInit is %s\n", err)
		dialog.ShowError(err, w)
	}

	durationEntry := widget.NewEntry()

	timerLabel := widget.NewLabel("...")

	startTimerFunc := func() {
		duration, er := time.ParseDuration(durationEntry.Text)
		if er != nil {
			dialog.ShowError(er, w)
		}
		remaining := int(duration.Seconds())
		for remaining > 0 {
			time.Sleep(1 * time.Second)
			s1 := fmt.Sprintf("%d", remaining)
			s2 := fmt.Sprintf("Time remaining: %d seconds", remaining)
			fyne.Do(func() {
				w.SetTitle(s1)
				timerLabel.SetText(s2)
				// timerLabel.Refresh()  this isn't in the example, and it works without it, so I'm leaving it out.
			})
			select {
			case <-stopTimerChan:
				fmt.Println("stopTimerChan received")
				fyne.Do(func() {
					w.SetTitle("Timer stopped")
					timerLabel.SetText("Timer stopped")
				})
				return
			default: // I may not need this w/ a buffered channel, but I'm leaving it in for now.
				// do nothing
			}
			remaining--
		}

		speaker.Play(roadRunnerBufStreamer)
		fyne.Do(func() {
			timerLabel.SetText("Time's up")
		})
		err = beeepTones(261.6256, 500) // Middle C, also called C4, or c' 1 line octave
		if err != nil {
			fmt.Printf("Error from beeep.Beep in beeepTones is %s\n", err)
			dialog.ShowError(err, w)
		}
		err = beeepTones(440, 500) // A4, a' or high A.
		if err != nil {
			fmt.Printf("Error from beeep.Beep in beeepTones is %s\n", err)
			dialog.ShowError(err, w)
		}
	}

	durationEntry.SetPlaceHolder(" Enter a duration string")
	durationEntry.OnSubmitted = func(_ string) {
		go startTimerFunc()
	}
	if pflag.NArg() > 0 {
		durationEntry.Text = pflag.Arg(0)
		if !unicode.IsLetter(rune(durationEntry.Text[len(durationEntry.Text)-1])) {
			durationEntry.Text += "s"
		}
		go startTimerFunc()
	}

	startTimerBtn := widget.NewButton("Start timer", func() {
		go startTimerFunc()
	})

	stopTimerBtn := widget.NewButton("Stop timer", func() {
		stopTimerChan <- true
	})

	quitBtn := widget.NewButton("Quit", func() {
		err = beeepTones(440, 500) // A4, a' or high A.
		if err != nil {
			fmt.Printf("Error from beeep.Beep in beeepTones is %s\n", err)
			dialog.ShowError(err, w)
		}
		w.Close()
	})

	c := container.NewVBox(timerLabel, durationEntry, startTimerBtn, stopTimerBtn, quitBtn)
	w.SetContent(c)
	w.ShowAndRun()

	err = beeepTones(261.6256, 250) // Middle C, also called C4, or c' 1 line octave
	if err != nil {
		dialog.ShowError(err, w)
	}
	err = beeepTones(440, 250) // A4, a' or high A.
	if err != nil {
		dialog.ShowError(err, w)
	}
	err = beeepTones(220, 250) // A.
	if err != nil {
		dialog.ShowError(err, w)
	}
}
