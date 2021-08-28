// fyne1, going thru examples from fyne.io website.
/*
REVISION HISTORY
-------- -------
27 Aug 21 -- Copied from a web example
*/

package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"runtime"
)

const LastModified = "August 28, 2021"
const maxWidth = 2500
const maxHeight = 2000

var globalA fyne.App
var globalW fyne.Window

// ---------------------------------------------------- main --------------------------------------------------
func main() {
	str := fmt.Sprintf("fyne widget.NewEntry example, last modified %s, compiled using %s", LastModified, runtime.Version())

	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	globalW.SetTitle(str)


	globalW.SetContent(widget.NewEntry())
	globalW.Resize(fyne.NewSize(200, 200))

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
	}
}
