package main

import "time"

/*
  Listing 4 from Linux Magazine 285 Aug 2024
  21 Aug 24 -- hold keeps to the buy and hold strategy.  That is, buy is and not sell it.
               buydrop buys the stock when it falls 2% of the previous days price, and sells it when the price changes by 10% up or down.
                 This is typical day trader behavior.
               firstweek strategy is invented for this article.  It buys at the start of each month and sells after 5 trading days.
               The dispatch table, disp, refers to the corresponding strategy function as selected by the command line flag.  This is a map (hash table) that assigns function pointers
               to methods because the trader struct "tr" is part of it.  This is typical for a receiver in Go.
               The trade function expects the timestamp of a trading day and the closing price that day as its params.
*/

// buy and hold forever.  The holds bool checks to see if the stock is held in the portfolio.  If not, call buy.  Else hold it until the last day covered here.  The framework's engine
// takes its turn at the end, sees that the stock is still in the portfolio, and sells it at the last closing price.

func (tr *trader) strat_hold() tradeFu {
	tradeFunc := func(dt time.Time, quote float64) {
		if !tr.holds {
			tr.buy(dt, quote)
		}
	}
	return tradeFunc
}

// buy the stock if the closing price < 2% of the prev day's quote (prevQ).  When the unrealized gain/loss is more than 10% of the cost, the sell method is called.

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

// This silly strategy only invests in the first week of very month.  The strategy function here returns an investor function to its caller, but can define variables in its
// environment.  The function then drags these along using the closure process.  Eg, the local held variable defines the number of days during which the strategy has already
// held the position.  The function has to determine that the first trading day of a month has arrived, and buy() is called.  After 5 days in the securities account,
// the condition "held > 5" becomes true and sell() is called.

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
