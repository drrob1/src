/*
imginfo.go

REVISION HISTORY
-------- -------
22 Aug 23 -- Now called imginfo.go.  I pulled most of this code from img.go.  I will determine the info of the image using the 2 ways I found.
             One is using the std library, and the other uses fyne code.

*/

package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	//"fyne.io/fyne/v2/internal/widget"
	//"fyne.io/fyne/v2/layout"
	//"fyne.io/fyne/v2/container"
	//"image/color"

	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2/storage"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const LastModified = "Aug 22, 2023"

var globalA fyne.App
var globalW fyne.Window

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, " Usage: imginfo <image file name>")
		os.Exit(1)
	}

	str := fmt.Sprintf("Image Info, last modified %s, compiled using %s", LastModified, runtime.Version())
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

	baseFilename := filepath.Base(imgfilename)
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
	fmt.Printf(" image.Config %s, %s, %s \n width=%g, height=%g, and aspect ratio=%.4g.  \n",
		imgfilename, fullFilename, baseFilename, width, height, aspectRatio)

	//fmt.Printf(" Type for DecodeConfig is %T \n", imgConfig) // answer is image.Config

	cwd := filepath.Dir(fullFilename)
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" os.Getwd failed w/ error of %s\n", err)
		os.Exit(1)
	}
	if cwd != workingDir {
		ctfmt.Printf(ct.Red, false, " cwd=%q, should equal workingDir=%q.  Don't know why these are not equal.  Will ignore workingDir.\n")
	}

	globalA = app.New() // this line must appear before any other uses of fyne.  But I may not need it if I don't load an image.
	globalW = globalA.NewWindow(str)
	globalW.Canvas().SetOnTypedKey(keyTyped)

	imageURI := storage.NewFileURI(fullFilename) // needs to be a type = fyne.CanvasObject
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullFilename, "is", err)
		os.Exit(1)
	}
	defer imgRead.Close()
	img, imgFmtName, err := image.Decode(imgRead) // imgFmtName is a string of the format name used during format registration by the init function.
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from image.Decode is", err)
		os.Exit(1)
	}
	bounds := img.Bounds()
	imgHeight := bounds.Max.Y
	imgWidth := bounds.Max.X
	fmt.Printf(" %s image.Decode: %dw x %dh, min X = %d, min Y = %d\n", imgFmtName, imgWidth, imgHeight, bounds.Min.X, bounds.Min.Y)

	imgTitle := fmt.Sprintf("%s, %d x %d", imgfilename, imgWidth, imgHeight)
	loadedImg := canvas.NewImageFromImage(img)
	loadedImg.ScaleMode = canvas.ImageScaleFastest

	globalW.SetTitle(imgTitle)
	globalW.SetContent(loadedImg)
	globalW.CenterOnScreen()
	globalW.Resize(fyne.NewSize(float32(imgWidth), float32(imgHeight)))
	globalW.ShowAndRun()

} // end main

// -------------------------------------------------------------------------------------------------------------------------------------------

func isImage(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

// -------------------------------------------------------- isNotImageStr ----------------------------------------
func isNotImageStr(name string) bool {
	return !isImage(name)
}

// ------------------------------------------------------------ keyTyped ------------------------------

func keyTyped(e *fyne.KeyEvent) { // index and shiftState are global var's
	switch e.Name {
	case fyne.KeyUp:
		globalW.Close()
	case fyne.KeyDown:
		globalA.Quit()
	case fyne.KeyLeft:
		globalA.Quit()
	case fyne.KeyRight:
		globalW.Close()
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
	case fyne.KeyHome:
		globalA.Quit()
	case fyne.KeyEnd:
		globalA.Quit()
	case fyne.KeyPageUp:
		globalA.Quit()
	case fyne.KeyPageDown:
		globalA.Quit()
	case fyne.KeyPlus, fyne.KeyAsterisk:
		globalA.Quit()
	case fyne.KeyMinus:
		globalA.Quit()
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		globalA.Quit()
	case fyne.KeyBackspace: // preserve always resetting zoomfactor here.  Hope I remember I'm doing this.
		globalW.Close()
	case fyne.KeySlash:
		globalW.Close()
	case fyne.KeyV:
		globalW.Close()
	case fyne.KeyZ:
		globalA.Quit()

	default:
		globalA.Quit()
	}
} // end keyTyped

/*
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
		*verboseFlag = !*verboseFlag
		fmt.Printf(" Verbose flag is now %t, Sticky is %t, and scaleFactor is %2.2g\n", *verboseFlag, sticky, scaleFactor)
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

*/
