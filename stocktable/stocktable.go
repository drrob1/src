package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"src/misc"
	"time"
)

/*
  22 Aug 2024 -- Going to try to build a stock table using twelvedata.com that I learned about from Linux Magazine 285 Aug 2024.
                 Then I played a bit on its website and I registered for a free account.
                 I'll store the opening and closing prices in the database, not just the closing prices that the article code did.
  23 Aug 2024 -- It works.  I used SQLiteStudio to check.
  24 Aug 2024 -- Now to test reading from the sqlite database by writing to a text file.
  28 Aug 2024 -- Summary of the logic here: main calls sql.Open, then it calls updater, then it calls replay to retrieve from the database just written.
                 updater constructs the SQL CREATE TABLE if needed, then calls fetchQ.
                 fetchQ uses the twelvedata API and then uses gjson to populate the slices that will be returned to updater.
                 updater takes these slices and uses them in the SQL INSERT OR REPLACE statement.
*/

const APIKEY = "0f6e5638d2b742509cf234f1956abcac"
const lastModified = "Aug 24, 2024"
const stockFilename = "stocktable.txt"

var verbose = flag.Bool("v", false, "Verbose mode")
var veryVerbose = flag.Bool("vv", false, "Very Verbose mode for output within a loop")

const jsonExt = ".json"

var jsonFilename string

type stockFileType struct {
	date         time.Time
	openingQuote float64
	closingQuote float64
	tickerSymbol string
}

/*
https://api.twelvedata.com/time_series?apikey=demo&symbol=qqq&interval=4h
{
    "meta": {
        "symbol": "QQQ",
        "interval": "4h",
        "currency": "USD",
        "exchange_timezone": "America/New_York",
        "exchange": "NASDAQ",
        "mic_code": "XNMS",
        "type": "ETF"
    },
    "values": [
        {
            "datetime": "2024-08-22 13:30:00",
            "open": "477.89001",
            "high": "478.07550",
            "low": "474.86499",
            "close": "476.17999",
            "volume": "7687721"
        },
        {
            "datetime": "2024-08-22 09:30:00",
            "open": "484.82999",
            "high": "485.51999",
            "low": "476.79999",
            "close": "477.91989",
            "volume": "19585036"
        },
        ...
*/

func fetchQ(symbols string) ([]gjson.Result, []gjson.Result, []gjson.Result, error) {
	//dates := []gjson.Result{}  From the original code.  I want to pre-allocate some memory so I'm using make below.
	//openQuotes := []gjson.Result{}
	//closeQuotes := []gjson.Result{}
	dates := make([]gjson.Result, 0, 500)
	openQuotes := make([]gjson.Result, 0, 500)
	closeQuotes := make([]gjson.Result, 0, 500)

	u := url.URL{Scheme: "https", Host: "api.twelvedata.com", Path: "time_series"}

	q := u.Query()
	q.Set("symbol", symbols)
	q.Set("interval", "1day")
	q.Set("start_date", "2022-01-02") // these keys are case sensitive.  I initially had this as Start_date, which did not give an error but it did not work.
	q.Set("apikey", APIKEY)           // here is where the login key goes.

	u.RawQuery = q.Encode()
	if *verbose {
		fmt.Printf(" in fetchQ: GET %s\n", u.String())
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return dates, openQuotes, closeQuotes, err
	}
	defer resp.Body.Close()

	//js, err := json.Marshal(resp.Body)
	//if err != nil {
	//	fmt.Printf(" Error from json marshal is %v.\n", err)
	//	os.Exit(1)
	//}
	//err = os.WriteFile(jsonFilename, js, 0666)
	//if err != nil {
	//	fmt.Printf(" Error from json WriteFile is %v.\n", err)
	//}
	//
	//if *veryVerbose {
	//	return dates, openQuotes, closeQuotes, err
	//}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dates, openQuotes, closeQuotes, err
	}
	if *verbose {
		fmt.Printf("fetchQ: len(body) = %d, err = %s\n response body = %v\n body = %v\n", len(body), err, resp.Body, body)
	}

	dates = gjson.Get(string(body), "values.#.datetime").Array()
	openQuotes = gjson.Get(string(body), "values.#.open").Array()
	closeQuotes = gjson.Get(string(body), "values.#.close").Array()

	if *verbose {
		fmt.Printf("fetchQ: Len of dates = %d, Open quotes = %d, Close quotes = %d\n", len(dates), openQuotes, closeQuotes)
	}

	return dates, openQuotes, closeQuotes, nil
}

func updater(db *sql.DB, ticker string) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS stockquotes (
        "date" DATE NOT NULL,
        "openquote" REAL NOT NULL,
        "closequote" REAL NOT NULL,
        "symbol" TEXT NOT NULL,
        UNIQUE(date, openquote, closequote)
    );`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	dates, openingQuotes, closingQuotes, err := fetchQ(ticker)
	if err != nil {
		return err
	}

	if *verbose {
		fmt.Printf(" updater: len(dates) = %d, len(openingquotes) = %q, len(closingquotes) = %d\n", len(dates), len(openingQuotes), len(closingQuotes))
	}

	insertSQL := `INSERT OR REPLACE INTO stockquotes (date, openquote, closequote, symbol) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer statement.Close()

	for i, date := range dates {
		openingquote := openingQuotes[i]
		closingquote := closingQuotes[i]
		if *veryVerbose {
			fmt.Printf(" updater date loop: date = %v, openingquote=%v, closingquote=%v\n", date, openingquote, closingquote)
		}
		_, err := statement.Exec(date.String(), openingquote.String(), closingquote.String(), ticker)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
   This is the function that handles the strategy part of the code from the article.
   replay() expects a database handle and executes the SQL SELECT statement which reads all fields and sorts in ascending order.
   It calls the cb() callback for each date/price value pair.  The callback passes the data to the appropriate strategy or trading function as determined by
   the dispatch table from listing 3.
   SQLITE stores timestamps as strings, which an app can interpret as it needs to.  That's why time.Parse() function is used to create an internal time.Time type,
   stored as prevDt.  This is so simple date arithmetic can be performed on these dates.
*/

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
		*verbose = true
	}

	fmt.Printf(" StockTable routine using sqlite3 database format, last modified %s.\n", lastModified)
	if *verbose {
		execName, _ := os.Executable()
		ExecFI, _ := os.Stat(execName)
		ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
		fmt.Printf("%s timestamp is %s, full exec is %s\n", execName, ExecTimeStamp, execName)
	}

	if flag.NArg() < 1 {
		fmt.Printf(" Need ticker symbol for the query.\n")
		return
	}
	tickerSymbol := flag.Arg(0)
	fmt.Printf(" Ticker symbol: %s\n", tickerSymbol)

	jsonFilename = tickerSymbol + jsonExt
	fmt.Printf(" json filename = %s\n", jsonFilename)

	db, err := sql.Open("sqlite3", "./stockquotes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = updater(db, tickerSymbol)
	if err != nil {
		log.Fatal(err)
	}

	// creating and writing to the database works, now to retrieve the elements and write to a text file.
	stockSlice, err := replay(db)
	if err != nil {
		fmt.Printf(" replay error: %s\n", err)
		return // return is better than os.Exit because of all the defer statements.
	}

	stockFile, stockBuf, err := misc.CreateOrAppendWithBuffer(stockFilename)
	layout := "Jan 02 2006"
	defer stockFile.Close()
	defer stockBuf.Flush()
	for _, stock := range stockSlice {
		_, err := stockBuf.WriteString(fmt.Sprintf("%s, %.2f, %.2f, %s\n",
			stock.date.Format(layout), stock.openingQuote, stock.closingQuote, stock.tickerSymbol))
		if err != nil {
			fmt.Printf(" stock file write error: %s\n", err)
			return
		}
	}

	fmt.Printf(" Stock file written with %d lines.\n", len(stockSlice))

}
