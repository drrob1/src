package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include "mainwindow.h"
MainWindow::MainWindow() {
    setWindowTitle("RMS System");
    setFixedSize(500, 500);

    setWindowIcon(QIcon("window_logo.png"));

    createIcons();
    setupCoreWidgets();
    createMenuBar();
    createToolBar();

    centralWidgetLayout->addLayout(formLayout);
    centralWidgetLayout->addWidget(appTable);
    //centralWidgetLayout->addStretch();
    centralWidgetLayout->addLayout(buttonsLayout);

    mainWidget->setLayout(centralWidgetLayout);

    setCentralWidget(mainWidget);

    setupSignalsAndSlots();

}

void MainWindow::createIcons() {
    newIcon = QPixmap("new.png");
    openIcon = QPixmap("open.png");
    closeIcon = QPixmap("close.png");
    clearIcon = QPixmap("clear.png");
    deleteIcon = QPixmap("delete.png");
}

void MainWindow::setupCoreWidgets() {
    mainWidget = new QWidget();
    centralWidgetLayout = new QVBoxLayout();
    formLayout = new QGridLayout();
    buttonsLayout = new QHBoxLayout();

    nameLabel = new QLabel("Name:");
    dateOfBirthLabel= new QLabel("Date Of Birth:");
    phoneNumberLabel = new QLabel("Phone Number");
    savePushButton = new QPushButton("Save");
    clearPushButton = new QPushButton("Clear All");

    nameLineEdit = new QLineEdit();
    dateOfBirthEdit = new QDateEdit(QDate(1980, 1, 1));
    phoneNumberLineEdit = new QLineEdit();

    // TableView
    appTable = new QTableView();
    model = new QStandardItemModel(1, 3, this);
    appTable->setContextMenuPolicy(Qt::CustomContextMenu);
    appTable->horizontalHeader()->setSectionResizeMode(QHeaderView::Stretch); // Note

	model->setHorizontalHeaderItem(0, new QStandardItem(QString("Name")));
	model->setHorizontalHeaderItem(1, new QStandardItem(QString("Date of Birth")));
	model->setHorizontalHeaderItem(2, new QStandardItem(QString("Phone Number")));

	appTable->setModel(model);
	QStandardItem *firstItem = new QStandardItem(QString("G. Sohne"));
	QDate dateOfBirth(1980, 1, 1);
	QStandardItem *secondItem = new QStandardItem(dateOfBirth.toString());
	QStandardItem *thirdItem = new QStandardItem(QString("05443394858"));
	model->setItem(0,0,firstItem);
	model->setItem(0,1,secondItem);
	model->setItem(0,2,thirdItem);

	//Layouts
	formLayout->addWidget(nameLabel, 0, 0);
	formLayout->addWidget(nameLineEdit, 0, 1);
	formLayout->addWidget(dateOfBirthLabel, 1, 0);
	formLayout->addWidget(dateOfBirthEdit, 1, 1);
	formLayout->addWidget(phoneNumberLabel, 2, 0);
	formLayout->addWidget(phoneNumberLineEdit, 2, 1);

	buttonsLayout->addStretch();
	buttonsLayout->addWidget(savePushButton);
	buttonsLayout->addWidget(clearPushButton);
}

void MainWindow::createMenuBar() {
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
}

void MainWindow::createToolBar() {
	// Setup Tool bar menu
	toolbar = addToolBar("main toolbar");
	// toolbar->setMovable( false );

	newToolBarAction = toolbar->addAction(QIcon(newIcon), "New File");
	openToolBarAction = toolbar->addAction(QIcon(openIcon), "Open File");
	toolbar->addSeparator();
	clearToolBarAction = toolbar->addAction(QIcon(clearIcon), "Clear All");
	deleteOneEntryToolBarAction = toolbar->addAction(QIcon(deleteIcon), "Delete a record");
	closeToolBarAction = toolbar->addAction(QIcon(closeIcon), "Quit Application");
}

void MainWindow::setupSignalsAndSlots() {
	// Setup Signals and Slots
	connect(quitAction, &QAction::triggered, this, &QApplication::quit);
	connect(aboutAction, SIGNAL(triggered()), this, SLOT(aboutDialog()));
	connect(clearToolBarAction, SIGNAL(triggered()), this, SLOT(clearAllRecords()));
	connect(closeToolBarAction, &QAction::triggered, this, &QApplication::quit);
	connect(deleteOneEntryToolBarAction, SIGNAL(triggered()), this, SLOT(deleteSavedRecord()));
	connect(savePushButton, SIGNAL(clicked()), this, SLOT(saveButtonClicked()));
	connect(clearPushButton, SIGNAL(clicked()), this, SLOT(clearAllRecords()));
}

void MainWindow::deleteSavedRecord() {
	bool ok;
	int rowId = QInputDialog::getInt(this, tr("Select Row to delete"),
	tr("Please enter Row ID of record (Eg. 1)"),
	1, 1, model->rowCount(), 1, &ok );
	if (ok) {
		model->removeRow(rowId-1);
	}
}

void MainWindow::saveButtonClicked() {
	QStandardItem *name = new QStandardItem(nameLineEdit->text());
	QStandardItem *dob = new QStandardItem(dateOfBirthEdit->date().toString());
	QStandardItem *phoneNumber = new QStandardItem(phoneNumberLineEdit->text());

	model->appendRow({ name, dob, phoneNumber});
	clearFields();

	QMessageBox::information(this, tr("RMS System"), tr("Record saved successfully!"),
	QMessageBox::Ok|QMessageBox::Default,
	QMessageBox::NoButton, QMessageBox::NoButton);
}

void MainWindow::clearFields() {
	nameLineEdit->clear();
	phoneNumberLineEdit->setText("");
	QDate dateOfBirth(1980, 1, 1);
	dateOfBirthEdit->setDate(dateOfBirth);
}

void MainWindow::clearAllRecords() {

//   int status = QMessageBox::question( this, tr("Delete Records ?"),
//                                       tr("You are about to delete all saved records "
//                                          "<p>Are you sure you want to delete all records "),
//                                       QMessageBox::No|QMessageBox::Default, QMessageBox::No|QMessageBox::Escape, QMessageBox::NoButton);
//   if (status == QMessageBox::Yes) return model->clear();

	int status = QMessageBox::question(this, tr("Delete all Records ?"),
				tr("This operation will delete all saved records. "
				"<p>Do you want to remove all saved records ? "
				), tr("Yes, Delete all records"), tr("No !"), 		QString(), 1, 1);
	if (status == 0) {
		int rowCount = model->rowCount();
		model->removeRows(0, rowCount);
	}
}

void MainWindow::aboutDialog() {
	QMessageBox::about(this, "About RMS System",
	"RMS System 2.0"
	"<p>Copyright &copy; 2005 Inc."
	"This is a simple application to demonstrate the use of windows,"
	"tool bars, menus and dialog boxes");
}

*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("Dialog Example translated into Go")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetFixedSize2(500, 500) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int)

	windowIcon := gui.QIcon_FromTheme2("window-icon", gui.NewQIcon5("window_logo.png"))
	window.SetWindowIcon(windowIcon)

	newIcon := gui.QIcon_FromTheme2("new", gui.NewQIcon5("new.png"))
	openIcon := gui.QIcon_FromTheme2("open", gui.NewQIcon5("open.png"))
	closeIcon := gui.QIcon_FromTheme2("close", gui.NewQIcon5("close.png"))
	clearIcon := gui.QIcon_FromTheme2("clear", gui.NewQIcon5("clear.png"))
	deleteIcon := gui.QIcon_FromTheme2("delete", gui.NewQIcon5("delete.png"))

	centralwidget := widgets.NewQWidget(nil, 0)
	vboxLayout := widgets.NewQVBoxLayout()  // this more closely matches the example code above
	formLayout := widgets.NewQGridLayout2()
	buttonsLayout := widgets.NewQHBoxLayout()

	nameLabel := widgets.NewQLabel2("Name:", centralwidget, 0)
	DOBLabel := widgets.NewQLabel2("Date of Birth:", centralwidget, 0)
	phoneNumberLabel := widgets.NewQLabel2("Phone Number", centralwidget, 0)
	savePushButton :=  widgets.NewQPushButton2("Save", centralwidget)
	clearPushButton := widgets.NewQPushButton2("Clear All", centralwidget)
	nameLineEdit := widgets.NewQLineEdit(centralwidget)
	DOBEdit := widgets.NewQDateEdit2(core.NewQDate3(1980,1,1),centralwidget)
	phoneNumberLineEdit := widgets.NewQLineEdit(centralwidget)

    // table view
    appTable := widgets.NewQTableView(centralwidget)
    model := gui.NewQStandardItemModel2(1,3,centralwidget)
    appTable.SetContextMenuPolicy(core.Qt__CustomContextMenu)
    appTable.HorizontalHeader().SetSectionResizeMode(widgets.QHeaderView__Stretch)
    

	centralwidget.SetLayout(hboxlayout)  // this line gives an error of attempting to set layout which already has a layout
	window.SetCentralWidget(centralwidget)


	urlLineEdit.SetPlaceholderText("Enter Url to export")
	urlLineEdit.SetFixedWidth(200)
	centralwidget.Layout().AddWidget(urlLineEdit)


	centralwidget.Layout().AddWidget(exportbutton)

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


	//quitIcon := gui.QIcon_FromTheme2("document-quit", gui.NewQIcon5("quit-512.png"))
	c := fileMenu.AddAction2(windowIcon, "&Quit")
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
