//              picviewer3.go
package main

/*
   REVISION HISTORY
   ======== =======
   12 Apr 20 -- Now picviewer3.go.  I want change it completely so that I only have one func imageViewer.
                  But I want to not lose the other version that mostly works, except that the view size does not match the displayArea on supsequent reads.
   17 Apr 20 -- func (*QGraphicsView) SetDragMode(mode QGraphicsView__DragMode) --> [ __NoDrag | __ScrollHandDrag | __RubberBandDrag ]
   18 Apr 20 -- func (*QGraphicsView) Scale(sx, sy float64) which is working.
                 and now trying to see what SetResizeAnchor does.
*/

import (
	"fmt"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
	"io/ioutil"
	"log"
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
)
var (
	imageReader *gui.QImageReader
	//displayArea *widgets.QMainWindow // was a QWidget, but I want something else, but it didn't work.  The scene was distorted.
	displayArea *widgets.QWidget
	scene       *widgets.QGraphicsScene
	view        *widgets.QGraphicsView
	item        *widgets.QGraphicsPixmapItem
	layout      *widgets.QVBoxLayout
	//maxqsize    *core.QSize
)

const maxQWidth = 1440
const maxQHeight = 1050
const maxWidth = 1400
const maxHeight = 1024

func imageViewer() *widgets.QWidget {

	//var imageReader *gui.QImageReader
	imageReader = gui.NewQImageReader3(imageFileName, core.NewQByteArray2("", 0)) // format is set by core.NewQByteArray2
	size := imageReader.Size()
	width := size.Width()
	fwidth := float64(width)
	height := size.Height()
	fheight := float64(height)
	fmt.Println()
	fmt.Println(" New Image: imagereader width=", width, ", height=", height, ", fwidth, fheight =", fwidth, fheight)

	firstTimeThru := false
	if displayArea == nil { // must be first pass thru this rtn.
		displayArea = widgets.NewQWidget(nil, 0)
		firstTimeThru = true
		displayArea.SetFixedSize2(maxQWidth, maxQHeight) // displayArea has to be slightly bigger than the pixmap.
		//scene = widgets.NewQGraphicsScene3(0, 0, fwidth, fheight, displayArea)
		scene = widgets.NewQGraphicsScene(displayArea)
		view = widgets.NewQGraphicsView(displayArea)
	} else {
		scene.Clear()
		//view.Close()
	}

	imageReader.SetAutoTransform(true)
	//	imageReader.SetAutoDetectImageFormat(true)  this is on by default.  Format refers to image file format.

	// test to see if we are dealing with animated GIF
	fmt.Println("Animated GIF : ", imageReader.SupportsAnimation())
	var pixmap = gui.NewQPixmap3(maxWidth, maxHeight)
	fmt.Printf(" pixmap just created to be %d wide and %d high. \n", pixmap.Width(), pixmap.Height())

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

		//var pixmap = gui.NewQPixmap3(maxWidth, maxHeight)
		pixmap.DestroyQPixmap()
		//pixmap = gui.NewQPixmap3(maxWidth, maxHeight)
		//pixmap.Load(imageFileName, "", core.Qt__AutoColor)
		pixmap = gui.NewQPixmap5(imageFileName, "", core.Qt__AutoColor) // this was changed fromNewQPixmap3 in before I had to redo Qt and therecipe.

		//size := pixmap.Size()
		width = pixmap.Width()
		height = pixmap.Height()
		fmt.Printf(" Pixmap image %s is %d wide and %d high \n", imageFileName, width, height)

		item = widgets.NewQGraphicsPixmapItem2(pixmap, nil)

		scene.AddItem(item)
	}

	view.SetScene(scene)
	//view.SetDragMode(widgets.QGraphicsView__RubberBandDrag)
	view.SetDragMode(widgets.QGraphicsView__ScrollHandDrag)
	view.EnsureVisible2(0, 0, fwidth, fheight, 10, 10)
	displayArea.SetWindowTitle(imageFileName)
	/*
		//create a button and connect the clicked signal.  Or not.
		var button = widgets.NewQPushButton2("Quit", nil)

		btnclicked := func(flag bool) {
			widgets.QApplication_Beep()
			//widgets.QMessageBox_Information(nil, "OK", "You clicked quit button!", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			mainApp.Quit()
		}
		button.ConnectClicked(btnclicked)
	*/
	if firstTimeThru {
		layout = widgets.NewQVBoxLayout()

		layout.AddWidget(view, 0, core.Qt__AlignCenter)
		//layout.AddWidget(button, 0, core.Qt__AlignCenter)

		displayArea.SetLayout(layout) // I tried not using a layout, but displayArea does not have an AddItem method.

		// Must test for combo keys before indiv keys, as indiv key test ignore the modifiers.
		// I discovered that testing N before Ctrl-N always found N and never ctrl-N.
		arrowEventclosure := func(ev *gui.QKeyEvent) {
			if false { // only keys without events will still call qmessagebox
				// do nothing, just so I can test this.
			} else if ev.Matches(gui.QKeySequence__New) { // ctrl-n
				//widgets.QMessageBox_Information(nil, "key New", "Ctrl-N hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				if currImgIdx < len(picfiles)-1 {
					currImgIdx++
				}
				imageFileName = picfiles[currImgIdx]
				imageViewer()
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
				view.SetResizeAnchor(widgets.QGraphicsView__AnchorViewCenter)
				view.Scale(1.25, 1.25) // factor is 5/4
				imageViewer().Show()
			} else if ev.Matches(gui.QKeySequence__ZoomOut) {
				//widgets.QMessageBox_Information(nil, "zoom out key", "zoom out key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				view.SetResizeAnchor(widgets.QGraphicsView__AnchorViewCenter)
				view.Scale(0.8, 0.8) // factor is 4/5
				imageViewer().Show()
			} else if ev.Matches(gui.QKeySequence__HelpContents) {
				widgets.QMessageBox_Information(nil, "key Help", "F1 key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			} else if ev.Key() == int(core.Qt__Key_B) {
				//widgets.QMessageBox_Information(nil, "B key", "B key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				if currImgIdx > 0 {
					currImgIdx--
				}
				imageFileName = picfiles[currImgIdx]
				imageViewer()
			} else if ev.Key() == int(core.Qt__Key_N) {
				//widgets.QMessageBox_Information(nil, "N key", "N key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				if currImgIdx < len(picfiles)-1 {
					currImgIdx++
				}
				imageFileName = picfiles[currImgIdx]
				imageViewer()
			} else if ev.Key() == int(core.Qt__Key_Q) {
				mainApp.Quit()
			} else if ev.Key() == int(core.Qt__Key_Equal) {
				//widgets.QMessageBox_Information(nil, "= key", "equal key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				view.SetResizeAnchor(widgets.QGraphicsView__AnchorViewCenter)
				view.Scale(1.25, 1.25) // factor is 5/4
				imageViewer().Show()
			} else if ev.Key() == int(core.Qt__Key_Minus) {
				//widgets.QMessageBox_Information(nil, "- key", "minus key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				view.SetResizeAnchor(widgets.QGraphicsView__AnchorViewCenter)
				view.Scale(0.8, 0.8) // factor is 4/5
				imageViewer().Show()
			}
		}
		displayArea.ConnectKeyPressEvent(arrowEventclosure)
	}

	return displayArea
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage : %s <image file>\n", os.Args[0])
		os.Exit(0)
	}

	imageFileName = os.Args[1]
	maxqsize = core.NewQSize2(maxWidth, maxHeight)

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

//	arrowEvent := gui.NewQKeyEvent(core.QEvent__KeyPress, int(core.Qt__Key_Up), core.Qt__NoModifier, "", false, 0)
//	displayArea.ConnectKeyPressEvent(func(ev *gui.QKeyEvent) {  This doesn't work.
//		widgets.QMessageBox_Information(nil, "OK", "Up arrow key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
//	})(arrowEvent)

/*
// -------------------------- NextPic --------------------------------
func nextPic(i int) int {
	j := i
	if j < len(picfiles)-1 {
		j++
	}
	//fmt.Println(" In NexPic.  prevImgIdx=", prevImgIdx, ", and currImgIdx=", currImgIdx)
	return j
}

// ------------------------- PrevPic -------------------------------
func prevPic(i int) int {
	j := i
	if j > 0 {
		j--
	}
	//fmt.Println(" In NexPic.  prevImgIdx=", prevImgIdx, ", and currImgIdx=", currImgIdx)
	return j
}

*/

/*
// ------------------------- DisplayImageByNumber ----------------------
func displayImageByNumber(i int) {
	currImgIdx = i
	imageFileName = picfiles[currImgIdx]
	fmt.Println(" in displayImageByNumber.  currImgIdx=", currImgIdx, ", imageFileName=", imageFileName)
	imageReader = gui.NewQImageReader3(imageFileName, core.NewQByteArray2("", 0)) // format is set by core.NewQByteArray2 and means image file format.
	imageReader.SetAutoTransform(true)
	var pic = gui.NewQPixmap5(imageFileName, "", core.Qt__AutoColor)
	scene.RemoveItem(item)
	item = widgets.NewQGraphicsPixmapItem2(pic, nil)
	width := pic.Width()
	height := pic.Height()
	var fwidth float64 = math.Trunc(float64(width) * 1.1)
	var fheight float64 = math.Trunc(float64(height) * 1.1)
	fmt.Printf(" displayImageByNumber %s is %d wide and %d high, goes to %g wide and %g high \n",
		imageFileName, width, height, fwidth, fheight)
	width1 := int(fwidth)
	if fwidth < 300 {
		width1 += 100
	}
	height1 := int(fheight)
	if fheight < 300 {
		height1 += 100
	}

	scene.AddItem(item)
	//fmt.Printf(" displayImageByNumber %s is %d wide and %d high \n", imageFileName, width, height)
	//displayArea.AdjustSize()  didn't do anything
	displayArea.Resize2(width1, height1) // slightly too small.
//	displayArea.SetContentsMargins(0,0,width,height)  Doen't do what I want.
    //displayArea.Scroll(-width/2, -height/2)  Doesn't do what I want, at all.
	displayArea.Show()
}

*/
