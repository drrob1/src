//              picviewer2.go
package main

/*
   REVISION HISTORY
   ======== =======
    7 Apr 20 -- Now called picviewer2.go.  I'm going to try the image reading trick I learned from the Qt example imageviewer.
    9 Apr 20 -- Will try to handle arrow keys.
*/

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"os"
)

var (
	displayArea   *widgets.QWidget
	scene         *widgets.QGraphicsScene
	view          *widgets.QGraphicsView
	item          *widgets.QGraphicsPixmapItem
	mainApp       *widgets.QApplication
	imageFileName string
)

func imageViewer() *widgets.QWidget {
	displayArea = widgets.NewQWidget(nil, 0)
	scene = widgets.NewQGraphicsScene(nil)
	view = widgets.NewQGraphicsView(nil)

	var imageReader *gui.QImageReader

	imageReader.SetAutoTransform(true)

	imageReader = gui.NewQImageReader3(imageFileName, core.NewQByteArray2("", 0))

//	arrowEvent := gui.NewQKeyEvent(core.QEvent__KeyPress, int(core.Qt__Key_Up), core.Qt__NoModifier, "", false, 0)
// Must test combo keys before indiv keys, as indiv key test ignore the modifiers.
// I discovered that testing N before Ctrl-N always found N and never ctrl-N.
	arrowEventclosure := func(ev *gui.QKeyEvent) {
		if false {
			// do nothing, just so I can test this.
		}else if ev.Matches(gui.QKeySequence__New) {
			widgets.QMessageBox_Information(nil, "key New", "Ctrl-N kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		} else if ev.Key() == int(core.Qt__Key_B) {
			widgets.QMessageBox_Information(nil, "B key", "B key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		} else if ev.Key() == int(core.Qt__Key_N) {
			widgets.QMessageBox_Information(nil, "N key", "N key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		} else if ev.Matches(gui.QKeySequence__Open) {
			widgets.QMessageBox_Information(nil, "key Open", "Ctrl-O key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		} else if ev.Matches(gui.QKeySequence__HelpContents) {
			widgets.QMessageBox_Information(nil, "key Help", "F1 key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		}
	}
	displayArea.ConnectKeyPressEvent(arrowEventclosure)


//	displayArea.ConnectKeyPressEvent(func(ev *gui.QKeyEvent) {
//		widgets.QMessageBox_Information(nil, "OK", "Up arrow key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
//	})(arrowEvent)


	// test to see if we are dealing with animated GIF
	fmt.Println("Animated GIF : ", imageReader.SupportsAnimation())

	if imageReader.SupportsAnimation() {
		// instead of reading from file(disk) again, we take from memory
		// HOWEVER, this will cause segmentation violation error ! :(
		//var movie = gui.NewQMovieFromPointer(imageReader.Pointer())
		var movie = gui.NewQMovie3(imageFileName, core.NewQByteArray2("", 0), nil)

		// see http://stackoverflow.com/questions/5769766/qt-how-to-show-gifanimated-image-in-qgraphicspixmapitem
		var movieLabel = widgets.NewQLabel(nil, core.Qt__Widget)
		movieLabel.SetMovie(movie)
		movie.Start()
		scene.AddWidget(movieLabel, core.Qt__Widget)
	} else {

		var pixmap = gui.NewQPixmap5(imageFileName, "", core.Qt__AutoColor) // this was changed fromNewQPixmap3 in before I had to redo Qt and therecipe.
		//size := pixmap.Size()
		width := pixmap.Width()
		height := pixmap.Height()
		fmt.Printf(" Image from file %s is %d wide and %d high \n", imageFileName, width, height)

		item = widgets.NewQGraphicsPixmapItem2(pixmap, nil)

		scene.AddItem(item)
	}

	view.SetScene(scene)

	//create a button and connect the clicked signal
	var button = widgets.NewQPushButton2("Quit", nil)

	btnclicked := func(flag bool) {
		//os.Exit(0)

		widgets.QApplication_Beep()
		//widgets.QMessageBox_Information(nil, "OK", "You clicked quit button!", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)

		// errmm... proper way to quit Qt application
		// https://godoc.org/github.com/therecipe/qt/widgets#QApplication.Quit
		mainApp.Quit()
	}
	button.ConnectClicked(btnclicked)

	var layout = widgets.NewQVBoxLayout()

	layout.AddWidget(view, 0, core.Qt__AlignCenter)
	layout.AddWidget(button, 0, core.Qt__AlignCenter)

	displayArea.SetLayout(layout)

	return displayArea
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage : %s <image file>\n", os.Args[0])
		os.Exit(0)
	}

	imageFileName = os.Args[1]

	fmt.Println("Loading image : ", imageFileName)

	mainApp = widgets.NewQApplication(len(os.Args), os.Args)

	imageViewer().Show()

	widgets.QApplication_Exec()
	//      mainApp.exec()    // I wonder if this will work.}
}
