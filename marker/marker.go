package main // marker.go from highlight.go

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"github.com/spf13/pflag"
)

/*
  REVISION HISTORY
  -------- -------
  30 Nov 25 -- First version.  Copied from Linux Magazine 299, Oct 2025, that I read in Aruba Nov 2025.
			I had to upgrade fyne to the current version to get it to compile.  Current is 2.7.1.  The old version was 2.2.3.
			This works by creating an overlay on top of the image.  The overlay is a rectangle that can be dragged to select a region.
			Here, the purpose of the overlay is to highlight the region of interest.

			From the article, the overlay is at the top of the stack, and the image is at the bottom.  These 2 layers form a composite in which
			the lower layer is visible unless the upper layer defines something to prevent this.  Initially, the top layer is translucent and the image below is displayed.
			The upper layer is clickable.  When the Open Image button callback loads a new image from the disk, the program uses Objects[0] to access the stack's container
			element array, grabs the first element (index 0) and replaces it w/ the image just loaded.  Refresh() is called so the container displays the new image.
			This is at lines 198 and 199 below.

			The button horizontal container is located vertically above the stack that holds both the image and the overlay.

			Back to the Original
            The layer design is ideal for more complex highlighting (e.g., additive selections that let the user mark several blocks at once).  Unfortunately, though, the original image's
            code has been lost, because I shrunk the image at the outset to display it in a more compact way on the desktop.  On top of this, the selection resides in a layer at a higher
            level.  You may be wondering how the code flattens the image later.  The idea is to save the modified image file including the markers on disk.

            The original file still exists in the main program's memory, and the highlighting is also there, but scaled down.  The highlighting needs to scale to match when the
			scaled-down image is enlarged again.  This is true for the rectangle origin point (x,y) and its size (h and w).

			The loadImage() function trimmed the image to a width of 800 pixels and a height resulting from a constant aspect ratio.  If the original width of the image was X pixels,
			this results in a downscaling factor of X/800.  To transfer the highlighting in the display layer to the original image, the code needs to multiply both the X,Y position
			and the width, height of the marker by the scaling factor.

			The AsImage() function expects the scaling factor as an artument for this stretching operation; it draws a rectangle w/ the new dimensions.  Note that Go's image package
			expects the coordinates of a rectangle in a different format than the Fyne framework.  While Fyne requires the top left corner as the origin and the length and width of the
			rectangle, the image package defines Min in the image.Point format as the starting point in the upper left corner and the coordinates of the bottom right corner as Max.

			Both are legitimate formats, but the code has to translate between the 2 formats.  The return of AsImage() uses RGB values and an alpha value to define the yellow-green
            highlighting color while keeping the text below it shining through.

			To save a screenshot with highlighting back to disk in its original format, the SaveBig() function clones the original image to big; this is because Go's image.Image
			structure does not allow any modifications by default.  The Clone() command returns a structure that uses the same interface, but allows drawing with Draw().
			The draw.Over param superimposes the semi-transparent rectangle over the image file on disk, like in the GUI.

			The three standard commands (go mod init, go mod tidy, go build) create the binary.  Then you just need to move the finalized GUI program to a path in the $PATH environment variable.
            There is nothing to stop you from highlighting images.  You can add code as you see fit.  Want to highlight more sections while holding down the Super key or save in a
			different path?  Go for it.
  13 Dec 25 -- Added saving to a random name, not to overwrite the original.  I needed to use Junie AI chat to learn how to convert from path string to a listableURI.
				I don't like the built-in fyne open dialog.  There is no way to narrow down the search.  I will use Junie AI chat to learn how to create a custom dialog.
  17 Dec 25 -- Separated out the AI-generated code for a custom menu to open files, so I can try to understand it better.
				I still don't understand it well.  So I asked perplexity.  I think I understand it now.  It's all about the SetFilter function that must return a bool.
  18 Dec 25 -- Added code to make sure it's a picture file, using the regexp below.  Now I'm adding the code to only show files that match the current base name, still using the
				standard dialog (that I hate).  It takes the base name of the file opened by the open file button function.
  19 Dec 25 -- Now called marker.go, from highlight.go.  The magazine article wanted it called marker after all.  I'm going to keep highlight as working code, and play more here.
                And I included the contents of fileopenstuff.go here, so I don't have a separate file here, so far.
  22 Dec 25 -- Junie told me that I can resize dialog boxes from code.  So I resized both the built-in file open dialog box and the custom box.  I like the custom box better
*/

const lastModified = "22 Dec 25"
const width = 800
const height = 600

var picRegexp *regexp.Regexp = regexp.MustCompile(`(?i)\.(jpg|jpeg|png|gif|bmp|webp)$`) // I added this line to make the search case insensitive, by AI.

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

// AsImage function expects the scaling factor as an artument for this stretching operation; it draws a rectangle w/ the new dimensions.  Note that Go's image package
// expects the coordinates of a rectangle in a different format than the Fyne framework.  While Fyne requires the top left corner as the origin and the length and width of the
// rectangle, the image package defines Min in the image.Point format as the starting point in the upper left corner and the coordinates of the bottom right corner as Max.
func (r *Rect) AsImage(zoom float64) image.Rectangle {
	pos, size := r.Dims()
	x := pos.X * float32(zoom)
	y := pos.Y * float32(zoom)
	w := size.Width * float32(zoom)
	h := size.Height * float32(zoom)
	rect := image.Rectangle{Min: image.Point{X: int(x), Y: int(y)}, Max: image.Point{X: int(x) + int(w), Y: int(y) + int(h)}}

	return rect
}

func (r *Rect) Color() color.NRGBA {
	return color.NRGBA{R: 204, G: 255, B: 0, A: 50}
}

func NewOverlay() *Overlay {
	over := Overlay{}
	over.ExtendBaseWidget(&over)            // This turns the overlay into a widget that detects mouse clicks and drags.
	over.con = container.NewWithoutLayout() // This is an empty container that holds the rectangle later in the code.
	over.rect = NewRect()
	return &over // pointer semantics
}

func (t *Overlay) CreateRenderer() fyne.WidgetRenderer { // This uses the container's standard renderer.   So now creating the customized widget is done.
	return widget.NewSimpleRenderer(t.con)
}

// Dragged is the function signature for the dragged event.
// The signature is required by the fyne.Draggable interface.
// When the user drags the mouse, this function is called because the dragged event is registered with the overlay by the function signature.
// The event is repeated continuously and sends the mouse pointer's current coordinates to the function avery few miliseconds as it moves over the surface.
// The sequence ends when the mouse button is released, and the DragEnd event signals the event to the callback function.
func (t *Overlay) Dragged(e *fyne.DragEvent) {
	if !t.inMotion { // the article explicitly compares to false.  StaticCheck cried foul and said to change it.
		t.inMotion = true
		t.rect.From = e.Position
		return
	}
	t.rect.To = e.Position
	pos, size := t.rect.Dims()
	t.DrawMarker(pos, size) // draws the rectangle each time the mouse moves.  It receives the new coordinates from the Dragged event each time it is called.
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
	t.con.Remove(t.marker) // remove previous rectangle
	t.marker = rect        // replace with new rectangle
	t.con.Add(rect)        // store the new rectangle into the container, using the current coordinates and size.
	t.Refresh()
}

// SaveBig -- To save a screenshot with highlighting back to disk in its original format, the SaveBig() function clones the original image to big; this is because Go's
// image.Image structure does not allow any modifications by default.  The Clone() command returns a structure that uses the same interface, but allows drawing with Draw().
// The draw.Over param superimposes the semi-transparent rectangle over the image file on disk, like in the GUI.
func (t *Overlay) SaveBig(big image.Image, path string) error {
	dimg := imaging.Clone(big)
	r := t.rect.AsImage(t.zoom)
	draw.Draw(dimg, r, &image.Uniform{t.rect.Color()}, r.Min, draw.Over)
	i := rand.IntN(1_000_000)
	s := strconv.Itoa(i)
	// b := strings.TrimSuffix(base, filepath.Ext(base)) is an alternate way to get the base name.
	ext := filepath.Ext(path)
	baseName := path[:len(path)-len(ext)]
	savedFilename := baseName + "_" + s + ext
	dialog.ShowInformation("Saved Filename is", savedFilename, w)
	err := imaging.Save(dimg, baseName+"_"+s+ext)
	return err // my simplification.  A linter told me to return err instead of if err != nil return err and then return nil.
}

// LoadImage
// The loadImage() function trims the image to a width of 800 pixels and a height resulting from a constant aspect ratio.  If the original width of the image was X pixels,
// this results in a downscaling factor of X/800.  To transfer the highlighting in the display layer to the original image, the code needs to multiply both the X, Y position
// and the width, height of the marker by the scaling factor.
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

var w fyne.Window // global so other functions have access to it.
func main() {
	a := app.NewWithID("com.example.Image_Highlighter")
	s := fmt.Sprintf("Image Highlighter, last modified %s, compiled with %s", lastModified, runtime.Version())
	w = a.NewWindow(s)
	w.Resize(fyne.NewSize(width, height))
	ov := NewOverlay()
	img := &canvas.Image{}
	var big image.Image
	var imgPath, basenameSearchStr string

	pflag.StringVarP(&basenameSearchStr, "search", "s", "", "String to search base filename against.  If not specified, the program opens the first image file in the current directory.")
	pflag.Parse()

	if pflag.NArg() > 0 {
		imgPath = pflag.Arg(0)
		file, err := os.Open(imgPath)
		if err != nil {
			panic(err)
		}
		defer file.Close() // this is not in the article.  The AI here added it.
		big, img = ov.LoadImage(file)
	}

	if basenameSearchStr == "" { // this will allow either setting the search string as a command line argument or not.
		basenameSearchStr = pflag.Arg(1)
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

	workingDir, err := os.Getwd()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	curURI, err := listableFromPath(workingDir)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	/*
	   Usage in your app (e.g., in marker.go):
	*/
	openFunc2 := func(r fyne.URI) {
		if r == nil {
			return
		}
		f, er := storage.Reader(r)
		if er != nil {
			dialog.ShowError(er, w)
			return
		}
		defer f.Close()

		imgPath = r.Path()
		ext := filepath.Ext(imgPath)
		basenameSearchStr = filepath.Base(imgPath)
		basenameSearchStr = strings.TrimSuffix(basenameSearchStr, ext)
		big, img = ov.LoadImage(f)
		stack.Objects[0] = img // the bottom of the stack, at position [0], is the image.
		stack.Refresh()
	}

	openBtn2 := widget.NewButton("Open…2", func() {
		NewOpenFileDialogWithPrefix(w, basenameSearchStr, []string{".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp"}, openFunc2)
	})

	fileOpenFunc := func(reader fyne.URIReadCloser, err error) { // this closure gets called AFTER the user has selected a file from the fyne dialog.
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		imgPath = reader.URI().Path()
		ext := filepath.Ext(imgPath)
		basenameSearchStr = filepath.Base(imgPath)
		basenameSearchStr = strings.TrimSuffix(basenameSearchStr, ext)
		big, img = ov.LoadImage(reader)
		stack.Objects[0] = img // the bottom of the stack, at position [0], is the image.
		stack.Refresh()
	}
	//openBtnFunc := func() { // article example.
	//	dialog.NewFileOpen(fileOpenFunc, w).Show()
	//}
	openBtnFunc := func() { // I want to specify starting directory 1st
		openDialog := dialog.NewFileOpen(fileOpenFunc, w)
		openDialog.SetLocation(curURI)
		// openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".gif", ".bmp", "webp"})) This line was ignored.  Can't have 2 filtering cond's.
		openDialog.SetFilter(&nameFilterType{search: basenameSearchStr}) // I hope if this is empty, then it matches everything.  And I hope I can have 2 filtering conditions.
		openDialog.Resize(fyne.NewSize(width, width))
		openDialog.Show()
	}
	openBtn := widget.NewButton("Open Image", openBtnFunc)

	saveBtnFunc := func() {
		ov.SaveBig(big, imgPath)
	}
	saveBtn := widget.NewButton("Save Image", saveBtnFunc)

	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	buttons := container.NewHBox(openBtn, openBtn2, saveBtn, quitBtn)

	w.SetContent(container.NewVBox(buttons, stack))

	w.ShowAndRun()

}

func listableFromPath(path string) (fyne.ListableURI, error) {
	u := storage.NewFileURI(path)
	listerURI, err := storage.ListerForURI(u)
	return listerURI, err
}

type FileFilterI interface {
	Matches(fyne.URI) bool
}

type nameFilterType struct {
	search string
}

func (f nameFilterType) Matches(u fyne.URI) bool { //  I'm going to add check against a directory
	// base name without directories
	name := u.Name()

	isDir, _ := storage.CanList(u) // this doesn't prevent the directories from being populated also
	if isDir {
		return false
	}
	isPic := picRegexp.MatchString(name)
	if !isPic {
		return false
	}

	// for case-insensitive substring match:
	return strings.Contains(
		strings.ToLower(name),
		strings.ToLower(f.search),
	)
}
