// (C) 1990-2016.  Robert W Solomon.  All rights reserved.
// rpn.go
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	//
	"getcommandline"
	hpcalc "hpcalc2"
	"makesubst"
	//{{{
	//                                                                                          "holidaycalc"
	//                                                                                              "timlibg"
	//                                                                                             "tokenize"
	//                                                                                              "timlibg"
	//}}}
)

const LastCompiled = "8 Nov 2020"

var suppressDump map[string]bool

func main() {
	/*
	   This module uses the HPCALC module to simulate an RPN type calculator.
	   REVISION HISTORY
	   ----------------
	    1 Dec 89 -- Changed prompt.
	   24 Dec 91 -- Converted to M-2 V 4.00.  Changed params to GETRESULT.
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
	   22 Aug 16 -- Started conversion to Go.
	   28 Aug 16 -- added makesubst capability for substitutions = -> + and ; -> *
	   28 Nov 16 -- Backported stringslice return of hpcalc so can use the new, improved hpcalc.
	   23 Feb 17 -- Made "?" equivalent to "help"
	   26 Mar 17 -- Changed startup message to include written in Go.
	   22 Aug 18 -- Learning about code folding
	    2 Oct 18 -- Now using code folding.  za normal mode command toggles the fold mode where cursor is.
	    8 Feb 20 -- Added PopX to hpcalc.go, and will test it here.
	    9 Apr 20 -- Will add the suppressdump map I've been using for a while in rpng.
	    8 Nov 20 -- Now will use hpcalc2.  I'm adding toclip, fromclip (based on code from "Go Standard Library Cookbook") to hpcalc2.
	                And finally removed code that was commented out long ago.
	*/

	var R float64
	var INBUF, ans string
	const StackFileName = "RPNStack.sav"
	var stringslice []string

	var Stk hpcalc.StackType // used when time to write out the stack upon exit.

	var err error

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
	//        suppressDump[""] = true

	allowDumpFlag := true

	StackFileExists := true
	InputByteSlice := make([]byte, 8*hpcalc.StackSize) // I hope this is a slice of 64 bytes, ie, 8*8.

	if InputByteSlice, err = ioutil.ReadFile(StackFileName); err != nil {
		fmt.Printf(" Error from ioutil.ReadFile.  Probably because no Stack File found: %v\n", err)
		StackFileExists = false
	}
	if StackFileExists { // I'll read all into memory.
		for i := 0; i < hpcalc.StackSize*8; i = i + 8 {
			buf := bytes.NewReader(InputByteSlice[i : i+8])
			err := binary.Read(buf, binary.LittleEndian, &R)
			if err != nil {
				fmt.Printf(" binary.Read failed with error of %v \n", err)
				StackFileExists = false
			}
			hpcalc.PUSHX(R)
		} // loop to extract each 8 byte chunk to convert to a longreal (float64) and push onto the hpcalc stack.
	} // stackfileexists

	hpcalc.PushMatrixStacks()

	fmt.Println(" HP-type RPN calculator written in Go.  Last compiled ", LastCompiled)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		INBUF = getcommandline.GetCommandLineString()
		INBUF = makesubst.MakeSubst(INBUF)
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
		INBUF = makesubst.MakeSubst(INBUF)
	} // if command tail exists

	hpcalc.PushMatrixStacks()

	for len(INBUF) > 0 {
		R, stringslice = hpcalc.GetResult(INBUF)
		ans = strconv.FormatFloat(R, 'g', -1, 64)
		ans = hpcalc.CropNStr(ans)
		if R > 10000 {
			ans = hpcalc.AddCommas(ans)
		}
		fmt.Println()
		fmt.Println()
		for _, ss := range stringslice {
			fmt.Println(ss)

			allowDumpFlag = false // Don't update stack if any strings were returned from hpcalc.GetResult()
		}

		if strings.ToLower(INBUF) == "about" {
			fmt.Println(" Last compiled rpn.go ", LastCompiled)
			allowDumpFlag = false
		}

		INBUF = strings.ToUpper(INBUF)
		if suppressDump[INBUF] {
			allowDumpFlag = false
		}
		//		if !(strings.HasPrefix(INBUF, "DUMP") || INBUF == "HELP" || INBUF == "?") { // old way of suppressing DUMP: if just did it, or if HELP called to not scroll help off screen.
		if allowDumpFlag {
			_, stringslice = hpcalc.GetResult("DUMP") // discard result.  Only need stack dump general executed.
			for _, ss := range stringslice {
				fmt.Println(ss)
			}
		}

		fmt.Println()
		fmt.Print("                                            Result = ")
		hpcalc.OutputFixedOrFloat(R)
		fmt.Println("         |    ", ans)
		fmt.Print(" Enter calculation, HELP or Enter to exit: ")
		scanner.Scan()
		INBUF = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		INBUF = makesubst.MakeSubst(INBUF)
		allowDumpFlag = true
	}

	// Now that I've got this working, I'm taking notes.  The binary.Write appends to the buf after each
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
			fmt.Printf(" binary.write into buf failed with error %v \n", err)
			os.Exit(1)
		}
		// {{{
		//    fmt.Println(" Got Stk.  buf.Bytes len =",len(buf.Bytes()),". buf.Bytes: ",buf.Bytes());
		//    OutputByteSlice = append(OutputByteSlice,buf.Bytes()...);
		//    fmt.Println(" Length of OutByteSlice after append operation ",len(OutputByteSlice));
		//    buf.Reset();
		// }}}
	}
	err = ioutil.WriteFile(StackFileName, buf.Bytes(), os.ModePerm) // os.ModePerm = 0777
	if err != nil {
		fmt.Printf(" ioutil.WriteFile failed with error %v \n", err)
	}
} // main in rpn.go
