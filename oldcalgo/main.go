// oldcalgo.go from calgo.go of Nov 14, 2022
// Copyright (C) 1987-2022  Robert Solomon MD.  All rights reserved.

package main

/*
  REVISION HISTORY
  ----------------
   6 Apr 88 -- 1) Converted to M2 V3.03.
               2) Response to 12 page question is now echoed to the terminal.
               3) Module name changed to CAL so not to conflict with the Logitech's CALENDAR library module.
   4 Nov 90 -- Updated the UTILLIB references to UL2, and recompiled under V3.4.
  28 Oct 91 -- Added FSA parse and indiv month printing abilities.
   2 Nov 91 -- Fixed problem w/ Zeller's congruence when Las2dgts is small enough to make the expression evaluate to a negative value.
  20 Jan 92 -- First page now does not begin with a FF.
----------------------------------------------------------------------------------------------------
  9 Nov 16 -- Converting to Go, using a CLI.  Input a year on the commandline, and output two files.
                A 1-page calendar meant for printing out, and a 12-page calendar meant for importing into Excel.
 10 Nov 16 -- First working version, based on Modula-2 code from 92.
 11 Nov 16 -- Code from January 2009 to import into Excel is working.
 12 Nov 16 -- Fixed bug in DATEASSIGN caused by not porting my own Modula-2 code correctly.
----------------------------------------------------------------------------------------------------
  3 Mar 17 -- Now calgo, and will use termbox to try to do what CALm2 does.
  3 Apr 17 -- Came back to this, after going thru Book of R.
  4 Apr 17 -- Will only write the calendar output files if they do not already exist.
  9 Apr 17 -- For Cal1, now every month also prints the 4 digit year.
 10 Apr 17 -- Will write func AssignYear and allow displaying this year and next year
 12 Apr 17 -- Tweaking display output
 13 Apr 17 -- Golint complained, so I added some comments
 29 Sep 17 -- Changed the output of the final line, and added exec detection code.
  5 Feb 18 -- Will close the calendar files immediately after writing them, instead of waiting for this pgm to exit.
  6 Feb 18 -- Tried to move global variables to main, but had to move them back.
  8 Feb 18 -- Cleaned up code to be more idiomatic, ie, use slices and not arrays.
 22 Nov 19 -- Adding use of flags.  Decided that will have month only be alphabetic, and year only numeric, so order does not matter.
 25 Dec 19 -- Fixed termbox, I hope.
 10 Jan 20 -- Removed ending termbox.flush and close, as they make windows panic.
 14 Nov 22 -- Fixed import list, so it's now module aware.  But I don't use this pgm now; I use calg instead.
*/

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"log"
	"os"
	"os/exec" // for the clear screen functions.
	"src/holidaycalc"
	"src/timlibg"
	"strconv"
	"strings"
	"unicode"
)

const lastCompiled = "Jan 10, 2020"

// BLANKCHR is probably not used much anymore, but golint needs a comment
const BLANKCHR = ' '

const horizTab = 9

// BlankLineWithTabs -- There are embedded <tab> chars here.
const BlankLineWithTabs = "  	  	  	  	  	  	  "

// These are the month number constants.  Golint complains unless I write this.
const (
	JAN = iota
	FEB
	MAR
	APR
	MAY
	JUN
	JUL
	AUG
	SEP
	OCT
	NOV
	DEC
	NumOfMonthsInYear
)

// DCM is now a synonym for December Month Number = 11, as Jan = 0.
const DCM = DEC

var OutCal1file, OutCal12file *bufio.Writer // must be global
var PROMPT, ExtDefault, YEARSTR string
var BLANKSTR2 = "  "
var BLANKSTR3 = "   "
var Cal1Filename, Cal12Filename string
var MN, MN2, MN3 int //  MNEnum Month Number Vars

// DateCell structure was added for termbox code.  Subscripts are [MN] [W] [DOW]
type DateCell struct {
	DateStr  string
	day      int
	ch1, ch2 rune
	fg, bg   termbox.Attribute
}
type WeekVector [7]DateCell
type MonthMatrix [6]WeekVector
type AllMonthsArray [NumOfMonthsInYear]MonthMatrix

var EntireYear AllMonthsArray

var (
	WIM                                          [NumOfMonthsInYear]int
	DIM                                          = []int{31, 0, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	MONNAMSHORT                                  = []string{"JANUARY", "FEBRUARY", "MARCH", "APRIL", "MAY", "JUNE", "JULY", "AUGUST", "SEPTEMBER", "OCTOBER", "NOVEMBER", "DECEMBER"}
	MONNAMLONG                                   [NumOfMonthsInYear]string
	Clear                                        map[string]func()
	BrightYellow, BrightCyan, BrightGreen, Black termbox.Attribute
	year, DOW, W, CurrentMonthNumber, RequestedMonthNumber, LineNum, TodaysDayNumber, CurrentYear,
	StartCol, StartRow, sigfig, MaxRow, MaxCol, TitleRow, StackRow, RegRow, OutputRow, DisplayCol,
	PromptRow, outputmode, n int
	DAYSNAMLONG      = "SUNDAY    MONDAY      TUESDAY     WEDNESDAY   THURSDAY    FRIDAY      SATURDAY"
	DayNamesWithTabs = "SUNDAY \t MONDAY \t TUESDAY \t WEDNESDAY \t THURSDAY \t FRIDAY \t SATURDAY"
	DAYSNAMSHORT     = "  S  M  T  W TH  F  S    "
)

//                      var MONNAMSHORT [NumOfMonthsInYear]string;  Non-idiomatic declaration and initialization
//                      var DAYSNAMLONG, DayNamesWithTabs, DAYSNAMSHORT string;

// DIM = Days In Month  Non-idiomatic initialization.
//  DIM[JAN] = 31; DIM[MAR] = 31; DIM[APR] = 30; DIM[MAY] = 31; DIM[JUN] = 30; DIM[JUL] = 31; DIM[AUG] = 31;
//  DIM[SEP] = 30; DIM[OCT] = 31; DIM[NOV] = 30; DIM[DCM] = 31;

// ------------------------------------------------------- init -----------------------------------
func init() { // start termbox in the init code doesn't work.  Don't know why.  But this init does work.
	Clear = make(map[string]func())
	Clear["linux"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	Clear["windows"] = func() { // this is a closure, or an anonymous function
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// Cap -- Uses UnquoteChar to return a capitalized letter
func Cap(c rune) rune {
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap

// Print_tb -- helper function for termbox output.
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

// Printf_tb -- helper function for termbox output.
func Printf_tb(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	Print_tb(x, y, fg, bg, s)
}

// ClearLine -- clears the rest of the line so unwanted characters already there don't pollute the display.  Not used now.
//func ClearLine(y int) {
//	if y > MaxRow {
//		y = MaxRow
//	}
//	for x := StartCol; x <= MaxCol; x++ {
//		termbox.SetCell(x, y, 0, Black, Black) // Don't know if it matters if the char is ' ' or nil.
//	}
//	err := termbox.Flush()
//	check(err, "")
//} // end ClearLine

// HardClearScreen -- a way to deal w/ an issue where the Clear function didn't seem to work.  Not needed now.
//func HardClearScreen() {
//	err := termbox.Clear(Black, Black)
//	check(err, "")
//	for row := StartRow; row <= MaxRow; row++ {
//		ClearLine(row)
//	}
//	err = termbox.Flush()
//	check(err, "")
//}

// ClearEOL -- clears the rest of the line to not pollute the output.
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
	check(err, "")
}

// ClearScreen -- uses the correct function to achieve this purpose.  Not used now.
//func ClearScreen() {
//	clearfunc, ok := Clear[runtime.GOOS]
//	if ok {
//		clearfunc()
//	} else { // unsupported platform
//		panic(" The ClearScreen platform is only supported on linux or windows, at the moment")
//	}
//}

// DAY2STR -- Could have been handled w/ a Sprintf call now, but I'm keeping this legacy func.
func DAY2STR(DAY int) (string, rune, rune) {
	/*
	   DAY TO STRING CONVERSION.
	   THIS ROUTINE WILL CONVERT THE 2 DIGIT DAY INTO A 2 CHAR STRING.
	   IF THE FIRST DIGIT IS ZERO, THEN THAT CHAR WILL BE BLANK.
	*/

	const digits = "0123456789"
	const ZERO = '0'

	bs := make([]byte, 3)

	TENSDGT := DAY / 10
	UNTSDGT := DAY % 10
	bs[0] = BLANKCHR
	if TENSDGT == 0 {
		bs[1] = BLANKCHR
	} else {
		bs[1] = digits[TENSDGT]
	}
	bs[2] = digits[UNTSDGT]
	return string(bs), rune(bs[1]), rune(bs[2]) // not sure if this is best as a string or as a byteslice
} //END DAY2STR;

func DATEASSIGN(MN int) {
	/*
	   --------------------------------------------------------- DATEASSIGN -------------------------------------------
	   DATE ASSIGNMENT FOR MONTH.
	   THIS ROUTINE WILL ASSIGN THE DATES FOR AN ENTIRE MONTH.  IT WILL PUT THE CHAR
	   REPRESENTATIONS OF THE DATE IN THE FIRST 2 BYTES.  THE EXTRA BYTES CAN BE USED
	   LATER FOR SEPCIAL PRINTER CONTROL CODES.

	   INPUT FROM GBL VAR'S : DIM(MN), DOW
	   OUTPUT TO  GBL VAR'S : DOW, MonthArray(MN,,), WIM(MN)

	*/

	W := 0 // W is for Week number, IE, which week of the month is this.
	for DATE := 1; DATE <= DIM[MN]; DATE++ {
		if DOW > 6 { // DOW = 0 for Sunday.
			W++
			DOW = 0
		} // ENDIF
		DATESTRING, TensRune, UnitsRune := DAY2STR(DATE)
		EntireYear[MN][W][DOW].DateStr = DATESTRING
		EntireYear[MN][W][DOW].day = DATE
		EntireYear[MN][W][DOW].ch1 = TensRune
		EntireYear[MN][W][DOW].ch2 = UnitsRune
		EntireYear[MN][W][DOW].fg = BrightCyan
		EntireYear[MN][W][DOW].bg = Black
		DOW++
	} // ENDFOR;
	WIM[MN] = W  /* Return number of weeks in this month */
	if DOW > 6 { // Don't return a DOW > 6, as that will make a blank first week for next month.
		DOW = 0
	} // if DOW > 6
} // END DATEASSIGN

// PRMONTH -- Prints one month per page.  Not currently used.
//func PRMONTH(MN int) { // Originally intended to print one month per page.  Not currently used.
//	s0 := fmt.Sprintf("%40s", MONNAMSHORT[MN])
//	s1 := fmt.Sprintf("%6s", YEARSTR)
//	_, err := OutCal12file.WriteString(s0)
//	check(err, "Error while writing month name short for big calendar")
//	_, err = OutCal12file.WriteString(s1)
//	check(err, "Error while writing yearstr for big calendar")
//	_, err = OutCal12file.WriteRune('\n')
//	check(err, "")
//	_, err = OutCal12file.WriteRune('\n')
//	check(err, "")
//	_, err = OutCal12file.WriteString(DAYSNAMLONG)
//	check(err, "")
//	_, err = OutCal12file.WriteRune('\n')
//	check(err, "")
//	_, err = OutCal12file.WriteRune('\n')
//	check(err, "")
//	for W := 0; W <= WIM[MN]; W++ {
//		_, err = OutCal12file.WriteString(" ")
//		check(err, "")
//		_, err = OutCal12file.WriteString(EntireYear[MN][W][0].DateStr) // write out Sunday
//		check(err, "")
//		_, err = OutCal12file.WriteString("      ")
//		check(err, "")
//		for I := 1; I < 6; I++ { // write out Monday ..  Friday
//			_, err = OutCal12file.WriteString(" ")
//			check(err, "")
//			_, err = OutCal12file.WriteString(EntireYear[MN][W][I].DateStr)
//			_, err = OutCal12file.WriteString("        ") // FWRBL(OUTUN1,8);
//			check(err, "")
//		}
//		_, err = OutCal12file.WriteString(" ")
//		check(err, "")
//		_, err = OutCal12file.WriteString(EntireYear[MN][W][6].DateStr) // write out Saturday
//		_, err = OutCal12file.WriteRune('\n')
//		check(err, "")
//	}
//} // END PRMONTH

// WrMonthForXL
// Intended to print in a format that can be read by Excel as a call schedule template.

func WrMonthForXL(MN int) {

	s0 := fmt.Sprintf("%s", MONNAMSHORT[MN])
	s1 := fmt.Sprintf("\t%6s", YEARSTR) // I'm going to add <tab> here to see if I like this effect
	_, err := OutCal12file.WriteString(s0)
	check(err, "Error while writing month name short for big calendar")
	_, err = OutCal12file.WriteString(s1)
	check(err, "Error while writing yearstr for big calendar")
	_, err = OutCal12file.WriteRune('\n')
	check(err, "")
	_, err = OutCal12file.WriteString(DayNamesWithTabs)
	check(err, "")
	_, err = OutCal12file.WriteRune('\n')
	check(err, "")

	for W := 0; W <= WIM[MN]; W++ {
		_, err = OutCal12file.WriteString(EntireYear[MN][W][0].DateStr) // write out Sunday
		check(err, "")
		err = OutCal12file.WriteByte(horizTab) // <tab>, or horizontal tab <HT>, to confirm that this does work
		check(err, "")

		for I := 1; I < 6; I++ { // write out Monday ..  Friday

			_, err = OutCal12file.WriteString(EntireYear[MN][W][I].DateStr)
			check(err, "")
			_, err = OutCal12file.WriteRune('\t') // <tab>, or horizontal tab <HT>, to see if this works
			check(err, "")

		} // ENDFOR I

		_, err = OutCal12file.WriteString(EntireYear[MN][W][6].DateStr) // write out Saturday
		check(err, "")
		_, err = OutCal12file.WriteRune('\n')
		check(err, "")
		_, err = OutCal12file.WriteString(BlankLineWithTabs)
		check(err, "")
		_, err = OutCal12file.WriteRune('\n')
		check(err, "")
		_, err = OutCal12file.WriteString(BlankLineWithTabs)
		check(err, "")
		_, err = OutCal12file.WriteRune('\n')
		check(err, "")
	} // ENDFOR W
	_, err = OutCal12file.WriteRune('\n')
	check(err, "")
	_, err = OutCal12file.WriteRune('\n')
	check(err, "")
} // END WrMonthForXL

// -------------------------------------- WrOnePageYear ----------------------------------

func WrOnePageYear() { // Each column must be exactly 25 characters for the spacing to work.
	var err error
	// Write one page calendar
	s := fmt.Sprintf("%40s", YEARSTR)
	//  _, err := OutCal1file.WriteString(s);
	//                                                check(err,"Error while writing YEARSTR to Cal 1 file");
	//  _, err = OutCal1file.WriteRune('\n');
	//                                                check(err,"Error while writing a newline rune to Cal 1 file");

	for MN = JAN; MN <= DCM; MN += 3 {
		MN2 = MN + 1
		MN3 = MN + 2

		//    _, err = OutCal1file.WriteRune('\n');
		//                                                  check(err,"Error while writing newline rune to Cal 1 file");
		if MN > JAN { // have fewer blank lines after year heading than btwn rows of months.
			_, err = OutCal1file.WriteRune('\n')
			check(err, "Error while writing newline rune to Cal 1 file")
			_, err = OutCal1file.WriteRune('\n')
			check(err, "Error while writing newline rune to Cal 1 file")
		}
		_, err := OutCal1file.WriteString(s)
		check(err, "Error while writing YEARSTR to Cal 1 file")
		_, err = OutCal1file.WriteRune('\n')
		check(err, "Error while writing a newline rune to Cal 1 file")
		_, err = OutCal1file.WriteRune('\n')
		check(err, "Error while writing a newline rune to Cal 1 file")
		_, err = OutCal1file.WriteString(MONNAMLONG[MN])
		check(err, "Error writing first long month name to cal 1 file")
		_, err = OutCal1file.WriteString(MONNAMLONG[MN2])
		check(err, "")
		_, err = OutCal1file.WriteString(MONNAMLONG[MN3])
		check(err, "")
		_, err = OutCal1file.WriteRune('\n')
		check(err, "Error while writing newline rune to Cal 1 file")
		_, err = OutCal1file.WriteRune('\n')
		check(err, "Error while writing newline rune to Cal 1 file")
		//    _, err = OutCal1file.WriteRune('\n');                         too many blank lines
		//    check(err,"Error while writing newline rune to Cal 1 file");
		_, err = OutCal1file.WriteString(DAYSNAMSHORT)
		check(err, "Error while writing day names to cal 1 file")
		_, err = OutCal1file.WriteString(DAYSNAMSHORT)
		check(err, "Error while writing day names to cal 1 file")
		_, err = OutCal1file.WriteString(DAYSNAMSHORT)
		check(err, "Error while writing day names to cal 1 file")
		_, err = OutCal1file.WriteRune('\n')
		check(err, "Error while writing newline rune to Cal 1 file")
		for W = 0; W < 6; W++ { // week number
			for I := 0; I < 7; I++ { // day of week positions for 1st month
				_, err = OutCal1file.WriteString(EntireYear[MN][W][I].DateStr)
				check(err, "Error while writing date string to cal 1 file")
			} // ENDFOR I
			_, err = OutCal1file.WriteString("    ")
			check(err, "")
			for I := 0; I < 7; I++ { // day of week positions for 2nd month
				_, err = OutCal1file.WriteString(EntireYear[MN2][W][I].DateStr)
				check(err, "Error while writing date string to cal 1 file")
			} // ENDFOR I
			_, err = OutCal1file.WriteString("    ")
			check(err, "")
			for I := 0; I < 7; I++ { // day of week position for 3rd month
				_, err = OutCal1file.WriteString(EntireYear[MN3][W][I].DateStr)
				check(err, "Error while writing date string to cal 1 file")
			} // ENDFOR I
			_, err = OutCal1file.WriteRune('\n')
			check(err, "Error while writing newline rune to Cal 1 file")
		} // ENDFOR W
	} // ENDFOR MN;
	_, err = OutCal1file.WriteRune('\n')
	check(err, "Error while writing newline rune to Cal 1 file")
	_, err = OutCal1file.WriteString(s)
	check(err, "Error while writing YEARSTR to Cal 1 file")
	_, err = OutCal1file.WriteRune('\n')
	check(err, "Error while writing a newline rune to Cal 1 file")

} // WrOnePageYear

// ----------------------------- ShowMonth ---------------------------------
func ShowMonth(col, row, mn int) {
	// col is the starting col for this month number.  Will likely be either 0, 25 or 50.
	// Each week is 21 char wide (3 x 7), and 4 spaces btwn months.
	// Print_tb should be able to handle this easily.  I have not yet coded the change in colors for a particular day.  I may process the holidays month by month or entire year.
	// And I want to have today's date be shown differently, also.
	// type DateCell struct { DateStr string; ch1,ch2 rune; fg,bg termbox.Attribute; }
	// func Printf_tb(x,y int, fg,bg termbox.Attribute, format string, args ...interface{})

	y := row

	Print_tb(col, y, BrightCyan, Black, MONNAMLONG[mn])
	y++
	Print_tb(col, y, BrightCyan, Black, DAYSNAMSHORT)
	y++
	for W = 0; W < 6; W++ { // week number
		x := col
		for I := 0; I < 7; I++ { // day of week positions for 1st month
			Print_tb(x, y, EntireYear[mn][W][I].fg, EntireYear[mn][W][I].bg, EntireYear[mn][W][I].DateStr)
			x += 3
		} // ENDFOR I
		y++
	}
} // END ShowMonth

// ----------------------------- HolidayAssign ---------------------------------

func HolidayAssign(year int) {

	var Holiday holidaycalc.HolType

	// type MDType struct { M,D int;}

	// type HolType struct {
	//         MLK,Pres,Easter,Mother,Memorial,Father,Labor,Columbus,Election,Thanksgiving MDType;
	//         Year int;
	//         Valid bool;
	//}                              TodaysDayNumber
	if year < 40 {
		year += 2000
	} else if year < 100 {
		year += 1900
	}
	Holiday = holidaycalc.GetHolidays(year)
	Holiday.Valid = true

	// New Year's Day
	julian := timlibg.JULIAN(1, 1, year)
	dow := julian % 7
	EntireYear[0][0][dow].fg = Black
	EntireYear[0][0][dow].bg = BrightYellow

	// MLK Day
	d := Holiday.MLK.D
MLKloop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[0][w][dow].day == d {
				EntireYear[0][w][dow].fg = Black
				EntireYear[0][w][dow].bg = BrightYellow
				break MLKloop
			}
		}
	}

	// President's Day
	d = Holiday.Pres.D
PresLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[1][w][dow].day == d {
				EntireYear[1][w][dow].fg = Black
				EntireYear[1][w][dow].bg = BrightYellow
				break PresLoop
			}
		}
	}

	// Easter
	m := Holiday.Easter.M - 1
	d = Holiday.Easter.D
EasterLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[m][w][dow].day == d {
				EntireYear[m][w][dow].fg = Black
				EntireYear[m][w][dow].bg = BrightYellow
				break EasterLoop
			}
		}
	}

	// Mother's Day
	d = Holiday.Mother.D
MotherLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[4][w][dow].day == d {
				EntireYear[4][w][dow].fg = Black
				EntireYear[4][w][dow].bg = BrightYellow
				break MotherLoop
			}
		}
	}

	// Memorial Day
	d = Holiday.Memorial.D
MemorialLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[4][w][dow].day == d {
				EntireYear[4][w][dow].fg = Black
				EntireYear[4][w][dow].bg = BrightYellow
				break MemorialLoop
			}
		}
	}

	// Father's Day
	d = Holiday.Father.D
FatherLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[5][w][dow].day == d {
				EntireYear[5][w][dow].fg = Black
				EntireYear[5][w][dow].bg = BrightYellow
				break FatherLoop
			}
		}
	}

	// July 4th
	d = 4
IndependenceLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[6][w][dow].day == d {
				EntireYear[6][w][dow].fg = Black
				EntireYear[6][w][dow].bg = BrightYellow
				break IndependenceLoop
			}
		}
	}

	// Labor Day
	d = Holiday.Labor.D
LaborLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[8][w][dow].day == d {
				EntireYear[8][w][dow].fg = Black
				EntireYear[8][w][dow].bg = BrightYellow
				break LaborLoop
			}
		}
	}

	// Columbus Day
	d = Holiday.Columbus.D
ColumbusLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[9][w][dow].day == d {
				EntireYear[9][w][dow].fg = Black
				EntireYear[9][w][dow].bg = BrightYellow
				break ColumbusLoop
			}
		}
	}

	// Election Day
	d = Holiday.Election.D
ElectionLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[10][w][dow].day == d {
				EntireYear[10][w][dow].fg = Black
				EntireYear[10][w][dow].bg = BrightYellow
				break ElectionLoop
			}
		}
	}

	// Veteran's Day
	d = 11
VeteranLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			//      Printf_tb(0,20,BrightYellow,Black," w = %d, dow = %d ",w,dow);
			//      Print_tb(0,MaxRow-1,BrightYellow,Black," Hit <enter> to continue.");
			//      termbox.SetCursor(26,MaxRow);
			//      _ = GetInputString(26,MaxRow);
			if EntireYear[10][w][dow].day == d {
				EntireYear[10][w][dow].fg = Black
				EntireYear[10][w][dow].bg = BrightYellow
				break VeteranLoop
			}
		}
	}

	// Thanksgiving Day
	d = Holiday.Thanksgiving.D
TGLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[10][w][dow].day == d {
				EntireYear[10][w][dow].fg = Black
				EntireYear[10][w][dow].bg = BrightYellow
				break TGLoop
			}
		}
	}

	// Christmas Day
	d = 25
XmasLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[11][w][dow].day == d {
				EntireYear[11][w][dow].fg = Black
				EntireYear[11][w][dow].bg = BrightYellow
				break XmasLoop
			}
		}
	}

	// Today
	if year == CurrentYear {
		d = TodaysDayNumber
		m = CurrentMonthNumber - 1
	TodayLoop:
		for w := 0; w < 6; w++ {
			for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
				if EntireYear[m][w][dow].day == d {
					EntireYear[m][w][dow].fg = Black
					EntireYear[m][w][dow].bg = BrightGreen
					break TodayLoop
				}
			}
		}
	}

} // END HolidayAssign

// ----------------------------- SetMonthNumber ----------------------------------

func SetMonthNumber(s string) int { // returns -1 if there was an error
	var month int

	month = -1
	t := strings.ToUpper(s)
	if unicode.IsLetter(rune(s[0])) { // determine the 3 letter month code
		for c := JAN; c < NumOfMonthsInYear; c++ {
			if strings.HasPrefix(MONNAMSHORT[c], t) {
				month = c
				break
			}
		}
	}

	//  Printf_tb(0,LineNum,BrightYellow,Black,
	//            " in SetMonthNumber: monnamshort[0] %s monnamshort[1] %s, token %#v, Requestedmon num %d",
	//            MONNAMSHORT[0],MONNAMSHORT[1],token,RequestedMonthNum);
	//  LineNum++
	//  LineNum++
	return month
}

// ----------------------------------- AssignYear ----------------------------------------------------
func AssignYear(y int) {

	if y < 40 {
		y += 2000
	} else if y < 100 {
		y += 1900
	} else if y < 1900 || y > 2100 {
		fmt.Printf("Year is %d, which is out of range (1900-2100).  Exiting.\n", y)
		os.Exit(1)
	}

	JulDate := timlibg.JULIAN(1, 1, y)
	JAN1DOW := JulDate % 7
	DOW = JAN1DOW
	FEBDAYS := 28

	if ((y % 4) == 0) && ((y % 100) != 0) {
		// YEAR IS DIVISIBLE BY 4 AND NOT BY 100
		FEBDAYS = 29
	} else if (y % 400) == 0 {
		FEBDAYS = 29
	} // ENDIF about leap year

	DIM[FEB] = FEBDAYS

	// Initialize the calendar to all BLANKSTR3, for correct spacing
	for m := JAN; m <= DEC; m++ { // month position
		for wk := 0; wk < 6; wk++ { // week position
			for dayofweek := 0; dayofweek < 7; dayofweek++ {
				EntireYear[m][wk][dayofweek].DateStr = BLANKSTR3
				EntireYear[m][wk][dayofweek].day = 0
				EntireYear[m][wk][dayofweek].ch1 = '0'
				EntireYear[m][wk][dayofweek].ch2 = '0'
				EntireYear[m][wk][dayofweek].fg = Black
				EntireYear[m][wk][dayofweek].bg = Black
			}
		}
	}

	// Make the calendar

	for MN := JAN; MN <= DCM; MN++ {
		DATEASSIGN(MN)
	} // ENDFOR;

} // END AssignYear

/*
--------------------- MAIN ---------------------------------------------
*/
func main() {
	var Cal1FilenameFlag, Cal12FilenameFlag bool
	//	var YearToken tokenize.TokenType

	BrightYellow = termbox.ColorYellow | termbox.AttrBold
	BrightCyan = termbox.ColorCyan | termbox.AttrBold
	BrightGreen = termbox.ColorGreen | termbox.AttrBold
	Black = termbox.ColorBlack
	//	fmt.Println(" Calendar Printing Program written in Go.  Last altered ", lastCompiled)
	//	fmt.Println()
	MONNAMLONG[JAN] = "    J A N U A R Y        "
	MONNAMLONG[FEB] = "   F E B R U A R Y       "
	MONNAMLONG[MAR] = "      M A R C H          "
	MONNAMLONG[APR] = "      A P R I L          "
	MONNAMLONG[MAY] = "        M A Y            "
	MONNAMLONG[JUN] = "       J U N E           "
	MONNAMLONG[JUL] = "       J U L Y           "
	MONNAMLONG[AUG] = "     A U G U S T         "
	MONNAMLONG[SEP] = "  S E P T E M B E R      "
	MONNAMLONG[OCT] = "    O C T O B E R        "
	MONNAMLONG[NOV] = "   N O V E M B E R       "
	MONNAMLONG[DCM] = "   D E C E M B E R       "

	PROMPT = " Enter Year : " // not currently used.
	Ext1Default := ".out"
	Ext12Default := ".xls"
	CurrentMonthNumber, TodaysDayNumber, CurrentYear = timlibg.TIME2MDY()

	// flag definitions and processing
	var nofilesflag = flag.Bool("no", false, "do not generate output cal1 and cal12 files.") // Ptr
	var NoFilesFlag = flag.Bool("n", false, "do not generate output cal1 and cal12 files.")  // Ptr

	var helpflag = flag.Bool("h", false, "print help message.") // pointer
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "help", false, "print help message.")

	var testFlag = flag.Bool("test", false, "test mode flag.") // pointer

	flag.Parse()

	if *helpflag || HelpFlag {
		fmt.Println()
		fmt.Println(" Calgo Calendar Printing Program, last altered", lastCompiled)
		fmt.Println(" Usage: calgo <flags> year month or month year, where month must be a month name string.")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *testFlag {
		fmt.Println()
		fmt.Println()
		fmt.Println(" calgo, a calendar printing program written in Go.  Last altered", lastCompiled)
		fmt.Println()
		//ans := ""
		//fmt.Print(" pausing, hit <enter> to resume")
		//fmt.Scanln(&ans)
		//fmt.Println()
	}

	// process command line parameters
	RequestedMonthNumber = CurrentMonthNumber - 1 // default value.
	MonthNotExplicitlySet := true
	if flag.NArg() > 0 {
		commandline := flag.Args()
		if flag.NArg() > 2 {
			commandline = commandline[:2] // if there are too many params, only use params 0 and 1, ie, up to 2 but not incl'g 2.
		}
		for _, commandlineparam := range commandline {
			if unicode.IsDigit(rune(commandlineparam[0])) { // have numeric parameter, must be a year
				YEARSTR = commandlineparam
				var err error
				year, err = strconv.Atoi(commandlineparam)
				if err != nil {
					fmt.Println(" Error from Atoi for year.  Using CurrentYear.  Entered string is", commandlineparam)
					year = CurrentYear
					fmt.Print(" pausing.  Hit <enter> to continue")
					ans := ""
					fmt.Scanln(&ans)
					fmt.Println()
				}
				if MonthNotExplicitlySet {
					RequestedMonthNumber = 0 // if a year is explicitily entered, start w/ January.
				}
			} else { // not a numeric parameter, process like it's a month abbrev code
				RequestedMonthNumber = SetMonthNumber(commandlineparam)
				if RequestedMonthNumber < 0 { // if error from SetMonthNumber, use current month
					fmt.Println(" Error from SetMonthNumber.  Using current month of ", CurrentMonthNumber)
					RequestedMonthNumber = CurrentMonthNumber - 1
					fmt.Print(" pausing.  Hit <enter> to continue ")
					ans := ""
					fmt.Scanln(&ans)
					fmt.Println()
				}
				MonthNotExplicitlySet = false
			}
		}
	} else {
		year = CurrentYear
	}

	if year < 40 {
		year += 2000
	} else if year < 100 {
		year += 1900
	} else if year < 1900 || year > 2100 {
		fmt.Printf("Year is %d, which is out of range (1900-2100).  Exiting.\n", year)
		os.Exit(1)
	}
	if *testFlag {
		fmt.Println()
		fmt.Println(" using year", year, ", using month", MONNAMSHORT[RequestedMonthNumber])
		fmt.Println()
		ans := ""
		fmt.Print(" pausing, hit <enter> to resume")
		fmt.Scanln(&ans)
		fmt.Println()
	}

	YEARSTR = strconv.Itoa(year) // This will always be a 4 digit year, regardless of what's entered on command line.

	AssignYear(year)

	HolidayAssign(year)

	AllowFilesFlag := !(*nofilesflag || *NoFilesFlag)
	Cal1FilenameFlag = false  // default value
	Cal12FilenameFlag = false // default value
	if AllowFilesFlag {
		BaseFilename := YEARSTR
		Cal1Filename = BaseFilename + "_cal1" + Ext1Default
		Cal12Filename = BaseFilename + "_cal12" + Ext12Default
		FI, err := os.Stat(Cal1Filename)

		if err == nil {
			//		Cal1FilenameFlag = false
			fmt.Printf(" %s already exists.  From stat call file created %s, filesize is %d.\n",
				Cal1Filename, FI.ModTime().Format("Jan-02-2006 15:04:05"), FI.Size())
		} else {
			Cal1FilenameFlag = true
			fmt.Printf(" %s does not already exist.\n", Cal1Filename)
		}

		FI, err = os.Stat(Cal12Filename)
		if err == nil {
			//		Cal12FilenameFlag = false
			fmt.Printf(" %s already exists.  From stat call file created %s, filesize is %d.\n",
				Cal12Filename, FI.ModTime().Format("Jan-02-2006 15:04:05"), FI.Size())
		} else {
			Cal12FilenameFlag = true
			fmt.Printf(" %s does not already exist.\n", Cal12Filename)
		}
	}

	if *testFlag {
		fmt.Println()
		fmt.Println(" Completed year matrix.  AllowFilesFlag is", AllowFilesFlag, ".  Ready to jump into termbox.")
		fmt.Print(" pausing.  Hit <enter> to contiue.")
		ans := ""
		fmt.Scanln(&ans)
		fmt.Println()
	}

	termerr := termbox.Init()
	if termerr != nil {
		log.Println(" TermBox init failed.")
		panic(termerr)
	}

	defer termbox.Close()
	defer termbox.Flush() // added 12/25/2019
	defer termbox.Sync()  // added 12/25/2019

	MaxCol, MaxRow = termbox.Size()
	MaxCol-- // These numbers are too large by 1
	MaxRow-- // So decrement them.
	e := termbox.Clear(Black, Black)
	check(e, "")
	e = termbox.Flush()
	check(e, "")

	Printf_tb(0, LineNum, BrightCyan, Black, " Calendar Printing Program written in Go.  Last altered %s", lastCompiled)
	LineNum++

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	Printf_tb(0, LineNum, BrightCyan, Black, "%s was last linked on %s.  Working directory is %s.", ExecFI.Name(), LastLinkedTimeStamp, workingdir)
	LineNum++
	Printf_tb(0, LineNum, BrightCyan, Black, " Full name of executable file is %s", execname)
	LineNum++

	/* {{{
		//                                               s := s1 + s2;
		//                                               Print_tb(0,LineNum,BrightCyan,Black,s);
		//                                               LineNum++
		//                                               LineNum++
	   }}} */
	var OutCal1, OutCal12 *os.File
	if Cal1FilenameFlag {
		OutCal1, err := os.Create(Cal1Filename)
		check(err, " Trying to create Cal1 output file")
		defer OutCal1.Close()
		OutCal1file = bufio.NewWriter(OutCal1)
		defer OutCal1file.Flush()

	}
	if Cal12FilenameFlag {
		OutCal12, e := os.Create(Cal12Filename)
		check(e, " Trying to create Cal12 output file")
		defer OutCal12.Close()
		OutCal12file = bufio.NewWriter(OutCal12)
		defer OutCal12file.Flush()
	}

	// WRITE 12 PAGE CALENDAR, ONE MONTH PER PAGE
	if Cal12FilenameFlag {
		for MN := JAN; MN <= DCM; MN++ {
			WrMonthForXL(MN)
		} // ENDFOR
		OutCal12file.Flush()
		OutCal12.Close()
	}

	// Write One Page Calendar
	if Cal1FilenameFlag {
		WrOnePageYear()
		OutCal1file.Flush()
		OutCal1.Close()
	}

	LineNum++
	LineNum++

	Printf_tb(MaxCol/3, LineNum, BrightYellow, Black, " Year %4d", year)
	LineNum++
	ShowMonth(0, LineNum, RequestedMonthNumber)
	if MaxCol > 48 && RequestedMonthNumber < DEC {
		ShowMonth(25, LineNum, RequestedMonthNumber+1)
	}
	if MaxCol > 72 && RequestedMonthNumber < NOV {
		ShowMonth(50, LineNum, RequestedMonthNumber+2)
	}
	if MaxCol > 98 && RequestedMonthNumber < OCT {
		ShowMonth(75, LineNum, RequestedMonthNumber+3)
	}
	if MaxCol > 122 && RequestedMonthNumber < SEP {
		ShowMonth(100, LineNum, RequestedMonthNumber+4)
	}
	if MaxCol > 148 && RequestedMonthNumber < AUG {
		ShowMonth(125, LineNum, RequestedMonthNumber+5)
	}
	if MaxCol > 172 && RequestedMonthNumber < JUL {
		ShowMonth(150, LineNum, RequestedMonthNumber+6)
	}
	if MaxCol > 198 && RequestedMonthNumber < JUN {
		ShowMonth(175, LineNum, RequestedMonthNumber+7)
	}

	// Now disploy next year.  No file writing.  Min 10 lines/calendar.
	if MaxRow > 30 {
		year++
		YEARSTR = strconv.Itoa(year)
		AssignYear(year)
		HolidayAssign(year)
		LineNum += 10
		RequestedMonthNumber = 0

		Printf_tb(MaxCol/3, LineNum, BrightYellow, Black, " Year %4d", year)
		LineNum++
		ShowMonth(0, LineNum, RequestedMonthNumber)
		if MaxCol > 48 && RequestedMonthNumber < DEC {
			ShowMonth(25, LineNum, RequestedMonthNumber+1)
		}
		if MaxCol > 72 && RequestedMonthNumber < NOV {
			ShowMonth(50, LineNum, RequestedMonthNumber+2)
		}
		if MaxCol > 98 && RequestedMonthNumber < OCT {
			ShowMonth(75, LineNum, RequestedMonthNumber+3)
		}
		if MaxCol > 122 && RequestedMonthNumber < SEP {
			ShowMonth(100, LineNum, RequestedMonthNumber+4)
		}
		if MaxCol > 148 && RequestedMonthNumber < AUG {
			ShowMonth(125, LineNum, RequestedMonthNumber+5)
		}
		if MaxCol > 172 && RequestedMonthNumber < JUL {
			ShowMonth(150, LineNum, RequestedMonthNumber+6)
		}
		if MaxCol > 198 && RequestedMonthNumber < JUN {
			ShowMonth(175, LineNum, RequestedMonthNumber+7)
		}
	}

	LineNum += 20
	Print_tb(0, LineNum, BrightYellow, Black, " Hit <enter> to continue.")
	termbox.SetCursor(26, LineNum)
	_ = GetInputString(26, LineNum)

	//	termbox.Flush() // added 12/25/2019 and removed 1/10/20 as duplicating deferred statements, making Windows unhappy.
	//	termbox.Close() // added 12/25/2019

} // end main func

// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
	if e != nil {
		log.Printf("%s : ", msg)
		panic(e)
	}
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
			check(err, "")
		case termbox.EventResize:
			err := termbox.Sync()
			check(err, "")
			err = termbox.Flush()
			check(err, "")
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

//END calgo

/*
   ColorBlack ColorRed ColorGreen ColorYellow ColorBlue ColorMagenta ColorCyan ColorWhite
   const ( AttrBold AttrUnderline AttrReverse)
*/
