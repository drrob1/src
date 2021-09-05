// From Fyne GUI book by Andrew Williams, Chapter 6, widget.go

package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"runtime"

	"src/hpcalc2"
)

const lastModified = "Sep 5, 2021"
var globalA fyne.App
var globalW fyne.Window

func main() {
	var f float64
	fmt.Printf(" rpnf.go, using fyne.io v2.  Last modified %s, compiled using %s.\n", lastModified, runtime.Version())
	globalA = app.New()
	globalW = globalA.NewWindow("Widget Binding")
	globalW.Canvas().SetOnTypedKey(keyTyped)

	f, _  = hpcalc2.GetResult("t")
	x := binding.BindFloat(&f)

	_, ss := hpcalc2.GetResult("dump")
	shorterSS := ss[1:len(ss)-1] // removes the first and last strings, which are only character delims

	f1, _ := x.Get()
	fmt.Println(" Should be today's julian date:", f1)

	stackstringslice := binding.BindStringList(&shorterSS)

	newlabelwidgetfunc := func() fyne.CanvasObject {
		return widget.NewLabel("The Stack")
	}
	bindingfunc := func(i binding.DataItem, o fyne.CanvasObject) {
		o.(*widget.Label).Bind(i.(binding.String))
	}
	listwidget := widget.NewListWithData(stackstringslice, newlabelwidgetfunc, bindingfunc)

	globalW.SetContent(listwidget)
	globalW.Resize(fyne.Size{300, 470})


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
