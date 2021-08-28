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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"image/color"
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

	text1 := canvas.NewText("Hello", color.White)
	text2 := canvas.NewText("there", color.White)
	text3 := canvas.NewText("(right)", color.White)
	content := container.New(layout.NewHBoxLayout(), text1, text2, layout.NewSpacer(), text3)

	text4 := canvas.NewText("centered", color.White)
	centered := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), text4, layout.NewSpacer())

	globalW.SetContent(container.New(layout.NewVBoxLayout(), content, centered))
	//	globalW.Resize(fyne.NewSize(500, 500))

	//	globalW.CenterOnScreen()

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
