package main;
/*
  REVISION HISTORY
  ----------------
  29 Nov 13 -- First version based on TestFilePickerBase, vlcshuffle and TestXMLtoken.
                This uses *.xspf as its default pattern.
  11 Jan 14 -- Discovered that if there is only 1 match to pattern, it fails.  And I will
                have new pattern option also display some potential matches.
  18 Jan 14 -- Add procedure to make sure pattern ends in .xspf.
  21 Sep 16 -- Started conversion to Go.  This is going to take a while.  After experimenting with encoding/xml, I decided
                not to use that package due to it's not ignoring garbage characters like <LF> <tab> <space> etc.
  26 Sep 16 -- Still converting.  It seems that peekXMLtoken is a misnomer.  I don't remember what I was thinking, but
                I cannot find a code section in which I need a peek function.  I always use GetXMLtoken.  I'm going to
                rewrite it so that there is no peek and next functions, just a get function (called a Getter in Go idioms).
                There is a need for PeekChar, or I could have written using GetChar and UngetChar, but I didn't because
                Modula-2 file functions had a Peek function, not Unget.
  30 Sep 16 --  First working version of the code completed.
                Changes from the Modula-2 version: command line interface for the filename, and much larger capacity.
                By using a slice, this pgm can handle more than MaxNumOfTracks, as the append function will enlarge the slice
                as needed.
   6 Oct 16 -- Added timing information to time the shuffling.  This seems to take most of the time, esp as the file sizes get bigger and bigger.
*/

import (
"os"
"bufio"
"fmt"
"runtime"
"strings"
"strconv"
"io"
"math/rand"
"time"
"path/filepath"
//
"getcommandline"
"timlibg"
)

const MaxNumOfTracks = 2048;
const blankline = "                                                                             "; // ~70 spaces
const sepline   = "-----------------------------------------------------------------------------";
const LastCompiled = "on or about Oct 1, 2016";

const (
  EMPTY = iota
  CONTENTS     // was string in the old Modula-2 code
  OPENINGHTML
  CLOSINGHTML
  OTHERERROR
      );

const   (   //  XMLcharType  which is an enumeration of states of a single character.
// I removed the EOL state, as it no longer applies.  Modula-2 would return a special EOL character, but this language follows the 
// C tradition of \r for <CR>, ASCII value of 13, and \n for <LF>, value of 10
  CTRL = iota
  OPENANGLE
  CLOSEANGLE
  SLASH
  PLAIN
        );

type TokenType struct {
  Str string;
  State int;
}

type CharType struct {
  Ch byte;
  State int;
}

// track was an array of TrackType.  Now it's a slice of pointers to TrackType, to make it easier to
// shuffle.  And so I don't need NumArray anymore which just shuffles an array of indices into
// TrackArray.
type  TrackType struct {
        location, title, creator, image, duration, extension string;
  }

var TrackSlice []*TrackType;  // Global variable.  But still needs to call make in func main.
// var haveValidTkn bool;     Don't need this anymore.

// var XMLtoken, peekXMLtoken TokenType;  Not sure that I want this here yet.

var lineDelim string;
var tabchar = '\t';
//  var indivTrackNum = 0;  Still don't think I need this.

func init() {
  rand.Seed(time.Now().UnixNano());
  TrackSlice = make([]*TrackType,0,MaxNumOfTracks);
}

/* ------------------------------------------- MakeDateStr ---------------------------------------------------* */

func MakeDateStr() (datestr string) {

const DateSepChar = "-";

  m,d,y := timlibg.TIME2MDY();
  timenow := timlibg.GetDateTime();

  MSTR := strconv.Itoa(m);
  DSTR := strconv.Itoa(d);
  YSTR := strconv.Itoa(y);
  Hr := strconv.Itoa(timenow.Hours);
  Min := strconv.Itoa(timenow.Minutes);
  Sec := strconv.Itoa(timenow.Seconds);


  datestr = "_" + MSTR + DateSepChar + DSTR + DateSepChar + YSTR + "_" + Hr + DateSepChar + Min + DateSepChar +
                 Sec + "__" + timenow.DayOfWeekStr;
  return datestr;
}  // MakeDateStr


/* -------------------------------------------- Shuffle ---------------------------------------------------- */

func Shuffle() {
/*
  Shuffle the array by passing once through the array, swapping each element with another, randomly chosen, element.
*/

  n := len(TrackSlice);

  for c := 1; c <= n; c++ {  // c is not used in the loop below.  It's just an outer loop counter.

    for i := n-1; i > 0; i-- {
/* swap element i with any element at or below that place.  Note that i is not allowed to be 0, but k
 * can be */
      k := rand.Intn(i);
      TrackSlice[i],TrackSlice[k] = TrackSlice[k],TrackSlice[i];   // Go swap idiom, to swap pointers to a track that's held in TrackSlice
    } // for inner loop
  } // for outer (ntimes) loop
}  // Shuffle;

// ---------------------------------------------------------------- PeekChar -----------------------------
// peeks at the next char without advancing fileptr.  The filepointer is advanced by ReadChar, below.
func PeekChar(f *bufio.Reader) (ch CharType, EOF bool){
  b := make([]byte,1);
  b,err := f.Peek(1);  // b is a byte slice with size of 1 byte.
  if err == io.EOF {  // basically any error is returned as EOF, because of the n==0 condition.
    return CharType{}, true;
  }else if err != nil {  // These 2 conditions are not essentially different.  They may be in the future.
    return CharType{}, true;
  }

  ch.Ch = b[0];

  if ch.Ch == '<' {
    ch.State = OPENANGLE;
  }else if ch.Ch == '>' {
    ch.State = CLOSEANGLE;
  }else if ch.Ch == '/' {
    ch.State = SLASH;
  }else if ch.Ch <= 31 {  // remember that 32 is a space.
    ch.State = CTRL
  }else{
   ch.State = PLAIN;
  }
  return ch, EOF;

/* set filepointer back one byte from its current position.  This will have Read get this byte again.
   Nevermind, by using bufio instead of os, I found a peek function.
  NextFileOffset, err = f.Seek(-1,1);
  if err != nil { fmt.Errorf(" Error from PeekChar call to Seek: %v\n",err); }
*/

}  // PeekChar

// ------------------------------------------------------------- NextChar ----------------------------
// advances filepointer.  PeekChar does not advance fileptr.  This rtn will throw the character away, as
// it assumes that PeekChar already got and processed the character.

func NextChar(f *bufio.Reader) {
  _,err := f.Discard(1);  // Discard 1 byte.  Throw away the return saying how many bytes were actually discarded
  check(err,"In DiscardChar and got err:");
} // DiscardChar


/* -------------------------------------------------------------------------- GetXMLtoken -------------------------------------------- */
func GetXMLtoken(f *bufio.Reader) (XMLtoken TokenType, EOFFLG bool) {
/*
This will use the bufio file operations as I want this as a character stream.  
The only delimiters are angle brackets.  This is the only routine where input characters are read and processed.
And I rewrote it to just exist as an XML token getter.  I don't need peeking functionality.  I guess when I
first wrote this, I thought I would need this capability.

*/

//  XMLtoken := TokenType{};  // nil literal not needed because Go automatically does this for params.

  tokenbyteslice := make([]byte,0,256);  // intermed type to make a string before returning.

  MainForLoop: for {
    ch,EOF := PeekChar(f);
    if EOF {
      return TokenType{},true;
    } // if EOFFLG then return empty TokenType and true for EOF

    switch XMLtoken.State {
    case EMPTY :
      switch ch.State {
      case PLAIN,SLASH :
        if ch.Ch != ' ' { // ignore leading blanks, but always go to NextChar.
          XMLtoken.State = CONTENTS;
          tokenbyteslice = append(tokenbyteslice,ch.Ch);  // build contents
        }
        NextChar(f);
      case OPENANGLE :
        XMLtoken.State = OPENINGHTML;
        NextChar(f);  /* discard byte, but change state to begin a tag */
      case CTRL :
        NextChar(f); /* discard these */
      case CLOSEANGLE :
        fmt.Errorf(" In peekXMLtoken and got an unexpected close angle.");
        XMLtoken.State = OTHERERROR;
        return XMLtoken,false;
      } /* case ch.state when the token state is empty */
    case CONTENTS :  // this case was STRING in the original Modula-2 code
      switch ch.State {
      case PLAIN,SLASH :
        tokenbyteslice = append(tokenbyteslice,ch.Ch);  // continue building the contents string
        NextChar(f);
      case CTRL :
        NextChar(f); /* ignore control char */
      case OPENANGLE : /* openangle char is still avail for next loop iteration */
        break MainForLoop;
      case CLOSEANGLE :
        fmt.Errorf(" In GetXMLToken.  String token got closeangle char");
      } /* case ch.state when the token state is STRING which is the value of the tag */
    case OPENINGHTML :
      switch ch.State {
      case  PLAIN,OPENANGLE :
        tokenbyteslice = append(tokenbyteslice,ch.Ch);
        NextChar(f);
      case SLASH :
        NextChar(f);
        if len(tokenbyteslice) == 0 {
          XMLtoken.State = CLOSINGHTML  // change state of this token from OPENING to CLOSING
        }else{
          tokenbyteslice = append(tokenbyteslice,ch.Ch);
        } /* if length == 0 */
      case CLOSEANGLE,CTRL :
        NextChar(f);
        break MainForLoop;
      } /* case chstate when the token state is OPENINGHTML */
    case CLOSINGHTML :
      switch ch.State {
      case PLAIN,SLASH,OPENANGLE :
        tokenbyteslice = append(tokenbyteslice,ch.Ch);
        NextChar(f);
      case CLOSEANGLE,CTRL :
        NextChar(f);
        break MainForLoop;
      } /* case chstate */
    default:
      fmt.Errorf(" In GetXMLtoken and tokenstate is in default clause of switch case.");
      XMLtoken.State = OTHERERROR;
      return XMLtoken,false;
    } /* case XMLtoken.State */
  } // indefinite for loop 

  XMLtoken.Str = string(tokenbyteslice);
  return XMLtoken, false;
} // GetXMLtoken




/* -------------------------------------------------- GetTrack -------------------------------------------- */
func GetTrack(f *bufio.Reader) (trk *TrackType) {

// This returns a pointer to TrackType now.  But I don't need to explicitly dereference this pointer in Go.

  trk = new(TrackType);
  for {
    XMLtoken, EOF := GetXMLtoken(f);
    if EOF {
      fmt.Println(" Trying to get XML record and got unexpected EOF condition.");
      break;
    }
    if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"LOCATION") {
      XMLtoken, EOF = GetXMLtoken(f);
      if EOF || (XMLtoken.State != CONTENTS) {
        fmt.Println(" Trying to get location XML tag and got unexpedted EOF condition or token is not CONTENTS.")
        break;
      }
      trk.location = XMLtoken.Str;
      _, _ = GetXMLtoken(f);  /* retrieve and discard the closinghtml for location */

    }else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"title") {
      XMLtoken, EOF = GetXMLtoken(f);
      if EOF || (XMLtoken.State != CONTENTS)  {
        fmt.Println(" Trying to get title XML tag and got unexpected EOF condition or token is not CONTENTS.");
        break;
      }
      trk.title = XMLtoken.Str;
      _,_ = GetXMLtoken(f);  /* retrieve and discard the closinghtml for title */

    }else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"creator") {
      XMLtoken, EOF = GetXMLtoken(f);
      if EOF || (XMLtoken.State != CONTENTS) {
        fmt.Println(" Trying to get creator XML tag and got unexpected EOF condition or token is not CONTENTS.")
        break;
      }
      trk.creator = XMLtoken.Str;
      _,_ = GetXMLtoken(f);  /* retrieve and discard the closinghtml for creator */

    }else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"image") {
      XMLtoken, EOF = GetXMLtoken(f);
      if EOF || (XMLtoken.State != CONTENTS) {
        fmt.Println(" Trying to get image XML record and got unexpected EOF condition or token is not CONTENTS.");
        break;
      }
      trk.image = XMLtoken.Str;
      _,_ = GetXMLtoken(f);  /* retrieve and discard the closinghtml for image */

    } else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"duration") {
      XMLtoken, EOF = GetXMLtoken(f);
      if EOF || (XMLtoken.State != CONTENTS) {
        fmt.Println(" Trying to get duration XML record and got unexpected EOF condition or token is not CONTENTS.");
        break;
      }
      trk.duration = XMLtoken.Str;
      _,EOF = GetXMLtoken(f);  /* retrieve and discard the closinghtml for duration */

    } else if (XMLtoken.State == OPENINGHTML) && strings.HasPrefix(strings.ToLower(XMLtoken.Str),"extension") {
/* this tag is more complicated than the others because it includes an application and a nested vlc:id tag */
      trk.extension = XMLtoken.Str;
/* retrieve and discard the vlc:id tag in its entirety */
      for  {
        XMLtoken, EOF = GetXMLtoken(f);
        if EOF || ((XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str,"extension")) {
          break
        }
      } // was REPEAT ... UNTIL in original Modula-2 code.

    } else if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str,"track") {
        break;
    } else if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str,"tracklist") {
        fmt.Println(" In GetTrack and came upon unexpected </tracklist>");
      break;
/* now should never get here because this tag is really part of the extension tag and it's swallowed there */
    } else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"track") {
      fmt.Println(" in GetTrack and found an unexpected opening track tag");
      break;
    }else{
/*      Have random white space here, typically either a space or a tab before an opening html tag.  Ignore it. */
    } /* if XMLtkn.state == whatever */
  } /* Outer for loop for all contents of this track */
  return trk;
} // GetTrack



/* -------------------------------------------------- PutTrack -------------------------------------------- */
func PutTrack(f *bufio.Writer, trk *TrackType, TrackNum int) {

// indivTrackNum used to be incremented here.  I'll have it incremented in the caller now.
// And the input param is now a pointer to the TrackType, and an array subscript of what was TrackArray in the Modula-2 version of the code.


// lineDelim is already set in main to be <CR><LF> for Windows and <LF> for everything else (Linux).


  _,err := f.WriteRune(tabchar);
  check(err," First WriteRune of tabchar in PutTrack and got ");
  f.WriteString("<track>");
  f.WriteString(lineDelim);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteString("<location>");
  f.WriteString(trk.location); // Remember that TrackSlice is a slice of pointers to a TrackType.
  f.WriteString("</location>");
  f.WriteString(lineDelim);

  if len(trk.title) > 0 { // I don't know if I'm required to explicitly dereference this pointer.
    f.WriteRune(tabchar);
    f.WriteRune(tabchar);
    f.WriteRune(tabchar);
    f.WriteString("<title>");
    f.WriteString(trk.title);
    f.WriteString("</title>");
    f.WriteString(lineDelim);
  } /* if have a title tag */

  if len(trk.creator) > 0 {
    f.WriteRune(tabchar);
    f.WriteRune(tabchar);
    f.WriteRune(tabchar);
    f.WriteString("<creator>");
    f.WriteString(trk.creator);
    f.WriteString("</creator>");
    f.WriteString(lineDelim);
  } /* if have a creator tag */

  if len(trk.image) > 0 {
    f.WriteRune(tabchar);
    f.WriteRune(tabchar);
    f.WriteRune(tabchar);
    f.WriteString("<image>");
    f.WriteString(trk.image);
    f.WriteString("</image>");
    f.WriteString(lineDelim);
  } /* if have an image tag */

  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteString("<duration>");
  f.WriteString(trk.duration);
  f.WriteString("</duration>");
  f.WriteString(lineDelim);

  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteString("<");
  f.WriteString(trk.extension);
  f.WriteString(">");
  f.WriteString(lineDelim);

  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteString("  <vlc:id>");
  nstr := strconv.Itoa(TrackNum);
  f.WriteString(nstr);
  f.WriteString("</vlc:id>");
  f.WriteString(lineDelim);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteString(" </extension>");
  f.WriteString(lineDelim);

  f.WriteRune(tabchar);
  f.WriteRune(tabchar);
  f.WriteString("</track>");
  _,err = f.WriteString(lineDelim);
  check(err," Last write of lineDelim in PutTrack, and got ");
  return;
} // PutTrack


/* ---------------------------------------------- ProcessXMLfile    ------------------------------------------ */

func ProcessXMLfile(inputfile *bufio.Reader, outputfile *bufio.Writer) {

  firstlineoffile,err := inputfile.ReadString('\n'); /* this is the ?xml version line, incl'g <CR><LF> chars */
  check(err,"Error when reading first line of input file.");
  _,err = outputfile.WriteString(firstlineoffile);
  check(err,"Error when writing first line of output file.");

  secondlineoffile,err := inputfile.ReadString('\n'); /* this is the playlist xmlns= line */
  check(err,"Error when reading second line of input file.");
  _,err = outputfile.WriteString(secondlineoffile);
  check(err,"Error when writing second line of output file.");


  thirdlineoffile,err := inputfile.ReadString('\n');  /* this is the title line */
  check(err,"Error when reading third line of input file.");
  _,err = outputfile.WriteString(thirdlineoffile);
  check(err,"Error when writing third line of output file.");

  for { // LOOP to ignoring white space until get the opening tracklist tag
    XMLtoken,EOF := GetXMLtoken(inputfile);
    if EOF {
      fmt.Errorf(" ProcessXMLfile and got EOF when trying to get <trackList>.  Ending.\n");
      return;
    } // if EOF

    if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"tracklist") {
      break;
    } /* if have tracklist */
  }/* loop until get opening tracklist tag */

  for { /* ignoring white space until get the opening track tag.  I'm not sure I still need this as a loop.  But it works. */
    XMLtoken,EOF := GetXMLtoken(inputfile);
    if EOF {
      fmt.Errorf(" Trying to get opening track tag and got EOF condition.  Ending.\n");
      return;
    }  // if EOF
    if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str,"track") {
      break;
    } /* if have track */
  } /* loop ignoring whitespace until get the opening track tag */

  TrackSlice = make([]*TrackType,0,MaxNumOfTracks)

  for { /* to read the continuous stream of track tokens */
    trackptr := GetTrack(inputfile);
    TrackSlice = append(TrackSlice,trackptr);  // I'm using a separate assignment to hold the pointer, so I can more easily debug the code, if needed.

/*
 This next token will either be a closing tracklist tag or an opening track tag.  If it is not
 a closing tracklist tag to end the loop, then we just swallowed the next opening track tag which
 is perfect for the GetTrack rtn anyway.
*/

    XMLtoken,EOF := GetXMLtoken(inputfile);  // this token should be <track> and then rtn loops again.
    if EOF {
      fmt.Errorf(" Trying to get another track tag and got EOF condition.  Ending.\n");
      return;
    }

    if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str,"tracklist") { // unexpected condition
        break;
    } /* if have closing tracklist */
  } /* loop to read in more tracks */


  NumOfTracks := len(TrackSlice);
  fmt.Println("Last track number read is ",NumOfTracks);

  t0 := time.Now();
/*
  need to shuffle here
*/

  Time := timlibg.GetDateTime();
  shuffling := Time.Month + Time.Day + Time.Hours + Time.Minutes + Time.Year + Time.Seconds;
  for k := 0; k < shuffling; k++ {
    Shuffle();
  }
/* Finished shuffling.    */

  timeToShuffle := time.Since(t0);  // timeToShuffle is a Duration type, which is an int64 but has methods.
  timeToShuffleString := timeToShuffle.String();
  fmt.Println(" It took ",timeToShuffleString," to shuffle this file.");
  fmt.Println();
  
  
/* Write the output file. */
  _,err = outputfile.WriteRune(tabchar);
  check(err," Starting to write the shuffled tracklist to the output file and got error: ");
  outputfile.WriteString("<trackList>");
  outputfile.WriteString(lineDelim);

  for c := 0; c < len(TrackSlice); c++ {
        PutTrack(outputfile,TrackSlice[c],c);
  }

  outputfile.WriteRune(tabchar);
  outputfile.WriteString("</trackList>");
  outputfile.WriteString(lineDelim);

  for { /* to read and write the rest of the lines */
    line,err := inputfile.ReadString('\n');
    if err == io.EOF { break }
    check(err," Reading final lines of the inputfile and got this error: ");
    _,err = outputfile.WriteString(line);
  } /* final read and write loop */


} // ProcessXMLfile


// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
  if e != nil {
    fmt.Errorf("%s : ",msg)
    panic(e);
  }
}



/* ---------------------------- MAIN -------------------------------- */
func main() {
  fmt.Println(" Shuffling program for the tracks in a vlc file.");

  if len(os.Args) <=1 {
    fmt.Println(" Usage: vlc <trackfile.xspf>");
    os.Exit(0);
}

  commandline := getcommandline.GetCommandLineString();
  cleancommandline := filepath.Clean(commandline);
  infile,err := os.Open(cleancommandline);
  if err != nil {
    fmt.Println(" Cannot open input file.  Does it exist?  Error is",err);
    os.Exit(1);
  }

  defer infile.Close();
  inputfile := bufio.NewReader(infile);



/*   Build outfilename */
  BaseFilename := filepath.Base(cleancommandline);
  ExtFilename := filepath.Ext(cleancommandline);
  lastIndex := strings.LastIndex(BaseFilename,".");
  base := BaseFilename[:lastIndex];  // base is the name without extension

  TodaysDateString := MakeDateStr();

  outfilename := base + TodaysDateString + ExtFilename;
  outfile,err := os.Create(outfilename);
  if err != nil {
    fmt.Println(" Cannot open outfilename ",outfilename,"  with error ",err);
    os.Exit(1);
  }
  defer outfile.Close();

  outputfile := bufio.NewWriter(outfile);
  defer outputfile.Flush();

//  indivTrackNum = 1;

  fmt.Println();
  fmt.Println(" GOOS =",runtime.GOOS,".  ARCH=",runtime.GOARCH);
  fmt.Println();
  if runtime.GOOS == "windows" {
    lineDelim = "\r\n"
  }else{
    lineDelim = "\n"
  }


  ProcessXMLfile(inputfile,outputfile);

} //  vlc main
