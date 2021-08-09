// From Go GUI with Fyne, Chap 4.  I believe it will be enhanced in later chapters, but this is what is it for now.
/*
REVISION HISTORY
-------- -------
 9 Aug 21 -- I realized that this will not be enhanced, as I went thru more of the book.  I'll have to enhance it myself.
             First, I'm changing the function constants to the version that's more readable to me.


*/

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/nfnt/resize"
)

const LastModified = "August 10, 2021"

type bgImageLoad struct {
	uri fyne.URI
	img *canvas.Image
}

var loads = make(chan bgImageLoad, 1024)

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

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif"
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

func main() {
	a := app.New()
	w := a.NewWindow("Image Browser")

	w.SetContent(makeUI(startDirectory()))
	w.Resize(fyne.NewSize(480, 360))

	chooseDirFunc := func() {
		chooseDirectory(w)
	}
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", fyne.NewMenuItem("Open Directory...", chooseDirFunc))))

	go doLoadImages()
	w.ShowAndRun()
}

/*

	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File",
		fyne.NewMenuItem("Open Directory...", func() {
			chooseDirectory(w)
		}))))
*/
