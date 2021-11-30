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
29 Nov 21 -- Now called fynereadmouse.go, to do same reading of mouse position but using the Fyne GUI interface.

*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"image/color"
	"os"
	"runtime"
	"strings"
	"time"
)

const LastModified = "Nov 29, 2021"
const clickedX = 450
const clickedY = 325
const incrementY = 100
const incrementX = 10
const timerDefault = 5
const minTimer = 5

type mousepoint struct {
	x, y int
}

var mousePoints []mousepoint

var globalA fyne.App
var globalW fyne.Window
var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var firstX, firstY int
var timedelay = flag.Int("t", timerDefault, "timer value in seconds to wait in between mouse cursor readings.")
var X = flag.Int("x", clickedX, "X (col) value")
var Y = flag.Int("y", clickedY, "Y (row) value")

// ---------------------------------------------------- main --------------------------------------------------
func main() {

	str := fmt.Sprintf("fynerobot last modified %s, compiled using %s", LastModified, runtime.Version())

	flag.Parse()

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
	mousePoints = make([]mousepoint, 0, 10)
	mousePoints = append(mousePoints, mousepoint{firstX, firstY})

	myCanvas.SetContent(vbox)
	go changeContent(myCanvas)

	globalW.ShowAndRun()

} // end main

// ---------------------------------------------------------- changeContent ---------------------------
func changeContent(cnvs fyne.Canvas) {
	time.Sleep(time.Duration(*timedelay) * time.Second)

	ticker := time.NewTicker(time.Duration(*timedelay) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			nowStr := now.Format("Mon Jan 2 2006 15:04:05 MST")
			fynetext := canvas.NewText(nowStr, green)
			fynenow := canvas.NewText(nowStr, blue)
			nextX, nextY := robotgo.GetMousePos()
			rawtext := fmt.Sprintf("mouseX (col) =%d, mouseY (row) =%d", nextX, nextY)
			mousecanvas := canvas.NewText(rawtext, green)
			mousecanvas.TextStyle.Bold = true
			mousePoints = append(mousePoints, mousepoint{nextX, nextY})
			mousePointsCanvas := showMousePoints()
			vbox := container.NewVBox(fynetext, fynenow, mousecanvas, mousePointsCanvas)
			cnvs.SetContent(vbox)
		default:
			// do nothing at the moment, but it will loop without blocking.

		}
	}
}

// ---------------------------------------------------------- showMousePoints ----------------------------
func showMousePoints() fyne.CanvasObject {

	mouseString := make([]string, 0, 10)
	for i, mouse := range mousePoints {
		s := fmt.Sprintf(" point %d: x = %d, y = %d \n", i, mouse.x, mouse.y)
		mouseString = append(mouseString, s)
	}

	mouseStr := strings.Join(mouseString, "\n")
	mouseLabel := widget.NewLabel(mouseStr)
	// mouseScroll := container.NewScroll(mouseLabel)  This looked terrible when tested.
	return mouseLabel
}

// ------------------------------------------------------------ keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index is a global var
	switch e.Name {
	case fyne.KeyUp:
	case fyne.KeyDown:
	case fyne.KeyLeft:
	case fyne.KeyRight:
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalA.Quit()
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyEnter, fyne.KeyReturn:
		globalW.Close()
	}
}
