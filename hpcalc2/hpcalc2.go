package hpcalc2

import (
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	//
	"src/holidaycalc"
	"src/timlibg"
	"src/tknptr"
)

/* (C) 1990-2024.  Robert W Solomon.  All rights reserved.
REVISION HISTORY
----------------
 1 Dec 89 -- Added the help command.
24 Dec 91 -- Converted to M-2 V 4.00.  Also changed the params to the GETRESULT proc to be more reasonable.
21 Mar 93 -- Added exponentiation and MOD, INT, FRAC and ROUND, as well as used the UL2 procs again.
25 Jul 93 -- Added JUL and GREG commands.
18 May 03 -- Win32 version using Stony Brook Modula-2 v 4
26 May 03 -- Allowed years to pivot for 2000 or 1900 in TIMLIB.
 4 Oct 03 -- Added LongCard2HexStr and its cmd.
31 Oct 03 -- Added HCF cmd.
 1 Nov 03 -- Fixed the var swap bug in HCF rtn.
27 Nov 04 -- Added pi as a cmd.
12 Mar 06 -- Made trig fcn's use arguments in degrees, changed stacksize to const, and moved stackregnames to def.
22 Jul 06 -- Added % operator does xy/100 but does not drop stack.  Just like old HP 25.
15 Apr 07 -- Added comma to push stack.  Removed it as a delim char and made it a ALLELSE.  And
              added ! ` and ~ as stack commands.  And updated i/o procedure use.
 1 Apr 13 -- Modified the DUMP routine so the characters are not printed on the same line as the numbers.  The
              alignment was all wrong anyway.  And added Dump2Console.
 2 May 13 -- Will use consolemode compilation flags
15 Oct 13 -- converting to gm2.
28 Jun 14 -- converting to Ada.
31 Aug 14 -- found that tknrtnsa does not process gettknreal correctly.  Had to rewrite it to match the Modula-2 version.
 2 Sep 14 -- Added ToHex back.
 1 Dec 14 -- Converted to cpp.
20 Dec 14 -- Added use of macros for date of last compile
31 Dec 14 -- Started coding HOL command, and changed the limits for the OutputFixedOrFloat proc to print Julian numbers in fixed format.
10 Jan 15 -- Coding use of string operations to supplement output of stack.  And number cropping, addcommas.
 5 Nov 15 -- New commands for recip, curt, vol.
22 Nov 15 -- Noticed that T1 and T2 stack operations are not correct.  This effects HP2cursed and rpnc.  Changed ARRAYOF Stack declaration.
              Was Stack[T1], now is Stack[StackSize].  The declaration is number of elements, not the high bound.
13 Apr 16 -- Adding undo and redo commands, which operate on the entire stack not just X register.
19 May 16 -- Fixing help text for the % commanded coded in 2006.  Oddly enough, the help never included it.
 2 Jul 16 -- Fixed help text for the PI command.  And changed the pivot for the JUL command to be the current year instead of the constant "30".  HOL command pivot remains 40.
 7 Jul 16 -- Added UP command.  Surprising I had not done this earlier.
 9 Jul 16 -- Fixed bug in timlibc.  When juldate is too small get infinite loop in GREGORIAN
18 Aug 16 -- Started conversion to Go
29 Aug 16 -- Added Prime command and support function IsPrime.  And PWRI came back into use.
 5 Sep 16 -- Changed output format verb params in fixed and general format.
11 Sep 16 -- Had to fix help text now that PWRI is used.  No longer do Abs(Y) before calling PWRI.
 7 Oct 16 -- Added the adjust command so that amounts are not .9999997 or something silly like that.  And trying NextAfter.
               Noticed that math.Cbrt exists, and removed Abs(Y) from ** operator implemented by a call to math.Pow();
               Added SigFig or Fix command for the strconv.FormatFloat fcn call.
 8 Oct 16 -- Fixed help by adding trunc command.  Seems I never added it to help when I added the command.
 3 Nov 16 -- This rtn will now return a string when it needs to, instead of doing its own output.  This is so that termbox-go is as smooth as rpng.
 4 Nov 16 -- Changed how DOW is handled.  It now returns its answer in stringslice.
28 Nov 16 -- Decided that the approximation for vol command was not necessary, and will use more exact formula not assuming pi = 3.
               And added piover6 command.  And added CHS (change sign) command which also allows underscore, _, as the symbol for this command.
29 Nov 16 -- Decided to reorder the statements in the main if statement of GetResults to optimize for probability of usage.  Mostly, the trig,
               log and exp functions were moved to the bottom, and I combined conditions into compound OR conditionals for clarity of function.
23 Feb 17 -- Removed a redundant line in the help command.  Changed ! to | so I can implement factorial.  Help can now be called by ?.
24 Feb 17 -- Added PrimeFactorization
16 Mar 17 -- Rephrased help text for vol command.
19 Mar 17 -- Made LOG a synonym for LN.
26 Mar 17 -- Fixed help text regarding ^ and ** operators.
 4 Apr 17 -- Added BEFORE command to use NextAfter towards 0.  And made "AFTER" a synonym for NEXTAFTER.
26 Apr 17 -- Fixed help text to remove stop, which blocked STO into the P register.  Edited the HeaderDivider to align a '+' with '|'.
               At some point, "?" became a synonym for "help"
29 May 17 -- Found bug in CropNStr.  If number is in scientific notation, and the exponent ends in 0, that will be removed.
13 July 17 -- Rewrote ToHex, based on code from the Python mooc I'm taking now.  And with more experience.
25 Feb 18 -- PrimeFactorMemoized added.
27 Feb 18 -- Fixed bug in PrimeFactorMemoized and support routines.
 8 Mar 18 -- Fixed bug in IsPrime rtn.
 2 Dec 18 -- Fixed comments regarding before and after commands.  And updated the help command to include NAME.
 5 Dec 18 -- Help from here will only produce help text for those commands processed here.
15 Apr 19 -- Increased size of stringslice
 1 May 19 -- Noticed that the character substitutions were never listed here in help as they are in the cpp version.  Now added that.
 3 Jun 19 -- Added t as an abbreviation for today
28 Jun 19 -- Added prev or previous as synonym for before
29 Dec 19 -- For dumpfloat and dumpgen, CropNstr and addcommas do not make sense, so I took them out.  Finally!
30 Dec 19 -- Reordered command tests, moving up PRIMEFAC
22 Jan 20 -- Noticed that holiday command, hol, only works if X register is a valid year.  Now prints a message to remind me of that.
 8 Feb 20 -- Added PopX, because discovered that ROLLDN does not affect X, by design.  I don't remember why.
 9 Feb 20 -- HCF now reports a message and does not alter the stack.  This one I coded in cpp first, as it turns out.
               And only a command that changes the stack needs to call PushMatrixStacks.  I removed that call from hol and a few others.
22 Mar 20 -- Shortened PrimeFac to PrimeF, just like the C++ version I wrote for Qt.  And fix bug of primefac of zero or a number near zero.
 7 Apr 20 -- Decided to comment out the break statements in the GetResult case statement, which is held over from my C++ code.  Doesn't belong here.
 9 Apr 20 -- Switched to tknptr from tokenize package.  I guess mostly to test it.  I should have done this when I first wrote it.
15 Apr 20 -- Fixed AddCommas to ignore the string if there is an 'E', ie, string is in scientific notation.
25 Jun 20 -- Changed vol command to take numbers in x,y,z and compute volume, and added dia command to get diameter of a sphere with volume in X.
 3 Jul 20 -- Added cbrt as synonym for curt.
 7 Aug 20 -- Now called hpcal2.go, and will use a map to get a commandNumber, and then a switch-case on command number.
               And made a minimal change in GetResult for variable I to make code more idiomatic.
 9 Aug 20 -- Cleaned out some old, unhelpful comments, and removed one extraneous "break" in GetResult tknptr.OP section.
24 Oct 20 -- Fixed a bug in that if cmd is < 3 character, subslicing a slice panic'd w/ an out of bounds error.
 8 Nov 20 -- Adding toclip, fromclip, based on code from "Go Standard Library Cookbook", by Radomir Sohlich, (c) 2018 Packtpub.
 9 Nov 20 -- Including use of comspec to find tcc on Windows.
 4 Dec 20 -- Thinking about how to add conversion factors.  1 lb = 453.59238 g; 1 oz = 28.34952 g; 1 m = 3.28084 ft; 1 mi = 1.609344 km
11 Dec 20 -- Fixed a line in the help command reporting this module as hpcalc instead of hpcalc2.
12 Dec 20 -- Adding mappedReg stuff.  And new commands mapsho, mapsto, maprcl, mapdel and mapclose.
14 Dec 20 -- Decided to sort mapsho output.
17 Dec 20 -- Will implement mapped register recall using abbreviations, ie, match prefix against a sorted list of the available mapped registers.
               and added C2F, F2C
21 Dec 20 -- Changed MAPRCL abbreviation concept from strings.HasPrefix to strings.Contains, so substrings are matched instead of just prefixes.
30 Jan 21 -- Results of the converstions functions also push their result onto the stack.
31 Jan 21 -- Added SigFig()
 3 Feb 21 -- Fixed bug in what gets pushed onto the stack with the c2f and f2c commands.
 4 Feb 21 -- Added H for help.
11 Feb 21 -- Added these commands that will be ignored, X, P and Q.  And took out come dead code.
 8 Apr 21 -- Converting to src module residing at ~/go/src.  What a coincidence!
14 Jun 21 -- Split off Result from GetResult
16 Jun 21 -- Adding os.UserHomeDir(), which became available as of Go 1.12.
17 Jun 21 -- Added "defer mapWriteAndClose()" to the init function, to see it this works.  It doesn't, so I removed it.
               Deferred code will be run at the end of the containing function.  But I can call defer mapWriteAndClose() at the top of a client pgm.
               And fixed help message regarding mapWriteAndClose, which is not automatic but needs to be deferred as I just wrote.
19 Jun 21 -- Changed MAP code so that it saves the file whenever writing or deleting, so don't need to call mapWriteAndClose directly anymore.
16 Sep 21 -- I increased the number of digits for the %g verb when output is dump'd.
 2 Nov 21 -- Adjusted dumpedfixed so that very large or very small numbers are output in general format to not mess up the display
 4 May 22 -- PIOVER6 never coded, so that's now added.  And I decided to use platform specific code for toclip/fromclip, contained in clippy.go.
               And stoclip and rclclip are now synonyms for toclip and fromclip.  But these only work at the command line,
               as sto and rcl are processed by rpng and rpnf without passing them to HPCALC2.
 7 May 22 -- Played a bit in clippy_linux.go, where I'm using make to initialize bytes.NewBuffer().
 7 Sep 22 -- Changed the pivot for the JUL command from the current year to a const of 30
21 Oct 22 -- golangci-lint says I have an unnecessary Sprintf call.  It's right.
21 Nov 22 -- static linter found a few more issues.
24 Jun 23 -- Will only close the map reg file when needed, ie, when I open it.  This is to not have rpnt and rpnf clobber each other.
               I'm not exporting the map close function, and I renamed it to mapWriteAndClose.  By not exporting it, I'm making sure that the client programs can't close this file and clobber one
               another.
               I decided to not read the map info in the init fcn, but instead I have to have it read the map file with every operation.  This is to prevent the local map from becoming stale.
 8 Jul 23 -- I'm testing the new TokenReal(), here in production.  Looks to be working.  I won't recompile the others just yet.  I'll try to shift into using rpn2 for a while.
 9 Jul 23 -- Now to fix hex input.  This is handled in tknptr.go.
20 Jul 23 -- Amended help text to show that to enter a negative exponent, must use '_' char.
15 Oct 23 -- Help doesn't report HCF (highest common factor), and I'll add a GCD synonym.
18 Oct 23 -- Writing the sieve code for the live project showed me that there's an off by 1 issue w/ my int sqrt routine.  I fixed it as I did in the sieve code.  It's like ceil(n).
23 Oct 23 -- Added the probably prime routines I learned about in the live project by Rod Stephens.  I'll add them to the prime command.
24 Oct 23 -- Added that the prime command will prime factor if the number is not prime.
28 Oct 23 -- Updated the message for the probably prime routine.
17 Dec 23 -- Updating probably prime routine to report how many guesses it took to say a number is probably not prime.  And added some comments.
 8 Jul 24 -- Adding the gcd routine as an alternative to hcf.
14 Aug 24 -- For the undo/redo, the matrix stack will be dynamically managed.
 9 Nov 24 -- Trying to figure out how to correct the small floating point errors that I see.  Next and Prev are manual.  I'm going to try using math.Floor.
23 Nov 24 -- Clean command was added several weeks ago.  It works.  I'm expanding it to allow a number, like fix does.
24 Nov 24 -- Improved some comments to make them clearer, and added more doc comments to the functions here
*/

const LastAlteredDate = "25 Nov 2024"

const HeaderDivider = "+-------------------+------------------------------+"
const SpaceFiller = "     |     "
const julPivot = 30

const (
	X = iota // StackRegNames as int.  No need for a separate type.
	Y
	Z
	T5
	T4
	T3
	T2
	T1
	StackSize
)

const Top = T1
const Bottom = X
const numOfFermatTests = 100 // Will set this at top of package so it's easy to change later.  // 20 gives a similar chance of being struck by lightning/yr, 30 gives better odds of winning powerball than the numbers not being prime.

var StackRegNamesString []string = []string{" X", " Y", " Z", "T5", "T4", "T3", "T2", "T1"}

// var FSATypeString []string = []string{"DELIM", "OP", "DGT", "AllElse"}  I am getting an unused variable, as the debugging statements are likely commented out.
var cmdMap map[string]int

type StackType [StackSize]float64

const mappedRegFilename = "mappedreg.gob"

type mappedRegStructType struct { // so the mapsho items can be sorted.
	key   string
	value float64
}

var mappedReg map[string]float64
var Stack StackType

// var StackUndoMatrix [StackSize]StackType
var StackUndoMatrix []StackType
var fullMappedRegFilename = homedir + string(os.PathSeparator) + mappedRegFilename

const PI = math.Pi // 3.141592653589793;
var LastX, MemReg float64
var sigfig = -1 // default significant figures of -1 for the strconv.FormatFloat call.
var homedir string
var mappedRegExists bool
var CurUndoRedoIdx int

const lb2g = 453.59238
const oz2g = 28.34952
const in2cm = 2.54
const m2ft = 3.28084
const mi2km = 1.609344
const veryLargeNumber = 1e10
const verySmallNumber = 1e-10

// -----------------------------------------------------------------------------------------------------------------------------
func init() {
	StackUndoMatrix = make([]StackType, 0, StackSize) // initial capacity of this is the same as it used to be.
	var err error
	cmdMap = make(map[string]int, 100)
	cmdMap["DUMP"] = 10
	cmdMap["DUMPFIX"] = 20
	cmdMap["DUMPFIXED"] = 20
	cmdMap["DUMPFLOAT"] = 30
	cmdMap["ADJ"] = 40
	cmdMap["ADJUST"] = 40
	cmdMap["NEXT"] = 50
	cmdMap["AFTER"] = 50
	cmdMap["BEFORE"] = 60
	cmdMap["PREV"] = 60
	cmdMap["PREVIOUS"] = 60
	cmdMap["SIG"] = 70
	cmdMap["SIGFIG"] = 70
	cmdMap["FIX"] = 70
	cmdMap["RECIP"] = 80
	cmdMap["CURT"] = 90
	cmdMap["CBRT"] = 90
	cmdMap["DIA"] = 100
	cmdMap["VOL"] = 110
	cmdMap["HELP"] = 120
	cmdMap["?"] = 120
	cmdMap["H"] = 120 // added 02/04/2021 8:54:30 AM
	cmdMap["STO"] = 130
	cmdMap["RCL"] = 135 // mistake -- was 130, so instead of renumbering all of it, I used my escape hatch
	cmdMap["UNDO"] = 140
	cmdMap["REDO"] = 150
	cmdMap["SWAP"] = 160
	cmdMap["~"] = 160
	cmdMap["`"] = 160 // that's a back tick
	cmdMap["LASTX"] = 170
	cmdMap["@"] = 170
	cmdMap["ROLLDN"] = 180
	cmdMap[","] = 190
	cmdMap["UP"] = 190
	cmdMap["|"] = 200
	cmdMap["DN"] = 200
	cmdMap["POP"] = 210
	cmdMap["INT"] = 215 // Missed this one on first pass thru assigning numbers
	cmdMap["PRIME"] = 220
	cmdMap["PRIMEF"] = 230
	cmdMap["PRIMEFA"] = 230
	cmdMap["PRIMEFAC"] = 230
	cmdMap["TRUNC"] = 240
	cmdMap["ROUND"] = 250
	cmdMap["CEIL"] = 260
	cmdMap["HEX"] = 270
	cmdMap["HCF"] = 280
	cmdMap["GCD"] = 285
	cmdMap["P"] = 290
	cmdMap["X"] = 290 // ignore these
	cmdMap["Q"] = 290
	cmdMap["FRAC"] = 300
	cmdMap["MOD"] = 310
	cmdMap["JUL"] = 320
	cmdMap["TODAY"] = 330
	cmdMap["T"] = 330
	cmdMap["GREG"] = 340
	cmdMap["DOW"] = 350
	cmdMap["PI"] = 360
	cmdMap["PIOVER6"] = 365
	cmdMap["CHS"] = 370
	cmdMap["_"] = 370
	cmdMap["HOL"] = 380
	cmdMap["ABOUT"] = 390
	cmdMap["SQR"] = 400
	cmdMap["SQRT"] = 410
	cmdMap["EXP"] = 420
	cmdMap["LOG"] = 430
	cmdMap["LN"] = 430
	cmdMap["SIN"] = 440
	cmdMap["COS"] = 450
	cmdMap["TAN"] = 460
	cmdMap["ARCSIN"] = 470
	cmdMap["ARCCOS"] = 480
	cmdMap["ARCTAN"] = 490
	cmdMap["D2R"] = 500
	cmdMap["R2D"] = 510
	cmdMap["TOCLIP"] = 520
	cmdMap["STOCLIP"] = 520
	cmdMap["FROMCLIP"] = 530
	cmdMap["RCLCLIP"] = 530
	cmdMap["LB2G"] = 540
	cmdMap["OZ2G"] = 550
	cmdMap["CM2IN"] = 560
	cmdMap["M2FT"] = 570
	cmdMap["MI2KM"] = 580
	cmdMap["G2LB"] = 590
	cmdMap["G2OZ"] = 600
	cmdMap["IN2CM"] = 610
	cmdMap["FT2M"] = 620
	cmdMap["KM2MI"] = 630
	cmdMap["C2F"] = 633
	cmdMap["F2C"] = 636
	cmdMap["MAP"] = 640 // mapsto, maprcl and mapsho are essentially subcommands of map.
	cmdMap["CLE"] = 650 // So I can process cleanN.  The code below is written for STO and RCL, so it uses the first 3 characters.
	cmdMap["CLEAN"] = 650
	cmdMap["CLEAN4"] = 650
	cmdMap["CLEAN5"] = 660

	/* commented out 6/16/21
	if runtime.GOOS == "linux" {
		homedir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		homedir = os.Getenv("userprofile")
	}
	*/
	homedir, err = os.UserHomeDir() // This func became available as of Go 1.12
	if err != nil {
		fmt.Fprintln(os.Stderr, " Error from os.UserHomeDir call is", err)
		os.Exit(1)
	}

	fullMappedRegFilename = homedir + string(os.PathSeparator) + mappedRegFilename

	mappedReg = make(map[string]float64, 100)
	mappedRegFile, er := os.Open(fullMappedRegFilename) // open for reading
	if os.IsNotExist(er) {
		mappedRegExists = false
	} else if er != nil {
		mappedRegExists = false
	} else {
		mappedRegExists = true
		mappedRegFile.Close()
	}

	//if mappedRegExists {
	//	defer mappedRegFile.Close()
	//	decoder := gob.NewDecoder(mappedRegFile) // decoder reads the file.
	//	err = decoder.Decode(&mappedReg)         // decoder reads the file.
	//	if err != nil {
	//		fmt.Fprintln(os.Stderr, err)
	//	}
	//}
} // init

// -----------------------------------------------------mapWriteAndClose --------------------------------------------------------------------

func mapWriteAndClose() {
	//fullmappedRegFilename := homedir + string(os.PathSeparator) + mappedRegFilename  this is global as of 6/23/23
	//mappedRegFile, err := os.OpenFile(fullMappedRegFilename, os.O_CREATE | os.O_APPEND, 0666)
	mappedRegFile, err := os.Create(fullMappedRegFilename) // open for writing.  6/24/23: I'm truncating the file each time I write it, but since I write the map from memory, this works.
	if err != nil {
		fmt.Fprintln(os.Stderr, "from os.Create", err)
	}
	defer mappedRegFile.Close()
	encoder := gob.NewEncoder(mappedRegFile) // encoder writes the file
	err = encoder.Encode(&mappedReg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "from gob encoder", err)
	}
} // end mapWriteAndClose

//------------------------------------------------------ ROUND ----------------------------------------------------------------------

func Round(f float64) float64 {
	sign := 1.0
	if math.Signbit(f) {
		sign = -1.0
	}
	result := math.Trunc(f + sign*0.5)
	return result
} // end Round

//------------------------------------------------------ STACKUP

func STACKUP() {
	for S := T2; S >= X; S-- {
		Stack[S+1] = Stack[S]
	}
} // STACKUP

//------------------------------------------------------ STACKDN

func STACKDN() {
	for S := Y; S < T1; S++ { // Does not affect X, so can do calculations and then remove Y and Z as needed.
		Stack[S] = Stack[S+1]
	}
} // STACKDN

//------------------------------------------------------ STACKROLLDN

func STACKROLLDN() {
	TEMP := Stack[X]
	Stack[X] = Stack[Y]
	STACKDN()
	Stack[T1] = TEMP
} // STACKROLLDN

// ----------------------------------------------------- PopX ---------------------

func PopX() float64 {
	x := Stack[X]
	for S := X; S < T1; S++ {
		Stack[S] = Stack[S+1]
	}
	return x
} // PopX

// ------------------------------------------------------ PUSHX

func PUSHX(R float64) {
	STACKUP()
	Stack[X] = R
}

//------------------------------------------------------ READX

func READX() float64 {
	return Stack[X]
} // READX

//------------------------------------------------------ SWAPXY

func SWAPXY() {
	Stack[X], Stack[Y] = Stack[Y], Stack[X]
} // SWAPXY

//------------------------------------------------------ GETSTACK

func GETSTACK() StackType {
	return Stack
} // GETSTACK

//-------------------------------------------------------------------- InsertByteSlice

func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//-----------------------------------------------------------------------------------------------------------------------------

func AddCommas(instr string) string {
	var i, decptposn int
	var Comma []byte = []byte{','}

	capstr := strings.ToUpper(instr)

	if strings.Contains(capstr, "E") {
		return instr
	}

	BS := make([]byte, 0, 100)
	//  outBS := make([]byte,0,100);
	decptposn = strings.LastIndex(instr, ".")
	BS = append(BS, instr...)

	if decptposn < 0 { // decimal point not found
		i = len(BS)
		BS = append(BS, '.')
	} else {
		i = decptposn
	}

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas

//-----------------------------------------------------------------------------------------------------------------------------

func CropNStr(instr string) string {

	//   A bug is if there is no decimal pt and there is a 0 in ones place, then that will no longer be
	//   removed.
	//   Another bug if in scientific notation.

	var outstr string
	var i int

	if strings.LastIndex(instr, "e") > 0 || strings.LastIndex(instr, "E") > 0 { // e char cannot be first char
		return instr
	}

	if strings.LastIndex(instr, ".") < 0 {
		return instr // ie, instr is unchanged.
	}
	upperbound := len(instr) - 1
	for i = upperbound; (i >= 0) && (instr[i] == '0'); i-- {
	} // looking for last non-zero character
	outstr = instr[:i+1]

	return strings.TrimSpace(outstr)
} // CropNStr

// DumpStackFixed -- Returns a slice of stack strings in fixed point format.
func DumpStackFixed() []string {
	//                                       var SRN int   I would never do this now, so I'm removing it
	var str string

	ss := make([]string, 0, StackSize+2)
	ss = append(ss, HeaderDivider)
	for SRN := T1; SRN >= X; SRN-- {
		if math.Abs(Stack[SRN]) < verySmallNumber || math.Abs(Stack[SRN]) > veryLargeNumber {
			str = strconv.FormatFloat(Stack[SRN], 'g', sigfig, 64)
			ss = append(ss, fmt.Sprintf("%2s: %10.2g %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))

		} else {
			str = strconv.FormatFloat(Stack[SRN], 'f', sigfig, 64)
			str = CropNStr(str)
			if Stack[SRN] > 10000 {
				str = AddCommas(str)
			}
			ss = append(ss, fmt.Sprintf("%2s: %10.2f %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))
		}
	}
	ss = append(ss, HeaderDivider) // call to sprintf was unneeded here.
	return ss
} // DumpStackFixed

// DumpStackFloat -- returns a slice of strings that were converted from float64 to string using FormatFloat.
func DumpStackFloat() []string {
	var SRN int
	var str string

	ss := make([]string, 0, StackSize+2)
	ss = append(ss, HeaderDivider)
	for SRN = T1; SRN >= X; SRN-- {
		str = strconv.FormatFloat(Stack[SRN], 'e', sigfig, 64)
		//		str = CropNStr(str)  makes no sense for numbers in exponential format
		//		if Stack[SRN] > 10000 {
		//			str = AddCommas(str)
		//		}
		ss = append(ss, fmt.Sprintf("%2s: %20.9e %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))
	}
	ss = append(ss, HeaderDivider)
	return ss
} // DumpStackFloat

// OutputFixedOrFloat -- Tries to output a number without an exponent, if possible.  Will output an exponent if it has to.
func OutputFixedOrFloat(r float64) {       //  Now only rpn.go (and probably rpn2.go) still uses this routine.
	if (r == 0) || math.Abs(r) < 1.0e-10 { // write 0.0
		fmt.Print("0.0")
	} else {
		str := strconv.FormatFloat(r, 'g', sigfig, 64) // when r >= 1e6 this switches to scientific notation.
		str = CropNStr(str)                            // bug here was caught by static linter.
		fmt.Print(str)
	}
} // OutputFixedOrFloat

// DumpStackGeneral -- Dumps the stack using a general format verb.
func DumpStackGeneral() []string {
	var SRN int
	var str string

	ss := make([]string, 0, StackSize+2)
	ss = append(ss, HeaderDivider)
	for SRN = T1; SRN >= X; SRN-- {
		str = strconv.FormatFloat(Stack[SRN], 'g', sigfig, 64)
		ss = append(ss, fmt.Sprintf("%2s: %10.7g %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))
	}
	ss = append(ss, HeaderDivider)
	return ss
} // DumpStackGeneral

// ToHex -- Uses an elegant algorithm I recently read about.
func ToHex(L float64) string {
	const hexDigits = "0123456789abcdef"

	IsNeg := false
	if L < 0 {
		IsNeg = true
		L = -L
	}
	// not changing sign of the value, so -1 may show as 0FFFFh, etc
	str := ""

	for h := int(math.Trunc(Round(L))); h > 0; h = h / 16 {
		d := h % 16                      // % is MOD op
		str = string(hexDigits[d]) + str // this line prepends the new digit with each iteration, so the result does not need to be reversed.
	}

	if IsNeg {
		return "Negative " + str + "H"
	}
	return str + "H"
} // ToHex

// IsPrime -- Just what its name says.
func IsPrime(real float64) bool { // The real input is to allow from stack.

	var t uint64 = 3
	var RoundSqrt uint64

	Uint := uint64(Round(math.Abs(real))) // just thoughts now, but will check in hpcalc

	if Uint == 0 || Uint == 1 {
		return false
	} else if Uint == 2 || Uint == 3 {
		return true
	} else if Uint%2 == 0 {
		return false
	}

	sqrt := math.Sqrt(real)
	RoundSqrt = uint64(Round(sqrt))

	for t <= RoundSqrt {
		if Uint%t == 0 {
			return false
		}
		t += 2
	}
	return true
} // IsPrime

// PrimeFactorMemoized -- returns a slice of uint prime factors using a memoization algorithm.
func PrimeFactorMemoized(U uint) []uint {

	if U == 0 {
		return nil
	}

	var val uint = 2

	PrimeUfactors := make([]uint, 0, 20)

	for u := U; u > 1; {
		fac, facflag := NextPrimeFac(u, val)

		if facflag {
			PrimeUfactors = append(PrimeUfactors, fac)
			u = u / fac
			val = fac
		} else { // this means that there are no more prime factors.
			PrimeUfactors = append(PrimeUfactors, u)
			break
		}
	}
	return PrimeUfactors
}

// NextPrimeFac -- Returns the next prime factor after the given one.  Supports the memoization algorithm.  Returns true when returning a prime factor.  False means there are no more prime factors.
func NextPrimeFac(n, startfac uint) (uint, bool) { // note that this is the reverse of IsPrime
	// This returns a prime factor or zero, and a bool which is true when the function returns a prime factor.
	// It will return false when there are no more prime factors.

	var t uint = startfac

	UintSqrt := usqrt(n)

	for t <= UintSqrt {
		if n%t == 0 {
			return t, true
		}
		if t == 2 {
			t = 3
		} else {
			t += 2
		}
	}
	return 0, false
} // IsPrime

// usqrt -- estimates a sqrt by dividing and averaging using a uint.
func usqrt(u uint) uint {

	sqrt := u / 2

	for i := 0; i < 30; i++ {
		guess := u / sqrt
		sqrt = (guess + sqrt) / 2
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}
	return sqrt + 1 // to fix an off by 1 issue I discovered by writing the Sieve code.
}

// PWRI -- Power function that multiplies the minimum number of times.  Exponent is an int that can be negative.
// Power Of I.
// This is a power function with a real base and integer exponent, using the optimized algorithm as discussed in PIM-2, V 2.
func PWRI(R float64, I int) float64 {
	Z := 1.0
	NEGFLAG := false
	if I < 0 {
		NEGFLAG = true
		I = -I
	}
	for I > 0 {
		if I%2 == 1 {
			Z *= R // Z = Z*R
		}
		R *= R // R = R squared
		I /= 2 // I = half I
	}
	if NEGFLAG {
		Z = 1 / Z
	}
	return Z
} // PWRI

// PushMatrixStacks -- Saves the stack state prior to an operation that will change the stack state.
func PushMatrixStacks() { // this is called from GetResult before an operation that would change the stack.
	//StacksMatrixUp()
	//StackUndoMatrix[Bottom] = Stack
	StackUndoMatrix = append(StackUndoMatrix, Stack)
	CurUndoRedoIdx = len(StackUndoMatrix) - 1
}

// UndoMatrixStacks -- performs an undo operation by reverting to a previous stack state.
func UndoMatrixStacks() {
	if CurUndoRedoIdx == len(StackUndoMatrix)-1 { // only push the current state if it's not already been pushed.  IE, first undo for this stack state.
		StackUndoMatrix = append(StackUndoMatrix, Stack)
	}
	if CurUndoRedoIdx <= len(StackUndoMatrix) && CurUndoRedoIdx > 0 {
		CurUndoRedoIdx--
		Stack = StackUndoMatrix[CurUndoRedoIdx]
	}
}

// Floor -- To automatically fix the small floating point errors introduced by the conversions.  Real can be negative, places cannot be negative or > 10.
func Floor(real, places float64) float64 { // written and debugged here before I put a copy if this in the misc package for the SQLite3 program to use.
	if places < 0 || places > 10 {
		places = 10
	}

	negFlag := real < 0
	result := real
	if negFlag {
		result *= -1
	}
	factor := math.Pow(10, places)
	result *= factor
	result = math.Floor(result + 0.5)
	result /= factor
	if negFlag {
		result *= -1
	}
	return result
}

// RedoMatrixStacks -- Redoes the stack matrix, that is, undoes the undo by reverting to a later stack state.
func RedoMatrixStacks() {
	if CurUndoRedoIdx >= 0 && CurUndoRedoIdx < len(StackUndoMatrix)-1 {
		CurUndoRedoIdx++
		Stack = StackUndoMatrix[CurUndoRedoIdx]
	}
}

//-------------------------------------------------------- HCF -------------------------------------

// HCF means highest common factor.
func HCF(a, b int) int {
	// a = bt + r, then hcf(a,b) = hcf(b,r)
	var r, a1, b1 int

	if a < b {
		a1 = b
		b1 = a
	} else {
		a1 = a
		b1 = b
	}
	for {
		r = a1 % b1 // % is MOD operator
		a1 = b1
		b1 = r
		if r == 0 {
			break
		}
	}
	return a1
} // HCF

//------------------------------------------------------------------------

// GCD means greatest common divisor, which is a synonym for HCF
func GCD(a, b int) int {
	// a = bt + r, then gcd(a,b) = gcd(b,r)

	var r int

	if a < b {
		a, b = b, a
	}

	for {
		r = a % b
		if r == 0 {
			break
		}
		a = b
		b = r
	}
	return b
}

// GetResult -- Input a string of commands and operations, return the result as a float64 and a message as a slice of strings.
func GetResult(s string) (float64, []string) {
	var token tknptr.TokenType
	var EOL bool
	var R float64
	var stringslice []string

	tokenPointer := tknptr.New(s) // Using the Go idiom, instead of INITKN(s)
	for {
		//token, EOL = tokenPointer.GETTKNREAL()
		token, EOL = tokenPointer.TokenReal() // here goes nothing
		if EOL {
			break
		}
		R, stringslice = Result(token)
	}
	return R, stringslice
}

func Result(tkn tknptr.TokenType) (float64, []string) {
	ss := make([]string, 0, 100) // ss is abbrev for stringslice.

outerloop:
	switch tkn.State {
	case tknptr.DELIM:

	case tknptr.DGT:
		PUSHX(tkn.Rsum)
		PushMatrixStacks()

	case tknptr.OP:
		I := tkn.Isum
		if (I == 6) || (I == 20) || (I == 1) || (I == 3) { // <>, ><, <, > will all SWAP
			SWAPXY()
		} else {
			LastX = Stack[X]
			PushMatrixStacks()
			switch I {
			case 5, 8: // allow = and + to both mean add.
				Stack[X] += Stack[Y]
			case 10:
				Stack[X] = Stack[Y] - Stack[X]
			case 12:
				Stack[X] *= Stack[Y]
			case 14:
				Stack[X] = Stack[Y] / Stack[X]
			case 16:
				Stack[X] = PWRI(Stack[Y], int(Round(Stack[X]))) // ^ op -> PWRI
			case 18:
				Stack[X] = math.Pow(Stack[Y], Stack[X]) // **
			case 22:
				Stack[X] *= Stack[Y] / 100.0 // percent
			default:
				ss = append(ss, fmt.Sprintf("%s is an unrecognized operation.", tkn.Str))
				STACKUP()
			}
			if I != 22 { // Do not move stack for % operator
				STACKDN()
			}
		} // opcode value condition
	case tknptr.ALLELSE:
		cmdnum := cmdMap[tkn.Str]
		if cmdnum == 0 && len(tkn.Str) > 2 {
			TokenStrShortened := tkn.Str[:3] // First 3 characters, ie, characters at positions 0, 1 and 2
			cmdnum = cmdMap[TokenStrShortened]
		}
		if cmdnum == 0 {
			ss = append(ss, fmt.Sprintf(" %s is an unrecognized command.", tkn.Str))
			break
		}

		switch cmdnum {
		case 10: // DUMP
			ss = append(ss, DumpStackGeneral()...)
		case 20: // DUMPFIX
			ss = append(ss, DumpStackFixed()...)
		case 30: // DUMPFLOAT
			ss = append(ss, DumpStackFloat()...)
		case 40: // ADJ or ADJUST
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] *= 100
			Stack[X] = Round(Stack[X])
			Stack[X] /= 100
		case 50: // NEXT, AFTER, for math.Nextafter to go up to a large number, here I use 1 billion.
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = math.Nextafter(LastX, 1e9) // correct up.
		case 60: // BEFORE, PREV, PREVIOUS, for math.Nextafter to go towards zero.
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = math.Nextafter(LastX, 0) // correct down.
		case 70: // SIGn, SIGFIGn, FIXn
			ch := tkn.Str[len(tkn.Str)-1] // ie, the last character.
			sigfig = GetRegIdx(ch)
			if sigfig > 9 { // If sigfig greater than this max value, make it -1 again.
				sigfig = -1
			}
		case 80: // RECIP
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = 1 / Stack[X]
		case 90: // CURT or CBRT
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = math.Cbrt(Stack[X])
		case 100: // DIA
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = math.Cbrt(Stack[X]) * 1.2407009817988 // constant is cube root of 6/Pi, so can multiply cube roots.
		case 110: // VOL
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = Stack[X] * Stack[Y] * Stack[Z] * PI / 6
			STACKDN()
			STACKDN()
		case 120: // HELP, H or ?
			ss = append(ss, " To enter a negative exponent, only '_' is allowed, which will be substituted for '-' before conversion to float64.")
			ss = append(ss, " SQRT,SQR -- X = sqrt(X) or sqr(X) register.")
			ss = append(ss, " CURT,CBRT -- X = cuberoot(X).")
			ss = append(ss, " RECIP -- X = 1/X.")
			ss = append(ss, " CHS,_ -- Change Sign,  X = -1 * X.")
			ss = append(ss, " DIA -- Given a volume in X, then X = estimated diameter for that volume, assuming a sphere.  Does not approximate Pi as 3.")
			ss = append(ss, " VOL -- Take values in X, Y, And Z and return a volume in X.  Does not approximate Pi as 3.")
			ss = append(ss, " TOCLIP, FROMCLIP, STOCLIP, RCLCLIP -- uses xclip on linux and tcc on Windows to access the clipboard.")
			ss = append(ss, " STO,RCL  -- store/recall the X register to/from the memory register.")
			ss = append(ss, " `,~,SWAP,SWAPXY,<>,><,<,> -- equivalent commands that swap the X and Y registers.")
			ss = append(ss, " @, LastX -- put the value of the LASTX register back into the X register.")
			ss = append(ss, " , comma -- stack up.  | vertical bar -- stack down.")
			ss = append(ss, " Pop -- displays X and then moves stack down.")
			ss = append(ss, " Dump, Dumpfixed, Dumpfloat, Sho -- dump the stack to the terminal.")
			ss = append(ss, " EXP,LN,LOG -- evaluate exp(X) or ln(X) and put result back into X.")
			ss = append(ss, " ^  -- Y to the X power using PWRI, put result in X and pop stack 1 reg.  Rounds X")
			ss = append(ss, " **  -- Y to the X power, put result in X and pop stack 1 reg, using Pow()")
			ss = append(ss, " INT, TRUNC, ROUND, CEIL, FRAC, PI, PIOVER6 -- do what their names suggest.")
			ss = append(ss, " MOD -- evaluate Y MOD X, put result in X and pop stack 1 reg.")
			ss = append(ss, " %   -- does XY/100, places result in X.  Leaves Y alone.")
			ss = append(ss, " SIN,COS,TAN,ARCTAN,ARCSIN,ARCCOS -- In deg.")
			ss = append(ss, " D2R, R2D -- perform degrees <--> radians conversion of the X register.")
			ss = append(ss, fmt.Sprintf(" JUL -- Return Julian date number of Z month, Y day, X year.  Pop stack x2.  Pivot is %d for 2 digit years.", julPivot))
			ss = append(ss, " TODAY, T -- Return Julian date number of today's date.  Pop stack x2.")
			ss = append(ss, " GREG-- Return Z month, Y day, X year of Julian date number in X.")
			ss = append(ss, " DOW -- Return day number 0..6 of julian date number in X register.")
			ss = append(ss, " HEX -- Round X register to a long_integer and output it in hex format.")
			ss = append(ss, " HCF -- Displays highest common factor for rounded X and Y.  GCD for greatest common denominator is a synonym.")
			ss = append(ss, " HOL -- Display holidays.")
			ss = append(ss, " UNDO, REDO -- entire stack.  More comprehensive than lastx.")
			ss = append(ss, " Prime, PrimeFactors -- evaluates X.")
			ss = append(ss, " Adjust -- X reg *100, Round, /100")
			ss = append(ss, " NextAfter,Before,Prev -- Reference factor for the fcn is 1e9 or 0.")
			ss = append(ss, " clean, clean4, clean5, cleanN -- Automatically correct the small floating point errors to 4, 5 or N decimal places.")
			ss = append(ss, " SigFigN,FixN -- Set the significant figures to N for the stack display string.  Default is -1.")
			ss = append(ss, " substitutions: = for +, ; for *.")
			ss = append(ss, " lb2g, oz2g, cm2in, m2ft, mi2km, c2f and their inverses -- unit conversions.")
			ss = append(ss, " mapsho, mapsto, maprcl, mapdel -- mappedReg commands.  mapWriteAndClose needs to be deferred after")
			ss = append(ss, "                                   first use of PushMatrixStacks.  !`~ become spaces in the name.")
			ss = append(ss, fmt.Sprintf(" last altered hpcalc2 %s.\n", LastAlteredDate))
		case 130: // STO
			MemReg = Stack[X]
		case 135: // RCL
			PUSHX(MemReg)
		case 140: // UNDO
			UndoMatrixStacks()
		case 150: // REDO
			RedoMatrixStacks()
		case 160: // SWAP or ~ or backtick; I removed SWAPXY
			PushMatrixStacks()
			SWAPXY()
		case 170: // LASTX or @
			PushMatrixStacks()
			PUSHX(LastX)
		case 180: // StackRolldn(), not StackDn()
			PushMatrixStacks()
			STACKROLLDN()
		case 190: // UP
			PushMatrixStacks()
			STACKUP()
		case 200: // DN or |
			PushMatrixStacks()
			Stack[X] = Stack[Y]
			STACKDN()
		case 210: // POP
			PushMatrixStacks()
			x := PopX()
			str := strconv.FormatFloat(x, 'g', sigfig, 64)
			ss = append(ss, str)
		case 215: // INT
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Floor(Stack[X])
		case 220: // PRIME
			n := Round(Stack[X])
			i := int(n)
			var primeFlag bool

			num, probably := isProbablyPrime(i, numOfFermatTests)
			if probably {
				ss = append(ss, fmt.Sprintf("%d is probably prime using %d tests.", i, num))
			} else {
				ss = append(ss, fmt.Sprintf("%d is NOT probably prime using %d tests.", i, num))
			}

			if IsPrime(n) {
				ss = append(ss, fmt.Sprintf("%d is prime.", i))
				primeFlag = true
			} else {
				ss = append(ss, fmt.Sprintf("%d is NOT prime.", i))
			}
			if !primeFlag {
				u := uint(i)
				if u < 2 {
					ss = append(ss, "PrimeFactors cmd of numbers < 2 ignored.")
				} else {
					PrimeUfactors := PrimeFactorMemoized(u)
					stringslice := make([]string, 0, 10)

					for _, pf := range PrimeUfactors {
						stringslice = append(stringslice, fmt.Sprintf("%d", pf))
					}
					ss = append(ss, strings.Join(stringslice, ", "))
				}
			}

		case 230: // PRIMEFAC, PRIMEF or PRIMEFA  Intended for PrimeFactors or PrimeFactorization
			U := uint(Round(Stack[X]))
			if U < 2 {
				ss = append(ss, "PrimeFactors cmd of numbers < 2 ignored.")
			} else {

				PrimeUfactors := PrimeFactorMemoized(U)
				stringslice := make([]string, 0, 10)

				for _, pf := range PrimeUfactors {
					stringslice = append(stringslice, fmt.Sprintf("%d", pf))
				}
				ss = append(ss, strings.Join(stringslice, ", "))
			}
		case 240: // TRUNC
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Trunc(Stack[X])
		case 250: // ROUND
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = Round(LastX)
		case 260: // CEIL
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Ceil(LastX)
		case 270: // HEX
			if (Stack[X] >= -2.0e9) && (Stack[X] <= 1.80e19) {
				ss = append(ss, fmt.Sprintf(" Value of X reg in hex: %s", ToHex(Stack[X])))
			} else {
				ss = append(ss, " Cannot convert X register to hex string, as number is out of range.") // use of Sprintf here was not needed.  Caught by the golangci-lint.
			} // Hex command
		case 280: // HCF
			c1 := int(math.Abs(Round(Stack[X])))
			c2 := int(math.Abs(Round(Stack[Y])))
			c := HCF(c2, c1)
			ss = append(ss, fmt.Sprintf("HCF of %d and %d is %d.", c1, c2, c))
		case 285: // GCD
			c1 := int(math.Abs(Round(Stack[X])))
			c2 := int(math.Abs(Round(Stack[Y])))
			c := GCD(c2, c1)
			PUSHX(float64(c))
			ss = append(ss, fmt.Sprintf("GCD of %d and %d is %d.", c1, c2, c))
		case 290: // P, Q, X
			//  essentially do nothing but print RESULT= line again.
		case 300: // FRAC
			PushMatrixStacks()
			LastX = Stack[X]
			_, frac := math.Modf(Stack[X])
			Stack[X] = frac
		case 310: // MOD
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Mod(Round(Stack[Y]), Round(Stack[X]))
			STACKDN()
		case 320: // JUL
			PushMatrixStacks()
			LastX = Stack[X]
			// allow for 2 digit years
			//_, _, year := timlibg.TIME2MDY()
			//if Stack[X] <= float64(year%100) { // changed the pivot for 2 digit years Sep 7, 2022
			if Stack[X] <= julPivot { // % is the MOD operator.  The pivot is 30 as of this writing.
				Stack[X] += 2000.0
			} else if Stack[X] < 100.0 {
				Stack[X] += 1900.0
			}
			Stack[X] = float64(timlibg.JULIAN(int(Round(Stack[Z])), int(Round(Stack[Y])), int(Round(Stack[X]))))
			STACKDN()
			STACKDN()
		case 330: // TODAY or T
			PushMatrixStacks()
			LastX = Stack[X]
			STACKUP()
			c1, c2, c3 := timlibg.TIME2MDY()
			Stack[X] = float64(timlibg.JULIAN(c1, c2, c3))
		case 340: // GREG
			PushMatrixStacks()
			LastX = Stack[X]
			STACKUP()
			STACKUP()
			c1, c2, c3 := timlibg.GREGORIAN(int(Round(Stack[X])))
			Stack[Z] = float64(c1)
			Stack[Y] = float64(c2)
			Stack[X] = float64(c3)
		case 350: // DOW
			dow := int(Round(Stack[X]))
			i := dow % 7 // % is the MOD operator only for int's
			s := fmt.Sprintf(" Day of Week for %d is a %s", dow, timlibg.DayNames[i])
			ss = append(ss, s)
		case 360: // PI
			PushMatrixStacks()
			PUSHX(PI)
		case 365: // PIOVER6
			PushMatrixStacks()
			PUSHX(PI / 6)
		case 370: // CHS or _
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = -1 * Stack[X]
		case 380: // HOL
			// PushMatrixStacks()  Doesn't change the stack.  I don't think it ever did.
			year := int(Round(Stack[X]))
			if year < 40 {
				year += 2000
			} else if year < 100 {
				year += 1900
			}
			if (year >= 1900) && (year <= 2100) {
				Holiday := holidaycalc.GetHolidays(year)
				Holiday.Valid = true
				ss = append(ss, fmt.Sprintf(" For year %d:", Holiday.Year))
				Y := Holiday.Year
				NYD := timlibg.JULIAN(1, 1, Y) % 7
				ss = append(ss, fmt.Sprintf("New Years Day is a %s, MLK Day is January %d, Pres Day is February %d, Easter Sunday is %s %d, Mother's Day is May %d",
					timlibg.DayNames[NYD], Holiday.MLK.D, Holiday.Pres.D, timlibg.MonthNames[Holiday.Easter.M], Holiday.Easter.D, Holiday.Mother.D))

				July4 := timlibg.JULIAN(7, 4, Y) % 7
				ss = append(ss, fmt.Sprintf("Memorial Day is May %d, Father's Day is June %d, July 4 is a %s, Labor Day is Septempber %d, Columbus Day is October %d",
					Holiday.Memorial.D, Holiday.Father.D, timlibg.DayNames[July4], Holiday.Labor.D, Holiday.Columbus.D))

				VetD := timlibg.JULIAN(11, 11, Y) % 7
				ChristmasD := timlibg.JULIAN(12, 25, Y) % 7
				ss = append(ss, fmt.Sprintf("Election Day is November %d, Veteran's Day is a %s, Thanksgiving is November %d, and Christmas Day is a %s.",
					Holiday.Election.D, timlibg.DayNames[VetD], Holiday.Thanksgiving.D, timlibg.DayNames[ChristmasD]))
			} else { // added 1/22/20.
				s := " X register is not a valid year.  Command ignored."
				ss = append(ss, s)
			}
		case 390: // ABOUT
			ss = append(ss, fmt.Sprintf(" last changed hpcalc2.go %s", LastAlteredDate))
		case 400: // SQR
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] *= Stack[X]
		case 410: // SQRT
			LastX = Stack[X]
			PushMatrixStacks()
			Stack[X] = math.Sqrt(Stack[X])
		case 420: // EXP
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Exp(Stack[X])
		case 430: // LOG or LN
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Log(math.Abs(Stack[X]))
		case 440: // SIN
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Sin(Stack[X] * PI / 180.0)
		case 450: // COS
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Cos(Stack[X] * PI / 180.0)
		case 460: // TAN
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Tan(Stack[X] * PI / 180.0)
		case 470: // ARCSIN
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Asin(LastX) * 180.0 / PI
		case 480: // ARCCOS
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Acos(Stack[X]) * 180.0 / PI
		case 490: // ARCTAN
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] = math.Atan(LastX) * 180.0 / PI
		case 500: // D2R
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] *= PI / 180.0
		case 510: // R2D
			PushMatrixStacks()
			LastX = Stack[X]
			Stack[X] *= 180.0 / PI
		case 520: // TOCLIP, now in platform specific code
			msg := toClip(READX())
			ss = append(ss, msg)

			//R := READX()
			//s := strconv.FormatFloat(R, 'g', -1, 64)
			//if runtime.GOOS == "linux" {
			//	linuxclippy := func(s string) {
			//		buf := []byte(s)
			//		rdr := bytes.NewReader(buf)
			//		cmd := exec.Command("xclip")
			//		cmd.Stdin = rdr
			//		cmd.Stdout = os.Stdout
			//		cmd.Run()
			//		ss = append(ss, fmt.Sprintf(" Sent %s to xclip.", s))
			//	}
			//	linuxclippy(s)
			//} else if runtime.GOOS == "windows" {
			//	comspec, ok := os.LookupEnv("ComSpec")
			//	if !ok {
			//		ss = append(ss, " Environment does not have ComSpec entry.  ToClip unsuccessful.")
			//		break outerloop
			//	}
			//	winclippy := func(s string) {
			//		//cmd := exec.Command("c:/Program Files/JPSoft/tcmd22/tcc.exe", "-C", "echo", s, ">clip:")
			//		cmd := exec.Command(comspec, "-C", "echo", s, ">clip:")
			//		cmd.Stdout = os.Stdout
			//		cmd.Run()
			//		ss = append(ss, fmt.Sprintf(" Sent %s to %s.", s, comspec))
			//	}
			//	winclippy(s)
			//}

		case 530: // FROMCLIP, now in platform specific code
			PushMatrixStacks()
			LastX = Stack[X]
			f, msg, err := fromClip()
			if err == nil {
				PUSHX(f)
			} else {
				msg = fmt.Sprintf("%s  Error fromClip is %v.", msg, err)
			}
			ss = append(ss, msg)

			//w := bytes.NewBuffer([]byte{}) // From "Go Standard Library Cookbook" as referenced above.
			//if runtime.GOOS == "linux" {
			//	cmdfromclip := exec.Command("xclip", "-o")
			//	cmdfromclip.Stdout = w
			//	cmdfromclip.Run()
			//	str := w.String()
			//	s := fmt.Sprintf(" Received %s from xclip.", str)
			//	str = strings.ReplaceAll(str, "\n", "")
			//	str = strings.ReplaceAll(str, "\r", "")
			//	str = strings.ReplaceAll(str, ",", "")
			//	str = strings.ReplaceAll(str, " ", "")
			//	s = s + fmt.Sprintf("  After removing all commas and spaces it becomes %s.", str)
			//	ss = append(ss, s)
			//	R, err := strconv.ParseFloat(str, 64)
			//	if err != nil {
			//		ss = append(ss, fmt.Sprintln(" fromclip on linux conversion returned error", err, ".  Value ignored."))
			//	} else {
			//		PUSHX(R)
			//	}
			//} else if runtime.GOOS == "windows" {
			//	comspec, ok := os.LookupEnv("ComSpec")
			//	if !ok {
			//		ss = append(ss, " Environment does not have ComSpec entry.  FromClip unsuccessful.")
			//		break outerloop
			//	}
			//
			//	cmdfromclip := exec.Command(comspec, "-C", "echo", "%@clip[0]")
			//	cmdfromclip.Stdout = w
			//	cmdfromclip.Run()
			//	lines := w.String()
			//	s := fmt.Sprint(" Received ", lines, "from ", comspec)
			//	linessplit := strings.Split(lines, "\n")
			//	str := strings.ReplaceAll(linessplit[1], "\"", "")
			//	str = strings.ReplaceAll(str, "\n", "")
			//	str = strings.ReplaceAll(str, "\r", "")
			//	str = strings.ReplaceAll(str, ",", "")
			//	str = strings.ReplaceAll(str, " ", "")
			//	s = s + fmt.Sprintln(", after post processing the string becomes", str)
			//	ss = append(ss, s)
			//	R, err := strconv.ParseFloat(str, 64)
			//	if err != nil {
			//		ss = append(ss, fmt.Sprintln(" fromclip", err, ".  Value ignored."))
			//	} else {
			//		PUSHX(R)
			//	}
			//}

		case 540: // lb2g = 453.59238
			r := READX() * lb2g
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s1 := fmt.Sprintf("%s pounds is %s grams", x, s0)
			ss = append(ss, s1)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 550: // oz2g = 28.34952
			r := READX() * oz2g
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s := fmt.Sprintf("%s oz is %s grams", x, s0)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 560: // cm2in = 2.54
			r := READX() / in2cm
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s := fmt.Sprintf("%s cm is %s inches", x, s0)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 570: // m2ft = 3.28084
			r := READX() * m2ft
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s := fmt.Sprintf("%s meters is %s feet", x, s0)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 580: // mi2km = 1.609344
			r := READX() * mi2km
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s := fmt.Sprintf("%s miles is %s km", x, s0)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 590: // g2lb
			r := READX() / lb2g
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s1 := fmt.Sprintf("%s grams is %s pounds", x, s0)
			ss = append(ss, s1)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 600: // g2oz
			r := READX() / oz2g
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s1 := fmt.Sprintf("%s grams is %s oz", x, s0)
			ss = append(ss, s1)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 610: //in2cm
			r := READX() * in2cm
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s1 := fmt.Sprintf("%s inches is %s cm", x, s0)
			ss = append(ss, s1)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 620: // ft2m
			r := READX() / m2ft
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s := fmt.Sprintf("%s ft is %s meters", x, s0)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 630: // km2mi
			r := READX() / mi2km
			x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
			s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
			s := fmt.Sprintf("%s km is %s mi", x, s0)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(r)

		case 633: // C2F
			x := READX()
			xstr := strconv.FormatFloat(x, 'f', sigfig, 64)
			fdeg := x*1.8 + 32
			fstr := strconv.FormatFloat(fdeg, 'f', sigfig, 64)
			s := fmt.Sprintf("%s deg C is %s deg F", xstr, fstr)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(fdeg)

		case 636: // F2C
			x := READX()
			xstr := strconv.FormatFloat(x, 'f', sigfig, 64)
			cdeg := (x - 32) / 1.8
			cstr := strconv.FormatFloat(cdeg, 'f', sigfig, 64)
			s := fmt.Sprintf("%s deg F is %s deg C", xstr, cstr)
			ss = append(ss, s)
			PushMatrixStacks()
			LastX = Stack[X]
			PUSHX(cdeg)

		case 640: // map.   Now to deal w/ subcommands mapsto, maprcl, mapdel and mapsho, etc.
			// Will read a fresh copy of the map reg file because I want it to not be stale.
			mappedRegFile, err := os.Open(fullMappedRegFilename) // open for reading
			if os.IsNotExist(err) {
				mappedRegExists = false
			} else if err != nil {
				mappedRegExists = false
				fmt.Printf(" Error from opening %s is %s\n", fullMappedRegFilename, err)
			} else {
				mappedRegExists = true
			}

			if mappedRegExists {
				mappedReg = nil
				decoder := gob.NewDecoder(mappedRegFile) // decoder reads the file.
				err := decoder.Decode(&mappedReg)        // decoder reads the file.
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				mappedRegFile.Close() // I need to close this file now, and not defer the close.  Else it may interfere when I write the file, since I now do a read and write in the same block.
			}
			subcmd := tkn.Str[3:] // slice off first three characters, which are map
			if strings.HasPrefix(subcmd, "STO") {
				regName := getMapRegName(subcmd)
				//                                             fmt.Println(" in mapsto section.  regname=", regname)
				if regName == "" {
					ss = append(ss, "mapsto needs a register label.  None found so command ignored.")
					break outerloop
				}
				mappedReg[regName] = READX()
				mapWriteAndClose() // added 6/19/21
				_, stringresult := GetResult("mapsho")
				ss = append(ss, stringresult...)

			} else if strings.HasPrefix(subcmd, "RCL") {
				if !mappedRegExists {
					ss = append(ss, "No Mapped Registers file exists.")
					break outerloop
				}
				regName := getMapRegName(subcmd)
				//                                             fmt.Println(" in maprcl section.  regname=", regname)
				if regName == "" {
					ss = append(ss, "maprcl needs a register label.  None found so command ignored.")
					break outerloop
				}
				r, ok := mappedReg[regName]
				if ok {
					PUSHX(r)
				} else { // call the abbreviation processing routine, that I have yet to write.
					name := getFullMatchingName(regName)
					if name == "" {
						s := fmt.Sprintf("register label %s not found in maprcl cmd.  Command ignored.", regName)
						ss = append(ss, s)
						break outerloop
					}
					r := mappedReg[name]
					PUSHX(r)
				}

			} else if strings.HasPrefix(subcmd, "SHO") || strings.HasPrefix(subcmd, "LS") || strings.HasPrefix(subcmd, "LIST") {
				if !mappedRegExists {
					ss = append(ss, "No Mapped Registers file exists.")
					break outerloop
				}
				// maybe sort this list in a later version of this code.  And maybe allow option to only show mappedReg specified in this subcmd.
				s0 := fmt.Sprint("Map length is ", len(mappedReg))
				ss = append(ss, s0)
				sliceRegVar := mappedRegSortedNames()

				for _, reg := range sliceRegVar {
					fmtValue := strconv.FormatFloat(reg.value, 'g', sigfig, 64)
					s := fmt.Sprintf("reg[%s] = %s", reg.key, fmtValue)
					ss = append(ss, s)
				}

			} else if strings.HasPrefix(subcmd, "DEL") {
				if !mappedRegExists {
					ss = append(ss, "No Mapped Registers file exists.")
					break outerloop
				}
				regName := getMapRegName(subcmd)
				if regName == "" {
					ss = append(ss, "mapdel needs a register label.  None found so command ignored.")
					break outerloop
				}
				delete(mappedReg, regName) // if key is not in the map, this does nothing but does not panic.
				s := fmt.Sprint("deleted ", regName)
				ss = append(ss, s)
				mapWriteAndClose() // added 6/19/21
				_, stringresult := GetResult("mapsho")
				ss = append(ss, stringresult...)

			}

		case 650: // CLEAN and CLEAN4.  As of 11/23/24, it's really just all CLEAN now.  More precisely, CLE followed by last character determines value of N.
			PushMatrixStacks()
			LastX = Stack[X]
			ch := tkn.Str[len(tkn.Str)-1] // ie, the last character.
			places := GetRegIdx(ch)
			placesReal := float64(places)
			if places > 10 { // If greater than this max value, make it 10, ie default is 10.
				placesReal = float64(10)
			}
			x := Floor(Stack[X], placesReal) // this is to correct the small floating point errors, to 4 decimal places.  Floor is my function, defined above.
			Stack[X] = x
			s := fmt.Sprintf("clean(x, %d) done", places)
			ss = append(ss, s)
		case 660: // CLEAN5  this is redundant now.  I'm not deleting it.
			PushMatrixStacks()
			LastX = Stack[X]
			x := Floor(Stack[X], 5) // this is to correct the small floating point errors, to 5 decimal places.  Floor is my function that's defined above.
			Stack[X] = x
			s := "clean5 done"
			ss = append(ss, s)

		case 999: // do nothing, ignore me but don't generate an error message.

		default:
			ss = append(ss, fmt.Sprintf(" %s is an unrecognized command.  And should not get here.", tkn.Str))
		} // main text command selection if statement
	}
	return Stack[X], ss
} // Result

// ----------------------------------------------------------- getMapRegName --------------------------------------------
func getMapRegName(cmd string) string {
	if len(cmd) < 4 {
		return ""
	}
	sub := cmd[3:] // slice off first three characters, which are the subcmd sto, rcl or sho
	inspected := string(sub[0])
	mappedregname := sub
	if strings.ContainsAny(inspected, "!~`") { // if first char represents a space, lop it off
		mappedregname = sub[1:]
	}

	mappedregname = MakeSubst(mappedregname)         // changes ~ ! and ` to a space
	mappedregname = strings.TrimSpace(mappedregname) // if there are any spaces at beginning or end, trim them off
	return mappedregname
}

// ------------------------------------------------------------ MakeSubst -----------------------------------------------

func MakeSubst(instr string) string {
	// substitute ! ~ ` chara for spaces.  Copied from rpntcell
	instr = strings.TrimSpace(instr)
	inRune := make([]rune, len(instr))

	for i, s := range instr {
		switch s {
		case '!', '`', '~':
			s = ' '
		}
		inRune[i] = s
	}
	return string(inRune)
}

// ------------------------------------------------------------ GetRegIdx ---------

func GetRegIdx(chr byte) int {
	/* Return 0..35 w/ A = 10 and Z = 35.  Copied from main. */

	ch := tknptr.CAP(chr)
	if (ch >= '0') && (ch <= '9') {
		ch = ch - '0'
	} else if (ch >= 'A') && (ch <= 'Z') {
		ch = ch - 'A' + 10
	}
	return int(ch)
} // GetRegIdx

// ---------------------------------------------------------- mappedRegSortedNames ---------------------
func mappedRegSortedNames() []mappedRegStructType {
	sliceregvar := make([]mappedRegStructType, 0, 50)
	for key, value := range mappedReg {
		m := mappedRegStructType{key, value} // using structured literal syntax.
		sliceregvar = append(sliceregvar, m)
	}
	sortlessfunction := func(i, j int) bool {
		return sliceregvar[i].key < sliceregvar[j].key
	}
	sort.Slice(sliceregvar, sortlessfunction)

	return sliceregvar
}

// ----------------------------------------------------------- getFullMatchingName -----------------------
func getFullMatchingName(abbrev string) string {
	sliceregvar := mappedRegSortedNames()
	for _, name := range sliceregvar {
		if strings.Contains(name.key, abbrev) {
			return name.key
		}
	}
	return ""
}

// ----------------------------------------------------------- SigFig --------------------------------------

func SigFig() int {
	return sigfig
}

// ----------------------------------------------------------- isProbablyPrime -----------------------------

func isProbablyPrime(p int, numTests int) (int, bool) {
	if p%2 == 0 {
		return 0, false
	}
	if p == 1 {
		return 0, false
	}
	// Run numTests number of Fermat's little theorem.  For any that fail, return false, if all succeed return true.  These are fast to check, so having numTests of 30, 50 or 100 will be fast.
	for i := 0; i < numTests; i++ {
		n := randRange(p/3, p)
		expMod := fastExpMod(n, p-1, p)
		//fmt.Printf(" n=%d, p=%d, ExpMod = %d\n", n, p, expMod)
		if expMod != 1 {
			return i + 1, false // so if the first test fails, it returns 1 instead of zero.
		}
	}
	return numTests, true
}

// ------------------------------------------------------------- fastExpMod ------------------------------

func fastExpMod(num, pow, mod int) int { // pow can't be negative, or else it will panic.
	Z := 1
	if pow < 0 || mod < 0 {
		s := fmt.Sprintf("fastExpMod pow or mod cannot be negative.  pow = %d, mod = %d", pow, mod)
		panic(s)
	}
	for pow > 0 {
		if pow%2 == 1 { // ie, if pow is odd
			Z = (Z * num) % mod // Z = Z * R
		}
		num = (num * num) % mod // R = R squared
		pow /= 2                // I = half I
	}
	return Z //% mod
}

// --------------------------------------------------------------- randRange ----------------------------

func randRange(minP, maxP int) int { // note that this is not cryptographically secure.  Writing a cryptographically secure pseudorandom number generator (CSPRNG) is beyond the scope of this exercise.
	return minP + rand.Intn(maxP-minP)
}
