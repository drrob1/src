/* From Fyne GUI book by Andrew Williams, (C) 2021 Packtpub.
 5 Sep 21 -- Started playing w/ the UI for rpn calculator.  I already have the code that works, so I just need the UI and some support code.
 8 Sep 21 -- Working as expected.  By george I think I've done it!
 9 Sep 21 -- Using the direct clipboard functions from fyne instead of the shelling out done in hpcal2.  Andy Williams had to help me for me to get this right.
10 Sep 21 -- Adding a way to have input box get input without having to click in it.  And it works!
13 Sep 21 -- After Andy wrote back as how to code what I want, I had already taken a stab at it.
               Turns out that keyTyped func is much more complex than it needs to be.  So I left in what I had already coded, and added what Andy suggested.
16 Sep 21 -- Made result output color yellow, defined yellow, and added output modes.
17 Sep 21 -- Fyne v 2.1.0 released today, and added a new widget.RichText that I'm going to use for the help output and see what happens.
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
	"fyne.io/fyne/v2/theme"
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

const lastModified = "Sep 17, 2021"

const ( // output modes
	outputfix = iota
	outputfloat
	outputgen
)
var outputMode int

var divider = "-------------------------------------------------------------------------------------------------------"

var globalA fyne.App
var globalW fyne.Window
var input *widget.Entry

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var yellow = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
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
var shiftState bool

const Storage1FileName = "RPNfyneStorage.gob" // Allows for a rotation of Storage files, in case of a mistake.
const Storage2FileName = "RPNfyneStorage2.gob"
const Storage3FileName = "RPNfyneStorage3.gob"
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
	fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme()) // Goland is saying that DarkTheme is depracated and will be removed in v3.
	//fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())  // Goland is saying that LightTheme is depracated and will be removed in v3.
	globalW = globalA.NewWindow("rpnf calculator using fyne")
	globalW.Canvas().SetOnTypedKey(keyTyped)


	populateUI()
	go Doit()
	if flag.NArg() > 0 {
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
					extra := make([]string, 0, 10)
					extra = append(extra, "STOn,RCLn  -- store/recall the X register to/from the register indicated by n.")
					extra = append(extra, "OutputFixed, OutputFloat, OutputGen -- set OutputMode to fixed, float or gen.")
					extra = append(extra, "SigN, FixN -- set significant figures for displayed numbers to N.  Default is -1.")
					str := fmt.Sprintf("%s last modifed on %s, \n and compiled w/ %s \n", os.Args[0], lastModified, runtime.Version())
					extra = append(extra, str)
					showHelp(extra)
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
				} else if rtkn.Str == "FROMCLIP" {
					contents := globalW.Clipboard().Content()
					contents = strings.TrimSpace(contents)
					f, err := strconv.ParseFloat(contents, 64)
					if err == nil {
						hpcalc2.PUSHX(f)
					} else {
						msg := fmt.Sprintf("Error from conversion of clipboard.  Clipboard=%s, err=%v.", contents, err)
						stringslice = append(stringslice, msg)
					}
				} else if rtkn.Str == "TOCLIP" {
					rStr := strconv.FormatFloat(hpcalc2.READX(), 'g', hpcalc2.SigFig(), 64)
					globalW.Clipboard().SetContent(rStr)
				} else if strings.HasPrefix(rtkn.Str, "OUTPUTFI") {
					outputMode = outputfix
				} else if strings.HasPrefix(rtkn.Str, "OUTPUTFL") || strings.HasPrefix(rtkn.Str, "OUTPUTR"){
					outputMode = outputfloat
				} else if strings.HasPrefix(rtkn.Str, "OUTPUTG") {
					outputMode = outputgen
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
func keyTyped(e *fyne.KeyEvent) { // Maybe better to first call input.TypedRune, and then change focus.  Else some keys were getting duplicated.
	switch e.Name {
	case fyne.KeyUp: // stack up
		inbufChan <- ","
	case fyne.KeyDown: // stack down
		_ = hpcalc2.PopX()
		inbufChan <- ""
	case fyne.KeyLeft:
		globalW.Canvas().Focus(input)
	case fyne.KeyRight:
		globalW.Canvas().Focus(input)
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//		(*globalA).Quit()
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyPageUp:
	case fyne.KeyPageDown:
	case fyne.KeySpace:
		globalW.Canvas().Focus(input)
		input.TypedRune(' ')
	case fyne.KeyBackspace, fyne.KeyDelete:
		globalW.Canvas().Focus(input)
		input.TypedRune('\b')
	case fyne.KeyPlus:
		input.TypedRune('+')
	case fyne.KeyAsterisk:
		input.TypedRune('*')
	case fyne.KeyF1, fyne.KeyF2, fyne.KeyF12:
		input.TypedRune('H') // for help
		inbufChan <- input.Text

	case fyne.KeyEnter, fyne.KeyReturn:
		inbufChan <- input.Text

	default:
		if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
			shiftState = true
			return
		}
		if shiftState {
			shiftState = false
			if e.Name == fyne.KeyEqual {
				input.TypedRune('+')
			} else if e.Name == fyne.KeySlash {
				input.TypedRune('?')
			} else if e.Name == fyne.KeyPeriod {
				input.TypedRune('>')
			} else if e.Name == fyne.KeyComma {
				input.TypedRune('<')
			} else if e.Name == fyne.KeyMinus {
				input.TypedRune('_')
			} else if e.Name == fyne.Key8 {
				input.TypedRune('*')
			} else if e.Name == fyne.Key6 {
				input.TypedRune('^')
			} else if e.Name == fyne.Key5 {
				input.TypedRune('%')
			} else if e.Name == fyne.Key2 {
				input.TypedRune('@')
			} else if e.Name == fyne.Key1 {
				input.TypedRune('!')
			} else if e.Name == fyne.KeyBackTick {
				input.TypedRune('~')
			}
			globalW.Canvas().Focus(input)
		} else {
			input.TypedRune(rune(e.Name[0]))
			globalW.Canvas().Focus(input)
		}

		//fmt.Printf(" in keyTyped, e.Name is: %q\n", e.Name) I saw LeftShift, RightShift, LeftControl, RightControl when I depressed the keys.
	}
} // end keyTyped

/*
// ---------------------------------------------------------- keyTyped --------------------------------------------
func keyTyped(e *fyne.KeyEvent) { // index is a global var
	switch e.Name {
	case fyne.KeyUp:
		globalW.Canvas().Focus(input)
	case fyne.KeyDown:
		globalW.Canvas().Focus(input)
	case fyne.KeyLeft:
		globalW.Canvas().Focus(input)
	case fyne.KeyRight:
		globalW.Canvas().Focus(input)
	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalW.Close() // quit's the app if this is the last window, which it is.
		//		(*globalA).Quit()
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyPageUp:
	case fyne.KeyPageDown:
	case fyne.KeySpace:
		input.TypedRune(' ')
		globalW.Canvas().Focus(input)
	case fyne.KeyBackspace, fyne.KeyDelete:
		globalW.Canvas().Focus(input)
		input.TypedRune('\b')
	case fyne.KeyPlus:
		input.TypedRune('+')
	case fyne.KeyAsterisk:
		input.TypedRune('*')
	case fyne.KeyF1, fyne.KeyF2, fyne.KeyF12:
		input.TypedRune('H') // for help
	case fyne.KeyEnter, fyne.KeyReturn:
		inbufChan <- input.Text
	default:
		if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
			shiftState = true
			return
		}
		if shiftState {
			shiftState = false
			if e.Name == fyne.KeyEqual {
				input.TypedRune('+')
			} else if e.Name == fyne.KeySlash {
				input.TypedRune('?')
			} else if e.Name == fyne.KeyPeriod {
				input.TypedRune('>')
			} else if e.Name == fyne.KeyComma {
				input.TypedRune('<')
			} else if e.Name == fyne.KeyMinus {
				input.TypedRune('_')
			} else if e.Name == fyne.Key8 {
				input.TypedRune('*')
			} else if e.Name == fyne.Key6 {
				input.TypedRune('^')
			} else if e.Name == fyne.Key5 {
				input.TypedRune('%')
			} else if e.Name == fyne.Key2 {
				input.TypedRune('@')
			} else if e.Name == fyne.Key1 {
				input.TypedRune('!')
			} else if e.Name == fyne.KeyBackTick {
				input.TypedRune('~')
			}
		} else {
			globalW.Canvas().Focus(input)
			input.TypedRune(rune(e.Name[0]))
		}

		//fmt.Printf(" in keyTyped.  e.Name is: %q\n", e.Name) I saw LeftShift, RightShift, LeftControl, RightControl when I depressed the keys.
	}
} // end keyTyped

 */

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

func showHelp(extra []string) {
	//time.Sleep(5 * time.Second)
	_, ss := hpcalc2.GetResult("help")
	ss = append(ss, extra...)
	helpStr := strings.Join(ss, "\n")
	//helpLabel := widget.NewLabel(helpStr)
	//dialog.ShowCustom("Help text", "OK", helpLabel, globalW)
	helpRichText := widget.NewRichTextWithText(helpStr)
	dialog.ShowCustom("Help", "OK", helpRichText, globalW)

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

	resultLabel := canvas.NewText("X = "+resultStr, yellow)
	if runtime.GOOS == "windows" {
		resultLabel = canvas.NewText("X = "+resultStr, green)
	}
	resultLabel.TextSize = 42
	resultLabel.Alignment = fyne.TextAlignCenter

	ss := make([]string, 0, 10)
	if outputMode == outputfix {
		_, ss = hpcalc2.GetResult("DUMPFIXED")
	} else if outputMode == outputfloat {
		_, ss = hpcalc2.GetResult("DUMPFLOAT")
	} else if outputMode == outputgen {
		_, ss = hpcalc2.GetResult("DUMP")
	}

	ssJoined := strings.Join(ss, "\n")
	stackLabel := widget.NewLabel(ssJoined)

	input = widget.NewEntry()
	input.PlaceHolder = "Enter expression or command"
	enterfunc := func(s string) {
		inbufChan <- s // send this string down the channel
		input.SetText("")
	}
	input.OnSubmitted = enterfunc

	regStr := OutputRegToString()
	regLabel := widget.NewLabel(regStr)

	outputFromHPlabel := widget.NewLabel(resultToOutput)

	spacerLabel := widget.NewLabel(divider)
	dividerLabel := widget.NewLabel(divider)

	leftColumn := container.NewVBox(input, resultLabel, stackLabel, regLabel, spacerLabel,outputFromHPlabel, dividerLabel)

	displayString := strings.Join(DisplayTape, "\n")
	displayLabel := widget.NewLabel(displayString)
	paddingLabel := widget.NewLabel("\n \n \n \n")

	_, mapString := hpcalc2.GetResult("mapsho")
	mapJoined := strings.Join(mapString, "\n")
	maplabel := widget.NewLabel(mapJoined)
	rightColumn := container.NewVBox(paddingLabel, displayLabel, maplabel)

	combinedColumns := container.NewHBox(leftColumn, rightColumn)

	globalW.SetContent(combinedColumns)
	globalW.Resize(fyne.Size{Width: 950, Height: 1000})

} // end populateUI
