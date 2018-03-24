package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {
	fmt.Println("vim-go")
	var wg sync.WaitGroup

	var urls = []string{
		"https://www.golang.org",
		"https://www.google.org",
		"https://www.nytimes.com",
		"https://www.cnn.com",
		"https://www.bing.com",
	}

	fetchurl := func(url string) { // an anonymous closure
		defer wg.Done()
		resp, err := http.Get(url)
		if err != nil {
			log.Println("err from http.Get:", err)
		}
		fmt.Println(" resp.Body from", url, ":", resp.Body)
		//		bytes, err := ioutil.ReadAll(resp.Body)
		//		if err != nil {
		//			log.Println("err from ioutil.ReadAll:", err)
		//		}
		//		fmt.Println("resp.Body:", bytes) // this is much more complex than I thought.
		resp.Body.Close()
	}

	for _, url := range urls {
		wg.Add(1)
		go fetchurl(url)
	}

	wg.Wait()
	fmt.Println("vim-stop")
}
