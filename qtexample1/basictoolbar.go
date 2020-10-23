package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
main.cpp
#include <QApplication>
#include "mainwindow.h"
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QCoreApplication::setAttribute(Qt::AA_DontUseNativeMenuBar); //
   MainWindow mainwindow;
   mainwindow.show();
return app.exec();
}

mainwindow.cpp
#include "mainwindow.h"
MainWindow::MainWindow()
{
   setWindowTitle("Form in Window");
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
   // Setup Tool bar menu
   toolbar = addToolBar("main toolbar");
   // toolbar->setMovable( false );
   newToolBarAction = toolbar->addAction(QIcon(newIcon), "New File");
   openToolBarAction = toolbar->addAction(QIcon(openIcon), "Open File");
   toolbar->addSeparator();
   closeToolBarAction = toolbar->addAction(QIcon(closeIcon), "Quit Application");
   // Setup Signals and Slots
   connect(quitAction, &QAction::triggered, this, &QApplication::quit);
   connect(closeToolBarAction, &QAction::triggered, this, &QApplication::quit);
}

*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0) //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("Toolbar Example") // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(500, 500)         //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int)

	//newIcon := gui.QIcon_FromTheme2("document-new", gui.NewQIcon5("new.png"))
	filenewmenuoption := func() {
		// need a function here.  I'll make it a dummy function
		widgets.QMessageBox_About(window, "File New", "File New Menu option was selected")
		return
	}

	//openIcon := gui.QIcon_FromTheme2("document-open", gui.NewQIcon5("open.png"))
	fileopenmenuoption := func() {
		// need a function here.  I'll make it a dummy function
		widgets.QMessageBox_About(window, "File Open", "File Open Menu option was selected")
		return
	}

	//closeIcon := gui.QIcon_FromTheme2("document-close", gui.NewQIcon5("close.png"))
	fileclosemenuoption := func() {
		widgets.QMessageBox_About(window, "File Close", "File Close menu option was selected")
		return
	}

	//helpIcon := gui.QIcon_FromTheme2("document-help", gui.NewQIcon5("help2.png"))
	filehelpmenuoption := func() {
		widgets.QMessageBox_About(window, "Help", "Help menu option was selected")
		return
	}

	//aboutIcon := gui.QIcon_FromTheme2("document-about", gui.NewQIcon5("about.png"))
	fileaboutmenuoption := func() {
		widgets.QMessageBox_About(window, "about", "About menu option was selected")
		return
	}

	//quitIcon := gui.QIcon_FromTheme2("document-quit", gui.NewQIcon5("quit-512.png"))
	filequitmenuoption := func() {
		widgets.QMessageBox_About(window, "File Quit", "Quit Menu option was selected")
		app.Quit()
		return
	}

	// toolbar stuff based on Hands on GUI programming in Go
	// The toolbar clobbers the menubar on screen.  I'm going to remove the SetToolButtonStyle, or play w/ it.
	// Removing SetToolButtonStyle didn't work.

	//toolbar := widgets.NewQToolBar("tools", window)
	toolbar := widgets.NewQToolBar2(window)
	toolbar.SetToolButtonStyle(core.Qt__ToolButtonTextOnly)
	//toolbar.SetToolButtonStyle(core.Qt__ToolButtonTextUnderIcon)
	//toolbar.SetToolButtonStyle(core.Qt__ToolButtonIconOnly)
	//toolbar.SetToolButtonStyle(core.Qt__ToolButtonFollowStyle)
	//toolbar.SetFloatable(true)  // this worked, but I actually had to move the toolbar to see all options
	toolbar.SetMovable(true) // this worked, but I actually had to move the toolbar to see all options.

	//docNew := toolbar.AddAction2(newIcon, "New")
	docNew := toolbar.AddAction("New")
	docNew.ConnectTriggered(func(checked bool) {
		filenewmenuoption()
		return
	})

	toolbar.AddSeparator()
	//docOpen := toolbar.AddAction2(openIcon, "Open")
	docOpen := toolbar.AddAction("Open")
	docOpen.ConnectTriggered(func(checked bool) {
		fileopenmenuoption()
		return
	})

	toolbar.AddSeparator()
	//docClose := toolbar.AddAction2(closeIcon, "Close")
	docClose := toolbar.AddAction("Close")
	docClose.ConnectTriggered(func(checked bool) {
		fileclosemenuoption()
	})

	toolbar.AddSeparator()
	//Help := toolbar.AddAction2(helpIcon, "Help")
	Help := toolbar.AddAction("Help")
	Help.ConnectTriggered(func(checked bool) {
		filehelpmenuoption()
		return
	})

	toolbar.AddSeparator()
	//About := toolbar.AddAction2(aboutIcon, "About")
	About := toolbar.AddAction("About")
	About.ConnectTriggered(func(checked bool) {
		fileaboutmenuoption()
		return
	})

	toolbar.AddSeparator()
	// Quit := toolbar.AddAction2(quitIcon, "Quit")
	Quit := toolbar.AddAction("Quit")
	Quit.ConnectTriggered(func(checked bool) {
		filequitmenuoption()
		return
	})

	//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
