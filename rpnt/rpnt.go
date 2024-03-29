// (C) 1990-2023.  Robert W.  Solomon.  All rights reserved.
// rpnt.go
package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
	"runtime"
	"src/timlibg"
	"src/tknptr"
	"strconv"
	"strings"
	"time"

	//"getcommandline"
	hpcalc "src/hpcalc2"
	"src/tokenize"
	//	runewidth "github.com/mattn/go-runewidth"  Not needed after I simplified puts()
)

// runtime.GOOS returns either linux or windows.  I have not tested mac.  I want either $HOME or %userprofile to set the write dir.  Supplanted by os.UserHomeDir.

/*
This module uses the HPCALC2 module to simulate an RPN type calculator.

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
16 Feb 20 -- Will use ! to recall a command in the history string slice.  Like what bash can do.  I don't need a factorial command anyway.
               Now it counts from bottom, where ! is last command, and !1 is next cmd.  Not sure if I want it to count from the top.
11 Mar 20 -- Noticed that my linux code runs on WSL on bash, except for this pgm.  I'm expanding the error reporting to see if I can track this and
               maybe file an issue.
 7 Apr 20 -- Will add the clipboard code I first wrote for rpng, and will have PgUp/PgDn manipulate the stack, while Up and Dn will allow command recall.
 2 Jul 20 -- Removed a duplicate help line.
 3 Jul 20 -- FIX command also sets output mode to be fixed.  So to use FIXn to set digits for float or gen, have to change outputmode AFTER using FIXn.
 8 Aug 20 -- Now using hpcalc2 in place of hpcalc.
23 Oct 20 -- Adding flag package to allow the -n flag, meaning no files read or written.
25 Oct 20 -- Trying to improve the display, as it looks terrible on (only) terminator for some reason.
 8 Nov 20 -- Fixed minor errors in the fmt messages of FROMCLIP.  I was looking because of "Go Standard Library Cookbook"
 9 Nov 20 -- Added use of comspec.
12 Dec 20 -- Started to think about mapped storage registers that can have string names, stored in a file.  I'm moving this to hpcalc2.
13 Dec 20 -- Tweaked help message and other minor stuff.  I commented out TOCLIP, FROMCLIP here, as it's now in hpcalc2.  Added LABEL as synonym for NAME.
15 Dec 20 -- Tweaked help message some more.
21 Dec 20 -- Changed LABEL to LABL so that it's 4 characters, which is needed for GetIndex logic.
 4 Feb 21 -- Added H for help.
17 Jun 21 -- Converted to modules, and added the stuff I've learned over the last 4 months.
24 Jul 21 -- tcell v2 is now imported.
31 Dec 21 -- Decided to call it rpnt.go, and will clear out dead code.
 5 May 22 -- HPCALC2 now uses OS specific code to handle clipboard events.  No changes made here.
31 Aug 22 -- Changing proccessing of commands to match rpnf and rpng.  IE, will respect commands like rcl+.  And removing the "!" command recall syntax.  I could
               not remember its syntax anyway, and I did not implement this in rpnf or rpng anyway.
               And I added verboseFlag, but I can't see what it's writing to the screen, so I'll write to a file in the current directory.
18 Feb 23 -- Changing from os.UserHomeDir to os.UserConfigDir.  This is %appdata% or $HOME/.config
24 Jun 23 -- Changed hpcalc2, so I have to recompile.  And I changed some comments.
 3 Sep 23 -- I noticed that entering a stack command discards whatever characters were already entered.  I fixed that, and I added subst tilde for backtick.
              And I added a buffer of 1 to the channels in InputString.  Before, these were unbuffered channels so they could block when I didn't expect it.
              This may have affected behavior.
*/

const LastAltered = "3 Sep 2023"

const InputPrompt = " Enter calculation, HELP or <return> to exit: "

type Register struct {
	Value float64
	Name  string
}

var Storage [36]Register // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var Divider string
var clear map[string]func()

var startCol, startRow, sigfig, maxRow, maxCol, titleRow, stackRow, regRow, outputRow, displayCol, promptRow, outputmode, n int

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
const verboseFileName = "verbose.txt"

/*
const SpaceFiller = "     |     "       Not used so I commented it out 6/17/21.
type keyStructType struct {             Not used, so I commented it out 6/17/21
	r    rune
	name string
}
*/

var gblrow int // = 0 by default

var style, plain, bold, reverse tcell.Style
var Green = style.Foreground(tcell.ColorGreen)
var Cyan = style.Foreground(tcell.ColorAqua)
var Yellow = style.Foreground(tcell.ColorYellow)
var Red = style.Foreground(tcell.ColorRed)
var BoldYellow = Yellow.Bold(true)
var BoldRed = Red.Bold(true)
var BoldGreen = Green.Bold(true)
var verboseFile *os.File

var scrn tcell.Screen

var verboseFlag = flag.Bool("v", false, "Verbose mode, mostly for debugging.")
var noFileFlag = flag.Bool("n", false, "no files read or written.") // pointer

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

// ---------------------------------------------------------------------------------------------------------------------
func puts(scrn tcell.Screen, style tcell.Style, x, y int, str string) { // orig designed to allow for non ASCII characters.  I removed that.
	for i, r := range str {
		scrn.SetContent(x+i, y, r, nil, style)
	}
	x += len(str)

	deleol(x, y) // no longer crashes here.
	//scrn.Show()
}

// ---------------------------------------------------------------------------------------------------------------------
func deleol(x, y int) {
	width, _ := scrn.Size() // don't need height for this calculation.
	//empty := width - x - 1
	for i := 0; i < width; i++ {
		scrn.SetContent(x+i, y, ' ', nil, plain) // making a blank slice kept crashing.  This direct method works.
	}
	//scrn.SetContent(width-1, y, '|', nil, plain) not needed.
}

// ---------------------------------------------------------------------------------------------------------------------
func clearline(line int) {
	deleol(0, line)
}

// ---------------------------------------------------------------------------------------------------------------------
func main() {
	var INBUF, HomeDir string

	var x int

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.
	var err error

	flag.Parse()

	scrn, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, " error from tcell.NewScreen %v \n", err)
		log.Println(" error from tcell.NewScreen call ", err)
		errorfile, e := os.Create("errorfile.txt")
		if e != nil {
			log.Println(" error while creating errorfile.txt.  Looks like I'm fucked.", e)
		}
		defer errorfile.Close()
		errorfilebuffered := bufio.NewWriter(errorfile)
		s := fmt.Sprintf(" Error from tcell.NewScreen is %v \n", err)
		errorfilebuffered.WriteString(s)
		errorfilebuffered.Flush()
		os.Exit(1)
	}

	if err = scrn.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, " error from scrn.Init %v \n", err)
		log.Println(" error from scrn.Init is ", err)
		error2file, e2 := os.Create("error2file.txt")
		if e2 != nil {
			log.Println(" error creating error2file.txt.  Looks like I'm fucked again.  ", e2)
		}
		defer error2file.Close()
		error2filebuffered := bufio.NewWriter(error2file)
		s := fmt.Sprintf(" Error from scrn.Init is %v \n", err)
		error2filebuffered.WriteString(s)
		error2filebuffered.Flush()
		os.Exit(1)
	}

	defer scrn.Fini()

	maxCol, maxRow = scrn.Size()

	//                       scrn.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack))
	scrn.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)) // this worked.
	scrn.Clear()

	style = tcell.StyleDefault
	plain = tcell.StyleDefault
	bold = style.Bold(true)
	reverse = style.Reverse(true)

	//	style = bold     looks ugly.  I'm removing it.
	putfln("RPN Calculator written in Go.  Last updated %s, and compiled w/ %s.", LastAltered, runtime.Version())
	style = plain

	stringslice = make([]string, 0, 50)
	sigfig = -1 // now only applies to WriteRegToScreen
	// StartRow := 0  Now defined globally,
	// StartCol := 0  and of course it's initialized to zero.
	outputmode = outputfix

	//HomeDir, err = os.UserHomeDir()
	HomeDir, err = os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error from os.UserConfigDir() is", err)
		os.Exit(1)
	}
	Divider = strings.Repeat("-", maxCol-startCol)

	x = startCol
	titleRow = startRow
	stackRow = startRow + 4
	regRow = stackRow + 12
	outputRow = regRow + 10
	displayCol = maxCol/2 + 2
	promptRow = startRow + 1
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	DisplayTape = make([]string, 0, 100)
	theFileExists := true
	StorageFullFilename := HomeDir + string(os.PathSeparator) + Storage1FileName
	Storage2FullFilename := HomeDir + string(os.PathSeparator) + Storage2FileName
	Storage3FullFilename := HomeDir + string(os.PathSeparator) + Storage3FileName

	var thefile *os.File
	if !*noFileFlag {
		thefile, err = os.Open(StorageFullFilename) // open for reading
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
	}

	WriteStack(x, stackRow)
	n = WriteRegToScreen(x, regRow)
	if n > 8 {
		outputRow = regRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
		promptRow = startRow + 1
	}

	args := flag.Args()
	if len(args) > 0 {
		INBUF = strings.Join(args, " ")
	} else {
		puts(scrn, Cyan, x, promptRow, InputPrompt)
		x += len(InputPrompt) + 2
		scrn.ShowCursor(x, promptRow)
		INBUF = GetInputString(x, promptRow)
		if strings.HasPrefix(INBUF, "Q") {
			os.Exit(0)
		}
		x = startCol
	} // if command tail exists
	scrn.Show()

	hpcalc.PushMatrixStacks()
	//defer hpcalc.mapWriteAndClose()  Part of the edits to hpcalc2 was to not have 1 calculator pgm clobber the map file of another.  Commented this out 6/24/23.

	if *verboseFlag {
		verboseFile, err = os.OpenFile(verboseFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from opening the verbose file is %v\n", err)
			*verboseFlag = false
		} else {
			verboseFile.WriteString(Divider)
			verboseFile.WriteString("\n")
			defer verboseFile.Close()
		}
	}

	for len(INBUF) > 0 { // Main processing loop for whole lines
		INBUF = strings.ToUpper(INBUF)
		DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
		realTknSlice := tknptr.RealTokenSlice(INBUF)
		if *verboseFlag {
			s := fmt.Sprintf(" INBUF=%q, length of realTknSlice=%d\n", INBUF, len(realTknSlice))
			verboseFile.WriteString(s)
		}
		for _, rtkn := range realTknSlice { // Main processing loop for tokens within a line.
			// check for new use history command patterned after bash, ie, using ! to start it.  Nevermind.  Removed 31 Aug 22.
			/*
				if strings.HasPrefix(INBUF, "!") {
					i := 0
					if len(INBUF) > 1 {
						ch := INBUF[1] // the 2nd position
						i = GetRegIdx(ch)
					}
					INBUF = GetHx(i)
				} else if INBUF == "UP" {
					INBUF = GetHx(1)

				} else if INBUF == "DN" {
					// leaving it alone will pop the stack down.

				} else if INBUF == "PGUP" {
					INBUF = "UP" // and sent into GetResult to push the stack up

				} else if INBUF == "PGDN" {
					INBUF = "DN" // and sent into GetResult to pop the stack down

				} else { // only put typed command lines into the hx.
					DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
				}

			*/
			if *verboseFlag {
				s := fmt.Sprintf(" rtkn.Str=%q, and %#v\n", rtkn.Str, rtkn)
				verboseFile.WriteString(s)
			}

			x = startCol
			// These commands are not run thru hpcalc2 as they are processed before calling GetResult.
			if rtkn.Str == "ZEROREG" {
				for c := range Storage {
					Storage[c].Value = 0.0
					Storage[c].Name = ""
				}
			} else if strings.HasPrefix(rtkn.Str, "STO") {
				i := 0
				if len(rtkn.Str) > 3 {
					ch := rtkn.Str[3] // The 4th position.
					i = GetRegIdx(ch)
				}
				Storage[i].Value = hpcalc.READX()
				n = WriteRegToScreen(x, regRow)
				if n > 8 {
					clearline(promptRow)
					clearline(outputRow)
					outputRow = regRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
					promptRow = startRow + 1   // used to be OutputRow -1
				}
				if i > 0 {
					Storage[i].Name = GetNameStr()
				}
			} else if strings.HasPrefix(rtkn.Str, "RCL") {
				i := 0
				if len(rtkn.Str) > 3 {
					ch := rtkn.Str[3] // the 4th position.
					i = GetRegIdx(ch)
				}
				hpcalc.PUSHX(Storage[i].Value)
				RepaintScreen(startCol)
			} else if strings.HasPrefix(rtkn.Str, "SHO") || rtkn.Str == "LS" || strings.HasPrefix(rtkn.Str, "LIST") {
				n = WriteRegToScreen(startCol, regRow)
				if n > 8 {
					outputRow = regRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
					promptRow = startRow + 1   // used to be OutputRow -1
				}
				WriteDisplayTapeToScreen(displayCol, stackRow)
			} else if strings.HasPrefix(rtkn.Str, "NAME") || strings.HasPrefix(rtkn.Str, "LABL") { // prefix must be 4 chars.
				//var ans string
				var i int // remember that this auto-zero'd
				if len(rtkn.Str) > 4 {
					ch := rtkn.Str[4] // the 5th position
					i = GetRegIdx(ch)
				}
				Storage[i].Name = GetNameStr()
			} else if strings.HasPrefix(rtkn.Str, "SIG") || strings.HasPrefix(rtkn.Str, "FIX") { // SigFigN command, or FIX
				ch := rtkn.Str[len(rtkn.Str)-1] // ie, the last character.
				sigfig = GetRegIdx(ch)
				if sigfig > 9 { // If sigfig greater than this max value, make it -1 again.
					sigfig = -1
				}
				_, _ = hpcalc.GetResult(rtkn.Str) // Have to send this into hpcalc also
				outputmode = outputfix
			} else if rtkn.Str == "HELP" || rtkn.Str == "?" || rtkn.Str == "H" {
				WriteHelp(startCol+2, startRow)
			} else if strings.HasPrefix(rtkn.Str, "DUMP") {
				// do nothing, ie, don't send it into hpcalc.GetResult

			} else if strings.HasPrefix(rtkn.Str, "OUTPUTFI") { // allow outputfix, etc
				outputmode = outputfix
			} else if strings.HasPrefix(rtkn.Str, "OUTPUTFL") { // allow outputfloat, etc
				outputmode = outputfloat
			} else if strings.HasPrefix(rtkn.Str, "OUTPUTGE") { // allow outputgen, etc.
				outputmode = outputgen
			} else if rtkn.Str == "CLEAR" || rtkn.Str == "CLS" {
				scrn.Clear()
				scrn.Fill(' ', plain)
				scrn.Sync()
				gblrow = 0
				RepaintScreen(startCol)
			} else if rtkn.Str == "REPAINT" {
				RepaintScreen(startCol)
			} else if rtkn.Str == "DEBUG" {
				style = Cyan
				putf(x, outputRow+8, " HP-type RPN calculator written in Go.  Last altered %s", LastAltered)
				//			Printf_tb(0, OutputRow+9, BrightCyan, Black, "%s was last linked on %s.  Full executable is %s.", ExecFI.Name(), LastLinkedTimeStamp, execname)
				putf(0, outputRow+9, "%s was last linked on %s.  Full executable is %s.", ExecFI.Name(), LastLinkedTimeStamp, execName)

				style = Yellow
				putf(startCol, outputRow+10, " StartCol=%d,StartRow=%d,MaxCol=%d,MaxRow=%d,TitleRow=%d,StackRow=%d,RegRow=%d,OutputRow=%d,PromptRow=%d",
					startCol, startRow, maxCol, maxRow, titleRow, stackRow, regRow, outputRow, promptRow)
				putf(startCol, outputRow+11, " DisplayCol=%d", displayCol)
				putf(x, outputRow+13, " StorageFullFilename 1:%s, 2:%s, 3:%s", StorageFullFilename, Storage2FullFilename, Storage3FullFilename)
				style = Cyan
			} else if strings.HasPrefix(rtkn.Str, ":W") || strings.HasPrefix(rtkn.Str, "WR") {
				xstring := GetXstring()
				XStringFile, err := os.OpenFile(TextFilenameOut, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
				if err != nil {
					style = Yellow
					putf(0, outputRow, " Error %v while opening %s", err, TextFilenameOut)
					style = Cyan
				}
				defer XStringFile.Close()
				XstringWriter := bufio.NewWriter(XStringFile)
				defer XstringWriter.Flush()
				today := time.Now()
				datestring := today.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.
				_, err = XstringWriter.WriteString("------------------------------------------------------\n")
				check(err)
				_, err = XstringWriter.WriteString(datestring)
				_, err = XstringWriter.WriteRune('\n')
				_, err = XstringWriter.WriteString(xstring)
				_, err = XstringWriter.WriteRune('\n')
				check(err)

				_, err = XstringWriter.WriteString("\n\n")
				check(err)
				XstringWriter.Flush()
				XStringFile.Close()
			} else if strings.HasPrefix(rtkn.Str, ":R") || rtkn.Str == "READ" || rtkn.Str == "RD" {
				XstringFileExists := true
				XstringFile, err := os.Open(TextFilenameIn) // open for reading
				if os.IsNotExist(err) {
					style = Yellow
					putf(0, outputRow, "\n %s does not exist for reading in this directory.  Command ignored.\n", TextFilenameIn)
					style = Cyan
					XstringFileExists = false
				} else if err != nil {
					style = BoldYellow
					putf(0, outputRow, "\n %s does not exist for reading in this directory.  Command ignored.\n", TextFilenameIn)
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
				//                                                _, stringslice = hpcalc.GetResult(INBUF) // Here is where GetResult is called
				//                                                The old GETRESULT is long gone.
				_, stringslice = hpcalc.Result(rtkn) // Now need to call Result which accepts a token instead of a string.
				// ----------------------------------------------------------------------------------------------
				y := outputRow
				for _, s := range stringslice {
					puts(scrn, Yellow, x, y, s)
					y++
				}

				for y < maxRow {
					clearline(y)
					y++
				}
			}

			//  These commands are processed after GetResult is called, so these commands are run thru hpcalc2.
			if strings.ToLower(rtkn.Str) == "about" { // I'm using ToLower here just to experiment a little.
				style = Yellow
				putf(x, outputRow+1, " Last altered rpntcell %s, last linked %s. ", LastAltered, LastLinkedTimeStamp)
				style = Cyan
			}

		} // End Main Processing For Loop where I range over the tokens in INBUF
		RepaintScreen(startCol) // I forgot 4/07/2020 3:02:52 PM how the screen keeps getting cleared.  I finally found how, and then I changed it.
		x = startCol
		puts(scrn, Cyan, x, promptRow, InputPrompt)
		x += len(InputPrompt) + 2
		scrn.ShowCursor(x, promptRow)
		scrn.Show()
		ans := GetInputString(x, promptRow)
		INBUF = strings.ToUpper(ans)
		if len(INBUF) == 0 || strings.HasPrefix(INBUF, "Q") || INBUF == "X" {
			fmt.Println()
			break
		}

	} // end for len(INBUF) > 0

	// Time to write files before exiting.

	if !*noFileFlag {
		// Rotate StorageFileNames and write
		err = os.Rename(Storage2FullFilename, Storage3FullFilename)
		if err != nil && !*noFileFlag {
			fmt.Printf(" Rename of storage 2 to storage 3 failed with error %v \n", err)
		}

		err = os.Rename(StorageFullFilename, Storage2FullFilename)
		if err != nil && !*noFileFlag {
			fmt.Printf(" Rename of storage 1 to storage 2 failed with error %v \n", err)
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
	check(err)
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

	// hpcalc.mapWriteAndClose()  This is now handled by the defer above, as of 6/17/21.
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
	for i, s := range DisplayTape {
		idx := len(DisplayTape) - 1 - i
		if idx < 0 {
			idx = 0
		}
		str := fmt.Sprintf("%d  %s", idx, s)
		puts(scrn, Green, x, y, str)
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
		fmt.Printf("%s %s \n", msg, err)
		panic(err)
	}
}

// --------------------------------------------------- Cap -----------------------------------------
/*
func Cap(c rune) rune {  // not used, so I'll take it out.
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap
*/

// --------------------------------------------------- GetInputString for tcell--------------------------------------

func GetInputString(x, y int) string {
	deleol(x, y)
	scrn.ShowCursor(x, y)
	scrn.Show()
	donechan := make(chan bool, 1)
	keychannl := make(chan rune, 1)
	helpchan := make(chan bool, 1)
	delchan := make(chan bool, 1)
	upchan := make(chan bool, 1)
	downchan := make(chan bool, 1)
	pgupchan := make(chan bool, 1)
	pgdnchan := make(chan bool, 1)
	homechan := make(chan bool, 1)
	endchan := make(chan bool, 1)
	leftchan := make(chan bool, 1)
	rightchan := make(chan bool, 1)

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

				case tcell.KeyUp:
					upchan <- true
					return

				case tcell.KeyDown:
					downchan <- true
					return

				case tcell.KeyPgUp:
					pgupchan <- true
					return

				case tcell.KeyPgDn:
					pgdnchan <- true
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
			if len(bs) > 0 {
				return string(bs) + " help"
			}
			return "help"

		case <-delchan:
			if len(bs) > 0 {
				bs = bs[:len(bs)-1]
			}
			puts(scrn, style, x+len(bs), y, " ")
			scrn.ShowCursor(x+len(bs), y)
			scrn.Show()

		case <-upchan:
			if len(bs) > 0 {
				return string(bs) + " up"
			}
			return "up"

		case <-downchan:
			if len(bs) > 0 {
				return string(bs) + " dn"
			}
			return "dn"

		case <-pgupchan:
			if len(bs) > 0 {
				return string(bs) + " pgup"
			}
			return "pgup"

		case <-pgdnchan:
			if len(bs) > 0 {
				return string(bs) + " pgdn"
			}
			return "pgdn"

		case <-homechan:
			if len(bs) > 0 {
				return string(bs) + " pgup"
			}
			return "pgup" // "home key"

		case <-endchan:
			if len(bs) > 0 {
				return string(bs) + " pgdn"
			}
			return "pgdn" //"end key"

		case <-rightchan:
			if len(bs) > 0 {
				return string(bs) + " ~"
			}
			return "~"

		case <-leftchan:
			if len(bs) > 0 {
				return string(bs) + " ~"
			}
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
			} else if key == '`' {
				key = '~'
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
	helpstringslice = append(helpstringslice, "")
	helpstringslice = append(helpstringslice, "")
	helpstringslice = append(helpstringslice, " STOn,RCLn  -- store/recall the X register to/from the register indicated by n.")
	helpstringslice = append(helpstringslice, " NAME -- NAME registers with strings, Use - for spaces in these strings.")
	helpstringslice = append(helpstringslice, " Clear, CLS -- clear screen.")
	helpstringslice = append(helpstringslice, " EXIT,(Q)uit -- Needed after switch to use ScanWords in bufio scanner.")
	helpstringslice = append(helpstringslice, fmt.Sprintf(" :w, wr -- write X register to text file %s.", TextFilenameOut))
	helpstringslice = append(helpstringslice, fmt.Sprintf(" :r, rd, read -- read X register from first line of %s.", TextFilenameIn))
	helpstringslice = append(helpstringslice, " Debug -- Print debugging message to screen.")
	helpstringslice = append(helpstringslice, " SigN, FixN -- set significant figures for displayed numbers to N.  Default is -1.")
	helpstringslice = append(helpstringslice, " outputfix, outputfloat, outputgen -- sets output mode for displayed numbers.")
	helpstringslice = append(helpstringslice, " !N -- Recall history, hx.  N=0 means last command.")
	helpstringslice = append(helpstringslice, " up -- execute command !1.")
	helpstringslice = append(helpstringslice, " dn, pgdn, pgup, home, end -- manipulate the stack.")
	// helpstringslice = append(helpstringslice, " toclip, fromclip -- access to clip is by shelling out to xclip or tcc.") moved to hpcalc2 Nov 2020

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

	if y+len(helpstringslice) > maxRow {
		FI, err := os.Stat(HelpFileName)
		check(err)
		//		Printf_tb(x, y, BrightYellow, Black, " Too many help lines for this small screen.  See %s.", HelpFileName)
		style = BoldGreen
		y = outputRow
		putf(0, y, "Too many help lines for this small screen.  %d lines exceed %d rows.  See %s.", len(helpstringslice), maxRow, HelpFileName)
		y++
		yr, m, d := FI.ModTime().Date()
		putf(0, y, "%s from %d/%d/%d is in current directory.", FI.Name(), m, d, yr)
		y++
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
	scrn.Sync()

	RepaintScreen(x)
} // end WriteHelp

// ------------------------------------------------------- Repaint ----------------------------------

func RepaintScreen(x int) {

	WriteStack(x, stackRow)
	n = WriteRegToScreen(x, regRow)
	if n > 8 {
		outputRow = regRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
		promptRow = startRow + 1   // PromptRow = OutputRow - 1 was prev assignment.
	}
	WriteDisplayTapeToScreen(displayCol, stackRow)
	//	Printf_tb(x, MaxRow-1, BrightCyan, Black, Divider)  Not needed for tcell
	gblrow = 0
	scrn.Sync()
}

// -------------------------------------------------- GetNameStr --------------------------------

func GetNameStr() string {
	var ans string
	promptstr := "   Input name string, making - or = into a space : "
	puts(scrn, Yellow, 1, promptRow, promptstr)
	ans = GetInputString(len(promptstr)+2, promptRow)
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

// ------------------------------------------------ GetHx --------------------------------------

func GetHx(posn int) string {
	// DisplayTape is global
	var str string

	if len(DisplayTape) < posn {
		return " " // need a space else the pgm will exit
	}

	// posn 0 is last entry, not first.  If it is to be the first entry then I have to allow multidigit int, not just 1 char.
	i := len(DisplayTape) - 1 - posn
	if i < 0 {
		return ""
	}
	str = DisplayTape[i]
	return str
} // GetHx

// ---------------------------------------------------- End rpnt.go ------------------------------
