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
#include <QSlider>
#include <QLCDNumber>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QVBoxLayout *layout = new QVBoxLayout;
   QDial *volumeDial= new QDial;
   QSlider *lengthSlider = new QSlider(Qt::Horizontal);
   QLCDNumber *volumeLCD = new QLCDNumber;
   volumeLCD->setPalette(Qt::red);
   lengthSlider->setTickPosition(QSlider::TicksAbove);   lengthSlider->setTickInterval(10);
   lengthSlider->setSingleStep(1);   lengthSlider->setMinimum(0);
   lengthSlider->setMaximum(100);
   volumeDial->setNotchesVisible(true);   volumeDial->setMinimum(0);   volumeDial->setMaximum(100);
   layout->addWidget(volumeDial);   layout->addWidget(lengthSlider);   layout->addWidget(volumeLCD);
   QObject::connect(volumeDial, SIGNAL(valueChanged(int)), volumeLCD,   SLOT(display(int)));
   QObject::connect(lengthSlider, SIGNAL(valueChanged(int)), volumeLCD, SLOT(display(int)));
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

	volumeslider := widgets.NewQSlider2(core.Qt__Horizontal, centralwidget)
	volumeslider.SetTickPosition(widgets.QSlider__TicksAbove)
	volumeslider.SetTickInterval(10) // interval betwen tick marks
	volumeslider.SetSingleStep(1)
	volumeslider.SetMinimum(0)
	volumeslider.SetMaximum(100)

	volumeDial := widgets.NewQDial(centralwidget)
	volumeDial.SetNotchesVisible(true)
	volumeDial.SetMinimum(0)
	volumeDial.SetMaximum(100)

	volumeLCD := widgets.NewQLCDNumber2(3,centralwidget)
	paletteRed := gui.NewQPalette3(core.Qt__red)
	volumeLCD.SetPalette(paletteRed)

	centralwidget.Layout().AddWidget(volumeDial)
	centralwidget.Layout().AddWidget(volumeslider)
	centralwidget.Layout().AddWidget(volumeLCD)

	LCDdisplaynum := func (n int) {
		volumeLCD.Display2(n)
	}

	volumeDial.ConnectValueChanged(LCDdisplaynum)
	volumeslider.ConnectValueChanged(LCDdisplaynum)

	window.Show()

	// Execute app
	app.Exec()
}
