/*
REVISION HISTORY
-------- -------
27 Aug 21 -- Copied from a web example.  Now called fynerobot.go
29 Aug 21 -- I really want robotgo mouse control.  Since it doesn't work as a terminal pgm, I'm trying as a fyne GUI pgm.
 2 Sep 21 -- I figured out how to get this to work.  Notes are in GoNotes.txt.  Conclusion is that I have to bundle in \msys2\mingw64\bin\zlib1.dll
             -- count down timer and have the title change w/ every count down number.  Time defaults to 900 sec, but can be set by a flag.
             -- when timer is zero, mouse is moved to coordinates which are defaulted, but can be set by flags for row and col, or X and Y.
             -- loops until exit by kbd or "X"-ing it's window closed.
 3 Sep 21 -- I found on Yussi's computer (empirically) that a value of X=450 and Y=325 works well, and each row is 100 pixels lower, ie, Y += 100.
               Just one spot double-clicked did not keep Epic awake.  I have to do more like what I do w/ the take command batch file.
               Maybe 3 lines in succession, each w/ X incremented or decremented by 10, and Y incremented each time by 100 pixels.
29 Jun 22 -- Fixed depracated MoveMouse -> Move and MouseClick -> Click.
 3 Jul 22 -- Now called gofshowtimer, so I can convert this code to have the ShowTimer function that I currently in Modula-2.
 3 Aug 22 -- Adding <tab> that will send the message "tabbed", and <space> is same as <enter>, X or Q.
 4 Aug 22 -- Seems that <tab> is reserved and not being passed to fyne, so I changed the message to 'cycled'
*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"io"

	//"fyne.io/fyne/v2/theme"
	//"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"image/color"
	"os"
	"runtime"
	"time"
)

const LastModified = "August 4, 2022"

//const clickedX = 450  This was needed in fynerobot but it's not needed here.
//const clickedY = 325
//const incrementY = 100
//const incrementX = 10

const timerDefault = 870
const minTimer = 5

var globalA fyne.App
var globalW fyne.Window
var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var firstX, firstY int
var timer = flag.Int("t", timerDefault, "timer value in seconds")

//var X = flag.Int("x", clickedX, "X (col) value")  Also needed in fynerobot but not here.
//var Y = flag.Int("y", clickedY, "Y (row) value")

// ---------------------------------------------------- main --------------------------------------------------
func main() {

	str := fmt.Sprintf("gofShowTimer last modified %s, compiled using %s", LastModified, runtime.Version())

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
	execname, _ := os.Executable()
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
	time.Sleep(500 * time.Millisecond) // I keep making this shorter and shorter.  It started out as 5 sec.
	countdowntimer := *timer

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C //:
		countdowntimer--
		timeStr := fmt.Sprintf("%d", countdowntimer)
		globalW.SetTitle(timeStr)
		now := time.Now()
		nowStr := now.Format("Mon Jan 2 2006 15:04:05 MST")
		fyneText := canvas.NewText(timeStr, green)
		fyneNow := canvas.NewText(nowStr, blue)
		msg := "esc, q, x: escaped \t return, space: early \t backtick, F1, F2: cycled"
		msgText := canvas.NewText(msg, blue)
		vbox := container.NewVBox(fyneText, fyneNow, msgText)
		cnvs.SetContent(vbox)

		if countdowntimer == 0 {
			io.WriteString(os.Stdout, "normal") // write this string so it's picked up by goclick.
			globalW.Close()

			// These are commented out and were part of fynerobot, but now based on ShowTimer so it's just to countdown and exit.
			//countdowntimer = *timer
			//currentX, currentY := robotgo.GetMousePos()
			//Xcol := *X
			//Yrow := *Y
			//robotgo.Move(Xcol, Yrow)
			//robotgo.Click("left", true) // button, double
			//time.Sleep(500 * time.Millisecond)
			//
			//Xcol += incrementX
			//Yrow += incrementY
			////robotgo.MoveMouse(Xcol, Yrow) depracated
			////robotgo.MouseClick("left", true) depracated
			//robotgo.Move(Xcol, Yrow)
			//robotgo.Click("left", true)
			//time.Sleep(400 * time.Millisecond) // used to be 500
			//
			//Xcol -= incrementX
			//Yrow += incrementY
			////robotgo.MoveMouse(Xcol, Yrow) depracated
			////robotgo.MouseClick("left", true) depracated
			//robotgo.Move(Xcol, Yrow)
			//robotgo.Click("left", true)
			//
			//time.Sleep(300 * time.Millisecond) // used to be 500
			//
			////robotgo.MoveMouse(currentX, currentY)
			//robotgo.Move(currentX, currentY)
		}
	}
} // end changeContent

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
		io.WriteString(os.Stdout, "escaped")
		globalW.Close() // quit's the app if this is the last window, which it is.
	case fyne.KeyHome:
		//firstImage()
	case fyne.KeyEnd:
		//lastImage()
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		io.WriteString(os.Stdout, "early")
		globalW.Close()

	case fyne.KeyBackTick, fyne.KeyF1, fyne.KeyF2:
		io.WriteString(os.Stdout, "cycled")
		globalA.Quit() // just for some variety.
	}

} // end keyTyped
