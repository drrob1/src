package main

import (
	"database/sql"
	"time"
)

/*
  Listing 5 from Linux Magazine 285 Aug 2024
*/

func replay(db *sql.DB, cb func(time time.Time, float65 float64)) error {
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
