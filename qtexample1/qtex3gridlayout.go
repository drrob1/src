package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
// The QGridLayout is used to arrange widgets by specifying the number of rows and columns that will be filled up by multiple widgets.  A grid-like
// structure mimics a table in that it has rows and columns and widgets are inserted as cells where a row and column meet.

#include <QApplication>
#include <QPushButton>
#include <QGridLayout>
#include <QLineEdit>
#include <QDateTimeEdit>
#include <QSpinBox>
#include <QComboBox>
#include <QLabel>
#include <QStringList>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QLabel *nameLabel = new QLabel("Open Happiness");
   QLineEdit *firstNameLineEdit= new QLineEdit;
   QLineEdit *lastNameLineEdit= new QLineEdit;
   QSpinBox *ageSpinBox = new QSpinBox;
   ageSpinBox->setRange(1, 100);
   QComboBox *employmentStatusComboBox= new QComboBox;
   QStringList employmentStatus = {"Unemployed", "Employed", "NA"};
   employmentStatusComboBox->addItems(employmentStatus);
   QGridLayout *layout = new QGridLayout;
   layout->addWidget(nameLabel, 0, 0);
   layout->addWidget(firstNameLineEdit, 0, 1);
   layout->addWidget(lastNameLineEdit, 0, 2);
   layout->addWidget(ageSpinBox, 1, 0);
   layout->addWidget(employmentStatusComboBox, 1, 1,1,2);
   window->setLayout(layout);
   window->show();
return app.exec();
}

*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication


	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("Grid Layout Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(300, 300) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	mainwidget := widgets.NewQWidget(nil, 0)

	nameLabel := widgets.NewQLabel2("Open Happiness", mainwidget, 0)
	firstNameLineEdit := widgets.NewQLineEdit(mainwidget)
	lastNameLineEdit := widgets.NewQLineEdit(mainwidget)
	ageSpinBox := widgets.NewQSpinBox(mainwidget)
	ageSpinBox.SetRange(1,100)

	employmentStatusComboBox := widgets.NewQComboBox(mainwidget)
	employmentStatusStringSlice := []string{"Unemployed", "Employed", "NA"}
	employmentStatusComboBox.AddItems(employmentStatusStringSlice)

	// Create grid layout
	layout := widgets.NewQGridLayout(mainwidget)
	layout.AddWidget2(nameLabel, 0, 0, core.Qt__AlignLeft)
	layout.AddWidget2(firstNameLineEdit, 0, 1, core.Qt__AlignLeft)
	layout.AddWidget2(lastNameLineEdit, 0, 2, core.Qt__AlignLeft)
	layout.AddWidget2(ageSpinBox, 1, 0, core.Qt__AlignLeft)
	layout.AddWidget2(employmentStatusComboBox, 1, 1, core.Qt__AlignLeft)

	window.SetCentralWidget(mainwidget)
//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()


/*
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
*/

	// Execute app
	app.Exec()
}
