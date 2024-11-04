package main // allcc.go

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/*
   3 Nov 24 -- First version of allcc.go.  This will allow me to insert single records into allcc-sqlite.db.  At least when I'm finished developing it.
				Then I'll write this for Citibank.db.
   4 Nov 24 -- Writing date validation routine using regexp.
*/

const lastAltered = "4 Nov 24"

type transaction struct {
	Date        string
	Amount      float64
	Description string
	Comment     string
}

// Filename is the name of the database file, to be set by the client.
var (
	Filename = "" // name of the database file, to be set by the client.
)

// openConnection() function is private and only accessed within the scope of the package
func openConnection() (*sql.DB, error) {
	if Filename == "" {
		return nil, errors.New("Filename is empty")
	}
	db, err := sql.Open("sqlite3", Filename) // SQLite3 does not require a username or a password and does not operate over a TCP/IP network.
	// Therefore, sql.Open() requires just a single parameter, which is the filename of the database.
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AddRecord(record transaction) error {
	if record.Date == "" {
		return errors.New("Date is empty")
	}
	if record.Amount == 0 {
		return errors.New("Amount is empty")
	}

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

	// This is how we construct an INSERT statement that accepts parameters. The presented statement requires four values.
	// With db.Exec() we pass the value of the parameters into the insertStatement variable.
	insertStatement := `INSERT INTO Allcc values (?,?,?,?)`
	_, err = db.Exec(insertStatement, record.Date, record.Amount, record.Description, record.Comment)
	if err != nil {
		return err
	}

	return nil
}

func checkDate(date string) bool {
	regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`) // staticcheck said to use raw string delimiter so I don't have to escape the backslash.
	regex2 := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
	regex3 := regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}$")
	result1 := regex.MatchString(date)
	result2 := regex2.MatchString(date)
	result3 := regex3.MatchString(date)
	fmt.Printf(" Result1 using raw string backslash d is %t, result2 using 0-9 is %t, and result 3 using double backslash is %t\n",
		result1, result2, result3)
	return result1 || result2 || result3
}

func main() {
	flag.Parse()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	fmt.Printf(" %s last altered %s, timestamp is %s, full exec name is %s\n\n", os.Args[0], lastAltered, ExecTimeStamp, execName)

	Filename = "allcc-test.db" // the db routines above need this to be defined.
	var date string
	fmt.Print(" Enter date as yyyy-mm-dd : ")
	n, err := fmt.Scanln(&date)
	if n != 1 || err != nil {
		fmt.Printf(" Entered date is in error.  Exiting\n")
		return
	}
	if !strings.Contains(date, "-") {
		fmt.Printf(" Entered date is in error as it's missing the '-'.  Using today.\n")
		date = time.Now().Format(time.DateOnly)
	}
	if len(date) != 10 {
		fmt.Printf(" Entered date is in wrong format as it's length is wrong.  %q was entered.  Using today.\n", date)
		date = time.Now().Format(time.DateOnly)
	}
	if checkDate(date) {
		fmt.Printf(" Entered date is valid.  Great.\n")
	} else {
		fmt.Printf(" Entered date is NOT valid.  Using today.\n")
		date = time.Now().Format(time.DateOnly)
	}

	fmt.Print(" Enter amount as float : ")
	var amt float64
	var amtstr string
	n, err = fmt.Scanln(&amtstr)
	if n != 1 || err != nil {
		fmt.Printf(" Entered amount is in error.  Exiting\n")
		return
	}
	amt, err = strconv.ParseFloat(amtstr, 64)
	if err != nil {
		ctfmt.Printf(ct.Red, true, "Error from parsing %s to float is %s\n", amtstr, err)
		return
	}
	var description string
	fmt.Print(" Enter description as text (can't have spaces -- just kidding as I fixed this.) : ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	description = scanner.Text()
	// n, err = fmt.Scanln(&description) spaces in the input causes this input to fail.
	//if n != 1 || err != nil {
	//	fmt.Printf(" Entered description is in error.  Exiting\n")
	//	return
	//}
	if description == "" {
		fmt.Printf(" Warning: entered description is empty.\n")
	}
	var comment string
	fmt.Print(" Enter comment as text (can't have spaces) -- I fixed this : ")
	scanner.Scan()
	comment = scanner.Text()
	//n, err = fmt.Scanln(&comment) spaces in the input causes this input to fail.
	//if n != 1 || err != nil {
	//	fmt.Printf(" Entered comment is in error.  Exiting\n")
	//}
	if comment == "" {
		fmt.Printf(" Warning: entered comment is empty.\n")
	}

	record := transaction{
		Date:        date,
		Amount:      amt,
		Description: description,
		Comment:     comment,
	}

	err = AddRecord(record)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from AddRecord is %s\n", err.Error())
	}
}
