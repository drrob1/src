// Hands on GUI programming book, first Qt example.  I'll add notes from the book as comments here.
package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)

	// Widget constructor usually takes 2 params, the parent widget and a flags param.  If there are more params, they will appear before these 2.
	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Hello World")

	widget := widgets.NewQWidget(window, 0)
	widget.SetLayout(widgets.NewQVBoxLayout()) // layout is set here

	window.SetCentralWidget(widget) // set window content

	label := widgets.NewQLabel2("Hello World!", window, 0) // equivalent to widgets.NewQLabel(window, 0).SetTitle(title)

	widget.Layout().AddWidget(label) // widget is added to layout set above.

	button := widgets.NewQPushButton2("Quit", window)
	buttonclicked := func(bool) {
		app.QuitDefault()
	}
	button.ConnectClicked(buttonclicked)

	widget.Layout().AddWidget(button)

	window.Show()
	widgets.QApplication_Exec() // I think this is equivalent to app.Exec() to start the event loop.
}
