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
  9 Nov 16 -- Converting to Go, using a CLI.  It will take a year on the commandline, and output two files.
                A 1 page calendar meant for printing out, and a 12 page calendar meant for importing into
                Excel.
*/


import (
  "os"
  "fmt"
  "path/filepath"
//
  "getcommandline"
  "timlibg"
  "tokenize"
)

  const LastCompiled = "12 Nov 16";
  const LNSIZ      = 80;
  const LNPPAG     = 60; // Used by PRMONTH proc.  Top 6 lines used by Month and day names.
  const FF         = 12
  const CR         = 13;
  const ESC        = 27;
//      NULL       = 0C;
  const BLANKCHR   = ' ';
  const BOXCHRVERT = '|';
  const BOXCHRHORZ = '_';

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
  const DCM = DEC;  // These are now synonyms for December.  Month Number = 11, as Jan = 0.


  var MN, MN2, MN3 int //  MNEnum Month Number Vars

// YearArray subscripts are [MN,W,DOW]
  type WeekVector [7] string;
  type MonthMatrix [6] WeekVector;
  type AllMonthsArray [NumOfMonthsInYear] MonthMatrix;
  var EntireYear AllMonthsArray;
//  var MONTH Was ARRAY [JAN..DCM],[1..6],[1..7] OF STR10TYP in Modula-2


  var year,DOW,W,JAN1DOW,FEBDAYS int // was CARDINAL in Modula-2
  var DIM, WIM [NumOfMonthsInYear]int   // was ARRAY[JAN..DCM] OF CARDINAL in Modula-2
  var MONNAMSHORT [NumOfMonthsInYear]string;  // was ARRAY[JAN..DCM] OF STR10TYP in Modula-2
  var MONNAMLONG  [NumOfMonthsInYear]string   // was ARRAY[JAN..DCM] OF LNTYP in Modula-2

// ------------------------------------------------------- DAY2STR  -------------------------------------
func DAY2STR(DAY int) string {
/*
********************* DAY2STR *****************************************
DAY TO STRING CONVERSION.
THIS ROUTINE WILL CONVERT THE 2 DIGIT DAY INTO A 2 CHAR STRING.
IF THE FIRST DIGIT IS ZERO, THEN THAT CHAR WILL BE BLANK.
*/

const digits = "0123456789"
const ZERO = '0';

  bs := make([]byte,3);

  TENSDGT := DAY / 10;
  UNTSDGT := DAY % 10;
  bs[0] = BLANKCHR
  if TENSDGT == 0 {
    bs[1] = BLANKCHR;
  }else{
    bs[1] = digits[TENSDGT];
  }
  bs[2] = digits[UNTSDGT];
  return string(bs);  // not sure if this is best as a string or as a byteslice
} //END DAY2STR;

func DATEASSIGN(mn int) {
/*
--------------------------------------------------------- DATEASSIGN -------------------------------------------
DATE ASSIGNMENT FOR MONTH.
THIS ROUTINE WILL ASSIGN THE DATES FOR AN ENTIRE MONTH.  IT WILL PUT THE CHAR
REPRESENTATIONS OF THE DATE IN THE FIRST 2 BYTES.  THE EXTRA BYTES CAN BE USED 
LATER FOR SEPCIAL PRINTER CONTROL CODES.

INPUT FROM GBL VAR'S : DIM(MN), DOW
OUTPUT TO  GBL VAR'S : MonthArray(MN,,), WIM(MN)

*/

  W := 0; // W is for Week number, IE, which week of the month is this.
  for DATE := 1; DATE <= DIM[mn]; DATE++ {
    if DOW > 6 {  // DOW = 0 for Sunday.
      W++
      DOW = 0;
    } // ENDIF
    DATESTR := DAY2STR(DATE);
    EntireYear[mn] [W] [DOW] = DATESTR;
    DOW++
  } // ENDFOR;
  WIM[mn] = W;  /* Return number of weeks in this month */
} // END DATEASSIGN

/* not yet
// ----------------------------------------------------------- PRMONTH --------------------------------------
func PRMONTH(MN int ) {

  FWRSTR(OUTUN1,PAGCHR);
  FWRBL(OUTUN1,35);  (* Line displacement for month name *)
  FWRSTR(OUTUN1,MONNAMSHORT[MN]);
  FWRSTR(OUTUN1,BLANKSTR2);
  FWRSTR(OUTUN1,YEARSTR);
  FWRLN(OUTUN1);
  FWRLN(OUTUN1);
  FWRSTR(OUTUN1,DAYSNAMLONG);
  FWRLN(OUTUN1);
  FWRSTR(OUTUN1,TOPLINE);
  FWRLN(OUTUN1);
  FOR W := 1 TO WIM[MN] DO
    FWRSTR(OUTUN1,BLANKSTR2);
    FWRSTR(OUTUN1,MONTH[MN,W,1]);
    FWRBL(OUTUN1,6);
    FOR I := 2 TO 6 DO
      FWRSTR(OUTUN1,BLANKSTR2);
      FWRSTR(OUTUN1,MONTH[MN,W,I]);
      FWRBL(OUTUN1,8);
    END(*FOR*);
    FWRSTR(OUTUN1,BLANKSTR2);
    FWRSTR(OUTUN1,MONTH[MN,W,7]);
    FWRSTR(OUTUN1,CR);
    FOR J := 1 TO LNPPAG DIV WIM[MN] - 1 DO
      FWRSTR(OUTUN1,ANYLINE);
      FWRLN(OUTUN1);
    END(*FOR*);
    IF W = WIM[MN] THEN
      BotLnWithMonth := BOTMLINE;
      BotLnWithMonth[34] := ' ';
      I := 35;
      FOR J := 0 TO STRLENFNT(MONNAMSHORT[MN])-1 DO
        BotLnWithMonth[I] := MONNAMSHORT[MN,J];
        INC(I);
      END(*FOR*);
      BotLnWithMonth[I] := ' ';
      INC(I);
      FOR J := 0 TO STRLENFNT(YEARSTR)-1 DO
        BotLnWithMonth[I] := YEARSTR[J];
        INC(I);
      END(*FOR*);
      BotLnWithMonth[I] := ' ';
      FWRSTR(OUTUN1,BotLnWithMonth);
    ELSE
      FWRSTR(OUTUN1,MIDLINE);
    END(*IF*);
    FWRLN(OUTUN1);
  END(*FOR*);
  PAGCHR := FF;
} // END PRMONTH
*/


/*
------------------------------------------------------------- MAIN ---------------------------------------------
*/
func main() {
  fmt.Println("Calendar Printing Program.  ",LastCompiled);
  fmt.Println();

  if len(os.Args) <=1 {
    fmt.Println(" Usage: cal <year>");
    os.Exit(0);
}

  ExtDefault := ".out";


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
    fmt.Printf("Year is %d, which is out of range (1900-2100).  Exiting.\n",year);
    os.Exit(1)
  }


  BaseFilename := YearToken.Str;
  Cal1Filename := BaseFilename + "_cal1" + ExtDefault;
  Cal12Filename := BaseFilename + "_cal12" + ExtDefault;

  fmt.Println(" Output Files are : ",Cal1Filename,Cal12Filename);
  fmt.Println();

/*
  OutCal1,err := os.Create(Cal1Filename);
  check(err," Trying to create Cal1 output file");
  defer OutCal1.Close();

  OutCal12,e := os.Create(Cal12Filename);
  check(e," Trying to create Cal12 output file");
  defer OutCal12.Close();


  OutCal1file := bufio.NewWriter(OutCal1);
  defer OutCal1file.Flush();

  OutCal12file := bufio.NewWriter(OutCal12);
*/
  MONNAMSHORT[JAN] = "JANUARY";
  MONNAMSHORT[FEB] = "FEBRUARY";
  MONNAMSHORT[MAR] = "ARCH";
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


// Time to make the calendar

    for MN := JAN; MN <= DCM; MN++ {
      DATEASSIGN(MN);
    } // ENDFOR;

  fmt.Printf("January %d\n",year);
  fmt.Println(EntireYear[JAN]);
  fmt.Printf("Feb %d\n",year);
  fmt.Println(EntireYear[FEB]);
  fmt.Printf("December %d\n",year);
  fmt.Println(EntireYear[DEC]);

} // end main func


// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
  if e != nil {
    fmt.Errorf("%s : ",msg)
    panic(e);
  }
}




//END CAL


