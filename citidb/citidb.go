package main // citidb.go, first developed as allcc.go

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
   3 Nov 24 -- First version of citidb.go.  This will allow me to insert single records when I'm finished developing it.
				This is primarily intended to enter voided checks into Citibank.db.
   4 Nov 24 -- Added checkDate, and using a scanner to read in text from the terminal, so spaces are allowed now.
  26 Apr 25 -- allowing date entry in m/d/y format
*/

const lastAltered = "27 Apr 25"

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

// Filename is the global name of the database file, to be set by the client.
var (
	Filename = "" // name of the database file, to be set by the client.
)

// openConnection() function is private and only accessed within the scope of the package
func openConnection() (*sql.DB, error) {
	if Filename == "" {
		return nil, errors.New("filename is empty")
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
		return errors.New("date is empty")
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
	fmt.Print(" Enter date as yyyy-mm-dd (default is today) : ")
	n, err := fmt.Scanln(&date)
	if n != 1 || err != nil {
		fmt.Printf(" Entered date is in error.  Will use today.\n")
		date = time.Now().Format(time.DateOnly)
	}

	if checkDate(date) {
		fmt.Printf(" Entered date is in YYYY-MM-DD format.\n")
	} else {
		formattedDate := tryDateMDY(date)
		if formattedDate == "" {
			fmt.Printf(" Entered date is NOT valid.  Using today.\n")
			date = time.Now().Format(time.DateOnly)
		}
		fmt.Printf(" Entered date is valid.  Using %s.\n", formattedDate)
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

	fmt.Print(" Enter amount (default is zero) : ")
	var amt float64
	var amtstr string
	n, err = fmt.Scanln(&amtstr)
	if n != 1 || err != nil {
		fmt.Printf(" Entered amount returned an error.  Assuming zero\n")
		amtstr = "0"
	}
	amt, err = strconv.ParseFloat(amtstr, 64)
	if err != nil {
		ctfmt.Printf(ct.Red, true, "Error from parsing %s to float is %s\n", amtstr, err)
		amt = 0.0
	}

	var description string
	fmt.Print(" Enter description as text (default is void) : ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	description = scanner.Text()
	if description == "" {
		fmt.Printf(" Entered description is empty.  Using void.\n")
		description = "void"
	}

	var accountName string
	fmt.Print(" Enter AccountName as text : ")
	scanner.Scan()
	accountName = scanner.Text()
	if accountName == "" {
		fmt.Printf(" Warning: entered accountName is empty.  Assuming checking\n")
		accountName = "checking"
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

func checkDate(date string) bool {
	regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`) // staticcheck said to use raw string delimiter so I don't have to escape the backslash.
	result := regex.MatchString(date)
	return result
}

func tryDateMDY(date string) string {
	regex1 := regexp.MustCompile(`\d{1,2}-\d{1,2}-\d{2}`)
	result := regex1.MatchString(date)
	if result { // Input date is in format m-d-yy
		// Input date is in format m-d-yy.  Will convert to yyyy-mm-dd
		formattedDate := parseMDYdash(date)
		return formattedDate // this may be blank if there was an error
	}
	regex2 := regexp.MustCompile(`\d{1,2}-\d{1,2}-\d{4}`)
	result = regex2.MatchString(date)
	if result {
		formattedDate := parseMDYdash(date)
		return formattedDate // this may be blank if there was an error
	}
	regex3 := regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{2}`)
	regex4 := regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{4}`)
	result = regex3.MatchString(date) || regex4.MatchString(date)
	if result {
		formattedDate := parseMDYslash(date)
		return formattedDate
	}
	regex5 := regexp.MustCompile(`\d{1,2}.\d{1,2}.\d{2}`)
	regex6 := regexp.MustCompile(`\d{1,2}.\d{1,2}.\d{4}`)
	result = regex5.MatchString(date) || regex6.MatchString(date)
	if result {
		formattedDate := parseMDYdot(date)
		return formattedDate
	}
	return ""
}

func parseMDYdash(date string) string {
	tokens := strings.Split(date, "-")
	if len(tokens) != 3 {
		return ""
	}
	month, err := strconv.Atoi(tokens[0])
	if err != nil {
		return ""
	}
	day, err := strconv.Atoi(tokens[1])
	if err != nil {
		return ""
	}
	year, err := strconv.Atoi(tokens[2])
	if err != nil {
		return ""
	}
	if year < 100 {
		year += 2000
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || year < 1900 || year > 2030 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func parseMDYslash(date string) string {
	tokens := strings.Split(date, "/")
	if len(tokens) != 3 {
		return ""
	}
	month, err := strconv.Atoi(tokens[0])
	if err != nil {
		return ""
	}
	day, err := strconv.Atoi(tokens[1])
	if err != nil {
		return ""
	}
	year, err := strconv.Atoi(tokens[2])
	if err != nil {
		return ""
	}
	if year < 100 {
		year += 2000
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || year < 1900 || year > 2030 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
func parseMDYdot(date string) string {
	tokens := strings.Split(date, ".")
	if len(tokens) != 3 {
		return ""
	}
	month, err := strconv.Atoi(tokens[0])
	if err != nil {
		return ""
	}
	day, err := strconv.Atoi(tokens[1])
	if err != nil {
		return ""
	}
	year, err := strconv.Atoi(tokens[2])
	if err != nil {
		return ""
	}
	if year < 100 {
		year += 2000
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || year < 1900 || year > 2030 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
