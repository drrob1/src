package main

import (
	"encoding/csv"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	"github.com/tealeg/xlsx/v3"
	"os"
	"path/filepath"
	"regexp"
	"src/filepicker"
	"src/misc"
	"strconv"
	"strings"
	"time"
)

/*
30 May 25 -- Starting to think about how I would import and use the csv package to process the lightning-bolt csv files.
31 May 25 -- It's working to read the csv and write the processed csv.  Now I want to add writing in Excel format.  Nevermind.
			I think I'll create 2 slices, one byAssignment and the other byDate.  So the data can be more easily retrievable by either.
			Basically, this is by column and by row.  I don't think I need by row after all, only by column, i.e., by date.
			Maybe I just need to populate the table so it can be viewed.  I may not need to do anything, just teach them to download this file and read it into Excel.
			I decided to use the xlsx package to write an Excel file.
			It works as intended.
*/

const LastAltered = "31 May 25"
const csvext = ".csv"

var verboseFlag bool
var veryVerboseFlag bool

func writeXLSX(baseFilename string, table [][]string) (string, error) {
	// I decided to populate an Excel type table, and then write it out.

	//workBook, err := xlsx.OpenFile(templateName)
	//if err != nil {
	//	return err
	//}

	workbook := xlsx.NewFile()
	comment := removeExt(baseFilename)
	if len(comment) > 31 { // this limit is set by Excel
		comment = comment[:30]
	}

	sheet, err := workbook.AddSheet(comment)
	if err != nil {
		return "", err
	}

	_, _ = sheet.Cell(0, 1) // just to allow this to compile, for now.

	for i, row := range table { // remember that xl is 1-based, but the xlsx routines handle this correctly
		for j, field := range row {
			cell, err := sheet.Cell(i, j)
			if err != nil {
				fmt.Println(" Error from fmt.Fprintln: ", err, ".  Exiting.")
				return "", err
			}
			if isDate(field) {
				timedate, err := time.Parse("1/2/2006", field)
				if err != nil {
					return "", err
				}
				cell.SetDate(timedate)
				continue
			}
			cell.SetString(field)
		}
	}

	number := misc.RandRange(1, 1000)
	numStr := strconv.Itoa(number)
	fn := baseFilename + "_" + numStr + ".xlsx"
	err = workbook.Save(fn) // I don't want to clober anything while I'm testing.

	return fn, err
}

func main() {

	fmt.Println(" schedwork.go lastModified is", LastAltered)
	var InFilename, BaseFilename string
	var InFileExists bool
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose mode")
	flag.BoolVarP(&veryVerboseFlag, "veryverbose", "w", false, "very verbose mode")
	flag.BoolVar(&veryVerboseFlag, "vv", false, "very verbose mode")
	flag.Parse()
	if veryVerboseFlag {
		verboseFlag = true
	}

	if flag.NArg() < 1 {
		filenames, err := filepicker.GetRegexFilenames("csv$") // $ matches end of line
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from GetRegexFilenames is %v, exiting\n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s \n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code= 999 . , / ;) : ")
		var ans string
		n, err := fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil || n == 0 { // these are redundant.  I'm playing now.
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == "/" || ans == ";" {
			fmt.Println(" Stop code entered.")
			return
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
			InFilename = BaseFilename + csvext
			_, err := os.Stat(InFilename)
			if err == nil {
				InFileExists = true
			}
		}

		if !InFileExists {
			fmt.Println(" File ", BaseFilename, BaseFilename+csvext, " or ", InFilename, " do not exist.  Exiting.")
			return
		}
		fmt.Println(" input filename is ", InFilename)
	}

	dir, rawBase := filepath.Split(InFilename)
	var err error
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			fmt.Println(" Error from os.Getwd: ", err, ".  Exiting.")
			return
		}
	}
	base := removeExt(rawBase)
	if verboseFlag {
		fmt.Printf(" dir is %s, base is %s, raw base is %s\n", dir, base, rawBase)
	}

	// Open the file for reading.
	f, err := os.ReadFile(InFilename)
	if err != nil {
		fmt.Println(" Error from os.ReadFile: ", err, ".  Exiting.")
		os.Exit(1)
	}

	// read in file using csv package.
	r := csv.NewReader(strings.NewReader(string(f)))
	r.Comment = '#'
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println(" Error from r.ReadAll: ", err, ".  Exiting.")
		return
	}
	fmt.Println(" Finished reading ", len(records), " records from ", InFilename)

	//for {
	//	record, err := r.Read()
	//	if err == io.EOF {
	//		break
	//	}
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	fmt.Println(record)
	//}

	// construct output file name.

	BaseFilename = base
	OutFilename := BaseFilename + "_processed.out"
	f2, err := os.OpenFile(OutFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error from os.Create: ", err, ".  Exiting.")
		os.Exit(1)
	}
	defer f2.Close()

	// write processed records to file.

	var sum int
	var recCount int
	for _, record := range records {
		for _, field := range record {
			n, err := fmt.Fprintf(f2, "%10s |", field)
			if err != nil {
				fmt.Println(" Error from fmt.Fprintln: ", err, ".  Exiting.")
				return
			}
			sum += n
		}
		recCount++
		fmt.Fprintln(f2)
	}

	fmt.Printf(" Finished writing %d bytes and %d records to %s. \n", sum, recCount, OutFilename)
	fmt.Printf(" \n\n\n Getting ready to populate the byAssignment and byDate slices. \n\n\n")

	// write out Excel

	outName, err := writeXLSX(BaseFilename, records)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from writeXLSX is %s\n", err)
	} else {
		ctfmt.Printf(ct.Green, true, " Finished writing Excel file %s\n", outName)
	}

}

func removeExt(filename string) string {
	if !strings.HasSuffix(filename, csvext) {
		return filename
	}
	return filename[:len(filename)-len(csvext)]
}

func isDate(instr string) bool {
	regexStr := `^[0-3]?[0-9]/[0-3]?[0-9]/(?:[0-9]{2})?[0-9]{2}$`
	regex := regexp.MustCompile(regexStr)
	isdate := regex.MatchString(instr)
	return isdate
}
