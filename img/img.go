package main

import (
	"fmt"
	"image"
	"image/gif"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	flag "github.com/spf13/pflag"
	_ "golang.org/x/image/webp"
	//_ "image/gif"
	"image/jpeg"
	//_ "image/jpeg"
	"image/png"
	//_ "image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	//ct "github.com/daviddengcn/go-colortext"
	//ctfmt "github.com/daviddengcn/go-colortext/fmt"
	//"fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/layout"
	//"fyne.io/fyne/v2/container"
	//"image/color"
	//"github.com/disintegration/imaging"
)

// Based on Go GUI with Fyne, Chap 4.
/*

This pgm works by the main thread initializing the image display and then starting the display message loop.
Then the keyTyped function handles the keyboard events.
I now want to change that so that keyTyped puts these events into a buffered channel that is handled by a different go routine.
Just to see if I can.


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
16 Mar 22 -- Only writing using fmt.Print calls if verbose or flags are set.
26 Mar 22 -- Expanding to allow a directory other than the current one.
21 Nov 22 -- Made changes recommended by static linter.
21 Aug 23 -- Made the -sticky flag default to on/true.  And I changed the displayed title string.
               Now I want to refactor the code so it doesn't need a filename as an arg.  If no filename is given, it will default to the first one in its slice.
               Here that would be the most recent image file.
24 Aug 23 -- Removed the old version of main(), and edited some comments.  And added help output.
26 Aug 23 -- Added call to SetFullScreen(), then removed it because it made most images look terrible.
 7 Sep 23 -- Added reverse sort flag, that would make the oldest first.
19 Feb 25 -- Starting to think about adding a rotate image command, likely 'r'.  And I might as well use pflag instead of flag.
20 Feb 25 -- It works, and uses repeated hits of 'r' to rotate clockwise 90 deg each time.  Or can use '1', '2', '3', or '4' to directly rotate that number of degrees.
22 Feb 25 -- Added '=' to mean set scaleFactor=1 and zero the rotatedTimes variable.
23 Jul 25 -- Got idea to save an image in its current size and degree of rotation.  This may take some time to get right.
 1 Dec 25 -- Added fyne.Do, as was supposed to happen all along.
18 Jan 26 -- Added CenterOnScreen() to the window in the loadTheImage routine.
*/

const LastModified = "Jan 18, 2026"
const keyCmdChanSize = 20
const (
	firstImgCmd = iota
	prevImgCmd
	nextImgCmd
	loadImgCmd
	lastImgCmd
)

// const maxWidth = 1800 // actual resolution is 1920 x 1080   \ unused
// const maxHeight = 900 // actual resolution is 1920 x 1080   /

var index int
var loadedimg *canvas.Image
var cwd string
var imageInfo []os.FileInfo
var globalA fyne.App
var globalW fyne.Window
var verboseFlag = flag.BoolP("verbose", "v", false, "verbose flag.")
var zoomFlag = flag.BoolP("zoom", "z", false, "set zoom flag to allow zooming up a lot.")
var stickyFlag = flag.BoolP("sticky", "s", true, "sticky flag for keeping zoom factor among images.") // defaults to on as of 8/21/23
var reverseFlag = flag.BoolP("reverse", "r", false, "reverse sort flag, ie, oldest first.")
var sticky bool
var scaleFactor float64 = 1
var shiftState bool
var keyCmdChan chan int
var rotatedTimes int64 // used in keyTyped.  And atomicadd so need this type.
var imageAsDisplayed image.Image

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
	var err error
	flag.Usage = func() {
		executable, err := os.Executable()
		if err != nil {
			panic(err)
		}
		ExecFI, _ := os.Stat(executable)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf(" %s last altered %s, compiled with %s, and linked %s. \n",
			os.Args[0], LastModified, runtime.Version(), ExecTimeStamp)
		fmt.Printf(" Usage information:\n")
		fmt.Printf(" z = zoom and also toggles sticky.\n")
		fmt.Printf(" v = verbose.\n")
		fmt.Printf(" r = rotate 90º clockwise.\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	sticky = *zoomFlag || *stickyFlag

	// Set up the slice of imgFileInfo, which is []os.FileInfo, sorted w/ newest first.  This slice is set up as a go routine and the result is passed back here in a channel.
	// And define the 3 channels to be used here.  One is a keystroke channel, another is the imgFileInfo channel, and the 3rd is a image # channel.

	cwd, err = os.Getwd()
	if err != nil {
		fmt.Printf(" os.Getwd failed w/ error of %s\n", err)
		os.Exit(1)
	}

	keyCmdChan = make(chan int, keyCmdChanSize)
	imgFileInfoChan := make(chan []os.FileInfo, 1) // unbuffered channel increases latency.  Will make it buffered now.  It only needs a buffer of 1 because it only receives once.
	indexChan := make(chan int, 1)                 // I'm now making this buffered as I don't need a guarantee of receipt.  This may reduce latency.
	go MyReadDirForImages(cwd, imgFileInfoChan)    // this go routine is started here, and in a few lines the channel read is assigned to the global imageInfo.

	str := fmt.Sprintf("Image Viewer last modified %s, compiled using %s", LastModified, runtime.Version())
	if *verboseFlag {
		fmt.Println(str)
	}

	globalA = app.New() // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)

	// This is how I initialize the imageInfo slice of FileInfos
	imageInfo = <-imgFileInfoChan // reading from a channel into a global slice here.  Channel read is a unary use of the channel operator.  I hope this does not introduce latency.

	if *verboseFlag {
		if isSorted(imageInfo) {
			fmt.Println(" imageInfo slice of FileInfo is sorted.  Length is", len(imageInfo))
		} else {
			fmt.Println(" imageInfo slice of FileInfo is NOT sorted.  Length is", len(imageInfo))
		}
		fmt.Println()
	}

	// I have code here to read an image, show it's params, but not use that info after it's written to the screen.  I'll change this so all populating the main display window is done from
	// the loadTheImage routine.  I used to do it here first before going into the keystroke loop.  I don't need that anymore.  It's much more flexible if I don't do that, so then I can
	// define a default image number if no filename is given.  That file name would have to be searched against the imageInfo slice as a linear sequential search.  Can't do a binary search
	// as the sort function is not alphabetical.
	// If there is no argument, then default to the first image in the imageInfo slice.

	imgFilename := flag.Arg(0)
	baseFilename := filepath.Base(imgFilename)

	if imgFilename == "" {
		// index = 0  I don't need to set it to zero as that's its default.
	} else {
		go filenameIndex(imageInfo, baseFilename, indexChan)
		_, err = os.Stat(imgFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  \n", imgFilename, err)
		}

		if isNotImageStr(imgFilename) {
			fmt.Println(imgFilename, "does not have an image extension.  ")
		}

		index = <-indexChan // syntax to read from a channel, using the channel operator as a unary operator.
	}

	if index < 0 {
		index = 0
	}
	loadTheImage(index)
	//globalW.SetFullScreen(true)
	//globalW.SetFullScreen(false)

	go processKeys()

	globalW.ShowAndRun()

} // end main

// --------------------------------------------------- processKeys -------------------------------
func processKeys() {
	for {
		keyCmd := <-keyCmdChan
		//                             fmt.Println("in processKeys go routine.  keycmd =", keyCmd)
		switch keyCmd {
		case firstImgCmd:
			firstImage()
		case prevImgCmd:
			prevImage()
		case nextImgCmd:
			nextImage()
		case loadImgCmd:
			loadTheImage(index)
		case lastImgCmd:
			lastImage()
		}
	}
}

// loadTheImage -- loads the image given by the index
func loadTheImage(idx int) {
	//                                          imgname := imageInfo[index].Name()  where index was a global.  I changed taking input from a global.
	imgName := imageInfo[idx].Name()
	fullFilename, err := filepath.Abs(imgName)
	if err != nil {
		fmt.Printf(" loadTheImage(%d): error is %s.  imgName=%s, fullFilename is %s \n", idx, err, imgName, fullFilename)
	}

	imageURI := storage.NewFileURI(fullFilename)
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from storage.Reader(%s) is %s.  Skipped.\n", fullFilename, err)
		return
	}
	defer imgRead.Close() // moved to here, after checking err, as recommended by static linter.

	img, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format name used during format registration by the init function.
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from image.Decode is %s.  Skipped\n", err)
		return
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X

	title := fmt.Sprintf(" %s %s, %d x %d, SF=%.2f \n", imgFmtName, imgName, imgWidth, imgHeight, scaleFactor)
	if *verboseFlag {
		fmt.Println(title)
	}
	/*  According to Andy, this is unnecessary.
	if imgWidth > maxWidth {
		img = resize.Resize(maxWidth, 0, img, resize.Lanczos3)
		title = title + "; resized."
	} else if imgHeight > maxHeight {
		img = resize.Resize(0, maxHeight, img, resize.Lanczos3)
		title = title + "; resized."
	}
	bounds = img.Bounds()
	imgHeight = bounds.Max.Y
	imgWidth = bounds.Max.X
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
		//                                title = fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s\n", imgname, imgWidth, imgHeight, imgFmtName, cwd)
		title = fmt.Sprintf("%s, %d x %d, SF=%.2f, %s \n", imgName, imgWidth, imgHeight, scaleFactor, imgFmtName)
	}

	if *verboseFlag {
		bounds = img.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		fmt.Println(" Scalefactor =", scaleFactor, "last height =", imgHeight, "last width =", imgWidth)
		fmt.Printf(" loadTheImage(%d): imgName=%s, fullFilename is %s \n", idx, imgName, fullFilename)
		fmt.Println()
	}

	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.ScaleMode = canvas.ImageScaleSmooth
	if !*zoomFlag {
		loadedimg.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
		//loadedimg.FillMode = canvas.ImageFillOriginal -- sets min size to be that of the original.
	}

	imageAsDisplayed = loadedimg.Image

	atomic.StoreInt64(&rotatedTimes, 0) // reset this counter when load a fresh image.

	fyne.Do(func() { // I was getting warnings from fyne about this being called from a non-GUI thread.
		// safe to touch widgets here
		globalW.SetContent(loadedimg)
		globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))
		globalW.SetTitle(title)
		globalW.CenterOnScreen() // added 1/18/26.  To see if it works.  It does.  I'm guessing it works because I'm centering the window after calling SetContent.
		globalW.Show()
	})

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
	//return  also redundant
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
			imgInfo, err := os.Lstat(dir + string(filepath.Separator) + name)
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

	if *reverseFlag {
		sortfcn = func(i, j int) bool {
			return fi[i].ModTime().Before(fi[j].ModTime()) // I want a oldest-first sort.  Added 9/7/23.
		}
	}

	sort.Slice(fi, sortfcn)
	elapsedtime := time.Since(t0)

	if *verboseFlag {
		fmt.Printf(" Length of the image fileinfo slice is %d, and sorted in %s\n", len(fi), elapsedtime.String())
		fmt.Println()
	}

	imageInfoChan <- fi
	// return  also redundant
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
// func nextImage(indx int) *canvas.Image {
func nextImage() {
	index++
	if index >= len(imageInfo) {
		index--
	}
	loadTheImage(index)
	// return  also redundant
} // end nextImage

// ------------------------------------------ prevImage -------------------------------------------------------
// func prevImage(indx int) *canvas.Image {
func prevImage() {
	index--
	if index < 0 {
		index++
	}
	loadTheImage(index)
	// return  also redundant
} // end prevImage

// ------------------------------------------ firstImage -----------------------------------------------------
func firstImage() {
	index = 0
	loadTheImage(index)
}

// ------------------------------------------ lastImage ---------------------------------------------------------
func lastImage() {
	index = len(imageInfo) - 1
	loadTheImage(index)
}

// ------------------------------------------------------------ keyTyped ------------------------------
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
		globalW.Close() // quits the app if this is the last window, which it is.
		//		globalA.Quit()
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
	case fyne.KeyEqual: // first added Feb 22, 2025.  I thought I had the from the beginning.  So it goes.
		scaleFactor = 1
		atomic.StoreInt64(&rotatedTimes, 0) // reset this counter when load a fresh image.
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
		*verboseFlag = !*verboseFlag
		fmt.Printf(" Verbose flag is now %t, Sticky is %t, and scaleFactor is %2.2g\n", *verboseFlag, sticky, scaleFactor)
	case fyne.KeyZ:
		sticky = !sticky
		*verboseFlag = !*verboseFlag
		if *verboseFlag {
			fmt.Println(" Sticky is now", sticky, "and scaleFactor is", scaleFactor)
		}
	case fyne.KeyR:
		atomic.AddInt64(&rotatedTimes, 1)
		rotateAndLoadTheImage(index, rotatedTimes) // index and rotatedTimes are global
	case fyne.Key1:
		rotatedTimes = 1
		rotateAndLoadTheImage(index, 1)
	case fyne.Key2:
		rotatedTimes = 2
		rotateAndLoadTheImage(index, 2)
	case fyne.Key3:
		rotatedTimes = 3
		rotateAndLoadTheImage(index, 3)
	case fyne.Key4, fyne.Key0:
		atomic.StoreInt64(&rotatedTimes, 0) // reset this counter when load a fresh image.
		rotateAndLoadTheImage(index, 0)

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

// rotateTheImage -- loads the image given by the index, and then rotates it before displaying it.
func rotateAndLoadTheImage(idx int, repeat int64) {
	imgName := imageInfo[idx].Name()
	fullFilename, err := filepath.Abs(imgName)
	if err != nil {
		fmt.Printf(" loadTheImage(%d): error is %s.  imgName=%s, fullFilename is %s \n", idx, err, imgName, fullFilename)
	}

	//imageURI := storage.NewFileURI(fullFilename)
	//imgRead, err := storage.Reader(imageURI)
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
	repeat = repeat % 4 // modulus operator
	switch repeat {     // trying a new way, just to see if it works
	case 0:
		imgImg = imgRead
	case 1:
		rotatedImg = imaging.Rotate90(imgRead)
		imgImg = imgImage(rotatedImg)
	case 2:
		rotatedImg = imaging.Rotate180(imgRead)
		imgImg = imgImage(rotatedImg)
	case 3:
		rotatedImg = imaging.Rotate270(imgRead)
		imgImg = imgImage(rotatedImg)
	}

	bounds := imgImg.Bounds()
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

	canvasImage := canvas.NewImageFromImage(imgImg)
	canvasImage.ScaleMode = canvas.ImageScaleSmooth
	if !*zoomFlag {
		canvasImage.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
		//                                                 loadedimg.FillMode = canvas.ImageFillOriginal -- sets min size to be that of the original.
	}

	imageAsDisplayed = canvasImage.Image

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
	sizeStr := fmt.Sprintf("%dx%d_rot_%d", imgWidth, imgHeight, rotatedTimes)
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
	sizeStr := fmt.Sprintf("%dx%d_rot_%d", imgHeight, imgWidth, rotatedTimes)
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
