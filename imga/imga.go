package main

import (
	"flag"
	"fmt"
	"image"
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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

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
26 Aug 21 -- Now called imga.go, to be sorted alphabetically as a slice of strings, not FileInfo's.
               This allows use of more of the library functions.
22 Sep 21 -- Added zoomfactor, and will account for shiftState.  And maxWidth, maxHeight more closely match the monitor limits of 1920 x 1080.
27 Sep 21 -- Added stickyFlag, sticky and 'z' zoom toggle.  When sticky is true, zoom factor is not cleared automatically.
30 Sep 21 -- Added KeyAsterisk and removed redundant code (as per Andy Williams)
 4 Dec 21 -- Adding a go routine to process the keystrokes.  And adding "v" to turn on verbose mode.  And other things from img.go.
16 Mar 22 -- Only writing using fmt.Print calls if verboseFlag is set.
26 Mar 22 -- Handles correctly when dir is not current dir; I did not need to port the code from img.go as it always worked here.
               It works because the sort is alphabetical, not by date, so I don't need to call Lstat.
21 Nov 22 -- Fixed some issues flagged by static linter.
21 Aug 23 -- Made the -sticky flag default to on.  And added a ScaleFactor value to the display window's title.
25 Aug 23 -- Will time how long it take to create the slice of filenames in MyReadDir.  It's ~1% of the time that img2.go takes, because here os.Lstat isn't used.
               And removed the duplicate code in main() that loads an image.
20 Feb 25 -- Porting code from img.go to here, allowing manual rotation of an image using repeated hits of 'r' to rotate clockwise 90 deg, or '1', '2', or '3'.
			It's too late now; I'll do this tomorrow.
			Added the rotateAndLoadImage and imgImage procedures, modified keyTyped and loadTheImage.  Fetching the image names is done w/ one goroutine; this is fast enough.
22 Feb 25 -- Added '=' to mean set scaleFactor=1 and zero the rotatedTimes variable.
24 Jul 25 -- Added ability to save an image in its current size and degree of rotation.  Developed first in img.go.
26 Jul 25 -- Using a different method to trim off ext of base filename, just to see if it works.  It does.
 1 Dec 25 -- Added fyne.Do, as was supposed to happen long ago.
18 Jan 26 -- Now centering output in loadTheImage.  Tested in img.go and it works, so I'm porting it here.
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

var index int
var loadedimg *canvas.Image
var cwd string
var imageInfo sort.StringSlice
var globalA fyne.App
var globalW fyne.Window
var verboseFlag = flag.Bool("v", false, "verbose flag")
var zoomFlag = flag.Bool("z", false, "set zoom flag to allow zooming up a lot")
var stickyFlag = flag.Bool("sticky", true, "sticky flag for keeping zoom factor among images.") // default changed to true 8/21/23
var sticky bool
var scaleFactor float64 = 1
var shiftState bool
var keyCmdChan chan int
var rotatedCtr int64 // used in keyTyped.  And atomicadd so need this type.
var imageAsDisplayed image.Image

// -------------------------------------------------------- isNotImageStr ----------------------------------------
func isNotImageStr(name string) bool {
	isImg := isImage(name)
	return !isImg
}

// ----------------------------------isImage ----------------------------------------------
func isImage(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	ext = strings.ToLower(ext)

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

// ---------------------------------------------------- main --------------------------------------------------
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

	str := fmt.Sprintf("Image Viewer last modified %s, compiled using %s", LastModified, runtime.Version())
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
	keyCmdChan = make(chan int, keyCmdChanSize)
	imgFileInfoChan := make(chan []string, 1) // unbuffered channel increases latency.  Will make it buffered now.  It only needs a buffer of 1 because it only receives once.

	go MyReadDirForImagesAlphabetically(cwd, imgFileInfoChan)

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

	if flag.NArg() >= 1 {
		go filenameAlphaIndex(imageInfo, baseFilename, indexChan)

		_, err := os.Stat(imgFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from os.Stat(%s) is %s.  \n", imgFilename, err)
		}

		if isNotImageStr(imgFilename) {
			fmt.Fprintln(os.Stderr, imgFilename, "does not have an image extension.")
		}

		index = <-indexChan // syntax to read from a channel.
	}
	if index < 0 {
		index = 0
	}

	loadTheImage()

	go processKeys()

	globalW.ShowAndRun()

} // end main

// --------------------------------------------------- processKeys -------------------------------
func processKeys() {
	for {
		switch <-keyCmdChan {
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
	imgname := imageInfo[index]
	fullfilename := cwd + string(filepath.Separator) + imgname
	imageURI := storage.NewFileURI(fullfilename)
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullfilename, "is", err)
		os.Exit(1)
	}
	defer imgRead.Close() // moved as recommended by static linter

	img, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format type, ie jpeg, png, webm used during format registration by the init function.
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from image.Decode is", err)
		os.Exit(1)
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X

	//              title := fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s\n", imgname, imgWidth, imgHeight, imgFmtName, cwd)
	title := fmt.Sprintf("%s, %d x %d, SF=%2f; %s \n", imgname, imgWidth, imgHeight, scaleFactor, imgFmtName)
	if *verboseFlag {
		fmt.Println(title, "and cwd=", cwd)
	}
	/* redundant code
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
		//                                     title = fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s\n", imgname, imgWidth, imgHeight, imgFmtName, cwd)
		title = fmt.Sprintf("%s, %d x %d, type=%s \n", imgname, imgWidth, imgHeight, imgFmtName)
	}

	if *verboseFlag {
		bounds = img.Bounds()
		imgHeight = bounds.Max.Y
		imgWidth = bounds.Max.X
		fmt.Println(" Scalefactor =", scaleFactor, "last height =", imgHeight, "last width =", imgWidth)
		fmt.Println()
	}

	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.ScaleMode = canvas.ImageScaleSmooth
	if !*zoomFlag {
		loadedimg.FillMode = canvas.ImageFillContain // this must be after the image is assigned else there's distortion.  And prevents blowing up the image a lot.
		//loadedimg.FillMode = canvas.ImageFillOriginal -- sets min size to be that of the original.
	}

	imageAsDisplayed = loadedimg.Image

	atomic.StoreInt64(&rotatedCtr, 0) // reset this counter when load a fresh image.

	fyne.Do(func() { // I was getting warnings from fyne about this being called from a non-GUI thread.
		// safe to touch widgets here
		globalW.SetContent(loadedimg)
		globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))
		globalW.SetTitle(title)
		globalW.CenterOnScreen() // added 1/18/26.  To see if it works.  It does.  I'm guessing it works because I'm centering the window after calling SetContent.
		globalW.Show()
	})

} // end loadTheImage

// ------------------------------- filenameAlphaIndex --------------------------------------
func filenameAlphaIndex(files sort.StringSlice, name string, intchan chan int) {
	if !sort.IsSorted(files) {
		intchan <- -1
		return
	}
	index := sort.SearchStrings(files, name)
	if index == len(files) { // this will not catch many not found errors, but I don't know what else to look for
		index = -1
	}
	intchan <- index
	// return
}

/*
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
*/
// ------------------------------- MyReadDirForImagesAlphabetically -----------------------------------

func MyReadDirForImagesAlphabetically(dir string, imageInfoChan chan []string) {
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
	imgNamesSlice := make([]string, 0, len(names))
	for _, name := range names {
		if isImage(name) {
			imgNamesSlice = append(imgNamesSlice, name) // note that this does not call Lstat for the files, so it's much faster than in img.go and img2.go.  Dir of 23K images take 3 ms here, but ~280 ms in img2.go.
		}
	}

	sort.Strings(imgNamesSlice)
	elapsedtime := time.Since(t0)

	if *verboseFlag {
		fmt.Printf(" Length of the image fileinfo slice is %d; created and sorted in %s\n", len(imgNamesSlice), elapsedtime.String())
		fmt.Println()
	}

	imageInfoChan <- imgNamesSlice
	// return
} // MyReadDirForImagesAlphabetically
/*
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


*/

// ------------------------------------------------------- isSorted -----------------------------------------------
//func isSortedAlpha(slice sort.StringSlice) bool {
//	return sort.IsSorted(slice)
//}

// ---------------------------------------------- nextImage -----------------------------------------------------
func nextImage() {
	index++
	if index >= len(imageInfo) {
		index--
	}
	loadTheImage()
	// return
} // end nextImage

// ------------------------------------------ prevImage -------------------------------------------------------
func prevImage() {
	index--
	if index < 0 {
		index++
	}
	loadTheImage()
	// return
} // end prevImage

// ------------------------------------------ firstImage -----------------------------------------------------
func firstImage() {
	index = 0
	loadTheImage()
}

// ------------------------------------------ lastImage
func lastImage() {
	index = len(imageInfo) - 1
	loadTheImage()
}

// ------------------------------------------------------------ keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index and shiftState are global var's
	switch e.Name {
	case fyne.KeyW:
		baseName := imageInfo[index]
		err := saveImage(imageAsDisplayed, baseName)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from saveImage(%s) is %s.  Skipped.\n", baseName, err)
		}
	case fyne.KeyS:
		baseName := imageInfo[index]
		err := imageSave(imageAsDisplayed, baseName)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from saveImage(%s) is %s.  Skipped.\n", baseName, err)
		}

	case fyne.KeyUp:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- prevImgCmd
	case fyne.KeyDown:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- nextImgCmd
	case fyne.KeyLeft:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- prevImgCmd
	case fyne.KeyRight:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- nextImgCmd
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//														(*globalA).Quit()
	case fyne.KeyHome:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- firstImgCmd
	case fyne.KeyEnd:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- lastImgCmd
	case fyne.KeyPageUp:
		scaleFactor *= 1.1 // I'm reversing what I did before.  PageUp now scales up
		keyCmdChan <- loadImgCmd
	case fyne.KeyPageDown:
		scaleFactor *= 0.9 // I'm reversing what I did before.  PageDn now scales down
		keyCmdChan <- loadImgCmd
	case fyne.KeyPlus, fyne.KeyAsterisk:
		scaleFactor *= 1.1
		keyCmdChan <- loadImgCmd
	case fyne.KeyMinus:
		scaleFactor *= 0.9
		keyCmdChan <- loadImgCmd
	case fyne.KeyEqual: // first added Feb 22, 2025.  I thought I had the from the beginning.  So it goes.
		scaleFactor = 1
		atomic.StoreInt64(&rotatedCtr, 0) // reset this counter when load a fresh image.
		keyCmdChan <- loadImgCmd
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		if !sticky {
			scaleFactor = 1
		}
		keyCmdChan <- nextImgCmd
	case fyne.KeyBackspace: // preserve always resetting zoomfactor here.  Hope I remember I'm doing this.
		scaleFactor = 1
		keyCmdChan <- prevImgCmd
	case fyne.KeySlash:
		scaleFactor *= 0.9
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
		atomic.AddInt64(&rotatedCtr, 1)
		rotateAndLoadTheImage(index, rotatedCtr) // index and rotatedCtr are global
	case fyne.Key1:
		rotatedCtr = 1
		rotateAndLoadTheImage(index, 1)
	case fyne.Key2:
		rotatedCtr = 2
		rotateAndLoadTheImage(index, 2)
	case fyne.Key3:
		rotatedCtr = 3
		rotateAndLoadTheImage(index, 3)
	case fyne.Key4, fyne.Key0:
		atomic.StoreInt64(&rotatedCtr, 0) // reset this counter when load a fresh image.
		rotateAndLoadTheImage(index, 4)

	default:
		if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
			shiftState = true
			return
		}
		if shiftState {
			shiftState = false
			if e.Name == fyne.KeyEqual { // shift equal is key plus
				scaleFactor *= 1.1
				keyCmdChan <- loadImgCmd
			} else if e.Name == fyne.KeyPeriod { // >
				keyCmdChan <- nextImgCmd
			} else if e.Name == fyne.KeyComma { // <
				keyCmdChan <- prevImgCmd
			} else if e.Name == fyne.Key8 { // *
				scaleFactor *= 1.1
				keyCmdChan <- loadImgCmd
			}
		} else if e.Name == fyne.KeyEqual {
			scaleFactor = 1
			keyCmdChan <- loadImgCmd
		}
	}
} // end keyTyped

// rotateAndLoadTheImage -- loads the image given by the index, and then rotates it before displaying it.
func rotateAndLoadTheImage(idx int, repeat int64) {
	imgName := imageInfo[idx]
	fullFilename, err := filepath.Abs(imgName)
	if err != nil {
		fmt.Printf(" loadTheImage(%d): error is %s.  imgName=%s, fullFilename is %s \n", idx, err, imgName, fullFilename)
	}

	imgRead, err := imaging.Open(fullFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from storage.Reader(%s) is %s.  Skipped.\n", fullFilename, err)
		return
	}

	var rotatedImg *image.NRGBA
	var imgImg image.Image

	for range repeat {
		rotatedImg = imaging.Rotate90(imgRead)
		imgImg = imgImage(rotatedImg) // need to convert from *image.NRGBA to image.Image
		imgRead = imgImg
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
		return fmt.Errorf("image passed to saveImage is nil")
	}
	ext := filepath.Ext(inputname)
	bounds := img.Bounds()
	imgWidth := bounds.Max.X
	imgHeight := bounds.Max.Y
	sizeStr := fmt.Sprintf("%dx%d_rot_%d", imgHeight, imgWidth, rotatedCtr)
	savedName := strings.TrimSuffix(inputname, ext) + "_saved_" + sizeStr + ext

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
