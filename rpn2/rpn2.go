// (C) 1990-2020.  Robert W Solomon.  All rights reserved.
// rpn2.go, testing hpcalc2.go
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"runtime"
	"src/tknptr"
	"strconv"
	"strings"
	//
	"src/getcommandline"
	"src/hpcalc2"
)

/*
This module uses the HPCALC2 module to simulate an RPN type calculator.
REVISION HISTORY
----------------
 1 Dec 89 -- Changed prompt.
24 Dec 91 -- Converted to M-2 V 4.00.  Changed params to GETRESULT.
25 Jul 93 -- Output result without trailing insignificant zeros, imported UL2, and changed prompt again.
 3 Mar 96 -- Fixed bug in string display of real2str fails because number is too large (ie, Avogadro's Number).
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
22 Aug 16 -- Started conversion to Go.
28 Aug 16 -- added makesubst capability for substitutions = -> + and ; -> *
28 Nov 16 -- Backported stringslice return of hpcalc so can use the new, improved hpcalc.
23 Feb 17 -- Made "?" equivalent to "help"
26 Mar 17 -- Changed startup message to include written in Go.
22 Aug 18 -- Learning about code folding
 2 Oct 18 -- Now using code folding.  za normal mode command toggles the fold mode where cursor is.
 8 Feb 20 -- Added PopX to hpcalc.go, and will test it here.
 9 Apr 20 -- Will add the suppressdump map I've been using for a while in rpng.
 8 Aug 20 -- Now called rpn2.go to test hpcalc2.go
 8 Nov 20 -- Removed unnecessary comments.
12 Dec 20 -- hpcalc2 now has MAP commands.
13 Dec 20 -- Shortened the lead space on displaying the Results.
31 Jan 21 -- Adding color.  And windowsFlag so color is better, ie, use bold flag on windows.
 8 Apr 21 -- Converted to module src residing at ~/go/src.  What a coincidence.
16 Jun 21 -- Will use strings.ReplaceAll instead of my makesubst to test that '=' now adds from hpcalc2.
               And ioutil package is depracated as of Go 1.16, so I removed it.
17 Jun 21 -- Testing to see if the defer I put there works.  It does.
19 Jun 21 -- Changed hpcalc2 MAP commands so that the STO and DEL call mapWriteAndClose (formerly MapClose), so I don't have to do that explicitly.
21 Jun 21 -- Changing the awkward looking code that reads in the stack from the stackfile.
10 Aug 22 -- "about" will print info about the exe file.
24 Jun 23 -- Checking to make sure that the code is correct after I stopped exporting MapClose and changed its name to mapWriteAndClose.
               It's fine, as I made that change 2 years ago.  I changed some comments here, and I have to recompile because I changed hpcalc2.
 8 Jul 23 -- I coded a simpler TokenReal(), to replace GETTKNREAL().  I'm testing it here in production.  The rtn already passed in tknptr_test.go.
 9 Jul 23 -- Added modification of tknptr to the about cmd.
*/

const LastCompiled = "July 9, 2023"

var suppressDump map[string]bool

var windowsFlag bool

func main() {
	var R float64
	var INBUF, ans string
	const StackFileName = "RPNStack.sav"
	var stringslice []string

	var Stk hpcalc2.StackType // used when time to write out the stack upon exit.
	var err error

	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	suppressDump = make(map[string]bool)
	suppressDump["PRIME"] = true
	suppressDump["HEX"] = true
	suppressDump["DOW"] = true
	suppressDump["HOL"] = true
	suppressDump["ABOUT"] = true
	suppressDump["HELP"] = true
	suppressDump["TOCLIP"] = true
	suppressDump["DUMP"] = true
	suppressDump["DUMPFIX"] = true
	suppressDump["DUMPFIXED"] = true
	suppressDump["DUMPFLOAT"] = true
	suppressDump["?"] = true

	allowDumpFlag := true
	windowsFlag = runtime.GOOS == "windows"

	StackFileExists := true
	//InputByteSlice := make([]byte, 8*hpcalc2.StackSize) // this is a slice of 64 bytes, ie, 8*8.  But I don't need to do this.

	InputByteSlice, err := os.ReadFile(StackFileName)
	if err != nil {
		fmt.Printf(" Error from os.ReadFile.  Probably because no Stack File found: %v\n", err)
		StackFileExists = false
	}
	if StackFileExists { // trying another way to read this file.  If it doesn't work, I'll use encoding/gob as in hpcalc2.  The original way that works is still in rpn.go
		buf := bytes.NewReader(InputByteSlice)
		for i := 0; i < hpcalc2.StackSize; i++ {
			err := binary.Read(buf, binary.LittleEndian, &R)
			if err != nil {
				fmt.Printf(" binary.Read failed with error of %v \n", err)
				StackFileExists = false
				break
			}
			hpcalc2.PUSHX(R)
		} // loop to read each 8 byte chunk to convert to a longreal (float64) and push onto the hpcalc stack.
	} // stackfileexists

	hpcalc2.PushMatrixStacks()

	fmt.Println(" HP-type RPN calculator written in Go.  Last compiled ", LastCompiled, "using", runtime.Version())
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
		//INBUF = makesubst.MakeSubst(INBUF)
		INBUF = strings.ReplaceAll(INBUF, ";", "*")
	} else {
		fmt.Print(" Enter calculation, HELP or Enter to exit: ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(INBUF) == 0 {
			os.Exit(0)
		}
		//INBUF = makesubst.MakeSubst(INBUF)
		INBUF = strings.ReplaceAll(INBUF, ";", "*")
	} // if command tail exists

	hpcalc2.PushMatrixStacks()

	for len(INBUF) > 0 {
		R, stringslice = hpcalc2.GetResult(INBUF)
		sigfig := hpcalc2.SigFig()
		ans = strconv.FormatFloat(R, 'g', sigfig, 64)
		ans = hpcalc2.CropNStr(ans)
		if R > 10000 {
			ans = hpcalc2.AddCommas(ans)
		}
		fmt.Println()
		fmt.Println()
		for _, ss := range stringslice {
			ctfmt.Println(ct.Green, windowsFlag, ss)

			allowDumpFlag = false // Don't show stack if any strings were returned from hpcalc.GetResult()
		}

		if strings.ToLower(INBUF) == "about" {
			ctfmt.Printf(ct.Cyan, windowsFlag, " tknptr last altered %s\n", tknptr.LastAltered)
			ctfmt.Println(ct.Cyan, windowsFlag, " Last changed rpn2.go ", LastCompiled)
			ctfmt.Printf(ct.Cyan, windowsFlag, " %s timestamp is %s.  Full exec name is %s.\n", ExecFI.Name(), ExecTimeStamp, execname)
			allowDumpFlag = false
		}

		INBUF = strings.ToUpper(INBUF)
		if suppressDump[INBUF] {
			allowDumpFlag = false
		}

		if allowDumpFlag { // display stack.
			_, stringslice = hpcalc2.GetResult("DUMP") // discard result.  Only need stack dump general executed.
			for _, ss := range stringslice {
				ctfmt.Println(ct.Cyan, windowsFlag, ss)
			}
		}

		fmt.Println()
		ctfmt.Print(ct.Yellow, windowsFlag, "                  Result = ")
		hpcalc2.OutputFixedOrFloat(R)
		ctfmt.Println(ct.Yellow, windowsFlag, "         |    ", ans)
		ctfmt.Print(ct.Blue, windowsFlag, " Enter calculation, HELP or Enter to exit: ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		//INBUF = makesubst.MakeSubst(INBUF)
		INBUF = strings.ReplaceAll(INBUF, ";", "*")
		allowDumpFlag = true
	}

	// Now that I've got this working, I'm taking notes.  The binary.Write appends to the buf after each call,
	// since I'm not doing anything to the bytes.Buffer to reset it.  I don't need a separate slice of
	// bytes to accumulate the stack for output.  I just have to reverse the order I write them out so that
	// they are read in correctly, without reversing the stack after each write.  I could reset the buf.Bytes
	// each time if I wanted.  I tested that and it works.  But it is unnecessary for my needs so I commented it out.

	Stk = hpcalc2.GETSTACK()
	buf := new(bytes.Buffer)

	for i := hpcalc2.T1; i >= hpcalc2.X; i-- { // in reverse.  for range cannot go in reverse.
		r := Stk[i]
		err := binary.Write(buf, binary.LittleEndian, r)
		if err != nil {
			ctfmt.Println(ct.Red, windowsFlag, err)
			fmt.Printf(" binary.write into buf failed with error %v \n", err)
			os.Exit(1)
		}
	}
	err = os.WriteFile(StackFileName, buf.Bytes(), os.ModePerm) // os.ModePerm = 0777
	if err != nil {
		fmt.Printf(" os.WriteFile failed with error %v \n", err)
	}
} // main in rpn2.go
