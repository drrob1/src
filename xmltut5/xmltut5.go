package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keyword  string
	Location string
}

func main() {
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/agg/", NewsAggHandler) // news aggregator
	http.ListenAndServe(":8000", nil)        // assumes localhost

}

func index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "output message from index_handler")
}

func NewsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s SitemapIndex
	var n News
	resp, err := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	xml.Unmarshal(bytes, &s)

	news_map := make(map[string]NewsMap)
	for _, location := range s.Locations {
		resp, err := http.Get(location)
		if err != nil {
			panic(err)
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		resp.Body.Close()
		xml.Unmarshal(bytes, &n)
		for idx := range n.Titles {
			tit := strings.TrimSpace(n.Titles[idx])
			k := strings.TrimSpace(n.Keywords[idx])
			news_map[tit] = NewsMap{k, n.Locations[idx]}
		}
	}
	p := NewsAggPage{Title: "A News Aggregator based on the Washington Post", News: news_map}
	t, err := template.ParseFiles("newsaggtemplate.gohtml")
	check(err)
	t.Execute(w, p)
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
