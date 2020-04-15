//              picviewer4.go
package main

/*
   REVISION HISTORY
   ======== =======
   12 Apr 20 -- Now picviewer3.go.  I want change it completely so that I only have one func imageViewer.
                  But I want to not lose the other version that mostly works, except that the view size does not match the displayArea on supsequent reads.
   14 Apr 20 -- Now picviewer4.go.  Now I want to base it on the QImageViewer example that uses QScrollArea and QLabel to display the pics.
                  And the Set methods have an implied clear operation.
*/

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

var (
	mainApp       *widgets.QApplication
	imageFileName string
	picfiles      sort.StringSlice
	currImgIdx    int
	origImgIdx    int
	prevImgIdx    int
	scaleFactor   float64 = 1  // running scalefactor
)
var (
	imageReader *gui.QImageReader
	//	displayArea   *widgets.QWidget
	//	scene         *widgets.QGraphicsScene
	//	view          *widgets.QGraphicsView
	//	item          *widgets.QGraphicsPixmapItem
	//	layout        *widgets.QVBoxLayout
	window *widgets.QMainWindow
	scrollArea   *widgets.QScrollArea
	displayLabel *widgets.QLabel
    pixmap       *gui.QPixmap
)

const maxWidth = 1440
const maxHeight = 960

func imageViewer() *widgets.QMainWindow {

	imageReader = gui.NewQImageReader3(imageFileName, core.NewQByteArray2("", 0)) // format is set by core.NewQByteArray2, and means image format.
	imageReader.SetAutoTransform(true)
	//	imageReader.SetAutoDetectImageFormat(true)  this is on by default.  Format refers to image file format.
	size := imageReader.Size()
	width := size.Width()
	fwidth := float64(width)
	height := size.Height()
	fheight := float64(height)
	fmt.Println(" imagereader width=", width, ", height=", height, ", fwidth, fheight -", fwidth, fheight)

	firstTimeThru := false
	if window == nil { // must be first pass thru this rtn.
		window = widgets.NewQMainWindow(nil, 0)
		scrollArea = widgets.NewQScrollArea(window)
		//window.SetMinimumSize2(1920, 1080)
		//window.SetBaseSize2(1440, 960)  this didn't do what I want.
		window.SetMinimumSize2(maxWidth, maxHeight)
		window.SetCentralWidget(scrollArea)
		firstTimeThru = true
		displayLabel = widgets.NewQLabel(scrollArea, core.Qt__Widget)
		scrollArea.SetWidget(displayLabel)
		scrollArea.SetWidgetResizable(true)
		pixmap = gui.NewQPixmap()
	} // else {
	//	scene.Clear()
	//	//view.Close()
	//} As there is an implied Clear, this explicit clear is not needed.

	// test to see if we are dealing with animated GIF
	fmt.Println("Animated GIF : ", imageReader.SupportsAnimation())

	if imageReader.SupportsAnimation() {
		// instead of reading from file(disk) again, we take from memory.  HOWEVER, this will cause segmentation violation error ! :(
		//var movie = gui.NewQMovieFromPointer(imageReader.Pointer())
		var movie = gui.NewQMovie3(imageFileName, core.NewQByteArray2("", 0), nil)

		// see http://stackoverflow.com/questions/5769766/qt-how-to-show-gifanimated-image-in-qgraphicspixmapitem
		var movieLabel = widgets.NewQLabel(nil, core.Qt__Widget)
		movieLabel.SetMovie(movie)
		movie.Start()
		//scene.AddWidget(movieLabel, core.Qt__Widget)
	} else {

		//var pixmap = gui.NewQPixmap5(imageFileName, "", core.Qt__AutoColor)  scaling isn't working, so let me try something else
		//var pixmap = gui.NewQPixmap()  Now a global.

		pixmap.Load(imageFileName,"", core.Qt__AutoColor)
		scaleFactor = 1

		//size := pixmap.Size()
		width := pixmap.Width()
		height := pixmap.Height()

		fmt.Printf(" Pixmap image %s is %d wide and %d high \n",
			imageFileName, width, height)

		displayLabel.SetPixmap(pixmap)
		window.SetWindowTitle(imageFileName)
	}

	if firstTimeThru {
		// Must test for combo keys before indiv keys, as indiv key test ignore the modifiers.
		// I discovered that testing N before Ctrl-N always found N and never ctrl-N.
		arrowEventclosure := func(ev *gui.QKeyEvent) {
			if false { // only keys without events will still call qmessagebox
				// do nothing, just so I can test this.
			} else if ev.Matches(gui.QKeySequence__New) { // ctrl-n
				//widgets.QMessageBox_Information(nil, "key New", "Ctrl-N hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				scaleFactor = 1
			} else if ev.Matches(gui.QKeySequence__Quit) { // ctrl-q
				//widgets.QMessageBox_Information(nil, "quit Key", "Ctrl-q hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				mainApp.Quit()
			} else if ev.Matches(gui.QKeySequence__Cancel) { // ESC
				//widgets.QMessageBox_Information(nil, "cancel", "cancel <Esc> hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				mainApp.Quit()
			} else if ev.Matches(gui.QKeySequence__Open) { // ctrl-oh
				//widgets.QMessageBox_Information(nil, "key Open", "Ctrl-O key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				origImgIdx, currImgIdx = currImgIdx, origImgIdx
				//scene.RemoveItem(item)
				imageFileName = picfiles[currImgIdx]
				imageViewer()
			} else if ev.Matches(gui.QKeySequence__HelpContents) {
				helpmsg := " n -- next image \n"
				helpmsg += " b -- prev image \n"
				helpmsg += " ctrl-o -- original image \n"
				helpmsg += " Esc, q, ctrl-q -- quit \n"
				helpmsg += " ctrl-n -- scalefactor = 1 \n"
				helpmsg += " zoom in -- by 1.25, or 5/4 \n"
				helpmsg += " zoom out -- by 0.8, or 4/5 \n"
				widgets.QMessageBox_Information(nil, "key Help", helpmsg, widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			} else if ev.Matches(gui.QKeySequence__ZoomIn) {
				//widgets.QMessageBox_Information(nil, "zoom in key", "zoom in key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				scaleFactor *= 1.25  // factor is 5/4
				imageViewer().Show()
			} else if ev.Matches(gui.QKeySequence__ZoomOut) {
				//widgets.QMessageBox_Information(nil, "zoom out key", "zoom out key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				scaleFactor *= 0.8  // factor is 4/5
				imageViewer().Show()
			} else if ev.Key() == int(core.Qt__Key_B) {
				//widgets.QMessageBox_Information(nil, "B key", "B key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				if currImgIdx > 0 {
					currImgIdx--
				}
				imageFileName = picfiles[currImgIdx]
				imageViewer().Show()
			} else if ev.Key() == int(core.Qt__Key_N) {
				//widgets.QMessageBox_Information(nil, "N key", "N key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				if currImgIdx < len(picfiles)-1 {
					currImgIdx++
				}
				imageFileName = picfiles[currImgIdx]
				imageViewer().Show()
			} else if ev.Key() == int(core.Qt__Key_Q) {
				mainApp.Quit()
			} else if ev.Key() == int(core.Qt__Key_Equal) {
				//widgets.QMessageBox_Information(nil, "= key", "equal key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				scaleFactor *= 1.25  // factor is 5/4
				zoomIn(scaleFactor)
			} else if ev.Key() == int(core.Qt__Key_Minus) {
				//widgets.QMessageBox_Information(nil, "- key", "minus key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				scaleFactor *= 0.8  // factor is 4/5
				zoomOut(scaleFactor)
			}
		}
		window.ConnectKeyPressEvent(arrowEventclosure)
	}

	return window
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

	// widgets.QApplication_Exec() doesn't work when placed here.  The code that follows looks like it doesn't get executed.

	workingdir, _ := os.Getwd()

	// populate the string slice of all picture filenames, and the index in this slice of the initial displayed image.
	files, err := ioutil.ReadDir(workingdir)
	if err != nil { // It seems that ReadDir itself stops when it gets an error of any kind, and I cannot change that.
		log.Println(err, "so calling my own MyReadDir.")
		files = MyReadDir(workingdir)
	}

	picfiles = make(sort.StringSlice, 0, len(files))
	for _, f := range files {
		if isPicFile(f.Name()) {
			picfiles = append(picfiles, f.Name())
		}
	}
	picfiles.Sort()
	currImgIdx = picfiles.Search(imageFileName)
	fmt.Println(" Current image index in the picfiles slice is", currImgIdx, "; there are", len(picfiles), "picture files in", workingdir)
	origImgIdx = currImgIdx

	widgets.QApplication_Exec()
	//      mainApp.exec()    // also works.}
} // end main

// ------------------------------- MyReadDir -----------------------------------
func MyReadDir(dir string) []os.FileInfo {

	dirname, err := os.Open(dir)
	//	dirname, err := os.OpenFile(dir, os.O_RDONLY,0777)
	if err != nil {
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return nil
	}

	fi := make([]os.FileInfo, 0, len(names))
	for _, s := range names {
		L, err := os.Lstat(s)
		if err != nil {
			log.Println(" Error from os.Lstat ", err)
			continue
		}
		fi = append(fi, L)
	}
	return fi
} // MyReadDir

// ---------------------------- isPicFile ------------------------------
func isPicFile(filename string) bool {
	picext := []string{".jpg", ".png", ".jpeg", ".gif", "xcf"}
	for _, ext := range picext {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// ----------------------------------- ZoomIn --------------------------------
func zoomIn(factor float64)  {
	width := pixmap.Width()
	height := pixmap.Height()

	scaledwidth := int(math.Trunc(factor*float64(width)))
	if scaledwidth > maxWidth {
		scaledwidth = maxWidth
	}
	scaledheight := int(math.Trunc(factor*float64(height)))
	if scaledheight > maxHeight {
		scaledheight = maxHeight
	}
	fmt.Printf(" From Zoomin: Pixmap image %s is %d wide and %d high, scalefactor=%g, scaled w x h= %d x %d \n",
		imageFileName, width, height, scaleFactor, scaledwidth, scaledheight)

	pixmap.Scaled2(scaledwidth, scaledheight, core.Qt__KeepAspectRatio, core.Qt__SmoothTransformation)
	displayLabel.SetPixmap(pixmap)
	displayLabel.Resize2(scaledwidth, scaledheight)
	window.Show()
} // end zoomIn

// ---------------------------------- ZoomOut --------------------------------------
func zoomOut(factor float64)  {
	width := pixmap.Width()
	height := pixmap.Height()

	scaledwidth := int(math.Trunc(factor*float64(width)))
	if scaledwidth > maxWidth {
		scaledwidth = maxWidth
	}
	scaledheight := int(math.Trunc(factor*float64(height)))
	if scaledheight > maxHeight {
		scaledheight = maxHeight
	}
	fmt.Printf(" From ZoomOut: Pixmap image %s is %d wide and %d high, scalefactor=%g, scaled w x h= %d x %d \n",
		imageFileName, width, height, scaleFactor, scaledwidth, scaledheight)

	pixmap.Scaled2(scaledwidth, scaledheight, core.Qt__KeepAspectRatio, core.Qt__SmoothTransformation)
	displayLabel.SetPixmap(pixmap)
	displayLabel.Resize2(scaledwidth, scaledheight)
	window.Show()
}

