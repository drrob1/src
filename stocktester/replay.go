package main

import (
	"database/sql"
	"time"
)

/*
  Listing 5 from Linux Magazine 285 Aug 2024

  21 Aug 24 -- The simulation driver goes thru all the trading days of the 2-year period and activates the currently selected strategy function.
               replay() expects a database handle and executes the SQL SELECT statement which reads all timestamps and closing prices, sorted in ascending order.
               It calls the cb() callback for each date/price value pair.  The callback passes the data to the appropriate strategy or trading function as determined by
               the dispatch table from listing 3.
               SQLITE stores timestamps as strings, which an app can interpret as it needs to.  That's why time.Parse() function is used to create an internal time.Time type,
                stored as prevDt.  This is so simple date arithmetic can be performed on these dates.
*/

func replay(db *sql.DB, cb func(time.Time, float64)) error { // this function handles the strategy part of the code, by using this callback function.
	query := `SELECT date, quote FROM quotes ORDER BY date ASC`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var quote float64
		err := rows.Scan(&date, &quote)
		if err != nil {
			return err
		}
		dt, err := time.Parse("2006-01-02T15:04:05Z07:00", date)
		if err != nil {
			return err
		}
		cb(dt, quote)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}
