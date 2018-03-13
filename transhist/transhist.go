// TransHist, using qfx2xls.mod, dateconvert.go and  Allcc as templates.
package main

import (
	"bufio"
	"encoding/csv"
	"filepicker"
	"fmt"
	"getcommandline"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tknptr"
	"tokenize"
	"unicode"
)

const lastModified = "12 Mar 2018"

/*
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
  19 Oct 17 -- Added filepicker code
   1 Nov 17 -- Added output of $ for footer amount.
  24 Dec 17 -- Now called dateconvert, meant to read csv from sqlite and change format to ISO8601 YYYY-MM-DD
               Fields are date,amount,descr,comment, like in allcc-sqlite.db.
  25 Dec 17 -- Now will do same thing for Allcc-Sqlite.db.  Too bad the fields are in a different order.
  27 Dec 17 -- Added automatic removal of blank lines, and fixed 00 problem.
  31 Dec 17 -- Fixed a panic when testing record[4].  There is no record[4].  What was I thinking???
               And I broadened the ExtractAmtFromString logic.
   5 Jan 18 -- Will have it change the output format of the date field to the opposite of the input file.
   6 Jan 18 -- Expanded ReformatToISO8601date to accept either 2 or 4 digit year in input.
                 And will use tknptr instead of tokenize, for variety in ReformatToStdDate
   2 Feb 18 -- Fixed formatting bug to ISO8601 format, in which January becomes 001 instead of 01.
  12 Mar 18 -- Now called transhist.go, to convert HSBC csv file to a usable format.
               TransHist = date not in ISO8601 format, description and amount.
			   Write out to date in ISO8601 format, amount, description and comment.  But there is
			   white space before the date that I want to remove.
*/

type Row struct {
	date, amount, descr, comment string
}

const CSVext = ".CSV"
const csvext = ".csv"

func main() {
	var e error
	var BaseFilename, ans, InFilename string
	var row Row

	rows := make([]Row, 0, 1000)
	outputstringslice := make([]string, 4)
	InFileExists := false

	fmt.Println(" transhist.go (date,amount,descr,comment) lastModified is", lastModified)
	if len(os.Args) <= 1 {
		filenames := filepicker.GetFilenames("*" + csvext)
		for i := 0; i < min(len(filenames), 10); i++ {
			fmt.Println("filename[", i, "] is", filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		fmt.Scanln(&ans)
		if len(ans) == 0 {
			ans = "0"
		}
		i, err := strconv.Atoi(ans)
		if err == nil {
			InFilename = filenames[i]
		} else {
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
			InFilename = BaseFilename + csvext
			_, err := os.Stat(InFilename)
			if err == nil {
				InFileExists = true
			} else {
				InFilename = BaseFilename + CSVext
				_, err := os.Stat(InFilename)
				if err == nil {
					InFileExists = true
				}
			}
		}

		if !InFileExists {
			fmt.Println(" File ", BaseFilename, BaseFilename+csvext, BaseFilename+CSVext, " or ", InFilename, " do not exist.  Exiting.")
			os.Exit(1)
		}
		fmt.Println(" input filename is ", InFilename)

	}

	fmt.Println()

	InputFile, err := os.Open(InFilename)
	if err != nil {
		fmt.Println(" Error while opening ", InFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer InputFile.Close()

	OutFilename := BaseFilename + "-converted" + csvext
	OutputFile, err := os.Create(OutFilename)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()

	// Process input file line by line.

	rdr := csv.NewReader(InputFile)

	for {
		record, err := rdr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else if len(record[0]) < 3 { // likely an empty field with just "", like a Header line.
			continue
		} else if unicode.IsLetter(rune(record[0][len(record[0])-1])) { // Header line.
			continue
		}
		datestring := strings.TrimSpace(record[0])
		if IsNotDate(datestring) {
			fmt.Print(" First field is not formated like a date. Continue? ")
			if Stop() {
				os.Exit(1)
			}
		}
		if IsISO8601(datestring) {
			row.date = ReformatToStdDate(datestring)
		} else {
			row.date = ReformatToISO8601date(datestring)
		}
		row.amount = ExtractAmtFromString(record[2])
		row.descr = strings.TrimSpace(record[1])
		row.comment = " "

		rows = append(rows, row)
	}

	InputFile.Close()

	// Output to file section

	wrtr := csv.NewWriter(OutputFile)
	defer wrtr.Flush()

	for ctr, r := range rows {
		outputstringslice[0] = r.date
		outputstringslice[1] = r.amount
		outputstringslice[2] = r.descr
		outputstringslice[3] = r.comment
		fmt.Printf(" %d: %q,%q,%q,%q \n", ctr, outputstringslice[0], outputstringslice[1], outputstringslice[2], outputstringslice[3])
		if e = wrtr.Write(outputstringslice); e != nil {
			log.Fatalln(" Error writing record to csv:", e)
		}
		if ctr%40 == 0 && ctr > 0 && ctr < 100 { // these files can have 6000 records to output.
			Pause()
		}
	}

	wrtr.Flush()
	if err := wrtr.Error(); err != nil {
		log.Fatal(err)
	}
	OutputFile.Close()

} // end main of this package

//---------------------------------------------------------------------------------------------------

func ReformatToISO8601date(in string) string {
	//                                          formerly  func ExtractDateFromString(in string) string {
	// need to construct YYYY-MM-DD from MM/DD/YYYY or MM/DD/YY
	var mstr, dstr, ystr string

	tokenize.INITKN(in)
	token, EOL := tokenize.GETTKN() // get MM
	if EOL || token.State != tokenize.DGT {
		return ""
	}
	L := len(token.Str)
	if L >= 2 {
		mstr = token.Str
	} else {
		mstr = "0" + token.Str
	}
	token, EOL = tokenize.GETTKN() // discard the / which is state of opcode.
	token, EOL = tokenize.GETTKN() // get DD
	if EOL || token.State != tokenize.DGT {
		return ""
	}

	L = len(token.Str)
	if L >= 2 {
		dstr = token.Str
	} else {
		dstr = "0" + token.Str
	}
	token, EOL = tokenize.GETTKN() // discard the / which is state of opcode.
	token, EOL = tokenize.GETTKN() // get YY or YYYY
	if EOL || token.State != tokenize.DGT {
		return ""
	}
	y := token.Isum
	if y < 100 {
		y += 2000
	}
	ystr = strconv.Itoa(y)

	out := ystr + "-" + mstr + "-" + dstr
	return out
} // end ReformatToISO8601date, formerly ExtractDateFromString
//-------------------------------------------------------
func ReformatToStdDate(in string) string {
	// need to construct MM/DD/YYYY from YYYY-MM-DD
	var mstr, dstr, ystr string

	tknP := tknptr.INITKN(in)
	token, EOL := tknP.GETTKN() // get YYYY
	if EOL || token.State != tknptr.DGT {
		return ""
	}
	y := token.Isum
	if y < 100 {
		y += 2000
	}
	ystr = strconv.Itoa(y)

	//	token, EOL = tknP.GETTKN() // discard the - which is state of opcode.
	token, EOL = tknP.GETTKN() // get MM, which is negative because of the dash as a minus sign effect.
	if EOL || token.State != tknptr.DGT {
		return ystr
	}
	m := token.Isum
	if m < 0 {
		m = -m
	}
	if m > 9 {
		mstr = token.Str
	} else {
		mstr = "0" + strconv.Itoa(m)
	}

	//  token, EOL = tknP.GETTKN() // discard the - which is state of opcode.
	token, EOL = tknP.GETTKN() // get DD, which is negative because of the dash as a minus sign effect
	if EOL || token.State != tknptr.DGT {
		return ystr + "-" + mstr
	}
	L := len(token.Str)
	if L >= 2 {
		dstr = token.Str // the leading "-" won't be in the Str field.
	} else {
		dstr = "0" + token.Str
	}

	out := mstr + "/" + dstr + "/" + ystr
	return out
} // end ReformatToStdDate
//-------------------------------------------------------
func ExtractAmtFromString(in string) string {
	// Need to make ($###.##) format to -###.##.  If already -###.##, leave alone.

	bs := []byte(in) // convert to a byte slice
	byteslice := make([]byte, 0, 20)

	if bs[0] == '(' && bs[1] == '$' {
		bs[0] = ' '
		bs[1] = '-'
	} else if bs[0] == '$' {
		bs[0] = ' '
	}

	for i := range bs {
		if bs[i] == ')' || bs[i] == '(' || bs[i] == '$' {
			bs[i] = ' '
		}
	}

	// remove blanks by copying over only non-blanks

	for i := range bs {
		if bs[i] != ' ' {
			byteslice = append(byteslice, bs[i])
		}
	}

	out := string(byteslice)
	return out
} // end ExtractAmtFromString

//-------------------------------------------------------
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//-------------------------------------------------------
func Pause() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Pausing.  Hit <enter> to continue  ")
	scanner.Scan()
	_ = scanner.Text()
}

//------------------------------------------------------------
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
	var Comma []byte = []byte{','}

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
func IsISO8601(instr string) bool {
	// Look for either a - or a / in the string.  That's it!
	isISOdate := strings.Contains(instr, "-")
	return isISOdate
}

//---------------------------------------------------------------------------------------------------
func IsNotDate(instr string) bool {
	NotAdate := !strings.ContainsAny(instr, "-/")
	return NotAdate
}

//---------------------------------------------------------------------------------------------------
func Stop() bool {
	var ans string
	_, _ = fmt.Scanln(&ans)
	ans = strings.ToUpper(ans)
	stopflag := false
	if len(ans) > 0 && strings.HasPrefix(ans, "N") {
		stopflag = true
	}
	return stopflag
}

//---------------------------------------------------------------------------------------------------
// END dateconvert based on ofx2csv.go based on qfx2xls.mod
