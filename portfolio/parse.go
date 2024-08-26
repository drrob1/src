package main

import "github.com/tidwall/gjson"

/*
  25 Aug 2024 -- From Linux Magazine 277 Dec 2023
                 Listing 2
*/

func parse(data string) qMap {
	everything := gjson.Get(data, "@this").Map()
	reslt := qMap{}

	for tickerSymbol := range everything {
		dates := gjson.Get(string(data), tickerSymbol+".values.#.datetime").Array()
		openingPrice := gjson.Get(string(data), tickerSymbol+".values.#.openingPrice").Array()
		closingPrice := gjson.Get(string(data), tickerSymbol+".values.#.closingPrice").Array()
		series := []dateVal{}

		for i, date := range dates {
			dateValue := dateVal{
				date:         date.String(),
				openingPrice: openingPrice[i].String(),
				closingPrice: closingPrice[i].String(),
			}
			series = append(series, dateValue)
		}

		reslt[tickerSymbol] = series
	}

	return reslt
}
