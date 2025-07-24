package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

// From Go GUI with Fyne, Chap 4, by Andrew Williams, (C) Packtpub.
/*
REVISION HISTORY
-------- -------
 9 Aug 21 -- I realized that this will not be enhanced, as I went thru more of the book.  I'll have to enhance it myself.
             First, I'm changing the function constants to the version that's more readable to me.  That's working, but I had to
             import more parts of fyne.io than the unmodified version.
12 Aug 21 -- Now called img.go, so I can display 1 image.  I'll start here.
13 Aug 21 -- Now called imgfyne.go.  Same purpose as img.go, but so I can test non-fyne code there and fyne code here.
15 Aug 21 -- Copied back to img.go after the code works and displays a single image from the command line.
               Will use imgfyne to add the arrow key navigation and img display.
18 Aug 21 -- It works!
20 Aug 21 -- Adding a verbose switch to print the messages, and not print them unless that switch is used.
21 Aug 21 -- Adding Q and X to exit.
23 Aug 21 -- Added webp format, after talking w/ Andy Williams, author of the book on fyne I read.  He's scottish.
 4 Sep 21 -- Added -, + and = keys
 4 Sep 21 -- Now called img2, and I intend to use a more complex display, w/ a central image container and a bottom text label.
               This will need a border layout.
22 Sep 21 -- I figured out that I need to deal w/ shift states in the keys to get magnification
29 Sep 21 -- Added stickyFlag, sticky and 'z' zoom toggle.  When sticky is true, zoom factor is not cleared automatically.
               Copied from img.go and imga.go.  When looking at my code, Andy commented that I don't need to resize, that's handled
               automatically.  I have a follow up question: how to I magnify a region of an image?  Answer: crop it.
 4 Dec 21 -- Made the channels buffered, cleaned up some channel code, and added "v" command to turn on verbose mode.  Ported from img.go.
26 Dec 21 -- Experimenting w/ detecting mouse scroll wheel movement
16 Mar 22 -- Will only write using fmt.Print calls if verboseFlag is set.
26 Mar 22 -- Expanding to work when display directory is not current directory
21 Oct 22 -- Fixed bad use of format verb caught by golangci-lint.
21 Nov 22 -- Fixed some issues caught by static linter.
21 Aug 23 -- Made the -sticky flag default to on.  And added scaleFactor to the window title.
24 Aug 23 -- Will add the new code I wrote for img to here.
25 Aug 23 -- Will time how long it takes to create the slice of fileInfos, and then sort them.
 7 Sep 23 -- I want to add a reverse sort flag, so the first image is the oldest, and I want to allow the use of the scroll wheel.  This may take a bit.
20 Feb 25 -- Porting code from img.go to here, allowing manual rotation of an image using repeated hits of 'r' to rotate clockwise 90 deg, or '1', '2', or '3'.
			It's too late now; I'll do this tomorrow.
			Added the rotateAndLoadImage and imgImage procedures, modified keyTyped and loadTheImage.  Fetching the image names is done w/ one goroutine; this is fast enough.
22 Feb 25 -- Added '=' to mean set scaleFactor=1 and zero the rotatedTimes variable.
			And here in img2 I'm going to rotate differently.  I'm going to use the routine that takes an amount to rotate in degrees.  Just to see if this, too, works.
			It does.  And I noticed another way the code here is different; it doesn't use a channel to send keystrokes to a receiver.  IE, loading images is not concurrent.
			The only concurrent code here is making the initial slice of strings containing all the image names in the working directory.
			I combined keys in the keyTyped routine here but not in the others.
			I added AutoOrientation to the rotateAndLoadTheImage and then to loadTheImage
24 Jul 25 -- Added ability to save an image in its current size and degree of rotation.  Developed first in img.go.
*/

const LastModified = "July 24, 2025"
const textboxheight = 20

// const maxWidth = 1800 // actual resolution is 1920 x 1080
// const maxHeight = 900 // actual resolution is 1920 x 1080

type ImageWidget struct {
	widget.BaseWidget
	img      *canvas.Image // flagged as unused, but I'm not changing it.
	x, y     int
	click    func(event *fyne.PointEvent)
	scrolled func(event *fyne.ScrollEvent)
	filename string // flagged as unused, but I'm not changing it.
}

func (iw *ImageWidget) Clicked(e *fyne.PointEvent) {
	if iw.click == nil {
		return
	}

	iw.click(e)
}

func (iw *ImageWidget) Scrolled(e *fyne.ScrollEvent) {
	if iw.scrolled == nil {
		return
	}

	iw.scrolled(e)
}

type imageRender struct {
	imgWdgt *ImageWidget
}

func (ir *imageRender) MinSize() fyne.Size { // This func is needed to define a renderer
	return fyne.NewSize(float32(ir.imgWdgt.x), float32(ir.imgWdgt.y))
}
func (ir *imageRender) Layout(_ fyne.Size) { // This func is needed to define a renderer
	ir.imgWdgt.Resize(ir.MinSize())
}
func (ir *imageRender) Destroy() { // This func is needed to define a renderer, but it can be empty
}
func (ir *imageRender) Refresh() { // This func is needed to define a renderer
	ir.imgWdgt.Refresh()
}
func (ir *imageRender) Objects() []fyne.CanvasObject { // This func is needed to define a renderer
	return []fyne.CanvasObject{ir.imgWdgt}
}

func (iw *ImageWidget) CreateRenderer() fyne.WidgetRenderer {
	rndrr := imageRender{imgWdgt: iw}
	return &rndrr
}

/*
func newImageWidget(fn string, clk func(e *fyne.PointEvent), scrl func(e *fyne.ScrollEvent)) *ImageWidget {
	e := ImageWidget{click: clk, scrolled: scrl}
	imgURI := storage.NewFileURI(fn)
	imgRead, err := storage.Reader(imgURI)
	if err != nil {
		panic(err)
	}
	defer imgRead.Close()

	imag, _, err := image.Decode(imgRead)
	if err != nil {
		panic(err)
	}

	bounds := imag.Bounds()
	e.y = bounds.Max.Y - bounds.Min.Y
	e.x = bounds.Max.X - bounds.Min.X

	e.img = canvas.NewImageFromImage(imag)

	e.ExtendBaseWidget(&e)
	return &e
}
*/
/*
//type ImageCanvas struct {
//	widget.BaseWidget
//	canvas.Image
//	click    func()
//	scrolled func()
//	filename string
//}

//func newImageCanvas(res fyne.Resource, clk func(), scrl func()) *ImageCanvas {
//	e := ImageCanvas{click: clk, scrolled: scrl}
//	e.Resource = res
//	e.ExtendBaseWidget(&e)
//	return &e
//}

//func (ic *ImageCanvas) Clicked(_ *fyne.PointEvent) {
//	if ic.click == nil {
//		return
//	}
//
//	ic.click()
//}

//func (ic *ImageCanvas) Scrolled(_ *fyne.ScrollEvent) {
//	if ic.scrolled == nil {
//		return
//	}

//	ic.scrolled()
//}
*/

var index int
var loadedimg *canvas.Image
var cwd string
var imageInfo []os.FileInfo
var globalA fyne.App
var globalW fyne.Window
var GUI fyne.CanvasObject
var verboseFlag = flag.Bool("v", false, "verbose flag")
var zoomFlag = flag.Bool("z", false, "set zoom flag to allow zooming up a lot")
var stickyFlag = flag.Bool("sticky", true, "sticky flag for keeping zoom factor among images.") // default of true from 8/21/23
var reverseFlag = flag.Bool("r", false, "reverse sort flag, ie, oldest image is first.")
var sticky bool
var rotatedCtr int64 // used in keyTyped.  And atomicadd so need this type.
var imageAsDisplayed image.Image

var scaleFactor float64 = 1
var shiftState bool // it must be global to preserve state btwn key presses.

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var cyan = color.NRGBA{R: 0, G: 255, B: 255, A: 255}

// isNotImageStr ----------------------------------------
func isNotImageStr(name string) bool {
	ismage := isImage(name)
	return !ismage
}

// isImage ----------------------------------------------
func isImage(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	ext = strings.ToLower(ext)

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

// main --------------------------------------------------
func main() {
	var err error
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " %s last altered %s, and compiled with %s. \n", os.Args[0], LastModified, runtime.Version())
		fmt.Fprintf(flag.CommandLine.Output(), " Usage information:\n")
		fmt.Fprintf(flag.CommandLine.Output(), " z = zoom and also toggles sticky.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " v = verbose.\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	sticky = *zoomFlag || *stickyFlag

	str := fmt.Sprintf("Image Viewer2 last modified %s, compiled using %s", LastModified, runtime.Version())
	if *verboseFlag {
		fmt.Println(str) // this works as intended
	}

	cwd, err = os.Getwd()
	if err != nil {
		ctfmt.Printf(ct.Red, true, " os.Getwd() err is %s.\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	indexChan := make(chan int, 1)
	imgFileInfoChan := make(chan []os.FileInfo, 1) // buffered channel

	go MyReadDirForImages(cwd, imgFileInfoChan)

	if *verboseFlag {
		ctfmt.Printf(ct.Red, true, " cwd = %s\n", cwd)
	}
	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)

	imageInfo = <-imgFileInfoChan // unary channel operator reads from the channel.
	if *verboseFlag {
		ctfmt.Printf(ct.Red, true, " after imageInfo channel read.  Len = %d, cap = %d\n", len(imageInfo), cap(imageInfo))
	}

	imgFilename := flag.Arg(0)
	baseFilename := filepath.Base(imgFilename)
	//ctfmt.Printf(ct.Red, true, " imgFilename = %s, baseFilename = %s\n", imgFilename, baseFilename)

	if flag.NArg() >= 1 {
		go filenameIndex(imageInfo, baseFilename, indexChan)
		_, err = os.Stat(imgFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  Skipped.", imgFilename, err)
		}

		if isNotImageStr(imgFilename) {
			fmt.Fprintln(os.Stderr, imgFilename, "does not have an image extension.  Skipped.")
		}

		index = <-indexChan // syntax to read from a channel.
	}

	if index < 0 {
		index = 0
	}

	loadTheImage()

	globalW.ShowAndRun()

} // end main

// loadTheImage ------------------------------
func loadTheImage() {
	imgName := imageInfo[index].Name()
	//                                       fullfilename := cwd + string(filepath.Separator) + imgname
	fullFilename, err := filepath.Abs(imgName)
	if err != nil {
		fmt.Printf(" loadTheImage(%d): error is %s.  imgName=%s, fullFilename is %s \n", index, err, imgName, fullFilename)

	}
	imageURI := storage.NewFileURI(fullFilename)
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullFilename, "is", err)
		return
	}
	//defer imgRead.Close() // moved based on static linter.  Not needed with imaging.Open

	_, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format name used during format registration by the init function.  Changed this line 2/22/25.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from image.Decode is %s.  Skipped.\n", err)
		return
	}
	img, err := imaging.Open(fullFilename, imaging.AutoOrientation(true)) // added 2/22/25
	if err != nil {
		fmt.Printf(" Error from imaging.Open is %s.  Skipped.\n", err)
		return
	}
	imgHeight := img.Bounds().Max.Y
	imgWidth := img.Bounds().Max.X
	//bounds := img.Bounds()
	//imgHeight := bounds.Max.Y
	//imgWidth := bounds.Max.X

	//                             title := fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s", imgname, imgWidth, imgHeight, imgFmtName, cwd)
	title := fmt.Sprintf("%s: %s %d x %d, SF=%.2f ", imgFmtName, imgName, imgWidth, imgHeight, scaleFactor)
	if *verboseFlag {
		fmt.Println(title)
	}

	//bounds = img.Bounds()   This seems to be redundant.  I don't know why it's here.
	//imgHeight = bounds.Max.Y
	//imgWidth = bounds.Max.X

	/*  Andy said that this code is not needed.
	if imgWidth > maxWidth {
		img = resize.Resize(maxWidth, 0, img, resize.Lanczos3)
		title = title + "; resized."
	} else if imgHeight > maxHeight {
		img = resize.Resize(0, maxHeight, img, resize.Lanczos3)
		title = title + "; resized."
	}
	*/

	if scaleFactor != 1 {
		if imgHeight > imgWidth { // resize the larger dimension, hoping for minimizing distortion.
			scaledHeight := float64(imgHeight) * scaleFactor
			intHeight := uint(math.Round(scaledHeight))
			img = resize.Resize(0, intHeight, img, resize.Lanczos3)
		} else {
			scaledWidth := float64(imgWidth) * scaleFactor
			intWidth := uint(math.Round(scaledWidth))
			img = resize.Resize(intWidth, 0, img, resize.Lanczos3)
		}
		imgHeight = img.Bounds().Max.Y
		imgWidth = img.Bounds().Max.X
	}

	if *verboseFlag {
		imgHeight = img.Bounds().Max.Y
		imgWidth = img.Bounds().Max.X
		fmt.Println(" Scalefactor =", scaleFactor, "last height =", imgHeight, "last width =", imgWidth)
		fmt.Println()
	}

	labelStr := fmt.Sprintf("%s %s; %dw x %dh", imgFmtName, imgName, imgWidth, imgHeight)
	label := canvas.NewText(labelStr, green)
	label.TextStyle.Bold = true
	label.Alignment = fyne.TextAlignCenter

	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.ScaleMode = canvas.ImageScaleFastest
	if !*zoomFlag {
		loadedimg.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
	}

	imageAsDisplayed = loadedimg.Image

	GUI = container.NewBorder(nil, label, nil, nil, loadedimg) // top, bottom, left, right, center
	atomic.StoreInt64(&rotatedCtr, 0)                          // reset this counter when load a fresh image.
	globalW.SetContent(GUI)
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight+textboxheight)))
	globalW.SetTitle(title)

	globalW.Show()
} // end loadTheImage

// filenameIndex --------------------------------------
func filenameIndex(fileinfos []os.FileInfo, name string, intchan chan int) {
	for i, fi := range fileinfos {
		if fi.Name() == name {
			intchan <- i
			return
		}
	}
	intchan <- -1
	// return  redundant
}

//  MyReadDirForImages -----------------------------------

func MyReadDirForImages(dir string, imageInfoChan chan []os.FileInfo) {
	dirname, err := os.Open(dir)
	if err != nil {
		return
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return
	}

	t0 := time.Now()
	fi := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		if isImage(name) {
			imgInfo, err := os.Lstat(dir + string(filepath.Separator) + name)
			if err != nil {
				fmt.Fprintln(os.Stderr, " Error from os.Lstat ", err)
				continue
			}
			fi = append(fi, imgInfo)
		}
	}

	sortfcn := func(i, j int) bool {
		return fi[i].ModTime().After(fi[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
	}

	if *reverseFlag {
		sortfcn = func(i, j int) bool {
			return fi[i].ModTime().Before(fi[j].ModTime()) // I want a oldest-first sort.  Added 9/7/23.
		}
	}

	sort.Slice(fi, sortfcn)
	elapsedtime := time.Since(t0)

	if *verboseFlag {
		fmt.Printf(" Length of the image fileinfo slice is %d; created and sorted in %s\n", len(fi), elapsedtime.String())
		fmt.Println()
	}

	imageInfoChan <- fi
} // MyReadDirForImages

/*
// isSorted -----------------------------------------------
func isSorted(slice []os.FileInfo) bool {
	for i := 0; i < len(slice)-1; i++ {
		if slice[i].ModTime().Before(slice[i+1].ModTime()) {
			fmt.Println(" debugging: i=", i, "Name[i]=", slice[i].Name(), " and Name[i+1]=", slice[i+1].Name())
			return false
		}
	}
	return true
}
*/

//  nextImage -----------------------------------------------------

func nextImage() {
	index++
	if index >= len(imageInfo) {
		index--
	}
	loadTheImage()
	//  return  redundant
} // end nextImage

//  prevImage -------------------------------------------------------

func prevImage() {
	index--
	if index < 0 {
		index++
	}
	loadTheImage()
	// return redundant
} // end prevImage

// firstImage -----------------------------------------------------
func firstImage() {
	index = 0
	loadTheImage()
}

// lastImage ---------------------------------------------------------
func lastImage() {
	index = len(imageInfo) - 1
	loadTheImage()
}

// keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index and shiftState are global var's
	switch e.Name {
	case fyne.KeyW:
		baseName := imageInfo[index].Name()
		err := saveImage(imageAsDisplayed, baseName)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from saveImage(%s) is %s.  Skipped.\n", baseName, err)
		}
	case fyne.KeyS:
		baseName := imageInfo[index].Name()
		err := imageSave(imageAsDisplayed, baseName)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from saveImage(%s) is %s.  Skipped.\n", baseName, err)
		}

	case fyne.KeyUp, fyne.KeyLeft:
		if !sticky {
			scaleFactor = 1
		}
		prevImage()
	case fyne.KeyDown, fyne.KeyRight:
		if !sticky {
			scaleFactor = 1
		}
		nextImage()
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quits the app if this is the last window, which it is.
	case fyne.KeyHome:
		if !sticky {
			scaleFactor = 1
		}
		firstImage()
	case fyne.KeyEnd:
		if !sticky {
			scaleFactor = 1
		}
		lastImage()
	case fyne.KeyPageUp, fyne.KeyPlus, fyne.KeyAsterisk:
		scaleFactor *= 1.1
		loadTheImage()
	case fyne.KeyPageDown, fyne.KeyMinus, fyne.KeySlash:
		scaleFactor *= 0.9
		loadTheImage()
	case fyne.KeyEqual: // first added Feb 22, 2025.  I thought I had the from the beginning.  So it goes.
		scaleFactor = 1
		atomic.StoreInt64(&rotatedCtr, 0) // reset this counter when load a fresh image.
		loadTheImage()
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		if !sticky {
			scaleFactor = 1
		}
		nextImage()
	case fyne.KeyBackspace: // preserve always resetting zoomfactor here.  Hope I remember I'm doing this.
		scaleFactor = 1
		prevImage()
	case fyne.KeyV:
		*verboseFlag = !*verboseFlag
		fmt.Printf(" Verbose flag is now %t, Sticky is %t and scaleFactor is %2.2g\n", *verboseFlag, sticky, scaleFactor)
	case fyne.KeyZ:
		sticky = !sticky
		if *verboseFlag {
			fmt.Println(" Sticky is now", sticky, "and scaleFactor is", scaleFactor)
		}
	case fyne.KeyR:
		atomic.AddInt64(&rotatedCtr, 1)
		rotateAndLoadTheImage(index, rotatedCtr) // index and rotatedCtr are global
	case fyne.Key1:
		rotateAndLoadTheImage(index, 1)
	case fyne.Key2:
		rotateAndLoadTheImage(index, 2)
	case fyne.Key3:
		rotateAndLoadTheImage(index, 3)
	case fyne.Key4, fyne.Key0:
		atomic.StoreInt64(&rotatedCtr, 0) // reset this counter when load a fresh image.
		rotateAndLoadTheImage(index, 0)

	default:
		if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
			shiftState = true
			return
		}
		if shiftState {
			shiftState = false
			if e.Name == fyne.KeyEqual { // like key plus
				scaleFactor *= 1.1
				loadTheImage()
			} else if e.Name == fyne.KeyPeriod { // >
				nextImage()
			} else if e.Name == fyne.KeyComma { // <
				prevImage()
			} else if e.Name == fyne.Key8 { // *
				scaleFactor *= 1.1
				loadTheImage()
			}
		} else if e.Name == fyne.KeyEqual {
			scaleFactor = 1
			loadTheImage()
		}
	}
}

// mouseClicked
func mouseClicked(e *fyne.PointEvent) {
	fmt.Println(" received clicked event.  X =", e.Position.X, " and Y =", e.Position.Y, "on image.")
}

// mouseScrolled
func mouseScrolled(e *fyne.ScrollEvent) {
	fmt.Println(" received scroll event.  scrolled delta X =", e.Scrolled.DX, " and scrolled delta Y =", e.Scrolled.DY, "at point X =",
		e.Position.X, "and point Y =", e.Position.Y)

}

// rotateAndLoadTheImage -- loads the image given by the index, and then rotates it before displaying it.
func rotateAndLoadTheImage(idx int, repeat int64) {
	imgName := imageInfo[idx].Name()
	fullFilename, err := filepath.Abs(imgName)
	if err != nil {
		fmt.Printf(" loadTheImage(%d): error is %s.  imgName=%s, fullFilename is %s \n", idx, err, imgName, fullFilename)
	}

	imgRead, err := imaging.Open(fullFilename, imaging.AutoOrientation(true))
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from storage.Reader(%s) is %s.  Skipped.\n", fullFilename, err)
		return
	}

	var rotatedImg *image.NRGBA
	var imgImg image.Image

	//for range repeat { old way
	//	rotatedImg = imaging.Rotate90(imgRead)
	//	imgImg = imgImage(rotatedImg) // need to convert from *image.NRGBA to image.Image
	//	imgRead = imgImg
	//}

	repeat = repeat % 4 // only allow 0..3
	rotateFac := float64(repeat) * 90
	rotatedImg = imaging.Rotate(imgRead, rotateFac, nil)
	imgImg = imgImage(rotatedImg) //  need to convert from *image.NRGBA to image.Image becuse of the resize fcn.

	bounds := rotatedImg.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X

	title := fmt.Sprintf(" %s, %d x %d, SF=%.2f \n", imgName, imgWidth, imgHeight, scaleFactor)
	if *verboseFlag {
		fmt.Println(title)
	}

	if scaleFactor != 1 {
		if imgHeight > imgWidth { // resize the larger dimension, hoping for minimizing distortion.
			scaledHeight := float64(imgHeight) * scaleFactor
			intHeight := uint(math.Round(scaledHeight))
			imgImg = resize.Resize(0, intHeight, imgImg, resize.Lanczos3)
		} else {
			scaledWidth := float64(imgWidth) * scaleFactor
			intWidth := uint(math.Round(scaledWidth))
			imgImg = resize.Resize(intWidth, 0, imgImg, resize.Lanczos3)
		}
		bounds = imgImg.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		//                                title = fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s\n", imgname, imgWidth, imgHeight, imgFmtName, cwd)
		title = fmt.Sprintf("%s, %d x %d, SF=%.2f \n", imgName, imgWidth, imgHeight, scaleFactor)
	}

	if *verboseFlag {
		bounds = imgImg.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		fmt.Println(" Scalefactor =", scaleFactor, "last height =", imgHeight, "last width =", imgWidth)
		fmt.Printf(" loadTheImage(%d): imgName=%s, fullFilename is %s \n", idx, imgName, fullFilename)
		fmt.Println()
	}

	imageAsDisplayed = imgImg

	canvasImage := canvas.NewImageFromImage(imgImg)
	canvasImage.ScaleMode = canvas.ImageScaleSmooth
	if !*zoomFlag {
		canvasImage.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
		//loadedimg.FillMode = canvas.ImageFillOriginal -- sets min size to be that of the original.
	}

	globalW.SetContent(canvasImage)
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))
	globalW.SetTitle(title)
	globalW.Show()

} // end rotateAndLoadTheImage

func imgImage(img *image.NRGBA) image.Image {
	return img
}

func saveImage(img image.Image, inputname string) error { // uses method 1
	if img == nil {
		return fmt.Errorf("image passed to saveImage is nil")
	}
	ext := filepath.Ext(inputname)
	bounds := img.Bounds()
	imgWidth := bounds.Max.X
	imgHeight := bounds.Max.Y
	sizeStr := fmt.Sprintf("%dx%d_rot_%d", imgWidth, imgHeight, rotatedCtr)
	savedName := inputname[:len(inputname)-len(ext)] + "_saved_" + sizeStr + ext // using strings.TrimSuffix would likely also work here
	err := imaging.Save(img, savedName)
	fmt.Printf(" Saved image %s with error of %v\n", savedName, err)
	return err
}

func imageSave(img image.Image, inputname string) error { // uses method 2, just to see if both work.
	if img == nil {
		return fmt.Errorf("image passed to imageSave is nil")
	}
	ext := filepath.Ext(inputname)
	bounds := img.Bounds()
	imgWidth := bounds.Max.X
	imgHeight := bounds.Max.Y
	sizeStr := fmt.Sprintf("%dx%d_rot_%d", imgHeight, imgWidth, rotatedCtr)
	savedName := inputname[:len(inputname)-len(ext)] + "_saved_" + sizeStr + ext

	f, err := os.Create(savedName)
	if err != nil {
		return fmt.Errorf("in imageSave, error creating file %s: %v", savedName, err)
	}
	defer f.Close()

	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	case ".png":
		err = png.Encode(f, img)
	case ".gif":
		err = gif.Encode(f, img, nil)
	default: // it seems that webp doesn't have an encode method.
		err = fmt.Errorf("cannot encode unsupported image format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("error encoding image: %v", err)
	}

	fmt.Printf(" Saved image %s\n", savedName)
	return nil
}
