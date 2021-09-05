// From Fyne GUI book by Andrew Williams, Chapter 6, widget.go

package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"runtime"

	"src/hpcalc2"
)

const lastModified = "Sep 5, 2021"

func makeUI() fyne.CanvasObject {

	fmt.Printf(" rpnf.go, using fyne.io v2.  Last modified %s, compiled using %s.\n", lastModified, runtime.Version())
	f := binding.NewFloat()

	prog := widget.NewProgressBarWithData(f)
	slide := widget.NewSliderWithData(0, 1, f)
	slide.Step = 0.01
	btnfunc := func() { // my edits
		_ = f.Set(0.5)
	}
	btn := widget.NewButton("Set to 0.5", btnfunc) // my edits

	return container.NewVBox(prog, slide, btn)
}

func main() {
	var f float64
	a := app.New()
	w := a.NewWindow("Widget Binding")

	f, _  = hpcalc2.GetResult("t")
	x := binding.BindFloat(&f)

	_, ss := hpcalc2.GetResult("dump")

	f1, _ := x.Get()
	fmt.Println(" Should be today's julian date:", f1)

	stackstringslice := binding.BindStringList(&ss)

	newlabelwidgetfunc := func() fyne.CanvasObject {
		return widget.NewLabel("The Stack")
	}
	bindingfunc := func(i binding.DataItem, o fyne.CanvasObject) {
		o.(*widget.Label).Bind(i.(binding.String))
	}
	listwidget := widget.NewListWithData(stackstringslice, newlabelwidgetfunc, bindingfunc)

	shorterSS := ss[1:len(ss)-1] // removes the first and last strings, which are only character delims
	shorterStackStringSlice := binding.BindStringList(&shorterSS)
	newlabelwidgetfunc2 := func() fyne.CanvasObject {
		return widget.NewLabel("The Stack")
	}
	bindingfunc2 := func(i binding.DataItem, o fyne.CanvasObject) {
		o.(*widget.Label).Bind(i.(binding.String))
	}
	shorterListWidget := widget.NewListWithData(shorterStackStringSlice, newlabelwidgetfunc2, bindingfunc2)

	content := container.NewHBox(listwidget, shorterListWidget)

	// w.SetContent(listwidget) from first run just using listwidget as content
	// w.Resize(fyne.Size{300, 470}) based on first run just using listwidget as content
	w.SetContent(content)
	w.Resize(fyne.Size{700, 470})
	w.ShowAndRun()
}
