package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"os"
	"src/misc"
	"strings"
	"time"
)

/*
  26 Aug 2024 -- First writing this file.  I'm going to use the code from portfolio that I got from linux mag 277, and SQLite3 code from linux mag 285
                   to do what I initially wanted, in that I can list several stocks, and it will parse and store them correctly.
                   I can't to a test run more often than 1/min because of the limits on the free tier of twelvedata.com.
                   Unless I test w/ only 2 ticker symbols?
  28 Aug 2024 -- main will call sql.Open and then call updater.
                 updater will construct SQL CREATE TABLE if needed, and then call fetchQ.
                 fetchQ uses the twelvedata API and calls parse to construct the qMap that is returned to main.
                 parse populates and returns the qMap to fetchQ.
  29 Aug 2024 -- Adding the replay func.
  31 Aug 2024 -- Adding options 1, 2 and 3, and allow entering stock tickers on command line.  If option1, option2 or option3 are true, then the program will ignore the command line.
                 The command line can have the stock ticker symbols comma separated, comma-space separated, or just space separated.
*/

const APIKEY = "0f6e5638d2b742509cf234f1956abcac"
const dateModified = "31 Aug 2024"
const debugFilename = "portbaseDebug.txt"
const portfolioDatabase = "portbase.db"
const outputTextFilename = "portbaseOutput.txt"
const firstGrouping = "amzn,goog,blk,glw,jpm,qqq"
const secondGrouping = "rsp,vmc,lyg,glw,lamr,txn,vti"
const thirdGrouping = "ebay,duk,sbux,ko,apd,nee,wfc"

var verboseFlag = flag.Bool("v", false, "Verbose mode")
var veryVerbose = flag.Bool("vv", false, "Very Verbose mode for output within a loop")
var option1 = flag.Bool("1", false, "Will get amzn,goog,blk,glw,jpm,qqq")
var option2 = flag.Bool("2", false, "Will get rsp,vmc,lyg,glw,lamr,txn,vti")
var option3 = flag.Bool("3", false, "Will get ebay,duk,sbux,ko,apd,nee,wfc")

type dateVal struct { // for creating the SQLite database
	date         string
	tdate        time.Time
	openingPrice string
	closingPrice string
	symbol       string
}

type stockFileType struct { // for reading from the SQLite database
	date         time.Time
	openingQuote float64
	closingQuote float64
	tickerSymbol string
}

type qMap map[string][]dateVal

func updater(db *sql.DB, w io.Writer, resultMap qMap) error { // from stocktable
	createTableSQL := `CREATE TABLE IF NOT EXISTS stockquotes (
        "date" DATE NOT NULL,
        "openquote" REAL NOT NULL,
        "closequote" REAL NOT NULL,
        "symbol" TEXT NOT NULL,
        UNIQUE(date, openquote, closequote, symbol)
    );`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	insertSQL := `INSERT OR REPLACE INTO stockquotes (date, openquote, closequote, symbol) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer statement.Close()

	for tickSymbol, dv := range resultMap { // dv means dateVal
		if *veryVerbose {
			multiWriteString(w, fmt.Sprintf(" updater resultMap loop: tickSymbol = %s, len(dv) = %d\n", tickSymbol, len(dv)))
		}
		for i := range dv {
			date := dv[i].date
			openingQuote := dv[i].openingPrice
			closingQuote := dv[i].closingPrice
			if *veryVerbose {
				str := fmt.Sprintf(" updater range dv loop: date = %s, openquote = %s, closequote = %s, symbol = %s, %s\n",
					date, openingQuote, closingQuote, dv[i].symbol, tickSymbol)
				multiWriteString(w, str)
			}
			_, err := statement.Exec(date, openingQuote, closingQuote, tickSymbol)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func fetchQ(w io.Writer, symbols string) (qMap, error) { // from portfolio
	u := url.URL{
		Scheme: "https",
		Host:   "api.twelvedata.com",
		Path:   "time_series",
	}

	now := time.Now()
	threeMonthsAgo := now.AddDate(0, -4, -7)
	threeMonthsAgoStr := threeMonthsAgo.Format("2006-01-02")

	q := u.Query()
	q.Set("symbol", symbols)
	q.Set("interval", "1day")
	q.Set("start_date", threeMonthsAgoStr) // these keys are case sensitive.  I initially had this as Start_date, which did not give an error but it did not work.
	q.Set("apikey", APIKEY)
	u.RawQuery = q.Encode()

	reslt := qMap{}

	if *verboseFlag {
		S := fmt.Sprintf("Fetching %s\n", u.String())
		err := multiWriteString(w, S)
		if err != nil {
			panic(err)
		}
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return reslt, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return reslt, err
	}

	reslt = parse(w, string(body))
	return reslt, nil
}

func parse(w io.Writer, data string) qMap { // from portfolio
	everything := gjson.Get(data, "@this").Map()
	if *veryVerbose {
		S := fmt.Sprintf("len(everything)= %d, everything: \n%v\n", len(everything), everything)
		err := multiWriteString(w, S)
		if err != nil {
			panic(err)
		}
	}

	/*
		type dateVal struct {
			date         string
			tdate        time.Time
			openingPrice string
			closingPrice string
			symbol       string
		}
		type qMap map[string][]dateVal
	*/
	//result := qMap{}  // this is from the original code in the article
	result := make(map[string][]dateVal, 500) // I want to pre-allocate the memory.

	for tickerSymbol := range everything {
		if *veryVerbose {
			S := fmt.Sprintf(" tickerSymbol = %s, everything[%s]=%v\n", tickerSymbol, tickerSymbol, everything[tickerSymbol])
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}
		dates := gjson.Get(data, tickerSymbol+".values.#.datetime").Array() // the last field here has to exactly match the field name in the json input
		if *veryVerbose {
			S := fmt.Sprintf("dates = %v\n", dates)
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}

		openingPrice := gjson.Get(data, tickerSymbol+".values.#.open").Array() // the last field here has to exactly match the field name in the json input
		if *veryVerbose {
			S := fmt.Sprintf("openingPrice = %v\n", openingPrice)
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}

		closingPrice := gjson.Get(data, tickerSymbol+".values.#.close").Array() // the last field here has to match what's in the json input
		if *veryVerbose {
			S := fmt.Sprintf("closingPrice = %v\n", closingPrice)
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}

		series := make([]dateVal, 0, 100) // pre-allocate some space.  I don't know if this will be enough, but it's better than nothing.

		for i, date := range dates {
			dt, err := time.Parse("2006-01-02", date.String())
			if err != nil {
				panic(err)
			}

			if *veryVerbose {
				S := fmt.Sprintf("date = %s, and %s\n", date.String(), dt.Format("2006-01-02T15:04:05Z07:00"))
				err := multiWriteString(w, S)
				if err != nil {
					panic(err)
				}
			}

			dateValue := dateVal{
				date:         date.String(),
				tdate:        dt,
				openingPrice: openingPrice[i].String(),
				closingPrice: closingPrice[i].String(),
				symbol:       tickerSymbol,
			}
			series = append(series, dateValue)
		}

		result[tickerSymbol] = series
	}

	return result
}

func replay(db *sql.DB) ([]stockFileType, error) {
	query := `SELECT date, openquote, closequote, symbol FROM stockquotes ORDER BY symbol, date ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stockSlice := make([]stockFileType, 0, 500) // just guessing the initial size of the slice I'll need

	for rows.Next() {
		var date, symbol string
		var openQuote, closeQuote float64
		err := rows.Scan(&date, &openQuote, &closeQuote, &symbol)
		if err != nil {
			return nil, err
		}
		dt, err := time.Parse("2006-01-02T15:04:05Z07:00", date)
		if err != nil {
			return nil, err
		}

		slice := stockFileType{
			date:         dt,
			openingQuote: openQuote,
			closingQuote: closeQuote,
			tickerSymbol: symbol,
		}
		stockSlice = append(stockSlice, slice)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stockSlice, nil
}

func main() {
	flag.Parse()
	if *veryVerbose {
		*verboseFlag = true
	}
	fmt.Printf(" Portfolio Base, using SQLite3 to create a stock portfolio database.  Last modified %s\n", dateModified)

	var symbols string
	if *option1 {
		symbols = firstGrouping // firstGrouping = "amzn,goog,blk,glw,jpm,qqq"
	} else if *option2 {
		symbols = secondGrouping // secondGrouping = "rsp,vmc,lyg,glw,lamr,txn,vti"
	} else if *option3 {
		symbols = thirdGrouping // ebay,duk,sbux,ko,apd,nee,wfc
	} else if flag.NArg() == 0 {
		fmt.Printf("Usage: portbase [options] tickerSymbols in a comma separated format without spaces.\n")
	} else if flag.NArg() == 1 {
		symbols = flag.Arg(0)
	} else if flag.NArg() > 1 {
		stringSlice := flag.Args()
		if strings.Contains(stringSlice[0], ",") {
			symbols = strings.Join(stringSlice, "")
		} else {
			symbols = strings.Join(stringSlice, ",")
		}
		symbols = removeAllSpaces(symbols) // this may be overkill, as the above statements may already achieve this goal
	}

	file, buf, err := misc.CreateOrAppendWithBuffer(debugFilename)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	defer buf.Flush()
	w := io.MultiWriter(buf, os.Stdout)

	now := time.Now()
	todayStr := now.Format("2006-01-02")
	threeMonthsAgo := now.AddDate(0, -4, -7)
	threeMonthsAgoStr := threeMonthsAgo.Format("2006-01-02")
	threeMonthsAgoOutput := fmt.Sprintf(" Today is %s, three months ago was %s", todayStr, threeMonthsAgoStr)
	err = multiWriteString(w, threeMonthsAgoOutput)
	if err != nil {
		fmt.Printf("Error writing threeMonthsAgo: %v\n", err)
	}

	if *verboseFlag {
		S := fmt.Sprintf(" symbols: %s\n", symbols)
		err := multiWriteString(w, S)
		if err != nil {
			panic(err)
		}
	}

	result, err := fetchQ(w, symbols) // result is a qMap
	if err != nil {
		panic(err)
	}

	// result is a qMap from twelvedata
	// Now to write to the database.

	db, err := sql.Open("sqlite3", portfolioDatabase)
	if err != nil {
		fmt.Printf("Error opening database %s: %v\n", portfolioDatabase, err)
		return
	}
	defer db.Close()

	err = updater(db, w, result)
	if err != nil {
		fmt.Printf(" Error updating %s: %v\n", portfolioDatabase, err)
		return
	}

	completionMessage := fmt.Sprintf("Portfolio database %s updated successfully!\n", portfolioDatabase)
	multiWriteString(w, completionMessage)

	// Now will read back the SQLite database.

	outputSlice, err := replay(db)
	if err != nil {
		fmt.Printf(" replay error: %s\n", err)
		return // return is better than os.Exit because of all the defer statements.
	}

	//outputFile, outputBuf, err := misc.CreateOrAppendWithBuffer(outputTextFilename)  I don't want the file to be continually appended to.
	outputFile, err := os.Create(outputTextFilename)
	if err != nil {
		fmt.Printf("Error creating output text file: %v\n", err)
		return
	}
	outputBuf := bufio.NewWriter(outputFile)
	layout := "Jan 02 2006"
	defer outputFile.Close()
	defer outputBuf.Flush()
	for _, stock := range outputSlice {
		_, err := outputBuf.WriteString(fmt.Sprintf("%s, %.2f, %.2f, %s\n",
			stock.date.Format(layout), stock.openingQuote, stock.closingQuote, stock.tickerSymbol))
		if err != nil {
			fmt.Printf(" output text file write error: %s\n", err)
			return
		}
	}

	fmt.Printf(" Stock file written with %d lines.\n", len(outputSlice))

}

func multiWriteString(w io.Writer, str string) error {
	_, err := w.Write([]byte(str))
	if err != nil {
		return err
	}
	w.Write([]byte("\n"))
	pause()
	return nil
}

func pause() {
	fmt.Printf(" ... pausing ... until hit <enter>")
	fmt.Scanln()
}

func removeAllSpaces(str string) string {
	if !strings.Contains(str, " ") {
		return str
	}
	return strings.ReplaceAll(str, " ", "")
}
