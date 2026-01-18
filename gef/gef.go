package main // gef.go, meaning gastric emptying in fyne
import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"src/whichexec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

/*
  25 Dec 25 -- This is going to be a GUI interface for my gastric emptying pgm, currently at version 3.
                My intention is to enter the data in a multiline entry box, save it to a gastric file, run gastric3, and then show the output.
                I may show both the standard output and the gastric file in separate widgets.
				Then I added an open file button, so I can use a file I've already created.  Primarily for testing.
  26 Dec 25 -- At work, when I made a desktop shortcut, the pgm couldn't find gastric3.exe.  Fixed.
				It does work if started from the command line.
  29 Dec 25 -- Trying embedding an icon of a stomach.  I found a relevant image on the internet and then converted it to .png using Irfanview.
  18 Jan 26 -- Added a default file name for the save dialog, using the SetFileName method.
*/

const lastModified = "Jan 18, 2026"
const width = 1600
const height = 900
const minRowsVisible = 40

var w fyne.Window // global so other functions have access to it.

//go:embed assets/stomach-4.png
var stomachIcon []byte

func main() {
	a := app.NewWithID("com.example.Gastric_Emptying_GUI")

	stomIconRes := fyne.NewStaticResource("stomach-4.png", stomachIcon)
	a.SetIcon(stomIconRes)
	//a.SetIcon(theme.FyneLogo())
	s := fmt.Sprintf("Gastric Emptying v 3, last modified %s, compiled with %s", lastModified, runtime.Version())
	w = a.NewWindow(s)
	w.Resize(fyne.NewSize(width, height))

	editWidget := widget.NewMultiLineEntry() // for entering the gastric emptying data
	editWidget.SetMinRowsVisible(minRowsVisible)
	editWidget.SetPlaceHolder("Enter gastric emptying data here. Need long string to size the window correctly.")
	editWidget.Resize(fyne.NewSize(width/2, height))

	filenameWidget := widget.NewEntry() // for entering the name of the output file
	filenameWidget.PlaceHolder = " Enter output gastric-<whatever>.txt filename"

	outputWidget := widget.NewMultiLineEntry() // for displaying the output of the gastric emptying computations
	outputWidget.SetMinRowsVisible(minRowsVisible)
	outputWidget.SetPlaceHolder("Gastric output will appear here. Need long string to size the window correctly.")
	outputWidget.Resize(fyne.NewSize(width/2, height))

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

	var filenameToSave string

	//   Using the custom open file dialog box
	openFileFunc := func(r fyne.URI) {
		if r == nil {
			return
		}
		f, er := storage.Reader(r)
		if er != nil {
			dialog.ShowError(er, w)
			return
		}
		defer f.Close()

		filenameToSave = r.Path()
		filenameWidget.SetText(filenameToSave)

		data, err := io.ReadAll(f)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		editWidget.SetText(string(data))
	}

	openFileBtn := widget.NewButton("Open gastric data file", func() {
		NewOpenFileDialogWithPrefix(w, "gastric", []string{".txt", ".text"}, openFileFunc)
	})

	showSaveFunc := func(win fyne.Window, data []byte) {
		writeCallback := func(wr fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if wr == nil { // user cancelled
				return
			}
			defer wr.Close()

			filenameToSave = wr.URI().Path()
			filenameWidget.SetText(filenameToSave)

			if _, err := wr.Write(data); err != nil {
				dialog.ShowError(err, win)
				return
			}
		}
		wrDialog := dialog.NewFileSave(writeCallback, win)
		wrDialog.SetLocation(curURI)
		wrDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".text"}))
		wrDialog.SetFileName("gastric-.txt") // this is a template file name to be set by user
		wrDialog.Show()
	}

	saveBtnFunc := func() {
		showSaveFunc(w, []byte(editWidget.Text))
	}
	saveBtn := widget.NewButton("Save gastric emptying file", saveBtnFunc)

	gastricBtnFunc := func() {
		if filenameToSave == "" {
			dialog.ShowError(fmt.Errorf("No filename specified.  Need to save file first."), w)
			return
		}
		_, err = os.Stat(filenameToSave)
		if err != nil {
			dialog.ShowError(fmt.Errorf("File %s does not exist.", filenameToSave), w)
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error from os.UserHomeDir: %v", err), w)
			return
		}
		//cmdPath := whichexec.Find("gastric3.exe", homeDir)  I forgot that I don't need the .exe part, that's automatically added on Windows.
		cmdPath := whichexec.Find("gastric3", homeDir)
		if cmdPath == "" {
			dialog.ShowError(fmt.Errorf("could not find gastric3.exe"), w)
			return
		}
		if cmdPath == "" {
			dialog.ShowError(fmt.Errorf("could not find gastric3.exe"), w)
			return
		}

		screen := bytes.NewBuffer(make([]byte, 0, 2048))
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creating gastricoutputfile.out: %v", err), w)
			return
		}

		gEmptyingCmd := exec.Command(cmdPath, filenameToSave)
		gEmptyingCmd.Stdout = screen
		gEmptyingCmd.Stderr = os.Stderr
		err = gEmptyingCmd.Run()
		if err != nil {
			dialog.ShowError(fmt.Errorf("error running gastric3.exe: %v", err), w)
			return
		}
		outputWidget.SetText(screen.String())
		outputWidget.Refresh()
	}

	gastricBtn := widget.NewButton("Run gastric3", gastricBtnFunc)

	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	buttons := container.NewHBox(openFileBtn, saveBtn, gastricBtn, quitBtn)
	gridbox := container.NewGridWithColumns(2, editWidget, outputWidget)
	vbox := container.NewVBox(buttons, filenameWidget, gridbox)
	w.SetContent(vbox)

	w.ShowAndRun()
}

func listableFromPath(path string) (fyne.ListableURI, error) {
	u := storage.NewFileURI(path)
	listerURI, err := storage.ListerForURI(u)
	return listerURI, err
}
