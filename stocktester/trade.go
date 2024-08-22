package main

import (
	"fmt"
	"time"
)

/*
  From listing 3 of linux magazine 285 Aug 2024.

  21 Aug 24 -- The trader struct is defined here.  A bool is used to indicate whether or not the trader holds the stock.
               There is a trade() wrapper that first calls the strategy and ensures that the previous day's price is included in the trader structure for later use.
               I noticed that these trader functions use pointer semantics to update the trader struct each time.
*/

type tradeFu func(time.Time, float64) // timestamp of a trading day and the closing price from that day.

type trader struct {
	holds    bool
	cost     float64   // price at which the stock was bought
	prevQ    float64   // previous day's price
	prevDt   time.Time // previous day's date
	ledger   float64   // running total of gain/loss from all previous transactions.
	runStrat tradeFu   // called each trading day to decide what to do.
}

func newTrader(strategy string) *trader {
	tr := trader{}
	disp := map[string]func() tradeFu{
		"hold":      tr.strat_hold,
		"buydrop":   tr.strat_buydrop,
		"firstweek": tr.strat_firstweek,
	}
	tr.runStrat = disp[strategy]()
	return &tr
}

func (tr *trader) sell(dt time.Time, quote float64) {
	tr.holds = false
	tr.ledger += quote - tr.cost
	fmt.Printf(" Selling %s %.2f (%+.2f) total %+.2f\n", dt.Format("2006-01-02"), quote, quote-tr.cost, tr.ledger)
}

func (tr *trader) buy(dt time.Time, quote float64) {
	fmt.Printf(" Buying %s %.2f\n", dt.Format("2006-01-02"), quote)
	tr.holds = true
	tr.cost += quote
}

func (tr *trader) trade(dt time.Time, quote float64) {
	tr.runStrat(dt, quote)
	tr.prevQ = quote
	tr.prevDt = dt
}
