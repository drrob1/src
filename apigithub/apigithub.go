package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

//type Reply struct {  moved below to be an anonymous struct
//	Name string
//	//Public_Repos int  this worked, but I want to do it now w/ field tags, just to demonstrate this.
//	NumOfRepos int `json:"public_repos"`
//}

func main() {
	//respBody, er := io.ReadAll(resp.Body)
	//if er != nil {
	//	log.Fatalf(" io.ReadAll(resp.Body) err returned is: %s\n", er)
	//}
	//fmt.Printf(" respnse header content-type is: %#v\n", resp.Header.Get("Content-Type"))

	//err = os.WriteFile("api_github_com_drrob1.txt", respBody, 0666)
	//if err != nil {
	//	fmt.Printf(" Error from os.WriteFile(api-github) is: %s\n", err)
	//}

	//fmt.Printf(" %#v\n", r)

	name, publicRepos, err := githubInfo("drrob1")
	if err != nil {
		log.Fatalf(" err from githubInfo is: %s\n", err)
	}
	fmt.Printf(" github name: %s, number of public repos = %d\n", name, publicRepos)
}

// Will take code from main (above) and place it into this func.  This is a user exercise.
func githubInfo(login string) (string, int, error) {
	urL := "https://api.github.com/users/" + url.PathEscape(login) // this is to make sure that the login string will result in a valid url.
	resp, err := http.Get(urL)
	if err != nil {
		log.Printf(" Github api err returned is: %s\n", err)
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf(" Github api err returned is: %s\n", err)
		return "", 0, err
	}

	var r struct { // now it's an anonymous struct, which is fine since it's not needed anywhere else in this code.
		Name string
		//Public_Repos int  this worked, but I want to do it now w/ field tags, just to demonstrate this.
		NumOfRepos int `json:"public_repos"`
	}
	decod := json.NewDecoder(resp.Body)      // Remember that unused fields in either the json or the Go struct will be ignored without being an error.
	if err := decod.Decode(&r); err != nil { // since 'r' has to be changed by the decoder, it must take a pointer instead of a value.
		log.Fatalf(" can't decode, error is %s\n", err)
	}
	return r.Name, r.NumOfRepos, nil
}
