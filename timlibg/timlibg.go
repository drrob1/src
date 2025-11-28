package timlibg

import (
	"fmt"
	"strconv"
	"time"
)

/*
--  REVISION HISTORY
--  ----------------
--  14 Apr 92 -- Created JULIAN and GREGORIAN procs, which are accurate beyond 3/1/2100.
--  25 Jul 93 -- Changed GREG2JUL and JUL2GREG limits to those imposed by the
--                algorithm, ie, only years <2100 are now allowed.
--  25 Jul 94 -- Changed limits of allowed years to 1700 from 1900.
--  10 Nov 02 -- Converted to SBM2 Win v4.
--  17 May 03 -- First Win32 version.
--  26 May 03 -- Adjusted algorithm for Julian fcn so year has a pivot.
--   6 Oct 13 -- Converted to gm2.
--  19 Jan 14 -- Converted to Ada.
    17 Nov 14 -- Converted to C++.
     7 Dec 14 -- Removed ElapsedSec from the exported data type.  See timlibc.h.
     9 Jul 16 -- Fixed bug in Gregorian in which there is an infinite loop if juldate is too small.
     6 Aug 16 -- Started converting to Go.
    13 Apr 17 -- golint complained, so I added some comments
     6 Aug 23 -- The julian date number is of type int.  This is a reminder comment.
    24 May 24 -- Adding the comments so that go doc will work.
     7 Nov 24 -- Adding converting to/from Excel-based date numbers.  Day zero was 12/31/1899, so day one was 1/1/1900.
	28 Nov 25 -- Fixed bug in JULIAN because "y" was not assigned correctly.
*/

// DateTimeType -- fields are Rawtime, Month, Day, Year, Hours, Minutes, Seconds, Nanosec, MonthStr and DayOfWeekStr.
type DateTimeType struct {
	Rawtime                                   time.Time
	Month, Day, Year, Hours, Minutes, Seconds int
	Nanosec                                   int64
	MonthStr, DayOfWeekStr                    string
}

var ( // I tried declaring these as const but this would not compile.
	// DayNames Sunday .. Saturday
	DayNames = [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	// MonthNames January .. December
	MonthNames = [...]string{"", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

	// ADIPM is Accumulated Days in Previous Month
	ADIPM = [...]int{0, 1, -1, 0, 0, 1, 1, 2, 3, 3, 4, 4}
)

//  ADIPM is a constant that represents the difference btwn the last day of the previous month and 30, assuming each month was 30 days long.
//  The variable name is an acronym of Accumulated Days In Previous Months.

//	*********************************** TIME2MDY *************************
//
// TIME2MDY System Time To Month, Day, and Year Conversion.
func TIME2MDY() (MM, DD, YY int) {

	var DateTime DateTimeType
	DateTime.Rawtime = time.Now()

	MM = int(DateTime.Rawtime.Month()) // +1 is not needed as January =1.
	DD = DateTime.Rawtime.Day()
	YY = DateTime.Rawtime.Year()
	return
} // TIME2MDY

// --------------------------------------- GetDateTime

// GetDateTime fills the DateTimeType structure.
func GetDateTime() DateTimeType {
	var DateTime DateTimeType

	DateTime.Rawtime = time.Now()

	DateTime.Month = int(DateTime.Rawtime.Month())
	DateTime.Day = DateTime.Rawtime.Day()
	DateTime.Year = DateTime.Rawtime.Year()
	DateTime.Hours = DateTime.Rawtime.Hour()
	DateTime.Minutes = DateTime.Rawtime.Minute()
	DateTime.Seconds = DateTime.Rawtime.Second()
	DateTime.Nanosec = DateTime.Rawtime.UnixNano()
	DateTime.MonthStr = DateTime.Rawtime.Month().String()
	DateTime.DayOfWeekStr = DayNames[JULIAN(DateTime.Month, DateTime.Day, DateTime.Year)%7]
	return DateTime
} // GetDateTime

// ***************************************** MDY2STR ***************************************************

// MDY2STR -- Input Month Day Year in, output String "mm/dd/yyyy".
func MDY2STR(M, D, Y int) string {

	const DateSepChar = "/"

	MSTR := strconv.Itoa(M)
	DSTR := strconv.Itoa(D)
	YSTR := strconv.Itoa(Y)
	IntermedStr := MSTR + DateSepChar + DSTR + DateSepChar + YSTR
	return IntermedStr
} // MDY2STR

// ----------------------------------------------- JULIAN

// JULIAN -- Since the numbers are < 800,000, 32-bit int would be enough, but this returns int.
func JULIAN(M, D, Y int) int {

	_, _, YY := TIME2MDY()
	YearPivot := YY%100 + 1

	if Y < YearPivot {
		Y += 2000
	} else if Y < 100 {
		Y += 1900
	}
	Y0 := Y - 1

	// Month, Day or Year is out of range
	if (M < 1) || (M > 12) || (D < 1) || (D > 31) || (Y < 1700) || (Y > 2500) {
		Juldate := 0
		return Juldate
	} // if stuff is out of range

	M0 := M - 1

	// Y0 and M0 are the previous years and months, then we have to add in the current partial year and month.
	// Y0/4 is the number of possible leap years since the year 1
	// Y0/100 is the number of century years since the year 1 that have to be subtracted
	// Y0/400 is the number of leap century years since the year 1 that have to be added back

	Juldate := Y0*365 + Y0/4 - Y0/100 + Y0/400 + ADIPM[M0] + M0*30 + D

	if (((Y%4 == 0) && (Y%100 != 0)) || (Y%400 == 0)) && (M > 2) { // add in current year if it's a leap year
		Juldate++
	}
	return Juldate
} // JULIAN

// ----------------------------------------- GREGORIAN

// GREGORIAN -- Input Juldate int, output M,D,Y int
func GREGORIAN(Juldate int) (M, D, Y int) {

	const MinJuldate = 630000

	if Juldate <= MinJuldate { // Found this bug 07/09/2016.  Else get infinite loop.
		M = 0
		D = 0
		Y = 0
		return
	}

	Y0 := Juldate / 365
	M0 := 1
	D0 := 1

	for JULIAN(M0, D0, Y0) > Juldate {
		Y0--
	}

	M0 = 12
	for JULIAN(M0, D0, Y0) > Juldate {
		M0--
	}

	for JULIAN(M0, D0, Y0) < Juldate {
		D0++
	}

	M = M0
	D = D0
	Y = Y0
	return
} // GREGORIAN

// FromExcelToJul will take the 5 digit date number from Excel, and return my 6-digit number that is typically 730,000+.
func FromExcelToJul(xl int) int {
	juldate := xl + 693594
	return juldate
}

// FromExcelToGreg converts from Excel date number and returns m, d, y.
func FromExcelToGreg(xl int) (int, int, int) {
	juldate := FromExcelToJul(xl)
	m, d, y := GREGORIAN(juldate)
	return m, d, y
}

// FromExcelToDateOnly converts from Excel date number and returns YYYY-MM-DD
func FromExcelToDateOnly(xl int) string {
	m, d, y := FromExcelToGreg(xl)

	s := fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	return s
}

// END timlibg
