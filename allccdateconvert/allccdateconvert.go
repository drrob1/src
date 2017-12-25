// dateconvert.go using ofx2cvs as a template.
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
	"tokenize"
)

const lastModified = "25 Dec 2017"

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
  19 Oct 17 -- Added filepicker code
   1 Nov 17 -- Added output of $ for footer amount.
  24 Dec 17 -- Now called dateconvert, meant to read csv from sqlite and change format to ISO8601 YYYY-MM-DD HH:MM:SS.SSS
  25 Dec 17 -- Now will do same thing for Allcc-Sqlite.db.  Too bad the fields are in a different order.
*/

type Row struct {
	date, descr, amount, comment string
}

const CSVext = ".CSV"
const csvext = ".csv"

func main() {
	var e error
	var BaseFilename, ans, InFilename string
	var row Row

	rows := make([]Row, 0, 10000)
	outputstringslice := make([]string, 4)
	InFileExists := false

	fmt.Println(" dateconvert.go lastModified is", lastModified)
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
		}
		datestring := record[0]
		row.date = ExtractDateFromString(datestring)
		row.amount = ExtractAmtFromString(record[1])
		row.descr = record[2]
		row.comment = record[3]

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

func ExtractDateFromString(in string) string {
	// need to construct YYYY-MM-DD from MM/DD/YYYY
	var mstr, dstr, ystr string

	tokenize.INITKN(in)
	token, EOL := tokenize.GETTKN()
	if EOL || token.State != tokenize.DGT {
		return ""
	}
	m := token.Isum
	if m > 9 {
		mstr = token.Str
	} else {
		mstr = "0" + token.Str
	}
	token, EOL = tokenize.GETTKN() // discard the / which is state of opcode.
	token, EOL = tokenize.GETTKN()
	if EOL || token.State != tokenize.DGT {
		return ""
	}
	d := token.Isum
	if d > 9 {
		dstr = token.Str
	} else {
		dstr = "0" + token.Str
	}
	token, EOL = tokenize.GETTKN() // discard the / which is state of opcode.
	token, EOL = tokenize.GETTKN()
	if EOL || token.State != tokenize.DGT {
		return ""
	}
	//	y := token.Isum
	ystr = token.Str

	out := ystr + "-" + mstr + "-" + dstr
	return out
} // end ExtractDateFromString
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
		if bs[i] == ')' {
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
// END dateconvert based on ofx2csv.go based on qfx2xls.mod
