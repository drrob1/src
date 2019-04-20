package main // multiple go routines using multiple channels, and now with a generator.

import (
	"fmt"
	"io/ioutil"
	"net/http"
	//	"os"
)

func getPage(url string) (int, error) {

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return len(body), nil
}

func worker(urlCh chan string, sizeCh chan string, id int) {

	for {
		url := <-urlCh
		length, err := getPage(url)
		if err == nil { // only handle this if no errors.
			sizeCh <- fmt.Sprintf("%s has length %d, id %d", url, length, id)
		} else {
			sizeCh <- fmt.Sprintf("Error getting %s(%d): %s", url, id, err)
		}
	}
}

func generator(url string, urlCh chan string) {
	urlCh <- url
}

func main() {

	urls := []string{"http://www.yahoo.com", "http://www.cnn.com", "http://www.drrws.com",
		"http://robsolomon.name", "http://drrws.net", "http://drrws.org", "http://bing.com",
		"http://www.robsolomon.info/"}

	urlCh := make(chan string)
	sizeCh := make(chan string)

	// starting 10 workers is a load balancing system so that 10 requests are handled without waiting.  If
	// there was fewer workers than the number of requests, then requests would queue up and wait.

	for i := 0; i < 10; i++ {
		go worker(urlCh, sizeCh, i)
	}

	for _, url := range urls {
		go generator(url, urlCh)
	}

	for i := 0; i < len(urls); i++ {
		fmt.Println(<-sizeCh)

	}

	fmt.Println()
}
