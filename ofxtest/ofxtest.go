package main

import (
	"bufio"
	"bytes"
	"fmt"
	"getcommandline"
	"io/ioutil"
	"os"
	"path/filepath"
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
const ofxext = ".OFX"
const qfxext = ".QFX"

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
	TRNTYPE   string
	DTPOSTED  string
	TRNAMTstr string
	TRNAMT    float64
	FITID     string
	CHECKNUM  int
	NAME      string
	MEMO      string
	Juldate   int
}

type citifootertype struct {
	BalAmt string
	DTasof string
}

var Transactions []citiTransactionType

func main() {

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
		fmt.Println(" File ", BaseFilename, BaseFilename+qfxext, BaseFilename+ofxext, " or ", InFilename, " do not exist.  Exiting.")
		os.Exit(1)
	}
	/* Can't get this to work.  I don't really need it, anyway.
	toascii := func(s string) { // anonymous function, also called a closure.
		cmd := exec.Command("cmd", "/c", "toascii", s)
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
	*/

	fmt.Println(" input filename is ", InFilename)
	filebyteslice = make([]byte, 0, MB) // 1 MB as initial capacity.
	filebyteslice, e = ioutil.ReadFile(InFilename)
	if e != nil {
		fmt.Println(" Error from ReadFile is ", e)
		os.Exit(1)
	}

	//	fmt.Println(" Length of filebyteslice is ", len(filebyteslice))

	bytesbuffer := bytes.NewBuffer(filebyteslice)

	//	fmt.Println(" Length of bytesbuffer is ", bytesbuffer.Len())

	// This code started as qfx2xls.mod, but I really want more like CitiFilterQIF.mod.  So I have to merge
	// in that code also.
	// And I need to use toascii in some way or another, either an exec function, or copying the
	// code here.  toascii deletes non UTF-8 code points, utf8toascii does not do this.

	Transactions = make([]citiTransactionType, 0, 200)

	header, footer := ProcessOFXFile(bytesbuffer)

	// now I have a header, footer, and a slice of all the individual transactions.  At this
	// point, I'll just display them, and pause in between.

	fmt.Println(" Header is ", header, ",  footer is ", footer, ", and number of transactions is ", len(Transactions))

	for cnt, t := range Transactions {
		fmt.Printf(" TRNTYPE=%s, DTPOSTED=%s,TRNAMTstr=%s,TRNAMT=%g,FITID=%s,CHECKNUM=%d,Name=%s,Memo=%s,Juldate=%d",
			t.TRNTYPE, t.DTPOSTED, t.TRNAMTstr, t.TRNAMT, t.FITID, t.CHECKNUM, t.NAME, t.MEMO, t.Juldate)
		if cnt%20 == 0 {
			Pause()
		}
	}

} // end main of this package

//---------------------------------------------------------------------------------------------------
func DateFieldReformat(datein string) (string, int) {
	//                                                                    01234567    01234567
	//  This procedure changes the date as it is input in a qfx file from yyyymmdd -> mm/dd/yy.
	// I have to look into if I want a 2 or 4 digit year

	var dateout string

	datebyteslice := make([]byte, 8)
	datebyteslice[0] = datein[4]
	datebyteslice[1] = datein[5]
	datebyteslice[2] = '/'
	datebyteslice[3] = datein[6]
	datebyteslice[4] = datein[7]
	datebyteslice[5] = '/'
	datebyteslice[6] = datein[2]
	datebyteslice[7] = datein[3]
	dateout = string(datebyteslice)
	m, _ := strconv.Atoi(datein[4:5])
	d, _ := strconv.Atoi(datein[6:7])
	y, err := strconv.Atoi(datein[2:3])
	juldate := timlibg.JULIAN(m, d, y)
	fmt.Println(" in DateFieldReformat datein=", datein, " m=", m, ",d=", d, ",y=", y, "err=", err, "juldate=", juldate, "dateout=", dateout)
	return dateout, juldate

} // END DateFieldReformat;

//--------------------------------------------------------------------------------------------------
func GetOfxToken(buf *bytes.Buffer) ofxTokenType {
	// -------------------------------------------------- GetQfxToken ----------------------------------
	// Delimiters are angle brackets and EOL.
	//   I forgot that break applies to switch-case as well as for.  I had to be more
	//   specific for this to work.

	var token ofxTokenType
	var char ofxCharType

MainProcessingLoop:
	for { // main processing loop

		// GetChar
		r, size, err := buf.ReadRune()
		if err != nil { // this includes the EOF condition
			break
		}
		for size > 1 { // discard non-ASCII runes
			r, size, err = buf.ReadRune()
			if err != nil { // this includes the EOF condition
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
				break MainProcessingLoop
			case openangle: // openangle char will be avail for next loop iteration
				_ = buf.UnreadRune()
				break MainProcessingLoop
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
	return token
} // END GetOfxToken;

// ---------------------------------------------------- GetTransactionData --------------------------
func GetTransactionData(buf *bytes.Buffer) citiTransactionType {
	// Returns nil as a sign of either normal or abnormal end.

	var OFXtoken ofxTokenType
	var transaction citiTransactionType

	for { // processing loop
		OFXtoken = GetOfxToken(buf)
		if OFXtoken.State == empty {
			fmt.Println(" Trying to get transaction record and got unexpected EOF condition.")
			break
		}

		if false {
			// do nothing but it allows the rest of the conditions to be in the else if form

		} else if OFXtoken.State == openinghtml && OFXtoken.Str == "TRNTYPE" {
			OFXtoken = GetOfxToken(buf)
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.TRNTYPE = OFXtoken.Str

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "DTPOSTED") {
			OFXtoken = GetOfxToken(buf) // Now need the string data of this token
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.DTPOSTED, transaction.Juldate = DateFieldReformat(OFXtoken.Str)

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "TRNAMT") {
			OFXtoken = GetOfxToken(buf)
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.TRNAMTstr = OFXtoken.Str
			transaction.TRNAMT, _ = strconv.ParseFloat(OFXtoken.Str, 64)

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "FITID") {
			OFXtoken = GetOfxToken(buf)
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.FITID = OFXtoken.Str

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "CHECKNUM") {
			OFXtoken = GetOfxToken(buf)
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.CHECKNUM, _ = strconv.Atoi(OFXtoken.Str)

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "NAME") {
			OFXtoken = GetOfxToken(buf)
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.NAME = OFXtoken.Str

		} else if (OFXtoken.State == openinghtml) && (OFXtoken.Str == "MEMO") {
			OFXtoken = GetOfxToken(buf)
			if OFXtoken.State == empty || (OFXtoken.State != strng) {
				fmt.Println(" after get ofxtoken, got unexpedted EOF or token is not a string.")
				break
			} // if EOF or token state not a string
			transaction.MEMO = OFXtoken.Str

		} else if (OFXtoken.State == closinghtml) && (OFXtoken.Str == "STMTTRN") {
			break

		} else if (OFXtoken.State == closinghtml) && (OFXtoken.Str == "BANKTRANLIST") { // should not happen
			break
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
		token = GetOfxToken(buf)
		if token.State == empty {
			fmt.Println(" Trying to get header info and got EOF condition.")
			break
		}

		if (token.State == openinghtml) && (token.Str == "ORG") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.ORG = token.Str

		} else if (token.State == openinghtml) && (token.Str == "ACCTID") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.ACCTID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTSERVER") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.DTSERVER = token.Str

		} else if (token.State == openinghtml) && (token.Str == "LANGUAGE") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.LANGUAGE = token.Str

		} else if (token.State == openinghtml) && (token.Str == "FID") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.FID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "CURDEF") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.CURDEF = token.Str

		} else if (token.State == openinghtml) && (token.Str == "BANKID") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a strng.")
				break
			}
			header.BANKID = token.Str

		} else if (token.State == openinghtml) && (token.Str == "ACCTTYPE") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a strng.")
				break
			}
			header.ACCTTYPE = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTSTART") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a strng.")
				break
			}
			header.DTSTART = token.Str

		} else if (token.State == openinghtml) && (token.Str == "DTEND") {
			token = GetOfxToken(buf)
			if token.State == empty || (token.State != strng) {
				fmt.Println(" Trying to get header info and got EOF condition or token is not a string.")
				break
			}
			header.DTEND = token.Str

		} else if (token.State == openinghtml) && (token.Str == "STMTTRN") {
			break

		} // END if token.State AND token.Str
	} // END loop for header info

	for { // LOOP to read multiple transactions
		transaction = GetTransactionData(buf)

		if transaction.TRNTYPE == "" && transaction.DTPOSTED == "" {
			break // either at EOF or there was an error from GetTransactionData
		}

		Transactions = append(Transactions, transaction)

	} // END LOOP for multiple transactions

	//  Get Footer containing ledgerbal, balamt, dtasof.  Stop when come TO </OFX>

	for { // loop to get the footer.   exit out of this loop at EOF or came to </OFX>
		token = GetOfxToken(buf)
		if token.State == empty {
			fmt.Println(" Trying to get footer info and got EOF condition.")
			return header, footer
		}

		if false {
			// do nothing
		} else if token.State == openinghtml && token.Str == "BALAMT" {
			token = GetOfxToken(buf)
			if token.State == empty {
				fmt.Println(" Trying to get footer info and got an error.")
				break
			}
			footer.BalAmt = token.Str

		} else if token.State == openinghtml && token.Str == "DTASOF" {
			token = GetOfxToken(buf)
			if token.State == empty {
				fmt.Println(" Trying to get footer info and got an error.")
				break
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

func Pause() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing.  Hit <enter> to continue  ")
	scanner.Scan()
	_ = scanner.Text()
}

// END ofx2csv.go based on qfx2xls.mod
