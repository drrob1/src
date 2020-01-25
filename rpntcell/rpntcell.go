// (C) 1990-2016.  Robert W.  Solomon.  All rights reserved.
// rpnterm.go
package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/gdamore/tcell"
	"log"
	"os"

	"runtime"
	"strconv"
	"strings"
	"time"
	"timlibg"

	"getcommandline"
	"hpcalc"
	"tokenize"
	//  "os/exec"      // for the clear screen functions.
	//	"github.com/gdamore/tcell"
	//	"github.com/gdamore/tcell/encoding"
	//	runewidth "github.com/mattn/go-runewidth"  Not needed after I simplified puts()
)

const LastAltered = "Jan 25, 2020"

// runtime.GOOS returns either linux or windows.  I have not tested mac.  I want either $HOME or %userprofile to set the write dir.

/*
This module uses the HPCALC module to simulate an RPN type calculator.
REVISION HISTORY
----------------
 1 Dec 89 -- Changed prompt.
24 Dec 91 -- Converted to M-2 V 4.00.  Changed params to GETRESULT.
25 Jul 93 -- Output result without trailing insignificant zeros, imported UL2, and changed prompt again.
 3 Mar 96 -- Fixed bug in string display if real2str fails because number is too large (ie, Avogadro's Number).
18 May 03 -- First Win32 version.  And changed name.
 1 Apr 13 -- Back to console mode pgm that will read from the cmdline.  Intended to be a quick and useful little utility.
		                   And will save/restore the stack to/from a file.
 2 May 13 -- Will use console mode flag for HPCALC, so it will write to console instead of the terminal module routines.
		                   And I now have the skipline included in MiscStdInOut so it is removed from here.
15 Oct 13 -- Now writing for gm2 under linux.
22 Jul 14 -- Converting to Ada.
 6 Dec 14 -- Converting to cpp.
20 Dec 14 -- Added macros for date and time last compiled.
31 Dec 14 -- Started coding HOL command.
 1 Jan 15 -- After getting HOL command to work, I did more fiddling to further understand c-strings and c++ string class.
10 Jan 15 -- Playing with string conversions and number formatting.
 5 Nov 15 -- Added the RECIP, CURT, VOL commands to hpcalc.cpp
22 Nov 15 -- Noticed that T1 and T2 stack operations are not correct.  This effects HP2cursed and rpnc.
13 Apr 16 -- Adding undo and redo commands, which operate on the entire stack not just X register.
 2 Jul 16 -- Fixed help to include PI command, and changed pivot for JUL command.  See hpcalcc.cpp
 7 Jul 16 -- Added UP command to hpcalcc.cpp
 8 Jul 16 -- Added display of stack dump to always happen, and a start up message.
22 Aug 16 -- Started conversion to Go, as rpn.go.
 8 Sep 16 -- Finished coding rpn.go, started 26 Aug 16 as rpng.go, adding functionality from hppanel, ie, persistant storage, a display tape and operator substitutions = or + and ; for *.
11 Sep 16 -- Changed stack and storage files to use gob package and only use one Storage file.
 1 Oct 16 -- Made the stack display when the program starts.
 4 Oct 16 -- Made the storage registers display when the program starts.
 7 Oct 16 -- Changed default dump to DUMPFIXED.   Input is now using splitfunc by space delimited words instead of whole lines.
		                    Conversion to ScanWords means I cannot get an empty string back unless ^C or ^D.  So needed (Q)uit, EXIT, STOP.
 8 Oct 16 -- Updated the prompt to say (Q)uit to exit instead of Enter to exit, and sto command calls WriteRegToScreen()
 9 Oct 16 -- Added ability to enter hex numbers like in C, or GoLang, etc, using the "0x" prefix.
21 Oct 16 -- Decided that sto command should not also dump stack, as the stack does not change.
28 Oct 16 -- Will clear screen in between command calls.  I took this out when I did next revision.
30 Oct 16 -- Decided to rename this to rpnterm, and will now start to use termbox-go.
31 Oct 16 -- Debugged GetInputString(x,y int) string
 4 Nov 16 -- Changed hpcalc.go to return a string slice for everything.  Added outputmodes.
 6 Nov 16 -- Windows returns <backspace> code of 8, which is std ASCII.  Seems linux does not do this.
23 Nov 16 -- Will clear screen before calling init termbox-go, to see if that helps some of the irregularities
		                    I've found with termbox-go.  It doesn't help.
11 Dec 16 -- Fixed bug in GetRegIdx when out of range char is passed in.
13 Jan 17 -- Removed stop as an exit command, as it meant that reg p could not be stored into.
23 Feb 17 -- Added "?" as equivalent to "help"
 4 Apr 17 -- Noticed that HardClearScreen is called before MaxRow is set.  I fixed this by moving the Size call.
		                  And I commented out a fmt.Println() call that didn't do anything anyway, and is ignored on Windows.
16 Aug 17 -- Added code from when.go to use timestamp on executable as link time.
25 Feb 18 -- PrimeFactorMemoized added.
 6 Apr 18 -- Wrote code to save DisplayTape as a text file.
22 Aug 18 -- learning about code folding
 2 Dec 18 -- Trying GoLand for editing this code now.  And exploring adding string names for the registers.
               Need to code name command, display these names, deal w/ the reg file.  Code a converter.
 4 Dec 18 -- Made STO also ask for NAME.  And used ClearLine when the OutputLine is increased.
 5 Dec 18 -- Help command will print from here those commands that are processed here, and from hpcalc those that are processed there.
 6 Dec 18 -- Added "today" for reg name string, and it will plug in today's date as a string.
 8 Dec 18 -- Added StrSubst for register name operation, so that = or - becomes a space.  Note that = becomes + in GetInputString.
10 Dec 18 -- Register 0 will not ask for name, to match my workflow using these registers.
17 Dec 18 -- Starting to code :w and :r to/from text files, intended for clipboard access via vim or another text editor.
18 Dec 18 -- Fixed help to show :r, rd, read commands.
31 Jan 19 -- Added prefix of  :w to write a text file, and prefix of :R to read a text file.
13 Apr 19 -- If on a small screen, like the System76 laptop, there are too many help lines, so it panics.  Started to fix that.
14 Apr 19 -- And took out a debug line at top that I should have done shortly after debugging this routine.
15 Apr 19 -- Changing the look somewhat.  I want the input to be on the top, like I did in C++.
 3 Jun 19 -- Added T as abbreviation for today in GetNameStr rtn, and in hpcalc since I liked the idea so much.
14 Dec 19 -- Moved prompt for register name string to top, from middle of screen where it's easy to miss.
29 Dec 19 -- Defer statement executes in a LIFO stack.  I have to reverse the order of the defer closing statements, and remove from explicit call at end.
               And checkmsg now uses fmt.Errorf so that I will see a message even if termbox is still active.  And need to respect output mode for registers.
               And fixed the condition that used to be INBUF != "CLEAR" || INBUF != "CLS", as that needed to be && there.  Picked up by go vet.
30 Dec 19 -- Generalizing outputfix, outputfloat, and outputgen.
17 Jan 20 -- Started converting from termbox to tcell.
19 Jan 20 -- Fixed bug in deleol.
20 Jan 10 -- Removed empiric fix in puts that was replaced by fixing deleol.  And decided that regular yellow is easier to see than boldyellow.
25 Jan 20 -- Substituted '=' to '+' and ';' to '*'.  Forgot about that earlier.
*/

const InputPrompt = " Enter calculation, HELP or <return> to exit: "

type Register struct {
	Value float64
	Name  string
}

var Storage [36]Register // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var Divider string
var clear map[string]func()

var StartCol, StartRow, sigfig, MaxRow, MaxCol, TitleRow, StackRow, RegRow, OutputRow, DisplayCol, PromptRow, outputmode, n int

const SpaceFiller = "     |     "

const ( // output modes
	outputfix = iota
	outputfloat
	outputgen
)

const Storage1FileName = "RPNStorageName.gob" // Allows for a rotation of Storage files, in case of a mistake.
const Storage2FileName = "RPNStorageName2.gob"
const Storage3FileName = "RPNStorageName3.gob"
const DisplayTapeFilename = "displaytape.txt"
const TextFilenameOut = "rpntcelloutput.txt"
const TextFilenameIn = "rpntcellinput.txt"
const HelpFileName = "helprpn.txt"

type keyStructType struct {
	r    rune
	name string
}

var gblrow int // = 0 by default

var style, plain, bold, reverse tcell.Style
var Green = style.Foreground(tcell.ColorGreen)
var Cyan = style.Foreground(tcell.ColorAqua)
var Yellow = style.Foreground(tcell.ColorYellow)
var Red = style.Foreground(tcell.ColorRed)
var BoldYellow = Yellow.Bold(true)
var BoldRed = Red.Bold(true)
var BoldGreen = Green.Bold(true)

var scrn tcell.Screen

func putln(str string) {
	puts(scrn, style, 1, gblrow, str)
	gblrow++
}

func putfln(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	puts(scrn, style, 1, gblrow, s)
	gblrow++
}

func putf(x, y int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	puts(scrn, style, x, y, s)
}

func puts(scrn tcell.Screen, style tcell.Style, x, y int, str string) { // orig designed to allow for non ASCII characters.  I removed that.
	for i, r := range str {
		scrn.SetContent(x+i, y, r, nil, style)
	}
	x += len(str)

	deleol(x, y) // no longer crashes here.
	scrn.Show()
}

func deleol(x, y int) {
	width, _ := scrn.Size() // don't need height for this calculation.
	empty := width - x - 1
	for i := 0; i < empty; i++ {
		scrn.SetContent(x+i,y,' ',nil, plain)  // making a blank slice kept crashing.  This direct method works.
	}
}

func clearline(line int) {
	deleol(0, line)
}

/*
func puts(scrn tcell.Screen, style tcell.Style, x, y int, str string) {
	i := 0
	var deferred []rune
	dwidth := 0
	zwj := false
	for _, r := range str {
		if r == '\u200d' {
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
			deferred = append(deferred, r)
			zwj = true
			continue
		}
		if zwj {
			deferred = append(deferred, r)
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				scrn.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				scrn.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		scrn.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
	scrn.Show()
}
*/
func main() {
	var INBUF, HomeDir string

	var x int

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.
	var err error

	scrn, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err = scrn.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defer scrn.Fini()

	MaxCol, MaxRow = scrn.Size()

	scrn.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	scrn.Clear()

	style = tcell.StyleDefault
	plain = tcell.StyleDefault
	bold = style.Bold(true)
	reverse = style.Reverse(true)

//	style = bold     looks ugly.  I'm removing it.
	putfln("RPN Calculator written in Go.  Last updated %s.", LastAltered)
	style = plain

	stringslice = make([]string, 0, 35)
	sigfig = -1 // now only applies to WriteRegToScreen
	StartRow := 0
	StartCol := 0
	outputmode = outputfix

	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
	} else { // then HomeDir will be empty.
		putln(" runtime.GOOS does not say linux or windows.  Is this a Mac?")
	}
	Divider = strings.Repeat("-", MaxCol-StartCol)

	x = StartCol
	TitleRow = StartRow
	StackRow = StartRow + 4
	RegRow = StackRow + 12
	OutputRow = RegRow + 10
	DisplayCol = MaxCol/2 + 2
	PromptRow = StartRow + 1
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	DisplayTape = make([]string, 0, 100)
	theFileExists := true
	StorageFullFilename := HomeDir + string(os.PathSeparator) + Storage1FileName
	Storage2FullFilename := HomeDir + string(os.PathSeparator) + Storage2FileName
	Storage3FullFilename := HomeDir + string(os.PathSeparator) + Storage3FileName

	thefile, err := os.Open(StorageFullFilename) // open for reading
	if os.IsNotExist(err) {
		log.Print(" thefile does not exist for reading. ")
		putln("thefile does not exist for reading.")
		theFileExists = false
	} else if err != nil {
		log.Printf(" Error from os.Open(Storage1FileName).  Possibly because no Stack File found: %v\n", err)
		putfln("Error from os.Open(Storage1FileName.  Possibly because nostack file found: %v ", err)
		theFileExists = false
	}
	if theFileExists {
		defer thefile.Close()
		decoder := gob.NewDecoder(thefile)       // decoder reads the file.
		err = decoder.Decode(&Stk)               // decoder reads the file.
		check(err)                               // theFileExists, so panic if this is an error.
		for i := hpcalc.T1; i >= hpcalc.X; i-- { // Push in reverse onto the stack in hpcalc.
			hpcalc.PUSHX(Stk[i])
		}

		err = decoder.Decode(&Storage) // decoder reads the file.
		check(err)                     // theFileExists, so panic if this is an error.

		thefile.Close()
	} // thefileexists for both the Stack variable, Stk, and the Storage registers.

	hpcalc.PushMatrixStacks()

	WriteStack(x, StackRow)
	n = WriteRegToScreen(x, RegRow)
	if n > 8 {
		OutputRow = RegRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
		PromptRow = StartRow + 1
	}

	//  Print_tb(x,PromptRow,BrightCyan,Black,InputPrompt);  Doesn't make any difference, it seems.
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
	} else {
		//		Print_tb(x, PromptRow, BrightCyan, Black, InputPrompt)
		puts(scrn, Cyan, x, PromptRow, InputPrompt)
		x += len(InputPrompt) + 2
		scrn.ShowCursor(x, PromptRow)
		INBUF = GetInputString(x, PromptRow)
		if strings.HasPrefix(INBUF, "Q") {
			os.Exit(0)
		}
		x = StartCol
	} // if command tail exists
	scrn.Show()

	INBUF = strings.ToUpper(INBUF)

	hpcalc.PushMatrixStacks()

	for len(INBUF) > 0 { // Main processing loop
		DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
		x = StartCol
		// These commands are not run thru hpcalc as they are processed before calling GetResult.
		if INBUF == "ZEROREG" {
			for c := range Storage {
				Storage[c].Value = 0.0
				Storage[c].Name = ""
			}
		} else if strings.HasPrefix(INBUF, "STO") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // The 4th position.
				i = GetRegIdx(ch)
			}
			Storage[i].Value = hpcalc.READX()
			n = WriteRegToScreen(x, RegRow)
			if n > 8 {
				clearline(PromptRow)
				clearline(OutputRow)
				OutputRow = RegRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
				PromptRow = StartRow + 1   // used to be OutputRow -1
			}
			if i > 0 {
				Storage[i].Name = GetNameStr()
			}
		} else if strings.HasPrefix(INBUF, "RCL") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // the 4th position.
				i = GetRegIdx(ch)
			}
			hpcalc.PUSHX(Storage[i].Value)
			RepaintScreen(StartCol)
		} else if strings.HasPrefix(INBUF, "SHO") { // so it will match SHOW and SHO
			n = WriteRegToScreen(StartCol, RegRow)
			if n > 8 {
				OutputRow = RegRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
				PromptRow = StartRow + 1   // used to be OutputRow -1
			}
			WriteDisplayTapeToScreen(DisplayCol, StackRow)
		} else if strings.HasPrefix(INBUF, "NAME") {
			//var ans string
			var i int // remember that this auto-zero'd
			if len(INBUF) > 4 {
				ch := INBUF[4] // the 5th position
				i = GetRegIdx(ch)
			}
			//promptstr := "   Input name string : "
			//Print_tb(1, OutputRow, BrightYellow, Black, promptstr)
			//ans = GetInputString(len(promptstr)+2, OutputRow)
			Storage[i].Name = GetNameStr()
		} else if strings.HasPrefix(INBUF, "SIG") || strings.HasPrefix(INBUF, "FIX") { // SigFigN command, or FIX
			ch := INBUF[len(INBUF)-1] // ie, the last character.
			sigfig = GetRegIdx(ch)
			if sigfig > 9 { // If sigfig greater than this max value, make it -1 again.
				sigfig = -1
			}
			_, _ = hpcalc.GetResult(INBUF) // Have to send this into hpcalc also
		} else if INBUF == "HELP" || INBUF == "?" {
			WriteHelp(StartCol+2, StackRow)
		} else if strings.HasPrefix(INBUF, "DUMP") {
			// do nothing, ie, don't send it into hpcalc.GetResult
		} else if strings.HasPrefix(INBUF, "OUTPUTFI") { // allow outputfix, etc
			outputmode = outputfix
		} else if strings.HasPrefix(INBUF, "OUTPUTFL") { // allow outputfloat, etc
			outputmode = outputfloat
		} else if strings.HasPrefix(INBUF, "OUTPUTGE") { // allow outputgen, etc.
			outputmode = outputgen
		} else if INBUF == "CLEAR" || INBUF == "CLS" {
			scrn.Clear()
			RepaintScreen(0)
		} else if INBUF == "REPAINT" {
			RepaintScreen(StartCol)
		} else if INBUF == "DEBUG" {
			//			Printf_tb(x, OutputRow+8, BrightCyan, Black, " HP-type RPN calculator written in Go.  Last altered %s", LastAltered)
			style = Cyan
			putf(x, OutputRow+8, " HP-type RPN calculator written in Go.  Last altered %s", LastAltered)
			//			Printf_tb(0, OutputRow+9, BrightCyan, Black, "%s was last linked on %s.  Full executable is %s.", ExecFI.Name(), LastLinkedTimeStamp, execname)
			putf(0, OutputRow+9, "%s was last linked on %s.  Full executable is %s.", ExecFI.Name(), LastLinkedTimeStamp, execname)

			style = Yellow
			putf(StartCol, OutputRow+10, " StartCol=%d,StartRow=%d,MaxCol=%d,MaxRow=%d,TitleRow=%d,StackRow=%d,RegRow=%d,OutputRow=%d,PromptRow=%d",
				StartCol, StartRow, MaxCol, MaxRow, TitleRow, StackRow, RegRow, OutputRow, PromptRow)
			putf(StartCol, OutputRow+11, " DisplayCol=%d", DisplayCol)
			putf(x, OutputRow+13, " StorageFullFilename 1:%s, 2:%s, 3:%s", StorageFullFilename, Storage2FullFilename, Storage3FullFilename)
			style = Cyan
		} else if strings.HasPrefix(INBUF, ":W") || strings.HasPrefix(INBUF, "WR") {
			xstring := GetXstring()
			XStringFile, err := os.OpenFile(TextFilenameOut, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				style = Yellow
				putf(0, OutputRow, " Error %v while opening %s", err, TextFilenameOut)
				style = Cyan
				//os.Exit(1)
			}
			defer XStringFile.Close()
			XstringWriter := bufio.NewWriter(XStringFile)
			defer XstringWriter.Flush()
			today := time.Now()
			datestring := today.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
			_, err = XstringWriter.WriteString("------------------------------------------------------\n")
			_, err = XstringWriter.WriteString(datestring)
			_, err = XstringWriter.WriteRune('\n')
			_, err = XstringWriter.WriteString(xstring)
			_, err = XstringWriter.WriteRune('\n')
			check(err)

			_, err = XstringWriter.WriteString("\n\n")
			check(err)
			XstringWriter.Flush()
			XStringFile.Close()
		} else if strings.HasPrefix(INBUF, ":R") || INBUF == "READ" || INBUF == "RD" {
			XstringFileExists := true
			XstringFile, err := os.Open(TextFilenameIn) // open for reading
			if os.IsNotExist(err) {
				style = Yellow
				putf(0, OutputRow, "\n %s does not exist for reading in this directory.  Command ignored.\n", TextFilenameIn)
				style = Cyan
				XstringFileExists = false
			} else if err != nil {
				style = BoldYellow
				putf(0, OutputRow, "\n %s does not exist for reading in this directory.  Command ignored.\n", TextFilenameIn)
				style = Cyan
				XstringFileExists = false
			}

			if XstringFileExists {
				defer XstringFile.Close()
				XstringScanner := bufio.NewScanner(XstringFile)
				XstringScanner.Scan()
				Xstring := strings.TrimSpace(XstringScanner.Text())
				if err := XstringScanner.Err(); err != nil {
					log.Println(" Error while reading from ", TextFilenameIn, ".  Error is ", err, ".  Command ignored.")
				} else {
					r, err := strconv.ParseFloat(Xstring, 64)
					check(err)
					// fmt.Println(" Read ", r, " from ", TextFilenameIn, ".")  a debugging statement
					hpcalc.PUSHX(r)
				}
			}
		} else {

			// ----------------------------------------------------------------------------------------------
			_, stringslice = hpcalc.GetResult(INBUF) // Here is where GetResult is called
			// ----------------------------------------------------------------------------------------------
			y := OutputRow
			for _, s := range stringslice {
				puts(scrn, Yellow, x, y, s)
				y++
			}

			for y < MaxRow {
				clearline(y)
				y++
			}
		}
		// Don't understand why this next line helps stabilize the output display, but it does.
		//Print_tb(x, OutputRow+len(stringslice)+1, BrightCyan, Black, "-------------")

		//  These commands are processed after GetResult is called, so these commands are run thru hpcalc.
		if strings.ToLower(INBUF) == "about" { // I'm using ToLower here just to experiment a little.
			style = Yellow
			putf(x, OutputRow+1, " Last altered rpntcell %s, last linked %s. ", LastAltered, LastLinkedTimeStamp)
			style = Cyan
		}

		if !(INBUF == "CLEAR" || INBUF == "CLS") {
			RepaintScreen(StartCol)
		}
		x = StartCol
		puts(scrn, Cyan, x, PromptRow, InputPrompt)
		x += len(InputPrompt) + 2
		scrn.ShowCursor(x, PromptRow)
		scrn.Show()
		ans := GetInputString(x, PromptRow)
		INBUF = strings.ToUpper(ans)
		if len(INBUF) == 0 || strings.HasPrefix(INBUF, "Q") || INBUF == "EXIT" {
			fmt.Println()
			break
		}
	} // End Main Processing For Loop

	// Time to write files before exiting.

	// Rotate StorageFileNames and write
	err = os.Rename(Storage2FullFilename, Storage3FullFilename)
	if err != nil {
		_ = fmt.Errorf(" Rename of storage 2 to storage 3 failed with error %v \n", err)
	}

	err = os.Rename(StorageFullFilename, Storage2FullFilename)
	if err != nil {
		_ = fmt.Errorf(" Rename of storage 1 to storage 2 failed with error %v \n", err)
	}

	thefile, err = os.Create(StorageFullFilename)        // for writing
	checkmsg(err, "After os.Create StorageFullFilename") // This should not fail, so panic if it does fail.
	defer thefile.Close()

	Stk = hpcalc.GETSTACK()
	encoder := gob.NewEncoder(thefile)        // encoder writes the file
	err = encoder.Encode(Stk)                 // encoder writes the file
	checkmsg(err, "after encoder.Encode Stk") // Panic if there is an error
	err = encoder.Encode(Storage)             // encoder writes the file
	checkmsg(err, "after encoder.Encode Storage")

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
		checkmsg(err, "after DisplayTapeWriter WriteSrting and WriteRune")
	}
	_, err = DisplayTapeWriter.WriteString("\n\n")
	checkmsg(err, "after last DisplayTapeWriter WriteString newline newline")

	err = DisplayTapeWriter.Flush()
	checkmsg(err, "After last DisplayTapeWriter flush")

	err = DisplayTapeFile.Close()
	checkmsg(err, "after DisplayTapeFile close")
} // main in rpntcell.go

/* ------------------------------------------------------------ GetRegIdx --------- */
func GetRegIdx(chr byte) int {
	// Return 0..35 w/ A = 10 and Z = 35
	ch := tokenize.CAP(chr)
	if (ch >= '0') && (ch <= '9') {
		ch = ch - '0'
	} else if (ch >= 'A') && (ch <= 'Z') {
		ch = ch - 'A' + 10
	} else { // added 12/11/2016 to fix bug
		ch = 0
	}
	return int(ch)
} // GetRegIdx
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

// ------------------------------------------------------------ WriteRegToScreen --------------
func WriteRegToScreen(x, y int) int { // Outputs the number of reg's that are not zero.
	FirstNonZeroStorageFlag := true
	n := 0

	for i, r := range Storage {
		if r.Value != 0.0 {
			if FirstNonZeroStorageFlag {
				s := "The following storage registers are not zero"
				puts(scrn, Yellow, x, y, s)
				y++
				FirstNonZeroStorageFlag = false
			} // if firstnonzerostorageflag
			ch := GetRegChar(i)
			s := ""
			if outputmode == outputfix {
				s = strconv.FormatFloat(r.Value, 'f', sigfig, 64) // sigfig of -1 means max sigfig.
				s = hpcalc.CropNStr(s)
				if r.Value >= 10000 {
					s = hpcalc.AddCommas(s)
				}
			} else if outputmode == outputfloat {
				s = strconv.FormatFloat(r.Value, 'e', sigfig, 64) // sigfig of -1 means max sigfig.
			} else { // outputmode has to be outputgen
				s = strconv.FormatFloat(r.Value, 'g', sigfig, 64) // sigfig of -1 means max sigfig.
			}

			//			Printf_tb(x, y, BrightCyan, Black, " Reg [%s], %s =  %s", ch, r.Name, s)
			style = Cyan
			putf(x, y, " Reg [%s], %s =  %s", ch, r.Name, s)
			//			deleol(x+len(s),y)
			y++
			n++
		} // if storage value is not zero
	} // for range over Storage
	if FirstNonZeroStorageFlag {
		//		Print_tb(x, y, BrightYellow, Black, " All storage registers are zero.")
		puts(scrn, Yellow, x, y, " All storage registers are zero.")
		y++
	}
	style = Cyan
	return n
} // WriteRegToScreen

// --------------------------------------------------------- WriteDisplayTapeToScreen ----------------
func WriteDisplayTapeToScreen(x, y int) {
	//	Print_tb(x, y, BrightCyan, Black, "DisplayTape")
	puts(scrn, Cyan, x, y, "DisplayTape")
	y++
	for _, s := range DisplayTape {
		//		Print_tb(x, y, BrightYellow, Black, s)
		puts(scrn, Green, x, y, s)
		y++
	} // for ranging over DisplayTape slice of strings
} // WriteDisplayTapeToScreen

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------- checkmsg -------------------------------
func checkmsg(err error, msg string) {
	if err != nil {
		_ = fmt.Errorf("%s %v \n", msg, err) // writes to stderr instead of stdout.
		panic(err)
	}
}

// --------------------------------------------------- Cap -----------------------------------------
func Cap(c rune) rune {
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap

// --------------------------------------------------- GetInputString for tcell--------------------------------------

func GetInputString(x, y int) string {

	deleol(x, y)
	scrn.ShowCursor(x, y)
	scrn.Show()
	donechan := make(chan bool)
	keychannl := make(chan rune)
	helpchan := make(chan bool)
	delchan := make(chan bool)
	upchan := make(chan bool)
	downchan := make(chan bool)
	homechan := make(chan bool)
	endchan := make(chan bool)
	leftchan := make(chan bool)
	rightchan := make(chan bool)

	pollevent := func() {
		for {
			event := scrn.PollEvent()
			switch event := event.(type) {
			case *tcell.EventKey:
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					donechan <- true // I don't have to send true to quit.
					return
				case tcell.KeyCtrlL:
					scrn.Sync()
				case tcell.KeyF1, tcell.KeyF2:
					// help
					helpchan <- true
					return

				case tcell.KeyBackspace, tcell.KeyDEL, tcell.KeyDelete:
					delchan <- true
					// do not return after any of these keys are hit, as an entry is being edited.

				case tcell.KeyPgUp, tcell.KeyUp:
					upchan <- true
					return

				case tcell.KeyPgDn, tcell.KeyDown:
					downchan <- true
					return

				case tcell.KeyRight, tcell.KeyUpRight, tcell.KeyDownRight:
					rightchan <- true
					return

				case tcell.KeyLeft, tcell.KeyUpLeft, tcell.KeyDownLeft:
					leftchan <- true
					return

				case tcell.KeyHome:
					homechan <- true
					return

				case tcell.KeyEnd:
					endchan <- true
					return

				case tcell.KeyRune:
					r := event.Rune()
					keychannl <- r
					if r == ' ' {
						return
					}
				}
			case *tcell.EventResize:
				scrn.Sync()
			}
		}
	}

	go pollevent()

	bs := make([]byte, 0, 100) // byteslice to build up the string to be returned.
	for {
		select {
		case <-donechan: // reading from quitchan will block until its closed.
			return string(bs)

		case <-helpchan:
			putfln("help message received.  %s", "enter key is delimiter")
			return "help"

		case <-delchan:
			if len(bs) > 0 {
				bs = bs[:len(bs)-1]
			}
			puts(scrn, style, x+len(bs), y, " ")
			scrn.ShowCursor(x+len(bs), y)
			scrn.Show()

		case <-upchan:
			return "up"

		case <-downchan:
			return "dn"

		case <-homechan:
			return  "up"   // "home key"

		case <-endchan:
			return  "dn"   //"end key"

		case <-rightchan:
			return "~"

		case <-leftchan:
			return "~"

		case key := <-keychannl:
			if key == ' ' {
				if len(bs) > 0 {
					return string(bs)
				} else {
					go pollevent() // need to restart the go routine to fetch more keys.
					continue       // discard this extaneous space
				}
			} else if key == '=' {
				key = '+'
			} else if key == ';' {
				key = '*'
			}
			bs = append(bs, byte(key))
			puts(scrn, style, x, y, string(bs))

			scrn.ShowCursor(x+len(bs), y)
			scrn.Show()
		}
	}
} // GetInputString for tcell

// ------------------------------------------- GetXstring ------------------------------------------
func GetXstring() string {

	if outputmode == outputfix {
		_, stringslice = hpcalc.GetResult("DUMPFIXED")
	} else if outputmode == outputfloat {
		_, stringslice = hpcalc.GetResult("DUMPFLOAT")
	} else if outputmode == outputgen {
		_, stringslice = hpcalc.GetResult("DUMP")
	}
	return stringslice[len(stringslice)-2]
}

// ------------------------------------------- WriteStack ------------------------------------------
func WriteStack(x, y int) {

	if outputmode == outputfix {
		_, stringslice = hpcalc.GetResult("DUMPFIXED")
	} else if outputmode == outputfloat {
		_, stringslice = hpcalc.GetResult("DUMPFLOAT")
	} else if outputmode == outputgen {
		_, stringslice = hpcalc.GetResult("DUMP")
	}

	puts(scrn, Yellow, x+10, y, stringslice[len(stringslice)-2]) // just gets X register to be output in Yellow
	//	deleol(x+10+len(GetXstring()), y)
	y++
	for _, s := range stringslice {
		puts(scrn, Cyan, x, y, s)
		deleol(x+len(s), y)
		y++
	}
} // end WriteStack

//--------------------------------------------- WriteHelp -------------------------------------------
func WriteHelp(x, y int) { // starts w/ help text from hpcalc, and then adds help from this module.
	var HelpFile *bufio.Writer

	_, helpstringslice := hpcalc.GetResult("HELP")
	helpstringslice = append(helpstringslice, " STOn,RCLn  -- store/recall the X register to/from the memory register.")
	helpstringslice = append(helpstringslice, " Outputfix, outputfloat, outputgen -- outputmodes for stack display.")
	helpstringslice = append(helpstringslice, " NAME -- NAME registers with strings, Use - for spaces in these strings.")
	helpstringslice = append(helpstringslice, " Clear, CLS -- clear screen.")
	helpstringslice = append(helpstringslice, " EXIT,(Q)uit -- Needed after switch to use ScanWords in bufio scanner.")
	helpstringslice = append(helpstringslice, fmt.Sprintf(" :w, wr -- write X register to text file %s.", TextFilenameOut))
	helpstringslice = append(helpstringslice, fmt.Sprintf(" :r, rd, read -- read X register from first line of %s.", TextFilenameIn))
	helpstringslice = append(helpstringslice, " Debug -- Print debugging message to screen.")
	helpstringslice = append(helpstringslice, " SigN, FixN -- set significant figures for displayed numbers to N.  Default is -1.")
	helpstringslice = append(helpstringslice, " outputfix, outputfloat, outputgen -- sets output mode for displayed numbers.")

	// Will always open this file in the current working directory instead of the HomeDir.
	// This is different than rpnterm, which only writes this file if it's not already there.
	HelpOut, err := os.Create(HelpFileName)
	check(err)
	defer HelpOut.Close()
	HelpFile = bufio.NewWriter(HelpOut)
	defer HelpFile.Flush()

	for _, s := range helpstringslice {
		HelpFile.WriteString(s)
		HelpFile.WriteRune('\n')
	}

	HelpFile.Flush()
	HelpOut.Close()

	if y+len(helpstringslice) >= MaxRow {
		FI, err := os.Stat(HelpFileName)
		check(err)
		//		Printf_tb(x, y, BrightYellow, Black, " Too many help lines for this small screen.  See %s.", HelpFileName)
		style = BoldGreen
		putf(x, y, " Too many help lines for this small screen.  See %s.", HelpFileName)
		yr, m, d := FI.ModTime().Date()
		putf(x, y, "%s from %d/%d/%d is in current directory.", FI.Name(), m, d, yr)
		style = Cyan
		return
	}

	//	scrn.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack))
	scrn.Clear()

	gblrow = y
	for _, s := range helpstringslice {
		putln(s)
	}

	style = Cyan
	putln(" pausing ")
	scrn.ShowCursor(x+11, gblrow)
	_ = GetInputString(x+11, gblrow)

	//	scrn.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack))
	scrn.Clear()

	RepaintScreen(x)
} // end WriteHelp

// ------------------------------------------------------- Repaint ----------------------------------
func RepaintScreen(x int) {

	// Printf_tb(x, TitleRow, BrightCyan, Black, " HP-type RPN calculator written in Go.  Last altered %s", LastAltered)  Removed 4/15/19
	WriteStack(x, StackRow)
	n = WriteRegToScreen(x, RegRow)
	if n > 8 {
		OutputRow = RegRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
		PromptRow = StartRow + 1   // PromptRow = OutputRow - 1 was prev assignment.
	}
	WriteDisplayTapeToScreen(DisplayCol, StackRow)
	//	Printf_tb(x, MaxRow-1, BrightCyan, Black, Divider)  Not needed for tcell
	gblrow = 0
}

// -------------------------------------------------- GetNameStr --------------------------------
func GetNameStr() string {
	var ans string
	promptstr := "   Input name string, making - or = into a space : "
	puts(scrn, Yellow, 1, PromptRow, promptstr)
	ans = GetInputString(len(promptstr)+2, PromptRow)
	answer := strings.ToUpper(ans) // don't return a ToUpper(ans)
	if answer == "TODAY" || answer == "T" {
		m, d, y := timlibg.TIME2MDY()
		ans = timlibg.MDY2STR(m, d, y)
	} else {
		ans = StrSubst(ans) // will make - or = into a space.
	}
	return ans
}

// -------------------------------------------------- StrSubst -----------------------------------
func StrSubst(instr string) string { // copied from makesubst package.

	instr = strings.TrimSpace(instr)
	inRune := make([]rune, len(instr))

	for i, s := range instr {
		switch s {
		case '+':
			s = ' '
		case '-':
			s = ' '
		}
		inRune[i] = s // was byte(s) before I made this a slice of runes.
	}
	return string(inRune)
} // makesubst

// ---------------------------------------------------- End rpntcell.go ------------------------------
