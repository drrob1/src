package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"os"
	"runtime"
	"src/hpcalc2"
	"src/makesubst"
	"src/timlibg"
	"src/tknptr"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	//"fyne.io/fyne/v2/data/binding"
	//ct "github.com/daviddengcn/go-colortext"
	//ctfmt "github.com/daviddengcn/go-colortext/fmt"
)

/*
 From Fyne GUI book by Andrew Williams, (C) 2021 Packtpub.
 5 Sep 21 -- Started playing w/ the UI for rpn calculator.  I already have the code that works, so I just need the UI and some support code.
 8 Sep 21 -- Working as expected.  By george I think I've done it!
 9 Sep 21 -- Using the direct clipboard functions from fyne instead of the shelling out done in hpcal2.  Andy Williams had to help me for me to get this right.
10 Sep 21 -- Adding a way to have input box get input without having to click in it.  And it works!
13 Sep 21 -- After Andy wrote back as how to code what I want, I had already taken a stab at it.
               Turns out that keyTyped func is much more complex than it needs to be.  So I left in what I had already coded, and added what Andy suggested.
16 Sep 21 -- Made result output color yellow, defined yellow, and added output modes.
17 Sep 21 -- Fyne v 2.1.0 released today, and added a new widget.RichText that I'm going to use for the help output and see what happens.
19 Sep 21 -- Added light and dark commands to change the theme.  And found container.NewScroll from the fyne conference 2021 talk.
29 Sep 21 -- playing w/ an idea for backspace operation. Turns out that it works.
30 Sep 21 -- changing function of <space>
 1 Oct 21 -- changing left, right arrows to swap X and Y, '=' will always send '+' and ';' will always send '*'
11 Oct 21 -- Starting to add a pop-up modal form for register names.  This was finished the next evening.
14 Oct 21 -- Added trim to the popup text
21 Oct 21 -- Added processing of backspace and del to the popup text.  That was an oversight.
               And added commas to big X display when > 10K instead of 1M.
31 Oct 21 -- Will allow fixed, float and gen to switch output modes.  So fix will also change modes, but sigfig will not.
               And fix, lastx both use letter X which immediately exits.  Now fixed.
 8 Jan 22 -- Will detect File Not Found error and handle it differently than other errors.  I now know how based on "Powerful Command Line Applications in Go."
               And will have keyTyped go back into the Entry widget.  I think it looks nicer.
12 Feb 22 -- Going back to not have keyTyped to into the entry widget.  This allows <space> to be a delimiter.  I like that better.
16 Mar 22 -- Removing fmt.Print calls so a terminal window doesn't appear, unless I use the -v flag.
 5 May 22 -- HPCALC2 was changed to use OS specific code. No changes here, though.
16 May 22 -- Removed a superfluous select statement in Doit.  I understand concurrency better now.
11 Aug 22 -- About command will give more info about the exe binary
21 Oct 22 -- golangci-lint caught that I have an unneeded Sprintf call.  I removed both of them.  And added to show when the binary was last linked for the ABOUT cmd.
18 Feb 23 -- Changing from os.UserHomeDir to os.UserConfigDir. This is %appdata% or $HOME/.config
 3 Sep 23 -- When entering an arrow key for stack manipulation, the entry is lost.  Time to fix that by sending the input string, if it exists, down the chan before sending the arrow key.
20 Oct 23 -- Removed KeyQ -> quit from key processing.  I couldn't enter sqrt otherwise easily.  I had to type directly into the input box until I coded this fix.
 5 May 25 -- Fixed a bug regarding not clearing old message so they don't get displayed after they've become stale.
 7 May 25 -- Edited "about" text.
22 May 25 -- Copied new code from rpnf2 that handles the map command.  So now there's no difference btwn rpnf and rpnf2
 1 Nov 25 -- Updated comments.  The screen is built in populateUI using NewVBox for the left and right parts, and then NewHBox to combine them.  SetContent is used to display the output.
				The input is handled in keyTyped and its brothers.  The explicit go routine is in DoIt, which starts the concurrent code to handle the input.
*/

const lastModified = "May 22, 2025"

const ( // output modes
	outputfix = iota
	outputfloat
	outputgen
)

var outputMode int

var divider = "-------------------------------------------------------------------------------------------------------"

var globalA fyne.App
var globalW, helpWindow, popupName fyne.Window
var input, nameLabelInput *widget.Entry

var green = color.NRGBA{R: 0, G: 100, B: 0, A: 255}
var yellow = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var red = color.NRGBA{R: 100, G: 0, B: 0, A: 255}
var blue = color.NRGBA{R: 0, G: 0, B: 100, A: 255}
var gray = color.Gray{Y: 100}
var cyan = color.NRGBA{R: 0, G: 255, B: 255, A: 255}
var lightcyan = color.NRGBA{R: 224, G: 255, B: 255, A: 255}

// var homeDir, INBUF, resultToOutput string
var homeDir, resultToOutput string
var windowsFlag bool

type register struct {
	Value float64
	Name  string
}

var Storage [36]register // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var inbufChan chan string
var shiftState, lightTheme bool

const Storage1FileName = "RPNfyneStorage.gob" // Allows for a rotation of Storage files, in case of a mistake.
const Storage2FileName = "RPNfyneStorage2.gob"
const Storage3FileName = "RPNfyneStorage3.gob"
const DisplayTapeFilename = "displaytape.txt"

var nofileflag = flag.Bool("n", false, "no files read or written.") // pointer
var verboseFlag = flag.Bool("v", false, "Verbose mode enabled.")
var screenWidth = flag.Float64("sw", 950, "Screen Width for the resize method in Doit.") // needed by distortion on H97N
var screenHeight = flag.Float64("sh", 1000, "screen height for the resizze method in Doit.")

var execname, execTimeStamp string
var execFI os.FileInfo

func main() {

	flag.Parse()

	if *verboseFlag {
		fmt.Printf(" rpnf.go, using fyne.io v2.  Last modified %s, compiled using %s.\n", lastModified, runtime.Version())
	}

	var Stk hpcalc2.StackType // used when it's time to write out the stack upon exit.
	var err error
	DisplayTape = make([]string, 0, 100)
	DisplayTape = append(DisplayTape, "History of entered commands")
	theFileExists := true
	inbufChan = make(chan string, 10)

	//homeDir, err = os.UserHomeDir() // this function became available as of Go 1.12
	homeDir, err = os.UserConfigDir() // this function became available as of Go 1.12
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error from os.UserConfigDir() is", err)
		os.Exit(1)
	}
	windowsFlag = runtime.GOOS == "windows"
	if windowsFlag {
		lightTheme = true
	}

	StorageFullFilename := homeDir + string(os.PathSeparator) + Storage1FileName
	Storage2FullFilename := homeDir + string(os.PathSeparator) + Storage2FileName
	Storage3FullFilename := homeDir + string(os.PathSeparator) + Storage3FileName

	var thefile *os.File

	thefile, err = os.Open(StorageFullFilename) // open for reading
	if err != nil {
		theFileExists = false
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "%s not found\n", Storage1FileName)
		} else {
			fmt.Fprintf(os.Stderr, " Error from os.Open(%s) is: %v\n", Storage1FileName, err)
		}
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
	globalW = globalA.NewWindow("rpnf calculator using fyne")
	globalW.Canvas().SetOnTypedKey(keyTyped)

	execname, _ = os.Executable()
	execFI, _ = os.Stat(execname)
	execTimeStamp = execFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

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
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, " Rename of storage 2 to storage 3 failed with error %v \n", err)
			}
		}

		err = os.Rename(StorageFullFilename, Storage2FullFilename)
		if err != nil && !*nofileflag {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, " Rename of storage 1 to storage 2 failed with error %v \n", err)
			}
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
	DisplayTapeWriter.WriteString("------------------------------------------------------\n")
	DisplayTapeWriter.WriteString(datestring)
	DisplayTapeWriter.WriteRune('\n')
	for _, s := range DisplayTape {
		DisplayTapeWriter.WriteString(s)
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
		INBUF = <-inbufChan // this is blocking
		INBUF = strings.TrimSpace(INBUF)
		if len(INBUF) > 0 {
			INBUF = makesubst.MakeShorterReplaced(INBUF)                            // doesn't alter backtick, only will change '=' to '+' and ';' to '*'
			if strings.HasPrefix(INBUF, "map") || strings.HasPrefix(INBUF, "MAP") { // map operations are processed differently now.  This code copied from rpnf2.
				_, stringslice = hpcalc2.GetResult(INBUF)
				if len(stringslice) > 0 {
					resultToOutput = strings.Join(stringslice, "\n")
				}
				populateUI()
				globalW.Show()
				continue
			}

			INBUF = strings.ToUpper(INBUF)
			DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
			// These commands are not run thru hpcalc as they are processed before calling it.
			realtknslice := tknptr.RealTokenSlice(INBUF)
			INBUF = "" // do this to stop endless processing of the same value of INBUF in a concurrent model.
			stringslice = nil

			for _, rtkn := range realtknslice {
				if rtkn.Str == "HELP" || rtkn.Str == "?" || rtkn.Str == "H" { // append more help lines to output from HPCALC2
					extra := make([]string, 0, 10)
					extra = append(extra, "STOn,RCLn  -- store/recall the X register to/from the register indicated by n.")
					extra = append(extra, "OutputFixed (fix), OutputFloat (float, real), OutputGen (gen) -- set OutputMode to fixed, float or gen.")
					extra = append(extra, "SigN, FixN -- set significant figures for displayed numbers to N.  Default is -1.")
					extra = append(extra, "dark, light -- set Fyne theme to dark or light.")
					str := fmt.Sprintf("%s last modified on %s and compiled w/ %s \n", os.Args[0], lastModified, runtime.Version())
					extra = append(extra, str)
					showHelp(extra)
				} else if rtkn.Str == "ZEROREG" {
					for c := range Storage {
						Storage[c].Value = 0
					}
				} else if strings.HasPrefix(rtkn.Str, "STO") {
					i := 0
					if len(rtkn.Str) > 3 {
						ch := rtkn.Str[3] // The 4th position.
						i = GetRegIdx(ch)
					}
					Storage[i].Value = hpcalc2.READX()
					if i > 0 {
						getNameFromPopup()
						// I took out a select {} statement here because it was not needed.
						name := <-inbufChan
						if strings.ToLower(name) == "t" || strings.ToLower(name) == "today" {
							m, d, y := timlibg.TIME2MDY()
							name = timlibg.MDY2STR(m, d, y)
						}
						Storage[i].Name = name
					}
				} else if strings.HasPrefix(rtkn.Str, "RCL") {
					i := 0
					if len(rtkn.Str) > 3 {
						ch := rtkn.Str[3] // the 4th position.
						i = GetRegIdx(ch)
					}
					hpcalc2.PUSHX(Storage[i].Value)
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
				} else if strings.HasPrefix(rtkn.Str, "OUTPUTFL") || strings.HasPrefix(rtkn.Str, "OUTPUTR") || rtkn.Str == "REAL" || rtkn.Str == "FLOAT" {
					outputMode = outputfloat
				} else if strings.HasPrefix(rtkn.Str, "OUTPUTG") || strings.HasPrefix(rtkn.Str, "GEN") {
					outputMode = outputgen
				} else if rtkn.Str == "DARK" {
					fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme()) // Goland is saying that DarkTheme is depracated and will be removed in v3.
					lightTheme = false
				} else if rtkn.Str == "LIGHT" {
					fyne.CurrentApp().Settings().SetTheme(theme.LightTheme()) // Goland is saying that LightTheme is depracated and will be removed in v3.
					lightTheme = true
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

					str = fmt.Sprintf(" %s timestamp is %s.\n Full exec name is %s.", execFI.Name(), execTimeStamp, execname)
					stringslice = append(stringslice, str)

				} else if strings.HasPrefix(rtkn.Str, "FIX") { // so fix, fixed, etc sets output mode AND number of significant figures.
					outputMode = outputfix
				}
				resultToOutput = "" // clears any old messages.
				if len(stringslice) > 0 {
					resultToOutput = strings.Join(stringslice, "\n")
				}

			}
		}
		populateUI()
		globalW.Show()
	}
} // end Doit

/*
// ---------------------------------------------------------- keyTyped --------------------------------------------
func keyTyped(e *fyne.KeyEvent) { // Maybe better to first call input.TypedRune, and then change focus.  Else some keys were getting duplicated.
	switch e.Name {
	case fyne.KeyUp: // stack up
		inbufChan <- ","
	case fyne.KeyDown: // stack down
		//                                          _ = hpcalc2.PopX()  but this prevents UNDO from working here.
		inbufChan <- "POP" // I'm sending a command thru instead of calling PopX so that UNDO will work correctly.
	case fyne.KeyLeft: // swap X, Y
		//                                                                                globalW.Canvas().Focus(input)
		inbufChan <- "~"
	case fyne.KeyRight: // swap X, Y
		//                                                                                globalW.Canvas().Focus(input)
		inbufChan <- "~"
	case fyne.KeyEscape, fyne.KeyQ:
		//                                                                                               globalA.Quit()
		//	                                                                                          (*globalA).Quit()
		globalW.Close() // quit's the app if this is the last window.
	case fyne.KeyX:
		if len(input.Text) == 0 { // only eXit if X is first character typed.
			globalW.Close()
		} else {
			input.TypedRune('X')
		}
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyPageUp:
	case fyne.KeyPageDown:
	case fyne.KeySpace:
		//                                                                                globalW.Canvas().Focus(input)
		//                                                                                         input.TypedRune(' ')
		inbufChan <- input.Text
	case fyne.KeyBackspace, fyne.KeyDelete:
		//                                                                                globalW.Canvas().Focus(input)
		//                                                                                        input.TypedRune('\b')
		text := input.Text
		if len(text) > 0 {
			text = text[:len(text)-1]
		}
		input.SetText(text)
	case fyne.KeyPlus:
		input.TypedRune('+')
	case fyne.KeyAsterisk:
		input.TypedRune('*')
	case fyne.KeyEqual:
		input.TypedRune('+')
	case fyne.KeySemicolon:
		input.TypedRune('*')
	case fyne.KeyF1, fyne.KeyF2, fyne.KeyF12:
		//                                                                             input.TypedRune('H') // for help
		//                                                                                      inbufChan <- input.Text
		inbufChan <- "H"

	case fyne.KeyEnter, fyne.KeyReturn:
		inbufChan <- input.Text

	default:
		if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
			shiftState = true
			return
		}
		if shiftState {
			shiftState = false
			if e.Name == fyne.KeySlash {
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
			//globalW.Canvas().Focus(input)
		} else {
			input.TypedRune(rune(e.Name[0]))
			//globalW.Canvas().Focus(input)
		}

		//fmt.Printf(" in keyTyped, e.Name is: %q\n", e.Name) I saw LeftShift, RightShift, LeftControl, RightControl when I depressed the keys.
	}
} // end keyTyped
*/

// ---------------------------------------------------------- keyTyped --------------------------------------------
func keyTyped(e *fyne.KeyEvent) { // Now calls input.TypedRune, and then change focus.  inbufChan is the string chan that is used to return these strings to be processed.
	switch e.Name {
	case fyne.KeyUp: // stack up
		if len(input.Text) > 0 {
			inbufChan <- input.Text
		}
		inbufChan <- ","
		return
	case fyne.KeyDown: // stack down
		if len(input.Text) > 0 {
			inbufChan <- input.Text
		}
		inbufChan <- "POP" // I'm sending a command thru instead of calling PopX so that UNDO will work correctly.
		return
	case fyne.KeyLeft: // swap X, Y
		if len(input.Text) > 0 {
			inbufChan <- input.Text
		}
		inbufChan <- "~"
		return
	case fyne.KeyRight: // swap X, Y
		if len(input.Text) > 0 {
			inbufChan <- input.Text
		}
		inbufChan <- "~"
		return
	case fyne.KeyEscape: // fyne.KeyQ  removed "Q" from quitting, because then I couldn't enter sqrt, for example.
		globalW.Close() // quits the app if this is the last window.
		//                                                                                               globalA.Quit()
		//	                                                                                          (*globalA).Quit()
	case fyne.KeyX:
		if len(input.Text) == 0 { // only eXit if X is first character typed.
			globalW.Close()
		} else {
			input.TypedRune('X')
		}
	case fyne.KeyHome:
	case fyne.KeyEnd:
	case fyne.KeyPageUp:
	case fyne.KeyPageDown:
	case fyne.KeySpace:
		//     inbufChan <- input.Text // removed 5/5/25
		globalW.Canvas().Focus(input) // added back 5/5/25
		input.TypedRune(' ')          // added back 5/5/25
		return
	case fyne.KeyBackspace, fyne.KeyDelete:
		text := input.Text
		if len(text) > 0 {
			text = text[:len(text)-1]
		}
		input.SetText(text)
		//                                                                                globalW.Canvas().Focus(input)
		//                                                                                        input.TypedRune('\b')

	case fyne.KeyPlus:
		input.TypedRune('+')
	case fyne.KeyAsterisk:
		input.TypedRune('*')
	case fyne.KeyEqual:
		input.TypedRune('+')
	case fyne.KeySemicolon:
		input.TypedRune('*')
	case fyne.KeyF1, fyne.KeyF2, fyne.KeyF12:
		if len(input.Text) > 0 {
			inbufChan <- input.Text
		}
		inbufChan <- "H"
		//                                                                             input.TypedRune('H') // for help
		//                                                                                      inbufChan <- input.Text
		return
	case fyne.KeyEnter, fyne.KeyReturn:
		inbufChan <- input.Text

	default:
		if e.Name == "LeftShift" || e.Name == "RightShift" || e.Name == "LeftControl" || e.Name == "RightControl" {
			shiftState = true
			// globalW.Canvas().Focus(input)
			return
		}
		if shiftState {
			shiftState = false
			if e.Name == fyne.KeySlash {
				input.TypedRune('?')
			} else if e.Name == fyne.KeyPeriod {
				inbufChan <- "~"
				//input.TypedRune('>')
				return
			} else if e.Name == fyne.KeyComma {
				inbufChan <- "~"
				//input.TypedRune('<')
				return
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
				if len(input.Text) > 0 {
					inbufChan <- input.Text
				}
				inbufChan <- "~"
				//                                                                                  input.TypedRune('~')
				return
			}
			// globalW.Canvas().Focus(input)  Not changing focus into the entry widget.  This is the line that changes focus.  Added to KeySpace, above May 2025.
		} else {
			input.TypedRune(rune(e.Name[0]))
			//globalW.Canvas().Focus(input)
		}

	}
	// globalW.Canvas().Focus(input) // first key typed that's not a command changes the focus to the entry widget.  Undone.  Added to KeySpace, above May 2025.
} // end keyTyped

// ---------------------------------------------------------- keyTypedHelp --------------------------------------------
func keyTypedHelp(e *fyne.KeyEvent) { // Maybe better to first call input.TypedRune, and then change focus.  Else some keys were getting duplicated.
	switch e.Name {
	//case fyne.KeySpace:
	//	globalW.Canvas().Focus(input)
	//	input.TypedRune(' ')

	case fyne.KeyEnter, fyne.KeyReturn, fyne.KeySpace:
		helpWindow.Close()

	case fyne.KeyEscape, fyne.KeyQ, fyne.KeyX:
		globalA.Quit()

	default:
		//input.TypedRune(rune(e.Name[0]))  ignore the key but change focus to the main window
		globalW.Canvas().Focus(input)
	}
} // end keyTypedHelp

// ---------------------------------------------------------- keyTypedPopup --------------------------------------------
func keyTypedPopup(e *fyne.KeyEvent) { // Maybe better to first call input.TypedRune, and then change focus.  Else some keys were getting duplicated.
	switch e.Name {
	case fyne.KeySpace:
		nameLabelInput.TypedRune(' ')

	case fyne.KeyEnter, fyne.KeyReturn:
		popupName.Close()
		inbufChan <- strings.TrimSpace(nameLabelInput.Text)

	case fyne.KeyQ, fyne.KeyX:
		globalA.Quit()

	case fyne.KeyEscape:
		popupName.Close()
		inbufChan <- " "

	case fyne.KeyBackspace, fyne.KeyDelete:
		text := nameLabelInput.Text
		if len(text) > 0 {
			text = text[:len(text)-1]
		}
		nameLabelInput.SetText(text)

	default:
		nameLabelInput.TypedRune(rune(e.Name[0]))
	}
} // end keyTypedPopup

// --------------------------------------------------------- getNameFromPopup ------------------------------------------

func getNameFromPopup() {
	nameLabelInput = widget.NewEntry()
	nameLabelInput.PlaceHolder = "Enter name label for register"
	enterFunc := func(s string) {
		inbufChan <- s
		popupName.Close()
	}
	nameLabelInput.OnSubmitted = enterFunc

	popupName = globalA.NewWindow("Get Name Label for register")

	popupName.SetContent(nameLabelInput)
	popupName.Canvas().SetOnTypedKey(keyTypedPopup)
	popupName.Resize(fyne.NewSize(500, 200))
	popupName.Show()
	//return  redundant
} // end getNameFromPopup

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

	for i, r := range Storage { // r for register
		if r.Value != 0.0 {
			if FirstNonZeroStorageFlag {
				//s := fmt.Sprintf("                The following storage registers are not zero")
				s := "                The following storage registers are not zero"
				ss = append(ss, s)
				FirstNonZeroStorageFlag = false
			}
			ch := GetRegChar(i)
			sigfig := hpcalc2.SigFig()
			s := strconv.FormatFloat(r.Value, 'g', sigfig, 64)
			s = hpcalc2.CropNStr(s)
			if r.Value >= 10000 {
				s = hpcalc2.AddCommas(s)
			}
			str := fmt.Sprintf("Reg [%s], %s = %s", ch, r.Name, s)
			ss = append(ss, str)
		} // if storage value is not zero
	} // for range over Storage
	if FirstNonZeroStorageFlag {
		//s := fmt.Sprintf("All storage registers are zero.")
		s := "All storage registers are zero."
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
	helpLabel := widget.NewLabel(helpStr)

	//dialog.ShowCustom("", "OK", helpLabel, globalW) // empty title
	//dialog.ShowInformation("", helpStr, globalW)  // empty title  I don't like its look, as each line is centered and not left aligned.
	//helpScroll := container.NewVScroll(helpLabel)
	helpScroll := container.NewScroll(helpLabel)
	helpWindow = globalA.NewWindow("Help")
	helpWindow.SetContent(helpScroll)
	helpWindow.Canvas().SetOnTypedKey(keyTypedHelp)
	helpWindow.Resize(fyne.NewSize(1000, 900))
	helpWindow.Show()
	//dialog.ShowCustom("", "OK", helpScroll, globalW)
	//dialog.ShowCustom("Help text", "OK", helpLabel, globalW)
	//helpRichText := widget.NewRichTextWithText(helpStr)
	//dialog.ShowCustom("Help", "OK", helpRichText, globalW)

	//return  redundant
} // end showHelp

// -------------------------------------------------------- PopulateUI -------------------------------------------

func populateUI() {
	R := hpcalc2.READX()
	sigfig := hpcalc2.SigFig()

	resultStr := strconv.FormatFloat(R, 'g', sigfig, 64)
	resultStr = hpcalc2.CropNStr(resultStr)
	if R > 10_000 {
		resultStr = hpcalc2.AddCommas(resultStr)
	}

	resultLabel := canvas.NewText("X = "+resultStr, yellow)
	if lightTheme {
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

	leftColumn := container.NewVBox(input, resultLabel, stackLabel, regLabel, spacerLabel, outputFromHPlabel, dividerLabel)

	displayString := strings.Join(DisplayTape, "\n")
	displayLabel := widget.NewLabel(displayString)
	paddingLabel := widget.NewLabel("\n \n \n \n")

	_, mapString := hpcalc2.GetResult("mapsho")
	mapJoined := strings.Join(mapString, "\n")
	maplabel := widget.NewLabel(mapJoined)
	rightColumn := container.NewVBox(paddingLabel, displayLabel, maplabel)

	combinedColumns := container.NewHBox(leftColumn, rightColumn)

	globalW.SetContent(combinedColumns)
	globalW.Resize(fyne.Size{
		Width:  float32(*screenWidth),
		Height: float32(*screenHeight),
	})

} // end populateUI
