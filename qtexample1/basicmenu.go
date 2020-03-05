package main

import (
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

	// set up menu bar, and maybe toolbar
	menubar := window.MenuBar()

	qactionpointerslice := make([]*widgets.QAction,0,5)
	// set up file menu option

	fileMenu := menubar.AddMenu2("&File")
    a := fileMenu.AddAction2(newIcon,"&New")  // a has type *QAction
    filenewmenuoption := func () {
    	// need a function here.  I'll make it a dummy function
		return
	}
	a.ConnectTriggered(func(checked bool) { filenewmenuoption() })  // function to execute when option is triggered
	qactionpointerslice = append(qactionpointerslice, a)

	a.SetPriority(widgets.QAction__LowPriority)
	a.SetShortcuts2(gui.QKeySequence__New)

    b := fileMenu.AddAction2(openIcon, "&Open") // b has type *QAction
    fileopenmenuoption := func () {
    	// need a function here.  I'll make it a dummy function
		return
	}
	b.ConnectTriggered(func(checked bool) { fileopenmenuoption() }) // function to execute when option is triggered
	qactionpointerslice = append(qactionpointerslice, b)
	b.SetPriority(widgets.QAction__LowPriority)
	b.SetShortcuts2(gui.QKeySequence__Open)
	fileMenu.AddActions(qactionpointerslice)

	// I stopped here.  I'm too tired to continue.  Wed Feb 4 9 pm
	openAction := widgets.NewQAction3(openIcon,"&Open", window)
	openAction.SetShortcuts2(gui.QKeySequence__Open)
	fileMenu.AddAction2(openIcon, "&Open")

	fileMenu.AddSeparator()
	fileMenu.AddAction2("Quit", quitAction, "quit")


	helpmenu := menubar.AddMenu2("Help")
	helpmenu.AddAction2(newIcon, "Help")


	quitAction := widgets.NewQAction3(closeIcon, "Quit", window)
	quitAction.SetShortcuts2(gui.QKeySequence__Quit)
	quitclicked := func () {
		app.Quit()
	}
	quitAction.ConnectTrigger(quitclicked)


	aboutMenu := menubar.AddMenu2("About")
	aboutAction := widgets.NewQAction2("About", window)
	aboutAction.SetShortcuts2(gui.QKeySequence__WhatsThis)
	aboutMenu.AddAction("About")






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
