// fyne1, going thru examples from fyne.io website.
/*
REVISION HISTORY
-------- -------
27 Aug 21 -- Copied from a web example
*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"os"
	"runtime"
)

const LastModified = "August 27, 2021"
const maxWidth = 2500
const maxHeight = 2000

var index int
var loadedimg *canvas.Image
var cwd string
var imageInfo []os.FileInfo
var globalA fyne.App
var globalW fyne.Window
var verboseFlag = flag.Bool("v", false, "verbose flag")

// ---------------------------------------------------- main --------------------------------------------------
func main() {
	str := fmt.Sprintf("fyne3 example for Box Layout, last modified %s, compiled using %s", LastModified, runtime.Version())

	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	globalW.SetTitle(str)

	bugURL, _ := url.Parse("https://github.com/fyne-io/fyne/issues/new")

	globalW.SetContent(widget.NewHyperlink("Report a Bug", bugURL))
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
