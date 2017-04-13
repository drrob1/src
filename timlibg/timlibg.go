package timlibg;

import (
        "time"
        "strconv"
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
*/


/*
 tm is a c standard datatype.
  time_t t=time(0);  absolute time in seconds, or -1 if unknown
  tm POINTER p = gmtime( ADROF t);   usually indicated as tm* p = whatever
    sec,min,hour,mday,mon (0-11), year (subt 1900), wday, yday, isdst

struct tm IS
  tm_sec    int	seconds after the minute  0-60*
  tm_min    int	minutes after the hour    0-59
  tm_hour   int	hours since midnight      0-23
  tm_mday   int	day of the month          1-31
  tm_mon    int	months since January      0-11
  tm_year   int	years since 1900
  tm_wday   int	days since Sunday         0-6
  tm_yday   int	days since January 1      0-365
  tm_isdst  int	Daylight Saving Time flag
END


For historical reasons, it is generally implemented as an integral value representing the number of seconds 
elapsed since 00:00 hours, Jan 1, 1970 UTC (i.e., a unix timestamp). Although libraries may implement this 
type using alternative time representations.
*/

type  DateTimeType struct { // golint wants a comment here.  I don't think I need one.
    Rawtime time.Time;
    Month,Day,Year,Hours,Minutes,Seconds int;
    Nanosec int64;
    MonthStr,DayOfWeekStr string;
  }


var (  // I tried declaring these as const but this would not compile.
  DayNames = [...]string{"Sunday","Monday","Tuesday","Wednesday", "Thursday","Friday","Saturday"};  // golint wants a comment here
  MonthNames = [...]string{"","January","February","March","April","May", "June","July","August", "September","October","November","December"};
  ADIPM = [...]int{0,1,-1,0,0,1,1,2,3,3,4,4};  // Accumulated Days in Previous Month
    )

//  ADIPM is a typed constant that represents the difference btwn the last day
//  of the previous month and 30, assuming each month was 30 days long.
//  The variable name is an acronym of Accumulated Days In Previous Months.



//              *********************************** TIME2MDY *************************
// TIME2MDY System Time To Month, Day, and Year Conversion.
func TIME2MDY()(MM, DD, YY int) {

  var DateTime DateTimeType;
  DateTime.Rawtime = time.Now();

  MM = int(DateTime.Rawtime.Month()) // +1 is not needed as January =1.
  DD = DateTime.Rawtime.Day();
  YY = DateTime.Rawtime.Year();
  return;
}// TIME2MDY

// **************************************************** GetDateTime ***********************************
// GetDateTime fills the structure.
func GetDateTime() DateTimeType {
  var DateTime DateTimeType;

  DateTime.Rawtime = time.Now();

  DateTime.Month = int(DateTime.Rawtime.Month());
  DateTime.Day = DateTime.Rawtime.Day();
  DateTime.Year = DateTime.Rawtime.Year();
  DateTime.Hours = DateTime.Rawtime.Hour();
  DateTime.Minutes = DateTime.Rawtime.Minute();
  DateTime.Seconds = DateTime.Rawtime.Second();
  DateTime.Nanosec = DateTime.Rawtime.UnixNano();
  DateTime.MonthStr = DateTime.Rawtime.Month().String();
  DateTime.DayOfWeekStr = DayNames[JULIAN(DateTime.Month,DateTime.Day,DateTime.Year) % 7];
  return DateTime;
}// GetDateTime

// ***************************************** MDY2STR ***************************************************
// MDY2STR Month Day Year Cardinals To String.  By both returning a string as a param and as a function I have
func MDY2STR(M, D, Y int) string{

  const DateSepChar = "/";
//  var MSTR,DSTR,YSTR string;
//  var IntermedStr string;


  MSTR := strconv.Itoa(M);
  DSTR := strconv.Itoa(D);
  YSTR := strconv.Itoa(Y);
  IntermedStr := MSTR + DateSepChar + DSTR + DateSepChar + YSTR;
  return IntermedStr;
} // MDY2STR

// ************************************************ JULIAN **********************************
// JULIAN used to need longint or longcard.  Since the numbers are < 800,000, regular 32 bit int are enough.
func JULIAN(M, D, Y int) int {

 var (
    M0,Y0 int;
    Juldate int;
)
  _, _, YY := TIME2MDY();
  YearPivot := YY % 100 + 1;


  if Y < YearPivot {
    Y0 = Y + 2000 - 1;
  }else if (Y < 100) {
    Y0 = Y + 1900 - 1;
  }else{
    Y0 = Y - 1;
  } // if Y

// Month, Day or Year is out of range
  if (M < 1) || (M > 12) || (D < 1) || (D > 31) || (Y < 1700) || (Y > 2500) {
    Juldate := 0;
    return Juldate;
  } // if stuff is out of range

  M0 = M - 1;

  Juldate =  Y0 * 365 +      // Number of days in previous normal years
             Y0 / 4 -       // Number of possible leap days
             Y0 / 100 +     // Subtract all century years
             Y0 / 400 +     // Add back the true leap century years
             ADIPM[M0] + M0 * 30 + D;

  if ((( Y % 4 == 0) && ( Y % 100 != 0)) || ( Y % 400 == 0)) &&  (M > 2) {
    Juldate++
  } // if have to increment Juldate
  return Juldate;
}// JULIAN

// **************************************** GREGORIAN ****************************************
func GREGORIAN(Juldate int) (M,D,Y int) {

  const MinJuldate = 630000;
//  var Y0,M0,D0 int;

  if Juldate <= MinJuldate {            // Found this bug 07/09/2016.  Else get infinite loop.
    M = 0;
    D = 0;
    Y = 0;
    return;
  };

  Y0 := Juldate / 365;
  M0 := 1;
  D0 := 1;

  for (JULIAN(M0,D0,Y0) > Juldate) { Y0-- }

  M0 = 12;
  for (JULIAN(M0,D0,Y0) > Juldate) { M0-- }

  for (JULIAN(M0,D0,Y0) < Juldate) { D0++ }

  M = M0;
  D = D0;
  Y = Y0;
  return;
}// GREGORIAN

// END timlibc
