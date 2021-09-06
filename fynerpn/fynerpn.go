// From Fyne GUI book by Andrew Williams, Chapter 6, widget.go

package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"runtime"
	"strconv"
	"strings"

	"src/hpcalc2"
)

const lastModified = "Sep 5, 2021"

var globalA fyne.App
var globalW fyne.Window

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var cyan = color.NRGBA{R: 0, G: 255, B: 255, A: 255}

func main() {
	var f float64
	fmt.Printf(" rpnf.go, using fyne.io v2.  Last modified %s, compiled using %s.\n", lastModified, runtime.Version())
	globalA = app.New()
	globalW = globalA.NewWindow("Widget Binding")
	globalW.Canvas().SetOnTypedKey(keyTyped)

	f, _ = hpcalc2.GetResult("t")
	x := binding.BindFloat(&f)

	_, ss := hpcalc2.GetResult("dump")
	ssJoined := strings.Join(ss, "\n")
	shorterSS := ss[1 : len(ss)-1] // removes the first and last strings, which are only character delims
	//shorterSSjoined := strings.Join(shorterSS, "\n")

	f1, _ := x.Get()
	fmt.Println(" Should be today's julian date:", f1)
	resultStr := strconv.FormatFloat(f1, 'g', -1, 64)
	resultStr = hpcalc2.CropNStr(resultStr)
	resultLabel := canvas.NewText("X = "+resultStr, cyan)
	resultLabel.TextSize = 42
	resultLabel.Alignment = fyne.TextAlignCenter

	stackstringslice := binding.BindStringList(&shorterSS)
/*
	Xlabel := widget.NewLabel(ss[1]) // ss[0] is a delimiter string that I don't want here.
	Ylabel := widget.NewLabel(ss[2])
	Zlabel := widget.NewLabel(ss[3])
	T5label := widget.NewLabel(ss[4])
	T4label := widget.NewLabel(ss[5])
	T3label := widget.NewLabel(ss[6])
	T2label := widget.NewLabel(ss[7])
	T1label := widget.NewLabel(ss[8]) // ss[9] is also a delimiter string.
	stackContainer := container.NewVBox(resultLabel, Xlabel, Ylabel, Zlabel, T5label, T4label, T3label, T2label, T1label)
 */

    stackLabel := widget.NewLabel(ssJoined)

	newlabelwidgetfunc := func() fyne.CanvasObject {
		return widget.NewLabel("The Stack")
	}
	bindingfunc := func(i binding.DataItem, o fyne.CanvasObject) {
		o.(*widget.Label).Bind(i.(binding.String))
	}
	listwidget := widget.NewListWithData(stackstringslice, newlabelwidgetfunc, bindingfunc)

	// globalW.SetContent(listwidget)
	// globalW.Resize(fyne.Size{300, 470})

	content := container.NewVBox(resultLabel,stackLabel, listwidget)
	globalW.SetContent(content)
	globalW.Resize(fyne.Size{400, 500})

	globalW.ShowAndRun()
}
func keyTyped(e *fyne.KeyEvent) { // index is a global var
	switch e.Name {
	case fyne.KeyUp:
	case fyne.KeyDown:
	case fyne.KeyLeft:
	case fyne.KeyRight:
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//		(*globalA).Quit()
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyPageUp:
	case fyne.KeyPageDown:
	case fyne.KeyPlus:
	case fyne.KeyMinus:
	case fyne.KeyEqual:
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		globalA.Quit()
	case fyne.KeyBackspace:
	}
}
