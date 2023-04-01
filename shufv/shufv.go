package main // shufv.go

/*
  REVISION HISTORY
  ----------------
  29 Nov 13 -- First version based on TestFilePickerBase, vlcshuffle and TestXMLtoken.
                This uses *.xspf as its default pattern.
  11 Jan 14 -- Discovered that if there is only 1 match to pattern, it fails.  And I will
                have new pattern option also display some potential matches.
  18 Jan 14 -- Add procedure to make sure pattern ends in .xspf.
  21 Sep 16 -- Started conversion to Go.  This is going to take a while.  After experimenting with encoding/xml, I decided
                not to use that package due to its not ignoring garbage characters like <LF> <tab> <space> etc.
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
  16 Oct 18 -- Learned about math/rand having a shuffle function.
  17 Oct 18 -- Adding filepicker for *.xspf pattern.
  18 Oct 18 -- Added check for empty stringslice of filenames.
  24 Aug 21 -- Instead of calling shuffle once, I'm going to imbed it in a loop to shuffle year + month + day + hour + min
                 It's been almost 3 yrs since I've been in this code.  Wow, I've changed my style a lot in that time.  I've switched to line comments
                 and stopped structure end comments for small number of lines, ie, I can see beginning { and ending } on one screen.
                 And I converted to modules.
  13 Feb 22 -- Converted to new API for filepicker.
   6 Feb 23 -- Fixed a comment and will now show number of shuffling iterations.
  10 Mar 23 -- Now called shufv, short for shufflevlc, based on vlcshuffle and xspf code.
  12 Mar 23 -- Came home from Phoenix last night.  Found that waiting for me to close this instance of vlc to delete the temp xspf file will tie up the terminal.
                 I'll change it so that I can delete them afterwards.  The pattern is vlc and a 10-digit number which I can delete myself.
   1 Apr 23 -- StaticCheck found a few issues.
*/

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/jonhadfield/findexec"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	//
	"src/filepicker"
	"src/getcommandline"
	"src/timlibg"
)

const LastCompiled = "Apr 1 2023"
const MaxNumOfTracks = 2048

// const blankline = "                                                                             " // ~70 spaces
// const sepline = "-----------------------------------------------------------------------------"

const (
	EMPTY    = iota
	CONTENTS // was string in the old Modula-2 code
	OPENINGHTML
	CLOSINGHTML
	OTHERERROR
)

//  XMLcharType  which is an enumeration of states of a single character.
// I removed the EOL state, as it no longer applies.  Modula-2 would return a special EOL character, but this language follows the
// C tradition of \r for <CR>, ASCII value of 13, and \n for <LF>, value of 10

const (
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
// shuffle.  And so I don't need NumArray anymore which just shuffles an array of indices into TrackArray.

type TrackType struct {
	location, title, creator, image, duration, extension string
}

var TrackSlice []*TrackType // Global variable.  But still needs to call make in func main.

var lineDelim string
var tabchar = '\t'

var tokenStateNames = []string{"empty", "contents", "openingHTML", "closingHTML", "otherError"}
var verboseFlag, veryVerBoseFlag bool
var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"

func init() {
	TrackSlice = make([]*TrackType, 0, MaxNumOfTracks)
	goVersion := runtime.Version()
	goVersion = goVersion[4:6] // this should be a string of characters 4 and 5, or the numerical digits after Go1.  At the time of writing this, it will be 20.
	goVersionInt, err := strconv.Atoi(goVersion)
	if err == nil {
		fmt.Printf(" Go 1 version is %d\n", goVersionInt)
		if goVersionInt >= 20 { // starting w/ go1.20, rand.Seed() is deprecated.  It will auto-seed if I don't call it, and it wants to do that itself.
			return
		}
	} else {
		fmt.Printf(" ERROR from Atoi: %s\n", err)
	}
	rand.Seed(time.Now().UnixNano())
}

// ---------------------------------------------------------------- getChar -----------------------------

func getChar(f *bytes.Reader) (CharType, error) { // I used to call this peek char, but I don't need a peek char function.  So I'll name this what it is.
	var ch CharType

	b, err := f.ReadByte() // b is a byte slice with size of 1 byte.
	if err != nil {        // These 2 conditions are not essentially different.  They may be in the future.
		return CharType{}, err
	}

	ch.Ch = b
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
	return ch, nil
} // getChar

// -------------------------------------------------------------------------- GetXMLtoken --------------------------------------------

func GetXMLToken(f *bytes.Reader) (TokenType, error) {
	//   The only delimiters are angle brackets.  This is the only routine where input characters are read and processed.
	//   And I rewrote it as an XML token getter.  I do use an unget call in the processing of open angle and close angle, < and >.

	var tokenString strings.Builder // intermediate type to make a string before returning.
	var XMLtoken TokenType

MainForLoop:
	for {
		ch, err := getChar(f)
		if veryVerBoseFlag {
			fmt.Printf("in top of GETXMLToken for loop.  err=%s, ch= %c:%d\n", err, ch.Ch, ch.State)
		}
		if err != nil { // including, and especially, when err is io.EOF
			return TokenType{}, err
		}

		switch XMLtoken.State {
		case EMPTY:
			switch ch.State {
			case PLAIN, SLASH:
				if ch.Ch != ' ' { // ignore leading blanks, but always go to NextChar.
					XMLtoken.State = CONTENTS
					tokenString.WriteByte(ch.Ch)
				}
			case OPENANGLE:
				XMLtoken.State = OPENINGHTML
			case CTRL:
			case CLOSEANGLE:
				XMLtoken.State = OTHERERROR
				f.UnreadByte()
				e := errors.New(" In peekXMLtoken and got an unexpected close angle")
				return XMLtoken, e
			} // case ch.state when the token state is empty
		case CONTENTS: // this case was STRING in the original Modula-2 code
			switch ch.State {
			case PLAIN, SLASH:
				tokenString.WriteByte(ch.Ch)
			case CTRL:
				// ignore control char
			case OPENANGLE: // make openangle char avail for next loop iteration
				f.UnreadByte()
				break MainForLoop
			case CLOSEANGLE:
				e := fmt.Errorf(" In GetXMLToken.  String token %q got closeangle char", tokenString.String())
				f.UnreadByte()
				return XMLtoken, e
			} // case ch.state when the token state is STRING which is the value of the tag
		case OPENINGHTML:
			switch ch.State {
			case PLAIN, OPENANGLE:
				tokenString.WriteByte(ch.Ch)
			case SLASH:
				if tokenString.Len() == 0 {
					XMLtoken.State = CLOSINGHTML // change state of this token from OPENING to CLOSING
				} else {
					tokenString.WriteByte(ch.Ch)
				} // if length == 0
			case CLOSEANGLE, CTRL:
				break MainForLoop
			} // case chstate when the token state is OPENINGHTML
		case CLOSINGHTML:
			switch ch.State {
			case PLAIN, SLASH, OPENANGLE:
				tokenString.WriteByte(ch.Ch)
			case CLOSEANGLE, CTRL:
				break MainForLoop
			} // case chstate
		default:
			f.UnreadByte()
			e := fmt.Errorf(" In GetXMLtoken and tokenstate is in default clause of switch case.  Token = %q", tokenString.String())
			XMLtoken.State = OTHERERROR
			return XMLtoken, e
		} // case XMLtoken.State
	} // indefinite for loop

	XMLtoken.Str = tokenString.String()
	if veryVerBoseFlag {
		fmt.Printf("Exiting GetXMLToken.  XMLToken.Str=%s, XMLToken.State=%s\n", XMLtoken.Str, tokenStateNames[XMLtoken.State])
	}
	return XMLtoken, nil
} // GetXMLToken

// -------------------------------------------------- GetTrack --------------------------------------------

func GetTrack(f *bytes.Reader) (*TrackType, error) {

	// This returns a pointer to TrackType now.  But I don't need to explicitly dereference this pointer in Go.
	// This returns a pointer to TrackType.  Based on what I've learned from Bill Kennedy, I'm rewriting this to more obviously use pointer semantics.

	var trk TrackType
	var err error
	var XMLToken TokenType

	for {
		XMLToken, err = GetXMLToken(f)
		if err != nil {
			return nil, err
			//if !errors.Is(err, io.EOF) { // only show message if not EOF.
			//	fmt.Printf(" Trying to get XML record and got ERROR of %s\n", err)
			//}
			//break
		}
		if (XMLToken.State == OPENINGHTML) && strings.EqualFold(XMLToken.Str, "LOCATION") {
			XMLToken, err = GetXMLToken(f)
			if err != nil {
				e := fmt.Errorf(" Trying to get location XML tag and got ERROR of %s", err)
				return nil, e
			}
			if XMLToken.State != CONTENTS {
				err = fmt.Errorf(" Trying to get location XML tag but token is %#v", XMLToken)
				return nil, err
			}
			trk.location = XMLToken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for location

		} else if (XMLToken.State == OPENINGHTML) && strings.EqualFold(XMLToken.Str, "title") {
			XMLToken, err = GetXMLToken(f)
			if err != nil {
				e := fmt.Errorf(" Trying to get title XML tag and got ERROR of %s", err)
				return nil, e
			}
			if XMLToken.State != CONTENTS {
				err = fmt.Errorf(" Trying to get title XML tag and got unexpected token of %#v", XMLToken)
				return nil, err
			}
			trk.title = XMLToken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for title

		} else if (XMLToken.State == OPENINGHTML) && strings.EqualFold(XMLToken.Str, "creator") {
			XMLToken, err = GetXMLToken(f)
			if err != nil {
				e := fmt.Errorf(" Trying to get creator XML tag and got unexpected ERR of %s", err)
				return nil, e
			}
			if XMLToken.State != CONTENTS {
				err = fmt.Errorf(" Trying to get creator XML tag and token is %#v", XMLToken)
				return nil, err
			}
			trk.creator = XMLToken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for creator

		} else if (XMLToken.State == OPENINGHTML) && strings.EqualFold(XMLToken.Str, "image") {
			XMLToken, err = GetXMLToken(f)
			if err != nil {
				e := fmt.Errorf(" Trying to get image XML record and got unexpected ERROR of %s", err)
				return nil, e
			}
			if XMLToken.State != CONTENTS {
				err = fmt.Errorf(" Trying to get image XML record and got unexpected token of %#v ", XMLToken)
				return nil, err
			}
			trk.image = XMLToken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for image

		} else if (XMLToken.State == OPENINGHTML) && strings.EqualFold(XMLToken.Str, "duration") {
			XMLToken, err = GetXMLToken(f)
			if err != nil {
				e := fmt.Errorf(" Trying to get duration XML record and got ERROR of %s", err)
				return nil, e
			}
			if XMLToken.State != CONTENTS {
				err = fmt.Errorf(" Trying to get duration XML record and got unexpected token of %#v", XMLToken)
				return nil, err
			}
			trk.duration = XMLToken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for duration

		} else if (XMLToken.State == OPENINGHTML) && strings.HasPrefix(strings.ToLower(XMLToken.Str), "extension") {
			// this tag is more complicated than the others because it includes an application and a nested vlc:id tag
			trk.extension = XMLToken.Str
			// retrieve and discard the vlc:id tag in its entirety
			for {
				XMLToken, err = GetXMLToken(f)
				if err != nil {
					return nil, err
				}
				if XMLToken.State == CLOSINGHTML && strings.EqualFold(XMLToken.Str, "extension") {
					break
				}
			} // was REPEAT ... UNTIL in original Modula-2 code.

		} else if (XMLToken.State == CLOSINGHTML) && strings.EqualFold(XMLToken.Str, "track") {
			break
		} else if (XMLToken.State == CLOSINGHTML) && strings.EqualFold(XMLToken.Str, "tracklist") {
			fmt.Println(" In GetTrack and came upon unexpected </tracklist>")
			break
			// now should never get here because this tag is really part of the extension tag and it's swallowed there
		} else if (XMLToken.State == OPENINGHTML) && strings.EqualFold(XMLToken.Str, "track") {
			fmt.Println(" in GetTrack and found an unexpected opening track tag")
			break
		} else {
			//      Have random white space here, typically either a space or a tab before an opening html tag.  Ignore it.
		} // if XMLtkn.state == whatever
	} // Outer for loop for all contents of this track
	return &trk, nil
} // GetTrack

// -------------------------------------------------- PutTrack --------------------------------------------

func PutTrack(f *bufio.Writer, trk *TrackType, TrackNum int) {

	// indivTrackNum used to be incremented here.  I'll have it incremented in the caller now.
	// And the input param is now a pointer to the TrackType, and an array subscript of what was TrackArray in the Modula-2 version of the code.

	// lineDelim is already set in main to be <CR><LF> for Windows and <LF> for everything else (Linux).

	_, err := f.WriteRune(tabchar)
	check(err, " First WriteRune of tabchar in PutTrack and got ")
	f.WriteString("<track>")
	f.WriteString(lineDelim)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteString("<location>")
	f.WriteString(trk.location) // Remember that TrackSlice is a slice of pointers to a TrackType.
	f.WriteString("</location>")
	f.WriteString(lineDelim)

	if len(trk.title) > 0 { // I don't know if I'm required to explicitly dereference this pointer.
		f.WriteRune(tabchar)
		f.WriteRune(tabchar)
		f.WriteRune(tabchar)
		f.WriteString("<title>")
		f.WriteString(trk.title)
		f.WriteString("</title>")
		f.WriteString(lineDelim)
	} /* if have a title tag */

	if len(trk.creator) > 0 {
		f.WriteRune(tabchar)
		f.WriteRune(tabchar)
		f.WriteRune(tabchar)
		f.WriteString("<creator>")
		f.WriteString(trk.creator)
		f.WriteString("</creator>")
		f.WriteString(lineDelim)
	} /* if have a creator tag */

	if len(trk.image) > 0 {
		f.WriteRune(tabchar)
		f.WriteRune(tabchar)
		f.WriteRune(tabchar)
		f.WriteString("<image>")
		f.WriteString(trk.image)
		f.WriteString("</image>")
		f.WriteString(lineDelim)
	} /* if have an image tag */

	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteString("<duration>")
	f.WriteString(trk.duration)
	f.WriteString("</duration>")
	f.WriteString(lineDelim)

	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteString("<")
	f.WriteString(trk.extension)
	f.WriteString(">")
	f.WriteString(lineDelim)

	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteString("  <vlc:id>")
	nstr := strconv.Itoa(TrackNum)
	f.WriteString(nstr)
	f.WriteString("</vlc:id>")
	f.WriteString(lineDelim)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteString(" </extension>")
	f.WriteString(lineDelim)

	f.WriteRune(tabchar)
	f.WriteRune(tabchar)
	f.WriteString("</track>")
	_, err = f.WriteString(lineDelim)
	check(err, " Last write of lineDelim in PutTrack, and got ")
	// return  This was flagged as redundant by StaticCheck.
} // PutTrack

// ---------------------------------------------- ProcessXMLfile    ------------------------------------------

func ProcessXMLfile(inputfile *bytes.Reader, outputfile *bufio.Writer) {

	firstlineoffile, err := readLine(inputfile) // this is the ?xml version line, incl'g <CR><LF> chars
	check(err, "Error when reading first line of input file.")
	_, err = outputfile.WriteString(firstlineoffile)
	check(err, "Error when writing first line of output file.")
	_, err = outputfile.WriteString(lineDelim)
	check(err, "Error when writing first lineDelim of output file.")

	secondlineoffile, err := readLine(inputfile) // this is the playlist xmlns= line
	check(err, "Error when reading second line of input file.")
	_, err = outputfile.WriteString(secondlineoffile)
	check(err, "Error when writing second line of output file.")
	_, err = outputfile.WriteString(lineDelim)
	check(err, "Error when writing 2nd lineDelim of output file.")

	thirdlineoffile, err := readLine(inputfile) // this is the title line
	check(err, "Error when reading third line of input file.")
	_, err = outputfile.WriteString(thirdlineoffile)
	check(err, "Error when writing third line of output file.")
	_, err = outputfile.WriteString(lineDelim)
	check(err, "Error when writing 3rd lineDelim of output file.")

	for { // LOOP to ignoring white space until get the opening tracklist tag
		XMLtoken, err := GetXMLToken(inputfile)
		if err != nil {
			fmt.Printf(" ProcessXMLfile and got ERROR of %s when trying to get <trackList>.  Ending.\n", err)
			return
		}

		if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "tracklist") {
			break // if have tracklist
		}
	} // loop until get opening tracklist tag

	for { // ignoring white space until get the opening track tag.
		XMLtoken, err := GetXMLToken(inputfile)
		if err != nil {
			fmt.Printf(" Trying to get opening track tag and got ERROR of %s.  Ending.\n", err)
			return
		}
		if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "track") {
			break
		}
	}

	TrackSlice = make([]*TrackType, 0, MaxNumOfTracks)

	for { // to read the continuous stream of track tokens
		trackptr, err := GetTrack(inputfile)
		if err != nil {
			fmt.Printf(" After GetTrack and got ERROR of %s\n", err)
			return
		}
		TrackSlice = append(TrackSlice, trackptr) // I'm using a separate assignment to hold the pointer, so I can more easily debug the code, if needed.

		//   This next token will either be a closing tracklist tag or an opening track tag.  If it is not
		//   a closing tracklist tag to end the loop, then we just swallowed the next opening track tag which
		//   is perfect for the GetTrack rtn anyway.

		XMLtoken, err := GetXMLToken(inputfile) // this token should be <track> and then rtn loops again.
		if err != nil {
			fmt.Printf(" Trying to get another track tag and got ERROR of %s.\n", err)
			return
		}

		if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "tracklist") { // unexpected condition
			break
		} // if have closing tracklist
	} // loop to read in more tracks

	NumOfTracks := len(TrackSlice)
	fmt.Println("Last track number read is ", NumOfTracks)

	t0 := time.Now()

	//   time to shuffle

	swapFnt := func(i, j int) {
		TrackSlice[i], TrackSlice[j] = TrackSlice[j], TrackSlice[i]
	}

	Time := timlibg.GetDateTime()
	shuffling := Time.Month + Time.Day + Time.Hours + Time.Minutes + Time.Year + Time.Seconds
	for k := 0; k < shuffling; k++ {
		rand.Shuffle(len(TrackSlice), swapFnt)
	}
	// Finished shuffling.

	timeToShuffle := time.Since(t0) // timeToShuffle is a Duration type, which is an int64 but has methods.
	fmt.Printf(" It took %s to shuffle this file %d times.\n", timeToShuffle.String(), shuffling)
	fmt.Println()

	// Write the output file.
	_, err = outputfile.WriteRune(tabchar)
	check(err, " Starting to write the shuffled tracklist to the output file and got error: ")
	outputfile.WriteString("<trackList>")
	outputfile.WriteString(lineDelim)

	for c := 0; c < len(TrackSlice); c++ {
		PutTrack(outputfile, TrackSlice[c], c)
	}

	outputfile.WriteRune(tabchar)
	outputfile.WriteString("</trackList>")
	outputfile.WriteString(lineDelim)

	for { // to read and write the rest of the lines
		line, err := readLine(inputfile)
		if errors.Is(err, io.EOF) {
			break
		}
		check(err, " Reading final lines of the inputfile and got this error: ")
		_, err = outputfile.WriteString(line)
		check(err, " Writing final line of the output file and got this error: ")
	} // final read and write loop

} // ProcessXMLfile

// -------------------------------------------- check ---------------------------------------------
func check(e error, msg string) {
	if e != nil {
		fmt.Printf("%s : ", msg)
		panic(e)
	}
}

// -------------------------------------------- min ---------------------------------------------
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// ------------------------------------------- MAIN --------------------------------
func main() {

	flag.BoolVar(&verboseFlag, "v", false, " Verbose Mode.")
	flag.BoolVar(&veryVerBoseFlag, "vv", false, " Very verbose mode.")
	flag.Parse()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	fmt.Printf(" %s is a Shuffling program for the tracks in a vlc file, last altered %s, compiled by %s, and timestamp is %s.\n\n",
		os.Args[0], LastCompiled, runtime.Version(), LastLinkedTimeStamp)

	InExtDefault := ".xspf"
	ans := ""
	Filename := ""
	FileExists := false

	if flag.NArg() == 0 { // need to use filepicker
		filenames, err := filepicker.GetFilenames("*.xspf")
		if err != nil {
			fmt.Printf(" filepicker returned error %s\n.  Exiting.", err)
			os.Exit(1)
		}
		if len(filenames) == 0 {
			fmt.Println(" No filenames found that match *.xspf pattern.  Exiting")
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ { // goes 0 .. 25, or a .. z
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		_, err = fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil {
			ans = "0"
		}
		i, er := strconv.Atoi(ans)
		if er == nil {
			Filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			Filename = filenames[i]
		}
		fmt.Println(" Picked filename is", Filename)
	} else { // will use filename entered on commandline
		ns := getcommandline.GetCommandLineString()
		Filename = filepath.Clean(ns)

		if strings.Contains(Filename, ".") {
			_, err := os.Stat(Filename)
			if err == nil {
				FileExists = true
			}
		} else {
			FullFilename := Filename + InExtDefault
			_, err := os.Stat(FullFilename)
			if err == nil {
				FileExists = true
				Filename = FullFilename
			}
		}

		if !FileExists {
			fmt.Println(" File", Filename, " does not exist.  Exiting.")
			os.Exit(1)
		}
		fmt.Println(" Filename is", Filename)
	}

	//infile, err := os.Open(Filename)
	fileBuf, err := os.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Cannot open input file,", Filename, ".  Does it exist?  Error is", err)
		os.Exit(1)
	}
	fileRdr := bytes.NewReader(fileBuf)

	//defer infile.Close()
	//inputfile := bufio.NewReader(infile)
	//   Build outfilename not needed, and will be replaced by a temp file.
	//BaseFilename := filepath.Base(Filename)
	//ExtFilename := filepath.Ext(Filename)
	//lastIndex := strings.LastIndex(BaseFilename, ".")
	//base := BaseFilename[:lastIndex] // base is the name without extension
	//TodaysDateString := MakeDateStr()
	//outfilename := base + TodaysDateString + ExtFilename

	//outfile, err := os.Create(outfilename)
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Printf(" os.Getwd() call ERROR is %s\n", err)
		os.Exit(1)
	}
	outputFile, err := os.CreateTemp(workingDir, "vlc")
	tempFilename := outputFile.Name()
	if err != nil {
		fmt.Printf(" os.CreateTemp ERROR is %s\n", err)
		os.Exit(1)
	}
	if verboseFlag {
		fmt.Printf(" Temp filename is %s\n", tempFilename)
	}
	outfileBuf := bufio.NewWriter(outputFile)

	if runtime.GOOS == "windows" {
		lineDelim = "\r\n"
	} else {
		lineDelim = "\n"
	}

	ProcessXMLfile(fileRdr, outfileBuf)
	outfileBuf.Flush()
	outputFile.Close()
	temp := tempFilename + ".xspf"
	os.Rename(tempFilename, temp)

	// Now have the output file written, flushed and closed.  Now to pass it to vlc

	var path, searchPath string // vlcPath is defined as a global, above, that hard codes the directory for vlc on Windows.
	path = os.Getenv("PATH")
	vPath, ok := os.LookupEnv("VLCPATH")
	if ok {
		vlcPath = strings.ReplaceAll(vPath, `"`, "") // Here I use back quotes to insert a literal quote.
	}
	if runtime.GOOS == "windows" {
		searchPath = vlcPath + ";" + path
	} else if runtime.GOOS == "linux" && ok {
		searchPath = vlcPath + ":" + path
	} else { // on linux and not ok, meaning environment variable VLCPATH is empty.
		searchPath = path
	}

	var vlcStr string
	if runtime.GOOS == "windows" {
		vlcStr = findexec.Find("vlc", searchPath) //Turns out that vlc was not in the path.  But it shows up when I use "which vlc".  So it seems that findexec doesn't find it on my win10 system.  So I added it to the path.
	} else if runtime.GOOS == "linux" {
		vlcStr = findexec.Find("vlc", "") // calling vlc without a console.
	}

	if vlcStr == "" {
		fmt.Printf(" vlcStr is null.  Exiting ")
		os.Exit(1)
	}

	execCmd := exec.Command(vlcStr, temp)

	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	//err = execCmd.Run()
	err = execCmd.Start()
	if err != nil {
		fmt.Printf(" Error returned by running vlc %s is %s\n", temp, err)
	}

	// And now need to erase the temp file.
	//os.Remove(temp)  No, I won't do this so the terminal is freed up right after this pgm runs.

} //  vlc main

// ----------------------------------------------------- readLine ------------------------------------------------------
// Needed as a bytes reader does not have a readString method.

func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byte, err := r.ReadByte()
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
		if byte == '\n' {
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byte)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine
/*
// ---------------------------------------------------------------- PeekChar -----------------------------
// peeks at the next char without advancing fileptr.  The filepointer is advanced by ReadChar, below.

func PeekChar(f *bufio.Reader) (ch CharType, EOF bool) {
	b := make([]byte, 1)
	b, err := f.Peek(1) // b is a byte slice with size of 1 byte.
	if err == io.EOF {  // basically any error is returned as EOF, because of the n==0 condition.
		return CharType{}, true
	} else if err != nil { // These 2 conditions are not essentially different.  They may be in the future.
		return CharType{}, true
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

	set filepointer back one byte from its current position.  This will have Read get this byte again.
	    Nevermind, by using bufio instead of os, I found a peek function.
	   NextFileOffset, err = f.Seek(-1,1);
	   if err != nil { fmt.Errorf(" Error from PeekChar call to Seek: %v\n",err); }


} // PeekChar

	// ------------------------------------------------------------- NextChar ----------------------------
// advances filepointer.  PeekChar does not advance fileptr.  This rtn will throw the character away, as
// it assumes that PeekChar already got and processed the character.

func NextChar(f *bufio.Reader) {
	_, err := f.Discard(1) // Discard 1 byte.  Throw away the return saying how many bytes were actually discarded
	check(err, "In DiscardChar and got err:")
} // DiscardChar
*/

/*
// -------------------------------------------------------------------------- GetXMLtoken --------------------------------------------


func GetXMLtoken(f *bufio.Reader) (XMLtoken TokenType, EOFFLG bool) {

	   This will use the bufio file operations as I want this as a character stream.
	   The only delimiters are angle brackets.  This is the only routine where input characters are read and processed.
	   And I rewrote it to just exist as an XML token getter.  I need peeking functionality, as in a few sections below I don't advance the file pointer after peeking at
	   the next byte.


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
				NextChar(f) // discard byte, but change state to begin a tag
			case CTRL:
				NextChar(f)  discard these
			case CLOSEANGLE:
				fmt.Errorf(" In peekXMLtoken and got an unexpected close angle.")
				XMLtoken.State = OTHERERROR
				return XMLtoken, false
			} case ch.state when the token state is empty
		case CONTENTS: // this case was STRING in the original Modula-2 code
			switch ch.State {
			case PLAIN, SLASH:
				tokenbyteslice = append(tokenbyteslice, ch.Ch) // continue building the contents string
				NextChar(f)
			case CTRL:
				NextChar(f) ignore control char
			case OPENANGLE: openangle char is still avail for next loop iteration
				break MainForLoop
			case CLOSEANGLE:
				fmt.Errorf(" In GetXMLToken.  String token got closeangle char")
			} // case ch.state when the token state is STRING which is the value of the tag
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
				} if length == 0
			case CLOSEANGLE, CTRL:
				NextChar(f)
				break MainForLoop
			} case chstate when the token state is OPENINGHTML
		case CLOSINGHTML:
			switch ch.State {
			case PLAIN, SLASH, OPENANGLE:
				tokenbyteslice = append(tokenbyteslice, ch.Ch)
				NextChar(f)
			case CLOSEANGLE, CTRL:
				NextChar(f)
				break MainForLoop
			} case chstate
		default:
			fmt.Errorf(" In GetXMLtoken and tokenstate is in default clause of switch case.")
			XMLtoken.State = OTHERERROR
			return XMLtoken, false
		} case XMLtoken.State
	} // indefinite for loop

	XMLtoken.Str = string(tokenbyteslice)
	return XMLtoken, false
} // GetXMLtoken
*/

/* ------------------------------------------- MakeDateStr ---------------------------------------------------

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

-------------------------------------------- Shuffle ----------------------------------------------------

func Shuffle() {  replaced by rand.Shuffle
	// Shuffle the array by passing once through the array, swapping each element with another, randomly chosen, element.

	n := len(TrackSlice)

	for c := 1; c < n; c++ { // c is not used in the loop below.  It's just an outer loop counter.
		for i := n - 1; i > 0; i-- {
			// swap element i with any element at or below that place.  Note that i is not allowed to be 0, but k can be
			k := rand.Intn(i)
			TrackSlice[i], TrackSlice[k] = TrackSlice[k], TrackSlice[i] // Go swap idiom, to swap pointers to a track that's held in TrackSlice
		}
	}
} // Shuffle;

*/
