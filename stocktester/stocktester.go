package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
)

/*
	From Linux Magazine 285, Aug 2024.  The stock tester routine they called back tester.

20 Aug 2024 -- First typed into GoLand from the article.
*/
func main() {
	update := flag.Bool("update", false, "update stock price quotes in db")
	strategy := flag.String("strategy", "hold", "stock trader strategy")
	flag.Parse()

	db, err := sql.Open("sqlite3", "./quotes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *update {
		err := updater(db, "nflx")
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	tr := newTrader(*strategy)
	err = replay(db, tr.trade)
	if err != nil {
		log.Fatal(err)
	}

	if tr.holds { // leftover
		tr.sell(tr.prevDt, tr.prevQ)
	}
	fmt.Printf(" Total: %+.2f\n", tr.ledger)
}
