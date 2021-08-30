// fyne1, going thru examples from fyne.io website.
/*
REVISION HISTORY
-------- -------
29 Aug 21 -- Copied from a web example
*/

package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
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
/*
Trying out alternate syntaxes.  This works as a func literal.
	input.OnSubmitted = func(s string) {
		log.Println(" func literal closure func: ENTER was hit:", s)
	}

 */

	submitted := func(s string) {
		log.Println(" func assigned closure ENTER was hit:", s)
	}
	input.OnSubmitted = submitted // after help from Andy Williams, a principal in the fyne.io project.

	btn := widget.NewButton("Wider save me", btnfunc)
	content := container.NewVBox(input, btn)
	//content = container.NewHBox(input, btn) // cute way to show both methods used.  But this doesn't display as hoped.

	globalW.SetContent(content)
	globalW.Resize(fyne.Size{200, 200})

	globalW.ShowAndRun()

} // end main

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
		log.Print(" ENTER/RETURN was hit,")
	}
}
