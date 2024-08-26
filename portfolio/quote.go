package main

import (
	"io"
	"net/http"
	"net/url"
)

/*
  From Linux Magazine 277 Dec 2023.  Terminal Stock Portfolio Display

  25 Aug 2024 -- First started typing this in, from listing 1.

*/

const APIKEY = "0f6e5638d2b742509cf234f1956abcac"

type dateVal struct {
	date         string
	openingPrice string
	closingPrice string
}
type qMap map[string][]dateVal

func fetchQ(symbols string) (qMap, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "api.twelvedata.com",
		Path:   "time_series",
	}

	q := u.Query()
	q.Set("symbol", symbols)
	q.Set("interval", "1day")
	q.Set("apikey", APIKEY)
	u.RawQuery = q.Encode()

	reslt := qMap{}

	resp, err := http.Get(u.String())
	if err != nil {
		return reslt, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return reslt, err
	}

	reslt = parse(string(body))
	return reslt, nil
}
