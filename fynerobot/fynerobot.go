/*
REVISION HISTORY
-------- -------
27 Aug 21 -- Copied from a web example
29 Aug 21 -- I really want robotgo mouse control.  Since it doesn't work as a terminal pgm, I'm trying as a fyne GUI pgm.
 2 Sep 21 -- I figured out how to get this to work.  Notes are in GoNotes.txt.  Conclusion is that I have to bundle in \msys2\mingw64\bin\zlib1.dll
             -- count down timer and have the title change w/ every count down number.  Time defaults to 900 sec, but can be set by a flag.
             -- when timer is zero, mouse is moved to coordinates which are defaulted, but can be set by flags for row and col, or X and Y.
             -- loops until exit by kbd or "X"-ing it's window closed.
 3 Sep 21 -- I found on Yussi's computer (empirically) that a value of X=450 and Y=325 works well, and each row is 100 pixels lower, ie, Y += 100.
               Just one spot double-clicked did not keep Epic awake.  I have to do more like what I do w/ the take command batch file.
               Maybe 3 lines in succession, each w/ X incremented or decremented by 10, and Y incremented each time by 100 pixels.
*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	//"fyne.io/fyne/v2/theme"
	//"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"image/color"
	"os"
	"runtime"
	"time"

)

const LastModified = "Sep 4, 2021"
const clickedX = 450
const clickedY = 325
const incrementY = 100
const incrementX = 10
const timerDefault = 870
const minTimer = 5

var globalA fyne.App
var globalW fyne.Window
var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var firstX, firstY int
var timer = flag.Int("timer", timerDefault, "timer value in seconds")
var X = flag.Int("x", clickedX, "X (col) value")
var Y = flag.Int("y", clickedY, "Y (row) value")

// ---------------------------------------------------- main --------------------------------------------------
func main() {

	str := fmt.Sprintf("fynerobot last modified %s, compiled using %s", LastModified, runtime.Version())

	flag.Parse()

	if *timer < minTimer { // need a minimum timer value else it's too hard to stop the pgm.
		*timer = minTimer
	}

	globalA = app.New()
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	globalW.SetTitle(str)

	myCanvas := globalW.Canvas()

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	headingStr := fmt.Sprintf("%s (%s) last modified %s, last linked %s, and working directory is %s",
		ExecFI.Name(), execname, LastModified, LastLinkedTimeStamp, workingdir)
	globalW.SetTitle(headingStr)

	headingFyne := canvas.NewText(headingStr, red)
	firstX, firstY = robotgo.GetMousePos()

	rawtext := fmt.Sprintf("mouseX (col) =%d, mouseY (row) =%d", firstX, firstY)
	fynetext := canvas.NewText(rawtext, green)
	fynetext.TextStyle.Bold = true

	vbox := container.NewVBox(headingFyne, fynetext)

	myCanvas.SetContent(vbox)
	go changeContent(myCanvas)

	//globalW.Resize(fyne.NewSize(100, 100))

	globalW.ShowAndRun()

} // end main

// ---------------------------------------------------------- changeContent ---------------------------
func changeContent(cnvs fyne.Canvas) {
	time.Sleep(10*time.Second)
	countdowntimer := *timer

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <- ticker.C:
			countdowntimer--
			timeStr := fmt.Sprintf("%d", countdowntimer)
			globalW.SetTitle(timeStr)
			now := time.Now()
			nowStr := now.Format("Mon Jan 2 2006 15:04:05 MST")
			fynetext := canvas.NewText(timeStr, green)
			fynenow := canvas.NewText(nowStr, blue)
			vbox := container.NewVBox(fynetext, fynenow)
			cnvs.SetContent(vbox)

			if countdowntimer == 0 {
				countdowntimer = *timer
				currentX, currentY := robotgo.GetMousePos()

				Xcol := *X
				Yrow := *Y
				robotgo.MoveMouse(Xcol, Yrow)
				robotgo.MouseClick("left", true)
				time.Sleep(500 * time.Millisecond)

				Xcol += incrementX
				Yrow += incrementY
				robotgo.MoveMouse(Xcol, Yrow)
				robotgo.MouseClick("left", true)
				time.Sleep(500 * time.Millisecond)

				Xcol -= incrementX
				Yrow += incrementY
				robotgo.MoveMouse(Xcol, Yrow)
				robotgo.MouseClick("left", true)
				time.Sleep(500 * time.Millisecond)

				robotgo.MoveMouse(currentX, currentY)
			}

		default:
			// do nothing at the moment, but it will loop without blocking.  I don't know which is better.

		}
	}
}
/*
// ---------------------------------------------------------- ShowAnother ----------------------------
func showAnother(a fyne.App) {
	time.Sleep(5 * time.Second)
	win2nd := a.NewWindow("2nd Window")
	win2nd.SetContent(widget.NewLabel("5 seconds later, closed 2 seconds after that"))
	win2nd.Resize(fyne.NewSize(400,400))
	win2nd.Show()

	time.Sleep(2*time.Second)
	win2nd.Close()
}

 */


// ------------------------------------------------------------ keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index is a global var
	switch e.Name {
	case fyne.KeyUp:
		//prevImage()
	case fyne.KeyDown:
		//nextImage()
	case fyne.KeyLeft:
		//prevImage()
	case fyne.KeyRight:
		//nextImage()
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		//globalW.Close() // quit's the app if this is the last window, which it is.
		globalA.Quit()
	case fyne.KeyHome:
		//firstImage()
	case fyne.KeyEnd:
		//lastImage()
	case fyne.KeyEnter, fyne.KeyReturn:
		globalW.Close()
	}
}
