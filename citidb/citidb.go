package main // citidb.go, first developed as allcc.go

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

/*
   3 Nov 24 -- First version of citidb.go.  This will allow me to insert single records when I'm finished developing it.
				This is primary intended to enter voided checks into Citibank.db.
*/

const lastAltered = "3 Nov 24"

type transaction struct {
	status         string
	Date           string
	transactionNum int
	Description    string
	Amount         float64
	AccountName    string
	AccountNumber  int
	UnknownNumber  string // I never found out why this is there.
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
	//if record.Amount == 0 { // amount can, and will likely be, zero.
	//	return errors.New("Amount is empty")
	//}

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
	insertStatement := `INSERT INTO Citibank values (?,?,?,?,?,?,?,?)`
	_, err = db.Exec(insertStatement, record.status, record.Date, record.transactionNum, record.Description, record.Amount,
		record.AccountName, record.AccountNumber, record.UnknownNumber)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	fmt.Printf(" %s last altered %s, timestamp is %s, full exec name is %s\n\n", os.Args[0], lastAltered, ExecTimeStamp, execName)

	//Filename = "Citi-test.db" // the db routines above need this to be defined.
	Filename = "Citibank.db" // the db routines above need this to be defined.
	var date string
	fmt.Print(" Enter date as yyyy-mm-dd : ")
	n, err := fmt.Scanln(&date)
	if n != 1 || err != nil {
		fmt.Printf(" Entered date is in error.  Exiting\n")
		return
	}
	if !strings.Contains(date, "-") {
		fmt.Printf(" Entered date is in error as it's missing the '-'.  Exiting\n")
		return
	}
	if len(date) != 10 {
		fmt.Printf(" Entered date is in wrong format as it's length is wrong.  %q was entered.  Exiting\n", date)
		return
	}

	var transactionNumStr string
	var transactionNum int
	fmt.Print(" Enter transaction number : ")
	n, err = fmt.Scanln(&transactionNumStr)
	if n != 1 || err != nil {
		fmt.Printf(" Entered transaction number is in error.  Exiting\n")
		return
	}
	transactionNum, err = strconv.Atoi(transactionNumStr)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from transaction num Atoi is %s\n", err.Error())
		return
	}

	fmt.Print(" Enter amount : ")
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
	fmt.Print(" Enter description as text (can't have spaces) : ")
	n, err = fmt.Scanln(&description)
	if n != 1 || err != nil {
		fmt.Printf(" Entered description is in error.  Exiting\n")
		return
	}
	if description == "" {
		fmt.Printf(" Warning: entered description is empty.\n")
	}

	var accountName string
	fmt.Print(" Enter AccountName as text (can't have spaces) : ")
	n, err = fmt.Scanln(&accountName)
	if n != 1 || err != nil {
		fmt.Printf(" Entered accountName is in error.  Exiting\n")
	}
	if accountName == "" {
		fmt.Printf(" Warning: entered accountName is empty.\n")
	}

	record := transaction{ // whatever I don't enter will be that type's zero value.
		Date:           date,
		transactionNum: transactionNum,
		Description:    description,
		Amount:         amt,
		AccountName:    accountName,
	}

	err = AddRecord(record)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from AddRecord is %s\n", err.Error())
	}
}
