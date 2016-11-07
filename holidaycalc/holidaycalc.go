package holidaycalc;

import (
"timlibg"
);

/*
  REVISION HISTORY
  ----------------
  5 Apr 88 -- 1) Converted to M2 V3.03.
              2) Imported the REALLIB and CALLIB modules, and deleted their
                  code from here.  They were originally created here, but
                  it seems more appropriate to leave them in a central
                  module and import them.
  19 Mar 89 -- 1) Imported the newly created HPCALC to allow date arithmetic.
                  The rest of the code was modified to take advantage of this
                  new capability.
               2) Fixed a bug in the error reporting from the EASTER proc
                  so that a FOR index variable does not get assigned 0.
  30 Mar 89 -- 1) Fixed bug in the PR and HOL cmds that ignored 2 digit
                  years
               2) Added reminder to GREG help line to use quotes to force
                  the date to be taken as an ALLELSE TKNSTATE.
  26 Dec 90 -- 1) Utilized the GETTKNSTR procedure where appropriate.
               2) UL2 is used instead of UTILLIB.
               3) Added GETTKNEOL proc to deal with GREG & DOW cmds.
  25 Jul 93 -- 1) Dropping requirement of CALCCMD by passing cmdline thru
                   to GETRESULT.
               2) Allowed empty command line to quit after confirmation.
               3) Eliminated writing of trailing insignificant 0's from
                   arithmetic functions.
               4) Imported TKNRTNS instead of TOKENIZE.
               5) Eliminated need for GETTKNSTR by improving algorithm.
               6) Deleted GETTKNEOL proc as it is no longer used.  If
                   needed, it may be imported from TKNRTNS now.
  18 May 03 -- Conversion to Win32 using Stony Brook Modula-2 V 4
  25 Dec 14 -- Converted to a module to get holiday dates by the calculator, using HolMod written for the Cal program.
   1 Jan 15 -- Converting to cpp, and combining HolMod into this module.
  25 Aug 16 -- Converting to Go.
*/

type MDType struct {  // MDType is a contraction of Month Day Type, and Go export caps rules apply.
         M,D int;
}

type HolType struct {
         MLK,Pres,Easter,Mother,Memorial,Father,Labor,Columbus,Election,Thanksgiving MDType;
         Year int;
         Valid bool;
}


//******************************** SUBTDAYS ********************************

func SUBTDAYS(C, Y int) int {
/*
Subtract Days.
Computes how many days to subtract from the holiday depending on the year.
Days to Subtract = C + [5/4 Y] - [3/4 (1 + [Y/100])  ]) MOD 7
*/
  return ((C + (5*Y/4) - 3*(1 + (Y/100)) / 4) % 7);  // % is the MOD operator
}// SUBTDAYS

/*****************************************************************************************/
func CalcMLK(year int) (int) {
/*
 Find the date of MLK day by finding which day back from Jan 21 is a Monday.
*/


  day := 21;
  J := timlibg.JULIAN(1,day,year);
  for J % 7 != 1 {
    day--;
    J = timlibg.JULIAN(1,day,year);
  }
  return day;
}// CalcMLK

//************************************ EASTER ******************************
func EASTER(YEAR int) (MM, DD int) {
/*
EASTER.
This routine computes the golden number for that year, then Easter Sunday is
the first Sunday following this date.  If the date is a Sunday, then Easter
is the following Sunday.
*/

  if (YEAR < 1900) || (YEAR > 2500) {
    MM = 0;
    DD = 0;
  }else{
    GOLDENNUM := (YEAR % 19) + 1;
    switch GOLDENNUM {
      case  1: // APR 14
        MM = 4;
        DD = 14;
        break;
      case  2: // APR 3
        MM = 4;
        DD = 3;
        break;
      case  3: // MAR 23
        MM = 3;
        DD = 23;
        break;
      case  4: // APR 11
        MM = 4;
        DD = 11;
        break;
      case  5: // MAR 31
        MM = 3;
        DD = 31;
        break;
      case  6: // APR 18
        MM = 4;
        DD = 18;
        break;
      case  7: // APR 8
        MM = 4;
        DD = 8;
        break;
      case  8: // MAR 28
        MM = 3;
        DD = 28;
        break;
      case  9: // APR 16
        MM = 4;
        DD = 16;
        break;
      case 10: // APR 5
        MM = 4;
        DD = 5;
        break;
      case 11: // MAR 25
        MM = 3;
        DD = 25;
        break;
      case 12: // APR 13
        MM = 4;
        DD = 13;
        break;
      case 13: // APR 2
        MM = 4;
        DD = 2;
        break;
      case 14: // MAR 22
        MM = 3;
        DD = 22;
        break;
      case 15: // APR 10
        MM = 4;
        DD = 10;
        break;
      case 16: // MAR 30
        MM = 3;
        DD = 30;
        break;
      case 17: // APR 17
        MM = 4;
        DD = 17;
        break;
      case 18: // APR 7
        MM = 4;
        DD = 7;
        break;
      case 19: // MAR 27
        MM = 3;
        DD = 27;
    } // endcase on GoldenNum
  } // endif
/*
  Now find next Sunday.
*/
// if M/D/Y starts on a sunday, the holiday is 1 week later.  So I need a post incr value.
  JULDATE := timlibg.JULIAN(MM,DD,YEAR) +1;
  for (JULDATE % 7) != 0 {
    JULDATE++;
  }
  MM, DD, _ = timlibg.GREGORIAN(JULDATE);
  return MM, DD;
} // EASTER

//*****************************************************************************************

func GetHolidays(y int) (HolType) {

  Holidays := HolType{};

  if y < 1900 || y > 2100 {
    return Holidays;  // returning a zeroed out Holidays, including Valid field being false
  }

  Holidays.Year = y;

  Holidays.MLK.M = 1;
  Holidays.MLK.D = CalcMLK(y);

  Holidays.Pres.M = 2;
  Holidays.Pres.D = 21 - SUBTDAYS(2,y-1);

  Holidays.Easter.M, Holidays.Easter.D = EASTER(y);

  Holidays.Mother.M = 5;
  Holidays.Mother.D =  14 - SUBTDAYS(0,y);

  Holidays.Memorial.M = 5;
  Holidays.Memorial.D = 31 - SUBTDAYS(2,y);

  Holidays.Father.M = 6;
  Holidays.Father.D = 21 - SUBTDAYS(3,y);

  Holidays.Labor.M = 9;
  Holidays.Labor.D = 7 - SUBTDAYS(3,y);

  Holidays.Columbus.M = 10;
  Holidays.Columbus.D = 14 - SUBTDAYS(5,y);

  Holidays.Election.M = 11;
  Holidays.Election.D = 8 - SUBTDAYS(1,y);

  Holidays.Thanksgiving.M = 11;
  Holidays.Thanksgiving.D = 28 - SUBTDAYS(5,y);

  return Holidays;
} // GetHolidays

// END holidaycalc.go
