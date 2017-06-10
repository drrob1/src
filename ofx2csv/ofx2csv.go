package main

const lastModified = "10 Jun 17"

import (
  "os"
  "fmt"
  "bufio"
  "io/ioutil"
  "getcommandline"
  "timlibg"

)
/*
MODULE qfx2xls;
  REVISION HISTORY
  ----------------
  13 Mar 04 -- It does not seem to be always creating the output file.
               It now copies the d/l file instead of renaming it.  And
               it echoes its output to the terminal.
  14 Mar 04 -- Made it a Text Window module
  21 Mar 04 -- Changed exit keys to remove <cr> and add <space>, <bs>, <tab>
  15 Apr 04 -- Decided to include <cr> again, as I fixed the pblm in Excel macros.
  31 Jul 04 -- Needed to change logic because Citibank now d/l nulls in fields that I need.
   8 Aug 04 -- Copied to process MMA file as well.  And fixed a minor output bug.
   8 Nov 04 -- Citibank changed their file format again.  Don't need ExtractLastNum and the
                description is now 2 fields instead of 1.
  17 Jun 06 -- Opened a new citi acnt they call eSavings.  Need to include this into database.
                And changed initial value of chknum to zero, and any key will exit now.
  18 Jun 06 -- Now uses command line for file names.
  19 Jun 06 -- Fixed bug of always writing the same acnt name and made it output filename.
  27 Jan 07 -- Noticed that the fileformat changed slightly as of Oct or so.  I have to remove
                squotes from acnt#.  And added a menu option to exit.
  29 Jan 07 -- While at ISET 2007, I decided to change the method of removing the squote so that
                all squotes are removed, even if Citibank gets cute and puts more in.
   2 Oct 07 -- Now has ability to use .qif files, and needed a module name change for this.
                Also used menu pick instead of cmd line params.
  21 Feb 08 -- Had to make output file .txt so that Access on P5 could import the file.  Don't know y.
                And I copied the .txt file to .out file so I don't have to change anything on P4.
  24 Mar 08 -- HSBC uses short date format and squote delim for 2 dgt year.
                 And I changed output file format to be more straightforward, reordering fields.
   9 Feb 09 -- Now does .qfx files, hence module name change.  And will use <tab> as output delim, just because.
                And since it really is meant for Excel to import the text file, module name change to xls.
   3 Mar 11 -- Noticed but in GetQfxToken in that read calls should all be to the param f, not the
                global infile.  I will fix this next time I have to recompile. 

   7 Jun 17 -- Converting to go.  I posted on go-nuts, and was told that the .qfx format is not xml, but ofx,
                which means open financial exchange (for/of information).  New name is ofx2cvs.go
		I think I will first process the file using something like toascii.
*/

  const ( // intended for ofxCharType
          eol = iota  // so eol = 0, and so on.  And the zero val needs to be DELIM.
          openangle
          closeangle
          slash
          plain
  )

  const ( // intended for ofxTokenType
	  empty = iota
	  strng
	  openinghtml
	  closinghtml
	  othererror
  )

  type ofxTokenType struct {
    Str string  // name or contents, depending on the State value
    State int
  }

  type ofxCharType struct {
    Ch byte
    State int
  }

var err error

const KB = 1024
const MB = KB * KB
const ofxext = ".ofx"
const qfxext = ".qfx"

type citiheadertype struct {
	DTSERVER string
	LANGUAGE string
	ORG      string
	FID      string
	CURDEF   string
	BANKID   string
	ACCTID   string
	ACCTTYPE string
	DTSTART  string
	DTEND    string
}

type citiTransactionType struct {
	TRNTYPE  string
	DTPOSTED string
	TRNAMT   int
	FITID    string
	CHECKNUM int
	NAME     string
	MEMO     string
	Juldate int
}

type citifootertype struct {
  BalAmt string
  DTasof string
}

var bsidx int // byte slice index for getting one ASCII character at a time.
var Transactions []citiTransactionType

func main () {

  var GblAcctIDStr,ledgerBalAmtStr,availBalAmtStr,BalAmtDateAsOfStr,commentStr,acntidStr string
  var ofxToken ofxTokenType
  var ofxChar ofxCharType
  var ofxDataElement ofxDataType
  var juldate1,juldate2,juldate3 uint32

  var e error
  var citiheader citixmlheadertype
  var cititrans cititransxmltype
  var filebyteslice []byte

  if len(os.Args) <= 1 {
    fmt.Println(" Usage: xmltest <hashFileName.ext> where .ext = [.qfx|.xml]")
    os.Exit(1)
  }

  inbuf := getcommandline.GetCommandLineString()
  BaseFilename := filepath.Clean(inbuf)
  InFilename := ""
  InFileExists := false

  if strings.Contains(BaseFilename, ".") { // there is an extension here
    InFilename = BaseFilename
    _, err := os.Stat(InFilename)
    if err == nil {
      InFileExists = true
    }
  } else {
    InFilename = BaseFilename + qfxext
    _, err := os.Stat(InFilename)
    if err == nil {
      InFileExists = true
    } else {
      InFilename = BaseFilename + ofxext
      _, err := os.Stat(InFilename)
      if err == nil {
        InFileExists = true
      }
    }
  }

  if !InFileExists {
		fmt.Println(" File ", BaseFilename, BaseFilename+qfxext, BaseFilename+xmlext, " or ", InFilename, " do not exist.  Exiting.")
		os.Exit(1)
	}

	toascii := func() {
		cmd := exec.Command("cmd","/c","toascii")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	if runtime.GOOS == "windows" {
      err = toascii(InFilename)
      if err != nil {
        fmt.Println(" Error from toascii ",err)
        os.Exit(1)
	  }
    }

	filebyteslice = make([]byte, 0, MB) // 1 MB as initial capacity.
	filebyteslice, e = ioutil.ReadFile(InFilename)
	if e != nil {
		fmt.Println(" Error from ReadFile is ", e)
		os.Exit(1)
	}

// This code started as qfx2xls, but I really want more like CitiFilterQIF.  So I have to merge
// in that code also.
// And I need to use toascii in some way or another, either an exec function, or copying the
// code here.  toascii deletes non UTF-8 code points, utf8toascii does not do this.

  Transactions = make([]citiTransactionType,0,100)

  header,footer :=  ProcessOFXFile(filebyteslice)

  // now I have a header, footer, and a slice of all the individual transactions.  At this
  // point, I'll just display them, and pause in between.

} // end main of this package

//---------------------------------------------------------------------------------------------------
func DateFieldReformat(datein string) (string int) {
//                                                                    01234567    01234567
//  This procedure changes the date as it is input in a qfx file from yyyymmdd -> mm/dd/yy.
// I have to look into if I want a 2 or 4 digit year 

  var dateout string
  var datebytearray [8]byte

  datebyteslice[0] = datein[4];
  datebyteslice[1] = datein[5];
  datebyteslice[2] = '/';
  datebyteslice[3] = datein[6];
  datebyteslice[4] = datein[7];
  datebyteslice[5] = '/';
  datebyteslice[6] = datein[2];
  datebyteslice[7] = datein[3];
  dateout = string(datebytearray)
  m := strconv.Atoi(datein[4:5])
  d := strconv.Atoi(datein[6:7])
  y := strconv.Atoi(datein[2:3])
  juldate := timlibg.Julian(m,d,y)
  return dateout,juldate

} // END DateFieldReformat;

//--------------------------------------------------------------------------------------------------
func GetOfxToken(bs []byte) ofxTokenType {
// -------------------------------------------------- GetQfxToken ----------------------------------
// Delimiters are fixed at angle brackets and EOL.


  var token ofxTokenType
  var char ofxCharType

  for { // main processing loop

    if bsidx >= len(bs) { // finished processing all bytes of the input filebyteslice
      return nil
    }

    // GetChar
    char.Ch = bs[bsidx]
    bsidx++

    // assign charstate
    switch char.Ch {
      case '\n','\r','\t' : char.State = eol
      case '<' : char.State = openangle;
      case '>' : char.State = closeangle;
      case '/' : char.State = slash;
      default :  char.State = plain;
    } // END switch case on ch 


    switch token.State {
    case empty :  // of token.State
      switch char.State {
      case plain,slash :
        token.State = strng;
        token.Str = string(char.Ch)
      case openangle :
        token.State = openinghtml;
      case eol :
        // do nothing
      case closeangle :
        fmt.Println(" In GetOfxToken.  Empty token got closeangle char")
      } // END case chstate is empty

    case strng : // of token.State
      switch char.State {
      case plain,slash :
        token.Str = token.Str + string(char.Ch)
      case eol :
        break
      case openangle : // openangle char is still avail for next loop iteration
        bsidx--
	  break
        case closeangle :
        fmt.Println(" In GetOfxToken.  String token got closeangle ch")
	  } // END case chtkn.State in ofxtkn.Str of strng
    case openinghtml : // of token.State
      switch  char.State {
      case plain,openangle :
        token.Str = token.Str + string(char.Ch)
      case slash :
        if len(token.Str) = 0 {
          token.State = closinghtml
	    } else {
          token.Str = token.Str + string(char.Ch)
	    } // END;
      case closeangle,eol :
        break
	  } // END case chtkn.State in openinghtml
    case closinghtml : // of token.State
      switch  char.State {
      case plain,slash,openangle :
        token.Str = token.Str + string(char.Ch)
      case closeangle,eol :
        break
      } //      END; (* case chstate in closinghtml *)
    default: // ofxtkn.State is othererror
        fmt.Println(" In GetQfxToken and tokenstate is othererror.")
    } // END case ofxtknstate
  } // END ofxtkn.State processing loop 

  return token
} // END GetOfxToken;




// ---------------------------------------------------- GetTransactionData --------------------------
func GetTransactionData(bs []byte) citiTransactionType {
// Returns nil as a sign of either normal or abnormal end.

var OFXtoken OfxTokenType
var transaction citiTransactionType

  for { // processing loop
    OFXtoken = GetOfxToken(bs)
    if OFXtoken == nil {
      fmt.Println(" Trying to get qfx record and got unexpected EOF condition.")
      return nil
    }

    if false {
   // do nothing but it allows the rest of the conditions to be in the else if form

    } else if OFXtoken.State == openinghtml && OFXtoken.Str == "TRNTYPE" {
      OFXtoken = GetOfxToken(bs)
      if OFXtoken == nil || (OFXtoken.State != strng) {
        fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
        return nil
      } // if EOF or token state not a string
      transaction.TRNTYPE = OFXtoken.Str

    } else if (OFXoken.State == openinghtml) && (OFXtoken.Str == "DTPOSTED") {
      OFXtoken = GetOfxToken(bs)    // Now need the string data of this token
      if OFXtoken == nil || (OFXtoken.State != strng) {
        fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
        return nil
      } // if EOF or token state not a string
      transaction.DTPOSTED,transaction.Juldate = DateFieldReformat(OFXtoken.Str)

    } else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "TRNAMT") {
      OFXtoken = GetOfxToken(bs)
      if OFXtoken == nil || (OFXtoken.State != strng) {
        fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
        return nil
      } // if EOF or token state not a string
      transaction.TRNAMT = strconv.Atoi(OFXtoken.Str)

    } else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "FITID") {
      OFXtoken = GetOfxToken(bs)
      if OFXtoken == nil || (OFXtoken.State != strng) {
        fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
        return nil
      } // if EOF or token state not a string
      transaction.FITID = OFXtoken.Str

    } else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "NAME") {
      OFXtoken = GetOfxToken(bs)
      if OFXtoken == nil || (OFXtoken.State != strng) {
        fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
        return nil
      } // if EOF or token state not a string
      transaction.NAME = OFXtoken.Str

    } else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "MEMO") {
      OFXtoken = GetOfxToken(bs)
      if OFXtoken == nil || (OFXtoken.State != strng) {
        fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
        return nil
      } // if EOF or token state not a string
      transaction.MEMO = OFXtoken.Str

    } else if (OFXtoken.State == closinghtml) && (OFXtoken.Str == "STMTTRN") {
      return transaction

    } else if (qfxtokenstate == closinghtml) && (OFXtoken.Str == "BANKTRANLIST") {
      return nil
    } // END if OFXoken.State condition 
  } // END processing loop for record contents
  return transaction
} // END GetTransactionData


//--------------------------------------------------------------------------------------------
func ProcessOFXFile(bs []byte) (citiheadertype,citifootertype) {

var header citiheadertype
var token ofxTokenType
var transaction citiTransactionType
var footer citifootertype


  for {  // loop to read the header
    token = GetOFXToken(bs)
    if token == nil {
      fmt.Println(" Trying to get header info and got EOF condition.")
      return nil
    }

    if (token.State == openinghtml) && (token.Str == "ORG") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.ORG := token.Str;

    } else if (token.State == openinghtml) && (token.Str == "ACCTID") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.ACCIT = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "DTSERVER") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.DTSERVER = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "LANGUAGE") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.LANGUAGE = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "FID") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.FID = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "CURDEF") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.CURDEF = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "BANKID") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.BANKID = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "ACCTTYPE") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.ACCTTYPE = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "DTSTART") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.DTSTART = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "DTEND") {
      token = GetOfxToken(bs)
      if token == nil || (token.State != string) {
        fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
        return nil,nil
      }
      header.DTEND = token.Str
      
    } else if (token.State == openinghtml) && (token.Str == "STMTTRN") {
		break
      
    } // END if token.State AND token.Str
  } // END loop for header info






  for { // LOOP to read multiple transactions
    transaction = GetTransactionData(filebyteslice)

    if transaction == nil { // either at EOF or there was an error from GetTransactionData
        break
    }

    Transactions = append(Transactions,transaction)

  } // END LOOP for multiple transactions

//  Get Footer containing ledgerbal, balamt, dtasof.  Stop when come TO </OFX>

  for { // loop to get the footer.   exit out of this loop at EOF or came to </OFX>
    token = GetOfxToken(bs)
    if token == nil {
      fmt.Println(" Trying to get footer info and got EOF condition.")
      return header,footer
    }

    if false {
	// do nothing
	} else if token.State == openinghtml && token.Str == "BALAMT" {
       token = GetOfxToken(bs)
       if token == nil {
         fmt.Println(" Trying to get footer info and got an error.")
		 return header,nil
       }
	   footer.BalAmt = token.Str 

    } else if token.State == openinghtml && token.Str == "DTASOF" {
       token = GetOfxToken(bs)
       if token == nil {
         fmt.Println(" Trying to get footer info and got an error.")
		 return header,nil
       }
	   footer.DTasof = token.Str 

    } else if token.State == closinghtml && token.Str == "LEDGERBAL" {
		break

    } else if token.State == closinghtml && token.Str == "BANKMSGSRSV1" {
		break

    } else if token.State == closinghtml && token.Str == "OFX" {
		break

	} // END if token.State
  } // END loop for footer info

  return header,footer
} // END ProcessOFXFile



(-++++-----------------------------------------------------------------)
PROCEDURE WndProcTW(tw : TextWindow; msg : TWMessageRec) : ResponseType;
(----------------------------------------------------------------------)
VAR
    clr         : Colors;
    x,y         : COORDINATE;
    i,int       : INTEGER;
    cmdline     : ARRAY [0..255] OF CHAR;
    cmdbuf,tkn  : BUFTYP;
    tknstate    : FSATYP;
    retcod,c5   : CARDINAL;
    filter,s    : STRTYP;

BEGIN
    CASE msg.msg OF
    TWM_CLOSE:  (* Turns out that this winmsg is being executed twice before the pgm closes.  I have no idea why *)
(*        BasicDialogs.MessageBox(outfilename,MsgInfo); *)
        Strings.Append('.xls',outfilename);
(*
        BasicDialogs.MessageBox(outfilename,MsgInfo);
        BasicDialogs.MessageBox(OUTFNAM.CHARS,MsgInfo);
*)
        FileFunc.CopyFile(OUTFNAM.CHARS,outfilename);
        IF msg.closeMode = CM_DICTATE THEN
            WinShell.TerminateDispatchMessages(0);
        END;
        RETURN OkayToClose;
    | TWM_CREATE:
        FUNC SetWindowIcon(tw, CitiIcon);

        xCaret := 0;
        yCaret := 0;
        inputline := '';
        juldate1 := 0;
        juldate2 := 0;
        juldate3 := 0;
        chknum := 0;

        INFNAM.CHARS := '';
        OUTFNAM.CHARS := '';


(*
  PROCEDURE BasicDialogs.PromptOpenFile(VAR INOUT name : ARRAY OF CHAR;
                                            filters : ARRAY OF CHAR;
                                            VAR INOUT defFilter : CARDINAL;
                                            defDir : ARRAY OF CHAR;
                                            defExt : ARRAY OF CHAR;
                                            title : ARRAY OF CHAR;
                                            createable : BOOLEAN) : BOOLEAN;
 Opens an operating system common dialog for opening  a file
   filters specifies a list of file extension filters that are
   separated by semicolons.
   The format for filters is as follows.
   defDir = the default directory to start the dialog in
   an empty string "" means use the current directory.
   defExt = the default file extension to use if the user does not
   provide an extension. "" means no default extension.
   the extension should *not* have a leading '.' character.
   title = the caption text of the dialog. title can be empty "".
   in this case the default operating system title is used.
   If createable = TRUE then the file need not already exist, otherwise
   the file must exist for the dialog to return successful.
   RETURNs TRUE is successful and name will contain the file specification
   for the file the user has given.
*)
        c5 := 1;
        DlgShell.ConvertToNulls(MenuSep,filter);
        bool := BasicDialogs.PromptOpenFile(infilename,filter,c5,'','','Open transaction text file',FALSE);
(*        BasicDialogs.MessageBox(infilename,MsgInfo); *)
        IF NOT bool THEN
          WriteString(tw,'Could not find file.  Does it exist?',a);
          HALT;
        END;

        IF NOT FileFunc.FileExists(infilename) THEN
          MiscM2.Error(' Could not find input file.  Does it exist?');
          HALT;
        END(*if*);
        OpenFile(infile,infilename,ReadOnlyDenyWrite);
        IF infile.status > 0 THEN
          WriteString(tw,' Error in opening/creating file ',a);
          WriteString(tw,inputline,a);
          WriteString(tw,'--',a);
          CASE TranslateFileError(infile) OF
            FileErrFileNotFound : WriteString(tw,'File not found.',a);
          | FileErrDiskFull : WriteString(tw,'Disk Full',a);
          ELSE
            WriteString(tw,'Nonspecific error occured.',a);
          END(*CASE*);
          WriteLn(tw);
          WriteString(tw,' Program Terminated.',a);
          WriteLn(tw);
          HALT;
        END(*IF infile.status*);
        SetFileBuffer(infile,InBuf);

        C := LENGTH(infilename);
        DEC(C);
        buf[0] := CAP(infilename[C-2]);
        buf[1] := CAP(infilename[C-1]);
        buf[2] := CAP(infilename[C]);
        buf[3] := 0C;

        IF STRCMPFNT(buf,'QFX') = 0 THEN
                csvqifqfxState := qfx;
        ELSIF STRCMPFNT(buf,'QIF') = 0 THEN
          csvqifqfxState := qif;
        ELSE
          csvqifqfxState := csv
        END;

    | TWM_SIZE:
        GetClientSize(tw,cxScreen,cyScreen);
        cxClient := msg.width;
        cyClient := msg.height;
        SnapWindowToFont(tw,TRUE);
        SetDisplayMode(tw,DisplayNormal);
        SetScrollRangeAllowed(tw,WA_VSCROLL,60);
        SetScrollBarPos(tw,WA_VSCROLL,0);
        SetScrollRangeAllowed(tw,WA_HSCROLL,100);
        SetScrollBarPos(tw,WA_HSCROLL,0);
        SetCaretType(tw,CtHalfBlock);
        MoveCaretTo(tw,xCaret,yCaret);
        MakeCaretVisible(tw);
        CaretOn(tw);
        SetWindowEnable(tw,TRUE);
        SetForegroundWindow(tw);

    | TWM_GAINFOCUS, TWM_ACTIVATEAPP :
        MoveCaretTo(tw,xCaret, yCaret);
        MakeCaretVisible(tw);
    | TWM_PAINT:
        CASE csvqifqfxState OF
          qfx: ProcessQFXFile(tw);
        | qif: MiscM2.Error(' This pgm will only process qfx files.');
        | csv: MiscM2.Error(' This pgm will only process qfx files.');
        END (*case*);

        RemoveFileBuffer(infile);
        CloseFile(infile);
        FCLOSE(OUTUN1);
        WriteLn(tw);
        WriteString(tw,OUTFNAM.CHARS,a);
        WriteString(tw,' file now closed.',a);
        WriteLn(tw);
        EraseToEOL(tw,a);
        WriteLn(tw);
        WriteLn(tw);
        INC(c32);
        FUNC FormatString(' Number of Paint msgs is: %c.',buf,c32);
        WriteString(tw,buf,a);
        WriteLn(tw);
        WriteStringAt(tw,0,cyClient-1,LastMod,a);

    | TWM_MENU:
(*
  a menu item has been selected menuId = the menu resource id number for the menu item
  TWM_MENU:
       msg.menuId      : INTEGER;
       msg.accel       : BOOLEAN;
*)
         CASE msg.menuId OF
         20  : (* exit *)
              CloseWindow(tw,CM_REQUEST);
      ELSE (* do nothing but not an error *)
      END; (* case menuId *)


    |
    TWM_KEY:
     FOR i := 0  TO INT(msg.k_count-1) DO
      IF (msg.k_special = KEY_NOTSPECIAL) THEN
        CASE msg.k_ch OF
          CHR(8) :                                     (* backspace       *)
          FUNC CloseWindow(tw, CM_REQUEST);
        | CHR(9) :                                     (* tab             *)
          FUNC CloseWindow(tw, CM_REQUEST);

        | CHR(10):                                     (* line feed       *)

        | CHR(13):                                     (* carriage RETURN *)
          FUNC CloseWindow(tw, CM_REQUEST);

        | CHR(27):                                     (* escape *)
          FUNC CloseWindow(tw, CM_REQUEST);
        | CHR(32):                                     (* space *)
          FUNC CloseWindow(tw, CM_REQUEST);
        | 'A','a': (* About *)
             BasicDialogs.MessageTitle := 'About';
             Strings.Assign('Last Modified and Compiled ',s);
             Strings.Append(LastMod,s);
             BasicDialogs.MessageBox(s, BasicDialogs.MsgInfo);
        ELSE (* CASE ch *)
          FUNC CloseWindow(tw, CM_REQUEST);
        END (* case ch *);
      ELSIF msg.k_special = KEY_PAGEUP THEN
      ELSIF msg.k_special = KEY_PAGEUP THEN
      ELSIF msg.k_special = KEY_PAGEDOWN THEN
      ELSIF msg.k_special = KEY_HOME THEN
      ELSIF msg.k_special = KEY_END THEN
      ELSIF msg.k_special = KEY_RIGHTARROW THEN
      ELSIF msg.k_special = KEY_LEFTARROW THEN
      ELSIF msg.k_special = KEY_UPARROW THEN
      ELSIF msg.k_special = KEY_DOWNARROW THEN
      ELSIF msg.k_special = KEY_INSERT THEN
      ELSIF msg.k_special = KEY_DELETE THEN

      ELSE (* msg.k_special *)
      END (*if*);
     END(* for *);
    ELSE (* case msg.msg *)
    END (* case msg.msg *);

    RETURN DEFAULT_HANDLE;
END WndProcTW;

PROCEDURE Start(param : ADDRESS);
BEGIN
    UNREFERENCED_PARAMETER(param);
    Win := CreateWindow(NIL, (* parent : WinShell.Window *)
                        "qfx To text xls Converter", (* name : ARRAY OF CHAR *)
                        "#100",        (* menu : ARRAY OF CHAR *)
                        "CitiIcon",        (* icon : ARRAY OF CHAR *)
                        -1,-1, (* x,y= the initial screen coordinates for the window to be displayed *)
                        110,20, (* xSize, ySize : COORDINATE *)
                        250,100, (* xBuffer, yBuffer : COORDINATE *)
                        FALSE,  (* gutter : BOOLEAN *)
                        DefaultFontInfo, (* font : FontInfo *)
                        ComposeAttribute(Black, White, NormalFont), (* background : ScreenAttribute *)
                        ToplevelWindow,  (* windowType : WindowTypes *)
                        WndProcTW,
                        NormalWindow + AddScrollBars,    (* attribs : WinAttrSet *)
                        NIL); (* createParam : ADDRESS *)
    IF Win = NIL THEN
        WinShell.TerminateDispatchMessages(127);
    END;
    SetAutoScroll(Win,TRUE);
END Start;

(********************************** Main body ************************************)
BEGIN
  LastModLen := LENGTH(LastMod);
  a := ComposeAttribute(Black, White, NormalFont);
  c32 := 0;
  FUNC WinShell.DispatchMessages(Start, NIL);

END qfx2xls.
