package main

import (
	"fmt"
	"net/http"
	"sync"
)

func main() {
	fmt.Println("vim-go")
	var wg sync.WaitGroup

	var urls = []string{
		"www.golang.org",
		"www.google.org",
		"www.nytimes.com",
		"www.cnn.com",
		"www.bing.com",
	}

	fetchurl := func(url string) {
		wg.Add(1)
		defer wg.Done()
		http.Get(url)
	}

	for _, url := range urls {
		go fetchurl(url)
	}

	wg.Wait()
	fmt.Println("vim-stop")
}
