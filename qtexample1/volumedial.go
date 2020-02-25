package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include <QApplication>
#include <QVBoxLayout>
#include <QLabel>
#include <QDial>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QVBoxLayout *layout = new QVBoxLayout;
   QLabel *volumeLabel = new QLabel("0");
   QDial *volumeDial= new QDial;
   layout->addWidget(volumeDial);
   layout->addWidget(volumeLabel);
   QObject::connect(volumeDial, SIGNAL(valueChanged(int)), volumeLabel,
   SLOT(setNum(int)));
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
	window.SetWindowTitle("volumedial Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(200, 200) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	centralwidget := widgets.NewQWidget(nil, 0)
	centralwidget.SetLayout(widgets.NewQVBoxLayout()) // from example code above
	window.SetCentralWidget(centralwidget)

	volumeLabel := widgets.NewQLabel2("0", centralwidget, 0)
	volumeDial := widgets.NewQDial(centralwidget)

	centralwidget.Layout().AddWidget(volumeDial)
	centralwidget.Layout().AddWidget(volumeLabel)

	slotsetnum := func (n int) {
		volumeLabel.SetNum(n)
	}

	volumeDial.ConnectValueChanged(slotsetnum)
	
//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
