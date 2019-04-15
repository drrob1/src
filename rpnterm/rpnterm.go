// (C) 1990-2016.  Robert W.  Solomon.  All rights reserved.
// rpnterm.go
package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec" // for the clear screen functions.
	"runtime"
	"strconv"
	"strings"
	"time"
	"timlibg"

	//	"github.com/nsf/termbox-go"
	//
	"getcommandline"
	"hpcalc"
	"tokenize"

	"github.com/nsf/termbox-go"
)

const LastAltered = "15 Apr 2019"
const InputPrompt = " Enter calculation, HELP or (Q)uit to exit: "

type Register struct {
	Value float64
	Name  string
}

var Storage [36]Register // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var Divider string
var clear map[string]func()

var StartCol, StartRow, sigfig, MaxRow, MaxCol, TitleRow, StackRow, RegRow, OutputRow, DisplayCol, PromptRow, outputmode, n int
var BrightYellow, BrightCyan, Black termbox.Attribute

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
const TextFilenameOut = "rpntermoutput.txt"
const TextFilenameIn = "rpnterminput.txt"
const HelpFileName = "helprpn.txt"

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
*/

func main() {
	var INBUF, HomeDir string

	var x int

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.
	var err error

	ClearScreen() // ClearScreen before termbox.init fcn, to see if this helps.
	fmt.Println()
	//  fmt.Println("  I hope this helps.  But it likely won't.  ");  It didn't help.  I'm commenting it out, finally.
	fmt.Println()
	fmt.Println()

	termerr := termbox.Init()
	if termerr != nil {
		log.Println(" TermBox init failed.")
		panic(termerr)
	}
	defer termbox.Close()
	BrightYellow = termbox.ColorYellow | termbox.AttrBold
	BrightCyan = termbox.ColorCyan | termbox.AttrBold
	Black = termbox.ColorBlack
	//	err = termbox.Clear(BrightYellow, Black) \  removed 4/15/19
	//	check(err)                               /
	e := termbox.Clear(Black, Black)
	check(e)
	e = termbox.Flush()
	check(e)
	MaxCol, MaxRow = termbox.Size()
	HardClearScreen() // I think there is a bug in termbox.

	stringslice = make([]string, 0, 35)
	sigfig = -1 // now only applies to WriteRegToScreen
	StartRow := 0
	StartCol := 0
	outputmode = 0

	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
		//	StartRow = 2 // Could this be the source of my problems with screen updating correctly?  Nope.  commended out 4/15/19
		//	StartCol = 1 // So I'll make it the same as under Windows, and I'll see what's going on, maybe?  Nope.  Comended 4/15/19
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
		//	StartRow = 2
		//	StartCol = 1
	} else { // then HomeDir will be empty.
		Print_tb(StartCol, StartRow+20, BrightYellow, Black, " runtime.GOOS does not say linux or windows.  Is this a Mac?")
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

	Printf_tb(StartCol, MaxRow-1, BrightCyan, Black, Divider)
	e = termbox.Flush()
	check(e)

	DisplayTape = make([]string, 0, 100)
	theFileExists := true

	StorageFullFilenameSlice := make([]string, 5)
	StorageFullFilenameSlice[0] = HomeDir
	StorageFullFilenameSlice[1] = string(os.PathSeparator)
	StorageFullFilenameSlice[2] = Storage1FileName
	StorageFullFilename := strings.Join(StorageFullFilenameSlice, "")

	Storage2FullFilenameSlice := make([]string, 5)
	Storage2FullFilenameSlice[0] = HomeDir
	Storage2FullFilenameSlice[1] = string(os.PathSeparator)
	Storage2FullFilenameSlice[2] = Storage2FileName
	Storage2FullFilename := strings.Join(Storage2FullFilenameSlice, "")

	Storage3FullFilenameSlice := make([]string, 5)
	Storage3FullFilenameSlice[0] = HomeDir
	Storage3FullFilenameSlice[1] = string(os.PathSeparator)
	Storage3FullFilenameSlice[2] = Storage3FileName
	Storage3FullFilename := strings.Join(Storage3FullFilenameSlice, "")

	thefile, err := os.Open(StorageFullFilename) // open for reading
	if os.IsNotExist(err) {
		log.Print(" thefile does not exist for reading. ")
		theFileExists = false
	} else if err != nil {
		log.Printf(" Error from os.Open(Storage1FileName).  Possibly because no Stack File found: %v\n", err)
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
		PromptRow = StartRow + 1   // used to be OutputRow -1
	}

	//  Print_tb(x,PromptRow,BrightCyan,Black,InputPrompt);  Doesn't make any difference, it seems.
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
	} else {
		Print_tb(x, PromptRow, BrightCyan, Black, InputPrompt)
		x = x + len(InputPrompt) + 2
		termbox.SetCursor(x, PromptRow)
		INBUF = GetInputString(x, PromptRow)
		if strings.HasPrefix(INBUF, "Q") {
			os.Exit(0)
		}
		x = StartCol
	} // if command tail exists
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
			// {{{
			/*  This is now handled in hpcalc directly, and returned in the stringslice.
			    }else if INBUF == "DOW" {
			      r,_ := hpcalc.GetResult("DOW");
			      i := int(r);
			      Printf_tb(x,OutputRow,BrightYellow,Black," Day of Week for X value is a %s",timlibg.DayNames[i]);
			*/
			// }}}
		} else if strings.HasPrefix(INBUF, "STO") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // The 4th position.
				i = GetRegIdx(ch)
			}
			Storage[i].Value = hpcalc.READX()
			n = WriteRegToScreen(x, RegRow)
			if n > 8 {
				ClearLine(PromptRow)       // Should have done this a long time ago.
				ClearLine(OutputRow)       // This too.
				OutputRow = RegRow + n + 3 // So there is enough space for all the reg's to be displayed above the output
				PromptRow = StartRow + 1   // used to be OutputRow -1
			}
			// Now ask for NAME
			// var ans string
			// promptstr := "   Input name string : "
			// Print_tb(1, OutputRow, BrightYellow, Black, promptstr)
			// ans = GetInputString(len(promptstr)+2, OutputRow)
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
			// do nothing, but don't send it into hpcalc.GetResult
		} else if strings.HasPrefix(INBUF, "OUTPUTFIX") { // allow outputfix or outputfixed
			outputmode = outputfix
		} else if INBUF == "OUTPUTFLOAT" {
			outputmode = outputfloat
		} else if INBUF == "OUTPUTGEN" {
			outputmode = outputgen
		} else if INBUF == "CLEAR" || INBUF == "CLS" {
			HardClearScreen()
			//      err = termbox.Clear(BrightYellow,Black);
			//      check(err);
		} else if INBUF == "REPAINT" {
			RepaintScreen(StartCol)
		} else if INBUF == "DEBUG" {
			Printf_tb(x, OutputRow+8, BrightCyan, Black, " HP-type RPN calculator written in Go.  Last altered %s", LastAltered)

			Printf_tb(0, OutputRow+9, BrightCyan, Black, "%s was last linked on %s.  Full executable is %s.", ExecFI.Name(), LastLinkedTimeStamp, execname)

			Printf_tb(StartCol, OutputRow+10, BrightYellow, Black, " StartCol=%d,StartRow=%d,MaxCol=%d,MaxRow=%d,TitleRow=%d,StackRow=%d,RegRow=%d,OutputRow=%d,PromptRow=%d",
				StartCol, StartRow, MaxCol, MaxRow, TitleRow, StackRow, RegRow, OutputRow, PromptRow)
			Printf_tb(StartCol, OutputRow+11, BrightYellow, Black, " DisplayCol=%d", DisplayCol)
			Printf_tb(x, OutputRow+13, BrightYellow, Black, " StorageFullFilename 1:%s, 2:%s, 3:%s", StorageFullFilename, Storage2FullFilename, Storage3FullFilename)
		} else if strings.HasPrefix(INBUF, ":W") || strings.HasPrefix(INBUF, "WR") {
			xstring := GetXstring()
			XStringFile, err := os.OpenFile(TextFilenameOut, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				Printf_tb(0, OutputRow, BrightYellow, Black, " Error %v while opening %s", err, TextFilenameOut)
				os.Exit(1)
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
				Printf_tb(0, OutputRow, BrightYellow, Black, "\n %s does not exist for reading in this directory.  Command ignored.\n", TextFilenameIn)
				XstringFileExists = false
			} else if err != nil {
				Printf_tb(0, OutputRow, BrightYellow, Black, "\n %s does not exist for reading in this directory.  Command ignored.\n", TextFilenameIn)
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
				Print_tb(x, y, BrightYellow, Black, s)
				y++
			}
			for y < MaxRow {
				ClearLine(y)
				y++
			}
		}
		// Don't understand why this next line helps stabilize the output display, but it does.
		Print_tb(x, OutputRow+len(stringslice)+1, BrightCyan, Black, "-------------")

		//  These commands are processed after GetResult is called, so these commands are run thru hpcalc.
		if strings.ToLower(INBUF) == "about" { // I'm using ToLower here just to experiment a little.
			Printf_tb(x, OutputRow+1, BrightYellow, Black, " Last altered rpnterm %s, last linked %s. ", LastAltered, LastLinkedTimeStamp)
		} // if INBUF
		if INBUF != "CLEAR" || INBUF != "CLS" {
			RepaintScreen(StartCol)
		}
		x = StartCol
		Print_tb(x, PromptRow, BrightCyan, Black, InputPrompt)
		x += len(InputPrompt) + 2
		termbox.SetCursor(x, PromptRow)
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
		fmt.Errorf(" Rename of storage 2 to storage 3 failed with error %v \n", err)
	}

	err = os.Rename(StorageFullFilename, Storage2FullFilename)
	if err != nil {
		fmt.Errorf(" Rename of storage 1 to storage 2 failed with error %v \n", err)
	}

	thefile, err = os.Create(StorageFullFilename) // for writing
	check(err)                                    // This should not fail, so panic if it does fail.
	defer thefile.Close()

	Stk = hpcalc.GETSTACK()
	encoder := gob.NewEncoder(thefile) // encoder writes the file
	err = encoder.Encode(Stk)          // encoder writes the file
	check(err)                         // Panic if there is an error
	err = encoder.Encode(Storage)      // encoder writes the file
	check(err)

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

} // main in rpnterm.go

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
				Print_tb(x, y, BrightYellow, Black, "The following storage registers are not zero")
				y++
				FirstNonZeroStorageFlag = false
			} // if firstnonzerostorageflag
			ch := GetRegChar(i)
			s := strconv.FormatFloat(r.Value, 'g', sigfig, 64) // sigfig of -1 means max sigfig.
			s = hpcalc.CropNStr(s)
			if r.Value >= 10000 {
				s = hpcalc.AddCommas(s)
			}
			Printf_tb(x, y, BrightCyan, Black, " Reg [%s], %s =  %s", ch, r.Name, s)
			y++
			n++
		} // if storage value is not zero
	} // for range over Storage
	if FirstNonZeroStorageFlag {
		Print_tb(x, y, BrightYellow, Black, " All storage registers are zero.")
		y++
	}
	return n
} // WriteRegToScreen

// --------------------------------------------------------- WriteDisplayTapeToScreen ----------------
func WriteDisplayTapeToScreen(x, y int) {
	Print_tb(x, y, BrightCyan, Black, "DisplayTape")
	y++
	for _, s := range DisplayTape {
		Print_tb(x, y, BrightYellow, Black, s)
		y++
	} // for ranging over DisplayTape slice of strings
} // WriteDisplayTapeToScreen

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------- init -----------------------------------
func init() { // start termbox in the init code doesn't work.  Don't know why.  But this init does work.
	clear = make(map[string]func())
	clear["linux"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	clear["windows"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// --------------------------------------------------- Cap -----------------------------------------
func Cap(c rune) rune {
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap

// --------------------------------------------------- Print_tb -----------------------------------
func Print_tb(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
	ClearEOL(x, y)
	e := termbox.Flush()
	if e != nil {
		panic(e)
	}
}

//----------------------------------------------------- Printf_tb ---------------------------------
func Printf_tb(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	Print_tb(x, y, fg, bg, s)
}

// --------------------------------------------------- GetInputString --------------------------------------

func GetInputString(x, y int) string {
	bs := make([]byte, 0, 100) // byteslice to build up the string to be returned.
	termbox.SetCursor(x, y)

MainEventLoop:
	for {
		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventKey:
			ch := event.Ch
			key := event.Key
			if key == termbox.KeySpace {
				ch = ' '
				if len(bs) > 0 { // ignore spaces if there is no string yet
					break MainEventLoop
				}
			} else if ch == 0 { // need to process backspace and del keys
				if key == termbox.KeyEnter {
					break MainEventLoop
				} else if key == termbox.KeyF1 || key == termbox.KeyF2 {
					bs = append(bs, "HELP"...)
					break MainEventLoop
				} else if key == termbox.KeyPgup || key == termbox.KeyArrowUp {
					bs = append(bs, "UP"...) // Code in C++ returned ',' here
					break MainEventLoop
				} else if key == termbox.KeyPgdn || key == termbox.KeyArrowDown {
					bs = append(bs, "DN"...) // Code in C++ returned '!' here
					break MainEventLoop
				} else if key == termbox.KeyArrowRight || key == termbox.KeyArrowLeft {
					bs = append(bs, '~') // Could return '<' or '>' or '<>' or '><' also
					break MainEventLoop
				} else if key == termbox.KeyEsc {
					bs = append(bs, 'Q')
					break MainEventLoop

					// this test must be last because all special keys above meet condition of key > '~'
					// except on Windows, where <backspace> returns 8, which is std ASCII.  Seems that linux doesn't.
				} else if (len(bs) > 0) && (key == termbox.KeyDelete || key > '~' || key == 8) {
					x--
					bs = bs[:len(bs)-1]
				}
			} else if ch == '=' {
				ch = '+'
			} else if ch == ';' {
				ch = '*'
			}
			termbox.SetCell(x, y, ch, BrightYellow, Black)
			if ch > 0 {
				x++
				bs = append(bs, byte(ch))
			}
			termbox.SetCursor(x, y)
			err := termbox.Flush()
			check(err)
		case termbox.EventResize:
			err := termbox.Sync()
			check(err)
			err = termbox.Flush()
			check(err)
		case termbox.EventError:
			panic(event.Err)
		case termbox.EventMouse:
		case termbox.EventInterrupt:
		case termbox.EventRaw:
		case termbox.EventNone:

		} // end switch-case on the Main Event  (Pun intended)

	} // MainEventLoop for ever

	return string(bs)
} // end GetInputString

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

	Print_tb(x+10, y, BrightYellow, Black, stringslice[len(stringslice)-2])
	y++
	for _, s := range stringslice {
		Print_tb(x, y, BrightCyan, Black, s)
		y++
	}
	// {{{
	/*  Old way was to GetStack and output.  Now use hpcalc to generate the output as a string.
	    var SRN int;  // IIRC, SRN = stack register name OR stack register number
	    var str string;
	    stk := hpcalc.GETSTACK();  // this is of type hpcalc.StackType which is [StackSize]float64 where StackSize=8
	    for SRN=hpcalc.T1; SRN >= hpcalc.X; SRN-- {
	      str = strconv.FormatFloat(stk[SRN],'g',sigfig,64); // sigfig of -1 means maximum precision.
	      str = hpcalc.CropNStr(str);
	      if stk[SRN] > 10000 {
	        str = hpcalc.AddCommas(str);
	      }
	      Printf_tb(x,y,BrightYellow,Black,"%2s:  %10.4g %s %s\n",hpcalc.StackRegNamesString[SRN],stk[SRN],
	                                                                                                SpaceFiller,str);
	      y++
	    }
	*/
	// }}}
} // end WriteStack

//--------------------------------------------- WriteHelp -------------------------------------------
func WriteHelp(x, y int) { // essentially moved to hpcalc module quite a while ago, but I didn't log when.
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

	FI, err := os.Stat(HelpFileName)
	if err != nil {
		// Will open this file in the current working directory instead of the HomeDir.
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
	}

	if y+len(helpstringslice) >= MaxRow {
		Printf_tb(x, y, BrightYellow, Black, " Too many help lines for this small screen.  See %s.", HelpFileName)
		yr, m, d := FI.ModTime().Date()
		Printf_tb(x, y, BrightYellow, Black, "%s from %d/%d/%d is in current directory.", FI.Name(), m, d, yr)
		return
	}
	err = termbox.Clear(BrightYellow, Black)
	check(err)
	P := Print_tb
	// Pf := Printf_tb  not needed now that the helpstringslice has been extended.

	y = 1
	for _, s := range helpstringslice {
		P(x, y, BrightYellow, Black, s)
		y++
	}

	P(x, y, BrightCyan, Black, " pausing ")
	termbox.SetCursor(x+11, y)
	_ = GetInputString(x+11, y)
	y++

	err = termbox.Clear(BrightYellow, Black)
	check(err)
	RepaintScreen(x)
} // end WriteHelp

// ----------------------------------------------------- ClearLine -----------------------------------
func ClearLine(y int) {
	if y > MaxRow {
		y = MaxRow
	}
	for x := StartCol; x <= MaxCol; x++ {
		termbox.SetCell(x, y, 0, Black, Black) // Don't know if it matters if the char is ' ' or nil.
	}
	err := termbox.Flush()
	check(err)
} // end ClearLine

// ----------------------------------------------------- HardClearScreen -----------------------------
func HardClearScreen() {
	err := termbox.Clear(Black, Black)
	check(err)
	for row := StartRow; row <= MaxRow; row++ {
		ClearLine(row)
	}
	err = termbox.Flush()
	check(err)
}

// ------------------------------------------------------ ClearEOL -----------------------------------
func ClearEOL(x, y int) {
	if y > MaxRow {
		y = MaxRow
	}
	if x > MaxCol {
		return
	}
	for i := x; i <= MaxCol; i++ {
		termbox.SetCell(i, y, 0, Black, Black) // Don't know if it matters if the char is ' ' or nil.
	}
	err := termbox.Flush()
	check(err)
}

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
	Printf_tb(x, MaxRow-1, BrightCyan, Black, Divider)
}

// ---------------------------------------------------- ClearScreen ------------------------------------
func ClearScreen() {
	clearfunc, ok := clear[runtime.GOOS]
	if ok {
		clearfunc()
	} else { // unsupported platform
		panic(" The ClearScreen platform is only supported on linux or windows, at the moment")
	}
}

// -------------------------------------------------- GetNameStr --------------------------------
func GetNameStr() string {
	var ans string
	promptstr := "   Input name string : "
	Print_tb(1, OutputRow, BrightYellow, Black, promptstr)
	ans = GetInputString(len(promptstr)+2, OutputRow)
	if strings.ToUpper(ans) == "TODAY" {
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

// ---------------------------------------------------- End rpnterm.go ------------------------------
