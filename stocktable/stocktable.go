package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

/*
  22 Aug 2024 -- Going to try to build a stock table using twelvedata.com that I learned about from Linux Magazine 285 Aug 2024.
                 Then I played a bit on its website and I registered for a free account.
                 I'll store the opening and closing prices in the database, not just the closing prices that the article code did.
*/

const APIKEY = "0f6e5638d2b742509cf234f1956abcac"

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
	dates := make([]gjson.Result, 0, 500)      // dates := []gjson.Result{} was original code in the article.  There are ~500 trading days in 2 yrs.
	openQuotes := make([]gjson.Result, 0, 500) // quotes := []gjson.Result{} was original code in the article.
	closeQuotes := make([]gjson.Result, 0, 500)

	u := url.URL{Scheme: "https", Host: "api.twelvedata.com", Path: "time_series"}

	q := u.Query()
	q.Set("symbols", symbols)
	q.Set("interval", "1day")
	q.Set("Start_date", "2022-01-01")
	q.Set("apikey", APIKEY) // here is where the login key goes.

	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return dates, openQuotes, closeQuotes, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dates, openQuotes, closeQuotes, err
	}

	dates = gjson.Get(string(body), "values.#.datetime").Array()
	openQuotes = gjson.Get(string(body), "values.#.open").Array()
	closeQuotes = gjson.Get(string(body), "values.#.close").Array()

	return dates, openQuotes, closeQuotes, nil
}

func updater(db *sql.DB, ticker string) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS stockquotes (
        "date" DATE NOT NULL,
        "openquote" REAL NOT NULL,
        "closequote" REAL NOT NULL,
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

	insertSQL := `INSERT OR REPLACE INTO stockquotes (date, openquote, closequote) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer statement.Close()

	for i, date := range dates {
		openingquote := openingQuotes[i]
		closingquote := closingQuotes[i]
		_, err := statement.Exec(date.String(), openingquote.String(), closingquote.String())
		if err != nil {
			return err
		}
	}

	return nil
}

/*
   This is the function that handles the strategy part of the code from the article.
   replay() expects a database handle and executes the SQL SELECT statement which reads all timestamps and closing prices, sorted in ascending order.
   It calls the cb() callback for each date/price value pair.  The callback passes the data to the appropriate strategy or trading function as determined by
   the dispatch table from listing 3.
   SQLITE stores timestamps as strings, which an app can interpret as it needs to.  That's why time.Parse() function is used to create an internal time.Time type,
   stored as prevDt.  This is so simple date arithmetic can be performed on these dates.
*/

func replay(db *sql.DB, cb func(time.Time, float64, float64)) error {
	query := `SELECT date, openquote, closequote FROM stockquotes ORDER BY date ASC`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var openQuote, closeQuote float64
		err := rows.Scan(&date, &openQuote, &closeQuote)
		if err != nil {
			return err
		}
		dt, err := time.Parse("2006-01-02T15:04:05Z07:00", date)
		if err != nil {
			return err
		}
		cb(dt, openQuote, closeQuote)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}

func main() {

	update := flag.Bool("update", false, "update stock price quotes in db")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./stockquotes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *update {
		err := updater(db, "qqq")
		if err != nil {
			log.Fatal(err)
		}
		return
	}
}
