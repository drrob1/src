// (C) 1990-2016.  Robert W. Solomon.  All rights reserved.
// rpng.go
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	//"path"
	//"path/filepath"
	//
	"src/getcommandline"
	"src/holidaycalc"
	"src/hpcalc"
	"src/makesubst"
	"src/timlibg"
	"src/tokenize"
	//                                                                                                      "timlibg"
)

const LastCompiled = "8 Sep 16"

var Storage [36]float64 // 0 .. 9, a .. z
var DisplayTape []string

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
	   - 24 Dec 91 -- Converted to M-2 V 4.00.  Changed params to GETRESULT.
	     25 Jul 93 -- Output result without trailing insignificant zeros,
	                   imported UL2, and changed prompt again.
	      3 Mar 96 -- Fixed bug in string display if real2str fails because
	                   number is too large (ie, Avogadro's Number).
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
	*/

	var R float64
	var Y, NYD, July4, VetD, ChristmasD int //  For Holiday cmd
	var INBUF, line, HomeDir string         // no longer use ans so I have to undeclare it.
	const StackFileName = "RPNStack.sav"
	const Storage1FileName = "RPNStorage.sav" // Allows for a rotation of Storage files, in case of a mistake.
	const Storage2FileName = "RPNStorage2.sav"
	const Storage3FileName = "RPNStorage3.sav"
	var Holidays holidaycalc.HolType

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.
	var err error

	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
	} else { // then HomeDir will be empty.
		fmt.Println(" runtime.GOOS does not say linux or windows.  Don't know why.")
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(" GOOS =", runtime.GOOS, ".  HomeDir =", HomeDir, ".  ARCH=", runtime.GOARCH)
	fmt.Println()
	fmt.Println()

	StackFullFileNameSlice := make([]string, 5)
	StackFullFileNameSlice[0] = HomeDir
	StackFullFileNameSlice[1] = string(os.PathSeparator)
	StackFullFileNameSlice[2] = StackFileName
	StackFullFileName := strings.Join(StackFullFileNameSlice, "")
	fmt.Print(" StackFullFilename:", StackFullFileName)

	AllowDumpFlag := false // I need to be able to supress the automatic DUMP under certain circumstances.
	DisplayTape = make([]string, 0, 100)
	StackFileExists := true
	InputByteSlice := make([]byte, 8*hpcalc.StackSize) // I hope this is a slice of 64 bytes, ie, 8*8.

	StorageFullFilenameSlice := make([]string, 5)
	StorageFullFilenameSlice[0] = HomeDir
	StorageFullFilenameSlice[1] = string(os.PathSeparator)
	StorageFullFilenameSlice[2] = Storage1FileName
	StorageFullFilename := strings.Join(StorageFullFilenameSlice, "")
	fmt.Println("     StorageFullFilename:", StorageFullFilename)
	fmt.Println()
	fmt.Println()

	if InputByteSlice, err = ioutil.ReadFile(StackFileName); err != nil {
		fmt.Errorf(" Error from ioutil.ReadFile.  Possibly because no Stack File found: %v\n", err)
		StackFileExists = false
	}
	if StackFileExists { // read all into memory.
		for i := 0; i < hpcalc.StackSize*8; i = i + 8 {
			buf := bytes.NewReader(InputByteSlice[i : i+8])
			err := binary.Read(buf, binary.LittleEndian, &R)
			if err != nil {
				fmt.Errorf(" binary.Read failed with error of %v \n", err)
				StackFileExists = false
			}
			hpcalc.PUSHX(R)
		} // loop to extract each 8 byte chunk to convert to a longreal (float64) and push onto the hpcalc stack.
	} // stackfileexists

	StorageFileExists := true
	InputStorageByteSlice := make([]byte, 36*8)
	if InputStorageByteSlice, err = ioutil.ReadFile(Storage1FileName); err != nil {
		fmt.Errorf(" Error from ioutil.ReadFile when trying to read StorageFileName: %v\n", err)
		StorageFileExists = false
	}

	if StorageFileExists {
		j := 0
		for i := 0; i < 36*8; i += 8 {
			buf := bytes.NewReader(InputStorageByteSlice[i : i+8])
			err := binary.Read(buf, binary.LittleEndian, &R)
			if err != nil {
				fmt.Errorf(" binary.Read failed when reading Storage file with error of %v\n", err)
				StorageFileExists = false
			}
			Storage[j] = R
			j++
		} // for loop to process each float64
	} // if StorageFileExists

	hpcalc.PushMatrixStacks()

	fmt.Println(" HP-type RPN calculator written in Go last compiled ", LastCompiled)
	fmt.Println()
	fmt.Println()
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
	} else {
		fmt.Print(" Enter calculation, HELP or Enter to exit: ")
		scanner.Scan()
		INBUF = strings.TrimSpace(scanner.Text())
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(INBUF) == 0 {
			os.Exit(0)
		}
	} // if command tail exists

	hpcalc.PushMatrixStacks()

	for len(INBUF) > 0 { // Main processing loop
		INBUF = makesubst.MakeSubst(INBUF)
		DisplayTape = append(DisplayTape, INBUF) // This is an easy way to capture everything.
		INBUF = strings.ToUpper(INBUF)
		// These commands are not run thru hpcalc as they are processed before calling GetResult.
		if INBUF == "ZEROREG" {
			for c := range Storage {
				Storage[c] = 0.0
			}
			AllowDumpFlag = false
		} else if INBUF == "DOW" {
			r, _ := hpcalc.GetResult("DOW") // discard results.  I think this is allowed.
			i := int(r)
			fmt.Println(" Day of Week for X value is a ", timlibg.DayNames[i])
			AllowDumpFlag = false
		} else if strings.HasPrefix(INBUF, "STO") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // The 4th position.
				i = GetRegIdx(ch)
			}
			Storage[i] = hpcalc.READX()
		} else if strings.HasPrefix(INBUF, "RCL") {
			i := 0
			if len(INBUF) > 3 {
				ch := INBUF[3] // the 4th position.
				i = GetRegIdx(ch)
			}
			hpcalc.PUSHX(Storage[i])
		} else if strings.HasPrefix(INBUF, "SHO") { // so it will match SHOW and SHO
			WriteRegToScreen()
			WriteDisplayTapeToScreen()
			AllowDumpFlag = false
		} else {
			_, Holidays = hpcalc.GetResult(INBUF)
			AllowDumpFlag = true
		}
		//                                                                        ans = strconv.FormatFloat(R,'g',-1,64);
		//                                                                                  ans = hpcalc.CropNStr(ans);
		//                                                                                  if R > 10000 {
		//                                                                                    ans = hpcalc.AddCommas(ans);
		//                                                                                  }
		//    fmt.Println();
		//    fmt.Println();

		//  These commands are processed after GetResult is called, so these commands are run thru hpcalc.
		if strings.ToLower(INBUF) == "about" { // I'm using ToLower here just to experiment a little.
			fmt.Println(" Last compiled rpng.go ", LastCompiled)
			AllowDumpFlag = false
		} else if strings.HasPrefix(INBUF, "DUMP") {
			AllowDumpFlag = false
		} else if INBUF == "HELP" {
			AllowDumpFlag = false
		} else if Holidays.Valid {
			AllowDumpFlag = false
			fmt.Println(" For year ", Holidays.Year, ":")
			Y = Holidays.Year
			NYD = (timlibg.JULIAN(1, 1, Y) % 7)
			line = fmt.Sprintf("New Years Day is a %s, MLK Day is January %d, Pres Day is February %d, Easter Sunday is %s %d, Mother's Day is May %d",
				timlibg.DayNames[NYD], Holidays.MLK.D, Holidays.Pres.D, timlibg.MonthNames[Holidays.Easter.M], Holidays.Easter.D, Holidays.Mother.D)
			fmt.Println(line)

			July4 = (timlibg.JULIAN(7, 4, Y) % 7)
			line = fmt.Sprintf("Memorial Day is May %d, Father's Day is June %d, July 4 is a %s, Labor Day is Septempber %d, Columbus Day is October %d",
				Holidays.Memorial.D, Holidays.Father.D, timlibg.DayNames[July4], Holidays.Labor.D, Holidays.Columbus.D)
			fmt.Println(line)

			VetD = (timlibg.JULIAN(11, 11, Y) % 7)
			ChristmasD = (timlibg.JULIAN(12, 25, Y) % 7)
			line = fmt.Sprintf("Election Day is November %d, Veteran's Day is a %s, Thanksgiving is November %d, and Christmas Day is a %s.",
				Holidays.Election.D, timlibg.DayNames[VetD], Holidays.Thanksgiving.D, timlibg.DayNames[ChristmasD])
			fmt.Println(line)
			Holidays.Valid = false
		} // if INBUF
		if AllowDumpFlag {
			hpcalc.GetResult("DUMP") // discard result.  Only need stack dump general executed.
		}
		fmt.Println()
		fmt.Print(" Enter calculation, HELP or Enter to exit: ")
		scanner.Scan()
		INBUF = strings.TrimSpace(scanner.Text())
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		//    fmt.Print("                                                               Result = ");
		//    hpcalc.OutputFixedOrFloat(R);
		//    fmt.Println("         |    ",ans);
	}

	// Now that I've got this binary stuff working, I'm taking notes.  The binary.Write appends to the buf after each
	// call, since I'm not doing anthing to the bytes.Buffer to reset it.  I don't need a separate slice of
	// bytes to accumulate the stack for output.  I just have to reverse the order I write them out so that
	// they are read in correctly, without reversing the stack after each write.  I could reset the
	// buf.Bytes each time if I wanted.  I tested that and it works.  But it is unnecessary for my needs so
	// I commented it out.

	Stk = hpcalc.GETSTACK()
	//  OutputByteSlice := make([]byte,0,8*hpcalc.StackSize);  // each stack element is a float64, and there are StackSize of these (now StackSize=8), so this is a slice of 64 bytes
	buf := new(bytes.Buffer)

	for i := hpcalc.T1; i >= hpcalc.X; i-- { // in reverse.  for range cannot go in reverse.
		r := Stk[i]
		err := binary.Write(buf, binary.LittleEndian, r)
		if err != nil {
			fmt.Errorf(" binary.write into buf failed with error %v \n", err)
			os.Exit(1)
		}
		//    fmt.Println(" Got Stk.  buf.Bytes len =",len(buf.Bytes()),". buf.Bytes: ",buf.Bytes());
		//    OutputByteSlice = append(OutputByteSlice,buf.Bytes()...);
		//    fmt.Println(" Length of OutByteSlice after append operation ",len(OutputByteSlice));
		//    buf.Reset();
	} // for range in reverse over the stack.
	err = ioutil.WriteFile(StackFileName, buf.Bytes(), os.ModePerm) // os.ModePerm = 0777
	if err != nil {
		fmt.Errorf(" ioutil.WriteFile of stack failed with error %v \n", err)
	}

	// Rotate StorageFileNames and write
	err = os.Rename(Storage2FileName, Storage3FileName)
	if err != nil {
		fmt.Errorf(" Rename of storage 2 to storage 3 failed with error %v \n", err)
	}

	err = os.Rename(Storage1FileName, Storage2FileName)
	if err != nil {
		fmt.Errorf(" Rename of storage 1 to storage 2 failed with error %v \n", err)
	}

	StorageBuf := new(bytes.Buffer)
	for _, r := range Storage {
		err := binary.Write(StorageBuf, binary.LittleEndian, r)
		if err != nil {
			fmt.Errorf(" binary.write into StorageBuf failed with error %v\n", err)
			os.Exit(1)
		}
	} // for range Storage

	//        fmt.Println(" StorageBuf.Bytes len =",len(StorageBuf.Bytes()),".  StorageBuf.Bytes",StorageBuf.Bytes());

	err = ioutil.WriteFile(Storage1FileName, StorageBuf.Bytes(), os.ModePerm) // os.ModePerm =0777
	if err != nil {
		fmt.Errorf(" ioutil.WriteFile Storage1FileName failed with error %v\n", err)
	}
} // main in rpng.go

/************************************************************* GetRegIdx ****/
func GetRegIdx(chr byte) int {
	/* Return 0..35 w/ A = 10 and Z = 35 */

	ch := tokenize.CAP(chr)
	if (ch >= '0') && (ch <= '9') {
		ch = ch - '0'
	} else if (ch >= 'A') && (ch <= 'Z') {
		ch = ch - 'A' + 10
	}
	return int(ch)
} // GetRegIdx
/************************************************************** GetRegChar ******/
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
// ---------------------
func WriteRegToScreen() {
	FirstNonZeroStorageFlag := true

	for i, r := range Storage {
		if r != 0.0 {
			if FirstNonZeroStorageFlag {
				fmt.Println(" The following storage registers are not zero")
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
// -----------------------
func WriteDisplayTapeToScreen() {
	fmt.Println(" DisplayTape ")
	for _, s := range DisplayTape {
		fmt.Println(s)
	} // for ranging over DisplayTape slice of strings
	fmt.Println()
} // WriteDisplayTapeToScreen

// ---------------------------------------------------- End rpng.go ------------------------------
