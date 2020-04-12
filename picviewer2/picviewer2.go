//              picviewer2.go
package main
/*
   REVISION HISTORY
   ======== =======
    7 Apr 20 -- Now called picviewer2.go.  I'm going to try the image reading trick I learned from the Qt example imageviewer.
    9 Apr 20 -- Will try to handle arrow keys.
   11 Apr 20 -- Won't handle arrow keys that I can get to work.  Will use N and B, I think.
   12 Apr 20 -- Now that the keys are working, I don't need a pushbutton.
                  And will make less dependent on globals.
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
	displayArea   *widgets.QWidget
	scene         *widgets.QGraphicsScene
	view          *widgets.QGraphicsView
	item          *widgets.QGraphicsPixmapItem
	mainApp       *widgets.QApplication
	imageFileName string
	picfiles      sort.StringSlice
	currImgIdx    int
	origImgIdx    int
	prevImgIdx    int
)

func imageViewer() *widgets.QWidget {
	displayArea = widgets.NewQWidget(nil, 0)
	scene = widgets.NewQGraphicsScene(displayArea)
	view = widgets.NewQGraphicsView(displayArea)

	var imageReader *gui.QImageReader

	imageReader.SetAutoTransform(true)

	imageReader = gui.NewQImageReader3(imageFileName, core.NewQByteArray2("", 0))

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
	var layout = widgets.NewQVBoxLayout()

	layout.AddWidget(view, 0, core.Qt__AlignCenter)
//	layout.AddWidget(button, 0, core.Qt__AlignCenter)

	displayArea.SetLayout(layout)  // I tried not using a layout, but displayArea does not have an AddItem method.

	// Must test combo keys before indiv keys, as indiv key test ignore the modifiers.
	// I discovered that testing N before Ctrl-N always found N and never ctrl-N.
	arrowEventclosure := func(ev *gui.QKeyEvent) {
		if false { // only keys without events will still call qmessagebox
			// do nothing, just so I can test this.
		} else if ev.Matches(gui.QKeySequence__New) { // ctrl-n
			//widgets.QMessageBox_Information(nil, "key New", "Ctrl-N hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			i := nextPic(currImgIdx)
			currImgIdx = i
			displayImageByNumber(i)
		} else if ev.Matches(gui.QKeySequence__Quit) { // ctrl-q
			//widgets.QMessageBox_Information(nil, "quit Key", "Ctrl-q hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			mainApp.Quit()
		} else if ev.Matches(gui.QKeySequence__Cancel) { // ESC
			//widgets.QMessageBox_Information(nil, "cancel", "cancel <Esc> hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			mainApp.Quit()
		} else if ev.Matches(gui.QKeySequence__Open) { // ctrl-oh
			//widgets.QMessageBox_Information(nil, "key Open", "Ctrl-O key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			origImgIdx, currImgIdx = currImgIdx, origImgIdx
			displayImageByNumber(currImgIdx)
		} else if ev.Matches(gui.QKeySequence__HelpContents) {
			widgets.QMessageBox_Information(nil, "key Help", "F1 key kit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		} else if ev.Key() == int(core.Qt__Key_B) {
			//widgets.QMessageBox_Information(nil, "B key", "B key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			i := prevPic(currImgIdx)
			currImgIdx = i
			displayImageByNumber(i)
		} else if ev.Key() == int(core.Qt__Key_N) {
			//widgets.QMessageBox_Information(nil, "N key", "N key hit", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			i := nextPic(currImgIdx)
			currImgIdx = i
			displayImageByNumber(i)
		} else if ev.Key() == int(core.Qt__Key_Q) {
			mainApp.Quit()
		}
	}
	displayArea.ConnectKeyPressEvent(arrowEventclosure)

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
func prevPic(i int)(int) {
	j := i
	if j > 0 {
		j--
	}
	//fmt.Println(" In NexPic.  prevImgIdx=", prevImgIdx, ", and currImgIdx=", currImgIdx)
	return j
}

// ------------------------- DisplayImageByNumber ----------------------
func displayImageByNumber(i int) {
	currImgIdx = i
	imageFileName = picfiles[currImgIdx]
	fmt.Println(" in displayImageByNumber.  currImgIdx=", currImgIdx, ", imageFileName=", imageFileName)
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