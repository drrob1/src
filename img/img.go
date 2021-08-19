// From Go GUI with Fyne, Chap 4.  I believe it will be enhanced in later chapters, but this is what is it for now.
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
*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2/app"
	//w "fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/layout"
	//"image/color"

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

	//_ "golang.org/x/image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	//"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"

	"github.com/nfnt/resize"
)

const LastModified = "August 18, 2021"
const maxWidth = 2500
const maxHeight = 2000

var index int
var loadedimg *canvas.Image
var cwd string
var imageInfo []os.FileInfo
var globalA fyne.App
var globalW fyne.Window

func isNotImageStr(name string) bool {
	ext := filepath.Ext(name)
	isImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
	return !isImage
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, " Usage: img <image file name>")
		os.Exit(1)
	}

	str := fmt.Sprintf("Single Image Viewer last modified %s, compiled using %s", LastModified, runtime.Version())
	fmt.Println(str) // this works as intended

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
	var aspectRatio = width/height
	if aspectRatio > 1 {
		aspectRatio = 1/aspectRatio
	}

	fmt.Println(" image.Config", imgfilename, fullFilename, basefilename, "width =", width, ", height =", height, "and aspect ratio =", aspectRatio)

	if width > maxWidth || height > maxHeight {
		width = maxWidth * aspectRatio
		height = maxHeight * aspectRatio
	}

	fmt.Println()
	//fmt.Printf(" Type for DecodeConfig is %T \n", imgConfig) // answer is image.Config
	fmt.Println(" adjusted image.Config width =", width, ", height =", height, " but these values are not used to show the image.")
	fmt.Println()


	cwd = filepath.Dir(fullFilename)
	imgFileInfoChan := make(chan []os.FileInfo)  // unbuffered channel
	go MyReadDirForImages(cwd, imgFileInfoChan)

	globalA = app.New()  // this line must appear before any other uses of fyne.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)

	imageURI := storage.NewFileURI(fullFilename) // needs to be a type = fyne.CanvasObject
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullFilename, "is", err)
		os.Exit(1)
	}
	defer imgRead.Close()
	img, imgFmtName, err := image.Decode(imgRead)  // imgFmtName is a string of the format name used during format registration by the init function.
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from image.Decode is", err)
		os.Exit(1)
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X
	fmt.Println(" image.Decode, width=", imgWidth, "and height=", imgHeight, ", imgFmtName=", imgFmtName, "and cwd=", cwd)
	fmt.Println()
	if imgWidth > maxWidth {
		img = resize.Resize(maxWidth, 0, img, resize.Lanczos3)
	} else if imgHeight > maxHeight {
		img = resize.Resize(0, maxHeight, img, resize.Lanczos3)
	}

	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.FillMode = canvas.ImageFillContain

	imgtitle := fmt.Sprintf("%s, %d x %d", imgfilename, imgWidth, imgHeight)
	globalW.SetTitle(imgtitle)
	globalW.SetContent(loadedimg)
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))

	select { // this syntax works and is blocking.
	case imageInfo = <- imgFileInfoChan :  // this ackward syntax is what's needed to read from a channel.
	}

	t0:= time.Now()
	sortfcn := func(i, j int) bool { // this is a closure anonymous function
		return imageInfo[i].ModTime().After(imageInfo[j].ModTime()) // I want a newest-first sort.  Changed 12/20/20
	}
	sort.Slice(imageInfo, sortfcn)
	elapsedtime := time.Since(t0)

	fmt.Printf(" Length of the image fileinfo slice is %d, and sorted in %s\n", len(imageInfo), elapsedtime.String())
	fmt.Println()

	t0 = time.Now()

	indexchan := make(chan int)
	go filenameIndex(imageInfo, basefilename, indexchan)
	select {
	case index = <- indexchan:
	}
	elapsedtime = time.Since(t0)

	fmt.Printf(" %s index is %d in the fileinfo slice; linear sequential search took %s.\n", basefilename, index, elapsedtime)
	fmt.Printf(" As a check, imageInfo[%d] = %s.\n", index, imageInfo[index].Name())
	fmt.Println()

	globalW.CenterOnScreen()
	globalW.ShowAndRun()

} // end main

// --------------------------------------------------- loadTheImage ------------------------------
                                                     //func loadTheImage(indx int) *canvas.Image {
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

	img, imgFmtName, err := image.Decode(imgRead)  // imgFmtName is a string of the format name used during format registration by the init function.
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from image.Decode is", err)
		os.Exit(1)
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X
	title := fmt.Sprintf("%s width=%d, height=%d, type=%s and cwd=%s\n", imgname, imgWidth, imgHeight, imgFmtName, cwd)
	fmt.Println(title)
	if imgWidth > maxWidth {
		img = resize.Resize(maxWidth, 0, img, resize.Lanczos3)
	} else if imgHeight > maxHeight {
		img = resize.Resize(0, maxHeight, img, resize.Lanczos3)
	}
	loadedimg = canvas.NewImageFromImage(img)
	loadedimg.FillMode = canvas.ImageFillContain
	globalW.SetContent(loadedimg)
	globalW.SetTitle(title)
	globalW.Show()
	return
}


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


// ----------------------------------isImage ----------------------------------------------
func isImage(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	ext = strings.ToLower(ext)

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif"
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

	imageInfoChan <- fi
	return
} // MyReadDirForImages

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

// ------------------------------------------ lastImage
func lastImage() {
	index = len(imageInfo) - 1
	loadTheImage()
}

// ------------------------------------------------------------ keyTyped ------------------------------
func keyTyped(e *fyne.KeyEvent) { // index is a global var
	switch e.Name {
	case fyne.KeyUp:
		prevImage()
	case fyne.KeyDown:
		nextImage()
	case fyne.KeyLeft:
		prevImage()
	case fyne.KeyRight:
		nextImage()
	case fyne.KeyEscape:
		globalW.Close()  // quit's the app if this is the last window, which it is.
//		(*globalA).Quit()
	case fyne.KeyHome:
		firstImage()
	case fyne.KeyEnd:
		lastImage()
	}
}

