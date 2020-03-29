// (C) 1990-2016.  Robert W.  Solomon.  All rights reserved.
// rpng.go
package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	//
	"getcommandline"
	"hpcalc"
	"makesubst"
	"tokenize"
)

const LastAlteredDate = "29 Mar 2020"

var Storage [36]float64 // 0 ..  9, a ..  z
var DisplayTape, stringslice []string
var clear map[string]func()
var clippy map[string]func(s string)
var SuppressDump map[string]bool

/*
 runtime.GOOS returns either linux or windows.  I have not tested mac.
 I want either $HOME or %userprofile to set the write dir.
*/

func main() {
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
	*/

	//  var Y,NYD,July4,VetD,ChristmasD int;     //  For Holiday cmd
	var INBUF, HomeDir string

	const Storage1FileName = "RPNStorage.gob" // Allows for a rotation of Storage files, in case of a mistake.
	const Storage2FileName = "RPNStorage2.gob"
	const Storage3FileName = "RPNStorage3.gob"
	const DisplayTapeFilename = "displaytape.txt"
	//  var Holidays holidaycalc.HolType;   not needed after hpcalc does this directly thru stringslice.

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.
	var err error
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
	} else { // then HomeDir will be empty.
		fmt.Println(" runtime.GOOS does not say linux or windows.  Is this a Mac?")
	}
	fmt.Println()
	fmt.Println(" GOOS =", runtime.GOOS, ".  HomeDir =", HomeDir, ".  ARCH=", runtime.GOARCH)
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
	// {{{
	//	fmt.Println(" StorageFullFilename (1,2,3):", StorageFullFilename, Storage2FullFilename,
	//	Storage3FullFilename)  This is not needed anymore.
	//	fmt.Println()
	// }}}

	thefile, err := os.Open(StorageFullFilename) // open for reading
	if err != nil {
		fmt.Errorf(" Error from os.Open(Storage1FileName).  Possibly because no Stack File found: %v\n", err)
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

	hpcalc.PushMatrixStacks()

	fmt.Println(" HP-type RPN calculator written in Go.  Code was last altered ", LastAlteredDate)
	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fmt.Println()
	_, stringslice = hpcalc.GetResult("DUMPFIXED")
	for _, ss := range stringslice {
		fmt.Println(ss)
	}
	WriteRegToScreen()
	fmt.Println()
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	//	scanner.Split(bufio.ScanWords).  Not by words, but bufio is back as of 10/07/2018 09:33:54 AM

	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
	} else {
		fmt.Print(" Enter calculation, HELP or (Q)uit to exit: ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			// _, err := fmt.Scan(&INBUF)  // removed 10/2/18.
			// _, err := fmt.Scanln(&INBUF) // added 10/02/2018 04:52:23 PM, and removed again on 10/07/2018 09:34:54 AM
			fmt.Fprintln(os.Stderr, "reading standard input from fmt.Scanln():", err)
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
		if INBUF == "ZEROREG" {
			for c := range Storage {
				Storage[c] = 0.0
			}
			AllowDumpFlag = false
		} else if strings.HasPrefix(INBUF, "STO") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // The 4th position.
				i = GetRegIdx(ch)
			}
			Storage[i] = hpcalc.READX()
			P()
			WriteRegToScreen()
			P()
			AllowDumpFlag = false
		} else if strings.HasPrefix(INBUF, "RCL") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // the 4th position.
				i = GetRegIdx(ch)
			}
			hpcalc.PUSHX(Storage[i])
		} else if strings.HasPrefix(INBUF, "SHO") { // so it will match SHOW and SHO
			fmt.Println()
			WriteRegToScreen()
			fmt.Println()
			WriteDisplayTapeToScreen()
			AllowDumpFlag = false
		} else if INBUF == "NO" || INBUF == "OFF" {
			ManualDump = false
		} else if INBUF == "YES" || INBUF == "ON" {
			ManualDump = true
		} else if INBUF == "TOCLIP" {
			R := hpcalc.READX()
			s := strconv.FormatFloat(R, 'g', -1, 64)
			if runtime.GOOS == "linux" {
				linuxclippy := func(s string) {
					buf := []byte(s)
					rdr := bytes.NewReader(buf)
					cmd := exec.Command("xclip")
					cmd.Stdin = rdr
					cmd.Stdout = os.Stdout
					fmt.Println("GOOS=", runtime.GOOS, " debugging string from clippy is:", cmd.String())
					cmd.Run()
				}
				linuxclippy(s)
			} else if runtime.GOOS == "windows" {
				winclippy := func(s string) {
					cmd := exec.Command("c:/Program Files/JPSoft/tcmd25/tcc.exe", "-C", "echo", s, ">clip:")
					cmd.Stdout = os.Stdout
					fmt.Println("GOOS=", runtime.GOOS, " debugging string from clippy is:", cmd.String())
					cmd.Run()
				}
				winclippy(s)
			}
		} else if INBUF == "FROMCLIP" {
			var w strings.Builder
			if runtime.GOOS == "linux" {
				cmdfromclip := exec.Command("xclip", "-o")
				cmdfromclip.Stdout = &w
				cmdfromclip.Run()
				str := w.String()
				str = strings.ReplaceAll(str, ",", "")
				R, err := strconv.ParseFloat(str, 64)
				if err != nil {
					log.Println(" fromclip on linux conversion returned error", err, ".  Value ignored.")
					AllowDumpFlag = false  // need to see the error msg.
				} else {
					hpcalc.PUSHX(R)
					AllowDumpFlag = true  // there is no error msg to see.
				}
			} else if runtime.GOOS == "windows" {
				cmdfromclip := exec.Command("c:/Program Files/JPSoft/tcmd25/tcc.exe", "-C", "echo", "%@clip[0]")
				cmdfromclip.Stdout = &w
				cmdfromclip.Run()
				lines := w.String()
				linessplit := strings.Split(lines, "\n")
				str := strings.ReplaceAll(linessplit[1], "\"", "")
				str = strings.ReplaceAll(str, "\n", "")
				str = strings.ReplaceAll(str, "\r", "")
				str = strings.ReplaceAll(str, ",", "")
				R, err := strconv.ParseFloat(str, 64)
				if err != nil {
					fmt.Println(" fromclip", err, ".  Value ignored.")
					AllowDumpFlag = false // need to see the error msg.
				} else {
					hpcalc.PUSHX(R)
					AllowDumpFlag = true // there's no error msg to see.
				}
			}
			// AllowDumpFlag = true

		} else {
			// -------------------------------------------------------------------------------------
			_, stringslice = hpcalc.GetResult(INBUF) //   Here is where GetResult is called
			// -------------------------------------------------------------------------------------
			AllowDumpFlag = true
			fmt.Println()
			if len(stringslice) > 0 {
				for _, ss := range stringslice {
					fmt.Println(ss)
				}
				AllowDumpFlag = false // if there is anything in the stringslice to output, suppress dump.
			} else if SuppressDump[INBUF] { // this may be redundant to the len(stringslice) conditional
				AllowDumpFlag = false
			} // end if len(stringslice) elsif SuppressDump
		}
		// -------------------------------------------------------------------------------------

		//  These commands are processed thru hpcalc first, then these are processed here.
		if strings.ToLower(INBUF) == "about" { // I'm using ToLower here just to experiment a little.
			fmt.Println(" Last altered the source of rpng.go", LastAlteredDate)
			AllowDumpFlag = false
		} else if strings.HasPrefix(INBUF, "DUMP") {
			AllowDumpFlag = false
		} else if INBUF == "HELP" || INBUF == "?" { // have more help lines to print
			fmt.Println(" NO,OFF,YES,ON -- Manually change a Dump flag for commands like PRIME and HEX.")
			fmt.Println(" toclip, fromclip -- uses shell commands to access the clipboard")
			AllowDumpFlag = false
		} else if INBUF == "HOL" {
			AllowDumpFlag = false
		} else if INBUF == "DOW" {
			AllowDumpFlag = false
		} // if INBUF
		if AllowDumpFlag && ManualDump {
			ClearScreen()
			RepaintScreen()
		}
		fmt.Println()
		fmt.Print(" Enter calculation, HELP or (Q)uit to exit: ")
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
		if strings.HasPrefix(INBUF, "Q") || INBUF == "EXIT" {
			fmt.Println()
			break
		}
		ClearScreen()
		RepaintScreen()
	}

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

} // main in rpng.go

/* ------------------------------------------------------------ GetRegIdx --------- */
func GetRegIdx(chr byte) int {
	// Return 0..35 with A = 10 and Z = 35
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
func WriteRegToScreen() {
	FirstNonZeroStorageFlag := true

	for i, r := range Storage {
		if r != 0.0 {
			if FirstNonZeroStorageFlag {
				fmt.Println("                The following storage registers are not zero")
				FirstNonZeroStorageFlag = false
			} // if firstnonzerostorageflag
			ch := GetRegChar(i)
			s := strconv.FormatFloat(r, 'g', -1, 64)
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
		fmt.Println(s)
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
	fmt.Println(" HP-type RPN calculator written in Go.  Last altered source of rpng.go", LastAlteredDate)
	//	fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fmt.Println()
	_, stringslice := hpcalc.GetResult("DUMPFIXED")
	fmt.Println("                                   ", stringslice[len(stringslice)-2])
	fmt.Println()
	for _, ss := range stringslice {
		fmt.Println(ss)
	}
	WriteRegToScreen()
	WriteDisplayTapeToScreen()
}

// ---------------------------------------------------- End rpng.go ------------------------------
