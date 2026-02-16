package main // taxproc.go
import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"src/filepicker"
	"src/misc"
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
   6 Nov 24 -- Started working on taxproc.
               Now that I know how to read and write .xlsx files, and also how to write to an SQLite3 .db file, I can write code to process my taxesyy file that I do once a year.
               This will take a few days, at least.  I'll think about it here and write as I think.

               I have to create a file to be read into access, update taxes.db, and update GasGraph.xlsm.

02/29/2020 8:55:10 AM Once a year I import the taxesyy.xlsm file into access and sqlite.  I'm putting the steps (instructions) here because I keep forgetting them.
                      1) delete the summation stuff, excess rows below the last entry to the item that says "end of database definition"
                      2) delete column that says account number (whatever it is)
                      3) column headings need to be date, description, amount, comment for trimmed Excel file.
                      4) Save taxesyy-trimmed.xlsm
                      5) Remove column headings row for SQLiteStudio
                      6) Change date format to yyyy-mm-dd and save taxesyy.csv

               The GasGraph is only 2 columns, Date and Amount which is the price/gal I paid before any discounts were applied.

				Use filepicker to select the desired file.
               I'll start by reading into a slice of struct, the date, description and amount.
               And into a separate slice of struct where description has gas, just date and price/gal
               The reading stops at the first empty date field.  It will report out how many entries it found for both slices, one for the trimmed data and one for GasGraph.

               Then I can write out separate .xlsx files for taxesyy-trimmed.xlsx and for gasyy.xlsx

				First, I'll focus on debugging the reading in the data and creating the slices for the taxesyy entries and for the gas price points.
				Then I'll write out the 2 different .xlsx files and debug that.
   7 Nov 24 -- The reading in the data to the taxType and gasType slices now works.  I think I'll come back to this tomorrow.
                 The problems I had was first that Excel uses a julian date number w/ day 1 being 1/1/1900.  I used a fudge factor to adjust that to my julian date number system.
                 The other problem I had was parsing the gas price field.  I forgot that I stopped using the '@' and had to just remove the 3 characters of "gas" and then TrimSpaces the result.
                 Tomorrow I'll start writing the code to write the Excel file to be read into Access, and then directly updating the taxes.db file.
                 I can use code from fromfx.go for both of these tasks, as I just wrote that for fromfx.go.
   8 Nov 24 -- Taxes.mdb fields are Date (mm/dd/yyyy), Description, Amount, Comment.
               taxes.db fields are Date (DATETIME), Descr (TEXT), Amount (REAL), Comment (TEXT).
   9 Nov 24 -- Added Floor to correct small floating point errors automatically
  10 Nov 24 -- Now that it works, I'm going to see if I can get it to work using the xlsx methods instead of my own timlibg methods to convert to/from Excel date julian numbers.
                And I'm using strings.TrimSpaces on the description and comment fields.  This is so that the HasPrefix looking for "gas" will succeed if I have a stray " " in
				the beginning of the cell.
  16 Feb 26 -- Updated message to make clear that it takes a taxesyy.xlsm file as input.  I don't remove the macros or change the format of the file.
				I had trouble today, turned out to be because of missing entry for amount.  I got an error from the FormatFloat function.  But I didn't know where.
				So now I added verboseFlag to show the row output before processing.  When it stops and shows the error, I can see that it's the next row w/ a problem.
				I can't tell if I have to remove the empty rows until the database ends.  I don't remember about this.  Now I see, on line 109 below, if there is a blank date,
				then the reading loop breaks.  The problem I have w/ the taxes25 file is that I'm making entries when I send a file to Billy but I don't intend for these
				entries to be extracted.  Maybe all I need is a blank row to end the reading loop and then have my notes about sending to Billy after a blank line.
*/

const lastModified = "16 Feb 26"
const tempFilename = "debugTaxes.txt"

type taxType struct {
	Date        string
	TimeDate    time.Time
	DateOnly    string
	Description string
	Amount      float64
	Comment     string
	XLdateStr   string
	XLdateNum   int
}

type gasType struct {
	Date     string
	TimeDate time.Time
	Price    float64
}

var verboseFlag = flag.Bool("v", false, "Verbose mode")

var SQliteDBname = "" // name of the database file, to be set below, for use in the direct updating of the SQLite3 database files below.

func readTaxes(xl string) ([]taxType, error) {
	workBook, err := xlsx.OpenFile(xl)
	if err != nil {
		return nil, err
	}

	sheets := workBook.Sheets
	taxSlice := make([]taxType, 0, 300)

	for i := 4; i < 300; i++ { // start at row 5 in Excel, which is row 4 here in a 0-origin system.  Excel is a 1-origin system, as origin is cell A1.
		row, err := sheets[0].Row(i)
		if err != nil {
			return nil, err
		}
		if *verboseFlag {
			ctfmt.Printf(ct.Cyan, false, "Processing row %d: row : %v\n", i, row)
		}
		XLdateStr := row.GetCell(0).Value
		if XLdateStr == "" {
			break
		}
		//dateTime, err := time.Parse(time.DateOnly, XLdateStr)  This errored out as it contains a string representation of the Excel date number.
		timedate, err := row.GetCell(0).GetTime(false)
		if err != nil {
			return nil, err
		}
		dateFormat := timedate.Format("2006-01-02")
		if dateFormat != timedate.Format(time.DateOnly) { // just to make sure I'm right that these are equivalent.
			ctfmt.Printf(ct.Red, false, "These 2 different ways to achieve the same format should be equal.  They're not.  first = %s, second = %s\n",
				dateFormat, timedate.Format(time.DateOnly))
		}

		XLdateNum, err := strconv.Atoi(XLdateStr)
		if err != nil {
			return nil, err
		}
		dateOnly := timlibg.FromExcelToDateOnly(XLdateNum)

		amountStr := row.GetCell(2).Value
		amountFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return nil, err
		}
		taxEntry := taxType{
			Date:        dateOnly,
			TimeDate:    timedate,
			DateOnly:    dateFormat,
			Description: strings.TrimSpace(row.GetCell(1).Value),
			Amount:      amountFloat,
			Comment:     strings.TrimSpace(row.GetCell(4).Value), // column number 3 is Account1, to be skipped.
			XLdateStr:   XLdateStr,
			XLdateNum:   XLdateNum,
		}
		taxSlice = append(taxSlice, taxEntry)
	}
	return taxSlice, nil
}

// gasStrToNum converts the description of a gas entry that looks like "gas  3.009" or "gas @3.009".  It may or may not have '@'
func gasStrToNum(str string) (float64, error) {
	var s string
	idx := strings.Index(str, "@")
	if idx == -1 { // no '@' found
		s = str[3:] // I'm skipping over the 3 char's "gas"
	} else {
		s = str[idx+1:]
	}
	s = strings.TrimSpace(s)
	price, err := strconv.ParseFloat(s, 64)
	return price, err
}

func gasData(taxes []taxType) ([]gasType, error) {
	gasSlice := make([]gasType, 0, 100)
	for _, tax := range taxes {
		descr := strings.ToLower(strings.TrimSpace(tax.Description))
		if strings.HasPrefix(descr, "gas ") { // this choked on the gas range line for 2024.  So I made it look for a prefix of "gas ", and this is working.
			price, err := gasStrToNum(descr)
			if err != nil {
				return nil, err
			}
			gas := gasType{
				Date:     tax.Date,
				TimeDate: tax.TimeDate,
				Price:    price,
			}
			gasSlice = append(gasSlice, gas)
		}
	}
	return gasSlice, nil
}

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	SQliteDBname = "taxes.db"

	fmt.Printf(" Tax Proc, last modified %s, exec binary time stamp is %s, SQLiteDBname = %s\n", lastModified, ExecTimeStamp, SQliteDBname)
	fmt.Printf(" Select a taxesyy.xlsm file, without altering its format or structure.  If there are rows to not be extracted, put them at the bottom after a blank line.\n")

	var filename, ans string

	flag.Parse()

	// filepicker stuff.

	if flag.NArg() == 0 {
		filenames, err := filepicker.GetRegexFilenames("taxes.*xls.$")
		if err != nil {
			ctfmt.Printf(ct.Red, false, " Error from filepicker is %s.  Exiting \n", err)
			return
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == ";" {
			fmt.Println(" Stop code entered.  Exiting.")
			return
		}
		i, e := strconv.Atoi(ans)
		if e == nil {
			filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			filename = filenames[i]
		}
		fmt.Println(" Picked spreadsheet is", filename)
	} else { // will use filename entered on commandline
		filename = flag.Arg(0)
	}

	if *verboseFlag {
		fmt.Printf(" spreadsheet picked is %s\n", filename)
	}
	fmt.Println()

	taxes, err := readTaxes(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from readTaxes(%s) is %s.  Exiting \n", filename, err)
		return
	}

	gasPrices, err := gasData(taxes)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from gasData(%s) is %s.  Exiting \n", filename, err)
		return
	}

	fmt.Printf("Found %d tax items, and %d gas price points.\n", len(taxes), len(gasPrices))
	showStuff(taxes, gasPrices)

	// constructing the output file names from the input taxesyy.xlsm file.
	base := filepath.Base(filename)
	idx := strings.Index(base, ".")

	if idx > -1 { // there's an extension here so need to trim that off
		base = base[:idx]
	}
	outTaxesFilename := "processed" + filename
	outGasFilename := "gas" + filename

	if *verboseFlag {
		fmt.Printf(" outTaxesFilename is %s, base = %s, idx = %d, outGasFilename = %s\n", outTaxesFilename, base, idx, outGasFilename)
	}

	err = writeOutExcelTaxesFile(outTaxesFilename, base, taxes)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from writeOutExcelTaxesFile(%s) is %s.  Exiting \n", filename, err)
	}

	err = TaxesAddRecords(base, taxes)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from TaxesAddRecords(%s) is %s.  Exiting \n", SQliteDBname, err)
	}

	err = writeOutGasExcelFile(outGasFilename, base, gasPrices)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from writeOutGasFile(%s) is %s.  Exiting \n", outGasFilename, err)
	}

	ctfmt.Printf(ct.Green, true, "Taxes and Gas files created successfully.  \n")
}

// showStuff writes the contents of the taxes and gas slices to the tempFilename, which is now debugTaxes.txt
func showStuff(taxes []taxType, gas []gasType) {
	debugFile, debugBuf, err := misc.CreateOrAppendWithBuffer(tempFilename)
	defer func() {
		err = debugBuf.Flush()
		if err != nil {
			panic(err)
		}
		err := debugFile.Close()
		if err != nil {
			panic(err)
		}
	}()

	debugBuf.WriteString("------------------------------------------------------------------------\n")
	now := time.Now()
	debugBuf.WriteString(now.Format(time.ANSIC))
	debugBuf.WriteString("\n Taxes:\n")
	for _, tax := range taxes {
		s := fmt.Sprintf("Date=%s, Description=%s, Amount=%.3f, Comment=%s, Excel Date as string=%s, Excel date as int=%d, TimeDate=%s, DateOnly=%s\n",
			tax.Date, tax.Description, tax.Amount, tax.Comment, tax.XLdateStr, tax.XLdateNum, tax.TimeDate.String(), tax.DateOnly)
		debugBuf.WriteString(s)
	}
	debugBuf.WriteString("\n Gas:\n")
	for _, g := range gas {
		s := fmt.Sprintf(" Date = %s, Price = %.3f\n", g.Date, g.Price)
		debugBuf.WriteString(s)
	}
	debugBuf.WriteString("\n\n")
}

// writeOutExcelFile writes the file to be read into Excel (ExcelTaxesFilename) and then for Taxes.mdb, fields are Date (mm/dd/yyyy), Description, Amount, Comment.
func writeOutExcelTaxesFile(fn string, base string, taxes []taxType) error {

	xlsx.SetDefaultFont(13, "Arial") // the size number doesn't work.  I'm finding it set to 11 when I open the sheet in Excel.

	const excelMoneyFormat = `$#,##0.00_);[Red](-$#,##0.00)`
	// const excelDateFormat = "*3/14/2012"  not used now that I'm outputting a TimeDate directly.

	// Need to make sure that the extension is .xlsx and not .xlsm
	lastChar := len(fn)
	fn = fn[:lastChar-1]
	fn = fn + "x"

	workbook := xlsx.NewFile()

	if len(base) > 31 { // this limit is set by Excel
		base = base[:10]
	}
	sheet, err := workbook.AddSheet(base)
	if err != nil {
		return err
	}

	firstRow := sheet.AddRow()
	cellFirst := firstRow.AddCell()
	cellFirst.SetString("Date")
	cellSecond := firstRow.AddCell()
	cellSecond.SetString("Description")
	cellThird := firstRow.AddCell()
	cellThird.SetString("Amount")
	firstRow.AddCell().SetString("Comment") // just trying this syntax to see if it works for the 4th column.  It does.

	//dateOptions := xlsx.DateTimeOptions{  This didn't work.  It panicked.  See updated comments in writeOutGasExcelFile below
	//	Location:        nil,
	//	ExcelTimeFormat: excelDateFormat,
	//}

	//  fields are Date (mm/dd/yyyy), Description, Amount, Comment.
	for _, t := range taxes {
		row := sheet.AddRow()
		cell0 := row.AddCell()
		//                                           cell0.SetString(t.Date) // now I don't want to set this as a string.  Excel thinks it's a string type this way.  But it does work.
		//                                           cell0.SetDateWithOptions(t.TimeDate, dateOptions)  This panicked
		cell0.SetDate(t.TimeDate)
		cell1 := row.AddCell()
		cell1.SetString(t.Description)
		cell2 := row.AddCell()
		cell2.SetFloatWithFormat(t.Amount, excelMoneyFormat)
		cell3 := row.AddCell()
		// I want to only write the comment if this is not a gas line because the comment would just be a number and is not wanted in the taxes.mdb file.
		descr := strings.ToLower(strings.TrimSpace(t.Description))
		if !strings.HasPrefix(descr, "gas ") { // only write out this comment if it's NOT a gas item, which would just be a number that I don't want in the databases.
			cell3.SetString(t.Comment)
		}
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

// TaxesAddRecords -- base means the base input filename and records is obvious.
func TaxesAddRecords(base string, taxes []taxType) error {
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

	// taxes.db fields are Date (DATETIME), Descr (TEXT), Amount (REAL), Comment (TEXT).
	for _, t := range taxes {
		if t.Date == "" {
			return errors.New("taxes.Date is empty")
		}
		if !checkSQLiteDate(t.Date) {
			s := fmt.Sprintf("taxes.Date is not in a valid format.  It is %s", t.Date)
			return errors.New(s)
		}

		// This is how we construct an INSERT statement that accepts parameters. The presented statement requires four values.
		// With db.Exec() we pass the value of the parameters into the insertStatement variable.
		amount := misc.Floor(t.Amount, 4)
		insertStatement := `INSERT INTO taxes values (?,?,?,?)`
		_, err = db.Exec(insertStatement, t.Date, t.Description, amount, t.Comment) // Date, Description, Amount and Comment for taxes.db.  Amount is automatically corrected by Floor.
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

// writeOutGasExcelFile -- fn is the filename that's written.  base is the sheet name that's created.
func writeOutGasExcelFile(fn string, base string, gas []gasType) error {

	xlsx.SetDefaultFont(13, "Arial") // the size number doesn't work.  I'm finding it set to 11 when I open the sheet in Excel.

	const excelMoneyFormat = `$#,##0.00_);[Red](-$#,##0.00)`
	//const excelDateFormat = "*3/14/2012"  Not used now that I'm outputting a TimDate directly.
	//newYork, err := time.LoadLocation("America/New_York")
	//if err != nil {
	//	return err
	//}
	//dateOptions := xlsx.DateTimeOptions{ //This panicked when I tried to use this, when I made Location nil.  Now it doesn't panic, but Excel complained about bad data when opened in Excel.
	//	Location:        newYork,
	//	ExcelTimeFormat: excelDateFormat,
	//}

	// Need to make sure that the extension is .xlsx and not .xlsm
	lastChar := len(fn)
	fn = fn[:lastChar-1]
	fn = fn + "x"

	workbook := xlsx.NewFile()

	if len(base) > 31 { // this limit is set by Excel
		base = base[:10]
	}
	sheet, err := workbook.AddSheet(base)
	if err != nil {
		return err
	}

	firstRow := sheet.AddRow()
	cellFirst := firstRow.AddCell()
	cellFirst.SetString("Date")
	cellSecond := firstRow.AddCell()
	cellSecond.SetString("Price")

	//  fields are Date (mm/dd/yyyy), Description, Amount, Comment.
	for _, g := range gas {
		row := sheet.AddRow()
		cell0 := row.AddCell()
		// cell0.SetString(g.Date)  This sets the cell as a string type, not a datetime type.  But it does work.
		// cell0.SetDateWithOptions(g.TimeDate, dateOptions) // This panicked before I made Loccation not nil.  Now it doesn't panic, but Excel complains about bad data when opened in Excel.
		cell0.SetDate(g.TimeDate)
		cell1 := row.AddCell()
		cell1.SetFloatWithFormat(g.Price, excelMoneyFormat)
	}

	err = workbook.Save(fn)
	return err
}
