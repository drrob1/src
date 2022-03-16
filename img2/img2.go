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
*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"image/color"
	"math"

	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/nfnt/resize"
)

const LastModified = "Mar 16, 2022"
const maxWidth = 1800 // actual resolution is 1920 x 1080
const maxHeight = 900 // actual resolution is 1920 x 1080
const textboxheight = 20

type ImageWidget struct {
	widget.BaseWidget
	img      *canvas.Image
	x, y     int
	click    func(event *fyne.PointEvent)
	scrolled func(event *fyne.ScrollEvent)
	filename string
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
var stickyFlag = flag.Bool("sticky", false, "sticky flag for keeping zoom factor among images.")
var sticky bool

var scaleFactor float64 = 1
var shiftState bool // it must be global to preserve state btwn key presses.

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var cyan = color.NRGBA{R: 0, G: 255, B: 255, A: 255}

//  isNotImageStr ----------------------------------------
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

//  main --------------------------------------------------
func main() {
	flag.Parse()
	sticky = *zoomFlag || *stickyFlag
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, " Usage: img2 <image file name>")
		os.Exit(1)
	}

	str := fmt.Sprintf("Single Image Viewer2 last modified %s, compiled using %s", LastModified, runtime.Version())
	if *verboseFlag {
		fmt.Println(str) // this works as intended
	}

	imgfilename := flag.Arg(0)
	_, err := os.Stat(imgfilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from os.Stat(", imgfilename, ") is", err)
		os.Exit(1)
	}

	if isNotImageStr(imgfilename) {
		fmt.Fprintln(os.Stderr, imgfilename, "does not have an image extension.")
		os.Exit(1)
	}

	basefilename := filepath.Base(imgfilename)
	fullFilename, err := filepath.Abs(imgfilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from filepath.Abs on", imgfilename, "is", err)
		os.Exit(1)
	}

	imgFileHandle, err := os.Open(fullFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from opening", fullFilename, "is", err)
		os.Exit(1)
	}

	imgConfig, _, err := image.DecodeConfig(imgFileHandle) // img is of type image.Config
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from decode config on", fullFilename, "is", err)
		os.Exit(1)
	}
	imgFileHandle.Close()

	var width = float32(imgConfig.Width)
	var height = float32(imgConfig.Height)
	var aspectRatio = width / height
	if aspectRatio > 1 {
		aspectRatio = 1 / aspectRatio
	}

	if *verboseFlag {
		fmt.Printf(" image.Config %s, %s, %s \n width=%g, height=%g, and aspect ratio=%.4g \n",
			imgfilename, fullFilename, basefilename, width, height, aspectRatio)
	}

	if width > maxWidth || height > maxHeight {
		width = maxWidth * aspectRatio
		height = maxHeight * aspectRatio
	}

	if *verboseFlag {
		fmt.Println()
		//                                               fmt.Printf(" Type for DecodeConfig is %T \n", imgConfig) // answer is image.Config
		fmt.Println(" adjusted image.Config width =", width, ", height =", height, " but these values are not used to show the image.")
		fmt.Println()
	}

	cwd = filepath.Dir(fullFilename)
	imgFileInfoChan := make(chan []os.FileInfo, 1) // buffered channel
	go MyReadDirForImages(cwd, imgFileInfoChan)

	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)

	imageURI := storage.NewFileURI(fullFilename) // needs to be a type = fyne.CanvasObject
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullFilename, "is", err)
		os.Exit(1)
	}
	defer imgRead.Close()
	img, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format type (jpeg, png, etc) used during format registration by the init function.
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from image.Decode is", err)
		os.Exit(1)
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X
	if *verboseFlag {
		fmt.Println(" image.Decode, width=", imgWidth, "and height=", imgHeight, ", imgFmtName=", imgFmtName, "and cwd=", cwd)
		fmt.Println()
	}
	/*  Andy commented that this code is not needed, as it's done automatically.
	if imgWidth > maxWidth {
		img = resize.Resize(maxWidth, 0, img, resize.Lanczos3)
	} else if imgHeight > maxHeight {
		img = resize.Resize(0, maxHeight, img, resize.Lanczos3)
	}
	*/

	// Time to make the GUI

	imgtitle := fmt.Sprintf("%s, %d x %d", imgfilename, imgWidth, imgHeight)
	label := canvas.NewText(imgtitle, green)
	label.TextStyle.Bold = true
	label.Alignment = fyne.TextAlignCenter

	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.ScaleMode = canvas.ImageScaleFastest
	if !*zoomFlag {
		loadedimg.FillMode = canvas.ImageFillContain
	}
	GUI = container.NewBorder(nil, label, nil, nil, loadedimg) // top, bottom, left, right, center

	globalW.SetTitle(imgtitle)
	//                                                                                    globalW.SetContent(loadedimg)
	globalW.SetContent(GUI)
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight+textboxheight)))

	imageInfo = <-imgFileInfoChan // unary channel operator reads from the channel.

	if *verboseFlag {
		if isSorted(imageInfo) {
			fmt.Println(" imageInfo slice of FileInfo is sorted.  Length is", len(imageInfo))
		} else {
			fmt.Println(" imageInfo slice of FileInfo is NOT sorted.  Length is", len(imageInfo))
		}
		fmt.Println()
	}

	indexchan := make(chan int, 1)
	t0 := time.Now()

	go filenameIndex(imageInfo, basefilename, indexchan)

	globalW.CenterOnScreen()

	index = <-indexchan // syntax to read from a channel.

	elapsedtime := time.Since(t0)

	fmt.Printf(" %s index is %d in the fileinfo slice of len %d; linear sequential search took %s.\n", basefilename, index, len(imageInfo), elapsedtime)
	fmt.Printf(" As a check, imageInfo[%d] = %s.\n", index, imageInfo[index].Name())
	fmt.Println()

	globalW.ShowAndRun()

} // end main

//  loadTheImage ------------------------------
func loadTheImage() {
	imgname := imageInfo[index].Name()
	fullfilename := cwd + string(filepath.Separator) + imgname
	imageURI := storage.NewFileURI(fullfilename)
	imgRead, err := storage.Reader(imageURI)
	defer imgRead.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullfilename, "is", err)
		os.Exit(1)
	}

	img, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format name used during format registration by the init function.
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from image.Decode is", err)
		os.Exit(1)
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X

	//                             title := fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s", imgname, imgWidth, imgHeight, imgFmtName, cwd)
	title := fmt.Sprintf("%s width=%d, height=%d and type=%s", imgname, imgWidth, imgHeight, imgFmtName)
	if *verboseFlag {
		fmt.Println(title)
	}

	bounds = img.Bounds()
	imgHeight = bounds.Max.Y
	imgWidth = bounds.Max.X

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
		bounds = img.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
	}

	if *verboseFlag {
		bounds = img.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		fmt.Println(" Scalefactor =", scaleFactor, "last height =", imgHeight, "last width =", imgWidth)
		fmt.Println()
	}

	labelStr := fmt.Sprintf("%s; width=%d, height=%d", imgname, imgWidth, imgHeight)
	label := canvas.NewText(labelStr, green)
	label.TextStyle.Bold = true
	label.Alignment = fyne.TextAlignCenter

	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.ScaleMode = canvas.ImageScaleFastest
	if !*zoomFlag {
		loadedimg.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
	}

	GUI = container.NewBorder(nil, label, nil, nil, loadedimg) // top, bottom, left, right, center
	globalW.SetContent(GUI)
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight+textboxheight)))
	globalW.SetTitle(title)

	globalW.Show()
	return
} // end loadTheImage

//  filenameIndex --------------------------------------
func filenameIndex(fileinfos []os.FileInfo, name string, intchan chan int) {
	for i, fi := range fileinfos {
		if fi.Name() == name {
			intchan <- i
			return
		}
	}
	intchan <- -1
	return
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

	fi := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		if isImage(name) {
			imgInfo, err := os.Lstat(name)
			if err != nil {
				fmt.Fprintln(os.Stderr, " Error from os.Lstat ", err)
				continue
			}
			fi = append(fi, imgInfo)
		}
	}

	t0 := time.Now()
	sortfcn := func(i, j int) bool {
		return fi[i].ModTime().After(fi[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
	}

	sort.Slice(fi, sortfcn)
	elapsedtime := time.Since(t0)

	if *verboseFlag {
		fmt.Printf(" Length of the image fileinfo slice is %d, and sorted in %s\n", len(fi), elapsedtime.String())
		fmt.Println()
	}

	imageInfoChan <- fi
	return
} // MyReadDirForImages

//  isSorted -----------------------------------------------
func isSorted(slice []os.FileInfo) bool {
	for i := 0; i < len(slice)-1; i++ {
		if slice[i].ModTime().Before(slice[i+1].ModTime()) {
			fmt.Println(" debugging: i=", i, "Name[i]=", slice[i].Name(), " and Name[i+1]=", slice[i+1].Name())
			return false
		}
	}
	return true
}

//  nextImage -----------------------------------------------------

func nextImage() {
	index++
	if index >= len(imageInfo) {
		index--
	}
	loadTheImage()
	return
} // end nextImage

//  prevImage -------------------------------------------------------

func prevImage() {
	index--
	if index < 0 {
		index++
	}
	loadTheImage()
	return
} // end prevImage

//  firstImage -----------------------------------------------------
func firstImage() {
	index = 0
	loadTheImage()
}

//  lastImage ---------------------------------------------------------
func lastImage() {
	index = len(imageInfo) - 1
	loadTheImage()
}

//  keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index and shiftState are global var's
	switch e.Name {
	case fyne.KeyUp:
		if !sticky {
			scaleFactor = 1
		}
		prevImage()
	case fyne.KeyDown:
		if !sticky {
			scaleFactor = 1
		}
		nextImage()
	case fyne.KeyLeft:
		if !sticky {
			scaleFactor = 1
		}
		prevImage()
	case fyne.KeyRight:
		if !sticky {
			scaleFactor = 1
		}
		nextImage()
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//		(*globalA).Quit()
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
	case fyne.KeyPageUp:
		scaleFactor *= 1.1
		loadTheImage()
	case fyne.KeyPageDown:
		scaleFactor *= 0.9
		loadTheImage()
	case fyne.KeyPlus, fyne.KeyAsterisk:
		scaleFactor *= 1.1
		loadTheImage()
	case fyne.KeyMinus:
		scaleFactor *= 0.9
		loadTheImage()
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		if !sticky {
			scaleFactor = 1
		}
		nextImage()
	case fyne.KeyBackspace: // preserve always resetting zoomfactor here.  Hope I remember I'm doing this.
		scaleFactor = 1
		prevImage()
	case fyne.KeySlash:
		scaleFactor *= 0.9
		loadTheImage()
	case fyne.KeyV:
		*verboseFlag = !*verboseFlag
		fmt.Printf(" Verbose flag is now %t, Sticky is and scaleFactor is %2.2g\n", *verboseFlag, sticky, scaleFactor)
	case fyne.KeyZ:
		sticky = !sticky
		if *verboseFlag {
			fmt.Println(" Sticky is now", sticky, "and scaleFactor is", scaleFactor)
		}

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
