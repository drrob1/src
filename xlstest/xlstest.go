package main // xlstest.go   Testing what happens when I add a sheet to a workbook in code
import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/tealeg/xlsx/v3"
	"os"
	"src/filepicker"
	"strconv"
	"strings"
)

/*
  11 Nov 24 -- First version.  Turns out that adding a sheet here adds it as the last sheet, not first.
*/

func main() {
	fmt.Printf(" xlstest.\n")

	flag.Parse()

	var filename, ans string

	// filepicker stuff.

	if flag.NArg() == 0 {
		filenames, err := filepicker.GetRegexFilenames("xlsx$")
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

	fmt.Printf(" spreadsheet picked is %s\n", filename)
	fmt.Println()

	workingDir, _ := os.Getwd()
	workBook, err := xlsx.OpenFile(filename)
	if err != nil {
		fmt.Printf("Error opening excel file %s in directory %s: %s\n", filename, workingDir, err)
		return
	}

	xlsx.SetDefaultFont(13, "Arial") // the size number doesn't work.  I'm finding it set to 11 when I open the sheet in Excel.

	base := "I am first"
	sheet, err := workBook.AddSheet(base)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error adding sheet %s to workbook: %s\n", base, err)
		return
	}

	fmt.Println("Sheets in this file:")
	for i, sh := range workBook.Sheets {
		fmt.Println(i, sh.Name)
	}

	cell021, err := sheet.Cell(0, 21)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 0,21. %s\n", err)
	}
	fmt.Printf(" cell 021 is %s\n", cell021.String())

	cell121, err := workBook.Sheets[0].Cell(1, 21)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 1,21. %s\n", err)
	}
	fmt.Printf(" cell 121 is %s\n", cell121.String())

	//cell210, err := workBook.Sheets[1].Cell(21, 0)
	//if err != nil {
	//	ctfmt.Printf(ct.Red, false, " Error getting cell 21,1 in sheet 2. %s\n", err)
	//}
	//fmt.Printf(" cell 210 in sheet 2 is %s\n", cell210.String())

	row := sheet.AddRow()
	firstCell := row.AddCell()
	firstCell.SetString("First cell")
	cell00, err := workBook.Sheets[0].Cell(0, 0)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 0,0. %s\n", err)
	}
	fmt.Printf(" cell00 is %s\n\n", cell00.String())

	err = workBook.Save(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error saving file %s to %s\n", filename, err)
	}
}
