package main

/*
  REVISION HISTORY
  ----------------
  29 Nov 13 -- First version based on TestFilePickerBase, vlcshuffle and TestXMLtoken.
                This uses *.xspf as its default pattern.
  11 Jan 14 -- Discovered that if there is only 1 match to pattern, it fails.  And I will
                have new pattern option also display some potential matches.
  18 Jan 14 -- Add procedure to make sure pattern ends in .xspf.
  21 Sep 16 -- Started conversion to Go.  This is going to take a while.  After experimenting with encoding/xml, I decided
                not to use that package due to it's not ignoring garbage characters like <LF> <tab> <space> etc
   5 Apr 25 -- Converted import to modules so it stops causing errors.
*/

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	//
	"src/getcommandline"
	"src/timlibg"
)

const MaxNumOfTracks = 512
const blankline = "                                                                             " // ~70 spaces
const sepline = "-----------------------------------------------------------------------------"
const LastCompiled = "on or about Oct 1, 2016"

// XMLtokenType which is an enumeration of states of an XMLtoken.
const (
	EMPTY    = iota
	CONTENTS // was string in the old Modula-2 code
	OPENINGHTML
	CLOSINGHTML
	OTHERERROR
)

const ( //  XMLcharType  which is an enumeration of states of a single character.
	// I removed the EOL state, as it no longer applies.  Modula-2 would return a special EOL character, but this language follows the
	// C tradition of \r for <CR>, ASCII value of 13, and \n for <LF>, value of 10
	CTRL = iota
	OPENANGLE
	CLOSEANGLE
	SLASH
	PLAIN
)

type TokenType struct {
	Str   string
	State int
}

type CharType struct {
	Ch    byte
	State int
}

// track was an array of TrackType.  Now it's a slice of pointers to TrackType, to make it easier to
// shuffle.  And so I don't need NumArray anymore which just shuffles an array of indices into
// TrackArray.
type TrackType struct {
	location, title, creator, image, duration, extension string
}

var TrackSlice []*TrackType // Global variable.  But still needs to call make in func main.
var haveValidTkn bool
var XMLtoken, peekXMLtoken TokenType
var CurrentFileOffset, PreviousFileOffset, NextFileOffset int64

/*************************************************************************************************/

func MakeDateStr() (datestr string) {

	const DateSepChar = "-"

	m, d, y := timlibg.TIME2MDY()
	timenow := timlibg.GetDateTime()

	MSTR := strconv.Itoa(m)
	DSTR := strconv.Itoa(d)
	YSTR := strconv.Itoa(y)
	Hr := strconv.Itoa(timenow.Hours)
	Min := strconv.Itoa(timenow.Minutes)
	Sec := strconv.Itoa(timenow.Seconds)

	datestr = "_" + MSTR + DateSepChar + DSTR + DateSepChar + YSTR + "_" + Hr + DateSepChar + Min + DateSepChar +
		Sec + "__" + timenow.DayOfWeekStr
	return datestr
} // MakeDateStr

// ---------------------------------------------------------------- PeekChar -----------------------------
// peeks at the next char without advancing fileptr.  The filepointer is advanced by ReadChar, below.
func PeekChar(f *bufio.Reader) (ch CharType, EOF bool) {
	b := make([]byte, 1)
	b, err := f.Peek(1) // b is a byte slice with size of 1 byte.
	if err == io.EOF {  // basically any error is returned as EOF, because of the n==0 condition.
		return CharType{}, true
	} else if err != nil { // These 2 conditions are not essentially different.  They may be in the future.
		check(err, " Peeking a char and err is:")
	}

	ch.Ch = b[0]

	if ch.Ch == '<' {
		ch.State = OPENANGLE
	} else if ch.Ch == '>' {
		ch.State = CLOSEANGLE
	} else if ch.Ch == '/' {
		ch.State = SLASH
	} else if ch.Ch <= 31 { // remember that 32 is a space.
		ch.State = CTRL
	} else {
		ch.State = PLAIN
	}
	return ch, EOF
} // PeekChar

// ------------------------------------------------------------- NextChar ----------------------------
// advances filepointer.  PeekChar does not advance fileptr.  This rtn will throw the character away, as
// it assumes that PeekChar already got and processed the character.

func NextChar(f *bufio.Reader) {
	_, err := f.Discard(1) // Discard 1 byte.  Throw away the return saying how many bytes were actually discarded
	check(err, "In DiscardChar and got err:")
} // DiscardChar

// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Errorf("%s : ", msg)
		panic(e)
	}
}

/* -------------------------------------------------------------------------- PeekXMLtoken -------------------------------------------- */
func GetXMLtoken(f *bufio.Reader) (XMLtoken TokenType, EOFFLG bool) {
	/*
	   This will use the bufio file operations as I want this as a character stream.
	   The only delimiters are angle brackets.  This is the only routine where input characters are read and processed.
	   And I rewrote it to just exist as an XML token getter.  I don't need peeking functionality.  I guess when I
	   first wrote this, I thought I would need this capability.

	*/

	//  XMLtoken := TokenType{};  // nil literal not needed because Go automatically does this for params.

	tokenbyteslice := make([]byte, 0, 256) // intermed type to make a string before returning.

MainForLoop:
	for {
		ch, EOF := PeekChar(f)
		if EOF {
			return TokenType{}, true
		} // if EOFFLG then return empty TokenType and true for EOF

		switch XMLtoken.State {
		case EMPTY:
			switch ch.State {
			case PLAIN, SLASH:
				if ch.Ch != ' ' { // ignore leading blanks, but always go to NextChar.
					XMLtoken.State = CONTENTS
					tokenbyteslice = append(tokenbyteslice, ch.Ch) // build contents
				}
				NextChar(f)
			case OPENANGLE:
				XMLtoken.State = OPENINGHTML
				NextChar(f) /* discard byte, but change state to begin a tag */
			case CTRL:
				NextChar(f) /* discard these */
			case CLOSEANGLE:
				fmt.Errorf(" In peekXMLtoken and got an unexpected close angle.")
				XMLtoken.State = OTHERERROR
				return XMLtoken, false
			} /* case ch.state when the token state is empty */
		case CONTENTS: // this case was STRING in the original Modula-2 code
			switch ch.State {
			case PLAIN, SLASH:
				tokenbyteslice = append(tokenbyteslice, ch.Ch) // continue building the contents string
				NextChar(f)
			case CTRL:
				NextChar(f) /* ignore control char */
			case OPENANGLE: /* openangle char is still avail for next loop iteration */
				break MainForLoop
			case CLOSEANGLE:
				fmt.Errorf(" In GetXMLToken.  String token got closeangle char")
			} /* case ch.state when the token state is STRING which is the value of the tag */
		case OPENINGHTML:
			switch ch.State {
			case PLAIN, OPENANGLE:
				tokenbyteslice = append(tokenbyteslice, ch.Ch)
				NextChar(f)
			case SLASH:
				NextChar(f)
				if len(tokenbyteslice) == 0 {
					XMLtoken.State = CLOSINGHTML // change state of this token from OPENING to CLOSING
				} else {
					tokenbyteslice = append(tokenbyteslice, ch.Ch)
				} /* if length == 0 */
			case CLOSEANGLE, CTRL:
				NextChar(f)
				break MainForLoop
			} /* case chstate when the token state is OPENINGHTML */
		case CLOSINGHTML:
			switch ch.State {
			case PLAIN, SLASH, OPENANGLE:
				tokenbyteslice = append(tokenbyteslice, ch.Ch)
				NextChar(f)
			case CLOSEANGLE, CTRL:
				NextChar(f)
				break MainForLoop
			} /* case chstate */
		default:
			fmt.Errorf(" In GetXMLtoken and tokenstate is in default clause of switch case.")
			XMLtoken.State = OTHERERROR
			return XMLtoken, false
		} /* case XMLtoken.State */
	} // indefinite for loop

	XMLtoken.Str = string(tokenbyteslice)
	return XMLtoken, false
} // GetXMLtoken

/* -------------------------------------------------- GetTrack -------------------------------------------- */

func GetTrack(f *bufio.Reader) (trk *TrackType) {
	trk = new(TrackType)
	for {
		XMLtoken, EOF := GetXMLtoken(f)
		if EOF {
			fmt.Println(" Trying to get XML record and got unexpected EOF condition.")
			break
		}
		if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "LOCATION") {
			XMLtoken, EOF = GetXMLtoken(f)
			if EOF || (XMLtoken.State != CONTENTS) {
				fmt.Println(" Trying to get location XML tag and got unexpedted EOF condition or token is not CONTENTS.")
				break
			}
			trk.location = XMLtoken.Str
			_, _ = GetXMLtoken(f) /* retrieve and discard the closinghtml for location */

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "title") {
			XMLtoken, EOF = GetXMLtoken(f)
			if EOF || (XMLtoken.State != CONTENTS) {
				fmt.Println(" Trying to get title XML tag and got unexpected EOF condition or token is not CONTENTS.")
				break
			}
			trk.title = XMLtoken.Str
			_, _ = GetXMLtoken(f) /* retrieve and discard the closinghtml for title */

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "creator") {
			XMLtoken, EOF = GetXMLtoken(f)
			if EOF || (XMLtoken.State != CONTENTS) {
				fmt.Println(" Trying to get creator XML tag and got unexpected EOF condition or token is not CONTENTS.")
				break
			}
			trk.creator = XMLtoken.Str
			_, _ = GetXMLtoken(f) /* retrieve and discard the closinghtml for creator */

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "image") {
			XMLtoken, EOF = GetXMLtoken(f)
			if EOF || (XMLtoken.State != CONTENTS) {
				fmt.Println(" Trying to get image XML record and got unexpected EOF condition or token is not CONTENTS.")
				break
			}
			trk.image = XMLtoken.Str
			_, _ = GetXMLtoken(f) /* retrieve and discard the closinghtml for image */

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "duration") {
			XMLtoken, EOF = GetXMLtoken(f)
			if EOF || (XMLtoken.State != CONTENTS) {
				fmt.Println(" Trying to get duration XML record and got unexpected EOF condition or token is not CONTENTS.")
				break
			}
			trk.duration = XMLtoken.Str
			_, EOF = GetXMLtoken(f) /* retrieve and discard the closinghtml for duration */

		} else if (XMLtoken.State == OPENINGHTML) && strings.HasPrefix(strings.ToLower(XMLtoken.Str), "extension") {
			/* this tag is more complicated than the others because it includes an application and a nested vlc:id tag */
			trk.extension = XMLtoken.Str
			/* retrieve and discard the vlc:id tag in its entirety */
			for {
				XMLtoken, EOF = GetXMLtoken(f)
				if EOF || ((XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "extension")) {
					break
				}
			} // was REPEAT ... UNTIL in original Modula-2 code.

		} else if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "track") {
			break
		} else if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "tracklist") {
			fmt.Println(" In GetTrack and came upon unexpected </tracklist>")
			break
			/* now should never get here because this tag is really part of the extension tag and it's swallowed there */
		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "track") {
			fmt.Println(" in GetTrack and found an unexpected opening track tag")
			break
		} else {
			/*      Have random white space here, typically either a space or a tab before an opening html tag.  Ignore it. */
		} /* if XMLtkn.state == whatever */
	} /* Outer for loop for all contents of this track */
	return trk
} // GetTrack

/***************************** MAIN **********************************/
func main() {

	fmt.Println(" Testing PeekChar, NextChar ")

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: vlc <VLC trackfile.xspf>")
		os.Exit(0)
	}

	fmt.Println()
	fmt.Println(" GOOS =", runtime.GOOS, ".  ARCH=", runtime.GOARCH)
	fmt.Println()

	/*  I may or may not need this later, when I have to start writing the output file.
	    lineDelim := '\n';
	    if runtime.GOOS == "windows" {
	      lineDelim = '\r';
	    }
	*/

	commandline := getcommandline.GetCommandLineString()
	cleancommandline := filepath.Clean(commandline)
	infile, err := os.Open(cleancommandline)
	if err != nil {
		fmt.Println(" Cannot open input file.  Does it exist?  Error is", err)
		os.Exit(1)
	}

	defer infile.Close()
	inputfile := bufio.NewReader(infile)

	/*   Build outfilename */
	BaseFilename := filepath.Base(cleancommandline)
	ExtFilename := filepath.Ext(cleancommandline)
	lastIndex := strings.LastIndex(BaseFilename, ".")
	base := BaseFilename[:lastIndex] // base is the name without extension

	TodaysDateString := MakeDateStr()

	outfilename := base + TodaysDateString + ExtFilename
	outfile, err := os.Create(outfilename)
	if err != nil {
		fmt.Println(" Cannot open outfilename ", outfilename, "  with error ", err)
		os.Exit(1)
	}
	defer outfile.Close()

	outputfile := bufio.NewWriter(outfile)
	defer outputfile.Flush()

	fmt.Println(" Input file name is ", cleancommandline, "  and output file name is ", outfilename)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	firstlineoffile, err := inputfile.ReadString('\n') /* this is the ?xml version line */
	check(err, "Error when reading first line of input file.")
	_, err = outputfile.WriteString(firstlineoffile)
	check(err, "Error when writing first line of output file.")
	//  outputfile.WriteByte('\n');  // I may need WriteRune( ) here.  Turns out this worked.

	secondlineoffile, err := inputfile.ReadString('\n') /* this is the playlist xmlns= line */
	check(err, "Error when reading second line of input file.")
	_, err = outputfile.WriteString(secondlineoffile)
	check(err, "Error when writing second line of output file.")
	//  outputfile.WriteByte('\n');  // But I don't need this at all, as the string includes it.

	thirdlineoffile, err := inputfile.ReadString('\n') /* this is the title line */
	check(err, "Error when reading third line of input file.")
	_, err = outputfile.WriteString(thirdlineoffile)
	check(err, "Error when writing third line of output file.")
	//  outputfile.WriteByte('\n');

	// var XMLcharTypeName = [...]string{"EOL","OpenAngle","CloseAngle","Slash","Plain","Ctrl"};
	var XMLtokenTypeName = [...]string{"EMPTY", "CONTENTS", "OPENINGHTML", "CLOSINGHTML", "OTHERERROR"}
	/*  Now need to test GetXMLtoken
	  for {
	    peekedchar,EOF := PeekChar(inputfile);
	    if EOF {
	      break;
	    }
	// using %c here is a bad idea because of the control characters <CR> and <LF>
	//    fmt.Printf(" peeked char type is %T, char as char is %c\n and %q, char as int is %d\n ",
	//                 peekedchar,peekedchar.Ch,peekedchar.Ch,peekedchar.Ch);
	    fmt.Printf(" peeked char type is %T, char is %q and %d and ",
	                 peekedchar,peekedchar.Ch,peekedchar.Ch);
	    fmt.Printf(" peeked char state is %d, and %s \n",peekedchar.State,XMLcharTypeName[peekedchar.State]);
	    NextChar(inputfile);
	//  I have tested and it works char by char, now I will test until EOF.  That worked.

	    fmt.Print(" pausing: ");
	    scanner.Scan();
	    line := scanner.Text();
	    if err := scanner.Err(); err != nil {
	      fmt.Errorf(" reading std input and got: %v",err);
	      os.Exit(1);
	    }
	    line = strings.ToLower(line);
	    if line == "quit" || line == "exit" {
	      break;
	    }

	  } // read until done
	*/

	for {
		XMLtoken, EOF := GetXMLtoken(inputfile)
		if EOF {
			break
		}
		fmt.Println(" XMLtoken.Str:", XMLtoken.Str, "   XMLtoken.State:", XMLtokenTypeName[XMLtoken.State])

		fmt.Print(" pausing: ")
		scanner.Scan()
		ans := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		ans = strings.TrimSpace(ans)
		ans = strings.ToUpper(ans)
		if ans == "QUIT" || ans == "EXIT" {
			break
		} // end if
		if XMLtoken.Str == "track" {
			trk := GetTrack(inputfile)
			fmt.Printf("Track read is: %#v\n", trk)
		}
	} // end for
} // end main
