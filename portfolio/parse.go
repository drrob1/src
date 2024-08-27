package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
)

/*
  25 Aug 2024 -- From Linux Magazine 277 Dec 2023
                 Listing 2
*/

func parse(w io.Writer, data string) qMap {
	everything := gjson.Get(data, "@this").Map()
	if *verboseFlag {
		S := fmt.Sprintf("len(everything)= %d, everything: \n%v\n", len(everything), everything)
		err := multiWriteString(w, S)
		if err != nil {
			panic(err)
		}
	}
	reslt := qMap{}

	for tickerSymbol := range everything {
		if *verboseFlag {
			S := fmt.Sprintf(" tickerSymbol = %s, everything[%s]=%v\n", tickerSymbol, tickerSymbol, everything[tickerSymbol])
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}
		dates := gjson.Get(data, tickerSymbol+".values.#.datetime").Array() // the last field here has to exactly match the field name in the json input
		if *verboseFlag {
			S := fmt.Sprintf("dates = %v\n", dates)
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}

		openingPrice := gjson.Get(data, tickerSymbol+".values.#.open").Array() // the last field here has to exactly match the field name in the json input
		if *verboseFlag {
			S := fmt.Sprintf("openingPrice = %v\n", openingPrice)
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}

		closingPrice := gjson.Get(data, tickerSymbol+".values.#.close").Array() // the last field here has to match what's in the json input
		if *verboseFlag {
			S := fmt.Sprintf("closingPrice = %v\n", closingPrice)
			err := multiWriteString(w, S)
			if err != nil {
				panic(err)
			}
		}
		series := []dateVal{}

		for i, date := range dates {
			if *verboseFlag {
				S := fmt.Sprintf("date = %s\n", date.String())
				err := multiWriteString(w, S)
				if err != nil {
					panic(err)
				}
			}
			dateValue := dateVal{
				date:         date.String(),
				openingPrice: openingPrice[i].String(), // here is where it's blowing up.
				closingPrice: closingPrice[i].String(),
			}
			series = append(series, dateValue)
		}

		reslt[tickerSymbol] = series
	}

	return reslt
}
