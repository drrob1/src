package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
)

/*
  From listing 2 of Linux Magazine 285 Aug 2024 for the stock trading program.

  21 Aug 24 -- The updater() function uses the Twelvedata API to retrieve the daily prices from Jan 1, 2022 to today, at once.
               The SQLite engine inserts the daily quotes into a database table, quotes.db.
               The actual API request is sent by fetchQ().  If successful, Twelvedata sends a json response from which the gjson library extracts the trading days and
                 closing prices as arrays using the values.*.datetime and values.#.close queries.
               The exec command uses INSERT OR REPLACE to insert the closing prices for each date into the database table.  The replace option is needed to avoid duplicates if an
                 entry already exists for a given date.
*/

const APIKEY = "0f6e5638d2b742509cf234f1956abcac"

func updater(db *sql.DB, ticker string) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS quotes (
        "date" DATE NOT NULL,
        "quote" REAL NOT NULL,
        UNIQUE(date, quote)
    );`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	dates, quotes, err := fetchQ(ticker)
	if err != nil {
		return err
	}

	insertSQL := `INSERT OR REPLACE INTO quotes (date, quote) VALUES (?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer statement.Close()

	for i, date := range dates {
		quote := quotes[i]
		_, err := statement.Exec(date.String(), quote.String())
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchQ(symbols string) ([]gjson.Result, []gjson.Result, error) {
	dates := make([]gjson.Result, 0, 500)  // dates := []gjson.Result{} was original code in the article.  There are ~500 trading days in 2 yrs.
	quotes := make([]gjson.Result, 0, 500) // quotes := []gjson.Result{} was original code in the article.
	u := url.URL{Scheme: "https", Host: "api.twelvedata.com", Path: "time_series"}

	q := u.Query()
	q.Set("symbol", symbols)
	q.Set("interval", "1day")
	q.Set("start_date", "2022-01-01")
	q.Set("apikey", APIKEY) // here is where the login key goes.

	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return dates, quotes, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dates, quotes, err
	}

	dates = gjson.Get(string(body), "values.#.datetime").Array()
	quotes = gjson.Get(string(body), "values.#.close").Array()

	return dates, quotes, nil
}
