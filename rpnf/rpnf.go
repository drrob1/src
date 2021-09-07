// From Fyne GUI book by Andrew Williams, Chapter 6, widget.go

package main

import (
	"encoding/gob"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"runtime"
	"src/hpcalc2"
	"src/tknptr"
	"strconv"
	"strings"
	"time"

	//"fyne.io/fyne/v2/data/binding"
)

const lastModified = "Sep 5, 2021"

var globalA fyne.App
var globalW fyne.Window

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var cyan = color.NRGBA{R: 0, G: 255, B: 255, A: 255}
var lightcyan = color.NRGBA{R: 224, G: 255, B: 255, A: 255}

var homeDir string
var windowsFlag bool
var Storage [36]float64 // 0 ..  9, a ..  z
var DisplayTape, stringslice []string

const Storage1FileName = "RPNStorage.gob" // Allows for a rotation of Storage files, in case of a mistake.
const Storage2FileName = "RPNStorage2.gob"
const Storage3FileName = "RPNStorage3.gob"
const DisplayTapeFilename = "displaytape.txt"

func main() {
	fmt.Printf(" rpnf.go, using fyne.io v2.  Last modified %s, compiled using %s.\n", lastModified, runtime.Version())

	var Stk hpcalc2.StackType // used when time to write out the stack upon exit.
	var err error
	DisplayTape = make([]string, 0, 100)
	DisplayTape = append(DisplayTape, "History of entered commands")
	theFileExists := true

	homeDir, err = os.UserHomeDir() // this function became available as of Go 1.12
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error from os.UserHomeDir() is", err)
		os.Exit(1)
	}
	windowsFlag = runtime.GOOS == "windows"

	StorageFullFilenameSlice := make([]string, 5)
	StorageFullFilenameSlice[0] = homeDir
	StorageFullFilenameSlice[1] = string(os.PathSeparator)
	StorageFullFilenameSlice[2] = Storage1FileName
	StorageFullFilename := strings.Join(StorageFullFilenameSlice, "")
	/*
		Storage2FullFilenameSlice := homeDir + string(os.PathSeparator) + Storage2FileName
		Storage3FullFilenameSlice := homeDir + string(os.PathSeparator) + Storage3FileName
	*/

	var thefile *os.File

	thefile, err = os.Open(StorageFullFilename) // open for reading
	if err != nil {
		fmt.Printf(" Error from os.Open(Storage1FileName).  Possibly because no Stack File found: %v\n", err)
		theFileExists = false
	}
	defer thefile.Close()
	if theFileExists {
		decoder := gob.NewDecoder(thefile)         // decoder reads the file.
		err = decoder.Decode(&Stk)                 // decoder reads the file.
		check(err)                                 // theFileExists, so panic if this is an error.
		for i := hpcalc2.T1; i >= hpcalc2.X; i-- { // Push in reverse onto the stack in hpcalc2.
			hpcalc2.PUSHX(Stk[i])
		}

		err = decoder.Decode(&Storage) // decoder reads the file.
		check(err)                     // theFileExists, so panic if this is an error.

		thefile.Close()
	}

	globalA = app.New()
	globalW = globalA.NewWindow("Widget Binding")
	globalW.Canvas().SetOnTypedKey(keyTyped)
	R, _ := hpcalc2.GetResult("t")
	_, ss := hpcalc2.GetResult("dump")
	ssJoined := strings.Join(ss, "\n")
	//shorterSS := ss[1 : len(ss)-1] // removes the first and last strings, which are only character delims
	//shorterSSjoined := strings.Join(shorterSS, "\n")

	resultStr := strconv.FormatFloat(R, 'g', -1, 64)
	resultStr = hpcalc2.CropNStr(resultStr)
	resultLabel := canvas.NewText("X = "+resultStr, cyan)
	resultLabel.TextSize = 42
	resultLabel.Alignment = fyne.TextAlignCenter

	stackLabel := widget.NewLabel(ssJoined)

	input := widget.NewEntry()
	input.PlaceHolder = "Enter expression or command"
	enterfunc := func(s string) {
		log.Println(" func assigned closure ENTER was hit:", s)
		input.SetText("")
	}
	input.OnSubmitted = enterfunc

	regStr := OutputRegToString()
	regLabel := widget.NewLabel(regStr)

	leftColumn := container.NewVBox(input, resultLabel, stackLabel, regLabel)

	displayString := strings.Join(DisplayTape, "\n")
	displayLabel := widget.NewLabel(displayString)
	paddingLabel := widget.NewLabel("\n \n \n \n")

	go showHelp()

	rightColumn := container.NewVBox(paddingLabel, displayLabel)

	combinedColumns := container.NewHBox(leftColumn, rightColumn)

	globalW.SetContent(combinedColumns)
	globalW.Resize(fyne.Size{Width: 950, Height: 950})

	globalW.CenterOnScreen()
	globalW.ShowAndRun()
}

func keyTyped(e *fyne.KeyEvent) { // index is a global var
	switch e.Name {
	case fyne.KeyUp:
	case fyne.KeyDown:
	case fyne.KeyLeft:
	case fyne.KeyRight:
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//		(*globalA).Quit()
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyPageUp:
	case fyne.KeyPageDown:
	case fyne.KeyPlus:
	case fyne.KeyMinus:
	case fyne.KeyEqual:
	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		globalA.Quit()
	case fyne.KeyBackspace:
	}
}

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

/* ------------------------------------------------------------ GetRegIdx --------- */

func GetRegIdx(chr byte) int {
	// Return 0..35 with A = 10 and Z = 35
	ch := tknptr.CAP(chr)
	if (ch >= '0') && (ch <= '9') {
		ch = ch - '0'
	} else if (ch >= 'A') && (ch <= 'Z') {
		ch = ch - 'A' + 10
	} else { // added 12/11/2016 to fix bug
		ch = 0
	}
	return int(ch)
} // end GetRegIdx

/*-------------------------------------------------------------- GetRegChar ------  */

func GetRegChar(idx int) string {
	/* Return '0'..'Z' with A = 10 and Z = 35 */
	const RegNames = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if (idx < 0) || (idx > 35) {
		return "0"
	}
	ch := RegNames[idx]
	return string(ch)
} // GetRegChar


// ------------------------------------------------------------ OutputRegToString --------------

func OutputRegToString() string {
	FirstNonZeroStorageFlag := true
	ss := make([]string, 0, 40)

	for i, r := range Storage {
		if r != 0.0 {
			if FirstNonZeroStorageFlag {
				s := fmt.Sprintf("                The following storage registers are not zero")
				ss = append(ss, s)
				FirstNonZeroStorageFlag = false
			}
			ch := GetRegChar(i)
			sigfig := hpcalc2.SigFig()
			s := strconv.FormatFloat(r, 'g', sigfig, 64)
			s = hpcalc2.CropNStr(s)
			if r >= 10000 {
				s = hpcalc2.AddCommas(s)
			}
			str := fmt.Sprintf("Reg [%s] = %s", ch, s)
			ss = append(ss, str)
		} // if storage value is not zero
	} // for range over Storage
	if FirstNonZeroStorageFlag {
		s := fmt.Sprintf("All storage registers are zero.")
		ss = append(ss, s)
	}
	jointedStr := strings.Join(ss, "\n")
	return jointedStr
} // end OutputRegToString

// --------------------------------------------------------- getHelpStr ------------------------------------------

func showHelp() {

	time.Sleep(5 * time.Second)
	_, ss := hpcalc2.GetResult("help")
	helpStr := strings.Join(ss, "\n")
	helpLabel := widget.NewLabel(helpStr)
	dialog.ShowCustom("Help","OK", helpLabel, globalW)

	return
}