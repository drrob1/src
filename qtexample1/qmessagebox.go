package main

import (
	"github.com/therecipe/qt/widgets"
	//"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
	"log"
	"os"
	"fmt"
)

func main() {
	// Create application
	app := widgets.NewQApplication(len(os.Args), os.Args) // func NewQApplication(argc int, argv []string) *QApplication

	// Create main window
	window := widgets.NewQMainWindow(nil, 0)  //  func NewQMainWindow(parent QWidget_ITF, flags core.Qt__WindowType) *QMainWindow
	window.SetWindowTitle("QHBox Layout Example")  // func (ptr *QGraphicsWidget) SetWindowTitle(title string)
	window.SetMinimumSize2(500, 100) //  func (ptr *QWidget) SetMinimumSize2(minw int, minh int) {

	centralwidget := widgets.NewQWidget(nil, 0)

	answer := widgets.QMessageBox_Information(window, "test title", "test string", widgets.QMessageBox__Ok, widgets.QMessageBox__Cancel )

	log.Printf(" MessageBox answer is %v", answer)
	s := fmt.Sprintf(" sprintf messagebox answer is %v", answer)
	fmt.Fprintln(os.Stderr, s)
	fmt.Println(" first messagebox answer is", answer, "using fmt.Println before calling window.Show() or app.Exec()")
	widgets.QMessageBox_About(window, "About box first answer",s)

    answer = widgets.QMessageBox_Information(window, "2nd title", "2nd information text", widgets.QMessageBox__Cancel, widgets.QMessageBox__Ok)


	hboxlayout := widgets.NewQHBoxLayout()  // this more closely matches the example code above
    centralwidget.SetLayout(hboxlayout)  // this line gives an error of attempting to set layout which already has a layout
	window.SetCentralWidget(centralwidget)

	urlLineEdit := widgets.NewQLineEdit(centralwidget)
	urlLineEdit.SetPlaceholderText("Enter Url to export")
	urlLineEdit.SetFixedWidth(200)
	centralwidget.Layout().AddWidget(urlLineEdit)

	exportbutton := widgets.NewQPushButton2("Export", centralwidget)
	centralwidget.Layout().AddWidget(exportbutton)

	window.Show()

	// Execute app
	app.Exec()


	s1 := fmt.Sprintf(" 2nd sprintf messagebox answer is %v", answer)
	_, _ = fmt.Fprintln(os.Stderr, s1)
	fmt.Println(" 2nd message box answer is", answer, "using fmt.Println")

}
