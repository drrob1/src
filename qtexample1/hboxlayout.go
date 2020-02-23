package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include <QApplication>
#include <QHBoxLayout>
#include <QPushButton>
#include <QLineEdit>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QLineEdit *urlLineEdit= new QLineEdit;
   QPushButton *exportButton = new QPushButton("Export");
   urlLineEdit->setPlaceholderText("Enter Url to export. Eg, http://yourdomain.com/items");
   urlLineEdit->setFixedWidth(400);
   QHBoxLayout *layout = new QHBoxLayout;
   layout->addWidget(urlLineEdit);
   layout->addWidget(exportButton);
   window->setLayout(layout);
   window->show();
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
	window.SetWindowTitle("QHBox Layout Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(500, 100) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	centralwidget := widgets.NewQWidget(nil, 0)
	//hboxlayout := widgets.NewQHBoxLayout2(centralwidget)
	hboxlayout := widgets.NewQHBoxLayout()  // this more closely matches the example code above
    centralwidget.SetLayout(hboxlayout)  // this line gives an error of attempting to set layout which already has a layout
	window.SetCentralWidget(centralwidget)

	urlLineEdit := widgets.NewQLineEdit(centralwidget)
	urlLineEdit.SetPlaceholderText("Enter Url to export")
	urlLineEdit.SetFixedWidth(200)
	centralwidget.Layout().AddWidget(urlLineEdit)

	exportbutton := widgets.NewQPushButton2("Export", centralwidget)
	centralwidget.Layout().AddWidget(exportbutton)

	window.Show()

	// Execute app
	app.Exec()
}
