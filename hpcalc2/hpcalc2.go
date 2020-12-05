package hpcalc2

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	//
	"holidaycalc"
	"timlibg"
	//	"tokenize"
	"tknptr"
)

const LastAlteredDate = "5 Dec 2020"

/* (C) 1990.  Robert W Solomon.  All rights reserved.
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
*/

const HeaderDivider = "+-------------------+------------------------------+"
const SpaceFiller = "     |     "

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
const Bottom = 0

var StackRegNamesString []string = []string{" X", " Y", " Z", "T5", "T4", "T3", "T2", "T1"}

//var FSATypeString []string = []string{"DELIM", "OP", "DGT", "AllElse"}  I am getting an unused variable, as the debugging statements are likely commented out.
var cmdMap map[string]int

type StackType [StackSize]float64

var Stack StackType
var StackUndoMatrix [StackSize]StackType

const PI = math.Pi // 3.141592653589793;
var LastX, MemReg float64
var sigfig = -1 // default significant figures of -1 for the strconv.FormatFloat call.

const lb2g = 453.59238
const oz2g = 28.34952
const cm2in = 2.54
const m2ft = 3.28084
const mi2km = 1.609344

//-----------------------------------------------------------------------------------------------------------------------------
func init() {
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
	cmdMap["P"] = 290
	cmdMap["FRAC"] = 300
	cmdMap["MOD"] = 310
	cmdMap["JUL"] = 320
	cmdMap["TODAY"] = 330
	cmdMap["T"] = 330
	cmdMap["GREG"] = 340
	cmdMap["DOW"] = 350
	cmdMap["PI"] = 360
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
	cmdMap["FROMCLIP"] = 530
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
}

//const lb2g = 453.59238
//const oz2g = 28.34952
//const m2ft = 3.28084
//const cm2in = 2.54
//const mi2km = 1.609344

//------------------------------------------------------ ROUND ----------------------------------------------------------------------
func Round(f float64) float64 {
	sign := 1.0
	if math.Signbit(f) {
		sign = -1.0
	}
	result := math.Trunc(f + sign*0.5)
	return result
}

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

//------------------------------------------------------------------ DumpStackFixed -----------------------------------------------------------
func DumpStackFixed() []string {
	var SRN int
	var str string

	ss := make([]string, 0, StackSize+2)
	ss = append(ss, fmt.Sprintf("%s", HeaderDivider))
	for SRN = T1; SRN >= X; SRN-- {
		str = strconv.FormatFloat(Stack[SRN], 'f', sigfig, 64)
		str = CropNStr(str)
		if Stack[SRN] > 10000 {
			str = AddCommas(str)
		}
		ss = append(ss, fmt.Sprintf("%2s: %10.2f %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))
	}
	ss = append(ss, fmt.Sprintf("%s", HeaderDivider))
	return ss
} // DumpStackFixed
// ************************************************* DumpStackFloat **************************
func DumpStackFloat() []string {
	var SRN int
	var str string

	ss := make([]string, 0, StackSize+2)
	ss = append(ss, fmt.Sprintf("%s", HeaderDivider))
	for SRN = T1; SRN >= X; SRN-- {
		str = strconv.FormatFloat(Stack[SRN], 'e', sigfig, 64)
		//		str = CropNStr(str)  makes no sense for numbers in exponential format
		//		if Stack[SRN] > 10000 {
		//			str = AddCommas(str)
		//		}
		ss = append(ss, fmt.Sprintf("%2s: %20.9e %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))
	}
	ss = append(ss, fmt.Sprintf("%s", HeaderDivider))
	return ss
} // DumpStackFloat
//************************************************* OutputFixedOrFloat *******************************
func OutputFixedOrFloat(r float64) { // All output is thru a string slice.  Now only rpn.go still uses this routine.
	if (r == 0) || math.Abs(r) < 1.0e-10 { // write 0.0
		fmt.Print("0.0")
	} else {
		str := strconv.FormatFloat(r, 'g', sigfig, 64) // when r >= 1e6 this switches to scientific notation.
		CropNStr(str)
		fmt.Print(str)
	}
} // OutputFixedOrFloat
//************************************************** DumpStackGeneral ***************************
func DumpStackGeneral() []string {
	var SRN int
	var str string

	ss := make([]string, 0, StackSize+2)
	ss = append(ss, fmt.Sprintf("%s", HeaderDivider))
	for SRN = T1; SRN >= X; SRN-- {
		str = strconv.FormatFloat(Stack[SRN], 'g', sigfig, 64)
		//		str = CropNStr(str)  makes no sense for numbers in exponential format
		//		if Stack[SRN] > 10000 {
		//			str = AddCommas(str)
		//		}
		ss = append(ss, fmt.Sprintf("%2s: %10.4g %s %s", StackRegNamesString[SRN], Stack[SRN], SpaceFiller, str))
	}
	ss = append(ss, fmt.Sprintf("%s", HeaderDivider))
	return ss
} // DumpStackGeneral

//------------------------------------------------- ToHex ------------------
// The new algorithm is elegantly simple.
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
		d := h % 16 // % is MOD op
		str = string(hexDigits[d]) + str
	} // until L = 0

	if IsNeg {
		return "Negative " + str + "H"
	}
	return str + "H"
} // ToHex

// ------------------------------------------------- IsPrime -----------------
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

// ------------------------------------------------- PrimeFactorization ---------------------------------
func PrimeFactorization(N int) []int {
	var PD = [...]int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47} // Prime divisors array

	PrimeFactors := make([]int, 0, 10)

	n := N
	for i := 0; i < len(PD); i++ { // outer loop to sequentially test the prime divisors
		for n > 0 && n%PD[i] == 0 {
			PrimeFactors = append(PrimeFactors, PD[i])
			n = n / PD[i]
		}
		if n == 0 || IsPrimeInt(n) {
			PrimeFactors = append(PrimeFactors, n)
			break
		}
	}
	return PrimeFactors

} // PrimeFactorization

// ------------------------------------------------- IsPrimeInt -----------------
func IsPrimeInt(n int) bool {

	var t uint64 = 3

	Uint := uint64(n)

	if Uint == 0 || Uint == 1 {
		return false
	} else if Uint == 2 || Uint == 3 {
		return true
	} else if Uint%2 == 0 {
		return false
	}

	sqrt := math.Sqrt(float64(Uint))
	UintSqrt := uint64(sqrt)

	for t <= UintSqrt {
		if Uint%t == 0 {
			return false
		}
		t += 2
	}
	return true
} // IsPrimeInt

// --------------------------------------- PrimeFactorMemoized -------------------
func PrimeFactorMemoized(U uint) []uint {

	if U == 0 {
		return nil
	}

	var val uint = 2

	PrimeUfactors := make([]uint, 0, 20)

	//	fmt.Print("u, fac, val, primeflag : ")
	for u := U; u > 1; {
		fac, facflag := NextPrimeFac(u, val)
		//		fmt.Print(u, " ", fac, " ", val, " ", primeflag, ", ")
		if facflag {
			PrimeUfactors = append(PrimeUfactors, fac)
			u = u / fac
			val = fac
		} else {
			PrimeUfactors = append(PrimeUfactors, u)
			break
		}
	}
	//	fmt.Println()
	return PrimeUfactors
}

// ------------------------------------------------- NextPrimeFac -----------------
func NextPrimeFac(n, startfac uint) (uint, bool) { // note that this is the reverse of IsPrime

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

//----------------------------------------------- usqrt ---------------------------
func usqrt(u uint) uint {

	sqrt := u / 2

	for i := 0; i < 30; i++ {
		guess := u / sqrt
		sqrt = (guess + sqrt) / 2
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}
	return sqrt
}

// ---------------------------------------- PWRI -------------------------------------------------
//POWER OF I.
// This is a power function with a real base and integer exponent, using the optimized algorithm as discussed in PIM-2, V.  2.
func PWRI(R float64, I int) float64 {
	Z := 1.0
	NEGFLAG := false
	if I < 0 {
		NEGFLAG = true
		I = -I
	}
	for I > 0 {
		if I%2 == 1 {
			Z = Z * R
		}
		R = R * R
		I = I / 2
	}
	if NEGFLAG {
		Z = 1 / Z
	}
	return Z
} // PWRI

//-------------------------------------------------------- StacksMatrixUp
func StacksMatrixUp() {
	for i := T2; i >= X; i-- {
		StackUndoMatrix[i+1] = StackUndoMatrix[i]
	} // FOR i
} // StacksMatrixUp
//-------------------------------------------------------- StacksMatrixDown
func StacksMatrixDown() {
	for i := Y; i <= T1; i++ {
		StackUndoMatrix[i-1] = StackUndoMatrix[i]
	} // FOR i
} // StacksMatrixDown
//-------------------------------------------------------- PushMatrixStacks
func PushMatrixStacks() {
	StacksMatrixUp()
	StackUndoMatrix[Bottom] = Stack
} // PushMatrixStacks
//-------------------------------------------------------- UndoMatrixStacks
func UndoMatrixStacks() { // RollDown operation for main stack
	TempStack := Stack
	Stack = StackUndoMatrix[Bottom]

	StacksMatrixDown()

	StackUndoMatrix[Top] = TempStack
} // UndoMatrixStacks  IE RollDown

//-------------------------------------------------------- RedoMatrixStacks
func RedoMatrixStacks() { // RollUp uperation for main stack
	TempStack := Stack
	Stack = StackUndoMatrix[Top]

	StacksMatrixUp()

	StackUndoMatrix[Bottom] = TempStack
} // RedoMatrixStacks  IE RollUp

//-------------------------------------------------------- HCF -------------------------------------
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

//------------------------------------------------------------------------- GetResults -----------
func GetResult(s string) (float64, []string) {
	var year int
	var Token tknptr.TokenType
	var EOL bool

	ss := make([]string, 0, 100) // ss is abbrev for stringslice.

	tokenPointer := tknptr.NewToken(s) // Using the Go idiom, instead of INITKN(s)
outerloop:
	for { //  UNTIL reached EOL
		Token, EOL = tokenPointer.GETTKNREAL()
		//    fmt.Println(" In GetResult after GetTknReal and R =",Token.Rsum,", Token.Str =",Token.Str,  ", TokenState = ", FSATypeString[Token.State]);
		fmt.Println()
		if EOL {
			break
		}

		switch Token.State {
		case tknptr.DELIM:

		case tknptr.DGT:
			PUSHX(Token.Rsum)
			PushMatrixStacks()

		case tknptr.OP:
			I := Token.Isum
			if (I == 6) || (I == 20) || (I == 1) || (I == 3) { // <>, ><, <, > will all SWAP
				SWAPXY()
			} else {
				LastX = Stack[X]
				PushMatrixStacks()
				switch I {
				case 8:
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
					ss = append(ss, fmt.Sprintf("%s is an unrecognized operation.", Token.Str))
					STACKUP()
				}
				if I != 22 { // Do not move stack for % operator
					STACKDN()
				}
			} // opcode value condition
		case tknptr.ALLELSE:
			cmdnum := cmdMap[Token.Str]
			if cmdnum == 0 && len(Token.Str) > 2 {
				TokenStrShortened := Token.Str[:3] // First 3 characters, ie, characters at positions 0, 1 and 2
				cmdnum = cmdMap[TokenStrShortened]
			}
			if cmdnum == 0 {
				ss = append(ss, fmt.Sprintf(" %s is an unrecognized command.", Token.Str))
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
				ch := Token.Str[len(Token.Str)-1] // ie, the last character.
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
			case 120: // HELP or ?
				ss = append(ss, " SQRT,SQR -- X = sqrt(X) or sqr(X) register.")
				ss = append(ss, " CURT,CBRT -- X = cuberoot(X).")
				ss = append(ss, " RECIP -- X = 1/X.")
				ss = append(ss, " CHS,_ -- Change Sign,  X = -1 * X.")
				ss = append(ss, " DIA -- Given a volume in X, then X = estimated diameter for that volume, assuming a sphere.  Does not approximate Pi as 3.")
				ss = append(ss, " VOL -- Take values in X, Y, And Z and return a volume in X.  Does not approximate Pi as 3.")
				ss = append(ss, " TOCLIP, FROMCLIP -- uses xclip on linux and tcc 22 on Windows to access the clipboard.")
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
				ss = append(ss, " JUL -- Return Julian date number of Z month, Y day, X year.  Pop stack x2.")
				ss = append(ss, " TODAY, T -- Return Julian date number of today's date.  Pop stack x2.")
				ss = append(ss, " GREG-- Return Z month, Y day, X year of Julian date number in X.")
				ss = append(ss, " DOW -- Return day number 0..6 of julian date number in X register.")
				ss = append(ss, " HEX -- Round X register to a long_integer and output it in hex format.")
				ss = append(ss, " HCF -- Push HCF(Y,X) onto stack without removing Y or X.")
				ss = append(ss, " HOL -- Display holidays.")
				ss = append(ss, " UNDO, REDO -- entire stack.  More comprehensive than lastx.")
				ss = append(ss, " Prime, PrimeFactors -- evaluates X.")
				ss = append(ss, " Adjust -- X reg *100, Round, /100")
				ss = append(ss, " NextAfter,Before,Prev -- Reference factor for the fcn is 1e9 or 0.")
				ss = append(ss, " SigFigN,FixN -- Set the significant figures to N for the stack display string.  Default is -1.")
				ss = append(ss, " substitutions: = for +, ; for *.")
				ss = append(ss, " lb2g, oz2g, cm2in, m2ft, mi2km, and their inverses -- unit conversions.")
				ss = append(ss, fmt.Sprintf(" last altered hpcalc %s.", LastAlteredDate))
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
				if IsPrime(n) {
					ss = append(ss, fmt.Sprintf("%d is prime.", int64(n)))
				} else {
					ss = append(ss, fmt.Sprintf("%d is NOT prime.", int64(n)))
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
					ss = append(ss, fmt.Sprintf(" Cannot convert X register to hex string, as number is out of range."))
				} // Hex command
			case 280: // HCF
				c1 := int(math.Abs(Round(Stack[X])))
				c2 := int(math.Abs(Round(Stack[Y])))
				c := HCF(c2, c1)
				ss = append(ss, fmt.Sprintf("HCF of %d and %d is %d.", c1, c2, c))
			case 290: // P
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
				_, _, year = timlibg.TIME2MDY()
				if Stack[X] <= float64(year%100) { // % is the MOD operator
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
				ss = append(ss, fmt.Sprintf(" last changed hpcalc.go %s", LastAlteredDate))
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
			case 520: // TOCLIP
				R := READX()
				s := strconv.FormatFloat(R, 'g', -1, 64)
				if runtime.GOOS == "linux" {
					linuxclippy := func(s string) {
						buf := []byte(s)
						rdr := bytes.NewReader(buf)
						cmd := exec.Command("xclip")
						cmd.Stdin = rdr
						cmd.Stdout = os.Stdout
						cmd.Run()
						ss = append(ss, fmt.Sprintf(" Sent %s to xclip.", s))
					}
					linuxclippy(s)
				} else if runtime.GOOS == "windows" {
					comspec, ok := os.LookupEnv("ComSpec")
					if !ok {
						ss = append(ss, " Environment does not have ComSpec entry.  ToClip unsuccessful.")
						break outerloop
					}
					winclippy := func(s string) {
						//cmd := exec.Command("c:/Program Files/JPSoft/tcmd22/tcc.exe", "-C", "echo", s, ">clip:")
						cmd := exec.Command(comspec, "-C", "echo", s, ">clip:")
						cmd.Stdout = os.Stdout
						cmd.Run()
						ss = append(ss, fmt.Sprintf(" Sent %s to %s.", s, comspec))
					}
					winclippy(s)
				}

			case 530: // FROMCLIP
				w := bytes.NewBuffer([]byte{}) // From "Go Standard Library Cookbook" as referenced above.
				if runtime.GOOS == "linux" {
					cmdfromclip := exec.Command("xclip", "-o")
					cmdfromclip.Stdout = w
					cmdfromclip.Run()
					str := w.String()
					s := fmt.Sprintf(" Received %s from xclip.", str)
					str = strings.ReplaceAll(str, "\n", "")
					str = strings.ReplaceAll(str, "\r", "")
					str = strings.ReplaceAll(str, ",", "")
					str = strings.ReplaceAll(str, " ", "")
					s = s + fmt.Sprintf("  After removing all commas and spaces it becomes %s.", str)
					ss = append(ss, s)
					R, err := strconv.ParseFloat(str, 64)
					if err != nil {
						ss = append(ss, fmt.Sprintln(" fromclip on linux conversion returned error", err, ".  Value ignored."))
					} else {
						PUSHX(R)
					}
				} else if runtime.GOOS == "windows" {
					comspec, ok := os.LookupEnv("ComSpec")
					if !ok {
						ss = append(ss, " Environment does not have ComSpec entry.  FromClip unsuccessful.")
						break outerloop
					}

					//cmdfromclip := exec.Command("c:/Program Files/JPSoft/tcmd22/tcc.exe", "-C", "echo", "%@clip[0]")
					cmdfromclip := exec.Command(comspec, "-C", "echo", "%@clip[0]")
					cmdfromclip.Stdout = w
					cmdfromclip.Run()
					lines := w.String()
					s := fmt.Sprint(" Received ", lines, "from ", comspec)
					linessplit := strings.Split(lines, "\n")
					str := strings.ReplaceAll(linessplit[1], "\"", "")
					str = strings.ReplaceAll(str, "\n", "")
					str = strings.ReplaceAll(str, "\r", "")
					str = strings.ReplaceAll(str, ",", "")
					str = strings.ReplaceAll(str, " ", "")
					s = s + fmt.Sprintln(", after post processing the string becomes", str)
					ss = append(ss, s)
					R, err := strconv.ParseFloat(str, 64)
					if err != nil {
						ss = append(ss, fmt.Sprintln(" fromclip", err, ".  Value ignored."))
					} else {
						PUSHX(R)
					}
				}

			case 540: // lb2g = 453.59238
				r := READX() * lb2g // X * conversion factor pounds to g
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s1 := fmt.Sprintf("%s pounds is %s grams", x, s0)
				ss = append(ss, s1)

			case 550: // oz2g = 28.34952
				r := READX() * oz2g // X * conversion factor ounce to g
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s := fmt.Sprintf("%s oz is %s grams", x, s0)
				ss = append(ss, s)

			case 560: // cm2in = 2.54
				r := READX() / cm2in // X / conversion factor of cm to inches
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s := fmt.Sprintf("%s cm is %s inches", x, s0)
				ss = append(ss, s)

			case 570: // m2ft = 3.28084
				r := READX() * m2ft
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s := fmt.Sprintf("%s meters is %s feet", x, s0)
				ss = append(ss, s)

			case 580: // mi2km = 1.609344
				r := READX() * mi2km
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s := fmt.Sprintf("%s miles is %s km", x, s0)
				ss = append(ss, s)

			case 590: // g2lb
				r := READX() / lb2g
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s1 := fmt.Sprintf("%s grams is %s pounds", x, s0)
				ss = append(ss, s1)

			case 600: // g2oz
				r := READX() / oz2g
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s1 := fmt.Sprintf("%s grams is %s oz", x, s0)
				ss = append(ss, s1)

			case 610: //in2cm
				r := READX() * cm2in
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s1 := fmt.Sprintf("%s inches is %s cm", x, s0)
				ss = append(ss, s1)

			case 620: // ft2m
				r := READX() / m2ft
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s := fmt.Sprintf("%s ft is %s meters", x, s0)
				ss = append(ss, s)

			case 630: // km2mi
				r := READX() / mi2km
				x := strconv.FormatFloat(READX(), 'f', sigfig, 64)
				s0 := strconv.FormatFloat(r, 'f', sigfig, 64)
				s := fmt.Sprintf("%s km is %s mi", x, s0)
				ss = append(ss, s)

			default:
				ss = append(ss, fmt.Sprintf(" %s is an unrecognized command.  And should not get here.", Token.Str))
			} // main text command selection if statement
		}
	}
	return Stack[X], ss
} // GETRESULT

/* ------------------------------------------------------------ GetRegIdx --------- */
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
