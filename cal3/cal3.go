// cal3.go
// Copyright (C) 1987-2016  Robert Solomon MD.  All rights reserved.

package main

/*
  REVISION HISTORY
  ----------------
   6 Apr 88 -- 1) Converted to M2 V3.03.
               2) Response to 12 page question is now echoed to the terminal.
               3) Module name changed to CAL so not to conflict with the
                   Logitech's CALENDAR library module.
   4 Nov 90 -- Updated the UTILLIB references to UL2, and recompiled under
               V3.4.
  28 Oct 91 -- Added FSA parse and indiv month printing abilities.
   2 Nov 91 -- Fixed problem w/ Zeller's congruence when Las2dgts is small
                enough to make the expression evaluate to a negative value.
  20 Jan 92 -- First page now does not begin with a FF.
  9 Nov 16 -- Converting to Go, using a CLI.  Input a year on the commandline, and output two files.
                A 1 page calendar meant for printing out, and a 12 page calendar meant for importing into Excel.
 10 Nov 16 -- First working version, based on Modula-2 code from 92.
 11 Nov 16 -- Code from January 2009 to import into Excel is working.  Now to try with week slices, and renamed to cal2.go
 12 Nov 16 -- Fixed bug in DATEASSIGN so first week in month is not skipped.
 12 Nov 16 -- Changed logic again to not use global for EntireYear[,,], just to see how this goes.
 13 Nov 16 -- Testing idiomatic Go to init DIM and MONNAMSHORT.  And removed DAYSNAMLONG as I don't use it.
*/

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	//
	"getcommandline"
	"timlibg"
	"tokenize"
)

const LastCompiled = "13 Nov 16"
const BLANKCHR = ' '
const HorizTab = 9                               // ASCII code, also ^I, or ctrl-I
const BlankLineWithTabs = "  	  	  	  	  	  	  " // There are embedded <tab> chars here, too

const ( // MNenum, ie, month number enumeration from the old Modula-2 code
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
const DCM = DEC // These are now synonyms for December Month Number = 11, as Jan = 0.

var OutputCal1, OutputCal12 os.File
var OutCal1file, OutCal12file *bufio.Writer
var YEARSTR, BLANKSTR2, BLANKSTR3 string
var Cal1Filename, Cal12Filename string
var DayNamesWithTabs, DAYSNAMSHORT string

// AllMonthsArray type subscripts are [MN] [W] [DOW]
// I will attempt to use week slices after I get a working excel version, just to see if I can.
// Then I won't need the WIM array.
type WeekVector [7]string

//  type MonthMatrix [6]WeekVector;  This is the array syntax
type MonthMatrix []WeekVector // This is slice syntax, just to see if I can get this to work.
type AllMonthsArray [NumOfMonthsInYear]MonthMatrix

//  var EntireYear AllMonthsArray;  Moved out of this global domain.
//                                          var MONTH Was ARRAY [JAN..DCM],[1..6],[1..7] OF STR10TYP in Modula-2

var MONNAMSHORT, MONNAMLONG [NumOfMonthsInYear]string

// ------------------------------------------------------- DAY2STR  -------------------------------------
func DAY2STR(DAY int) string {
	/*
	   DAY TO STRING conversion.
	   This routine will convert the DAY into a 3 char string.
	   If the first digit is zero, then that char will be blank, too.
	*/

	const digits = "0123456789"

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
	return string(bs) // not sure if this is best as a string or as a byteslice
} //END DAY2STR;

func DATEASSIGN(DOW, DIM int) (int, MonthMatrix) {
	/*
	   --------------------------------------------------------- DATEASSIGN -------------------------------------------
	   Date ASSIGNment for month.
	   This routine will assign the dates for an entire month.  It will put the char
	   representations of the date in the first 3 bytes.

	   Former INPUT FROM GBL VAR'S : DIM(MN), DOW
	   Former OUTPUT TO  GBL VAR'S : DOW, EntireYear[MN,,].  Formerly WIM(MN) but this is replaced by slices here now.

	   As type MonthMatrix is a slice of WeekVector, the MonthMatrix type is itself a pointer already.  I decided to
	   not worry about pointer params for this reason.
	*/
	var week WeekVector
	for w := range week {
		week[w] = BLANKSTR3 // This initialization now occurs here for entire week regardless of DOW.
	}
	aMonth := make(MonthMatrix, 0, 6) // will need to append weeks as we go to each month-by-month

	// No longer need a var like W to keep track of the # of weeks in a particular month.
	for date := 1; date <= DIM; date++ {
		if DOW > 6 { // DOW = 0 for Sunday.
			DOW = 0
			aMonth = append(aMonth, week)
			for w := range week {
				week[w] = "   " // Re-initialize entire week with 3 blanks per day.
			} // endfor w
		} // ENDIF DOW
		DATESTR := DAY2STR(date)
		week[DOW] = DATESTR
		DOW++
	} // ENDFOR date;
	if DOW > 0 {
		aMonth = append(aMonth, week)
	}
	if DOW > 6 { // Don't return a DOW > 6, as that will make a blank first week for next month.
		DOW = 0
	} // if DOW > 6

	return DOW, aMonth
} // END DATEASSIGN

// --------------------------------------------------------- PrMonthForXL --------------------------------------
// Intended to print in a format that can be read by Excel as a call schedule template.

func PrMonthForXL(MN int, aMonth MonthMatrix) {

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

	for W := range aMonth {
		_, err = OutCal12file.WriteString(aMonth[W][0]) // write out Sunday
		check(err, "")
		err = OutCal12file.WriteByte(HorizTab) // <tab>, or horizontal tab <HT>, to confirm that this works
		check(err, "")

		for I := 1; I < 6; I++ { // write out Monday .. Friday

			_, err = OutCal12file.WriteString(aMonth[W][I])
			check(err, "")
			_, err = OutCal12file.WriteRune('\t') // <tab>, or horizontal tab <HT>, to see if this works
			check(err, "")

		} // ENDFOR I

		_, err = OutCal12file.WriteString(aMonth[W][6]) // write out Saturday
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
} // END PrMonthForXL

/*
---------------------------------------------------------- MAIN ---------------------------------------------
*/
func main() {
	BLANKSTR2 = "  "
	BLANKSTR3 = "   "
	var year, DOW, W, JAN1DOW, FEBDAYS int
	var DIM [NumOfMonthsInYear]int // Don't need WIM array anymore.
	var MN, MN2, MN3 int           //  MNEnum Month Number Vars

	//                                                          type WeekVector [7]string;
	//                                                          type MonthMatrix []WeekVector;
	//                                                          type AllMonthsArray [NumOfMonthsInYear]MonthMatrix;
	var aMonth MonthMatrix
	var EntireYear AllMonthsArray
	fmt.Println("Calendar Printing Program in Go.  ", LastCompiled)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: cal <year>")
		os.Exit(0)
	}

	Ext1Default := ".out"
	Ext12Default := ".xls"

	commandline := getcommandline.GetCommandLineString()
	cleancommandline := filepath.Clean(commandline)
	tokenize.INITKN(cleancommandline)
	YearToken, _ := tokenize.GETTKN()
	if YearToken.State != tokenize.DGT {
		fmt.Println(" Numeric token not found on command line.  Exiting")
		os.Exit(1)
	}

	year = YearToken.Isum
	if year < 40 {
		year += 2000
	} else if year < 100 {
		year += 1900
	} else if year < 1900 || year > 2100 {
		fmt.Printf("Year is %d, which is out of range (1900-2100).  Exiting.\n")
		os.Exit(1)
	}
	YEARSTR = strconv.Itoa(year)

	BaseFilename := YearToken.Str
	Cal1Filename = BaseFilename + "_cal1" + Ext1Default
	Cal12Filename = BaseFilename + "_cal12" + Ext12Default

	fmt.Println(" Output Files are : ", Cal1Filename, Cal12Filename)
	fmt.Println()

	OutCal1, err := os.Create(Cal1Filename)
	check(err, " Trying to create Cal1 output file")
	defer OutCal1.Close()

	OutCal12, e := os.Create(Cal12Filename)
	check(e, " Trying to create Cal12 output file")
	defer OutCal12.Close()

	OutCal1file = bufio.NewWriter(OutCal1)
	defer OutCal1file.Flush()

	OutCal12file = bufio.NewWriter(OutCal12)
	defer OutCal12file.Flush()

	MONNAMSHORT = [NumOfMonthsInYear]string{"JANUARY", "FEBRUARY", "MARCH", "APRIL", "MAY", "JUNE", "JULY",
		"AUGUST", "SEPTEMBER", "OCTOBER", "NOVEMBER", "DECEMBER"}
	/*
	   MONNAMSHORT[JAN] = "JANUARY";
	   MONNAMSHORT[FEB] = "FEBRUARY";
	   MONNAMSHORT[MAR] = "MARCH";
	   MONNAMSHORT[APR] = "APRIL";
	   MONNAMSHORT[MAY] = "MAY";
	   MONNAMSHORT[JUN] = "JUNE";
	   MONNAMSHORT[JUL] = "JULY";
	   MONNAMSHORT[AUG] = "AUGUST";
	   MONNAMSHORT[SEP] = "SEPTEMBER";
	   MONNAMSHORT[OCT] = "OCTOBER";
	   MONNAMSHORT[NOV] = "NOVEMBER";
	   MONNAMSHORT[DCM] = "DECEMBER";
	*/
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

	DayNamesWithTabs = "SUNDAY \t MONDAY \t TUESDAY \t WEDNESDAY \t THURSDAY \t FRIDAY \t SATURDAY"

	DAYSNAMSHORT = "  S  M  T  W TH  F  S    "

	// DIM = Days In Month
	DIM = [NumOfMonthsInYear]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	/*
	   DIM[JAN] = 31;
	   DIM[MAR] = 31;
	   DIM[APR] = 30;
	   DIM[MAY] = 31;
	   DIM[JUN] = 30;
	   DIM[JUL] = 31;
	   DIM[AUG] = 31;
	   DIM[SEP] = 30;
	   DIM[OCT] = 31;
	   DIM[NOV] = 30;
	   DIM[DCM] = 31;
	*/
	MN = JAN

	JulDate := timlibg.JULIAN(1, 1, year)
	JAN1DOW = JulDate % 7
	DOW = JAN1DOW

	if ((year % 4) == 0) && ((year % 100) != 0) {
		// YEAR IS DIVISIBLE BY 4 AND NOT BY 100
		FEBDAYS = 29
	} else if (year % 400) == 0 {
		FEBDAYS = 29
	} else {
		// HAVE EITHER A NON-LEAP YEAR OR A CENTURY YEAR
		FEBDAYS = 28
	} // ENDIF about leap year
	DIM[FEB] = FEBDAYS

	// Time to make the calendar
	for MN := JAN; MN <= DCM; MN++ {
		DOW, aMonth = DATEASSIGN(DOW, DIM[MN])
		EntireYear[MN] = aMonth
	}

	// WRITE 12 PAGE CALENDAR, ONE MONTH PER PAGE
	for MN := JAN; MN <= DCM; MN++ {
		PrMonthForXL(MN, EntireYear[MN])
	} // ENDFOR

	// Write one page calendar
	s := fmt.Sprintf("%40s", YEARSTR)
	_, err = OutCal1file.WriteString(s)
	check(err, "Error while writing YEARSTR to Cal 1 file")
	_, err = OutCal1file.WriteRune('\n')
	check(err, "Error while writing a newline rune to Cal 1 file")

	for MN = JAN; MN <= DCM; MN += 3 {
		MN2 = MN + 1
		MN3 = MN + 2
		_, err = OutCal1file.WriteRune('\n')
		check(err, "Error while writing newline rune to Cal 1 file")
		if MN > JAN { // have fewer blank lines after year heading than btwn rows of months.
			_, err = OutCal1file.WriteRune('\n')
			check(err, "Error while writing newline rune to Cal 1 file")
			_, err = OutCal1file.WriteRune('\n')
			check(err, "Error while writing newline rune to Cal 1 file")
		}
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
		blanks21 := strings.Repeat(" ", 21)
		for W = 0; W < 6; W++ { // week number
			if W < len(EntireYear[MN]) {
				for I := 0; I < 7; I++ { // day of week positions for 1st month
					_, err = OutCal1file.WriteString(EntireYear[MN][W][I])
					check(err, "Error while writing date string to cal 1 file")
				} // endfor I
			} else {
				_, err = OutCal1file.WriteString(blanks21)
				check(err, "")
			} // ENDIF W
			_, err = OutCal1file.WriteString("    ")
			check(err, "")
			if W < len(EntireYear[MN2]) {
				for I := 0; I < 7; I++ { // day of week positions for 2nd month
					_, err = OutCal1file.WriteString(EntireYear[MN2][W][I])
					check(err, "Error while writing date string to cal 1 file")
				} // endfor I
			} else {
				_, err = OutCal1file.WriteString(blanks21)
				check(err, "")
			} // ENDIF W
			_, err = OutCal1file.WriteString("    ")
			check(err, "")
			if W < len(EntireYear[MN3]) {
				for I := 0; I < 7; I++ { // day of week position for 3rd month
					_, err = OutCal1file.WriteString(EntireYear[MN3][W][I])
					check(err, "Error while writing date string to cal 1 file")
				} // endfor I
			} else {
				_, err = OutCal1file.WriteString(blanks21)
				check(err, "")
			} // ENDIF W
			_, err = OutCal1file.WriteRune('\n')
			check(err, "Error while writing newline rune to Cal 1 file")
		} // ENDFOR W
	} // ENDFOR MN;
	_, err = OutCal1file.WriteRune('\n')
	check(err, "Error while writing newline rune to Cal 1 file")
	_, err = OutCal1file.WriteString(s)
	check(err, "Error while writing YEARSTR to Cal 1 file")
	_, err = OutCal1file.WriteRune('\n')
	check(err, "Error while writing newline rune to Cal 1 file")
} // end main func

// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Errorf("%s : ", msg)
		panic(e)
	}
}

//END CAL
