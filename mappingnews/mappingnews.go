package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keyword  []string
	Location string
}

func main() {
	resp, err := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	var s SitemapIndex
	xml.Unmarshal(bytes, &s)

	var n News
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
			kslice := strings.Split(k, ",")
			news_map[tit] = NewsMap{kslice, n.Locations[idx]}
		}
	}
	/*
	   	fmt.Println(" News.Titles length: ", len(n.Titles), ", Keywords: ", len(n.Keywords), ", Locations: ", len(n.Locations))
	   turns out that all of the slices are of equal length.  There is only one title and location per entry, but the multiple keywords are comma separated.  That is ignored in the orig code here.  I'm not sure that my string splitting code is better.
	   	var ans string
	   	fmt.Println()
	   	fmt.Print(" hit Enter to continue.")
	   	fmt.Scanln(&ans)
	   	if strings.ToLower(ans) == "q" {
	   		fmt.Println()
	   		fmt.Println()
	   		os.Exit(0)
	   	}
	*/
	for idx, data := range news_map {
		//		fmt.Println("Title: ", idx, "\nKeyword: ", data.Keyword, "\nLocation: ", data.Location)
		fmt.Print("Title: ", idx, "\nLocation: ", data.Location, "\nKeywords")
		for _, k := range data.Keyword {
			fmt.Print(", ", k)
		}
		fmt.Println()
		fmt.Println("Keywords:", strings.Join(data.Keyword, ", "))
	}

}
