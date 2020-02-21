package main

import (
	"github.com/therecipe/qt/widgets"
	"github.com/therecipe/qt/gui"
	"os"
)

/*
Widgets-getting started w/ Qt 5, first example I found in the book.  Its purpose is to show that a widget without a parent object becomes the main window.

#include <QApplication>
#include <QPushButton>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QPushButton myButton(QIcon("filesaveas.png"),"Push Me");
   myButton.setToolTip("Click this to turn back the hands of time");
   myButton.show();
return app.exec();
}


 */


// these are going to be annotated with the definitations now present in widgets.go that I d/l from github.com on 02/19/2020 9:38:34 PM
func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication  

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("Hello World Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(200, 200) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	// Create main layout
	layout := widgets.NewQVBoxLayout()  //  func NewQVBoxLayout() *QVBoxLayout

	// Create main widget and set the layout
	mainWidget := widgets.NewQWidget(nil, 0)  // func NewQWidget(parent QWidget_ITF, ff core.Qt__WindowType) *QWidget
	mainWidget.SetLayout(layout)  //  func (ptr *QWidget) SetLayout(layout QLayout_ITF)

	// Create a line edit and add it to the layout
	input := widgets.NewQLineEdit(mainWidget)    //  func NewQLineEdit(parent QWidget_ITF) *QLineEdit or func NewQLineEdit2(contents string, parent QWidget_ITF) *QLineEdit
	input.SetPlaceholderText("1. write something")  //   func (ptr *QLineEdit) SetPlaceholderText(vqs string)
	layout.AddWidget(input, 0, 0)  //  func (ptr *QBoxLayout) AddWidget(widget QWidget_ITF, stretch int, alignment core.Qt__AlignmentFlag)

	// create a QIcon
//	icon := gui.NewQIcon5("Nike.jpg")  this works
	icon := gui.NewQIcon5("Nike.tif")  // this works, too

	// Create a button and add it to the layout
//	button := widgets.NewQPushButton2("2. click me", mainWidget)  //  func NewQPushButton2(text string, parent QWidget_ITF) *QPushButton
	button := widgets.NewQPushButton3(icon,"2. click me", mainWidget)  //  func NewQPushButton3(icon gui.QIcon_ITF, text string, parent QWidget_ITF) *QPushButton
	button.SetToolTip("Click this to turn back the hands of time.")
	layout.AddWidget(button, 0, 0)  //  func (ptr *QBoxLayout) AddWidget(widget QWidget_ITF, stretch int, alignment core.Qt__AlignmentFlag)

	// Connect event for button using my preferred syntax
	onclicked := func(checked bool) {
		widgets.QMessageBox_Information(mainWidget, "QLineEdit widget", input.Text(), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	}
	button.ConnectClicked(onclicked)  //  func (ptr *QAbstractButton) ConnectClicked(f func(checked bool))

	// Set main widget as the central widget of the window
	window.SetCentralWidget(mainWidget)  //  func (ptr *QMainWindow) SetCentralWidget(widget QWidget_ITF)

	// Show the window
	window.Show()  //  

	// Execute app
	app.Exec()  //  
}
