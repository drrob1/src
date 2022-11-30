package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("https://api.github.com/users/drrob1")
	if err != nil {
		log.Fatalf(" Github api err returned is: %s\n", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf(" Github api err returned is: %s\n", err)
	}
	respBody, er := io.ReadAll(resp.Body)
	if er != nil {
		log.Fatalf(" io.ReadAll(resp.Body) err returned is: %s\n", er)
	}
	err = os.WriteFile("api-github-com-drrob1.txt", respBody, 0666)
	if err != nil {
		fmt.Printf(" Error from os.WriteFile(api-github) is: %s\n", err)
	}
}
