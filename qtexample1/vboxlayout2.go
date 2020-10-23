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
*/

// doesn't work as of Feb 22, 2020.
func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)      //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("QVBox Layout Example") // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(400, 400)              //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	vboxlayout := widgets.NewQVBoxLayout2(window)
	mainwidget := widgets.NewQWidget(nil, 0)
	mainwidget.SetLayout(vboxlayout)

	usernameLabel := widgets.NewQLabel2("Username", nil, 0)
	pwdLabel := widgets.NewQLabel2("Password", nil, 0)
	usernameLineEdit := widgets.NewQLineEdit(nil)
	pwdLineEdit := widgets.NewQLineEdit(nil)
	pwdLineEdit.SetEchoMode(widgets.QLineEdit__Password)
	pwdLineEdit.SetPlaceholderText("Enter password")

	loginbutton := widgets.NewQPushButton2("Login", window)
	registerbutton := widgets.NewQPushButton2("Register", window)

	vboxlayout.AddWidget(usernameLabel, 0, 0)
	vboxlayout.AddWidget(usernameLineEdit, 0, 0)
	vboxlayout.AddWidget(pwdLabel, 0, 0)
	vboxlayout.AddWidget(pwdLineEdit, 0, 0)
	vboxlayout.AddWidget(loginbutton, 0, 0)
	vboxlayout.AddWidget(registerbutton, 0, 0)

	// window.SetCentralWidget(vboxlayout)

	window.Show()

	// Execute app
	app.Exec()
}
