package main

import (
	"bytes"
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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/godoes/printers"
	_ "golang.org/x/image/webp"
)

/*
15 July 26 -- Created by Codex.
*/

const (
	targetWidth  = 1800
	targetHeight = 900
)

var imageExts = map[string]bool{
	".jpeg": true,
	".jpg":  true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

type imageEntry struct {
	name    string
	path    string
	modTime time.Time
}

var (
	globalApp     fyne.App
	globalWindow  fyne.Window
	viewerImage   *canvas.Image
	imageEntries  []imageEntry
	currentIndex  int
	currentScale  float64 = 1
	currentAuto   float64 = 1
	currentRotate int
	currentImage  image.Image
	shiftState    bool
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: %s [starting image filename]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
		os.Exit(1)
	}

	imageEntries, err = scanImages(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan images: %v\n", err)
		os.Exit(1)
	}
	if len(imageEntries) == 0 {
		fmt.Fprintln(os.Stderr, "no images found in the current directory")
		os.Exit(1)
	}

	startName := ""
	if flag.NArg() > 0 {
		startName = flag.Arg(0)
	}
	currentIndex = resolveStartIndex(imageEntries, startName)

	globalApp = app.New()
	globalWindow = globalApp.NewWindow("Image Viewer")
	globalWindow.Canvas().SetOnTypedKey(keyTyped)
	globalWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("HELP",
			fyne.NewMenuItem("HELP", showHelp),
			fyne.NewMenuItem("ABOUT", showAbout),
			fyne.NewMenuItem("PRINT", showPrintDialog),
			fyne.NewMenuItem("QUIT", func() { globalWindow.Close() }),
		),
	))

	if err := renderCurrent(true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	globalWindow.ShowAndRun()
}

func scanImages(dir string) ([]imageEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	images := make([]imageEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !imageExts[strings.ToLower(filepath.Ext(name))] {
			continue
		}
		full := filepath.Join(dir, name)
		fi, err := os.Stat(full)
		if err != nil {
			continue
		}
		images = append(images, imageEntry{
			name:    name,
			path:    full,
			modTime: fi.ModTime(),
		})
	}

	sort.Slice(images, func(i, j int) bool {
		if images[i].modTime.Equal(images[j].modTime) {
			return strings.ToLower(images[i].name) < strings.ToLower(images[j].name)
		}
		return images[i].modTime.After(images[j].modTime)
	})

	return images, nil
}

func resolveStartIndex(entries []imageEntry, start string) int {
	if start == "" {
		return 0
	}
	base := filepath.Base(start)
	for i, entry := range entries {
		if entry.name == base || strings.EqualFold(entry.name, base) {
			return i
		}
	}
	return 0
}

func renderCurrent(resetScale bool) error {
	entry := imageEntries[currentIndex]

	img, err := imaging.Open(entry.path, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("open %s: %w", entry.name, err)
	}

	if currentRotate != 0 {
		img = rotateImage(img, currentRotate)
	}

	autoScale := fitScale(img)
	if resetScale {
		currentScale = autoScale
	} else if approxEqual(currentScale, currentAuto) {
		currentScale = autoScale
	}
	currentAuto = autoScale

	rendered := scaleImage(img, currentScale)
	currentImage = rendered
	updateWindow(entry, rendered)
	return nil
}

func rotateImage(img image.Image, quarterTurns int) image.Image {
	switch quarterTurns % 4 {
	case 1:
		return imaging.Rotate90(img)
	case 2:
		return imaging.Rotate180(img)
	case 3:
		return imaging.Rotate270(img)
	default:
		return img
	}
}

func scaleImage(img image.Image, scale float64) image.Image {
	if scale <= 0 || approxEqual(scale, 1) {
		return img
	}
	b := img.Bounds()
	w := uint(math.Max(1, math.Round(float64(b.Dx())*scale)))
	h := uint(math.Max(1, math.Round(float64(b.Dy())*scale)))
	return imaging.Resize(img, int(w), int(h), imaging.Lanczos)
}

func fitScale(img image.Image) float64 {
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return 1
	}
	ws := float64(targetWidth) / float64(b.Dx())
	hs := float64(targetHeight) / float64(b.Dy())
	if ws < hs {
		return ws
	}
	return hs
}

func updateWindow(entry imageEntry, img image.Image) {
	b := img.Bounds()
	title := fmt.Sprintf("%s  %dx%d  scale=%.2f  rot=%d", entry.name, b.Dx(), b.Dy(), currentScale, currentRotate*90)
	globalWindow.SetTitle(title)

	if viewerImage == nil {
		viewerImage = canvas.NewImageFromImage(img)
		viewerImage.FillMode = canvas.ImageFillOriginal
		globalWindow.SetContent(viewerImage)
	} else {
		viewerImage.Image = img
		viewerImage.Refresh()
	}

	size := fyne.NewSize(float32(b.Dx()), float32(b.Dy()))
	viewerImage.Resize(size)
	globalWindow.Resize(size)
	globalWindow.CenterOnScreen()
}

func showHelp() {
	text := strings.Join([]string{
		"HELP",
		"left/up = previous image",
		"right/down = next image",
		"home = first image",
		"end = last image",
		"x/q/escape = quit",
		"pageup/+ = magnify 10%",
		"pagedown/- = minify 10%",
		"= = 100%",
		"r = rotate 90 degrees clockwise",
		"s/w = save current image",
	}, "\n")
	dialog.ShowInformation("HELP", text, globalWindow)
}

func showAbout() {
	text := "Image viewer written in Go using Fyne."
	dialog.ShowInformation("ABOUT", text, globalWindow)
}

func showPrintDialog() {
	if runtime.GOOS != "windows" {
		dialog.ShowInformation("PRINT", "Printing is only supported on Windows.", globalWindow)
		return
	}

	names, err := printers.ReadNames()
	if err != nil {
		dialog.ShowError(fmt.Errorf("read printers: %w", err), globalWindow)
		return
	}
	if len(names) == 0 {
		dialog.ShowInformation("PRINT", "No installed printers were found.", globalWindow)
		return
	}

	selectWidget := widget.NewSelect(names, nil)
	selectWidget.PlaceHolder = "Choose a printer"
	if defaultPrinter, err := printers.GetDefault(); err == nil {
		selectWidget.SetSelected(defaultPrinter)
	}
	if selectWidget.Selected == "" {
		selectWidget.SetSelected(names[0])
	}

	dialog.NewCustomConfirm("PRINT", "PRINT", "CANCEL",
		container.NewVBox(widget.NewLabel("Choose an installed printer:"), selectWidget),
		func(ok bool) {
			if !ok {
				return
			}
			if err := printCurrent(selectWidget.Selected); err != nil {
				dialog.ShowError(err, globalWindow)
			}
		},
		globalWindow,
	).Show()
}

func keyTyped(e *fyne.KeyEvent) {
	if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
		shiftState = true
		return
	}
	if shiftState {
		shiftState = false
		if e.Name == fyne.KeyEqual {
			currentScale *= 1.1
			_ = renderCurrent(false)
		} else if e.Name == fyne.KeyPeriod {
			nextImage()
		} else if e.Name == fyne.KeyComma {
			prevImage()
		} else if e.Name == fyne.Key8 {
			currentScale *= 1.1
			_ = renderCurrent(false)
		}
		return
	}

	switch e.Name {
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalWindow.Close()

	case fyne.KeyLeft, fyne.KeyUp:
		prevImage()
	case fyne.KeyRight, fyne.KeyDown, fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		nextImage()
	case fyne.KeyHome:
		firstImage()
	case fyne.KeyEnd:
		lastImage()

	case fyne.KeyPageUp, fyne.KeyPlus, fyne.KeyAsterisk:
		currentScale *= 1.1
		_ = renderCurrent(false)
	case fyne.KeyPageDown, fyne.KeyMinus, fyne.KeySlash:
		currentScale *= 0.9
		_ = renderCurrent(false)
	case fyne.KeyEqual:
		currentScale = 1
		_ = renderCurrent(false)

	case fyne.KeyR:
		currentRotate = (currentRotate + 1) % 4
		_ = renderCurrent(false)
	case fyne.KeyS, fyne.KeyW:
		if err := saveCurrent(); err != nil {
			dialog.ShowError(err, globalWindow)
		}
	case fyne.KeyV:
		// no-op: reserved for future verbose output
	}
}

func prevImage() {
	if currentIndex > 0 {
		currentIndex--
	}
	currentRotate = 0
	_ = renderCurrent(true)
}

func nextImage() {
	if currentIndex < len(imageEntries)-1 {
		currentIndex++
	}
	currentRotate = 0
	_ = renderCurrent(true)
}

func firstImage() {
	currentIndex = 0
	currentRotate = 0
	_ = renderCurrent(true)
}

func lastImage() {
	currentIndex = len(imageEntries) - 1
	currentRotate = 0
	_ = renderCurrent(true)
}

func saveCurrent() error {
	if currentImage == nil {
		return fmt.Errorf("no image loaded")
	}
	entry := imageEntries[currentIndex]
	ext := filepath.Ext(entry.name)
	base := strings.TrimSuffix(entry.name, ext)
	b := currentImage.Bounds()
	saveName := fmt.Sprintf("%s_%dx%d_rot%d%s", base, b.Dx(), b.Dy(), currentRotate*90, ext)
	full := filepath.Join(filepath.Dir(entry.path), saveName)

	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()

	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, currentImage, &jpeg.Options{Quality: 95})
	case ".png":
		return png.Encode(f, currentImage)
	case ".gif":
		return gif.Encode(f, currentImage, nil)
	case ".webp":
		return webp.Encode(f, currentImage, &webp.Options{Lossless: false, Quality: 95})
	default:
		return fmt.Errorf("unsupported format for save: %s", ext)
	}
}

func printCurrent(printerName string) error {
	if currentImage == nil {
		return fmt.Errorf("no image loaded")
	}
	if printerName == "" {
		return fmt.Errorf("no printer selected")
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, currentImage, &jpeg.Options{Quality: 95}); err != nil {
		return err
	}

	p, err := printers.Open(printerName)
	if err != nil {
		return err
	}
	defer p.Close()

	if err := p.StartDocument("Image Print", "RAW"); err != nil {
		return err
	}
	if err := p.StartPage(); err != nil {
		return err
	}
	if _, err := p.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := p.EndPage(); err != nil {
		return err
	}
	if err := p.EndDocument(); err != nil {
		return err
	}
	return nil
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}
