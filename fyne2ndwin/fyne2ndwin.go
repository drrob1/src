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
	"fyne.io/fyne/v2/app"
	"image/color"

	//"fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/layout"
	//"fyne.io/fyne/v2/container"
	//"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// -------------------------------------------------------- isNotImageStr ----------------------------------------
func isNotImageStr(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	isImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
	return !isImage
}
// ---------------------------------------------------- main --------------------------------------------------
func main() {
	//	verboseFlag = flag.Bool("v", false, "verbose flag")
	//  flag.Parse()
	//if flag.NArg() < 1 {
	//	fmt.Fprintln(os.Stderr, " Usage: img <image file name>")
	//	os.Exit(1)
	//}

	str := fmt.Sprintf("fyne1 example last modified %s, compiled using %s", LastModified, runtime.Version())


	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	rect := canvas.NewRectangle(color.White)

	globalW.SetTitle(str)
	globalW.SetContent(rect)
	globalW.Resize(fyne.NewSize(500, 500))


	globalW.CenterOnScreen()





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
