package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include <QApplication>
#include <QFormLayout>
#include <QPushButton>
#include <QLineEdit>
#include <QSpinBox>
#include <QComboBox>
#include <QStringList>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QLineEdit *firstNameLineEdit= new QLineEdit;
   QLineEdit *lastNameLineEdit= new QLineEdit;
   QSpinBox *ageSpingBox = new QSpinBox;
   QComboBox *employmentStatusComboBox= new QComboBox;
   QStringList employmentStatus = {"Unemployed", "Employed", "NA"};
   ageSpingBox->setRange(1, 100);
   employmentStatusComboBox->addItems(employmentStatus);
   QFormLayout *personalInfoformLayout = new QFormLayout;
   personalInfoformLayout->addRow("First Name:", firstNameLineEdit);
   personalInfoformLayout->addRow("Last Name:", lastNameLineEdit);
   personalInfoformLayout->addRow("Age", ageSpingBox);
   personalInfoformLayout->addRow("Employment Status", employmentStatusComboBox);
   window->setLayout(personalInfoformLayout);
   window->show();
return app.exec();
}
*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)      //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("QForm Layout Example") // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(400, 400)              //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	mainwidget := widgets.NewQWidget(nil, 0)

	//nameLabel := widgets.NewQLabel2("Open Happiness", mainwidget, 0)
	firstNameLineEdit := widgets.NewQLineEdit(mainwidget)
	lastNameLineEdit := widgets.NewQLineEdit(mainwidget)
	ageSpinBox := widgets.NewQSpinBox(mainwidget)
	ageSpinBox.SetRange(1, 100)

	employmentStatusComboBox := widgets.NewQComboBox(mainwidget)
	employmentStatusStringSlice := []string{"Unemployed", "Employed", "NA"}
	employmentStatusComboBox.AddItems(employmentStatusStringSlice)

	// Create QForm layout
	personnellInfoformLayout := widgets.NewQFormLayout(mainwidget)
	personnellInfoformLayout.AddRow3("First Name:", firstNameLineEdit)
	personnellInfoformLayout.AddRow3("Last Name:", lastNameLineEdit)
	personnellInfoformLayout.AddRow3("Age", ageSpinBox)
	personnellInfoformLayout.AddRow3("Employment Status", employmentStatusComboBox)

	window.SetCentralWidget(mainwidget)
	//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
