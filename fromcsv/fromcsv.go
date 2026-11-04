// fromcsv.go from fromfx.go based on ofx2csv.go.
package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"src/filepicker"
	"src/timlibg"
	"strconv"
	"strings"
	"time"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/tealeg/xlsx/v3"

	_ "github.com/mattn/go-sqlite3"
)

/*
  REVISION HISTORY
  ----------------
  13 Mar 04 -- ofx2csf: It does not seem to be always creating the output file.
               It now copies the d/l file instead of renaming it.  And it echoes its output to the terminal.
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
   3 Mar 11 -- Noticed bug in GetQfxToken in that read calls should all be to the param f, not the
                global infile.  I will fix this next time I have to recompile.
------------------------------------------------------------------------------------------------------------------------------------------------------
   7 Jun 17 -- Converting to go.  I posted on go-nuts, and was told that the .qfx format is not XML, but ofx,
                which means open financial exchange (for/of information).  New name is ofx2csv.go
		I think I will first process the file using something like toascii.
	 This code started as qfx2xls.mod, but I really want more like CitiFilterQIF.mod.  So I have to merge in that code also.
	 And I need to use toascii because it deletes non UTF-8 code points, utf8toascii does not do this.

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
------------------------------------------------------------------------------------------------------------------------------------------------------
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
   2 Oct 20 -- qbo files will populate the filepicker menu.  Filepicker now uses case-insensitive flag.  Stop code added.
   4 Oct 20 -- Will ignore empty tokens
  17 Oct 20 -- Removed the strings.ToLower for output filenames.
   8 Jan 22 -- Converted to modules; it shows [a .. z] as well as [0 .. 26] as I allow letter input also, and I removed use of getcommandline.
                 Added verbose flag to control the display of pauses.  Removed use of ioutil that was deprecated as of Go 1.16
  21 Jan 24 -- Expanded stop code to include "," and "."
  16 Oct 24 -- Got the idea that this program place the title row for all .xls files, i.e., Date, Amt, Description, Comment, delimited as it has to be processed correctly.
               Will do this by adding a func instead of adding it to main.  It'll be easier to debug as a func.
  21 Oct 24 -- Since I forgot, I'll summarize how this works.  First it reads in all the transactions and parses them into a slice of generalTransactionType, then
                 it writes out this slice to 3 files: a csv to be read by sqlite3 database, a txt file named as xls for Excel,
                 and now .xlsx file directly for Excel using the GitHub code I recently learned about.
   5 Nov 24 -- Today's Election Day, but that's not important right now.
                 I got the idea to have this routine directly update the sqlite3 databases I use, i.e., allcc-sqlite.db and citibank.db.  I'd need a separate routine to update taxes.db.
                 This routine handles 2 different kinds of input files, the CHK_6861_Current_View.qfx, and one of the several credit card files in qfx or ofx format.
                 I have to make sure that the input state is consistent throughout the code.  Then send the slice of transactions to the rtn that will update the correct .db file.
					As far as I can tell, it worked on the first compile and run.  I'm going to change the database filenames to the real ones instead of the test ones.
  10 Nov 24 -- I'm going to port the new code I wrote in taxproc.go to more correctly write the Excel file.
               I added a field to generalTransactionType, added code to the getTransactionDate to populate the new DateTime field, and then added code to write it to the Excel file.
  11 Nov 24 -- I considered adding the sheet to Allcc.xlsx in code, but that doesn't work because the sheet gets added as the last one, and not the first one.
  12 Dec 24 -- Now that writing .xlsx works, I find that outputting .xls (really a txt file) is confusing me.  So I'll stop writing the files that aren't needed anymore.
				I'll use an output flag to write the xls and csv files, like most of the linux tools, I'll use -o.
				It looks like the output of the citifile and credit card files are intertwined.  The best way for me to organize this is to separate them into separate functions.
				That will make it much easier to read the code.
				Turns out that this is a major refactoring if I want it to still be able to write out all the files.  I have yet to test the citifile stuff.
                To define the output more completely:
					always: output for allcc the xlsx file and update the .db, and for citi always output the citifile.txt and update the .db
					if output flag set: also output for allcc the .csv and .xls (in txt format) files, and for citi output the citifile.csv.
					Need filenames for allcc.csv, allcc.xls, allcc.xlsx, and citifile.txt and citifile.csv.
					The correct formating of the date for Sqlite and Access or Excel is handled by the different fields in the transaction struct.
  14 Dec 24 -- The code is much easier to read.
  16 Dec 24 -- Ok, now that it works, I'm going to clean up the output section of main, so that it has one check of inputmode, instead of the spaghetti structure it used to be.
  10 Oct 26 -- Cleaned up comments and removed the min function that became built into the compiler as of Go 1.20 (Aug 2023) or so.  No need to recompile.
------------------------------------------------------------------------------------------------------------------------------------------------------
  30 Oct 26 -- Now called fromcsv.go.  WellsFargo has 2 years of data in csv format, but only 3 months of data in ofx format.  So I'm going to use csv for allcc and citi.
         		This will take a while because I'm also doing the 80-hour course.
				I'll use a map for the field names, starting at 1.  This leaves 0 for undefined input to be ignored.  Nevermind, csv format needs a simple loop since field position
				is what matters, not field name.
				"DATE", "DESCRIPTION", "AMOUNT", "CHECK #", "STATUS" are the fields in the Wells Fargo csv file.
   2 Nov 26 -- So far, processCSVfile is working, and writing to the SQL file is working.  I'll remove the code that parses the HTML tags.  That took a lot of code.
   3 Nov 26 -- Adding a start-up message that says the field names expected.
   4 Nov 26 -- Removed all the code that handled HTML tags.
*/

const lastModified = "3 Nov 26"

/*
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
)

type ofxTokenType struct {
	Str   string // name or contents, depending on the State value
	State int
}

type ofxCharType struct {
	Ch    byte
	State int
}

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

type generalFooterType struct {
	BalAmt      string
	BalAmtFloat float64
	DTasof      string
}

var bankTranListEnd bool
var EOF bool
var outputFlag = flag.Bool("o", false, "Set output mode, currently controls writing the .xls and .csv output files.")

const CSVext = ".csv"
const XLSext = ".xls"
*/

const ofxext = ".OFX"
const qfxext = ".QFX"
const xlsxext = ".xlsx"

// "DATE", "DESCRIPTION", "AMOUNT",    "CHECK #",     "STATUS" are the fields in the Wells Fargo csv file.
// Juldate  Descript       TRNAMT       CHECKNUM       TRNTYPE
// DateTime                TRNAMTfloat  CHECKNUMint
// DTPOSTEDtxt
// DTPOSTEDcsv
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
	DateTime    time.Time
}

// This uses the std library flag package
var verboseFlag = flag.Bool("v", false, "Set verbose mode.")

// SQliteDBname = "Allcc-Sqlite.db"
// SQliteDBname = "Allcc-test.db"
var SQliteDBname = "Allcc-Sqlite.db" // name of the database file, for use in the direct updating of the SQLite3 database files below.

func main() {
	var BaseFilename, ans, InFilename, XLoutFilename string

	flag.Parse()

	InFileExists := false

	fmt.Printf(" fromcsv.go lastModified is %s, expecting fields DATE, DESCRIPTION, AMOUNT, CHECK #, STATUS in the csv file.\n", lastModified)
	if flag.NArg() < 1 {
		filenames, err := filepicker.GetFilenames("*csv")
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from GetRegexFilenames is %v, exiting\n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s \n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code=999 . , / ;) : ")
		n, err := fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil || n == 0 { // these are redundant.  I'm playing now.
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == "/" || ans == ";" {
			fmt.Println(" Stop code entered.")
			os.Exit(0)
		}

		i, err := strconv.Atoi(ans)
		if err == nil {
			InFilename = filenames[i]
		} else { // allow entering 'a' .. 'z' for 0 to 25.
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			if i > 25 {
				fmt.Printf(" Index out of bounds.  It is %d.\n", i)
				return
			}
			InFilename = filenames[i]
		}
		fmt.Println(" Picked filename is", InFilename)
		BaseFilename = InFilename
	} else {
		inBuf := flag.Arg(0)
		BaseFilename = filepath.Clean(inBuf)

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

	fmt.Println()

	// Open the input file
	inputFile, err := os.Open(InFilename)
	if err != nil {
		fmt.Printf(" Error from Open(%s) is %s\n", InFilename, err)
		os.Exit(1)
	}
	defer inputFile.Close()

	// Process the input file

	transactions := processCSVfile(inputFile)

	if transactions == nil {
		ctfmt.Println(ct.Red, true, " No transactions found in input file.  Look above for the error message from processCSVfile.  Exiting.")
		os.Exit(1)
	}

	fmt.Println(" Number of transactions is ", len(transactions))

	// Output Section
	// input state is always cc, for credit card transactions, as this is the only type of transaction we process in this program.

	XLoutFilename = BaseFilename + xlsxext
	err = writeOutExcelFile(XLoutFilename, BaseFilename, transactions)
	if err != nil {
		fmt.Printf(" Error writing excel formatted file %s is %s \n", XLoutFilename, err)
	}

	err = AllccAddRecords(BaseFilename, transactions)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from AllccAddRecords is %s\n\n", err.Error())
	}

	/*
		if *outputFlag {
			err := writeCreditCardFileTXT4XLS(BaseFilename, transactions)
			if err != nil {
				ctfmt.Printf(ct.Red, true, "%s\n", err.Error())
			}

			err = writeCreditCardFileCSV4Sqlite(BaseFilename, transactions)
			if err != nil {
				ctfmt.Printf(ct.Red, true, "%s\n", err.Error())
			}
		}

	*/

	fmt.Println()

} // end main

func processCSVfile(inputFile *os.File) []generalTransactionType {
	transactions := make([]generalTransactionType, 0, 100) // magic number I pulled out of the WellsFargo.csv file

	csvReader := csv.NewReader(inputFile)
	rec, err := csvReader.Read() // read the first line which is a header line containing field names.
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from first csvReader.Read() is %s\n", err)
		return nil
	}
	if rec[1] != "DESCRIPTION" {
		ctfmt.Printf(ct.Red, true, " Have the wrong format in the cvs file.  The 2nd field is not DESCRIPTION.\n")
		return nil
	}

	for { // loop through the csv file
		record, err := csvReader.Read()
		if err == io.EOF { // this is the only legit break from this loop.  The rest are errors.
			break
		}
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Error from csvReader.Read() is %s\n", err.Error())
			return nil
		}
		// for WellsFargo:
		// "DATE", "DESCRIPTION", "AMOUNT",    "CHECK #",     "STATUS" are the fields in the Wells Fargo csv file.
		// Juldate  Descript       TRNAMT       CHECKNUM       TRNTYPE
		// DateTime                TRNAMTfloat  CHECKNUMint
		// DTPOSTEDtxt
		// DTPOSTEDcsv

		// Start with DATE
		dateStr := record[0] // in m/d/yyyy format.
		dt, err := time.Parse("1/2/2006", dateStr)
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Error from time.Parse(%s) is %s\n", dateStr, err)
			return nil
		}
		juldate := timlibg.Julian(int(dt.Month()), dt.Day(), dt.Year())
		dateSQL := fmt.Sprintf("%04d-%02d-%02d", dt.Year(), int(dt.Month()), dt.Day())
		if *verboseFlag {
			fmt.Printf("dateStr = %s, dateSQL = %s\n", dateStr, dateSQL)
		}

		// next is DESCRIPTION
		description := record[1]

		// next is AMOUNT
		amountStr := record[2]
		if *verboseFlag {
			fmt.Printf("amountStr = %s\n", amountStr)
		}
		amountFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Error from strconv.ParseFloat(%s) is %s\n", amountStr, err)
			return nil
		}

		// next is CHECK #
		checkNumStr := record[3]
		if *verboseFlag {
			fmt.Printf("checkNumStr = %s\n", checkNumStr)
		}
		checkNumInt := 0
		if checkNumStr != "" {
			checkNumInt, err = strconv.Atoi(checkNumStr)
			if err != nil {
				ctfmt.Printf(ct.Red, true, " Error from strconv.Atoi(%s) is %s\n", checkNumStr, err)
				return nil
			}
		}

		// last is STATUS
		status := record[4]

		transaction := generalTransactionType{
			Juldate:     juldate,
			DateTime:    dt,
			DTPOSTEDtxt: dateStr,
			DTPOSTEDcsv: dateSQL,
			Descript:    description,
			TRNAMT:      amountStr,
			TRNAMTfloat: amountFloat,
			CHECKNUM:    checkNumStr,
			CHECKNUMint: checkNumInt,
			TRNTYPE:     status,
		}

		transactions = append(transactions, transaction)
	}
	return transactions
}

//---------------------------------------------------------------------------------------------------

func writeOutExcelFile(fn string, base string, transactions []generalTransactionType) error {

	if fn == "" || base == "" {
		return fmt.Errorf("writeOutExcelFile: no filename or base name given")
	}

	xlsx.SetDefaultFont(13, "Arial") // the size number doesn't work.  I'm finding it set to 11 when I open the sheet in Excel.

	const excelFormat = `$#,##0.00_);[Red](-$#,##0.00)`
	workbook := xlsx.NewFile()

	comment := base
	if len(base) > 31 { // this limit is set by Excel
		base = base[:10]
	}
	sheet, err := workbook.AddSheet(base)
	if err != nil {
		return err
	}

	// write out heading names
	// "Date \t Amt \t Description \t Comment"
	firstRow := sheet.AddRow()
	cellFirst := firstRow.AddCell()
	cellFirst.SetString("Date")
	cellSecond := firstRow.AddCell()
	cellSecond.SetString("Amt")
	cellThird := firstRow.AddCell()
	cellThird.SetString("Description")
	firstRow.AddCell().SetString("Comment") // just trying this syntax to see if it works.  It does.

	for _, t := range transactions {
		row := sheet.AddRow()
		cell0 := row.AddCell()
		cell0.SetDate(t.DateTime)
		cell1 := row.AddCell()
		cell1.SetFloatWithFormat(t.TRNAMTfloat, excelFormat)
		cell2 := row.AddCell()
		cell2.SetString(t.Descript)
		cell3 := row.AddCell()
		cell3.SetString(comment)
		//                                      cell0.SetString(t.DTPOSTEDtxt)  I want this field to have a date time type, not string type in Excel.
		//                                      cell1.SetString(t.TRNAMT)
	}

	err = workbook.Save(fn)
	return err
}

// openConnection() function is private and only accessed within the scope of the package
func openConnection() (*sql.DB, error) {
	if SQliteDBname == "" {
		return nil, errors.New("database name is empty")
	}
	db, err := sql.Open("sqlite3", SQliteDBname) // SQLite3 does not require a username or a password and does not operate over a TCP/IP network.
	// Therefore, sql.Open() requires just a single parameter, which is the filename of the database.
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AllccAddRecords(base string, records []generalTransactionType) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer func() {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}()

	for _, record := range records {
		if record.DTPOSTEDcsv == "" {
			return errors.New("DTPOSTEDcsv is empty")
		}
		if !checkSQLiteDate(record.DTPOSTEDcsv) {
			s := fmt.Sprintf("DTPOSTEDcsv is not in a valid format.  It is %s", record.DTPOSTEDcsv)
			return errors.New(s)
		}

		// This is how we construct an INSERT statement that accepts parameters. The presented statement requires four values.
		// With db.Exec() we pass the value of the parameters into the insertStatement variable.
		insertStatement := `INSERT INTO Allcc values (?,?,?,?)`
		_, err = db.Exec(insertStatement, record.DTPOSTEDcsv, record.TRNAMTfloat, record.Descript, base) // date, amount, description and comment for Allcc-Sqlite.db
		if err != nil {
			return err
		}
	}

	return nil
}

func checkSQLiteDate(date string) bool {
	regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`) // staticcheck said to use raw string delimiter so I don't have to escape the backslash.
	result := regex.MatchString(date)
	return result
}

//---------------------------------------------------------------------------------------------------
// END fromcsv.go based on fromfx.go based on ofx2csv.go based on qfx2xls.mod
