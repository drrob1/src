package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

//type Reply struct {  moved below to be an anonymous struct
//	Name string
//	//Public_Repos int  this worked, but I want to do it now w/ field tags, just to demonstrate this.
//	NumOfRepos int `json:"public_repos"`
//}

//  He later comes back to enhance this code to include a context timeout.  So he rewrites the githubInfo func.

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

	//name, publicRepos, err := githubInfo("drrob1")
	//fmt.Printf(" github name: %s, number of public repos = %d\n", name, publicRepos)
	//if err != nil {
	//	log.Fatalf(" err from githubInfo is: %s\n", err)
	//}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	fmt.Println(githubInfo(ctx, "drrob1"))
}

// Will take code from main (above) and place it into this func.  This is a user exercise.
func githubInfo(ctx context.Context, login string) (string, int, error) { // refactored to include a context time out
	urL := "https://api.github.com/users/" + url.PathEscape(login) // this is to make sure that the login string will result in a valid url.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urL, nil)
	if err != nil {
		log.Printf(" Github api err returned is: %s\n", err)
		return "", 0, err
	}

	resp, err := http.DefaultClient.Do(req)
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
		//Public_Repos int // this worked, but I want to do it now w/ field tags, just to demonstrate this.  And now I have both, to see of both work at same time.
		NumOfRepos int `json:"public_repos"`
		//NumOfRepos int  having both Public_Repos and NumOfRepos didn't work.  NumOfRepos was right but not the other oee.
	}
	decod := json.NewDecoder(resp.Body)      // Remember that unused fields in either the json or the Go struct will be ignored without being an error.
	if err := decod.Decode(&r); err != nil { // since 'r' has to be changed by the decoder, it must take a pointer instead of a value.
		log.Fatalf(" can't decode, error is %s\n", err)
	}
	//fmt.Printf(" in githubinfo: github name = %s, public_repos = %d, NumOfRepos = %d\n", r.Name, r.Public_Repos, r.NumOfRepos)
	return r.Name, r.NumOfRepos, nil
}

/*
func githubInfo(login string) (string, int, error) {  does not use a timeout.
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
		//Public_Repos int // this worked, but I want to do it now w/ field tags, just to demonstrate this.  And now I have both, to see of both work at same time.
		NumOfRepos int `json:"public_repos"`
		//NumOfRepos int  having both Public_Repos and NumOfRepos didn't work.  NumOfRepos was right but not the other oee.
	}
	decod := json.NewDecoder(resp.Body)      // Remember that unused fields in either the json or the Go struct will be ignored without being an error.
	if err := decod.Decode(&r); err != nil { // since 'r' has to be changed by the decoder, it must take a pointer instead of a value.
		log.Fatalf(" can't decode, error is %s\n", err)
	}
	//fmt.Printf(" in githubinfo: github name = %s, public_repos = %d, NumOfRepos = %d\n", r.Name, r.Public_Repos, r.NumOfRepos)
	return r.Name, r.NumOfRepos, nil
}
*/