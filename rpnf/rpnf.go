/* From Fyne GUI book by Andrew Williams, Chapter 6, widget.go
5 Sep 21 -- Started playing w/ the UI for rpn calculator.  I already have the code that works, so I just need the UI and some support code.
8 Sep 21 -- Working as expected.  By george I think I've done it!


*/
package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"os"
	"runtime"
	"src/hpcalc2"
	"src/makesubst"
	"src/tknptr"
	"strconv"
	"strings"
	"time"
	//"fyne.io/fyne/v2/data/binding"
	//ct "github.com/daviddengcn/go-colortext"
	//ctfmt "github.com/daviddengcn/go-colortext/fmt"
)

const lastModified = "Sep 8, 2021"

var globalA fyne.App
var globalW fyne.Window

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var cyan = color.NRGBA{R: 0, G: 255, B: 255, A: 255}
var lightcyan = color.NRGBA{R: 224, G: 255, B: 255, A: 255}

//var homeDir, INBUF, resultToOutput string
var homeDir, resultToOutput string
var windowsFlag bool
var Storage [36]float64 // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var inbufChan chan string

const Storage1FileName = "RPNStorage.gob" // Allows for a rotation of Storage files, in case of a mistake.
const Storage2FileName = "RPNStorage2.gob"
const Storage3FileName = "RPNStorage3.gob"
const DisplayTapeFilename = "displaytape.txt"

func main() {
	fmt.Printf(" rpnf.go, using fyne.io v2.  Last modified %s, compiled using %s.\n", lastModified, runtime.Version())

	var nofileflag = flag.Bool("n", false, "no files read or written.") // pointer
	flag.Parse()

	var Stk hpcalc2.StackType // used when time to write out the stack upon exit.
	var err error
	DisplayTape = make([]string, 0, 100)
	DisplayTape = append(DisplayTape, "History of entered commands")
	theFileExists := true
	inbufChan = make(chan string)

	homeDir, err = os.UserHomeDir() // this function became available as of Go 1.12
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error from os.UserHomeDir() is", err)
		os.Exit(1)
	}
	windowsFlag = runtime.GOOS == "windows"

	StorageFullFilename := homeDir + string(os.PathSeparator) + Storage1FileName
	Storage2FullFilename := homeDir + string(os.PathSeparator) + Storage2FileName
	Storage3FullFilename := homeDir + string(os.PathSeparator) + Storage3FileName

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
	globalW = globalA.NewWindow("rpnf -- fyne")
	globalW.Canvas().SetOnTypedKey(keyTyped)

	populateUI()
	go Doit()
	if flag.NArg() > 0 { // there will always be at least 1 here, as os.Args[0] is the program name itself.
		inbufChan <- strings.Join(flag.Args(), " ")
	}

	globalW.CenterOnScreen()
	globalW.ShowAndRun()

	// Time to write files before exiting, if the flag says so.

	if !*nofileflag {
		// Rotate StorageFileNames and write
		err = os.Rename(Storage2FullFilename, Storage3FullFilename)
		if err != nil && !*nofileflag {
			fmt.Fprintf(os.Stderr, " Rename of storage 2 to storage 3 failed with error %v \n", err)
		}

		err = os.Rename(StorageFullFilename, Storage2FullFilename)
		if err != nil && !*nofileflag {
			fmt.Fprintf(os.Stderr, " Rename of storage 1 to storage 2 failed with error %v \n", err)
		}

		thefile, err = os.Create(StorageFullFilename) // for writing
		check(err)                                    // This should not fail, so panic if it does fail.
		defer thefile.Close()

		Stk = hpcalc2.GETSTACK()
		encoder := gob.NewEncoder(thefile) // encoder writes the file
		err = encoder.Encode(Stk)          // encoder writes the file
		check(err)                         // Panic if there is an error
		err = encoder.Encode(Storage)      // encoder writes the file
		check(err)
	}

	// Will open this file in the current working directory instead of the HomeDir.
	DisplayTapeFile, err := os.OpenFile(DisplayTapeFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error while opening DisplayTapeFilename", err)
		os.Exit(1)
	}
	defer DisplayTapeFile.Close()
	DisplayTapeWriter := bufio.NewWriter(DisplayTapeFile)
	defer DisplayTapeWriter.Flush()
	today := time.Now()
	datestring := today.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
	_, err = DisplayTapeWriter.WriteString("------------------------------------------------------\n")
	_, err = DisplayTapeWriter.WriteString(datestring)
	_, err = DisplayTapeWriter.WriteRune('\n')
	for _, s := range DisplayTape {
		_, err = DisplayTapeWriter.WriteString(s)
		_, err = DisplayTapeWriter.WriteRune('\n')
		check(err)
	}
	_, err = DisplayTapeWriter.WriteString("\n\n")
	check(err)

} // end main

// ---------------------------------------------------------- Doit ------------------------------------------------

func Doit() {
	INBUF := ""
	for { // main processing loop
		select {
		case INBUF = <-inbufChan: // this is blocking
		}
		if len(INBUF) > 0 {
			INBUF = makesubst.MakeSubst(INBUF)
			INBUF = strings.ToUpper(INBUF)
			DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
			// These commands are not run thru hpcalc as they are processed before calling it.
			realtknslice := tknptr.RealTokenSlice(INBUF)
			INBUF = "" // do this to stop endless processing of INBUF in a concurrent model.
			stringslice = nil
			//fmt.Println(" after setting stringslice to nil.  len =", len(stringslice), " and cap =", cap(stringslice))  output is 0 and 0

			for _, rtkn := range realtknslice {
				if rtkn.Str == "HELP" || rtkn.Str == "?" || rtkn.Str == "H" { // have more help lines to print
					str := fmt.Sprintf("%s last modifed on %s, and compiled w/ %s", os.Args[0], lastModified, runtime.Version())
					showHelp(str)
				} else if rtkn.Str == "ZEROREG" {
					for c := range Storage {
						Storage[c] = 0
					}
				} else if strings.HasPrefix(rtkn.Str, "STO") {
					i := 0
					if len(rtkn.Str) > 3 {
						ch := rtkn.Str[3] // The 4th position.
						i = GetRegIdx(ch)
					}
					Storage[i] = hpcalc2.READX()
				} else if strings.HasPrefix(rtkn.Str, "RCL") {
					i := 0
					if len(rtkn.Str) > 3 {
						ch := rtkn.Str[3] // the 4th position.
						i = GetRegIdx(ch)
					}
					hpcalc2.PUSHX(Storage[i])
				} else {
					// -------------------------------------------------------------------------------------
					_, stringslice = hpcalc2.Result(rtkn) //   Here is where GetResult is called -> Result
					// -------------------------------------------------------------------------------------
				}
				// -------------------------------------------------------------------------------------

				//  These commands are processed thru GetResult() first, then these are processed here.
				if strings.ToLower(rtkn.Str) == "about" { // I'm using ToLower here just to experiment a little.
					str := fmt.Sprintf("Last altered the source of rpnf.go %s, compiled w/ %s", lastModified, runtime.Version())
					stringslice = append(stringslice, str)
				}
				if len(stringslice) > 0 {
					resultToOutput = strings.Join(stringslice, "\n")
				}

			}
		}
		populateUI()
		globalW.Show()
	}

} // end Doit

// ---------------------------------------------------------- keyTyped --------------------------------------------
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
	//case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
	//globalA.Quit()
	case fyne.KeyBackspace:
	}
} // end keyTyped

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
} // end check

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

// --------------------------------------------------------- showHelp ------------------------------------------

func showHelp(extra string) {
	//time.Sleep(5 * time.Second)
	_, ss := hpcalc2.GetResult("help")
	ss = append(ss, extra)
	helpStr := strings.Join(ss, "\n")
	helpLabel := widget.NewLabel(helpStr)
	dialog.ShowCustom("Help text", "OK", helpLabel, globalW)

	return
} // end showHelp

// -------------------------------------------------------- PopulateUI -------------------------------------------

func populateUI() {
	R := hpcalc2.READX()
	sigfig := hpcalc2.SigFig()

	resultStr := strconv.FormatFloat(R, 'g', sigfig, 64)
	resultStr = hpcalc2.CropNStr(resultStr)
	if R > 1_000_000 {
		resultStr = hpcalc2.AddCommas(resultStr)
	}

	resultLabel := canvas.NewText("X = "+resultStr, green)
	resultLabel.TextSize = 42
	resultLabel.Alignment = fyne.TextAlignCenter

	_, ss := hpcalc2.GetResult("dump")
	ssJoined := strings.Join(ss, "\n")
	stackLabel := widget.NewLabel(ssJoined)

	input := widget.NewEntry()
	input.PlaceHolder = "Enter expression or command"
	enterfunc := func(s string) {
		inbufChan <- s // send this string down the channel
		input.SetText("")
	}
	input.OnSubmitted = enterfunc

	regStr := OutputRegToString()
	regLabel := widget.NewLabel(regStr)

	outputFromHPlabel := widget.NewLabel(resultToOutput)

	leftColumn := container.NewVBox(input, resultLabel, stackLabel, regLabel, outputFromHPlabel)

	displayString := strings.Join(DisplayTape, "\n")
	displayLabel := widget.NewLabel(displayString)
	paddingLabel := widget.NewLabel("\n \n \n \n")

	_, mapString := hpcalc2.GetResult("mapsho")
	mapJoined := strings.Join(mapString, "\n")
	maplabel := widget.NewLabel(mapJoined)
	rightColumn := container.NewVBox(paddingLabel, displayLabel, maplabel)

	combinedColumns := container.NewHBox(leftColumn, rightColumn)

	globalW.SetContent(combinedColumns)
	globalW.Resize(fyne.Size{Width: 950, Height: 950})

} // end populateUI
