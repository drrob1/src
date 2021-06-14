// (C) 1990-2021.  Robert W.  Solomon.  All rights reserved.
// rpng.go
package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"src/tknptr"
	"strconv"
	"strings"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	hpcalc "src/hpcalc2"
	"src/makesubst"
)

/*
ct go-colortext
const (None, Black, Red, Green, Yellow, Blue, Magenta, Cyan, White)
var Writer io.Writer = os.Stdout  uses regular fmt.Fprintln(writer," whatever ");
type Color int
func Background(cl Color, bright bool)
func ChangeColor(fg Color, fgBright bool, bg Color, bgBright bool)
func Foreground(cl Color, bright bool)
func ResetColor()

ctfmt go-colortext/fmt
   func Print(cl ct.Color, bright bool, a ...interface{}) (n int, err error)  n = number of bytes written
   func Printf(cl ct.Color, bright bool, format string, a ...interface{}) (n int, err error)
   func Printfln(cl ct.Color, bright bool, format string, a ...interface{}) (n int, err error)
   func Println(cl ct.Color, bright bool, a ...interface{}) (n int, err error)
*/

const lastAlteredDate = "14 Jun 2021"

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
 8 Sep 16 -- Finished coding started 26 Aug 16 as rpng.go, adding functionality from hppanel, ie, persistant storage, a display tape and operator substitutions = or + and ; for *.
11 Sep 16 -- Changed stack and storage files to use gob package and only use one Storage file.
 1 Oct 16 -- Made the stack display when the program starts.
 4 Oct 16 -- Made the storage registers display when the program starts.
 7 Oct 16 -- Changed default dump to DUMPFIXED.   Input is now using splitfunc by space delimited words instead of whole lines.
      				Conversion to ScanWords means I cannot get an empty string back unless ^C or ^D.  So needed (Q)uit, EXIT, STOP.
 8 Oct 16 -- Updated the prompt to say (Q)uit to exit instead of Enter to exit, and sto command calls WriteRegToScreen()
 9 Oct 16 -- Added ability to enter hex numbers like in C, or GoLang, etc, using the "0x" prefix.
21 Oct 16 -- Decided that sto command should not also dump stack, as the stack does not change.
28 Oct 16 -- Will clear screen in between command calls.  This is before I play with termbox or goterm
31 Oct 16 -- Since clear screen works, I'll have the stack and reg's displayed more often.
      				Turns out that this cleared output like PRIME or HEX.  Now have YES,NO,ON,OFF for manual
      				control of DUMP output.  And added a HELP line printed here after that from hpcalc.
 3 Nov 16 -- Changed hpcalc to return strings for output by the client rtn, instead of it doing its own output.  So now have
      				to make the corresponding changes here.
 4 Nov 16 -- Added the RepaintScreen rtn, and changed how DOW is done.
 5 Nov 16 -- Added SuppressDump map code
 6 Nov 16 -- Added blank lines before writing output from hpcalc.
11 Dec 16 -- Fixed bug in GetRegIdx when a char is passed in that is not 0..9, A..Z
13 Jan 17 -- Removed stop as an exit command, as it meant that reg p could not be stored into.
23 Feb 17 -- Made "?" equivalent to "help"
17 Jul 17 -- Discovered bug in that command history is not displayed correctly here.  It is displayed
      				correctly in rpnterm.  It will take me a little longer to find and fix this bug.
31 Aug 17 -- Stopped working on the command hx issue.  Instead of bufio, I will use fmt.Scan().
   			And will check timestamp of the rpng exec file.
 6 Apr 18 -- Wrote out DisplayTapeFile.
22 Aug 18 -- learning about code folding
 2 Oct 18 -- Changed fmt.Scan to fmt.Scanln so now empty input will exit again.  It works, but I had to convert error to string.
 7 Oct 18 -- A bug doesn't allow entering several numbers on a line.  Looks like buffered input is needed, afterall.
 1 Jan 19 -- There is a bug in entering several space delimited terms.  So I removed a trimspace and I'll see what happens next.
26 Mar 20 -- I'm attempting to get output to the clipboard.  On linux using xsel or xclip utilities.
29 Mar 20 -- Need to removed commas from a number string received from the clip.
30 Mar 20 -- Will make it use tcc 22 as I have that at work also.
 1 Apr 20 -- Removed all spaces from the string returned from xclip.  The conversion fails if extraneous spaces are there.
 8 Aug 20 -- Now uses hpcalc2, which seems somewhat faster.  But only somewhat.
23 Oct 20 -- Adding flag package to allow the -n flag, meaning no files read or written.
 8 Nov 20 -- Found minor error in fmt messages of fromclip command.  I was looking because of "Go Standard Library Cookbook"
 9 Nov 20 -- Added use of comspec, copied from hpcalc2.go.
11 Dec 20 -- Went back to delimiting by spaces, not <CR>.  And, as of Nov 2020, toclip, fromclip are implemented in hpcalc2
13 Dec 20 -- Compiled w/ map registers now included in hpcalc2.
30 Jan 21 -- Starting coding the colorization of the output, using "github.com/daviddengcn/go-colortext" and its documentation at
   	            https://godoc.org/github.com/daviddengcn/go-colortext.
31 Jan 21 -- Wrote hpcalc2.SigFig for the conversion routine here.  And color for Windows will be bold.
 4 Feb 21 -- Will display stack before displaying any returned strings from hpcalc2.  And fixed bug of ignoring a command line param.
 5 Feb 21 -- Removed an extra PushMatrixStacks() while initializing everything.
11 Feb 21 -- Added X for exit, to copy PACS.  And changed the prompt
 8 Apr 21 -- Converted to module name src, that happens to reside at ~/go/src.  Go figure!
12 Jun 21 -- Now that I have RealTokenSlice in tknptr, I'll use it to allow more flexibility when entering commands.  And removed tokenize.CAP().
14 Jun 21 -- Testing new routine in hpcalc2, called Result that takes a token as a param instead of a string.
*/

var Storage [36]float64 // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var clear map[string]func()
var SuppressDump map[string]bool
var WindowsFlag bool

func main() {
	var INBUF, HomeDir string

	const Storage1FileName = "RPNStorage.gob" // Allows for a rotation of Storage files, in case of a mistake.
	const Storage2FileName = "RPNStorage2.gob"
	const Storage3FileName = "RPNStorage3.gob"
	const DisplayTapeFilename = "displaytape.txt"

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.
	var err error

	var nofileflag = flag.Bool("n", false, "no files read or written.") // pointer
	flag.Parse()

	SuppressDump = make(map[string]bool)
	SuppressDump["PRIME"] = true
	SuppressDump["HEX"] = true
	SuppressDump["DOW"] = true
	SuppressDump["HOL"] = true
	SuppressDump["ABOUT"] = true
	SuppressDump["HELP"] = true
	SuppressDump["TOCLIP"] = true

	P := fmt.Println // a cute trick I just learned, from gobyexample.com.
	ClearScreen()

	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
		WindowsFlag = true
	} else { // then HomeDir will be empty.
		fmt.Println(" runtime.GOOS does not say linux or windows.  Is this a Mac?")
	}
	fmt.Println()
	ctfmt.Println(ct.Blue, WindowsFlag, " GOOS =", runtime.GOOS, ".  HomeDir =", HomeDir, ".  ARCH=", runtime.GOARCH)
	fmt.Println()

	AllowDumpFlag := false // I need to be able to supress the automatic DUMP under certain circumstances.
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

	var thefile *os.File
	if !*nofileflag {
		thefile, err = os.Open(StorageFullFilename) // open for reading
		if err != nil {
			fmt.Printf(" Error from os.Open(Storage1FileName).  Possibly because no Stack File found: %v\n", err)
			theFileExists = false
		}
		defer thefile.Close()
		if theFileExists {
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

	// hpcalc.PushMatrixStacks()  I think this is an extra one and is not needed.

	ctfmt.Println(ct.Yellow, WindowsFlag, " HP-type RPN calculator written in Go.  Code was last altered ", lastAlteredDate,
		", NoFileFlag is ", *nofileflag)
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	ctfmt.Println(ct.Cyan, WindowsFlag, ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fmt.Println()
	_, stringslice = hpcalc.GetResult("DUMPFIXED")
	for _, ss := range stringslice {
		ctfmt.Println(ct.Yellow, WindowsFlag, ss)
	}
	WriteRegToScreen()
	fmt.Println()
	fmt.Println()

	args := flag.Args()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords) // as of 12/11/20, back to word by word, not line by line.

	if len(args) > 0 { // fixed bug here 2/4/21.  Used to be > 1 which is not correct.
		// INBUF = getcommandline.GetCommandLineString()  No longer works now that I'm using the flag package.
		INBUF = strings.Join(args, " ")
	} else {
		fmt.Print(" Enter calculation, Help,  Quit, or eXit: ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			// _, err := fmt.Scan(&INBUF)  // removed 10/2/18.
			// _, err := fmt.Scanln(&INBUF) // added 10/02/2018 04:52:23 PM, and removed again on 10/07/2018 09:34:54 AM
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(INBUF) == 0 { // it seems that this can never be empty when using fmt.Scan(), but it can with fmt.Scanln().
			os.Exit(0)
		}
	} // if command tail exists

	ManualDump := true

	hpcalc.PushMatrixStacks()

	for len(INBUF) > 0 { // Main processing loop
		INBUF = makesubst.MakeSubst(INBUF)
		INBUF = strings.ToUpper(INBUF)
		DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
		// These commands are not run thru hpcalc as they are processed before calling GetResult.
		realtknslice := tknptr.RealTokenSlice(INBUF)

		for _, rtkn := range realtknslice {
			if rtkn.Str == "ZEROREG" {
				for c := range Storage {
					Storage[c] = 0.0
				}
				AllowDumpFlag = false
			} else if strings.HasPrefix(rtkn.Str, "STO") {
				i := 0
				if len(rtkn.Str) > 3 {
					ch := rtkn.Str[3] // The 4th position.
					i = GetRegIdx(ch)
				}
				Storage[i] = hpcalc.READX()
				P()
				WriteRegToScreen()
				P()
				AllowDumpFlag = false
			} else if strings.HasPrefix(rtkn.Str, "RCL") {
				i := 0
				if len(rtkn.Str) > 3 {
					ch := rtkn.Str[3] // the 4th position.
					i = GetRegIdx(ch)
				}
				hpcalc.PUSHX(Storage[i])
			} else if strings.HasPrefix(rtkn.Str, "SHO") { // so it will match SHOW and SHO
				fmt.Println()
				WriteRegToScreen()
				fmt.Println()
				WriteDisplayTapeToScreen()
				AllowDumpFlag = false
			} else if rtkn.Str == "NO" || rtkn.Str == "OFF" {
				ManualDump = false
			} else if rtkn.Str == "YES" || rtkn.Str == "ON" {
				ManualDump = true
			} else {
				// -------------------------------------------------------------------------------------
				_, stringslice = hpcalc.Result(rtkn) //   Here is where GetResult is called -> Result
				// -------------------------------------------------------------------------------------
				ClearScreen()   // added 02/04/2021 9:07:12 AM to always update the stack, before displaying any returned strings from GetResult.
				RepaintScreen() // So I don't think I need this complex system to allow or SuppressDump.  I'll keep it for a while but turn it off.
				//	AllowDumpFlag = true
				AllowDumpFlag = false
				fmt.Println()
				if len(stringslice) > 0 {
					for _, ss := range stringslice {
						ctfmt.Println(ct.Cyan, WindowsFlag, ss)
					}
					AllowDumpFlag = false // if there is anything in the stringslice to output, suppress dump.
				} else if SuppressDump[rtkn.Str] { // this may be redundant to the len(stringslice) conditional
					AllowDumpFlag = false
				}
			}
			// -------------------------------------------------------------------------------------

			//  These commands are processed thru GetResult() first, then these are processed here.
			if strings.ToLower(rtkn.Str) == "about" { // I'm using ToLower here just to experiment a little.
				ctfmt.Println(ct.Cyan, WindowsFlag, " Last altered the source of rpng.go", lastAlteredDate)
				AllowDumpFlag = false
			} else if strings.HasPrefix(rtkn.Str, "DUMP") {
				AllowDumpFlag = false
			} else if rtkn.Str == "HELP" || rtkn.Str == "?" || rtkn.Str == "H" { // have more help lines to print
				fmt.Println(" NO,OFF,YES,ON -- Manually change a Dump flag for commands like PRIME and HEX.")
				AllowDumpFlag = false
			} else if rtkn.Str == "HOL" {
				AllowDumpFlag = false
			} else if rtkn.Str == "DOW" {
				AllowDumpFlag = false
			}
			// AllowDumpFlag = false // added 02/04/2021 1:02:58 PM to always turn this off, and removed 6/12/21
			if AllowDumpFlag && ManualDump {
				ClearScreen()
				RepaintScreen()
			}
		} // end of for _, rtkn := range realtokenslice
		fmt.Println()
		fmt.Print(" Enter calculation, Help, Quit or eXit: ")
		//{{{
		//_, err := fmt.Scan(&INBUF)  removed 10/2/18
		//_, err := fmt.Scanln(&INBUF) // added 10/02/2018 1:58:57 PM, and removed again 10/07/2018 09:37:51 AM
		// }}}
		scanner.Scan()
		INBUF = scanner.Text()
		INBUF = makesubst.MakeSubst(INBUF)
		INBUF = strings.ToUpper(INBUF)
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading 2nd buffered input:", err)
			os.Exit(1)
		}
		if len(INBUF) == 0 {
			break
		}
		INBUF = strings.ToUpper(INBUF)
		if strings.HasPrefix(INBUF, "Q") || INBUF == "EXIT" || INBUF == "X" {
			fmt.Println()
			break
		}
	} // end main for loop

	// Time to write files before exiting, if the flag says so.

	if !*nofileflag {
		// Rotate StorageFileNames and write
		err = os.Rename(Storage2FullFilename, Storage3FullFilename)
		if err != nil && !*nofileflag {
			fmt.Printf(" Rename of storage 2 to storage 3 failed with error %v \n", err)
		}

		err = os.Rename(StorageFullFilename, Storage2FullFilename)
		if err != nil && !*nofileflag {
			fmt.Printf(" Rename of storage 1 to storage 2 failed with error %v \n", err)
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

	hpcalc.MapClose()
} // main in rpng.go

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
func WriteRegToScreen() {
	FirstNonZeroStorageFlag := true

	for i, r := range Storage {
		if r != 0.0 {
			if FirstNonZeroStorageFlag {
				fmt.Println("                The following storage registers are not zero")
				FirstNonZeroStorageFlag = false
			} // if firstnonzerostorageflag
			ch := GetRegChar(i)
			sigfig := hpcalc.SigFig()
			s := strconv.FormatFloat(r, 'g', sigfig, 64)
			s = hpcalc.CropNStr(s)
			if r >= 10000 {
				s = hpcalc.AddCommas(s)
			}
			fmt.Println(" Reg [", ch, "] = ", s)
		} // if storage value is not zero
	} // for range over Storage
	if FirstNonZeroStorageFlag {
		fmt.Println(" All storage registers are zero.")
		fmt.Println()
	}
} // WriteRegToScreen

// --------------------------------------------------------- WriteDisplayTapeToScreen ----------------
func WriteDisplayTapeToScreen() {
	fmt.Println("        DisplayTape ")
	for _, s := range DisplayTape {
		ctfmt.Println(ct.Blue, WindowsFlag, s)
	} // for ranging over DisplayTape slice of strings
	fmt.Println()
} // WriteDisplayTapeToScreen

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------- init -----------------------------------
func init() {
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

// ---------------------------------------------------- ClearScreen ------------------------------------
func ClearScreen() {
	clearfunc, ok := clear[runtime.GOOS]
	if ok {
		clearfunc()
	} else { // unsupported platform
		panic(" The ClearScreen platform is only supported on linux or windows, at the moment")
	}
}

// ---------------------------------------------------- RepaintScreen ----------------------------------
func RepaintScreen() { // ExecFI, ExecTimeStamp and execname are not global.  I'll not print them here.
	ctfmt.Println(ct.Cyan, WindowsFlag, " HP-type RPN calculator written in Go.  Last altered source of rpng.go", lastAlteredDate)
	//	fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fmt.Println()
	_, stringslice := hpcalc.GetResult("DUMPFIXED")
	ctfmt.Println(ct.Yellow, WindowsFlag, "                                   ", stringslice[len(stringslice)-2])
	fmt.Println()
	for _, ss := range stringslice {
		ctfmt.Println(ct.Cyan, WindowsFlag, ss)
	}
	WriteRegToScreen()
	WriteDisplayTapeToScreen()
}

// ---------------------------------------------------- End rpng.go ------------------------------
