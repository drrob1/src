// fyne1, going thru examples from fyne.io website.
/*
REVISION HISTORY
-------- -------
28 Aug 21 -- Copied from a web example
*/

package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"time"

	//"fyne.io/fyne/v2/container"
	//"fyne.io/fyne/v2/layout"
	//"golang.org/x/exp/rand"
	"image/color"
	"math/rand"
	"runtime"
)

const LastModified = "August 28, 2021"
const maxWidth = 2500
const maxHeight = 2000

var globalA fyne.App
var globalW fyne.Window

// ---------------------------------------------------- rasterfunc ------------------------------------------
func rasterfunc(_, _, w, h int) color.Color {
	return color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 0xff}
}


// ---------------------------------------------------- main --------------------------------------------------
func main() {
	str := fmt.Sprintf("fyneraster example, last modified %s, compiled using %s", LastModified, runtime.Version())

	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	globalW.SetTitle(str)

	raster := canvas.NewRasterWithPixels(rasterfunc)
	// raster := canvas.NewRasterFromImage()

	globalW.SetContent(raster)
	globalW.Resize(fyne.NewSize(300, 300))

	go changeContent(raster)
	globalW.ShowAndRun()

} // end main

// ---------------------------------------------------------- changeContent ---------------------------
func changeContent(cr *canvas.Raster) {
	time.Sleep(1*time.Second)
	for {
		globalW.SetContent(cr)
		time.Sleep(500*time.Millisecond)

		size := cr.Size()
		x := size.Width + 10
		y := size.Height + 10
		globalW.Resize(fyne.Size{x,y})
	}
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
	case fyne.KeyHome:
		//firstImage()
	case fyne.KeyEnd:
		//lastImage()
	}
}
