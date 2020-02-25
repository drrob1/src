package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include <QApplication>
#include <QPushButton>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QPushButton *quitButton = new QPushButton("Quit");
   QObject::connect(quitButton, SIGNAL(clicked()),
           &app, SLOT(quit()));
   quitButton->show();
return app.exec();
}

from github.com/therecipe/examples/ ... / advanced.  I'm not including the func def's from this file into here.
func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(250, 200)
	window.SetWindowTitle("listview Example")
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(widget)
	listview := widgets.NewQListView(nil)
	model := NewCustomListModel(nil)
	listview.SetModel(model)
	widget.Layout().AddWidget(listview)
	remove := widgets.NewQPushButton2("remove last item", nil)
	remove.ConnectClicked(func(bool) {
		model.Remove()
	})
	widget.Layout().AddWidget(remove)
	add := widgets.NewQPushButton2("add new item", nil)
	add.ConnectClicked(func(bool) {
		model.Add(ListItem{"john", "doe"})
	})
	widget.Layout().AddWidget(add)
	edit := widgets.NewQPushButton2("edit last item", nil)
	edit.ConnectClicked(func(bool) {
		model.Edit("bob", "omb")
	})
	widget.Layout().AddWidget(edit)
	window.Show()
	app.Exec()
}

*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("QButton Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(200, 200) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	centralwidget := widgets.NewQWidget(nil, 0)
	centralwidget.SetLayout(widgets.NewQVBoxLayout()) // from example code above
	window.SetCentralWidget(centralwidget)

	quitbutton := widgets.NewQPushButton2("quit", centralwidget)

	centralwidget.Layout().AddWidget(quitbutton)

	quitclicked := func (checked bool) {
		app.Quit()
	}
	quitbutton.ConnectClicked(quitclicked)

//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
