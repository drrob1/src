package main

import (
	"github.com/therecipe/qt/widgets"
	"os"
)

/*
#include <QApplication>
#include <QVBoxLayout>
#include <QPushButton>
#include <QLabel>
#include <QLineEdit>
int main(int argc, char *argv[])
{
   QApplication app(argc, argv);
   QWidget *window = new QWidget;
   QLabel *label1 = new QLabel("Username");
   QLabel *label2 = new QLabel("Password");
   QLineEdit *usernameLineEdit = new QLineEdit;
   usernameLineEdit->setPlaceholderText("Enter your username");
   QLineEdit *passwordLineEdit = new QLineEdit;
   passwordLineEdit->setEchoMode(QLineEdit::Password);
   passwordLineEdit->setPlaceholderText("Enter your password");
   QPushButton *button1 = new QPushButton("&Login");
   QPushButton *button2 = new QPushButton("&Register");
   QVBoxLayout *layout = new QVBoxLayout;
   layout->addWidget(label1);
   layout->addWidget(usernameLineEdit);
   layout->addWidget(label2);
   layout->addWidget(passwordLineEdit);
   layout->addWidget(button1);
   layout->addWidget(button2);
   window->setLayout(layout);
   window->show();
return app.exec();
}

from github.com/therecipe/examples/ ... / advanced.  I'm not including the func def's from this file into here.
func main() {
	app := widgets.NewQApplication(len(os.Args), os.Args)
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(250, 200)
	window.SetWindowTitle("listview Example")
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(widget)
	listview := widgets.NewQListView(nil)
	model := NewCustomListModel(nil)
	listview.SetModel(model)
	widget.Layout().AddWidget(listview)
	remove := widgets.NewQPushButton2("remove last item", nil)
	remove.ConnectClicked(func(bool) {
		model.Remove()
	})
	widget.Layout().AddWidget(remove)
	add := widgets.NewQPushButton2("add new item", nil)
	add.ConnectClicked(func(bool) {
		model.Add(ListItem{"john", "doe"})
	})
	widget.Layout().AddWidget(add)
	edit := widgets.NewQPushButton2("edit last item", nil)
	edit.ConnectClicked(func(bool) {
		model.Edit("bob", "omb")
	})
	widget.Layout().AddWidget(edit)
	window.Show()
	app.Exec()
}

*/

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("QVBox Layout Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(200, 200) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	mainwidget := widgets.NewQWidget(nil, 0)
	mainwidget.SetLayout(widgets.NewQVBoxLayout()) // from example code above
	window.SetCentralWidget(mainwidget)

	usernameLabel := widgets.NewQLabel2("Username", mainwidget, 0)
	pwdLabel := widgets.NewQLabel2("Password", mainwidget, 0)
	usernameLineEdit := widgets.NewQLineEdit(mainwidget)
	pwdLineEdit := widgets.NewQLineEdit(mainwidget)
	pwdLineEdit.SetEchoMode(widgets.QLineEdit__Password)
	pwdLineEdit.SetPlaceholderText("Enter password")

	loginbutton := widgets.NewQPushButton2("Login", mainwidget)
	registerbutton := widgets.NewQPushButton2("Register", mainwidget)

	mainwidget.Layout().AddWidget(usernameLabel)
	mainwidget.Layout().AddWidget(usernameLineEdit)
	mainwidget.Layout().AddWidget(pwdLabel)
	mainwidget.Layout().AddWidget(pwdLineEdit)
	mainwidget.Layout().AddWidget(loginbutton)
	mainwidget.Layout().AddWidget(registerbutton)

//	window.SetLayout(vboxlayout)  this line gives an error of attempting to set Qlayout which already has a layout

	window.SetCentralWidget(mainwidget)
//	window.SetLayout(layout)  I'm getting an error that says attempting to set layout on QMainWindow which already has a layout
	window.Show()

	// Execute app
	app.Exec()
}
