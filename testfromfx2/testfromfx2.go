// testfromfx2.go
package main

import (
	"bufio"
	"bytes"
	//	"encoding/csv"
	"filepicker"
	"fmt"
	"getcommandline"
	//	"go/scanner"
	"io/ioutil"
	//	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"timlibg"
	//	"tknptr"
	"tokenize"
)

const lastModified = "19 Oct 20"

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
                which means open financial exchange (for/of information).  New name is ofx2csv.go
		I think I will first process the file using something like toascii.
  19 Oct 17 -- Added filepicker code
   1 Nov 17 -- Added output of $ for footer amount.
  25 Dec 17 -- Decided to try changing date format to match ISO8601 YYYY-MM-DD required for sqlite.
  30 Dec 17 -- Discovered that Access won't handle yyyy-mm-dd format, only Excel will.  Now I need
                 to write out 2 different files, one for Access and one for Sqlite.db.  And I need to
				 append 2 commas for the sqlite file.
  31 Mar 19 -- Noticed that this pgm defaults to .qfx files.  I decided to have it default to both .ofx and .qfx files.
   5 Sep 20 -- Still not showing .ofx files, and removed default CHK part of the pattern.
                 Put back the CHK part of pattern because this file is only designed for the Citibank checking files.
   6 Sep 20 -- Added a comment about passing Transactions slice globally.
   7 Sep 20 -- Now called fromfx, to mean from qfx or ofx.  My intent is to cover both the CHK files and cc files.
                 I'll look for CHK in the selected filename to distinguish.
  17 Sep 20 -- Using the csv routines from Go did not work for SQLiteStudio.  It's not reading the date again.
                 I have to explicitly quote the output.  And also always write Windows line endings \r\n.  Else
                 SQLiteStudio does not process the lines of the file correctly.  I just checked, and the
                 Modula-2 version always wrote Windows line endings.
                 This Go routine reads all the lines into a slice of records/struct's.  I did not write it that way
                 in Modula-2.  In Modula-2 I read 1 line and then wrote out 1 line, one by one.  I probably could
                 have created an array of RECORD type, but I didn't.
                 I guess that Go makes dynamic arrays just so easy that I never did that in Modula-2.  I could have
                 defined a large enough static array.  I just never went down that road, I guess.
                 The Modula-2 code writes the memo field as FITID + "  " + memo + ": " + comment, where I enter comment
                 myself w/ each run of the pgm.  I'm trying out adding the FITID numbers to Descript and see how I
                 like it.
   2 Oct 20 -- qbo files will populate the filepicker menu.  Filepicker now uses case insensitive flag.  Stop code added.
   3 Oct 20 -- Now called fromfx2, and I intend this to have an ungettoken, that I have to implement by storing a token
                 and checking if there is a valid token that was ungotten before fetching a new one.
   4 Oct 20 -- Now called testfromfx2, to be able to test the ungettoken code I wrote yesterday.
  19 Oct 20 -- Removed useless comments of dead code.
*/

const ( // intended for ofxCharType
	eol = iota // so eol = 0, and so on.  And the zero val needs to be DELIM.
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
	// othererror  // unused at the moment, but I'll just comment it out.
)

const ( // intended for inputstate
	citichecking = iota
	cc
)

type ofxTokenType struct {
	Str   string // name or contents, depending on the State value
	State int
}

type ofxCharType struct {
	Ch    byte
	State int
}

const KB = 1024
const MB = KB * KB
const ofxext = ".OFX"
const qfxext = ".QFX"
const sqliteoutfile = "citifile.csv"
const accessoutfile = "citifile.txt"

type generalHeaderType struct {
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

type generalTransactionType struct {
	TRNTYPE     string
	DTPOSTEDtxt string // intended for Excel or Access
	DTPOSTEDcsv string // intended for SQLite
	TRNAMT      string
	TRNAMTfloat float64
	FITID       string
	CHECKNUM    string
	CHECKNUMint int
	NAME        string
	MEMO        string
	Descript    string
	Juldate     int
}

type generalFooterType struct {
	BalAmt      string
	BalAmtFloat float64
	DTasof      string
}

var Transactions []generalTransactionType
var inputstate int
var bankTranListEnd bool
var EOF bool
var unToken ofxTokenType
var unTokenFlag bool

func main() {
	var e error
	var filebyteslice []byte
	var BaseFilename, ans, InFilename string
	InFileExists := false

	fmt.Println(" testfromfx2.go lastModified is", lastModified)
	if len(os.Args) <= 1 {
		filenames := filepicker.GetRegexFilenames("(ofx$)|(qfx$)|(qbo$)") // $ matches end of line
		for i := 0; i < min(len(filenames), 30); i++ {
			fmt.Println("filename[", i, "] is", filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code=999) : ")
		fmt.Scanln(&ans)
		if len(ans) == 0 {
			ans = "0"
		} else if ans == "999" {
			fmt.Println(" Stop code entered.")
			os.Exit(0)
		}

		i, err := strconv.Atoi(ans)
		if err == nil {
			InFilename = filenames[i]
		} else { // allow entering 'a' .. 'z' for 0 to 25.  However, it seems I never use it.
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			InFilename = filenames[i]
		}
		fmt.Println(" Picked filename is", InFilename)
		BaseFilename = InFilename
	} else {
		inbuf := getcommandline.GetCommandLineString()
		BaseFilename = filepath.Clean(inbuf)

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
			fmt.Println(" File ", BaseFilename, BaseFilename+qfxext, BaseFilename+ofxext, " or ", InFilename, " do not exist.  Exiting.")
			os.Exit(1)
		}
		fmt.Println(" input filename is ", InFilename)
	}

	if !strings.HasPrefix(BaseFilename, "CHK") { // remember that inputstate starts as 0, so it starts as citichecking.
		inputstate = cc
	}

	fmt.Println()

	filebyteslice = make([]byte, 0, MB) // 1 MB as initial capacity.
	filebyteslice, e = ioutil.ReadFile(InFilename)
	if e != nil {
		fmt.Println(" Error from ReadFile is ", e)
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(filebyteslice)

	const ( // intended for ofxTokenType
		empty = iota
		strng
		openinghtml
		closinghtml
		// othererror  // unused at the moment, but I'll just comment it out.
	)

	for !EOF {
		token := getOfxToken(bytesbuffer)
		fmt.Printf(" TokenType : %s, Token : %q   ", translateTokenType(token), token.Str) // intentional no newline here.
		if EOF {
			break
		}
		fmt.Print(" call UnGetTkn? (y,q,e,N) ")
		fmt.Scanln(&ans) // ignore the error as <enter> by itself is an error of "unexpected newline"

		ans = strings.TrimSpace(ans)
		ans = strings.ToUpper(ans)
		if strings.HasPrefix(ans, "Y") {
			ungetToken()
		} else if ans == "Q" || ans == "QUIT" || ans == "E" || ans == "EXIT" {
			os.Exit(0)
		}
	}
} // end main for test routine.

//---------------------------------------------------------------------------------------------------
func translateTokenType(t ofxTokenType) string {
	switch t.State {
	case empty:
		return "empty      "
	case strng:
		return "strng      "
	case openinghtml:
		return "OpeningHTML"
	case closinghtml:
		return "ClosingHTML"
	default:
		return "unknown value"
	}
}

//---------------------------------------------------------------------------------------------------

func DateFieldReformatAccess(datein string) (string, int) {
	//                                                                0123456789     01234567
	//  This procedure changes the date as it is input in a qfx file: mm/dd/yyyy <-- YYYYMMDD

	var dateout string

	datebyteslice := make([]byte, 10)
	datebyteslice[0] = datein[4]
	datebyteslice[1] = datein[5]
	datebyteslice[2] = '/'
	datebyteslice[3] = datein[6]
	datebyteslice[4] = datein[7]
	datebyteslice[5] = '/'
	datebyteslice[6] = datein[0]
	datebyteslice[7] = datein[1]
	datebyteslice[8] = datein[2]
	datebyteslice[9] = datein[3]
	dateout = string(datebyteslice)
	m, _ := strconv.Atoi(datein[4:6])
	d, _ := strconv.Atoi(datein[6:8])
	y, _ := strconv.Atoi(datein[0:4])
	juldate := timlibg.JULIAN(m, d, y)
	return dateout, juldate
} // END DateFieldReformatAccess;

func DateFieldAccessToSQlite(datein string) string {
	//                                   0123456789     0123456789
	//  This procedure changes the date: MM/DD/YYYY --> YYYY-MM-DD
	// Written after I learned that Access won't handle the YYYY-MM-DD format, so now I have to
	// write out 2 files.  First I'll write the Access file, then reprocess the date fields to
	// write out the Sqlite format.  The Juldate doesn't change, but I don't think I use it
	// anyway.

	var dateout string

	datebyteslice := make([]byte, 10)
	datebyteslice[0] = datein[6]
	datebyteslice[1] = datein[7]
	datebyteslice[2] = datein[8]
	datebyteslice[3] = datein[9]
	datebyteslice[4] = '-'
	datebyteslice[5] = datein[0]
	datebyteslice[6] = datein[1]
	datebyteslice[7] = '-'
	datebyteslice[8] = datein[3]
	datebyteslice[9] = datein[4]
	dateout = string(datebyteslice)
	return dateout
} // END DateFieldAccessToSQlite

//--------------------------------------------------------------------------------------------------
func ungetToken() {
	unTokenFlag = true
}

//--------------------------------------------------------------------------------------------------
func getOfxToken(buf *bytes.Buffer) ofxTokenType {
	// -------------------------------------------------- GetQfxToken ----------------------------------
	// Delimiters are angle brackets and EOL.
	//   I forgot that break applies to switch-case as well as for loop.  I had to be more specific for this to work.
	// Now allows unget a token, so it has to check this first.  So the allowEmptyToken is not needed, and was removed.

	var token ofxTokenType
	var char ofxCharType

	if unTokenFlag {
		unTokenFlag = false
		return unToken
	}

MainProcessingLoop:
	for { // main processing loop

		// GetChar
		r, size, err := buf.ReadRune()
		if err != nil { // this includes the EOF condition among others
			EOF = true
			break
		}
		for size > 1 { // discard non-ASCII runes
			r, size, err = buf.ReadRune()
			if err != nil { // this includes the EOF condition
				//      i noticed that FITID last 4 digits, from positions 9..12, or [9:13] are a sequence number for that dayonly
				//	  And name = Bill Payment is when I have to extract the number > 12000 at the end for the CHECKNUM field
				//	  I'll have to code this later.
				//
				EOF = true
				break MainProcessingLoop
			}
		}

		char.Ch = byte(r)

		// assign charstate
		switch char.Ch {
		case '\n', '\r', '\t':
			char.State = eol
		case '<':
			char.State = openangle
		case '>':
			char.State = closeangle
		case '/':
			char.State = slash
		default:
			char.State = plain
		} // END switch case on ch

		switch token.State {
		case empty: // of token.State, which is the initial state.
			switch char.State {
			case plain, slash:
				token.State = strng
				token.Str = string(char.Ch)
			case openangle:
				token.State = openinghtml
			case eol:
				// ignore newlines and just get next token on the next line
				break MainProcessingLoop

			case closeangle:
				fmt.Println(" In getOfxToken.  Empty token got closeangle char")
			} // END case chstate is empty

		case strng: // of token.State
			switch char.State {
			case plain, slash:
				token.Str = token.Str + string(char.Ch)
			case eol:
				break MainProcessingLoop
			case openangle: // openangle char will be avail for next loop iteration
				_ = buf.UnreadRune()
				break MainProcessingLoop
			case closeangle:
				fmt.Println(" In getOfxToken.  String token got closeangle ch")
			} // END case chtkn.State in ofxtkn.Str of strng
		case openinghtml: // of token.State
			switch char.State {
			case plain, openangle:
				token.Str = token.Str + string(char.Ch)
			case slash:
				if len(token.Str) == 0 {
					token.State = closinghtml
				} else {
					token.Str = token.Str + string(char.Ch)
				} // END;
			case closeangle, eol:
				break MainProcessingLoop
			} // END case chtkn.State in openinghtml
		case closinghtml: // of token.State
			switch char.State {
			case plain, slash, openangle:
				token.Str = token.Str + string(char.Ch)
			case closeangle, eol:
				break MainProcessingLoop
			} //      END; (* case chstate in closinghtml *)
		default: // ofxtkn.State is othererror
			fmt.Println(" In GetQfxToken and tokenstate is othererror.")
		} // END case ofxtknstate
	} // END ofxtkn.State processing loop
	unToken = token
	return token
} // END getOfxToken;

// ---------------------------------------------------- getTransactionData --------------------------
func getTransactionData(buf *bytes.Buffer) generalTransactionType {
	// Returns nil as a sign of either normal or abnormal end.

	var OFXtoken ofxTokenType
	var transaction generalTransactionType

	for { // processing loop
		OFXtoken = getOfxToken(buf) // get opening tag name, ie, <tagname>
		//fmt.Println(" in gettransactiondata.  OFXtoken is", OFXtoken)
		if EOF {
			fmt.Println(" Trying to get transaction record and got unexpected EOF condition.")
			break // will return an empty transaction
		}

		if false {
			// do nothing but it allows the rest of the conditions to be in the else if form

		} else if OFXtoken.State == openinghtml && OFXtoken.Str == "TRNTYPE" {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				ungetToken()
				//	fmt.Println(" after TRNTYPE got unexpedted token:", OFXtoken)
				//	break
			}
			transaction.TRNTYPE = OFXtoken.Str

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "DTPOSTED") {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				fmt.Println(" after DTPOSTED got token:", OFXtoken)
				ungetToken()
				break
			}
			transaction.DTPOSTEDtxt, transaction.Juldate = DateFieldReformatAccess(OFXtoken.Str)
			transaction.DTPOSTEDcsv = DateFieldAccessToSQlite(transaction.DTPOSTEDtxt)
		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "TRNAMT") {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				fmt.Println(" after TRNAMT got unexpedted token:", OFXtoken)
				ungetToken()
				break
			} // if EOF or token state not a string
			transaction.TRNAMT = OFXtoken.Str
			transaction.TRNAMTfloat, _ = strconv.ParseFloat(OFXtoken.Str, 64)

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "FITID") {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				ungetToken()
				//	fmt.Println(" after FITID got unexpedted token:", OFXtoken)
				//	break
			}
			transaction.FITID = OFXtoken.Str

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "CHECKNUM") {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				ungetToken()
				//	fmt.Println(" after CHECKNUM got unexpedted token:", OFXtoken)
				//	break
			}
			transaction.CHECKNUM = OFXtoken.Str
			transaction.CHECKNUMint, _ = strconv.Atoi(OFXtoken.Str)

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "NAME") {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				fmt.Println(" after NAME got unexpedted token:", OFXtoken)
				ungetToken()
				break
			}
			if strings.ContainsRune(OFXtoken.Str, '&') {
				OFXtoken.Str = strings.ReplaceAll(OFXtoken.Str, "amp;", "")
			}
			transaction.NAME = OFXtoken.Str

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "MEMO") {
			OFXtoken = getOfxToken(buf) // tag contents must be on same line as tagname
			if OFXtoken.State != strng {
				ungetToken()
				//	fmt.Println(" after MEMO got unexpedted token:", OFXtoken)
				//	break
			}
			if strings.ContainsRune(OFXtoken.Str, '&') {
				OFXtoken.Str = strings.ReplaceAll(OFXtoken.Str, "amp;", "")
			}
			transaction.MEMO = OFXtoken.Str

		} else if (OFXtoken.State == closinghtml) && (OFXtoken.Str == "STMTTRN") {
			break

		} else if (OFXtoken.State == closinghtml) && (OFXtoken.Str == "BANKTRANLIST") {
			bankTranListEnd = true
			break
		} // END if OFXoken.State condition
	} // END processing loop for record contents
	return transaction
} // END getTransactionData

//--------------------------------------------------------------------------------------------
func processOFXFile(buf *bytes.Buffer) (generalHeaderType, generalFooterType) {
	// transactions slice is passed as a global

	var header generalHeaderType
	var token ofxTokenType
	var transaction generalTransactionType
	var footer generalFooterType

	for { // loop to read the header
		token = getOfxToken(buf) // get tagname, as in <tagname>
		if EOF {
			fmt.Println(" Trying to get header info and got EOF condition.")
			break
		}

		if (token.State == openinghtml) && (token.Str == "ORG") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get header ORG and got error.  Token is", token)
				//	break
			}
			header.ORG = token.Str

		} else if (token.State == openinghtml) && (token.Str == "ACCTID") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get ACCTID header and got error.  Token is", token)
				//	break
			}
			header.ACCTID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTSERVER") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get DTSERVER header and got error.  Token is", token)
				//	break
			}
			header.DTSERVER = token.Str

		} else if (token.State == openinghtml) && (token.Str == "LANGUAGE") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get LANGUAGE header and got error.  Token is", token)
				//	break
			}
			header.LANGUAGE = token.Str

		} else if (token.State == openinghtml) && (token.Str == "FID") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get FID header and got error.  Token is", token)
				//	break
			}
			header.FID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "CURDEF") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get CURDEF header and got error.  Token is", token)
				//	break
			}
			header.CURDEF = token.Str

		} else if (token.State == openinghtml) && (token.Str == "BANKID") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get BANKID header and got error.  Token is", token)
				//	break
			}
			header.BANKID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "ACCTTYPE") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get ACCTTYPE header and got error. Token is", token)
				//	break
			}
			header.ACCTTYPE = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTSTART") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get DTSTART header and got error. Token is", token)
				//	break
			}
			header.DTSTART = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTEND") {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				//	fmt.Println(" Trying to get DTEND header and got error condition.  Token is", token,
				//		", header is", header)
				//	break
			}
			header.DTEND = token.Str

		} else if (token.State == openinghtml) && (token.Str == "STMTTRN") { // header finished, transactions will follow.
			break

		} // END if token.State AND token.Str
	} // END loop for header info

	for { // LOOP to read multiple transactions
		transaction = getTransactionData(buf)

		if bankTranListEnd {
			break // either at EOF or there was an error from getTransactionData
		}

		Transactions = append(Transactions, transaction)

	} // END LOOP for multiple transactions

	if EOF {
		fmt.Println(" Unexpected EOF.")
		os.Exit(1)
	}

	//  Get Footer containing ledgerbal, balamt, dtasof.  Stop when come TO </OFX>

	for { // loop to get the footer.   exit out of this loop at EOF or came to </OFX>
		token = getOfxToken(buf) // get tagname, as in <tagname>
		if EOF {
			fmt.Println(" Trying to get footer info and got EOF condition.")
			return header, footer
		}

		if false {
			// do nothing
		} else if token.State == openinghtml && token.Str == "BALAMT" {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				continue
				//	fmt.Println(" Trying to get footer BALAMT and got an error.")
				//	break
			}
			footer.BalAmt = token.Str
			var err error
			footer.BalAmtFloat, err = strconv.ParseFloat(footer.BalAmt, 64)
			if err != nil {
				fmt.Println(" error converting string footer.BalAmt to float:", err)
			}
			if footer.BalAmtFloat > 9999 {
				footer.BalAmt = AddCommas(footer.BalAmt)
			}

		} else if token.State == openinghtml && token.Str == "DTASOF" {
			token = getOfxToken(buf) // tag contents must be on same line as tagname
			if token.State != strng {
				ungetToken()
				fmt.Println(" Trying to get footer DTASOF and got an error.")
				continue
				//break
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

	return header, footer
} // END processOFXFile

//--------------------------------------------------------------------------------------------
func Pause() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing.  Hit <enter> to continue  ")
	scanner.Scan()
	_ = scanner.Text()
}

//------------------------------------------------------------

func ExtractNumberFromString(s string) (string, int) {
	var chknum string
	var chknumint int

	tokenize.INITKN(s)
	for {
		token, EOL := tokenize.GETTKN()
		if EOL {
			return "", 0
		}
		if token.State == tokenize.DGT && token.Isum > 12000 && token.Isum < 30000 {
			chknum = token.Str
			chknumint = token.Isum
			return chknum, chknumint
		}
	}
} // end ExtractNumberFromString

//-------------------------------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

//-------------------------------------------------------
func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var Comma = []byte{','}

	BS := make([]byte, 0, 15)
	BS = append(BS, instr...)

	i := len(BS) - 3 // account for a decimal point and 2 decimal digits.

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//---------------------------------------------------------------------------------------------------
// END testfromfx2.go based on fromfx2.go based on fromfx.go based on ofx2csv.go based on qfx2xls.mod
