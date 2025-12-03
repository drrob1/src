package main // highlight.go

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

/*
  REVISION HISTORY
  -------- -------
  30 Nov 25 -- First version.  Copied from Linux Magazine 299, Oct 2025, that I read in Aruba Nov 2025.
				I had to upgrade fyne to the current version to get it to compile.  Current is 2.7.1.  The version was 2.2.3.

*/

const lastModified = "3 Dec 25"
const width = 800
const height = 600

type Overlay struct {
	widget.BaseWidget
	con      *fyne.Container
	marker   *canvas.Rectangle
	rect     *Rect
	inMotion bool
	zoom     float64
}

type Rect struct {
	From fyne.Position
	To   fyne.Position
	Zoom float64
}

func NewOverlay() *Overlay {
	over := Overlay{}
	over.ExtendBaseWidget(&over)
	over.con = container.NewWithoutLayout()
	over.rect = NewRect()
	return &over // pointer semantics
}

//func NewOverlay() *Overlay { Original function.  I rewrote it in the format recommended by Bill Kennedy.
//	over := &Overlay{}
//	over.ExtendBaseWidget(over)
//	over.con = container.NewWithoutLayout()
//	over.rect = NewRect()
//	return over // pointer semantics
//}

func NewRect() *Rect {
	return &Rect{}
}

func (r *Rect) Dims() (fyne.Position, fyne.Size) {
	x := r.From.X
	y := r.From.Y
	w := r.To.X - r.From.X
	h := r.To.Y - r.From.Y
	if r.To.X < r.From.X {
		x = r.To.X
		w = -w
	}
	if r.To.Y < r.From.Y {
		y = r.To.Y
		h = -h
	}

	return fyne.NewPos(x, y), fyne.NewSize(w, h)
}

func (r *Rect) AsImage(zoom float64) image.Rectangle {
	pos, size := r.Dims()
	x := pos.X * float32(zoom)
	y := pos.Y * float32(zoom)
	w := size.Width * float32(zoom)
	h := size.Height * float32(zoom)
	rect := image.Rectangle{Min: image.Point{X: int(x), Y: int(y)}, Max: image.Point{X: int(x + w), Y: int(y + h)}}

	return rect
}

func (r *Rect) Color() color.NRGBA {
	return color.NRGBA{R: 204, G: 255, B: 0, A: 50}
}

func (t *Overlay) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.con)
}

func (t *Overlay) Dragged(e *fyne.DragEvent) {
	if t.inMotion == false {
		t.inMotion = true
		t.rect.From = e.Position
		return
	}
	t.rect.To = e.Position
	pos, size := t.rect.Dims()
	t.DrawMarker(pos, size)
}

func (t *Overlay) DragEnd() {
	t.inMotion = false
}

func (t *Overlay) DrawMarker(pos fyne.Position, size fyne.Size) {
	rect := canvas.NewRectangle(t.rect.Color())
	rect.Resize(size)
	rect.Move(pos)
	if t.marker == nil {
		t.marker = rect
		return
	}
	t.con.Remove(t.marker)
	t.marker = rect
	t.con.Add(rect)
	t.Refresh()
	return
}

func (t *Overlay) SaveBig(big image.Image, path string) error {
	dimg := imaging.Clone(big)
	r := t.rect.AsImage(t.zoom)
	draw.Draw(dimg, r, &image.Uniform{t.rect.Color()}, r.Min, draw.Over)
	err := imaging.Save(dimg, path)
	return err
}

func (t *Overlay) LoadImage(r io.Reader) (image.Image, *canvas.Image) {
	big, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		panic(err)
	}
	shrunk := imaging.Resize(big, width, 0, imaging.Lanczos)
	t.zoom = float64(big.Bounds().Dx()) / float64(width)
	img := canvas.NewImageFromImage(shrunk)
	img.FillMode = canvas.ImageFillOriginal

	return big, img
}

func main() {
	a := app.NewWithID("com.example.imagehighlighter")
	s := fmt.Sprintf("Image Highlighter, last modified %s, compiled with %s", lastModified, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(width, height))
	ov := NewOverlay()
	img := &canvas.Image{}
	var big image.Image
	var imgPath string

	if len(os.Args) == 2 {
		imgPath = os.Args[1]
		file, err := os.Open(imgPath)
		if err != nil {
			panic(err)
		}
		defer file.Close() // this is not in the article.  The AI here added it.
		big, img = ov.LoadImage(file)
	}

	stack := container.NewStack(img, ov)
	typedKey := func(ev *fyne.KeyEvent) { // I separated this out so I can more easily understand it.
		key := string(ev.Name)
		switch key {
		case "Q", "Escape", "X":
			os.Exit(0)
		}
	}
	w.Canvas().SetOnTypedKey(typedKey)

	fileOpenFunc := func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		imgPath = reader.URI().Path()
		big, img = ov.LoadImage(reader)
		stack.Objects[0] = img
		stack.Refresh()
	}
	openBtnFunc := func() {
		dialog.NewFileOpen(fileOpenFunc, w).Show()
	}
	openBtn := widget.NewButton("Open Image", openBtnFunc)

	saveBtnFunc := func() {
		ov.SaveBig(big, imgPath)
	}
	saveBtn := widget.NewButton("Save Image", saveBtnFunc)

	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	buttons := container.NewHBox(openBtn, saveBtn, quitBtn)

	w.SetContent(container.NewVBox(buttons, stack))

	w.ShowAndRun()

}
