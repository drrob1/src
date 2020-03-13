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
    	widgets.QMessageBox_About(window, "File New", "File New Menu option was selected")
		return
	}
	a.ConnectTriggered(func(checked bool) {
		filenewmenuoption()
		return
	})  // function to execute when option is triggered

	a.SetPriority(widgets.QAction__LowPriority)
	a.SetShortcuts2(gui.QKeySequence__New)

	qactionpointerslice = append(qactionpointerslice, a)

	b := fileMenu.AddAction2(openIcon, "&Open") // b has type *QAction
    fileopenmenuoption := func () {
    	// need a function here.  I'll make it a dummy function
    	widgets.QMessageBox_About(window, "File Open", "File Open Menu option was selected")
		return
	}
	b.ConnectTriggered(func(checked bool) {
		fileopenmenuoption()
		return
	}) // function to execute when option is triggered

	b.SetPriority(widgets.QAction__LowPriority)
	b.SetShortcuts2(gui.QKeySequence__Open)

	qactionpointerslice = append(qactionpointerslice, b)

	fileMenu.AddSeparator()

	e := fileMenu.AddAction2(closeIcon, "&Close")
	fileclosemenuoption := func() {
		widgets.QMessageBox_About(window, "File Close", "File Close menu option was selected")
		return
	}
	e.ConnectTriggered(func(checked bool) {
		fileclosemenuoption()
		return
	})
	e.SetPriority(widgets.QAction__LowPriority)
	e.SetShortcuts2(gui.QKeySequence__Close)
	qactionpointerslice = append(qactionpointerslice, e)


	quitIcon := gui.QIcon_FromTheme2("document-quit", gui.NewQIcon5("quit-512.png"))
	c := fileMenu.AddAction2(quitIcon, "&Quit")
	filequitmenuoption := func() {
		widgets.QMessageBox_About(window, "File Quit", "File Quit Menu option was selected")
		app.Quit()
		return
	}
	c.ConnectTriggered(func(checked bool) {
		filequitmenuoption()
		return
	})
	c.SetPriority(widgets.QAction__LowPriority)
	c.SetShortcuts2(gui.QKeySequence__Quit)
	qactionpointerslice = append(qactionpointerslice, c)

	helpIcon := gui.QIcon_FromTheme2("document-help", gui.NewQIcon5("help2.png"))
	d := fileMenu.AddAction2(helpIcon, "&Help")
	filehelpmenuoption := func() {
		widgets.QMessageBox_About(window, "Help", "File Help menu option was selected")
		return
	}
	d.ConnectTriggered(func(checked bool) {
		filehelpmenuoption()
		return
	})
	d.SetPriority(widgets.QAction__LowPriority)
	d.SetShortcuts2(gui.QKeySequence__HelpContents)
	qactionpointerslice = append(qactionpointerslice, d)

	aboutIcon := gui.QIcon_FromTheme2("document-about", gui.NewQIcon5("about.png"))
	f := fileMenu.AddAction2(aboutIcon, "&About")
	fileaboutmenuoption := func() {
		widgets.QMessageBox_About(window, "about", "File About menu option was selected")
		return
	}
	f.ConnectTriggered(func(checked bool) {
		fileaboutmenuoption()
		return
	})
	f.SetPriority(widgets.QAction__LowPriority)
	f.SetShortcuts2(gui.QKeySequence__WhatsThis)
	qactionpointerslice = append(qactionpointerslice, f)

// turns out that this is not needed
	//fileMenu.AddActions(qactionpointerslice)

//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
