package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
)

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)                   //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("Dialog Example translated into Go") // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(500, 500)                             //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int)

	windowIcon := gui.QIcon_FromTheme2("window-icon", gui.NewQIcon5("window_logo.png"))
	window.SetWindowIcon(windowIcon)

	newIcon := gui.QIcon_FromTheme2("new", gui.NewQIcon5("new.png"))
	openIcon := gui.QIcon_FromTheme2("open", gui.NewQIcon5("open.png"))
	closeIcon := gui.QIcon_FromTheme2("close", gui.NewQIcon5("close.png"))
	clearIcon := gui.QIcon_FromTheme2("clear", gui.NewQIcon5("clear.png"))
	deleteIcon := gui.QIcon_FromTheme2("delete", gui.NewQIcon5("delete.png"))

	centralwidget := widgets.NewQWidget(nil, 0)

	// table view
	appTable := widgets.NewQTableView(centralwidget)
	appTable.SetContextMenuPolicy(core.Qt__CustomContextMenu)
	appTable.HorizontalHeader().SetSectionResizeMode(widgets.QHeaderView__Stretch)

	// model := widgets.NewQTableView(appTable)   Now a bunch of other code doesn't work.  Nevermind.
	model := gui.NewQStandardItemModel2(1, 3, appTable)
	model.SetHorizontalHeaderItem(0, gui.NewQStandardItem2("Name"))
	model.SetHorizontalHeaderItem(1, gui.NewQStandardItem2("Date of Birth"))
	model.SetHorizontalHeaderItem(2, gui.NewQStandardItem2("Phone Number"))

	appTable.SetModel(model)
	firstItem := gui.NewQStandardItem2("G. Sohne")
	dateOfBirth := core.NewQDate3(1980, 1, 1)
	//seconditem := gui.NewQStandardItem2(dateOfBirth.ToString2(core.Qt__TextDate))  // I'll pick one of these
	//seconditem := gui.NewQStandardItem2(dateOfBirth.ToString2(core.Qt__ISODate))  // whichever I like the best.
	seconditem := gui.NewQStandardItem2(dateOfBirth.ToString2(core.Qt__LocalDate)) //
	thirditem := gui.NewQStandardItem2("123-456-7890")
	model.SetItem(0, 0, firstItem)
	model.SetItem(0, 1, seconditem)
	model.SetItem(0, 2, thirditem)

	// layouts

	window.SetCentralWidget(centralwidget)
	nameLabel := widgets.NewQLabel2("Name:", centralwidget, 0)
	DOBLabel := widgets.NewQLabel2("Date of Birth:", centralwidget, 0)
	phoneNumberLabel := widgets.NewQLabel2("Phone Number", centralwidget, 0)
	savePushButton := widgets.NewQPushButton2("Save", centralwidget)
	clearPushButton := widgets.NewQPushButton2("Clear All", centralwidget)
	nameLineEdit := widgets.NewQLineEdit(centralwidget)
	DOBEdit := widgets.NewQDateEdit2(core.NewQDate3(1980, 1, 1), centralwidget)
	phoneNumberLineEdit := widgets.NewQLineEdit(centralwidget)

	formLayout := widgets.NewQGridLayout(centralwidget)
	centralwidget.SetLayout(formLayout)
	formLayout.AddWidget2(nameLabel, 0, 0, core.Qt__AlignCenter)
	formLayout.AddWidget2(nameLineEdit, 0, 1, core.Qt__AlignCenter)
	formLayout.AddWidget2(DOBLabel, 1, 0, core.Qt__AlignCenter)
	formLayout.AddWidget2(DOBEdit, 1, 1, core.Qt__AlignCenter)
	formLayout.AddWidget2(phoneNumberLabel, 2, 0, core.Qt__AlignCenter)
	formLayout.AddWidget2(phoneNumberLineEdit, 2, 1, core.Qt__AlignCenter)

	buttonsLayout := widgets.NewQHBoxLayout2(centralwidget)
	centralwidget.SetLayout(buttonsLayout)

	buttonsLayout.AddStretch(1)
	buttonsLayout.AddWidget(savePushButton, 2, core.Qt__AlignLeft)
	buttonsLayout.AddWidget(clearPushButton, 2, core.Qt__AlignLeft)


	// set up menu bar, and maybe toolbar
	menubar := window.MenuBar()

	// set up file menu option
	fileMenu := menubar.AddMenu2("&File")
	a := fileMenu.AddAction2(newIcon, "&New") // a has type *QAction
	filenewmenuoption := func() {
		// need a function here.  I'll make it a dummy function
		widgets.QMessageBox_About(window, "File New", "File New Menu option was selected")
		return
	}
	a.ConnectTriggered(func(checked bool) {
		filenewmenuoption()
		return
	}) // function to execute when option is triggered

	a.SetPriority(widgets.QAction__LowPriority)
	a.SetShortcuts2(gui.QKeySequence__New)

	b := fileMenu.AddAction2(openIcon, "&Open") // b has type *QAction
	fileopenmenuoption := func() {
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

	fileMenu.AddSeparator()

	g := fileMenu.AddAction2(clearIcon, "clea&R")
	clearmenuoption := func() {
		widgets.QMessageBox_About(window, "Clear All", "Clear All menu option was selected")
		// return  I think this is not needed
	}
	g.ConnectTriggered(func(checked bool) {
		clearmenuoption()
	})
	g.SetPriority(widgets.QAction__LowPriority)

	h := fileMenu.AddAction2(deleteIcon, "&Delete")
	deletemenuoption := func() {
		widgets.QMessageBox_About(window, "Delete", "Delete an entry was selected")
		var ok bool
		//numofrows := model.RowCount(appTable)
		//rowId := widgets.QInputDialog_GetInt(window,"Delete One Row","Select row to delete",1,1, numofrows,1,&ok,core.Qt__Dialog)
		if ok {
		//	model.RemoveRow(rowId-1, nil)
		}
	}
	h.ConnectTriggered(func(checked bool) {
		deletemenuoption()
	})
	h.SetPriority(widgets.QAction__LowPriority)

	QuitMenu := menubar.AddMenu2("&Quit")
	quitIcon := gui.QIcon_FromTheme2("quit", gui.NewQIcon5("quit.png"))
	c := QuitMenu.AddAction2(quitIcon, "&Quit")
	filequitmenuoption := func() {
		//widgets.QMessageBox_About(window, "Quit", "Quit Menu option was selected")
		app.Quit()
		return
	}
	c.ConnectTriggered(func(checked bool) {
		filequitmenuoption()
		return
	})
	c.SetPriority(widgets.QAction__LowPriority)
	c.SetShortcuts2(gui.QKeySequence__Quit)

	HelpMenu := menubar.AddMenu2("&Help")
	helpIcon := gui.QIcon_FromTheme2("document-help", gui.NewQIcon5("help2.png"))
	d := HelpMenu.AddAction2(helpIcon, "&Help")
	filehelpmenuoption := func() {
		widgets.QMessageBox_About(window, "Help", "Help menu option was selected")
		return
	}
	d.ConnectTriggered(func(checked bool) {
		filehelpmenuoption()
		return
	})
	d.SetPriority(widgets.QAction__LowPriority)
	d.SetShortcuts2(gui.QKeySequence__HelpContents)

	aboutIcon := gui.QIcon_FromTheme2("document-about", gui.NewQIcon5("about.png"))
	f := HelpMenu.AddAction2(aboutIcon, "&About")
	fileaboutmenuoption := func() {
		widgets.QMessageBox_About(window, "About dialog.go", "system <p> &copy; Whatever")
		return
	}
	f.ConnectTriggered(func(checked bool) {
		fileaboutmenuoption()
		return
	})
	f.SetPriority(widgets.QAction__LowPriority)
	f.SetShortcuts2(gui.QKeySequence__WhatsThis)

	// toolbar
	toolbar := widgets.NewQToolBar2(window)
	//	toolbar.SetToolButtonStyle(core.Qt__ToolButtonTextOnly)
	//toolbar.SetToolButtonStyle(core.Qt__ToolButtonTextUnderIcon)
	//toolbar.SetToolButtonStyle(core.Qt__ToolButtonIconOnly)
	//toolbar.SetToolButtonStyle(core.Qt__ToolButtonFollowStyle)
	//toolbar.SetFloatable(true)  // this worked, but I actually had to move the toolbar to see all options

	//docNew := toolbar.AddAction2(newIcon, "New")
	docNew := toolbar.AddAction("New")
	docNew.ConnectTriggered(func(checked bool) {
		filenewmenuoption()
		// return
	})

	toolbar.AddSeparator()
	//docOpen := toolbar.AddAction2(openIcon, "Open")
	docOpen := toolbar.AddAction("Open")
	docOpen.ConnectTriggered(func(checked bool) {
		fileopenmenuoption()
		// return
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
		// return
	})

	toolbar.AddSeparator()
	//About := toolbar.AddAction2(aboutIcon, "About")
	About := toolbar.AddAction("About")
	About.ConnectTriggered(func(checked bool) {
		fileaboutmenuoption()
		// return
	})

	toolbar.AddSeparator()

	// Quit := toolbar.AddAction2(quitIcon, "Quit")
	Quit := toolbar.AddAction("Quit")
	Quit.ConnectTriggered(func(checked bool) {
		filequitmenuoption()
		// return  I don't think this is needed
	})

	del := toolbar.AddAction2(deleteIcon, "Del")
	del.ConnectTriggered(func(checked bool) {
		deletemenuoption()
	})

	clearall := toolbar.AddAction2(clearIcon, "Clear All")
	clearall.ConnectTriggered(func(checked bool) {
		clearmenuoption()
	})

	//centralwidget.Layout().AddLayout(formLayout, 1)
	//centralwidget.Layout().AddWidget(appTable)  I'm not sure this is needed
	//centralwidget.Layout().AddLayout(buttonsLayout, 1)

	clearfields := func() {
		nameLineEdit.Clear()
		DOBEdit.Clear()
		phoneNumberLineEdit.Clear()
	}

	savepushbuttonclicked := func(checked bool) {
		name := gui.NewQStandardItem2(nameLineEdit.Text())
		dob := gui.NewQStandardItem2(DOBEdit.Date().ToString2(core.Qt__LocalDate))
		phonenumber := gui.NewQStandardItem2(phoneNumberLineEdit.Text())

		rowitems := make([]*gui.QStandardItem,0,3)
		rowitems = append(rowitems, name.QStandardItem_PTR())
		rowitems = append(rowitems, dob.QStandardItem_PTR())
		rowitems = append(rowitems, phonenumber.QStandardItem_PTR())

		model.AppendRow(rowitems)
		clearfields()
		widgets.QMessageBox_About(window, "dialog example in Go", "record saved successfully")
	}
	savePushButton.ConnectClicked(savepushbuttonclicked)

    clearallRecords := func() {
    	status := widgets.QMessageBox_Question(window, "Delete All Records", "Are you sure about deleting all saved records?",
    		widgets.QMessageBox__No, widgets.QMessageBox__Yes)
    	if status == widgets.QMessageBox__Yes {
    		// rowcount := model->rowcount()
    		// model->removerows(0,rowcount)
		}
	}

	clearPushButton.ConnectClicked(func(checked bool) {
		clearallRecords()
	})




	window.Show()

	// Execute app
	app.Exec()
}
