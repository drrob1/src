package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Location struct {
	Loc string `xml:"loc"`
}

type SitemapIndex struct {
	Locations []Location `xml:"sitemap"`
}

func (L Location) String() string {
	return fmt.Sprintf(L.Loc)
}

func main() {
	fmt.Println("vim-go")
	resp, err := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	//	string_body := string(bytes)
	//	fmt.Println(string_body)

	fmt.Println("unmarshalling now")
	var s SitemapIndex
	err = xml.Unmarshal(bytes, &s)
	if err != nil {
		panic(err)
	}

	fmt.Println(s)
	fmt.Println()

	for _, q := range s.Locations {
		fmt.Println(q) // implicit method prints fields with {} delimiters
	}
	fmt.Println()

}
