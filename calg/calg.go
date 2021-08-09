// calg.go, derived from caltcell.go, which was derived from  calgo.go.  This version uses colortext
// Copyright (C) 1987-2021  Robert Solomon MD.  All rights reserved.

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
  9 Nov 16 -- Converting to Go, using a CLI.  Input a year on the commandline, and output two files.
                A 1 page calendar meant for printing out, and a 12 page calendar meant for importing into Excel.
 10 Nov 16 -- First working version, based on Modula-2 code from 92.
 11 Nov 16 -- Code from January 2009 to import into Excel is working.
 12 Nov 16 -- Fixed bug in DATEASSIGN caused by not porting my own Modula-2 code correctly.
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
 19 Jan 20 -- Now moved to tcell as terminal interface.  Mostly copied code from rpntcell.go.
 20 Jan 20 -- Removed deleol call from puts, as it's not needed when scrn is written only once.
 18 Feb 21 -- Back to cal.go.  And will convert to colortext calls, removing all tcell stuff as that won't run correctly in tcc.
 20 Feb 21 -- Experimenting w/ allowing reverse colors using ColorText.
 21 Feb 21 -- Adding a comment field to the datecell struct, so holiday string can be output.  And cleaning up the code a bit.
 22 Feb 21 -- Removing text for Columbus and Veteran Days as these are not hospital holidays.
 23 Mar 21 -- Will allow years from 1800 - 2100.  This came up while reading about Apr 14, 1865, which was a Friday.
                And discovered a bug when a 4 digit year is entered.
 18 Jun 21 -- Juneteenth added, as it became a legal federal holiday yesterday, signed into law by Biden.
                And converted to modules.
  9 Aug 21 -- Added -v to be a synonym of test.
*/

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec" // for the clear screen functions.

	"runtime"
	"strconv"
	"strings"
	"unicode"

	"src/holidaycalc"
	"src/timlibg"
	//"tokenize"  I don't use tokenization to parse the params anymore.

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
)

// LastCompiled needs a comment according to golint
const LastCompiled = "Aug 9, 2021"

// BLANKCHR is used in DAY2STR.

const BLANKCHR = ' '

// HorizTab needs comment according to golint
const HorizTab = 9

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
var PROMPT = " Enter Year : "               // not currently used.
var ExtDefault, YEARSTR string
var BLANKSTR2 = "  "
var BLANKSTR3 = "   "
var Cal1Filename, Cal12Filename string
var MN, MN2, MN3 int //  MNEnum Month Number Vars

// DateCell structure was added for termbox code.  Subscripts are [MN] [W] [DOW].  It was adapted for tcell, and now for colortext
type DateCell struct {
	DateStr, comment string
	day              int
	ch1, ch2         rune
	fg, bg           ct.Color
}

type WeekVector [7]DateCell
type MonthMatrix [6]WeekVector
type AllMonthsArray [NumOfMonthsInYear]MonthMatrix

var EntireYear AllMonthsArray
var windowsFlag bool

var (
	WIM                                                                          [NumOfMonthsInYear]int
	DIM                                                                          = []int{31, 0, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	MONNAMSHORT                                                                  = []string{"JANUARY", "FEBRUARY", "MARCH", "APRIL", "MAY", "JUNE", "JULY", "AUGUST", "SEPTEMBER", "OCTOBER", "NOVEMBER", "DECEMBER"}
	MONNAMLONG                                                                   [NumOfMonthsInYear]string
	clear                                                                        map[string]func()
	DOW, W                                                                       int // these are global so the date assign can do their jobs correctly
	year, CurrentMonthNumber, RequestedMonthNumber, TodaysDayNumber, CurrentYear int
	DAYSNAMLONG                                                                  = "SUNDAY    MONDAY      TUESDAY     WEDNESDAY   THURSDAY    FRIDAY      SATURDAY"
	DayNamesWithTabs                                                             = "SUNDAY \t MONDAY \t TUESDAY \t WEDNESDAY \t THURSDAY \t FRIDAY \t SATURDAY"
	DAYSNAMSHORT                                                                 = "  S  M  T  W TH  F  S    "
)

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

// ------------------------------------------------------- DAY2STR  -------------------------------------
func DAY2STR(DAY int) (string, rune, rune) {
	/*
	   DAY TO STRing conversion.
	   This routine will convert the 2 digit day into a 2 char string.  If the first digit is zero, then that char will be blank.
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
	return string(bs), rune(bs[1]), rune(bs[2])
} //END DAY2STR

func DATEASSIGN(MN int) {
	/*
	   --------------------------------------------------------- DATEASSIGN -------------------------------------------
	   DATE ASSIGNment for month.
	   This routine will assign the dates for an entire month.  It will put the char representations of the date in the first 2 bytes.

	   INPUT FROM GBL VAR'S : DIM(MN), DOW
	   OUTPUT TO  GBL VAR'S : DOW, MonthArray(MN,,), WIM(MN)
	*/

	W := 0 // W is for Week number, IE, which week of the month is this.
	for DATE := 1; DATE <= DIM[MN]; DATE++ {
		if DOW > 6 { // DOW = 0 for Sunday.
			W++
			DOW = 0
		}
		DATESTRING, TensRune, UnitsRune := DAY2STR(DATE)
		EntireYear[MN][W][DOW].DateStr = DATESTRING
		EntireYear[MN][W][DOW].day = DATE
		EntireYear[MN][W][DOW].ch1 = TensRune
		EntireYear[MN][W][DOW].ch2 = UnitsRune
		EntireYear[MN][W][DOW].fg = ct.Cyan
		EntireYear[MN][W][DOW].bg = ct.Black
		DOW++
	}
	WIM[MN] = W  /* Return number of weeks in this month */
	if DOW > 6 { // Don't return a DOW > 6, as that will make a blank first week for next month.
		DOW = 0
	}
} // END DATEASSIGN

// ----------------------------------------------------------- WrMonthForXL --------------------------------------
// Intended to print in a format that can be read by Excel as a call schedule template.

func WrMonthForXL(MN int) {

	s0 := fmt.Sprintf("%s", MONNAMSHORT[MN])
	s1 := fmt.Sprintf("\t%6s", YEARSTR) // I like the effect here of adding <tab>
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

	for wk := 0; wk <= WIM[MN]; wk++ { // so won't shadow the global W
		s := fmt.Sprintf("%s  %s", EntireYear[MN][wk][0].comment, EntireYear[MN][wk][0].DateStr)
		_, err = OutCal12file.WriteString(s) // write out Sunday
		check(err, "")
		err = OutCal12file.WriteByte(HorizTab) // <tab>, or horizontal tab <HT>, to confirm that this does work
		check(err, "")

		for I := 1; I < 6; I++ { // write out Monday ..  Friday
			s = fmt.Sprintf("%s  %s", EntireYear[MN][wk][I].comment, EntireYear[MN][wk][I].DateStr)
			_, err = OutCal12file.WriteString(s)
			check(err, "")
			_, err = OutCal12file.WriteRune('\t') // <tab>, or horizontal tab <HT>, to see if this works
			check(err, "")
		}

		s = fmt.Sprintf("%s  %s", EntireYear[MN][wk][6].comment, EntireYear[MN][wk][6].DateStr)
		_, err = OutCal12file.WriteString(s) // write out Saturday
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
	}
	_, err = OutCal12file.WriteRune('\n')
	check(err, "")
	_, err = OutCal12file.WriteRune('\n')
	check(err, "")
} // END WrMonthForXL

// -------------------------------------- WrOnePageYear ----------------------------------

func WrOnePageYear() { // Each column must be exactly 25 characters for the spacing to work.

	// Write one page calendar

	var err error

	s := fmt.Sprintf("%40s", YEARSTR)

	for MN = JAN; MN < NumOfMonthsInYear; MN += 3 {
		MN2 = MN + 1
		MN3 = MN + 2

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

// -------------------------------------- Show3MonthRow ----------------------------------

func Show3MonthRow(mn int) { // Modified from WrOnePageYear.  main() makes sure mn is in range.
	// Display 3 months per row for 2 rows on the terminal using ColorText.
	s := fmt.Sprintf("%40s", YEARSTR)
	ctfmt.Println(ct.Yellow, windowsFlag, s)
	fmt.Println()

	for i := 0; i < 2; i++ { // just display 2 rows with 3 months each
		MN = mn + 3*i
		MN2 = MN + 1
		MN3 = MN + 2

		if i > 0 { // have fewer blank lines after year heading than btwn rows of months.
			fmt.Println()
			fmt.Println()
		}
		fmt.Println()
		fmt.Println()
		fmt.Print(MONNAMLONG[MN])
		fmt.Print(MONNAMLONG[MN2])
		fmt.Print(MONNAMLONG[MN3])
		fmt.Println()
		fmt.Println()
		fmt.Print(DAYSNAMSHORT)
		fmt.Print(DAYSNAMSHORT)
		fmt.Print(DAYSNAMSHORT)
		fmt.Println()

		for W = 0; W < 6; W++ { // week number
			for I := 0; I < 7; I++ { // day of week positions for 1st month
				if EntireYear[MN][W][I].bg == ct.Black {
					ctfmt.Print(EntireYear[MN][W][I].fg, windowsFlag, EntireYear[MN][W][I].DateStr)
				} else { // need this construct because background set to black isn't really black.
					ct.Foreground(EntireYear[MN][W][I].fg, windowsFlag)
					ct.Background(EntireYear[MN][W][I].bg, windowsFlag)
					fmt.Fprint(ct.Writer, EntireYear[MN][W][I].DateStr)
					ct.ResetColor()
				}
			}
			fmt.Print("    ")
			for I := 0; I < 7; I++ { // day of week positions for 2nd month
				if EntireYear[MN2][W][I].bg == ct.Black {
					ctfmt.Print(EntireYear[MN2][W][I].fg, windowsFlag, EntireYear[MN2][W][I].DateStr)
				} else {
					ct.Foreground(EntireYear[MN2][W][I].fg, windowsFlag)
					ct.Background(EntireYear[MN2][W][I].bg, windowsFlag)
					fmt.Fprint(ct.Writer, EntireYear[MN2][W][I].DateStr)
					ct.ResetColor()
				}
			}
			fmt.Print("    ")
			for I := 0; I < 7; I++ { // day of week position for 3rd month
				if EntireYear[MN3][W][I].bg == ct.Black {
					ctfmt.Print(EntireYear[MN3][W][I].fg, windowsFlag, EntireYear[MN3][W][I].DateStr)
				} else {
					ct.Foreground(EntireYear[MN3][W][I].fg, windowsFlag)
					ct.Background(EntireYear[MN3][W][I].bg, windowsFlag)
					fmt.Fprint(ct.Writer, EntireYear[MN3][W][I].DateStr)
					ct.ResetColor()
				}
			}
			fmt.Println()
		} // END FOR W
	} // END FOR i
	fmt.Println()
	fmt.Println()
	fmt.Println(s)
	fmt.Println()
} // Show3MonthRow

// ----------------------------- HolidayAssign ---------------------------------

func HolidayAssign(year int) {

	var Holiday holidaycalc.HolType

	Holiday = holidaycalc.GetHolidays(year)
	Holiday.Valid = true

	/*
		fmt.Println(" Debugging holiday assign.")
		fmt.Println(Holiday)
		fmt.Print("hit <enter> to continue. ...")
		ans := ""
		fmt.Scanln(&ans)
		fmt.Println()
	*/

	// New Year's Day
	julian := timlibg.JULIAN(1, 1, year)
	DOW := julian % 7
	EntireYear[JAN][0][DOW].comment = "NYD"
	EntireYear[JAN][0][DOW].fg = ct.Yellow
	EntireYear[JAN][0][DOW].bg = ct.Black

	// MLK Day
	d := Holiday.MLK.D
MLKloop:
	for w := 1; w < 6; w++ { // start looking at 2nd week
		for dow := 0; dow < 7; dow++ {
			if EntireYear[JAN][w][dow].day == d {
				EntireYear[JAN][w][dow].comment = "MLK Day"
				EntireYear[JAN][w][dow].fg = ct.Yellow
				EntireYear[JAN][w][dow].bg = ct.Black
				break MLKloop
			}
		}
	}

	// President's Day
	d = Holiday.Pres.D
PresLoop:
	for w := 1; w < 6; w++ { // start looking at the 2nd week
		for dow := 0; dow < 7; dow++ {
			if EntireYear[FEB][w][dow].day == d {
				EntireYear[FEB][w][dow].comment = "Pres Day"
				EntireYear[FEB][w][dow].fg = ct.Yellow
				EntireYear[FEB][w][dow].bg = ct.Black
				break PresLoop
			}
		}
	}

	// Easter
	m := Holiday.Easter.M - 1 // convert to a zero origin system
	d = Holiday.Easter.D
EasterLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ {
			if EntireYear[m][w][dow].day == d {
				EntireYear[m][w][dow].comment = "Easter"
				EntireYear[m][w][dow].fg = ct.Yellow
				EntireYear[m][w][dow].bg = ct.Black
				break EasterLoop
			}
		}
	}

	// Mother's Day
	d = Holiday.Mother.D
MotherLoop:
	for w := 1; w < 6; w++ { // start looking at the 2nd week
		for dow := 0; dow < 7; dow++ {
			if EntireYear[MAY][w][dow].day == d {
				EntireYear[MAY][w][dow].comment = "Mom Day"
				EntireYear[MAY][w][dow].fg = ct.Yellow
				EntireYear[MAY][w][dow].bg = ct.Black
				break MotherLoop
			}
		}
	}

	// Memorial Day
	d = Holiday.Memorial.D
MemorialLoop:
	for w := 2; w < 6; w++ { // start looking at the 3rd week
		for dow := 0; dow < 7; dow++ {
			if EntireYear[MAY][w][dow].day == d {
				EntireYear[MAY][w][dow].comment = "Meml Day"
				EntireYear[MAY][w][dow].fg = ct.Yellow
				EntireYear[MAY][w][dow].bg = ct.Black
				break MemorialLoop
			}
		}
	}

	// Juneteenth
	d = 19
JuneteenthLoop:
	for w := 2; w < 6; w++ { // start looking at the 3rd week
		for dow := 0; dow < 7; dow++ {
			if EntireYear[JUN][w][dow].day == d {
				EntireYear[JUN][w][dow].comment = "Juneteenth"
				EntireYear[JUN][w][dow].fg = ct.Yellow
				EntireYear[JUN][w][dow].bg = ct.Black
				break JuneteenthLoop
			}
		}
	}

	// Father's Day
	d = Holiday.Father.D
FatherLoop:
	for w := 1; w < 6; w++ { // start looking at the 2nd week
		for dow := 0; dow < 7; dow++ {
			if EntireYear[JUN][w][dow].day == d {
				EntireYear[JUN][w][dow].comment = "Dad Day"
				EntireYear[JUN][w][dow].fg = ct.Yellow
				EntireYear[JUN][w][dow].bg = ct.Black
				break FatherLoop
			}
		}
	}

	// July 4th
	d = 4
IndependenceLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ {
			if EntireYear[JUL][w][dow].day == d {
				EntireYear[JUL][w][dow].comment = "Indpnc Day"
				EntireYear[JUL][w][dow].fg = ct.Yellow
				EntireYear[JUL][w][dow].bg = ct.Black
				break IndependenceLoop
			}
		}
	}

	// Labor Day
	d = Holiday.Labor.D
LaborLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ {
			if EntireYear[SEP][w][dow].day == d {
				EntireYear[SEP][w][dow].comment = "Labor Day"
				EntireYear[SEP][w][dow].fg = ct.Yellow
				EntireYear[SEP][w][dow].bg = ct.Black
				break LaborLoop
			}
		}
	}

	// Columbus Day
	d = Holiday.Columbus.D
ColumbusLoop:
	for w := 1; w < 6; w++ { // start looking at the 2nd week
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[OCT][w][dow].day == d {
				// EntireYear[OCT][w][dow].comment = "Columbus D"  not hospital holiday
				EntireYear[OCT][w][dow].fg = ct.Yellow
				EntireYear[OCT][w][dow].bg = ct.Black
				break ColumbusLoop
			}
		}
	}

	// Election Day
	d = Holiday.Election.D
ElectionLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[NOV][w][dow].day == d {
				EntireYear[NOV][w][dow].comment = "Electn Day"
				EntireYear[NOV][w][dow].fg = ct.Yellow
				EntireYear[NOV][w][dow].bg = ct.Black
				break ElectionLoop
			}
		}
	}

	// Veteran's Day
	d = 11
VeteranLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ {
			if EntireYear[NOV][w][dow].day == d {
				// EntireYear[NOV][w][dow].comment = "Vetrns Day"  not hospital holiday
				EntireYear[NOV][w][dow].fg = ct.Yellow
				EntireYear[NOV][w][dow].bg = ct.Black
				break VeteranLoop
			}
		}
	}

	// Thanksgiving Day
	d = Holiday.Thanksgiving.D
TGLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[NOV][w][dow].day == d {
				EntireYear[NOV][w][dow].comment = "ThanksGvg"
				EntireYear[NOV][w][dow].fg = ct.Yellow
				EntireYear[NOV][w][dow].bg = ct.Black
				break TGLoop
			}
		}
	}

	// Christmas Day
	d = 25
XmasLoop:
	for w := 0; w < 6; w++ {
		for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
			if EntireYear[DEC][w][dow].day == d {
				EntireYear[DEC][w][dow].comment = "Xmas Day"
				EntireYear[DEC][w][dow].fg = ct.Yellow
				EntireYear[DEC][w][dow].bg = ct.Black
				break XmasLoop
			}
		}
	}

	// Today
	if year == CurrentYear {
		d = TodaysDayNumber
		m = CurrentMonthNumber - 1 // convert to a zero origin system
	TodayLoop:
		for w := 0; w < 6; w++ {
			for dow := 0; dow < 7; dow++ { // note that this dow is a shadow of NYD dow
				if EntireYear[m][w][dow].day == d {
					EntireYear[m][w][dow].fg = ct.White
					EntireYear[m][w][dow].bg = ct.Red
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
	return month
}

// ----------------------------------- AssignYear ----------------------------------------------------

func AssignYear(y int) {

	if y < 1800 || y > 2100 {
		fmt.Printf("Year in AssignYear is %d, which is out of range of 1800..2100.  Exiting.\n", y)
		os.Exit(1)
	}

	JAN1DOW := timlibg.JULIAN(1, 1, y) % 7 // julian date number of Jan 1 of input year MOD 7.
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
	for m := JAN; m < NumOfMonthsInYear; m++ { // month position
		for wk := 0; wk < 6; wk++ { // week position
			for dayofweek := 0; dayofweek < 7; dayofweek++ {
				EntireYear[m][wk][dayofweek].DateStr = BLANKSTR3
				EntireYear[m][wk][dayofweek].day = 0
				EntireYear[m][wk][dayofweek].ch1 = '0'
				EntireYear[m][wk][dayofweek].ch2 = '0'
				EntireYear[m][wk][dayofweek].fg = ct.White
				EntireYear[m][wk][dayofweek].bg = ct.Black
			}
		}
	}

	// Make the calendar
	for MN := JAN; MN < NumOfMonthsInYear; MN++ {
		DATEASSIGN(MN)
	}
} // END AssignYear

/*
--------------------- MAIN ---------------------------------------------
*/
func main() {
	var Cal1FilenameFlag, Cal12FilenameFlag bool
	windowsFlag = runtime.GOOS == "windows" // intended for bold color on windows, not on linux

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

	Ext1Default := ".out"
	Ext12Default := ".xls"
	CurrentMonthNumber, TodaysDayNumber, CurrentYear = timlibg.TIME2MDY()

	ClearScreen()

	// flag definitions and processing
	var nofilesflag = flag.Bool("no", false, "do not generate output cal1 and cal12 files.") // Ptr
	var NoFilesFlag = flag.Bool("n", false, "do not generate output cal1 and cal12 files.")  // Ptr

	var helpflag = flag.Bool("h", false, "print help message.") // pointer
	var HelpFlag bool
	flag.BoolVar(&HelpFlag, "help", false, "print help message.")

	var testFlag bool
	flag.BoolVar(&testFlag,"test", false, "test mode flag.")
	flag.BoolVar(&testFlag, "v", false, "Verbose (test) mode.")

	flag.Parse()

	if *helpflag || HelpFlag {
		fmt.Println()
		fmt.Println(" Calendar Printing Program, last altered", LastCompiled)
		fmt.Println(" Usage: calg [flags] year month or month year, where month must be a month name string.")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(0)
	}

	fmt.Printf(" Calg, a calendar display program written in Go.  Last altered %s, using %s.\n", LastCompiled, runtime.Version())

	// process command line parameters
	RequestedMonthNumber = CurrentMonthNumber - 1 // default value converted to a zero origin reference.
	year = CurrentYear

	if flag.NArg() > 0 {
		commandline := flag.Args()
		if flag.NArg() > 2 {
			commandline = commandline[:2] // if there are too many params, only use params 0 and 1, ie, up to 2 but not incl'g 2.
		}
		for _, commandlineparam := range commandline {
			if unicode.IsDigit(rune(commandlineparam[0])) { // have numeric parameter, must be a year
				YEARSTR = commandlineparam
				var err error
				year, err = strconv.Atoi(YEARSTR)
				if err != nil {
					fmt.Println(" Error from Atoi for year.  Using CurrentYear.  Entered string is", commandlineparam)
					year = CurrentYear
					fmt.Print(" pausing.  Hit <enter> to continue")
					ans := ""
					fmt.Scanln(&ans)
					fmt.Println()
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
			}
		}
	} else {
		year = CurrentYear
	}

	if year < 100 && (year < (CurrentYear%100) || year < 30) {
		year += 2000
	} else if year < 100 {
		year += 1900
	}
	if year < 1800 || year > 2100 {
		fmt.Printf("Year is %d, which is out of range (1800-2100).  Exiting.\n", year)
		os.Exit(1)
	}

	YEARSTR = strconv.Itoa(year) // This will always be a 4 digit year, regardless of what's entered on command line.

	if RequestedMonthNumber > 6 { // If request after July, make it July because of the 6 month display.
		RequestedMonthNumber = 6
	}

	if testFlag {
		fmt.Println()
		fmt.Println(" using year", year, ", using month", MONNAMSHORT[RequestedMonthNumber])
		fmt.Println()
		ans := ""
		fmt.Print(" pausing, hit <enter> to resume")
		fmt.Scanln(&ans)
		fmt.Println()
	}

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
			fmt.Printf(" %s already exists.  From stat call file created %s, filesize is %d.\n",
				Cal1Filename, FI.ModTime().Format("Jan-02-2006 15:04:05"), FI.Size())
		} else {
			Cal1FilenameFlag = true
			fmt.Printf(" %s does not already exist.\n", Cal1Filename)
		}

		FI, err = os.Stat(Cal12Filename)
		if err == nil {
			fmt.Printf(" %s already exists.  From stat call file created %s, filesize is %d.\n",
				Cal12Filename, FI.ModTime().Format("Jan-02-2006 15:04:05"), FI.Size())
		} else {
			Cal12FilenameFlag = true
			fmt.Printf(" %s does not already exist.\n", Cal12Filename)
		}
	}

	if testFlag {
		fmt.Println()
		fmt.Println(" Completed year matrix.  AllowFilesFlag is", AllowFilesFlag, ".")
		fmt.Print(" pausing.  Hit <enter> to contiue.")
		ans := ""
		fmt.Scanln(&ans)
		fmt.Println()
	}

	var err error

	workingdir, _ := os.Getwd()
	execname, _ := os.Executable() // from memory, check at home
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	if testFlag {
		fmt.Printf(" %s was last linked on %s.  Working directory is %s. \n", ExecFI.Name(), LastLinkedTimeStamp, workingdir)
		fmt.Printf(" Full name of executable file is %s \n", execname)
	}

	var OutCal1, OutCal12 *os.File
	if Cal1FilenameFlag {
		OutCal1, err = os.Create(Cal1Filename)
		check(err, " Trying to create Cal1 output file")
		defer OutCal1.Close()
		OutCal1file = bufio.NewWriter(OutCal1)
		defer OutCal1file.Flush()
	}

	if Cal12FilenameFlag {
		OutCal12, err = os.Create(Cal12Filename)
		check(err, " Trying to create Cal12 output file")
		defer OutCal12.Close()
		OutCal12file = bufio.NewWriter(OutCal12)
		defer OutCal12file.Flush()
	}

	// write to file 12 page calendar, one month/page
	if Cal12FilenameFlag {
		for MN := JAN; MN <= DCM; MN++ {
			WrMonthForXL(MN)
		} // ENDFOR
		OutCal12file.Flush()
		OutCal12.Close()
	}

	// Write to file One Page Calendar
	if Cal1FilenameFlag {
		WrOnePageYear()
		OutCal1file.Flush()
		OutCal1.Close()
	}

	fmt.Println()
	fmt.Println()

	Show3MonthRow(RequestedMonthNumber)

	/*
			// Now disploy next year.  No file writing.  Min 10 lines/calendar.  Nevermind.
		    year++
		    YEARSTR = strconv.Itoa(year)
		    AssignYear(year)
		    HolidayAssign(year)
		    RequestedMonthNumber = 0
		    Show3MonthRow(RequestedMonthNumber)

	*/

} // end main func

// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
	if e != nil {
		log.Printf("%s : ", msg)
		panic(e)
	}
}

//END calg
