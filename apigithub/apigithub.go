package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	//respBody, er := io.ReadAll(resp.Body)
	//if er != nil {
	//	log.Fatalf(" io.ReadAll(resp.Body) err returned is: %s\n", er)
	//}
	//fmt.Printf(" respnse header content-type is: %#v\n", resp.Header.Get("Content-Type"))

	//err = os.WriteFile("api_github_com_drrob1.txt", respBody, 0666)
	//if err != nil {
	//	fmt.Printf(" Error from os.WriteFile(api-github) is: %s\n", err)
	//}

	type Reply struct {
		Name string
		//Public_Repos int  this worked, but I want to do it now w/ field tags, just to demonstrate this.
		NumOfRepos int `json:"public_repos"`
	}
	var r Reply

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&r); err != nil { // since 'r' has to be changed by the decoder, it must take a pointer instead of a value.
		log.Fatalf(" can't decode, error is %s\n", err)
	}
	fmt.Printf(" %#v\n", r)
}
