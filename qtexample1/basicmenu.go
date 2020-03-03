package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include "mainwindow.h"
MainWindow::MainWindow()
{
   setWindowTitle("SRM System");
   setFixedSize(500, 500);
   QPixmap newIcon("new.png");
   QPixmap openIcon("open.png");
   QPixmap closeIcon("close.png");
   // Setup File Menu
   fileMenu = menuBar()->addMenu("&File");
   quitAction = new QAction(closeIcon, "Quit", this);
   quitAction->setShortcuts(QKeySequence::Quit);
   newAction = new QAction(newIcon, "&New", this);
   newAction->setShortcut(QKeySequence(Qt::CTRL + Qt::Key_C));
   openAction = new QAction(openIcon, "&New", this);
   openAction->setShortcut(QKeySequence(Qt::CTRL + Qt::Key_O));
   fileMenu->addAction(newAction);
   fileMenu->addAction(openAction);
   fileMenu->addSeparator();
   fileMenu->addAction(quitAction);
   helpMenu = menuBar()->addMenu("Help");
   aboutAction = new QAction("About", this);
   aboutAction->setShortcut(QKeySequence(Qt::CTRL + Qt::Key_H));
   helpMenu->addAction(aboutAction);
   // Setup Signals and Slots
   connect(quitAction, &QAction::triggered, this, &QApplication::quit);
}


*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("SRM System Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetFixedSize2(500, 500) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int)
	newIcon := gui.QIcon_FromTheme2("document-new", gui.NewQIcon5("new.png"))
	openIcon := gui.QIcon_FromTheme2("document-open", gui.NewQIcon5("open.png"))
	closeIcon := gui.QIcon_FromTheme2("document-close", gui.NewQIcon5("close.png"))

	// set up file menu

	fileMenu := widgets.NewQMenuBar(window)

	quitAction := widgets.NewQAction3(closeIcon, "Quit", window)
	quitAction.SetShortcuts2(gui.QKeySequence__Quit)
	quitclicked := func () {
		app.Quit()
	}
	quitAction.ConnectTrigger(quitclicked)

	newAction := widgets.NewQAction3(openIcon,"&New", window)
	newAction.SetShortcuts2(gui.QKeySequence__New)

	openAction := widgets.NewQAction3(openIcon,"&Open", window)
	openAction.SetShortcuts2(gui.QKeySequence__Open)

	fileMenu.AddAction2("New", newAction,"newAction")
	fileMenu.AddAction2("Open", openAction, "openAction")

	fileMenu.AddSeparator()
	fileMenu.AddAction2("Quit", quitAction, "quit")

	menubar := window.MenuBar()
	helpmenu := menubar.AddMenu2("Help")

	aboutAction := widgets.NewQAction2("About", window)
	aboutAction.SetShortcuts2(gui.QKeySequence__WhatsThis)
	
	helpmenu.AddActions(aboutAction)



/*
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
*/

//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
