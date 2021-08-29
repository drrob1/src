// fyne1, going thru examples from fyne.io website.
/*
REVISION HISTORY
-------- -------
29 Aug 21 -- Copied from a web example
*/

package main

import (
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"time"

	//"fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/layout"
	//"fyne.io/fyne/v2/container"
	//"image/color"

	"fyne.io/fyne/v2"
	"runtime"
)

const LastModified = "August 29, 2021"

var globalA fyne.App
var globalW fyne.Window
var red, green, blue color.Color
var input *widget.Entry

// ---------------------------------------------------- main --------------------------------------------------
func main() {

	str := fmt.Sprintf("fyneEntry example last modified %s, compiled using %s", LastModified, runtime.Version())

	globalA = app.New()
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	globalW.SetTitle(str)

	input = widget.NewEntry()
	input.SetPlaceHolder(" Entry Widget ...")
	btnfunc := func() {
		log.Println(" text entered:", input.Text)
	}
	btn := widget.NewButton("Save me", btnfunc)
	content := container.NewVBox(input, btn)
	//content = container.NewHBox(input, btn) // cute way to show both methods used.  But this doesn't display as hoped.

	globalW.SetContent(content)
	globalW.Resize(fyne.Size{200, 200})

	globalW.ShowAndRun()

} // end main

// ---------------------------------------------------------- changeContent ---------------------------
/*
func changeContent(c fyne.Canvas) {
	time.Sleep(2*time.Second)

	blue := color.NRGBA{R: 0, G: 0, B: 100, A: 255}
	c.SetContent(canvas.NewRectangle(blue))

	time.Sleep(2*time.Second)
	gray := color.Gray{Y: 100}
	c.SetContent(canvas.NewLine(gray))

	time.Sleep(2*time.Second)
	red := color.NRGBA{R: 0xff, G: 0x33, B: 0x33, A: 0xff}
	circle := canvas.NewCircle(color.White)
	circle.StrokeWidth = 4
	circle.StrokeColor = red
	c.SetContent(circle)

	time.Sleep(2*time.Second)
	c.SetContent(canvas.NewImageFromResource(theme.FyneLogo()))
}

 */

// ---------------------------------------------------------- ShowAnother ----------------------------
func showAnother(a fyne.App) {
	time.Sleep(5 * time.Second)
	win2nd := a.NewWindow("2nd Window")
	win2nd.Canvas().SetOnTypedKey(keyTyped)
	text1 := canvas.NewText("Green Hello", green)
	text1.TextStyle.Bold = true

	text2 := canvas.NewText("there", red)
	text2.Move(fyne.NewPos(20,20))

	content := container.New(layout.NewGridLayout(2), text1, text2)

	win2nd.SetContent(content)
	win2nd.Resize(fyne.NewSize(400,400))
	win2nd.Show()

//	time.Sleep(2*time.Second)
//	win2nd.Close()
}


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
	case fyne.KeyEnter:
		log.Print(" ENTER was hit,")
		log.Println(" and text was entered:", input.Text)
	case fyne.KeyHome:
		//firstImage()
	case fyne.KeyEnd:
		//lastImage()
	}
}
