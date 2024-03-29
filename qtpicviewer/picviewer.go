package main

/*
  REVISION HISTORY
  -------- -------
   3 Apr 20 -- Attempting to be able to rotate an image 90 deg by a push button press.

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
	itemrotated   *widgets.QGraphicsPixmapItem
	mainApp       *widgets.QApplication
	imageFileName string
)

func imageViewer() *widgets.QWidget {
	displayArea = widgets.NewQWidget(nil, 0)
	scene = widgets.NewQGraphicsScene(nil)
	view = widgets.NewQGraphicsView(nil)

	var imageReader = gui.NewQImageReader3(imageFileName, core.NewQByteArray2("", 0))
	//var angle float64 = 0
	imageReader.SetAutoTransform(true)

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

		var pixmap = gui.NewQPixmap5(imageFileName, "", core.Qt__AutoColor)
		item = widgets.NewQGraphicsPixmapItem2(pixmap, nil)
		size := pixmap.Size()
		width := size.Width()
		height := size.Height()
		halfwidth := float64(width) / 2
		halfheight := float64(height) / 2
		fmt.Println(" Picture width=", width, ", picture height=", height, ", half width and height are", halfwidth, halfheight)

		//scene.AddItem(item)
		scene.AddItem(item)
	}

	view.SetScene(scene)
	// view.SetAlignment(core.Qt__AlignCenter)  doesn't do anything.
	//view.CenterOn2(halfwidth, halfheight)

	//create a button and connect the clicked signal
	var quitbutton = widgets.NewQPushButton2("Quit", nil)
	//var rotatebutton = widgets.NewQPushButton2("Rotate", nil)

	quitbtnclicked := func(flag bool) {
		//os.Exit(0)

		widgets.QApplication_Beep()
		//widgets.QMessageBox_Information(nil, "OK", "You clicked quit button!", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)

		// errmm... proper way to quit Qt application
		// https://godoc.org/github.com/therecipe/qt/widgets#QApplication.Quit
		mainApp.Quit()
	}
	quitbutton.ConnectClicked(quitbtnclicked)
	/*
		rotateclicked := func(flag bool) {
			angle = angle + 90
			itemrotated.SetRotation(angle)

			view.SetScene(scene)
			// view.SetAlignment(core.Qt__AlignCenter)  doesn't work
			view.CenterOn2(halfwidth, halfheight)
		}
		rotatebutton.ConnectClicked(rotateclicked)
	*/
	var layout = widgets.NewQVBoxLayout()

	layout.AddWidget(view, 0, core.Qt__AlignCenter)
	layout.AddWidget(quitbutton, 0, core.Qt__AlignCenter)
	//	layout.AddWidget(rotatebutton, 0, core.Qt__AlignCenter)

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
