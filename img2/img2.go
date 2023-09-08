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
*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
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

const LastModified = "Sep 7, 2023"
const textboxheight = 20

// const maxWidth = 1800 // actual resolution is 1920 x 1080
// const maxHeight = 900 // actual resolution is 1920 x 1080

type ImageWidget struct {
	widget.BaseWidget
	widget.Icon
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
	if *verboseFlag {
		fmt.Printf(" scroll event found.  X= %.4f, dX=%.4f; Y=%.4f, dY=%.4f, for %s\n", e.Position.X, e.Scrolled.DX, e.Position.Y, e.Scrolled.DY,
			iw.Resource.Name())
	}

	if e.Scrolled.DY < 0 || e.Scrolled.DX < 0 {
		nextImage()
		return
	}
	prevImage()
}

type imageRender struct {
	//imgWdgt *ImageWidget
	*ImageWidget // making this an embedded type fixed an issue w/ MinSize.
}

func (ir *imageRender) MinSize() fyne.Size { // This func is needed to define a renderer
	return fyne.NewSize(float32(ir.x), float32(ir.y))
}
func (ir *imageRender) Layout(_ fyne.Size) { // This func is needed to define a renderer
	ir.Resize(ir.MinSize())
}
func (ir *imageRender) Destroy() { // This func is needed to define a renderer, but it can be empty
}

// func (ir *imageRender) Refresh() { // This func is needed to define a renderer
//		ir.Refresh()
// }

func (ir *imageRender) Objects() []fyne.CanvasObject { // This func is needed to define a renderer
	return []fyne.CanvasObject{ir}
}

func (iw *ImageWidget) CreateRenderer() fyne.WidgetRenderer {
	//rndrr := imageRender{imgWdgt: iw}
	rndrr := imageRender{iw}
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
	defer imgRead.Close() // moved based on static linter

	img, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format name used during format registration by the init function.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from image.Decode is %s.  Skipped.\n", err)
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

	GUI = container.NewBorder(nil, label, nil, nil, loadedimg) // top, bottom, left, right, center
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
	// return  redundant
} // MyReadDirForImages

// isSorted ----------------------------------------------- is unused
//func isSorted(slice []os.FileInfo) bool {
//	for i := 0; i < len(slice)-1; i++ {
//		if slice[i].ModTime().Before(slice[i+1].ModTime()) {
//			fmt.Println(" debugging: i=", i, "Name[i]=", slice[i].Name(), " and Name[i+1]=", slice[i+1].Name())
//			return false
//		}
//	}
//	return true
//}

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
		fmt.Printf(" Verbose flag is now %t, Sticky is %t and scaleFactor is %2.2g\n", *verboseFlag, sticky, scaleFactor)
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

// mouseClicked unused
//func mouseClicked(e *fyne.PointEvent) {
//	fmt.Println(" received clicked event.  X =", e.Position.X, " and Y =", e.Position.Y, "on image.")
//}

// mouseScrolled unused
//func mouseScrolled(e *fyne.ScrollEvent) {
//	fmt.Println(" received scroll event.  scrolled delta X =", e.Scrolled.DX, " and scrolled delta Y =", e.Scrolled.DY, "at point X =",
//		e.Position.X, "and point Y =", e.Position.Y)
//}
