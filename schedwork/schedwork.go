package main

import (
	"encoding/csv"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"path/filepath"
	"src/filepicker"
	"strconv"
	"strings"
)

/*
30 May 25 -- Starting to think about how I would import and use the csv package to process the lightning-bolt csv files.
*/

const LastAltered = "31 May 25"
const csvext = ".csv"

func main() {

	fmt.Println(" schedwork.go lastModified is", LastAltered)
	var InFilename, BaseFilename string
	var InFileExists bool
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

	// write processed records to file.

	BaseFilename = filepath.Base(InFilename)
	OutFilename := BaseFilename + "_processed.out"
	f2, err := os.OpenFile(OutFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error from os.Create: ", err, ".  Exiting.")
		os.Exit(1)
	}
	defer f2.Close()

	var sum int
	var recCount int
	for _, record := range records {
		n, err := fmt.Fprintf(f2, "%320s", record)
		if err != nil {
			fmt.Println(" Error from fmt.Fprintln: ", err, ".  Exiting.")
			return
		}
		sum += n
		recCount++
	}
	fmt.Printf(" Finished writing %d bytes and %d records to %s. \n", sum, recCount, OutFilename)
}
