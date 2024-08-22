package main

import "time"

/*
  Listing 4 from Linux Magazine 285 Aug 2024
*/

func (tr *trader) strat_hold() tradeFu {
	tradeFunc := func(dt time.Time, quote float64) {
		if !tr.holds {
			tr.buy(dt, quote)
		}
	}
	return tradeFunc
}

func (tr *trader) strat_buydrop() tradeFu {
	tradeFunc := func(dt time.Time, quote float64) {
		if tr.prevQ != 0 {
			if tr.holds {
				if quote > 1.1*tr.cost || quote < 0.9*tr.cost {
					tr.sell(dt, quote)
				}
			} else {
				if quote < 0.98*tr.prevQ {
					tr.buy(dt, quote)
				}
			}
		}
	}
	return tradeFunc
}

func (tr *trader) strat_firstweek() tradeFu {
	held := 0
	tradeFunc := func(dt time.Time, quote float64) {
		if tr.holds {
			held++
			if held > 5 {
				tr.sell(dt, quote)
				held = 0
			}
		} else {
			if dt.Day() < 7 {
				tr.buy(dt, quote)
			}
		}
	}
	return tradeFunc
}
