package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

func main() {
	var ans string
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
		for i := range n.Titles {
			fmt.Print(" Title: ", n.Titles[i])
			fmt.Println()
			fmt.Print(" Keywords: ", n.Keywords[i])
			fmt.Println()
			fmt.Print(" URL: ", n.Locations[i])
			fmt.Println()
			fmt.Print(" hit Enter to continue.")
			fmt.Scanln(&ans)
			if strings.ToLower(ans) == "q" {
				fmt.Println()
				fmt.Println()
				os.Exit(0)
			}
		}
	}
	fmt.Println()
	fmt.Println()
}
