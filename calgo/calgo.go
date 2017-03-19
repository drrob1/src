// calgo.go
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
 11 Nov 16 -- Code from January 2009 to import into Excel is working.
 12 Nov 16 -- Fixed bug in DATEASSIGN caused by not porting my own Modula-2 code correctly.
  3 Mar 17 -- Now calgo, and will use termbox to try to do what CALm2 does.
*/


import (
  "os"
  "bufio"
  "fmt"
  "path/filepath"
  "strconv"
  "termbox"
//
  "getcommandline"
  "timlibg"
  "tokenize"
)

  const LastCompiled = "3 Mar 17";
  const BLANKCHR   = ' ';
  const HorizTab = 9;  // ASCII code, also ^I, or ctrl-I
  const BlankLineWithTabs = "  	  	  	  	  	  	  "; // There are embedded <tab> chars here, too

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
        );
  const DCM = DEC;  // These are now synonyms for December Month Number = 11, as Jan = 0.


  var OutputCal1,OutputCal12 os.File;
  var OutCal1file, OutCal12file *bufio.Writer;
  var PROMPT,ExtDefault,YEARSTR,BLANKSTR2,BLANKSTR3 string;
  var Cal1Filename,Cal12Filename string;
  var DAYSNAMLONG, DayNamesWithTabs, DAYSNAMSHORT string;
  var MN, MN2, MN3 int //  MNEnum Month Number Vars

// AllMonthsArray type subscripts are [MN] [W] [DOW]
// I will attempt to use week slices after I get a working excel version, just to see if I can.
// Then I won't need the WIM array.
  type WeekVector [7] string;
  type MonthMatrix [6] WeekVector;
  type AllMonthsArray [NumOfMonthsInYear] MonthMatrix;
  var EntireYear AllMonthsArray;
//                                          var MONTH Was ARRAY [JAN..DCM],[1..6],[1..7] OF STR10TYP in Modula-2


  var year,DOW,W,JAN1DOW,FEBDAYS int
  var DIM, WIM [NumOfMonthsInYear]int
  var MONNAMSHORT [NumOfMonthsInYear]string;
  var MONNAMLONG  [NumOfMonthsInYear]string;
  var clear map[string]func()


// ------------------------------------------------------- init -----------------------------------
func init() {  // start termbox in the init code doesn't work.  Don't know why.  But this init does work.
  clear = make(map[string]func());
  clear["linux"] = func() {  // this is a closure, or an anonymous function
    cmd := exec.Command("clear");
    cmd.Stdout = os.Stdout;
    cmd.Run();
  }

  clear["windows"] = func() {  // this is a closure, or an anonymous function
    cmd := exec.Command("cmd","/c","cls");
    cmd.Stdout = os.Stdout;
    cmd.Run();
  }
}

// --------------------------------------------------- Cap -----------------------------------------
func Cap(c rune) rune {
  r,_,_,_ := strconv.UnquoteChar(strings.ToUpper(string(c)),0);
  return r;
} // Cap

// --------------------------------------------------- Print_tb -----------------------------------
func Print_tb(x,y int, fg,bg termbox.Attribute, msg string) {
  for _,c := range msg {
    termbox.SetCell(x,y,c,fg,bg);
    x++;
  }
  ClearEOL(x,y);
  e := termbox.Flush();
  if e != nil {
    panic(e);
  }
}

//----------------------------------------------------- Printf_tb ---------------------------------
func Printf_tb(x,y int, fg,bg termbox.Attribute, format string, args ...interface{}) {
  s := fmt.Sprintf(format,args...);
  Print_tb(x,y,fg,bg,s);
}

// ----------------------------------------------------- ClearLine -----------------------------------
func ClearLine(y int) {
  if y > MaxRow {
    y = MaxRow
  }
  for x := StartCol; x <= MaxCol; x++ {
    termbox.SetCell(x,y,0,Black,Black);  // Don't know if it matters if the char is ' ' or nil.
  }
  err := termbox.Flush();
  check(err);
}  // end ClearLine

// ----------------------------------------------------- HardClearScreen -----------------------------
func HardClearScreen () {
  err := termbox.Clear(Black,Black);
  check(err);
  for row := StartRow; row <= MaxRow; row ++ {
    ClearLine(row);
  }
  err = termbox.Flush();
  check(err);
}

// ------------------------------------------------------ ClearEOL -----------------------------------
func ClearEOL(x,y int) {
  if y > MaxRow {
    y = MaxRow
  }
  if x > MaxCol {
    return
  }
  for i := x; i <= MaxCol; i++ {
    termbox.SetCell(i,y,0,Black,Black);  // Don't know if it matters if the char is ' ' or nil.
  }
  err := termbox.Flush();
  check(err);
}


// ------------------------------------------------------- Repaint ----------------------------------
func RepaintScreen(x int) {

  Printf_tb(x,TitleRow,BrightCyan,Black," HP-type RPN calculator written in Go.  Last compiled %s",LastCompiled);
  WriteStack(x,StackRow);
  n = WriteRegToScreen(x,RegRow);
  if n > 8 {
    OutputRow = RegRow + n + 3// So there is enough space for all the reg's to be displayed above the output
    PromptRow = OutputRow -1;
  }
  WriteDisplayTapeToScreen(DisplayCol,StackRow);
  Printf_tb(x,MaxRow-1,BrightCyan,Black,Divider);
}


// ---------------------------------------------------- ClearScreen ------------------------------------
func ClearScreen() {
  clearfunc, ok := clear[runtime.GOOS]
  if ok {
    clearfunc();
  }else{  // unsupported platform
    panic(" The ClearScreen platform is only supported on linux or windows, at the moment");
  }
}

// ------------------------------------------------------- DAY2STR  -------------------------------------
func DAY2STR(DAY int) string {
/*
DAY TO STRING CONVERSION.
THIS ROUTINE WILL CONVERT THE 2 DIGIT DAY INTO A 2 CHAR STRING.
IF THE FIRST DIGIT IS ZERO, THEN THAT CHAR WILL BE BLANK.
*/

const digits = "0123456789"
const ZERO = '0';

  bs := make([]byte,3);

  TENSDGT := DAY / 10;
  UNTSDGT := DAY % 10;
  bs[0] = BLANKCHR;
  if TENSDGT == 0 {
    bs[1] = BLANKCHR;
  }else{
    bs[1] = digits[TENSDGT];
  }
  bs[2] = digits[UNTSDGT];
  return string(bs);  // not sure if this is best as a string or as a byteslice
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

  W := 0; // W is for Week number, IE, which week of the month is this.
  for DATE := 1; DATE <= DIM[MN]; DATE++ {
    if DOW > 6 {  // DOW = 0 for Sunday.
      W++
      DOW = 0;
    } // ENDIF
    DATESTR := DAY2STR(DATE);
    EntireYear[MN] [W] [DOW] = DATESTR;
    DOW++
  } // ENDFOR;
  WIM[MN] = W;  /* Return number of weeks in this month */
  if DOW > 6 { // Don't return a DOW > 6, as that will make a blank first week for next month.
    DOW = 0;
  } // if DOW > 6
} // END DATEASSIGN


// ----------------------------------------------------------- PRMONTH --------------------------------------
func PRMONTH(MN int ) { // Originally intended to print one month per page.

  s0 := fmt.Sprintf("%40s",MONNAMSHORT[MN]);
  s1 := fmt.Sprintf("%6s",YEARSTR);
  _, err := OutCal12file.WriteString(s0);
  check(err,"Error while writing month name short for big calendar");
  _, err = OutCal12file.WriteString(s1);
  check(err,"Error while writing yearstr for big calendar");
  _, err = OutCal12file.WriteRune('\n');
  check(err,"");
  _, err = OutCal12file.WriteRune('\n');
  check(err,"");
  _, err = OutCal12file.WriteString(DAYSNAMLONG);
  check(err,"");
  _, err = OutCal12file.WriteRune('\n');
  check(err,"");
  _, err = OutCal12file.WriteRune('\n');
  check(err,"");
  for W := 0; W <= WIM[MN]; W++ {
    _, err = OutCal12file.WriteString(" ");
    check(err,"");
    _, err = OutCal12file.WriteString(EntireYear[MN] [W] [0]); // write out Sunday
    check(err,"");
    _, err = OutCal12file.WriteString("      ");
    check(err,"");
    for I := 1; I < 6; I++ { // write out Monday .. Friday
      _, err = OutCal12file.WriteString(" ");
      check(err,"");
      _, err = OutCal12file.WriteString(EntireYear[MN] [W] [I]);
      _, err = OutCal12file.WriteString("        "); // FWRBL(OUTUN1,8);
      check(err,"");
    } // ENDFOR I
    _, err = OutCal12file.WriteString(" ");
    check(err,"");
    _, err = OutCal12file.WriteString(EntireYear[MN] [W] [6]); // write out Saturday
    _, err = OutCal12file.WriteRune('\n');
    check(err,"");
  } // ENDFOR W;
} // END PRMONTH

// ----------------------------------------------------------- PrMonthForXL --------------------------------------
// Intended to print in a format that can be read by Excel as a call schedule template.
func PrMonthForXL(MN int) {

  s0 := fmt.Sprintf("%s",MONNAMSHORT[MN]);
  s1 := fmt.Sprintf("\t%6s",YEARSTR);     // I'm going to add <tab> here to see if I like this effect
  _, err := OutCal12file.WriteString(s0);
                                       check(err,"Error while writing month name short for big calendar");
  _, err = OutCal12file.WriteString(s1);
                                                check(err,"Error while writing yearstr for big calendar");
  _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");
  _, err = OutCal12file.WriteString(DayNamesWithTabs);
                                                check(err,"");
  _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");

  for W := 0; W <= WIM[MN]; W++ {
    _, err = OutCal12file.WriteString(EntireYear[MN] [W] [0]); // write out Sunday
                                                check(err,"");
    err = OutCal12file.WriteByte(HorizTab); // <tab>, or horizontal tab <HT>, to confirm that this does work
                                                check(err,"");

    for I := 1; I < 6; I++ {                                  // write out Monday .. Friday

      _, err = OutCal12file.WriteString(EntireYear[MN] [W] [I]);
                                                check(err,"");
      _, err = OutCal12file.WriteRune('\t'); // <tab>, or horizontal tab <HT>, to see if this works
                                                check(err,"");


    } // ENDFOR I

    _, err = OutCal12file.WriteString(EntireYear[MN] [W] [6]); // write out Saturday
                                                check(err,"");
    _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");
    _, err = OutCal12file.WriteString(BlankLineWithTabs);
                                                check(err,"");
    _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");
    _, err = OutCal12file.WriteString(BlankLineWithTabs);
                                                check(err,"");
    _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");
  } // ENDFOR W
  _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");
  _, err = OutCal12file.WriteRune('\n');
                                                check(err,"");
} // END PrMonthForXL


/*
--------------------- MAIN ---------------------------------------------
*/
func main() {
  BLANKSTR2 = "  ";
  BLANKSTR3 = "   ";
  fmt.Println("Calendar Printing Program.  ",LastCompiled);
  fmt.Println();

  if len(os.Args) <=1 {
    fmt.Println(" Usage: cal <year>");
    os.Exit(0);
}

  PROMPT = " Enter Year : ";
  Ext1Default := ".out";
  Ext12Default := ".xls";

  commandline := getcommandline.GetCommandLineString();
  cleancommandline := filepath.Clean(commandline);
  tokenize.INITKN(cleancommandline);
  YearToken,_ := tokenize.GETTKN();
  if YearToken.State != tokenize.DGT {
    fmt.Println(" Numeric token not found on command line.  Exiting");
    os.Exit(1);
  }

  year = YearToken.Isum;
  if year < 40 {
    year += 2000;
  }else if year < 100 {
    year += 1900;
  }else if year < 1900 || year > 2100 {
    fmt.Printf("Year is %d, which is out of range (1900-2100).  Exiting.\n");
    os.Exit(1)
  }
  YEARSTR = strconv.Itoa(year);


  BaseFilename := YearToken.Str;
  Cal1Filename = BaseFilename + "_cal1" + Ext1Default;
  Cal12Filename = BaseFilename + "_cal12" + Ext12Default;

  fmt.Println(" Output Files are : ",Cal1Filename,Cal12Filename);
  fmt.Println();


  OutCal1,err := os.Create(Cal1Filename);
  check(err," Trying to create Cal1 output file");
  defer OutCal1.Close();

  OutCal12,e := os.Create(Cal12Filename);
  check(e," Trying to create Cal12 output file");
  defer OutCal12.Close();


  OutCal1file = bufio.NewWriter(OutCal1);
  defer OutCal1file.Flush();

  OutCal12file = bufio.NewWriter(OutCal12);
  defer OutCal12file.Flush();

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

  MONNAMLONG[JAN] = "    J A N U A R Y        ";
  MONNAMLONG[FEB] = "   F E B R U A R Y       ";
  MONNAMLONG[MAR] = "      M A R C H          ";
  MONNAMLONG[APR] = "      A P R I L          ";
  MONNAMLONG[MAY] = "        M A Y            ";
  MONNAMLONG[JUN] = "       J U N E           ";
  MONNAMLONG[JUL] = "       J U L Y           ";
  MONNAMLONG[AUG] = "     A U G U S T         ";
  MONNAMLONG[SEP] = "  S E P T E M B E R      ";
  MONNAMLONG[OCT] = "    O C T O B E R        ";
  MONNAMLONG[NOV] = "   N O V E M B E R       ";
  MONNAMLONG[DCM] = "   D E C E M B E R       ";

  DAYSNAMLONG = "SUNDAY    MONDAY      TUESDAY     WEDNESDAY   THURSDAY    FRIDAY      SATURDAY";
  DayNamesWithTabs = "SUNDAY \t MONDAY \t TUESDAY \t WEDNESDAY \t THURSDAY \t FRIDAY \t SATURDAY";

  DAYSNAMSHORT = "  S  M  T  W TH  F  S    ";


// DIM = Days In Month
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

  MN = JAN;

  JulDate := timlibg.JULIAN(1,1,year);
  JAN1DOW = JulDate % 7;
  DOW = JAN1DOW;

  if ((year % 4) == 0) && ((year % 100) != 0) {
// YEAR IS DIVISIBLE BY 4 AND NOT BY 100 
    FEBDAYS = 29;
  }else if (year % 400) == 0 {
    FEBDAYS = 29;
  }else{
// HAVE EITHER A NON-LEAP YEAR OR A CENTURY YEAR 
    FEBDAYS = 28;
  } // ENDIF about leap year
  DIM[FEB] = FEBDAYS;

// Initialize the calendar to all BLANKSTR3, for correct spacing
  for m := JAN; m <= DEC; m++ { // month position
    for wk := 0; wk < 6; wk++ { // week position
      for dayofweek := 0; dayofweek < 7; dayofweek++ {
        EntireYear[m] [wk] [dayofweek] = BLANKSTR3;
      }
    }
  }

// Time to make the calendar

  for MN := JAN; MN <= DCM; MN++ {
    DATEASSIGN(MN);
  } // ENDFOR;


// WRITE 12 PAGE CALENDAR, ONE MONTH PER PAGE 
  for MN := JAN; MN <= DCM; MN++ {
        PrMonthForXL(MN);
  } // ENDFOR

// Write one page calendar
  s := fmt.Sprintf("%40s",YEARSTR);
  _, err = OutCal1file.WriteString(s);
  check(err,"Error while writing YEARSTR to Cal 1 file");
  _, err = OutCal1file.WriteRune('\n');
  check(err,"Error while writing a newline rune to Cal 1 file");

  for MN = JAN; MN <= DCM; MN += 3 {
    MN2 = MN + 1;
    MN3 = MN + 2;

    _, err = OutCal1file.WriteRune('\n');
    check(err,"Error while writing newline rune to Cal 1 file");
    if MN > JAN {  // have fewer blank lines after year heading than btwn rows of months.
      _, err = OutCal1file.WriteRune('\n');
      check(err,"Error while writing newline rune to Cal 1 file");
      _, err = OutCal1file.WriteRune('\n');
      check(err,"Error while writing newline rune to Cal 1 file");
    }
    _, err = OutCal1file.WriteString(MONNAMLONG[MN]);
    check(err,"Error writing first long month name to cal 1 file");
    _, err = OutCal1file.WriteString(MONNAMLONG[MN2]);
    check(err,"");
    _, err = OutCal1file.WriteString(MONNAMLONG[MN3]);
    check(err,"");
    _, err = OutCal1file.WriteRune('\n');
    check(err,"Error while writing newline rune to Cal 1 file");
    _, err = OutCal1file.WriteRune('\n');
    check(err,"Error while writing newline rune to Cal 1 file");
//    _, err = OutCal1file.WriteRune('\n');                         too many blank lines
//    check(err,"Error while writing newline rune to Cal 1 file");
    _, err = OutCal1file.WriteString(DAYSNAMSHORT);
    check(err,"Error while writing day names to cal 1 file");
    _, err = OutCal1file.WriteString(DAYSNAMSHORT);
    check(err,"Error while writing day names to cal 1 file");
    _, err = OutCal1file.WriteString(DAYSNAMSHORT);
    check(err,"Error while writing day names to cal 1 file");
    _, err = OutCal1file.WriteRune('\n');
    check(err,"Error while writing newline rune to Cal 1 file");
    for W = 0; W < 6; W++ { // week number
      for I := 0; I < 7; I++ { // day of week positions for 1st month
        _, err = OutCal1file.WriteString(EntireYear[MN] [W] [I]);
        check(err,"Error while writing date string to cal 1 file");
      } // ENDFOR I
      _,err = OutCal1file.WriteString("    ");
      check(err,"");
      for I := 0; I < 7; I++ { // day of week positions for 2nd month
        _, err = OutCal1file.WriteString(EntireYear[MN2] [W] [I]);
        check(err,"Error while writing date string to cal 1 file");
      } // ENDFOR I
      _,err = OutCal1file.WriteString("    ");
      check(err,"");
      for I := 0; I < 7; I++ { // day of week position for 3rd month 
        _, err = OutCal1file.WriteString(EntireYear[MN3] [W] [I]);
        check(err,"Error while writing date string to cal 1 file");
      } // ENDFOR I
      _, err = OutCal1file.WriteRune('\n');
      check(err,"Error while writing newline rune to Cal 1 file");
    } // ENDFOR W
  } // ENDFOR MN;
  _, err = OutCal1file.WriteRune('\n');
  check(err,"Error while writing newline rune to Cal 1 file");
  _, err = OutCal1file.WriteString(s);
  check(err,"Error while writing YEARSTR to Cal 1 file");
  _, err = OutCal1file.WriteRune('\n');
  check(err,"Error while writing a newline rune to Cal 1 file");
} // end main func


// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
  if e != nil {
    fmt.Errorf("%s : ",msg)
    panic(e);
  }
}




//END CAL


