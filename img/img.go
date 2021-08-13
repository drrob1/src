// From Go GUI with Fyne, Chap 4.  I believe it will be enhanced in later chapters, but this is what is it for now.
/*
REVISION HISTORY
-------- -------
 9 Aug 21 -- I realized that this will not be enhanced, as I went thru more of the book.  I'll have to enhance it myself.
             First, I'm changing the function constants to the version that's more readable to me.  That's working, but I had to
             import more parts of fyne.io than the unmodified version.
12 Aug 21 -- Now called img.go, so I can display 1 image.  I'll start here.
*/

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	//_ "golang.org/x/image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/nfnt/resize"
)

const LastModified = "August 13, 2021"
const maxWidth = 3000
const maxHeight = 2000

type bgImageLoad struct {
	uri fyne.URI
	img *canvas.Image
}

var loads = make(chan bgImageLoad, 1024)

var imgInfoChan = make(chan []os.FileInfo)  // unbuffered channel

func scaleImage(img image.Image) image.Image {
	return resize.Thumbnail(320, 240, img, resize.Lanczos3)
}

func doLoadImage(u fyne.URI, img *canvas.Image) {
	//read, err := storage.OpenFileFromURI(u)  I'm getting a message from goland that this function is depracated.
	read, err := storage.Reader(u)
	if err != nil {
		log.Println("Error opening image", err)
		return
	}

	defer read.Close()
	raw, _, err := image.Decode(read)
	if err != nil {
		log.Println("Error decoding image", err)
		return
	}

	img.Image = scaleImage(raw)
	img.Refresh()
}

func doLoadImages() {
	for load := range loads {
		doLoadImage(load.uri, load.img)
	}
}

func loadImage(u fyne.URI) fyne.CanvasObject {
	img := canvas.NewImageFromResource(nil)
	img.FillMode = canvas.ImageFillContain

	loads <- bgImageLoad{u, img} // typed constant
	return img
}

type itemLayout struct {
	bg, text, gradient fyne.CanvasObject
}

type imgLayout struct {
	bg, text, gradient fyne.CanvasObject
}

func (i *itemLayout) MinSize([]fyne.CanvasObject) fyne.Size { // I removed an underscore from this line that I thought was a mistake.
	return fyne.NewSize(160, 120)
}

func (i *itemLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	textHeight := float32(22)
	for _, o := range objs {
		if o == i.text {
			o.Move(fyne.NewPos(0, size.Height-textHeight))
			o.Resize(fyne.NewSize(size.Width, textHeight))
		} else if o == i.bg {
			o.Move(fyne.NewPos(0, size.Height-textHeight))
			o.Resize(fyne.NewSize(size.Width, textHeight))
		} else if o == i.gradient {
			o.Move(fyne.NewPos(0, size.Height-(textHeight*1.5)))
			o.Resize(fyne.NewSize(size.Width, textHeight/2))
		} else {
			o.Move(fyne.NewPos(0, 0))
			o.Resize(size)
		}
	}
}

func makeImageItem(u fyne.URI) fyne.CanvasObject {
	label := canvas.NewText(u.Name(), color.Gray{128})
	label.Alignment = fyne.TextAlignCenter

	bgColor := &color.NRGBA{R: 255, G: 255, B: 255, A: 224}
	bg := canvas.NewRectangle(bgColor)
	fade := canvas.NewLinearGradient(color.Transparent, bgColor, 0)
	return container.New(&itemLayout{text: label, bg: bg, gradient: fade},
		loadImage(u), bg, fade, label)
}

func isImage(file fyne.URI) bool {
	ext := strings.ToLower(file.Extension())

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp"
}

func filterImages(files []fyne.URI) []fyne.URI {
	images := []fyne.URI{}

	for _, file := range files {
		if isImage(file) {
			images = append(images, file)
		}
	}

	return images
}

func makeImageGrid(images []fyne.URI) fyne.CanvasObject {
	items := []fyne.CanvasObject{}

	for _, item := range images {
		img := makeImageItem(item)
		items = append(items, img)
	}

	cellSize := fyne.NewSize(160, 120)
	return container.NewGridWrap(cellSize, items...)
}

func makeStatus(dir fyne.ListableURI, images []fyne.URI) fyne.CanvasObject {
	status := fmt.Sprintf("Directory %s, %d items", dir.Name(), len(images))
	return canvas.NewText(status, color.Gray{128})
}

func makeUI(dir fyne.ListableURI) fyne.CanvasObject {
	list, err := dir.List()
	if err != nil {
		log.Println("Error listing directory", err)
	}
	images := filterImages(list)
	status := makeStatus(dir, images)
	content := container.NewScroll(makeImageGrid(images))
	return container.NewBorder(nil, status, nil, nil, status, content)
}

func chooseDirectory(w fyne.Window) {
	listableURIfunc := func(dir fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		w.SetContent(makeUI(dir))
	}
	dialog.ShowFolderOpen(listableURIfunc, w)
}

/*
	dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		w.SetContent(makeUI(dir))
	}, w)
*/

func startDirectory() fyne.ListableURI {
	flag.Parse()
	if len(flag.Args()) < 1 {
		cwd, _ := os.Getwd()
		list, _ := storage.ListerForURI(storage.NewFileURI(cwd))
		return list
	}

	dir, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Println("Could not find directory", dir)
		cwd, _ := os.Getwd()
		list, _ := storage.ListerForURI(storage.NewFileURI(cwd))
		return list
	}

	list, _ := storage.ListerForURI(storage.NewFileURI(dir))
	return list
}

func isNotImage(name string) bool {
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

	imgfilename := flag.Arg(0)
	_, err := os.Stat(imgfilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from os.Stat(", imgfilename, ") is", err)
		os.Exit(1)
	}

	if isNotImage(imgfilename) {
		fmt.Fprintln(os.Stderr, imgfilename, "does not have an image extension.")
		os.Exit(1)
	}

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

	width := imgConfig.Width
	height := imgConfig.Height

	if width > maxWidth {
		width = maxWidth
	}
	if height > maxHeight {
		height = maxHeight
	}

	fmt.Println()
	fmt.Printf(" Type for DecodeConfig is %T \n", imgConfig) // answer is image.Config
	fmt.Println(" Image", imgfilename, fullFilename, "width =", width, " and height =", height)
	fmt.Println()

	a := app.New()
	str := fmt.Sprintf("Single Image Viewer last modified %s, compiled using %s", LastModified, runtime.Version())
	fmt.Println(str) // this works as intended
	w := a.NewWindow(str)
	w.Resize(fyne.NewSize(maxWidth, maxHeight))
	w.Show()

        cwd, _ := os.Getwd()
	imageURI := storage.NewFileURI(fullFilename) // needs to be a type = fyne.CanvasObject
	imgRead, err := storage.Reader(imageURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from storage.Reader of", fullFilename, "is", err)
		os.Exit(1)
	}
	defer imgRead.Close()
	img, imgFmtName, err := image.Decode(imgRead)  // imgFmtName is a string of the format name used during format registration by the init function.
	bounds := img.Bounds()
	imgHeight := bounds.Max.X
	imgWidth := bounds.Max.Y
	fmt.Println(" Using image.Decode, width=", imgWidth, "and height=", imgHeight, "and imgFmtName=", imgFmtName)
	fmt.Println()

	label := canvas.NewText(imgfilename, color.Gray{128})
	label.Alignment = fyne.TextAlignCenter
	imgRect := canvas.NewRectangle(color.Black)


        if imgWidth > maxWidth || imgHeight > maxHeight {
            img = resize.Resize(0, maxWidth, resize.Lanczos3)
        }

        imgPic := canvas.NewImageFromResource(img)  // code above has nil here
        imgPic.FillMode = canvas.ImageFillContain
	imgAndTitle := container.NewBorder(nil, label, nil, nil, imgRect) // top, left and right are all nil here.  bottom=label, center=imgRect

/*
	bgColor := &color.NRGBA{R: 255, G: 255, B: 255, A: 224}
	bg := canvas.NewRectangle(bgColor)
	fade := canvas.NewLinearGradient(color.Transparent, bgColor, 0)
	imagelayoutliteral := &imgLayout{text: label, bg: bg, gradient: fade} // needs to be a type = fyne.Layout
	picContainer := container.New(imagelayoutliteral, uri, bg, fade, label)
*/
	//	w.SetContent(makeUI(startDirectory()))
	//	w.Resize(fyne.NewSize(480, 360))
	//    w.SetFullScreen(true)

	//	chooseDirFunc := func() {
	//		chooseDirectory(w)
	//	}
	//	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", fyne.NewMenuItem("Open Directory...", chooseDirFunc))))

	//	go doLoadImages()
        go MyReadDirForImages(cwd, imageInfoChan)
	w.ShowAndRun()

        imageInfo := make([]os.FileInfo,0, 1024)

        for {
            select {
            case  imageinfo <- imageInfoChan:  break
            default:
                    // do nothing but don't block.
            }
        }

        fmt.Println(" Have the slice of image file infos.  Len =", len(imageingo))
        fmt.Println()


} // end main

// ----------------------------------isImage // ----------------------------------------------
func isImage(file string) bool {
	ext := strings.ToLower(file.Extension())
        ext = strings.ToLower(ext)

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif"
}

// ------------------------------- MyReadDirForImages -----------------------------------
func MyReadDirForImages(dir string, imageInfoChan chan []os.FileInfo) {

	dirname, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return nil
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

