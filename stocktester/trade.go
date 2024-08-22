package main

import (
	"fmt"
	"time"
)

/*
  From listing 3 of linux magazine 285 Aug 2024.
*/

type tradeFu func(time.Time, float64)

type trader struct {
	holds    bool
	cost     float64
	prevQ    float64
	prevDt   time.Time
	ledger   float64
	runStrat tradeFu
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
