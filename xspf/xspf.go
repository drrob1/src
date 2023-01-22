package main

/*
  REVISION HISTORY
  ----------------
  29 Nov 13 -- First version based on TestFilePickerBase, vlcshuffle and TestXMLtoken. This uses *.xspf as its default pattern.
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
  16 Oct 18 -- Learned about math/rand having a shuffle function.
  17 Oct 18 -- Adding filepicker for *.xspf pattern.
  18 Oct 18 -- Added check for empty stringslice of filenames.
  24 Aug 21 -- Instead of calling shuffle once, I'm going to imbed it in a loop to shuffle year + month + day + hour + min
                 It's been almost 3 yrs since I've been in this code.  Wow, I've changed my style a lot in that time.  I've switched to line comments
                 and stopped structure end comments for small number of lines, ie, I can see beginning { and ending } on one screen.
                 And I converted to modules.
  13 Feb 22 -- Converted to new API for filepicker.
  19 Jan 23 -- Now called xspf.go.  It will read a vlc xsfp file, shuffle the filenames/titles it finds, and then call vlc w/ possibly a subslice of those names.
  20 Jan 23 -- It's finally working.  I'm going to stop now.  Maybe I'll come back to this another time, but I've had enough for now.
*/

import (
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
	"src/timlibg"
)

const LastCompiled = "20 Jan 2023"
const MaxNumOfTracks = 2048 // Initial capacity
const extension = ".xspf"

//const blankline = "                                                                             " // ~70 spaces
//const sepline = "-----------------------------------------------------------------------------"

const ( // States of tokens
	EMPTY    = iota
	CONTENTS // was string in the old Modula-2 code
	OPENINGHTML
	CLOSINGHTML
	OTHERERROR
)

var tokenStateNames = []string{"empty", "contents", "openingHTML", "closingHTML", "otherError"}

//  XMLcharType  which is an enumeration of states of a single character.
// I removed the EOL state, as it no longer applies.  Modula-2 would return a special EOL character, but this language follows the
// C tradition of \r for <CR>, ASCII value of 13, and \n for <LF>, value of 10

const ( // States of characters
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
var veryVerboseFlag bool
var verboseFlag bool

//var lineDelim string
//var tabChar = '\t'
//  var indivTrackNum = 0;  Still don't think I need this.

func init() {
	rand.Seed(time.Now().UnixNano())
	TrackSlice = make([]*TrackType, 0, MaxNumOfTracks)
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
	/*
	   The only delimiters are angle brackets.  This is the only routine where input characters are read and processed.
	   And I rewrote it as an XML token getter.  I do use an unget call in the processing of open angle and close angle, < and >.
	*/

	var tokenString strings.Builder // intermediate type to make a string before returning.
	var XMLtoken TokenType

MainForLoop:
	for {
		ch, err := getChar(f)
		if veryVerboseFlag {
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
				e := fmt.Errorf(" In peekXMLtoken and got an unexpected close angle.")
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
	if veryVerboseFlag {
		fmt.Printf("Exiting GetXMLToken.  XMLToken.Str=%s, XMLToken.State=%s\n", XMLtoken.Str, tokenStateNames[XMLtoken.State])
	}
	return XMLtoken, nil
} // GetXMLToken

// -------------------------------------------------- GetTrack --------------------------------------------

func GetTrack(f *bytes.Reader) (*TrackType, error) {

	// This returns a pointer to TrackType.  Based on what I've learned from Bill Kennedy, I'm rewriting this to more obviously use pointer semantics.

	var trk TrackType
	var err error
	var XMLtoken TokenType

	//for { // LOOP to ignoring white space until get the opening track tag
	//	XMLtoken, err := GetXMLToken(f)
	//	if verboseFlag {
	//		fmt.Printf("GetTrack looking for opening track XMLtoken at about line 238: token= %s, state=%s\n", XMLtoken.Str, tokenStateNames[XMLtoken.State])
	//	}
	//	if err != nil {
	//		e := fmt.Errorf(" GetTrack and ERROR %s occurred when trying to get <track>.", err)
	//		return nil, e
	//	}
	//
	//	if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "track") {
	//		break // if have tracklist, break this inner for loop.
	//	}
	//} // loop until get opening tracklist tag

	for {
		XMLtoken, err = GetXMLToken(f)
		if verboseFlag {
			fmt.Printf("In GetTrack.  err=%s, XMLtoken = %#v\n", err, XMLtoken)
		}
		if err != nil {
			if !errors.Is(err, io.EOF) { // only show message if not EOF.
				fmt.Printf(" Trying to get XML record and got error of %s\n", err)
			}
			break
		}
		if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "LOCATION") { // This is the full filename
			XMLtoken, err = GetXMLToken(f)
			if err != nil || (XMLtoken.State != CONTENTS) {
				fmt.Printf(" Trying to get location XML tag and got %s, or token is not CONTENTS.\n", err)
				break
			}
			trk.location = XMLtoken.Str
			if verboseFlag {
				fmt.Printf(" in GetTrack.  trk.Location=%s, XMLtoken=%s, and state = %s\n", trk.location, XMLtoken.Str, tokenStateNames[XMLtoken.State])
			}
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for location

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "title") {
			XMLtoken, err = GetXMLToken(f)
			if err != nil || (XMLtoken.State != CONTENTS) {
				fmt.Printf(" Trying to get title XML tag and got error %s, or token is not CONTENTS.  XMLtoken=%s, XMLState=%s\n", err, XMLtoken.Str, tokenStateNames[XMLtoken.State])
				break
			}
			trk.title = XMLtoken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for title

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "creator") {
			XMLtoken, err = GetXMLToken(f)
			if err != nil || (XMLtoken.State != CONTENTS) {
				fmt.Printf(" Trying to get creator XML tag and got %s, or token is not CONTENTS.\n", err)
				break
			}
			trk.creator = XMLtoken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for creator

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "image") {
			XMLtoken, err = GetXMLToken(f)
			if err != nil || (XMLtoken.State != CONTENTS) {
				fmt.Println(" Trying to get image XML record and got unexpected EOF condition or token is not CONTENTS.")
				break
			}
			trk.image = XMLtoken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for image

		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "duration") {
			XMLtoken, err = GetXMLToken(f)
			if err != nil || (XMLtoken.State != CONTENTS) {
				fmt.Printf(" Trying to get duration XML record and got %s, or token is not CONTENTS.", err)
				break
			}
			trk.duration = XMLtoken.Str
			_, _ = GetXMLToken(f) // retrieve and discard the closinghtml for duration

		} else if (XMLtoken.State == OPENINGHTML) && strings.HasPrefix(strings.ToLower(XMLtoken.Str), "extension") {
			// this tag is more complicated than the others because it includes an application and a nested vlc:id tag
			trk.extension = XMLtoken.Str
			// retrieve and discard the vlc:id tag in its entirety
			for {
				XMLtoken, err = GetXMLToken(f)
				if err != nil || ((XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "extension")) {
					break
				}
			} // was REPEAT ... UNTIL in original Modula-2 code.

		} else if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "track") {
			if verboseFlag {
				fmt.Printf(" GetTrack line 322: Got Closing HTML of track, so will break for loop here.\n")
			}
			break
		} else if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "tracklist") {
			//fmt.Println(" In GetTrack and came upon unexpected </tracklist>")
			//break
			// now should never get here because this tag is really part of the extension tag and it's swallowed there
		} else if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "track") {
			//fmt.Println(" in GetTrack and found an unexpected opening track tag")
			//break
		} else {
			//      Have random white space here, typically either a space or a tab before an opening html tag.  Ignore it.
		} // end if XMLtkn.state == whatever
	} // Outer for loop for all contents of this track

	if verboseFlag {
		fmt.Printf("Exiting GetTrack.  Err=%s,  Track= %#v\n", err, trk)
	}

	//if err == io.EOF && trk.location != "" { // If got EOF but already collected a track, clear the error before returning it.
	//	err = nil
	//}

	return &trk, err // more obviously uses pointer semantics.

} // GetTrack

// ---------------------------------------------- getFilenames, formerly ProcessXMLfile ------------------------------------------
// Will read the xspf file and return a randomized slice of filenames/titles as a []string

func getFileNames(inputFile *bytes.Reader) ([]string, error) {
	var fn []string
	/*  Since I'm not writing a file, I don't need to capture these first 3 lines.
	_, err := readLine(inputFile) // this is the ?xml version (first) line, incl'g <CR><LF> chars
	if err != nil {
		return nil, err
	}

	_, err = readLine(inputFile) // this is the playlist xmlns= (second) line
	if err != nil {
		return nil, err
	}

	_, err = readLine(inputFile) // this is the title (third) line
	if err != nil {
		return nil, err
	}
	*/
	for { // LOOP to ignoring white space until get the opening tracklist tag
		XMLtoken, err := GetXMLToken(inputFile)
		if verboseFlag {
			fmt.Printf("getFileNames looking for opening tracklist XMLtoken: %#v\n", XMLtoken)
		}
		if err != nil {
			e := fmt.Errorf(" ProcessXMLfile and %s when trying to get <trackList>.  Ending.", err)
			return nil, e
		}

		if (XMLtoken.State == OPENINGHTML) && strings.EqualFold(XMLtoken.Str, "tracklist") {
			break // if have tracklist, break this inner for loop.
		}
	} // loop until get opening tracklist tag

	TrackSlice = make([]*TrackType, 0, MaxNumOfTracks)

	for { // to read the continuous stream of track tokens
		trackPtr, err := GetTrack(inputFile)
		if verboseFlag {
			//fmt.Printf("in getFileNames after GetTrack(inputfile) call: err = %s, and *track is %#v\n", err, *trackPtr) Don't need to dereference the pointer.
			fmt.Printf("in getFileNames after GetTrack(inputfile) call at about line 386: err = %s, track is %#v\n", err, trackPtr)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf(" In getFileNames at about line 393.  err = %s, trackPtr is %#v\n", err, trackPtr)
			return nil, err
		}
		TrackSlice = append(TrackSlice, trackPtr) // I'm using a separate assignment to hold the pointer, so I can more easily debug the code, if needed.
		fn = append(fn, trackPtr.location)

		//   This next token will either be a closing tracklist tag or an opening track tag.  If it is not
		//   a closing tracklist tag to end the loop, then we just swallowed the next opening track tag which
		//   is perfect for the GetTrack rtn anyway.

		XMLtoken, err := GetXMLToken(inputFile) // this token should be <track> and then rtn loops again.
		if verboseFlag {
			fmt.Printf(" In GetFileNames at about line 405.  XMLtoken=%s, XMLtoken.State=%s\n", XMLtoken.Str, tokenStateNames[XMLtoken.State])
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Printf("In getFileNames and got EOF\n")
				break
			}
			e := fmt.Errorf(" Trying to get another track tag and got %s.  Ending.", err)
			return nil, e
		}

		//if (XMLtoken.State == CLOSINGHTML) && strings.EqualFold(XMLtoken.Str, "tracklist") { // unexpected condition
		//	e := fmt.Errorf(" Unexpected tracklist token.  Token = %#V", XMLtoken)
		//	return nil, e
		//} // if have closing tracklist
	} // loop to read in more tracks

	fmt.Printf("In getFileNames before sort.  Length of TrackSlice = %d, and length of fn = %d\n", len(TrackSlice), len(fn))

	t0 := time.Now()

	//   It's time to shuffle

	//swapfnt := func(i, j int) {
	//	TrackSlice[i], TrackSlice[j] = TrackSlice[j], TrackSlice[i]
	//}

	swapFcn := func(i, j int) {
		fn[i], fn[j] = fn[j], fn[i]
	}

	Time := timlibg.GetDateTime()
	rand.Seed(t0.Unix())
	shuffling := Time.Month + Time.Day + Time.Hours + Time.Minutes + Time.Year + Time.Seconds
	for k := 0; k < shuffling; k++ {
		rand.Shuffle(len(fn), swapFcn)
	}
	// Finished shuffling.

	timeToShuffle := time.Since(t0) // timeToShuffle is a Duration type, which is an int64 but has methods.
	fmt.Printf(" It took %s to shuffle this file.\n", timeToShuffle.String())
	fmt.Println()

	// Used to write the output file, here.  Now that I've created the slice to feed to vlc, I'll return it.
	return fn, nil

	//_, err = outputfile.WriteRune(tabchar)
	//check(err, " Starting to write the shuffled tracklist to the output file and got error: ")
	//outputfile.WriteString("<trackList>")
	//outputfile.WriteString(lineDelim)
	//
	//for c := 0; c < len(TrackSlice); c++ {
	//	PutTrack(outputfile, TrackSlice[c], c)
	//}
	//
	//outputfile.WriteRune(tabchar)
	//outputfile.WriteString("</trackList>")
	//outputfile.WriteString(lineDelim)
	//
	//for { // to read and write the rest of the lines
	//	line, err := inputfile.ReadString('\n')
	//	if err == io.EOF {
	//		break
	//	}
	//	check(err, " Reading final lines of the inputfile and got this error: ")
	//	_, err = outputfile.WriteString(line)
	//} // final read and write loop

} // end getFilenames, formerly ProcessXMLFile

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
	var ans, filename string
	var fileExistsFlag bool
	var searchPath, path, vPath string
	var vlcPath = "C:\\Program Files\\VideoLAN\\VLC"
	var numNames int

	fmt.Printf(" %s for the tracks in a vlc file with xsfp extension.  Last altered %s, compiled by %s\n\n", os.Args[0], LastCompiled, runtime.Version())

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

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

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), " This pgm will match an input regexp using smart case, against all filenames in the current directory\n")
		fmt.Fprintf(flag.CommandLine.Output(), " shuffle them, and then output 'n' of them on the command line to vlc.\n")
		fmt.Fprintf(flag.CommandLine.Output(), " %s has timestamp of %s, full name of executable is %s and vlcPath is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, execName, vlcPath)
		fmt.Fprintf(flag.CommandLine.Output(), " Usage: launchv <options> <input-regex> where <input-regex> cannot be empty. \n")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	flag.BoolVar(&verboseFlag, "v", false, "Verbose mode flag.")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "Verbose mode flag.")
	flag.IntVar(&numNames, "n", 40, " Number of file names to output on the commandline to vlc.")
	flag.Parse()

	if veryVerboseFlag {
		verboseFlag = true
	}

	if verboseFlag {
		fmt.Printf(" %s has timestamp of %s, and full name of executable is %s.\n",
			ExecFI.Name(), LastLinkedTimeStamp, execName)
		//fmt.Printf(" vlcPath = %s, searchPath is: \n", vlcPath)
		//fmt.Printf(" SearchPath: %#v\n", searchPath)
	}

	if flag.NArg() == 0 { // need to use filepicker
		filenames, err := filepicker.GetFilenames("*.xspf")
		if err != nil {
			fmt.Fprintf(os.Stderr, " filepicker returned error %v\n.  Exiting.", err)
			os.Exit(1)
		}
		if len(filenames) == 0 {
			fmt.Println(" No filenames found that match *.xspf pattern.  Exiting")
			os.Exit(1)
		}

		// display filenames found by filepicker routine
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
			filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			filename = filenames[i]
		}
		fmt.Printf(" Picked filename is %s\n", filename)
	} else { // will use filename entered on commandline
		filename = filepath.Clean(flag.Arg(0))

		if strings.Contains(filename, ".") { // assume that if there's a dot, then don't need to append the extension.
			_, err := os.Stat(filename)
			if err == nil {
				fileExistsFlag = true
			}
		} else { // no dot, so need the extension.
			fullFilename := filename + extension
			_, err := os.Stat(fullFilename)
			if err == nil {
				fileExistsFlag = true
				filename = fullFilename
			}
		}

		if !fileExistsFlag {
			fmt.Println(" File", filename, " does not exist.  Exiting.")
			os.Exit(1)
		}
		fmt.Println(" Filename is", filename)
	}

	startTime := time.Now()
	infile, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(" Cannot open input file,", filename, ".  Does it exist?  Error is", err)
		os.Exit(1)
	}
	inFileReader := bytes.NewReader(infile)

	//   Build outfilename.  Nevermind.  Don't need this here
	//BaseFilename := filepath.Base(filename)
	//ExtFilename := filepath.Ext(filename)
	//lastIndex := strings.LastIndex(BaseFilename, ".")
	//base := BaseFilename[:lastIndex] // base is the name without extension
	//
	//TodaysDateString := MakeDateStr()
	//
	//outfilename := base + TodaysDateString + ExtFilename
	//outfile, err := os.Create(outfilename)
	//if err != nil {
	//	fmt.Println(" Cannot open outfilename ", outfilename, "  with error ", err)
	//	os.Exit(1)
	//}
	//defer outfile.Close()
	//
	//outputfile := bufio.NewWriter(outfile)
	//defer outputfile.Flush()

	//fmt.Println()
	//fmt.Println(" GOOS =", runtime.GOOS, ".  ARCH=", runtime.GOARCH)
	//fmt.Println()
	//if runtime.GOOS == "windows" {
	//	lineDelim = "\r\n"
	//} else {
	//	lineDelim = "\n"
	//}

	fileNames, err := getFileNames(inFileReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, " Error from getFileNames is %s.  Good-bye\n", err)
		os.Exit(1)
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

	var execCmd *exec.Cmd

	variadicParam := make([]string, 0, len(fileNames))
	variadicParam = append(variadicParam, fileNames...)
	n := minInt(numNames, len(fileNames))
	if n > 0 {
		variadicParam = variadicParam[:n]
	}

	if runtime.GOOS == "windows" {
		execCmd = exec.Command(vlcStr, variadicParam...)
	} else if runtime.GOOS == "linux" { // I'm ignoring this for now.  I'll come back to it after I get the Windows code working.
		execCmd = exec.Command(vlcStr, fileNames...)
	}

	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	e := execCmd.Start()
	if e != nil {
		fmt.Printf(" Error returned by running vlc %s is %v\n", variadicParam, e)
	}
	fmt.Printf(" It took %s for the entire %s program to run.\n", time.Since(startTime), os.Args[0])

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

// ------------------------------- minInt ----------------------------------------

func minInt(i, j int) int {
	if i <= j {
		return i
	}
	return j
}

/* ------------------------------------------- MakeDateStr ---------------------------------------------------* */
/*
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

*/

/* -------------------------------------------- Shuffle ----------------------------------------------------

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
// -------------------------------------------------- PutTrack --------------------------------------------
/*
//func PutTrack(f *bufio.Writer, trk *TrackType, TrackNum int) {
//
//	// indivTrackNum used to be incremented here.  I'll have it incremented in the caller now.
//	// And the input param is now a pointer to the TrackType, and an array subscript of what was TrackArray in the Modula-2 version of the code.
//
//	// lineDelim is already set in main to be <CR><LF> for Windows and <LF> for everything else (Linux).
//
//	_, err := f.WriteRune(tabchar)
//	check(err, " First WriteRune of tabchar in PutTrack and got ")
//	f.WriteString("<track>")
//	f.WriteString(lineDelim)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteString("<location>")
//	f.WriteString(trk.location) // Remember that TrackSlice is a slice of pointers to a TrackType.
//	f.WriteString("</location>")
//	f.WriteString(lineDelim)
//
//	if len(trk.title) > 0 { // I don't know if I'm required to explicitly dereference this pointer.
//		f.WriteRune(tabchar)
//		f.WriteRune(tabchar)
//		f.WriteRune(tabchar)
//		f.WriteString("<title>")
//		f.WriteString(trk.title)
//		f.WriteString("</title>")
//		f.WriteString(lineDelim)
//	}
//
//	if len(trk.creator) > 0 {
//		f.WriteRune(tabchar)
//		f.WriteRune(tabchar)
//		f.WriteRune(tabchar)
//		f.WriteString("<creator>")
//		f.WriteString(trk.creator)
//		f.WriteString("</creator>")
//		f.WriteString(lineDelim)
//	}
//
//	if len(trk.image) > 0 {
//		f.WriteRune(tabchar)
//		f.WriteRune(tabchar)
//		f.WriteRune(tabchar)
//		f.WriteString("<image>")
//		f.WriteString(trk.image)
//		f.WriteString("</image>")
//		f.WriteString(lineDelim)
//	}
//
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteString("<duration>")
//	f.WriteString(trk.duration)
//	f.WriteString("</duration>")
//	f.WriteString(lineDelim)
//
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteString("<")
//	f.WriteString(trk.extension)
//	f.WriteString(">")
//	f.WriteString(lineDelim)
//
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteString("  <vlc:id>")
//	nstr := strconv.Itoa(TrackNum)
//	f.WriteString(nstr)
//	f.WriteString("</vlc:id>")
//	f.WriteString(lineDelim)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteString(" </extension>")
//	f.WriteString(lineDelim)
//
//	f.WriteRune(tabchar)
//	f.WriteRune(tabchar)
//	f.WriteString("</track>")
//	_, err = f.WriteString(lineDelim)
//	check(err, " Last write of lineDelim in PutTrack, and got ")
//	return
//} // PutTrack
//
*/
// -------------------------------------------- check ---------------------------------------------
//func check(e error, msg string) {
//	if e != nil {
//		fmt.Printf("%s : ", msg)
//		panic(e)
//	}
//}
