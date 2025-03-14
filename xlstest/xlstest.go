package main // xlstest.go   Testing what happens when I add a sheet to a workbook in code
import (
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path/filepath"
	"src/filepicker"
	"src/misc"
	"strconv"
	"strings"
)

/*
  11 Nov 24 -- First version.  Turns out that adding a sheet here adds it as the last sheet, not first.
  14 Mar 25 -- Today's Pi day.  But that's not important now.  Testing if open file takes a full filename.  It does.  Tested on linux so far.
				And filepath.Split() does split correctly if the last part is a regexp.  Tested here on linux so far, and the windows syntax strings don't parse correctly.
*/

var lastModified = "14 Mar 2025"

func main() {
	fmt.Printf(" xlstest, last modified %s.\n", lastModified)

	flag.Parse()

	var filename, ans string

	// test filepicker.Split
	testString := "/home/rob/Documents/xlsx$"
	dir, name := filepath.Split(testString)
	fmt.Printf("dir: %q name: %q \n", dir, name)

	testString = "o:\\xlsx$" // doesn't work on linux
	dir, name = filepath.Split(testString)
	fmt.Printf("dir: %q name: %q \n", dir, name)

	testString = "c:\\users\\rsolomon\\documents\\xlsx$" // doesn't work on linux
	dir, name = filepath.Split(testString)
	fmt.Printf("dir: %q name: %q \n\n", dir, name)

	testString = "c:/users/rsolomon/documents/xlsx$" // this does parse correctly, on linux.
	dir, name = filepath.Split(testString)
	fmt.Printf("dir: %q name: %q \n\n", dir, name)

	// filepicker stuff.

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Printf("homeDir: %q , workingDir: %q\n", homeDir, workingDir)

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

	fmt.Println()
	fmt.Printf("filename: %q \n", filename)

	filename = workingDir + string(os.PathSeparator) + filename
	fmt.Printf(" after adding workingDir, filename: %q \n", filename) // this worked, so far on linux.

	workBook, err := xlsx.OpenFile(filename)
	if err != nil {
		fmt.Printf("Error opening excel file %s in directory %s: %s\n", filename, workingDir, err)
		return
	}

	xlsx.SetDefaultFont(13, "Arial") // the size number doesn't work.  I'm finding it set to 11 when I open the sheet in Excel.

	n := misc.RandRange(1000, 100_000)

	base := strconv.Itoa(n)
	sheet, err := workBook.AddSheet(base)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error adding sheet %s to workbook: %s\n", base, err)
		return
	}

	fmt.Println("Sheets in this file:")
	for i, sh := range workBook.Sheets {
		fmt.Println(i, sh.Name)
	}

	cell021, err := sheet.Cell(0, 21) // getting contents of cells before they're allocated actually allocates them.
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 0,21. %s\n", err)
	}
	fmt.Printf(" cell 021 is %q\n", cell021.String())

	cell121, err := workBook.Sheets[0].Cell(1, 21)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 1,21. %s\n", err)
	}
	fmt.Printf(" cell 121 is %q\n", cell121.String())

	//cell210, err := workBook.Sheets[1].Cell(21, 0) // getting the contents here actually allocated all rows before, too.  So the AddRow call below added after this row.
	//if err != nil {
	//	ctfmt.Printf(ct.Red, false, " Error getting cell 21,1 in sheet 2. %s\n", err)
	//}
	//fmt.Printf(" cell 210 in sheet 2 is %s\n", cell210.String())

	originCell, err := workBook.Sheets[1].Cell(0, 0)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 0,0. %s\n", err)
	}
	originCell.SetString("this is at 0,0")
	row := sheet.AddRow()
	firstCell := row.AddCell()
	firstCell.SetString("First cell")
	cell100, err := workBook.Sheets[1].Cell(0, 0)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 0,0. %s\n", err)
	}
	fmt.Printf(" cell00 should be this is at 0,0, and is %q\n\n", cell100.String()) // this works

	cell210, err := workBook.Sheets[1].Cell(1, 0)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error getting cell 1,0. %s\n", err)
	}
	fmt.Printf(" cel210 should be First Cell and is %q\n\n", cell210.String()) // this works

	err = workBook.Save(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, false, " Error saving file %s to %s\n", filename, err)
	}
}
