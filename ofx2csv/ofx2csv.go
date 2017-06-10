package main

import (
	"bytes"
	"fmt"
	"getcommandline"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"timlibg"
)

const lastModified = "10 Jun 17"

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
	othererror
)

type ofxTokenType struct {
	Str   string // name or contents, depending on the State value
	State int
}

type ofxCharType struct {
	Ch    byte
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
	Juldate  int
}

type citifootertype struct {
	BalAmt string
	DTasof string
}

var bsidx int // byte slice index for getting one ASCII character at a time.
var Transactions []citiTransactionType

func main() {

	var ofxToken ofxTokenType
	var ofxChar ofxCharType
	//	var juldate1, juldate2, juldate3 int   soon but not yet.

	var e error
	var filebyteslice []byte

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: ofx2csv <FileName.ext> where .ext = [.qfx|.ofx]")
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
		cmd := exec.Command("cmd", "/c", "toascii")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	if runtime.GOOS == "windows" {
		err = toascii(InFilename)
		if err != nil {
			fmt.Println(" Error from toascii ", err)
			os.Exit(1)
		}
	}

	filebyteslice = make([]byte, 0, MB) // 1 MB as initial capacity.
	filebyteslice, e = ioutil.ReadFile(InFilename)
	if e != nil {
		fmt.Println(" Error from ReadFile is ", e)
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(filebyteslice)

	// This code started as qfx2xls.mod, but I really want more like CitiFilterQIF.mod.  So I have to merge
	// in that code also.
	// And I need to use toascii in some way or another, either an exec function, or copying the
	// code here.  toascii deletes non UTF-8 code points, utf8toascii does not do this.

	Transactions = make([]citiTransactionType, 0, 200)

	header, footer := ProcessOFXFile(filebyteslice)

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

	datebyteslice[0] = datein[4]
	datebyteslice[1] = datein[5]
	datebyteslice[2] = '/'
	datebyteslice[3] = datein[6]
	datebyteslice[4] = datein[7]
	datebyteslice[5] = '/'
	datebyteslice[6] = datein[2]
	datebyteslice[7] = datein[3]
	dateout = string(datebytearray)
	m := strconv.Atoi(datein[4:5])
	d := strconv.Atoi(datein[6:7])
	y := strconv.Atoi(datein[2:3])
	juldate := timlibg.Julian(m, d, y)
	return dateout, juldate

} // END DateFieldReformat;

//--------------------------------------------------------------------------------------------------
func GetOfxToken(buf *bytes.Buffer) ofxTokenType {
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
		case empty: // of token.State
			switch char.State {
			case plain, slash:
				token.State = strng
				token.Str = string(char.Ch)
			case openangle:
				token.State = openinghtml
			case eol:
				// do nothing
			case closeangle:
				fmt.Println(" In GetOfxToken.  Empty token got closeangle char")
			} // END case chstate is empty

		case strng: // of token.State
			switch char.State {
			case plain, slash:
				token.Str = token.Str + string(char.Ch)
			case eol:
				break
			case openangle: // openangle char is still avail for next loop iteration
				bsidx--
				break
			case closeangle:
				fmt.Println(" In GetOfxToken.  String token got closeangle ch")
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
				break
			} // END case chtkn.State in openinghtml
		case closinghtml: // of token.State
			switch char.State {
			case plain, slash, openangle:
				token.Str = token.Str + string(char.Ch)
			case closeangle, eol:
				break
			} //      END; (* case chstate in closinghtml *)
		default: // ofxtkn.State is othererror
			fmt.Println(" In GetQfxToken and tokenstate is othererror.")
		} // END case ofxtknstate
	} // END ofxtkn.State processing loop

	return token
} // END GetOfxToken;

// ---------------------------------------------------- GetTransactionData --------------------------
func GetTransactionData(buf *bytes.Buffer) citiTransactionType {
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
			OFXtoken = GetOfxToken(bs) // Now need the string data of this token
			if OFXtoken == nil || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				return nil
			} // if EOF or token state not a string
			transaction.DTPOSTED, transaction.Juldate = DateFieldReformat(OFXtoken.Str)

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
func ProcessOFXFile(buf *bytes.Buffer) (citiheadertype, citifootertype) {

	var header citiheadertype
	var token ofxTokenType
	var transaction citiTransactionType
	var footer citifootertype

	for { // loop to read the header
		token = GetOFXToken(bs)
		if token == nil {
			fmt.Println(" Trying to get header info and got EOF condition.")
			return nil
		}

		if (token.State == openinghtml) && (token.Str == "ORG") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.ORG = token.Str

		} else if (token.State == openinghtml) && (token.Str == "ACCTID") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.ACCIT = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTSERVER") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.DTSERVER = token.Str

		} else if (token.State == openinghtml) && (token.Str == "LANGUAGE") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.LANGUAGE = token.Str

		} else if (token.State == openinghtml) && (token.Str == "FID") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.FID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "CURDEF") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.CURDEF = token.Str

		} else if (token.State == openinghtml) && (token.Str == "BANKID") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.BANKID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "ACCTTYPE") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.ACCTTYPE = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTSTART") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
			}
			header.DTSTART = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTEND") {
			token = GetOfxToken(bs)
			if token == nil || (token.State != string) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				return nil, nil
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

		Transactions = append(Transactions, transaction)

	} // END LOOP for multiple transactions

	//  Get Footer containing ledgerbal, balamt, dtasof.  Stop when come TO </OFX>

	for { // loop to get the footer.   exit out of this loop at EOF or came to </OFX>
		token = GetOfxToken(bs)
		if token == nil {
			fmt.Println(" Trying to get footer info and got EOF condition.")
			return header, footer
		}

		if false {
			// do nothing
		} else if token.State == openinghtml && token.Str == "BALAMT" {
			token = GetOfxToken(bs)
			if token == nil {
				fmt.Println(" Trying to get footer info and got an error.")
				return header, nil
			}
			footer.BalAmt = token.Str

		} else if token.State == openinghtml && token.Str == "DTASOF" {
			token = GetOfxToken(bs)
			if token == nil {
				fmt.Println(" Trying to get footer info and got an error.")
				return header, nil
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
} // END ProcessOFXFile

// END ofx2csv.go based on qfx2xls.mod
