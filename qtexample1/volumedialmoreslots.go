package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include <QApplication>
#include <QVBoxLayout>
#include <QLabel>
#include <QDial>
#include <QLCDNumber>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QVBoxLayout *layout = new QVBoxLayout;
   QLabel *volumeLabel = new QLabel("0");
   QDial *volumeDial= new QDial;
   QLCDNumber *volumeLCD = new QLCDNumber;
   volumeLCD->setPalette(Qt::red);
   volumeLabel->setAlignment(Qt::AlignHCenter);
   volumeDial->setNotchesVisible(true);
   volumeDial->setMinimum(0);
   volumeDial->setMaximum(100);
   layout->addWidget(volumeDial);
   layout->addWidget(volumeLabel);
   layout->addWidget(volumeLCD);
   QObject::connect(volumeDial, SIGNAL(valueChanged(int)), volumeLabel,
   SLOT(setNum(int)));
   QObject::connect(volumeDial, SIGNAL(valueChanged(int)), volumeLCD ,
   SLOT(display(int)));
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

	volumeLCD := widgets.NewQLCDNumber2(3,centralwidget)
	paletteRed := gui.NewQPalette3(core.Qt__red)
	volumeLCD.SetPalette(paletteRed)
	volumeLabel.SetAlignment(core.Qt__AlignCenter)
	volumeDial.SetNotchesVisible(true)
	volumeDial.SetMinimum(0)
	volumeDial.SetMaximum(100)

	centralwidget.Layout().AddWidget(volumeDial)
	centralwidget.Layout().AddWidget(volumeLabel)
	centralwidget.Layout().AddWidget(volumeLCD)

	labelsetnum := func (n int) {
		volumeLabel.SetNum(n)
	}

	LCDdisplaynum := func (n int) {
		volumeLCD.Display2(n)
	}

	volumeDial.ConnectValueChanged(labelsetnum)
	volumeDial.ConnectValueChanged(LCDdisplaynum)
	
//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
