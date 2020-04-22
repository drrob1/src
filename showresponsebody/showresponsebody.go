/*
  From Black Hat Go.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Printf(" Usage: %s URLs \n", os.Args[0])
		os.Exit(1)
	}

	urls := make([]string, 0, 5)

	urls = append(urls, "https://www.google.com/robots.txt")

	for _, url := range os.Args[1:] {
		urls = append(urls, "http://"+url)
	}

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			continue
		}

		// Print HTTP Status
		fmt.Printf(" URL: %s, Status string: %s, status number is %d \n", url, resp.Status, resp.StatusCode)

		// Read and display response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panicln(err)
		}

		fmt.Println(string(body))
		resp.Body.Close()
		var ans string
		fmt.Print("hit <enter> to continue")
		fmt.Scanln(&ans)
		fmt.Print(ans)
	}
	fmt.Println()
	fmt.Println(" Finished.")
	fmt.Println()
}
