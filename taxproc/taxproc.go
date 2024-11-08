package main // taxproc.go
import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/tealeg/xlsx/v3"
	"os"
	"src/filepicker"
	"src/misc"
	"src/timlibg"
	"strconv"
	"strings"
	"time"
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
   7 Nov 24 -- The reading in the data to the taxType and gasType slices now works.
*/

const lastModified = "7 Nov 24"
const tempFilename = "debugTaxes.txt"

type taxType struct {
	Date        string
	Description string
	Amount      float64
	Comment     string
	XLdateStr   string
	XLdateNum   int
}

type gasType struct {
	Date   string
	Amount float64
}

var verboseFlag = flag.Bool("v", false, "Verbose mode")

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
		XLdateStr := row.GetCell(0).Value
		if XLdateStr == "" {
			break
		}
		//dateTime, err := time.Parse(time.DateOnly, XLdateStr)  This errored out as it contains a string representation of the excel date number.
		//dateFormat := dateTime.Format("2006-01-02")
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
			Description: row.GetCell(1).Value,
			Amount:      amountFloat,
			Comment:     row.GetCell(4).Value, // column number 3 is Account1, to be skipped.
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
		if strings.Contains(descr, "gas ") {
			price, err := gasStrToNum(descr)
			if err != nil {
				return nil, err
			}
			gas := gasType{
				Date:   tax.Date,
				Amount: price,
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
	fmt.Printf(" Tax Proc, last modified %s, exec binary time stamp is %s\n", lastModified, ExecTimeStamp)

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

	showStuff(taxes, gasPrices)

}

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
		s := fmt.Sprintf("Date = %s, Description = %s, Amount = %.3f, Comment = %s, Excel Date as a string = %s, Excel date as an int = %d\n",
			tax.Date, tax.Description, tax.Amount, tax.Comment, tax.XLdateStr, tax.XLdateNum)
		debugBuf.WriteString(s)
	}
	debugBuf.WriteString("\n Gas:\n")
	for _, g := range gas {
		s := fmt.Sprintf(" Date = %s, Amount = %.3f\n", g.Date, g.Amount)
		debugBuf.WriteString(s)
	}
	debugBuf.WriteString("\n\n")
}
