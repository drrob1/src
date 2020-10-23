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

// this one will be closer to the book, using nil as parent and see if it works without a mainwidget as central widget.
// Turns out that it does not work when I use nil, but does not give any errors.
// And it doesn't work without a mainwidget, even when I create the widgets with window as the parent.
func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)     //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("Grid Layout Example") // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(300, 300)             //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	//mainwidget := widgets.NewQWidget(nil, 0)

	nameLabel := widgets.NewQLabel2("Open Happiness", window, 0)
	firstNameLineEdit := widgets.NewQLineEdit(window)
	lastNameLineEdit := widgets.NewQLineEdit(window)
	ageSpinBox := widgets.NewQSpinBox(window)
	ageSpinBox.SetRange(1, 100)

	employmentStatusComboBox := widgets.NewQComboBox(window)
	employmentStatusStringSlice := []string{"Unemployed", "Employed", "NA"}
	employmentStatusComboBox.AddItems(employmentStatusStringSlice)

	// Create grid layout
	layout := widgets.NewQGridLayout(window)
	layout.AddWidget2(nameLabel, 0, 0, core.Qt__AlignLeft)
	layout.AddWidget2(firstNameLineEdit, 0, 1, core.Qt__AlignLeft)
	layout.AddWidget2(lastNameLineEdit, 0, 2, core.Qt__AlignLeft)
	layout.AddWidget2(ageSpinBox, 1, 0, core.Qt__AlignLeft)
	layout.AddWidget2(employmentStatusComboBox, 1, 1, core.Qt__AlignLeft)

	//	window.SetCentralWidget(mainwidget)
	window.SetLayout(layout)
	window.Show()

	// Execute app
	app.Exec()
}
