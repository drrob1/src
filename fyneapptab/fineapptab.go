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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"runtime"
	"time"
)

const LastModified = "August 29, 2021"

var globalA fyne.App
var globalW fyne.Window
var red, green color.Color

// ---------------------------------------------------- main --------------------------------------------------
func main() {

	str := fmt.Sprintf("fyne AppTab example last modified %s, compiled using %s", LastModified, runtime.Version())

	globalA = app.New()
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)
	globalW.SetTitle(str)

	tabs := container.NewAppTabs(
		container.NewTabItem("Tab 1", widget.NewLabel("Hello")),
		container.NewTabItem("Tab 2", widget.NewLabel("World")),
	)

	tabs.Append(container.NewTabItemWithIcon("Home", theme.HomeIcon(), widget.NewLabel("Home Tab")))
	tabs.SetTabLocation(container.TabLocationLeading)

	globalW.SetContent(tabs)

	//go showAnother(globalA)  Not for this example
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
	text1 := canvas.NewText("Green Hello", green)
	text1.TextStyle.Bold = true

	text2 := canvas.NewText("there", red)
	text2.Move(fyne.NewPos(20, 20))

	content := container.New(layout.NewGridLayout(2), text1, text2)

	win2nd.SetContent(content)
	win2nd.Resize(fyne.NewSize(400, 400))
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
	case fyne.KeyHome:
		//firstImage()
	case fyne.KeyEnd:
		//lastImage()
	}
}
