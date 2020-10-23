// From Hands On GUI Application Development in Go
package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)

	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Hello World")

	widget := widgets.NewQWidget(window, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())

	window.SetCentralWidget(widget)

	label := widgets.NewQLabel2("Hello World!", window, 0)

	widget.Layout().AddWidget(label)

	button := widgets.NewQPushButton2("Quit", window)
	onclicked := func(bool) {
		app.QuitDefault()
	}
	button.ConnectClicked(onclicked)

	widget.Layout().AddWidget(button)

	window.Show()

	widgets.QApplication_Exec()
}
