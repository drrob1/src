package main

import (
	"fmt"
	"io/ioutil"
)
import "net/http"

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

	string_body := string(bytes)
	resp.Body.Close()
	fmt.Println(string_body)

}
