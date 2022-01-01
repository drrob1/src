// From Go GUI with Fyne, Chap 4.  img.go -> imgscroll.go -> imgs.go
/*

This pgm works by the main thread initializing the image display and then starting the display message loop.
Then the keyTyped handles the keyboard events.
I now want to change that so that keyTyped puts these events into a buffered channel that is handled by a different go routine.
Just to see if I can now.


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
22 Sep 21 -- Added shiftState to deal w/ shift= is the plus sign
27 Sep 21 -- Added stickyFlag, sticky and 'z' zoom toggle.  When sticky is true, zoom factor is not cleared automatically.
30 Sep 21 -- Added keyAsterisk, and removed the unneeded scaling code (according to Andy Williams).
 2 Dec 21 -- After listening to Bill Kennedy's Go talks, I made the image channel buffered.
 3 Dec 21 -- Some clean up that I learned from Bill Kennedy.
 4 Dec 21 -- Adding a go routine to process the keystrokes.  And adding "v" to turn on verbose mode.
26 Dec 21 -- Adding display of the image minsize.  I didn't know it existed until today.
27 Dec 21 -- Now called imgcsroll.go.  I'm going to take a stab at adding mouse wheel detection.
28 Dec 21 -- It works!  Now to get it to do what I want w/ the mouse wheel.  Mouse clicks will still print a message if verbose is set.
31 Dec 21 -- Decided to call it imgs, and removed dead code.
*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"math"

	//"fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/layout"
	//"fyne.io/fyne/v2/container"
	//"image/color"

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
	"github.com/nfnt/resize"
)

const LastModified = "Dec 31, 2021"
const maxWidth = 1800 // actual resolution is 1920 x 1080
const maxHeight = 900 // actual resolution is 1920 x 1080
const keyCmdChanSize = 20
const (
	firstImgCmd = iota
	prevImgCmd
	nextImgCmd
	loadImgCmd
	lastImgCmd
)

var index int

var cwd string
var imageInfo []os.FileInfo
var globalA fyne.App
var globalW fyne.Window
var verboseFlag = flag.Bool("v", false, "verbose flag.")
var zoomFlag = flag.Bool("z", false, "set zoom flag to allow zooming up a lot.")
var stickyFlag = flag.Bool("sticky", false, "sticky flag for keeping zoom factor among images.")
var sticky bool
var scaleFactor float64 = 1
var shiftState bool
var keyCmdChan chan int

type scrollClickImg struct {
	widget.Icon
	cImage *canvas.Image
	iImage image.Image
	imgFmt string
}

func NewScrollClickImg(res fyne.Resource) *scrollClickImg {
	scrClkImg := scrollClickImg{}
	scrClkImg.ExtendBaseWidget(&scrClkImg)
	scrClkImg.SetResource(res)
	scrClkImg.cImage = canvas.NewImageFromResource(res)
	scrClkImg.cImage.ScaleMode = canvas.ImageScaleFastest
	buf := bytes.NewReader(scrClkImg.Resource.Content())
	var err error
	scrClkImg.iImage, scrClkImg.imgFmt, err = image.Decode(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, " image decoding error: %v \n", err)
	}
	return &scrClkImg
}

func (sI *scrollClickImg) Scrolled(se *fyne.ScrollEvent) {
	if *verboseFlag {
		fmt.Printf(" scroll event found.  X= %.4f, dX=%.4f; Y=%.4f, dY=%.4f, for %s\n", se.Position.X, se.Scrolled.DX, se.Position.Y, se.Scrolled.DY,
			sI.Resource.Name())
	}
	if se.Scrolled.DY < 0 { // will ignore DX as the scroll wheel only changes DY.
		keyCmdChan <- nextImgCmd
	} else { // I forgot to do this, so both would be sent in this branch.  Oops.
		keyCmdChan <- prevImgCmd
	}
}
func (sI *scrollClickImg) Tapped(pe *fyne.PointEvent) {
	if *verboseFlag {
		fmt.Printf(" leftclick event found.  X=%.4f, Y=%.4f for %s\n", pe.Position.X, pe.Position.Y, sI.Resource.Name())
	}
}
func (sI *scrollClickImg) TappedSecondary(pe *fyne.PointEvent) {
	if *verboseFlag {
		fmt.Printf(" rightclick event found.  X=%.4f, Y=%.4f for %s\n", pe.Position.X, pe.Position.Y, sI.Resource.Name())
	}
}

// -------------------------------------------------------- isNotImageStr ----------------------------------------
func isNotImageStr(name string) bool {
	return !isImage(name)
}

// ----------------------------------------------------------isImage ----------------------------------------------
func isImage(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

// ---------------------------------------------------- main --------------------------------------------------
func main() {
	flag.Parse()
	sticky = *zoomFlag || *stickyFlag
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, " Usage: img <image file name>")
		os.Exit(1)
	}

	str := fmt.Sprintf("Single Image Viewer last modified %s, compiled using %s", LastModified, runtime.Version())
	fmt.Println(str)

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

	keyCmdChan = make(chan int, keyCmdChanSize)
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
		fmt.Printf(" image.Config %s, %s, %s \n width=%g, height=%g, and aspect ratio=%.4g.  Sticky=%t \n",
			imgfilename, fullFilename, basefilename, width, height, aspectRatio, sticky)
	}

	if width > maxWidth || height > maxHeight {
		width = maxWidth * aspectRatio
		height = maxHeight * aspectRatio
	}

	if *verboseFlag {
		fmt.Println()
		//fmt.Printf(" Type for DecodeConfig is %T \n", imgConfig) // answer is image.Config
		fmt.Println(" adjusted image.Config width =", width, ", height =", height, " but these values are not used to show the image.")
		fmt.Println()
	}

	cwd = filepath.Dir(fullFilename)
	imgFileInfoChan := make(chan []os.FileInfo, 1) // unbuffered channel increases latency.  Will make it buffered now.  It only needs a buffer of 1 because it only receives once.
	go MyReadDirForImages(cwd, imgFileInfoChan)

	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)

	imageResource, err := fyne.LoadResourceFromPath(fullFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from LoadResourceFromPath.  File=", fullFilename, "err:", err)
	}
	if *verboseFlag {
		fmt.Printf(" Resource name=%s\n ", imageResource.Name())
	}
	scrollClickImage := NewScrollClickImg(imageResource)

	bounds := scrollClickImage.iImage.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X
	if *verboseFlag {
		fmt.Println(" image.Decode, width=", imgWidth, "and height=", imgHeight, ", imgFmtName=", scrollClickImage.imgFmt, "and cwd=", cwd, ".  Min x =", bounds.Min.X,
			"and min y =", bounds.Min.Y)
		fmt.Println()
	}

	//loadedimg = canvas.NewImageFromImage(img) original code I wrote that I know works.
	//loadedimg = canvas.NewImageFromResource(imageResource)  Goland does not flag this line w/ an error.
	if !*zoomFlag {
		scrollClickImage.cImage.FillMode = canvas.ImageFillContain
	}

	imgtitle := fmt.Sprintf("%s, %d x %d", imgfilename, imgWidth, imgHeight)
	globalW.SetTitle(imgtitle)
	//globalW.SetContent(scrollClickImage.cImage)  Doesn't respond to the mouse.
	globalW.SetContent(scrollClickImage) // this does respond to the mouse.  Yea!!!
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))

	// this syntax works and was blocking until I made the channel buffered.
	imageInfo = <-imgFileInfoChan // reading from a channel is a unary use of the channel operator.

	if *verboseFlag {
		if isSorted(imageInfo) {
			fmt.Println(" imageInfo slice of FileInfo is sorted.  Length is", len(imageInfo))
		} else {
			fmt.Println(" imageInfo slice of FileInfo is NOT sorted.  Length is", len(imageInfo))
		}
		fmt.Println()
	}

	indexchan := make(chan int, 1) // I'm now making this buffered as I don't need a guarantee of receipt.  This may reduce latency.
	t0 := time.Now()

	go filenameIndex(imageInfo, basefilename, indexchan)

	globalW.CenterOnScreen()

	index = <-indexchan // syntax to read from a channel, using the channel operator as a unary operator.
	elapsedtime := time.Since(t0)

	if *verboseFlag {
		fmt.Printf(" %s index is %d in the fileinfo slice of len %d; linear sequential search took %s.\n", basefilename, index, len(imageInfo), elapsedtime)
		fmt.Printf(" As a check, imageInfo[%d] = %s.\n", index, imageInfo[index].Name())
		fmt.Println()
	}

	go processKeys()

	globalW.ShowAndRun()

} // end main

// --------------------------------------------------- processKeys -------------------------------
func processKeys() {
	for {
		keyCmd := <-keyCmdChan
		//fmt.Println("in processKeys go routine.  keycmd =", keyCmd)
		switch keyCmd {
		case firstImgCmd:
			firstImage()
		case prevImgCmd:
			prevImage()
		case nextImgCmd:
			nextImage()
		case loadImgCmd:
			loadTheImage()
		case lastImgCmd:
			lastImage()
		}
	}
}

// --------------------------------------------------- loadTheImage ------------------------------
func loadTheImage() {
	imgname := imageInfo[index].Name()
	//fullfilename := cwd + string(filepath.Separator) + imgname
	fullFilename, err := filepath.Abs(imgname)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from filepath.Abs in loadTheImage is %v\n", err)
	}
	imageResource, err := fyne.LoadResourceFromPath(fullFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, " In loadTheImage, error from LoadResourceFromPath is %v\n", err)
	}
	scrollClickImage := NewScrollClickImg(imageResource)

	bounds := scrollClickImage.iImage.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X

	title := fmt.Sprintf("%s, %d x %d, type=%s \n", imgname, imgWidth, imgHeight, scrollClickImage.imgFmt)
	if *verboseFlag {
		fmt.Println(title)
	}

	if scaleFactor != 1 {
		if imgHeight > imgWidth { // resize the larger dimension, hoping for minimizing distortion.
			scaledHeight := float64(imgHeight) * scaleFactor
			intHeight := uint(math.Round(scaledHeight))
			scrollClickImage.iImage = resize.Resize(0, intHeight, scrollClickImage.iImage, resize.Lanczos3)
		} else {
			scaledWidth := float64(imgWidth) * scaleFactor
			intWidth := uint(math.Round(scaledWidth))
			scrollClickImage.iImage = resize.Resize(intWidth, 0, scrollClickImage.iImage, resize.Lanczos3)
		}
		bounds = scrollClickImage.iImage.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		title = fmt.Sprintf("%s, %d x %d, type=%s \n", imgname, imgWidth, imgHeight, scrollClickImage.imgFmt)
	}

	if *verboseFlag {
		bounds = scrollClickImage.iImage.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		fmt.Println(" Scalefactor =", scaleFactor, "last height =", imgHeight, "last width =", imgWidth)
		fmt.Println()
	}

	scrollClickImage.cImage.ScaleMode = canvas.ImageScaleSmooth
	if !*zoomFlag {
		scrollClickImage.cImage.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
		//loadedimg.FillMode = canvas.ImageFillOriginal -- sets min size to be that of the original.
	}

	//globalW.SetContent(scrollClickImage.cImage) // doesn't respond to the mouse.
	globalW.SetContent(scrollClickImage) // This does respond to the mouse.  Yea!!!
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))
	globalW.SetTitle(title)

	globalW.Show()
	return
} // end loadTheImage

// ------------------------------- filenameIndex --------------------------------------
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

// ------------------------------- MyReadDirForImages -----------------------------------

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

// ------------------------------------------------------- isSorted -----------------------------------------------
func isSorted(slice []os.FileInfo) bool {
	for i := 0; i < len(slice)-1; i++ {
		if slice[i].ModTime().Before(slice[i+1].ModTime()) {
			fmt.Println(" debugging: i=", i, "Name[i]=", slice[i].Name(), " and Name[i+1]=", slice[i+1].Name())
			return false
		}
	}
	return true
}

// ---------------------------------------------- nextImage -----------------------------------------------------
//func nextImage(indx int) *canvas.Image {
func nextImage() {
	index++
	if index >= len(imageInfo) {
		index--
	}
	loadTheImage()
	return
} // end nextImage

// ------------------------------------------ prevImage -------------------------------------------------------
//func prevImage(indx int) *canvas.Image {
func prevImage() {
	index--
	if index < 0 {
		index++
	}
	loadTheImage()
	return
} // end prevImage

// ------------------------------------------ firstImage -----------------------------------------------------
func firstImage() {
	index = 0
	loadTheImage()
}

// ------------------------------------------ lastImage ---------------------------------------------------------
func lastImage() {
	index = len(imageInfo) - 1
	loadTheImage()
}

// ------------------------------------------------------------ keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index and shiftState are global var's
	switch e.Name {
	case fyne.KeyUp:
		if !sticky {
			scaleFactor = 1
		}
		//prevImage()
		keyCmdChan <- prevImgCmd
	case fyne.KeyDown:
		if !sticky {
			scaleFactor = 1
		}
		//nextImage()
		keyCmdChan <- nextImgCmd
	case fyne.KeyLeft:
		if !sticky {
			scaleFactor = 1
		}
		//prevImage()
		keyCmdChan <- prevImgCmd
	case fyne.KeyRight:
		if !sticky {
			scaleFactor = 1
		}
		//nextImage()
		keyCmdChan <- nextImgCmd
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//		(*globalA).Quit()
	case fyne.KeyHome:
		if !sticky {
			scaleFactor = 1
		}
		//firstImage()
		keyCmdChan <- firstImgCmd
	case fyne.KeyEnd:
		if !sticky {
			scaleFactor = 1
		}
		//lastImage()
		keyCmdChan <- lastImgCmd
	case fyne.KeyPageUp:
		scaleFactor *= 1.1 // I'm reversing what I did before.  PageUp now scales up
		//loadTheImage()
		keyCmdChan <- loadImgCmd
	case fyne.KeyPageDown:
		scaleFactor *= 0.9 // I'm reversing what I did before.  PageDn now scales down
		//loadTheImage()
		keyCmdChan <- loadImgCmd
	case fyne.KeyPlus, fyne.KeyAsterisk:
		scaleFactor *= 1.1
		//loadTheImage()
		keyCmdChan <- loadImgCmd
	case fyne.KeyMinus:
		scaleFactor *= 0.9
		//loadTheImage()
		keyCmdChan <- loadImgCmd
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		if !sticky {
			scaleFactor = 1
		}
		//nextImage()
		keyCmdChan <- nextImgCmd
	case fyne.KeyBackspace: // preserve always resetting zoomfactor here.  Hope I remember I'm doing this.
		scaleFactor = 1
		//prevImage()
		keyCmdChan <- prevImgCmd
	case fyne.KeySlash:
		scaleFactor *= 0.9
		//loadTheImage()
		keyCmdChan <- loadImgCmd
	case fyne.KeyV:
		*verboseFlag = true
		fmt.Println(" Verbose flag is now on, and Sticky is", sticky, ", and scaleFactor is", scaleFactor)
	case fyne.KeyZ:
		sticky = !sticky
		*verboseFlag = !*verboseFlag
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
			if e.Name == fyne.KeyEqual { // shift equal is key plus
				scaleFactor *= 1.1
				//loadTheImage()
				keyCmdChan <- loadImgCmd
			} else if e.Name == fyne.KeyPeriod { // >
				//nextImage()
				keyCmdChan <- nextImgCmd
			} else if e.Name == fyne.KeyComma { // <
				//prevImage()
				keyCmdChan <- prevImgCmd
			} else if e.Name == fyne.Key8 { // *
				scaleFactor *= 1.1
				//loadTheImage()
				keyCmdChan <- loadImgCmd
			}
		} else if e.Name == fyne.KeyEqual {
			scaleFactor = 1
			//loadTheImage()
			keyCmdChan <- loadImgCmd
		}
	}
} // end keyTyped
